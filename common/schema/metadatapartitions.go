package schema

import "strconv"

//protobuf:15
type MetadataPartitions struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Partitions int `parser:"'+' 'partitions' @Number" protobuf:"2"`
}

var _ Metadata = (*MetadataPartitions)(nil)

func (m *MetadataPartitions) Position() Position { return m.Pos }

func (m *MetadataPartitions) String() string {
	return "+partitions " + strconv.Itoa(m.Partitions)
}

func (m *MetadataPartitions) schemaChildren() []Node { return nil }
func (m *MetadataPartitions) schemaMetadata()        {}
