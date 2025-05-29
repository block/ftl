package localscaling

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/backend/runner"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/artefacts"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/localdebug"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
)

var _ scaling.RunnerScaling = &localScaling{}

type localScaling struct {
	runnerContext context.Context

	lock     sync.Mutex
	cacheDir string
	// Deployments -> info
	runners map[string]*deploymentInfo
	// Module -> Port
	debugPorts map[string]*localdebug.DebugInfo
	// Module -> runner sequence
	runnerCounts  map[string]int64
	leaseAddress  *url.URL
	schemaAddress *url.URL

	prevRunnerSuffix int
	ideSupport       optional.Option[localdebug.IDEIntegration]
	storage          *artefacts.OCIArtefactService
	enableOtel       bool

	devModeEndpointsUpdates <-chan dev.LocalEndpoint
	devModeEndpoints        map[string]*devModeRunner

	LogConfig    log.Config
	routeTable   *routing.RouteTable
	schemaClient ftlv1connect.SchemaServiceClient
	adminClient  adminpbconnect.AdminServiceClient
}

func (l *localScaling) StartDeployment(ctx context.Context, deployment string, sch *schema.Module, hasCron bool, hasIngress bool) (url.URL, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	deploymentKey, err := key.ParseDeploymentKey(deployment)
	if err != nil {
		return url.URL{}, errors.Wrap(err, "failed to parse deployment key")
	}

	logger := log.FromContext(l.runnerContext).Scope("localScaling").Module(deploymentKey.Payload.Module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Starting deployment for %s", deployment)
	dep := &deploymentInfo{runner: optional.None[runnerInfo](), key: deploymentKey, language: sch.Runtime.Base.Language}
	l.runners[deployment] = dep
	//  Make sure we have all endpoint updates
	for {
		select {
		case devEndpoints := <-l.devModeEndpointsUpdates:
			l.updateDevModeEndpoint(ctx, devEndpoints)
			continue
		default:
		}
		break
	}

	if err := l.startRunner(ctx, dep.key, dep, sch); err != nil {
		logger.Errorf(err, "Failed to start runner")
		return url.URL{}, errors.WithStack(err)
	}

	if r, ok := dep.runner.Get(); ok {
		return r.getURL(), nil
	}
	return url.URL{}, errors.Errorf("runner not found")
}

func (l *localScaling) UpdateDeployment(ctx context.Context, deployment string, sch *schema.Module) error {
	// NOOP for local
	return nil
}

func (l *localScaling) TerminateDeployment(ctx context.Context, deployment string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	logger := log.FromContext(ctx)
	logger.Debugf("Terminating previous deployments for %s", deployment)
	dep := l.runners[deployment]
	if dep == nil {
		return nil
	}
	if r, ok := dep.runner.Get(); ok {
		r.cancelFunc(errors.Wrap(context.Canceled, "deployment terminated"))
	}
	delete(l.runners, deployment)
	return nil
}

type devModeRunner struct {
	uri          string
	hotReloadURI string
	// The deployment key of the deployment that is currently running
	deploymentKey optional.Option[key.Deployment]
	debugPort     int
	scehamVersion int64
}

func (l *localScaling) Start(ctx context.Context) error {
	go func() {
		for devEndpoints := range channels.IterContext(ctx, l.devModeEndpointsUpdates) {
			l.lock.Lock()
			l.updateDevModeEndpoint(ctx, devEndpoints)
			l.lock.Unlock()
		}
	}()
	return nil
}

// updateDevModeEndpoint updates the dev mode endpoint for a module
// Must be called under lock
func (l *localScaling) updateDevModeEndpoint(ctx context.Context, devEndpoints dev.LocalEndpoint) {
	l.devModeEndpoints[devEndpoints.Module] = &devModeRunner{
		uri:           devEndpoints.Endpoint,
		debugPort:     devEndpoints.DebugPort,
		hotReloadURI:  devEndpoints.HotReloadEndpoint,
		scehamVersion: devEndpoints.Version,
	}
	if ide, ok := l.ideSupport.Get(); ok {
		if devEndpoints.DebugPort != 0 {
			if debug, ok := l.debugPorts[devEndpoints.Module]; ok {
				debug.Port = devEndpoints.DebugPort
			} else {
				l.debugPorts[devEndpoints.Module] = &localdebug.DebugInfo{
					Port:     devEndpoints.DebugPort,
					Language: devEndpoints.Language,
				}
			}
		}
		ide.SyncIDEDebugIntegrations(ctx, l.debugPorts)

	}
}

type deploymentInfo struct {
	runner   optional.Option[runnerInfo]
	key      key.Deployment
	language string
}
type runnerInfo struct {
	cancelFunc context.CancelCauseFunc
	port       int
	host       string
}

func (r runnerInfo) getURL() url.URL {
	return url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", r.host, r.port),
	}
}

