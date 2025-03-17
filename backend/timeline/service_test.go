package timeline

import (
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/internal/channels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetTimeline(t *testing.T) {
	t.Parallel()
	t.Run("Limits", func(t *testing.T) {
		t.Parallel()

		entryCount := 100
		service := createTestService(t, callEventsFixture(entryCount))

		// Test with different limits
		for _, limit := range []int32{0, 10, 33, 110} {
			resp, err := service.GetTimeline(t.Context(), connect.NewRequest(&timelinepb.GetTimelineRequest{
				Query: &timelinepb.TimelineQuery{
					Order: timelinepb.TimelineQuery_ORDER_DESC,
					Limit: limit,
					Filters: []*timelinepb.TimelineQuery_Filter{
						evetTypeFilter(timelinepb.EventType_EVENT_TYPE_CALL),
					},
				},
			}))
			if limit == 0 {
				assert.Error(t, err, "invalid_argument: limit must be > 0")
				continue
			}
			assert.NoError(t, err)
			if limit == 0 || limit > int32(entryCount) {
				assert.Equal(t, entryCount, len(resp.Msg.Events))
			} else {
				assert.Equal(t, int(limit), len(resp.Msg.Events))
			}
		}
	})

	t.Run("Delete old events", func(t *testing.T) {
		t.Parallel()

		service := createTestService(t, timestampFixture(100))

		// Delete half the events (everything older than 3 seconds)
		_, err := service.DeleteOldEvents(t.Context(), connect.NewRequest(&timelinepb.DeleteOldEventsRequest{
			AgeSeconds: 3,
			EventType:  timelinepb.EventType_EVENT_TYPE_UNSPECIFIED,
		}))
		assert.NoError(t, err)
		assert.Equal(t, len(service.events), 150, "expected only half the events to be deleted")
	})
}

func createTestService(t *testing.T, dataFixture *timelinepb.CreateEventsRequest) *service {
	t.Helper()

	service := &service{notifier: channels.NewNotifier(t.Context())}
	service.events = []*timelinepb.Event{}
	_, err := service.CreateEvents(t.Context(), connect.NewRequest(dataFixture))
	assert.NoError(t, err)
	return service
}

func callEventsFixture(entryCount int) *timelinepb.CreateEventsRequest {
	entries := []*timelinepb.CreateEventsRequest_EventEntry{}
	for i := range entryCount {
		entries = append(entries, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamppb.New(time.Now()),
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		})
	}

	return &timelinepb.CreateEventsRequest{Entries: entries}
}

func timestampFixture(entryCount int) *timelinepb.CreateEventsRequest {
	// Create a bunch of entries of different types
	entries := []*timelinepb.CreateEventsRequest_EventEntry{}
	for i := range entryCount {
		var timestamp *timestamppb.Timestamp
		if i < 50 {
			timestamp = timestamppb.New(time.Now().Add(-3 * time.Second))
		} else {
			timestamp = timestamppb.New(time.Now())
		}
		entries = append(entries, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Log{
				Log: &timelinepb.LogEvent{
					Message: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_ChangesetCreated{
				ChangesetCreated: &timelinepb.ChangesetCreatedEvent{
					Key:       strconv.Itoa(i),
					CreatedAt: timestamp,
				},
			},
		})
	}
	return &timelinepb.CreateEventsRequest{Entries: entries}
}

func evetTypeFilter(eventTypes ...timelinepb.EventType) *timelinepb.TimelineQuery_Filter {
	return &timelinepb.TimelineQuery_Filter{
		Filter: &timelinepb.TimelineQuery_Filter_EventTypes{
			EventTypes: &timelinepb.TimelineQuery_EventTypeFilter{
				EventTypes: eventTypes,
			},
		},
	}
}
