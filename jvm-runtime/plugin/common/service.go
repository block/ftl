package common

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/beevik/etree"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl"
	hotreloadpb "github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1/hotreloadpbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/plugin"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute
const SchemaFile = "schema.pb"
const ErrorFile = "errors.pb"

var ErrInvalidateDependencies = errors.New("dependencies need to be updated")

type buildContextUpdatedEvent struct {
	buildCtx      buildContext
	schemaChanged bool
}

// buildContext contains contextual information needed to build.
type buildContext struct {
	ID           string
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, errors.Wrap(err, "could not parse schema from proto")
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	return buildContext{
		ID:           proto.Id,
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
	}, nil
}

type Service struct {
	updatesTopic          *pubsub.Topic[buildContextUpdatedEvent]
	acceptsContextUpdates atomic.Value[bool]
	buildContext          atomic.Value[buildContext]
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New() *Service {
	return &Service{
		updatesTopic: pubsub.New[buildContextUpdatedEvent](),
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {

	if req.Msg.Schema == nil {
		return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
	}
	sch, err := schema.FromProto(req.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema from proto")
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	_, err = s.writeGenericSchemaFiles(ctx, sch, config, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
}

// Build the module and stream back build events.
//
// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
// end of the build.
//
// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
// calls.
func (s *Service) Build(ctx context.Context, req *connect.Request[langpb.BuildRequest], stream *connect.ServerStream[langpb.BuildResponse]) error {
	logger := log.FromContext(ctx)
	_ = os.Setenv("QUARKUS_ANALYTICS_DISABLED", "true") //nolint:errcheck
	logger = logger.Module(req.Msg.BuildContext.ModuleConfig.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return errors.WithStack(err)
	}
	s.buildContext.Store(buildCtx)
	changed, err := s.writeGenericSchemaFiles(ctx, buildCtx.Schema, buildCtx.Config, true)
	if err != nil {
		return errors.Wrap(err, "failed to write generic schema files")
	}
	if s.acceptsContextUpdates.Load() {
		// Already running in dev mode, we don't need to rebuild
		s.updatesTopic.Publish(buildContextUpdatedEvent{buildCtx: buildCtx, schemaChanged: changed})
		return nil
	}

	projectConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	if req.Msg.RebuildAutomatically {
		return s.runDevMode(ctx, projectConfig, buildCtx, stream)
	}

	// Initial build
	if err := buildAndSend(ctx, stream, projectConfig, buildCtx, false); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Service) runDevMode(ctx context.Context, projectConfig projectconfig.Config, buildCtx buildContext, stream *connect.ServerStream[langpb.BuildResponse]) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "stopping JVM language plugin (devw"))

	s.acceptsContextUpdates.Store(true)
	defer s.acceptsContextUpdates.Store(false)

	watchPatterns, err := relativeWatchPatterns(buildCtx.Config.Dir, buildCtx.Config.Watch)
	if err != nil {
		return errors.WithStack(err)
	}
	ensureCorrectFTLVersion(ctx, buildCtx)
	watcher := watch.NewWatcher(optional.None[string](), watchPatterns...)
	fileEvents := make(chan watch.WatchEventModuleChanged, 32)
	ensureCorrectFTLVersion(ctx, buildCtx)
	if err := watchFiles(ctx, watcher, buildCtx, fileEvents); err != nil {
		return errors.WithStack(err)
	}

	firstResponseSent := &atomic.Value[bool]{}
	firstResponseSent.Store(false)
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context cancelled")
		default:

		}
		if firstResponseSent.Load() {
			err := stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_AutoRebuildStarted{AutoRebuildStarted: &langpb.AutoRebuildStarted{ContextId: buildCtx.ID}}})
			if err != nil {
				logger.Errorf(err, "Could not send build event")
				return errors.Wrap(err, "could not sent build event on new process launch")
			}
		}

		err := s.runQuarkusDev(ctx, projectConfig, buildCtx.Config.Realm, buildCtx.Config.Module, stream, firstResponseSent, fileEvents, cancel)
		if err != nil {
			logger.Errorf(err, "Dev mode process exited")
		}
		id := s.buildContext.Load().ID
		if !s.waitForFileChanges(ctx, fileEvents) {
			return nil
		}
		if id != s.buildContext.Load().ID {
			// The build context was updated, we need to mark this as an explicit build
			firstResponseSent.Store(false)
		}
	}
}

