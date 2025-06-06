// Code generated by FTL. DO NOT EDIT.
package pubsub

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server"
)

type CookPizzaClient func(context.Context, Pizza) error

type DeliverPizzaClient func(context.Context, Pizza) error

type OrderPizzaClient func(context.Context, OrderPizzaRequest) (OrderPizzaResponse, error)

func init() {
	reflection.Register(

		reflection.ProvideResourcesForVerb(
			CookPizza,
			server.TopicHandle[Pizza, ftl.SinglePartitionMap[Pizza]]("pubsub", "pizzaReadyTopic"),
		),

		reflection.ProvideResourcesForVerb(
			DeliverPizza,
		),

		reflection.ProvideResourcesForVerb(
			OrderPizza,
			server.TopicHandle[Pizza, PizzaPartitionMapper]("pubsub", "newOrderTopic"),
		),
	)
}
