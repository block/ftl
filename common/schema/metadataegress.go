package schema

// MetadataEgress identifies a verb that serves as a Egress boundary.
//
//protobuf:22,optional
type MetadataEgress struct {
	Pos     Position `parser:"" protobuf:"1,optional"`
	Targets []string `parser:"'+' 'egress' @String(',' @String)*" protobuf:"2"`
}

var _ Metadata = (*MetadataEgress)(nil)

func (*MetadataEgress) schemaMetadata()          {}
func (m *MetadataEgress) schemaChildren() []Node { return nil }
func (m *MetadataEgress) Position() Position     { return m.Pos }
func (m *MetadataEgress) String() string {
	sb := "+egress"
	for _, t := range m.Targets {
		sb += " \"" + t + "\""
	}
	return sb
}
