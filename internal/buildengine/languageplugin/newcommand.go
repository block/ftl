package languageplugin

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/projectconfig"
)

// PrepareNewCmd adds language specific flags to kong
// This allows the new command to have good support for language specific flags like:
// - help text (ftl module new go --help)
// - default values
// - environment variable overrides
//
// Language plugins take time to launch, so we return the one we created so it can be reused in Run().
func PrepareNewCmd(ctx context.Context, projectConfig projectconfig.Config, k *kong.Kong, args []string) (optionalPlugin InitializedPlugins, err error) {
	if len(args) < 2 {
		return optionalPlugin, nil
	} else if args[0] != "new" {
		return optionalPlugin, nil
	}

	language := args[1]
	// Default to `new` command handler if no language is provided, or option is specified on `new` command.
	if len(language) == 0 || language[0] == '-' {
		return optionalPlugin, nil
	}

	newCmdNode, ok := slices.Find(k.Model.Children, func(n *kong.Node) bool {
		return n.Name == "new"
	})
	if !ok {
		return optionalPlugin, fmt.Errorf("could not find new command")
	}

	plugin, err := CreateLanguagePlugin(ctx, projectConfig, language)
	if err != nil {
		return optionalPlugin, fmt.Errorf("could not create plugin for %v: %w", language, err)
	}
	flags, err := plugin.GetCreateModuleFlags(ctx)
	if err != nil {
		return optionalPlugin, fmt.Errorf("could not get CLI flags for %v plugin: %w", language, err)
	}

	registry := kong.NewRegistry().RegisterDefaults()
	for _, flag := range flags {
		var str string
		strPtr := &str
		flag.Target = reflect.ValueOf(strPtr).Elem()
		flag.Mapper = registry.ForValue(flag.Target)
		flag.Group = &kong.Group{
			Title: "Flags for " + strings.ToTitle(language[0:1]) + language[1:] + " modules",
			Key:   "languageSpecificFlags",
		}
	}
	newCmdNode.Flags = append(newCmdNode.Flags, flags...)
	return InitializedPlugins{plugin: map[string]*LanguagePlugin{language: plugin}}, nil
}

func CreateLanguagePlugin(ctx context.Context, projectConfig projectconfig.Config, language string) (plugin *LanguagePlugin, err error) {
	plugin, err = New(ctx, projectConfig.Root(), language, "new")

	if err != nil {
		return plugin, fmt.Errorf("could not create plugin for %v: %w", language, err)
	}

	return plugin, nil
}

type InitializedPlugins struct {
	plugin map[string]*LanguagePlugin
}

func (r *InitializedPlugins) Plugin(ctx context.Context, projectConfig projectconfig.Config, language string) (*LanguagePlugin, error) {
	if r.plugin == nil {
		r.plugin = map[string]*LanguagePlugin{}
	}
	pl := r.plugin[language]
	if pl != nil {
		return pl, nil
	}
	p, err := CreateLanguagePlugin(ctx, projectConfig, language)
	if err != nil {
		return nil, err
	}
	r.plugin[language] = p
	return p, nil
}

func (r *InitializedPlugins) Close() {
	if r.plugin == nil {
		return
	}
	for _, p := range r.plugin {
		_ = p.Kill() //nolint:errcheck
	}
	r.plugin = nil
}
