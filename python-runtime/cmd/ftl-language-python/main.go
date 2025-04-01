package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/internal/clirpc"
	"github.com/block/ftl/internal/log"
	pythonplugin "github.com/block/ftl/python-runtime/python-plugin"
)

var cli struct {
	Logging log.Config       `embed:"" prefix:"log-"`
	Version kong.VersionFlag `help:"Show version."`
	Name    string           `env:"FTL_NAME" help:"Name of plugin as provided by plugin host."`
	Command string           `arg:"" optional:"" help:"Command to run synchronously. Request is passed as proto-encoded bytes on stdin, and response returned on stdout."`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Kotlin`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)
	if cli.Command == "" {
		plugin.Start(context.Background(),
			cli.Name,
			createService,
			languagepbconnect.LanguageServiceName,
			languagepbconnect.NewLanguageServiceHandler)
	} else {
		ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.Logging))
		path, handler := languagepbconnect.NewLanguageCommandServiceHandler(pythonplugin.CmdService{})
		err := clirpc.Invoke(ctx, handler, path, cli.Command, os.Stdin, os.Stdout)
		kctx.FatalIfErrorf(err)
	}
	kctx.FatalIfErrorf(kctx.Run())
}

func createService(ctx context.Context, config any) (context.Context, *pythonplugin.Service, error) {
	return ctx, pythonplugin.New(), nil
}
