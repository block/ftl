package buildengine

import (
	"context"
	"runtime"
	"time"

	"golang.org/x/exp/maps"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/pubsub"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"
)

// isBuildComplete returns true if the module is considered complete based on the deployAfterBuild flag
func isBuildComplete(event *buildenginepb.EngineEvent) bool {
	switch event.Event.(type) {
	case *buildenginepb.EngineEvent_ModuleBuildSuccess,
		*buildenginepb.EngineEvent_ModuleDeployWaiting,
		*buildenginepb.EngineEvent_ModuleDeployStarted,
		*buildenginepb.EngineEvent_ModuleDeploySuccess,
		*buildenginepb.EngineEvent_ModuleDeployFailed:
		return true
	default:
		return false
	}
}

// isModuleComplete returns true if the module is considered complete based on the deployAfterBuild flag
func isModuleComplete(event *buildenginepb.EngineEvent, deployAfterBuild bool) bool {
	switch event.Event.(type) {
	case *buildenginepb.EngineEvent_ModuleBuildSuccess:
		return !deployAfterBuild
	case *buildenginepb.EngineEvent_ModuleDeploySuccess:
		return deployAfterBuild
	case *buildenginepb.EngineEvent_ModuleBuildFailed,
		*buildenginepb.EngineEvent_ModuleDeployFailed:
		return true
	default:
		return false
	}
}

// moduleMeta is a wrapper around a module that includes the last build's start time.
type moduleMetaV2 struct {
	module           Module
	plugin           *languageplugin.LanguagePlugin
	configDefaults   moduleconfig.CustomDefaults
	state            *buildenginepb.EngineEvent
	deployAfterBuild bool
}

type EngineV2 struct {
	projectConfig projectconfig.Config
	parallelism   int
	moduleDirs    []string
	moduleMetas   *xsync.MapOf[string, *moduleMetaV2]
	builtModules  *xsync.MapOf[string, *schema.Module]
	stateChanges  *pubsub.Topic[StateChange]
	buildEnv      []string
	os            string
	arch          string
	devMode       bool

	// TODO: Can we remove this?
	devModeEndpointUpdates chan dev.LocalEndpoint
}

type EngineV2Option func(o *EngineV2)

func BuildEnvV2(env []string) EngineV2Option {
	return func(o *EngineV2) {
		o.buildEnv = env
	}
}

func ParallelismV2(n int) EngineV2Option {
	return func(o *EngineV2) {
		o.parallelism = n
	}
}

// StateChange represents a state change event for a specific module.
type StateChange struct {
	Module string
	Event  *buildenginepb.EngineEvent
}

