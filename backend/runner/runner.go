// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"

	mysql "github.com/block/ftl-mysql-auth-proxy"
	hotreloadpb "github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1/hotreloadpbconnect"
	ftlleaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/pubsub/v1/pubsubpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/runner/observability"
	"github.com/block/ftl/backend/runner/proxy"
	"github.com/block/ftl/backend/runner/pubsub"
	"github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/artefacts"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/exec"
	ftlobservability "github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/pgproxy"
	"github.com/block/ftl/internal/rpc"
	timeline "github.com/block/ftl/internal/timelineclient"
	"github.com/block/ftl/internal/unstoppable"
)

type Config struct {
	Config                []string                `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
	Bind                  *url.URL                `help:"Endpoint the Runner should bind to and advertise." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	HealthBind            *url.URL                `help:"Endpoint the Runner should bind to for health check" env:"FTL_HEALTH_BIND"`
	Key                   key.Runner              `help:"Runner key (auto)."`
	SchemaEndpoint        *url.URL                `name:"schema-endpoint" help:"Schema server endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	AdminEndpoint         *url.URL                `name:"admin-endpoint" help:"Admin server endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"` // This is temporary, a quick temp hack to allow kube to get secrets / config, remove once this is fixed
	LeaseEndpoint         *url.URL                `name:"ftl-lease-endpoint" help:"Lease endpoint endpoint." env:"FTL_LEASE_ENDPOINT" default:"http://127.0.0.1:8895"`
	TimelineEndpoint      *url.URL                `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
	DeploymentKeepHistory int                     `help:"Number of deployments to keep history for." default:"3"`
	HeartbeatPeriod       time.Duration           `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter       time.Duration           `help:"Jitter to add to heartbeat period." default:"2s"`
	Deployment            key.Deployment          `help:"The deployment this runner is for." env:"FTL_DEPLOYMENT"`
	DebugPort             int                     `help:"The port to use for debugging." env:"FTL_DEBUG_PORT"`
	DevEndpoint           optional.Option[string] `help:"An existing endpoint to connect to in development mode" hidden:""`
	DevHotReloadEndpoint  optional.Option[string] `help:"The gRPC enpoint to send runner into to for hot reload." hidden:""`
	LocalRunners          bool                    `help:"Set to true if we are running with local runners" hidden:""`
	DevModeSchemaSequence int64                   `help:"Schema sequence number for dev mode runner. " hidden:""`
	DevModeRunnerSequence int64                   `help:"Runner sequence number  for dev mode runner. " hidden:""`
}

func Start(ctx context.Context, config Config, deploymentArtifactProvider artefacts.DeploymentArtefactProvider, deploymentContextProvider deploymentcontext.DeploymentContextProvider, schemaClient ftlv1connect.SchemaServiceClient) error {
	ctx, doneFunc := context.WithCancelCause(ctx)
	defer doneFunc(errors.Wrap(context.Canceled, "runner terminated"))
	hostname, err := os.Hostname()
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return errors.Wrap(err, "failed to get hostname")
	}
	pid := os.Getpid()

	runnerKey := config.Key
	if runnerKey.IsZero() {
		runnerKey = key.NewRunnerKey(config.Bind.Hostname(), config.Bind.Port())
	}

	logger := log.FromContext(ctx).Attrs(map[string]string{"runner": runnerKey.String()})
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Starting FTL Runner for %s", config.Deployment.String())
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return errors.WithStack(err)
	}

	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Using Schema endpoint %s", config.SchemaEndpoint.String())

	labels, err := structpb.NewStruct(map[string]any{
		"hostname": hostname,
		"pid":      pid,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	})
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return errors.Wrap(err, "failed to marshal labels")
	}
	timelineClient := timeline.NewClient(ctx, config.TimelineEndpoint)

	svc := &Service{
		key:                  runnerKey,
		config:               config,
		storage:              storage,
		deploymentProvider:   deploymentArtifactProvider,
		schemaClient:         schemaClient,
		timelineClient:       timelineClient,
		timelineLogSink:      timeline.NewLogSink(timelineClient, log.Debug),
		labels:               labels,
		cancelFunc:           doneFunc,
		devEndpoint:          config.DevEndpoint,
		devHotReloadEndpoint: config.DevHotReloadEndpoint,
		deploymentContextProvider: deploymentContextProvider,
	}

	module, err := svc.getModule(ctx, config.Deployment)
	if err != nil {
		return errors.Wrap(err, "failed to get module")
	}
	var git *schema.MetadataGit
	for m := range slices.FilterVariants[*schema.MetadataGit, schema.Metadata](module.Metadata) {
		git = m
		break
	}
	if git != nil {
		logger = logger.Attrs(map[string]string{"git-commit": git.Commit})
		ctx = log.ContextWithLogger(ctx, logger)
	}

	startedLatch := &sync.WaitGroup{}
	startedLatch.Add(2)
	g, ctx := errgroup.WithContext(ctx)
	dbAddresses := xsync.NewMapOf[string, string]()
	g.Go(func() error {
		return errors.WithStack(svc.startPgProxy(ctx, module, startedLatch, dbAddresses))
	})
	g.Go(func() error {
		return errors.WithStack(svc.startMySQLProxy(ctx, module, startedLatch, dbAddresses))
	})
	g.Go(func() error {
		startedLatch.Wait()
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		default:
			svc.queryService, err = query.New(ctx, module, dbAddresses)
			if err != nil {
				return errors.Wrap(err, "failed to create query service")
			}
			return errors.WithStack(svc.startDeployment(ctx, config.Deployment, module, dbAddresses))
		}
	})

	return errors.Wrap(g.Wait(), "failure in runner")
}

func (s *Service) startDeployment(ctx context.Context, key key.Deployment, module *schema.Module, dbAddresses *xsync.MapOf[string, string]) error {
	err := s.deploy(ctx, key, module, dbAddresses)
	if err != nil {
		// If we fail to deploy we just exit
		// Kube or local scaling will start a new instance to continue
		// This approach means we don't have to handle error states internally
		// It is managed externally by the scaling system
		return errors.WithStack(err)
	}
	go s.timelineLogSink.RunLogLoop(ctx)
	opts := []rpc.Option{
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s),
		rpc.GRPC(querypbconnect.NewQueryServiceHandler, s.queryService),
		rpc.GRPC(pubsubpbconnect.NewPubSubAdminServiceHandler, s.pubSub),
		rpc.HTTP("/", s),
		rpc.HealthCheck(s.healthCheck)}
	if s.config.HealthBind != nil {
		opts = append(opts, rpc.WithHealthCheckBind(s.config.HealthBind))
	}
	return errors.Wrap(rpc.Serve(ctx, s.config.Bind, opts...,
	), "failure in runner")
}

var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type deployment struct {
	key key.Deployment
	// Cancelled when plugin terminates
	ctx          context.Context
	cmd          optional.Option[exec.Cmd]
	endpoint     string // The endpoint the plugin is listening on.
	client       ftlv1connect.VerbServiceClient
	reverseProxy *httputil.ReverseProxy
}

type Service struct {
	key        key.Runner
	lock       sync.Mutex
	deployment atomic.Value[optional.Option[*deployment]]
	readyTime  atomic.Value[time.Time]

	config           Config
	deploymentProvider artefacts.DeploymentArtefactProvider
	deploymentContextProvider <-chan deploymentcontext.DeploymentContext
	schemaClient     ftlv1connect.SchemaServiceClient
	timelineClient   *timeline.Client
	timelineLogSink  *timeline.LogSink
	// Failed to register with the Controller
	registrationFailure  atomic.Value[optional.Option[error]]
	labels               *structpb.Struct
	cancelFunc           context.CancelCauseFunc
	devEndpoint          optional.Option[string]
	devHotReloadEndpoint optional.Option[string]
	proxy                *proxy.Service
	pubSub               *pubsub.Service
	queryService         *query.Service
	proxyBindAddress     *url.URL
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, errors.WithStack(connect.NewError(connect.CodeUnavailable, errors.New("no deployment")))
	}
	response, err := deployment.client.Call(ctx, req)
	if err != nil {
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(err, "Call to deployments %s failed to perform gRPC call", deployment.key)
		return nil, errors.WithStack(connect.NewError(connect.CodeOf(err), err))
	} else if response.Msg.GetError() != nil {
		// This is a user level error (i.e. something wrong in the users app)
		// Log it to the deployment logger
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(errors.Errorf("%v", response.Msg.GetError().GetMessage()), "Call to deployments %s failed", deployment.key)
	}

	return connect.NewResponse(response.Msg), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) getModule(ctx context.Context, key key.Deployment) (*schema.Module, error) {
	gdResp, err := s.schemaClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: s.config.Deployment.String()}))
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, errors.Wrap(err, "failed to get deployment from schema service")
	}

	module, err := schema.ValidatedModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	return module, nil
}

func (s *Service) deploy(ctx context.Context, key key.Deployment, module *schema.Module, dbAddresses *xsync.MapOf[string, string]) error {
	logger := log.FromContext(ctx)

	if err, ok := s.registrationFailure.Load().Get(); ok {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return errors.WithStack(connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to register runner")))
	}

	observability.Deployment.Started(ctx, key.String())
	defer observability.Deployment.Completed(ctx, key.String())

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.WithStack(errors.New("already deployed"))
	}

	deploymentLogger := s.getDeploymentLogger(ctx, key)
	ctx = log.ContextWithLogger(ctx, deploymentLogger)

	leaseServiceClient := rpc.Dial(ftlleaseconnect.NewLeaseServiceClient, s.config.LeaseEndpoint.String(), log.Error)

	s.proxy = proxy.New(s.deploymentContextProvider, leaseServiceClient, s.timelineClient,
		s.config.Bind.String(), s.config.Deployment, s.config.LocalRunners)

	pubSub, err := pubsub.New(module, key, s, s.timelineClient)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.Wrap(err, "failed to set up pubsub")
	}
	s.pubSub = pubSub

	parse, err := url.Parse("http://127.0.0.1:0")
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	proxyServer, err := rpc.NewServer(ctx, parse,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s.proxy),
		rpc.GRPC(ftlv1connect.NewDeploymentContextServiceHandler, s.proxy),
		rpc.GRPC(ftlleaseconnect.NewLeaseServiceHandler, s.proxy),
		rpc.GRPC(pubsubpbconnect.NewPublishServiceHandler, s.pubSub),
		rpc.GRPC(querypbconnect.NewQueryServiceHandler, s.queryService),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}
	urls := proxyServer.Bind.Subscribe(nil)
	go func() {
		err := proxyServer.Serve(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(err, "failed to serve")
			return
		}
	}()
	s.proxyBindAddress = <-urls

	// Make sure the proxy is up and running before we start the deployment
	// We get the bind address slightly before the server is ready, so we ping to be sure
	proxyPingClient := rpc.Dial(ftlv1connect.NewDeploymentContextServiceClient, s.proxyBindAddress.String(), log.Error)
	err = rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, proxyPingClient)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.Wrap(err, "failed to ping proxy")
	}
	var dep *deployment
	if ep, ok := s.devEndpoint.Get(); ok {
		if hotRelaodEp, ok := s.devHotReloadEndpoint.Get(); ok {
			hotReloadClient := rpc.Dial(hotreloadpbconnect.NewHotReloadServiceClient, hotRelaodEp, log.Error)
			err = rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, hotReloadClient)
			if err != nil {
				return errors.Wrap(err, "failed to ping hot reload endpoint")
			}
			var databases []*hotreloadpb.Database
			dbAddresses.Range(func(key string, value string) bool {
				databases = append(databases, &hotreloadpb.Database{
					Name:    key,
					Address: value,
				})
				return true
			})

			if err := s.queryService.UpdateConnections(ctx, module, dbAddresses); err != nil {
				return errors.Wrap(err, "failed to update query service connections")
			}

			go func() {
				delay := 50 * time.Millisecond
				connected := false
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(delay):
						info, err := hotReloadClient.RunnerInfo(ctx, connect.NewRequest(&hotreloadpb.RunnerInfoRequest{
							Deployment:    s.config.Deployment.String(),
							Address:       s.proxyBindAddress.String(),
							SchemaVersion: s.config.DevModeSchemaSequence,
							RunnerVersion: s.config.DevModeRunnerSequence,
							Databases:     databases,
						}))
						if err == nil {
							delay = 300 * time.Second
							if !connected {
								logger.Debugf("Runner connected to backend with schema version %d and runner no %d", s.config.DevModeSchemaSequence, s.config.DevModeRunnerSequence)
							}
							connected = true
							if info.Msg.Outdated {
								logger.Debugf("Runner is outdated, exiting")
								return
							}
						} else if connected {
							logger.Debugf("Lost connection to running dev mode process")
							connected = false
						}
					}
				}
			}()

		}
		client := rpc.Dial(ftlv1connect.NewVerbServiceClient, ep, log.Error)
		err = rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, client)
		dep = &deployment{
			ctx:      ctx,
			key:      key,
			cmd:      optional.None[exec.Cmd](),
			endpoint: ep,
			client:   client,
		}
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return errors.Wrap(err, "failed to ping dev endpoint")
		}
	} else {
		//err := download.ArtefactsFromOCI(ctx, s.schemaClient, key, deploymentDir, s.storage)
		deploymentDir, err := s.deploymentProvider.Provide()
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return errors.Wrap(err, "failed to download artefacts")
		}

		logger.Debugf("Setting FTL_RUNNER_ENDPOINT to %s", s.proxyBindAddress.String())
		envVars := []string{
			"FTL_RUNNER_ENDPOINT=" + s.proxyBindAddress.String(),
			"FTL_CONTROLLER_ENDPOINT=" + s.proxyBindAddress.String(),
			"FTL_CONFIG=" + strings.Join(s.config.Config, ","),
			"FTL_DEPLOYMENT=" + s.config.Deployment.String()}
		if s.config.DebugPort > 0 {
			envVars = append(envVars, fmt.Sprintf("FTL_DEBUG_PORT=%d", s.config.DebugPort))
		}

		verbCtx := log.ContextWithLogger(ctx, deploymentLogger.Attrs(map[string]string{"module": module.Name}))
		deployment, cmdCtx, err := plugin.Spawn(
			unstoppable.Context(verbCtx),
			log.FromContext(ctx).GetLevel(),
			module.Name,
			module.Name,
			deploymentDir,
			"./launch",
			ftlv1connect.NewVerbServiceClient,
			true,
			plugin.WithEnvars(
				envVars...,
			),
		)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return errors.Wrap(err, "failed to spawn plugin")
		}
		dep = s.makeDeployment(cmdCtx, key, deployment)
	}

	s.readyTime.Store(time.Now().Add(time.Second * 2)) // Istio is a bit flakey, add a small delay for readiness
	s.deployment.Store(optional.Some(dep))
	logger.Debugf("Deployed %s", key)

	err = s.pubSub.Consume(ctx)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.Wrap(err, "failed to set up pubsub consumption")
	}

	context.AfterFunc(ctx, func() {
		err := s.Close()
		if err != nil {
			// This is very common, as the process will have been explicitly killed.
			logger := log.FromContext(ctx)
			logger.Debugf("failed to terminate deployment: %s", err.Error())
		}
	})

	return nil
}

func (s *Service) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.queryService != nil {
		s.queryService.Close()
	}

	depl, ok := s.deployment.Load().Get()
	if !ok {
		return errors.WithStack(connect.NewError(connect.CodeNotFound, errors.New("no deployment")))
	}
	if cmd, ok := depl.cmd.Get(); ok {
		// Soft kill.
		err := cmd.Kill(syscall.SIGTERM)
		if err != nil {
			return errors.Wrap(err, "failed to kill plugin")
		}
		done := make(chan struct{})
		go func() {
			err := cmd.Wait()
			if err == nil {
				// Plugin exited cleanly.
				// If wait fails we try the sigterm hard kill below.
				close(done)
			}
		}()
		// Hard kill after 10 seconds.
		select {
		case <-done:
		case <-depl.ctx.Done():
		case <-time.After(10 * time.Second):
			if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
				err := cmd.Kill(syscall.SIGKILL)
				if err != nil {
					return errors.Wrap(err, "failed to kill plugin")
				}
			}
		}
	}
	s.deployment.Store(optional.None[*deployment]())
	return nil

}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		http.Error(w, "no deployment", http.StatusNotFound)
		return
	}
	deployment.reverseProxy.ServeHTTP(w, r)

}

func (s *Service) makeDeployment(ctx context.Context, key key.Deployment, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient, ftlv1.PingRequest, ftlv1.PingResponse, *ftlv1.PingResponse]) *deployment {
	proxy := httputil.NewSingleHostReverseProxy(plugin.Endpoint)
	return &deployment{
		ctx:          ctx,
		key:          key,
		cmd:          optional.Ptr(plugin.Cmd),
		endpoint:     plugin.Endpoint.String(),
		client:       plugin.Client,
		reverseProxy: proxy,
	}
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentKey key.Deployment) *log.Logger {
	attrs := map[string]string{"deployment": deploymentKey.String()}
	if requestKey, _ := rpc.RequestKeyFromContext(ctx); requestKey.Ok() { //nolint:errcheck // best effort
		attrs["request"] = requestKey.MustGet().String()
	}
	ctx = ftlobservability.AddSpanContextToLogger(ctx)

	return log.FromContext(ctx).AddSink(s.timelineLogSink).Attrs(attrs)
}

func (s *Service) healthCheck(writer http.ResponseWriter, request *http.Request) {
	if s.deployment.Load().Ok() {
		if s.readyTime.Load().After(time.Now()) {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	}
	writer.WriteHeader(http.StatusServiceUnavailable)
}

func (s *Service) startPgProxy(ctx context.Context, module *schema.Module, started *sync.WaitGroup, addresses *xsync.MapOf[string, string]) error {
	databases := map[string]*schema.Database{}
	for db := range slices.FilterVariants[*schema.Database](module.Decls) {
		if db.Type == "postgres" {
			databases[db.Name] = db
		}
	}

	if len(databases) == 0 {
		started.Done()
		return nil
	}

	// Map the channel to the optional.Option[pgproxy.Started] type.
	channel := make(chan pgproxy.Started)
	go func() {
		select {
		case pgProxy := <-channel:
			address := fmt.Sprintf("127.0.0.1:%d", pgProxy.Address.Port)
			for db := range databases {
				addresses.Store(db, address)
			}
			os.Setenv("FTL_PROXY_POSTGRES_ADDRESS", address)
			started.Done()
		case <-ctx.Done():
			started.Done()
			return
		}
	}()

	if err := pgproxy.New("127.0.0.1:0", func(ctx context.Context, params map[string]string) (string, error) {
		db, ok := databases[params["database"]]
		if !ok {
			return "", errors.Errorf("database %s not found", params["database"])
		}

		dsn, err := dsn.ResolvePostgresDSN(ctx, db.Runtime.Connections.Write)
		if err != nil {
			return "", errors.Wrap(err, "failed to resolve postgres DSN")
		}

		return dsn, nil
	}).Start(ctx, channel); err != nil {
		started.Done()
		return errors.Wrap(err, "failed to start pgproxy")
	}

	return nil
}

func (s *Service) startMySQLProxy(ctx context.Context, module *schema.Module, latch *sync.WaitGroup, addresses *xsync.MapOf[string, string]) error {
	defer latch.Done()
	logger := log.FromContext(ctx)

	databases := map[string]*schema.Database{}
	for _, decl := range module.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Type == "mysql" {
			databases[db.Name] = db
		}
	}

	if len(databases) == 0 {
		return nil
	}
	for db, decl := range databases {
		logger.Debugf("Starting MySQL proxy for %s", db)
		logger := log.FromContext(ctx)
		portC := make(chan int)
		errorC := make(chan error)
		databaseRuntime := decl.Runtime
		proxy := mysql.NewProxy("localhost", 0, func(ctx context.Context) (*mysql.Config, error) {
			cfg, err := dsn.ResolveMySQLConfig(ctx, databaseRuntime.Connections.Write)
			if err != nil {
				return nil, errors.Wrap(err, "failed to resolve MySQL DSN")
			}
			return cfg, nil
		}, &mysqlLogger{logger: logger}, portC)
		go func() {
			err := proxy.ListenAndServe(ctx)
			if err != nil {
				errorC <- err
			}
		}()
		port := 0
		select {
		case err := <-errorC:
			return errors.Wrap(err, "error")
		case port = <-portC:
		}

		address := fmt.Sprintf("127.0.0.1:%d", port)
		addresses.Store(decl.Name, address)
		os.Setenv(strings.ToUpper("FTL_PROXY_MYSQL_ADDRESS_"+decl.Name), address)
	}
	return nil
}

var _ mysql.Logger = (*mysqlLogger)(nil)

type mysqlLogger struct {
	logger *log.Logger
}

func (m *mysqlLogger) Print(v ...any) {
	for _, s := range v {
		m.logger.Infof("mysql: %v", s)
	}
}
