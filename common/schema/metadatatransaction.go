package schema

// MetadataTransaction identifies a verb that serves as a transaction boundary.
//
//protobuf:21,optional
type MetadataTransaction struct {
	Pos Position `parser:"'+' 'transaction'" protobuf:"1,optional"`
}

var _ Metadata = (*MetadataTransaction)(nil)

func (*MetadataTransaction) schemaMetadata()          {}
func (m *MetadataTransaction) schemaChildren() []Node { return nil }
func (m *MetadataTransaction) Position() Position     { return m.Pos }
func (m *MetadataTransaction) String() string {
	return "+transaction"
}
