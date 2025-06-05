package buildengine

// // Close stops the Engine's schema sync.
// func (e *Engine) Close() error {
// 	e.cancel(errors.Wrap(context.Canceled, "build engine stopped"))
// 	return nil
// }

// func (e *Engine) GetSchema() (*schema.Schema, bool) {
// 	sch := e.targetSchema.Load()
// 	if sch == nil {
// 		return nil, false
// 	}
// 	return sch, true
// }

// func (e *Engine) GetModuleSchema(moduleName string) (*schema.Module, bool) {
// 	sch := e.targetSchema.Load()
// 	if sch == nil {
// 		return nil, false
// 	}
// 	module, ok := slices.Find(sch.InternalModules(), func(m *schema.Module) bool {
// 		return m.Name == moduleName
// 	})
// 	if !ok {
// 		return nil, false
// 	}
// 	return module, true
// }

// Import manually imports a schema for a module as if it were retrieved from
// the FTL controller.
// func (e *Engine) Import(ctx context.Context, realmName string, moduleSch *schema.Module) {
// 	sch := reflect.DeepCopy(e.targetSchema.Load())
// 	for _, realm := range sch.Realms {
// 		if realm.Name != realmName {
// 			continue
// 		}
// 		realm.Modules = slices.Filter(realm.Modules, func(m *schema.Module) bool {
// 			return m.Name != moduleSch.Name
// 		})
// 		realm.Modules = append(realm.Modules, moduleSch)
// 		break
// 	}
// 	e.targetSchema.Store(sch)
// }

// // Each iterates over all local modules.
// func (e *Engine) Each(fn func(Module) error) (err error) {
// 	e.moduleMetas.Range(func(key string, value moduleMeta) bool {
// 		if ferr := fn(value.module); ferr != nil {
// 			err = errors.Wrapf(ferr, "%s", key)
// 			return false
// 		}
// 		return true
// 	})
// 	err = errors.WithStack(err)
// 	return
// }

// // Modules returns the names of all modules.
// func (e *Engine) Modules() []string {
// 	var moduleNames []string
// 	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
// 		moduleNames = append(moduleNames, name)
// 		return true
// 	})
// 	return moduleNames
// }

// type moduleState int

// const (
// 	moduleStateBuildWaiting moduleState = iota
// 	moduleStateExplicitlyBuilding
// 	moduleStateAutoRebuilding
// 	moduleStateBuilt
// 	moduleStateDeployWaiting
// 	moduleStateDeploying
// 	moduleStateDeployed
// 	moduleStateFailed
// )

// func (e *Engine) getDependentModuleNames(moduleName string) []string {
// 	dependentModuleNames := map[string]bool{}
// 	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
// 		for _, dep := range meta.module.Dependencies(AlwaysIncludeBuiltin) {
// 			if dep == moduleName {
// 				dependentModuleNames[name] = true
// 			}
// 		}
// 		return true
// 	})
// 	return maps.Keys(dependentModuleNames)
// }

// // BuildAndDeploy attempts to build and deploy all local modules.
// func (e *Engine) BuildAndDeploy(ctx context.Context, replicas optional.Option[int32], waitForDeployOnline bool, singleChangeset bool, moduleNames ...string) (err error) {
// 	logger := log.FromContext(ctx)
// 	if len(moduleNames) == 0 {
// 		moduleNames = e.Modules()
// 	}
// 	if len(moduleNames) == 0 {
// 		return nil
// 	}

// 	defer func() {
// 		if err == nil {
// 			return
// 		}
// 		pendingInitialBuilds := []string{}
// 		e.modulesToBuild.Range(func(name string, value bool) bool {
// 			if value {
// 				pendingInitialBuilds = append(pendingInitialBuilds, name)
// 			}
// 			return true
// 		})

// 		// Print out all modules that have yet to build if there are any errors
// 		if len(pendingInitialBuilds) > 0 {
// 			logger.Infof("Modules waiting to build: %s", strings.Join(pendingInitialBuilds, ", "))
// 		}
// 	}()

