package transform

import (
	"fmt"
	"go/types"
	"reflect"
)

// FromGoTypes creates an adapter from [go/types.Type] to [Type].
func FromGoTypes(t types.Type) Type {
	return &goType{t}
}

var _ Type = (*goType)(nil)

type goType struct {
	t types.Type
}

func (g *goType) Elem() Type {
	switch t := g.t.(type) {
	case *types.Pointer:
		return &goType{t.Elem()}
	case *types.Array:
		return &goType{t.Elem()}
	case *types.Slice:
		return &goType{t.Elem()}
	case *types.Map:
		return &goType{t.Elem()}
	case *types.Chan:
		return &goType{t.Elem()}
	default:
		panic(fmt.Sprintf("type %T does not have an element type", g.t))
	}
}

func (g *goType) Fields() []Field {
	switch t := g.t.(type) {
	case *types.Struct:
		fields := make([]Field, t.NumFields())
		for i := range t.NumFields() {
			field := t.Field(i)
			fields[i] = Field{
				Name: field.Name(),
				Type: &goType{field.Type()},
			}
		}
		return fields
	case *types.Named:
		if s, ok := t.Underlying().(*types.Struct); ok {
			fields := make([]Field, s.NumFields())
			for i := range s.NumFields() {
				field := s.Field(i)
				fields[i] = Field{
					Name: field.Name(),
					Type: &goType{field.Type()},
				}
			}
			return fields
		}
		return nil
	default:
		panic("type does not have fields")
	}
}

func (g *goType) Key() Type {
	switch t := g.t.(type) {
	case *types.Map:
		return &goType{t.Key()}
	default:
		panic("type does not have a key type")
	}
}

func (g *goType) Kind() reflect.Kind {
	return typesKind(g.t)
}

func (g *goType) Methods() []Function {
	return append(goTypeMethod(g.t), goTypeMethod(types.NewPointer(g.t))...)
}

func (g *goType) String() string { return g.Kind().String() }

func (g *goType) Size() int {
	a, ok := g.t.(*types.Array)
	if !ok {
		panic("not an array")
	}
	return int(a.Len())
}

func typesKind(t types.Type) reflect.Kind {
	switch t := t.(type) {
	case *types.Struct:
		return reflect.Struct

	case *types.Interface:
		return reflect.Interface

	case *types.Array:
		return reflect.Array

	case *types.Slice:
		return reflect.Slice

	case *types.Chan:
		return reflect.Chan

	case *types.Pointer:
		return reflect.Pointer

	case *types.Map:
		return reflect.Map

	case *types.Signature:
		return reflect.Func

	case *types.Basic:
		return typesToReflectKind[t.Kind()]

	case *types.Named:
		return typesKind(t.Underlying())

	default:
		panic("unimplemented")
	}
}

func goTypeMethod(t types.Type) []Function {
	methods := types.NewMethodSet(types.NewPointer(t))
	out := make([]Function, methods.Len())
	for i := range methods.Len() {
		method := methods.At(i)
		funcObj := method.Obj().(*types.Func)
		sig := funcObj.Type().(*types.Signature)

		// Convert parameters
		params := sig.Params()
		inParams := make([]Type, params.Len())
		for j := range params.Len() {
			inParams[j] = &goType{params.At(j).Type()}
		}

		// Convert results
		results := sig.Results()
		outParams := make([]Type, results.Len())
		for j := range results.Len() {
			outParams[j] = &goType{results.At(j).Type()}
		}

		out[i] = Function{
			Name: funcObj.Name(),
			In:   inParams,
			Out:  outParams,
		}
	}
	return out
}

var typesToReflectKind = []reflect.Kind{
	types.Bool:          reflect.Bool,
	types.Int:           reflect.Int,
	types.Int8:          reflect.Int8,
	types.Int16:         reflect.Int16,
	types.Int32:         reflect.Int32,
	types.Int64:         reflect.Int64,
	types.Uint:          reflect.Uint,
	types.Uint8:         reflect.Uint8,
	types.Uint16:        reflect.Uint16,
	types.Uint32:        reflect.Uint32,
	types.Uint64:        reflect.Uint64,
	types.Float32:       reflect.Float32,
	types.Float64:       reflect.Float64,
	types.Complex64:     reflect.Complex64,
	types.Complex128:    reflect.Complex128,
	types.String:        reflect.String,
	types.UnsafePointer: reflect.UnsafePointer,
}
