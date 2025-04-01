package main

import (
	"context"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/internal/log"
)

func main() {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	plugin.Start(ctx, "test-client", newServer, ftlv1connect.VerbServiceName, ftlv1connect.NewVerbServiceHandler)
}

type verbService struct{}

var _ ftlv1connect.VerbServiceHandler = (*verbService)(nil)

func (verbService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
func (verbService) Call(context.Context, *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	return connect.NewResponse(&ftlv1.CallResponse{}), nil
}

func newServer(ctx context.Context, config any) (context.Context, *verbService, error) {
	return ctx, &verbService{}, nil
}