// 	modulesToDeploy := [](*pendingModule){}
// 	buildErr := e.buildWithCallback(ctx, func(buildCtx context.Context, module Module, moduleSch *schema.Module, tmpDeployDir string, deployPaths []string) error {
// 		e.modulesToBuild.Store(module.Config.Module, false)
// 		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 			Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
// 				ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
// 					Module: module.Config.Module,
// 				},
// 			},
// 		}
// 		pendingDeployModule := newPendingModule(module, tmpDeployDir, deployPaths, moduleSch)
// 		if singleChangeset {
// 			modulesToDeploy = append(modulesToDeploy, pendingDeployModule)
// 			return nil
// 		}
// 		deployErr := make(chan error, 1)
// 		go func() {
// 			deployErr <- e.deployCoordinator.deploy(ctx, []*pendingModule{pendingDeployModule}, replicas)
// 		}()
// 		if waitForDeployOnline {
// 			return errors.WithStack(<-deployErr)
// 		}
// 		return nil
// 	}, moduleNames...)
// 	if buildErr != nil {
// 		return errors.WithStack(buildErr)
// 	}

// 	deployGroup := &errgroup.Group{}
// 	deployGroup.Go(func() error {
// 		// Wait for all build attempts to complete
// 		if singleChangeset {
// 			// Queue the modules for deployment instead of deploying directly
// 			return errors.WithStack(e.deployCoordinator.deploy(ctx, modulesToDeploy, replicas))
// 		}
// 		return nil
// 	})
// 	if waitForDeployOnline {
// 		err := deployGroup.Wait()
// 		return errors.WithStack(err) //nolint:wrapcheck
// 	}
// 	return nil
// }

// type buildCallback func(ctx context.Context, module Module, moduleSch *schema.Module, tmpDeployDir string, deployPaths []string) error

// func (e *Engine) handleDependencyCycleError(ctx context.Context, depErr DependencyCycleError, graph map[string][]string, callback buildCallback) error {
// 	// Mark each cylic module as having an error
// 	for _, module := range depErr.Modules {
// 		meta, ok := e.moduleMetas.Load(module)
// 		if !ok {
// 			return errors.Errorf("module %q not found in dependency cycle", module)
// 		}
// 		configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
// 		if err != nil {
// 			return errors.Wrap(err, "failed to marshal module config")
// 		}
// 		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 			Timestamp: timestamppb.Now(),
// 			Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
// 				ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
// 					Config: configProto,
// 					Errors: &langpb.ErrorList{
// 						Errors: []*langpb.Error{
// 							{
// 								Msg:   depErr.Error(),
// 								Level: langpb.Error_ERROR_LEVEL_ERROR,
// 								Type:  langpb.Error_ERROR_TYPE_FTL,
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}
// 	}

// 	// Build the remaining modules
// 	remaining := slices.Filter(maps.Keys(graph), func(module string) bool {
// 		return !slices.Contains(depErr.Modules, module) && module != "builtin"
// 	})
// 	if len(remaining) == 0 {
// 		return nil
// 	}
// 	remainingModulesErr := e.buildWithCallback(ctx, callback, remaining...)
// 	return errors.WithStack(remainingModulesErr)
// }

// func (e *Engine) tryBuild(ctx context.Context, mustBuild map[string]bool, moduleName string, builtModules map[string]*schema.Module, schemas chan *schema.Module, callback buildCallback) error {
// 	logger := log.FromContext(ctx)

// 	if !mustBuild[moduleName] {
// 		return errors.WithStack(e.mustSchema(ctx, moduleName, builtModules, schemas))
// 	}

// 	meta, ok := e.moduleMetas.Load(moduleName)
// 	if !ok {
// 		return errors.Errorf("module %q not found", moduleName)
// 	}

// 	for _, dep := range meta.module.Dependencies(Raw) {
// 		if _, ok := builtModules[dep]; !ok {
// 			logger.Warnf("build skipped because dependency %q failed to build", dep)
// 			return nil
// 		}
// 	}

