package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/watch"
)

// moduleMeta is a wrapper around a module that includes the last build's start time.
type moduleMeta struct {
	module         Module
	plugin         *languageplugin.LanguagePlugin
	events         chan languageplugin.PluginEvent
	configDefaults moduleconfig.CustomDefaults
}

// copyMetaWithUpdatedDependencies finds the dependencies for a module and returns a
// copy with those dependencies populated.
func copyMetaWithUpdatedDependencies(ctx context.Context, m moduleMeta) (moduleMeta, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting dependencies for %q", m.module.Config.Module)

	dependencies, err := m.plugin.GetDependencies(ctx, m.module.Config)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not get dependencies for %v: %w", m.module.Config.Module, err)
	}

	m.module = m.module.CopyWithDependencies(dependencies)
	return m, nil
}

//sumtype:decl
type rebuildEvent interface {
	rebuildEvent()
}

// rebuildRequestEvent is published when a module needs to be rebuilt when a module
// failed to build due to a change in dependencies.
type rebuildRequestEvent struct {
	module string
}

func (rebuildRequestEvent) rebuildEvent() {}

// rebuildRequiredEvent is published when a module needs to be rebuilt when a module
// failed to build due to a change in dependencies.
type autoRebuildCompletedEvent struct {
	module string
	schema *schema.Module
}

type pendingDeploy struct {
	modules  []Module
	err      chan error
	replicas int32
}

func (autoRebuildCompletedEvent) rebuildEvent() {}

// Engine for building a set of modules.
type Engine struct {
	adminClient      AdminClient
	schemaSource     *schemaeventsource.EventSource
	moduleMetas      *xsync.MapOf[string, moduleMeta]
	projectConfig    projectconfig.Config
	moduleDirs       []string
	watcher          *watch.Watcher // only watches for module toml changes
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schema.Notification]
	cancel           context.CancelCauseFunc
	parallelism      int
	modulesToBuild   *xsync.MapOf[string, bool]
	buildEnv         []string
	startTime        optional.Option[time.Time]

	// events coming in from plugins
	pluginEvents chan languageplugin.PluginEvent

	// requests to rebuild modules due to dependencies changing or plugins dying
	rebuildEvents chan rebuildEvent

	// internal channel for raw engine updates (does not include all state changes)
	rawEngineUpdates chan *buildenginepb.EngineEvent

	// topic to subscribe to engine events
	EngineUpdates *pubsub.Topic[*buildenginepb.EngineEvent]

	devModeEndpointUpdates chan dev.LocalEndpoint
	devMode                bool

	// deployment queue and state tracking
	deploymentQueue chan pendingDeploy
	os              string
	arch            string
}

type Option func(o *Engine)

func Parallelism(n int) Option {
	return func(o *Engine) {
		o.parallelism = n
	}
}

func BuildEnv(env []string) Option {
	return func(o *Engine) {
		o.buildEnv = env
	}
}

// WithDevMode sets the engine to dev mode.
func WithDevMode(updates chan dev.LocalEndpoint) Option {
	return func(o *Engine) {
		o.devModeEndpointUpdates = updates
		o.devMode = true
	}
}

