package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/backend/provisioner/scaling/k8sscaling"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/oci"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
	timeline "github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version               kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig   observability.Config `embed:"" prefix:"o11y-"`
	LogConfig             log.Config           `embed:"" prefix:"log-"`
	ProvisionerConfig     provisioner.Config   `embed:""`
	ConfigFlag            string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
	ArtefactConfig        oci.ArtefactConfig   `prefix:"oci-artefact-" embed:""`
	ImageConfig           oci.ImageConfig      `prefix:"oci-image-" embed:""`
	InstanceName          string               `help:"Instance name, use to differentiate ownership when there are multiple FTL instances ina cluster." env:"FTL_INSTANCE_NAME" default:"ftl"`
	UserNamespace         string               `help:"Namespace to use for user resources." env:"FTL_USER_NAMESPACE"`
	CronServiceAccount    string               `help:"Service account for cron." env:"FTL_CRON_SERVICE_ACCOUNT"`
	ConsoleServiceAccount string               `help:"Service account for console." env:"FTL_CONSOLE_SERVICE_ACCOUNT"`
	AdminServiceAccount   string               `help:"Service account for admin." env:"FTL_ADMIN_SERVICE_ACCOUNT"`
	HTTPServiceAccount    string               `help:"Service account for http." env:"FTL_HTTP_SERVICE_ACCOUNT"`
	DefaultRunnerImage    string               `help:"Default image to use for a runner." env:"FTL_DEFAULT_RUNNER_IMAGE" default:"ftl0/ftl-runner"`
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
	adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, cli.ProvisionerConfig.AdminEndpoint.String(), log.Error)
	err := observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.ProvisionerConfig.SchemaEndpoint.String(), log.Error)
	var mapper k8sscaling.NamespaceMapper
	var routeTemplate string
	if cli.UserNamespace != "" {
		routeTemplate = fmt.Sprintf("http://${module}.%s:8892", cli.UserNamespace)
		mapper = func(module string, realm string, systemNamespace string) string {
			return cli.UserNamespace
		}
	} else {
		routeTemplate = "http://${module}.${module}-${realm}:8892"
		mapper = func(module string, realm string, systemNamespace string) string {
			return module + "-" + realm
		}
	}

	artefactService, err := oci.NewArtefactService(ctx, cli.ArtefactConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")

	imageService, err := oci.NewImageService(ctx, artefactService, &cli.ImageConfig)
	kctx.FatalIfErrorf(err, "failed to create image service")

	scaling := k8sscaling.NewK8sScaling(false, cli.InstanceName, mapper, routeTemplate, cli.CronServiceAccount, cli.AdminServiceAccount, cli.ConsoleServiceAccount, cli.HTTPServiceAccount)
	err = scaling.Start(ctx)
	kctx.FatalIfErrorf(err, "error starting k8s scaling")
	registry, err := provisioner.RegistryFromConfigFile(ctx, cli.ProvisionerConfig.WorkingDir, cli.ProvisionerConfig.PluginConfigFile, scaling, adminClient, imageService)
	kctx.FatalIfErrorf(err, "failed to create provisioner registry")

	// Use in mem sql-migration provisioner as fallback for sql-migration provisioning if no other provisioner is registered
	if _, ok := slices.Find(registry.Bindings, func(binding *provisioner.ProvisionerBinding) bool {
		return slices.Contains(binding.Types, schema.ResourceTypeSQLMigration)
	}); !ok {

		sqlMigrationProvisioner := provisioner.NewSQLMigrationProvisioner(artefactService)
		sqlMigrationBinding := registry.Register("in-mem-sql-migration", sqlMigrationProvisioner, schema.ResourceTypeSQLMigration)
		logger.Debugf("Registered provisioner %s as fallback for sql-migration", sqlMigrationBinding)
	}

	// Use default image provisioner
	if _, ok := slices.Find(registry.Bindings, func(binding *provisioner.ProvisionerBinding) bool {
		return slices.Contains(binding.Types, schema.ResourceTypeImage)
	}); !ok {
		ociProvisioner := provisioner.NewOCIImageProvisioner(imageService, cli.DefaultRunnerImage)
		runnerBinding := registry.Register("oci-image", ociProvisioner, schema.ResourceTypeImage)
		logger.Debugf("Registered provisioner %s as fallback for image", runnerBinding)
	}

	// Use k8s scaling as fallback for runner provisioning if no other provisioner is registered
	if _, ok := slices.Find(registry.Bindings, func(binding *provisioner.ProvisionerBinding) bool {
		return slices.Contains(binding.Types, schema.ResourceTypeRunner)
	}); !ok {
		runnerProvisioner := provisioner.NewRunnerScalingProvisioner(scaling, false)
		runnerBinding := registry.Register("kubernetes", runnerProvisioner, schema.ResourceTypeRunner)
		logger.Debugf("Registered provisioner %s as fallback for runner", runnerBinding)
	}

	err = provisioner.Start(ctx, registry, schemaClient, timelineClient)
	kctx.FatalIfErrorf(err, "failed to start provisioner")
}
