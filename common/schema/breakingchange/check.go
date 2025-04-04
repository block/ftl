package breakingchange

import (
	"fmt"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/maps"
)

// Error returned by Validate
type Error struct {
	Pos schema.Position

	Err error
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Err) }
func (e Error) Unwrap() error { return e.Err }

// Validate that a type has not changed in a backwards incompatible way.
//
// For data structures, we support the following operations:
//
//   - Add new fields.
//   - Remove optional fields.
func Validate(s *schema.Schema, prev, next schema.Type) error {
	return validate(s, map[schema.Kind]bool{}, prev, next)
}

func validate(s *schema.Schema, seen map[schema.Kind]bool, prev, next schema.Type) error {
	if prev.Kind() != next.Kind() {
		return errorf(next, "schema changed type from %s to %s", prev.Kind(), next.Kind())
	}
	switch prev := prev.(type) {
	case *schema.Ref:
		// This is a lot of stuff!
		prevDecl, ok := s.Resolve(prev).Get()
		if !ok {
			return errorf(prev, "failed to resolve reference %s", prev)
		}
		nextDecl, ok := s.Resolve(next.(*schema.Ref)).Get() //nolint:forcetypeassert
		if !ok {
			return errorf(next, "failed to resolve reference %s", next)
		}
		nextType, ok := nextDecl.(schema.Type)
		if !ok {
			return errorf(next, "resolved reference was %T, not a type", next)
		}
		prevType, ok := prevDecl.(schema.Type)
		if !ok {
			return errorf(next, "resolved reference was %T, not a type", prev)
		}
		return validate(s, seen, prevType, nextType)

	case *schema.Array:
		next := next.(*schema.Array) //nolint:forcetypeassert
		return validate(s, seen, prev.Element, next.Element)

	case *schema.Map:
		next := next.(*schema.Map) //nolint:forcetypeassert
		if err := validate(s, seen, prev.Key, next.Key); err != nil {
			return err
		}
		if err := validate(s, seen, prev.Value, next.Value); err != nil {
			return err
		}
		return nil

	case *schema.Optional:
		next := next.(*schema.Optional) //nolint:forcetypeassert
		return validate(s, seen, prev.Type, next.Type)

	case *schema.Data:
		// Don't check generated data structures.
		if _, ok := slices.FindVariant[*schema.MetadataGenerated](prev.Metadata); ok {
			return nil
		}
		next := next.(*schema.Data) //nolint:forcetypeassert
		return validateData(s, seen, prev, next)

	case *schema.Enum:
		next := next.(*schema.Enum) //nolint:forcetypeassert
		// TODO: do we allow new enum values to be added? I think that breaks backwards compatibility.
		if !prev.Equal(next) {
			return errorf(prev, "enum values changed")
		}
		return nil
	}
	return nil
}

func validateData(s *schema.Schema, seen map[schema.Kind]bool, prev, next *schema.Data) error {
	prevf := maps.FromSlice(prev.Fields, func(f *schema.Field) (string, *schema.Field) { return f.Name, f })
	nextf := maps.FromSlice(next.Fields, func(f *schema.Field) (string, *schema.Field) { return f.Name, f })
	for name, prevField := range prevf {
		nextField, ok := nextf[name]
		if !ok {
			return errorf(prevField, "field %q removed", name)
		}
		if err := validate(s, seen, prevField.Type, nextField.Type); err != nil {
			return err
		}
	}
	return nil
}

func errorf(pos interface{ Position() schema.Position }, format string, args ...any) error {
	return Error{Pos: pos.Position(), Err: fmt.Errorf(format, args...)}
}
