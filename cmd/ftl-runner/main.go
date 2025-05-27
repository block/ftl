package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/runner"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	LogConfig           log.Config           `prefix:"log-" embed:""`
	ObservabilityConfig observability.Config `prefix:"o11y-" embed:""`
	RunnerConfig        runner.Config        `embed:""`
	DeploymentDir       string               `help:"Directory to store deployments in." default:"/deployments"`
}

func main() {
	kctx := kong.Parse(&cli, kong.Description(`
FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.

The Runner is the component of FTL that coordinates with the Controller to spawn
and route to user code.
	`), kong.Vars{
		"version": ftl.Version,
	})
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err := observability.Init(ctx, false, "", "ftl-runner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialise observability")
	// Substitute in the runner key into the deployment directory.
	cli.DeploymentDir = os.Expand(cli.DeploymentDir, func(key string) string {
		if key == "runner" {
			return cli.RunnerConfig.Key.String()
		}
		return key
	})

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.RunnerConfig.SchemaEndpoint.String(), log.Error)
	adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, cli.RunnerConfig.AdminEndpoint.String(), log.Error)
	routeTable := routing.New(ctx, schemaeventsource.New(ctx, "runner-deployment-context", schemaClient))
	dp, err := deploymentcontext.NewAdminProvider(ctx, cli.RunnerConfig.Deployment, routeTable, schemaClient, adminClient)
	kctx.FatalIfErrorf(err)
	deploymentProvider := func() (string, error) {

		return cli.DeploymentDir, nil
	}
	err = runner.Start(ctx, cli.RunnerConfig, deploymentProvider, dp, schemaClient)
	kctx.FatalIfErrorf(err)
}
