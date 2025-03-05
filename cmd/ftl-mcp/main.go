package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/log"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Version         kong.VersionFlag `help:"Show version."`
	LogConfig       log.Config       `embed:"" prefix:"log-"`
	AdminEndpoint   *url.URL         `help:"Admin endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	UpdatesEndpoint *url.URL         `help:"Socket to bind to." default:"http://127.0.0.1:8900" env:"FTL_BUILD_UPDATES_ENDPOINT"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Model Context Protocol Server`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	buildEngineClient := rpc.Dial(buildenginepbconnect.NewBuildEngineServiceClient, cli.UpdatesEndpoint.String(), log.Error)
	adminClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)

	s := server.NewMCPServer(
		"FTL",
		ftl.Version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	s.AddTool(mcp.NewTool(
		"Status",
		mcp.WithDescription("Get the current status of each FTL module and the current schema"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return statusTool(ctx, buildEngineClient, adminClient)
	})

	// Start the server
	err := server.ServeStdio(s)
	kctx.FatalIfErrorf(err, "failed to start mcp")
}

func contextFromServerContext(ctx context.Context) context.Context {
	return log.ContextWithLogger(ctx, log.Configure(os.Stderr, cli.LogConfig).Scope("mcp"))
}
