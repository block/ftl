package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/maps"
)

//protobuf:4
//protobuf:14 Type
type Enum struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments   []string       `parser:"@Comment*" protobuf:"2"`
	Visibility Visibility     `parser:"@@?" protobuf:"3"`
	Name       string         `parser:"'enum' @Ident" protobuf:"4"`
	Type       Type           `parser:"(':' @@)?" protobuf:"5,optional"`
	Variants   []*EnumVariant `parser:"'{' @@* '}'" protobuf:"6"`
}

var _ Type = (*Enum)(nil)
var _ Decl = (*Enum)(nil)
var _ Symbol = (*Enum)(nil)

func (e *Enum) Position() Position { return e.Pos }

func (e *Enum) Equal(other Type) bool {
	o, ok := other.(*Enum)
	if !ok {
		return false
	}
	if (e.Type == nil) != (o.Type == nil) {
		return false
	}
	if e.Type != nil && !e.Type.Equal(o.Type) {
		return false
	}
	if len(e.Variants) != len(o.Variants) {
		return false
	}
	ourVariants := maps.FromSlice(e.Variants, func(v *EnumVariant) (string, *EnumVariant) { return v.Name, v })
	otherVariants := maps.FromSlice(o.Variants, func(v *EnumVariant) (string, *EnumVariant) { return v.Name, v })
	for name, variant := range ourVariants {
		if otherVariant, ok := otherVariants[name]; !ok || !variant.Value.Equal(otherVariant.Value) {
			return false
		}
	}
	return true
}

func (e *Enum) schemaType() {}

func (e *Enum) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(e.Comments))
	formatTokens(w,
		e.Visibility.String(),
		"enum",
		e.Name,
	)
	if e.Type != nil {
		fmt.Fprintf(w, ": %s", e.Type)
	}
	fmt.Fprint(w, " {\n")
	for _, v := range e.Variants {
		fmt.Fprintln(w, indent(v.String()))
	}
	fmt.Fprint(w, "}")
	return w.String()
}
func (*Enum) schemaDecl()   {}
func (*Enum) schemaSymbol() {}
func (*Enum) Kind() Kind    { return KindEnum }
func (e *Enum) schemaChildren() []Node {
	var children []Node
	for _, v := range e.Variants {
		children = append(children, v)
	}
	if e.Type != nil {
		children = append(children, e.Type)
	}
	return children
}
func (e *Enum) GetName() string           { return e.Name }
func (e *Enum) GetVisibility() Visibility { return e.Visibility }
func (e *Enum) IsGenerated() bool         { return false }

// IsValueEnum determines whether this is a type or value enum using `e.Type` alone
// because value enums must always have a unified type across all variants, whereas type
// enums by definition cannot have a unified type.
func (e *Enum) IsValueEnum() bool {
	return e.Type != nil
}

func (e *Enum) VariantForName(name string) optional.Option[*EnumVariant] {
	for _, v := range e.Variants {
		if name == v.Name {
			return optional.Some(v)
		}
	}
	return optional.None[*EnumVariant]()
}

type EnumVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"@Ident" protobuf:"3"`
	Value    Value    `parser:"(('=' @@) | @@)!" protobuf:"4"`
}

// Equal implements Type.
func (e *EnumVariant) Equal(o *EnumVariant) bool {
	if e.Name != o.Name {
		return false
	}
	return e.Value.Equal(o.Value)
}

func (e *EnumVariant) Position() Position { return e.Pos }

func (e *EnumVariant) schemaChildren() []Node { return []Node{e.Value} }

func (e *EnumVariant) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(e.Comments))
	fmt.Fprint(w, e.Name)
	if e.Value != nil {
		if _, ok := e.Value.(*TypeValue); ok {
			fmt.Fprint(w, " ", e.Value)
		} else {
			fmt.Fprint(w, " = ", e.Value)
		}
	}
	return w.String()
}