// initializeModule handles the initialization of a single module, including:
// - Getting language-specific defaults
// - Validating and filling config defaults
// - Creating the language plugin
// - Creating the module instance
func NewV2(
	ctx context.Context,
	_ interface{}, // schemaSource placeholder
	projectConfig projectconfig.Config,
	moduleDirs []string,
	_ bool, // logChanges placeholder
	options ...EngineV2Option,
) (*EngineV2, error) {
	logger := log.FromContext(ctx).Scope("engine")
	ctx = log.ContextWithLogger(ctx, logger)

	e := &EngineV2{
		projectConfig: projectConfig,
		parallelism:   runtime.NumCPU(),
		moduleDirs:    moduleDirs,
		moduleMetas:   xsync.NewMapOf[string, *moduleMetaV2](),
		builtModules:  xsync.NewMapOf[string, *schema.Module](),
		stateChanges:  pubsub.New[StateChange](),
	}
	for _, option := range options {
		option(e)
	}

	// Discover modules
	configs, err := watch.DiscoverModules(ctx, moduleDirs)
	if err != nil {
		return nil, errors.Wrap(err, "could not discover modules")
	}

	err = CleanStubs(ctx, projectConfig.Root(), configs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to clean stubs")
	}

	logger.Infof("Initializing modules")

	wg := &errgroup.Group{}
	for _, config := range configs {
		wg.Go(func() error {
			return e.initModuleMeta(ctx, config)
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}

	go e.processChanges(ctx)

	return e, nil
}

// checkModuleState verifies that a module exists and its current state matches the expected state.
func (e *EngineV2) checkModuleState(ctx context.Context, module string, expectedEvent *buildenginepb.EngineEvent) (*moduleMetaV2, error) {
	logger := log.FromContext(ctx).Scope("engine")
	meta, ok := e.moduleMetas.Load(module)
	if !ok {
		logger.Warnf("Module %s not found", module)
		return nil, errors.Errorf("module %s not found", module)
	}
	if meta.state.Event != expectedEvent.Event {
		logger.Tracef("Module %s state mismatch: got %T(%+v), expected %T(%+v)", module, meta.state.Event, meta.state.Event, expectedEvent.Event, expectedEvent.Event)
		return nil, errors.Errorf("module %s state mismatch: got %T(%+v), expected %T(%+v)", module, meta.state.Event, meta.state.Event, expectedEvent.Event, expectedEvent.Event)
	}
	return meta, nil
}

func (e *EngineV2) processChanges(ctx context.Context) {
	ch := make(chan StateChange, 128)
	e.stateChanges.Subscribe(ch)
	logger := log.FromContext(ctx).Scope("engine")
	logger.Infof("Processing changes with parallelism %d", e.parallelism)

	wg := errgroup.Group{}
	wg.SetLimit(e.parallelism)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case stateChange := <-ch:
			wg.Go(func() error {
				logger := log.FromContext(ctx).Scope(stateChange.Module).Module(stateChange.Module)
				ctx = log.ContextWithLogger(ctx, logger)

				meta, err := e.checkModuleState(ctx, stateChange.Module, stateChange.Event)
				if err != nil {
					return err
				}

				switch stateChange.Event.Event.(type) {
				case *buildenginepb.EngineEvent_ModuleBuildWaiting:
					logger.Infof("Build waiting...")
					e.updateModuleState(ctx, stateChange.Module, &buildenginepb.EngineEvent{
						Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
							ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{},
						},
					})
				case *buildenginepb.EngineEvent_ModuleBuildStarted:
					logger.Infof("Building...")

					err := e.buildModule(ctx, meta)
					if err != nil {
						return err
					}
				case *buildenginepb.EngineEvent_ModuleBuildFailed:
					logger.Infof("Build failed")
				case *buildenginepb.EngineEvent_ModuleBuildSuccess:
					logger.Infof("Build success")
					dependentModules := e.getDependentModuleNames(stateChange.Module)
					for _, dependentModule := range dependentModules {
						e.tryStartBuild(ctx, dependentModule)
					}

					if meta.deployAfterBuild {
						e.updateModuleState(ctx, stateChange.Module, &buildenginepb.EngineEvent{
							Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
								ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{},
							},
						})
					}

				case *buildenginepb.EngineEvent_ModuleDeployWaiting:
					logger.Infof("Deploy waiting...")
				case *buildenginepb.EngineEvent_ModuleDeployStarted:
					logger.Infof("Deploying...")
					time.Sleep(1 * time.Second)
					e.updateModuleState(ctx, stateChange.Module, &buildenginepb.EngineEvent{
						Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
							ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{},
						},
					})
				case *buildenginepb.EngineEvent_ModuleDeployFailed:
					logger.Infof("Deploy failed")
				case *buildenginepb.EngineEvent_ModuleDeploySuccess:
					logger.Infof("Deploy success")
				default:
					logger.Infof("State change: %v", stateChange)
				}
				return nil
			})
		}
	}
}

func (e *EngineV2) buildModule(ctx context.Context, meta *moduleMetaV2) error {
	logger := log.FromContext(ctx)
	logger.Infof("Building module %s", meta.module.Config.Module)

	builtModules := make([]*schema.Module, 0)
	e.builtModules.Range(func(key string, module *schema.Module) bool {
		builtModules = append(builtModules, module)
		return true
	})

	logger.Infof("Built modules: %v", builtModules)

	sch := &schema.Schema{Realms: []*schema.Realm{{Modules: builtModules}}} //nolint:exptostd

	moduleSchema, _, _, err := build(ctx, e.projectConfig, meta.module, meta.plugin, languageplugin.BuildContext{
		Config:       meta.module.Config,
		Schema:       sch,
		Dependencies: meta.module.Dependencies(Raw),
		BuildEnv:     e.buildEnv,
		Os:           e.os,
		Arch:         e.arch,
	}, e.devMode, e.devModeEndpointUpdates)
	if err != nil {
		logger.Errorf(err, "Failed to build module %s", meta.module.Config.Module)
		return errors.Wrap(err, "failed to build module")
	}
	logger.Infof("Built module %s", meta.module.Config.Module)

	e.builtModules.Store(meta.module.Config.Module, moduleSchema)
	e.updateModuleState(ctx, meta.module.Config.Module, &buildenginepb.EngineEvent{
		Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
			ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{},
		},
	})

	return nil
}

// updateModuleState updates the state of a module and publishes the change.
func (e *EngineV2) updateModuleState(ctx context.Context, module string, event *buildenginepb.EngineEvent) error {
	logger := log.FromContext(ctx).Scope("engine")
	logger.Tracef("Updating module state: %s, %v", module, event)
	meta, ok := e.moduleMetas.Load(module)
	if !ok {
		return errors.Errorf("module %s not found", module)
	}
	meta.state = event
	e.moduleMetas.Store(module, meta)
	e.stateChanges.Publish(StateChange{
		Module: module,
		Event:  event,
	})
	return nil
}