func relativeWatchPatterns(moduleDir string, watchPaths []string) ([]string, error) {
	relativePaths := make([]string, len(watchPaths))
	for i, path := range watchPaths {
		relative, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return nil, errors.Wrap(err, "could create relative path for watch pattern")
		}
		relativePaths[i] = relative
	}
	return relativePaths, nil
}

// Waits for file changes or context cancellation.
// If file changes are found, the chan is drained and true is returned.
func (s *Service) waitForFileChanges(ctx context.Context, fileEvents chan watch.WatchEventModuleChanged) bool {
	updates := s.updatesTopic.Subscribe(nil)
	defer s.updatesTopic.Unsubscribe(updates)

	select {
	case <-ctx.Done():
		return false
	case <-fileEvents:
		// Files changed. Now consume the rest of the events so that it is empty for the next build attempt
		for {
			select {
			case <-fileEvents:
			default:
				return true
			}
		}
	case <-updates:
		return true
	}
}

// watchFiles begin watching files in the module directory
// This is only used to restart quarkus:dev if it ends (such as when the initial build fails).
func watchFiles(ctx context.Context, watcher *watch.Watcher, buildCtx buildContext, events chan watch.WatchEventModuleChanged) error {
	logger := log.FromContext(ctx)
	watchTopic, err := watcher.Watch(ctx, time.Second, []string{buildCtx.Config.Dir})
	if err != nil {
		return errors.Wrap(err, "could not watch for file changes")
	}
	log.FromContext(ctx).Debugf("Watching for file changes: %s", buildCtx.Config.Dir)
	watchEvents := make(chan watch.WatchEvent, 32)
	watchTopic.Subscribe(watchEvents)

	// We need watcher to calculate file hashes before we do initial build so we can detect changes
	select {
	case e := <-watchEvents:
		_, ok := e.(watch.WatchEventModuleAdded)
		if !ok {
			return errors.Errorf("expected module added event, got: %T", e)
		}
	case <-time.After(3 * time.Second):
		return errors.Errorf("expected module added event, got no event")
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "context done")
	}
	stubsDir := filepath.Join(buildCtx.Config.Dir, "src", "main", "ftl-module-schema")
	go func() {
		for e := range channels.IterContext(ctx, watchEvents) {
			if change, ok := e.(watch.WatchEventModuleChanged); ok {
				// Ignore changes to external protos. If a depenency was updated, the plugin will receive a new build context.
				// Also ignore changes to queries and test files. This is a bit hacky ATM
				change.Changes = islices.Filter(change.Changes, func(c watch.FileChange) bool {
					if strings.HasPrefix(c.Path, stubsDir) {
						return false
					}
					if strings.HasSuffix(c.Path, ".sql") && strings.Contains(c.Path, "queries") {
						return false
					}
					return !strings.Contains(c.Path, "src/test")
				})
				if len(change.Changes) == 0 {
					continue
				}
				logger.Infof("File change detected: %v", e)
				events <- change
			}
		}
	}()
	return nil
}

type buildResult struct {
	state               *hotreloadpb.SchemaState
	buildContextUpdated bool
	failed              bool
	bctx                buildContext
}

