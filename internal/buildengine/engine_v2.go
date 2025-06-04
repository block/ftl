package buildengine

import (
	"context"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/pubsub"
	"github.com/alecthomas/types/result"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/realm"
	"github.com/block/ftl/internal/schema/schemaeventsource"
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
	pendingDeploy    *pendingModule
}

type EngineV2 struct {
	adminClient    AdminClient
	projectConfig  projectconfig.Config
	parallelism    int
	moduleDirs     []string
	moduleMetas    *xsync.MapOf[string, *moduleMetaV2]
	builtModules   *xsync.MapOf[string, *schema.Module]
	stateChanges   *pubsub.Topic[StateChange]
	buildEnv       []string
	os             string
	arch           string
	devMode        bool
	externalRealms []*schema.Realm
	schemaSource   *schemaeventsource.EventSource
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

// WithDevMode sets the engine to dev mode.
func WithDevModeV2(updates chan dev.LocalEndpoint) EngineV2Option {
	return func(o *EngineV2) {
		o.devModeEndpointUpdates = updates
		o.devMode = true
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
	adminClient AdminClient,
	schemaSource *schemaeventsource.EventSource,
	projectConfig projectconfig.Config,
	moduleDirs []string,
	_ bool, // logChanges placeholder
	options ...EngineV2Option,

) (*EngineV2, error) {
	logger := log.FromContext(ctx).Scope("engine")
	ctx = log.ContextWithLogger(ctx, logger)

	e := &EngineV2{
		adminClient:    adminClient,
		projectConfig:  projectConfig,
		parallelism:    runtime.NumCPU(),
		moduleDirs:     moduleDirs,
		moduleMetas:    xsync.NewMapOf[string, *moduleMetaV2](),
		builtModules:   xsync.NewMapOf[string, *schema.Module](),
		stateChanges:   pubsub.New[StateChange](),
		externalRealms: []*schema.Realm{},
		schemaSource:   schemaSource,
		arch:           runtime.GOARCH, // Default to the local env, we attempt to read these from the cluster later
		os:             runtime.GOOS,
	}
	for _, option := range options {
		option(e)
	}

	updateTerminalWithEngineEventsV2(ctx, e.stateChanges)

	// Ensure schema sync at startup if we have an admin client
	if e.adminClient != nil {
		info, err := adminClient.ClusterInfo(ctx, connect.NewRequest(&adminpb.ClusterInfoRequest{}))
		if err != nil {
			log.FromContext(ctx).Debugf("failed to get cluster info: %s", err)
		} else {
			e.os = info.Msg.Os
			e.arch = info.Msg.Arch
		}
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

	jvm := false
	wg := &errgroup.Group{}
	for _, config := range configs {
		if config.Language == "java" || config.Language == "kotlin" {
			jvm = true
		}
		wg.Go(func() error {
			return e.initModuleMeta(ctx, config)
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}

	if jvm {
		// Huge hack that is just for development
		// In release builds this is a noop
		// This makes sure the JVM jars are up to date when running from source
		buildRequiredJARS(ctx)
	}

	// Initialize builtModules with builtins
	builtModules := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	// Create metasMap from moduleMetas
	metasMap := map[string]moduleMeta{}
	e.moduleMetas.Range(func(name string, meta *moduleMetaV2) bool {
		metasMap[name] = moduleMeta{
			module: meta.module,
			plugin: meta.plugin,
		}
		return true
	})

	// Generate stubs for builtins after all modules are initialized
	logger.Infof("Generating stubs for builtins module")
	err = GenerateStubs(ctx, e.projectConfig.Root(), maps.Values(builtModules), metasMap)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for name, cfg := range projectConfig.ExternalRealms {
		realm, err := realm.GetExternalRealm(ctx, e.projectConfig.ExternalRealmPath(), name, cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read external realm %s", name)
		}
		e.externalRealms = append(e.externalRealms, realm)
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
				case *buildenginepb.EngineEvent_ModuleBuildStarted:
					logger.Infof("Building...")

					err := e.buildModule(ctx, meta)
					if err != nil {
						return e.handleModuleError(ctx, stateChange, err)
					}
				case *buildenginepb.EngineEvent_ModuleBuildFailed:
					logger.Infof("Build failed")
				case *buildenginepb.EngineEvent_ModuleBuildSuccess:
					logger.Infof("Build success")

					dependentModules := e.getDependentModuleNames(stateChange.Module)
					for _, dependentModule := range dependentModules {
						e.startBuild(ctx, dependentModule)
					}

					if meta.deployAfterBuild {
						e.updateModuleState(ctx, stateChange.Module, &buildenginepb.EngineEvent{
							Timestamp: timestamppb.Now(),
							Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
								ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
									Module: meta.module.Config.Module,
								},
							},
						})
					}

				case *buildenginepb.EngineEvent_ModuleDeployWaiting:
					logger.Infof("Deploy waiting...")
				case *buildenginepb.EngineEvent_ModuleDeployStarted:
					logger.Infof("Deploying...")
					err := e.deployModule(ctx, meta)
					if err != nil {
						return e.handleModuleError(ctx, stateChange, err)
					}
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

	// Ensure all dependencies are built and their stubs are generated
	dependencies := meta.module.Dependencies(AlwaysIncludeBuiltin)
	for _, dep := range dependencies {
		if dep == "builtin" {
			continue
		}
		depMeta, ok := e.moduleMetas.Load(dep)
		if !ok {
			return errors.Errorf("dependency %s not found", dep)
		}
		if !isBuildComplete(depMeta.state) {
			logger.Debugf("Dependency %s not built", dep)
			return nil
		}
	}

	builtModules := make([]*schema.Module, 0)
	e.builtModules.Range(func(key string, module *schema.Module) bool {
		builtModules = append(builtModules, module)
		return true
	})

	sch := &schema.Schema{Realms: []*schema.Realm{{Modules: builtModules}}} //nolint:exptostd

	moduleSchema, tmpDeployDir, deployPaths, err := build(ctx, e.projectConfig, meta.module, meta.plugin, languageplugin.BuildContext{
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

	// Generate stubs for the successfully built module
	metasMap := map[string]moduleMeta{}
	e.moduleMetas.Range(func(name string, meta *moduleMetaV2) bool {
		metasMap[name] = moduleMeta{
			module: meta.module,
			plugin: meta.plugin,
		}
		return true
	})
	err = GenerateStubs(ctx, e.projectConfig.Root(), []*schema.Module{moduleSchema}, metasMap)
	if err != nil {
		logger.Errorf(err, "Failed to generate stubs for module %s", meta.module.Config.Module)
		return err
	}

	configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
	if err != nil {
		logger.Errorf(err, "Failed to marshal module config")
		return err
	}

	pendingDeploy := newPendingModule(meta.module, tmpDeployDir, deployPaths, moduleSchema)
	meta.pendingDeploy = pendingDeploy
	e.moduleMetas.Store(meta.module.Config.Module, meta)
	e.builtModules.Store(meta.module.Config.Module, moduleSchema)

	return e.updateModuleState(ctx, meta.module.Config.Module, &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
			ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
				Config: configProto,
			},
		},
	})
}

func (e *EngineV2) deployModule(ctx context.Context, meta *moduleMetaV2) error {
	logger := log.FromContext(ctx)
	logger.Infof("Deploying module %s", meta.module.Config.Module)

	sch, ok := e.builtModules.Load(meta.module.Config.Module)
	if !ok {
		return errors.Errorf("module %s not found", meta.module.Config.Module)
	}

	if sch.Runtime == nil {
		sch.Runtime = &schema.ModuleRuntime{
			Base: schema.ModuleRuntimeBase{
				CreateTime: time.Now(),
			},
		}
	}
	sch.Runtime.Base.Language = meta.module.Config.Language

	// Upload artifacts first
	if err := uploadArtefacts(ctx, meta.pendingDeploy, e.adminClient); err != nil {
		return errors.WithStack(err)
	}

	sch.Metadata = meta.pendingDeploy.schema.Metadata
	e.builtModules.Store(meta.module.Config.Module, sch)

	// Deploy the module
	keyChan := make(chan result.Result[key.Changeset], 1)
	if err := deploy(ctx, e.projectConfig.Name, []*schema.Module{sch}, e.adminClient, keyChan, e.externalRealms); err != nil {
		return errors.WithStack(err)
	}

	// Handle deployment completion
	if key, ok := (<-keyChan).Get(); ok {
		logger.Debugf("Created changeset %s for module %s", key, meta.module.Config.Module)
	}

	return e.updateModuleState(ctx, meta.module.Config.Module, &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
			ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
				Module: meta.module.Config.Module,
			},
		},
	})
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

func (e *EngineV2) handleModuleError(ctx context.Context, stateChange StateChange, err error) error {
	logger := log.FromContext(ctx).Scope(stateChange.Module)
	logger.Errorf(err, "Error in state %T", stateChange.Event.Event)
	var event *buildenginepb.EngineEvent

	// You may need to get the config proto for the module
	meta, ok := e.moduleMetas.Load(stateChange.Module)
	var configProto *langpb.ModuleConfig
	if ok {
		configProto, err = langpb.ModuleConfigToProto(meta.module.Config.Abs())
		if err != nil {
			logger.Errorf(err, "Failed to marshal module config")
			return err
		}
	}

	switch stateChange.Event.Event.(type) {
	case *buildenginepb.EngineEvent_ModuleBuildStarted:
		event = &buildenginepb.EngineEvent{
			Timestamp: timestamppb.Now(),
			Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
				ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
					Config:        configProto,
					IsAutoRebuild: false,
					Errors: &langpb.ErrorList{
						Errors: errorToLangError(err),
					},
				},
			},
		}
	case *buildenginepb.EngineEvent_ModuleDeployStarted:
		event = &buildenginepb.EngineEvent{
			Timestamp: timestamppb.Now(),
			Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
				ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
					Module: meta.module.Config.Module,
					Errors: &langpb.ErrorList{
						Errors: errorToLangError(err),
					},
				},
			},
		}
	default:
		return err
	}

	return e.updateModuleState(ctx, stateChange.Module, event)
}

// BuildV2 builds all modules and optionally deploys them.
func (e *EngineV2) BuildV2(ctx context.Context, deployAfterBuild bool, waitForDeployOnline bool) error {
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
		e.startBuild(ctx, key)
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

func (e *EngineV2) startBuild(ctx context.Context, module string) error {
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
			logger.Debugf("Dependency %s not built for %s", dependency, module)
			configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
			if err != nil {
				logger.Errorf(err, "Failed to marshal module config")
				return err
			}
			e.updateModuleState(ctx, module, &buildenginepb.EngineEvent{
				Timestamp: timestamppb.Now(),
				Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
					ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
						Config: configProto,
					},
				},
			})
			return nil
		}
	}

	logger.Infof("Starting build for %s", module)

	configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
	if err != nil {
		logger.Errorf(err, "Failed to marshal module config")
		return err
	}
	e.updateModuleState(ctx, module, &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
			ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
				Config: configProto,
			},
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
			Timestamp: timestamppb.Now(),
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
		Event:  meta.state,
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
