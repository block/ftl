package schemaservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/statemachine"
)

type Service struct {
	State *statemachine.SingleQueryHandle[struct{}, SchemaState, SchemaEvent]
}

var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	view, err := s.State.View(ctx)
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
	deployment, err := key.ParseDeploymentKey(req.Msg.Deployment)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller state: %w", err)
	}
	module, err := view.GetDeployment(deployment)
	if err != nil {
		return nil, fmt.Errorf("could not get schema: %w", err)
	}
	if module.Runtime == nil {
		module.Runtime = &schema.ModuleRuntime{}
	}
	event, err := schema.ModuleRuntimeEventFromProto(req.Msg.Event)
	if err != nil {
		return nil, fmt.Errorf("could not parse event: %w", err)
	}
	module.Runtime.ApplyEvent(event)
	err = s.State.Publish(ctx, &DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
	if err != nil {
		return nil, fmt.Errorf("could not update schema for module %s: %w", module.Name, err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) watchModuleChanges(ctx context.Context, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx)

	uctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stateIter, err := s.State.StateIter(uctx)
	if err != nil {
		return fmt.Errorf("failed to get schema state iterator: %w", err)
	}

	view, err := s.State.View(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema state: %w", err)
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
	for key, initial := range seedDeployments {
		initialCount--
		module := initial.ToProto()
		err := sendChange(&ftlv1.PullSchemaResponse{
			ModuleName:    module.Name,
			DeploymentKey: proto.String(key.String()),
			Schema:        module,
			ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
			More:          initialCount > 0,
		})
		if err != nil {
			return err
		}
	}
	logger.Tracef("Seeded %d deployments", initialCount)

	for notification := range iterops.Changes(stateIter, EventExtractor) {
		switch event := notification.(type) {
		case *DeploymentCreatedEvent:
			err := sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    event.Schema.Name,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        event.Schema.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
			})
			if err != nil {
				return err
			}
		case *DeploymentDeactivatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    dep.Name,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        dep.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED,
				ModuleRemoved: event.ModuleRemoved,
			})
			if err != nil {
				return err
			}
		case *DeploymentSchemaUpdatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				ModuleName:    dep.Name,
				DeploymentKey: proto.String(event.Key.String()),
				Schema:        dep.ToProto(),
				ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
