package buildengine

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"connectrpc.com/connect"

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

type updatesService struct {
	engine *Engine
}

var _ enginepbconnect.BuildEngineServiceHandler = &updatesService{}

func (e *Engine) startUpdatesService(ctx context.Context, endpoint *url.URL) error {
	logger := log.FromContext(ctx).Scope("build:updates")

	svc := &updatesService{
		engine: e,
	}

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
	events := make(chan *buildenginepb.EngineEvent, 64)
	u.engine.EngineUpdates.Subscribe(events)
	defer u.engine.EngineUpdates.Unsubscribe(events)

	for event := range channels.IterContext(ctx, events) {
		err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
			Event: event,
		})
		if err != nil {
			return fmt.Errorf("failed to send engine event: %w", err)
		}
	}

	return nil
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
				Level: langpb.Error_ErrorLevel(berr.Level),
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
