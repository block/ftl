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
	metas []moduleMeta
}

func (addMetasEvent) internalEvent() {}

type removeMetaEvent struct {
	config moduleconfig.UnvalidatedModuleConfig
}

func (removeMetaEvent) internalEvent() {}

type schemaUpdateEvent struct {
	newSchema                  *schema.Schema
	modulesWithBreakingChanges []string
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
	// schemaSource *schemaeventsource.EventSource,
	projectConfig projectconfig.Config,
	moduleDirs []string,
	logChanges bool,
	options ...Option,
) (*Engine, error) {
	fmt.Printf("Creating engine\n")
	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("build-engine"))
	// rawEngineUpdates := make(chan *buildenginepb.EngineEvent, 128)

	e := &Engine{
		adminClient:   adminClient,
		projectConfig: projectConfig,
		moduleDirs:    moduleDirs,
		// moduleMetas:   xsync.NewMapOf[string, moduleMeta](),
		// watcher:       watch.NewWatcher(optional.Some(projectConfig.WatchModulesLockPath())),
		// pluginEvents:     make(chan languageplugin.PluginEvent, 128),
		parallelism: runtime.NumCPU(),
		// modulesToBuild:   xsync.NewMapOf[string, bool](),
		// rebuildEvents:    make(chan rebuildEvent, 128),
		// rawEngineUpdates: rawEngineUpdates,
		engineUpdates:  pubsub.New[*buildenginepb.EngineEvent](),
		arch:           runtime.GOARCH, // Default to the local env, we attempt to read these from the cluster later
		os:             runtime.GOOS,
		internalEvents: make(chan internalEvent, 64),
	}

	for _, option := range options {
		option(e)
	}

	// ctx, cancel := context.WithCancelCause(ctx)
	// e.cancel = cancel

	// configs, err := watch.DiscoverModules(ctx, moduleDirs)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not find modules")
	// }

	// err = CleanStubs(ctx, projectConfig.Root(), configs)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to clean stubs")
	// }

	updateTerminalWithEngineEvents(ctx, e.engineUpdates)

	e.updatesService = e.startUpdatesService(ctx)

	// TODO: figure out how to know if there are initial modules
	// go e.watchForEventsToPublish(ctx, true) //len(configs) > 0)

	// wg := &errgroup.Group{}
	// for _, config := range configs {
	// 	wg.Go(func() error {
	// 		meta, err := e.newModuleMeta(ctx, config)
	// 		if err != nil {
	// 			return errors.WithStack(err)
	// 		}
	// 		meta, err = copyMetaWithUpdatedDependencies(ctx, meta)
	// 		if err != nil {
	// 			return errors.WithStack(err)
	// 		}
	// 		e.moduleMetas.Store(config.Module, meta)
	// 		e.modulesToBuild.Store(config.Module, true)
	// 		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
	// 			Timestamp: timestamppb.Now(),
	// 			Event: &buildenginepb.EngineEvent_ModuleAdded{
	// 				ModuleAdded: &buildenginepb.ModuleAdded{
	// 					Module: config.Module,
	// 				},
	// 			},
	// 		}
	// 		return nil
	// 	})
	// }
	// if err := wg.Wait(); err != nil {
	// 	return nil, errors.WithStack(err) //nolint:wrapcheck
	// }
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
	// TODO: logchanges param?
	// TODO: events...
	externalRealms := []*schema.Realm{}
	for name, cfg := range e.projectConfig.ExternalRealms {
		realm, err := realm.GetExternalRealm(ctx, e.projectConfig.ExternalRealmPath(), name, cfg)
		if err != nil {
			return errors.Wrapf(err, "failed to read external realm %s", name)
		}
		externalRealms = append(externalRealms, realm)
	}
	deployCoordinator := NewDeployCoordinator(ctx, e.adminClient, schemaSource, nil, false, e.projectConfig, externalRealms)
	// Save initial schema
	initialSchemaEvent := <-deployCoordinator.SchemaUpdates

	go watchForNewOrRemovedModules(ctx, e.projectConfig, e.moduleDirs, period, e.internalEvents)
	go watchSchemaUpdates(ctx, initialSchemaEvent.schema, deployCoordinator.SchemaUpdates, e.internalEvents)

	// watch for module additions and revovals
	return errors.WithStack(e.processEvents(ctx, initialSchemaEvent.schema, buildModuleAndPublish, moduleWatcherWithPeriod(period)))
}

