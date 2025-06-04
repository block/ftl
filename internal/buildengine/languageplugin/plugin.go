package languageplugin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
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
	ModifiedFiles     []string

	DebugPort           int
	redeployNotRequired bool
}

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
		client:       client,
		buildRunning: &sync.Mutex{},
	}
	go plugin.watchForCmdError(ctx)

	return plugin
}

type LanguagePlugin struct {
	client pluginClient

	buildRunning *sync.Mutex
}

// Kill stops the plugin and cleans up any resources.
func (p *LanguagePlugin) Kill() error {
	if p == nil {
		return nil
	}
	if err := p.client.kill(); err != nil {
		return errors.Wrap(err, "failed to kill language plugin")
	}
	return nil
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
func (p *LanguagePlugin) Build(ctx context.Context, projectConfig projectconfig.Config, stubsRoot string, bctx BuildContext, devModeBuild bool) (BuildResult, error) {
	p.buildRunning.Lock()
	defer p.buildRunning.Unlock()
	startTime := time.Now()

	configProto, err := langpb.ModuleConfigToProto(bctx.Config.Abs())
	if err != nil {
		return BuildResult{}, errors.Wrapf(err, "failed to marshal module config")
	}

	schemaProto := bctx.Schema.ToProto()
	result, err := p.client.build(ctx, connect.NewRequest(&langpb.BuildRequest{
		ProjectConfig: langpb.ProjectConfigToProto(projectConfig),
		StubsRoot:     stubsRoot,
		DevModeBuild:  devModeBuild,
		BuildContext: &langpb.BuildContext{
			ModuleConfig: configProto,
			Schema:       schemaProto,
			Dependencies: bctx.Dependencies,
			BuildEnv:     bctx.BuildEnv,
			Os:           bctx.Os,
			Arch:         bctx.Arch,
		},
	}))

	if err != nil {
		return BuildResult{}, errors.WithStack(err)
	}
	return buildResultFromProto(result.Msg, startTime)
}

func buildResultFromProto(result *langpb.BuildResponse, startTime time.Time) (buildResult BuildResult, err error) {
	switch et := result.Event.(type) {
	case *langpb.BuildResponse_BuildSuccess:
		buildSuccess := et.BuildSuccess

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
			Errors:              errs,
			Schema:              moduleSch,
			Deploy:              buildSuccess.Deploy,
			StartTime:           startTime,
			DevEndpoint:         optional.Ptr(buildSuccess.DevEndpoint),
			HotReloadEndpoint:   optional.Ptr(buildSuccess.DevHotReloadEndpoint),
			HotReloadVersion:    optional.Ptr(buildSuccess.DevHotReloadVersion),
			DebugPort:           port,
			ModifiedFiles:       buildSuccess.ModifiedFiles,
			redeployNotRequired: buildSuccess.RedeployNotRequired,
		}, nil
	case *langpb.BuildResponse_BuildFailure:
		buildFailure := et.BuildFailure

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
			ModifiedFiles:          buildFailure.ModifiedFiles,
		}, nil
	default:
		panic(fmt.Sprintf("unexpected result type %T", result))
	}
}

type buildInfo struct {
	projectConfig projectconfig.Config
	stubsRoot     string
	bctx          BuildContext
}

func (p *LanguagePlugin) watchForCmdError(ctx context.Context) {
	select {
	case err := <-p.client.cmdErr():
		if err == nil {
			// closed
			return
		}
		// TODO: handle this
		// p.updates.Publish(PluginDiedEvent{
		// 	Plugin: p,
		// 	Error:  err,
		// })

	case <-ctx.Done():

	}
}
