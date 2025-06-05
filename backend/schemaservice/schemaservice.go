package schemaservice

import (
	"context"
	"maps"
	gslices "slices"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/iterops"
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
}

type Service struct {
	State          *statemachine.SingleQueryHandle[struct{}, SchemaState, EventWrapper]
	Config         Config
	timelineClient timelineclient.Publisher
	receiverClient optional.Option[ftlv1connect.SchemaMirrorServiceClient] // TODO: rename as mirror
	devMode        bool
	creationLock   sync.Mutex
	rpcOpts        []rpc.Option
}

var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)
var _ rpc.Service = (*Service)(nil)

func (s *Service) GetDeployment(ctx context.Context, c *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	v, err := s.State.View(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schema state")
	}
	deploymentKey, err := key.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key")))
	}
	d, _, err := v.FindDeployment(deploymentKey)
	if err != nil {
		log.FromContext(ctx).Errorf(err, "failed to find deployment")
		return nil, errors.Wrap(err, "failed to find deployment")
	}
	return connect.NewResponse(&ftlv1.GetDeploymentResponse{Schema: d.ToProto()}), nil
}

var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

func NewLocalService(ctx context.Context, config Config, timelineClient timelineclient.Publisher, devMode bool) *Service {
	s := &Service{
		State:          statemachine.NewSingleQueryHandle(statemachine.NewLocalHandle(newStateMachine(ctx, "")), struct{}{}),
		Config:         config,
		timelineClient: timelineClient,
		receiverClient: optional.None[ftlv1connect.SchemaMirrorServiceClient](),
		devMode:        devMode,
	}
	s.rpcOpts = append(s.rpcOpts, rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, s))
	return s
}

func (s *Service) StartServices(context.Context) ([]rpc.Option, error) {
	return s.rpcOpts, nil
}

// Start the SchemaService. Blocks until the context is cancelled.
func New(
	ctx context.Context,
	config Config,
	timelineClient timelineclient.Publisher,
	receiverClient optional.Option[ftlv1connect.SchemaMirrorServiceClient],
	realm string,
	devMode bool,
) *Service {
	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL schema service")

	var rpcOpts []rpc.Option

	var svc *Service
	if config.Raft.DataDir == "" {
		// in local dev mode, use an inmemory state machine
		svc = NewLocalService(ctx, config, timelineClient, devMode)
	} else {
		clusterBuilder := raft.NewBuilder(&config.Raft)
		schemaShard := raft.AddShard(ctx, clusterBuilder, 1, newStateMachine(ctx, realm))
		cluster := clusterBuilder.Build(ctx)
		rpcOpts = append(rpcOpts, raft.RPCOption(cluster))
		svc = &Service{
			State:          statemachine.NewSingleQueryHandle(schemaShard, struct{}{}),
			Config:         config,
			timelineClient: timelineClient,
			receiverClient: receiverClient,
			rpcOpts:        rpcOpts,
		}
		svc.rpcOpts = append(svc.rpcOpts, rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc), rpc.StartHook(func(ctx context.Context) error {
			go svc.pushSchema(ctx)
			return nil
		}))
	}

	svc.devMode = devMode

	return svc
}
func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get controller state")
	}

	changesets := slices.Map(
		gslices.Collect(maps.Values(view.GetChangesets())),
		func(c *schema.Changeset) *schemapb.Changeset {
			return c.ToProto()
		},
	)

	return connect.NewResponse(&ftlv1.GetSchemaResponse{
		Schema:     view.GetCanonicalSchema().ToProto(),
		Changesets: changesets,
	}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	logger := log.FromContext(ctx)
	logger.Debugf("PullSchema subscription: %s", req.Msg.SubscriptionId)

	return errors.WithStack(s.watchModuleChanges(ctx, req.Msg.SubscriptionId, func(response *ftlv1.PullSchemaResponse) error {
		return errors.WithStack(stream.Send(response))
	}))
}

