package schema

import (
	"fmt"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/slices"
)

// A Data structure.
//
//protobuf:1
//protobuf:13 Type
type Data struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments       []string         `parser:"@Comment*" protobuf:"2"`
	Visibility     Visibility       `parser:"@@?" protobuf:"3"`
	Name           string           `parser:"'data' @Ident" protobuf:"4"`
	TypeParameters []*TypeParameter `parser:"( '<' @@ (',' @@)* '>' )? '{'" protobuf:"5"`
	Metadata       []Metadata       `parser:"@@*" protobuf:"7"`
	Fields         []*Field         `parser:"@@* '}'" protobuf:"6"`
}

var _ Type = (*Data)(nil)
var _ Decl = (*Data)(nil)
var _ Symbol = (*Data)(nil)
var _ Scoped = (*Data)(nil)

func (d *Data) Equal(other Type) bool {
	o, ok := other.(*Data)
	if !ok {
		return false
	}
	if d.Name != o.Name {
		return false
	}
	if len(d.TypeParameters) != len(o.TypeParameters) {
		return false
	}
	if len(d.Metadata) != len(o.Metadata) {
		return false
	}
	if len(d.Fields) != len(o.Fields) {
		return false
	}
	for _, f := range d.Fields {
		fo := o.FieldByName(f.Name)
		if fo == nil || f.Name != fo.Name || !f.Type.Equal(fo.Type) {
			return false
		}
	}
	return true
}
func (d *Data) schemaType() {}
func (d *Data) Scope() Scope {
	scope := Scope{}
	for _, t := range d.TypeParameters {
		scope[t.Name] = ModuleDecl{Symbol: t}
	}
	return scope
}

// FieldByName returns the field with the given name, or nil if it doesn't exist.
func (d *Data) FieldByName(name string) *Field {
	for _, f := range d.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// Monomorphise this data type with the given type arguments.
//
// If this data type has no type parameters, it will be returned as-is.
//
// This will return a new Data structure with all type parameters replaced with
// the given types.
func (d *Data) Monomorphise(ref *Ref) (*Data, error) {
	if len(d.TypeParameters) != len(ref.TypeParameters) {
		return nil, errors.Errorf("%s: expected %d type arguments, got %d", ref.Pos, len(d.TypeParameters), len(ref.TypeParameters))
	}
	if len(d.TypeParameters) == 0 {
		return d, nil
	}
	names := map[string]Type{}
	for i, t := range d.TypeParameters {
		names[t.Name] = ref.TypeParameters[i]
	}
	monomorphised := reflect.DeepCopy(d)
	monomorphised.TypeParameters = nil

	// Because we don't have parent links in the AST allowing us to visit on
	// Type and replace it on the parent, we have to do a full traversal to find
	// the parents of all the Type nodes we need to replace. This will be a bit
	// tricky to maintain, but it's basically any type that has parametric
	// types: maps, slices, fields, etc.
	err := Visit(monomorphised, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *Map:
			k, err := maybeMonomorphiseType(n.Key, names)
			if err != nil {
				return errors.Wrapf(err, "%s: map key", n.Key.Position())
			}
			v, err := maybeMonomorphiseType(n.Value, names)
			if err != nil {
				return errors.Wrapf(err, "%s: map value", n.Value.Position())
			}
			n.Key = k
			n.Value = v

		case *Array:
			t, err := maybeMonomorphiseType(n.Element, names)
			if err != nil {
				return errors.Wrapf(err, "%s: array element", n.Element.Position())
			}
			n.Element = t

		case *Field:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return errors.Wrapf(err, "%s: field type", n.Type.Position())
			}
			n.Type = t

		case *Optional:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return errors.Wrapf(err, "%s: optional type", n.Type.Position())
			}
			n.Type = t

		case *Config:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return errors.Wrapf(err, "%s: config type", n.Type.Position())
			}
			n.Type = t

		case *Secret:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return errors.Wrapf(err, "%s: secret type", n.Type.Position())
			}
			n.Type = t

		case Type, Metadata, Value, IngressPathComponent, DatabaseConnector,
			*Module, *Schema, *TypeParameter, *Verb, *Enum, *TypeAlias, *Topic, *Database,
			*DatabaseRuntimeConnections, *Data, *EnumVariant, *DatabaseRuntime, *Realm:
		}
		return errors.WithStack(next())
	})
	if err != nil {
		return nil, errors.Wrapf(err, "%s: failed to monomorphise", d.Pos)
	}
	return monomorphised, nil
}

func (d *Data) Position() Position { return d.Pos }
func (*Data) schemaDecl()          {}
func (*Data) schemaSymbol()        {}
func (d *Data) schemaChildren() []Node {
	children := make([]Node, 0, len(d.Fields)+len(d.Metadata)+len(d.TypeParameters))
	for _, t := range d.TypeParameters {
		children = append(children, t)
	}
	for _, f := range d.Fields {
		children = append(children, f)
	}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}

func (d *Data) GetName() string           { return d.Name }
func (d *Data) GetVisibility() Visibility { return d.Visibility }

func (d *Data) IsGenerated() bool {
	_, found := slices.FindVariant[*MetadataGenerated](d.Metadata)
	return found
}

func (d *Data) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	typeParameters := ""
	if len(d.TypeParameters) > 0 {
		typeParameters = "<"
		for i, t := range d.TypeParameters {
			if i != 0 {
				typeParameters += ", "
			}
			typeParameters += t.String()
		}
		typeParameters += ">"
	}

	formatTokens(w,
		d.Visibility.String(),
		"data",
		fmt.Sprintf("%s%s", d.Name, typeParameters),
		"{\n",
	)

	for _, metadata := range d.Metadata {
		fmt.Fprintln(w, indent(metadata.String()))
	}
	for _, f := range d.Fields {
		fmt.Fprintln(w, indent(f.String()))
	}
	fmt.Fprintf(w, "}")
	return w.String()
}

func (d *Data) Kind() Kind { return KindData }

// MonoType returns the monomorphised type of this data type if applicable, or returns the original type.
func maybeMonomorphiseType(t Type, typeParameters map[string]Type) (Type, error) {
	if t, ok := t.(*Ref); ok && t.Module == "" {
		if tp, ok := typeParameters[t.Name]; ok {
			return tp, nil
		}
		return nil, errors.Errorf("%s: unknown type parameter %q", t.Position(), t)
	}
	return t, nil
}
