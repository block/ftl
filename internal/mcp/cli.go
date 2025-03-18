package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/common/reflection"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/projectconfig"
)

const (
	ModuleRegex = "[a-z][a-z0-9_]*"
	RefRegex    = "[a-z][a-z0-9_]*\\.[a-zA-Z_][a-zA-Z0-9_]*"
)

type CLIToolOption func(inputOptions *CLIConfig)

func AddHelp(help string) CLIToolOption {
	return func(inputOptions *CLIConfig) {
		inputOptions.ExtraHelp = append(inputOptions.ExtraHelp, help)
	}
}

func IncludeOptional(name string) CLIToolOption {
	return func(inputOptions *CLIConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &CLIOptionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.IncludeOptional = true
	}
}

func Pattern(name string, pattern optional.Option[string]) CLIToolOption {
	return func(inputOptions *CLIConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &CLIOptionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.Pattern = pattern
	}
}

func Ignore(model any, name string) CLIToolOption {
	return func(inputOptions *CLIConfig) {
		o, ok := inputOptions.InputOptions[name]
		if !ok {
			o = &CLIOptionConfig{}
			inputOptions.InputOptions[name] = o
		}
		o.IgnoreInModel = optional.Some(model)
	}
}

func Args(args ...string) CLIToolOption {
	return func(inputOptions *CLIConfig) {
		inputOptions.ExtraArgs = append(inputOptions.ExtraArgs, args...)
	}
}

func IncludeStatus() CLIToolOption {
	return func(inputOptions *CLIConfig) {
		inputOptions.IncludeStatus = true
	}
}

// AutoReadFilePaths enables filepath detection on the result of the CLI command.
// If a filepath within the project directory is detected, it will be read in and included in the result for the assistant.
// This speeds up the process of performing actions and then waiting for the assistant to read the file.
func AutoReadFilePaths() CLIToolOption {
	return func(inputOptions *CLIConfig) {
		inputOptions.AutoReadFilePaths = true
	}
}

type CLIOptionConfig struct {
	IgnoreInModel   optional.Option[any]
	IncludeOptional bool
	Pattern         optional.Option[string]
}

type CLIConfig struct {
	InputOptions      map[string]*CLIOptionConfig
	ExtraHelp         []string
	ExtraArgs         []string
	IncludeStatus     bool
	AutoReadFilePaths bool
}

func (c CLIConfig) Option(name string) *CLIOptionConfig {
	if inputOptions, ok := c.InputOptions[name]; ok {
		return inputOptions
	}
	return &CLIOptionConfig{}
}

type CLIExecutor func(ctx context.Context, k *kong.Kong, args []string) error

func ToolFromCLI(serverCtx context.Context, k *kong.Kong, projectConfig projectconfig.Config, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient adminpbconnect.AdminServiceClient, executor CLIExecutor, title string, cmdPath []string, toolOptions ...CLIToolOption) (mcp.Tool, server.ToolHandlerFunc) {
	config := &CLIConfig{
		InputOptions: map[string]*CLIOptionConfig{},
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
			panic(fmt.Sprintf("could not find command %s in %v", strings.Join(cmdPath[:i+1], " "), islices.Map(nodes[len(nodes)-1].Children, func(n *kong.Node) string {
				return n.Name
			})))
		}
	}
	helps := append([]string{nodes[len(nodes)-1].Help}, config.ExtraHelp...)
	opts := []mcp.ToolOption{
		mcp.WithDescription(strings.Join(helps, "\n")),
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
			panic(fmt.Sprintf("ftl %v: could not find option %q in:\n%v", strings.Join(cmdPath, " "), name, strings.Join(islices.Map(slices.Collect(maps.Keys(all)), func(name string) string { //nolint: exptostd
				if included[name] {
					return name
				}
				return name + " (skipped)"
			}), "\n")))
		}
	}
	return mcp.NewTool(title, opts...), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := []string{}
		args = append(args, cmdPath...)
		for _, parser := range parsers {
			newArgs, err := parser(request.Params.Arguments)
			if err != nil {
				return nil, err
			}
			args = append(args, newArgs...)
		}
		args = append(args, config.ExtraArgs...)
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
		if err := executor(serverCtx, k, args); err != nil {
			return nil, err
		}
		write.Close()
		var buf strings.Builder
		_, err = io.Copy(&buf, read)
		if err != nil {
			return nil, fmt.Errorf("could not read output: %w", err)
		}
		cliResult := buf.String()
		content := []mcp.Content{
			annotateTextContent(mcp.NewTextContent(cliResult), []mcp.Role{mcp.RoleAssistant}, 1.0),
		}
		if config.AutoReadFilePaths {
			content = append(content, autoReadFilePaths(serverCtx, projectConfig, cliResult)...)
		}
		if config.IncludeStatus {
			if statusContent, err := statusContent(serverCtx, buildEngineClient, adminClient); err == nil {
				content = append(content, statusContent)
			}
		}

		return &mcp.CallToolResult{
			Content: content,
			IsError: false,
		}, nil
	}
}

