package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/dev"
	imaps "github.com/block/ftl/internal/maps"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/realm"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/watch"
)

var _ rpc.Service = (*Engine)(nil)

// moduleMeta is a wrapper around a module that includes the last build's start time.
type moduleMeta struct {
	module         Module
	plugin         *languageplugin.LanguagePlugin
	configDefaults moduleconfig.CustomDefaults
}

// copyMetaWithUpdatedDependencies finds the dependencies for a module and returns a
// copy with those dependencies populated.
func copyMetaWithUpdatedDependencies(ctx context.Context, m moduleMeta) (moduleMeta, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting dependencies for %q", m.module.Config.Module)

	dependencies, err := m.plugin.GetDependencies(ctx, m.module.Config)
	if err != nil {
		return moduleMeta{}, errors.Wrapf(err, "could not get dependencies for %v", m.module.Config.Module)
	}

	m.module = m.module.CopyWithDependencies(dependencies)
	return m, nil
}

//sumtype:decl
type internalEvent interface {
	internalEvent()
}

type addMetasEvent struct {
	metas map[string]moduleMeta
}

func (addMetasEvent) internalEvent() {}

type removeMetaEvent struct {
	config moduleconfig.UnvalidatedModuleConfig
}

func (removeMetaEvent) internalEvent() {}

type schemaUpdateEvent struct {
	newSchema                   *schema.Schema
	modulesWithInterfaceChanges []schema.ModuleRefKey
	modulesWithBreakingChanges  []schema.ModuleRefKey
}

func (schemaUpdateEvent) internalEvent() {}

type moduleNeedsToBuildEvent struct {
	module string
}

func (moduleNeedsToBuildEvent) internalEvent() {}

type moduleBuildEndedEvent struct {
	config       moduleconfig.ModuleConfig
	moduleSchema *schema.Module
	tmpDeployDir string
	deployPaths  []string
	err          error
}

func (moduleBuildEndedEvent) internalEvent() {}

type moduleDeployStartedEvent struct {
	modules []string
}

func (moduleDeployStartedEvent) internalEvent() {}

type moduleDeployEndedEvent struct {
	modules []string
	err     error
}

func (moduleDeployEndedEvent) internalEvent() {}

// Engine for building a set of modules.
type Engine struct {
	adminClient AdminClient
	// deployCoordinator *DeployCoordinator
	// moduleMetas       *xsync.MapOf[string, moduleMeta]
	// externalRealms    *xsync.MapOf[string, *schema.Realm]
	projectConfig projectconfig.Config
	moduleDirs    []string
	// watcher           *watch.Watcher // only watches for module toml changes
	// targetSchema      atomic.Value[*schema.Schema]
	// cancel      context.CancelCauseFunc
	parallelism int
	// modulesToBuild    *xsync.MapOf[string, bool]
	buildEnv  []string
	startTime optional.Option[time.Time]

	internalEvents chan internalEvent

	// internal channel for raw engine updates (does not include all state changes)
	// rawEngineUpdates chan *buildenginepb.EngineEvent

	// topic to subscribe to engine events
	engineUpdates *pubsub.Topic[*buildenginepb.EngineEvent]

	devModeEndpointUpdates chan dev.LocalEndpoint
	devMode                bool

	os             string
	arch           string
	updatesService rpc.Service
}

func (e *Engine) StartServices(ctx context.Context) ([]rpc.Option, error) {
	services, err := e.updatesService.StartServices(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start updates service")
	}
	return services, nil
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
// "moduleDirs" are directories to scan for local modules.
func New(
	ctx context.Context,
	adminClient AdminClient,
	projectConfig projectconfig.Config,
	moduleDirs []string,
	logChanges bool,
	options ...Option,
) (*Engine, error) {
	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("build-engine"))

	e := &Engine{
		adminClient:    adminClient,
		projectConfig:  projectConfig,
		moduleDirs:     moduleDirs,
		parallelism:    runtime.NumCPU(),
		engineUpdates:  pubsub.New[*buildenginepb.EngineEvent](),
		arch:           runtime.GOARCH, // Default to the local env, we attempt to read these from the cluster later
		os:             runtime.GOOS,
		internalEvents: make(chan internalEvent, 64),
	}

	for _, option := range options {
		option(e)
	}

	updateTerminalWithEngineEvents(ctx, e.engineUpdates)

	e.updatesService = e.startUpdatesService(ctx)

	if adminClient != nil {
		info, err := adminClient.ClusterInfo(ctx, connect.NewRequest(&adminpb.ClusterInfoRequest{}))
		if err != nil {
			log.FromContext(ctx).Debugf("failed to get cluster info: %s", err)
		} else {
			e.os = info.Msg.Os
			e.arch = info.Msg.Arch
		}
	}
	return e, nil
}