func (s *Service) runQuarkusDev(parentCtx context.Context, projectConfig projectconfig.Config, realm, module string, stream *connect.ServerStream[langpb.BuildResponse], firstResponseSent *atomic.Value[bool], fileEvents chan watch.WatchEventModuleChanged, parentCancel context.CancelCauseFunc) error {
	logger := log.FromContext(parentCtx)
	ctx, cancel := context.WithCancelCause(parentCtx)
	output := &errorDetector{
		logger: logger,
	}
	defer func() {
		cancel(errors.Wrap(context.Canceled, "stopping JVM language plugin (Quarkus dev mode)"))

		// the context is done, which means Quarkus has exited
		// we need to send a build failure event
		ers := langpb.ErrorsToProto(output.FinalizeCapture(true))
		ers.Errors = append(ers.Errors, &langpb.Error{Msg: "The dev mode process exited", Level: langpb.Error_ERROR_LEVEL_ERROR, Type: langpb.Error_ERROR_TYPE_COMPILER})
		auto := firstResponseSent.Load()
		firstResponseSent.Store(true)
		err := stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				IsAutomaticRebuild: auto,
				ContextId:          s.buildContext.Load().ID,
				Errors:             ers,
			}}})
		if err != nil {
			logger.Errorf(err, "could not send build event")
			parentCancel(err)
		}
	}()

	events := make(chan buildContextUpdatedEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)
	release, err := flock.Acquire(ctx, s.buildContext.Load().Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return errors.Wrap(err, "could not acquire build lock")
	}
	defer release() //nolint:errcheck
	address, err := plugin.AllocatePort()
	if err != nil {
		return errors.Wrap(err, "could not allocate port")
	}
	buildCtx := s.buildContext.Load()
	ctx = log.ContextWithLogger(ctx, logger)
	devModeEndpoint := fmt.Sprintf("http://localhost:%d", address.Port)
	bind := devModeEndpoint
	devModeBuild := buildCtx.Config.DevModeBuild
	debugPort, err := plugin.AllocatePort()
	debugPort32 := int32(debugPort.Port)

	if err == nil {
		devModeBuild = fmt.Sprintf("%s -Ddebug=%d", devModeBuild, debugPort.Port)
	}
	hotReloadPort, err := plugin.AllocatePort()
	if err != nil {
		return errors.Wrap(err, "could not allocate port")
	}
	devModeBuild = fmt.Sprintf("%s -Dftl.language.port=%d", devModeBuild, hotReloadPort.Port)

	if os.Getenv("FTL_SUSPEND") == "true" {
		devModeBuild += " -Dsuspend "
	}
	launchQuarkusProcessAsync(ctx, devModeBuild, projectConfig, buildCtx, bind, output, cancel)

	// Wait for the plugin to start.
	hotReloadEndpoint := fmt.Sprintf("http://localhost:%d", hotReloadPort.Port)
	client := rpc.Dial(hotreloadpbconnect.NewHotReloadServiceClient, hotReloadEndpoint, log.Trace)
	err = s.connectReloadClient(ctx, client)
	if err != nil || client == nil {
		_ = output.FinalizeCapture(true)
		return errors.WithStack(err)
	}
	logger.Debugf("Dev mode process started")
	reloadEvents := make(chan *buildResult, 32)

	go rpc.RetryStreamingServerStream(ctx, "hot-reload", backoff.Backoff{Max: time.Millisecond * 100}, &hotreloadpb.WatchRequest{}, client.Watch, func(ctx context.Context, stream *hotreloadpb.WatchResponse) error { //nolint
		_ = output.FinalizeCapture(false)
		reloadEvents <- &buildResult{state: stream.GetState(), bctx: buildCtx}
		return nil
	}, func(err error) bool {
		return true
	})
	go func() {
		s.watchReloadEvents(ctx, reloadEvents, firstResponseSent, stream, devModeEndpoint, hotReloadEndpoint, debugPort32, cancel)
	}()

	errs := make(chan error)
	for {
		newKey := key.NewDeploymentKey(realm, module)
		select {
		case err := <-errs:
			_ = output.FinalizeCapture(true)
			return errors.Wrap(err, "hot reload failed")
		case bc := <-events:
			logger.Debugf("Build context updated")
			_ = output.FinalizeCapture(false)
			go func() {
				err = s.doReload(ctx, client, &hotreloadpb.ReloadRequest{SchemaChanged: bc.schemaChanged}, reloadEvents, true, newKey)
				if err != nil {
					errs <- err
				}
			}()
		case <-fileEvents:
			_ = output.FinalizeCapture(false)
			newDeps, err := extractDependencies(buildCtx.Config.Module, buildCtx.Config.Dir)
			if err != nil {
				logger.Errorf(err, "could not extract dependencies")
			} else if !slices.Equal(islices.Sort(newDeps), islices.Sort(s.buildContext.Load().Dependencies)) {
				err := stream.Send(&langpb.BuildResponse{
					Event: &langpb.BuildResponse_BuildFailure{
						BuildFailure: &langpb.BuildFailure{
							ContextId:              buildCtx.ID,
							IsAutomaticRebuild:     true,
							InvalidateDependencies: true,
						},
					},
				})
				if err != nil {
					parentCancel(err)
					return errors.Wrap(err, "could not send build event")
				}
				continue
			}

			go func() {
				err = s.doReload(ctx, client, &hotreloadpb.ReloadRequest{}, reloadEvents, false, newKey)
				if err != nil {
					errs <- err
				}

			}()
		case <-ctx.Done():

			return errors.Wrap(ctx.Err(), "context cancelled")
		}
	}
}

