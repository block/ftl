package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

// KafkaTopicSetup is an executor that sets up a Kafka topic in an existing cluster.
type KafkaTopicSetup struct {
	inputs []state.State
}

func NewKafkaTopicSetup() *KafkaTopicSetup {
	return &KafkaTopicSetup{}
}

var _ provisioner.Executor = (*KafkaTopicSetup)(nil)

func (e *KafkaTopicSetup) Prepare(ctx context.Context, input state.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}

func (e *KafkaTopicSetup) Execute(ctx context.Context) ([]state.State, error) {
	logger := log.FromContext(ctx)
	var result []state.State
	for _, input := range e.inputs {
		if input, ok := input.(state.TopicClusterReady); ok {
			topic := input.Topic

			topicID := fmt.Sprintf("%s.%s", input.Module, topic)
			logger.Debugf("Provisioning topic: %s", topicID)

			config := sarama.NewConfig()
			config.Admin.Timeout = 30 * time.Second
			admin, err := sarama.NewClusterAdmin(input.Brokers, config)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create kafka admin client")
			}
			defer admin.Close()

			topicMetas, err := admin.DescribeTopics([]string{topicID})
			if err != nil {
				return nil, errors.Wrap(err, "failed to describe topic")
			}
			if len(topicMetas) != 1 {
				return nil, errors.Errorf("expected topic metadata from kafka but received none")
			}
			if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
				logger.Debugf("Topic %s does not exist. Creating it.", topicID)
				// No topic exists yet. Create it
				err = admin.CreateTopic(topicID, &sarama.TopicDetail{
					NumPartitions:     int32(input.Partitions), //nolint:gosec
					ReplicationFactor: 1,
					ReplicaAssignment: nil,
				}, false)
				if err != nil {
					return nil, errors.Wrap(err, "failed to create topic")
				}
				logger.Infof("Kafka topic created: %s", topicID) //nolint
			} else if topicMetas[0].Err != sarama.ErrNoError {
				return nil, errors.Wrapf(topicMetas[0].Err, "failed to describe topic %q", topicID)
			} else if len(topicMetas[0].Partitions) > input.Partitions {
				var plural string
				if len(topicMetas[0].Partitions) == 1 {
					plural = "partition"
				} else {
					plural = "partitions"
				}
				logger.Warnf("Using existing topic %s with %d %s instead of %d", topicID, len(topicMetas[0].Partitions), plural, input.Partitions)
			} else if len(topicMetas[0].Partitions) < input.Partitions {
				logger.Infof("Increasing partitions for topic %s from %d to %d", topicID, len(topicMetas[0].Partitions), input.Partitions) //nolint
				if err := admin.CreatePartitions(topicID, int32(input.Partitions), nil, false); err != nil {                               //nolint:gosec
					return nil, errors.Wrap(err, "failed to increase partitions")
				}
			} else {
				logger.Debugf("Topic %s already has %d partitions", topicID, input.Partitions)
			}
			output := state.OutputTopic{
				Module: input.Module,
				Topic:  topic,
				Runtime: &schema.TopicRuntime{
					KafkaBrokers: input.Brokers,
					TopicID:      topicID,
				},
			}
			result = append(result, output)

		}
	}
	return result, nil
}
