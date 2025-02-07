package schema

//protobuf:4
type VerbRuntime struct {
	Subscription *VerbRuntimeSubscription `protobuf:"1,optional"`
}

var _ Runtime = (*VerbRuntime)(nil)

func (m *VerbRuntime) runtimeElement() {
}

type VerbRuntimeSubscription struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*VerbRuntimeSubscription) verbRuntime() {}
