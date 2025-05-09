package languageplugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/alecthomas/types/result"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

const BuildLockTimeout = time.Minute

type BuildResult struct {
	StartTime time.Time

	Schema *schema.Module
	Errors []builderrors.Error

	// Files to deploy, relative to the module config's DeployDir
	Deploy []string

	// Whether the module needs to recalculate its dependencies
	InvalidateDependencies bool

	// Endpoint of an instance started by the plugin to use in dev mode
	DevEndpoint optional.Option[string]

	// File that the runner can use to pass info into the hot reload endpoint
	HotReloadEndpoint optional.Option[string]
	HotReloadVersion  optional.Option[int64]

	DebugPort int
}

// PluginEvent is used to notify of updates from the plugin.
//
//sumtype:decl
type PluginEvent interface {
	pluginEvent()
}

type PluginBuildEvent interface {
	PluginEvent
	ModuleName() string
}

// AutoRebuildStartedEvent is sent when the plugin starts an automatic rebuild.
type AutoRebuildStartedEvent struct {
	Module string
}

func (AutoRebuildStartedEvent) pluginEvent()         {}
func (e AutoRebuildStartedEvent) ModuleName() string { return e.Module }

// AutoRebuildEndedEvent is sent when the plugin ends an automatic rebuild.
type AutoRebuildEndedEvent struct {
	Module string
	Result result.Result[BuildResult]
}

func (AutoRebuildEndedEvent) pluginEvent()         {}
func (e AutoRebuildEndedEvent) ModuleName() string { return e.Module }

// PluginDiedEvent is sent when the plugin dies.
type PluginDiedEvent struct {
	// Plugins do not always have an associated module name, so we include the module
	Plugin *LanguagePlugin
	Error  error
}

func (PluginDiedEvent) pluginEvent() {}

// BuildContext contains contextual information needed to build.
//
// Any change to the build context would require a new build.
type BuildContext struct {
	Config       moduleconfig.ModuleConfig
	Schema       *schema.Schema
	Dependencies []string
	BuildEnv     []string
	Os           string
	Arch         string
}

var ErrPluginNotRunning = errors.New("language plugin no longer running")

// New creates a new language plugin from the given config.
func New(ctx context.Context, dir, language, name string) (p *LanguagePlugin, err error) {
	impl, err := newClientImpl(ctx, dir, language, name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return newPluginForTesting(ctx, impl), nil
}

func newPluginForTesting(ctx context.Context, client pluginClient) *LanguagePlugin {
	plugin := &LanguagePlugin{
		client:   client,
		commands: make(chan buildCommand, 64),
		updates:  pubsub.New[PluginEvent](),
	}

	var runCtx context.Context
	runCtx, plugin.cancel = context.WithCancelCause(ctx)
	go plugin.run(runCtx)
	go plugin.watchForCmdError(runCtx)

	return plugin
}

type buildCommand struct {
	BuildContext
	projectConfig        projectconfig.Config
	stubsRoot            string
	rebuildAutomatically bool

	startTime time.Time
	result    chan result.Result[BuildResult]
}

type LanguagePlugin struct {
	client pluginClient

	// cancels the run() context
	cancel context.CancelCauseFunc

	// commands to execute
	commands chan buildCommand

	updates *pubsub.Topic[PluginEvent]
}

// Kill stops the plugin and cleans up any resources.
func (p *LanguagePlugin) Kill() error {
	p.cancel(errors.Wrap(context.Canceled, "killing language plugin"))
	if err := p.client.kill(); err != nil {
		return errors.Wrap(err, "failed to kill language plugin")
	}
	return nil
}

// Updates topic for all update events from the plugin
// The same topic must be returned each time this method is called
func (p *LanguagePlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.updates
}

// GetDependencies returns the dependencies of the module.
func (p *LanguagePlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "could not convert module config to proto")
	}
	resp, err := p.client.getDependencies(ctx, connect.NewRequest(&langpb.GetDependenciesRequest{
		ModuleConfig: configProto,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dependencies from plugin")
	}
	return resp.Msg.Modules, nil
}

// GenerateStubs for the given module.
func (p *LanguagePlugin) GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error {
	moduleProto := module.ToProto()
	configProto, err := langpb.ModuleConfigToProto(moduleConfig.Abs())
	if err != nil {
		return errors.Wrap(err, "could not create proto for module config")
	}
	var nativeConfigProto *langpb.ModuleConfig
	if config, ok := nativeModuleConfig.Get(); ok {
		nativeConfigProto, err = langpb.ModuleConfigToProto(config.Abs())
		if err != nil {
			return errors.Wrap(err, "could not create proto for native module config")
		}
	}
	_, err = p.client.generateStubs(ctx, connect.NewRequest(&langpb.GenerateStubsRequest{
		Dir:                dir,
		Module:             moduleProto,
		ModuleConfig:       configProto,
		NativeModuleConfig: nativeConfigProto,
	}))
	if err != nil {
		return errors.Wrap(err, "plugin failed to generate stubs")
	}
	return nil
}

// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
// references to external modules, regardless of whether they are dependencies.
//
// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
// import the modules when users start reference them.
//
// It is optional to do anything with this call.
func (p *LanguagePlugin) SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string, view *schema.Schema) error {
	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return errors.Wrap(err, "could not create proto for native module config")
	}
	_, err = p.client.syncStubReferences(ctx, connect.NewRequest(&langpb.SyncStubReferencesRequest{
		StubsRoot:    dir,
		Modules:      moduleNames,
		ModuleConfig: configProto,
		Schema:       view.ToProto(),
	}))
	if err != nil {
		return errors.Wrap(err, "plugin failed to sync stub references")
	}
	return nil
}

