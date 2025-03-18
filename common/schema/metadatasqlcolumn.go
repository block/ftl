package schema

import (
	"fmt"
	"strconv"
)

// MetadataSQLColumn designates a database column.
//
//protobuf:17,optional
type MetadataSQLColumn struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Table string `parser:"'+' 'sql' Whitespace 'column' Whitespace (@Ident | @String)" protobuf:"2"`
	Name  string `parser:"'.' (@Ident | @String)" protobuf:"3"`
}

var _ Metadata = (*MetadataSQLColumn)(nil)

func (*MetadataSQLColumn) schemaMetadata()          {}
func (m *MetadataSQLColumn) schemaChildren() []Node { return nil }
func (m *MetadataSQLColumn) Position() Position     { return m.Pos }
func (m *MetadataSQLColumn) String() string {
	table := strconv.Quote(m.Table)
	if ValidateName(m.Table) {
		table = m.Table
	}
	column := strconv.Quote(m.Name)
	if ValidateName(m.Name) {
		column = m.Name
	}
	return fmt.Sprintf("+sql column %s.%s", table, column)
}
