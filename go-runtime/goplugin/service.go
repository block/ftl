package goplugin

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/block/scaffolder"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/glob"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	goruntime "github.com/block/ftl/go-runtime"
	"github.com/block/ftl/go-runtime/compile"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute

//sumtype:decl
type updateEvent interface{ updateEvent() }

type buildContextUpdatedEvent struct {
	buildCtx buildContext
}

func (buildContextUpdatedEvent) updateEvent() {}

type filesUpdatedEvent struct {
	changes []watch.FileChange
}

func (filesUpdatedEvent) updateEvent() {}

// buildContext contains contextual information needed to build.
type buildContext struct {
	ID           string
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
	BuildEnv     []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, fmt.Errorf("could not parse schema from proto: %w", err)
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	bc := buildContext{
		ID:           proto.Id,
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
		BuildEnv:     proto.BuildEnv,
	}
	arch := false
	os := false
	for _, i := range bc.BuildEnv {
		if strings.HasPrefix(i, "GOARCH=") {
			arch = true
		} else if strings.HasPrefix(i, "GOOS=") {
			os = true
		}
	}
	if !arch {
		bc.BuildEnv = append(bc.BuildEnv, "GOARCH="+proto.Arch)
	}
	if !os {
		bc.BuildEnv = append(bc.BuildEnv, "GOOS="+proto.Os)
	}

	return bc, nil
}

type Service struct {
	updatesTopic          *pubsub.Topic[updateEvent]
	acceptsContextUpdates atomic.Value[bool]
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New() *Service {
	return &Service{
		updatesTopic: pubsub.New[updateEvent](),
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	log.FromContext(ctx).Debugf("Received Ping")
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetCreateModuleFlagsResponse{
		Flags: []*langpb.GetCreateModuleFlagsResponse_Flag{
			{
				Name:        "replace",
				Help:        "Replace a module import path with a local path in the initialised FTL module.",
				Envar:       optional.Some("FTL_INIT_GO_REPLACE").Ptr(),
				Short:       optional.Some("r").Ptr(),
				Placeholder: optional.Some("OLD=NEW,...").Ptr(),
			},
		},
	}), nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

// CreateModule generates files for a new module with the requested name
func (s *Service) CreateModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	logger := log.FromContext(ctx)
	flags := req.Msg.Flags.AsMap()
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := scaffoldingContext{
		Name:      req.Msg.Name,
		GoVersion: runtime.Version()[2:],
		Replace:   map[string]string{},
	}
	if replaceValue, ok := flags["replace"]; ok && replaceValue != "" {
		replaceStr, ok := replaceValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid replace flag is not a string: %v", replaceValue)
		}
		for _, replace := range strings.Split(replaceStr, ",") {
			parts := strings.Split(replace, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid replace flag (format: A=B,C=D): %q", replace)
			}
			sctx.Replace[parts[0]] = parts[1]
		}
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy: %s", req.Msg.Dir)
	if err := exec.Command(ctx, log.Debug, req.Msg.Dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return nil, fmt.Errorf("could not tidy: %w", err)
	}
	return connect.NewResponse(&langpb.CreateModuleResponse{}), nil
}

// ModuleConfigDefaults provides default values for ModuleConfig for values that are not configured in the ftl.toml file.
func (s *Service) ModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	deployDir := ".ftl"
	watch := []string{"**/*.go", "**/*.sql", "go.mod", "go.sum"}
	additionalWatch, err := replacementWatches(req.Msg.Dir, deployDir)
	watch = append(watch, additionalWatch...)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&langpb.ModuleConfigDefaultsResponse{
		Watch:           watch,
		DeployDir:       deployDir,
		SqlMigrationDir: "db/schema",
		SqlQueryDir:     "db/queries",
	}), nil
}

