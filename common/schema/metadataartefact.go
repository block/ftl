package schema

import (
	"fmt"

	"github.com/block/ftl/common/sha256"
)

//protobuf:14
type MetadataArtefact struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Path       string        `parser:"'+' 'artefact' Whitespace @String" protobuf:"2"`
	Digest     sha256.SHA256 `parser:"@String" protobuf:"3"`
	Executable bool          `parser:"@'executable'?" protobuf:"4"`
}

var _ Metadata = (*MetadataArtefact)(nil)

func (m *MetadataArtefact) Position() Position { return m.Pos }

func (m *MetadataArtefact) String() string {
	return fmt.Sprintf("+artefact %q %q %t", m.Path, m.Digest, m.Executable)
}

func (m *MetadataArtefact) schemaChildren() []Node { return nil }
func (m *MetadataArtefact) schemaMetadata()        {}
