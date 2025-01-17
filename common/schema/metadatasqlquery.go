package schema

import (
	"fmt"
)

// MetadataSQLQuery designates a query verb; a verb generated from a SQL query.
//
//protobuf:16,optional
type MetadataSQLQuery struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Command string `parser:"'+' 'sql' Whitespace 'query' Whitespace ':' @Ident" protobuf:"2"`
	Query   string `parser:"Whitespace @String" protobuf:"3"`
}

var _ Metadata = (*MetadataSQLMigration)(nil)

func (*MetadataSQLQuery) schemaMetadata()          {}
func (m *MetadataSQLQuery) schemaChildren() []Node { return nil }
func (m *MetadataSQLQuery) Position() Position     { return m.Pos }
func (m *MetadataSQLQuery) String() string {
	return fmt.Sprintf("+sql query :%s %q", m.Command, m.Query)
}
