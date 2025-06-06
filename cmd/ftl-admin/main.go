package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/config/kubeconfig"
	"github.com/block/ftl/internal/kube"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/oci"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Bind                *url.URL              `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Version             kong.VersionFlag      `help:"Show version."`
	ObservabilityConfig observability.Config  `embed:"" prefix:"o11y-"`
	LogConfig           log.Config            `embed:"" prefix:"log-"`
	AdminConfig         admin.Config          `embed:"" prefix:"admin-"`
	SchemaEndpoint      *url.URL              `help:"Schema endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineConfig      timelineclient.Config `embed:""`
	Realm               string                `help:"Realm name." env:"FTL_REALM" required:""`
	Config              string                `help:"Path to FTL configuration file." env:"FTL_CONFIG" required:""`
	RegistryConfig      oci.ArtefactConfig    `embed:"" prefix:"oci-"`
	UseASM              bool                  `help:"Use AWS Secrets Manager to administer secrets" default:"false" env:"FTL_USE_ASM"`
	KubeConfig          kube.KubeConfig       `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Admin`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err := observability.Init(ctx, false, "", "ftl-admin", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	kubeClient, err := kube.CreateClientSet()
	kctx.FatalIfErrorf(err, "failed to initialize kube client")
	mapper := cli.KubeConfig.NamespaceMapper()
	cm := kubeconfig.NewKubeConfigProvider(kubeClient, mapper, cli.Realm)

	var sm config.Provider[config.Secrets]
	if cli.UseASM {
		// FTL currently only supports AWS Secrets Manager as a secrets provider.
		awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
		kctx.FatalIfErrorf(err)
		asm := config.NewASM(cli.Realm, secretsmanager.NewFromConfig(awsConfig))
		sm = config.NewCacheDecorator(ctx, asm)
	} else {
		sm = kubeconfig.NewKubeSecretProvider(kubeClient, mapper, cli.Realm)
	}
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, "admin", schemaClient)

	storage, err := oci.NewArtefactService(ctx, cli.RegistryConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")
	client := timelineclient.NewClient(ctx, cli.TimelineConfig)
	svc := admin.NewAdminService(cli.AdminConfig, cm, sm, schemaClient, eventSource, storage, routing.NewVerbRouter(ctx, eventSource, client), client, []string{})

	kctx.FatalIfErrorf(err, "failed to start admin service handlers")
	logger.Debugf("Admin service listening on: %s", cli.Bind)
	err = rpc.Serve(ctx, cli.Bind,
		rpc.WithServices(svc),
	)
	kctx.FatalIfErrorf(err, "failed to start admin service")
}
