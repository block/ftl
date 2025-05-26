package timeline

import (
	"context"
	"iter"
	"sort"
	"sync"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	timelineconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/rpc"
)

type Config struct {
	EventLogRetention *time.Duration `help:"Delete call logs after this time period. 0 to disable" env:"FTL_EVENT_LOG_RETENTION" default:"24h"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type Service struct {
	config Config
	lock   sync.RWMutex
	nextID int
	events []*timelinepb.Event

	notifier *channels.Notifier
}

var _ timelineconnect.TimelineServiceHandler = (*Service)(nil)
var _ rpc.Service = (*Service)(nil)

func New(ctx context.Context, config Config) (*Service, error) {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("timeline")
	ctx = log.ContextWithLogger(ctx, logger)
	svc, err := newService(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create timeline service")
	}

	return svc, nil
}
func (s *Service) StartServices(ctx context.Context) ([]rpc.Option, error) {
	go s.reapCallEvents(ctx)
	return []rpc.Option{rpc.GRPC(timelineconnect.NewTimelineServiceHandler, s)}, nil
}

func newService(ctx context.Context, config Config) (*Service, error) {
	return &Service{
		config:   config,
		nextID:   0,
		notifier: channels.NewNotifier(ctx),
		events:   make([]*timelinepb.Event, 0),
	}, nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) CreateEvents(ctx context.Context, req *connect.Request[timelinepb.CreateEventsRequest]) (*connect.Response[timelinepb.CreateEventsResponse], error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	entries := make([]*timelinepb.CreateEventsRequest_EventEntry, 0, len(req.Msg.Entries))
	entries = append(entries, req.Msg.Entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.AsTime().Before(entries[j].Timestamp.AsTime())
	})

	for _, entry := range entries {
		if entry.Timestamp == nil {
			return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.New("timestamp is required")))
		}
		event := &timelinepb.Event{
			Id:        int64(s.nextID),
			Timestamp: entry.Timestamp,
		}
		switch entry := entry.Entry.(type) {
		case *timelinepb.CreateEventsRequest_EventEntry_Log:
			event.Entry = &timelinepb.Event_Log{
				Log: entry.Log,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_Call:
			event.Entry = &timelinepb.Event_Call{
				Call: entry.Call,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_Ingress:
			event.Entry = &timelinepb.Event_Ingress{
				Ingress: entry.Ingress,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_CronScheduled:
			event.Entry = &timelinepb.Event_CronScheduled{
				CronScheduled: entry.CronScheduled,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_PubsubPublish:
			event.Entry = &timelinepb.Event_PubsubPublish{
				PubsubPublish: entry.PubsubPublish,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_PubsubConsume:
			event.Entry = &timelinepb.Event_PubsubConsume{
				PubsubConsume: entry.PubsubConsume,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_ChangesetCreated:
			event.Entry = &timelinepb.Event_ChangesetCreated{
				ChangesetCreated: entry.ChangesetCreated,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_ChangesetStateChanged:
			event.Entry = &timelinepb.Event_ChangesetStateChanged{
				ChangesetStateChanged: entry.ChangesetStateChanged,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_DeploymentRuntime:
			event.Entry = &timelinepb.Event_DeploymentRuntime{
				DeploymentRuntime: entry.DeploymentRuntime,
			}
		}
		s.events = append(s.events, event)
		s.nextID++
	}
	s.notifier.Notify(ctx)
	return connect.NewResponse(&timelinepb.CreateEventsResponse{}), nil
}

func (s *Service) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ascending := isAscending(req.Msg.Query)
	fctx := filtersFromQuery(req.Msg.Query)
	if req.Msg.Query.Limit == 0 {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0")))
	}
	// Get 1 more than the requested limit to determine if there are more results.
	limit := int(req.Msg.Query.Limit)
	fetchLimit := limit + 1

	results := []*timelinepb.Event{}

	var firstIdx, step int
	var idCheck func(int64) bool

	if ascending {
		firstIdx = s.findIndexWithLargerID(fctx.higherThan)
		step = 1
		idCheck = func(i int64) bool { return i <= fctx.lowerThan }
	} else {
		firstIdx = s.findIndexWithLargerID(fctx.lowerThan) - 1
		step = -1
		idCheck = func(i int64) bool { return i >= fctx.higherThan }
	}

	for i := firstIdx; i >= 0 && i < len(s.events) && idCheck(s.events[i].Id); i += step {
		event := s.events[i]
		_, didNotMatchAFilter := slices.Find(fctx.filters, func(filter TimelineFilter) bool {
			return !filter(event)
		})
		if didNotMatchAFilter {
			continue
		}
		results = append(results, s.events[i])
		if fetchLimit != 0 && len(results) >= fetchLimit {
			break
		}
	}

	var cursor *int64
	// Return only the requested number of results.
	if len(results) > limit {
		id := results[len(results)-1].Id
		results = results[:limit]
		cursor = &id
	}
	return connect.NewResponse(&timelinepb.GetTimelineResponse{
		Events: results,
		Cursor: cursor,
	}), nil
}

func (s *Service) findIndexWithLargerID(id int64) int {
	idx := sort.Search(len(s.events), func(i int) bool {
		return s.events[i].Id >= id
	})
	return idx
}

// We want to throttle the number of updates we send to the client via the stream.
const minUpdateInterval = 50 * time.Millisecond

func (s *Service) streamTimelineIter(ctx context.Context, req *timelinepb.StreamTimelineRequest) (iter.Seq[result.Result[*timelinepb.StreamTimelineResponse]], error) {
	sub := s.notifier.Subscribe(ctx)
	lastUpdate := time.Now()
	query := req.Query

	reverseOrder := timelinepb.TimelineQuery_ORDER_DESC
	if query.Order == reverseOrder {
		reverseOrder = timelinepb.TimelineQuery_ORDER_ASC
	}

	updateInterval := minUpdateInterval
	if req.UpdateInterval != nil && req.UpdateInterval.AsDuration() > minUpdateInterval {
		updateInterval = req.UpdateInterval.AsDuration()
	}

	if query.Limit == 0 {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0")))
	}

	first := true
	var events []*timelinepb.Event
	var lastEventID optional.Option[int64]

	return func(yield func(result.Result[*timelinepb.StreamTimelineResponse]) bool) {
		if first {
			// we are returning the first batch
			first = false
			// fetch the initial batch of events, up to the limit
			resp, err := s.GetTimeline(ctx, connect.NewRequest(&timelinepb.GetTimelineRequest{Query: &timelinepb.TimelineQuery{
				Order:   reverseOrder,
				Limit:   query.Limit,
				Filters: query.Filters,
			}}))
			if err != nil {
				yield(result.Err[*timelinepb.StreamTimelineResponse](errors.Wrap(err, "failed to get timeline")))
				return
			}
			events = resp.Msg.Events
			slices.Reverse(events)
			lastEventID = updatedMaxEventID(events, optional.None[int64]())

			if !yield(result.Ok(&timelinepb.StreamTimelineResponse{Events: events})) {
				return
			}
		}

		for {
			newQuery := &timelinepb.TimelineQuery{
				Order: timelinepb.TimelineQuery_ORDER_ASC,
				Limit: query.Limit,
				Filters: append(query.Filters, &timelinepb.TimelineQuery_Filter{
					Filter: &timelinepb.TimelineQuery_Filter_Id{
						Id: &timelinepb.TimelineQuery_IDFilter{
							HigherThan: lastEventID.Ptr(),
						},
					},
				}),
			}

			resp, err := s.GetTimeline(ctx, connect.NewRequest(&timelinepb.GetTimelineRequest{Query: newQuery}))
			if err != nil {
				yield(result.Err[*timelinepb.StreamTimelineResponse](errors.Wrap(err, "failed to get timeline")))
				return
			}

			events = resp.Msg.Events
			if query.Order == timelinepb.TimelineQuery_ORDER_DESC {
				slices.Reverse(events)
			}

			if len(events) > 0 {
				lastEventID = updatedMaxEventID(events, lastEventID)
				if !yield(result.Ok(&timelinepb.StreamTimelineResponse{Events: events})) {
					return
				}
			} else {
				// no more events to send, wait for a new events or the context to be done
				select {
				case <-sub:
				case <-ctx.Done():
					return
				}
			}
			// throttle the updates to the client
			time.Sleep(time.Until(lastUpdate.Add(updateInterval)))
			lastUpdate = time.Now()
		}
	}, nil
}

func updatedMaxEventID(events []*timelinepb.Event, prevMaxEventID optional.Option[int64]) optional.Option[int64] {
	if len(events) == 0 {
		return prevMaxEventID
	}
	if events[len(events)-1].Id > events[0].Id {
		return optional.Some(events[len(events)-1].Id)
	}
	return optional.Some(events[0].Id)
}

func (s *Service) StreamTimeline(ctx context.Context, req *connect.Request[timelinepb.StreamTimelineRequest], stream *connect.ServerStream[timelinepb.StreamTimelineResponse]) error {
	iter, err := s.streamTimelineIter(ctx, req.Msg)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := rpc.IterAsGrpc(iter, stream); err != nil {
		return errors.Wrap(err, "failed to stream timeline")
	}
	return nil
}

func (s *Service) DeleteOldEvents(ctx context.Context, req *connect.Request[timelinepb.DeleteOldEventsRequest]) (*connect.Response[timelinepb.DeleteOldEventsResponse], error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Events that match all these filters will be deleted
	cutoff := time.Now().Add(-1 * time.Duration(req.Msg.AgeSeconds) * time.Second)
	deletionFilters := []TimelineFilter{
		FilterTypes(&timelinepb.TimelineQuery_EventTypeFilter{
			EventTypes: []timelinepb.EventType{req.Msg.EventType},
		}),
		FilterTimeRange(&timelinepb.TimelineQuery_TimeFilter{
			OlderThan: timestamppb.New(cutoff),
		}),
	}

	filtered := []*timelinepb.Event{}
	deleted := int64(0)
	for _, event := range s.events {
		_, didNotMatchAFilter := slices.Find(deletionFilters, func(filter TimelineFilter) bool {
			return !filter(event)
		})
		if didNotMatchAFilter {
			filtered = append(filtered, event)
		} else {
			deleted++
		}
	}
	s.events = filtered
	return connect.NewResponse(&timelinepb.DeleteOldEventsResponse{
		DeletedCount: deleted,
	}), nil
}

func (s *Service) reapCallEvents(ctx context.Context) {
	logger := log.FromContext(ctx)
	var interval time.Duration
	if s.config.EventLogRetention == nil {
		interval = time.Hour
	} else {
		interval = *s.config.EventLogRetention / 20
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range channels.IterContext(ctx, ticker.C) {
		if s.config.EventLogRetention == nil {
			logger.Tracef("Event log retention is disabled, will not prune.")
			continue
		}

		resp, err := s.DeleteOldEvents(ctx, connect.NewRequest(&timelinepb.DeleteOldEventsRequest{
			EventType:  timelinepb.EventType_EVENT_TYPE_CALL,
			AgeSeconds: int64(s.config.EventLogRetention.Seconds()),
		}))
		if err != nil {
			logger.Errorf(err, "Failed to prune call events")
			continue
		}
		if resp.Msg.DeletedCount > 0 {
			logger.Debugf("Pruned %d call events older than %s", resp.Msg.DeletedCount, s.config.EventLogRetention)
		}
	}
}
