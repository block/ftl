package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/exp/maps"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	ireflect "github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/devstate"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
)

type mcpCmd struct{}

const (
	moduleRegex = "[a-z][a-z0-9_]*"
	refRegex    = "[a-z][a-z0-9_]*\\.[a-zA-Z_][a-zA-Z0-9_]*"
)

func (m mcpCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, bindContext KongContextBinder) error {

	// TODO: resource for `schema ebnf` / `schema example`

	s := newMCPServer(ctx, k, projectConfig, buildEngineClient, adminClient, bindContext)
	err := server.ServeStdio(s)
	k.FatalIfErrorf(err, "failed to start mcp")
	return nil
}

// newMCPServer creates a new mcp server with all the tools and resources
// Panics if a tool cannot be added
func newMCPServer(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	adminClient adminpbconnect.AdminServiceClient, bindContext KongContextBinder) *server.MCPServer {
	logger := log.FromContext(ctx)
	s := server.NewMCPServer(
		"FTL",
		ftl.Version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	addTool := func(tool mcp.Tool, handler server.ToolHandlerFunc) {
		json, err := tool.MarshalJSON()
		if err != nil {
			logger.Errorf(err, "failed to marshal tool %s", tool.Name)
		} else {
			logger.Debugf("Adding tool %s", string(json))
		}
		s.AddTool(tool, handler)
	}

	addTool(mcp.NewTool(
		"Status",
		mcp.WithDescription("Get the current status of each FTL module and the current schema"),
	), func(serverCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return statusTool(ctx, buildEngineClient, adminClient)
	})
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "NewModule", []string{"module", "new"}, IncludeOptional("dir"), Pattern("name", optional.Some(moduleRegex))))
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "CallVerb", []string{"call"}, IncludeOptional("request")))
	// all secret commands, with xor group of bools for providers
	// all config commands, with xor group of bools for providers
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "ResetSubscription", []string{"pubsub", "subscription", "reset"}))
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "NewMySQLDatabase", []string{"mysql", "new"}, Pattern("datasource", optional.Some(refRegex))))
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "NewMySQLMigration", []string{"mysql", "new", "migration"}, Ignore(newSQLCmd{}, "datasource")))
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "NewPostgresDatabase", []string{"postgres", "new"}))
	addTool(toolFromCLI(ctx, k, projectConfig, bindContext, "NewPostgresMigration", []string{"postgres", "new", "migration"}, Ignore(newSQLCmd{}, "datasource")))
	return s
}

func contextFromServerContext(ctx context.Context) context.Context {
	return log.ContextWithLogger(ctx, log.Configure(os.Stderr, cli.LogConfig).Scope("mcp"))
}

type statusOutput struct {
	Modules []devstate.ModuleState
	Schema  string
}

