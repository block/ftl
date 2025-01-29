package schemaservice

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"
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
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/statemachine"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8897" env:"FTL_BIND"`
}

type Service struct {
	State *statemachine.SingleQueryHandle[struct{}, SchemaState, schema.Event]
}

var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

func New(ctx context.Context) *Service {
	return &Service{State: NewInMemorySchemaState(ctx)}
}

// Start the SchemaService. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL schema service")

	svc := New(ctx)
	logger.Debugf("Listening on %s", config.Bind)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc),
			rpc.PProf(),
		)
	})

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("failed to start schema service: %w", err)
	}
	return nil
}
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
	deployment, err := key.ParseDeploymentKey(*req.Msg.Event.DeploymentKey)
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
	_, err = view.ApplyEvent(event)
	if err != nil {
		return nil, fmt.Errorf("could not apply event: %w", err)
	}
	err = s.State.Publish(ctx, &schema.DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
	if err != nil {
		return nil, fmt.Errorf("could not update schema for module %s: %w", module.Name, err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) UpdateSchema(ctx context.Context, req *connect.Request[ftlv1.UpdateSchemaRequest]) (*connect.Response[ftlv1.UpdateSchemaResponse], error) {
	event, err := schema.EventFromProto(req.Msg.Event)
	if err != nil {
		return nil, fmt.Errorf("could not parse event: %w", err)
	}
	if err = s.State.Publish(ctx, event); err != nil {
		return nil, fmt.Errorf("could not apply event: %w", err)
	}
	return connect.NewResponse(&ftlv1.UpdateSchemaResponse{}), nil
}
func (s *Service) GetDeployments(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentsRequest]) (*connect.Response[ftlv1.GetDeploymentsResponse], error) {
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema state: %w", err)
	}
	deployments := view.GetDeployments()
	activeDeployments := view.GetAllActiveDeployments()
	var result []*ftlv1.DeployedSchema
	for key, deployment := range deployments {
		_, activeOk := activeDeployments[key]
		result = append(result, &ftlv1.DeployedSchema{
			DeploymentKey: key.String(),
			Schema:        deployment.ToProto(),
			IsActive:      activeOk,
		})
	}
	return connect.NewResponse(&ftlv1.GetDeploymentsResponse{Schema: result}), nil
}

// CreateChangeset creates a new changeset.
func (s *Service) CreateChangeset(ctx context.Context, req *connect.Request[ftlv1.CreateChangesetRequest]) (*connect.Response[ftlv1.CreateChangesetResponse], error) {
	modules, err := slices.MapErr(req.Msg.Modules, func(m *schemapb.Module) (*schema.Module, error) {
		out, err := schema.ModuleFromProto(m)
		if err != nil {
			return nil, fmt.Errorf("invalid module %s: %w", m.Name, err)
		}
		// Allocate a deployment key for the module.
		out.Runtime = &schema.ModuleRuntime{}
		out.Runtime.Deployment = &schema.ModuleRuntimeDeployment{
			DeploymentKey: key.NewDeploymentKey(m.Name),
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

	err = s.State.Publish(ctx, &schema.ChangesetCreatedEvent{
		Changeset: changeset,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create changeset %w", err)
	}

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

	uctx, cancel := context.WithCancelCause(ctx)
	defer cancel(fmt.Errorf("schemaservice: stopped watching for module changes"))
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
		case *schema.DeploymentCreatedEvent:
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
		case *schema.DeploymentDeactivatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key, optional.Ptr(event.Changeset))
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
		case *schema.DeploymentSchemaUpdatedEvent:
			view, err := s.State.View(ctx)
			if err != nil {
				return fmt.Errorf("failed to get schema state: %w", err)
			}
			dep, err := view.GetDeployment(event.Key, optional.Ptr(event.Changeset))
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
		case *schema.ChangesetCreatedEvent:
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				Event: &ftlv1.PullSchemaResponse_ChangesetCreated_{
					ChangesetCreated: &ftlv1.PullSchemaResponse_ChangesetCreated{
						// TODO: include changeset info
						Changeset: event.Changeset.ToProto(),
					},
				},
			})
			if err != nil {
				return err
			}
		case *schema.ChangesetFailedEvent:
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				Event: &ftlv1.PullSchemaResponse_ChangesetFailed_{
					ChangesetFailed: &ftlv1.PullSchemaResponse_ChangesetFailed{
						// TODO: include changeset info
						Key:   event.Key.String(),
						Error: event.Error,
					},
				},
			})
			if err != nil {
				return err
			}
		case *schema.ChangesetCommittedEvent:
			err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
				Event: &ftlv1.PullSchemaResponse_ChangesetCommitted_{
					ChangesetCommitted: &ftlv1.PullSchemaResponse_ChangesetCommitted{
						// TODO: include changeset info
						Key: event.Key.String(),
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
