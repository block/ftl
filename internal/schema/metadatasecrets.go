package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// MetadataSecrets represents a metadata block with a list of config items that are used.
//
//protobuf:11,optional
type MetadataSecrets struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Secrets []*Ref `parser:"'+' 'secrets' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataSecrets)(nil)

func (m *MetadataSecrets) Position() Position { return m.Pos }
func (m *MetadataSecrets) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+secrets ")
	w := 6
	for i, secret := range m.Secrets {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := secret.String()
		if w+len(str) > 70 {
			w = 6
			fmt.Fprint(out, "\n      ")
		}
		w += len(str)
		fmt.Fprint(out, str)
	}
	fmt.Fprint(out)
	return out.String()
}

func (m *MetadataSecrets) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Secrets))
	for _, ref := range m.Secrets {
		out = append(out, ref)
	}
	return out
}
func (*MetadataSecrets) schemaMetadata() {}

func (m *MetadataSecrets) ToProto() proto.Message {
	return &schemapb.MetadataSecrets{
		Pos:     posToProto(m.Pos),
		Secrets: nodeListToProto[*schemapb.Ref](m.Secrets),
	}
}
