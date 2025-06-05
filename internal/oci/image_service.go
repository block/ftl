package oci

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	googleremote "github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
)

const SchemaLabel = "ftl.schema.digest"
const SchemaLocation = "deployments/ftl-full-schema.pb"

type ImageConfig struct {
	AllowInsecureImages bool     `help:"Allows the use of insecure HTTP based registries." env:"FTL_IMAGE_REPOSITORY_ALLOW_INSECURE"`
	Registry            Registry `help:"Registry to use for the image service." env:"FTL_IMAGE_REGISTRY"`
	RepositoryTemplate  string   `help:"Repository template to use for the image service." env:"FTL_IMAGE_REPOSITORY_TEMPLATE" default:"ftl-$${realm}-$${module}"`
	TagTemplate         string   `help:"Tag template to use for the image service." env:"FTL_IMAGE_TAG_TEMPLATE" default:"$${tag}"`
}

type ImageService struct {
	config   *ImageConfig
	puller   *googleremote.Puller
	logger   *log.Logger
	keyChain *keyChain
}

func NewImageService(ctx context.Context, config *ImageConfig) (*ImageService, error) {
	logger := log.FromContext(ctx)
	o := &ImageService{
		config: config,
		keyChain: &keyChain{
			resources:       map[string]*registryAuth{},
			originalContext: ctx,
		},
		logger: logger,
	}

	puller, err := googleremote.NewPuller(googleremote.WithAuthFromKeychain(o.keyChain))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create puller")
	}
	o.puller = puller
	return o, nil
}

