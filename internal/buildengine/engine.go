package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"runtime"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/realm"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/watch"
)

// moduleWatcherFunc is a function that watches a module for changes
//
// It returns a way to create file transactions and a cancel function to stop watching the module
type moduleWatcherFunc func(ctx context.Context, config moduleconfig.ModuleConfig, internalEvents chan internalEvent) (transactionProviderFunc, context.CancelCauseFunc, error)

// deployFunc is a function that might deploy a module.
//
// It returns true if the module is queued for deployment, or false otherwise.
type deployFunc func(ctx context.Context, module *pendingModule) (willDeploy bool)

//sumtype:decl
type internalEvent interface {
	internalEvent()
}

type addMetasEvent struct {
	preparedModules map[string]preparedModule
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
	adminClient   AdminClient
	projectConfig projectconfig.Config
	moduleDirs    []string
	parallelism   int
	buildEnv      []string
	startTime     optional.Option[time.Time]

	internalEvents chan internalEvent
	// topic to subscribe to engine events
	engineUpdates *pubsub.Topic[*buildenginepb.EngineEvent]

	devModeEndpointUpdates chan dev.LocalEndpoint
	devMode                bool

	os             string
	arch           string
	updatesService rpc.Service
}

var _ rpc.Service = (*Engine)(nil)

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
	_, err := e.processEvents(ctx, initialSchemaEvent.schema, false, moduleWatcherWithPeriod(period), buildModuleAndPublish, func(ctx context.Context, module *pendingModule) bool {
		go deployCoordinator.deploy(ctx, map[string]*pendingModule{module.schema.Name: module}, optional.None[int32]())
		return true
	})
	return errors.WithStack(err)
}

func (e *Engine) Build(ctx context.Context, schemaSource *schemaeventsource.EventSource) error {
	_, _, err := e.buildAndCollect(ctx, schemaSource)
	return err
}

func (e *Engine) BuildAndDeploy(ctx context.Context, schemaSource *schemaeventsource.EventSource, replicas optional.Option[int32]) error {
	moduleStates, pendingModules, err := e.buildAndCollect(ctx, schemaSource)
	// TODO: need a way to get module states...
	if err != nil {
		return err
	}
	for _, module := range pendingModules {
		e.engineUpdates.Publish(newModuleDeployStartedEvent(module.schema.Name))
	}
	deployCoordinator := NewDeployCoordinator(ctx, e.adminClient, schemaSource, e.internalEvents, true, e.projectConfig, nil)
	if err := deployCoordinator.deploy(ctx, pendingModules, replicas); err != nil {
		// TODO: handle
	}
	for _, module := range pendingModules {
		e.engineUpdates.Publish(newModuleDeploySuccessEvent(module.schema.Name))
	}
	// TODO: pass in module states
	e.engineUpdates.Publish(newEngineEndedEvent(moduleStates))
	return nil
}

// buildAndCollect builds all modules and returns the module states and pending modules for deployment.
func (e *Engine) buildAndCollect(ctx context.Context, schemaSource *schemaeventsource.EventSource) (map[string]*moduleState, map[string]*pendingModule, error) {
	sch := schemaSource.CanonicalView()
	if _, ok := sch.FirstInternalRealm().Get(); !ok {
		sch.Realms = append(sch.Realms, &schema.Realm{
			Name: e.projectConfig.Name,
		})
	}
	sch = sch.WithBuiltins()

	schemaUpdates := make(chan SchemaUpdatedEvent, 32)

	configs, err := watch.DiscoverModules(ctx, e.moduleDirs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not find modules")
	}
	preparedModuleMap := newPreparedModulesForConfigs(ctx, configs, e.projectConfig)
	if len(preparedModuleMap) > 0 {
		e.internalEvents <- addMetasEvent{preparedModules: preparedModuleMap}
	}

	go watchSchemaUpdates(ctx, reflect.DeepCopy(sch), schemaUpdates, e.internalEvents)

	pendingModules := map[string]*pendingModule{}
	// watch for module additions and revovals
	moduleStates, err := e.processEvents(ctx, reflect.DeepCopy(sch), true, nil, buildModuleAndPublish, func(ctx context.Context, module *pendingModule) bool {
		realm := sch.FirstInternalRealm().MustGet()
		realm.Modules = slices.Filter(realm.Modules, func(m *schema.Module) bool {
			return module.moduleName() != m.Name
		})
		realm.Modules = append(realm.Modules, module.schema)
		schemaUpdates <- SchemaUpdatedEvent{
			schema: reflect.DeepCopy(sch),
			updatedModules: []schema.ModuleRefKey{
				{Realm: realm.Name, Module: module.moduleName()},
			},
		}
		pendingModules[module.moduleName()] = module
		return false
	})
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return moduleStates, pendingModules, nil
}

