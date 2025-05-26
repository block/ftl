package test

import (
	"context"

	"connectrpc.com/connect"

	testpb "github.com/block/ftl/backend/protos/xyz/block/ftl/test/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/test/v1/testpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/rpc"
)

var _ testpbconnect.TestServiceHandler = (*Service)(nil)

func NewTestService(realm string, updated chan dev.LocalEndpoint) *Service {
	return &Service{realm: realm, devUpdates: updated}
}

type Service struct {
	devUpdates chan dev.LocalEndpoint
	realm      string
}

// Ping implements testpbconnect.TestServiceHandler.
func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// StartTestRun implements testpbconnect.TestServiceHandler.
func (s *Service) StartTestRun(ctx context.Context, req *connect.Request[testpb.StartTestRunRequest]) (*connect.Response[testpb.StartTestRunResponse], error) {
	s.devUpdates <- dev.LocalEndpoint{Module: req.Msg.ModuleName, Endpoint: req.Msg.Endpoint, HotReloadEndpoint: req.Msg.HotReloadEndpoint}
	dk := key.NewDeploymentKey(s.realm, req.Msg.ModuleName)
	return connect.NewResponse(&testpb.StartTestRunResponse{DeploymentKey: dk.String()}), nil

}

func (s *Service) StartServices(ctx context.Context) ([]rpc.Option, error) {
	return []rpc.Option{rpc.GRPC(testpbconnect.NewTestServiceHandler, s)}, nil
}
