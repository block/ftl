package oci

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/atomic"
	errors "github.com/alecthomas/errors"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	googleremote "github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
)

const (
	FTLFullSchemaPath = "ftl-full-schema.pb"
)

type DeploymentArtefactProvider func() (string, error)

type ArtefactReader interface {
	io.ReadCloser
}

// Metadata container for an artefact's metadata
type Metadata struct {
	Executable bool
	Size       int64
	Path       string
}

type ArtefactUpload struct {
	Digest  sha256.SHA256
	Size    int64
	Content io.ReadCloser
}

// Artefact container for an artefact's payload and metadata
type Artefact struct {
	Digest   sha256.SHA256
	Metadata Metadata
	Content  io.ReadCloser
}

type ArtefactKey struct {
	Digest sha256.SHA256
}

type ReleaseArtefact struct {
	Artefact   ArtefactKey
	Path       string
	Executable bool
}

// Registry is a string that represents an OCI container registry.
// For example, "123456789012.dkr.ecr.us-west-2.amazonaws.com"
type Registry string

// Repository is a string that represents an OCI container repository.
// For example, "123456789012.dkr.ecr.us-west-2.amazonaws.com/ftl-tests"
type Repository string

type RepositoryConfig struct {
	Repository    Repository `help:"OCI container repository, in the form host[:port]/repository" env:"FTL_ARTEFACT_REPOSITORY"`
	Username      string     `help:"OCI container repository username" env:"FTL_ARTEFACT_REPOSITORY_USERNAME"`
	Password      string     `help:"OCI container repository password" env:"FTL_ARTEFACT_REPOSITORY_PASSWORD"`
	AllowInsecure bool       `help:"Allows the use of insecure HTTP based registries." env:"FTL_ARTEFACT_REPOSITORY_ALLOW_INSECURE"`
}

type ArtefactService struct {
	keyChain *keyChain

	puller       *googleremote.Puller
	targetConfig RepositoryConfig
	logger       *log.Logger
}

type ArtefactRepository struct {
	ModuleDigest     sha256.SHA256
	MediaType        string
	ArtefactType     string
	RepositoryDigest digest.Digest
	Size             int64
}

type ArtefactBlobs struct {
	Digest    sha256.SHA256
	MediaType string
	Size      int64
}

type registryAuth struct {
	delegate authn.Authenticator
	auth     atomic.Value[*authn.AuthConfig]
}

// Authorization implements authn.Authenticator.
func (r *registryAuth) Authorization() (*authn.AuthConfig, error) {
	if r.delegate != nil {
		auth, err := r.delegate.Authorization()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to authorize container registry")
		}
		if auth != nil {
			return auth, nil
		}
	}
	return r.auth.Load(), nil
}

func NewNewArtefactServiceForTesting() *ArtefactService {
	storage, err := NewArtefactService(context.TODO(), RepositoryConfig{Repository: "127.0.0.1:15000/ftl-tests", AllowInsecure: true})
	if err != nil {
		panic(err)
	}
	return storage
}

func NewArtefactService(ctx context.Context, config RepositoryConfig) (*ArtefactService, error) {
	logger := log.FromContext(ctx)
	o := &ArtefactService{
		keyChain: &keyChain{
			repositories:    map[string]*registryAuth{},
			targetConfig:    config,
			originalContext: ctx,
		},
		targetConfig: config,
		logger:       logger,
	}

	puller, err := googleremote.NewPuller(googleremote.WithAuthFromKeychain(o.keyChain))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create puller")
	}
	o.puller = puller
	return o, nil
}

func (s *ArtefactService) GetRepository() Repository {
	return s.targetConfig.Repository
}

func (s *ArtefactService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	repo, err := s.repoFactory()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to connect to container repository '%s'", s.targetConfig.Repository)
	}
	set := make(map[sha256.SHA256]bool)
	for _, d := range digests {
		set[d] = true
	}
	keys = make([]ArtefactKey, 0)
	blobs := repo.Blobs()
	for _, d := range digests {
		_, err := blobs.Resolve(ctx, fmt.Sprintf("sha256:%s", d))
		if err != nil {
			if errors.Is(err, errdef.ErrNotFound) {
				continue
			}
			return nil, nil, errors.Wrapf(err, "unable to resolve digest '%s'", d)
		}
		keys = append(keys, ArtefactKey{Digest: d})
		delete(set, d)
	}
	missing = make([]sha256.SHA256, 0)
	for d := range set {
		missing = append(missing, d)
	}
	return keys, missing, nil
}

