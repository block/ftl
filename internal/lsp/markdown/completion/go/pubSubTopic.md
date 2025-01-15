Declare a topic for publishing events.

Topics are where events are sent in the PubSub system. A topic can be exported to allow other modules to subscribe to it. Each topic can have multiple subscriptions.

```go
type PubSubEvent struct {
	Time time.Time
}

//ftl:export
type TestTopic = ftl.TopicHandle[PubSubEvent, ftl.SinglePartitionMap[PubSubEvent]]

//ftl:verb
func Publish(ctx context.Context, topic TestTopic) error {
	logger := ftl.LoggerFromContext(ctx)
	t := time.Now()
	logger.Infof("Publishing %v", t)
	return topic.Publish(ctx, PubSubEvent{Time: t})
}
```

See https://block.github.io/ftl/docs/reference/pubsub/
---

type ${1:Event} struct {
	${2:// Event fields}
}

//ftl:export
type ${3:TopicName} = ftl.TopicHandle[${1:Event}, ftl.SinglePartitionMap[${1:Event}]]

//ftl:verb
func ${4:Publish}(ctx context.Context, topic ${3:TopicName}) error {
	return topic.Publish(ctx, ${1:Event}{})
}
