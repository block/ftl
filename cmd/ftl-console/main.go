package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/console"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	Bind                *url.URL             `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ConsoleConfig       console.Config       `embed:"" prefix:"console-"`
	TimelineEndpoint    *url.URL             `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	AdminEndpoint       *url.URL             `help:"Admin endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	BuildEngineEndpoint *url.URL             `help:"Build engine endpoint." env:"FTL_BUILD_UPDATES_ENDPOINT" default:"http://127.0.0.1:8900"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Console`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig).Scope("console"))
	err := observability.Init(ctx, false, "", "ftl-console", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	timelineClient := timelineclient.NewClient(ctx, cli.TimelineEndpoint)
	adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)
	buildEngineClient := rpc.Dial(buildenginepbconnect.NewBuildEngineServiceClient, cli.BuildEngineEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, "console", adminClient)

	routeManager := routing.NewVerbRouter(ctx, eventSource, timelineClient)

	svc := console.New(eventSource, timelineClient, adminClient, routeManager, buildEngineClient, cli.Bind, cli.ConsoleConfig)
	err = rpc.Serve(ctx, cli.Bind, rpc.WithServices(svc))
	kctx.FatalIfErrorf(err, "failed to start console service")
}
