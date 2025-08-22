package provisioner

import (
	"context"
	"os"
	goslices "slices"
	"strings"

	"github.com/alecthomas/errors"
	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/oci"
)

type OCIImageProvisionerConfig struct {
	Env map[string]string `toml:"env"`
}

func NewOCIImageProvisioner(storage *oci.ImageService, astorage *oci.ArtefactService, defaultImage string, config oci.ImageConfig, cfg OCIImageProvisionerConfig) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeImage: provisionOCIImage(storage, astorage, defaultImage, config, cfg),
	}, map[schema.ResourceType]InMemResourceProvisionerFn{})
}

func provisionOCIImage(storage *oci.ImageService, astorage *oci.ArtefactService, defaultImage string, config oci.ImageConfig, cfg OCIImageProvisionerConfig) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, moduleSch *schema.Module) (*schema.RuntimeElement, error) {
		images := slices.FilterVariants[*schema.MetadataImage](moduleSch.Metadata)
		for img := range images {
			if img.Image != "" {
				return &schema.RuntimeElement{
					Deployment: deployment,
					Element: &schema.ModuleRuntimeImage{
						Image: img.Image,
					},
				}, nil
			}
		}
		if moduleSch.GetRuntime().Image != nil && moduleSch.GetRuntime().Image.Image != "" {
			// We already have an image defined in the runtime module schema, so we don't need to build one.
			return nil, nil
		}
		logger := log.FromContext(ctx)
		variants := goslices.Collect(slices.FilterVariants[*schema.MetadataArtefact](moduleSch.Metadata))

		tempDir, err := os.MkdirTemp("", "image")
		if err != nil {
			return nil, errors.Wrap(err, "unable to create temp dir")
		}

		image := defaultImage
		if moduleSch.ModRuntime().Base.Image != "" {
			image = moduleSch.ModRuntime().Base.Image
			image = strings.ReplaceAll(image, "ftl0/ftl-runner", defaultImage)
		}

		image += ":"
		if !ftl.IsRelease(ftl.Version) || ftl.Version != ftl.BaseVersion(ftl.Version) {
			image += "latest"
		} else {
			image += "v"
			image += ftl.Version
		}
		logger.Debugf("Using base image %s from default %s", image, defaultImage)
		imageRef, err := storage.ParseName(image, config.AllowInsecureImages)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse image name %s", image)
		}

		tag := "latest"
		git, ok := slices.FindVariant[*schema.MetadataGit](moduleSch.Metadata)

		if ok {
			tag = git.Commit
		}

		target, err := storage.Image(config, deployment.Payload.Realm, deployment.Payload.Module, tag)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get image target for %s", deployment.Payload.Module)
		}
		err = storage.BuildOCIImageFromRemote(ctx, astorage, imageRef, target, tempDir, moduleSch, deployment, variants, cfg.Env, oci.WithRemotePush())
		if err != nil {
			return nil, errors.Wrap(err, "failed to build image")
		}
		return &schema.RuntimeElement{
			Deployment: deployment,
			Element: &schema.ModuleRuntimeImage{
				Image: target.String(),
			},
		}, nil
	}
}
