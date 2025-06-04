package pubsub

import (
	"context"

	"github.com/alecthomas/errors" // Import the FTL SDK.

	"github.com/block/ftl/go-runtime/ftl"
)

type PizzaPartitionMapper struct{}

func (PizzaPartitionMapper) PartitionKey(pizza Pizza) string {
	return pizza.Customer
}

//ftl:export
type NewOrderTopic = ftl.TopicHandle[Pizza, PizzaPartitionMapper]

//ftl:export
type PizzaReadyTopic = ftl.TopicHandle[Pizza, ftl.SinglePartitionMap[Pizza]]

type Pizza struct {
	ID       int
	Type     string
	Customer string
}

//ftl:verb
//ftl:subscribe newOrderTopic from=beginning
func CookPizza(ctx context.Context, pizza Pizza, topic PizzaReadyTopic) error {
	ftl.LoggerFromContext(ctx).Infof("Cooking pizza: %v", pizza)
	return errors.WithStack(topic.Publish(ctx, pizza))
}
