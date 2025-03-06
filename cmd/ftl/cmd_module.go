package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kong"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

type moduleCmd struct {
	New moduleNewCmd `cmd:"" help:"Create a new FTL module."`
}

type moduleNewCmd struct {
	Language string `arg:"" help:"Language of the module to create."`
	Name     string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
	Dir      string `arg:"" help:"Directory to initialize the module in." default:"."`

	AllowedDirs []string `help:"Directory that modules are required to be in (unless --force is set)." env:"FTL_DEV_DIRS" hidden:""`
	Force       bool     `help:"Force creation of module without checking allowed directories." short:"f"`
}

func (i moduleNewCmd) Run(ctx context.Context, ktctx *kong.Context, config projectconfig.Config, pluginHolder languageplugin.InitializedPlugins) error {
	logger := log.FromContext(ctx)
	name, path, err := validateModule(i.Dir, i.Name)
	if err != nil {
		return err
	}

	if !i.Force && len(i.AllowedDirs) > 0 {
		allowedAbsDirs := make([]string, len(i.AllowedDirs))
		for i, d := range i.AllowedDirs {
			absDir, err := filepath.Abs(d)
			if err != nil {
				return fmt.Errorf("could not make %q an absolute path: %w", d, err)
			}
			allowedAbsDirs[i] = absDir
		}
		if _, ok := slices.Find(allowedAbsDirs, func(d string) bool {
			return strings.HasPrefix(path, d)
		}); !ok {
			return fmt.Errorf("module directory %s is not within the expected module directories (%v). Please choose an appropriate path or force create the module outside of the expected directories by using the --force argument", path, strings.Join(allowedAbsDirs, ", "))
		}
	}

	logger.Debugf("Creating FTL %s module %q in %s", i.Language, name, path)

	moduleConfig := moduleconfig.ModuleConfig{
		Module:   name,
		Language: i.Language,
		Dir:      path,
	}

	flags := map[string]string{}
	for _, f := range ktctx.Selected().Flags {
		if f.Name == "allowed-dirs" || f.Name == "force" {
			continue
		}
		flagValue, ok := f.Target.Interface().(string)
		if !ok {
			return fmt.Errorf("expected %v value to be a string but it was %T", f.Name, f.Target.Interface())
		}
		flags[f.Name] = flagValue
	}

	plugin, err := pluginHolder.Plugin(ctx, config, i.Language)
	if err != nil {
		return err
	}

	release, err := flock.Acquire(ctx, config.WatchModulesLockPath(), 30*time.Second)
	if err != nil {
		return fmt.Errorf("could not acquire file lock: %w", err)
	}
	defer release() //nolint:errcheck

	err = plugin.CreateModule(ctx, config, moduleConfig, flags)
	if err != nil {
		return err
	}

	_, ok := internal.GitRoot(i.Dir).Get()
	if !config.NoGit && ok {
		logger.Debugf("Adding files to git")
		if err := maybeGitAdd(ctx, i.Dir, filepath.Join(i.Name, "*")); err != nil {
			return err
		}
	}
	_ = plugin.Kill() //nolint:errcheck

	fmt.Printf("Successfully created %s module %q in %s\n", i.Language, name, path)
	return nil
}

func validateModule(dir string, name string) (string, string, error) {
	if strings.Contains(name, string(filepath.Separator)) && (dir == "." || dir == "") {
		components := strings.Split(name, string(filepath.Separator))
		suggestedDir := strings.Join(components[:len(components)-1], string(filepath.Separator))
		return "", "", fmt.Errorf("module name %q cannot contain path separators. Did you mean 'ftl module new %s %s'", name, components[len(components)-1], suggestedDir)
	}
	if dir == "" {
		return "", "", fmt.Errorf("directory is required")
	}
	if name == "" {
		name = filepath.Base(dir)
	}
	if !schema.ValidateModuleName(name) {
		return "", "", fmt.Errorf("module name %q is invalid", name)
	}
	path := filepath.Join(dir, name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("could not make %q an absolute path: %w", path, err)
	}
	if _, err := os.Stat(absPath); err == nil {
		return "", "", fmt.Errorf("module directory %s already exists", path)
	}
	return name, absPath, nil
}
