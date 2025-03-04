// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"github.com/otiai10/copy"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"

	mysql "github.com/block/ftl-mysql-auth-proxy"
	"github.com/block/ftl/backend/controller/artefacts"
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
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	ftlobservability "github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/pgproxy"
	"github.com/block/ftl/internal/rpc"
	timeline "github.com/block/ftl/internal/timelineclient"
	"github.com/block/ftl/internal/unstoppable"
)

type Config struct {
	Config                []string                `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
	Bind                  *url.URL                `help:"Endpoint the Runner should bind to and advertise." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Key                   key.Runner              `help:"Runner key (auto)."`
	ControllerEndpoint    *url.URL                `name:"ftl-controller-endpoint" help:"Controller endpoint." env:"FTL_CONTROLLER_ENDPOINT" default:"http://127.0.0.1:8893"`
	SchemaEndpoint        *url.URL                `name:"schema-endpoint" help:"Schema server endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	LeaseEndpoint         *url.URL                `name:"ftl-lease-endpoint" help:"Lease endpoint endpoint." env:"FTL_LEASE_ENDPOINT" default:"http://127.0.0.1:8895"`
	QueryEndpoint         *url.URL                `name:"ftl-query-endpoint" help:"Query endpoint." env:"FTL_QUERY_ENDPOINT" default:"http://127.0.0.1:8897"`
	TimelineEndpoint      *url.URL                `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	TemplateDir           string                  `help:"Template directory to copy into each deployment, if any." type:"existingdir"`
	DeploymentDir         string                  `help:"Directory to store deployments in." default:"${deploymentdir}"`
	DeploymentKeepHistory int                     `help:"Number of deployments to keep history for." default:"3"`
	HeartbeatPeriod       time.Duration           `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter       time.Duration           `help:"Jitter to add to heartbeat period." default:"2s"`
	Deployment            key.Deployment          `help:"The deployment this runner is for." env:"FTL_DEPLOYMENT"`
	DebugPort             int                     `help:"The port to use for debugging." env:"FTL_DEBUG_PORT"`
	DevEndpoint           optional.Option[string] `help:"An existing endpoint to connect to in development mode" hidden:""`
	DevHotReloadEndpoint  optional.Option[string] `help:"The gRPC enpoint to send runner into to for hot reload." hidden:""`
}

func Start(ctx context.Context, config Config, storage *artefacts.OCIArtefactService) error {
	ctx, doneFunc := context.WithCancelCause(ctx)
	defer doneFunc(fmt.Errorf("runner terminated: %w", context.Canceled))
	hostname, err := os.Hostname()
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return fmt.Errorf("failed to get hostname: %w", err)
	}
	pid := os.Getpid()

	runnerKey := config.Key
	if runnerKey.IsZero() {
		runnerKey = key.NewRunnerKey(config.Bind.Hostname(), config.Bind.Port())
	}

	logger := log.FromContext(ctx).Attrs(map[string]string{"runner": runnerKey.String()})
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Starting FTL Runner")

	err = manageDeploymentDirectory(logger, config)
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return err
	}

	logger.Debugf("Using FTL endpoint: %s", config.ControllerEndpoint)
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Using Schema endpoint %s", config.SchemaEndpoint.String())

	controllerClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, config.ControllerEndpoint.String(), log.Error)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, config.SchemaEndpoint.String(), log.Error)

	labels, err := structpb.NewStruct(map[string]any{
		"hostname": hostname,
		"pid":      pid,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	})
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return fmt.Errorf("failed to marshal labels: %w", err)
	}
	timelineClient := timeline.NewClient(ctx, config.TimelineEndpoint)

	svc := &Service{
		key:                  runnerKey,
		config:               config,
		storage:              storage,
		controllerClient:     controllerClient,
		schemaClient:         schemaClient,
		timelineClient:       timelineClient,
		labels:               labels,
		deploymentLogQueue:   make(chan log.Entry, 10000),
		cancelFunc:           doneFunc,
		devEndpoint:          config.DevEndpoint,
		devHotReloadEndpoint: config.DevHotReloadEndpoint,
		queryServices:        query.NewMultiService(),
	}

	module, err := svc.getModule(ctx, config.Deployment)
	if err != nil {
		return fmt.Errorf("failed to get module: %w", err)
	}

	startedLatch := &sync.WaitGroup{}
	startedLatch.Add(2)
	g, ctx := errgroup.WithContext(ctx)
	dbAddresses := xsync.NewMapOf[string, string]()
	g.Go(func() error {
		return svc.startPgProxy(ctx, module, startedLatch, dbAddresses)
	})
	g.Go(func() error {
		return svc.startMySQLProxy(ctx, module, startedLatch, dbAddresses)
	})
	g.Go(func() error {
		startedLatch.Wait()
		g.Go(func() error {
			return svc.startQueryServices(ctx, module, dbAddresses)
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return svc.startDeployment(ctx, config.Deployment, module, dbAddresses)
		}
	})

	return fmt.Errorf("failure in runner: %w", g.Wait())
}

func (s *Service) startDeployment(ctx context.Context, key key.Deployment, module *schema.Module, dbAddresses *xsync.MapOf[string, string]) error {
	err := s.deploy(ctx, key, module, dbAddresses)
	if err != nil {
		// If we fail to deploy we just exit
		// Kube or local scaling will start a new instance to continue
		// This approach means we don't have to handle error states internally
		// It is managed externally by the scaling system
		return err
	}
	go s.streamLogsLoop(ctx)
	go func() {
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, s.controllerClient.RegisterRunner, s.registrationLoop)
	}()
	return fmt.Errorf("failure in runner: %w", rpc.Serve(ctx, s.config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s),
		rpc.GRPC(pubsubpbconnect.NewPubSubAdminServiceHandler, s.pubSub),
		rpc.HTTP("/", s),
		rpc.HealthCheck(s.healthCheck),
	))
}

// manageDeploymentDirectory ensures the deployment directory exists and removes old deployments.
func manageDeploymentDirectory(logger *log.Logger, config Config) error {
	logger.Debugf("Deployment directory: %s", config.DeploymentDir)
	err := os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Clean up old deployments.
	modules, err := os.ReadDir(config.DeploymentDir)
	if err != nil {
		return fmt.Errorf("failed to read deployment directory: %w", err)
	}

	for _, module := range modules {
		if !module.IsDir() {
			continue
		}

		moduleDir := filepath.Join(config.DeploymentDir, module.Name())
		deployments, err := os.ReadDir(moduleDir)
		if err != nil {
			return fmt.Errorf("failed to read module directory: %w", err)
		}

		if len(deployments) < config.DeploymentKeepHistory {
			continue
		}

		stats, err := slices.MapErr(deployments, func(d os.DirEntry) (os.FileInfo, error) {
			return d.Info()
		})
		if err != nil {
			return fmt.Errorf("failed to stat deployments: %w", err)
		}

		// Sort deployments by modified time, remove anything past the history limit.
		sort.Slice(deployments, func(i, j int) bool {
			return stats[i].ModTime().After(stats[j].ModTime())
		})

		for _, deployment := range deployments[config.DeploymentKeepHistory:] {
			old := filepath.Join(moduleDir, deployment.Name())
			logger.Debugf("Removing old deployment: %s", old)

			err := os.RemoveAll(old)
			if err != nil {
				// This is not a fatal error, just log it.
				logger.Errorf(err, "Failed to remove old deployment: %s", deployment.Name())
			}
		}
	}

	return nil
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
	storage          *artefacts.OCIArtefactService
	controllerClient ftlv1connect.ControllerServiceClient
	schemaClient     ftlv1connect.SchemaServiceClient
	timelineClient   *timeline.Client
	// Failed to register with the Controller
	registrationFailure  atomic.Value[optional.Option[error]]
	labels               *structpb.Struct
	deploymentLogQueue   chan log.Entry
	cancelFunc           context.CancelCauseFunc
	devEndpoint          optional.Option[string]
	devHotReloadEndpoint optional.Option[string]
	proxy                *proxy.Service
	pubSub               *pubsub.Service
	queryServices        *query.MultiService
	proxyBindAddress     *url.URL
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("no deployment"))
	}
	response, err := deployment.client.Call(ctx, req)
	if err != nil {
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(err, "Call to deployments %s failed to perform gRPC call", deployment.key)
		return nil, connect.NewError(connect.CodeOf(err), err)
	} else if response.Msg.GetError() != nil {
		// This is a user level error (i.e. something wrong in the users app)
		// Log it to the deployment logger
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(fmt.Errorf("%v", response.Msg.GetError().GetMessage()), "Call to deployments %s failed", deployment.key)
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
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	module, err := schema.ValidatedModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid module: %w", err)
	}
	return module, nil
}

func (s *Service) deploy(ctx context.Context, key key.Deployment, module *schema.Module, dbAddresses *xsync.MapOf[string, string]) error {
	logger := log.FromContext(ctx)

	if err, ok := s.registrationFailure.Load().Get(); ok {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to register runner: %w", err))
	}

	observability.Deployment.Started(ctx, key.String())
	defer observability.Deployment.Completed(ctx, key.String())

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.New("already deployed")
	}

	deploymentLogger := s.getDeploymentLogger(ctx, key)
	ctx = log.ContextWithLogger(ctx, deploymentLogger)

	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, key.String())
	if s.config.TemplateDir != "" {
		err := copy.Copy(s.config.TemplateDir, deploymentDir)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to copy template directory: %w", err)
		}
	} else {
		err := os.MkdirAll(deploymentDir, 0700)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to create deployment directory: %w", err)
		}
	}

	leaseServiceClient := rpc.Dial(ftlleaseconnect.NewLeaseServiceClient, s.config.LeaseEndpoint.String(), log.Error)

	s.proxy = proxy.New(s.controllerClient, leaseServiceClient, s.timelineClient, s.queryServices,
		s.config.Bind.String(), s.config.Deployment)

	pubSub, err := pubsub.New(module, key, s, s.timelineClient)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to set up pubsub: %w", err)
	}
	s.pubSub = pubSub

	parse, err := url.Parse("http://127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}
	proxyServer, err := rpc.NewServer(ctx, parse,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s.proxy),
		rpc.GRPC(ftlv1connect.NewControllerServiceHandler, s.proxy),
		rpc.GRPC(ftlleaseconnect.NewLeaseServiceHandler, s.proxy),
		rpc.GRPC(pubsubpbconnect.NewPublishServiceHandler, s.pubSub),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
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
	proxyPingClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, s.proxyBindAddress.String(), log.Error)
	err = rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, proxyPingClient)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to ping proxy: %w", err)
	}
	var dep *deployment
	if ep, ok := s.devEndpoint.Get(); ok {
		pingCtx, cancel := context.WithCancelCause(ctx)
		defer cancel(fmt.Errorf("complete"))
		if hotRelaodEp, ok := s.devHotReloadEndpoint.Get(); ok {
			hotReloadClient := rpc.Dial(hotreloadpbconnect.NewHotReloadServiceClient, hotRelaodEp, log.Error)
			err = rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, hotReloadClient)
			if err != nil {
				return fmt.Errorf("failed to ping hot reload endpoint: %w", err)
			}
			var databases []*hotreloadpb.Database
			dbAddresses.Range(func(key string, value string) bool {
				databases = append(databases, &hotreloadpb.Database{
					Name:    key,
					Address: value,
				})
				return true
			})
			go func() {
				for {
					select {
					case <-pingCtx.Done():
						return
					case <-time.After(time.Millisecond * 50):
						info, err := hotReloadClient.RunnerInfo(pingCtx, connect.NewRequest(&hotreloadpb.RunnerInfoRequest{
							Deployment: s.config.Deployment.String(),
							Address:    s.proxyBindAddress.String(),
							Databases:  databases,
						}))
						if err == nil {
							if info.Msg.Outdated {
								logger.Warnf("Runner is outdated, exiting")
								cancel(fmt.Errorf("runner is outdated, exiting"))
							}
						}
					}
				}
			}()

		}
		client := rpc.Dial(ftlv1connect.NewVerbServiceClient, ep, log.Error)
		err = rpc.Wait(pingCtx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute, client)
		dep = &deployment{
			ctx:      ctx,
			key:      key,
			cmd:      optional.None[exec.Cmd](),
			endpoint: ep,
			client:   client,
		}
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to ping dev endpoint: %w", err)
		}
	} else {
		err := download.ArtefactsFromOCI(ctx, s.schemaClient, key, deploymentDir, s.storage)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to download artefacts: %w", err)
		}

		logger.Debugf("Setting FTL_CONTROLLER_ENDPOINT to %s", s.proxyBindAddress.String())
		envVars := []string{"FTL_CONTROLLER_ENDPOINT=" + s.proxyBindAddress.String(),
			"FTL_CONFIG=" + strings.Join(s.config.Config, ","),
			"FTL_DEPLOYMENT=" + s.config.Deployment.String(),
			"FTL_OBSERVABILITY_ENDPOINT=" + s.config.ControllerEndpoint.String()}
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
			plugin.WithEnvars(
				envVars...,
			),
		)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to spawn plugin: %w", err)
		}
		dep = s.makeDeployment(cmdCtx, key, deployment)
	}

	s.readyTime.Store(time.Now().Add(time.Second * 2)) // Istio is a bit flakey, add a small delay for readiness
	s.deployment.Store(optional.Some(dep))
	logger.Debugf("Deployed %s", key)

	err = s.pubSub.Consume(ctx)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to set up pubsub consumption: %w", err)
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

	if s.queryServices != nil {
		s.queryServices.Close()
	}

	depl, ok := s.deployment.Load().Get()
	if !ok {
		return connect.NewError(connect.CodeNotFound, errors.New("no deployment"))
	}
	if cmd, ok := depl.cmd.Get(); ok {
		// Soft kill.
		err := cmd.Kill(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("failed to kill plugin: %w", err)
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
					return fmt.Errorf("failed to kill plugin: %w", err)
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

func (s *Service) registrationLoop(ctx context.Context, send func(request *ftlv1.RegisterRunnerRequest) error) error {
	logger := log.FromContext(ctx)

	// Figure out the appropriate state.
	var deploymentKey *string
	depl, ok := s.deployment.Load().Get()
	if ok {
		dkey := depl.key.String()
		deploymentKey = &dkey
		select {
		case <-depl.ctx.Done():
			err := context.Cause(depl.ctx)
			s.getDeploymentLogger(ctx, depl.key).Errorf(err, "Deployment terminated")
			s.deployment.Store(optional.None[*deployment]())
			s.cancelFunc(fmt.Errorf("deployment %s terminated: %w", depl.key, err))
			return nil
		default:
		}
	}

	logger.Tracef("Registering with Controller for deployment %s", s.config.Deployment)
	err := send(&ftlv1.RegisterRunnerRequest{
		Key:        s.key.String(),
		Endpoint:   s.config.Bind.String(),
		Labels:     s.labels,
		Deployment: s.config.Deployment.String(),
	})
	if err != nil {
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey))
		return fmt.Errorf("failed to register with Controller: %w", err)
	}

	// Wait for the next heartbeat.
	delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
	logger.Tracef("Registered with Controller, next heartbeat in %s", delay)
	observability.Runner.Registered(ctx, optional.Ptr(deploymentKey))
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey))
		return err

	case <-time.After(delay):
	}
	return nil
}

func (s *Service) streamLogsLoop(ctx context.Context) {
	for entry := range channels.IterContext(ctx, s.deploymentLogQueue) {
		dep, ok := entry.Attributes["deployment"]
		var deploymentKey key.Deployment
		var err error
		if ok {
			deploymentKey, err = key.ParseDeploymentKey(dep)
			if err != nil {
				continue
			}
		}

		var errorString *string
		if entry.Error != nil {
			errStr := entry.Error.Error()
			errorString = &errStr
		}
		var request optional.Option[key.Request]
		if reqStr, ok := entry.Attributes["request"]; ok {
			req, err := key.ParseRequestKey(reqStr) //nolint:errcheck // best effort
			if err == nil {
				request = optional.Some(req)
			}
		}
		s.timelineClient.Publish(ctx, &timeline.Log{
			DeploymentKey: deploymentKey,
			RequestKey:    request,
			Level:         int32(entry.Level),
			Time:          entry.Time,
			Attributes:    entry.Attributes,
			Message:       entry.Message,
			Error:         optional.Ptr(errorString),
		})
	}
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentKey key.Deployment) *log.Logger {
	attrs := map[string]string{"deployment": deploymentKey.String()}
	if requestKey, _ := rpc.RequestKeyFromContext(ctx); requestKey.Ok() { //nolint:errcheck // best effort
		attrs["request"] = requestKey.MustGet().String()
	}
	ctx = ftlobservability.AddSpanContextToLogger(ctx)

	sink := newDeploymentLogsSink(s.deploymentLogQueue)
	return log.FromContext(ctx).AddSink(sink).Attrs(attrs)
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
			return "", fmt.Errorf("database %s not found", params["database"])
		}

		dsn, err := dsn.ResolvePostgresDSN(ctx, db.Runtime.Connections.Write)
		if err != nil {
			return "", fmt.Errorf("failed to resolve postgres DSN: %w", err)
		}

		return dsn, nil
	}).Start(ctx, channel); err != nil {
		started.Done()
		return fmt.Errorf("failed to start pgproxy: %w", err)
	}

	return nil
}

// Run query services for each database
func (s *Service) startQueryServices(ctx context.Context, module *schema.Module, addresses *xsync.MapOf[string, string]) error {
	logger := log.FromContext(ctx)

	g, ctx := errgroup.WithContext(ctx)

	for _, decl := range module.Decls {
		db, ok := decl.(*schema.Database)
		if !ok {
			continue
		}

		logger.Debugf("Initializing query service for %s database: %s.%s", db.Type, module.Name, db.Name)
		dbadr, ok := addresses.Load(db.Name)
		if !ok {
			return fmt.Errorf("could not find DSN for DB %s", db.Name)
		}
		parts := strings.Split(dbadr, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("failed to parse port: %w", err)
		}
		host := parts[0]
		var sdsn string

		switch db.Type {
		case schema.MySQLDatabaseType:
			sdsn = dsn.MySQLDSN(db.Name, dsn.Host(host), dsn.Port(port))
		case schema.PostgresDatabaseType:
			sdsn = dsn.PostgresDSN(db.Name, dsn.Host(host), dsn.Port(port))
		default:
			return fmt.Errorf("unsupported database type: %T", db.Type)
		}

		baseURL, err := url.Parse("http://127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to parse base URL: %w", err)
		}

		querySvc, err := query.New(ctx, query.Config{
			DSN:      sdsn,
			Engine:   db.Type,
			Endpoint: baseURL,
		})
		if err != nil {
			return fmt.Errorf("failed to create query service for database %s: %w", db.Name, err)
		}

		server, err := rpc.NewServer(ctx, baseURL,
			rpc.GRPC(querypbconnect.NewQueryServiceHandler, querySvc),
		)
		if err != nil {
			querySvc.Close()
			return fmt.Errorf("failed to create server for database %s: %w", db.Name, err)
		}

		urls := server.Bind.Subscribe(nil)
		dbName := db.Name
		g.Go(func() error {
			serverGroup, serverCtx := errgroup.WithContext(ctx)

			serverGroup.Go(func() error {
				if err := server.Serve(serverCtx); err != nil && !errors.Is(err, context.Canceled) {
					return fmt.Errorf("query service server for database %s stopped unexpectedly: %w", dbName, err)
				}
				return nil
			})

			select {
			case <-serverCtx.Done():
				return serverCtx.Err()
			case endpoint := <-urls:
				client := rpc.Dial(querypbconnect.NewQueryServiceClient, endpoint.String(), log.Error)
				if err := rpc.Wait(serverCtx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Second*5, client); err != nil {
					return fmt.Errorf("query service for database %s failed health check: %w", dbName, err)
				}
				querySvc.SetEndpoint(endpoint)
				if err := s.queryServices.AddService(module.Name, dbName, querySvc); err != nil {
					return fmt.Errorf("failed to add query service for database %s: %w", dbName, err)
				}
				return serverGroup.Wait()
			}
		})
	}

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("failed to start query services: %w", err)
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
				return nil, fmt.Errorf("failed to resolve MySQL DSN: %w", err)
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
			return fmt.Errorf("error: %w", err)
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
