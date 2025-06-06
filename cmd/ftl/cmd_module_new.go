package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/profiles"
)

type moduleNewCmd struct {
	Language string `arg:"" help:"Language of the module to create."`
	Name     string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
	Dir      string `arg:"" help:"Directory to initialize the module in." default:""`

	AllowedDirs []string `help:"Directory that modules are required to be in (unless --force is set)." env:"FTL_DEV_DIRS" hidden:""`
	Force       bool     `help:"Force creation of module without checking allowed directories." short:"f"`
}

func (i moduleNewCmd) Run(ctx context.Context, ktctx *kong.Context, config profiles.ProjectConfig) error {
	logger := log.FromContext(ctx)

	dir := i.moduleDir(config)

	name, path, err := validateModule(dir, i.Name)
	if err != nil {
		return errors.WithStack(err)
	}

	if !i.Force && len(i.AllowedDirs) > 0 {
		allowedAbsDirs := make([]string, len(i.AllowedDirs))
		for i, d := range i.AllowedDirs {
			absDir, err := filepath.Abs(d)
			if err != nil {
				return errors.Wrapf(err, "could not make %q an absolute path", d)
			}
			allowedAbsDirs[i] = absDir
		}
		if _, ok := slices.Find(allowedAbsDirs, func(d string) bool {
			return strings.HasPrefix(path, d)
		}); !ok {
			return errors.Errorf("module directory %s is not within the expected module directories (%v). Please choose an appropriate path or force create the module outside of the expected directories by using the --force argument", path, strings.Join(allowedAbsDirs, ", "))
		}
	}

	logger.Debugf("Creating FTL %s module %q in %s", i.Language, name, path)

	moduleConfig := moduleconfig.ModuleConfig{
		Realm:    config.Realm,
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
			return errors.Errorf("expected %v value to be a string but it was %T", f.Name, f.Target.Interface())
		}
		flags[f.Name] = flagValue
	}

	lockPath := config.WatchModulesLockPath()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0700); err != nil {
		return errors.Wrap(err, "could not create directory for file lock")
	}
	release, err := flock.Acquire(ctx, lockPath, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "could not acquire file lock")
	}
	defer release() //nolint:errcheck

	err = languageplugin.NewModule(ctx, i.Language, config, moduleConfig, flags)
	if err != nil {
		return errors.WithStack(err)
	}

	_, ok := internal.GitRoot(dir).Get()
	if config.Git && ok {
		logger.Debugf("Adding files to git")
		if err := maybeGitAdd(ctx, dir, filepath.Join(i.Name, "*")); err != nil {
			return errors.WithStack(err)
		}
	}

	fmt.Printf("Successfully created %s module %q in %s\n", i.Language, name, path)
	return nil
}

// Use the first module directory in the project config if available, otherwise use the current directory
func (i moduleNewCmd) moduleDir(config profiles.ProjectConfig) string {
	dir := i.Dir
	if dir == "" {
		if len(config.ModuleRoots) > 0 {
			dir = config.ModuleRoots[0]
		} else {
			dir = "."
		}
	}
	return dir
}

func validateModule(dir string, name string) (string, string, error) {
	if strings.Contains(name, string(filepath.Separator)) && (dir == "." || dir == "") {
		components := strings.Split(name, string(filepath.Separator))
		suggestedDir := strings.Join(components[:len(components)-1], string(filepath.Separator))
		return "", "", errors.Errorf("module name %q cannot contain path separators. Did you mean 'ftl module new %s %s'", name, components[len(components)-1], suggestedDir)
	}
	if dir == "" {
		return "", "", errors.Errorf("directory is required")
	}
	if name == "" {
		name = filepath.Base(dir)
	}
	if !schema.ValidateModuleName(name) {
		return "", "", errors.Errorf("module name %q is invalid", name)
	}
	path := filepath.Join(dir, name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", errors.Wrapf(err, "could not make %q an absolute path", path)
	}
	if _, err := os.Stat(absPath); err == nil {
		return "", "", errors.Errorf("module directory %s already exists", path)
	}
	return name, absPath, nil
}
