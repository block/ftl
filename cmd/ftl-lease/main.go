package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/lease"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	LeaseConfig         lease.Config         `embed:"" `
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

	err = lease.Start(ctx, cli.LeaseConfig)
	kctx.FatalIfErrorf(err, "failed to start lease service")
}