// Build builds the module with the latest config and schema.
// In dev mode, plugin is responsible for automatically rebuilding as relevant files within the module change,
// and publishing these automatic builds updates to Updates().
func (p *LanguagePlugin) Build(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, bctx BuildContext, rebuildAutomatically bool) (BuildResult, error) {
	cmd := buildCommand{
		BuildContext:         bctx,
		projectConfig:        projectConfig,
		stubsRoot:            stubsRoot,
		rebuildAutomatically: rebuildAutomatically,
		startTime:            time.Now(),
		result:               make(chan result.Result[BuildResult]),
	}
	p.commands <- cmd
	select {
	case r := <-cmd.result:
		result, err := r.Result()
		if err != nil {
			return BuildResult{}, errors.WithStack(err) //nolint:wrapcheck
		}
		return result, nil
	case <-ctx.Done():
		return BuildResult{}, errors.Wrap(ctx.Err(), "error waiting for build to complete")
	}
}

func (p *LanguagePlugin) watchForCmdError(ctx context.Context) {
	select {
	case err := <-p.client.cmdErr():
		if err == nil {
			// closed
			return
		}
		p.updates.Publish(PluginDiedEvent{
			Plugin: p,
			Error:  err,
		})

	case <-ctx.Done():

	}
}

func (p *LanguagePlugin) run(ctx context.Context) {

	// State
	var bctx BuildContext
	var stubsRoot string
	var projectConfig *langpb.ProjectConfig

	// if a current build stream is active, this is non-nil
	// this does not indicate if the stream is listening to automatic rebuilds
	var streamChan chan result.Result[*langpb.BuildResponse]
	var streamCancel streamCancelFunc

	// if an explicit build command is active, this is non-nil
	// if this is nil, streamChan may still be open for automatic rebuilds
	var activeBuildCmd optional.Option[buildCommand]

	// if an automatic rebuild was started, this is the time it started
	var autoRebuildStartTime optional.Option[time.Time]

	// build counter is used to generate build request ids
	var contextCounter = 0

	// can not scope logger initially without knowing module name
	logger := log.FromContext(ctx)

	for {
		select {
		// Process incoming commands
		case c := <-p.commands:
			bctx = c.BuildContext
			projectConfig = langpb.ProjectConfigToProto(c.projectConfig)
			stubsRoot = c.stubsRoot

			// module name may have changed, update logger scope
			logger = log.FromContext(ctx).Scope(bctx.Config.Module)

			if _, ok := activeBuildCmd.Get(); ok {
				c.result <- result.Err[BuildResult](errors.Errorf("build already in progress"))
				continue
			}
			configProto, err := langpb.ModuleConfigToProto(bctx.Config.Abs())
			if err != nil {
				c.result <- result.Err[BuildResult](err)
				continue
			}

			schemaProto := bctx.Schema.ToProto()

			// update state
			contextCounter++
			if streamChan != nil {
				// tell plugin about new build context so that it rebuilds in existing build stream
				_, err = p.client.buildContextUpdated(ctx, connect.NewRequest(&langpb.BuildContextUpdatedRequest{
					BuildContext: &langpb.BuildContext{
						Id:           contextID(bctx.Config, contextCounter),
						ModuleConfig: configProto,
						Schema:       schemaProto,
						Dependencies: bctx.Dependencies,
						BuildEnv:     c.BuildEnv,
						Os:           c.Os,
						Arch:         c.Arch,
					},
				}))
				if err != nil {
					c.result <- result.Err[BuildResult](errors.Wrap(err, "failed to send updated build context to plugin"))
					continue
				}
				activeBuildCmd = optional.Some[buildCommand](c)
				continue
			}

			newStreamChan, newCancelFunc, err := p.client.build(ctx, connect.NewRequest(&langpb.BuildRequest{
				ProjectConfig:        projectConfig,
				StubsRoot:            stubsRoot,
				RebuildAutomatically: c.rebuildAutomatically,
				BuildContext: &langpb.BuildContext{
					Id:           contextID(bctx.Config, contextCounter),
					ModuleConfig: configProto,
					Schema:       schemaProto,
					Dependencies: bctx.Dependencies,
					BuildEnv:     c.BuildEnv,
					Os:           c.Os,
					Arch:         c.Arch,
				},
			}))
			if err != nil {
				c.result <- result.Err[BuildResult](errors.Wrap(err, "failed to start build stream"))
				continue
			}
			activeBuildCmd = optional.Some[buildCommand](c)
			streamChan = newStreamChan
			streamCancel = newCancelFunc

		// Receive messages from the current build stream
		case r := <-streamChan:
			e, err := r.Result()
			if err != nil {
				// Stream failed
				if c, ok := activeBuildCmd.Get(); ok {
					c.result <- result.Err[BuildResult](err)
					activeBuildCmd = optional.None[buildCommand]()
				}
				streamCancel = nil
				streamChan = nil
			}
			if e == nil {
				streamChan = nil
				streamCancel = nil
				continue
			}

			switch e.Event.(type) {
			case *langpb.BuildResponse_AutoRebuildStarted:
				if _, ok := activeBuildCmd.Get(); ok {
					logger.Debugf("ignoring automatic rebuild started during explicit build")
					continue
				}
				autoRebuildStartTime = optional.Some(time.Now())
				p.updates.Publish(AutoRebuildStartedEvent{
					Module: bctx.Config.Module,
				})
			case *langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure:
				streamEnded := false
				cmdEnded := false
				result, eventContextID, isAutomaticRebuild := getBuildSuccessOrFailure(e)
				if activeBuildCmd.Ok() == isAutomaticRebuild {
					if isAutomaticRebuild {
						logger.Debugf("ignoring automatic rebuild while expecting explicit build")
					} else {
						// This is likely a language plugin bug, but we can ignore it
						logger.Warnf("ignoring explicit build while none was requested")
					}
					continue
				} else if eventContextID != contextID(bctx.Config, contextCounter) {
					logger.Debugf("received build for outdated context %q; expected %q", eventContextID, contextID(bctx.Config, contextCounter))
					continue
				}

				var startTime time.Time
				if cmd, ok := activeBuildCmd.Get(); ok {
					startTime = cmd.startTime
				} else if t, ok := autoRebuildStartTime.Get(); ok {
					startTime = t
				} else {
					// Plugin did not declare when it started to build.
					startTime = time.Now()
				}
				streamEnded, cmdEnded = p.handleBuildResult(bctx.Config.Module, result, activeBuildCmd, startTime)
				if streamEnded {
					streamCancel()
					streamCancel = nil
					streamChan = nil
				}
				if cmdEnded {
					activeBuildCmd = optional.None[buildCommand]()
				}
			}

		case <-ctx.Done():
			if streamCancel != nil {
				streamCancel()
			}
			return
		}
	}
}

