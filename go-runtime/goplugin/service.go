package goplugin

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/go-runtime/compile"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute

//sumtype:decl
type updateEvent interface{ updateEvent() }

type filesUpdatedEvent struct {
	changes []watch.FileChange
}

func (filesUpdatedEvent) updateEvent() {}

// buildContext contains contextual information needed to build.
type buildContext struct {
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
	BuildEnv     []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, errors.Wrap(err, "could not parse schema from proto")
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	bc := buildContext{
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
	updatesTopic *pubsub.Topic[updateEvent]
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

// GetDependencies extracts dependencies for a module
func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	deps, err := compile.ExtractDependencies(config)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract dependencies")
	}
	return connect.NewResponse(&langpb.GetDependenciesResponse{
		Modules: deps,
	}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.Module.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	moduleSchema, err := schema.ValidatedModuleFromProto(req.Msg.Module)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	var nativeConfig optional.Option[moduleconfig.AbsModuleConfig]
	if req.Msg.NativeModuleConfig != nil {
		nativeConfig = optional.Some(langpb.ModuleConfigFromProto(req.Msg.NativeModuleConfig))
	}

	err = compile.GenerateStubs(ctx, req.Msg.Dir, moduleSchema, config, nativeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate stubs")
	}
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.ModuleConfig.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	err := compile.SyncGeneratedStubReferences(ctx, config, req.Msg.StubsRoot, req.Msg.Modules)
	if err != nil {
		return nil, errors.Wrap(err, "could not sync stub references")
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
func (s *Service) Build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (*connect.Response[langpb.BuildResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.BuildContext.ModuleConfig.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	events := make(chan updateEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)

	// cancel context when stream ends so that watcher can be stopped
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "build stream ended"))

	projectConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	watchPatterns, err := relativeWatchPatterns(buildCtx.Config.Dir, buildCtx.Config.Watch)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	watcher := watch.NewWatcher(optional.None[string](), watchPatterns...)

	ongoingState := &compile.OngoingState{}

	// Initial build
	return buildAndSend(ctx, projectConfig, req.Msg.StubsRoot, buildCtx, false, watcher.GetTransaction(buildCtx.Config.Dir), ongoingState, req.Msg.DevModeBuild)

}

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, buildCtx buildContext,
	isAutomaticRebuild bool, transaction watch.ModifyFilesTransaction, ongoingState *compile.OngoingState, devMode bool) (*connect.Response[langpb.BuildResponse], error) {
	buildEvent, err := build(ctx, projectConfig, stubsRoot, buildCtx, isAutomaticRebuild, transaction, ongoingState, devMode)
	if err != nil {
		buildEvent = buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		})
	}
	return connect.NewResponse(buildEvent), nil
}

func build(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, buildCtx buildContext, isAutomaticRebuild bool, transaction watch.ModifyFilesTransaction,
	ongoingState *compile.OngoingState, devMode bool) (*langpb.BuildResponse, error) {
	release, err := flock.Acquire(ctx, buildCtx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not acquire build lock")
	}
	defer release() //nolint:errcheck

	m, invalidateDeps, buildErrs := compile.Build(ctx, projectConfig, stubsRoot, buildCtx.Config, buildCtx.Schema, buildCtx.Dependencies, buildCtx.BuildEnv, transaction, ongoingState, devMode)
	if invalidateDeps {
		return &langpb.BuildResponse{
			Event: &langpb.BuildResponse_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					InvalidateDependencies: true,
				},
			},
		}, nil
	}
	if _, hasErrs := slices.Find(buildErrs, func(e builderrors.Error) bool { //nolint:errcheck
		return e.Level == builderrors.ERROR
	}); hasErrs {
		return buildFailure(buildCtx, isAutomaticRebuild, buildErrs...), nil
	}
	module, ok := m.Get()
	if !ok {
		return buildFailure(buildCtx, isAutomaticRebuild, buildErrs...), nil
	}
	if _, err := schema.ValidateModuleInSchema(buildCtx.Schema, optional.Some(module)); err != nil {
		return buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		}), nil
	}

	moduleProto := module.ToProto()
	deploy := []string{"main", "launch"}
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				Errors: langpb.ErrorsToProto(buildErrs),
				Module: moduleProto,
				Deploy: deploy,
			},
		},
	}, nil
}

// buildFailure creates a BuildFailure event based on build errors.
func buildFailure(buildCtx buildContext, isAutomaticRebuild bool, errs ...builderrors.Error) *langpb.BuildResponse {
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				Errors:                 langpb.ErrorsToProto(errs),
				InvalidateDependencies: false,
			},
		},
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
