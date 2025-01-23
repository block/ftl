package common

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/beevik/etree"
	"github.com/block/scaffolder"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/block/ftl"
	hotreloadpb "github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1/hotreloadpbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/plugin"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute
const SchemaFile = "schema.pb"
const ErrorFile = "errors.pb"

type buildContextUpdatedEvent struct {
	buildCtx buildContext
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
		return buildContext{}, fmt.Errorf("could not parse schema from proto: %w", err)
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
	scaffoldFiles         *zip.Reader
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New(scaffoldFiles *zip.Reader) *Service {
	return &Service{
		updatesTopic:  pubsub.New[buildContextUpdatedEvent](),
		scaffoldFiles: scaffoldFiles,
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetCreateModuleFlagsResponse{
		Flags: []*langpb.GetCreateModuleFlagsResponse_Flag{
			{
				Name: "group",
				Help: "The Maven groupId of the project.",
			},
		},
	}), nil
}

// CreateModule generates files for a new module with the requested name
func (s *Service) CreateModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.Name)
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)
	groupAny, ok := req.Msg.Flags.AsMap()["group"]
	if !ok {
		groupAny = ""
	}
	group, ok := groupAny.(string)
	if !ok {
		return nil, fmt.Errorf("group not a string")
	}
	if group == "" {
		group = "ftl." + req.Msg.Name
	}

	packageDir := strings.ReplaceAll(group, ".", "/")
	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"), // This is still needed, as there is an 'ignore' module in the scaffold dir
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	version := ftl.BaseVersion(ftl.Version)
	if !ftl.IsRelease(version) {
		version = "1.0-SNAPSHOT"
	}
	sctx := struct {
		Dir        string
		Name       string
		Group      string
		PackageDir string
		Version    string
	}{
		Dir:        projConfig.Path,
		Name:       req.Msg.Name,
		Group:      group,
		PackageDir: packageDir,
		Version:    version,
	}
	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(s.scaffoldFiles, parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	return connect.NewResponse(&langpb.CreateModuleResponse{}), nil
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
		return nil, fmt.Errorf("failed to parse schema from proto: %w", err)
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	err = s.writeGenericSchemaFiles(ctx, sch, config)
	if err != nil {
		return nil, err
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
	logger = logger.Module(req.Msg.BuildContext.ModuleConfig.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return err
	}
	err = s.writeGenericSchemaFiles(ctx, buildCtx.Schema, buildCtx.Config)
	if err != nil {
		return fmt.Errorf("failed to write generic schema files: %w", err)
	}
	if req.Msg.RebuildAutomatically {
		return s.runDevMode(ctx, req, buildCtx, stream)
	}

	if s.acceptsContextUpdates.Load() {
		// Already running in dev mode, we don't need to rebuild
		s.updatesTopic.Publish(buildContextUpdatedEvent{buildCtx: buildCtx})
		return nil
	}

	// Initial build
	if err := buildAndSend(ctx, stream, buildCtx, false); err != nil {
		return err
	}

	return nil
}

func (s *Service) runDevMode(ctx context.Context, req *connect.Request[langpb.BuildRequest], buildCtx buildContext, stream *connect.ServerStream[langpb.BuildResponse]) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.acceptsContextUpdates.Store(true)
	defer s.acceptsContextUpdates.Store(false)

	watchPatterns, err := relativeWatchPatterns(buildCtx.Config.Dir, buildCtx.Config.Watch)
	if err != nil {
		return err
	}
	watcher := watch.NewWatcher(optional.None[string](), watchPatterns...)
	fileEvents := make(chan watch.WatchEventModuleChanged, 32)
	if err := watchFiles(ctx, watcher, buildCtx, fileEvents); err != nil {
		return err
	}

	first := true
	for {
		if !first {
			err := stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_AutoRebuildStarted{AutoRebuildStarted: &langpb.AutoRebuildStarted{ContextId: buildCtx.ID}}})
			if err != nil {
				return fmt.Errorf("could not send build event: %w", err)
			}
		}
		err := s.runQuarkusDev(ctx, req, stream, first)
		first = false
		if err != nil {
			log.FromContext(ctx).Errorf(err, "Dev mode process exited")
			return err
		}
		if !waitForFileChanges(ctx, fileEvents) {
			return nil
		}
	}
}

