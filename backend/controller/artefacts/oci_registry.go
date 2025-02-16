package artefacts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/atomic"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	googleremote "github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/log"
)

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
	auth          *atomic.Value[authn.AuthConfig]
	puller        *googleremote.Puller
	registry      string
	allowInsecure bool
	logger        *log.Logger
}

func (s *OCIArtefactService) Authorization() (*authn.AuthConfig, error) {
	out := s.auth.Load()
	return &out, nil
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
		auth:          &atomic.Value[authn.AuthConfig]{},
		registry:      config.Registry,
		allowInsecure: config.AllowInsecure,
		logger:        logger,
	}
	puller, err := googleremote.NewPuller(googleremote.WithAuth(o))
	if err != nil {
		return nil, fmt.Errorf("unable to create puller for registry '%s': %w", config.Registry, err)
	}
	o.puller = puller

	if isECRRepository(config.Registry) {

		username, password, err := getECRCredentials(ctx)
		if err != nil {
			return nil, err
		}
		logger.Debugf("Using ECR credentials for registry '%s'", config.Registry)
		o.auth.Store(authn.AuthConfig{Username: username, Password: password})
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Hour):
					username, password, err := getECRCredentials(ctx)
					if err != nil {
						logger.Errorf(err, "failed to refresh ECR credentials")
					}
					o.auth.Store(authn.AuthConfig{Username: username, Password: password})
				}
			}
		}()
	} else {
		o.auth.Store(authn.AuthConfig{Username: config.Username, Password: config.Password})
	}
	return o, nil
}

func getECRCredentials(ctx context.Context) (string, string, error) {
	// Load AWS Config
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)
	// Get authorization token
	resp, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get authorization token: %w", err)
	}

	if len(resp.AuthorizationData) == 0 {
		return "", "", fmt.Errorf("no authorization data: %w", err)
	}
	authData := resp.AuthorizationData[0]
	token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode auth token: %w", err)
	}

	splitToken := strings.SplitN(string(token), ":", 2)
	if len(splitToken) != 2 {
		return "", "", fmt.Errorf("failed to decode auth token due to invalid format: %w", err)
	}

	username := splitToken[0]
	password := splitToken[1]
	return username, password, nil
}

func (s *OCIArtefactService) GetDigestsKeys(ctx context.Context, digests []sha256.SHA256) (keys []ArtefactKey, missing []sha256.SHA256, err error) {
	repo, err := s.repoFactory()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to container registry '%s': %w", s.registry, err)
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
			return nil, nil, fmt.Errorf("unable to resolve digest '%s': %w", d, err)
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
		return fmt.Errorf("unable to connect to repository '%s': %w", s.registry, err)
	}

	parseSHA256, err := sha256.ParseSHA256(artefact.Digest.String())
	if err != nil {
		return fmt.Errorf("unable to parse sha %w", err)
	}

	logger.Debugf("Pushing artefact blob")
	contentDesc := ocispec.Descriptor{
		MediaType: "application/vnd.ftl.artifact",
		Digest:    digest.Digest("sha256:" + artefact.Digest.String()),
		Size:      artefact.Size,
	}
	err = repo.Push(ctx, contentDesc, artefact.Content)
	if err != nil {
		return fmt.Errorf("unable to push to in memory repository %w", err)
	}

	tag := contentDesc.Digest.Hex()
	artefact.Digest = parseSHA256

	logger.Debugf("Tagging module blob with digest '%s'", tag)
	fileDescriptors := []ocispec.Descriptor{contentDesc}
	config := ocispec.ImageConfig{} // Create a new image config
	config.Labels = map[string]string{"type": "ftl-artifact"}
	configBlob, err := json.Marshal(config) // Marshal the config to json
	if err != nil {
		return fmt.Errorf("unable to marshal OCI image config: %w", err)
	}
	configDesc, err := pushBlob(ctx, ocispec.MediaTypeImageConfig, configBlob, repo) // push config blob
	if err != nil {
		return fmt.Errorf("unable to push OCI image config to OCI registry: %w", err)
	}

	manifestBlob, err := generateManifestContent(configDesc, fileDescriptors...)
	if err != nil {
		return fmt.Errorf("unable to generate manifest content: %w", err)
	}
	manifestDesc, err := pushBlob(ctx, ocispec.MediaTypeImageManifest, manifestBlob, repo) // push manifest blob
	if err != nil {
		return fmt.Errorf("unable to push manifest to OCI registry: %w", err)
	}
	if err = repo.Tag(ctx, manifestDesc, tag); err != nil {
		return fmt.Errorf("unable to tag OCI registry: %w", err)
	}

	return nil
}

func (s *OCIArtefactService) Download(ctx context.Context, dg sha256.SHA256) (io.ReadCloser, error) {
	// ORAS is really annoying, and needs you to know the size of the blob you're downloading
	// So we are using google's go-containerregistry to do the actual download
	// This is not great, we should remove oras at some point
	opts := []name.Option{}
	if s.allowInsecure {
		opts = append(opts, name.Insecure)
	}
	newDigest, err := name.NewDigest(fmt.Sprintf("%s@sha256:%s", s.registry, dg.String()), opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create digest '%s': %w", dg, err)
	}
	auth := s.auth.Load()
	layer, err := googleremote.Layer(newDigest, googleremote.WithAuth(authn.FromConfig(auth)), googleremote.Reuse(s.puller))
	if err != nil {
		return nil, fmt.Errorf("unable to read layer '%s': %w", newDigest, err)
	}
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("unable to read uncompressed layer '%s': %w", newDigest, err)
	}
	return uncompressed, nil
}

func (s *OCIArtefactService) repoFactory() (*remote.Repository, error) {
	reg, err := remote.NewRepository(s.registry)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to container registry '%s': %w", s.registry, err)
	}

	a := s.auth.Load()
	s.logger.Debugf("Connecting to registry '%s'", s.registry)
	reg.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: func(ctx context.Context, hostport string) (auth.Credential, error) {
			return auth.Credential{
				Username: a.Username,
				Password: a.Password,
			}, nil
		},
	}
	reg.PlainHTTP = s.allowInsecure
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
		return desc, fmt.Errorf("unable to push blob: %w", err)
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
		return nil, fmt.Errorf("unable to marshal manifest content: %w", err)
	}
	return json, nil // Get json content
}