// 	configProto, err := langpb.ModuleConfigToProto(meta.module.Config.Abs())
// 	if err != nil {
// 		return errors.Wrap(err, "failed to marshal module config")
// 	}
// 	e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 		Timestamp: timestamppb.Now(),
// 		Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
// 			ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
// 				Config:        configProto,
// 				IsAutoRebuild: false,
// 			},
// 		},
// 	}

// 	moduleSch, tmpDeployDir, deployPaths, err := e.build(ctx, moduleName, builtModules, schemas)
// 	if err == nil && callback != nil {
// 		// load latest meta as it may have been updated
// 		meta, ok = e.moduleMetas.Load(moduleName)
// 		if !ok {
// 			return errors.Errorf("module %q not found", moduleName)
// 		}
// 		return errors.WithStack(callback(ctx, meta.module, moduleSch, tmpDeployDir, deployPaths))
// 	}

// 	return errors.WithStack(err)
// }

// // Publish either the schema from the FTL controller, or from a local build.
// func (e *Engine) mustSchema(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
// 	if sch, ok := e.GetModuleSchema(moduleName); ok {
// 		schemas <- sch
// 		return nil
// 	}
// 	sch, _, _, err := e.build(ctx, moduleName, builtModules, schemas)
// 	schemas <- sch
// 	return errors.WithStack(err)
// }

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
// func (e *Engine) build(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) (moduleSch *schema.Module, tmpDeployDir string, deployPaths []string, err error) {
// 	meta, ok := e.moduleMetas.Load(moduleName)
// 	if !ok {
// 		return nil, "", nil, errors.Errorf("module %q not found", moduleName)
// 	}

// 	sch := &schema.Schema{Realms: []*schema.Realm{{Modules: maps.Values(builtModules)}}} //nolint:exptostd

// 	config := meta.module.Config.Abs()
// 	configProto, err := langpb.ModuleConfigToProto(config)
// 	if err != nil {
// 		return nil, "", nil, errors.Wrap(err, "failed to marshal module config")
// 	}
// 	if meta.module.SQLError != nil {
// 		meta.module = meta.module.CopyWithSQLErrors(nil)
// 		e.moduleMetas.Store(moduleName, meta)
// 	}
// 	transaction := meta.watcher.GetTransaction(config.Dir)
// 	if err := transaction.Begin(); err != nil {
// 		return nil, "", nil, errors.Wrapf(err, "failed to begin file transaction for %s", config.Dir)
// 	}
// 	moduleSchema, tmpDeployDir, deployPaths, err := build(ctx, e.projectConfig, meta.module, meta.plugin, transaction, languageplugin.BuildContext{
// 		Config:       meta.module.Config,
// 		Schema:       sch,
// 		Dependencies: meta.module.Dependencies(Raw),
// 		BuildEnv:     e.buildEnv,
// 		Os:           e.os,
// 		Arch:         e.arch,
// 	}, e.devMode, e.devModeEndpointUpdates)
// 	if err := transaction.End(); err != nil {
// 		return nil, "", nil, errors.Wrapf(err, "failed to end file transaction for %s", config.Dir)
// 	}

// 	if err != nil {
// 		if errors.Is(err, errSQLError) {
// 			// Keep sql error around so that subsequent auto rebuilds from the plugin keep the sql error
// 			meta.module = meta.module.CopyWithSQLErrors(err)
// 			e.moduleMetas.Store(moduleName, meta)
// 		}
// 		if errors.Is(err, errInvalidateDependencies) {
// 			e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 				Timestamp: timestamppb.Now(),
// 				Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
// 					ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
// 						Config: configProto,
// 					},
// 				},
// 			}
// 			// Do not start a build directly as we are already building out a graph of modules.
// 			// Instead we send to a chan so that it can be processed after.
// 			e.rebuildEvents <- rebuildRequestEvent{module: moduleName}
// 			return nil, "", nil, errors.WithStack(err)
// 		}
// 		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 			Timestamp: timestamppb.Now(),
// 			Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
// 				ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
// 					Config:        configProto,
// 					IsAutoRebuild: false,
// 					Errors: &langpb.ErrorList{
// 						Errors: errorToLangError(err),
// 					},
// 				},
// 			},
// 		}
// 		return nil, "", nil, errors.WithStack(err)
// 	}

