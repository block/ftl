package schema

//protobuf:20
type MetadataFixture struct {
	Pos    Position `parser:"" protobuf:"1,optional"`
	Manual bool     `parser:"'+' 'fixture' @'manual'?" protobuf:"2"`
}

var _ Metadata = (*MetadataFixture)(nil)

func (m *MetadataFixture) Position() Position { return m.Pos }
func (m *MetadataFixture) String() string {
	if m.Manual {
		return "+fixture manual"
	}
	return "+fixture"
}

func (m *MetadataFixture) schemaChildren() []Node {
	return nil
}

func (*MetadataFixture) schemaMetadata() {}
