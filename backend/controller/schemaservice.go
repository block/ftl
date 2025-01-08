package controller

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/block/ftl/backend/controller/state"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/model"
)

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	view, err := s.controllerState.Query(ctx, struct{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get controller state: %w", err)
	}
	schemas := view.GetActiveDeploymentSchemas()
	modules := []*schemapb.Module{
		schema.Builtins().ToProto(),
	}
	modules = append(modules, slices.Map(schemas, func(d *schema.Module) *schemapb.Module { return d.ToProto() })...)
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: &schemapb.Schema{Modules: modules}}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return s.watchModuleChanges(ctx, func(response *ftlv1.PullSchemaResponse) error {
		return stream.Send(response)
	})
}

func (s *Service) UpdateDeploymentRuntime(ctx context.Context, req *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error) {
	deployment, err := model.ParseDeploymentKey(req.Msg.Deployment)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	view, err := s.controllerState.Query(ctx, struct{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get controller state: %w", err)
	}
	dep, err := view.GetDeployment(deployment)
	if err != nil {
		return nil, fmt.Errorf("could not get schema: %w", err)
	}
	module := dep.Schema
	if module.Runtime == nil {
		module.Runtime = &schema.ModuleRuntime{}
	}
	event := schema.ModuleRuntimeEventFromProto(req.Msg.Event)
	module.Runtime.ApplyEvent(event)
	err = s.controllerState.Update(ctx, &state.DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
	if err != nil {
		return nil, fmt.Errorf("could not update schema for module %s: %w", module.Name, err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}