// 	e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 		Timestamp: timestamppb.Now(),
// 		Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
// 			ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
// 				Config:        configProto,
// 				IsAutoRebuild: false,
// 			},
// 		},
// 	}

// 	schemas <- moduleSchema
// 	return moduleSchema, tmpDeployDir, deployPaths, nil
// }

// Construct a combined schema for a module and its transitive dependencies.
// func (e *Engine) gatherSchemas(
// 	moduleSchemas map[string]*schema.Module,
// 	out map[string]*schema.Module,
// ) error {
// 	for _, sch := range e.targetSchema.Load().InternalModules() {
// 		out[sch.Name] = sch
// 	}

// 	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
// 		if _, ok := moduleSchemas[name]; ok {
// 			out[name] = moduleSchemas[name]
// 		} else {
// 			// We don't want to use a remote schema if we have it locally
// 			delete(out, name)
// 		}
// 		return true
// 	})

// 	return nil
// }

// func (e *Engine) syncNewStubReferences(ctx context.Context, newModules map[string]*schema.Module, metasMap map[string]moduleMeta) error {
// 	fullSchema := &schema.Schema{} //nolint:exptostd
// 	for _, r := range e.targetSchema.Load().Realms {
// 		realm := &schema.Realm{
// 			Name:     r.Name,
// 			External: r.External,
// 		}
// 		if !realm.External {
// 			realm.Modules = maps.Values(newModules)
// 		}

// 		for _, module := range r.Modules {
// 			if _, ok := newModules[module.Name]; !ok || realm.External {
// 				realm.Modules = append(realm.Modules, module)
// 			}
// 		}
// 		sort.SliceStable(realm.Modules, func(i, j int) bool {
// 			return realm.Modules[i].Name < realm.Modules[j].Name
// 		})
// 		fullSchema.Realms = append(fullSchema.Realms, realm)
// 	}

// 	return errors.WithStack(SyncStubReferences(ctx,
// 		e.projectConfig.Root(),
// 		slices.Map(fullSchema.InternalModules(), func(m *schema.Module) string { return m.Name }),
// 		metasMap,
// 		fullSchema))
// }

// func (e *Engine) newModuleMeta(ctx context.Context, config moduleconfig.UnvalidatedModuleConfig) (moduleMeta, error) {

// // watchForEventsToPublish listens for raw build events, collects state, and publishes public events to BuildUpdates topic.
// func (e *Engine) watchForEventsToPublish(ctx context.Context, hasInitialModules bool) {
// 	logger := log.FromContext(ctx)

// 	moduleErrors := map[string]*langpb.ErrorList{}
// 	moduleStates := map[string]moduleState{}

// 	idle := true
// 	var endTime time.Time
// 	var becomeIdleTimer <-chan time.Time

// 	isFirstRound := hasInitialModules

// 	addTimestamp := func(evt *buildenginepb.EngineEvent) {
// 		if evt.Timestamp == nil {
// 			evt.Timestamp = timestamppb.Now()
// 		}
// 	}

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return

// 		case <-becomeIdleTimer:
// 			becomeIdleTimer = nil
// 			if !e.isIdle(moduleStates) {
// 				continue
// 			}
// 			idle = true

