package main

import (
	"fmt"
	"go/types"

	"github.com/alecthomas/types/optional"
)

// An Expr represents an expression of the given type.
type Expr interface {
	// Type of the Expr wrapped in result.Result[T]
	Type() types.Type
	Expr() string
}

type Variable struct {
	Name string
	Typ  types.Type
}

var _ Expr = Variable{}

func Var(name string, typ types.Type) Variable { return Variable{name, typ} }
func (v Variable) Type() types.Type            { return v.Typ }
func (v Variable) Expr() string                { return fmt.Sprintf("result.Ok(%s)", v.Name) }

type Call struct {
	Package optional.Option[string]
	Name    string
	Arg     Expr
	Out     types.Type
}

var _ Expr = Call{}

func (c Call) Type() types.Type { return c.Out }
func (c Call) Expr() string {
	name := ""
	if pkg, ok := c.Package.Get(); ok {
		name += pkg + "."
	}
	name += c.Name
	return fmt.Sprintf("result.Map(%s, %s)", c.Arg.Expr(), name)
}

// basicType looks up a basic type by name.
func basicType(name string) types.Type {
	return types.Universe.Lookup(name).(*types.TypeName).Type()
}

func Transform(from Expr, to types.Type) (Expr, bool) {
	transforms := findTransform(from.Type(), to)
	if len(transforms) == 0 {
		return nil, false
	}
	out := from
	for _, transformation := range transforms {
		out = transformation.Transform(out)
	}
	return out, true
}

func findTransform(from types.Type, to types.Type) []Transformation {
	for _, transformation := range transformations {
		if tto, ok := transformation.Test(from); ok && types.Identical(tto, to) {
			return []Transformation{transformation}
		} else if ok {
			children := findTransform(tto, to)
			if len(children) != 0 {
				return append([]Transformation{transformation}, children...)
			}
		}
	}
	return nil
}

type Transformation struct {
	// Test function to determine if the transformation can be applied to the given type, and if so what it's resulting
	// type will be.
	Test      func(from types.Type) (to types.Type, ok bool)
	Transform func(from Expr) Expr
	Imports   []string
	Helper    string
}

var transformations = []Transformation{
	// string -> []byte
	{
		Test: func(from types.Type) (to types.Type, ok bool) {
			if from != basicType("string") {
				return nil, false
			}
			return types.NewSlice(basicType("byte")), true
		},
		Transform: func(from Expr) Expr {
			return Call{
				Name: "stringToBytes",
				Arg:  from,
				Out:  basicType("string"),
			}
		},
		Helper: `
			func stringToBytes(v string) ([]byte, error) {
				return []byte(v), nil
			}
		`,
	},
	// encoding.BinaryMarshaler -> []byte
	{
		Test: func(from types.Type) (to types.Type, ok bool) {
			if !implements(from, binaryMarshaler) {
				return nil, false
			}
			return types.NewSlice(basicType("byte")), true
		},
		Transform: func(from Expr) Expr {
			return Call{
				Name: "marshalBinary",
				Arg:  from,
				Out:  types.NewSlice(basicType("byte")),
			}
		},
		Imports: []string{
			"encoding",
			"github.com/alecthomas/types/result",
		},
		Helper: `
			func marshalBinary[T encoding.BinaryMarshaler](v T) ([]byte, error) {
				return return v.MarshalBinary()
			}
		`,
	},
	// optional.Option[T] -> T
	{
		Test: func(from types.Type) (to types.Type, ok bool) {
			named, ok := from.(*types.Named)
			if !ok {
				return nil, false
			}
			obj := named.Obj()
			if obj.Pkg().Path() != "github.com/alecthomas/types/optional" && obj.Name() == "Optional" {
				return nil, false
			}
			return named.TypeArgs().At(0), true
		},
		Transform: func(from Expr) Expr {
			return Call{
				Name: "unwrapOptional",
				Arg:  from,
				Out:  from.Type().(*types.Named).TypeParams().At(0),
			}
		},
		Helper: `
			func unwrapOptional[T any](v optional.Option[T]) (T, error) {
				out, _ := v.Get()
				return out, nil
			}
		`,
	},
}
