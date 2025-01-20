//protobuf:package xyz.block.ftl.go2proto.test
//protobuf:option go_package="github.com/block/ftl/cmd/go2proto/testdata/testdatapb"
package testdata

import (
	"errors"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/block/ftl/cmd/go2proto/testdata/external"
	"github.com/block/ftl/internal/key"
)

//protobuf:export
type Root struct {
	Int             int                        `protobuf:"1"`
	String          string                     `protobuf:"2"`
	MessagePtr      *Message                   `protobuf:"4"`
	Enum            Enum                       `protobuf:"5"`
	SumType         SumType                    `protobuf:"6"`
	OptionalInt     int                        `protobuf:"7,optional"`
	OptionalIntPtr  *int                       `protobuf:"8,optional"`
	OptionalMsg     *Message                   `protobuf:"9,optional"`
	RepeatedInt     []int                      `protobuf:"10"`
	RepeatedMsg     []*Message                 `protobuf:"11"`
	URL             *url.URL                   `protobuf:"12"`
	OptionalWrapper optional.Option[string]    `protobuf:"13"`
	ExternalRoot    external.Root              `protobuf:"14"`
	Key             key.Deployment             `protobuf:"15"`
	OptionalTime    optional.Option[time.Time] `protobuf:"16"`
	OptionalMessage optional.Option[Message]   `protobuf:"17"`
}

type Message struct {
	Time           time.Time     `protobuf:"1"`
	OptTime        time.Time     `protobuf:"2,optional"`
	Duration       time.Duration `protobuf:"3"`
	Invalid        bool          `protobuf:"4"`
	Nested         Nested        `protobuf:"5"`
	RepeatedNested []Nested      `protobuf:"6"`
}

func (m *Message) Validate() error {
	if m.Invalid {
		return errors.New("invalid message")
	}
	return nil
}

type Nested struct {
	Nested string `protobuf:"1"`
}

type Enum int

const (
	EnumA Enum = iota
	EnumB
)

type SumType interface {
	sumType()
}

//protobuf:1
type SumTypeA struct {
	A string `protobuf:"1"`
}

func (SumTypeA) sumType() {}

//protobuf:2
type SumTypeB struct {
	B int `protobuf:"1"`
}

func (SumTypeB) sumType() {}

//protobuf:3
type SumTypeC struct {
	C float64 `protobuf:"1"`
}

func (SumTypeC) sumType() {}

type SubSumType interface {
	SumType

	subSumType()
}

//protobuf:1
//protobuf:4 SumType
type SubSumTypeA struct {
	A string `protobuf:"1"`
}

func (SubSumTypeA) subSumType() {}
func (SubSumTypeA) sumType()    {}

//protobuf:2
//protobuf:5 SumType
type SubSumTypeB struct {
	A string `protobuf:"1"`
}

func (SubSumTypeB) subSumType() {}
func (SubSumTypeB) sumType()    {}
