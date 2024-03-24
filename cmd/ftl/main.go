package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type CLI struct {
	Version    kong.VersionFlag `help:"Show version."`
	LogConfig  log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Endpoint   *url.URL         `default:"http://127.0.0.1:8892" help:"FTL endpoint to bind/connect to." env:"FTL_ENDPOINT"`
	ConfigFlag []string         `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`

	Authenticators map[string]string `help:"Authenticators to use for FTL endpoints." mapsep:"," env:"FTL_AUTHENTICATORS" placeholder:"HOST=EXE,…"`

	Status   statusCmd   `cmd:"" help:"Show FTL status."`
	Init     initCmd     `cmd:"" help:"Initialize a new FTL module."`
	Dev      devCmd      `cmd:"" help:"Develop FTL modules. Will start the FTL cluster, build and deploy all modules found in the specified directories, and watch for changes."`
	PS       psCmd       `cmd:"" help:"List deployments."`
	Serve    serveCmd    `cmd:"" help:"Start the FTL server."`
	Call     callCmd     `cmd:"" help:"Call an FTL function."`
	Update   updateCmd   `cmd:"" help:"Update a deployment."`
	Kill     killCmd     `cmd:"" help:"Kill a deployment."`
	Schema   schemaCmd   `cmd:"" help:"FTL schema commands."`
	Build    buildCmd    `cmd:"" help:"Build all modules found in the specified directories."`
	Deploy   deployCmd   `cmd:"" help:"Build and deploy all modules found in the specified directories."`
	Download downloadCmd `cmd:"" help:"Download a deployment."`
	Secret   secretCmd   `cmd:"" help:"Manage secrets."`
	Config   configCmd   `cmd:"" help:"Manage configuration."`
}

var cli CLI

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
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
			"version": ftl.Version,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"numcpu":  strconv.Itoa(runtime.NumCPU()),
		},
	)

	rpc.InitialiseClients(cli.Authenticators)

	// Set some envars for child processes.
	os.Setenv("LOG_LEVEL", cli.LogConfig.Level.String())

	ctx, cancel := context.WithCancel(context.Background())

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	config, _ := projectconfig.LoadConfig(ctx, cli.ConfigFlag)
	kctx.Bind(config)

	sr := cf.ProjectConfigResolver[cf.Secrets]{Config: cli.ConfigFlag}
	cr := cf.ProjectConfigResolver[cf.Configuration]{Config: cli.ConfigFlag}
	kctx.BindTo(sr, (*cf.Resolver[cf.Secrets])(nil))
	kctx.BindTo(cr, (*cf.Resolver[cf.Configuration])(nil))

	// Propagate to runner processes.
	// TODO: This is a bit of a hack until we get proper configuration
	// management through the Controller.
	os.Setenv("FTL_CONFIG", strings.Join(projectconfig.ConfigPaths(cli.ConfigFlag), ","))

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Debugf("FTL terminating with signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	controllerServiceClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, cli.Endpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, controllerServiceClient)
	kctx.BindTo(controllerServiceClient, (*ftlv1connect.ControllerServiceClient)(nil))

	verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, cli.Endpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, verbServiceClient)
	kctx.BindTo(verbServiceClient, (*ftlv1connect.VerbServiceClient)(nil))

	kctx.Bind(cli.Endpoint)
	kctx.BindTo(ctx, (*context.Context)(nil))

	err := kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}
