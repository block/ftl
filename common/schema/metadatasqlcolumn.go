package schema

import (
	"fmt"
)

// MetadataSQLColumn designates a database column.
//
//protobuf:17,optional
type MetadataSQLColumn struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Table string `parser:"'+' 'sql' Whitespace 'column' Whitespace @String" protobuf:"2"`
	Name  string `parser:"'.' @String" protobuf:"3"`
}

var _ Metadata = (*MetadataSQLColumn)(nil)

func (*MetadataSQLColumn) schemaMetadata()          {}
func (m *MetadataSQLColumn) schemaChildren() []Node { return nil }
func (m *MetadataSQLColumn) Position() Position     { return m.Pos }
func (m *MetadataSQLColumn) String() string {
	return fmt.Sprintf("+sql column %q.%q", m.Table, m.Name)
}
