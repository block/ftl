package mcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"

	"maps"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/block/ftl/common/reflection"
	islices "github.com/block/ftl/common/slices"
)

const (
	ModuleRegex = "[a-z][a-z0-9_]*"
	RefRegex    = "[a-z][a-z0-9_]*\\.[a-zA-Z_][a-zA-Z0-9_]*"
)

type CLIToolOption func(inputOptions *CLIConfig)

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

type CLIOptionConfig struct {
	IgnoreInModel   optional.Option[any]
	IncludeOptional bool
	Pattern         optional.Option[string]
}

type CLIConfig struct {
	InputOptions map[string]*CLIOptionConfig
}

func (c CLIConfig) Option(name string) *CLIOptionConfig {
	if inputOptions, ok := c.InputOptions[name]; ok {
		return inputOptions
	}
	return &CLIOptionConfig{}
}

type CLIExecutor func(ctx context.Context, k *kong.Kong, args []string) error

func ToolFromCLI(ctx context.Context, k *kong.Kong, executor CLIExecutor, title string, cmdPath []string, toolOptions ...CLIToolOption) (mcp.Tool, server.ToolHandlerFunc) {
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
			panic(fmt.Sprintf("ftl %v: could not find option %q in:\n%v", strings.Join(cmdPath, " "), name, strings.Join(islices.Map(slices.Collect(maps.Keys(all)), func(name string) string { //nolint: exptostd
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
		if err := executor(ctx, k, args); err != nil {
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