// GetDependencies extracts dependencies for a module
func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	deps, err := compile.ExtractDependencies(config)
	if err != nil {
		return nil, fmt.Errorf("could not extract dependencies: %w", err)
	}
	return connect.NewResponse(&langpb.GetDependenciesResponse{
		Modules: deps,
	}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	moduleSchema, err := schema.ValidatedModuleFromProto(req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("invalid module: %w", err)
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	var nativeConfig optional.Option[moduleconfig.AbsModuleConfig]
	if req.Msg.NativeModuleConfig != nil {
		nativeConfig = optional.Some(langpb.ModuleConfigFromProto(req.Msg.NativeModuleConfig))
	}

	err = compile.GenerateStubs(ctx, req.Msg.Dir, moduleSchema, config, nativeConfig)
	if err != nil {
		return nil, fmt.Errorf("could not generate stubs: %w", err)
	}
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	err := compile.SyncGeneratedStubReferences(ctx, config, req.Msg.StubsRoot, req.Msg.Modules)
	if err != nil {
		return nil, fmt.Errorf("could not sync stub references: %w", err)
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
	events := make(chan updateEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)

	// cancel context when stream ends so that watcher can be stopped
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(fmt.Errorf("build stream ended: %w", context.Canceled))

	projectConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return err
	}

	watchPatterns, err := relativeWatchPatterns(buildCtx.Config.Dir, buildCtx.Config.Watch)
	if err != nil {
		return err
	}
	watcher := watch.NewWatcher(optional.None[string](), watchPatterns...)

	ongoingState := &compile.OngoingState{}

	if req.Msg.RebuildAutomatically {
		s.acceptsContextUpdates.Store(true)
		defer s.acceptsContextUpdates.Store(false)

		if err := watchFiles(ctx, watcher, buildCtx, events, watchPatterns); err != nil {
			return err
		}
	}

	// Initial build
	if err := buildAndSend(ctx, stream, projectConfig, req.Msg.StubsRoot, buildCtx, false, watcher.GetTransaction(buildCtx.Config.Dir), ongoingState, req.Msg.RebuildAutomatically); err != nil {
		return err
	}
	if !req.Msg.RebuildAutomatically {
		return nil
	}

	// Watch for changes and build as needed
	for e := range channels.IterContext(ctx, events) {
		var isAutomaticRebuild bool
		buildCtx, isAutomaticRebuild = buildContextFromPendingEvents(ctx, buildCtx, events, e, ongoingState)
		if isAutomaticRebuild {
			err = stream.Send(&langpb.BuildResponse{
				Event: &langpb.BuildResponse_AutoRebuildStarted{
					AutoRebuildStarted: &langpb.AutoRebuildStarted{
						ContextId: buildCtx.ID,
					},
				},
			})
			if err != nil {
				return fmt.Errorf("could not send auto rebuild started event: %w", err)
			}
		}
		if err = buildAndSend(ctx, stream, projectConfig, req.Msg.StubsRoot, buildCtx, isAutomaticRebuild, watcher.GetTransaction(buildCtx.Config.Dir), ongoingState, true); err != nil {
			return err
		}
	}
	if ctx.Err() != nil {
		log.FromContext(ctx).Infof("Build call ending - ctx cancelled")
	}
	return nil
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
	s.updatesTopic.Publish(buildContextUpdatedEvent{
		buildCtx: buildCtx,
	})
	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

func watchFiles(ctx context.Context, watcher *watch.Watcher, buildCtx buildContext, events chan updateEvent, watchPatterns []string) error {
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
				// this event stream is built with different watch patterns than the plugin, so 
				// it may include changes to files which are not watched by this plugin. we should
				// exclude those.
				changes := slices.Filter(change.Changes, func(c watch.FileChange) bool {
					return glob.MatchAny(watchPatterns, c.Path)
				})
				if len(changes) > 0 {
					log.FromContext(ctx).Infof("Found file changes: %s", change)
					events <- filesUpdatedEvent{changes: changes}
				}
			}
		}
	}()
	return nil
}

