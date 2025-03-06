package schema

import (
	"fmt"
)

//protobuf:19
type MetadataGit struct {
	Pos        Position `parser:"" protobuf:"1,optional"`
	Repository string   `parser:"'+' 'git' Whitespace @String" protobuf:"2"`
	Commit     string   `parser:"@String" protobuf:"3"`
	Dirty      bool     `parser:"@'dirty'?" protobuf:"4"`
}

var _ Metadata = (*MetadataGit)(nil)

func (m *MetadataGit) Position() Position { return m.Pos }

func (m *MetadataGit) String() string {
	if m.Dirty {
		return fmt.Sprintf("+git \"%s\" \"%s\" dirty", m.Repository, m.Commit)
	}
	return fmt.Sprintf("+git \"%s\" \"%s\"", m.Repository, m.Commit)
}

func (m *MetadataGit) schemaChildren() []Node { return nil }
func (m *MetadataGit) schemaMetadata()        {}
