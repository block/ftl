package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/timelineclient"
)

type publisher struct {
	module     string
	deployment key.Deployment
	topic      *schema.Topic
	producer   sarama.SyncProducer

	timelineClient *timelineclient.Client
}

func newPublisher(module string, t *schema.Topic, deployment key.Deployment, timelineClient *timelineclient.Client) (*publisher, error) {
	if t.Runtime == nil {
		return nil, fmt.Errorf("topic %s has no runtime", t.Name)
	}
	if len(t.Runtime.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("topic %s has no Kafka brokers", t.Name)
	}
	if t.Runtime.TopicID == "" {
		return nil, fmt.Errorf("topic %s has no topic ID", t.Name)
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(t.Runtime.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer for topic %s: %w", t.Name, err)
	}
	return &publisher{
		module:     module,
		deployment: deployment,
		topic:      t,
		producer:   producer,

		timelineClient: timelineClient,
	}, nil
}

func (p *publisher) publish(ctx context.Context, data []byte, key string, caller schema.Ref) error {
	logger := log.FromContext(ctx)
	requestKey, err := rpc.RequestKeyFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get request key: %w", err)
	}
	var requestKeyStr optional.Option[string]
	if r, ok := requestKey.Get(); ok {
		requestKeyStr = optional.Some(r.String())
	}

	timelineEvent := timelineclient.PubSubPublish{
		DeploymentKey: p.deployment,
		RequestKey:    requestKeyStr,
		Time:          time.Now(),
		SourceVerb:    caller,
		Topic:         p.topic.Name,
		Request:       data,
	}

	partition, offset, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic.Runtime.TopicID,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(key),
	})
	if err != nil {
		timelineEvent.Error = optional.Some(err.Error())
		logger.Errorf(err, "Failed to publish message to %s", p.topic.Name)
		return fmt.Errorf("failed to publish message to %s: %w", p.topic.Name, err)
	}
	timelineEvent.Partition = int(partition)
	timelineEvent.Offset = int(offset)
	p.timelineClient.Publish(ctx, timelineEvent)
	logger.Debugf("Published to %v[%v:%v]", p.topic.Name, partition, offset)
	return nil
}
