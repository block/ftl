package module

import (
	"context"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

// Module represents an FTL Module in the build engine
type Module struct {
	Config moduleconfig.ModuleConfig

	// SQLErrors are errors that occurred during SQL build which overrule any errors from the language plugin
	SQLError error

	// do not read directly, use Dependecies() instead
	dependencies []string
}

func newModule(config moduleconfig.ModuleConfig) Module {
	return Module{
		Config:       config,
		dependencies: []string{},
	}
}

func (m Module) CopyWithDependencies(dependencies []string) Module {
	module := reflect.DeepCopy(m)
	module.dependencies = dependencies
	return module
}

func (m Module) CopyWithSQLErrors(err error) Module {
	module := reflect.DeepCopy(m)
	module.SQLError = err
	return module
}

// DependencyMode is an enum for dependency modes
type DependencyMode string

const (
	Raw                  DependencyMode = "Raw"
	AlwaysIncludeBuiltin DependencyMode = "AlwaysIncludingBuiltin"
)

// Dependencies returns the dependencies of the Module
// Mode allows us to control how dependencies are returned.
//
// When calling language plugins, use Raw mode to ensure plugins receive the same
// dependencies that were declared.
func (m Module) Dependencies(mode DependencyMode) []string {
	dependencies := m.dependencies
	switch mode {
	case Raw:
		// leave as is
	case AlwaysIncludeBuiltin:
		containsBuiltin := false
		for _, dep := range dependencies {
			if dep == "builtin" {
				containsBuiltin = true
				break
			}
		}
		if !containsBuiltin {
			dependencies = append(dependencies, "builtin")
		}
	}
	return dependencies
}

// EngineModule is a wrapper around a Module that includes the last build's start time.
type EngineModule struct {
	Module         Module
	plugin         *languageplugin.LanguagePlugin
	events         chan languageplugin.PluginEvent
	configDefaults moduleconfig.CustomDefaults
	watch          *watch.Watcher

	devModeEndpointUpdates chan dev.LocalEndpoint
	devMode                bool

	os       string
	arch     string
	buildEnv []string
}

// Build a Module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (m *EngineModule) Build(ctx context.Context, projectConfig projectconfig.Config, sch *schema.Schema) (result *schema.Module, tmpDeployDir string, deployPaths []string, err error) {

	if m.Module.SQLError != nil {
		m.Module = m.Module.CopyWithSQLErrors(nil)
	}
	moduleSchema, tmpDeployDir, deployPaths, err := build(ctx, projectConfig, m.Module, m.plugin, languageplugin.BuildContext{
		Config:       m.Module.Config,
		Schema:       sch,
		Dependencies: m.Module.Dependencies(Raw),
		BuildEnv:     m.buildEnv,
		Os:           m.os,
		Arch:         m.arch,
	}, m.devMode, m.devModeEndpointUpdates)

	if err != nil {
		if errors.Is(err, errSQLError) {
			// Keep sql error around so that subsequent auto rebuilds from the plugin keep the sql error
			m.Module = m.Module.CopyWithSQLErrors(err)
		}
		return nil, "", nil, errors.WithStack(err)
	}

	return moduleSchema, tmpDeployDir, deployPaths, nil
}

func (m *EngineModule) SyncStubReferences(ctx context.Context, projectRoot string, moduleNames []string, view *schema.Schema) error {
	stubsRoot := stubsLanguageDir(projectRoot, m.Module.Config.Language)
	if err := m.plugin.SyncStubReferences(ctx, m.Module.Config, stubsRoot, moduleNames, view); err != nil {
		return errors.Wrapf(err, "failed to sync go stub references for %s", m.Module.Config.Module)
	}
	return nil
}

// GenerateStubs Syncs stub references for the given Module (not necessarily this Module).
// This is basically using this modules plugin to generate stubs for the other Module. This is not confusing at all.
func (m *EngineModule) GenerateStubs(ctx context.Context, projectRoot string, otherModule *EngineModule, otherModuleSchema *schema.Module) error {
	path := stubsModuleDir(projectRoot, m.Language(), m.Name())
	var nativeConfig optional.Option[moduleconfig.ModuleConfig]
	if otherModule.Module.Config.Module == "builtin" || otherModule.Module.Config.Language != m.Module.Config.Language {
		nativeConfig = optional.Some(m.Module.Config)
	}
	if err := m.plugin.GenerateStubs(ctx, path, otherModuleSchema, otherModule.Module.Config, nativeConfig); err != nil {
		return errors.WithStack(err) //nolint:wrapcheck
	}
	return nil
}

func (m *EngineModule) Language() string {
	return m.Module.Config.Language
}
func (m *EngineModule) Name() string {
	return m.Module.Config.Module
}

func (m *EngineModule) Dependencies(mode DependencyMode) []string {
	m.Module.Dependencies(mode)
}
