package schemaservice

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os/signal"
	gslices "slices"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
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
	"github.com/block/ftl/internal/timelineclient"
)

type CommonSchemaServiceConfig struct {
}

type Config struct {
	CommonSchemaServiceConfig
	Raft raft.RaftConfig `embed:"" prefix:"raft-"`
	Bind *url.URL        `help:"Socket to bind to." default:"http://127.0.0.1:8897" env:"FTL_BIND"`
}

type Service struct {
	State          *statemachine.SingleQueryHandle[struct{}, SchemaState, EventWrapper]
	Config         Config
	timelineClient *timelineclient.Client
	devMode        bool
	creationLock   sync.Mutex
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

func New(ctx context.Context, handle statemachine.Handle[struct{}, SchemaState, EventWrapper], config Config, timelineClient *timelineclient.Client) *Service {
	return &Service{
		State:          statemachine.NewSingleQueryHandle(handle, struct{}{}),
		Config:         config,
		timelineClient: timelineClient,
	}
}

// Start the SchemaService. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	timelineClient *timelineclient.Client,
	devMode bool,
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

	svc := New(ctx, shard, config, timelineClient)
	svc.devMode = devMode
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

	realm := &schemapb.Realm{
		Name:    "default", // TODO: implement
		Modules: modules,
	}

	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: &schemapb.Schema{Realms: []*schemapb.Realm{realm}}, Changesets: changesets}), nil
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
	err = s.publishEvent(ctx, &schema.DeploymentRuntimeEvent{Changeset: changeset, Payload: event, Realm: req.Msg.Realm})
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
	s.creationLock.Lock()
	defer s.creationLock.Unlock()
	realmChanges, err := slices.MapErr(req.Msg.RealmChanges, func(r *ftlv1.RealmChange) (*schema.RealmChange, error) {
		modules := make([]*schema.Module, 0, len(r.Modules))
		for _, m := range r.Modules {
			out, err := schema.ModuleFromProto(m)
			if err != nil {
				return nil, fmt.Errorf("invalid module %s: %w", m.Name, err)
			}
			if !out.ModRuntime().ModDeployment().DeploymentKey.IsZero() {
				// In dev mode we relax this restriction to allow for hot reload endpoints to allocate a deployment key.
				if !s.devMode {
					return nil, fmt.Errorf("deployment key cannot be set on changeset creation, it must be allocated by the schema service")
				}
			} else {
				// Allocate a deployment key for the module.
				out.ModRuntime().ModDeployment().DeploymentKey = key.NewDeploymentKey(m.Name)
			}
		}

		return &schema.RealmChange{
			Modules:  modules,
			ToRemove: r.ToRemove,
		}, nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	changeset := &schema.Changeset{
		Key:          key.NewChangesetKey(),
		State:        schema.ChangesetStatePreparing,
		RealmChanges: realmChanges,
		CreatedAt:    time.Now(),
	}

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

	switch e := event.(type) {
	case *schema.ChangesetCreatedEvent:
		// TODO: support multiple realms in the timeline
		modules := e.Changeset.InternalModules()
		moduleNames := make([]string, 0, len(modules))
		for _, module := range modules {
			moduleNames = append(moduleNames, module.Name)
		}
		s.timelineClient.Publish(ctx, timelineclient.ChangesetCreated{
			Key:       e.Changeset.Key,
			CreatedAt: e.Changeset.CreatedAt,
			Modules:   moduleNames,
			ToRemove:  e.Changeset.InternalToRemove(),
		})
	case *schema.ChangesetPreparedEvent:
		s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
			Key:   e.Key,
			State: schema.ChangesetStatePrepared,
		})
	case *schema.ChangesetCommittedEvent:
		s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
			Key:   e.Key,
			State: schema.ChangesetStateCommitted,
		})
	case *schema.ChangesetDrainedEvent:
		s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
			Key:   e.Key,
			State: schema.ChangesetStateDrained,
		})
	case *schema.ChangesetFinalizedEvent:
		s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
			Key:   e.Key,
			State: schema.ChangesetStateFinalized,
		})
	case *schema.ChangesetRollingBackEvent:
		s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
			Key:   e.Key,
			State: schema.ChangesetStateRollingBack,
			Error: optional.Some(e.Error),
		})
	case *schema.ChangesetFailedEvent:
		changeset, err := state.GetChangeset(e.Key)
		if err == nil {
			s.timelineClient.Publish(ctx, timelineclient.ChangesetStateChanged{
				Key:   e.Key,
				State: schema.ChangesetStateFailed,
				Error: optional.Some(changeset.Error),
			})
		}
	case *schema.DeploymentRuntimeEvent:
		var changeset optional.Option[key.Changeset]
		if e.Changeset != nil {
			changeset = optional.Some(*e.Changeset)
		}
		s.timelineClient.Publish(ctx, timelineclient.DeploymentRuntime{
			Deployment: e.Payload.Deployment,
			Element:    e.Payload,
			UpdatedAt:  time.Now(),
			Changeset:  changeset,
		})
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
		Schema: &schema.Schema{Realms: []*schema.Realm{{
			Name:    "default", // TODO: implement
			Modules: modules,
		}}},
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
