package schema

import (
	"strconv"
)

var _ Value = (*IntValue)(nil)

//protobuf:2
//protobuf:17 Type
type IntValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value int `parser:"@Number" protobuf:"2"`
}

func (x *IntValue) Equal(other Type) bool {
	o, ok := other.(*IntValue)
	if !ok {
		return false
	}
	return x.Value == o.Value
}

func (x *IntValue) Kind() Kind         { return KindIntValue }
func (x *IntValue) schemaType()        {}
func (i *IntValue) Position() Position { return i.Pos }

func (i *IntValue) schemaChildren() []Node { return nil }

func (i *IntValue) String() string {
	return strconv.Itoa(i.Value)
}

func (i *IntValue) GetValue() any { return i.Value }

func (*IntValue) schemaValueType() Type { return &Int{} }
