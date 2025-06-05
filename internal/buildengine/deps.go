package buildengine

import (
	"golang.org/x/exp/maps"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

// Graph returns the dependency graph for the given modules.
//
// If no modules are provided, the entire graph is returned. An error is returned if
// any dependencies are missing.
func Graph(metas map[string]moduleMeta, sch *schema.Schema, moduleNames ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(moduleNames) == 0 {
		moduleNames = maps.Keys(metas)
	}
	for _, name := range moduleNames {
		if err := buildGraph(metas, sch, name, out); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return out, nil
}

func buildGraph(metas map[string]moduleMeta, sch *schema.Schema, moduleName string, out map[string][]string) error {
	var deps []string
	// Short-circuit previously explored nodes
	if _, ok := out[moduleName]; ok {
		return nil
	}
	foundModule := false
	if meta, ok := metas[moduleName]; ok {
		foundModule = true
		deps = meta.module.Dependencies(AlwaysIncludeBuiltin)
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
		if err := buildGraph(metas, sch, dep, out); err != nil {
			return errors.Wrapf(err, "module %q requires dependency %q", moduleName, dep)
		}
	}
	return nil
}