func (s *Service) doReload(ctx context.Context, client hotreloadpbconnect.HotReloadServiceClient, request *hotreloadpb.ReloadRequest, reloadEvents chan *buildResult, buildContextUpdated bool, newKey key.Deployment) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Sending hot reload request")
	result, err := client.Reload(ctx, connect.NewRequest(request))

	if err != nil {
		logger.Debugf("Reload failed, attempting to reconnect")
		// If the connection has failed we try again
		err = s.connectReloadClient(ctx, client)

		if err != nil {
			logger.Debugf("Reconnect failed, unable to connect to client")
			return err
		}
		result, err = client.Reload(ctx, connect.NewRequest(&hotreloadpb.ReloadRequest{}))
	}
	if err != nil {
		logger.Debugf("Unable to invoke reload on the JVM") //TODO: restart
		return errors.Wrap(err, "unable to invoke hot reload")
	}
	handleReloadResponse(result, newKey)
	reloadEvents <- &buildResult{state: result.Msg.GetState(), failed: result.Msg.Failed, bctx: s.buildContext.Load(), buildContextUpdated: buildContextUpdated}
	return nil
}

func handleReloadResponse(result *connect.Response[hotreloadpb.ReloadResponse], newKey key.Deployment) {
	if result.Msg.State.Module != nil {
		if result.Msg.State.Module.Runtime == nil {
			result.Msg.State.Module.Runtime = &schemapb.ModuleRuntime{}
		}
		if result.Msg.State.Module.Runtime.Deployment == nil {
			result.Msg.State.Module.Runtime.Deployment = &schemapb.ModuleRuntimeDeployment{}
		}
		result.Msg.State.Module.Runtime.Deployment.DeploymentKey = newKey.String()
	}
}

