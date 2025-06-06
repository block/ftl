package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/oci"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type imageBuildCmd struct {
	Parallelism     int             `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs            []string        `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv        []string        `help:"Environment variables to set for the build."`
	ImageConfig     oci.ImageConfig `embed:""`
	Tag             string          `help:"The image tag" default:"latest"`
	RunnerImage     string          `help:"An override of the runner base image"`
	Push            bool            `help:"Push the image to the registry after building." default:"false"`
	SkipLocalDaemon bool            `help:"Skip pushing to the local docker daemon." default:"false"`
	TarFile         string          `help:"File system path to push the image to"`
	Deploy          bool            `help:"Deploy the images after they are built." default:"false"`
}

func (b *imageBuildCmd) Run(
	ctx context.Context,
	adminClient adminpbconnect.AdminServiceClient,
	schemaSource *schemaeventsource.EventSource,
	projConfig projectconfig.Config,
) error {
	logger := log.FromContext(ctx)
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.WithStack(errors.New("no directories specified"))
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "build stopped"))
	engine, err := buildengine.New(
		ctx,
		adminClient,
		schemaSource,
		projConfig,
		b.Dirs,
		false,
		buildengine.BuildEnv(b.BuildEnv),
		buildengine.Parallelism(b.Parallelism),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(engine.Modules()) == 0 {
		logger.Warnf("No modules were found to build")
		return nil
	}
	imageService, err := oci.NewImageService(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to init OCI")
	}
	var images []string
	if err := engine.BuildWithCallback(ctx, func(ctx context.Context, module buildengine.Module, moduleSch *schema.Module, tmpDeployDir string, deployPaths []string) error {
		artifacts := []*schema.MetadataArtefact{}

		for _, i := range deployPaths {
			s, err := os.Stat(i)
			if err != nil {
				return errors.Wrapf(err, "failed to stat file")
			}

			path, err := filepath.Rel(tmpDeployDir, i)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve file")
			}
			executable := s.Mode().Perm()&0111 != 0
			bytes, err := os.ReadFile(i)
			if err != nil {
				return errors.Wrapf(err, "failed to read file %s", path)
			}

			artifacts = append(artifacts, &schema.MetadataArtefact{Path: path, Executable: executable, Digest: sha256.Sum(bytes)})
		}
		var baseImage string
		if b.RunnerImage != "" {
			baseImage = b.RunnerImage
		} else {
			baseImage = "ftl0/ftl-runner"
			if moduleSch.ModRuntime().Base.Image != "" {
				baseImage = moduleSch.ModRuntime().Base.Image
			}
			baseImage += ":"
			if ftl.IsRelease(ftl.Version) && ftl.Version == ftl.BaseVersion(ftl.Version) {
				baseImage += "v"
				baseImage += ftl.Version
			} else {
				baseImage += "latest"
			}
		}
		tgt, err := imageService.Image(b.ImageConfig, projConfig.Name, moduleSch.Name, b.Tag)
		if err != nil {
			return errors.Wrapf(err, "failed to get image target for %s", moduleSch.Name)
		}
		baseRef, err := imageService.ParseName(baseImage, b.ImageConfig.AllowInsecureImages)
		if err != nil {
			return errors.Wrapf(err, "failed to parse base image name %s", baseImage)
		}
		moduleSch.Metadata = append(moduleSch.Metadata, &schema.MetadataImage{Image: tgt.String()})
		targets := []oci.ImageTarget{}
		if !b.SkipLocalDaemon {
			targets = append(targets, oci.WithLocalDeamon())
		}
		if b.Push || b.Deploy {
			targets = append(targets, oci.WithRemotePush())
		}
		if b.TarFile != "" {
			targets = append(targets, oci.WithDiskImage(b.TarFile))
		}
		// TODO: we need to properly sync the deployment with the actual deployment key
		// this is just a hack to get the module and realm to the runner
		deployment := key.NewDeploymentKey(projConfig.Name, moduleSch.Name)
		err = imageService.BuildOCIImage(ctx, moduleSch, baseRef, tgt, tmpDeployDir, deployment, artifacts, nil, targets...)
		if err != nil {
			return errors.Wrapf(err, "failed to build image")
		}
		images = append(images, tgt.String())
		return nil
	}); err != nil {
		return errors.Wrap(err, "build failed")
	}
	if b.Deploy {
		return errors.Wrapf(deployImages(ctx, adminClient, images, b.ImageConfig.AllowInsecureImages), "failed to deploy images %v", images)
	}
	return nil
}
