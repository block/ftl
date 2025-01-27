package controller

import (
	"context"
	sha "crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/controller/leases"
	"github.com/block/ftl/backend/controller/observability"
	"github.com/block/ftl/backend/controller/scheduledtask"
	"github.com/block/ftl/backend/controller/state"
	ftldeployment "github.com/block/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	deploymentconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/deployment/v1/deploymentpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	internalobservability "github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

// CommonConfig between the production controller and development server.
type CommonConfig struct {
	IdleRunners    int           `help:"Number of idle runners to keep around (not supported in production)." default:"3"`
	WaitFor        []string      `help:"Wait for these modules to be deployed before becoming ready." placeholder:"MODULE"`
	CronJobTimeout time.Duration `help:"Timeout for cron jobs." default:"5m"`
}

type Config struct {
	Bind                         *url.URL       `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Key                          key.Controller `help:"Controller key (auto)." placeholder:"KEY"`
	Advertise                    *url.URL       `help:"Endpoint the Controller should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_ADVERTISE"`
	RunnerTimeout                time.Duration  `help:"Runner heartbeat timeout." default:"10s"`
	ControllerTimeout            time.Duration  `help:"Controller heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration  `help:"Deployment reservation timeout." default:"120s"`
	ModuleUpdateFrequency        time.Duration  `help:"Frequency to send module updates." default:"1s"` //TODO: FIX this, this should be based on streaming events, 1s is a temp workaround for the lack of dependencies within changesets
	ArtefactChunkSize            int            `help:"Size of each chunk streamed to the client." default:"1048576"`
	CommonConfig
}

func (c *Config) SetDefaults() {
	if c.Advertise == nil {
		c.Advertise = c.Bind
	}
}

func (c *Config) OpenDBAndInstrument(dsn string) (*sql.DB, error) {
	conn, err := internalobservability.OpenDBAndInstrument(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}
	conn.SetMaxIdleConns(10)
	conn.SetMaxOpenConns(10)
	return conn, nil
}

// Start the Controller. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	storage *artefacts.OCIArtefactService,
	adminClient ftlv1connect.AdminServiceClient,
	timelineClient *timelineclient.Client,
	schemaClient ftlv1connect.SchemaServiceClient,
	devel bool,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL controller")

	svc, err := New(ctx, adminClient, timelineClient, schemaClient, storage, config, devel)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
			rpc.GRPC(deploymentconnect.NewDeploymentServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
			rpc.PProf(),
		)
	})

	return g.Wait()
}

var _ ftlv1connect.ControllerServiceHandler = (*Service)(nil)

type clients struct {
	verb ftlv1connect.VerbServiceClient
}

type Service struct {
	leaser             leases.Leaser
	key                key.Controller
	deploymentLogsSink *deploymentLogsSink
	adminClient        ftlv1connect.AdminServiceClient

	tasks          *scheduledtask.Scheduler
	timelineClient *timelineclient.Client
	storage        *artefacts.OCIArtefactService

	// Map from runnerKey.String() to client.
	clients    *ttlcache.Cache[string, clients]
	clientLock sync.Mutex

	config Config

	routeTable   *routing.RouteTable
	schemaClient ftlv1connect.SchemaServiceClient
	runnerState  eventstream.EventStream[state.RunnerState, state.RunnerEvent]
}

