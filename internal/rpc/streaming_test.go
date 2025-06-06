package rpc

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/channels"
)

// The Connect gRPC service interface.
type SchemaService interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
	StreamingPing(ctx context.Context, req *connect.Request[ftlv1.PingRequest], stream *connect.ServerStream[ftlv1.PingResponse]) error
}

// The proxy object that implements the gRPC interface over our mockable wrapper.
type SchemaServiceProxy struct {
	*serviceImpl
}

func (s *SchemaServiceProxy) StreamingPing(ctx context.Context, req *connect.Request[ftlv1.PingRequest], stream *connect.ServerStream[ftlv1.PingResponse]) error {
	return s.serviceImpl.StreamingPing(ctx, req, stream)
}

// Implementation of the mockable service.
type serviceImpl struct{}

func (s *serviceImpl) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *serviceImpl) StreamingPing(ctx context.Context, req *connect.Request[ftlv1.PingRequest], stream ServerStream[ftlv1.PingResponse]) error {
	notReady := "not ready"
	if err := stream.Send(&ftlv1.PingResponse{NotReady: &notReady}); err != nil {
		return errors.WithStack(err)
	}
	if err := stream.Send(&ftlv1.PingResponse{}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func TestStreaming(t *testing.T) {
	pings, stream := NewServerStream[ftlv1.PingResponse]()

	impl := &serviceImpl{}

	_, err := impl.Ping(t.Context(), connect.NewRequest(&ftlv1.PingRequest{}))
	assert.NoError(t, err)

	wg := errgroup.Group{}
	wg.Go(func() error {
		return impl.StreamingPing(t.Context(), connect.NewRequest(&ftlv1.PingRequest{}), stream)
	})

	ping := channels.Take(t, time.Second, pings)
	assert.Equal(t, "not ready", ping.GetNotReady())

	ping = channels.Take(t, time.Second, pings)
	assert.Equal(t, "", ping.GetNotReady())

	err = wg.Wait()
	assert.NoError(t, err)
}
