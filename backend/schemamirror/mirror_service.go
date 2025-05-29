package schemamirror

import (
	"context"
	"sync/atomic"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type MirrorService struct {
	receiving   *atomic.Bool
	eventSource *schemaeventsource.EventSource
}

func NewMirrorService() *MirrorService {
	return &MirrorService{
		receiving:   &atomic.Bool{},
		eventSource: schemaeventsource.NewUnattached(),
	}
}

var _ ftlv1connect.SchemaMirrorServiceHandler = (*MirrorService)(nil)

func (s *MirrorService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *MirrorService) PushSchema(ctx context.Context, stream *connect.ClientStream[ftlv1.PushSchemaRequest]) (*connect.Response[ftlv1.PushSchemaResponse], error) {
	logger := log.FromContext(ctx)
	if !s.receiving.CompareAndSwap(false, true) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("mirror is already receiving schema updates"))
	}
	defer s.receiving.Store(false)

	logger.Debugf("Started receiving schema stream pushes")
	for stream.Receive() {
		req := stream.Msg()
		notification, err := schema.NotificationFromProto(req.Event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert notification from proto")
		}
		err = s.eventSource.Publish(notification)
		if err != nil {
			return nil, errors.Wrap(err, "failed to publish schema notification")
		}
	}
	if err := stream.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to receive schema push stream")
	}
	return connect.NewResponse(&ftlv1.PushSchemaResponse{}), nil
}