// inputReader converts mcp request arguments into command line args
type inputReader func(map[string]any) ([]string, error)

func shouldIncludeInput(cmdNodes []*kong.Node, value *kong.Value, config *CLIOptionConfig) bool {
	if (!value.Required || value.HasDefault) && !config.IncludeOptional {
		return false
	}
	if ignoreInModel, ok := config.IgnoreInModel.Get(); ok && cmdNodes[len(cmdNodes)-1].Target.Type() == reflect.TypeOf(ignoreInModel) {
		return false
	}
	return true
}

func optionForInput(value *kong.Value, description string, flag bool, config *CLIOptionConfig) (mcp.ToolOption, inputReader) {
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
			defaultValue, ok := value.DefaultValue.Interface().(bool)
			if !ok {
				panic("expected default value to be a bool")
			}
			opts = append(opts, mcp.DefaultBool(defaultValue))
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
			defaultValue, ok := value.DefaultValue.Interface().(string)
			if !ok {
				panic("expected default value to be a string")
			}
			opts = append(opts, mcp.DefaultString(defaultValue))
		}
		return newStringOption(name, flag, opts)
	case reflect.Struct:
		t := value.Target.Type()
		if t == reflect.TypeOf(reflection.Ref{}) {
			opts = append(opts, mcp.Pattern(RefRegex))
			return newStringOption(name, flag, opts)
		}

	default:
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

// statusContent returns the status of the FTL after the tool was run.
func statusContent(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, adminClient adminpbconnect.AdminServiceClient) (mcp.Content, error) {
	output, err := getStatusOutput(ctx, buildEngineClient, adminClient)
	if err != nil {
		// Fallback to just returning the tool result
		return nil, err
	}
	wrapper := struct {
		Explanation string       `json:"explanation,omitempty"`
		Status      statusOutput `json:"status"`
	}{
		Explanation: "FTL status was retreived after the changes were made. Here is the result.",
		Status:      output,
	}

	statusJSON, err := json.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("could not marshal status: %w", err)
	}

	return annotateTextContent(mcp.NewTextContent(string(statusJSON)), []mcp.Role{mcp.RoleAssistant}, 1.0), nil
}

func autoReadFilePaths(ctx context.Context, projectConfig projectconfig.Config, cliOutput string) []mcp.Content {
	var contents []mcp.Content
	root := projectConfig.Root()
	for _, word := range strings.Fields(cliOutput) {
		if !strings.HasPrefix(word, root) {
			continue
		}
		if _, err := os.Stat(word); err == nil {
			fileContent, err := os.ReadFile(word)
			if err != nil {
				continue
			}
			token, err := tokenForFileContent(fileContent)
			if err != nil {
				continue
			}

			result, err := newReadResult(fileContent, token, false,
				`File at " + word + " may be relevant to the changes that were just made so it has been pre-fetched and included. 
			The WriteVerificationToken can be used to update the file.`)
			if err != nil {
				continue
			}
			contents = append(contents, result.Content...)
		}
	}

	return contents
}