// Dev builds and deploys all local modules and watches for changes, redeploying as necessary.
func (e *Engine) Dev(ctx context.Context, period time.Duration, schemaSource *schemaeventsource.EventSource) error {
	externalRealms := []*schema.Realm{}
	for name, cfg := range e.projectConfig.ExternalRealms {
		realm, err := realm.GetExternalRealm(ctx, e.projectConfig.ExternalRealmPath(), name, cfg)
		if err != nil {
			return errors.Wrapf(err, "failed to read external realm %s", name)
		}
		externalRealms = append(externalRealms, realm)
	}
	// TODO: logchanges param?
	deployCoordinator := NewDeployCoordinator(ctx, e.adminClient, schemaSource, e.internalEvents, false, e.projectConfig, externalRealms)
	// Save initial schema
	initialSchemaEvent := <-deployCoordinator.SchemaUpdates

	go watchForNewOrRemovedModules(ctx, e.projectConfig, e.moduleDirs, period, e.internalEvents)
	go watchSchemaUpdates(ctx, initialSchemaEvent.schema, deployCoordinator.SchemaUpdates, e.internalEvents)

	// watch for module additions and revovals
	err := errors.WithStack(e.processEvents(ctx, initialSchemaEvent.schema, moduleWatcherWithPeriod(period), buildModuleAndPublish, func(ctx context.Context, module *pendingModule) bool {
		go deployCoordinator.deploy(ctx, module, optional.None[int32]())
		return true
	}))
	fmt.Printf("process events returned: %v\n", err)
	time.Sleep(time.Second * 2)
	return err
}

func (e *Engine) Build(ctx context.Context, schemaSource *schemaeventsource.EventSource) error {
	sch := schemaSource.CanonicalView()
	schemaUpdates := make(chan SchemaUpdatedEvent, 32)

	configs, err := watch.DiscoverModules(ctx, e.moduleDirs)
	if err != nil {
		return errors.Wrap(err, "could not find modules")
	}
	metaMap := newModuleMetasForConfigs(ctx, configs, e.projectConfig)
	if len(metaMap) > 0 {
		e.internalEvents <- addMetasEvent{metas: metaMap}
	}

	go watchSchemaUpdates(ctx, reflect.DeepCopy(sch), schemaUpdates, e.internalEvents)

	// watch for module additions and revovals
	return errors.WithStack(e.processEvents(ctx, reflect.DeepCopy(sch), nil, buildModuleAndPublish, func(ctx context.Context, module *pendingModule) bool {
		realm := sch.FirstInternalRealm().MustGet()
		realm.Modules = slices.Filter(realm.Modules, func(m *schema.Module) bool {
			return module.moduleName() != m.Name
		})
		realm.Modules = append(realm.Modules, module.schema)
		schemaUpdates <- SchemaUpdatedEvent{
			schema: sch,
			updatedModules: []schema.ModuleRefKey{
				{Realm: realm.Name, Module: module.moduleName()},
			},
		}
		return false
	}))
}

func watchForNewOrRemovedModules(ctx context.Context, projectConfig projectconfig.Config, moduleDirs []string, period time.Duration, internalEvents chan internalEvent) {
	logger := log.FromContext(ctx)
	watcher := watch.NewWatcher(optional.Some(projectConfig.WatchModulesLockPath()))
	moduleListChanges := make(chan watch.WatchEvent, 16)
	moduleListTopic, err := watcher.Watch(ctx, period, moduleDirs)
	if err != nil {
		logger.Errorf(err, "failed to start watcher for module directories %v", moduleDirs)
		return
	}
	moduleListTopic.Subscribe(moduleListChanges)

	fmt.Printf("waiting to receive module list changes...\n")
	for event := range channels.IterContext(ctx, moduleListChanges) {
		fmt.Printf("received module list change event: %T\n", event)
		switch event := event.(type) {
		case watch.WatchEventModulesAdded:
			metaMap := newModuleMetasForConfigs(ctx, event.Configs, projectConfig)
			if len(metaMap) > 0 {
				internalEvents <- addMetasEvent{metas: metaMap}
			}
		case watch.WatchEventModuleRemoved:
			internalEvents <- removeMetaEvent{config: event.Config}

		case watch.WatchEventModuleChanged:
			// Changes within a module are not handled here
		}
	}
}

