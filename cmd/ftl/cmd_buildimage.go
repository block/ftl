package main

import (
	"context"
	"os"
	"path/filepath"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/artefacts"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type buildImageCmd struct {
	Parallelism     int                      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs            []string                 `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv        []string                 `help:"Environment variables to set for the build."`
	RegistryConfig  artefacts.RegistryConfig `embed:""`
	Tag             string                   `help:"The image tag" default:"latest"`
	RunnerImage     string                   `help:"An override of the runner base image"`
	Push            bool                     `help:"Push the image to the registry after building." default:"false"`
	SkipLocalDaemon bool                     `help:"Skip pushing to the local docker daemon." default:"false"`
}

func (b *buildImageCmd) Run(
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
	service, err := artefacts.NewOCIRegistryStorage(ctx, b.RegistryConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to init OCI")
	}
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
			artifacts = append(artifacts, &schema.MetadataArtefact{Path: path, Executable: executable})
		}
		var image string
		if b.RunnerImage != "" {
			image = b.RunnerImage
		} else {
			image = "ftl0/ftl-runner"
			if moduleSch.ModRuntime().Base.Image != "" {
				image = moduleSch.ModRuntime().Base.Image
			}
			image += ":"
			if ftl.IsRelease(ftl.Version) && ftl.Version == ftl.BaseVersion(ftl.Version) {
				image += ftl.Version
			} else {
				image += "latest"
			}
		}
		tgt := b.RegistryConfig.Registry
		tgt += ":"
		tgt += b.Tag
		targets := []artefacts.ImageTarget{}
		if !b.SkipLocalDaemon {
			targets = append(targets, artefacts.WithLocalDeamon())
		}
		if b.Push {
			targets = append(targets, artefacts.WithRemotePush())
		}
		err := service.BuildOCIImage(ctx, image, tgt, tmpDeployDir, artifacts, targets...)
		if err != nil {
			return errors.Wrapf(err, "failed to build image")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "build failed")
	}
	return nil
}
