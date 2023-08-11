// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	"context"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	"github.com/jpillora/backoff"
	"github.com/otiai10/copy"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TBD54566975/ftl/backend/common/download"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/plugin"
	rpc2 "github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/common/unstoppable"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type Config struct {
	Bind               *url.URL        `help:"Endpoint the Runner should bind to and advertise." default:"http://localhost:8893" env:"FTL_RUNNER_BIND"`
	Advertise          *url.URL        `help:"Endpoint the Runner should advertise (use --bind if omitted)." default:"" env:"FTL_RUNNER_ADVERTISE"`
	Key                model.RunnerKey `help:"Runner key (auto)." placeholder:"R<ULID>" default:"R00000000000000000000000000"`
	ControllerEndpoint *url.URL        `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://localhost:8892"`
	TemplateDir        string          `help:"Template directory to copy into each deployment, if any." type:"existingdir"`
	DeploymentDir      string          `help:"Directory to store deployments in." default:"${deploymentdir}" type:"existingdir"`
	Language           []string        `short:"l" help:"Languages the runner supports." env:"FTL_LANGUAGE" required:""`
	HeartbeatPeriod    time.Duration   `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter    time.Duration   `help:"Jitter to add to heartbeat period." default:"2s"`
}

func Start(ctx context.Context, config Config) error {
	if config.Advertise.String() == "" {
		config.Advertise = config.Bind
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "failed to get hostname")
	}
	pid := os.Getpid()

	client := rpc2.Dial(ftlv1connect.NewVerbServiceClient, config.ControllerEndpoint.String(), log.Error)
	ctx = rpc2.ContextWithClient(ctx, client)

	logger := log.FromContext(ctx).Sub(map[string]string{"runner": config.Key.String()})
	logger.Infof("Starting FTL Runner")
	logger.Infof("Deployment directory: %s", config.DeploymentDir)
	err = os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment directory")
	}
	logger.Infof("Using FTL endpoint: %s", config.ControllerEndpoint)
	logger.Infof("Listening on %s", config.Bind)

	controllerClient := rpc2.Dial(ftlv1connect.NewControllerServiceClient, config.ControllerEndpoint.String(), log.Error)

	key := config.Key
	if key == (model.RunnerKey{}) {
		key = model.NewRunnerKey()
	}
	labels, err := structpb.NewStruct(map[string]any{
		"hostname":  hostname,
		"pid":       pid,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"languages": slices.Map(config.Language, func(t string) any { return t }),
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal labels")
	}

	svc := &Service{
		key:              key,
		config:           config,
		controllerClient: controllerClient,
		forceUpdate:      make(chan struct{}, 16),
		labels:           labels,
		logger:           logger,
	}
	svc.state.Store(ftlv1.RunnerState_RUNNER_IDLE)

	go rpc2.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.RegisterRunner, svc.registrationLoop)
	go rpc2.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.StreamDeploymentLogs, svc.streamLogsLoop)

	return rpc2.Serve(ctx, config.Bind,
		rpc2.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc2.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
	)
}

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type deployment struct {
	key    model.DeploymentKey
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	// Cancelled when plugin terminates
	ctx      context.Context
	logger   *log.Logger
	logQueue chan log.Entry
}

type Service struct {
	key         model.RunnerKey
	lock        sync.Mutex
	state       atomic.Value[ftlv1.RunnerState]
	forceUpdate chan struct{}
	deployment  atomic.Value[types.Option[*deployment]]

	config           Config
	controllerClient ftlv1connect.ControllerServiceClient
	// Failed to register with the Controller
	registrationFailure atomic.Value[types.Option[error]]
	labels              *structpb.Struct
	logger              *log.Logger
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("no deployment"))
	}
	return deployment.plugin.Client.Call(ctx, req)
}

func (s *Service) Reserve(ctx context.Context, c *connect.Request[ftlv1.ReserveRequest]) (*connect.Response[ftlv1.ReserveResponse], error) {
	if !s.state.CompareAndSwap(ftlv1.RunnerState_RUNNER_IDLE, ftlv1.RunnerState_RUNNER_RESERVED) {
		return nil, errors.Errorf("can only reserve from IDLE state, not %s", s.state.Load())
	}
	return connect.NewResponse(&ftlv1.ReserveResponse{}), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Deploy(ctx context.Context, req *connect.Request[ftlv1.DeployRequest]) (response *connect.Response[ftlv1.DeployResponse], err error) {
	if err, ok := s.registrationFailure.Load().Get(); ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to register runner"))
	}

	key, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	logQueue := make(chan log.Entry, 100)
	sink := newDeploymentLogsSink(key, s.key, logQueue)
	logger := s.logger.AddSink(sink).Level(log.Trace)
	ctx = log.ContextWithLogger(ctx, logger)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("already deployed"))
	}

	// Set the state of the Runner, and update the Controller.
	setState := func(state ftlv1.RunnerState) {
		s.state.Store(state)
		s.forceUpdate <- struct{}{}
	}

	setState(ftlv1.RunnerState_RUNNER_RESERVED)
	defer func() {
		if err != nil {
			setState(ftlv1.RunnerState_RUNNER_IDLE)
			s.deployment.Store(types.None[*deployment]())
		}
	}()

	gdResp, err := s.controllerClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: req.Msg.DeploymentKey}))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, key.String())
	if s.config.TemplateDir != "" {
		err = copy.Copy(s.config.TemplateDir, deploymentDir)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy template directory")
		}
	} else {
		err = os.MkdirAll(deploymentDir, 0700)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create deployment directory")
		}
	}
	err = download.Artefacts(ctx, s.controllerClient, key, deploymentDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download artefacts")
	}
	deployment, cmdCtx, err := plugin.Spawn(
		unstoppable.Context(ctx),
		gdResp.Msg.Schema.Name,
		deploymentDir,
		"./main",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars(
			"FTL_ENDPOINT="+s.config.ControllerEndpoint.String(),
			"FTL_OBSERVABILITY_ENDPOINT="+s.config.ControllerEndpoint.String(),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to spawn plugin")
	}

	dep := s.makeDeployment(cmdCtx, key, logger, logQueue, deployment)
	s.deployment.Store(types.Some(dep))

	setState(ftlv1.RunnerState_RUNNER_ASSIGNED)
	return connect.NewResponse(&ftlv1.DeployResponse{}), nil
}

func (s *Service) Terminate(ctx context.Context, c *connect.Request[ftlv1.TerminateRequest]) (*connect.Response[ftlv1.RegisterRunnerRequest], error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	depl, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no deployment"))
	}
	deploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}
	if depl.key != deploymentKey {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("deployment key mismatch"))
	}

	// Soft kill.
	err = depl.plugin.Cmd.Kill(syscall.SIGTERM)
	if err != nil {
		return nil, errors.Wrap(err, "failed to kill plugin")
	}
	// Hard kill after 10 seconds.
	select {
	case <-depl.ctx.Done():
	case <-time.After(10 * time.Second):
		err := depl.plugin.Cmd.Kill(syscall.SIGKILL)
		if err != nil {
			// Should we os.Exit(1) here?
			return nil, errors.Wrap(err, "failed to kill plugin")
		}
	}
	s.deployment.Store(types.None[*deployment]())
	s.state.Store(ftlv1.RunnerState_RUNNER_IDLE)
	return connect.NewResponse(&ftlv1.RegisterRunnerRequest{
		Key:      s.key.String(),
		Endpoint: s.config.Advertise.String(),
		State:    ftlv1.RunnerState_RUNNER_IDLE,
		Labels:   s.labels,
	}), nil
}

func (s *Service) makeDeployment(ctx context.Context, key model.DeploymentKey, logger *log.Logger, logQueue chan log.Entry, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *deployment {
	return &deployment{ctx: ctx, key: key, logger: logger, plugin: plugin, logQueue: logQueue}
}

func (s *Service) registrationLoop(ctx context.Context, send func(request *ftlv1.RegisterRunnerRequest) error) error {
	logger := s.getLogger()

	// Figure out the appropriate state.
	state := s.state.Load()
	var errPtr *string
	var deploymentKey *string
	depl, ok := s.deployment.Load().Get()
	if ok {
		dkey := depl.key.String()
		deploymentKey = &dkey
		select {
		case <-depl.ctx.Done():
			state = ftlv1.RunnerState_RUNNER_IDLE
			err := context.Cause(depl.ctx)
			errStr := err.Error()
			errPtr = &errStr
			logger.Errorf(err, "Deployment terminated")
			s.deployment.Store(types.None[*deployment]())

		default:
			state = ftlv1.RunnerState_RUNNER_ASSIGNED
		}
		s.state.Store(state)
	}

	logger.Tracef("Registering with Controller as %s", state)
	err := send(&ftlv1.RegisterRunnerRequest{
		Key:        s.key.String(),
		Endpoint:   s.config.Advertise.String(),
		Labels:     s.labels,
		Deployment: deploymentKey,
		State:      state,
		Error:      errPtr,
	})
	if err != nil {
		logger.Errorf(err, "failed to register with Controller")
		s.registrationFailure.Store(types.Some(err))
		return errors.WithStack(err)
	}
	s.registrationFailure.Store(types.None[error]())

	// Wait for the next heartbeat.
	delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
	logger.Tracef("Registered with Controller, next heartbeat in %s", delay)
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
		s.registrationFailure.Store(types.Some(err))
		return errors.WithStack(err)

	case <-s.forceUpdate:

	case <-time.After(delay):
	}
	return nil
}

func (s *Service) streamLogsLoop(ctx context.Context, send func(request *ftlv1.StreamDeploymentLogsRequest) error) error {
	delay := time.Millisecond * 500

	deployment, ok := s.deployment.Load().Get()
	if !ok {
		// No deployment, nothing to do.
		select {
		case <-ctx.Done():
			err := context.Cause(ctx)
			return errors.WithStack(err)
		case <-time.After(delay):
			return nil
		}
	}

	select {
	case entry := <-deployment.logQueue:
		var errorString string
		if entry.Error != nil {
			errorString = entry.Error.Error()
		}
		err := send(&ftlv1.StreamDeploymentLogsRequest{
			DeploymentKey: deployment.key.String(),
			RunnerKey:     s.key.String(),
			TimeStamp:     entry.Time.Unix(),
			LogLevel:      int32(entry.Level.Severity()),
			Attributes:    entry.Attributes,
			Message:       entry.Message,
			Error:         &errorString,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	case <-time.After(delay):
	}

	return nil
}

func (s *Service) getLogger() *log.Logger {
	deployment, ok := s.deployment.Load().Get()
	if ok {
		if deployment.logger == nil {
			return deployment.logger
		}
	}

	return s.logger
}