func newModuleMetasForConfigs(ctx context.Context, configs []moduleconfig.UnvalidatedModuleConfig, projectConfig projectconfig.Config) map[string]moduleMeta {
	logger := log.FromContext(ctx)
	newMetas := make(chan moduleMeta, len(configs))
	group := errgroup.Group{}

	for _, config := range configs {
		group.Go(func() error {
			plugin, err := languageplugin.New(ctx, config.Dir, config.Language, config.Module)
			if err != nil {
				return errors.Wrapf(err, "could not create plugin for %s", config.Module)
			}
			// update config with defaults
			customDefaults, err := languageplugin.GetModuleConfigDefaults(ctx, config.Language, config.Dir)
			if err != nil {
				return errors.Wrapf(err, "could not get defaults provider for %s", config.Module)
			}
			validConfig, err := config.FillDefaultsAndValidate(customDefaults, projectConfig)
			if err != nil {
				return errors.Wrapf(err, "could not apply defaults for %s", config.Module)
			}
			meta := moduleMeta{
				module:         newModule(validConfig),
				plugin:         plugin,
				configDefaults: customDefaults,
			}
			meta, err = copyMetaWithUpdatedDependencies(ctx, meta)
			if err != nil {
				return errors.Wrapf(err, "could not copy meta with updated dependencies for %s", config.Module)
			}
			newMetas <- meta
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		logger.Errorf(err, "failed to create module metas for new modules")
	}
	metaMap := map[string]moduleMeta{}
collectMetas:
	for {
		select {
		case m := <-newMetas:
			metaMap[m.module.Config.Module] = m
		default:
			break collectMetas
		}
	}
	return metaMap
}

type moduleWatcherFunc func(ctx context.Context, config moduleconfig.ModuleConfig, internalEvents chan internalEvent) (transactionProviderFunc, context.CancelCauseFunc, error)

func moduleWatcherWithPeriod(period time.Duration) moduleWatcherFunc {
	return func(ctx context.Context, config moduleconfig.ModuleConfig, internalEvents chan internalEvent) (transactionProviderFunc, context.CancelCauseFunc, error) {
		patterns := config.Watch
		patterns = append(patterns, "ftl.toml", "**/*.sql")
		watcher := watch.NewWatcher(optional.None[string](), patterns...)

		ctx, cancel := context.WithCancelCause(ctx)
		updates, err := watcher.Watch(ctx, period, []string{config.Abs().Dir})
		if err != nil {
			cancel(context.Canceled)
			return nil, nil, errors.Wrapf(err, "failed to watch module directory %s", config.Dir)
		}
		events := make(chan watch.WatchEvent, 16)
		updates.Subscribe(events)

		go func() {
			for event := range channels.IterContext(ctx, events) {
				switch event.(type) {
				case watch.WatchEventModulesAdded:
					// not handled here
				case watch.WatchEventModuleRemoved:
					// not handled here
				case watch.WatchEventModuleChanged:
					internalEvents <- moduleNeedsToBuildEvent{module: config.Module}
				}
			}
		}()
		return func() watch.ModifyFilesTransaction {
			return watcher.GetTransaction(config.Abs().Dir)
		}, cancel, nil
	}
}

func watchSchemaUpdates(ctx context.Context,
	initialSchema *schema.Schema,
	schemaUpdates chan SchemaUpdatedEvent,
	internalEvents chan internalEvent) {
	logger := log.FromContext(ctx)
	moduleHashes := map[schema.ModuleRefKey][]byte{}
	// TODO: do not just do internal modules
	for _, realm := range initialSchema.Realms {
		for _, module := range realm.Modules {
			hash, err := computeModuleHash(module)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", module.Name)
				continue
			}
			moduleHashes[schema.ModuleRefKey{Realm: realm.Name, Module: module.Name}] = hash
		}
	}

	for event := range channels.IterContext(ctx, schemaUpdates) {
		modulesWithInterfaceChanges := []schema.ModuleRefKey{}
		materiallyChangedModules := []schema.ModuleRefKey{}
		for _, moduleRef := range event.updatedModules {
			moduleSch, ok := event.schema.Module(moduleRef.Realm, moduleRef.Module).Get()
			if !ok {
				logger.Logf(log.Error, "module %s not found in schema", moduleRef)
				continue
			}
			hash, err := computeModuleHash(moduleSch)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", moduleRef)
				continue
			}
			existingHash, ok := moduleHashes[moduleRef]
			if !ok {
				existingHash = []byte{}
			}
			modulesWithInterfaceChanges = append(modulesWithInterfaceChanges, moduleRef)
			if bytes.Equal(hash, existingHash) {
				logger.Tracef("schema for %s has not changed", moduleRef)
				continue
			}

			moduleHashes[moduleRef] = hash
			materiallyChangedModules = append(materiallyChangedModules, moduleRef)
		}
		internalEvents <- schemaUpdateEvent{
			newSchema:                   reflect.DeepCopy(event.schema),
			modulesWithInterfaceChanges: modulesWithInterfaceChanges,
			modulesWithBreakingChanges:  materiallyChangedModules,
		}
	}
}