// WithStartTime sets the start time to report total startup time
func WithStartTime(startTime time.Time) Option {
	return func(o *Engine) {
		o.startTime = optional.Some(startTime)
	}
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
//
// "dirs" are directories to scan for local modules.
func New(
	ctx context.Context,
	adminClient AdminClient,
	schemaSource *schemaeventsource.EventSource,
	projectConfig projectconfig.Config,
	moduleDirs []string,
	updatesEndpoint *url.URL,
	options ...Option,
) (*Engine, error) {
	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("build-engine"))
	ctx = rpc.ContextWithClient(ctx, adminClient)
	e := &Engine{
		adminClient:      adminClient,
		schemaSource:     schemaSource,
		projectConfig:    projectConfig,
		moduleDirs:       moduleDirs,
		moduleMetas:      xsync.NewMapOf[string, moduleMeta](),
		watcher:          watch.NewWatcher(optional.Some(projectConfig.WatchModulesLockPath()), "ftl.toml"),
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schema.Notification](),
		pluginEvents:     make(chan languageplugin.PluginEvent, 128),
		parallelism:      runtime.NumCPU(),
		modulesToBuild:   xsync.NewMapOf[string, bool](),
		rebuildEvents:    make(chan rebuildEvent, 128),
		rawEngineUpdates: make(chan *buildenginepb.EngineEvent, 128),
		EngineUpdates:    pubsub.New[*buildenginepb.EngineEvent](),
		deploymentQueue:  make(chan pendingDeploy, 128),
		arch:             runtime.GOARCH, // Default to the local env, we attempt to read these from the cluster later
		os:               runtime.GOOS,
	}
	for _, option := range options {
		option(e)
	}
	e.controllerSchema.Store("builtin", schema.Builtins())
	ctx, cancel := context.WithCancelCause(ctx)
	e.cancel = cancel

	err := CleanStubs(ctx, projectConfig.Root())
	if err != nil {
		return nil, fmt.Errorf("failed to clean stubs: %w", err)
	}

	updateTerminalWithEngineEvents(ctx, e.EngineUpdates)

	go e.watchForPluginEvents(ctx)
	go e.watchForEventsToPublish(ctx)
	go func() {
		if err := e.startUpdatesService(ctx, updatesEndpoint); err != nil && !errors.Is(err, context.Canceled) {
			log.FromContext(ctx).Errorf(err, "updates service failed")
		}
	}()

	// Start the deployment queue processor
	go e.processDeploymentQueue(ctx)

	configs, err := watch.DiscoverModules(ctx, moduleDirs)
	if err != nil {
		return nil, fmt.Errorf("could not find modules: %w", err)
	}

	wg := &errgroup.Group{}
	for _, config := range configs {
		wg.Go(func() error {
			meta, err := e.newModuleMeta(ctx, config)
			if err != nil {
				return err
			}
			meta, err = copyMetaWithUpdatedDependencies(ctx, meta)
			if err != nil {
				return err
			}
			e.moduleMetas.Store(config.Module, meta)
			e.modulesToBuild.Store(config.Module, true)
			e.rawEngineUpdates <- &buildenginepb.EngineEvent{
				Timestamp: timestamppb.Now(),
				Event: &buildenginepb.EngineEvent_ModuleAdded{
					ModuleAdded: &buildenginepb.ModuleAdded{
						Module: config.Module,
					},
				},
			}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, err //nolint:wrapcheck
	}
	if adminClient != nil {
		info, err := adminClient.ClusterInfo(ctx, connect.NewRequest(&ftlv1.ClusterInfoRequest{}))
		if err != nil {
			log.FromContext(ctx).Debugf("failed to get cluster info: %s", err)
		} else {
			e.os = info.Msg.Os
			e.arch = info.Msg.Arch
		}
	}
	if adminClient == nil {
		return e, nil
	}
	e.startSchemaSync(ctx)
	return e, nil
}

// Sync module schema changes from the FTL controller, as well as from manual
// updates, and merge them into a single schema map.
func (e *Engine) startSchemaSync(ctx context.Context) {
	logger := log.FromContext(ctx)
	if !e.schemaSource.Live() {
		logger.Debugf("Schema source is not live, skipping initial sync.")
	} else {
		e.schemaSource.WaitForInitialSync(ctx)
		for _, module := range e.schemaSource.CanonicalView().Modules {
			e.controllerSchema.Store(module.Name, module)
		}
	}

	go func() {
		events := e.schemaSource.Subscribe(ctx)
		for event := range channels.IterContext(ctx, events) {
			e.processEvent(event)
		}
	}()
}

func (e *Engine) processEvent(event schema.Notification) {
	switch event := event.(type) {
	case *schema.ChangesetCommittedNotification:
		adding := map[string]bool{}
		for _, a := range event.Changeset.Modules {
			adding[a.Name] = true
		}
		for _, removed := range event.Changeset.RemovingModules {
			// If a module has been explicitly killed we only find out about it here
			e.controllerSchema.Delete(removed.Name)
			if !adding[removed.Name] {
				e.rawEngineUpdates <- &buildenginepb.EngineEvent{
					Timestamp: timestamppb.Now(),
					Event: &buildenginepb.EngineEvent_ModuleRemoved{
						ModuleRemoved: &buildenginepb.ModuleRemoved{
							Module: removed.Name,
						},
					},
				}
			}
		}
		for _, module := range event.Changeset.Modules {
			e.controllerSchema.Store(module.Name, module)
		}
	default:

	}
	e.schemaChanges.Publish(event)
}

// Close stops the Engine's schema sync.
func (e *Engine) Close() error {
	e.cancel(fmt.Errorf("build engine stopped: %w", context.Canceled))
	return nil
}

// Graph returns the dependency graph for the given modules.
//
// If no modules are provided, the entire graph is returned. An error is returned if
// any dependencies are missing.
func (e *Engine) Graph(moduleNames ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(moduleNames) == 0 {
		e.moduleMetas.Range(func(name string, _ moduleMeta) bool {
			moduleNames = append(moduleNames, name)
			return true
		})
	}
	for _, name := range moduleNames {
		if err := e.buildGraph(name, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Engine) buildGraph(moduleName string, out map[string][]string) error {
	var deps []string
	// Short-circuit previously explored nodes
	if _, ok := out[moduleName]; ok {
		return nil
	}
	foundModule := false
	if meta, ok := e.moduleMetas.Load(moduleName); ok {
		foundModule = true
		deps = meta.module.Dependencies(AlwaysIncludeBuiltin)
	}
	if !foundModule {
		if sch, ok := e.controllerSchema.Load(moduleName); ok {
			foundModule = true
			deps = append(deps, sch.Imports()...)
		}
	}
	if !foundModule {
		return fmt.Errorf("module %q not found", moduleName)
	}
	deps = slices.Unique(deps)
	out[moduleName] = deps
	for _, dep := range deps {
		if err := e.buildGraph(dep, out); err != nil {
			return err
		}
	}
	return nil
}

// Import manually imports a schema for a module as if it were retrieved from
// the FTL controller.
func (e *Engine) Import(ctx context.Context, schema *schema.Module) {
	e.controllerSchema.Store(schema.Name, schema)
}

// Build attempts to build all local modules.
func (e *Engine) Build(ctx context.Context) error {
	return e.buildWithCallback(ctx, nil)
}

// Each iterates over all local modules.
func (e *Engine) Each(fn func(Module) error) (err error) {
	e.moduleMetas.Range(func(key string, value moduleMeta) bool {
		if ferr := fn(value.module); ferr != nil {
			err = fmt.Errorf("%s: %w", key, ferr)
			return false
		}
		return true
	})
	return
}

// Modules returns the names of all modules.
func (e *Engine) Modules() []string {
	var moduleNames []string
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		moduleNames = append(moduleNames, name)
		return true
	})
	return moduleNames
}

// Dev builds and deploys all local modules and watches for changes, redeploying as necessary.
func (e *Engine) Dev(ctx context.Context, period time.Duration) error {
	return e.watchForModuleChanges(ctx, period)
}

// watchForModuleChanges watches for changes and all build start and event state changes.
func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	schemaChanges := make(chan schema.Notification, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer func() {
		e.schemaChanges.Unsubscribe(schemaChanges)
	}()

	watchEvents := make(chan watch.WatchEvent, 128)
	ctx, cancel := context.WithCancelCause(ctx)
	topic, err := e.watcher.Watch(ctx, period, e.moduleDirs)
	if err != nil {
		cancel(fmt.Errorf("watch failed: %w: %w", context.Canceled, err))
		return err
	}
	topic.Subscribe(watchEvents)
	defer func() {
		// Cancel will close the topic and channel
		cancel(fmt.Errorf("watch stopped: %w", context.Canceled))
	}()

	// Build and deploy all modules first.
	err = e.BuildAndDeploy(ctx, 1, true)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Errorf(err, "Initial deploy failed")
	}

	moduleHashes := map[string][]byte{}
	e.controllerSchema.Range(func(name string, sch *schema.Module) bool {
		hash, err := computeModuleHash(sch)
		if err != nil {
			logger.Errorf(err, "compute hash for %s failed", name)
			return false
		}
		moduleHashes[name] = hash
		return true
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event := <-watchEvents:
			switch event := event.(type) {
			case watch.WatchEventModuleAdded:
				logger.Debugf("Module %q added", event.Config.Module)
				config := event.Config
				if _, exists := e.moduleMetas.Load(config.Module); !exists {
					meta, err := e.newModuleMeta(ctx, config)
					logger.Debugf("generated meta for %q", event.Config.Module)
					if err != nil {
						logger.Errorf(err, "could not add module %s", config.Module)
						continue
					}
					e.moduleMetas.Store(config.Module, meta)
					e.rawEngineUpdates <- &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_ModuleAdded{
							ModuleAdded: &buildenginepb.ModuleAdded{
								Module: config.Module,
							},
						},
					}
					logger.Debugf("calling build and deploy %q", event.Config.Module)
					_ = e.BuildAndDeploy(ctx, 1, true, config.Module) //nolint:errcheck
				}
			case watch.WatchEventModuleRemoved:
				err := terminateModuleDeployment(ctx, e.schemaSource, e.adminClient, event.Config.Module)
				if err != nil {
					logger.Errorf(err, "terminate %s failed", event.Config.Module)
				}
				if meta, ok := e.moduleMetas.Load(event.Config.Module); ok {
					meta.plugin.Updates().Unsubscribe(meta.events)
					err := meta.plugin.Kill()
					if err != nil {
						logger.Errorf(err, "terminate %s plugin failed", event.Config.Module)
					}
				}
				e.moduleMetas.Delete(event.Config.Module)
				e.rawEngineUpdates <- &buildenginepb.EngineEvent{
					Timestamp: timestamppb.Now(),
					Event: &buildenginepb.EngineEvent_ModuleRemoved{
						ModuleRemoved: &buildenginepb.ModuleRemoved{
							Module: event.Config.Module,
						},
					},
				}
			case watch.WatchEventModuleChanged:
				// ftl.toml file has changed
				meta, ok := e.moduleMetas.Load(event.Config.Module)
				if !ok {
					logger.Warnf("Module %q not found", event.Config.Module)
					continue
				}

				updatedConfig, err := moduleconfig.LoadConfig(event.Config.Dir)
				if err != nil {
					logger.Errorf(err, "Could not load updated toml for %s", event.Config.Module)
					continue
				}
				validConfig, err := updatedConfig.FillDefaultsAndValidate(meta.configDefaults)
				if err != nil {
					logger.Errorf(err, "Could not configure module config defaults for %s", event.Config.Module)
					continue
				}
				meta.module.Config = validConfig
				e.moduleMetas.Store(event.Config.Module, meta)

				_ = e.BuildAndDeploy(ctx, 1, true, event.Config.Module) //nolint:errcheck
			}
		case event := <-schemaChanges:
			switch event := event.(type) {
			case *schema.ChangesetCommittedNotification:
				inCs := map[string]bool{}
				for _, module := range event.Changeset.Modules {
					inCs[module.Name] = true
				}
				for _, module := range event.Changeset.Modules {
					existingHash, ok := moduleHashes[module.Name]
					if !ok {
						existingHash = []byte{}
					}

					hash, err := computeModuleHash(module)
					if err != nil {
						logger.Errorf(err, "compute hash for %s failed", module.Name)
						continue
					}

					if bytes.Equal(hash, existingHash) {
						logger.Tracef("schema for %s has not changed", module.Name)
						continue
					}

					moduleHashes[module.Name] = hash

					dependentModuleNames := e.getDependentModuleNames(module.Name)
					dependentModuleNames = slices.Filter(dependentModuleNames, func(name string) bool {
						// We don't update if this was already part of the same changeset
						return !inCs[name]
					})
					if len(dependentModuleNames) > 0 {
						logger.Infof("%s's schema changed; processing %s", module.Name, strings.Join(dependentModuleNames, ", "))
						_ = e.BuildAndDeploy(ctx, 1, true, dependentModuleNames...) //nolint:errcheck
					}
				}
			default:

			}

		case event := <-e.rebuildEvents:
			events := []rebuildEvent{event}
		readLoop:
			for {
				select {
				case event := <-e.rebuildEvents:
					events = append(events, event)
				default:
					break readLoop
				}
			}

			// Batch generate stubs for all auto rebuilds
			//
			// This is normally part of each group in the build topology, but auto rebuilds do not go through that flow
			builtModules := map[string]*schema.Module{}
			for _, event := range events {
				event, ok := event.(autoRebuildCompletedEvent)
				if !ok {
					continue
				}
				builtModules[event.module] = event.schema
			}
			if len(builtModules) > 0 {
				metasMap := map[string]moduleMeta{}
				e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
					metasMap[name] = meta
					return true
				})
				err = GenerateStubs(ctx, e.projectConfig.Root(), maps.Values(builtModules), metasMap)
				if err != nil {
					logger.Errorf(err, "Failed to generate stubs")
				}

				// Sync references to stubs if needed by the runtime
				err = e.syncNewStubReferences(ctx, builtModules, metasMap)
				if err != nil {
					logger.Errorf(err, "Failed to sync stub references")
				}

				// Deploy modules
				var modulesToDeploy = []Module{}
				for _, module := range builtModules {
					moduleToDeploy, ok := e.moduleMetas.Load(module.Name)
					if ok {
						modulesToDeploy = append(modulesToDeploy, moduleToDeploy.module)
					}
				}
				_ = e.deploy(ctx, modulesToDeploy, 1, false) //nolint:errcheck
			}

			// Batch together all new builds requested
			modulesToBuild := map[string]bool{}
			for _, event := range events {
				event, ok := event.(rebuildRequestEvent)
				if !ok {
					continue
				}
				modulesToBuild[event.module] = true
			}
			if len(modulesToBuild) > 0 {
				_ = e.BuildAndDeploy(ctx, 1, true, maps.Keys(modulesToBuild)...) //nolint:errcheck
			}
		}
	}
}

