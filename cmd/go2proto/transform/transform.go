package transform

import (
	"fmt"
	"go/token"
	"go/types"

	"github.com/alecthomas/types/optional"
	"golang.org/x/tools/go/packages"
)

var (
	fset            = token.NewFileSet()
	textMarshaler   = loadInterface("encoding", "TextMarshaler")
	binaryMarshaler = loadInterface("encoding", "BinaryMarshaler")
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

// Depth-first search for full match.
func findTransform(from types.Type, to types.Type) []*Transformation {
	for _, probe := range probes {
		if tf := probe(from); tf != nil && types.Identical(tf.To, to) {
			return []*Transformation{tf}
		} else if tf != nil {
			children := findTransform(tf.To, to)
			if len(children) != 0 {
				return append([]*Transformation{tf}, children...)
			}
		}
	}
	return nil
}

type Priority int

const (
	LowPriority Priority = iota - 1
	MediumPriority
	HighPriority
)

// Probe function to determine if the transformation can be applied to the given type.
type Probe func(from types.Type) *Transformation

type Transformation struct {
	To        types.Type
	Priority  Priority
	Transform func(from Expr) Expr
	Imports   []string
	Helper    string
}

var probes = []Probe{
	// T -> P using the method "T.ToProto() (P, error)"
	// func(from types.Type) *Transformation {
	// 	named, ok := from.(*types.Named)
	// 	if !ok {
	// 		return nil
	// 	}
	// 	var toProto *types.Func
	// 	for method := range named.Methods() {
	// 		if method.Name() == "ToProto" && method.Signature().Results().Len() == 1 {
	// 			toProto = method
	// 			break
	// 		}
	// 	}
	// 	if toProto == nil {
	// 		return nil
	// 	}
	// 	result := toProto.Signature().Results().At(0)
	// 	return &Transformation{
	// 		To:       result.Type(),
	// 		Priority: LowPriority,
	// 		Transform: func(from Expr) Expr {
	// 			return Call{
	// 				Name: "toProto",
	// 				Arg:  from,
	// 				Out:  result.Type(),
	// 			}
	// 		},
	// 		Helper: `
	// 			func toProto[P, T interface { ToProto() P }](v T) (P, error) {
	// 				return v.ToProto()
	// 			}
	// 		`,
	// 	}
	// },
	// []byte -> string
	func(from types.Type) *Transformation {
		if !types.Identical(from, types.NewSlice(basicType("byte"))) {
			return nil
		}
		return &Transformation{
			To: basicType("string"),
			Transform: func(from Expr) Expr {
				return Call{
					Name: "bytesToString",
					Arg:  from,
					Out:  types.NewSlice(basicType("byte")),
				}
			},
			Helper: `
				func bytesToString(v []byte) (string, error) {
					return string(v), nil
				}
			`,
		}
	},
	// string -> []byte
	func(from types.Type) *Transformation {
		if !types.Identical(from, basicType("string")) {
			return nil
		}
		return &Transformation{
			To: types.NewSlice(basicType("byte")),
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
		}
	},
	// encoding.BinaryMarshaler -> []byte
	func(from types.Type) *Transformation {
		if !implements(from, binaryMarshaler) {
			return nil
		}
		return &Transformation{
			To: types.NewSlice(basicType("byte")),
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
				func marshalBinary(v encoding.BinaryMarshaler) ([]byte, error) {
					return return v.MarshalBinary()
				}
			`,
		}
	},
	// optional.Option[T] -> T
	func(from types.Type) *Transformation {
		named, ok := from.(*types.Named)
		if !ok {
			return nil
		}
		obj := named.Obj()
		if obj.Pkg().Path() != "github.com/alecthomas/types/optional" && obj.Name() == "Optional" {
			return nil
		}
		return &Transformation{
			To: named.TypeArgs().At(0),
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
		}
	},
}

func loadInterface(pkg, symbol string) *types.Interface {
	name := loadObject(pkg, symbol)
	if t, ok := name.(*types.TypeName); ok {
		if t.Name() == symbol {
			return t.Type().Underlying().(*types.Interface) //nolint:forcetypeassert
		}
	}
	panic("could not find " + pkg + "." + symbol)
}

func loadObject(pkgName, symbol string) types.Object {
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, pkgName)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(symbol)
		if obj != nil {
			return obj
		}
	}
	panic("could not find " + pkgName + "." + symbol)
}

func implements(t types.Type, i *types.Interface) bool {
	return types.Implements(t, i) || types.Implements(types.NewPointer(t), i)
}
