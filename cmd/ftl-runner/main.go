package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/runner"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
)

var cli struct {
	Version             kong.VersionFlag         `help:"Show version."`
	LogConfig           log.Config               `prefix:"log-" embed:""`
	RegistryConfig      artefacts.RegistryConfig `prefix:"oci-" embed:""`
	ObservabilityConfig observability.Config     `prefix:"o11y-" embed:""`
	RunnerConfig        runner.Config            `embed:""`
}

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	kctx := kong.Parse(&cli, kong.Description(`
FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.

The Runner is the component of FTL that coordinates with the Controller to spawn
and route to user code.
	`), kong.Vars{
		"version":       ftl.Version,
		"deploymentdir": filepath.Join(cacheDir, "ftl-runner", "${runner}", "deployments"),
	})
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err = observability.Init(ctx, false, "", "ftl-runner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialise observability")
	storage, err := artefacts.NewOCIRegistryStorage(ctx, cli.RegistryConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")
	// Substitute in the runner key into the deployment directory.
	cli.RunnerConfig.DeploymentDir = os.Expand(cli.RunnerConfig.DeploymentDir, func(key string) string {
		if key == "runner" {
			return cli.RunnerConfig.Key.String()
		}
		return key
	})

	err = runner.Start(ctx, cli.RunnerConfig, storage)
	kctx.FatalIfErrorf(err)
}