func New(
	ctx context.Context,
	adminClient ftlv1connect.AdminServiceClient,
	timelineClient *timelineclient.Client,
	schemaClient ftlv1connect.SchemaServiceClient,
	storage *artefacts.OCIArtefactService,
	config Config,
	devel bool,
) (*Service, error) {
	logger := log.FromContext(ctx)
	logger = logger.Scope("controller")
	ctx = log.ContextWithLogger(ctx, logger)
	controllerKey := config.Key
	if config.Key.IsZero() {
		controllerKey = key.NewControllerKey(config.Bind.Hostname(), config.Bind.Port())
	}
	config.SetDefaults()

	// Override some defaults during development mode.
	if devel {
		config.RunnerTimeout = time.Second * 5
		config.ControllerTimeout = time.Second * 5
	}

	ldb := leases.NewClientLeaser(ctx)
	scheduler := scheduledtask.New(ctx, controllerKey, ldb)

	routingTable := routing.New(ctx, schemaeventsource.New(ctx, rpc.ClientFromContext[ftlv1connect.SchemaServiceClient](ctx)))

	svc := &Service{
		tasks:          scheduler,
		timelineClient: timelineClient,
		leaser:         ldb,
		key:            controllerKey,
		clients:        ttlcache.New(ttlcache.WithTTL[string, clients](time.Minute)),
		config:         config,
		routeTable:     routingTable,
		storage:        storage,
		schemaClient:   schemaClient,
		runnerState:    state.NewInMemoryRunnerState(ctx),
		adminClient:    adminClient,
	}

	svc.deploymentLogsSink = newDeploymentLogsSink(ctx, timelineClient)

	// Use min, max backoff if we are running in production, otherwise use
	// (1s, 1s) (or develBackoff). Will also wrap the job such that it its next
	// runtime is capped at 1s.
	maybeDevelTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) (backoff.Backoff, scheduledtask.Job) {
		if len(develBackoff) > 1 {
			panic("too many devel backoffs")
		}
		chain := job

		// Trace controller operations
		job = func(ctx context.Context) (time.Duration, error) {
			ctx, span := observability.Controller.BeginSpan(ctx, name)
			defer span.End()
			return chain(ctx)
		}

		if devel {
			chain := job
			job = func(ctx context.Context) (time.Duration, error) {
				next, err := chain(ctx)
				// Cap at 1s in development mode.
				return min(next, maxNext), err
			}
			if len(develBackoff) == 1 {
				return develBackoff[0], job
			}
			return backoff.Backoff{Min: time.Second, Max: time.Second}, job
		}
		return makeBackoff(minDelay, maxDelay), job
	}

	singletonTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) {
		maybeDevelJob, backoff := maybeDevelTask(job, name, maxNext, minDelay, maxDelay, develBackoff...)
		svc.tasks.Singleton(name, maybeDevelJob, backoff)
	}

	// Singleton tasks use leases to only run on a single controller.
	singletonTask(svc.reapStaleRunners, "reap-stale-runners", time.Second*2, time.Second, time.Second*10)
	return svc, nil
}

