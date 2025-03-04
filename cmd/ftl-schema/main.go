package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	SchemaServiceConfig schemaservice.Config `embed:""`
	TimelineEndpoint    *url.URL             `help:"Timeline Service endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8898"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	timelineClient := timelineclient.NewClient(ctx, cli.TimelineEndpoint)

	err = schemaservice.Start(ctx, cli.SchemaServiceConfig, timelineClient, false)
	kctx.FatalIfErrorf(err)
}