// Upload uploads the specific artifact as a raw blob and links it to a manifest to prevent GC
func (s *ArtefactService) Upload(ctx context.Context, artefact ArtefactUpload) error {
	repo, err := s.repoFactory()
	logger := log.FromContext(ctx).Scope("oci:" + artefact.Digest.String())
	if err != nil {
		return errors.Wrapf(err, "unable to connect to repository '%s'", s.targetConfig.Repository)
	}

	parseSHA256, err := sha256.ParseSHA256(artefact.Digest.String())
	if err != nil {
		return errors.Wrap(err, "unable to parse sh")
	}

	logger.Debugf("Pushing artefact blob")
	contentDesc := ocispec.Descriptor{
		MediaType: "application/vnd.ftl.artifact",
		Digest:    digest.Digest("sha256:" + artefact.Digest.String()),
		Size:      artefact.Size,
	}
	err = repo.Push(ctx, contentDesc, artefact.Content)
	if err != nil {
		return errors.Wrap(err, "unable to push to in memory repositor")
	}

	tag := contentDesc.Digest.Hex()
	artefact.Digest = parseSHA256

	logger.Debugf("Tagging module blob with digest '%s'", tag)
	fileDescriptors := []ocispec.Descriptor{contentDesc}
	config := ocispec.ImageConfig{} // Create a new image config
	config.Labels = map[string]string{"type": "ftl-artifact"}
	configBlob, err := json.Marshal(config) // Marshal the config to json
	if err != nil {
		return errors.Wrap(err, "unable to marshal OCI image config")
	}
	configDesc, err := pushBlob(ctx, ocispec.MediaTypeImageConfig, configBlob, repo) // push config blob
	if err != nil {
		return errors.Wrap(err, "unable to push OCI image config to OCI registry")
	}

	manifestBlob, err := generateManifestContent(configDesc, fileDescriptors...)
	if err != nil {
		return errors.Wrap(err, "unable to generate manifest content")
	}
	manifestDesc, err := pushBlob(ctx, ocispec.MediaTypeImageManifest, manifestBlob, repo) // push manifest blob
	if err != nil {
		return errors.Wrap(err, "unable to push manifest to OCI registry")
	}
	if err = repo.Tag(ctx, manifestDesc, tag); err != nil {
		return errors.Wrap(err, "unable to tag OCI registry")
	}

	return nil
}

func (s *ArtefactService) Download(ctx context.Context, dg sha256.SHA256) (io.ReadCloser, error) {
	// ORAS is really annoying, and needs you to know the size of the blob you're downloading
	// So we are using google's go-containerregistry to do the actual download
	// This is not great, we should remove oras at some point
	opts := []name.Option{}
	if s.targetConfig.AllowInsecure {
		opts = append(opts, name.Insecure)
	}
	newDigest, err := name.NewDigest(fmt.Sprintf("%s@sha256:%s", s.targetConfig.Repository, dg.String()), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create digest '%s'", dg)
	}
	layer, err := googleremote.Layer(newDigest, googleremote.WithAuthFromKeychain(s.keyChain), googleremote.Reuse(s.puller))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read layer '%s'", newDigest)
	}
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read uncompressed layer '%s'", newDigest)
	}
	return uncompressed, nil
}

func (s *ArtefactService) repoFactory() (*remote.Repository, error) {
	reg, err := remote.NewRepository(string(s.targetConfig.Repository))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to container repository '%s'", s.targetConfig.Repository)
	}

	ref, err := name.NewRepository(string(s.targetConfig.Repository))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse registry")
	}
	a, err := s.keyChain.Resolve(ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve authenticator")
	}
	acfg, err := a.Authorization()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to authenticate")
	}

	s.logger.Debugf("Connecting to repository '%s'", s.targetConfig.Repository)
	reg.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: func(ctx context.Context, hostport string) (auth.Credential, error) {
			return auth.Credential{
				Username: acfg.Username,
				Password: acfg.Password,
			}, nil
		},
	}
	reg.PlainHTTP = s.targetConfig.AllowInsecure
	return reg, nil
}