// ProcessList lists "processes" running on the cluster.
func (s *Service) ProcessList(ctx context.Context, req *connect.Request[ftlv1.ProcessListRequest]) (*connect.Response[ftlv1.ProcessListResponse], error) {
	currentState, err := s.runnerState.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get runner state: %w", err)
	}
	runners := currentState.Runners()

	deployments, err := s.schemaClient.GetDeployments(ctx, connect.NewRequest(&ftlv1.GetDeploymentsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("failed to get schema deployments: %w", err)
	}

	deploymentMap := map[string]*ftlv1.DeployedSchema{}
	for _, deployment := range deployments.Msg.Schema {
		deploymentMap[deployment.DeploymentKey] = deployment
	}

	out, err := slices.MapErr(runners, func(p state.Runner) (*ftlv1.ProcessListResponse_Process, error) {
		runner := &ftlv1.ProcessListResponse_ProcessRunner{
			Key:      p.Key.String(),
			Endpoint: p.Endpoint,
		}
		minReplicas := int32(0)
		deployment, ok := deploymentMap[p.Deployment.String()]
		if ok {
			minReplicas = deployment.Schema.GetRuntime().GetScaling().GetMinReplicas()
		}
		return &ftlv1.ProcessListResponse_Process{
			Deployment:  p.Deployment.String(),
			Runner:      runner,
			MinReplicas: minReplicas,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.ProcessListResponse{Processes: out}), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	currentRunnerState, err := s.runnerState.View(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get runner state: %w", err)
	}

	runners := currentRunnerState.Runners()
	deploymentsResponse, err := s.schemaClient.GetDeployments(ctx, connect.NewRequest(&ftlv1.GetDeploymentsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("failed to get schema deployments: %w", err)
	}
	activeDeployments := map[string]*schema.Module{}
	for _, deployment := range deploymentsResponse.Msg.Schema {
		if deployment.IsActive {
			module, err := schema.ModuleFromProto(deployment.Schema)
			if err != nil {
				return nil, fmt.Errorf("failed to get schema module: %w", err)
			}
			activeDeployments[deployment.DeploymentKey] = module
		}
	}
	allModules := s.routeTable.Current()
	routes := slices.Map(allModules.Schema().Modules, func(module *schema.Module) (out *ftlv1.StatusResponse_Route) {
		key := ""
		endpoint := ""
		if dkey := module.GetRuntime().GetDeployment().GetDeploymentKey(); !dkey.IsZero() {
			key = dkey.String()
			endpoint = module.GetRuntime().GetDeployment().Endpoint
		}
		return &ftlv1.StatusResponse_Route{
			Module:     module.Name,
			Deployment: key,
			Endpoint:   endpoint,
		}
	})
	replicas := map[string]int32{}
	protoRunners, err := slices.MapErr(runners, func(r state.Runner) (*ftlv1.StatusResponse_Runner, error) {
		asString := r.Deployment.String()
		deployment := &asString
		replicas[asString]++
		return &ftlv1.StatusResponse_Runner{
			Key:        r.Key.String(),
			Endpoint:   r.Endpoint,
			Deployment: deployment,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	var deployments []*ftlv1.StatusResponse_Deployment
	for key, deployment := range activeDeployments {
		var minReplicas int32
		if deployment.Runtime != nil && deployment.Runtime.Scaling != nil {
			minReplicas = deployment.Runtime.Scaling.MinReplicas
		}
		deployments = append(deployments, &ftlv1.StatusResponse_Deployment{
			Key:         key,
			Language:    deployment.Runtime.Base.Language,
			Name:        deployment.Name,
			MinReplicas: minReplicas,
			Replicas:    replicas[key],
			Schema:      deployment.ToProto(),
		})
	}
	resp := &ftlv1.StatusResponse{
		Controllers: []*ftlv1.StatusResponse_Controller{{
			Key:      s.key.String(),
			Endpoint: s.config.Bind.String(),
			Version:  ftl.Version,
		}},
		Runners:     protoRunners,
		Deployments: deployments,
		Routes:      routes,
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) setDeploymentReplicas(ctx context.Context, key key.Deployment, minReplicas int) (err error) {
	deployments, err := s.schemaClient.GetDeployments(ctx, connect.NewRequest(&ftlv1.GetDeploymentsRequest{}))

	if err != nil {
		return fmt.Errorf("failed to get schema deployments: %w", err)
	}
	deploymentMap := map[string]*ftlv1.DeployedSchema{}
	for _, deployment := range deployments.Msg.Schema {
		deploymentMap[deployment.DeploymentKey] = deployment
	}
	deployment, ok := deploymentMap[key.String()]
	if !ok {
		return fmt.Errorf("could not get deployment: %w", err)
	}

	_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
		Msg: &ftlv1.UpdateSchemaRequest{
			Event: &schemapb.Event{
				Value: &schemapb.Event_DeploymentReplicasUpdatedEvent{
					DeploymentReplicasUpdatedEvent: &schemapb.DeploymentReplicasUpdatedEvent{
						Key:      key.String(),
						Replicas: int64(minReplicas),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not update deployment replicas: %w", err)
	}
	if minReplicas == 0 {
		_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
			Msg: &ftlv1.UpdateSchemaRequest{
				Event: &schemapb.Event{
					Value: &schemapb.Event_DeploymentDeactivatedEvent{
						DeploymentDeactivatedEvent: &schemapb.DeploymentDeactivatedEvent{
							Key:           key.String(),
							ModuleRemoved: true,
						},
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("could not deactivate deployment: %w", err)
		}
	} else if deployment.Schema.GetRuntime().GetScaling().GetMinReplicas() == 0 {
		_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
			Msg: &ftlv1.UpdateSchemaRequest{
				Event: &schemapb.Event{
					Value: &schemapb.Event_DeploymentActivatedEvent{
						DeploymentActivatedEvent: &schemapb.DeploymentActivatedEvent{
							Key:         key.String(),
							MinReplicas: int64(minReplicas),
							ActivatedAt: timestamppb.New(time.Now()),
						},
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("could not activate deployment: %w", err)
		}
	}
	s.timelineClient.Publish(ctx, timelineclient.DeploymentUpdated{
		DeploymentKey:   key,
		MinReplicas:     minReplicas,
		PrevMinReplicas: int(deployment.Schema.GetRuntime().GetScaling().GetMinReplicas()),
	})

	return nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, c *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	newDeploymentKey, err := key.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger := s.getDeploymentLogger(ctx, newDeploymentKey)
	logger.Debugf("Replace deployment for: %s", newDeploymentKey)

	deployments, err := s.schemaClient.GetDeployments(ctx, connect.NewRequest(&ftlv1.GetDeploymentsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("failed to get schema deployments: %w", err)
	}
	deploymentMap := map[string]*ftlv1.DeployedSchema{}
	for _, deployment := range deployments.Msg.Schema {
		deploymentMap[deployment.DeploymentKey] = deployment
	}

	newDeployment, ok := deploymentMap[newDeploymentKey.String()]
	if !ok {
		return nil, fmt.Errorf("deployment not found: %s", newDeploymentKey)
	}
	minReplicas := int(c.Msg.MinReplicas)
	_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
		Msg: &ftlv1.UpdateSchemaRequest{
			Event: &schemapb.Event{
				Value: &schemapb.Event_DeploymentActivatedEvent{
					DeploymentActivatedEvent: &schemapb.DeploymentActivatedEvent{
						Key:         newDeploymentKey.String(),
						MinReplicas: int64(minReplicas),
						ActivatedAt: timestamppb.New(time.Now()),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("replace deployment failed to activate: %w", err)
	}

	// If there's an existing deployment, set its desired replicas to 0
	var replacedDeploymentKey optional.Option[key.Deployment]
	// TODO: remove all this, it needs to be event driven
	var oldDeployment *schema.Module
	var oldKey key.Deployment

	activeDeployments := map[string]*schema.Module{}
	for _, deployment := range deployments.Msg.Schema {
		if deployment.IsActive {
			module, err := schema.ModuleFromProto(deployment.Schema)
			if err != nil {
				return nil, fmt.Errorf("failed to get schema module: %w", err)
			}
			activeDeployments[deployment.DeploymentKey] = module
		}
	}

	for k, dep := range activeDeployments {
		if dep.Name == newDeployment.Schema.Name {
			oldDeployment = dep
			newKey, err := key.ParseDeploymentKey(k)
			if err != nil {
				return nil, fmt.Errorf("invalid deployment key: %w", err)
			}
			oldKey = newKey
			break
		}
	}
	if oldDeployment != nil {
		if oldKey.String() == newDeploymentKey.String() {
			return nil, fmt.Errorf("replace deployment failed: deployment already exists from %v to %v", oldKey, newDeploymentKey)
		}
		_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
			Msg: &ftlv1.UpdateSchemaRequest{
				Event: &schemapb.Event{
					Value: &schemapb.Event_DeploymentReplicasUpdatedEvent{
						DeploymentReplicasUpdatedEvent: &schemapb.DeploymentReplicasUpdatedEvent{
							Key:      newDeploymentKey.String(),
							Replicas: int64(minReplicas),
						},
					},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to set new deployment replicas from %v to %v: %w", oldKey, newDeploymentKey, err)
		}
		_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
			Msg: &ftlv1.UpdateSchemaRequest{
				Event: &schemapb.Event{
					Value: &schemapb.Event_DeploymentDeactivatedEvent{
						DeploymentDeactivatedEvent: &schemapb.DeploymentDeactivatedEvent{
							Key:           oldKey.String(),
							ModuleRemoved: true,
						},
					},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to deactivate old deployment %v: %w", oldKey, err)
		}
		replacedDeploymentKey = optional.Some(oldKey)
	} else {
		// Set the desired replicas for the new deployment
		_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
			Msg: &ftlv1.UpdateSchemaRequest{
				Event: &schemapb.Event{
					Value: &schemapb.Event_DeploymentReplicasUpdatedEvent{
						DeploymentReplicasUpdatedEvent: &schemapb.DeploymentReplicasUpdatedEvent{
							Key:      newDeploymentKey.String(),
							Replicas: int64(minReplicas),
						},
					},
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to set replicas for %v: %w", newDeploymentKey, err)
		}
	}

	s.timelineClient.Publish(ctx, timelineclient.DeploymentCreated{
		DeploymentKey:      newDeploymentKey,
		ModuleName:         newDeployment.Schema.Name,
		MinReplicas:        minReplicas,
		ReplacedDeployment: replacedDeploymentKey,
		Language:           newDeployment.Schema.Runtime.Base.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("replace deployment failed to create event: %w", err)
	}

	return connect.NewResponse(&ftlv1.ReplaceDeployResponse{}), nil
}

func (s *Service) RegisterRunner(ctx context.Context, stream *connect.ClientStream[ftlv1.RegisterRunnerRequest]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {

	deferredDeregistration := false

	logger := log.FromContext(ctx)
	for stream.Receive() {
		msg := stream.Msg()
		endpoint, err := url.Parse(msg.Endpoint)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid endpoint: %w", err))
		}
		if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid endpoint scheme %q", endpoint.Scheme))
		}
		runnerKey, err := key.ParseRunnerKey(msg.Key)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid key: %w", err))
		}

		runnerStr := fmt.Sprintf("%s (%s)", endpoint, runnerKey)
		logger.Tracef("Heartbeat received from runner %s", runnerStr)

		deploymentKey, err := key.ParseDeploymentKey(msg.Deployment)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		// The created event does not matter if it is a new runner or not.
		err = s.runnerState.Publish(ctx, &state.RunnerRegisteredEvent{
			Key:        runnerKey,
			Endpoint:   msg.Endpoint,
			Deployment: deploymentKey,
			Time:       time.Now(),
		})
		if err != nil {
			return nil, err
		}
		if !deferredDeregistration {
			// Deregister the runner if the Runner disconnects.
			defer func() {
				err := s.runnerState.Publish(ctx, &state.RunnerDeletedEvent{Key: runnerKey})
				if err != nil {
					logger.Errorf(err, "Could not deregister runner %s", runnerStr)
				}
			}()
			deferredDeregistration = true
		}
		if err != nil {
			return nil, fmt.Errorf("could not sync routes: %w", err)
		}
	}
	if stream.Err() != nil {
		return nil, stream.Err()
	}
	return connect.NewResponse(&ftlv1.RegisterRunnerResponse{}), nil
}

func (s *Service) GetDeployment(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	dkey, err := key.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	deployment, err := s.getDeployment(ctx, dkey)
	if err != nil {
		return nil, err
	}

	logger := s.getDeploymentLogger(ctx, dkey)
	logger.Debugf("Get deployment for: %s", dkey.String())

	artefacts := []*ftlv1.DeploymentArtefact{}
	for artefact := range slices.FilterVariants[schema.MetadataArtefact](deployment.Metadata) {
		artefacts = append(artefacts, &ftlv1.DeploymentArtefact{
			Digest:     artefact.Digest,
			Path:       artefact.Path,
			Executable: artefact.Executable,
		})
	}

	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema: deployment.ToProto(),
	}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	dkey, err := key.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	deployment, err := s.getDeployment(ctx, dkey)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}

	logger := s.getDeploymentLogger(ctx, dkey)
	logger.Debugf("Get deployment artefacts for: %s", dkey.String())

	chunk := make([]byte, s.config.ArtefactChunkSize)
nextArtefact:
	for artefact := range slices.FilterVariants[schema.MetadataArtefact](deployment.Metadata) {
		digest, err := sha256.ParseSHA256(artefact.Digest)
		if err != nil {
			return fmt.Errorf("invalid digest: %w", err)
		}
		deploymentArtefact := &state.DeploymentArtefact{
			Digest:     digest,
			Path:       artefact.Path,
			Executable: artefact.Executable,
		}
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(ftlv1.ArtefactToProto(deploymentArtefact), clientArtefact) {
				continue nextArtefact
			}
		}
		reader, err := s.storage.Download(ctx, digest)
		if err != nil {
			return fmt.Errorf("could not download artefact: %w", err)
		}
		defer reader.Close()
		for {

			n, err := reader.Read(chunk)
			if n != 0 {
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: ftlv1.ArtefactToProto(deploymentArtefact),
					Chunk:    chunk[:n],
				}); err != nil {
					return fmt.Errorf("could not send artefact chunk: %w", err)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return fmt.Errorf("could not read artefact chunk: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	if len(s.config.WaitFor) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	routeView := s.routeTable.Current()
	// It's not actually ready until it is in the routes table
	var missing []string
	for _, module := range s.config.WaitFor {
		if _, ok := routeView.GetForModule(module).Get(); !ok {
			missing = append(missing, module)
		}
	}
	if len(missing) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	msg := fmt.Sprintf("waiting for deployments: %s", strings.Join(missing, ", "))
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: &msg}), nil
}

// GetDeploymentContext retrieves config, secrets and DSNs for a module.
func (s *Service) GetDeploymentContext(ctx context.Context, req *connect.Request[ftldeployment.GetDeploymentContextRequest], resp *connect.ServerStream[ftldeployment.GetDeploymentContextResponse]) error {
	logger := log.FromContext(ctx)
	updates := s.routeTable.Subscribe()
	defer s.routeTable.Unsubscribe(updates)
	depName := req.Msg.Deployment
	key, err := key.ParseDeploymentKey(depName)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	deployment, err := s.getDeployment(ctx, key)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}
	module := deployment.Name

	// Initialize checksum to -1; a zero checksum does occur when the context contains no settings
	lastChecksum := int64(-1)

	callableModules := map[string]bool{}
	for _, decl := range deployment.Decls {
		switch entry := decl.(type) {
		case *schema.Verb:
			for _, md := range entry.Metadata {
				if calls, ok := md.(*schema.MetadataCalls); ok {
					for _, call := range calls.Calls {
						callableModules[call.Module] = true
					}
				}
			}
		default:

		}
	}
	callableModuleNames := maps.Keys(callableModules)
	callableModuleNames = slices.Sort(callableModuleNames)
	logger.Debugf("Modules %s can call %v", module, callableModuleNames)
	for {
		h := sha.New()

		routeView := s.routeTable.Current()
		configsResp, err := s.adminClient.MapConfigsForModule(ctx, &connect.Request[ftlv1.MapConfigsForModuleRequest]{Msg: &ftlv1.MapConfigsForModuleRequest{Module: module}})
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get configs: %w", err))
		}
		configs := configsResp.Msg.Values
		routeTable := map[string]string{}
		for _, module := range callableModuleNames {
			if module == deployment.Name {
				continue
			}
			deployment, ok := routeView.GetDeployment(module).Get()
			if !ok {
				continue
			}
			if route, ok := routeView.Get(deployment).Get(); ok && route.String() != "" {
				routeTable[deployment.String()] = route.String()
			}
		}

		secretsResp, err := s.adminClient.MapSecretsForModule(ctx, &connect.Request[ftlv1.MapSecretsForModuleRequest]{Msg: &ftlv1.MapSecretsForModuleRequest{Module: module}})
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get secrets: %w", err))
		}
		secrets := secretsResp.Msg.Values

		if err := hashConfigurationMap(h, configs); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on configs: %w", err))
		}
		if err := hashConfigurationMap(h, secrets); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on secrets: %w", err))
		}
		if err := hashRoutesTable(h, routeTable); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on routes: %w", err))
		}

		checksum := int64(binary.BigEndian.Uint64((h.Sum(nil))[0:8]))

		if checksum != lastChecksum {
			logger.Debugf("Sending module context for: %s routes: %v", module, routeTable)
			response := deploymentcontext.NewBuilder(module).AddConfigs(configs).AddSecrets(secrets).AddRoutes(routeTable).Build().ToProto()

			if err := resp.Send(response); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("could not send response: %w", err))
			}

			lastChecksum = checksum
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.config.ModuleUpdateFrequency):
		case <-updates:

		}
	}
}

// hashConfigurationMap computes an order invariant checksum on the configuration
// settings supplied in the map.
func hashConfigurationMap(h hash.Hash, m map[string][]byte) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return fmt.Errorf("error hashing configuration: %w", err)
		}
	}
	return nil
}

// hashRoutesTable computes an order invariant checksum on the routes
func hashRoutesTable(h hash.Hash, m map[string]string) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return fmt.Errorf("error hashing routes: %w", err)
		}
	}
	return nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	return s.callWithRequest(ctx, headers.CopyRequestForForwarding(req), optional.None[key.Request](), optional.None[key.Request]())
}

func (s *Service) callWithRequest(
	ctx context.Context,
	req *connect.Request[ftlv1.CallRequest],
	requestKeyOpt optional.Option[key.Request],
	parentKey optional.Option[key.Request],
) (*connect.Response[ftlv1.CallResponse], error) {
	logger := log.FromContext(ctx)
	start := time.Now()
	ctx, span := observability.Calls.BeginSpan(ctx, req.Msg.Verb)
	defer span.End()

	if req.Msg.Verb == nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: missing verb"))
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("verb is required"))
	}
	if req.Msg.Body == nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: missing body"))
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("body is required"))
	}

	routes := s.routeTable.Current()
	sch := routes.Schema()

	verbRef, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid verb"))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid verb: %w", err))
	}
	verb := &schema.Verb{}
	logger = logger.Module(verbRef.Module)

	if err := sch.ResolveToType(verbRef, verb); err != nil {
		if errors.Is(err, schema.ErrNotFound) {
			observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb not found"))
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb resolution failed"))
		return nil, err
	}

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get callers"))
		return nil, err
	}

	var currentCaller *schema.Ref // might be nil but that's fine. just means that it's not a cal from another verb
	if len(callers) > 0 {
		currentCaller = callers[len(callers)-1]
	}

	var requestKey key.Request
	var isNewRequestKey bool
	if k, ok := requestKeyOpt.Get(); ok {
		requestKey = k
		isNewRequestKey = false
	} else {
		k, ok, err := headers.GetRequestKey(req.Header())
		if err != nil {
			observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get request key"))
			return nil, err
		} else if !ok {
			requestKey = key.NewRequestKey(key.OriginIngress, "grpc")
			isNewRequestKey = true
		} else {
			requestKey = k
			isNewRequestKey = false
		}
	}
	if isNewRequestKey {
		headers.SetRequestKey(req.Header(), requestKey)
	}

	module := verbRef.Module
	deployment, ok := routes.GetDeployment(module).Get()
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment"))
		return nil, fmt.Errorf("deployment not found for module %q", module)
	}

	callEvent := &timelineclient.Call{
		DeploymentKey:    deployment,
		RequestKey:       requestKey,
		ParentRequestKey: parentKey,
		StartTime:        start,
		DestVerb:         verbRef,
		Callers:          callers,
		Request:          req.Msg,
	}

	route, ok := routes.Get(deployment).Get()
	if !ok {
		err = fmt.Errorf("no routes for module %q", module)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("no routes for module"))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if currentCaller != nil && currentCaller.Module != module && !verb.IsExported() {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: verb not exported"))
		err = connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
	}

	err = validateCallBody(req.Msg.Body, verb, sch)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid call body"))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		return nil, err
	}

	client := s.clientsForEndpoint(route.String())

	if pk, ok := parentKey.Get(); ok {
		ctx = rpc.WithParentRequestKey(ctx, pk)
	}
	ctx = rpc.WithRequestKey(ctx, requestKey)
	ctx = rpc.WithVerbs(ctx, append(callers, verbRef))
	reqVerb, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid verb"))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid verb: %w", err))
	}
	headers.AddCaller(req.Header(), reqVerb)

	response, err := client.verb.Call(ctx, req)
	var resp *connect.Response[ftlv1.CallResponse]
	if err == nil {
		resp = connect.NewResponse(response.Msg)
		callEvent.Response = result.Ok(resp.Msg)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	} else {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		logger.Errorf(err, "Call failed to verb %s for module %s", verbRef.String(), module)
	}

	s.timelineClient.Publish(ctx, callEvent)
	return resp, err
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := slices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, err
	}
	_, need, err := s.storage.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: slices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Uploading artefact")
	digest, err := s.storage.Upload(ctx, artefacts.Artefact{Content: req.Msg.Content})
	if err != nil {
		return nil, err
	}
	logger.Debugf("Created new artefact %s", digest)
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: digest[:]}), nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)

	ms := req.Msg.Schema
	if ms.Runtime == nil {
		err := errors.New("missing runtime metadata")
		logger.Errorf(err, "Missing runtime metadata")
		return nil, err
	}

	module, err := schema.ValidatedModuleFromProto(ms)
	if err != nil {
		logger.Errorf(err, "Invalid module schema")
		return nil, fmt.Errorf("invalid module schema: %w", err)
	}

	dkey := key.NewDeploymentKey(module.Name)
	module.ModRuntime().ModDeployment().CreatedAt = time.Now()
	_, err = s.schemaClient.UpdateSchema(ctx, &connect.Request[ftlv1.UpdateSchemaRequest]{
		Msg: &ftlv1.UpdateSchemaRequest{
			Event: &schemapb.Event{
				Value: &schemapb.Event_DeploymentCreatedEvent{
					DeploymentCreatedEvent: &schemapb.DeploymentCreatedEvent{
						Key:    dkey.String(),
						Schema: module.ToProto(),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Errorf(err, "Could not create deployment event")
		return nil, fmt.Errorf("could not create deployment event: %w", err)
	}

	deploymentLogger := s.getDeploymentLogger(ctx, dkey)
	deploymentLogger.Debugf("Created deployment %s", dkey)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: dkey.String()}), nil
}

func (s *Service) getDeployment(ctx context.Context, dkey key.Deployment) (*schema.Module, error) {
	deployments, err := s.schemaClient.GetDeployments(ctx, &connect.Request[ftlv1.GetDeploymentsRequest]{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}
	deploymentMap := map[string]*schema.Module{}
	for _, deployment := range deployments.Msg.Schema {
		module, err := schema.ModuleFromProto(deployment.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to get module from proto: %w", err)
		}
		deploymentMap[deployment.DeploymentKey] = module
	}

	deployment, ok := deploymentMap[dkey.String()]
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not retrieve deployment: %s", dkey))
	}
	return deployment, nil
}

// Return or create the RunnerService and VerbService clients for an endpoint.
func (s *Service) clientsForEndpoint(endpoint string) clients {
	clientItem := s.clients.Get(endpoint)
	if clientItem != nil {
		return clientItem.Value()
	}
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	// Double check it was not added while we were waiting for the lock
	clientItem = s.clients.Get(endpoint)
	if clientItem != nil {
		return clientItem.Value()
	}

	client := clients{
		verb: rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Error),
	}
	s.clients.Set(endpoint, client, time.Minute)
	return client
}

func (s *Service) reapStaleRunners(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)
	cs, err := s.runnerState.View(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get runner state: %w", err)
	}

	for _, runner := range cs.Runners() {
		if runner.LastSeen.Add(s.config.RunnerTimeout).Before(time.Now()) {
			runnerKey := runner.Key
			logger.Debugf("Reaping stale runner %s with last seen %v", runnerKey, runner.LastSeen.String())
			if err := s.runnerState.Publish(ctx, &state.RunnerDeletedEvent{Key: runnerKey}); err != nil {
				return 0, fmt.Errorf("failed to publish runner deleted event: %w", err)
			}
		}
	}
	return s.config.RunnerTimeout, nil
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentKey key.Deployment) *log.Logger {
	attrs := map[string]string{"deployment": deploymentKey.String()}
	if requestKey, _ := rpc.RequestKeyFromContext(ctx); requestKey.Ok() { //nolint:errcheck // best effort?
		attrs["request"] = requestKey.MustGet().String()
	}

	return log.FromContext(ctx).AddSink(s.deploymentLogsSink).Attrs(attrs)
}

func makeBackoff(min, max time.Duration) backoff.Backoff {
	return backoff.Backoff{
		Min:    min,
		Max:    max,
		Jitter: true,
		Factor: 2,
	}
}

func validateCallBody(body []byte, verb *schema.Verb, sch *schema.Schema) error {
	var root any
	err := json.Unmarshal(body, &root)
	if err != nil {
		return fmt.Errorf("request body is not valid JSON: %w", err)
	}

	var opts []schema.EncodingOption
	if e, ok := slices.FindVariant[*schema.MetadataEncoding](verb.Metadata); ok && e.Lenient {
		opts = append(opts, schema.LenientMode())
	}
	err = schema.ValidateJSONValue(verb.Request, []string{verb.Request.String()}, root, sch, opts...)
	if err != nil {
		return fmt.Errorf("could not validate call request body: %w", err)
	}
	return nil
}
