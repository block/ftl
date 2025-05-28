package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	Bind                *url.URL             `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	SchemaServiceConfig schemaservice.Config `embed:""`
	TimelineEndpoint    *url.URL             `help:"Timeline Service endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8898"`
	MirrorEndpoint      *url.URL             `help:"Schema Mirror endpoint." env:"FTL_SCHEMA_MIRROR_ENDPOINT"`
	Realm               string               `help:"Realm to use." env:"FTL_REALM" default:"ftl"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)
	logger := log.Configure(os.Stderr, cli.LogConfig).Scope("schema")
	ctx := log.ContextWithLogger(context.Background(), logger)
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer cancel()
	err := observability.Init(ctx, false, "", "ftl-schema", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	timelineClient := timelineclient.NewClient(ctx, cli.TimelineEndpoint)

	var pushClient optional.Option[ftlv1connect.SchemaMirrorServiceClient]
	if cli.MirrorEndpoint != nil {
		pushClient = optional.Some(rpc.Dial(ftlv1connect.NewSchemaMirrorServiceClient, cli.MirrorEndpoint.String(), log.Debug))
	}

	svc := schemaservice.New(ctx, cli.SchemaServiceConfig, timelineClient, pushClient, cli.Realm, false)
	err = rpc.Serve(ctx, cli.Bind, rpc.WithServices(svc))
	logger.Debugf("Listening on %s", cli.Bind)
	kctx.FatalIfErrorf(err)
}