type moduleState int

const (
	moduleStateBuildWaiting moduleState = iota
	moduleStateExplicitlyBuilding
	moduleStateAutoRebuilding
	moduleStateBuilt
	moduleStateDeployWaiting
	moduleStateDeploying
	moduleStateDeployed
	moduleStateFailed
)

func isIdle(moduleStates map[string]moduleState) bool {
	if len(moduleStates) == 0 {
		return true
	}
	for _, state := range moduleStates {
		switch state {
		case moduleStateExplicitlyBuilding,
			moduleStateAutoRebuilding,
			moduleStateDeploying:
			return false

		case moduleStateFailed,
			moduleStateDeployed,
			moduleStateBuildWaiting, // Modules can stay in this state if dependant modules fail to build
			moduleStateDeployWaiting,
			moduleStateBuilt:
		}
	}
	return true
}

// watchForEventsToPublish listens for raw build events, collects state, and publishes public events to BuildUpdates topic.
func (e *Engine) watchForEventsToPublish(ctx context.Context) {
	moduleErrors := map[string]*langpb.ErrorList{}
	moduleStates := map[string]moduleState{}

	idle := true
	var endTime time.Time
	var becomeIdleTimer <-chan time.Time

	isFirstRound := true

	addTimestamp := func(evt *buildenginepb.EngineEvent) {
		if evt.Timestamp == nil {
			evt.Timestamp = timestamppb.Now()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-becomeIdleTimer:
			becomeIdleTimer = nil
			if !isIdle(moduleStates) {
				continue
			}
			idle = true

			if e.devMode && isFirstRound {
				logger := log.FromContext(ctx)
				if len(moduleErrors) > 0 {
					var errs []error
					for module, errList := range moduleErrors {
						if errList != nil && len(errList.Errors) > 0 {
							moduleErr := fmt.Errorf("%s: %s", module, langpb.ErrorListString(errList))
							errs = append(errs, moduleErr)
						}
					}
					logger.Errorf(errors.Join(errs...), "Initial build failed")
				} else if start, ok := e.startTime.Get(); ok {
					e.startTime = optional.None[time.Time]()
					logger.Infof("All modules deployed in %.2fs, watching for changes...", endTime.Sub(start).Seconds())
				} else {
					logger.Infof("All modules deployed, watching for changes...")
				}
			}
			isFirstRound = false

			modulesOutput := []*buildenginepb.EngineEnded_Module{}
			for module := range moduleStates {
				meta, ok := e.moduleMetas.Load(module)
				if !ok {
					continue
				}
				modulesOutput = append(modulesOutput, &buildenginepb.EngineEnded_Module{
					Module: module,
					Path:   meta.module.Config.Dir,
					Errors: moduleErrors[module],
				})
			}
			evt := &buildenginepb.EngineEvent{
				Timestamp: timestamppb.Now(),
				Event: &buildenginepb.EngineEvent_EngineEnded{
					EngineEnded: &buildenginepb.EngineEnded{
						Modules: modulesOutput,
					},
				},
			}
			addTimestamp(evt)
			e.EngineUpdates.Publish(evt)

		case evt := <-e.rawEngineUpdates:
			switch rawEvent := evt.Event.(type) {
			case *buildenginepb.EngineEvent_ModuleAdded:

			case *buildenginepb.EngineEvent_ModuleRemoved:
				delete(moduleErrors, rawEvent.ModuleRemoved.Module)
				delete(moduleStates, rawEvent.ModuleRemoved.Module)

			case *buildenginepb.EngineEvent_ModuleBuildWaiting:
				moduleStates[rawEvent.ModuleBuildWaiting.Config.Name] = moduleStateBuildWaiting

			case *buildenginepb.EngineEvent_ModuleBuildStarted:
				if idle {
					idle = false
					started := &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_EngineStarted{
							EngineStarted: &buildenginepb.EngineStarted{},
						},
					}
					addTimestamp(started)
					e.EngineUpdates.Publish(started)
				}
				if rawEvent.ModuleBuildStarted.IsAutoRebuild {
					moduleStates[rawEvent.ModuleBuildStarted.Config.Name] = moduleStateAutoRebuilding
				} else {
					moduleStates[rawEvent.ModuleBuildStarted.Config.Name] = moduleStateExplicitlyBuilding
				}
				delete(moduleErrors, rawEvent.ModuleBuildStarted.Config.Name)
				log.FromContext(ctx).Module(rawEvent.ModuleBuildStarted.Config.Name).Scope("build").Infof("Building module")
			case *buildenginepb.EngineEvent_ModuleBuildFailed:
				moduleStates[rawEvent.ModuleBuildFailed.Config.Name] = moduleStateFailed
				moduleErrors[rawEvent.ModuleBuildFailed.Config.Name] = rawEvent.ModuleBuildFailed.Errors
				moduleErr := fmt.Errorf("%s: %s", rawEvent.ModuleBuildFailed.Config.Name, langpb.ErrorListString(rawEvent.ModuleBuildFailed.Errors))
				log.FromContext(ctx).Module(rawEvent.ModuleBuildFailed.Config.Name).Scope("build").Errorf(moduleErr, "Build failed")
			case *buildenginepb.EngineEvent_ModuleBuildSuccess:
				moduleStates[rawEvent.ModuleBuildSuccess.Config.Name] = moduleStateBuilt
				delete(moduleErrors, rawEvent.ModuleBuildSuccess.Config.Name)
			case *buildenginepb.EngineEvent_ModuleDeployWaiting:
				moduleStates[rawEvent.ModuleDeployWaiting.Module] = moduleStateDeployWaiting
			case *buildenginepb.EngineEvent_ModuleDeployStarted:
				if idle {
					idle = false
					started := &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_EngineStarted{
							EngineStarted: &buildenginepb.EngineStarted{},
						},
					}
					addTimestamp(started)
					e.EngineUpdates.Publish(started)
				}
				moduleStates[rawEvent.ModuleDeployStarted.Module] = moduleStateDeploying
				delete(moduleErrors, rawEvent.ModuleDeployStarted.Module)
			case *buildenginepb.EngineEvent_ModuleDeployFailed:
				moduleStates[rawEvent.ModuleDeployFailed.Module] = moduleStateFailed
				moduleErrors[rawEvent.ModuleDeployFailed.Module] = rawEvent.ModuleDeployFailed.Errors
			case *buildenginepb.EngineEvent_ModuleDeploySuccess:
				moduleStates[rawEvent.ModuleDeploySuccess.Module] = moduleStateDeployed
				delete(moduleErrors, rawEvent.ModuleDeploySuccess.Module)
			}

			addTimestamp(evt)
			e.EngineUpdates.Publish(evt)
		}
		if !idle && isIdle(moduleStates) {
			endTime = time.Now()
			becomeIdleTimer = time.After(time.Second * 2)
		}
	}
}

