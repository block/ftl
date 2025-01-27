package schemaservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"github.com/alecthomas/types/optional"
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
	schemas := view.GetCanonicalDeploymentSchemas()
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
	var changeset optional.Option[key.Changeset]
	if req.Msg.GetChangeset() != "" {
		c, err := key.ParseChangesetKey(req.Msg.GetChangeset())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
		}
		changeset = optional.Some(c)
	}
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller state: %w", err)
	}
	module, err := view.GetDeployment(deployment, changeset)
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

// CreateChangeset creates a new changeset.
func (s *Service) CreateChangeset(ctx context.Context, req *connect.Request[ftlv1.CreateChangesetRequest]) (*connect.Response[ftlv1.CreateChangesetResponse], error) {
	modules, err := slices.MapErr(req.Msg.Modules, func(m *schemapb.Module) (*schema.Module, error) {
		out, err := schema.ModuleFromProto(m)
		if err != nil {
			return nil, fmt.Errorf("invalid module %s: %w", m.Name, err)
		}
		return out, nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	changeset := &schema.Changeset{
		Key:     key.NewChangesetKey(),
		State:   schema.ChangesetStateProvisioning,
		Modules: modules,
	}

	// TODO: validate changeset schema with canonical schema

	s.State.Publish(ctx, &ChangesetCreatedEvent{
		Changeset: changeset,
	})
	return connect.NewResponse(&ftlv1.CreateChangesetResponse{}), nil
}

// CommitChangeset makes all deployments for the changeset part of the canonical schema.
func (s *Service) CommitChangeset(context.Context, *connect.Request[ftlv1.CommitChangesetRequest]) (*connect.Response[ftlv1.CommitChangesetResponse], error) {
	return connect.NewResponse(&ftlv1.CommitChangesetResponse{}), nil
}

// FailChangeset fails an active changeset.
func (s *Service) FailChangeset(context.Context, *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	return connect.NewResponse(&ftlv1.FailChangesetResponse{}), nil
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

	// TODO: update this to send changesets as well

	// Seed the notification channel with the current deployments.
	seedDeployments := view.GetCanonicalDeployments()
	initialCount := len(seedDeployments)

	builtins := schema.Builtins().ToProto()
	builtinsResponse := &ftlv1.PullSchemaResponse{
		Event: &ftlv1.PullSchemaResponse_DeploymentCreated_{
			DeploymentCreated: &ftlv1.PullSchemaResponse_DeploymentCreated{
				Schema: builtins,
			},
		},
		More: initialCount > 0,
	}
	err = sendChange(builtinsResponse)
	if err != nil {
		return err
	}
	for _, initial := range seedDeployments {
		initialCount--
		module := initial.ToProto()
		err := sendChange(&ftlv1.PullSchemaResponse{
			Event: &ftlv1.PullSchemaResponse_DeploymentCreated_{
				DeploymentCreated: &ftlv1.PullSchemaResponse_DeploymentCreated{
					Schema: module,
				},
			},
			More: initialCount > 0,
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
				Event: &ftlv1.PullSchemaResponse_DeploymentCreated_{
					DeploymentCreated: &ftlv1.PullSchemaResponse_DeploymentCreated{
						Schema: event.Schema.ToProto(),
					},
				},
			})
			if err != nil {
				return err
			}
		case *DeploymentDeactivatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key, event.Changeset)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				Event: &ftlv1.PullSchemaResponse_DeploymentRemoved_{
					DeploymentRemoved: &ftlv1.PullSchemaResponse_DeploymentRemoved{
						Key:        proto.String(event.Key.String()),
						ModuleName: dep.Name,
						// If this is true then the module was removed as well as the deployment.
						ModuleRemoved: event.ModuleRemoved,
					},
				},
			})
			if err != nil {
				return err
			}
		case *DeploymentSchemaUpdatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key, event.Changeset)
			if err != nil {
				logger.Errorf(err, "Deployment not found: %s", event.Key)
				continue
			}
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				Event: &ftlv1.PullSchemaResponse_DeploymentUpdated_{
					DeploymentUpdated: &ftlv1.PullSchemaResponse_DeploymentUpdated{
						// TODO: include changeset info
						Changeset: nil,
						Schema:    dep.ToProto(),
					},
				},
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
