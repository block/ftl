package schema

//protobuf:4
type VerbRuntime struct {
	SubscriptionConnector SubscriptionConnector `protobuf:"1,optional"`
	EgressRuntime         *EgressRuntime        `protobuf:"2,optional"`
}

var _ Runtime = (*EgressRuntime)(nil)
var _ Runtime = (SubscriptionConnector)(nil)
var _ SubscriptionConnector = (*PlaintextKafkaSubscriptionConnector)(nil)

// SubscriptionConnector is a connector to subscribe to a topic.
type SubscriptionConnector interface {
	subscriptionConnector()
	runtimeElement()
}

// PlaintextKafkaSubscriptionConnector is a non TLS subscription connector to a kafka cluster.
//
//protobuf:8
type PlaintextKafkaSubscriptionConnector struct {
	KafkaBrokers []string `protobuf:"1"`
}

func (x *PlaintextKafkaSubscriptionConnector) runtimeElement() {

}

func (*PlaintextKafkaSubscriptionConnector) subscriptionConnector() {}

// EgressRuntime stores the actual egress target.
//
//protobuf:7
type EgressRuntime struct {
	Targets []EgressElement `protobuf:"1"`
}

func (e EgressRuntime) runtimeElement() {

}

type EgressElement struct {
	Expression string `protobuf:"1"`
	Target     string `protobuf:"2"`
}
