package buildengine

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	enginepbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	ftlerrors "github.com/block/ftl/common/errors"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
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
}

var _ enginepbconnect.BuildEngineServiceHandler = &updatesService{}

func (e *Engine) startUpdatesService(ctx context.Context, endpoint *url.URL) error {
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
	svc.engine.EngineUpdates.Subscribe(events)

	// Start goroutine to collect events
	go func() {
		for event := range channels.IterContext(ctx, events) {
			svc.lock.Lock()
			svc.events = append(svc.events, event)
			svc.lock.Unlock()
		}
	}()

	// Start cache cleanup goroutine
	go svc.cleanupCache(ctx)

	logger.Debugf("Build updates service listening on: %s", endpoint)
	err := rpc.Serve(ctx, endpoint,
		rpc.GRPC(enginepbconnect.NewBuildEngineServiceHandler, svc),
	)

	if err != nil {
		return fmt.Errorf("build updates service stopped serving: %w", err)
	}
	return nil
}

func (u *updatesService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (u *updatesService) StreamEngineEvents(ctx context.Context, req *connect.Request[buildenginepb.StreamEngineEventsRequest], stream *connect.ServerStream[buildenginepb.StreamEngineEventsResponse]) error {
	// Only send cached events if replay_history is true
	if req.Msg.ReplayHistory {
		u.lock.RLock()
		for _, event := range u.events {
			err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
				Event: event,
			})
			if err != nil {
				u.lock.RUnlock()
				return fmt.Errorf("failed to send cached engine event: %w", err)
			}
		}
		u.lock.RUnlock()
	}

	// Then subscribe to new events
	events := make(chan *buildenginepb.EngineEvent, 64)
	u.engine.EngineUpdates.Subscribe(events)
	defer u.engine.EngineUpdates.Unsubscribe(events)

	for event := range channels.IterContext(ctx, events) {
		// Add timestamp to event
		if event.Timestamp == nil {
			event.Timestamp = timestamppb.Now()
		}

		// Cache the event
		u.lock.Lock()
		u.events = append(u.events, event)
		u.lock.Unlock()

		err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
			Event: event,
		})
		if err != nil {
			return fmt.Errorf("failed to send engine event: %w", err)
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
	for _, e := range ftlerrors.DeduplicateErrors(ftlerrors.UnwrapAll(err)) {
		if !ftlerrors.Innermost(e) {
			continue
		}

		var berr builderrors.Error
		if !errors.As(e, &berr) {
			// If not a builderrors.Error, create a basic error
			errs = append(errs, &langpb.Error{
				Msg:   e.Error(),
				Level: langpb.Error_ERROR_LEVEL_ERROR,
			})
			continue
		}

		pbError := &langpb.Error{
			Msg:   berr.Msg,
			Level: langpb.Error_ErrorLevel(berr.Level),
		}

		// Add position information if available
		if pos, ok := berr.Pos.Get(); ok {
			pbError.Pos = &langpb.Position{
				Filename:    pos.Filename,
				Line:        int64(pos.Line),
				StartColumn: int64(pos.StartColumn),
				EndColumn:   int64(pos.EndColumn),
			}
		}

		errs = append(errs, pbError)
	}

	return errs
}