func NewLocalScaling(
	ctx context.Context,
	controllerAddresse *url.URL,
	schemaAddress *url.URL,
	leaseAddress *url.URL,
	configPath string,
	enableVSCodeIntegration bool,
	enableIntellijIntegration bool,
	storage *artefacts.OCIArtefactService,
	enableOtel bool,
	devModeEndpoints <-chan dev.LocalEndpoint,
	routeTable *routing.RouteTable,
	schemaClient ftlv1connect.SchemaServiceClient,
	adminClient adminpbconnect.AdminServiceClient,
) (scaling.RunnerScaling, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	local := localScaling{
		runnerContext:           ctx,
		lock:                    sync.Mutex{},
		cacheDir:                cacheDir,
		runners:                 map[string]*deploymentInfo{},
		leaseAddress:            leaseAddress,
		schemaAddress:           schemaAddress,
		prevRunnerSuffix:        -1,
		debugPorts:              map[string]*localdebug.DebugInfo{},
		storage:                 storage,
		enableOtel:              enableOtel,
		devModeEndpointsUpdates: devModeEndpoints,
		devModeEndpoints:        map[string]*devModeRunner{},
		runnerCounts:            map[string]int64{},
		routeTable:              routeTable,
		schemaClient:            schemaClient,
		adminClient:             adminClient,
	}
	if configPath != "" {
		local.ideSupport = optional.Ptr(localdebug.NewIDEIntegration(configPath, enableVSCodeIntegration, enableIntellijIntegration))
	}

	return &local, nil
}

