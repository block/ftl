package schema

var _ Value = (*TypeValue)(nil)

//protobuf:3
//protobuf:19 Type
type TypeValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value Type `parser:"@@" protobuf:"2"`
}

func (x *TypeValue) Equal(other Type) bool {
	o, ok := other.(*TypeValue)
	if !ok {
		return false
	}
	return x.Value.Equal(o.Value)
}

func (x *TypeValue) Kind() Kind { return KindTypeValue }

func (x *TypeValue) schemaType() {}

func (t *TypeValue) Position() Position { return t.Pos }

func (t *TypeValue) schemaChildren() []Node { return []Node{t.Value} }

func (t *TypeValue) String() string {
	return t.Value.String()
}

func (t *TypeValue) GetValue() any { return t.Value.String() }

func (t *TypeValue) schemaValueType() Type { return t.Value }