func (s *Service) watchReloadEvents(ctx context.Context, reloadEvents chan *buildResult, firstResponseSent *atomic.Value[bool], stream *connect.ServerStream[langpb.BuildResponse], devModeEndpoint string, hotReloadEndpoint string, debugPort32 int32, cancel context.CancelCauseFunc) {
	logger := log.FromContext(ctx)
	lastFailed := false
	for event := range channels.IterContext(ctx, reloadEvents) {
		changed := event.state.GetNewRunnerRequired()
		errorList := event.state.GetErrors()
		logger.Debugf("Checking for schema changes: changed: %v failed: %v buildContextUpdated: %v newRunnerRequired: %v schemaVersion %v", changed, event.failed, event.buildContextUpdated, event.state.NewRunnerRequired, event.state.Version)

		if changed || event.buildContextUpdated || event.failed || lastFailed {
			lastFailed = false
			buildCtx := event.bctx
			auto := firstResponseSent.Load() && !event.buildContextUpdated
			if auto {
				logger.Debugf("sending auto build event")
				err := stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_AutoRebuildStarted{AutoRebuildStarted: &langpb.AutoRebuildStarted{ContextId: buildCtx.ID}}})
				if err != nil {
					logger.Errorf(err, "could not send build event")
					cancel(err)
					return
				}
			}
			if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(errorList)) || event.failed {
				lastFailed = true
				// skip reading schema
				logger.Warnf("Build failed, skipping schema, sending build failure")
				err := stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
					BuildFailure: &langpb.BuildFailure{
						IsAutomaticRebuild: auto,
						ContextId:          buildCtx.ID,
						Errors:             errorList,
					}}})
				if err != nil {
					logger.Errorf(err, "Could not send build event")
					cancel(err)
					return
				}
				firstResponseSent.Store(true)
				continue
			}
			moduleProto := event.state.GetModule()
			moduleSch, err := schema.ModuleFromProto(moduleProto)
			if err != nil {
				err := stream.Send(buildFailure(buildCtx, auto, builderrors.Error{
					Type:  builderrors.FTL,
					Level: builderrors.ERROR,
					Msg:   fmt.Sprintf("Could not parse schema from proto: %v", err),
				}))
				if err != nil {
					logger.Errorf(err, "Could not send build event")
					cancel(err)
					return
				}
				firstResponseSent.Store(true)
				continue
			}
			if _, validationErr := schema.ValidateModuleInSchema(buildCtx.Schema, optional.Some(moduleSch)); validationErr != nil {
				err := stream.Send(buildFailure(buildCtx, auto, builderrors.Error{
					Type:  builderrors.FTL,
					Level: builderrors.ERROR,
					Msg:   validationErr.Error(),
				}))
				if err != nil {
					logger.Errorf(err, "Could not send build event")
					cancel(err)
					return
				}
				firstResponseSent.Store(true)
				continue
			}
			if firstResponseSent.Load() {
				logger.Debugf("Live reload schema changed, sending build success event")
			}

			firstResponseSent.Store(true)
			err = stream.Send(&langpb.BuildResponse{
				Event: &langpb.BuildResponse_BuildSuccess{
					BuildSuccess: &langpb.BuildSuccess{
						ContextId:            buildCtx.ID,
						IsAutomaticRebuild:   auto,
						Module:               moduleSch.ToProto(),
						DevEndpoint:          ptr(devModeEndpoint),
						DevHotReloadEndpoint: ptr(hotReloadEndpoint),
						DebugPort:            &debugPort32,
						Deploy:               []string{SchemaFile},
						DevHotReloadVersion:  &event.state.Version,
					},
				},
			})
			if err != nil {
				logger.Errorf(err, "could not send build event")
				cancel(err)
				return
			}
		}
	}
}

func launchQuarkusProcessAsync(ctx context.Context, devModeBuild string, projectConfig projectconfig.Config, buildCtx buildContext, bind string, stdout *errorDetector, cancel context.CancelCauseFunc) {
	go func() {
		logger := log.FromContext(ctx)
		logger.Infof("Using dev mode build command '%s'", devModeBuild)
		command := exec.Command(ctx, log.Debug, buildCtx.Config.Dir, "bash", "-c", devModeBuild)
		if os.Getenv("MAVEN_OPTS") == "" {
			command.Env = append(command.Env, "MAVEN_OPTS=-Xmx2048m")
		}
		command.Env = append(command.Env, fmt.Sprintf("FTL_BIND=%s", bind), "FTL_MODULE_NAME="+buildCtx.Config.Module, "FTL_PROJECT_ROOT="+projectConfig.Root())
		command.Stdout = stdout
		command.Stderr = os.Stderr
		err := command.Run()
		if err != nil {
			stdout.FinalizeCapture(true)
			logger.Errorf(err, "Dev mode process exited with error")
			cancel(errors.Wrap(errors.Join(err, context.Canceled), "dev mode process exited with error"))
		} else {
			logger.Infof("Dev mode process exited")
			cancel(errors.Wrap(context.Canceled, "dev mode process exited"))
		}
	}()
}

func (s *Service) connectReloadClient(ctx context.Context, client hotreloadpbconnect.HotReloadServiceClient) error {
	logger := log.FromContext(ctx)
	err := rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute*100, client)
	if err != nil {
		logger.Infof("Dev mode process failed to start")
		select {
		case <-ctx.Done():
			return errors.Errorf("dev mode process exited")
		default:
		}
		return errors.Wrap(err, "timed out waiting for star")
	}
	return nil
}

