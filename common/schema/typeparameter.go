package schema

type TypeParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"@Ident" protobuf:"2"`
}

var _ Symbol = (*TypeParameter)(nil)

func (t *TypeParameter) Position() Position { return t.Pos }
func (t *TypeParameter) String() string     { return t.Name }

func (t *TypeParameter) schemaChildren() []Node { return nil }
func (t *TypeParameter) schemaSymbol()          {}
func (t *TypeParameter) GetName() string        { return t.Name }