func relativeWatchPatterns(moduleDir string, watchPaths []string) ([]string, error) {
	relativePaths := make([]string, len(watchPaths))
	for i, path := range watchPaths {
		relative, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return nil, fmt.Errorf("could create relative path for watch pattern: %w", err)
		}
		relativePaths[i] = relative
	}
	return relativePaths, nil
}

// Waits for file changes or context cancellation.
// If file changes are found, the chan is drained and true is returned.
func waitForFileChanges(ctx context.Context, fileEvents chan watch.WatchEventModuleChanged) bool {
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
	}
}

// watchFiles begin watching files in the module directory
// This is only used to restart quarkus:dev if it ends (such as when the initial build fails).
func watchFiles(ctx context.Context, watcher *watch.Watcher, buildCtx buildContext, events chan watch.WatchEventModuleChanged) error {
	watchTopic, err := watcher.Watch(ctx, time.Second, []string{buildCtx.Config.Dir})
	if err != nil {
		return fmt.Errorf("could not watch for file changes: %w", err)
	}
	log.FromContext(ctx).Debugf("Watching for file changes: %s", buildCtx.Config.Dir)
	watchEvents := make(chan watch.WatchEvent, 32)
	watchTopic.Subscribe(watchEvents)

	// We need watcher to calculate file hashes before we do initial build so we can detect changes
	select {
	case e := <-watchEvents:
		_, ok := e.(watch.WatchEventModuleAdded)
		if !ok {
			return fmt.Errorf("expected module added event, got: %T", e)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("expected module added event, got no event")
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	}

	go func() {
		for e := range channels.IterContext(ctx, watchEvents) {
			if change, ok := e.(watch.WatchEventModuleChanged); ok {
				events <- change
			}
		}
	}()
	return nil
}

func (s *Service) runQuarkusDev(ctx context.Context, req *connect.Request[langpb.BuildRequest], stream *connect.ServerStream[langpb.BuildResponse], firstAttempt bool) error {
	logger := log.FromContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events := make(chan buildContextUpdatedEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return err
	}
	release, err := flock.Acquire(ctx, buildCtx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return fmt.Errorf("could not acquire build lock: %w", err)
	}
	defer release() //nolint:errcheck
	address, err := plugin.AllocatePort()
	if err != nil {
		return fmt.Errorf("could not allocate port: %w", err)
	}

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
		return fmt.Errorf("could not allocate port: %w", err)
	}
	devModeBuild = fmt.Sprintf("%s -Dftl.language.port=%d", devModeBuild, hotReloadPort.Port)

	if os.Getenv("FTL_SUSPEND") == "true" {
		devModeBuild += " -Dsuspend "
	}
	go func() {
		logger.Infof("Using dev mode build command '%s'", devModeBuild)
		command := exec.Command(ctx, log.Debug, buildCtx.Config.Dir, "bash", "-c", devModeBuild)
		command.Env = append(command.Env, fmt.Sprintf("FTL_BIND=%s", bind), "MAVEN_OPTS=-Xmx1024m")
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err = command.Run()
		if err != nil {
			logger.Errorf(err, "Dev mode process exited with error")
		} else {
			logger.Infof("Dev mode process exited")
		}
		cancel()
	}()

	// Wait for the plugin to start.
	hotReloadEndpoint := fmt.Sprintf("http://localhost:%d", hotReloadPort.Port)
	client := rpc.Dial(hotreloadpbconnect.NewHotReloadServiceClient, hotReloadEndpoint, log.Trace)
	err = rpc.Wait(ctx, backoff.Backoff{}, time.Minute, client)
	if err != nil {
		logger.Infof("Dev mode process failed to start")
		return fmt.Errorf("timed out waiting for start %w", err)
	}
	logger.Debugf("Dev mode process started")

	schemaHash := sha256.SHA256{}
	errorHash := sha256.SHA256{}
	migrationHash := watch.FileHashes{}

	if fileExists(buildCtx.Config.SQLMigrationDirectory) {
		migrationHash, err = watch.ComputeFileHashes(buildCtx.Config.SQLMigrationDirectory, true, []string{"**/*.sql"})
		if err != nil {
			return fmt.Errorf("could not compute file hashes: %w", err)
		}
	}

	type buildResult struct {
		state       *hotreloadpb.SchemaState
		forceReload bool
	}

	reloadEvents := make(chan *buildResult, 32)
	go func() {
		firstAttempt := true
		var module *schemapb.Module
		for event := range channels.IterContext(ctx, reloadEvents) {
			changed := false
			if event.state.GetErrors() != nil {
				bytes, err := proto.Marshal(event.state.GetErrors())
				if err != nil {
					logger.Errorf(err, "failed to marshal errors")
					continue
				}
				sum := sha256.Sum(bytes)
				if sum != errorHash {
					changed = true
					errorHash = sum
				}
			}
			if event.state.GetModule() != nil {
				module = event.state.GetModule()
				bytes, err := proto.Marshal(module)
				if err != nil {
					logger.Errorf(err, "failed to marshal schema")
					continue
				}
				sum := sha256.Sum(bytes)
				if sum != schemaHash {
					changed = true
					schemaHash = sum
				}
			}
			logger.Tracef("Checking for schema changes %v %v %v", changed, schemaHash, errorHash)

			if changed || event.forceReload {
				auto := !firstAttempt
				if auto {
					logger.Debugf("sending auto build event")
					err = stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_AutoRebuildStarted{AutoRebuildStarted: &langpb.AutoRebuildStarted{ContextId: buildCtx.ID}}})
					if err != nil {
						logger.Errorf(err, "could not send build event")
						continue
					}
				}
				buildErrs, err := loadProtoErrors(buildCtx.Config)
				if err != nil {
					// This is likely a transient error
					logger.Errorf(err, "failed to load build errors")
					continue
				}
				if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(buildErrs)) {
					// skip reading schema
					err = stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
						BuildFailure: &langpb.BuildFailure{
							IsAutomaticRebuild: auto,
							ContextId:          buildCtx.ID,
							Errors:             buildErrs,
						}}})
					if err != nil {
						logger.Errorf(err, "Could not send build event")
					}
					firstAttempt = false
					continue
				}

				if !firstAttempt {
					logger.Infof("Live reload schema changed, sending build success event")
				}
				firstAttempt = false
				err = stream.Send(&langpb.BuildResponse{
					Event: &langpb.BuildResponse_BuildSuccess{
						BuildSuccess: &langpb.BuildSuccess{
							ContextId:            buildCtx.ID,
							IsAutomaticRebuild:   auto,
							Module:               module,
							DevEndpoint:          ptr(devModeEndpoint),
							DevHotReloadEndpoint: ptr(hotReloadEndpoint),
							DebugPort:            &debugPort32,
							Deploy:               []string{SchemaFile},
						},
					},
				})
				if err != nil {
					logger.Errorf(err, "could not send build event")
				}
			}
		}
	}()

	schemaWatch, err := client.Watch(ctx, connect.NewRequest(&hotreloadpb.WatchRequest{}))
	if !schemaWatch.Receive() {
		err := schemaWatch.Err()
		if err != nil {
			return fmt.Errorf("failed to receive initial schema watch response: %w", err)
		}
		return fmt.Errorf("failed to receive initial schema watch response")
	}
	reloadEvents <- &buildResult{state: schemaWatch.Msg().GetState()}
	go func() {
		for schemaWatch.Receive() {
			reloadEvents <- &buildResult{state: schemaWatch.Msg().GetState()}
		}
	}()

	schemaChangeTicker := time.NewTicker(500 * time.Millisecond)
	defer schemaChangeTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debugf("Context done")
			// the context is done before we notified the build engine
			// we need to send a build failure event
			err = stream.Send(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					IsAutomaticRebuild: true,
					ContextId:          buildCtx.ID,
					Errors:             &langpb.ErrorList{Errors: []*langpb.Error{{Msg: "The dev mode process exited", Level: langpb.Error_ERROR_LEVEL_ERROR, Type: langpb.Error_ERROR_TYPE_COMPILER}}},
				}}})
			if err != nil {
				return fmt.Errorf("could not send build event: %w", err)
			}
			return nil
		case bc := <-events:
			logger.Debugf("Build context updated")
			buildCtx = bc.buildCtx
			result, err := client.Reload(ctx, connect.NewRequest(&hotreloadpb.ReloadRequest{Force: true}))
			if err != nil {
				return fmt.Errorf("failed to invoke hot reload for build context update %w", err)
			}
			reloadEvents <- &buildResult{state: result.Msg.GetState(), forceReload: true}
		case <-schemaChangeTicker.C:
			changed := false
			logger.Debugf("Calling reload")
			result, err := client.Reload(ctx, connect.NewRequest(&hotreloadpb.ReloadRequest{Force: false}))
			logger.Debugf("Called reload")
			if err != nil {
				return fmt.Errorf("failed to invoke hot reload for build context update %w", err)
			}
			logger.Debugf("Checking for schema changes")

			if fileExists(buildCtx.Config.SQLMigrationDirectory) {
				newMigrationHash, err := watch.ComputeFileHashes(buildCtx.Config.SQLMigrationDirectory, true, []string{"**/*.sql"})
				if err != nil {
					logger.Errorf(err, "could not compute file hashes")
				} else if !reflect.DeepEqual(newMigrationHash, migrationHash) {
					changed = true
					migrationHash = newMigrationHash
				}
			}
			reloadEvents <- &buildResult{state: result.Msg.GetState(), forceReload: changed}

		}
	}
}

