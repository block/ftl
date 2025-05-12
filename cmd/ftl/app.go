package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/trace"
	"strconv"
	"strings"
	"syscall"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	"github.com/alecthomas/types/optional"
	kongcompletion "github.com/jotaen/kong-completion"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/editor"
	"github.com/block/ftl/internal/log"
	_ "github.com/block/ftl/internal/prodinit" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
	"github.com/block/ftl/internal/timelineclient"
)

type SharedCLI struct {
	Version          kong.VersionFlag `help:"Show version."`
	Project          string           `short:"P" name:"project" help:"Path to FTL project root directory. The git root will be used if not found in the current directory." env:"FTL_PROJECT" placeholder:"DIR" default:""`
	ConfigFlag       string           `name:"config" short:"C" help:"Path to FTL project configuration file." env:"FTL_CONFIG" placeholder:"FILE"`
	TimelineEndpoint *url.URL         `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
	AdminEndpoint    *url.URL         `help:"Admin endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	Trace            string           `help:"File to write golang runtime/trace output to." hidden:""`

	Ping      pingCmd      `cmd:"" help:"Ping the FTL cluster."`
	Init      initCmd      `cmd:"" help:"Initialize a new FTL project."`
	Profile   profileCmd   `cmd:"" help:"Manage profiles."`
	Module    moduleCmd    `cmd:"" help:"Manage modules."`
	PS        psCmd        `cmd:"" help:"List deployments."`
	Call      callCmd      `cmd:"" help:"Call an FTL verb."`
	Changeset changesetCmd `cmd:"" help:"Work with changesets."`
	Bench     benchCmd     `cmd:"" help:"Benchmark an FTL verb."`
	Replay    replayCmd    `cmd:"" help:"Call an FTL verb with the same request body as the last invocation."`
	Update    updateCmd    `cmd:"" help:"Update a deployment."`
	Kill      killCmd      `cmd:"" help:"Kill a deployment."`
	Schema    schemaCmd    `cmd:"" help:"FTL schema commands."`
	Download  downloadCmd  `cmd:"" help:"Download a deployment."`
	Secret    secretCmd    `cmd:"" help:"Manage secrets."`
	Config    configCmd    `cmd:"" help:"Manage configuration."`
	Pubsub    pubsubCmd    `cmd:"" help:"Manage pub/sub."`
	Goose     gooseCmd     `cmd:"" help:"Run a goose command."`
	Mysql     mySQLCmd     `cmd:"" help:"Manage MySQL databases."`
	Postgres  postgresCmd  `cmd:"" help:"Manage PostgreSQL databases."`
	Edit      editCmd      `cmd:"" help:"Edit a declaration in an IDE."`
	Realm     realmCmd     `cmd:"" help:"Manage realms." hidden:""`
}

type BuildAndDeploy struct {
	Build  buildCmd  `cmd:"" help:"Build all modules found in the specified directories."`
	Deploy deployCmd `cmd:"" help:"Build and deploy all modules found in the specified directories."`
}

type CLI struct {
	SharedCLI
	BuildAndDeploy
	LogConfig log.Config `embed:"" prefix:"log-" group:"Logging:"`

	Authenticators map[string]string `help:"Authenticators to use for FTL endpoints." mapsep:"," env:"FTL_AUTHENTICATORS" placeholder:"HOST=EXE,…"`
	Insecure       bool              `help:"Skip TLS certificate verification. Caution: susceptible to machine-in-the-middle attacks."`
	Plain          bool              `help:"Use a plain console with no color or status line." env:"FTL_PLAIN"`

	Interactive interactiveCmd            `cmd:"" help:"Interactive mode." default:""`
	Dev         devCmd                    `cmd:"" help:"Develop FTL modules. Will start the FTL cluster, build and deploy all modules found in the specified directories, and watch for changes."`
	Serve       serveCmd                  `cmd:"" help:"Start the FTL server."`
	Completion  kongcompletion.Completion `cmd:"" help:"Outputs shell code for initialising tab completions."`
	Logs        logsCmd                   `cmd:"" help:"View logs from FTL modules."`

	// Specify the 1Password vault to access secrets from.
	Vault string `name:"opvault" help:"1Password vault to be used for secrets. The name of the 1Password item will be the <ref> and the secret will be stored in the password field." placeholder:"VAULT"`

	LSP lspCmd `cmd:"" help:"Start the LSP server."`
	MCP mcpCmd `cmd:"" help:"Start the MCP server."`
}

