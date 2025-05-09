package pubsub

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type PartitionMapper struct{}

var _ ftl.TopicPartitionMap[PayinEvent] = PartitionMapper{}

func (PartitionMapper) PartitionKey(event PayinEvent) string {
	return event.Name
}

type PayinEvent struct {
	Name string
}

//ftl:topic partitions=4
type Payins = ftl.TopicHandle[PayinEvent, PartitionMapper]

//ftl:verb
func Payin(ctx context.Context, topic Payins) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Test"}); err != nil {
		return errors.Wrap(err, "failed to publish event")
	}
	return nil
}

//ftl:verb
//ftl:subscribe payins from=beginning
func ProcessPayin(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received PubSub event: %v", event)
	return nil
}

// publicBroadcast is a topic that broadcasts payin events to the public.
// out of order with subscription registration to test ordering doesn't matter.
//
//ftl:export
type PublicBroadcast = ftl.TopicHandle[PayinEvent, ftl.SinglePartitionMap[PayinEvent]]

//ftl:verb export
func Broadcast(ctx context.Context, topic PublicBroadcast) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Broadcast"}); err != nil {
		return errors.Wrap(err, "failed to publish broadcast event")
	}
	return nil
}

//ftl:verb
//ftl:subscribe publicBroadcast from=beginning
//ftl:retry 10 1s
func ProcessBroadcast(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received broadcast event: %v", event)
	return nil
}