// buildContextFromPendingEvents processes all pending events to determine the latest context and whether the build is automatic.
func buildContextFromPendingEvents(ctx context.Context, buildCtx buildContext, events chan updateEvent, firstEvent updateEvent, ongoingState *compile.OngoingState) (newBuildCtx buildContext, isAutomaticRebuild bool) {
	allEvents := []updateEvent{firstEvent}
	// find any other events in the queue
	for {
		select {
		case e := <-events:
			allEvents = append(allEvents, e)
		case <-ctx.Done():
			return buildCtx, false
		default:
			// No more events waiting to be processed
			hasExplicitBuilt := false
			for _, e := range allEvents {
				switch e := e.(type) {
				case buildContextUpdatedEvent:
					buildCtx = e.buildCtx
					hasExplicitBuilt = true
				case filesUpdatedEvent:
					ongoingState.DetectedFileChanges(buildCtx.Config, e.changes)
				}
			}
			switch e := firstEvent.(type) {
			case buildContextUpdatedEvent:
				buildCtx = e.buildCtx
				hasExplicitBuilt = true
			case filesUpdatedEvent:
				ongoingState.DetectedFileChanges(buildCtx.Config, e.changes)
			}
			return buildCtx, !hasExplicitBuilt
		}
	}
}

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, stream *connect.ServerStream[langpb.BuildResponse], projectConfig projectconfig.Config, stubsRoot string, buildCtx buildContext,
	isAutomaticRebuild bool, transaction watch.ModifyFilesTransaction, ongoingState *compile.OngoingState, devMode bool) error {
	buildEvent, err := build(ctx, projectConfig, stubsRoot, buildCtx, isAutomaticRebuild, transaction, ongoingState, devMode)
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

func build(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, buildCtx buildContext, isAutomaticRebuild bool, transaction watch.ModifyFilesTransaction,
	ongoingState *compile.OngoingState, devMode bool) (*langpb.BuildResponse, error) {
	release, err := flock.Acquire(ctx, buildCtx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, fmt.Errorf("could not acquire build lock: %w", err)
	}
	defer release() //nolint:errcheck

	m, buildErrs, err := compile.Build(ctx, projectConfig, stubsRoot, buildCtx.Config, buildCtx.Schema, buildCtx.Dependencies, buildCtx.BuildEnv, transaction, ongoingState, devMode)
	if err != nil {
		if errors.Is(err, compile.ErrInvalidateDependencies) {
			return &langpb.BuildResponse{
				Event: &langpb.BuildResponse_BuildFailure{
					BuildFailure: &langpb.BuildFailure{
						ContextId:              buildCtx.ID,
						IsAutomaticRebuild:     isAutomaticRebuild,
						InvalidateDependencies: true,
					},
				},
			}, nil
		}
		return buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.COMPILER,
			Level: builderrors.ERROR,
			Msg:   "compile: " + err.Error(),
		}), nil
	}
	module, ok := m.Get()
	if !ok {
		return buildFailure(buildCtx, isAutomaticRebuild, buildErrs...), nil
	}

	moduleProto := module.ToProto()
	deploy := []string{"main", "launch"}
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				ContextId:          buildCtx.ID,
				IsAutomaticRebuild: isAutomaticRebuild,
				Errors:             langpb.ErrorsToProto(buildErrs),
				Module:             moduleProto,
				Deploy:             deploy,
			},
		},
	}, nil
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

func relativeWatchPatterns(moduleDir string, watchPaths []string) ([]string, error) {
	relativePaths := make([]string, len(watchPaths))
	for i, path := range watchPaths {
		var relative string
		var err error
		if strings.HasPrefix(path, "!") {
			relative, err = filepath.Rel(moduleDir, strings.TrimPrefix(path, "!"))
			relative = "!" + relative
		} else {
			relative, err = filepath.Rel(moduleDir, path)
		}
		if err != nil {
			return nil, fmt.Errorf("could create relative path for watch pattern: %w", err)
		}
		relativePaths[i] = relative
	}
	return relativePaths, nil
}
