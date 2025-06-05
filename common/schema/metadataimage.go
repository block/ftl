package schema

import (
	"fmt"

	"github.com/block/ftl/common/sha256"
)

//protobuf:23
type MetadataImage struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Image  string        `parser:"'+' 'image' Whitespace @String" protobuf:"2"`
	Digest sha256.SHA256 `parser:"':' @String" protobuf:"3"`
}

var _ Metadata = (*MetadataImage)(nil)

func (m *MetadataImage) Position() Position { return m.Pos }

func (m *MetadataImage) String() string {
	return fmt.Sprintf("+artefact %q@%q", m.Image, m.Digest)
}

func (m *MetadataImage) schemaChildren() []Node { return nil }
func (m *MetadataImage) schemaMetadata()        {}