func build(ctx context.Context, bctx buildContext, autoRebuild bool) (*langpb.BuildResponse, error) {
	logger := log.FromContext(ctx)
	release, err := flock.Acquire(ctx, bctx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, fmt.Errorf("could not acquire build lock: %w", err)
	}
	defer release() //nolint:errcheck

	deps, err := extractDependencies(bctx.Config.Module, bctx.Config.Dir)
	if err != nil {
		return nil, fmt.Errorf("could not extract dependencies: %w", err)
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
	config := bctx.Config
	javaConfig, err := loadJavaConfig(config.LanguageConfig, config.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	if javaConfig.BuildTool == JavaBuildToolMaven {
		if err := setPOMProperties(ctx, config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Warnf("unable to update ftl.version in %s: %s", config.Dir, err.Error())
		}
	}
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	err = command.RunBuffered(ctx)
	if err != nil {
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{&langpb.BuildFailure{
			IsAutomaticRebuild: autoRebuild,
			ContextId:          bctx.ID,
			Errors:             &langpb.ErrorList{Errors: []*langpb.Error{{Msg: err.Error(), Level: langpb.Error_ERROR_LEVEL_ERROR, Type: langpb.Error_ERROR_TYPE_COMPILER}}},
		}}}, nil
	}

	buildErrs, err := loadProtoErrors(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load build errors: %w", err)
	}
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
		return nil, err
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

func readSchema(bctx buildContext) (*schemapb.Module, error) {
	path := filepath.Join(bctx.Config.DeployDir, SchemaFile)
	moduleSchema, err := schema.ModuleFromProtoFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema for module: %s from %s %w", bctx.Config.Module, path, err)
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
		return nil, fmt.Errorf("plugin does not accept context updates because these is no build stream allowing rebuilds")
	}
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, err
	}

	err = s.writeGenericSchemaFiles(ctx, buildCtx.Schema, buildCtx.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to write generic schema files: %w", err)
	}

	s.updatesTopic.Publish(buildContextUpdatedEvent{
		buildCtx: buildCtx,
	})

	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, stream *connect.ServerStream[langpb.BuildResponse], buildCtx buildContext, isAutomaticRebuild bool) error {
	buildEvent, err := build(ctx, buildCtx, isAutomaticRebuild)
	if err != nil {
		buildEvent = buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		})
	}
	if err = stream.Send(buildEvent); err != nil {
		return fmt.Errorf("could not send build event: %w", err)
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
		return JavaConfig{}, fmt.Errorf("failed to decode %s config: %w", language, err)
	}
	return javaConfig, nil
}

