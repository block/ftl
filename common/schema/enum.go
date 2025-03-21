package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
)

//protobuf:4
type Enum struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string       `parser:"@Comment*" protobuf:"2"`
	Export   bool           `parser:"@'export'?" protobuf:"3"`
	Name     string         `parser:"'enum' @Ident" protobuf:"4"`
	Type     Type           `parser:"(':' @@)?" protobuf:"5,optional"`
	Variants []*EnumVariant `parser:"'{' @@* '}'" protobuf:"6"`
}

var _ Decl = (*Enum)(nil)
var _ Symbol = (*Enum)(nil)

func (e *Enum) Position() Position { return e.Pos }

func (e *Enum) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(e.Comments))
	if e.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "enum %s", e.Name)
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
func (e *Enum) GetName() string   { return e.Name }
func (e *Enum) IsExported() bool  { return e.Export }
func (e *Enum) IsGenerated() bool { return false }

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