func computeModuleHash(module *schema.Module) ([]byte, error) {
	hasher := sha256.New()
	data := []byte(module.String())
	if _, err := hasher.Write(data); err != nil {
		return nil, err // Handle errors that might occur during the write
	}

	return hasher.Sum(nil), nil
}

func (e *Engine) getDependentModuleNames(moduleName string) []string {
	dependentModuleNames := map[string]bool{}
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		for _, dep := range meta.module.Dependencies(AlwaysIncludeBuiltin) {
			if dep == moduleName {
				dependentModuleNames[name] = true
			}
		}
		return true
	})
	return maps.Keys(dependentModuleNames)
}

// BuildAndDeploy attempts to build and deploy all local modules.
func (e *Engine) BuildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, moduleNames ...string) (err error) {
	logger := log.FromContext(ctx)
	if len(moduleNames) == 0 {
		moduleNames = e.Modules()
	}
	if len(moduleNames) == 0 {
		return nil
	}

	defer func() {
		if err == nil {
			return
		}
		pendingInitialBuilds := []string{}
		e.modulesToBuild.Range(func(name string, value bool) bool {
			if value {
				pendingInitialBuilds = append(pendingInitialBuilds, name)
			}
			return true
		})

		// Print out all modules that have yet to build if there are any errors
		if len(pendingInitialBuilds) > 0 {
			logger.Infof("Modules waiting to build: %s", strings.Join(pendingInitialBuilds, ", "))
		}
	}()

	buildGroup := errgroup.Group{}

	modulesToDeploy := []Module{}
	buildGroup.Go(func() error {
		return e.buildWithCallback(ctx, func(buildCtx context.Context, module Module) error {
			e.modulesToBuild.Store(module.Config.Module, false)
			modulesToDeploy = append(modulesToDeploy, module)
			e.rawEngineUpdates <- &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
					ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
						Module: module.Config.Module,
					},
				},
			}
			return nil
		}, moduleNames...)
	})

	// Wait for all build attempts to complete
	buildErr := buildGroup.Wait()
	if buildErr != nil {
		return fmt.Errorf("build failed: %w", buildErr)
	}

	// Queue the modules for deployment instead of deploying directly
	return e.deploy(ctx, modulesToDeploy, replicas, waitForDeployOnline)
}

