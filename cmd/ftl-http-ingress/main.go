package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/ingress"
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
	Bind                 *url.URL              `help:"Socket to bind to for ingress." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Version              kong.VersionFlag      `help:"Show version."`
	ObservabilityConfig  observability.Config  `embed:"" prefix:"o11y-"`
	LogConfig            log.Config            `embed:"" prefix:"log-"`
	HTTPIngressConfig    ingress.Config        `embed:""`
	SchemaServerEndpoint *url.URL              `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineConfig       timelineclient.Config `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - HTTP Ingress`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig).Scope("http-ingress"))
	err := observability.Init(ctx, false, "", "ftl-http-ingress", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaServerEndpoint.String(), log.Error)
	timelineClient := timelineclient.NewClient(ctx, cli.TimelineConfig)
	eventSource := schemaeventsource.New(ctx, "http-ingress", schemaClient)
	routeManager := routing.NewVerbRouter(ctx, eventSource, timelineClient)
	err = ingress.Start(ctx, cli.Bind, cli.HTTPIngressConfig, eventSource, routeManager, timelineClient)
	kctx.FatalIfErrorf(err, "failed to start HTTP ingress")
}
