package main

import (
	"context"
	"os"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/runner"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	LogConfig           log.Config           `prefix:"log-" embed:""`
	ObservabilityConfig observability.Config `prefix:"o11y-" embed:""`
	RunnerConfig        runner.Config        `embed:""`
	DeploymentDir       string               `help:"Directory to store deployments in." default:"/deployments"`
	SchemaLocation      string               `help:"Location of the schema file." env:"FTL_SCHEMA_LOCATION"`
	RouteTemplate       string               `help:"Template to use to construct routes to other services" env:"FTL_ROUTE_TEMPLATE"`
	SecretsPath         string               `help:"Path to the directory containing secret files" env:"FTL_SECRETS_PATH" default:"/etc/ftl/secrets"`
	ConfigsPath         string               `help:"Path to the directory containing config files" env:"FTL_CONFIGS_PATH" default:"/etc/ftl/configs"`
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
	sch, err := schemaFromDisk(cli.SchemaLocation)
	kctx.FatalIfErrorf(err, "failed to load schema")
	var module *schema.Module
	found := ""
	for _, rlm := range sch.Realms {
		if rlm.External {
			continue
		}
		for _, mod := range rlm.Modules {
			found += " " + mod.Name
			if mod.Name == cli.RunnerConfig.Deployment.Payload.Module {
				module = mod
				break
			}
		}
	}
	if module == nil {
		kctx.Fatalf("Failed to find module %s in schema, found %s", cli.RunnerConfig.Deployment.Payload.Module, found)
	}

	secProvider := deploymentcontext.NewDiskProvider(cli.SecretsPath)
	configProvider := deploymentcontext.NewDiskProvider(cli.ConfigsPath)
	dp := deploymentcontext.NewProvider(cli.RunnerConfig.Deployment, deploymentcontext.NewTemplateRouter(cli.RouteTemplate, cli.RunnerConfig.Deployment.Payload.Realm), module, secProvider, configProvider)
	deploymentProvider := func() (string, error) {

		return cli.DeploymentDir, nil
	}
	err = runner.Start(ctx, cli.RunnerConfig, deploymentProvider, dp, module)
	kctx.FatalIfErrorf(err)
}

func schemaFromDisk(path string) (*schema.Schema, error) {
	pb, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load schema")
	}
	schemaProto := &schemapb.Schema{}
	err = proto.Unmarshal(pb, schemaProto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schema")
	}
	sch, err := schema.FromProto(schemaProto)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema")
	}
	return sch, nil
}
