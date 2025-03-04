package schema

//protobuf:4
type VerbRuntime struct {
	SubscriptionConnector SubscriptionConnector `protobuf:"1,optional"`
}

var _ Runtime = (*VerbRuntime)(nil)

func (m *VerbRuntime) runtimeElement() {
}

// SubscriptionConnector is a connector to subscribe to a topic.
type SubscriptionConnector interface {
	subscriptionConnector()
}

// PlaintextKafkaSubscriptionConnector is a non TLS subscription connector to a kafka cluster.
//
//protobuf:1
type PlaintextKafkaSubscriptionConnector struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (*PlaintextKafkaSubscriptionConnector) subscriptionConnector() {}
