package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/internal/mcp"
	"github.com/block/ftl/internal/projectconfig"
)

type mcpCmd struct{}

func (m mcpCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, bindContext KongContextBinder) error {

	s := newMCPServer(ctx, k, projectConfig, buildEngineClient, adminClient, bindContext)
	k.FatalIfErrorf(s.Serve(), "failed to serve MCP")
	return nil
}

func newMCPServer(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, bindContext KongContextBinder) *mcp.Server {
	s := mcp.New()

	executor := func(ctx context.Context, k *kong.Kong, args []string) error {
		return runInnerCmd(ctx, k, projectConfig, bindContext, args, nil)
	}

	s.AddTool(mcp.StatusTool(ctx, buildEngineClient, adminClient))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "NewModule", []string{"module", "new"}, mcp.IncludeOptional("dir"), mcp.Pattern("name", optional.Some(mcp.ModuleRegex))))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "CallVerb", []string{"call"}, mcp.IncludeOptional("request")))
	// TODO: all secret commands, with xor group of bools for providers
	// TODO: all config commands, with xor group of bools for providers
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "ResetSubscription", []string{"pubsub", "subscription", "reset"}))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "NewMySQLDatabase", []string{"mysql", "new"}, mcp.Pattern("datasource", optional.Some(mcp.RefRegex))))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "NewMySQLMigration", []string{"mysql", "new", "migration"}, mcp.Ignore(newSQLCmd{}, "datasource")))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "NewPostgresDatabase", []string{"postgres", "new"}))
	s.AddTool(mcp.ToolFromCLI(ctx, k, executor, "NewPostgresMigration", []string{"postgres", "new", "migration"}, mcp.Ignore(newSQLCmd{}, "datasource")))
	return s
}