// getBuildSuccessOrFailure takes a BuildFailure or BuildSuccess event and returns the shared fields and an either wrapped result.
// This makes it easier to have some shared logic for both event types.
func getBuildSuccessOrFailure(e *langpb.BuildResponse) (result either.Either[*langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure], contextID string, isAutomaticRebuild bool) {
	switch e := e.Event.(type) {
	case *langpb.BuildResponse_BuildSuccess:
		return either.LeftOf[*langpb.BuildResponse_BuildFailure](e), e.BuildSuccess.ContextId, e.BuildSuccess.IsAutomaticRebuild
	case *langpb.BuildResponse_BuildFailure:
		return either.RightOf[*langpb.BuildResponse_BuildSuccess](e), e.BuildFailure.ContextId, e.BuildFailure.IsAutomaticRebuild
	default:
		panic(fmt.Sprintf("unexpected event type %T", e))
	}
}

// handleBuildResult processes the result of a build and publishes the appropriate events.
func (p *LanguagePlugin) handleBuildResult(module string, r either.Either[*langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure],
	activeBuildCmd optional.Option[buildCommand], startTime time.Time) (streamEnded, cmdEnded bool) {
	buildResult, err := buildResultFromProto(r, startTime)
	if cmd, ok := activeBuildCmd.Get(); ok {
		// handle explicit build
		cmd.result <- result.From(buildResult, err)

		cmdEnded = true
		if !cmd.rebuildAutomatically {
			streamEnded = true
		}
		return
	}
	// handle auto rebuild
	p.updates.Publish(AutoRebuildEndedEvent{
		Module: module,
		Result: result.From(buildResult, err),
	})
	return
}

