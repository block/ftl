package schema

type VerbRuntime struct {
	Subscription *VerbRuntimeSubscription `protobuf:"1,optional"`
}

type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}