func (s *Service) ModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	defaults := langpb.ModuleConfigDefaultsResponse{
		LanguageConfig:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
		Watch:           []string{"pom.xml", "src/**", "build/generated", "target/generated-sources", "src/main/resources/db"},
		SqlMigrationDir: "src/main/resources/db/schema",
		SqlQueryDir:     "src/main/resources/db/queries",
	}
	dir := req.Msg.Dir
	pom := filepath.Join(dir, "pom.xml")
	buildGradle := filepath.Join(dir, "build.gradle")
	buildGradleKts := filepath.Join(dir, "build.gradle.kts")
	if fileExists(pom) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolMaven)
		defaults.DevModeBuild = ptr("mvn -Dquarkus.console.enabled=false -q clean quarkus:dev")
		defaults.Build = ptr("mvn -B clean package")
		defaults.DeployDir = "target"
	} else if fileExists(buildGradle) || fileExists(buildGradleKts) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolGradle)
		defaults.DevModeBuild = ptr("gradle clean quarkusDev")
		defaults.Build = ptr("gradle clean build")
		defaults.DeployDir = "build"
	} else {
		return nil, fmt.Errorf("could not find JVM build file in %s", dir)
	}

	return connect.NewResponse[langpb.ModuleConfigDefaultsResponse](&defaults), nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	modules, err := extractDependencies(req.Msg.ModuleConfig.Name, req.Msg.ModuleConfig.Dir)
	if err != nil {
		return nil, err
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
			return fmt.Errorf("failed to walk directory: %w", err)
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
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
		return scanner.Err()
	})

	// We only error out if they both failed
	if err != nil && kotlinErr != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Java module: %w", moduleName, err)
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
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file while extracting dependencies: %w", err)
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
		return scanner.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Kotlin module: %w", self, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

// setPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func setPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.BaseVersion(ftl.Version)
	if !ftl.IsRelease(ftlVersion) {
		ftlVersion = "1.0-SNAPSHOT"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return fmt.Errorf("unable to read %s: %w", pomFile, err)
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
				version.SetText(ftlVersion)
				versionSet = true
			}
		}
	}

	err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return err
	}

	err = tree.WriteToFile(pomFile)
	if err != nil {
		return fmt.Errorf("unable to write %s: %w", pomFile, err)
	}
	if err != nil {
		return fmt.Errorf("could not mark %s as modified: %w", pomFile, err)
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) error {
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)
	return nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*langpb.ErrorList, error) {
	errorsPath := filepath.Join(config.DeployDir, "errors.pb")
	if _, err := os.Stat(errorsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(errorsPath)
	if err != nil {
		return nil, fmt.Errorf("could not load build errors file: %w", err)
	}

	errorspb := &langpb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal build errors %w", err)
	}
	return errorspb, nil
}

func ptr(s string) *string {
	return &s
}

func (s *Service) writeGenericSchemaFiles(ctx context.Context, v *schema.Schema, config moduleconfig.AbsModuleConfig) error {

	if v == nil {
		return nil
	}
	logger := log.FromContext(ctx)
	modPath := filepath.Join(config.Dir, "src", "main", "ftl-module-schema")
	err := os.MkdirAll(modPath, 0750)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", modPath, err)
	}
	changed := false

	for _, mod := range v.Modules {
		if mod.Name == config.Module {
			continue
		}
		deps := v.ModuleDependencies(mod.Name)
		if deps[config.Module] != nil {
			continue
		}
		data, err := schema.ModuleToBytes(mod)
		if err != nil {
			return fmt.Errorf("failed to export module schema for module %s %w", mod.Name, err)
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
			return fmt.Errorf("failed to write schema file for module %s %w", mod.Name, err)
		}
	}
	if !changed {
		return nil
	}
	return nil
}