func buildResultFromProto(result either.Either[*langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure], startTime time.Time) (buildResult BuildResult, err error) {
	switch result := result.(type) {
	case either.Left[*langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure]:
		buildSuccess := result.Get().BuildSuccess

		moduleSch, err := schema.ModuleFromProto(buildSuccess.Module)
		if err != nil {
			return BuildResult{}, errors.Wrap(err, "failed to unmarshal module from proto")
		}
		if moduleSch.Runtime != nil && len(strings.Split(moduleSch.Runtime.Base.Image, ":")) != 1 {
			return BuildResult{}, errors.Errorf("image tag not supported in runtime image: %s", moduleSch.Runtime.Base.Image)
		}

		errs := langpb.ErrorsFromProto(buildSuccess.Errors)
		builderrors.SortErrorsByPosition(errs)
		port := 0
		if buildSuccess.DebugPort != nil {
			port = int(*buildSuccess.DebugPort)
		}
		return BuildResult{
			Errors:            errs,
			Schema:            moduleSch,
			Deploy:            buildSuccess.Deploy,
			StartTime:         startTime,
			DevEndpoint:       optional.Ptr(buildSuccess.DevEndpoint),
			HotReloadEndpoint: optional.Ptr(buildSuccess.DevHotReloadEndpoint),
			HotReloadVersion:  optional.Ptr(buildSuccess.DevHotReloadVersion),
			DebugPort:         port,
		}, nil
	case either.Right[*langpb.BuildResponse_BuildSuccess, *langpb.BuildResponse_BuildFailure]:
		buildFailure := result.Get().BuildFailure

		errs := langpb.ErrorsFromProto(buildFailure.Errors)
		builderrors.SortErrorsByPosition(errs)

		if !builderrors.ContainsTerminalError(errs) {
			// This happens if the language plugin returns BuildFailure but does not include any errors with level ERROR.
			// Language plugins should always include at least one error with level ERROR in the case of a build failure.
			errs = append(errs, builderrors.Error{
				Msg:   "unexpected build failure without error level ERROR",
				Level: builderrors.ERROR,
				Type:  builderrors.FTL,
			})
		}

		return BuildResult{
			StartTime:              startTime,
			Errors:                 errs,
			InvalidateDependencies: buildFailure.InvalidateDependencies,
		}, nil
	default:
		panic(fmt.Sprintf("unexpected result type %T", result))
	}
}

func contextID(config moduleconfig.ModuleConfig, counter int) string {
	return fmt.Sprintf("%v-%v", config.Module, counter)
}
