package schema

import (
	"fmt"
	"time"

	"github.com/block/ftl/common/slices"
)

type VerbRuntime struct {
	Base         VerbRuntimeBase          `protobuf:"1"`
	Subscription *VerbRuntimeSubscription `protobuf:"2,optional"`
}

//protobuf:4 RuntimeEvent
//protobuf:export
type VerbRuntimeEvent struct {
	ID      string             `protobuf:"1"`
	Payload VerbRuntimePayload `protobuf:"2"`
}

var _ RuntimeEvent = (*VerbRuntimeEvent)(nil)

func (v *VerbRuntimeEvent) runtimeEvent() {}

func (v *VerbRuntimeEvent) ApplyTo(m *Module) {
	for verb := range slices.FilterVariants[*Verb](m.Decls) {
		if verb.Name == v.ID {
			if verb.Runtime == nil {
				verb.Runtime = &VerbRuntime{}
			}

			switch payload := v.Payload.(type) {
			case *VerbRuntimeBase:
				verb.Runtime.Base = *payload
			case *VerbRuntimeSubscription:
				verb.Runtime.Subscription = payload
			default:
				panic(fmt.Sprintf("unknown verb runtime payload type: %T", payload))
			}
		}
	}
}

//sumtype:decl
type VerbRuntimePayload interface {
	verbRuntime()
}

//protobuf:1
type VerbRuntimeBase struct {
	CreateTime time.Time `protobuf:"1,optional"`
	StartTime  time.Time `protobuf:"2,optional"`
}

func (*VerbRuntimeBase) verbRuntime() {}

//protobuf:2
type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}
