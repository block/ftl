package schema

import (
	"fmt"
	"strings"
)

//protobuf:7
type Secret struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"'secret' @Ident" protobuf:"3"`
	Type     Type     `parser:"@@" protobuf:"4"`
}

var _ Decl = (*Secret)(nil)
var _ Symbol = (*Secret)(nil)

func (s *Secret) Validate() error {
	if s.Name == "" {
		return errorf(s, "secret name cannot be empty")
	}
	if s.Type == nil {
		return errorf(s, "%s: missing secret type", s.Name)
	}
	return nil
}
func (s *Secret) GetName() string           { return s.Name }
func (s *Secret) GetVisibility() Visibility { return VisibilityScopeNone }
func (s *Secret) IsGenerated() bool         { return false }
func (s *Secret) Position() Position        { return s.Pos }
func (s *Secret) String() string {
	w := &strings.Builder{}

	fmt.Fprint(w, EncodeComments(s.Comments))
	fmt.Fprintf(w, "secret %s %s", s.Name, s.Type)

	return w.String()
}

func (s *Secret) schemaChildren() []Node { return []Node{s.Type} }
func (s *Secret) schemaDecl()            {}
func (s *Secret) schemaSymbol()          {}