type buildCallback func(ctx context.Context, module Module) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, moduleNames ...string) error {
	logger := log.FromContext(ctx)
	if len(moduleNames) == 0 {
		e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
			moduleNames = append(moduleNames, name)
			return true
		})
	}

	mustBuildChan := make(chan moduleconfig.ModuleConfig, len(moduleNames))
	wg := errgroup.Group{}
	for _, name := range moduleNames {
		wg.Go(func() error {
			meta, ok := e.moduleMetas.Load(name)
			if !ok {
				return fmt.Errorf("module %q not found", name)
			}

			meta, err := copyMetaWithUpdatedDependencies(ctx, meta)
			if err != nil {
				return fmt.Errorf("could not get dependencies for %s: %w", name, err)
			}

			e.moduleMetas.Store(name, meta)
			mustBuildChan <- meta.module.Config
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return err //nolint:wrapcheck
	}
	close(mustBuildChan)
	mustBuild := map[string]bool{}
	for config := range mustBuildChan {
		mustBuild[config.Module] = true
		proto, err := langpb.ModuleConfigToProto(config.Abs())
		if err != nil {
			logger.Errorf(err, "failed to marshal module config")
			continue
		}
		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
			Timestamp: timestamppb.Now(),
			Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
				ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
					Config: proto,
				},
			},
		}
	}

	graph, err := e.Graph(moduleNames...)
	if err != nil {
		return err
	}
	builtModules := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	metasMap := map[string]moduleMeta{}
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		metasMap[name] = meta
		return true
	})
	err = GenerateStubs(ctx, e.projectConfig.Root(), maps.Values(builtModules), metasMap)
	if err != nil {
		return err
	}

	topology, err := TopologicalSort(graph)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1024)
	for _, group := range topology {
		knownSchemas := map[string]*schema.Module{}
		err := e.gatherSchemas(builtModules, knownSchemas)
		if err != nil {
			return err
		}

		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))

		wg := errgroup.Group{}
		wg.SetLimit(e.parallelism)

		logger.Debugf("Building group: %v", group)
		for _, moduleName := range group {
			wg.Go(func() error {
				logger := log.FromContext(ctx).Module(moduleName).Scope("build")
				ctx := log.ContextWithLogger(ctx, logger)
				err := e.tryBuild(ctx, mustBuild, moduleName, builtModules, schemas, callback)
				if err != nil {
					errCh <- err
				}
				return nil
			})
		}

		err = wg.Wait()
		if err != nil {
			return err
		}

		// Now this group is built, collect all the schemas.
		close(schemas)
		newSchemas := []*schema.Module{}
		for sch := range schemas {
			builtModules[sch.Name] = sch
			newSchemas = append(newSchemas, sch)
		}

		err = GenerateStubs(ctx, e.projectConfig.Root(), newSchemas, metasMap)
		if err != nil {
			return err
		}

		// Sync references to stubs if needed by the runtime
		err = e.syncNewStubReferences(ctx, builtModules, metasMap)
		if err != nil {
			return err
		}
	}

	close(errCh)
	allErrors := []error{}
	for err := range errCh {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}