// Main entry for testing
func (e *EngineV2) BuildV2(ctx context.Context, deployAfterBuild bool) error {
	logger := log.FromContext(ctx).Scope("engine")
	logger.Infof("Building modules")

	// Create a channel to receive state changes
	stateCh := make(chan StateChange, 64)
	e.stateChanges.Subscribe(stateCh)
	defer e.stateChanges.Unsubscribe(stateCh)

	// Track module build states
	moduleStates := make(map[string]bool)
	totalModules := 0
	completedModules := 0

	// Initialize module states and start builds
	e.moduleMetas.Range(func(key string, meta *moduleMetaV2) bool {
		moduleStates[key] = false
		totalModules++
		meta.deployAfterBuild = deployAfterBuild
		e.moduleMetas.Store(key, meta)
		e.tryStartBuild(ctx, key)
		return true
	})

	// Monitor state changes until all modules are complete
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("build cancelled: %w", ctx.Err())
		case stateChange := <-stateCh:
			module := stateChange.Module
			switch stateChange.Event.Event.(type) {
			case *buildenginepb.EngineEvent_ModuleBuildSuccess,
				*buildenginepb.EngineEvent_ModuleBuildFailed,
				*buildenginepb.EngineEvent_ModuleDeploySuccess,
				*buildenginepb.EngineEvent_ModuleDeployFailed:
				if !moduleStates[module] && isModuleComplete(stateChange.Event, deployAfterBuild) {
					moduleStates[module] = true
					completedModules++
					switch stateChange.Event.Event.(type) {
					case *buildenginepb.EngineEvent_ModuleBuildFailed:
						logger.Infof("Module %s build failed", module)
					case *buildenginepb.EngineEvent_ModuleDeployFailed:
						logger.Infof("Module %s deploy failed", module)
					}
				}
			}

			// Check if all modules are complete
			if completedModules == totalModules {
				// Verify if any modules failed
				hasFailures := false
				for module, completed := range moduleStates {
					if !completed {
						hasFailures = true
						logger.Warnf("Module %s did not complete successfully", module)
					}
				}
				if hasFailures {
					return errors.Errorf("build failed: one or more modules failed to build")
				}
				// Sleep to allow logger to flush
				time.Sleep(100 * time.Millisecond)
				if deployAfterBuild {
					logger.Infof("All modules built and deployed successfully")
				} else {
					logger.Infof("All modules built successfully")
				}
				return nil
			}
		}
	}
}

func (e *EngineV2) tryStartBuild(ctx context.Context, module string) error {
	logger := log.FromContext(ctx).Scope("engine")

	meta, ok := e.moduleMetas.Load(module)
	if !ok {
		return errors.Errorf("module %s not found", module)
	}

	dependencies := meta.module.Dependencies(AlwaysIncludeBuiltin)
	for _, dependency := range dependencies {
		if dependency == "builtin" {
			continue
		}
		meta, ok := e.moduleMetas.Load(dependency)
		if !ok {
			logger.Warnf("Dependency %s not found for %s", dependency, module)
			return nil
		}
		if !isBuildComplete(meta.state) {
			logger.Warnf("Dependency %s not built for %s", dependency, module)
			e.updateModuleState(ctx, module, &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
					ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{},
				},
			})
			return nil
		}
	}

	logger.Infof("Starting build for %s", module)

	e.updateModuleState(ctx, module, &buildenginepb.EngineEvent{
		Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
			ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{},
		},
	})

	return nil
}

func (e *EngineV2) initModuleMeta(ctx context.Context, config moduleconfig.UnvalidatedModuleConfig) error {
	plugin, err := languageplugin.New(ctx, config.Dir, config.Language, config.Module)
	if err != nil {
		return errors.Wrapf(err, "could not create plugin for %s", config.Module)
	}

	customDefaults, err := languageplugin.GetModuleConfigDefaults(ctx, config.Language, config.Dir)
	if err != nil {
		return errors.Wrapf(err, "could not get defaults provider for %s", config.Module)
	}
	validConfig, err := config.FillDefaultsAndValidate(customDefaults, e.projectConfig)
	if err != nil {
		return errors.Wrapf(err, "could not apply defaults for %s", config.Module)
	}
	meta := moduleMetaV2{
		module:         newModule(validConfig),
		plugin:         plugin,
		configDefaults: customDefaults,
		state: &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleAdded{
				ModuleAdded: &buildenginepb.ModuleAdded{
					Module: config.Module,
				},
			},
		},
		deployAfterBuild: true,
	}

	dependencies, err := meta.plugin.GetDependencies(ctx, meta.module.Config)
	if err != nil {
		return errors.Wrapf(err, "could not get dependencies for %v", meta.module.Config.Module)
	}

	meta.module = meta.module.CopyWithDependencies(dependencies)

	e.moduleMetas.Store(config.Module, &meta)
	e.stateChanges.Publish(StateChange{
		Module: config.Module,
		Event:  &buildenginepb.EngineEvent{Event: &buildenginepb.EngineEvent_ModuleAdded{ModuleAdded: &buildenginepb.ModuleAdded{Module: config.Module}}},
	})

	return nil
}

func (e *EngineV2) getDependentModuleNames(moduleName string) []string {
	dependentModuleNames := map[string]bool{}
	e.moduleMetas.Range(func(name string, meta *moduleMetaV2) bool {
		for _, dep := range meta.module.Dependencies(AlwaysIncludeBuiltin) {
			if dep == moduleName {
				dependentModuleNames[name] = true
			}
		}
		return true
	})
	return maps.Keys(dependentModuleNames)
}
