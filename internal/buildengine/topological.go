package buildengine

import (
	"fmt"
	"sort"

	"github.com/alecthomas/errors"
	"golang.org/x/exp/maps"
)

type DependencyCycleError struct {
	Modules []string
}

var _ error = DependencyCycleError{}

func (e DependencyCycleError) Error() string {
	return fmt.Sprintf("detected a module dependency cycle that impacts these modules: %q", e.Modules)
}

// TopologicalSort attempts to order the modules supplied in the graph based on
// their topologically sorted order. A cycle in the module dependency graph
// will cause this sort to be incomplete. The sorted modules are returned as a
// sequence of `groups` of modules that may be built in parallel. The `unsorted`
// modules impacted by a dependency cycle get reported as an error.
func TopologicalSort(graph map[string][]string) (groups [][]string, cycleError error) {
	modulesByName := map[string]bool{}
	for module := range graph {
		modulesByName[module] = true
	}

	// Modules that have already been "built"
	built := map[string]bool{"builtin": true}

	for len(modulesByName) > 0 {
		// Current group of modules that can be built in parallel.
		group := map[string]bool{}
	nextModule:
		for module := range modulesByName {
			// Check that all dependencies have been built.
			for _, dep := range graph[module] {
				if !built[dep] {
					continue nextModule
				}
			}
			group[module] = true
			delete(modulesByName, module)
		}
		// A module dependency cycle prevents further sorting
		if len(group) == 0 {
			// The remaining modules are either a member of the cyclical
			// dependency chain or depend (directly or transitively) on
			// a member of the cyclical dependency chain
			modules := maps.Keys(modulesByName)
			sort.Strings(modules)
			cycleError = DependencyCycleError{Modules: modules}
			break
		}
		orderedGroup := maps.Keys(group)
		sort.Strings(orderedGroup)
		for _, module := range orderedGroup {
			built[module] = true
		}
		groups = append(groups, orderedGroup)
	}
	return groups, errors.WithStack(cycleError)
}