func (s *Service) UpdateDeploymentRuntime(ctx context.Context, req *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error) {
	event, err := schema.RuntimeElementFromProto(req.Msg.Update)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse event")
	}
	if event.Deployment.IsZero() {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Errorf("deployment key is required")))
	}
	var changeset *key.Changeset
	if req.Msg.Changeset != nil {
		cs, err := key.ParseChangesetKey(*req.Msg.Changeset)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse changeset")
		}
		changeset = &cs
	}
	err = s.publishEvent(ctx, &schema.DeploymentRuntimeEvent{Changeset: changeset, Payload: event})
	if err != nil {
		return nil, errors.Wrap(err, "could not apply event")
	}
	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) GetDeployments(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentsRequest]) (*connect.Response[ftlv1.GetDeploymentsResponse], error) {
	view, err := s.State.View(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schema state")
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

	var internalRealm *ftlv1.RealmChange
	var externalRealms []*ftlv1.RealmChange
	for _, realmChange := range req.Msg.RealmChanges {
		if !realmChange.External {
			if internalRealm != nil {
				return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Errorf("only one internal realm is allowed in a changeset")))
			}
			internalRealm = realmChange
		} else {
			externalRealms = append(externalRealms, realmChange)
		}
	}
	if internalRealm == nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Errorf("internal realm is required")))
	}

	modules, err := slices.MapErr(internalRealm.Modules, func(m *schemapb.Module) (*schema.Module, error) {
		out, err := schema.ModuleFromProto(m)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid module %s", m.Name)
		}
		if out.ModRuntime().ModDeployment().DeploymentKey.IsZero() {
			// Allocate a deployment key for the module, if it has not been baked into the image
			out.ModRuntime().ModDeployment().DeploymentKey = key.NewDeploymentKey(internalRealm.Name, m.Name)
		}
		return out, nil
	})
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, err))
	}
	changes := []*schema.RealmChange{{
		Name:     internalRealm.Name,
		External: false,
		Modules:  modules,
		ToRemove: internalRealm.ToRemove,
	}}
	for _, externalRealm := range externalRealms {
		var modules []*schema.Module
		for _, pb := range externalRealm.Modules {
			module, err := schema.ModuleFromProto(pb)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse module")
			}
			modules = append(modules, module)
		}
		changes = append(changes, &schema.RealmChange{
			Name:     externalRealm.Name,
			External: true,
			Modules:  modules,
		})
	}
	changeset := &schema.Changeset{
		Key:          key.NewChangesetKey(),
		State:        schema.ChangesetStatePreparing,
		RealmChanges: changes,
		CreatedAt:    time.Now(),
	}

	err = s.publishEvent(ctx, &schema.ChangesetCreatedEvent{Changeset: changeset})
	if err != nil {
		return nil, errors.Wrap(err, "could not create changeset")
	}

	return connect.NewResponse(&ftlv1.CreateChangesetResponse{Changeset: changeset.Key.String()}), nil
}

// PrepareChangeset prepares an active changeset for deployment.
func (s *Service) PrepareChangeset(ctx context.Context, req *connect.Request[ftlv1.PrepareChangesetRequest]) (*connect.Response[ftlv1.PrepareChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	err = s.publishEvent(ctx, &schema.ChangesetPreparedEvent{Key: changesetKey})
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare changeset")
	}
	return connect.NewResponse(&ftlv1.PrepareChangesetResponse{}), nil
}

