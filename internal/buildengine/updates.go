package buildengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	enginepbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	ftlerrors "github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/rpc"
)

type Config struct {
	EventCacheRetention *time.Duration `help:"Delete cached events after this time period. 0 to disable" env:"FTL_EVENT_CACHE_RETENTION" default:"24h"`
}

type updatesService struct {
	engine *Engine
	config Config
	lock   sync.RWMutex
	events []*buildenginepb.EngineEvent
	// Channel for subscribers to receive new events
	subscribers sync.Map
}

var _ enginepbconnect.BuildEngineServiceHandler = &updatesService{}

func (e *Engine) startUpdatesService(ctx context.Context) rpc.Service {
	logger := log.FromContext(ctx).Scope("build:updates")

	svc := &updatesService{
		engine: e,
		config: Config{
			EventCacheRetention: &[]time.Duration{24 * time.Hour}[0],
		},
		events: make([]*buildenginepb.EngineEvent, 0),
	}

	// Subscribe to engine updates from the start
	events := make(chan *buildenginepb.EngineEvent, 128)
	svc.engine.engineUpdates.Subscribe(events)

	// Start goroutine to collect events
	go func() {
		defer svc.engine.engineUpdates.Unsubscribe(events)
		for event := range channels.IterContext(ctx, events) {
			// Add timestamp to event if not present
			if event.Timestamp == nil {
				event.Timestamp = timestamppb.Now()
			}

			svc.lock.Lock()
			svc.events = append(svc.events, event)
			svc.lock.Unlock()

			// Broadcast to all subscribers
			svc.subscribers.Range(func(key, value interface{}) bool {
				if ch, ok := value.(chan *buildenginepb.EngineEvent); ok {
					select {
					case ch <- event:
					default:
						// If channel is full, skip the event for this subscriber
						logger.Warnf("Subscriber channel is full, dropping event")
					}
				}
				return true
			})
		}
	}()

	// Start cache cleanup goroutine
	go svc.cleanupCache(ctx)
	return svc
}
func (u *updatesService) StartServices(ctx context.Context) ([]rpc.Option, error) {
	return []rpc.Option{rpc.GRPC(enginepbconnect.NewBuildEngineServiceHandler, u)}, nil
}

func (u *updatesService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (u *updatesService) StreamEngineEvents(ctx context.Context, req *connect.Request[buildenginepb.StreamEngineEventsRequest], stream *connect.ServerStream[buildenginepb.StreamEngineEventsResponse]) error {
	events := make(chan *buildenginepb.EngineEvent, 64)
	subscriberKey := fmt.Sprintf("subscriber-%p", events)
	u.subscribers.Store(subscriberKey, events)
	defer u.subscribers.Delete(subscriberKey)

	// Always send atleast one event so clients can start the stream (the first event contains the headers)
	err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
		Event: &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_EngineEnded{
				EngineEnded: &buildenginepb.EngineEnded{},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to send engine event")
	}
	// Send cached events if replay_history is true
	if req.Msg.ReplayHistory {
		u.lock.RLock()
		for _, event := range u.events {
			err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
				Event: event,
			})
			if err != nil {
				u.lock.RUnlock()
				return errors.Wrap(err, "failed to send cached engine event")
			}
		}
		u.lock.RUnlock()

		err = stream.Send(&buildenginepb.StreamEngineEventsResponse{
			Event: &buildenginepb.EngineEvent{
				Timestamp: timestamppb.Now(),
				Event: &buildenginepb.EngineEvent_ReachedEndOfHistory{
					ReachedEndOfHistory: &buildenginepb.ReachedEndOfHistory{},
				},
			},
		})
		if err != nil {
			return errors.Wrap(err, "failed to send event to indicate the end of historical events")
		}
	}

	// Process new events
	for event := range channels.IterContext(ctx, events) {
		err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
			Event: event,
		})
		if err != nil {
			return errors.Wrap(err, "failed to send engine event")
		}
	}

	return nil
}

func (u *updatesService) cleanupCache(ctx context.Context) {
	logger := log.FromContext(ctx)
	if u.config.EventCacheRetention == nil || *u.config.EventCacheRetention == 0 {
		logger.Debugf("Event cache retention is disabled, will not clean up cache")
		return
	}

	ticker := time.NewTicker(*u.config.EventCacheRetention / 20)
	defer ticker.Stop()

	for range channels.IterContext(ctx, ticker.C) {
		u.lock.Lock()
		cutoff := time.Now().Add(-*u.config.EventCacheRetention)

		// Filter out old events
		filtered := make([]*buildenginepb.EngineEvent, 0, len(u.events))
		for _, event := range u.events {
			// Keep events that are newer than the cutoff
			if event.Timestamp != nil && event.Timestamp.AsTime().After(cutoff) {
				filtered = append(filtered, event)
			}
		}

		removed := len(u.events) - len(filtered)
		if removed > 0 {
			logger.Debugf("Cleaned up %d cached events older than %s", removed, u.config.EventCacheRetention)
		}

		u.events = filtered
		u.lock.Unlock()
	}
}

// errorToLangError converts a single error to a slice of protobuf Error messages
func errorToLangError(err error) []*langpb.Error {
	errs := []*langpb.Error{}

	// Unwrap and deduplicate all nested errors
	for _, e := range ftlerrors.Deduplicate(errors.UnwrapAllInnermost(err)) {
		if !errors.Innermost(e) {
			continue
		}

		var berr builderrors.Error
		if !errors.As(e, &berr) {
			// If not a builderrors.Error, create a basic error
			errs = append(errs, &langpb.Error{
				Msg:   e.Error(),
				Level: langpb.Error_ERROR_LEVEL_ERROR,
				Type:  langpb.Error_ERROR_TYPE_UNSPECIFIED,
			})
			continue
		}

		errs = append(errs, langpb.ErrorToProto(berr))
	}

	return errs
}
