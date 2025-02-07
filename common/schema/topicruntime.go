package schema

//protobuf:5
type TopicRuntime struct {
	KafkaBrokers []string `parser:"" protobuf:"1"`
	TopicID      string   `parser:"" protobuf:"2"`
}

var _ Runtime = (*TopicRuntime)(nil)

func (m *TopicRuntime) runtimeElement() {}
