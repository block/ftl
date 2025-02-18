package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type buildCmd struct {
	Parallelism     int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs            []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv        []string `help:"Environment variables to set for the build."`
	UpdatesEndpoint *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8900" env:"FTL_BUILD_UPDATES_ENDPOINT"`
}

func (b *buildCmd) Run(
	ctx context.Context,
	adminClient ftlv1connect.AdminServiceClient,
	schemaSource *schemaeventsource.EventSource,
	projConfig projectconfig.Config,
) error {
	logger := log.FromContext(ctx)
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(fmt.Errorf("build stopped: %w", context.Canceled))
	engine, err := buildengine.New(
		ctx,
		adminClient,
		schemaSource,
		projConfig,
		b.Dirs,
		b.UpdatesEndpoint,
		buildengine.BuildEnv(b.BuildEnv),
		buildengine.Parallelism(b.Parallelism),
	)
	if err != nil {
		return err
	}
	if len(engine.Modules()) == 0 {
		logger.Warnf("No modules were found to build")
		return nil
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}
