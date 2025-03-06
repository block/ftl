package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/devstate"
	"github.com/block/ftl/internal/log"
)

type mcpCmd struct{}

func (m mcpCmd) Run(ctx context.Context, k *kong.Kong, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient ftlv1connect.AdminServiceClient) error {
	s := server.NewMCPServer(
		"FTL",
		ftl.Version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	s.AddTool(mcp.NewTool(
		"Status",
		mcp.WithDescription("Get the current status of each FTL module and the current schema"),
	), func(serverCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return statusTool(ctx, buildEngineClient, adminClient)
	})

	// Start the server
	err := server.ServeStdio(s)
	k.FatalIfErrorf(err, "failed to start mcp")
	return nil
}

func contextFromServerContext(ctx context.Context) context.Context {
	return log.ContextWithLogger(ctx, log.Configure(os.Stderr, cli.LogConfig).Scope("mcp"))
}

type statusOutput struct {
	Modules []devstate.ModuleState
	Schema  string
}

func statusTool(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient ftlv1connect.AdminServiceClient) (*mcp.CallToolResult, error) {
	ctx = contextFromServerContext(ctx)
	result, err := devstate.WaitForDevState(ctx, buildEngineClient, adminClient)
	if err != nil {
		return nil, fmt.Errorf("could not get status: %w", err)
	}

	sch := reflect.DeepCopy(result.Schema)
	for _, module := range sch.Modules {
		moduleState, ok := slices.Find(result.Modules, func(m devstate.ModuleState) bool {
			return m.Name == module.Name
		})
		if !ok {
			continue
		}
		for _, decl := range module.Decls {
			switch decl := decl.(type) {
			case *schema.Topic:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Verb:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Config:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Secret:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Database:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Data:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.Enum:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			case *schema.TypeAlias:
				c, err := commentForPath(decl.Pos, moduleState.Path)
				if err != nil {
					return nil, err
				}
				decl.Comments = append(decl.Comments, c)
			}
		}
	}

	output := statusOutput{
		Modules: result.Modules,
		Schema:  sch.String(),
	}
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("could not marshal status: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

func commentForPath(pos schema.Position, modulePath string) (string, error) {
	if pos.Filename == "" {
		return "", nil
	}
	// each position has a prefix of "ftl/modulename". We want to replace that with the module file path
	parts := strings.SplitN(pos.Filename, string(filepath.Separator), 3)
	if len(parts) > 2 {
		parts = parts[1:]
		parts[0] = modulePath
	} else {
		return "", fmt.Errorf("unexpected path format: %s", pos.Filename)
	}
	components := []string{
		filepath.Join(parts...),
	}

	if pos.Line != 0 {
		components = append(components, strconv.Itoa(pos.Line))
		if pos.Column != 0 {
			components = append(components, strconv.Itoa(pos.Column))
		}
	}
	return "Code at " + strings.Join(components, ":"), nil
}
