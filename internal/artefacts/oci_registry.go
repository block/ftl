package artefacts

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/atomic"
	errors "github.com/alecthomas/errors"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
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

type RegistryConfig struct {
	Registry      string `help:"OCI container registry, in the form host[:port]/repository" env:"FTL_ARTEFACT_REGISTRY" required:""`
	Username      string `help:"OCI container registry username" env:"FTL_ARTEFACT_REGISTRY_USERNAME"`
	Password      string `help:"OCI container registry password" env:"FTL_ARTEFACT_REGISTRY_PASSWORD"`
	AllowInsecure bool   `help:"Allows the use of insecure HTTP based registries." env:"FTL_ARTEFACT_REGISTRY_ALLOW_INSECURE"`
}

type OCIArtefactService struct {
	originalContext context.Context
	puller          *googleremote.Puller
	targetConfig    RegistryConfig
	logger          *log.Logger
	registries      map[string]*registryAuth
	registryLock    sync.Mutex
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

func NewForTesting() *OCIArtefactService {
	storage, err := NewOCIRegistryStorage(context.TODO(), RegistryConfig{Registry: "127.0.0.1:15000/ftl-tests", AllowInsecure: true})
	if err != nil {
		panic(err)
	}
	return storage
}

func isECRRepository(repo string) bool {
	ecrRegex := regexp.MustCompile(`(?i)^\d{12}\.dkr\.ecr\.[a-z0-9-]+\.amazonaws\.com/`)
	return ecrRegex.MatchString(repo)
}

func NewOCIRegistryStorage(ctx context.Context, config RegistryConfig) (*OCIArtefactService, error) {
	// Connect the registry targeting the specified container

	logger := log.FromContext(ctx)
	o := &OCIArtefactService{
		originalContext: ctx,
		registries:      map[string]*registryAuth{},
		targetConfig:    config,
		logger:          logger,
	}

	puller, err := googleremote.NewPuller(googleremote.WithAuthFromKeychain(o))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create puller")
	}
	o.puller = puller
	return o, nil
}

// Resolve implements authn.Keychain.
func (s *OCIArtefactService) Resolve(r authn.Resource) (authn.Authenticator, error) {
	s.registryLock.Lock()
	defer s.registryLock.Unlock()

	logger := log.FromContext(s.originalContext)
	registry := r.String()
	existing := s.registries[registry]
	if existing != nil {
		return existing, nil
	}
	cfg := &registryAuth{}
	s.registries[registry] = cfg
	cfg.auth.Store(&authn.AuthConfig{})

	if registry == s.targetConfig.Registry &&
		s.targetConfig.Username != "" &&
		s.targetConfig.Password != "" {
		// The user has explicitly supplied credentials, lets use them
		cfg.auth.Store(&authn.AuthConfig{
			Username: s.targetConfig.Username,
			Password: s.targetConfig.Password,
		})
		return cfg, nil
	}

	dctx, err := authn.DefaultKeychain.ResolveContext(s.originalContext, r)
	if err == nil {
		// Local docker config takes precidence
		cfg.delegate = dctx
		return cfg, nil
	}

	if isECRRepository(registry) {

		username, password, err := getECRCredentials(s.originalContext)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		logger.Debugf("Using ECR credentials for registry '%s'", registry)
		cfg.auth.Store(&authn.AuthConfig{Username: username, Password: password})
		go func() {
			for {
				select {
				case <-s.originalContext.Done():
					return
				case <-time.After(time.Hour):
					username, password, err := getECRCredentials(s.originalContext)
					if err != nil {
						logger.Errorf(err, "failed to refresh ECR credentials")
					}
					cfg.auth.Store(&authn.AuthConfig{Username: username, Password: password})
				}
			}
		}()
	}
	return cfg, nil
}

func (s *OCIArtefactService) GetRegistry() string {
	return s.targetConfig.Registry
}

func getECRCredentials(ctx context.Context) (string, string, error) {
	// Load AWS Config
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to load AWS config")
	}

	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)
	// Get authorization token
	resp, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get authorization token")
	}

	if len(resp.AuthorizationData) == 0 {
		return "", "", errors.Wrap(err, "no authorization data")
	}
	authData := resp.AuthorizationData[0]
	token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to decode auth token")
	}

	splitToken := strings.SplitN(string(token), ":", 2)
	if len(splitToken) != 2 {
		return "", "", errors.Wrap(err, "failed to decode auth token due to invalid format")
	}

	username := splitToken[0]
	password := splitToken[1]
	return username, password, nil
}

