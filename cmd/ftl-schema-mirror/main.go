package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/schemamirror"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	Bind                *url.URL             `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)
	logger := log.Configure(os.Stderr, cli.LogConfig).Scope("schemamirror")
	ctx := log.ContextWithLogger(context.Background(), logger)
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM)
	defer cancel()
	err := observability.Init(ctx, false, "", "ftl-schema-mirror", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	svc := schemamirror.New(ctx)
	err = rpc.Serve(ctx, cli.Bind, rpc.GRPC(ftlv1connect.NewSchemaMirrorServiceHandler, svc), rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc))
	logger.Debugf("Listening on %s", cli.Bind)
	kctx.FatalIfErrorf(err)
}