// watchForNewOrRemovedModules watches for new or removed modules in the project directory.
//
// It uses a watcher to monitor the module directories and sends events to the internalEvents channel when modules are added or removed.
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

	for event := range channels.IterContext(ctx, moduleListChanges) {
		switch event := event.(type) {
		case watch.WatchEventModulesAdded:
			preparedModuleMap := newPreparedModulesForConfigs(ctx, event.Configs, projectConfig)
			if len(preparedModuleMap) > 0 {
				internalEvents <- addMetasEvent{preparedModules: preparedModuleMap}
			}
		case watch.WatchEventModuleRemoved:
			internalEvents <- removeMetaEvent{config: event.Config}

		case watch.WatchEventModuleChanged:
			// Changes within a module are not handled here
		}
	}
}

type preparedModule struct {
	module         Module
	plugin         *languageplugin.LanguagePlugin
	configDefaults moduleconfig.CustomDefaults
}

func newPreparedModulesForConfigs(ctx context.Context, configs []moduleconfig.UnvalidatedModuleConfig, projectConfig projectconfig.Config) map[string]preparedModule {
	logger := log.FromContext(ctx)
	preparedModules := make(chan preparedModule, len(configs))
	group := errgroup.Group{}

	for _, config := range configs {
		// Creating a plugin takes a while, so we do this in parallel.
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
			prep := preparedModule{
				module:         newModule(validConfig),
				plugin:         plugin,
				configDefaults: customDefaults,
			}
			logger.Debugf("Extracting dependencies for %q", prep.module.Config.Module)

			dependencies, err := prep.plugin.GetDependencies(ctx, prep.module.Config)
			if err != nil {
				return errors.Wrapf(err, "could not get dependencies for %v", prep.module.Config.Module)
			}
			prep.module = prep.module.CopyWithDependencies(dependencies)
			preparedModules <- prep
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		logger.Errorf(err, "failed to create module metas for new modules")
	}
	preparedModuleMap := map[string]preparedModule{}
collectMetas:
	for {
		select {
		case m := <-preparedModules:
			preparedModuleMap[m.module.Config.Module] = m
		default:
			break collectMetas
		}
	}
	return preparedModuleMap
}

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

// TODO: combine with moduleMeta? Make sure moduleState does not escape...
type moduleState struct {
	module         Module
	plugin         *languageplugin.LanguagePlugin
	configDefaults moduleconfig.CustomDefaults

	needsToBuild        bool
	lastEvent           *buildenginepb.EngineEvent
	cancelModuleWatch   context.CancelCauseFunc
	transactionProvider optional.Option[transactionProviderFunc]
}