func pushBlob(ctx context.Context, mediaType string, blob []byte, target oras.Target) (desc ocispec.Descriptor, err error) {
	desc = ocispec.Descriptor{ // Generate descriptor based on the media type and blob content
		MediaType: mediaType,
		Digest:    digest.FromBytes(blob), // Calculate digest
		Size:      int64(len(blob)),       // Include blob size
	}
	err = target.Push(ctx, desc, bytes.NewReader(blob)) // Push the blob to the registry target
	if err != nil {
		return desc, errors.Wrap(err, "unable to push blob")
	}
	return desc, nil
}

func generateManifestContent(config ocispec.Descriptor, layers ...ocispec.Descriptor) ([]byte, error) {
	content := ocispec.Manifest{
		Config:    config, // Set config blob
		Layers:    layers, // Set layer blobs
		Versioned: specs.Versioned{SchemaVersion: 2},
	}
	json, err := json.Marshal(content)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal manifest content")
	}
	return json, nil // Get json content
}

// DownloadArtifacts downloads artefacts for a deployment from an OCI registry.
func (s *ArtefactService) DownloadArtifacts(ctx context.Context, dest string, artifacts []*schema.MetadataArtefact) error {
	logger := log.FromContext(ctx)
	start := time.Now()
	count := 0
	for _, artefact := range artifacts {
		res, err := s.Download(ctx, artefact.Digest)
		if err != nil {
			return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
		}
		count++
		if !filepath.IsLocal(artefact.Path) {
			return errors.Errorf("path %q is not local", artefact.Path)
		}
		logger.Debugf("Downloading %s", filepath.Join(dest, artefact.Path))
		err = os.MkdirAll(filepath.Join(dest, filepath.Dir(artefact.Path)), 0700)
		if err != nil {
			return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
		}
		var mode os.FileMode = 0600
		if artefact.Executable {
			mode = 0755
		}
		w, err := os.OpenFile(filepath.Join(dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
		}
		defer w.Close()
		buf := make([]byte, 1024)
		read := 0
		for {
			read, err = res.Read(buf)
			if read > 0 {
				_, e2 := w.Write(buf[:read])
				if e2 != nil {
					return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return errors.Wrapf(err, "failed to download artifact %q", artefact.Digest)
			}
		}
	}
	logger.Debugf("Downloaded %d artefacts in %s", count, time.Since(start))
	return nil
}

// createLayer returns a v1.Layer with a single text file.
func createLayer(path string, artifacts []*schema.MetadataArtefact) (v1.Layer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, a := range artifacts {
		if err := addFileToTar(tw, path, a.Path, a.Executable); err != nil {
			return nil, errors.Wrapf(err, "failed to add file to layer")
		}
	}
	if err := tw.Close(); err != nil {
		return nil, errors.Wrapf(err, "failed to create layer")
	}
	// TODO: use a file
	return tarball.LayerFromReader(&buf) //nolint
}

// addFileToTar adds a single file to the tar writer.
func addFileToTar(tw *tar.Writer, basepath string, path string, execuable bool) error {
	file, err := os.Open(filepath.Join(basepath, path))
	if err != nil {
		return errors.Wrapf(err, "failed to open file")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return errors.Wrapf(err, "failed to stat file")
	}
	if stat.IsDir() {
		return errors.Errorf("directories not supported: %s", path)
	}

	mode := int64(0644)
	if execuable {
		mode = 0755
	}

	// TODO: hard coded deployments path
	hdr := &tar.Header{
		Name:    "deployments/" + path,
		Mode:    mode,
		Size:    stat.Size(),
		ModTime: time.Now(),
		Uid:     1000,
		Gid:     1000,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return errors.Wrap(err, "failed to write tar header")
	}

	_, err = io.Copy(tw, file)
	return errors.Wrap(err, "failed to copy files to tar")
}
