package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type schemaChange struct {
	ChangeType ftlv1.DeploymentChangeType
	*schema.Module
}

// Engine for building a set of modules.
type Engine struct {
	client           ftlv1connect.ControllerServiceClient
	projects         map[string]Project
	dirs             []string
	externalDirs     []string
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schemaChange]
	cancel           func()
	parallelism      int
}

type Option func(o *Engine)

func Parallelism(n int) Option {
	return func(o *Engine) {
		o.parallelism = n
	}
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
//
// "dirs" are directories to scan for local modules.
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient, dirs []string, externalDirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		dirs:             dirs,
		externalDirs:     externalDirs,
		projects:         map[string]Project{},
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schemaChange](),
		parallelism:      runtime.NumCPU(),
	}
	for _, option := range options {
		option(e)
	}
	e.controllerSchema.Store("builtin", schema.Builtins())
	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	modules, err := DiscoverModules(ctx, dirs...)
	if err != nil {
		return nil, err
	}
	for _, module := range modules {
		project, err := UpdateDependencies(ctx, Project(&module))
		if err != nil {
			return nil, err
		}
		e.projects[module.Key()] = project
	}

	for _, dir := range externalDirs {
		lib, err := LoadExternalLibrary(dir)
		if err != nil {
			return nil, err
		}
		project, err := UpdateDependencies(ctx, Project(&lib))
		if err != nil {
			return nil, err
		}
		e.projects[project.Key()] = project
	}

	if client == nil {
		return e, nil
	}
	schemaSync := e.startSchemaSync(ctx)
	go rpc.RetryStreamingServerStream(ctx, backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, schemaSync)
	return e, nil
}

// Sync module schema changes from the FTL controller, as well as from manual
// updates, and merge them into a single schema map.
func (e *Engine) startSchemaSync(ctx context.Context) func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	logger := log.FromContext(ctx)
	// Blocking schema sync from the controller.
	psch, err := e.client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err == nil {
		sch, err := schema.FromProto(psch.Msg.Schema)
		if err == nil {
			for _, module := range sch.Modules {
				e.controllerSchema.Store(module.Name, module)
			}
		} else {
			logger.Debugf("Failed to parse schema from controller: %s", err)
		}
	} else {
		logger.Debugf("Failed to get schema from controller: %s", err)
	}

	// Sync module schema changes from the controller into the schema event source.
	return func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
		switch msg.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
			sch, err := schema.ModuleFromProto(msg.Schema)
			if err != nil {
				return err
			}
			e.controllerSchema.Store(sch.Name, sch)
			e.schemaChanges.Publish(schemaChange{ChangeType: msg.ChangeType, Module: sch})

		case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
			e.controllerSchema.Delete(msg.ModuleName)
			e.schemaChanges.Publish(schemaChange{ChangeType: msg.ChangeType, Module: nil})
		}
		return nil
	}
}

// Close stops the Engine's schema sync.
func (e *Engine) Close() error {
	e.cancel()
	return nil
}