func (e *Engine) tryBuild(ctx context.Context, mustBuild map[string]bool, moduleName string, builtModules map[string]*schema.Module, schemas chan *schema.Module, callback buildCallback) error {
	logger := log.FromContext(ctx)

	if !mustBuild[moduleName] {
		return e.mustSchema(ctx, moduleName, builtModules, schemas)
	}

	meta, ok := e.moduleMetas.Load(moduleName)
	if !ok {
		return fmt.Errorf("module %q not found", moduleName)
	}

	for _, dep := range meta.module.Dependencies(Raw) {
		if _, ok := builtModules[dep]; !ok {
			logger.Warnf("build skipped because dependency %q failed to build", dep)
			return nil
		}
	}

	configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
	if err != nil {
		return fmt.Errorf("failed to marshal module config: %w", err)
	}
	e.rawEngineUpdates <- &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
			ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
				Config:        configProto,
				IsAutoRebuild: false,
			},
		},
	}

	err = e.build(ctx, moduleName, builtModules, schemas)
	if err != nil {
		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
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
	} else {
		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
			Timestamp: timestamppb.Now(),
			Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
				ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
					Config:        configProto,
					IsAutoRebuild: false,
				},
			},
		}
	}
	if err == nil && callback != nil {
		// load latest meta as it may have been updated
		meta, ok = e.moduleMetas.Load(moduleName)
		if !ok {
			return fmt.Errorf("module %q not found", moduleName)
		}
		return callback(ctx, meta.module)
	}

	return err
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load(moduleName); ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, moduleName, builtModules, schemas)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	meta, ok := e.moduleMetas.Load(moduleName)
	if !ok {
		return fmt.Errorf("module %q not found", moduleName)
	}

	sch := &schema.Schema{Modules: maps.Values(builtModules)}

	moduleSchema, deploy, err := build(ctx, meta.plugin, e.projectConfig, languageplugin.BuildContext{
		Config:       meta.module.Config,
		Schema:       sch,
		Dependencies: meta.module.Dependencies(Raw),
		BuildEnv:     e.buildEnv,
		Os:           e.os,
		Arch:         e.arch,
	}, e.devMode, e.devModeEndpointUpdates)
	if err != nil {
		if errors.Is(err, errInvalidateDependencies) {
			// Do not start a build directly as we are already building out a graph of modules.
			// Instead we send to a chan so that it can be processed after.
			e.rebuildEvents <- rebuildRequestEvent{module: moduleName}
		}
		return err
	}
	// update files to deploy
	e.moduleMetas.Compute(moduleName, func(meta moduleMeta, exists bool) (out moduleMeta, shouldDelete bool) {
		if !exists {
			return moduleMeta{}, true
		}
		meta.module = meta.module.CopyWithDeploy(deploy)
		return meta, false
	})
	schemas <- moduleSchema
	return nil
}

