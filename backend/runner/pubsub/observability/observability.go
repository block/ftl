package observability

import "github.com/alecthomas/errors"

var (
	PubSub *PubSubMetrics
)

func init() {
	var err error
	PubSub, err = initPubSubMetrics()
	if err != nil {
		panic(errors.Wrap(err, "could not initialize pubsub metrics"))
	}
}
