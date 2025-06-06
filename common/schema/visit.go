package schema

import (
	"slices"

	"github.com/alecthomas/errors"
)

// Visit all nodes in the schema.
func Visit(n Node, visit func(n Node, next func() error) error) error {
	return errors.WithStack(visit(n, func() error {
		if n == nil {
			return nil
		}
		for _, child := range n.schemaChildren() {
			if err := Visit(child, visit); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}))
}

// VisitWithParents visits all nodes in the schema providing the parent nodes on each visit
func VisitWithParents(n Node, parents []Node, visit func(n Node, parents []Node, next func() error) error) error {
	return errors.WithStack(visit(n, parents, func() error {
		for _, child := range n.schemaChildren() {
			childParents := slices.Clone(parents)
			childParents = append(childParents, n)
			if err := VisitWithParents(child, childParents, visit); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}))
}

// VisitExcludingMetadataChildren visits all nodes in the schema except the children of metadata nodes.
// This is used when generating external modules to avoid adding imports only referenced in the bodies of
// stubbed verbs.
func VisitExcludingMetadataChildren(n Node, visit func(n Node, next func() error) error) error {
	return errors.WithStack(visit(n, func() error {
		if d, ok := n.(Decl); ok {
			if !d.GetVisibility().Exported() {
				// Skip non-exported nodes
				return nil
			}
		}
		if _, ok := n.(Metadata); !ok {
			for _, child := range n.schemaChildren() {
				_, isParentVerb := n.(*Verb)
				_, isChildUnit := child.(*Unit)
				if isParentVerb && isChildUnit {
					// Skip visiting children of a verb that are units as the scaffolded code will not inclue them
					continue
				}
				if err := VisitExcludingMetadataChildren(child, visit); err != nil {
					return errors.WithStack(err)
				}
			}
		}
		return nil
	}))
}
