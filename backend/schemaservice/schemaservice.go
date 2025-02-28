package schemaservice

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os/signal"
	gslices "slices"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/statemachine"
)

type CommonSchemaServiceConfig struct {
}

type Config struct {
	CommonSchemaServiceConfig
	Raft raft.RaftConfig `embed:"" prefix:"raft-"`
	Bind *url.URL        `help:"Socket to bind to." default:"http://127.0.0.1:8897" env:"FTL_BIND"`
}

type Service struct {
	State  *statemachine.SingleQueryHandle[struct{}, SchemaState, EventWrapper]
	Config Config
}

func (s *Service) GetDeployment(ctx context.Context, c *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	v, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema state: %w", err)
	}
	deploymentKey, err := key.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	d, _, err := v.FindDeployment(deploymentKey)
	if err != nil {
		log.FromContext(ctx).Errorf(err, "failed to find deployment")
		return nil, fmt.Errorf("failed to find deployment: %w", err)
	}
	return connect.NewResponse(&ftlv1.GetDeploymentResponse{Schema: d.ToProto()}), nil
}

var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

func New(ctx context.Context, handle statemachine.Handle[struct{}, SchemaState, EventWrapper], config Config) *Service {
	return &Service{
		State: statemachine.NewSingleQueryHandle(handle, struct{}{}),
	}
}

// Start the SchemaService. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL schema service")

	g, gctx := errgroup.WithContext(ctx)
	gctx, cancel := signal.NotifyContext(gctx, syscall.SIGTERM)
	defer cancel()

	var rpcOpts []rpc.Option

	var shard statemachine.Handle[struct{}, SchemaState, EventWrapper]
	if config.Raft.DataDir == "" {
		// in local dev mode, use an inmemory state machine
		shard = statemachine.NewLocalHandle(newStateMachine(ctx))
	} else {
		clusterBuilder := raft.NewBuilder(&config.Raft)
		schemaShard := raft.AddShard(gctx, clusterBuilder, 1, newStateMachine(ctx))
		cluster := clusterBuilder.Build(gctx)
		shard = schemaShard

		rpcOpts = append(rpcOpts, raft.RPCOption(cluster))
	}

	svc := New(ctx, shard, config)
	logger.Debugf("Listening on %s", config.Bind)

	g.Go(func() error {
		return rpc.Serve(gctx, config.Bind,
			append(rpcOpts,
				rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc),
				rpc.PProf(),
			)...,
		)
	})

	err := g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to run schema service: %w", err)
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
	changesets := slices.Map(gslices.Collect(maps.Values(view.GetChangesets())), func(c *schema.Changeset) *schemapb.Changeset { return c.ToProto() })
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: &schemapb.Schema{Modules: modules}, Changesets: changesets}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	logger := log.FromContext(ctx)
	logger.Debugf("PullSchema subscription: %s", req.Msg.SubscriptionId)

	return s.watchModuleChanges(ctx, req.Msg.SubscriptionId, func(response *ftlv1.PullSchemaResponse) error {
		return stream.Send(response)
	})
}

func (s *Service) UpdateDeploymentRuntime(ctx context.Context, req *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error) {
	event, err := schema.RuntimeElementFromProto(req.Msg.Update)
	if err != nil {
		return nil, fmt.Errorf("could not parse event: %w", err)
	}
	if event.Deployment.IsZero() {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("deployment key is required"))
	}
	var changeset *key.Changeset
	if req.Msg.Changeset != nil {
		cs, err := key.ParseChangesetKey(*req.Msg.Changeset)
		if err != nil {
			return nil, fmt.Errorf("could not parse changeset: %w", err)
		}
		changeset = &cs
	}
	err = s.publishEvent(ctx, &schema.DeploymentRuntimeEvent{Changeset: changeset, Payload: event})
	if err != nil {
		return nil, fmt.Errorf("could not apply event: %w", err)
	}
	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) GetDeployments(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentsRequest]) (*connect.Response[ftlv1.GetDeploymentsResponse], error) {
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema state: %w", err)
	}
	deployments := view.GetDeployments()
	activeDeployments := view.GetCanonicalDeployments()
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
		if out.Runtime == nil {
			out.Runtime = &schema.ModuleRuntime{}
		}
		out.Runtime.Deployment = &schema.ModuleRuntimeDeployment{
			DeploymentKey: key.NewDeploymentKey(m.Name),
		}
		return out, nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	changeset := &schema.Changeset{
		Key:       key.NewChangesetKey(),
		State:     schema.ChangesetStatePreparing,
		Modules:   modules,
		CreatedAt: time.Now(),
		ToRemove:  req.Msg.ToRemove,
	}

	// TODO: validate changeset schema with canonical schema
	err = s.publishEvent(ctx, &schema.ChangesetCreatedEvent{Changeset: changeset})
	if err != nil {
		return nil, fmt.Errorf("could not create changeset %w", err)
	}

	return connect.NewResponse(&ftlv1.CreateChangesetResponse{Changeset: changeset.Key.String()}), nil
}