func build(ctx context.Context, projectConfig projectconfig.Config, bctx buildContext, autoRebuild bool) (*langpb.BuildResponse, error) {
	logger := log.FromContext(ctx)
	release, err := flock.Acquire(ctx, bctx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not acquire build lock")
	}
	defer release() //nolint:errcheck

	deps, err := extractDependencies(bctx.Config.Module, bctx.Config.Dir)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract dependencies")
	}

	if !slices.Equal(islices.Sort(deps), islices.Sort(bctx.Dependencies)) {
		// dependencies have changed
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				ContextId:              bctx.ID,
				IsAutomaticRebuild:     autoRebuild,
				InvalidateDependencies: true,
			},
		}}, nil
	}
	ensureCorrectFTLVersion(ctx, bctx)
	output := &errorDetector{
		logger: logger,
	}
	config := bctx.Config
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	command.Env = append(command.Env, "FTL_MODULE_NAME="+bctx.Config.Module, "FTL_PROJECT_ROOT="+projectConfig.Root())
	command.Stdout = output
	command.Stderr = os.Stderr
	err = command.Run()

	if err != nil {
		buildErrs := output.FinalizeCapture(true)
		if len(buildErrs) == 0 {
			buildErrs = []builderrors.Error{{Msg: "Compile process unexpectedly exited without reporting any errors", Level: builderrors.ERROR, Type: builderrors.COMPILER}}
		}
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{BuildFailure: &langpb.BuildFailure{
			IsAutomaticRebuild: autoRebuild,
			ContextId:          bctx.ID,
			Errors:             langpb.ErrorsToProto(buildErrs),
		}}}, nil
	}

	buildErrs, err := loadProtoErrors(config)
	capturedErrors := output.FinalizeCapture(false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load build errors")
	}
	buildErrs.Errors = append(buildErrs.Errors, langpb.ErrorsToProto(capturedErrors).Errors...)
	if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(buildErrs)) {
		// skip reading schema
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				IsAutomaticRebuild: autoRebuild,
				ContextId:          bctx.ID,
				Errors:             buildErrs,
			}}}, nil
	}

	moduleProto, err := readSchema(bctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				IsAutomaticRebuild: autoRebuild,
				ContextId:          bctx.ID,
				Errors:             buildErrs,
				Module:             moduleProto,
				Deploy:             []string{"launch", "quarkus-app"},
			},
		},
	}, nil
}

func ensureCorrectFTLVersion(ctx context.Context, bctx buildContext) {
	logger := log.FromContext(ctx)
	javaConfig, err := loadJavaConfig(bctx.Config.LanguageConfig, bctx.Config.Language)
	if err != nil {
		logger.Errorf(err, "unable to read JVM config %s", bctx.Config.Dir)
		return
	}
	if javaConfig.BuildTool == JavaBuildToolMaven {
		if err := setPOMProperties(ctx, bctx.Config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Errorf(err, "unable to update ftl.version in %s", bctx.Config.Dir)
		}
	}
}

func readSchema(bctx buildContext) (*schemapb.Module, error) {
	path := filepath.Join(bctx.Config.DeployDir, SchemaFile)
	moduleSchema, err := schema.ModuleFromProtoFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read schema for module: %s from %s", bctx.Config.Module, path)
	}

	moduleSchema.Runtime = &schema.ModuleRuntime{
		Base: schema.ModuleRuntimeBase{
			CreateTime: time.Now(),
			Language:   bctx.Config.Language,
			Image:      "ftl0/ftl-runner-jvm",
		},
		Scaling: &schema.ModuleRuntimeScaling{
			MinReplicas: 1,
		},
	}

	moduleProto := moduleSchema.ToProto()
	return moduleProto, nil
}

