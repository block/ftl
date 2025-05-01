package schema

import (
	"slices"
	"strings"

	errors "github.com/alecthomas/errors"
	"golang.org/x/exp/maps"
)

// GraphNode provides inbound and outbound edges for a node.
type GraphNode struct {
	Decl Decl
	In   []RefKey
	Out  []RefKey
}

// Graph returns a biderectional graph representation of the schema.
func Graph(s *Schema) map[RefKey]GraphNode {
	// Build graph with outbound edges.
	result := map[RefKey]GraphNode{}

	for _, module := range s.InternalModules() {
		Visit(module, func(s Node, next func() error) error { //nolint:errcheck
			d, ok := s.(Decl)
			if !ok {
				return errors.WithStack(next())
			}

			// Ignore type parameters.
			ignoredRefs := map[RefKey]bool{}
			if data, ok := d.(*Data); ok {
				for _, tp := range data.TypeParameters {
					ignoredRefs[RefKey{Name: tp.GetName()}] = true
				}
			}

			result[RefKey{Module: module.Name, Name: d.GetName()}] = GraphNode{
				Decl: d,
				Out:  OutboundEdges(d, ignoredRefs),
				In:   []RefKey{},
			}
			return errors.WithStack(next())
		})
	}
	// Derive inbound edges.
	for ref, node := range result {
		for _, out := range node.Out {
			if target, ok := result[out]; ok {
				target.In = append(target.In, ref)
				result[out] = target
			}
		}
	}
	// Normalise
	for _, node := range result {
		slices.SortFunc(node.In, func(i, j RefKey) int {
			return strings.Compare(i.String(), j.String())
		})
		slices.SortFunc(node.Out, func(i, j RefKey) int {
			return strings.Compare(i.String(), j.String())
		})
	}
	return result
}

// OutboundEdges returns all the outbound edges of a node.
func OutboundEdges(n Node, ignoredRefs map[RefKey]bool) []RefKey {
	out := map[RefKey]bool{}
	if r, ok := n.(*Ref); ok {
		out[r.ToRefKey()] = true
	}
	Visit(n, func(n Node, next func() error) error { //nolint:errcheck
		if r, ok := n.(*Ref); ok && !ignoredRefs[r.ToRefKey()] {
			out[r.ToRefKey()] = true
		}
		return errors.WithStack(next())
	})
	return maps.Keys(out)
}