// PrepareChangeset prepares an active changeset for deployment.
func (s *Service) PrepareChangeset(ctx context.Context, req *connect.Request[ftlv1.PrepareChangesetRequest]) (*connect.Response[ftlv1.PrepareChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	err = s.publishEvent(ctx, &schema.ChangesetPreparedEvent{Key: changesetKey})
	if err != nil {
		return nil, fmt.Errorf("could not prepare changeset %w", err)
	}
	return connect.NewResponse(&ftlv1.PrepareChangesetResponse{}), nil
}

// CommitChangeset makes all deployments for the changeset part of the canonical schema.
func (s *Service) CommitChangeset(ctx context.Context, req *connect.Request[ftlv1.CommitChangesetRequest]) (*connect.Response[ftlv1.CommitChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	err = s.publishEvent(ctx, &schema.ChangesetCommittedEvent{Key: changesetKey})
	if err != nil {
		return nil, fmt.Errorf("could not commit changeset %w", err)
	}
	v, err := s.State.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("could get changeset state after commit %w", err)
	}
	return connect.NewResponse(&ftlv1.CommitChangesetResponse{
		Changeset: v.changesets[changesetKey].ToProto(),
	}), nil
}

func (s *Service) DrainChangeset(ctx context.Context, req *connect.Request[ftlv1.DrainChangesetRequest]) (*connect.Response[ftlv1.DrainChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	err = s.publishEvent(ctx, &schema.ChangesetDrainedEvent{Key: changesetKey})
	if err != nil {
		return nil, fmt.Errorf("could not drain changeset %w", err)
	}
	return connect.NewResponse(&ftlv1.DrainChangesetResponse{}), nil
}

func (s *Service) FinalizeChangeset(ctx context.Context, req *connect.Request[ftlv1.FinalizeChangesetRequest]) (*connect.Response[ftlv1.FinalizeChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	err = s.publishEvent(ctx, &schema.ChangesetFinalizedEvent{Key: changesetKey})
	if err != nil {
		return nil, fmt.Errorf("could not de-provision changeset %w", err)
	}
	return connect.NewResponse(&ftlv1.FinalizeChangesetResponse{}), nil
}

func (s *Service) RollbackChangeset(ctx context.Context, req *connect.Request[ftlv1.RollbackChangesetRequest]) (*connect.Response[ftlv1.RollbackChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	logger := log.FromContext(ctx)
	logger.Infof("Rolling back changeset %s: %s", changesetKey, req.Msg.Error)
	err = s.publishEvent(ctx, &schema.ChangesetRollingBackEvent{
		Key:   changesetKey,
		Error: req.Msg.Error,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fail changeset %w", err)
	}
	return connect.NewResponse(&ftlv1.RollbackChangesetResponse{}), nil
}

// FailChangeset fails an active changeset.
func (s *Service) FailChangeset(ctx context.Context, req *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid changeset key: %w", err))
	}
	logger := log.FromContext(ctx)
	logger.Infof("Failing changeset %s", changesetKey)
	err = s.publishEvent(ctx, &schema.ChangesetFailedEvent{Key: changesetKey})
	if err != nil {
		return nil, fmt.Errorf("could not fail changeset %w", err)
	}
	return connect.NewResponse(&ftlv1.FailChangesetResponse{}), nil
}

func (s *Service) publishEvent(ctx context.Context, event schema.Event) error {
	// Verify the event against the latest known state before publishing
	state, err := s.State.View(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema state: %w", err)
	}
	if err := state.VerifyEvent(ctx, event); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	if err := s.State.Publish(ctx, EventWrapper{Event: event}); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

func (s *Service) watchModuleChanges(ctx context.Context, subscriptionID string, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx).Scope(subscriptionID)

	uctx, cancel2 := context.WithCancelCause(ctx)
	defer cancel2(fmt.Errorf("schemaservice: stopped watching for module changes: %w", context.Canceled))
	stateIter, err := s.State.StateIter(uctx)
	if err != nil {
		return fmt.Errorf("failed to get schema state iterator: %w", err)
	}

	view, err := s.State.View(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema state: %w", err)
	}

	modules := append([]*schema.Module{}, schema.Builtins())
	modules = append(modules, gslices.Collect(maps.Values(view.GetCanonicalDeployments()))...)

	notification := &schema.FullSchemaNotification{
		Schema:     &schema.Schema{Modules: modules},
		Changesets: gslices.Collect(maps.Values(view.GetChangesets())),
	}
	err = sendChange(&ftlv1.PullSchemaResponse{
		Event: &schemapb.Notification{Value: &schemapb.Notification_FullSchemaNotification{notification.ToProto()}},
	})
	if err != nil {
		return fmt.Errorf("failed to send initial schema: %w", err)
	}

	for notification := range iterops.Changes(stateIter, EventExtractor) {
		logger.Debugf("Send Notification (subscription: %s, event: %T)", subscriptionID, notification.Event.Value)
		err := sendChange(notification)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
