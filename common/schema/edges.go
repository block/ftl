package schema

import (
	"slices"
	"strings"
)

// GraphNode provides inbound and outbound edges for a node.
type GraphNode struct {
	Decl
	In  []RefKey
	Out []RefKey
}

// Graph returns a biderectional graph representation of the schema.
func Graph(s *Schema) map[RefKey]GraphNode {
	// Build graph with outbound edges.
	result := map[RefKey]GraphNode{}

	for _, module := range s.Modules {
		Visit(module, func(s Node, next func() error) error {
			d, ok := s.(Decl)
			if !ok {
				return next()
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
				Out:  outboundEdges(d, ignoredRefs),
				In:   []RefKey{},
			}
			return next()
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

// outboundEdges returns all the outbound edges of a node.
func outboundEdges(n Node, ignoredRefs map[RefKey]bool) []RefKey {
	out := []RefKey{}
	if r, ok := n.(*Ref); ok {
		out = append(out, r.ToRefKey())
	}
	Visit(n, func(n Node, next func() error) error {
		r, ok := n.(*Ref)
		if !ok {
			return next()
		}
		if !ignoredRefs[r.ToRefKey()] {
			out = append(out, r.ToRefKey())
		}
		return next()
	})
	return out
}
