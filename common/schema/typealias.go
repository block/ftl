package schema

import (
	"fmt"
	"strings"
)

//protobuf:5
//protobuf:16 Type
type TypeAlias struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Export   bool       `parser:"@'export'?" protobuf:"3"`
	Name     string     `parser:"'typealias' @Ident" protobuf:"4"`
	Type     Type       `parser:"@@" protobuf:"5"`
	Metadata []Metadata `parser:"@@*" protobuf:"6"`
}

var _ Type = (*TypeAlias)(nil)
var _ Decl = (*TypeAlias)(nil)
var _ Symbol = (*TypeAlias)(nil)

func (t *TypeAlias) Position() Position { return t.Pos }

func (t *TypeAlias) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(t.Comments))
	if t.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "typealias %s %s", t.Name, t.Type)
	fmt.Fprint(w, indent(encodeMetadata(t.Metadata)))
	return w.String()
}
func (*TypeAlias) schemaDecl()   {}
func (*TypeAlias) schemaSymbol() {}
func (t *TypeAlias) schemaChildren() []Node {
	children := make([]Node, 0, len(t.Metadata)+1)
	for _, m := range t.Metadata {
		children = append(children, m)
	}
	if t.Type != nil {
		children = append(children, t.Type)
	}
	return children
}

func (t *TypeAlias) GetName() string   { return t.Name }
func (t *TypeAlias) IsExported() bool  { return t.Export }
func (t *TypeAlias) IsGenerated() bool { return false }
func (t *TypeAlias) Equal(other Type) bool {
	o, ok := other.(*TypeAlias)
	if !ok {
		return false
	}
	return t.Name == o.Name && t.Type.Equal(o.Type)
}

func (t *TypeAlias) Kind() Kind  { return KindTypeAlias }
func (t *TypeAlias) schemaType() {}
