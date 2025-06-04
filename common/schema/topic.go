package schema

import (
	"fmt"
	"strings"
)

//protobuf:9
type Topic struct {
	Pos     Position      `parser:"" protobuf:"1,optional"`
	Runtime *TopicRuntime `parser:"" protobuf:"31634,optional"`

	Comments   []string   `parser:"@Comment*" protobuf:"2"`
	Visibility Visibility `parser:"@@?" protobuf:"3"`
	Name       string     `parser:"'topic' @Ident" protobuf:"4"`
	Event      Type       `parser:"@@" protobuf:"5"`
	Metadata   []Metadata `parser:"@@*" protobuf:"6"`
}

var _ Decl = (*Topic)(nil)
var _ Symbol = (*Topic)(nil)
var _ Provisioned = (*Topic)(nil)

func (t *Topic) Position() Position { return t.Pos }
func (*Topic) schemaDecl()          {}
func (*Topic) schemaSymbol()        {}
func (t *Topic) provisioned()       {}
func (t *Topic) schemaChildren() []Node {
	children := []Node{}
	for _, c := range t.Metadata {
		children = append(children, c)
	}
	if t.Event != nil {
		children = append(children, t.Event)
	}
	return children
}

func (t *Topic) GetName() string           { return t.Name }
func (t *Topic) GetVisibility() Visibility { return t.Visibility }
func (t *Topic) IsGenerated() bool         { return false }

func (t *Topic) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(t.Comments))
	formatTokens(w,
		t.Visibility.String(),
		"topic",
		t.Name,
		t.Event.String(),
	)
	fmt.Fprint(w, indent(encodeMetadata(t.Metadata)))
	return w.String()
}
func (t *Topic) GetProvisioned() ResourceSet {
	return ResourceSet{
		{Kind: ResourceTypeTopic, Config: &Topic{Name: t.Name}, State: t.Runtime},
	}
}

func (t *Topic) ResourceID() string {
	return t.Name
}
