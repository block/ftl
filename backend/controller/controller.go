package controller

import (
	"context"
	sha "crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/controller/leases"
	"github.com/block/ftl/backend/controller/observability"
	"github.com/block/ftl/backend/controller/scheduledtask"
	"github.com/block/ftl/backend/controller/state"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	internalobservability "github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
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
	adminClient ftlv1connect.AdminServiceClient,
	schemaClient ftlv1connect.SchemaServiceClient,
	leaseClient leasepbconnect.LeaseServiceClient,
	devel bool,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL controller")

	svc, err := New(ctx, adminClient, schemaClient, leaseClient, config, devel)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
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
	leaser      leases.Leaser
	key         key.Controller
	adminClient ftlv1connect.AdminServiceClient

	tasks *scheduledtask.Scheduler

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
	schemaClient ftlv1connect.SchemaServiceClient,
	leaserClient leasepbconnect.LeaseServiceClient,
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

	ldb := leases.NewClientLeaser(ctx, leaserClient)
	scheduler := scheduledtask.New(ctx, controllerKey, ldb)

	eventSource := schemaeventsource.New(ctx, "controller", schemaClient)
	routingTable := routing.New(ctx, eventSource)

	svc := &Service{
		tasks:        scheduler,
		leaser:       ldb,
		key:          controllerKey,
		clients:      ttlcache.New(ttlcache.WithTTL[string, clients](time.Minute)),
		config:       config,
		routeTable:   routingTable,
		schemaClient: schemaClient,
		runnerState:  state.NewInMemoryRunnerState(ctx),
		adminClient:  adminClient,
	}

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
		endpoint := module.GetRuntime().GetRunner().GetEndpoint()
		if dkey := module.GetRuntime().GetDeployment().GetDeploymentKey(); !dkey.IsZero() {
			key = dkey.String()
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

	logger := log.FromContext(ctx)
	logger.Debugf("Get deployment for: %s", dkey.String())

	artefacts := []*ftlv1.DeploymentArtefact{}
	for artefact := range slices.FilterVariants[schema.MetadataArtefact](deployment.Metadata) {
		artefacts = append(artefacts, &ftlv1.DeploymentArtefact{
			Digest:     artefact.Digest[:],
			Path:       artefact.Path,
			Executable: artefact.Executable,
		})
	}

	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema: deployment.ToProto(),
	}), nil
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
func (s *Service) GetDeploymentContext(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentContextRequest], resp *connect.ServerStream[ftlv1.GetDeploymentContextResponse]) error {
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

func makeBackoff(min, max time.Duration) backoff.Backoff {
	return backoff.Backoff{
		Min:    min,
		Max:    max,
		Jitter: true,
		Factor: 2,
	}
}
