package resetsub

import (
	"context"
	"strconv"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
)

type PublishRequest struct {
	Count int
	// Duration to consume each event in milliseconds
	Duration float64
}

type PublishResponse struct {
}

type Event struct {
	ID string
	// Duration to consume each event in milliseconds
	Duration float64
}

type PartitionMap struct {
}

var _ ftl.TopicPartitionMap[Event] = PartitionMap{}

func (PartitionMap) PartitionKey(e Event) string {
	return e.ID
}

//ftl:topic partitions=16
type Topic = ftl.TopicHandle[Event, PartitionMap]

//ftl:verb export
func Publish(ctx context.Context, req PublishRequest, topic Topic) (PublishResponse, error) {
	for i := range req.Count {
		err := topic.Publish(ctx, Event{ID: strconv.Itoa(i), Duration: req.Duration})
		if err != nil {
			return PublishResponse{}, err
		}
	}
	return PublishResponse{}, nil
}

//ftl:verb
//ftl:subscribe topic from=beginning
func Consume(ctx context.Context, e Event) error {
	ftl.LoggerFromContext(ctx).Infof("Consuming event %s", e.ID)
	time.Sleep(time.Duration(e.Duration) * time.Millisecond)
	return nil
}