// BuildContextUpdated is called whenever the build context is update while a Build call with "rebuild_automatically" is active.
//
// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
// event with the updated build context id with "is_automatic_rebuild" as false.
func (s *Service) BuildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	if !s.acceptsContextUpdates.Load() {
		return nil, errors.Errorf("plugin does not accept context updates because these is no build stream allowing rebuilds")
	}
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s.buildContext.Store(buildCtx)
	changed, err := s.writeGenericSchemaFiles(ctx, buildCtx.Schema, buildCtx.Config, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write generic schema files")
	}

	s.updatesTopic.Publish(buildContextUpdatedEvent{
		buildCtx:      buildCtx,
		schemaChanged: changed,
	})

	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, stream *connect.ServerStream[langpb.BuildResponse], projectConfig projectconfig.Config, buildCtx buildContext, isAutomaticRebuild bool) error {
	buildEvent, err := build(ctx, projectConfig, buildCtx, isAutomaticRebuild)
	if err != nil {
		buildEvent = buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		})
	}
	if err = stream.Send(buildEvent); err != nil {
		return errors.Wrap(err, "could not send build event")
	}
	return nil
}

// buildFailure creates a BuildFailure event based on build errors.
func buildFailure(buildCtx buildContext, isAutomaticRebuild bool, errs ...builderrors.Error) *langpb.BuildResponse {
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				ContextId:              buildCtx.ID,
				IsAutomaticRebuild:     isAutomaticRebuild,
				Errors:                 langpb.ErrorsToProto(errs),
				InvalidateDependencies: false,
			},
		},
	}
}

const JavaBuildToolMaven string = "maven"
const JavaBuildToolGradle string = "gradle"

type JavaConfig struct {
	BuildTool string `mapstructure:"build-tool"`
}

