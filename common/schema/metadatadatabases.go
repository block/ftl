package schema

import (
	"fmt"
	"strings"
)

//protobuf:4
type MetadataDatabases struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Calls []*Ref `parser:"'+' 'database' 'uses' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataDatabases)(nil)

func (m *MetadataDatabases) Append(call *Ref)   { m.Calls = append(m.Calls, call) }
func (m *MetadataDatabases) Position() Position { return m.Pos }
func (m *MetadataDatabases) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+database uses ")
	w := 6
	for i, call := range m.Calls {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := call.String()
		if w+len(str) > 70 {
			w = 6
			fmt.Fprint(out, "\n      ")
		}
		w += len(str)
		fmt.Fprint(out, str)
	}
	fmt.Fprint(out)
	return out.String()
}

func (m *MetadataDatabases) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Calls))
	for _, ref := range m.Calls {
		out = append(out, ref)
	}
	return out
}
func (*MetadataDatabases) schemaMetadata() {}
