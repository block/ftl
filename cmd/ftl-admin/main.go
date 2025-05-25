package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/configuration/routers"
	"github.com/block/ftl/internal/kube"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Bind                *url.URL                 `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Version             kong.VersionFlag         `help:"Show version."`
	ObservabilityConfig observability.Config     `embed:"" prefix:"o11y-"`
	LogConfig           log.Config               `embed:"" prefix:"log-"`
	AdminConfig         admin.Config             `embed:"" prefix:"admin-"`
	SchemaEndpoint      *url.URL                 `help:"Schema endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineEndpoint    *url.URL                 `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
	UserNamespace       string                   `help:"Namespace to use for kube user resources." env:"FTL_USER_NAMESPACE"`
	RegistryConfig      artefacts.RegistryConfig `embed:"" prefix:"oci-"`
	Realm               string                   `help:"Realm to use for the admin service." env:"FTL_REALM"`
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

	cs, err := kube.CreateClientSet()
	kctx.FatalIfErrorf(err, "failed to initialize kube client")

	mapper := kube.NewNamespaceMapper(cli.UserNamespace)

	configResolver := routers.NewKubeConfigRouter(cs, mapper, cli.Realm)
	cm, err := manager.New(ctx, configResolver, providers.NewInline[cf.Configuration]())
	kctx.FatalIfErrorf(err)

	// FTL currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := providers.NewASM(secretsmanager.NewFromConfig(awsConfig))
	dbSecretResolver := routers.NewKubeSecretRouter(cs, mapper, cli.Realm)
	sm, err := manager.New(ctx, dbSecretResolver, asmSecretProvider)
	kctx.FatalIfErrorf(err)

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, "admin", schemaClient)

	storage, err := artefacts.NewOCIRegistryStorage(ctx, cli.RegistryConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")
	client := timelineclient.NewClient(ctx, cli.TimelineEndpoint)
	svc := admin.NewAdminService(cli.AdminConfig, cm, sm, schemaClient, eventSource, storage, routing.NewVerbRouter(ctx, eventSource, client), client, []string{})

	kctx.FatalIfErrorf(err, "failed to start admin service handlers")
	logger.Debugf("Admin service listening on: %s", cli.Bind)
	err = rpc.Serve(ctx, cli.Bind,
		rpc.WithServices(svc),
	)
	kctx.FatalIfErrorf(err, "failed to start admin service")
}
