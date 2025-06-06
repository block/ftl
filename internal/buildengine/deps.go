package buildengine

import (
	"golang.org/x/exp/maps"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	imaps "github.com/block/ftl/internal/maps"
)

type customDependencyProvider func() []string

func GraphFromMetas(metas map[string]moduleMeta, sch *schema.Schema, moduleNames ...string) (map[string][]string, error) {
	return Graph(imaps.MapValues(metas, func(_ string, meta moduleMeta) customDependencyProvider {
		return func() []string { return meta.module.Dependencies(AlwaysIncludeBuiltin) }
	}), sch, moduleNames...)
}

// Graph returns the dependency graph for the given modules.
//
// If no modules are provided, the entire graph is returned. An error is returned if
// any dependencies are missing.
func Graph(customProviders map[string]customDependencyProvider, sch *schema.Schema, moduleNames ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(moduleNames) == 0 {
		moduleNames = maps.Keys(customProviders)
		moduleNames = append(moduleNames, slices.Map(sch.InternalModules(), func(m *schema.Module) string { return m.Name })...)
	}
	for _, name := range moduleNames {
		if err := buildGraph(customProviders, sch, name, out); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return out, nil
}

func buildGraph(customProviders map[string]customDependencyProvider, sch *schema.Schema, moduleName string, out map[string][]string) error {
	var deps []string
	// Short-circuit previously explored nodes
	if _, ok := out[moduleName]; ok {
		return nil
	}
	foundModule := false
	if customProvider, ok := customProviders[moduleName]; ok {
		foundModule = true
		deps = customProvider()
	}
	if !foundModule {
		if sch, ok := slices.Find(sch.InternalModules(), func(m *schema.Module) bool { return m.Name == moduleName }); ok {
			foundModule = true
			deps = append(deps, sch.Imports()...)
		}
	}
	if !foundModule {
		return errors.Errorf("module %q not found. does the module exist and is the ftl.toml file correct?", moduleName)
	}
	deps = slices.Unique(deps)
	out[moduleName] = deps
	for i := range deps {
		dep := deps[i]
		if err := buildGraph(customProviders, sch, dep, out); err != nil {
			return errors.Wrapf(err, "module %q requires dependency %q", moduleName, dep)
		}
	}
	return nil
}