func computeModuleHash(module *schema.Module) ([]byte, error) {
	hasher := sha256.New()
	data := []byte(module.String())
	if _, err := hasher.Write(data); err != nil {
		return nil, errors.WithStack(err) // Handle errors that might occur during the write
	}
	return hasher.Sum(nil), nil
}

type moduleState struct {
	meta                moduleMeta
	needsToBuild        bool
	lastEvent           *buildenginepb.EngineEvent
	cancelModuleWatch   context.CancelCauseFunc
	transactionProvider optional.Option[transactionProviderFunc]
}

func (e *Engine) processEvents(ctx context.Context, initialSchema *schema.Schema, moduleWatcher moduleWatcherFunc, builder buildFunc, deployer deployFunc) error {
	logger := log.FromContext(ctx)
	sch := initialSchema
	moduleStates := map[string]*moduleState{}
	metas := func() map[string]moduleMeta {
		return imaps.MapValues(moduleStates, func(_ string, m *moduleState) moduleMeta { return m.meta })
	}

	idle := true
	// var endTime time.Time
	var becomeIdleTimer <-chan time.Time

	for {
		events := []internalEvent{}
		select {
		case <-becomeIdleTimer:
			becomeIdleTimer = nil
			if !isIdle(moduleStates) {
				continue
			}
			idle = true

			// TODO: pass in module errors
			e.engineUpdates.Publish(newEngineEndedEvent(moduleStates))
			continue
		case event := <-e.internalEvents:
			events = append(events, event)
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context cancelled while waiting for events")
		}

	drainEvents:
		for {
			select {
			case event := <-e.internalEvents:
				events = append(events, event)
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "context cancelled while waiting for events")
			default:
				break drainEvents
			}
		}

		//
		// Process each event and update internal state
		//

		for _, event := range events {
			switch event := event.(type) {
			case addMetasEvent:
				for _, meta := range event.metas {
					name := meta.module.Config.Module
					newLanguage := len(slices.Filter(maps.Values(moduleStates), func(m *moduleState) bool {
						return m.meta.module.Config.Language == meta.module.Config.Language
					})) == 0
					if newLanguage {
						// clean stubs for the language if no modules are present
						// TODO: does this clean more than language specific stuff?
						CleanStubs(ctx, e.projectConfig.Root(), meta.module.Config.Language)

					}
					extEvent, err := newModuleBuildWaitingEvent(meta.module.Config)
					if err != nil {
						logger.Errorf(err, "failed to watch module %s", name)
						continue
					}

					var cancelModuleWatch context.CancelCauseFunc
					var transactionProvider optional.Option[transactionProviderFunc]
					if moduleWatcher != nil {
						txProvider, cancel, err := moduleWatcher(ctx, meta.module.Config, e.internalEvents)
						if err != nil {
							logger.Errorf(err, "failed to watch module %s", name)
							continue
						}
						cancelModuleWatch = cancel
						transactionProvider = optional.Some(txProvider)
					}
					moduleStates[name] = &moduleState{
						meta:                meta,
						needsToBuild:        true,
						lastEvent:           extEvent,
						cancelModuleWatch:   cancelModuleWatch,
						transactionProvider: transactionProvider,
					}
					e.engineUpdates.Publish(extEvent)

					if newLanguage {
						// TODO: not a good place for this
						// TODO: not just internal modules
						err := GenerateStubs(ctx, e.projectConfig.Root(), sch.InternalModules(), metas())
						if err != nil {
							// TODO: do not return here
							return errors.WithStack(err)
						}
					}
				}

				// New modules need to know which stubs have already been generated
				// TODO: not just internal modules
				SyncStubReferences(ctx, e.projectConfig.Root(), slices.Map(sch.InternalModules(), func(m *schema.Module) string { return m.Name }), event.metas, sch)

			case removeMetaEvent:
				if state, ok := moduleStates[event.config.Module]; ok && state.cancelModuleWatch != nil {
					state.cancelModuleWatch(errors.Wrap(context.Canceled, "module removed"))
				}
				delete(moduleStates, event.config.Module)
				e.engineUpdates.Publish(newModuleRemovedEvent(event.config.Module))

			case schemaUpdateEvent:
				sch = event.newSchema
				deps, err := GraphFromMetas(metas(), sch, slices.Map(event.modulesWithBreakingChanges, func(moduleRef schema.ModuleRefKey) string { return moduleRef.Module })...)
				if err != nil {
					logger.Errorf(err, "failed to get dependencies")
					continue
				}
				for _, moduleRef := range event.modulesWithBreakingChanges {
					depsForModule, ok := deps[moduleRef.Module]
					if ok {
						for _, dep := range depsForModule {
							if dep == "builtin" {
								continue
							}
							// mark all transitive dependencies as dirty
							if state, ok := moduleStates[dep]; ok {
								state.needsToBuild = true
							}
						}
					}
				}
				// TODO: not just internal modules
				if err := GenerateStubs(ctx, e.projectConfig.Root(), slices.Map(event.modulesWithInterfaceChanges, func(moduleRef schema.ModuleRefKey) *schema.Module {
					// TODO: remove MustGet() usage
					return sch.Module(moduleRef.Realm, moduleRef.Module).MustGet()
				}), metas()); err != nil {
					logger.Errorf(err, "failed to generate stubs for updated modules")
				}
				// All modules need to know which stubs have been generated
				SyncStubReferences(ctx, e.projectConfig.Root(), slices.Map(sch.InternalModules(), func(m *schema.Module) string { return m.Name }), metas(), sch)
			case moduleNeedsToBuildEvent:
				if state, ok := moduleStates[event.module]; ok {
					state.needsToBuild = true
				}

			case moduleBuildEndedEvent:
				state, ok := moduleStates[event.config.Module]
				if !ok {
					logger.Logf(log.Error, "module %s not found in module states", event.config.Module)
					continue
				}
				if event.err != nil {
					extEvent, err := newModuleBuildFailedEvent(event.config, event.err)
					if err != nil {
						logger.Errorf(err, "failed to create build failed event for module %s", event.config.Module)
						continue
					}
					state.lastEvent = extEvent
					e.engineUpdates.Publish(extEvent)
					continue
				}
				extEvent, err := newModuleBuildSuccessEvent(event.config)
				if err != nil {
					logger.Errorf(err, "failed to create build failed event for module %s", event.config.Module)
					continue
				}
				state.lastEvent = extEvent
				e.engineUpdates.Publish(extEvent)

				if deployer(ctx, newPendingModule(state.meta.module, event.tmpDeployDir, event.deployPaths, event.moduleSchema)) {
					extEvent := newModuleDeployWaitingEvent(event.config.Module)
					state.lastEvent = extEvent
					e.engineUpdates.Publish(extEvent)
				}

			case moduleDeployStartedEvent:
				for _, module := range event.modules {
					state, ok := moduleStates[module]
					if !ok {
						logger.Logf(log.Error, "module %s not found in module states", module)
						continue
					}
					extEvent := newModuleDeployStartedEvent(module)
					state.lastEvent = extEvent
					e.engineUpdates.Publish(extEvent)
				}
			case moduleDeployEndedEvent:
				for _, module := range event.modules {
					state, ok := moduleStates[module]
					if !ok {
						logger.Logf(log.Error, "module %s not found in module states", module)
						continue
					}
					var extEvent *buildenginepb.EngineEvent
					if event.err != nil {
						extEvent = newModuleDeployFailedEvent(module, event.err)
					} else {
						extEvent = newModuleDeploySuccessEvent(module)
					}
					state.lastEvent = extEvent
					e.engineUpdates.Publish(extEvent)
				}
			}
		}

		//
		// Kick off any builds that we can
		//
		modulesToBuild := slices.Filter(maps.Values(moduleStates), func(state *moduleState) bool {
			if !state.needsToBuild {
				return false
			}
			name := state.meta.module.Config.Module
			// TODO: calc deps once before?
			deps, err := GraphFromMetas(metas(), sch, name)
			if err != nil {
				// TODO: handle error
				return false
			}
			switch state.lastEvent.Event.(type) {
			case *buildenginepb.EngineEvent_ModuleBuildStarted,
				*buildenginepb.EngineEvent_ModuleDeployWaiting,
				*buildenginepb.EngineEvent_ModuleDeployStarted:
				return false
			default:
			}
			for _, dep := range deps[name] {
				if depState, ok := moduleStates[dep]; ok {
					if depState.needsToBuild {
						return false
					}
					if depState.lastEvent.Event != nil && depState.lastEvent.GetModuleBuildStarted() != nil {
						return false
					}
				}
				if _, ok := sch.Module(sch.FirstInternalRealm().MustGet().Name, dep).Get(); !ok {
					return false
				}
			}
			return true
		})
		buildCount := len(slices.Filter(maps.Values(moduleStates), func(state *moduleState) bool {
			return state.lastEvent.Event != nil && state.lastEvent.GetModuleBuildStarted() != nil
		}))
		for _, state := range modulesToBuild {
			if buildCount >= e.parallelism {
				break
			}
			engineEvent, err := newModuleBuildStartedEvent(state.meta.module.Config)
			if err != nil {
				logger.Errorf(err, "failed to create build started event for module %s", state.meta.module.Config.Module)
				continue
			}
			transactionProvider, ok := state.transactionProvider.Get()
			var fileTransaction watch.ModifyFilesTransaction
			if ok {
				fileTransaction = transactionProvider()
			} else {
				fileTransaction = watch.NoOpFilesTransation{}
			}
			state.needsToBuild = false
			state.lastEvent = engineEvent
			e.engineUpdates.Publish(engineEvent)

			go builder(ctx, e.projectConfig, state.meta.module, state.meta.plugin, languageplugin.BuildContext{
				Config:       state.meta.module.Config,
				Schema:       reflect.DeepCopy(sch),
				Dependencies: state.meta.module.Dependencies(Raw),
				BuildEnv:     e.buildEnv,
				Os:           e.os,
				Arch:         e.arch,
			}, e.devMode, e.devModeEndpointUpdates, fileTransaction, e.internalEvents)
			buildCount++
		}

		if !idle && isIdle(moduleStates) {
			// endTime = time.Now()
			becomeIdleTimer = time.After(time.Millisecond * 200)
		}
	}
}

func newModuleRemovedEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleRemoved{
			ModuleRemoved: &buildenginepb.ModuleRemoved{
				Module: module,
			},
		},
	}
}

func newModuleBuildWaitingEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
			ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
				Config: proto,
			},
		},
	}, nil
}

func newModuleBuildStartedEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
		ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
			Config: proto,
		},
	},
	}, nil
}

func newModuleBuildFailedEvent(config moduleconfig.ModuleConfig, buildErr error) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
			ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
				Config: proto,
				Errors: &langpb.ErrorList{Errors: errorToLangError(buildErr)},
			},
		},
	}, nil
}

func newModuleBuildSuccessEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
			ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
				Config: proto,
			},
		},
	}, nil
}

func newModuleDeployWaitingEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
			ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
				Module: module,
			},
		},
	}
}

func newModuleDeployStartedEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
			ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
				Module: module,
			},
		},
	}
}

func newModuleDeployFailedEvent(module string, deployErr error) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
			ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
				Module: module,
				Errors: &langpb.ErrorList{Errors: errorToLangError(deployErr)},
			},
		},
	}
}

func newModuleDeploySuccessEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
			ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
				Module: module,
			},
		},
	}
}

func newEngineEndedEvent(moduleStates map[string]*moduleState) *buildenginepb.EngineEvent {
	modulesOutput := []*buildenginepb.EngineEnded_Module{}
	for name, state := range moduleStates {
		modulesOutput = append(modulesOutput, &buildenginepb.EngineEnded_Module{
			Module: name,
			Path:   state.meta.module.Config.Dir,
			Errors: state.lastEvent.GetModuleBuildFailed().GetErrors(),
		})
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_EngineEnded{
			EngineEnded: &buildenginepb.EngineEnded{
				Modules: modulesOutput,
			},
		},
	}
}

func isIdle(moduleStates map[string]*moduleState) bool {
	if len(moduleStates) == 0 {
		return true
	}
	for _, state := range moduleStates {
		switch state.lastEvent.Event.(type) {
		case *buildenginepb.EngineEvent_ModuleBuildStarted,
			*buildenginepb.EngineEvent_ModuleDeployStarted:
			return false
		default:
		}
	}
	return true
}
