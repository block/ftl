package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

//protobuf:2
type MetadataIngress struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type   string                 `parser:"'+' 'ingress' @('http')?" protobuf:"2"`
	Method string                 `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')" protobuf:"3"`
	Path   []IngressPathComponent `parser:"('/' @@)+" protobuf:"4"`
}

var _ Metadata = (*MetadataIngress)(nil)

func (m *MetadataIngress) Position() Position { return m.Pos }
func (m *MetadataIngress) String() string {
	return fmt.Sprintf("+ingress %s %s %s", m.Type, strings.ToUpper(m.Method), m.PathString())
}

// PathString returns the path as a string, with parameters enclosed in curly braces.
//
// For example, /foo/{bar}
func (m *MetadataIngress) PathString() string {
	path := make([]string, len(m.Path))
	for i, p := range m.Path {
		switch v := p.(type) {
		case *IngressPathLiteral:
			path[i] = v.Text
		case *IngressPathParameter:
			path[i] = fmt.Sprintf("{%s}", v.Name)
		}
	}
	return "/" + strings.Join(path, "/")
}

func (m *MetadataIngress) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Path))
	for _, ref := range m.Path {
		out = append(out, ref)
	}
	return out
}

func (*MetadataIngress) schemaMetadata() {}

func ingressPathComponentListToSchema(s []*schemapb.IngressPathComponent) []IngressPathComponent {
	var out []IngressPathComponent
	for _, n := range s {
		switch n := n.Value.(type) {
		case *schemapb.IngressPathComponent_IngressPathLiteral:
			out = append(out, &IngressPathLiteral{
				Pos:  PosFromProto(n.IngressPathLiteral.Pos),
				Text: n.IngressPathLiteral.Text,
			})
		case *schemapb.IngressPathComponent_IngressPathParameter:
			out = append(out, &IngressPathParameter{
				Pos:  PosFromProto(n.IngressPathParameter.Pos),
				Name: n.IngressPathParameter.Name,
			})
		}
	}

	return out
}

type IngressPathComponent interface {
	Node
	schemaIngressPathComponent()
}

//protobuf:1
type IngressPathLiteral struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Text string `parser:"@~(Whitespace | '/' | '{' | '}')+" protobuf:"2"`
}

var _ IngressPathComponent = (*IngressPathLiteral)(nil)

func (l *IngressPathLiteral) Position() Position        { return l.Pos }
func (l *IngressPathLiteral) String() string            { return l.Text }
func (*IngressPathLiteral) schemaChildren() []Node      { return nil }
func (*IngressPathLiteral) schemaIngressPathComponent() {}

//protobuf:2
type IngressPathParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'{' @Ident '}'" protobuf:"2"`
}

var _ IngressPathComponent = (*IngressPathParameter)(nil)

func (l *IngressPathParameter) Position() Position        { return l.Pos }
func (l *IngressPathParameter) String() string            { return l.Name }
func (*IngressPathParameter) schemaChildren() []Node      { return nil }
func (*IngressPathParameter) schemaIngressPathComponent() {}
