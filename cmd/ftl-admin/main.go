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
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var cli struct {
	Version             kong.VersionFlag         `help:"Show version."`
	ObservabilityConfig observability.Config     `embed:"" prefix:"o11y-"`
	LogConfig           log.Config               `embed:"" prefix:"log-"`
	AdminConfig         admin.Config             `embed:"" prefix:"admin-"`
	SchemaEndpoint      *url.URL                 `help:"Schema endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8897"`
	TimelineEndpoint    *url.URL                 `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	Config              string                   `help:"Path to FTL configuration file." env:"FTL_CONFIG" required:""`
	Secrets             string                   `help:"Path to FTL secrets file." env:"FTL_SECRETS" required:""`
	RegistryConfig      artefacts.RegistryConfig `embed:"" prefix:"oci-"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Admin`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
		},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-admin", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	configResolver := routers.NewFileRouter[cf.Configuration](cli.Config)
	cm, err := manager.New(ctx, configResolver, providers.NewInline[cf.Configuration]())
	kctx.FatalIfErrorf(err)

	// FTL currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := providers.NewASM(secretsmanager.NewFromConfig(awsConfig))
	dbSecretResolver := routers.NewFileRouter[cf.Secrets](cli.Secrets)
	sm, err := manager.New(ctx, dbSecretResolver, asmSecretProvider)
	kctx.FatalIfErrorf(err)

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, "admin", schemaClient)

	storage, err := artefacts.NewOCIRegistryStorage(ctx, cli.RegistryConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")
	err = admin.Start(ctx, cli.AdminConfig, cm, sm, schemaClient, eventSource, timelineclient.NewClient(ctx, cli.TimelineEndpoint), storage, nil)
	kctx.FatalIfErrorf(err, "failed to start timeline service")
}