// 			if e.devMode && isFirstRound {
// 				if len(moduleErrors) > 0 {
// 					var errs []error
// 					for module, errList := range moduleErrors {
// 						if errList != nil && len(errList.Errors) > 0 {
// 							moduleErr := errors.Errorf("%s: %s", module, langpb.ErrorListString(errList))
// 							errs = append(errs, moduleErr)
// 						}
// 					}
// 					if len(errs) > 1 {
// 						logger.Logf(log.Error, "Initial build failed:\n%s", strings.Join(slices.Map(errs, func(err error) string {
// 							return fmt.Sprintf("  %s", err)
// 						}), "\n"))
// 					} else {
// 						logger.Errorf(errors.Join(errs...), "Initial build failed")
// 					}
// 				} else if start, ok := e.startTime.Get(); ok {
// 					e.startTime = optional.None[time.Time]()
// 					logger.Infof("All modules deployed in %.2fs, watching for changes...", endTime.Sub(start).Seconds())
// 				} else {
// 					logger.Infof("All modules deployed, watching for changes...")
// 				}
// 			}
// 			isFirstRound = false

// 			modulesOutput := []*buildenginepb.EngineEnded_Module{}
// 			for module := range moduleStates {
// 				meta, ok := e.moduleMetas.Load(module)
// 				if !ok {
// 					continue
// 				}
// 				modulesOutput = append(modulesOutput, &buildenginepb.EngineEnded_Module{
// 					Module: module,
// 					Path:   meta.module.Config.Dir,
// 					Errors: moduleErrors[module],
// 				})
// 			}
// 			evt := &buildenginepb.EngineEvent{
// 				Timestamp: timestamppb.Now(),
// 				Event: &buildenginepb.EngineEvent_EngineEnded{
// 					EngineEnded: &buildenginepb.EngineEnded{
// 						Modules: modulesOutput,
// 					},
// 				},
// 			}
// 			addTimestamp(evt)
// 			e.engineUpdates.Publish(evt)

// 		case evt := <-e.rawEngineUpdates:
// 			switch rawEvent := evt.Event.(type) {
// 			case *buildenginepb.EngineEvent_ModuleAdded:

// 			case *buildenginepb.EngineEvent_ModuleRemoved:
// 				delete(moduleErrors, rawEvent.ModuleRemoved.Module)
// 				delete(moduleStates, rawEvent.ModuleRemoved.Module)

// 			case *buildenginepb.EngineEvent_ModuleBuildWaiting:
// 				moduleStates[rawEvent.ModuleBuildWaiting.Config.Name] = moduleStateBuildWaiting

// 			case *buildenginepb.EngineEvent_ModuleBuildStarted:
// 				if idle {
// 					idle = false
// 					started := &buildenginepb.EngineEvent{
// 						Timestamp: timestamppb.Now(),
// 						Event: &buildenginepb.EngineEvent_EngineStarted{
// 							EngineStarted: &buildenginepb.EngineStarted{},
// 						},
// 					}
// 					addTimestamp(started)
// 					e.engineUpdates.Publish(started)
// 				}
// 				if rawEvent.ModuleBuildStarted.IsAutoRebuild {
// 					moduleStates[rawEvent.ModuleBuildStarted.Config.Name] = moduleStateAutoRebuilding
// 				} else {
// 					moduleStates[rawEvent.ModuleBuildStarted.Config.Name] = moduleStateExplicitlyBuilding
// 				}
// 				delete(moduleErrors, rawEvent.ModuleBuildStarted.Config.Name)
// 				logger.Module(rawEvent.ModuleBuildStarted.Config.Name).Scope("build").Debugf("Building...")
// 			case *buildenginepb.EngineEvent_ModuleBuildFailed:
// 				moduleStates[rawEvent.ModuleBuildFailed.Config.Name] = moduleStateFailed
// 				moduleErrors[rawEvent.ModuleBuildFailed.Config.Name] = rawEvent.ModuleBuildFailed.Errors
// 				moduleErr := errors.Errorf("%s", langpb.ErrorListString(rawEvent.ModuleBuildFailed.Errors))
// 				logger.Module(rawEvent.ModuleBuildFailed.Config.Name).Scope("build").Errorf(moduleErr, "Build failed")
// 			case *buildenginepb.EngineEvent_ModuleBuildSuccess:
// 				moduleStates[rawEvent.ModuleBuildSuccess.Config.Name] = moduleStateBuilt
// 				delete(moduleErrors, rawEvent.ModuleBuildSuccess.Config.Name)
// 			case *buildenginepb.EngineEvent_ModuleDeployWaiting:
// 				moduleStates[rawEvent.ModuleDeployWaiting.Module] = moduleStateDeployWaiting
// 			case *buildenginepb.EngineEvent_ModuleDeployStarted:
// 				if idle {
// 					idle = false
// 					started := &buildenginepb.EngineEvent{
// 						Timestamp: timestamppb.Now(),
// 						Event: &buildenginepb.EngineEvent_EngineStarted{
// 							EngineStarted: &buildenginepb.EngineStarted{},
// 						},
// 					}
// 					addTimestamp(started)
// 					e.engineUpdates.Publish(started)
// 				}
// 				moduleStates[rawEvent.ModuleDeployStarted.Module] = moduleStateDeploying
// 				delete(moduleErrors, rawEvent.ModuleDeployStarted.Module)
// 			case *buildenginepb.EngineEvent_ModuleDeployFailed:
// 				moduleStates[rawEvent.ModuleDeployFailed.Module] = moduleStateFailed
// 				moduleErrors[rawEvent.ModuleDeployFailed.Module] = rawEvent.ModuleDeployFailed.Errors
// 			case *buildenginepb.EngineEvent_ModuleDeploySuccess:
// 				moduleStates[rawEvent.ModuleDeploySuccess.Module] = moduleStateDeployed
// 				delete(moduleErrors, rawEvent.ModuleDeploySuccess.Module)
// 			}