func statusTool(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient adminpbconnect.AdminServiceClient) (*mcp.CallToolResult, error) {
	ctx = contextFromServerContext(ctx)
	result, err := devstate.WaitForDevState(ctx, buildEngineClient, adminClient)
	if err != nil {
		return nil, fmt.Errorf("could not get status: %w", err)
	}

	sch := ireflect.DeepCopy(result.Schema)
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

type cliToolOption func(inputOptions *cliConfig)

func IncludeOptional(name string) cliToolOption {
	return func(inputOptions *cliConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &optionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.IncludeOptional = true
	}
}

func Pattern(name string, pattern optional.Option[string]) cliToolOption {
	return func(inputOptions *cliConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &optionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.Pattern = pattern
	}
}

func Ignore(model any, name string) cliToolOption {
	return func(inputOptions *cliConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &optionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.IgnoreInModel = optional.Some(model)
	}
}

type optionConfig struct {
	IgnoreInModel   optional.Option[any]
	IncludeOptional bool
	Pattern         optional.Option[string]
}

type cliConfig struct {
	InputOptions map[string]*optionConfig
}

func (c cliConfig) Option(name string) *optionConfig {
	if inputOptions, ok := c.InputOptions[name]; ok {
		return inputOptions
	}
	return &optionConfig{}
}

func toolFromCLI(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, bindContext KongContextBinder, title string, cmdPath []string, toolOptions ...cliToolOption) (mcp.Tool, server.ToolHandlerFunc) {
	config := &cliConfig{
		InputOptions: map[string]*optionConfig{},
	}
	for _, opt := range toolOptions {
		opt(config)
	}

	nodes := make([]*kong.Node, 0, len(cmdPath)+1)
	nodes = append(nodes, k.Model.Node)
	for i, name := range cmdPath {
		var found bool
		for _, node := range nodes[len(nodes)-1].Children {
			if node.Type != kong.CommandNode {
				continue
			}
			if node.Name == name {
				nodes = append(nodes, node)
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("could not find command %s in %v", strings.Join(cmdPath[:i+1], " "), slices.Map(nodes[len(nodes)-1].Children, func(n *kong.Node) string {
				return n.Name
			})))
		}
	}
	opts := []mcp.ToolOption{
		mcp.WithDescription(nodes[len(nodes)-1].Help),
	}
	parsers := []inputReader{}
	included := map[string]bool{}
	all := map[string]bool{}
	for i, n := range nodes {
		for _, child := range n.Children {
			if child.Type != kong.ArgumentNode {
				continue
			}
			all[child.Name] = true
			c := config.Option(child.Name)
			cmdNodes := append([]*kong.Node{}, nodes[:i+1]...)
			cmdNodes = append(cmdNodes, child)
			if !shouldIncludeInput(cmdNodes, child.Argument, c) {
				continue
			}
			opt, parser := optionForInput(child.Argument, child.Help, false, c)
			opts = append(opts, opt)
			parsers = append(parsers, parser)
			included[child.Name] = true
		}
		for _, child := range n.Positional {
			all[child.Name] = true
			c := config.Option(child.Name)
			if !shouldIncludeInput(nodes[:i+1], child, c) {
				continue
			}
			opt, parser := optionForInput(child, child.Help, false, c)
			opts = append(opts, opt)
			parsers = append(parsers, parser)
			included[child.Name] = true
		}
		for _, flag := range n.Flags {
			all[flag.Name] = true
			c := config.Option(flag.Name)
			if !shouldIncludeInput(nodes[:i+1], flag.Value, c) {
				continue
			}
			opt, parser := optionForInput(flag.Value, flag.Help, true, c)
			opts = append(opts, opt)
			parsers = append(parsers, parser)
			included[flag.Name] = true
		}
	}
	// validate that all configured options were found
	for name := range config.InputOptions {
		if !included[name] {
			panic(fmt.Sprintf("ftl %v: could not find option %q in:\n%v", strings.Join(cmdPath, " "), name, strings.Join(slices.Map(maps.Keys(all), func(name string) string {
				if included[name] {
					return name
				}
				return name + " (skipped)"
			}), "\n")))
		}
	}
	return mcp.NewTool(title, opts...), func(serverCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := []string{}
		args = append(args, cmdPath...)
		for _, parser := range parsers {
			newArgs, err := parser(request.Params.Arguments)
			if err != nil {
				return nil, err
			}
			args = append(args, newArgs...)
		}
		oldOut := os.Stdout
		oldErr := os.Stderr
		defer func() {
			os.Stdout = oldOut
			os.Stderr = oldErr
		}()
		read, write, err := os.Pipe()

		if err != nil {
			return nil, fmt.Errorf("could not create pipe: %w", err)
		}
		os.Stdout = write
		os.Stderr = write
		if err := runInnerCmd(ctx, k, projectConfig, bindContext, args, nil); err != nil {
			return nil, err
		}
		write.Close()
		var buf strings.Builder
		_, err = io.Copy(&buf, read)
		if err != nil {
			return nil, fmt.Errorf("could not read output: %w", err)
		}
		return mcp.NewToolResultText(buf.String()), nil
	}
}

// inputReader converts mcp request arguments into command line args
type inputReader func(map[string]any) ([]string, error)

func shouldIncludeInput(cmdNodes []*kong.Node, value *kong.Value, config *optionConfig) bool {
	if (!value.Required || value.HasDefault) && !config.IncludeOptional {
		return false
	}
	if ignoreInModel, ok := config.IgnoreInModel.Get(); ok && cmdNodes[len(cmdNodes)-1].Target.Type() == reflect.TypeOf(ignoreInModel) {
		return false
	}
	return true
}

func optionForInput(value *kong.Value, description string, flag bool, config *optionConfig) (mcp.ToolOption, inputReader) {
	name := value.Name
	if value.Name == "" {
		panic("name is required")
	}
	opts := []mcp.PropertyOption{}
	if description != "" {
		opts = append(opts, mcp.Description(description))
	}
	if value.Required {
		opts = append(opts, mcp.Required())
	}
	if pattern, ok := config.Pattern.Get(); ok {
		opts = append(opts, mcp.Pattern(pattern))
	}
	switch value.Target.Kind() {
	case reflect.Bool:
		if !flag {
			panic("unhandled bool argument")
		}
		if value.HasDefault {
			opts = append(opts, mcp.DefaultBool(value.DefaultValue.Interface().(bool)))
		}
		return mcp.WithBoolean(name, opts...), func(args map[string]any) ([]string, error) {
			anyValue, ok := args[name]
			if ok {
				value, ok := anyValue.(bool)
				if !ok {
					return nil, fmt.Errorf("expected %s to be a bool but it was %T", name, anyValue)
				}
				if value {
					return []string{"--" + name}, nil
				}
			}
			return nil, nil
		}

	case reflect.String:
		if value.HasDefault {
			opts = append(opts, mcp.DefaultString(value.DefaultValue.Interface().(string)))
		}
		return newStringOption(name, flag, opts)
	case reflect.Struct:
		t := value.Target.Type()
		if t == reflect.TypeOf(reflection.Ref{}) {
			opts = append(opts, mcp.Pattern(refRegex))
			return newStringOption(name, flag, opts)
		}
	}

	panic(fmt.Sprintf("implement type %v %v for %s (hasDefault = %v)", value.Target.Type(), value.Target.Kind(), name, value.HasDefault))
}

func newStringOption(name string, flag bool, opts []mcp.PropertyOption) (mcp.ToolOption, inputReader) {
	return mcp.WithString(name, opts...), func(args map[string]any) ([]string, error) {
		value, ok := args[name]
		if !ok {
			return nil, nil
		}
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected %s to be a string but it was %T", name, value)
		}
		if flag {
			return []string{"--" + name, str}, nil
		}
		return []string{str}, nil
	}
}
