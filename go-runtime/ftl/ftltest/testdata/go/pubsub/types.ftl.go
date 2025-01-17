// Code generated by FTL. DO NOT EDIT.
package pubsub

import (
	"context"
	"github.com/block/ftl/go-runtime/ftl/reflection"
	"github.com/block/ftl/go-runtime/server"
)

type ConsumeEventClient func(context.Context, Event) error

type ErrorsAfterASecondClient func(context.Context, Event) error

type PropagateToTopic2Client func(context.Context, Event) error

type PublishToTopicOneClient func(context.Context, Event) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			ConsumeEvent,
		),
		reflection.ProvideResourcesForVerb(
			ErrorsAfterASecond,
		),
		reflection.ProvideResourcesForVerb(
			PropagateToTopic2,
			server.TopicHandle[Event]("pubsub", "topic2"),
		),
		reflection.ProvideResourcesForVerb(
			PublishToTopicOne,
			server.TopicHandle[Event]("pubsub", "topic1"),
		),
	)
}
