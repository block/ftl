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

func NewOCIImageProvisioner(storage *oci.ImageService, defaultImage string) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeImage: provisionOCIImage(storage, defaultImage),
	}, map[schema.ResourceType]InMemResourceProvisionerFn{})
}

func provisionOCIImage(storage *oci.ImageService, defaultImage string) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, moduleSch *schema.Module) (*schema.RuntimeElement, error) {
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

		tag := "latest"
		git, ok := slices.FindVariant[*schema.MetadataGit](moduleSch.Metadata)

		if ok {
			tag = git.Commit
		}

		target := string(storage.Image(deployment.Payload.Realm, deployment.Payload.Module, tag))
		err = storage.BuildOCIImageFromRemote(ctx, image, target, tempDir, moduleSch, variants, oci.WithRemotePush())
		if err != nil {
			return nil, errors.Wrap(err, "failed to build image")
		}
		return &schema.RuntimeElement{
			Deployment: deployment,
			Element: &schema.ModuleRuntimeImage{
				Image: target,
			},
		}, nil
	}
}
