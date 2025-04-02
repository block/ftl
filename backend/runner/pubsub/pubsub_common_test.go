//go:build infrastructure || integration

package pubsub_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/common/slices"
	in "github.com/block/ftl/internal/integration"
)

func setupPubsubTests() []in.ActionOrOption {

	calls := 20
	events := calls * 10
	return mapValues(
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),

		// After a deployment is "ready" it can take a second before a consumer group claims partitions.
		// "publisher.local" has "from=latest" so we need that group to be ready before we start publishing
		// otherwise it will start from the latest offset after claiming partitions.
		in.Sleep(time.Second*1),

		// publish half the events before subscriber is deployed
		publishToTestAndLocalTopics(calls/2),

		in.Deploy("subscriber"),

		// publish the other half of the events after subscriber is deployed
		publishToTestAndLocalTopics(calls/2),

		in.Sleep(time.Second*4),

		// check that there are the right amount of consumed events, depending on "from" offset option
		checkConsumed("publisher", "local", true, events, optional.None[string]()),
		checkConsumed("subscriber", "consume", true, events, optional.None[string]()),
		checkConsumed("subscriber", "consumeFromLatest", true, events/2, optional.None[string]()),
	)
}

func mapValues(option ...in.ActionOrOption) []in.ActionOrOption {
	return option
}

func publishToTestAndLocalTopics(calls int) in.Action {
	// do this in parallel because we want to test race conditions
	return func(t testing.TB, ic in.TestContext) {
		actions := []in.Action{
			in.Repeat(calls, in.Call("publisher", "publishTen", in.Obj{}, func(t testing.TB, resp in.Obj) {})),
			in.Repeat(calls, in.Call("publisher", "publishTenLocal", in.Obj{}, func(t testing.TB, resp in.Obj) {})),
		}
		wg := &sync.WaitGroup{}
		for _, action := range actions {
			wg.Add(1)
			go func() {
				action(t, ic)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func checkConsumed(module, verb string, success bool, count int, needle optional.Option[string]) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		if needle, ok := needle.Get(); ok {
			in.Infof("Checking for %v call(s) to %s.%s with needle %v", count, module, verb, needle)
		} else {
			in.Infof("Checking for %v call(s) to %s.%s", count, module, verb)
		}
		resp, err := ic.Timeline.GetTimeline(ic.Context, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Query: &timelinepb.TimelineQuery{
				Limit: 100000,
				Filters: []*timelinepb.TimelineQuery_Filter{
					{
						Filter: &timelinepb.TimelineQuery_Filter_EventTypes{
							EventTypes: &timelinepb.TimelineQuery_EventTypeFilter{
								EventTypes: []timelinepb.EventType{
									timelinepb.EventType_EVENT_TYPE_CALL,
								},
							},
						},
					},
					{
						Filter: &timelinepb.TimelineQuery_Filter_Module{
							Module: &timelinepb.TimelineQuery_ModuleFilter{
								Module: module,
								Verb:   &verb,
							},
						},
					},
				},
			},
		}))
		assert.NoError(t, err)
		calls := slices.Filter(slices.Map(resp.Msg.Events, func(e *timelinepb.Event) *timelinepb.CallEvent {
			return e.GetCall()
		}), func(c *timelinepb.CallEvent) bool {
			if c == nil {
				return false
			}
			assert.NotEqual(t, nil, c.RequestKey, "pub sub calls need a request key")
			assert.NoError(t, err)
			if needle, ok := needle.Get(); ok && !strings.Contains(c.Request, needle) {
				return false
			}
			return true
		})
		successfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error == nil
		})
		unsuccessfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error != nil
		})
		if success {
			assert.Equal(t, count, len(successfulCalls), "expected %v successful calls (failed calls: %v)", count, len(unsuccessfulCalls))
		} else {
			assert.Equal(t, count, len(unsuccessfulCalls), "expected %v unsuccessful calls (successful calls: %v)", count, len(successfulCalls))
		}
	}
}

func checkPublished(module, topic string, count int) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking for %v published events for %s.%s", count, module, topic)
		resp, err := ic.Timeline.GetTimeline(ic.Context, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Query: &timelinepb.TimelineQuery{
				Limit: 100000,
				Filters: []*timelinepb.TimelineQuery_Filter{
					{
						Filter: &timelinepb.TimelineQuery_Filter_EventTypes{
							EventTypes: &timelinepb.TimelineQuery_EventTypeFilter{
								EventTypes: []timelinepb.EventType{
									timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH,
								},
							},
						},
					},
				},
			},
		}))
		assert.NoError(t, err)
		events := slices.Filter(slices.Map(resp.Msg.Events, func(e *timelinepb.Event) *timelinepb.PubSubPublishEvent {
			return e.GetPubsubPublish()
		}), func(e *timelinepb.PubSubPublishEvent) bool {
			if e == nil {
				return false
			}
			if e.Topic != topic {
				return false
			}
			return true
		})
		assert.Equal(t, count, len(events), "expected %v published events", count)
	}
}

func checkGroupMembership(module, subscription string, expectedCount int) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		consumerGroup := module + "." + subscription
		in.Infof("Checking group membership for %v", consumerGroup)

		client, err := sarama.NewClient(in.RedPandaBrokers, sarama.NewConfig())
		assert.NoError(t, err)
		defer client.Close()

		clusterAdmin, err := sarama.NewClusterAdminFromClient(client)
		assert.NoError(t, err)
		defer clusterAdmin.Close()

		groups, err := clusterAdmin.DescribeConsumerGroups([]string{consumerGroup})
		assert.NoError(t, err)
		assert.Equal(t, len(groups), 1)
		assert.Equal(t, len(groups[0].Members), expectedCount, "expected consumer group %v to have %v members", consumerGroup, expectedCount)
	}
}
