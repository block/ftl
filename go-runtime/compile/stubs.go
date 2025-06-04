package compile

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/block/scaffolder"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/block/ftl"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/watch"
)

type ExternalDeploymentContext struct {
	GoVersion    string
	FTLVersion   string
	Module       *schema.Module
	Replacements []*modfile.Replace
}

func GenerateStubs(ctx context.Context, dir string, moduleSch *schema.Module, config moduleconfig.AbsModuleConfig, nativeConfig optional.Option[moduleconfig.AbsModuleConfig]) error {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dir)
	}

	var goModVersion string
	var replacements []*modfile.Replace
	var err error

	// If there's no module config, use the go.mod file for the first config we find.
	if config.Module == "builtin" || config.Language != "go" {
		nativeConfig, ok := nativeConfig.Get()
		if !ok {
			return errors.Errorf("no native module config provided")
		}
		goModPath := filepath.Join(nativeConfig.Dir, "go.mod")
		_, goModVersion, err = updateGoModule(goModPath, nativeConfig.Module, optional.None[watch.ModifyFilesTransaction]())
		if err != nil {
			return errors.Errorf("could not read go.mod %s", goModPath)
		}
		if goModVersion == "" {
			// The best we can do here if we don't have a module to read from is to use the current Go version.
			goModVersion = runtime.Version()[2:]
		}
		replacements = []*modfile.Replace{}
		if !ftl.IsRelease(ftl.Version) {
			path, err := os.Executable()
			if err != nil {
				return errors.Wrap(err, "failed to get executable pat")
			}
			ftlpath := "block/ftl"
			idx := strings.LastIndex(path, ftlpath)
			if idx > 0 {
				path = path[:idx+len(ftlpath)]
				replacements = append(replacements, &modfile.Replace{Old: module.Version{Path: "github.com/block/ftl"}, New: module.Version{Path: path}})
			}
		}
	} else {
		replacements, goModVersion, err = updateGoModule(filepath.Join(config.Dir, "go.mod"), config.Module, optional.None[watch.ModifyFilesTransaction]())
		if err != nil {
			return errors.WithStack(err)
		}
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return errors.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	context := ExternalDeploymentContext{
		GoVersion:    goModVersion,
		FTLVersion:   ftlVersion,
		Module:       moduleSch,
		Replacements: replacements,
	}

	funcs := maps.Clone(scaffoldFuncs)
	err = internal.ScaffoldZip(externalModuleTemplateFiles(), dir, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
	if err != nil {
		return errors.Wrap(err, "failed to scaffold zip")
	}

	if err := exec.Command(ctx, log.Debug, dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return errors.Wrap(err, "failed to tidy go.mod")
	}
	return nil
}

func SyncGeneratedStubReferences(ctx context.Context, config moduleconfig.AbsModuleConfig, stubsDir string, stubbedModules []string) error {
	sharedModulePaths := []string{}
	relativeStubsDir, err := filepath.Rel(config.Dir, stubsDir)
	if err != nil {
		return errors.Wrap(err, "failed to get relative path")
	}
	for _, mod := range stubbedModules {
		if mod == config.Module {
			continue
		}
		sharedModulePaths = append(sharedModulePaths, filepath.Join(relativeStubsDir, mod))
	}

	_, goModVersion, err := updateGoModule(filepath.Join(config.Dir, "go.mod"), config.Module, optional.None[watch.ModifyFilesTransaction]())
	if err != nil {
		return errors.WithStack(err)
	}

	funcs := maps.Clone(scaffoldFuncs)
	if err := internal.ScaffoldZip(mainWorkTemplateFiles(), config.Dir, MainWorkContext{
		GoVersion:          goModVersion,
		SharedModulesPaths: sharedModulePaths,
		IncludeMainPackage: mainPackageExists(config),
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return errors.Wrap(err, "failed to scaffold zip")
	}
	return nil
}

func mainPackageExists(config moduleconfig.AbsModuleConfig) bool {
	// check if main package exists, otherwise do not include it
	_, err := os.Stat(filepath.Join(buildDir(config.Dir), "go", "main", "go.mod"))
	return !errors.Is(err, os.ErrNotExist)
}
