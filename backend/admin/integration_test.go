//go:build integration

package admin

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/alecthomas/assert/v2"
	in "github.com/block/ftl/internal/integration"
)

func TestResetSubscription(t *testing.T) {
	in.Run(t,
		in.WithPubSub(),

		in.CopyModule("resetsub"),
		in.Deploy("resetsub"),

		// Create enough events that are quick to consume so that each partition has a bunch
		in.Call[in.Obj, in.Obj]("resetsub", "publish", in.Obj{
			"count":    1_000,
			"duration": 50, // 50ms
		}, nil),
		// Create many events that are slow to consume to make sure we never naturally consume them all
		in.Call[in.Obj, in.Obj]("resetsub", "publish", in.Obj{
			"count":    1_000,
			"duration": 10_000, //10s
		}, nil),

		in.Repeat(3, in.Chain(
			in.Sleep(time.Second*2),
			// Reset to latest
			resetSubscription("resetsub.consume", true),
			// Confirm that all partitions are on latest
			checkOffsets("resetsub.topic", "resetsub.consume", true),
			// Reset to earliest
			resetSubscription("resetsub.consume", false),
			// Confirm that all partitions are not on latest (some events may have been consumed by this point)
			checkOffsets("resetsub.topic", "resetsub.consume", false),
		)),
	)
}

// resets a subscription to the latest or earliest offset
func resetSubscription(subscription string, latest bool) in.Action {
	var offsetFlag string
	if latest {
		offsetFlag = "--latest"
	} else {
		offsetFlag = "--beginning"
	}
	return in.Exec("ftl", "pubsub", "subscription", "reset", subscription, offsetFlag)
}

// checks if a subscription is at the latest offset
func checkOffsets(topic, subscription string, latest bool) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		if latest {
			in.Infof("Checking if subscription %s is at the latest offset", subscription)
		} else {
			in.Infof("Checking if subscription %s is not at the latest offset", subscription)
		}
		config := sarama.NewConfig()
		client, err := sarama.NewClient(in.RedPandaBrokers, config)
		assert.NoError(t, err)
		admin, err := sarama.NewClusterAdmin(in.RedPandaBrokers, config)
		assert.NoError(t, err)

		partitionCount, err := kafkaPartitionCount(ic.Context, in.RedPandaBrokers, topic)
		assert.NoError(t, err)

		requestedPartitions := []int32{}
		for p := range int32(partitionCount) {
			requestedPartitions = append(requestedPartitions, p)
		}
		offsetsResp, err := admin.ListConsumerGroupOffsets(subscription, map[string][]int32{topic: requestedPartitions})
		assert.NoError(t, err)
		for p, block := range offsetsResp.Blocks[topic] {
			latestOffset, err := client.GetOffset(topic, p, sarama.OffsetNewest)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, latestOffset, sarama.OffsetNewest, "did not expect a partition to have the latest offset equal to 0 which means we can't tell if resets are working")
			if latest {
				assert.Equal(t, latestOffset, block.Offset, "expected partition %d to be at the latest offset", p)
			} else {
				assert.NotEqual(t, latestOffset, block.Offset, "expected partition %d to not be at the latest offset", p)
			}
		}
	}
}
