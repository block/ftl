package schema

import (
	"time"
)

type VerbRuntime struct {
	Base         VerbRuntimeBase          `protobuf:"1"`
	Subscription *VerbRuntimeSubscription `protobuf:"2,optional"`
}

type VerbRuntimeBase struct {
	CreateTime time.Time `protobuf:"1,optional"`
	StartTime  time.Time `protobuf:"2,optional"`
}

func (*VerbRuntimeBase) verbRuntime() {}

type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}