type ImageTarget func(ctx context.Context, s *ImageService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error

func WithRemotePush() ImageTarget {
	return func(ctx context.Context, s *ImageService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error {
		logger := log.FromContext(ctx)
		repo, err := name.NewRepository(targetImage.Repository.String())
		if err != nil {
			return errors.Wrapf(err, "unable to parse repo")
		}
		authOpt := googleremote.WithAuthFromKeychain(s.keyChain)

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
	return func(ctx context.Context, s *ImageService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error {

		logger := log.FromContext(ctx)
		if _, err := daemon.Write(targetImage, image); err != nil {
			return errors.Errorf("writing layout: %w", err)
		}
		logger.Infof("Wrote image %s to local daemon", targetImage) //nolint
		return nil
	}
}

func WithDiskImage(path string) ImageTarget {
	return func(ctx context.Context, s *ImageService, targetImage name.Tag, imageIndex v1.ImageIndex, image v1.Image, layers []v1.Layer) error {

		logger := log.FromContext(ctx)
		file, err := os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", path)
		}
		if err := tarball.Write(targetImage, image, file); err != nil {
			return errors.Wrap(err, "writing layout")
		}
		logger.Infof("Wrote image %s to %s", targetImage, path) //nolint
		return nil
	}
}

func (s *ImageService) Image(realm, module, tag string) Image {
	expFunc := func(k string) string {
		switch k {
		case "realm":
			return realm
		case "module":
			return module
		case "tag":
			return tag
		}
		return ""
	}

	return Image(fmt.Sprintf("%s/%s:%s",
		s.config.Registry,
		os.Expand(s.config.RepositoryTemplate, expFunc),
		os.Expand(s.config.TagTemplate, expFunc),
	))
}

func (s *ImageService) BuildOCIImageFromRemote(
	ctx context.Context,
	artefactService *ArtefactService,
	baseImage string,
	targetImage Image,
	tempDir string,
	module *schema.Module,
	deployment key.Deployment,
	artifacts []*schema.MetadataArtefact,
	targets ...ImageTarget,
) error {
	target, err := os.MkdirTemp(tempDir, "ftl-image-")
	if err != nil {
		return errors.Wrapf(err, "unable to create temp dir in %s", tempDir)
	}
	defer os.RemoveAll(target)
	err = artefactService.DownloadArtifacts(ctx, target, artifacts)
	if err != nil {
		return errors.Wrapf(err, "failed to download artifacts")
	}

	schemaPath := filepath.Join(target, FTLFullSchemaPath)
	schemaBytes, err := os.ReadFile(schemaPath)
	if err == nil {
		// Only update the schema if it exists
		schpb := &schemapb.Schema{}
		if err := proto.Unmarshal(schemaBytes, schpb); err != nil {
			return errors.Wrapf(err, "failed to unmashal schema")
		}
		schema, err := schema.FromProto(schpb)
		if err != nil {
			return errors.Wrapf(err, "failed to unmashal schema")
		}
		if realm, ok := schema.FirstInternalRealm().Get(); ok {
			realm.UpsertModule(module)
		}
		bytes, err := proto.Marshal(schema.ToProto())
		if err != nil {
			return errors.Wrapf(err, "failed to marshal schema")
		}
		err = os.WriteFile(schemaPath, bytes, 0644) //nolint:gosec
		if err != nil {
			return errors.Wrapf(err, "failed to write schema")
		}
	} else {
		log.FromContext(ctx).Errorf(err, "Unable to update schema file")
	}

	if err != nil {
		return errors.Wrapf(err, "failed to download artifacts")
	}
	return s.BuildOCIImage(ctx, baseImage, targetImage, target, deployment, artifacts, targets...)

}

func (s *ImageService) BuildOCIImage(
	ctx context.Context,
	baseImage string,
	targetImage Image,
	apath string,
	deployment key.Deployment,
	allArtifacts []*schema.MetadataArtefact,
	targets ...ImageTarget,
) error {
	var artifacts []*schema.MetadataArtefact
	var schemaArtifacts []*schema.MetadataArtefact
	for _, i := range allArtifacts {
		if i.Path == FTLFullSchemaPath {
			schemaArtifacts = append(schemaArtifacts, i)
		} else {
			artifacts = append(artifacts, i)
		}
	}
	if len(schemaArtifacts) > 0 {
		err := enhanceSchemaMetadata(filepath.Join(apath, schemaArtifacts[0].Path), string(targetImage), deployment)
		if err != nil {
			return errors.Wrapf(err, "failed to enhance schema metadata with image and deployment information")
		}
	}

	opts := []name.Option{}
	// TODO: use http:// scheme for allow/disallow insecure
	if s.config.AllowInsecureImages {
		opts = append(opts, name.Insecure)
	}
	logger := log.FromContext(ctx)
	logger.Infof("Building %s with %s as a base image", targetImage, baseImage) //nolint
	ref, err := name.ParseReference(baseImage, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to parse image name")
	}
	targetRef, err := name.NewTag(string(targetImage))
	if err != nil {
		return errors.Wrapf(err, "failed to parse target image")
	}

	base, err := daemon.Image(ref)
	if err != nil {
		desc, err := googleremote.Get(ref, googleremote.WithContext(ctx), googleremote.WithAuthFromKeychain(s.keyChain), googleremote.Reuse(s.puller))
		if err != nil {
			return errors.Errorf("getting base image metadata: %w", err)
		}

		base, err = desc.Image()
		if err != nil {
			return errors.Errorf("loading base image: %w", err)
		}
	} else {
		logger.Infof("Using image %s from local docker daemon", ref.String()) //nolint
	}

	layer, err := createLayer(apath, artifacts)
	if err != nil {
		return errors.Errorf("creating layer: %w", err)
	}
	schLayer, err := createLayer(apath, schemaArtifacts)
	if err != nil {
		return errors.Errorf("creating layer: %w", err)
	}
	schDigest, err := schLayer.Digest()
	if err != nil {
		return errors.Errorf("getting schema layer digest: %w", err)
	}

	// Append the layer to the base image
	newImg, err := mutate.AppendLayers(base, layer, schLayer)
	if err != nil {
		return errors.Errorf("appending layer: %w", err)
	}

	cfg, err := newImg.ConfigFile()
	if err != nil {
		return errors.Errorf("getting config file: %w", err)
	}
	cfg.Config.Env = append(cfg.Config.Env, "FTL_SCHEMA_LOCATION=/deployments/ftl-full-schema.pb")
	cfg.Config.Env = append(cfg.Config.Env, fmt.Sprintf("FTL_DEPLOYMENT=%s", deployment.String()))
	cfg.Config.Env = append(cfg.Config.Env, "LOG_LEVEL=DEBUG")
	cfg.Config.Env = append(cfg.Config.Env, "LOG_FORMAT=json")
	cfg.Config.Labels[SchemaLabel] = schDigest.String()
	newImg, err = mutate.Config(newImg, cfg.Config)
	if err != nil {
		return errors.Errorf("setting environment var: %w", err)
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

func (s *ImageService) PullSchema(ctx context.Context, image string) (*schema.Schema, error) {
	opts := []name.Option{}
	// TODO: use http:// scheme for allow/disallow insecure
	if s.config.AllowInsecureImages {
		opts = append(opts, name.Insecure)
	}
	logger := log.FromContext(ctx)
	ref, err := name.ParseReference(image, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse image name")
	}

	img, err := daemon.Image(ref)
	if err != nil {
		desc, err := googleremote.Get(ref, googleremote.WithContext(ctx), googleremote.WithAuthFromKeychain(s.keyChain), googleremote.Reuse(s.puller))
		if err != nil {
			return nil, errors.Errorf("getting base image metadata: %w", err)
		}

		img, err = desc.Image()
		if err != nil {
			return nil, errors.Errorf("loading base image: %w", err)
		}
	} else {
		logger.Infof("Using image %s from local docker daemon", ref.String()) //nolint
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, errors.Errorf("getting config file: %w", err)
	}
	lbl, ok := cfg.Config.Labels[SchemaLabel]
	if !ok {
		return nil, errors.Errorf("image %s does not contain schema label %s", image, SchemaLabel)
	}
	var layer v1.Layer
	layers, err := img.Layers()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get layers from image")
	}
	for _, i := range layers {
		d, err := i.Digest()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get layer digest from image")
		}
		if d.String() == lbl {
			layer = i
			break
		}
	}
	if layer == nil {
		return nil, errors.Errorf("image %s does not contain schema layer with digest %s", image, lbl)
	}
	reader, err := layer.Uncompressed()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to uncompress layer from layer %s", lbl)
	}
	defer reader.Close() // nolint:errcheck
	tar := tar.NewReader(reader)
	hdr, err := tar.Next()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read tar header from layer %s", lbl)
	}
	if hdr.Name != SchemaLocation {
		return nil, errors.Errorf("expected schema file at %s, got %s", SchemaLocation, hdr.Name)
	}
	bytes, err := io.ReadAll(tar)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read schema file from layer %s", lbl)
	}
	sch := schemapb.Schema{}
	err = proto.Unmarshal(bytes, &sch)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal schema from from layer %s", lbl)
	}
	schema, err := schema.FromProto(&sch)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert schema proto to schema from from layer %s", lbl)
	}
	logger.Infof("Pulled schema from from layer %s", lbl) //nolint
	return schema, nil
}

func enhanceSchemaMetadata(path string, image string, deployment key.Deployment) error {
	schemaBytes, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read schema file %s", path)
	}
	schp := schemapb.Schema{}
	err = proto.Unmarshal(schemaBytes, &schp)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal schema")
	}
	sch, err := schema.FromProto(&schp)
	if err != nil {
		return errors.Wrapf(err, "failed to convert schema proto to schema")
	}
	it, ok := sch.FirstInternalRealm().Get()
	if !ok {
		return errors.Errorf("realm %s not found in schema", deployment.Payload.Realm)
	}
	module, ok := it.Module(deployment.Payload.Module).Get()
	if !ok {
		return errors.Errorf("module %s/%s not found in schema", deployment.Payload.Realm, deployment.Payload.Module)
	}
	module.Metadata = append(module.Metadata, &schema.MetadataImage{Image: image})
	module.ModRuntime().ModDeployment().DeploymentKey = deployment
	tw, err := proto.Marshal(sch.ToProto())
	if err != nil {
		return errors.Wrapf(err, "failed to marshal schema")
	}
	err = os.WriteFile(path, tw, 0644) //nolint:gosec
	if err != nil {
		return errors.Wrapf(err, "failed to write schema file %s", path)
	}
	return nil
}
