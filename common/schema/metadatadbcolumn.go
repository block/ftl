package schema

import (
	"fmt"
)

// MetadataDBColumn designates a database column.
//
//protobuf:17,optional
type MetadataDBColumn struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Table string `parser:"'+' 'data' Whitespace 'column' Whitespace @String" protobuf:"2"`
	Name  string `parser:"'.' @String" protobuf:"3"`
}

var _ Metadata = (*MetadataDBColumn)(nil)

func (*MetadataDBColumn) schemaMetadata()          {}
func (m *MetadataDBColumn) schemaChildren() []Node { return nil }
func (m *MetadataDBColumn) Position() Position     { return m.Pos }
func (m *MetadataDBColumn) String() string {
	return fmt.Sprintf("+data column %q.%q", m.Table, m.Name)
}