// Construct a combined schema for a module and its transitive dependencies.
func (e *Engine) gatherSchemas(
	moduleSchemas map[string]*schema.Module,
	out map[string]*schema.Module,
) error {
	e.controllerSchema.Range(func(name string, sch *schema.Module) bool {
		out[name] = sch
		return true
	})

	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		if _, ok := moduleSchemas[name]; ok {
			out[name] = moduleSchemas[name]
		} else {
			// We don't want to use a remote schema if we have it locally
			delete(out, name)
		}
		return true
	})

	return nil
}

func (e *Engine) syncNewStubReferences(ctx context.Context, newModules map[string]*schema.Module, metasMap map[string]moduleMeta) error {
	fullSchema := &schema.Schema{Modules: maps.Values(newModules)}
	for _, module := range e.schemaSource.CanonicalView().Modules {
		if _, ok := newModules[module.Name]; !ok {
			fullSchema.Modules = append(fullSchema.Modules, module)
		}
	}
	sort.SliceStable(fullSchema.Modules, func(i, j int) bool { return fullSchema.Modules[i].Name < fullSchema.Modules[j].Name })

	return SyncStubReferences(ctx,
		e.projectConfig.Root(),
		slices.Map(fullSchema.Modules, func(m *schema.Module) string { return m.Name }),
		metasMap,
		fullSchema)
}

func (e *Engine) newModuleMeta(ctx context.Context, config moduleconfig.UnvalidatedModuleConfig) (moduleMeta, error) {
	plugin, err := languageplugin.New(ctx, config.Dir, config.Language, config.Module)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not create plugin for %s: %w", config.Module, err)
	}
	events := make(chan languageplugin.PluginEvent, 64)
	plugin.Updates().Subscribe(events)

	// pass on plugin events to the main event channel
	// make sure we do not pass on nil (chan closure) events
	go func() {
		for {
			select {
			case event := <-events:
				if event == nil {
					// chan closed
					return
				}
				e.pluginEvents <- event
			case <-ctx.Done():
				return
			}
		}
	}()

	// update config with defaults
	customDefaults, err := plugin.ModuleConfigDefaults(ctx, config.Dir)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not get defaults provider for %s: %w", config.Module, err)
	}
	validConfig, err := config.FillDefaultsAndValidate(customDefaults)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not apply defaults for %s: %w", config.Module, err)
	}
	return moduleMeta{
		module:         newModule(validConfig),
		plugin:         plugin,
		events:         events,
		configDefaults: customDefaults,
	}, nil
}