// Graph returns the dependency graph for the given modules.
//
// If no modules are provided, the entire graph is returned. An error is returned if
// any dependencies are missing.
func (e *Engine) Graph(projects ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}
	for _, name := range projects {
		if err := e.buildGraph(name, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Engine) buildGraph(name string, out map[string][]string) error {
	var deps []string
	if project, ok := e.projects[name]; ok {
		deps = project.Dependencies()
	} else if sch, ok := e.controllerSchema.Load(name); ok {
		deps = sch.Imports()
	} else {
		return fmt.Errorf("module %q not found", name)
	}
	out[name] = deps
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

// Deploy attempts to build and deploy all local modules.
func (e *Engine) Deploy(ctx context.Context, replicas int32, waitForDeployOnline bool) error {
	return e.buildAndDeploy(ctx, replicas, waitForDeployOnline)
}

// Dev builds and deploys all local modules and watches for changes, redeploying as necessary.
func (e *Engine) Dev(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	// Build and deploy all modules first.
	err := e.buildAndDeploy(ctx, 1, true)
	if err != nil {
		logger.Errorf(err, "initial deploy failed")
	}

	logger.Infof("All modules deployed, watching for changes...")

	// Then watch for changes and redeploy.
	return e.watchForModuleChanges(ctx, period)
}

func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

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

	schemaChanges := make(chan schemaChange, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer e.schemaChanges.Unsubscribe(schemaChanges)

	watchEvents := make(chan WatchEvent, 128)
	watch := Watch(ctx, period, e.dirs, e.externalDirs)
	watch.Subscribe(watchEvents)
	defer watch.Unsubscribe(watchEvents)
	defer watch.Close()

	// Watch for file and schema changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watchEvents:
			switch event := event.(type) {
			case WatchEventProjectAdded:
				if _, exists := e.projects[event.Project.Key()]; !exists {
					e.projects[event.Project.Key()] = event.Project
					err := e.buildAndDeploy(ctx, 1, true, event.Project.Key())
					if err != nil {
						logger.Errorf(err, "deploy %s failed", event.Project.Key())
					}
				}
			case WatchEventProjectRemoved:
				if _, ok := event.Project.(Module); ok {
					err := teminateModuleDeployment(ctx, e.client, event.Project.Key())
					if err != nil {
						logger.Errorf(err, "terminate %s failed", event.Project.Key())
					}
				}
				delete(e.projects, event.Project.Key())
			case WatchEventProjectChanged:
				err := e.buildAndDeploy(ctx, 1, true, event.Project.Key())
				if err != nil {
					if _, ok := event.Project.(Module); ok {
						logger.Errorf(err, "deploy %s failed", event.Project.Key())
					} else {
						logger.Errorf(err, "stub generation for %s failed", event.Project.Key())
					}
				}
			}
		case change := <-schemaChanges:
			if change.ChangeType != ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED {
				continue
			}

			hash, err := computeModuleHash(change.Module)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", change.Name)
				continue
			}

			if bytes.Equal(hash, moduleHashes[change.Name]) {
				logger.Tracef("schema for %s has not changed", change.Name)
				continue
			}

			moduleHashes[change.Name] = hash

			dependantModules := e.getDependentModules(change.Name)
			if len(dependantModules) > 0 {
				//TODO: inaccurate log message for ext libs
				logger.Infof("%s's schema changed; redeploying %s", change.Name, strings.Join(dependantModules, ", "))
				err = e.buildAndDeploy(ctx, 1, true, dependantModules...)
				if err != nil {
					logger.Errorf(err, "deploy %s failed", change.Name)
				}
			}
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

func (e *Engine) getDependentModules(projectName string) []string {
	dependentProjects := map[string]bool{}
	for key, project := range e.projects {
		for _, dep := range project.Dependencies() {
			if dep == projectName {
				dependentProjects[key] = true
			}
		}
	}
	return maps.Keys(dependentProjects)
}

func (e *Engine) buildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, projects ...string) error {
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}

	deployQueue := make(chan Project, len(projects))
	wg, ctx := errgroup.WithContext(ctx)

	// Build all projects and enqueue the modules for deployment.
	wg.Go(func() error {
		defer close(deployQueue)

		//TODO: don't always include external libs
		return e.buildWithCallback(ctx, func(ctx context.Context, project Project) error {
			if _, ok := project.(Module); ok {
				select {
				case deployQueue <- project:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}, projects...)
	})

	// Process deployment queue.
	for i := 0; i < len(projects); i++ {
		wg.Go(func() error {
			for {
				select {
				case project, ok := <-deployQueue:
					if !ok {
						return nil
					}
					if module, ok := project.(Module); ok {
						if err := Deploy(ctx, module, replicas, waitForDeployOnline, e.client); err != nil {
							return err
						}
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	return wg.Wait()
}

type buildCallback func(ctx context.Context, project Project) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, projects ...string) error {
	mustBuild := map[string]bool{}
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}
	for _, name := range projects {
		project, ok := e.projects[name]
		if !ok {
			return fmt.Errorf("project %q not found", project)
		}
		// Update dependencies before building.
		var err error
		project, err = UpdateDependencies(ctx, project)
		if err != nil {
			return err
		}
		e.projects[name] = project
		mustBuild[name] = true
	}
	graph, err := e.Graph(projects...)
	if err != nil {
		return err
	}
	built := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	topology := TopologicalSort(graph)
	for _, group := range topology {
		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))
		wg, ctx := errgroup.WithContext(ctx)
		wg.SetLimit(e.parallelism)
		for _, key := range group {
			wg.Go(func() error {
				if !mustBuild[key] {
					return e.mustSchema(ctx, key, built, schemas)
				}

				fmt.Printf("building/generating %s\n", key)
				switch e.projects[key].(type) {
				case Module:
					err := e.build(ctx, key, built, schemas)
					if err == nil && callback != nil {
						return callback(ctx, e.projects[key])
					}
					return err
				case ExternalLibrary:
					if err := e.generateStubsForExternalLib(ctx, key, built); err != nil {
						return fmt.Errorf("could not build stubs for external library %s: %w", key, err)
					}
					return nil
				default:
					panic("unknown project type")
				}
			})
		}
		err := wg.Wait()
		if err != nil {
			return err
		}
		// Now this group is built, collect all the schemas.
		close(schemas)
		for sch := range schemas {
			built[sch.Name] = sch
		}
	}

	return nil
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, name string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load(name); ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, name, built, schemas)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, key string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	project, ok := e.projects[key]
	if !ok {
		return fmt.Errorf("project %q not found", key)
	}
	module, ok := project.(Module)
	if !ok {
		return fmt.Errorf("can not build project %q as it is not a module", key)
	}

	combined := map[string]*schema.Module{}
	if err := e.gatherSchemas(built, project, combined); err != nil {
		return err
	}
	sch := &schema.Schema{Modules: maps.Values(combined)}

	err := Build(ctx, sch, module)
	if err != nil {

		return err
	}
	moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(module.Dir(), module.DeployDir, module.Schema))
	if err != nil {
		return fmt.Errorf("load schema %s: %w", key, err)
	}
	schemas <- moduleSchema
	return nil
}

func (e *Engine) generateStubsForExternalLib(ctx context.Context, key string, built map[string]*schema.Module) error {
	//TODO: support kotlin
	project, ok := e.projects[key]
	if !ok {
		return fmt.Errorf("project %q not found", key)
	}
	lib, ok := project.(ExternalLibrary)
	if !ok {
		return fmt.Errorf("project %q is not a library", key)
	}

	logger := log.FromContext(ctx)

	goModFile, replacements, err := compile.GoModFileWithReplacements(filepath.Join(lib.Dir(), "go.mod"))
	if err != nil {
		return fmt.Errorf("failed to propogate replacements %s: %w", lib.Dir(), err)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	combined := map[string]*schema.Module{}
	if err := e.gatherSchemas(built, project, combined); err != nil {
		return err
	}
	sch := &schema.Schema{Modules: maps.Values(combined)}

	err = compile.GenerateExternalModules(compile.ExternalModuleContext{
		ModuleDir:    lib.Dir(),
		GoVersion:    goModFile.Go.Version,
		FTLVersion:   ftlVersion,
		Schema:       sch,
		Replacements: replacements,
	})
	if err != nil {
		return fmt.Errorf("failed to generate external modules: %w", err)
	}
	logger.Infof("Generated stubs %v for %s", lib.Dependencies(), lib.Dir())
	return nil
}

// Construct a combined schema for a module and its transitive dependencies.
func (e *Engine) gatherSchemas(
	moduleSchemas map[string]*schema.Module,
	project Project,
	out map[string]*schema.Module,
) error {
	latestModule, ok := e.projects[project.Key()]
	if !ok {
		latestModule = project
	}
	for _, dep := range latestModule.Dependencies() {
		out[dep] = moduleSchemas[dep]
		if dep != "builtin" {
			depModule, ok := e.projects[dep]
			if !ok {
				return fmt.Errorf("dependency %q not found", dep)
			}
			e.gatherSchemas(moduleSchemas, depModule, out)
		}
	}
	return nil
}
