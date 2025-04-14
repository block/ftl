package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/lease"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	Bind                *url.URL             `help:"Socket to bind to." default:"http://127.0.0.1:8895" env:"FTL_BIND"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Lease`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig).Scope("lease"))
	err := observability.Init(ctx, false, "", "ftl-lease", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	svc := lease.New(ctx)
	err = rpc.Serve(ctx, cli.Bind, rpc.WithServices(svc))
	kctx.FatalIfErrorf(err, "failed to start lease service")
}
