package buildengine

import (
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/internal/moduleconfig"
)

// Module represents an FTL module in the build engine
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

// Dependencies returns the dependencies of the module
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
