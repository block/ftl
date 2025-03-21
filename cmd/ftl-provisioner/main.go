package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/backend/provisioner/scaling/k8sscaling"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
	timeline "github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version               kong.VersionFlag         `help:"Show version."`
	ObservabilityConfig   observability.Config     `embed:"" prefix:"o11y-"`
	LogConfig             log.Config               `embed:"" prefix:"log-"`
	ProvisionerConfig     provisioner.Config       `embed:""`
	ConfigFlag            string                   `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
	RegistryConfig        artefacts.RegistryConfig `prefix:"oci-" embed:""`
	InstanceName          string                   `help:"Instance name, use to differentiate ownership when there are multiple FTL instances ina cluster." env:"FTL_INSTANCE_NAME" default:"ftl"`
	UserNamespace         string                   `help:"Namespace to use for user resources." env:"FTL_USER_NAMESPACE"`
	ModulePerNamespace    bool                     `help:"If module per namespace mode is enabled" env:"FTL_MODULE_PER_NAMESPACE" default:"false"`
	CronServiceAccount    string                   `help:"Service account for cron." env:"FTL_CRON_SERVICE_ACCOUNT"`
	ConsoleServiceAccount string                   `help:"Service account for console." env:"FTL_CONSOLE_SERVICE_ACCOUNT"`
	AdminServiceAccount   string                   `help:"Service account for admin." env:"FTL_ADMIN_SERVICE_ACCOUNT"`
	HTTPServiceAccount    string                   `help:"Service account for http." env:"FTL_HTTP_SERVICE_ACCOUNT"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)
	cli.ProvisionerConfig.SetDefaults()

	logger := log.Configure(os.Stderr, cli.LogConfig).Scope("provisioner")
	ctx := log.ContextWithLogger(context.Background(), logger)
	timelineClient := timeline.NewClient(ctx, cli.ProvisionerConfig.TimelineEndpoint)
	err := observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.ProvisionerConfig.SchemaEndpoint.String(), log.Error)
	var mapper k8sscaling.NamespaceMapper
	if cli.ModulePerNamespace {
		mapper = func(module string, systemNamespace string) string {
			return module + "-ftl"
		}
	} else if cli.UserNamespace != "" {
		mapper = func(module string, systemNamespace string) string {
			return cli.UserNamespace
		}
	} else {
		mapper = func(module string, systemNamespace string) string {
			return systemNamespace
		}
	}
	scaling := k8sscaling.NewK8sScaling(false, cli.ProvisionerConfig.ControllerEndpoint.String(), cli.InstanceName, mapper, cli.CronServiceAccount, cli.AdminServiceAccount, cli.ConsoleServiceAccount, cli.HTTPServiceAccount)
	err = scaling.Start(ctx)
	kctx.FatalIfErrorf(err, "error starting k8s scaling")
	registry, err := provisioner.RegistryFromConfigFile(ctx, cli.ProvisionerConfig.WorkingDir, cli.ProvisionerConfig.PluginConfigFile, scaling)
	kctx.FatalIfErrorf(err, "failed to create provisioner registry")

	// Use in mem sql-migration provisioner as fallback for sql-migration provisioning if no other provisioner is registered
	if _, ok := slices.Find(registry.Bindings, func(binding *provisioner.ProvisionerBinding) bool {
		return slices.Contains(binding.Types, schema.ResourceTypeSQLMigration)
	}); !ok {
		storage, err := artefacts.NewOCIRegistryStorage(ctx, cli.RegistryConfig)
		kctx.FatalIfErrorf(err, "failed to create OCI registry storage")

		sqlMigrationProvisioner := provisioner.NewSQLMigrationProvisioner(storage)
		sqlMigrationBinding := registry.Register("in-mem-sql-migration", sqlMigrationProvisioner, schema.ResourceTypeSQLMigration)
		logger.Debugf("Registered provisioner %s as fallback for sql-migration", sqlMigrationBinding)
	}

	// Use k8s scaling as fallback for runner provisioning if no other provisioner is registered
	if _, ok := slices.Find(registry.Bindings, func(binding *provisioner.ProvisionerBinding) bool {
		return slices.Contains(binding.Types, schema.ResourceTypeRunner)
	}); !ok {
		runnerProvisioner := provisioner.NewRunnerScalingProvisioner(scaling)
		runnerBinding := registry.Register("kubernetes", runnerProvisioner, schema.ResourceTypeRunner)
		logger.Debugf("Registered provisioner %s as fallback for runner", runnerBinding)
	}

	err = provisioner.Start(ctx, registry, schemaClient, timelineClient)
	kctx.FatalIfErrorf(err, "failed to start provisioner")
}