// watchForPluginEvents listens for build updates from language plugins and reports them to the listener.
// These happen when a plugin for a module detects a change and automatically rebuilds.
func (e *Engine) watchForPluginEvents(originalCtx context.Context) {
	for {
		select {
		case event := <-e.pluginEvents:
			switch event := event.(type) {
			case languageplugin.PluginBuildEvent, languageplugin.AutoRebuildStartedEvent, languageplugin.AutoRebuildEndedEvent:
				buildEvent := event.(languageplugin.PluginBuildEvent) //nolint:forcetypeassert
				logger := log.FromContext(originalCtx).Module(buildEvent.ModuleName()).Scope("build")
				ctx := log.ContextWithLogger(originalCtx, logger)
				meta, ok := e.moduleMetas.Load(buildEvent.ModuleName())
				if !ok {
					logger.Warnf("module not found for build update")
					continue
				}
				configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
				if err != nil {
					continue
				}
				switch event := buildEvent.(type) {
				case languageplugin.AutoRebuildStartedEvent:
					e.rawEngineUpdates <- &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
							ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
								Config:        configProto,
								IsAutoRebuild: true,
							},
						},
					}

				case languageplugin.AutoRebuildEndedEvent:
					moduleSch, deploy, err := handleBuildResult(ctx, e.projectConfig, meta.module.Config, event.Result, e.devModeEndpointUpdates)
					if err != nil {
						e.rawEngineUpdates <- &buildenginepb.EngineEvent{
							Timestamp: timestamppb.Now(),
							Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
								ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
									Config:        configProto,
									IsAutoRebuild: true,
									Errors: &langpb.ErrorList{
										Errors: errorToLangError(err),
									},
								},
							},
						}
						if errors.Is(err, errInvalidateDependencies) {
							// Do not block this goroutine by building a module here.
							// Instead we send to a chan so that it can be processed elsewhere.
							e.rebuildEvents <- rebuildRequestEvent{module: event.ModuleName()}
						}
						continue
					}
					meta.module.Deploy = deploy
					e.moduleMetas.Store(event.ModuleName(), meta)

					e.rawEngineUpdates <- &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
							ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
								Config:        configProto,
								IsAutoRebuild: true,
							},
						},
					}
					e.rebuildEvents <- autoRebuildCompletedEvent{module: event.ModuleName(), schema: moduleSch}
				}
			case languageplugin.PluginDiedEvent:
				e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
					if meta.plugin != event.Plugin {
						return true
					}
					logger := log.FromContext(originalCtx).Module(name)
					logger.Errorf(event.Error, "Plugin died, recreating")

					c, err := moduleconfig.LoadConfig(meta.module.Config.Dir)
					if err != nil {
						logger.Errorf(err, "Could not recreate plugin: could not load config")
						return false
					}
					newMeta, err := e.newModuleMeta(originalCtx, c)
					if err != nil {
						logger.Errorf(err, "Could not recreate plugin")
						return false
					}
					e.moduleMetas.Store(name, newMeta)
					e.rebuildEvents <- rebuildRequestEvent{module: name}
					return false
				})
			}
		case <-originalCtx.Done():
			// kill all plugins
			e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
				err := meta.plugin.Kill()
				if err != nil {
					log.FromContext(originalCtx).Errorf(err, "could not kill plugin")
				}
				return true
			})
			return
		}
	}
}

func (e *Engine) deploy(ctx context.Context, modules []Module, replicas int32, wait bool) error {
	for _, module := range modules {
		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
				ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
					Module: module.Config.Module,
				},
			},
		}
	}
	errChan := make(chan error, 1)
	e.deploymentQueue <- pendingDeploy{modules: modules, replicas: replicas, err: errChan}
	if wait {
		select {
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck
		case err := <-errChan:
			return err
		}
	}
	return nil
}

// processDeploymentQueue handles the deployment queue and groups pending deployments into changesets
func (e *Engine) processDeploymentQueue(ctx context.Context) {
	logger := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return

		case deployment := <-e.deploymentQueue:
			// Collect any additional modules that have been queued
			modules := []Module{}
			modules = append(modules, deployment.modules...)
			errChans := []chan error{deployment.err}
			more := true
			for more {
				select {
				case newModules := <-e.deploymentQueue:
					errChans = append(errChans, newModules.err)
					modules = append(modules, newModules.modules...)
				default:
					more = false
				}
			}

			logger.Debugf("Deploying %d modules", len(modules))

			// Deploy all collected modules
			for _, module := range modules {
				e.rawEngineUpdates <- &buildenginepb.EngineEvent{
					Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
						ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
							Module: module.Config.Module,
						},
					},
				}
			}
			err := Deploy(ctx, e.projectConfig, modules, deployment.replicas, true, e.adminClient)
			if err != nil {
				// Handle deployment failure
				for _, module := range modules {
					e.rawEngineUpdates <- &buildenginepb.EngineEvent{
						Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
							ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
								Module: module.Config.Module,
								Errors: &langpb.ErrorList{
									Errors: errorToLangError(err),
								},
							},
						},
					}
				}
			} else {
				// Handle deployment success
				for _, module := range modules {
					e.rawEngineUpdates <- &buildenginepb.EngineEvent{
						Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
							ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
								Module: module.Config.Module,
							},
						},
					}
				}
			}
			for _, errChan := range errChans {
				errChan <- err
			}
		}
	}
}
