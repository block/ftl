package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/pgproxy"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`

	pgproxy.Config
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig).Scope("proxy-pg"))
	err := observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	proxy := pgproxy.New(cli.Config.Listen, func(ctx context.Context, params map[string]string) (string, error) {
		return "postgres://localhost:5432/postgres?user=" + params["user"], nil
	})
	if err := proxy.Start(ctx, nil); err != nil {
		kctx.FatalIfErrorf(err, "failed to start proxy")
	}
}
