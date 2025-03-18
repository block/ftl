package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinepbconnect"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/mcp"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
)

type mcpCmd struct{}

func (m mcpCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, bindContext KongContextBinder) error {
	timelineClient := rpc.Dial(timelinepbconnect.NewTimelineServiceClient, cli.TimelineEndpoint.String(), log.Error)

	s := newMCPServer(ctx, k, projectConfig, buildEngineClient, adminClient, timelineClient, bindContext)
	k.FatalIfErrorf(s.Serve(), "failed to serve MCP")
	return nil
}

func newMCPServer(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, timelineClient timelinepbconnect.TimelineServiceClient, bindContext KongContextBinder) *mcp.Server {
	s := mcp.New()

	executor := func(ctx context.Context, k *kong.Kong, args []string) error {
		return runInnerCmd(ctx, k, projectConfig, bindContext, args, nil)
	}

	s.AddTool(mcp.StatusTool(ctx, buildEngineClient, adminClient))
	s.AddTool(mcp.TimelineTool(ctx, timelineClient))
	s.AddTool(mcp.ReadTool())
	s.AddTool(mcp.WriteTool(ctx, buildEngineClient, adminClient))

	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "NewModule", []string{"module", "new"},
		mcp.IncludeOptional("dir"), mcp.Pattern("name", optional.Some(mcp.ModuleRegex)),
		mcp.IncludeStatus()))
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "CallVerb", []string{"call"},
		mcp.IncludeOptional("request"), mcp.Args("-v")))
	// TODO: all secret commands, with xor group of bools for providers
	// TODO: all config commands, with xor group of bools for providers
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "ResetSubscription", []string{"pubsub", "subscription", "reset"},
		mcp.AddHelp("This does not return any info about the state of the subscription."),
		mcp.AddHelp("The user MUST explicitly ask for the subscription to be reset.")))
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "NewMySQLDatabase", []string{"mysql", "new"},
		mcp.Pattern("datasource", optional.Some(mcp.RefRegex)),
		mcp.IncludeStatus(),
		mcp.AutoReadFilePaths()))
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "NewMySQLMigration", []string{"mysql", "new", "migration"},
		mcp.Ignore(newSQLCmd{}, "datasource"),
		mcp.Pattern("datasource", optional.Some(mcp.RefRegex)),
		mcp.AutoReadFilePaths()))
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "NewPostgresDatabase", []string{"postgres", "new"},
		mcp.Pattern("datasource", optional.Some(mcp.RefRegex)),
		mcp.IncludeStatus(),
		mcp.AutoReadFilePaths()))
	s.AddTool(mcp.ToolFromCLI(ctx, k, projectConfig, buildEngineClient, adminClient, executor, "NewPostgresMigration", []string{"postgres", "new", "migration"},
		mcp.Ignore(newSQLCmd{}, "datasource"),
		mcp.Pattern("datasource", optional.Some(mcp.RefRegex)),
		mcp.AutoReadFilePaths()))
	return s
}