func (l *localScaling) startRunner(ctx context.Context, deploymentKey key.Deployment, info *deploymentInfo, sch *schema.Module) error {
	logger := log.FromContext(ctx)
	select {
	case <-ctx.Done():
		// In some cases this gets called with an expired context, generally after the lease is released
		// We don't want to start a runner in that case
		return nil
	default:
	}

	var deploymentProvider artefacts.DeploymentArtefactProvider

	module := info.key.Payload.Module
	devEndpoint := l.devModeEndpoints[module]
	devURI := optional.None[string]()
	devHotReloadURI := optional.None[string]()
	debugPort := 0
	var runnerSeq int64
	var schemaSeq int64
	if devEndpoint != nil {
		devURI = optional.Some(devEndpoint.uri)
		devHotReloadURI = optional.Some(devEndpoint.hotReloadURI)
		if devKey, ok := devEndpoint.deploymentKey.Get(); ok && devKey.Equal(deploymentKey) {
			// Already running, don't start another
			return nil
		}
		devEndpoint.deploymentKey = optional.Some(deploymentKey)
		debugPort = devEndpoint.debugPort
		runnerSeq = l.runnerCounts[deploymentKey.Payload.Module]
		l.runnerCounts[deploymentKey.Payload.Module] = runnerSeq + 1
		schemaSeq = devEndpoint.scehamVersion
		logger.Debugf("Starting runner with schema version %d and runner sequence %d", schemaSeq, runnerSeq)
		deploymentProvider = func() (string, error) {
			return "", nil
		}
	} else if ide, ok := l.ideSupport.Get(); ok {
		var debug *localdebug.DebugInfo
		debugBind, err := plugin.AllocatePort()
		if err != nil {
			return errors.Wrap(err, "failed to start runner")
		}
		debug = &localdebug.DebugInfo{
			Language: info.language,
			Port:     debugBind.Port,
		}
		l.debugPorts[module] = debug
		ide.SyncIDEDebugIntegrations(ctx, l.debugPorts)
		debugPort = debug.Port

		deploymentProvider = func() (string, error) {

			deploymentDir := filepath.Join(l.cacheDir, deploymentKey.String())
			err = download.ArtefactsFromOCI(ctx, l.schemaClient, deploymentKey, deploymentDir, l.storage)
			if err != nil {
				return "", errors.Wrapf(err, "failed to download artifacts")
			}
			return deploymentDir, nil
		}
		// Download the required artifacts
	}

	bind, err := plugin.AllocatePort()
	if err != nil {
		return errors.Wrap(err, "failed to start runner")
	}

	keySuffix := l.prevRunnerSuffix + 1
	l.prevRunnerSuffix = keySuffix

	bindURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", bind.Port))
	if err != nil {
		return errors.Wrap(err, "failed to start runner")
	}
	config := runner.Config{
		Bind:                  bindURL,
		LeaseEndpoint:         l.leaseAddress,
		Key:                   key.NewLocalRunnerKey(keySuffix),
		Deployment:            deploymentKey,
		DebugPort:             debugPort,
		DevEndpoint:           devURI,
		DevHotReloadEndpoint:  devHotReloadURI,
		LocalRunners:          true,
		DevModeSchemaSequence: schemaSeq,
		DevModeRunnerSequence: runnerSeq,
	}

	simpleName := fmt.Sprintf("runner%d", keySuffix)
	if err := kong.ApplyDefaults(&config, kong.Vars{
		"deploymentdir": filepath.Join(l.cacheDir, "ftl-runner", simpleName, "deployments"),
		// TODO: This doesn't seem like it should be here.
		"language": "go,kotlin,java",
	}); err != nil {
		return errors.Wrap(err, "failed to apply defaults")
	}
	config.HeartbeatPeriod = time.Second
	config.HeartbeatJitter = time.Millisecond * 100

	runnerCtx := log.ContextWithLogger(ctx, logger.Scope(simpleName).Module(module))

	runnerCtx, cancel := context.WithCancelCause(runnerCtx)
	info.runner = optional.Some(runnerInfo{cancelFunc: cancel, port: bind.Port, host: "127.0.0.1"})

	dcproc, err := deploymentcontext.NewAdminProvider(ctx, info.key, l.routeTable, sch, l.adminClient)
	if err != nil {
		return errors.Wrapf(err, "Failed to create deployment context provider")
	}

	go func() {
		err := runner.Start(runnerCtx, config, deploymentProvider, dcproc, sch)
		cancel(errors.Wrap(err, "runner exited"))
		l.lock.Lock()
		defer l.lock.Unlock()
		if devEndpoint != nil {
			// Runner is complete, clear the deployment key
			devEndpoint.deploymentKey = optional.None[key.Deployment]()
		}
		info.runner = optional.None[runnerInfo]()
	}()
	if devEndpoint != nil {
		// We know this is already running
		return nil
	}
	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, bindURL.String(), log.Error)
	timeout := time.After(1 * time.Minute)
	for {
		select {
		case <-runnerCtx.Done():
			return errors.WithStack(context.Cause(runnerCtx))
		case <-timeout:
			return errors.Errorf("timed out waiting for runner to be ready")
		case <-time.After(time.Millisecond * 100):
			_, err := client.Ping(runnerCtx, connect.NewRequest(&ftlv1.PingRequest{}))
			if err == nil {
				return nil
			}
		}
	}
}