// 			addTimestamp(evt)
// 			e.engineUpdates.Publish(evt)
// 		}
// 		if !idle && e.isIdle(moduleStates) {
// 			endTime = time.Now()
// 			becomeIdleTimer = time.After(time.Millisecond * 200)
// 		}
// 	}
// }

// BuildWithCallback attempts to build all local modules, and calls back with the result
// func (e *Engine) BuildWithCallback(ctx context.Context, callback func(ctx context.Context, module Module, moduleSch *schema.Module, tmpDeployDir string, deployPaths []string) error) error {
// 	schemas := make(chan *schema.Module, e.moduleMetas.Size())
// 	if err := e.buildWithCallback(ctx, func(ctx context.Context, module Module, moduleSch *schema.Module, tmpDeployDir string, deployPaths []string) error {
// 		schemas <- moduleSch
// 		if callback != nil {
// 			err := callback(ctx, module, moduleSch, tmpDeployDir, deployPaths)
// 			if err != nil {
// 				return errors.Wrapf(err, "build callback failed")
// 			}
// 		}
// 		return nil
// 	}); err != nil {
// 		return errors.WithStack(err)
// 	}
// 	close(schemas)

// 	realm := &schema.Realm{
// 		Name: e.projectConfig.Name,
// 	}
// 	for moduleSch := range schemas {
// 		realm.Modules = append(realm.Modules, moduleSch)
// 	}
// 	e.targetSchema.Store(&schema.Schema{
// 		Realms: []*schema.Realm{realm},
// 	})
// 	return nil
// }

// func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, moduleNames ...string) error {
// 	logger := log.FromContext(ctx)
// 	if len(moduleNames) == 0 {
// 		e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
// 			moduleNames = append(moduleNames, name)
// 			return true
// 		})
// 	}

// 	mustBuildChan := make(chan moduleconfig.ModuleConfig, len(moduleNames))
// 	wg := errgroup.Group{}
// 	for _, name := range moduleNames {
// 		wg.Go(func() error {
// 			meta, ok := e.moduleMetas.Load(name)
// 			if !ok {
// 				return errors.Errorf("module %q not found", name)
// 			}

// 			meta, err := copyMetaWithUpdatedDependencies(ctx, meta)
// 			if err != nil {
// 				return errors.Wrapf(err, "could not get dependencies for %s", name)
// 			}

