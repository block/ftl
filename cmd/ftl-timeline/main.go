package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Bind                *url.URL             `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	TimelineConfig      timeline.Config      `embed:"" prefix:"timeline-"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Timeline`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	logger := log.Configure(os.Stderr, cli.LogConfig).Scope("timeline")
	ctx := log.ContextWithLogger(context.Background(), logger)
	err := observability.Init(ctx, false, "", "ftl-timeline", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	svc, err := timeline.New(ctx, cli.TimelineConfig)
	kctx.FatalIfErrorf(err, "failed to create timeline service")
	logger.Debugf("Timeline Service listening on: %s", cli.Bind)
	err = rpc.Serve(ctx, cli.Bind, rpc.WithServices(svc))
	kctx.FatalIfErrorf(err, "failed to start timeline service")
}