func (e *Engine) Build(ctx context.Context) error {
	// TODO implement
	return nil
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
			newMetas := make(chan moduleMeta, len(event.Configs))
			group := errgroup.Group{}

			for _, config := range event.Configs {
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
			fmt.Printf("waited for new metas\n")
			collectedMetas := []moduleMeta{}
		collectMetas:
			for {
				select {
				case m := <-newMetas:
					collectedMetas = append(collectedMetas, m)
				default:
					break collectMetas
				}
			}
			fmt.Printf("collected new metas : %v\n", len(collectedMetas))
			if len(collectedMetas) > 0 {
				internalEvents <- addMetasEvent{metas: collectedMetas}
				fmt.Printf("published : %v\n", len(collectedMetas))
			}
		case watch.WatchEventModuleRemoved:
			internalEvents <- removeMetaEvent{config: event.Config}

		case watch.WatchEventModuleChanged:
			// Changes within a module are not handled here
		}
	}
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
	moduleHashes := map[string][]byte{}
	for _, module := range initialSchema.InternalModules() {
		hash, err := computeModuleHash(module)
		if err != nil {
			logger.Errorf(err, "compute hash for %s failed", module.Name)
			continue
		}
		moduleHashes[module.Name] = hash
	}

	internalRealm, ok := slices.Find(initialSchema.Realms, func(r *schema.Realm) bool {
		return r.External == false
	})
	if !ok {
		logger.Logf(log.Error, "no internal realm found in schema")
		return
	}

	for event := range channels.IterContext(ctx, schemaUpdates) {
		materiallyChangedModules := slices.Filter(maps.Keys(event.updatedModules), func(module string) bool {
			moduleSch, ok := event.schema.Module(internalRealm.Name, module).Get()
			if !ok {
				logger.Logf(log.Error, "module %s not found in schema", module)
				return false
			}
			hash, err := computeModuleHash(moduleSch)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", module)
				return false
			}
			existingHash, ok := moduleHashes[module]
			if !ok {
				existingHash = []byte{}
			}

			if bytes.Equal(hash, existingHash) {
				logger.Tracef("schema for %s has not changed", module)
				return false
			}

			moduleHashes[module] = hash
			return true
		})
		internalEvents <- schemaUpdateEvent{
			newSchema:                  event.schema,
			modulesWithBreakingChanges: materiallyChangedModules,
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

func (e *Engine) processEvents(ctx context.Context, initialSchema *schema.Schema, builder buildFunc, moduleWatcher moduleWatcherFunc) error {
	logger := log.FromContext(ctx)
	sch := initialSchema
	dirtyModules := map[string]bool{}
	moduleMetas := map[string]moduleMeta{}
	moduleStates := map[string]*buildenginepb.EngineEvent{}
	cancelModuleWatchMap := map[string]context.CancelCauseFunc{}
	moduleTransactionProviders := map[string]transactionProviderFunc{}

	fmt.Printf("Initial schema: %s\n", sch.String())

	idle := true
	// var endTime time.Time
	var becomeIdleTimer <-chan time.Time

	for {

		fmt.Printf("Waiting for events...\n")

		events := []internalEvent{}
		select {
		case <-becomeIdleTimer:
			becomeIdleTimer = nil
			if !isIdle(moduleStates) {
				continue
			}
			idle = true

			// TODO: pass in module errors
			e.engineUpdates.Publish(newEngineEndedEvent(moduleStates, moduleMetas, nil))
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
			fmt.Printf("received event: %T\n", event)
			switch event := event.(type) {
			case addMetasEvent:
				for _, meta := range event.metas {
					name := meta.module.Config.Module
					newLanguage := len(slices.Filter(maps.Values(moduleMetas), func(m moduleMeta) bool {
						return m.module.Config.Language == meta.module.Config.Language
					})) == 0
					if newLanguage {
						// clean stubs for the language if no modules are present
						// TODO: does this clean more than language specific stuff?
						CleanStubs(ctx, e.projectConfig.Root(), meta.module.Config.Language)

					}
					state, err := newModuleBuildWaitingEvent(meta.module.Config)
					if err != nil {
						logger.Errorf(err, "failed to watch module %s", name)
						continue
					}
					moduleMetas[name] = meta
					transactionProvider, cancel, err := moduleWatcher(ctx, meta.module.Config, e.internalEvents)
					if err != nil {
						logger.Errorf(err, "failed to watch module %s", name)
						continue
					}
					cancelModuleWatchMap[name] = cancel
					moduleTransactionProviders[name] = transactionProvider
					dirtyModules[name] = true
					moduleStates[name] = state
					e.engineUpdates.Publish(state)

					if newLanguage {
						// TODO: not a good place for this
						// TODO: stub external modules too...
						err := GenerateStubs(ctx, e.projectConfig.Root(), sch.InternalModules(), moduleMetas)
						if err != nil {
							return errors.WithStack(err)
						}
					}
				}

			case removeMetaEvent:
				if cancel, ok := cancelModuleWatchMap[event.config.Module]; ok {
					cancel(errors.Wrap(context.Canceled, "module removed"))
				}
				delete(moduleMetas, event.config.Module)
				delete(cancelModuleWatchMap, event.config.Module)
				delete(dirtyModules, event.config.Module)
				delete(moduleStates, event.config.Module)
				delete(moduleTransactionProviders, event.config.Module)
				e.engineUpdates.Publish(newModuleRemovedEvent(event.config.Module))

			case schemaUpdateEvent:
				sch = event.newSchema
				for _, module := range event.modulesWithBreakingChanges {
					deps, err := Graph(moduleMetas, sch, module)
					if err != nil {
						// TODO: handle error
						continue
					}
					for _, dep := range deps[module] {
						// mark all transitive dependencies as dirty
						dirtyModules[dep] = true
					}
				}
			case moduleNeedsToBuildEvent:
				dirtyModules[event.module] = true

			case moduleBuildEndedEvent:
				if event.err != nil {
					extEvent, err := newModuleBuildFailedEvent(event.config, event.err)
					if err != nil {
						logger.Errorf(err, "failed to create build failed event for module %s", event.config.Module)
						continue
					}
					moduleStates[event.config.Module] = extEvent
					e.engineUpdates.Publish(extEvent)
					continue
				}

				err := GenerateStubs(ctx, e.projectConfig.Root(), []*schema.Module{event.moduleSchema}, moduleMetas)
				if err != nil {
					return errors.WithStack(err)
				}
			}
		}

		//
		// Kick off any builds that we can
		//
		modulesToBuild := slices.Filter(maps.Keys(dirtyModules), func(module string) bool {
			deps, err := Graph(moduleMetas, sch, module)
			if err != nil {
				// TODO: handle error
				return false
			}
			switch moduleStates[module].Event.(type) {
			case *buildenginepb.EngineEvent_ModuleBuildStarted,
				*buildenginepb.EngineEvent_ModuleDeployStarted:
				return false
			default:
			}
			for _, dep := range deps[module] {
				if _, ok := dirtyModules[dep]; ok {
					return false // dependency is dirty, cannot build yet
				}
				if state, ok := moduleStates[dep]; ok && state.Event != nil && state.GetModuleBuildStarted() != nil {
					return false
				}
			}
			return true
		})
		buildCount := len(slices.Filter(maps.Values(moduleStates), func(state *buildenginepb.EngineEvent) bool {
			return state.Event != nil && state.GetModuleBuildStarted() != nil
		}))
		for _, module := range modulesToBuild {
			if buildCount >= e.parallelism {
				break
			}
			meta, ok := moduleMetas[module]
			if !ok {
				logger.Logf(log.Error, "module %s not found in module metas", module)
				continue
			}
			engineEvent, err := newModuleBuildStartedEvent(meta.module.Config)
			if err != nil {
				logger.Errorf(err, "failed to create build started event for module %s", module)
				continue
			}
			transactionProvider := moduleTransactionProviders[module]

			dirtyModules[module] = false
			moduleStates[module] = engineEvent
			e.engineUpdates.Publish(engineEvent)

			go builder(ctx, e.projectConfig, meta.module, moduleMetas[module].plugin, languageplugin.BuildContext{
				Config:       meta.module.Config,
				Schema:       sch,
				Dependencies: meta.module.Dependencies(Raw),
				BuildEnv:     e.buildEnv,
				Os:           e.os,
				Arch:         e.arch,
			}, e.devMode, e.devModeEndpointUpdates, transactionProvider, e.internalEvents)
			buildCount++
		}

		//
		// Kick off any deploys that we can
		//

		// TODO: implement deploy

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

func newEngineEndedEvent(moduleStates map[string]*buildenginepb.EngineEvent, metas map[string]moduleMeta, moduleErrors map[string]*langpb.ErrorList) *buildenginepb.EngineEvent {
	modulesOutput := []*buildenginepb.EngineEnded_Module{}
	for module := range moduleStates {
		meta, ok := metas[module]
		if !ok {
			continue
		}
		modulesOutput = append(modulesOutput, &buildenginepb.EngineEnded_Module{
			Module: module,
			Path:   meta.module.Config.Dir,
			Errors: moduleErrors[module],
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

func isIdle(moduleStates map[string]*buildenginepb.EngineEvent) bool {
	if len(moduleStates) == 0 {
		return true
	}
	for _, state := range moduleStates {
		switch state.Event.(type) {
		case *buildenginepb.EngineEvent_ModuleBuildStarted,
			*buildenginepb.EngineEvent_ModuleDeployStarted:
			return false
		default:
		}
	}
	return true
}