// 			e.moduleMetas.Store(name, meta)
// 			mustBuildChan <- meta.module.Config
// 			return nil
// 		})
// 	}
// 	if err := wg.Wait(); err != nil {
// 		return errors.WithStack(err) //nolint:wrapcheck
// 	}
// 	close(mustBuildChan)
// 	mustBuild := map[string]bool{}
// 	jvm := false
// 	for config := range mustBuildChan {
// 		if config.Language == "java" || config.Language == "kotlin" {
// 			jvm = true
// 		}
// 		mustBuild[config.Module] = true
// 		proto, err := langpb.ModuleConfigToProto(config.Abs())
// 		if err != nil {
// 			logger.Errorf(err, "failed to marshal module config")
// 			continue
// 		}
// 		e.rawEngineUpdates <- &buildenginepb.EngineEvent{
// 			Timestamp: timestamppb.Now(),
// 			Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
// 				ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
// 					Config: proto,
// 				},
// 			},
// 		}
// 	}
// 	if jvm {
// 		// Huge hack that is just for development
// 		// In release builds this is a noop
// 		// This makes sure the JVM jars are up to date when running from source
// 		buildRequiredJARS(ctx)
// 	}

// 	graph, err := e.Graph(moduleNames...)
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
// 	builtModules := map[string]*schema.Module{
// 		"builtin": schema.Builtins(),
// 	}

// 	metasMap := map[string]moduleMeta{}
// 	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
// 		metasMap[name] = meta
// 		return true
// 	})
// 	err = GenerateStubs(ctx, e.projectConfig.Root(), maps.Values(builtModules), metasMap)
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}

// 	topology, topoErr := TopologicalSort(graph)
// 	if topoErr != nil {
// 		var dependencyCycleErr DependencyCycleError
// 		if !errors.As(topoErr, &dependencyCycleErr) {
// 			return errors.WithStack(topoErr)
// 		}
// 		if err := e.handleDependencyCycleError(ctx, dependencyCycleErr, graph, callback); err != nil {
// 			return errors.WithStack(errors.Join(err, topoErr))
// 		}
// 		return errors.WithStack(topoErr)
// 	}
// 	errCh := make(chan error, 1024)
// 	for _, group := range topology {
// 		knownSchemas := map[string]*schema.Module{}
// 		err := e.gatherSchemas(builtModules, knownSchemas)
// 		if err != nil {
// 			return errors.WithStack(err)
// 		}

// 		// Collect schemas to be inserted into "built" map for subsequent groups.
// 		schemas := make(chan *schema.Module, len(group))

// 		wg := errgroup.Group{}
// 		wg.SetLimit(e.parallelism)

// 		logger.Debugf("Building group: %v", group)
// 		for _, moduleName := range group {
// 			wg.Go(func() error {
// 				logger := log.FromContext(ctx).Module(moduleName).Scope("build")
// 				ctx := log.ContextWithLogger(ctx, logger)
// 				err := e.tryBuild(ctx, mustBuild, moduleName, builtModules, schemas, callback)
// 				if err != nil {
// 					errCh <- err
// 				}
// 				return nil
// 			})
// 		}

// 		err = wg.Wait()
// 		if err != nil {
// 			return errors.WithStack(err)
// 		}

// 		// Now this group is built, collect all the schemas.
// 		close(schemas)
// 		newSchemas := []*schema.Module{}
// 		for sch := range schemas {
// 			builtModules[sch.Name] = sch
// 			newSchemas = append(newSchemas, sch)
// 		}

// 		err = GenerateStubs(ctx, e.projectConfig.Root(), newSchemas, metasMap)
// 		if err != nil {
// 			return errors.WithStack(err)
// 		}

// 		// Sync references to stubs if needed by the runtime
// 		err = e.syncNewStubReferences(ctx, builtModules, metasMap)
// 		if err != nil {
// 			return errors.WithStack(err)
// 		}
// 	}

// 	close(errCh)
// 	allErrors := []error{}
// 	for err := range errCh {
// 		allErrors = append(allErrors, err)
// 	}

// 	if len(allErrors) > 0 {
// 		return errors.WithStack(errors.Join(allErrors...))
// 	}

// 	return nil
// }
