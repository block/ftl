package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/controller"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	_ "github.com/block/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ControllerConfig    controller.Config    `embed:""`
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
	DisableIstio        bool                 `help:"Disable Istio integration. This will prevent the creation of Istio policies to limit network traffic." env:"FTL_DISABLE_ISTIO"`
	LeaseEndpoint       *url.URL             `help:"Lease endpoint." env:"FTL_LEASE_ENDPOINT" default:"http://127.0.0.1:8895"`
	AdminEndpoint       *url.URL             `help:"Admin endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	SchemaEndpoint      *url.URL             `help:"Schema endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8897"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)
	cli.ControllerConfig.SetDefaults()

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	leaseClient := rpc.Dial(leasepbconnect.NewLeaseServiceClient, cli.LeaseEndpoint.String(), log.Error)

	ctx = rpc.ContextWithClient(ctx, leaseClient)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, schemaClient)

	adminClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, adminClient)

	err = controller.Start(ctx, cli.ControllerConfig, adminClient, schemaClient, false)
	kctx.FatalIfErrorf(err)
}
