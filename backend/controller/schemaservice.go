package controller

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/backend/controller/state"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
)

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	view, err := s.controllerState.View(ctx)
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
	view, err := s.controllerState.View(ctx)
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
	err = s.controllerState.Publish(ctx, &state.DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
	if err != nil {
		return nil, fmt.Errorf("could not update schema for module %s: %w", module.Name, err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) Watch(ctx context.Context, req *connect.Request[ftlv1.WatchRequest], stream *connect.ServerStream[ftlv1.WatchResponse]) error {
	view, err := s.controllerState.View(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller state: %w", err)
	}
	// Send the initial schema as the first message
	err = stream.Send(&ftlv1.WatchResponse{
		Schema: &schemapb.Schema{
			Modules: slices.Map(view.GetActiveDeploymentSchemas(), func(m *schema.Module) *schemapb.Module { return m.ToProto() }),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send initial schema: %w", err)
	}

	return s.watchModuleChanges(ctx, func(response *ftlv1.PullSchemaResponse) error {
		return stream.Send(&ftlv1.WatchResponse{
			Schema: &schemapb.Schema{
				Modules: slices.Map(view.GetActiveDeploymentSchemas(), func(m *schema.Module) *schemapb.Module { return m.ToProto() }),
			},
		})
	})
}

func (s *Service) watchModuleChanges(ctx context.Context, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx)

	updates := s.controllerState.Updates().Subscribe(nil)
	defer s.controllerState.Updates().Unsubscribe(updates)
	view, err := s.controllerState.View(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller state: %w", err)
	}

	// Seed the notification channel with the current deployments.
	seedDeployments := view.GetActiveDeployments()
	initialCount := len(seedDeployments)

	builtins := schema.Builtins().ToProto()
	builtinsResponse := &ftlv1.PullSchemaResponse{
		ModuleName: builtins.Name,
		Schema:     builtins,
		ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
		More:       initialCount > 0,
	}

	err = sendChange(builtinsResponse)
	if err != nil {
		return err
	}
	for _, initial := range seedDeployments {
		initialCount--
		module := initial.Schema.ToProto()
		err := sendChange(&ftlv1.PullSchemaResponse{
			ModuleName:    module.Name,
			DeploymentKey: proto.String(initial.Key.String()),
			Schema:        module,
			ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
			More:          initialCount > 0,
		})
		if err != nil {
			return err
		}
	}
	logger.Tracef("Seeded %d deployments", initialCount)

	for notification := range channels.IterContext(ctx, updates) {
		switch event := notification.(type) {
		case *state.DeploymentCreatedEvent:
			err := sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    event.Module,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        event.Schema.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
			})
			if err != nil {
				return err
			}
		case *state.DeploymentDeactivatedEvent:
			view, err := s.controllerState.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get controller state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    dep.Module,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        dep.Schema.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED,
				ModuleRemoved: event.ModuleRemoved,
			})
			if err != nil {
				return err
			}
		case *state.DeploymentSchemaUpdatedEvent:
			view, err := s.controllerState.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get controller state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    dep.Module,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        event.Schema.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