func (s *OCIArtefactService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	repo, err := s.repoFactory()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to connect to container registry '%s'", s.targetConfig.Registry)
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
func (s *OCIArtefactService) Upload(ctx context.Context, artefact ArtefactUpload) error {
	repo, err := s.repoFactory()
	logger := log.FromContext(ctx).Scope("oci:" + artefact.Digest.String())
	if err != nil {
		return errors.Wrapf(err, "unable to connect to repository '%s'", s.targetConfig.Registry)
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

func (s *OCIArtefactService) Download(ctx context.Context, dg sha256.SHA256) (io.ReadCloser, error) {
	// ORAS is really annoying, and needs you to know the size of the blob you're downloading
	// So we are using google's go-containerregistry to do the actual download
	// This is not great, we should remove oras at some point
	opts := []name.Option{}
	if s.targetConfig.AllowInsecure {
		opts = append(opts, name.Insecure)
	}
	newDigest, err := name.NewDigest(fmt.Sprintf("%s@sha256:%s", s.targetConfig.Registry, dg.String()), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create digest '%s'", dg)
	}
	layer, err := googleremote.Layer(newDigest, googleremote.WithAuthFromKeychain(s), googleremote.Reuse(s.puller))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read layer '%s'", newDigest)
	}
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read uncompressed layer '%s'", newDigest)
	}
	return uncompressed, nil
}

func (s *OCIArtefactService) repoFactory() (*remote.Repository, error) {
	reg, err := remote.NewRepository(s.targetConfig.Registry)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to container registry '%s'", s.targetConfig.Registry)
	}

	ref, err := name.NewRepository(s.targetConfig.Registry)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse registry")
	}
	a, err := s.Resolve(ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve authenticator")
	}
	acfg, err := a.Authorization()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to authenticate")
	}

	s.logger.Debugf("Connecting to registry '%s'", s.targetConfig.Registry)
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
func (s *OCIArtefactService) DownloadArtifacts(ctx context.Context, dest string, artifacts []*schema.MetadataArtefact) error {
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

func WithRemotePush() ImageTarget {
	return func(ctx context.Context, s *OCIArtefactService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error {
		logger := log.FromContext(ctx)
		repo, err := name.NewRepository(s.targetConfig.Registry)
		if err != nil {
			return errors.Wrapf(err, "unable to parse repo")
		}
		authOpt := googleremote.WithAuthFromKeychain(s)

		for _, l := range layers {
			if err := googleremote.WriteLayer(repo, l, authOpt); err != nil {
				return errors.Errorf("writing layer: %w", err)
			}
		}
		// Also push up any other layers
		existing, err := image.Layers()
		if err != nil {
			return errors.Wrapf(err, "unable to get image layers")
		}

		for _, l := range existing {
			if err := googleremote.WriteLayer(repo, l, authOpt); err != nil {
				return errors.Errorf("writing layer: %w", err)
			}
		}
		if err := googleremote.Write(targetImage, image, authOpt); err != nil {
			return errors.Errorf("writing image: %w", err)
		}
		logger.Infof("Wrote image %s to remote repository", targetImage) //nolint
		return nil
	}
}

func WithLocalDeamon() ImageTarget {
	return func(ctx context.Context, s *OCIArtefactService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error {

		logger := log.FromContext(ctx)
		if _, err := daemon.Write(targetImage, image); err != nil {
			return errors.Errorf("writing layout: %w", err)
		}
		logger.Infof("Wrote image %s to local daemon", targetImage) //nolint
		return nil
	}
}

type ImageTarget func(ctx context.Context, s *OCIArtefactService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error

func (s *OCIArtefactService) BuildOCIImageFromRemote(ctx context.Context, baseImage string, targetImage string, tempDir string, artifacts []*schema.MetadataArtefact, targets ...ImageTarget) error {
	target, err := os.MkdirTemp(tempDir, "ftl-image-")
	if err != nil {
		return errors.Wrapf(err, "unable to create temp dir in %s", tempDir)
	}
	defer os.RemoveAll(target)
	err = s.DownloadArtifacts(ctx, target, artifacts)
	if err != nil {
		return err
	}
	return s.BuildOCIImage(ctx, baseImage, targetImage, target, artifacts, targets...)

}

func (s *OCIArtefactService) BuildOCIImage(ctx context.Context, baseImage string, targetImage string, apath string, allArtifacts []*schema.MetadataArtefact, targets ...ImageTarget) error {
	var artifacts []*schema.MetadataArtefact
	var schemaArtifacts []*schema.MetadataArtefact
	for _, i := range allArtifacts {
		if i.Path == FTLFullSchemaPath {
			schemaArtifacts = append(schemaArtifacts, i)
		} else {
			artifacts = append(artifacts, i)
		}
	}

	opts := []name.Option{}
	// TODO: use http:// scheme for allow/disallow insecure
	if s.targetConfig.AllowInsecure {
		opts = append(opts, name.Insecure)
	}
	logger := log.FromContext(ctx)
	logger.Infof("Building %s with %s as a base image", targetImage, baseImage) //nolint
	ref, err := name.ParseReference(baseImage, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to parse image name")
	}
	targetRef, err := name.NewTag(targetImage)
	if err != nil {
		return errors.Wrapf(err, "failed to parse target image")
	}

	desc, err := googleremote.Get(ref, googleremote.WithContext(ctx), googleremote.WithAuthFromKeychain(s), googleremote.Reuse(s.puller))
	if err != nil {
		return errors.Errorf("getting base image metadata: %w", err)
	}

	base, err := desc.Image()
	if err != nil {
		return errors.Errorf("loading base image: %w", err)
	}

	layer, err := createLayer(apath, artifacts)
	if err != nil {
		return errors.Errorf("creating layer: %w", err)
	}
	schLayer, err := createLayer(apath, schemaArtifacts)
	if err != nil {
		return errors.Errorf("creating layer: %w", err)
	}

	// Append the layer to the base image
	newImg, err := mutate.AppendLayers(base, layer, schLayer)
	if err != nil {
		return errors.Errorf("appending layer: %w", err)
	}

	idx := mutate.AppendManifests(empty.Index, mutate.IndexAddendum{Add: newImg})

	for _, i := range targets {
		err = i(ctx, s, targetRef, idx, newImg, []v1.Layer{layer})
		if err != nil {
			return errors.Wrapf(err, "failed to write image")
		}
	}

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
