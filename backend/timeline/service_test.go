package timeline

import (
	"iter"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	sops "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/iterops"
)

func BenchmarkTimelineInsert(b *testing.B) {
	service := createTestService(b, &timelinepb.CreateEventsRequest{})
	fixture := callEventsFixture(10_000)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := service.CreateEvents(b.Context(), connect.NewRequest(fixture))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTimelineFilterID(b *testing.B) {
	service := createTestService(b, callEventsFixture(10_000_000))

	req := connect.NewRequest(&timelinepb.GetTimelineRequest{
		Query: &timelinepb.TimelineQuery{
			Order: timelinepb.TimelineQuery_ORDER_DESC,
			Limit: 1000,
			Filters: []*timelinepb.TimelineQuery_Filter{
				&timelinepb.TimelineQuery_Filter{
					Filter: &timelinepb.TimelineQuery_Filter_Id{
						Id: &timelinepb.TimelineQuery_IDFilter{
							HigherThan: proto.Int64(int64(service.nextID) - 1000),
						},
					},
				},
			},
		},
	})
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := service.GetTimeline(b.Context(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

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

func TestStreamTimeline(t *testing.T) {
	t.Parallel()

	t.Run("DESC order", func(t *testing.T) {
		t.Parallel()
		query := &timelinepb.TimelineQuery{Order: timelinepb.TimelineQuery_ORDER_DESC, Limit: 5}

		t.Run("returns an empty batch initially if there are no events", func(t *testing.T) {
			service := createTestService(t, callEventsFixture(0))
			iter, err := service.streamTimelineIter(t.Context(), &timelinepb.StreamTimelineRequest{Query: query})
			assert.NoError(t, err)
			res, _ := iterops.Next(iter)
			assert.Equal(t, []int{}, sops.Map(assertSuccees(t, res).Events, getID))
		})

		service := createTestService(t, callEventsFixture(20))
		iter, err := service.streamTimelineIter(t.Context(), &timelinepb.StreamTimelineRequest{Query: query})
		assert.NoError(t, err)

		t.Run("with the first <Limit> events in the first batch are returned in descending order", func(t *testing.T) {
			res, _ := iterops.Next(iter)
			events := assertSuccees(t, res).Events
			assert.Equal(t, []int{4, 3, 2, 1, 0}, sops.Map(events, getID))
		})
		t.Run("after the first batch, the remaining events in batches are in descending order", func(t *testing.T) {
			_, err = service.CreateEvents(t.Context(), connect.NewRequest(callEventsFixture(100)))
			assert.NoError(t, err)

			events := readEventIDs(t, 2, iter)
			assert.Equal(t, [][]int{
				{9, 8, 7, 6, 5},
				{14, 13, 12, 11, 10},
			}, events)
		})
	})
	t.Run("ASC order", func(t *testing.T) {
		t.Parallel()
		query := &timelinepb.TimelineQuery{Order: timelinepb.TimelineQuery_ORDER_ASC, Limit: 5}

		t.Run("returns an empty batch initially if there are no events", func(t *testing.T) {
			service := createTestService(t, callEventsFixture(0))
			iter, err := service.streamTimelineIter(t.Context(), &timelinepb.StreamTimelineRequest{Query: query})
			assert.NoError(t, err)
			res, _ := iterops.Next(iter)
			assert.Equal(t, []int{}, sops.Map(assertSuccees(t, res).Events, getID))
		})

		service := createTestService(t, callEventsFixture(20))
		iter, err := service.streamTimelineIter(t.Context(), &timelinepb.StreamTimelineRequest{Query: query})
		assert.NoError(t, err)

		t.Run("with the last <Limit> events in the first batch are returned in ascending order", func(t *testing.T) {
			res, _ := iterops.Next(iter)
			events := assertSuccees(t, res).Events
			assert.Equal(t, []int{15, 16, 17, 18, 19}, sops.Map(events, getID))
		})
		t.Run("after the first batch, the next batches are in ascending order", func(t *testing.T) {
			_, err = service.CreateEvents(t.Context(), connect.NewRequest(callEventsFixture(100)))
			assert.NoError(t, err)

			events := readEventIDs(t, 2, iter)
			assert.Equal(t, [][]int{
				{20, 21, 22, 23, 24},
				{25, 26, 27, 28, 29},
			}, events)
		})
	})
}

func TestQueries(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		query   *timelinepb.TimelineQuery
		wantIDs []int
		fixture *timelinepb.CreateEventsRequest
	}

	testCases := []testCase{{
		name: "events in descending order",
		query: &timelinepb.TimelineQuery{
			Limit: 5,
			Order: timelinepb.TimelineQuery_ORDER_DESC,
		},
		fixture: callEventsFixture(50),
		wantIDs: []int{4, 3, 2, 1, 0},
	}, {
		name: "events in ascending order",
		query: &timelinepb.TimelineQuery{
			Limit: 5,
			Order: timelinepb.TimelineQuery_ORDER_ASC,
		},
		fixture: callEventsFixture(50),
		wantIDs: []int{45, 46, 47, 48, 49},
	}, {
		name: "events in ascending order with a lower than id filter",
		query: &timelinepb.TimelineQuery{
			Limit: 10,
			Order: timelinepb.TimelineQuery_ORDER_ASC,
			Filters: []*timelinepb.TimelineQuery_Filter{
				higherIDThanFilter(45),
			},
		},
		fixture: callEventsFixture(50),
		wantIDs: []int{46, 47, 48, 49},
	}, {
		name: "events in descending order with a higher than id filter",
		query: &timelinepb.TimelineQuery{
			Limit: 10,
			Order: timelinepb.TimelineQuery_ORDER_DESC,
			Filters: []*timelinepb.TimelineQuery_Filter{
				lowerIDThanFilter(5),
			},
		},
		fixture: callEventsFixture(50),
		wantIDs: []int{4, 3, 2, 1, 0},
	}, {
		name: "events filtered by ingress type",
		query: &timelinepb.TimelineQuery{
			Limit: 10,
			Order: timelinepb.TimelineQuery_ORDER_DESC,
			Filters: []*timelinepb.TimelineQuery_Filter{
				evetTypeFilter(timelinepb.EventType_EVENT_TYPE_INGRESS),
			},
		},
		fixture: &timelinepb.CreateEventsRequest{Entries: []*timelinepb.CreateEventsRequest_EventEntry{{
			Timestamp: timestamppb.New(time.Now()),
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Ingress{
				Ingress: &timelinepb.IngressEvent{},
			},
		}, {
			Timestamp: timestamppb.New(time.Now()),
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Call{
				Call: &timelinepb.CallEvent{
					Request:  "request",
					Response: "response",
				},
			},
		}}},
		wantIDs: []int{0},
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := createTestService(t, tc.fixture)
			iter, err := service.streamTimelineIter(t.Context(), &timelinepb.StreamTimelineRequest{Query: tc.query})
			assert.NoError(t, err)
			events := readEventIDs(t, 1, iter)
			assert.Equal(t, tc.wantIDs, events[0])
		})
	}
}

func readEventIDs(t *testing.T, n int, iter iter.Seq[result.Result[*timelinepb.StreamTimelineResponse]]) [][]int {
	var eventsIDs [][]int
	for res := range iter {
		events := assertSuccees(t, res).Events
		eventsIDs = append(eventsIDs, sops.Map(events, getID))

		if len(eventsIDs) >= n {
			break
		}
	}
	return eventsIDs
}

func createTestService(t testing.TB, dataFixture *timelinepb.CreateEventsRequest) *Service {
	t.Helper()

	service, err := newService(t.Context(), Config{})
	assert.NoError(t, err)
	_, err = service.CreateEvents(t.Context(), connect.NewRequest(dataFixture))
	assert.NoError(t, err)
	return service
}

func assertSuccees[T any](t *testing.T, r result.Result[T]) T {
	v, err := r.Result()
	assert.NoError(t, err)
	return v
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

func higherIDThanFilter(id int64) *timelinepb.TimelineQuery_Filter {
	return &timelinepb.TimelineQuery_Filter{
		Filter: &timelinepb.TimelineQuery_Filter_Id{
			Id: &timelinepb.TimelineQuery_IDFilter{
				HigherThan: &id,
			},
		},
	}
}

func lowerIDThanFilter(id int64) *timelinepb.TimelineQuery_Filter {
	return &timelinepb.TimelineQuery_Filter{
		Filter: &timelinepb.TimelineQuery_Filter_Id{
			Id: &timelinepb.TimelineQuery_IDFilter{
				LowerThan: &id,
			},
		},
	}
}

func getID(event *timelinepb.Event) int {
	return int(event.Id)
}
