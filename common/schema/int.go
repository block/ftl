package schema

//protobuf:1
type Int struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Int bool `parser:"@'Int'" protobuf:"-"`
}

var _ Type = (*Int)(nil)
var _ Symbol = (*Int)(nil)

func (i *Int) Equal(other Type) bool { _, ok := other.(*Int); return ok }
func (i *Int) Position() Position    { return i.Pos }
func (*Int) schemaSymbol()           {}
func (*Int) schemaChildren() []Node  { return nil }
func (*Int) schemaType()             {}
func (*Int) String() string          { return "Int" }
func (*Int) GetName() string         { return "Int" }
func (*Int) Kind() Kind              { return KindInt }
