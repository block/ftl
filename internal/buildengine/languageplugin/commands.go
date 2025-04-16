package languageplugin

import (
	"bytes"
	"context"
	"reflect"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

// GetNewModuleFlags returns the flags that can be used to create a module for this language.
func GetNewModuleFlags(ctx context.Context, language string) ([]*kong.Flag, error) {
	res, err := runCommand[*langpb.GetNewModuleFlagsResponse](ctx, "GetNewModuleFlags", language, &langpb.GetNewModuleFlagsRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get create module flags from plugin")
	}
	return errors.WithStack2(kongFlagsFromProto(res.Flags))
}

func kongFlagsFromProto(protoFlags []*langpb.GetNewModuleFlagsResponse_Flag) ([]*kong.Flag, error) {
	flags := []*kong.Flag{}
	shorts := map[rune]string{}
	for _, f := range protoFlags {
		flag := &kong.Flag{
			Value: &kong.Value{
				Name: f.Name,
				Help: f.Help,
				Tag:  &kong.Tag{},
			},
		}
		if f.Envar != nil && *f.Envar != "" {
			flag.Tag.Envs = []string{*f.Envar}
		}
		if f.Default != nil && *f.Default != "" {
			flag.HasDefault = true
			flag.Default = *f.Default
		}
		if f.Short != nil && *f.Short != "" {
			if len(*f.Short) > 1 {
				return nil, errors.Errorf("invalid flag declared: short flag %q for %v must be a single character", *f.Short, f.Name)
			}
			short := rune((*f.Short)[0])
			if existingFullName, ok := shorts[short]; ok {
				return nil, errors.Errorf("multiple flags declared with the same short name: %v and %v", existingFullName, f.Name)
			}
			flag.Short = short
			shorts[short] = f.Name

		}
		if f.Placeholder != nil && *f.Placeholder != "" {
			flag.PlaceHolder = *f.Placeholder
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

// NewModule creates a new module in the given directory with the given name and language.
func NewModule(ctx context.Context, language string, projConfig projectconfig.Config, moduleConfig moduleconfig.ModuleConfig, flags map[string]string) error {
	genericFlags := map[string]any{}
	for k, v := range flags {
		genericFlags[k] = v
	}
	flagsProto, err := structpb.NewStruct(genericFlags)
	if err != nil {
		return errors.Wrap(err, "failed to convert flags to proto")
	}
	_, err = runCommand[*langpb.NewModuleResponse](ctx, "NewModule", language, &langpb.NewModuleRequest{
		Name:          moduleConfig.Module,
		Dir:           moduleConfig.Dir,
		ProjectConfig: langpb.ProjectConfigToProto(projConfig),
		Flags:         flagsProto,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create module")
	}
	return nil
}

// GetModuleConfigDefaults provides custom defaults for the module config.
//
// The result may be cached by FTL, so defaulting logic should not be changing due to normal module changes.
// For example, it is valid to return defaults based on which build tool is configured within the module directory,
// as that is not expected to change during normal operation.
// It is not recommended to read the module's toml file to determine defaults, as when the toml file is updated,
// the module defaults will not be recalculated.
func GetModuleConfigDefaults(ctx context.Context, language string, dir string) (moduleconfig.CustomDefaults, error) {
	result, err := runCommand[*langpb.GetModuleConfigDefaultsResponse](ctx, "GetModuleConfigDefaults", language, &langpb.GetModuleConfigDefaultsRequest{
		Dir: dir,
	})
	if err != nil {
		return moduleconfig.CustomDefaults{}, errors.Wrap(err, "failed to get module config defaults from plugin")
	}
	return customDefaultsFromProto(result), nil
}

func customDefaultsFromProto(proto *langpb.GetModuleConfigDefaultsResponse) moduleconfig.CustomDefaults {
	return moduleconfig.CustomDefaults{
		DeployDir:      proto.DeployDir,
		Watch:          proto.Watch,
		Build:          optional.Ptr(proto.Build),
		DevModeBuild:   optional.Ptr(proto.DevModeBuild),
		BuildLock:      optional.Ptr(proto.BuildLock),
		LanguageConfig: proto.LanguageConfig.AsMap(),
		SQLRootDir:     proto.SqlRootDir,
	}
}

func runCommand[Resp proto.Message](ctx context.Context, name string, language string, req proto.Message) (out Resp, err error) {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return out, errors.Wrap(err, "failed to marshal command")
	}
	cmdPath, err := cmdPathForLanguage(language)
	if err != nil {
		return out, errors.WithStack(err)
	}
	cliCmd := exec.Command(ctx, log.Debug, ".", cmdPath, name)
	cliCmd.Stdin = bytes.NewReader(reqBytes)
	outBytes, err := cliCmd.Capture(ctx)
	if err != nil {
		return out, errors.Wrap(err, "failed to run command")
	}
	out, ok := reflect.New(reflect.TypeOf(out).Elem()).Interface().(Resp)
	if !ok {
		return out, errors.Errorf("failed to create response type: %T", out)
	}
	err = proto.Unmarshal(outBytes, out)
	if err != nil {
		return out, errors.Wrap(err, "failed to unmarshal result")
	}
	return out, nil
}

// PrepareNewCmd adds language specific flags to kong
// This allows the new command to have good support for language specific flags like:
// - help text (ftl module new go --help)
// - default values
// - environment variable overrides
//
// Language plugins take time to launch, so we return the one we created so it can be reused in Run().
func PrepareNewCmd(ctx context.Context, projectConfig projectconfig.Config, k *kong.Kong, args []string) error {
	if len(args) < 2 {
		return nil
	} else if args[0] != "new" {
		return nil
	}

	language := args[1]
	// Default to `new` command handler if no language is provided, or option is specified on `new` command.
	if len(language) == 0 || language[0] == '-' {
		return nil
	}

	newCmdNode, ok := slices.Find(k.Model.Children, func(n *kong.Node) bool {
		return n.Name == "new"
	})
	if !ok {
		return errors.Errorf("could not find new command")
	}

	flags, err := GetNewModuleFlags(ctx, language)
	if err != nil {
		return errors.Wrapf(err, "could not get CLI flags for %v plugin", language)
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
	return nil
}
