package main

import (
	"context"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type buildCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv    []string `help:"Environment variables to set for the build."`
}

func (b *buildCmd) Run(
	ctx context.Context,
	schemaSource *schemaeventsource.EventSource,
	projConfig projectconfig.Config,
) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.WithStack(errors.New("no directories specified"))
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "build stopped"))
	engine, err := buildengine.NewV2(
		ctx,
		schemaSource,
		projConfig,
		b.Dirs,
		false,
		buildengine.BuildEnvV2(b.BuildEnv),
		buildengine.ParallelismV2(b.Parallelism),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := engine.BuildV2(ctx, true); err != nil {
		return errors.Wrap(err, "build failed")
	}
	return nil
}
