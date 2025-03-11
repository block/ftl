package mcp

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/block/ftl"
)

type Server struct {
	mcpServer *server.MCPServer
}

type CommandExecutor func(ctx context.Context, k *kong.Kong, args []string, additionalExit func(int)) error

// New creates a new mcp server with all the tools and resources
func New() *Server {
	mcpServer := server.NewMCPServer(
		"FTL",
		ftl.Version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	s := &Server{mcpServer: mcpServer}
	return s
}

// Adds a tool to the server
// Panics if the tool is not able to marshal to json
func (s *Server) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	_, err := tool.MarshalJSON()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal tool %s: %s", tool.Name, err))
	}
	s.mcpServer.AddTool(tool, handler)
}

func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
