package transform

import (
	"reflect"
	"slices"
)

// Type is similar to reflect.Type but handles both go/types.Type and "synthetic" types that are created for the
// purposes of matching transformations.
type Type interface {
	Kind() reflect.Kind
	// Only relevant for map types.
	Key() Type
	// For types this is the element type.
	Elem() Type
	// Size of array.
	Size() int
	Methods() []Function
	Fields() []Field
}

type Function struct {
	Name string
	In   []Type
	Out  []Type
}

type Field struct {
	Name string
	Type Type
}

// Identical returns true if a and b are identical types.
func Identical(a, b Type) bool {
	if a.Kind() != b.Kind() {
		return false
	}
	switch a.Kind() {
	case reflect.Array:
		return a.Size() == b.Size() && Identical(a.Elem(), b.Elem())

	case reflect.Slice, reflect.Chan, reflect.Pointer:
		return Identical(a.Elem(), b.Elem())

	case reflect.Struct:
		return slices.EqualFunc(a.Fields(), b.Fields(), func(af, bf Field) bool {
			return af.Name == bf.Name && Identical(af.Type, bf.Type)
		})

	case reflect.Interface:
		return slices.EqualFunc(a.Methods(), b.Methods(), func(am, bm Function) bool {
			return am.Name == bm.Name && slices.EqualFunc(am.In, bm.In, Identical) && slices.EqualFunc(am.Out, bm.Out, Identical)
		})

	case reflect.Map:
		return Identical(a.Key(), b.Key()) && Identical(a.Elem(), b.Elem())

	default:
		return true
	}
}