// CommitChangeset makes all deployments for the changeset part of the canonical schema.
func (s *Service) CommitChangeset(ctx context.Context, req *connect.Request[ftlv1.CommitChangesetRequest]) (*connect.Response[ftlv1.CommitChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	err = s.publishEvent(ctx, &schema.ChangesetCommittedEvent{Key: changesetKey})
	if err != nil {
		return nil, errors.Wrap(err, "could not commit changeset")
	}
	v, err := s.State.View(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could get changeset state after commi")
	}
	cs, ok := v.GetChangeset(changesetKey).Get()
	if !ok {
		return nil, errors.Errorf("changeset %s not found", changesetKey)
	}
	return connect.NewResponse(&ftlv1.CommitChangesetResponse{
		Changeset: cs.ToProto(),
	}), nil
}

func (s *Service) DrainChangeset(ctx context.Context, req *connect.Request[ftlv1.DrainChangesetRequest]) (*connect.Response[ftlv1.DrainChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	err = s.publishEvent(ctx, &schema.ChangesetDrainedEvent{Key: changesetKey})
	if err != nil {
		return nil, errors.Wrap(err, "could not drain changeset")
	}
	return connect.NewResponse(&ftlv1.DrainChangesetResponse{}), nil
}

func (s *Service) FinalizeChangeset(ctx context.Context, req *connect.Request[ftlv1.FinalizeChangesetRequest]) (*connect.Response[ftlv1.FinalizeChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	err = s.publishEvent(ctx, &schema.ChangesetFinalizedEvent{Key: changesetKey})
	if err != nil {
		return nil, errors.Wrap(err, "could not de-provision changeset")
	}
	return connect.NewResponse(&ftlv1.FinalizeChangesetResponse{}), nil
}

func (s *Service) RollbackChangeset(ctx context.Context, req *connect.Request[ftlv1.RollbackChangesetRequest]) (*connect.Response[ftlv1.RollbackChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	logger := log.FromContext(ctx)
	logger.Infof("Rolling back changeset %s: %s", changesetKey, req.Msg.Error)
	err = s.publishEvent(ctx, &schema.ChangesetRollingBackEvent{
		Key:   changesetKey,
		Error: req.Msg.Error,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not fail changeset")
	}
	return connect.NewResponse(&ftlv1.RollbackChangesetResponse{}), nil
}

// FailChangeset fails an active changeset.
func (s *Service) FailChangeset(ctx context.Context, req *connect.Request[ftlv1.FailChangesetRequest]) (*connect.Response[ftlv1.FailChangesetResponse], error) {
	changesetKey, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid changeset key")))
	}
	logger := log.FromContext(ctx)
	logger.Infof("Failing changeset %s", changesetKey)
	err = s.publishEvent(ctx, &schema.ChangesetFailedEvent{Key: changesetKey})
	if err != nil {
		return nil, errors.Wrap(err, "could not fail changeset")
	}
	return connect.NewResponse(&ftlv1.FailChangesetResponse{}), nil
}

func (s *Service) publishEvent(ctx context.Context, event schema.Event) error {
	// Verify the event against the latest known state before publishing
	state, err := s.State.View(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get schema state")
	}
	if err := state.VerifyEvent(ctx, event); err != nil {
		return errors.Wrap(err, "invalid event")
	}
	if err := s.State.Publish(ctx, EventWrapper{Event: event}); err != nil {
		return errors.Wrap(err, "failed to publish event")
	}

	switch e := event.(type) {
	case *schema.ChangesetCreatedEvent:
		moduleNames := make([]string, 0, len(e.Changeset.InternalRealm().Modules))
		for _, module := range e.Changeset.InternalRealm().Modules {
			moduleNames = append(moduleNames, module.Name)
		}
		s.timelineClient.Publish(ctx, timelineclient.ChangesetCreated{
			Key:       e.Changeset.Key,
			CreatedAt: e.Changeset.CreatedAt,
			Modules:   moduleNames,
			ToRemove:  e.Changeset.InternalRealm().ToRemove,
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
		changeset, ok := state.GetChangeset(e.Key).Get()
		if ok {
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
	defer cancel2(errors.Wrap(context.Canceled, "schemaservice: stopped watching for module changes"))
	stateIter, err := s.State.StateIter(uctx)
	if err != nil {
		return errors.Wrap(err, "failed to get schema state iterator")
	}

	view, err := s.State.View(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get schema state")
	}

	notification := &schema.FullSchemaNotification{
		Schema:     view.GetCanonicalSchema(),
		Changesets: gslices.Collect(maps.Values(view.GetChangesets())),
	}
	err = sendChange(&ftlv1.PullSchemaResponse{
		Event: &schemapb.Notification{
			Value: &schemapb.Notification_FullSchemaNotification{
				FullSchemaNotification: notification.ToProto(),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to send initial schema")
	}

	for notification := range iterops.Changes(stateIter, EventExtractor) {
		logger.Debugf("Send Notification (subscription: %s, event: %T)", subscriptionID, notification.Event.Value)
		err := sendChange(notification)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) pushSchema(ctx context.Context) {
	logger := log.FromContext(ctx)
	client, ok := s.receiverClient.Get()
	if !ok {
		return
	}
	for {
		delay := time.After(time.Second * 5)

		// create stream and send empty message to initiate the stream
		logger.Tracef("Attempting to push schema to receiver")
		stream := client.PushSchema(ctx)
		if err := stream.Send(nil); err != nil {
			logger.Errorf(err, "Error while pushing initial schema to receiver")
		} else {
			err := s.watchModuleChanges(ctx, "push-schema", func(response *ftlv1.PullSchemaResponse) error {
				return errors.WithStack(stream.Send(&ftlv1.PushSchemaRequest{
					Event: response.Event,
				}))
			})
			if connect.CodeOf(err) == connect.CodeFailedPrecondition {
				// This is expected if the receiver is already receiving schema updates.
				logger.Tracef("Could not begin pushing schema to receiver: %s", err)
			} else if err != nil {
				logger.Errorf(err, "Error while pushing schema to receiver")
			} else {
				logger.Logf(log.Error, "Ended pushing schema to receiver without an error")
			}
			_, _ = stream.CloseAndReceive() //nolint:errcheck
		}
		select {
		case <-ctx.Done():
			return
		case <-delay:
			// Prevent tight loop
		}
	}
}
