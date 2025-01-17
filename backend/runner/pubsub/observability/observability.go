package observability

import "fmt"

var (
	PubSub *PubSubMetrics
)

func init() {
	var err error
	PubSub, err = initPubSubMetrics()
	if err != nil {
		panic(fmt.Errorf("could not initialize pubsub metrics: %w", err))
	}
}