func (e *Engine) processEvents(ctx context.Context, initialSchema *schema.Schema, endWhenIdle bool, moduleWatcher moduleWatcherFunc, builder buildFunc, deployer deployFunc) (map[string]*moduleState, error) {
	logger := log.FromContext(ctx)
	sch := initialSchema
	moduleStates := map[string]*moduleState{}

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

			if endWhenIdle {
				return moduleStates, nil
			}
			continue
		case event := <-e.internalEvents:
			events = append(events, event)
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "context cancelled while waiting for events")
		}

	drainEvents:
		for {
			select {
			case event := <-e.internalEvents:
				events = append(events, event)
			case <-ctx.Done():
				return nil, errors.Wrap(ctx.Err(), "context cancelled while waiting for events")
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
				if err := e.handleAddMetasEvent(ctx, event, sch, moduleStates, moduleWatcher); err != nil {
					logger.Errorf(err, "failed to handle add metas event")
					continue
				}
			case removeMetaEvent:
				// TODO: this needs to go through deploy coordinator
				if state, ok := moduleStates[event.config.Module]; ok && state.cancelModuleWatch != nil {
					state.cancelModuleWatch(errors.Wrap(context.Canceled, "module removed"))
				}
				delete(moduleStates, event.config.Module)
				e.engineUpdates.Publish(newModuleRemovedEvent(event.config.Module))

			case schemaUpdateEvent:
				sch = event.newSchema
				if err := e.handleSchemaUpdateEvent(ctx, event, sch, moduleStates); err != nil {
					logger.Errorf(err, "failed to handle schema update event")
					continue
				}
			case moduleNeedsToBuildEvent:
				if state, ok := moduleStates[event.module]; ok {
					state.needsToBuild = true
				}

			case moduleBuildEndedEvent:
				if err := e.handleBuildEndedEvent(ctx, event, moduleStates, deployer); err != nil {
					logger.Errorf(err, "failed to handle build ended event")
					continue
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
		modulesReadyToBuild, deps, err := modulesReadyToBuild(ctx, sch, moduleStates)
		if err != nil {
			logger.Errorf(err, "failed to get modules ready to build")
			continue
		}
		if idle && len(modulesReadyToBuild) > 0 {
			e.engineUpdates.Publish(newEngineStartedEvent())
			idle = false
		}
		if err := e.handleAnyModulesReadyToBuild(ctx, sch, moduleStates, modulesReadyToBuild, deps, builder); err != nil {
			logger.Errorf(err, "failed to handle any modules ready to build")
		}

		if !idle && isIdle(moduleStates) {
			// endTime = time.Now()
			becomeIdleTimer = time.After(time.Millisecond * 200)
		}
	}
}

func (e *Engine) handleAddMetasEvent(ctx context.Context, event addMetasEvent, sch *schema.Schema, moduleStates map[string]*moduleState, moduleWatcher moduleWatcherFunc) error {
	newLanguages := map[string]bool{}
	for _, preparedModule := range event.preparedModules {
		newLanguage := !newLanguages[preparedModule.module.Config.Language] && len(slices.Filter(maps.Values(moduleStates), func(m *moduleState) bool {
			return m.module.Config.Language == preparedModule.module.Config.Language
		})) == 0
		if newLanguage {
			newLanguages[preparedModule.module.Config.Language] = true
		}
	}
	if len(newLanguages) > 0 {
		// clean stubs for the language if no modules are present
		// TODO: does this clean more than language specific stuff?
		CleanStubs(ctx, e.projectConfig.Root(), maps.Keys(newLanguages)...)
	}
	newStates := map[string]*moduleState{}
	for _, preparedModule := range event.preparedModules {
		name := preparedModule.module.Config.Module
		newLanguage := len(slices.Filter(maps.Values(moduleStates), func(m *moduleState) bool {
			return m.module.Config.Language == preparedModule.module.Config.Language
		})) == 0
		extEvent, err := newModuleBuildWaitingEvent(preparedModule.module.Config)
		if err != nil {
			return errors.Wrapf(err, "failed to watch module %s", name)
		}

		var cancelModuleWatch context.CancelCauseFunc
		var transactionProvider optional.Option[transactionProviderFunc]
		if moduleWatcher != nil {
			txProvider, cancel, err := moduleWatcher(ctx, preparedModule.module.Config, e.internalEvents)
			if err != nil {
				return errors.Wrapf(err, "failed to watch module %s", name)
			}
			cancelModuleWatch = cancel
			transactionProvider = optional.Some(txProvider)
		}
		state := &moduleState{
			module:              preparedModule.module,
			plugin:              preparedModule.plugin,
			configDefaults:      preparedModule.configDefaults,
			needsToBuild:        true,
			lastEvent:           extEvent,
			cancelModuleWatch:   cancelModuleWatch,
			transactionProvider: transactionProvider,
		}
		moduleStates[name] = state
		newStates[name] = state
		e.engineUpdates.Publish(extEvent)

		if newLanguage {
			if err := GenerateStubs(ctx, e.projectConfig.Root(), sch.InternalModules(), moduleStates); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	// New modules need to know which stubs have already been generated
	return errors.Wrapf(SyncStubReferences(ctx, e.projectConfig.Root(), slices.Map(sch.InternalModules(), func(m *schema.Module) string { return m.Name }), newStates, sch), "could not sync stub references after adding new modules")
}

func (e *Engine) handleSchemaUpdateEvent(ctx context.Context, event schemaUpdateEvent, sch *schema.Schema, moduleStates map[string]*moduleState) error {
	deps, err := GraphFromStates(moduleStates, sch, slices.Map(event.modulesWithBreakingChanges, func(moduleRef schema.ModuleRefKey) string { return moduleRef.Module })...)
	if err != nil {
		return errors.Wrapf(err, "failed to get dependencies")
	}
	for _, state := range moduleStates {
		deps := deps[state.module.Config.Module]
		if _, foundBreakingChange := slices.Find(deps, func(dep string) bool {
			return slices.Contains(event.modulesWithBreakingChanges, schema.ModuleRefKey{Realm: sch.FirstInternalRealm().MustGet().Name, Module: dep})
		}); foundBreakingChange {
			// mark all transitive dependencies as dirty
			state.needsToBuild = true
		}
	}

	// TODO: not just internal modules
	if err := GenerateStubs(ctx, e.projectConfig.Root(), slices.Map(event.modulesWithInterfaceChanges, func(moduleRef schema.ModuleRefKey) *schema.Module {
		// TODO: remove MustGet() usage
		return sch.Module(moduleRef.Realm, moduleRef.Module).MustGet()
	}), moduleStates); err != nil {
		return errors.Wrapf(err, "failed to generate stubs for updated modules")
	}
	// All modules need to know which stubs have been generated
	return SyncStubReferences(ctx, e.projectConfig.Root(), slices.Map(sch.InternalModules(), func(m *schema.Module) string { return m.Name }), moduleStates, sch)
}

func (e *Engine) handleBuildEndedEvent(ctx context.Context, event moduleBuildEndedEvent, moduleStates map[string]*moduleState, deployer deployFunc) error {
	logger := log.FromContext(ctx)
	state, ok := moduleStates[event.config.Module]
	if !ok {
		return errors.Errorf("module %s not found in module states", event.config.Module)
	}
	if event.err != nil {
		logger.Scope(event.config.Module).Errorf(event.err, "Build failed")
		extEvent, err := newModuleBuildFailedEvent(event.config, event.err)
		if err != nil {
			return errors.Wrapf(err, "failed to create build failed event for module %s", event.config.Module)
		}
		state.lastEvent = extEvent
		e.engineUpdates.Publish(extEvent)
		return nil
	}
	extEvent, err := newModuleBuildSuccessEvent(event.config)
	if err != nil {
		return errors.Wrapf(err, "failed to create build failed event for module %s", event.config.Module)
	}
	state.lastEvent = extEvent
	e.engineUpdates.Publish(extEvent)

	if deployer(ctx, newPendingModule(state.module, event.tmpDeployDir, event.deployPaths, event.moduleSchema)) {
		extEvent := newModuleDeployWaitingEvent(event.config.Module)
		state.lastEvent = extEvent
		e.engineUpdates.Publish(extEvent)
	}
	return nil
}

func modulesReadyToBuild(ctx context.Context, sch *schema.Schema, moduleStates map[string]*moduleState) (readyToBuild []string, deps map[string][]string, err error) {
	deps, err = GraphFromStates(moduleStates, sch, maps.Keys(moduleStates)...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get dependencies for modules")
	}
	modulesToBuild := slices.Filter(maps.Values(moduleStates), func(state *moduleState) bool {
		if !state.needsToBuild {
			return false
		}
		name := state.module.Config.Module

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
	return slices.Map(modulesToBuild, func(s *moduleState) string {
		return s.module.Config.Module
	}), deps, nil
}

func (e *Engine) handleAnyModulesReadyToBuild(ctx context.Context, sch *schema.Schema, moduleStates map[string]*moduleState, modulesToBuild []string, deps map[string][]string, builder buildFunc) error {
	buildCount := len(slices.Filter(maps.Values(moduleStates), func(state *moduleState) bool {
		return state.lastEvent.Event != nil && state.lastEvent.GetModuleBuildStarted() != nil
	}))
	for _, name := range modulesToBuild {
		if buildCount >= e.parallelism {
			return nil
		}
		state, ok := moduleStates[name]
		if !ok {
			return errors.Errorf("module %s not found in module states", name)
		}
		engineEvent, err := newModuleBuildStartedEvent(state.module.Config)
		if err != nil {
			return errors.Wrapf(err, "failed to create build started event for module %s", state.module.Config.Module)
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

		strippedSch := reflect.DeepCopy(sch)
		modulesToKeep := map[string]bool{}
		visitModuleDependencies(state.module.Config.Module, modulesToKeep, deps)
		for _, module := range strippedSch.InternalModules() {
			if !modulesToKeep[module.Name] {
				// remove module from schema
				strippedSch.RemoveModule(strippedSch.FirstInternalRealm().MustGet().Name, module.Name)
			}
		}

		go builder(ctx, e.projectConfig, state.module, state.plugin, languageplugin.BuildContext{
			Config:       state.module.Config,
			Schema:       strippedSch,
			Dependencies: state.module.Dependencies(Raw),
			BuildEnv:     e.buildEnv,
			Os:           e.os,
			Arch:         e.arch,
		}, e.devMode, e.devModeEndpointUpdates, fileTransaction, e.internalEvents)
		buildCount++
	}
	return nil
}

func visitModuleDependencies(module string, visited map[string]bool, deps map[string][]string) error {
	moduleDeps := deps[module]
	for _, dep := range moduleDeps {
		if !visited[dep] {
			visited[dep] = true
			if err := visitModuleDependencies(dep, visited, deps); err != nil {
				return err
			}
		}
	}
	return nil
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
