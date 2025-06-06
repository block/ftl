package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/cron"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	CronConfig          cron.Config          `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Cron`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig).Scope("cron"))
	err := observability.Init(ctx, false, "", "ftl-cron", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.CronConfig.SchemaServiceEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, "cron", schemaClient)

	timelineClient := timelineclient.NewClient(ctx, cli.CronConfig.TimelineConfig)
	routeManager := routing.NewVerbRouter(ctx, eventSource, timelineClient)

	err = cron.Start(ctx, cli.CronConfig, eventSource, routeManager, timelineClient)
	kctx.FatalIfErrorf(err, "failed to start cron")
}
