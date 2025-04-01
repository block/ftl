package main

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/block/ftl"
	languagepb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/internal/clirpc"
	pythonplugin "github.com/block/ftl/python-runtime/python-plugin"
)

var cli struct {
	Version kong.VersionFlag `help:"Show version."`
	Serve   serve            `cmd:"" default:"1" help:"Run the service."`

	GetNewModuleFlags       GetNewModuleFlagsCmd       `cmd:"" help:"Get language specific flags for a new module."`
	NewModule               NewModuleCmd               `cmd:"" help:"Create a new module."`
	GetModuleConfigDefaults GetModuleConfigDefaultsCmd `cmd:"" help:"Get the default config for a module."`
}

type serve struct {
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Python`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)
	kctx.FatalIfErrorf(kctx.Run())
}

func (s serve) Run() error {
	plugin.Start(context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler)
	return nil
}

func createService(ctx context.Context, config any) (context.Context, *pythonplugin.Service, error) {
	return ctx, pythonplugin.New(), nil
}

type GetNewModuleFlagsCmd struct {
	clirpc.Command
}

func (cmd GetNewModuleFlagsCmd) Run() error {
	cmdSvc := pythonplugin.CmdService{}
	return clirpc.Run(cmd.Command, func(ctx context.Context, req *languagepb.GetNewModuleFlagsRequest) (*languagepb.GetNewModuleFlagsResponse, error) {
		resp, err := cmdSvc.GetNewModuleFlags(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		return resp.Msg, nil
	})
}

type NewModuleCmd struct {
	clirpc.Command
}

func (cmd NewModuleCmd) Run() error {
	cmdSvc := pythonplugin.CmdService{}
	return clirpc.Run(cmd.Command, func(ctx context.Context, req *languagepb.NewModuleRequest) (*languagepb.NewModuleResponse, error) {
		resp, err := cmdSvc.NewModule(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		return resp.Msg, nil
	})
}

type GetModuleConfigDefaultsCmd struct {
	clirpc.Command
}

func (cmd GetModuleConfigDefaultsCmd) Run() error {
	cmdSvc := pythonplugin.CmdService{}
	return clirpc.Run(cmd.Command, func(ctx context.Context, req *languagepb.GetModuleConfigDefaultsRequest) (*languagepb.GetModuleConfigDefaultsResponse, error) {
		resp, err := cmdSvc.GetModuleConfigDefaults(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		return resp.Msg, nil
	})
}
