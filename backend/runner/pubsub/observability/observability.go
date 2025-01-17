package observability

import "fmt"

var (
	PubSub *PubSubMetrics
	// TODO: remove
	AsyncCalls *AsyncCallMetrics
)

func init() {
	var err error
	PubSub, err = initPubSubMetrics()
	if err != nil {
		panic(fmt.Errorf("could not initialize pubsub metrics: %w", err))
	}

	AsyncCalls, err = initAsyncCallMetrics()
	if err != nil {
		panic(fmt.Errorf("could not initialize async call metrics: %w", err))
	}
}