func loadJavaConfig(languageConfig any, language string) (JavaConfig, error) {
	var javaConfig JavaConfig
	err := mapstructure.Decode(languageConfig, &javaConfig)
	if err != nil {
		return JavaConfig{}, errors.Wrapf(err, "failed to decode %s config", language)
	}
	return javaConfig, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	modules, err := extractDependencies(req.Msg.ModuleConfig.Name, req.Msg.ModuleConfig.Dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse[langpb.GetDependenciesResponse](&langpb.GetDependenciesResponse{Modules: modules}), nil
}

func extractDependencies(moduleName string, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	// We also attempt to look at kotlin files
	// As the Java module supports both
	kotin, kotlinErr := extractKotlinFTLImports(moduleName, dir)
	if kotlinErr == nil {
		// We don't really care about the error case, its probably a Java project
		for _, imp := range kotin {
			dependencies[imp] = true
		}
	}
	javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to walk directory")
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := javaImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == moduleName {
					continue
				}
				dependencies[module] = true
			}
		}
		return errors.WithStack(scanner.Err())
	})

	// We only error out if they both failed
	if err != nil && kotlinErr != nil {
		return nil, errors.Wrapf(err, "%s: failed to extract dependencies from Java module", moduleName)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

func extractKotlinFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "could not open file while extracting dependencies")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == self {
					continue
				}
				dependencies[module] = true
			}
		}
		return errors.WithStack(scanner.Err())
	})

	if err != nil {
		return nil, errors.Wrapf(err, "%s: failed to extract dependencies from Kotlin module", self)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

// setPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func setPOMProperties(ctx context.Context, baseDir string) error {
	changed := false
	logger := log.FromContext(ctx)
	ftlVersion := ftl.BaseVersion(ftl.Version)
	if !ftl.IsRelease(ftlVersion) || ftlVersion != ftl.BaseVersion(ftl.Version) {
		ftlVersion = "1.0-SNAPSHOT"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return errors.Wrapf(err, "unable to read %s", pomFile)
	}
	root := tree.Root()

	parent := root.SelectElement("parent")
	versionSet := false
	if parent != nil {
		// You can't use properties in the parent
		// If they are using our parent then we want to update the version
		group := parent.SelectElement("groupId")
		artifact := parent.SelectElement("artifactId")
		if group.Text() == "xyz.block.ftl" && (artifact.Text() == "ftl-build-parent-java" || artifact.Text() == "ftl-build-parent-kotlin") {
			version := parent.SelectElement("version")
			if version != nil {
				if version.Text() != ftlVersion {
					version.SetText(ftlVersion)
					changed = true
				}
				versionSet = true
			}
		}
	}

	propChanged, err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return errors.WithStack(err)
	}
	if propChanged || changed {
		err = tree.WriteToFile(pomFile)
		if err != nil {
			return errors.Wrapf(err, "unable to write %s", pomFile)
		}
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) (bool, error) {
	properties := root.SelectElement("properties")
	if properties == nil {
		return false, errors.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return false, errors.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	if version.Text() == ftlVersion {
		return false, nil
	}
	version.SetText(ftlVersion)
	return true, nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*langpb.ErrorList, error) {
	errorsPath := filepath.Join(config.DeployDir, "errors.pb")
	if _, err := os.Stat(errorsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(errorsPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not load build errors file")
	}

	errorspb := &langpb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal build error")
	}
	return errorspb, nil
}

func ptr(s string) *string {
	return &s
}

func (s *Service) writeGenericSchemaFiles(ctx context.Context, v *schema.Schema, config moduleconfig.AbsModuleConfig, containsGeneratedSchema bool) (bool, error) {
	if v == nil {
		return false, nil
	}
	logger := log.FromContext(ctx)
	modPath := filepath.Join(config.Dir, "src", "main", "ftl-module-schema")
	err := os.MkdirAll(modPath, 0750)
	if err != nil {
		return false, errors.Wrapf(err, "failed to create directory %s", modPath)
	}
	changed := false

	for _, mod := range v.InternalModules() {
		if mod.Name == config.Module {
			if containsGeneratedSchema {
				logger.Debugf("writing generated schema files for %s", mod.Name)
				sw, err := s.writeGeneratedSchemaFiles(mod, modPath)
				if err != nil {
					return false, errors.WithStack(err)
				}
				changed = changed || sw
			}
			continue
		}
		deps := v.ModuleDependencies(mod.Name)
		if deps[config.Module] != nil {
			continue
		}
		data, err := schema.ModuleToBytes(mod)
		if err != nil {
			return false, errors.Wrapf(err, "failed to export module schema for module %s", mod.Name)
		}
		schemaFile := filepath.Join(modPath, mod.Name+".pb")
		if fileExists(schemaFile) {
			existing, err := os.ReadFile(schemaFile)
			if err == nil {
				// We ignore errors, but if the read succeeded we need to check if the file has changed
				if bytes.Equal(existing, data) {
					continue
				}
			}
		}
		changed = true
		err = os.WriteFile(schemaFile, data, 0644) // #nosec
		logger.Debugf("writing schema files for %s to %s", mod.Name, schemaFile)
		if err != nil {
			return false, errors.Wrapf(err, "failed to write schema file for module %s", mod.Name)
		}
	}
	return changed, nil
}

func (s *Service) writeGeneratedSchemaFiles(m *schema.Module, schemaDir string) (bool, error) {
	if m == nil {
		return false, nil
	}
	generatedDir := filepath.Join(schemaDir, "generated")
	err := os.MkdirAll(generatedDir, 0750)
	if err != nil {
		return false, errors.Wrapf(err, "failed to create directory %s", generatedDir)
	}
	schemaFile := filepath.Join(generatedDir, m.Name+".pb")

	generatedModule := m.ToGeneratedModule()
	if len(generatedModule.Decls) == 0 {
		if fileExists(schemaFile) {
			err := os.Remove(schemaFile)
			if err != nil {
				return false, errors.Wrapf(err, "failed to remove generated schema file for module %s", m.Name)
			}
		}
		return false, nil
	}

	data, err := schema.ModuleToBytes(generatedModule)
	if err != nil {
		return false, errors.Wrapf(err, "failed to export generated module schema for module %s", m.Name)
	}

	if fileExists(schemaFile) {
		existing, err := os.ReadFile(schemaFile)
		if err == nil {
			// file has not changed, no need to write
			if bytes.Equal(existing, data) {
				return false, nil
			}
		}
	}

	err = os.WriteFile(schemaFile, data, 0644) // #nosec
	if err != nil {
		return false, errors.Wrapf(err, "failed to write generated schema file for module %s", m.Name)
	}
	return true, nil
}