// InteractiveCLI is the CLI that is used when running in interactive mode.
type InteractiveCLI struct {
	SharedCLI
	BuildAndDeploy
}

// DevModeCLI is the embedded CLI when running in dev mode.
type DevModeCLI struct {
	SharedCLI
	Logs struct {
		Level logsSetLevelCmd `cmd:"" help:"Set the current log level"`
	} `cmd:"" help:"Log commands."`
}

var cli CLI

type App struct {
	app *kong.Kong
	csm *currentStatusManager
}

func New(ctx context.Context) (*App, error) {
	csm := &currentStatusManager{}

	app := createKongApplication(&cli, csm)

	err := kong.ApplyDefaults(&cli.LogConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pluginCtx := log.ContextWithLogger(ctx, log.Configure(os.Stdout, cli.LogConfig))
	// TODO: don't do this
	projectConfig, _ := projectconfig.Load(ctx, optional.None[string]()) //nolint:errcheck
	err = languageplugin.PrepareNewCmd(pluginCtx, projectConfig, app, os.Args[1:])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &App{app, csm}, nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	kctx, err := a.app.Parse(os.Args[1:])
	if err != nil {
		return errors.WithStack(err)
	}

	if cli.Trace != "" {
		file, err := os.OpenFile(cli.Trace, os.O_CREATE|os.O_WRONLY, 0644) // #nosec
		if err != nil {
			return errors.Errorf("failed to open trace file: %s", err.Error())
		}
		err = trace.Start(file)
		if err != nil {
			return errors.Errorf("failed to start tracing: %s", err.Error())
		}
		defer trace.Stop()
	}
	ctx, cancel := context.WithCancelCause(ctx)

	if !cli.Plain {
		sm := terminal.NewStatusManager(ctx)
		a.csm.statusManager = optional.Some(sm)
		ctx = sm.IntoContext(ctx)
	}
	rpc.InitialiseClients(cli.Authenticators, cli.Insecure)

	// Set some envars for child processes.
	os.Setenv("LOG_LEVEL", cli.LogConfig.Level.String())

	configPath := cli.ConfigFlag
	if configPath == "" {
		var ok bool
		configPath, ok = projectconfig.DefaultConfigPath().Get()
		if !ok {
			err = errors.Errorf("could not determine default config path, either place an ftl-project.toml file in the root of your project, use --config=FILE, or set the FTL_CONFIG envar")
			cancel(err)
			return err
		}
	}
	if terminal.IsANSITerminal(ctx) {
		cli.LogConfig.Color = true
	}

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	if cli.Insecure {
		logger.Warnf("--insecure skips TLS certificate verification")
	}

	os.Setenv("FTL_CONFIG", configPath)

	bindContext := makeBindContext(logger, cancel, a.csm)
	ctx = bindContext(ctx, kctx)

	err = kctx.Run(ctx)

	if sm, ok := a.csm.statusManager.Get(); ok {
		sm.Close()
	}
	kctx.Exit = os.Exit

	if err != nil {
		kctx.FatalIfErrorf(err)
	}
	return nil
}

func createKongApplication(cli any, csm *currentStatusManager) *kong.Kong {
	gitRoot, _ := internal.GitRoot(".").Get()
	app := kong.Must(cli,
		kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`),
		kong.Configuration(kongtoml.Loader, ".ftl.toml", "~/.ftl.toml"),
		kong.ShortUsageOnError(),
		kong.HelpOptions{Compact: true, WrapUpperBound: 80},
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			node, ok := parent.(*kong.Command)
			if !ok {
				return nil
			}
			return &kong.Group{Key: node.Name, Title: "Command flags:"}
		}),
		kong.Vars{
			"version":          ftl.FormattedVersion,
			"os":               runtime.GOOS,
			"arch":             runtime.GOARCH,
			"numcpu":           strconv.Itoa(runtime.NumCPU()),
			"gitroot":          gitRoot,
			"supportedEditors": strings.Join(editor.SupportedEditors, ","),
		},
		kong.Exit(func(code int) {
			if sm, ok := csm.statusManager.Get(); ok {
				sm.Close()
			}
			signal.Ignore(syscall.SIGINT)                       // We don't want to die with the process group, as that prevents os.Exit from executing.
			_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck // best effort
			os.Exit(code)
		},
		))
	return app
}

func makeBindContext(logger *log.Logger, cancel context.CancelCauseFunc, csm *currentStatusManager) KongContextBinder {
	var bindContext KongContextBinder
	bindContext = func(ctx context.Context, kctx *kong.Context) context.Context {
		err := kctx.BindToProvider(func(cli *SharedCLI) (projectconfig.Config, error) {
			config, err := projectconfig.Load(ctx, optional.Zero(cli.ConfigFlag))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return config, errors.WithStack(err)
			}
			return config, nil
		})
		kctx.FatalIfErrorf(err)
		kctx.Bind(logger)
		kctx.Bind(csm)
		kctx.Bind(&cli.SharedCLI)

		timelineClient := timelineclient.NewClient(ctx, cli.TimelineEndpoint)
		kctx.Bind(timelineClient)

		adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)
		kctx.BindTo(adminClient, (*adminpbconnect.AdminServiceClient)(nil))

		buildEngineClient := rpc.Dial(buildenginepbconnect.NewBuildEngineServiceClient, cli.AdminEndpoint.String(), log.Error)
		kctx.BindTo(buildEngineClient, (*buildenginepbconnect.BuildEngineServiceClient)(nil))

		err = kctx.BindToProvider(func() (*providers.Registry[configuration.Configuration], error) {
			return providers.NewDefaultConfigRegistry(), nil
		})
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(func() (*providers.Registry[configuration.Secrets], error) {
			return providers.NewDefaultSecretsRegistry(), nil
		})
		kctx.FatalIfErrorf(err)

		source := schemaeventsource.New(ctx, "cli", adminClient)
		kctx.BindTo(source, (**schemaeventsource.EventSource)(nil))
		kongcompletion.Register(kctx.Kong, kongcompletion.WithPredictors(terminal.Predictors(source.ViewOnly())))

		verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, cli.AdminEndpoint.String(), log.Error)
		kctx.BindTo(verbServiceClient, (*ftlv1connect.VerbServiceClient)(nil))

		err = kctx.BindToProvider(manager.NewDefaultConfigurationManagerFromConfig)
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(manager.NewDefaultSecretsManagerFromConfig)
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(config.NewConfigurationRegistry)
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(config.NewSecretsRegistry)
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(func(projectConfig projectconfig.Config, secretsRegistry *config.Registry[config.Secrets], configRegistry *config.Registry[config.Configuration]) (*profiles.Project, error) {
			return errors.WithStack2(profiles.Open(filepath.Dir(projectConfig.Path), secretsRegistry, configRegistry))
		})
		kctx.FatalIfErrorf(err)

		err = kctx.BindToProvider(provideAdminClient)
		kctx.FatalIfErrorf(err)

		kctx.BindTo(ctx, (*context.Context)(nil))
		kctx.Bind(bindContext)
		kctx.BindTo(cancel, (*context.CancelCauseFunc)(nil))
		kctx.Bind(cli.AdminEndpoint)
		return ctx
	}
	return bindContext
}

type currentStatusManager struct {
	statusManager optional.Option[terminal.StatusManager]
}

func provideAdminClient(
	ctx context.Context,
	cli *SharedCLI,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projectConfig projectconfig.Config,
	adminClient adminpbconnect.AdminServiceClient,
) (client admin.EnvironmentClient, err error) {
	shouldUseLocalClient, err := admin.ShouldUseLocalClient(ctx, adminClient, cli.AdminEndpoint)
	if err != nil {
		return client, errors.Wrap(err, "could not create admin client")
	}
	if shouldUseLocalClient {
		return admin.NewLocalClient(projectConfig, cm, sm), nil
	}
	return adminClient, nil
}
