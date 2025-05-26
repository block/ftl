package buildengine

import (
	"context"
	"encoding/hex"
	"io"
	"maps"
	"os"
	"path/filepath"
	stdslices "slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type AdminClient interface {
	ApplyChangeset(ctx context.Context, req *connect.Request[adminpb.ApplyChangesetRequest]) (*connect.ServerStreamForClient[adminpb.ApplyChangesetResponse], error)
	ClusterInfo(ctx context.Context, req *connect.Request[adminpb.ClusterInfoRequest]) (*connect.Response[adminpb.ClusterInfoResponse], error)
	GetArtefactDiffs(ctx context.Context, req *connect.Request[adminpb.GetArtefactDiffsRequest]) (*connect.Response[adminpb.GetArtefactDiffsResponse], error)
	UploadArtefact(ctx context.Context) *connect.ClientStreamForClient[adminpb.UploadArtefactRequest, adminpb.UploadArtefactResponse]
	StreamLogs(ctx context.Context, req *connect.Request[adminpb.StreamLogsRequest]) (*connect.ServerStreamForClient[adminpb.StreamLogsResponse], error)
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type DependencyGrapher interface {
	Graph(moduleNames ...string) (map[string][]string, error)
}

type pendingModule struct {
	module Module

	deployPaths  []string
	tmpDeployDir string

	schema *schema.Module
}

func newPendingModule(module Module, tmpDeployDir string, deployPaths []string, schema *schema.Module) *pendingModule {
	return &pendingModule{
		module:       module,
		deployPaths:  deployPaths,
		tmpDeployDir: tmpDeployDir,
		// Schema is mutated by deploy coordinator
		schema: reflect.DeepCopy(schema),
	}
}

func (p *pendingModule) moduleName() string {
	return p.module.Config.Module
}

type pendingDeploy struct {
	modules  map[string]*pendingModule
	replicas optional.Option[int32]

	publishInSchema bool
	changeset       optional.Option[key.Changeset]
	err             chan error

	waitingForModules map[string]bool
	superseded        bool
	supercededModules []*pendingDeploy
}

type SchemaUpdatedEvent struct {
	schema *schema.Schema
	// marks which modules were changed together (ie. in the same changeset or queued together)
	updatedModules map[string]bool
}

// DeployCoordinator manages the deployment of modules through changesets. It ensures that changesets are deployed
// in the correct order and that changesets are not deployed if they are superseded by a newer changeset.
//
// The DeployCoordinator also maintains a schema based on active changesets and queued modules.
// This allows the build engine to build modules against where the schema is moving to. For example if module A is dependant on
// module B and module B builds with a breaking schema change deploy coordinator will put deployment of A into a pending state,
// but publish it as part of the its schema. This allows the build engine to react and build module A against the new schema for module B.
// The DeployCoordinator will then create a changeset of A and B together.
type DeployCoordinator struct {
	adminClient       AdminClient
	schemaSource      *schemaeventsource.EventSource
	dependencyGrapher DependencyGrapher

	// for publishing deploy events
	engineUpdates chan *buildenginepb.EngineEvent

	// deployment queue and state tracking
	deploymentQueue chan pendingDeploy

	SchemaUpdates chan SchemaUpdatedEvent

	logChanges bool // log changes from timeline

	projectConfig  projectconfig.Config
	externalRealms []*schema.Realm
}

func NewDeployCoordinator(
	ctx context.Context,
	adminClient AdminClient,
	schemaSource *schemaeventsource.EventSource,
	dependencyGrapher DependencyGrapher,
	engineUpdates chan *buildenginepb.EngineEvent,
	logChanges bool,
	projectConfig projectconfig.Config,
	externalRealms *xsync.MapOf[string, *schema.Realm],
) *DeployCoordinator {
	c := &DeployCoordinator{
		adminClient:       adminClient,
		schemaSource:      schemaSource,
		dependencyGrapher: dependencyGrapher,
		engineUpdates:     engineUpdates,
		deploymentQueue:   make(chan pendingDeploy, 128),
		SchemaUpdates:     make(chan SchemaUpdatedEvent, 128),
		logChanges:        logChanges,
		projectConfig:     projectConfig,
	}

	externalRealms.Range(func(key string, value *schema.Realm) bool {
		c.externalRealms = append(c.externalRealms, value)
		return true
	})

	// Start the deployment queue processor
	go c.processEvents(ctx)

	return c
}

func (c *DeployCoordinator) deploy(ctx context.Context, modules []*pendingModule, replicas optional.Option[int32]) error {
	logger := log.FromContext(ctx)
	for _, module := range modules {
		c.engineUpdates <- &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
				ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
					Module: module.module.Config.Module,
				},
			},
		}
	}
	pendingModules := make(map[string]*pendingModule, len(modules))
	for _, m := range modules {
		pendingModules[m.moduleName()] = m

		defer func() {
			if err := os.RemoveAll(m.tmpDeployDir); err != nil {
				logger.Errorf(err, "failed to remove tmp deploy dir %s", m.tmpDeployDir)
			}
		}()
	}

	errChan := make(chan error, 1)
	c.deploymentQueue <- pendingDeploy{
		modules:  pendingModules,
		replicas: replicas,
		err:      errChan}
	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err()) //nolint:wrapcheck
	case err := <-errChan:
		if err != nil {
			logger.Errorf(err, "Failed to deploy %s", strings.Join(slices.Map(modules, func(m *pendingModule) string { return m.moduleName() }), ", "))
		}
		return errors.WithStack(err)
	}
}

// processEvents handles the deployment queue and groups pending deployments into changesets
// It also maintains schema updates and constructing a target view of the schema for the build engine.
func (c *DeployCoordinator) processEvents(ctx context.Context) {
	logger := log.FromContext(ctx)
	events := c.schemaSource.Subscribe(ctx)
	var sch *schema.Schema
	if !c.schemaSource.Live() {
		logger.Debugf("Schema source is not live, skipping initial sync.")
		sch = &schema.Schema{
			Realms: []*schema.Realm{{
				Name:    c.projectConfig.Name,
				Modules: []*schema.Module{schema.Builtins()},
			}},
		}
	} else {
		c.schemaSource.WaitForInitialSync(ctx)

		// If there are no realms yet, initialise the internal.
		sch = c.schemaSource.CanonicalView()
		if len(sch.Realms) == 0 {
			sch.Realms = []*schema.Realm{{
				Name:    c.projectConfig.Name,
				Modules: []*schema.Module{schema.Builtins()},
			}}
		}
	}

	sch.UpdateRealms(c.externalRealms)
	c.SchemaUpdates <- SchemaUpdatedEvent{schema: sch}

	toDeploy := []*pendingDeploy{}
	deploying := []*pendingDeploy{}

	for {
		select {
		case <-ctx.Done():
			return

		case _d := <-c.deploymentQueue:
			deployment := &_d
			// Prepare by collecting module schemas and uploading artifacts
			err := prepareForDeploy(ctx, deployment.modules, c.adminClient)
			if err != nil {
				deployment.err <- err
				continue
			}

			// Check if there are older deployments that are superceded by this one or can be joined with this one
			for _, existing := range toDeploy {
				for _, mod := range existing.modules {
					if _, ok := deployment.modules[mod.moduleName()]; ok {
						existing.superseded = true
					}
				}
				for mod := range existing.waitingForModules {
					if _, ok := deployment.modules[mod]; ok {
						existing.superseded = true
					}
				}
				if existing.superseded {
					newDeployment, err := c.mergePendingDeployment(deployment, existing)
					if err != nil {
						// Fail new deployment attempt as it is incompatible with a dependency that is already in the queue
						deployment.err <- err
						continue
					}
					deployment = newDeployment
				}
			}
			toDeploy = slices.Filter(toDeploy, func(d *pendingDeploy) bool {
				return !d.superseded
			})

			// Check for modules that need to be rebuilt for this change to be valid
			// Try and deploy, unless there are conflicting changesets this will happen immediately
			graph, err := c.dependencyGrapher.Graph()
			if err != nil {
				log.FromContext(ctx).Errorf(err, "could not build graph to order deployment")
				continue
			}

			modulesToValidate := []string{}
			for module, dependencies := range graph {
				if _, ok := slices.Find(dependencies, func(s string) bool {
					_, ok := deployment.modules[s]
					return ok
				}); !ok {
					continue
				}

				modulesToValidate = append(modulesToValidate, module)
			}
			deployment.waitingForModules = c.invalidModulesForDeployment(c.schemaSource.CanonicalView(), deployment, modulesToValidate)
			if len(deployment.waitingForModules) > 0 {
				deployment.publishInSchema = true
			}

			if c.tryDeployFromQueue(ctx, deployment, toDeploy, graph) {
				if deployment.changeset.Ok() {
					deploying = append(deploying, deployment)
				}
			} else {
				// We could not deploy, add to the list of pending deployments
				toDeploy = append(toDeploy, deployment)
			}
			if deployment.publishInSchema {
				c.publishUpdatedSchema(ctx, stdslices.Collect(maps.Keys(deployment.modules)), toDeploy, deploying)
			}
		case notification := <-events:
			var key key.Changeset
			var updatedModules []string
			switch e := notification.(type) {
			case *schema.ChangesetCommittedNotification:
				key = e.Changeset.Key
				updatedModules = slices.Map(e.Changeset.InternalRealm().Modules, func(m *schema.Module) string { return m.Name })

				for _, m := range e.Changeset.InternalRealm().RemovingModules {
					if _, ok := slices.Find(updatedModules, func(s string) bool { return s == m.Name }); ok {
						continue
					}
					c.engineUpdates <- &buildenginepb.EngineEvent{
						Timestamp: timestamppb.Now(),
						Event: &buildenginepb.EngineEvent_ModuleRemoved{
							ModuleRemoved: &buildenginepb.ModuleRemoved{
								Module: m.Name,
							},
						},
					}
				}
			case *schema.ChangesetRollingBackNotification:
				key = e.Changeset.Key
				updatedModules = slices.Map(e.Changeset.InternalRealm().Modules, func(m *schema.Module) string { return m.Name })
			default:
				continue
			}

			tmp := []*pendingDeploy{}
			graph, err := c.dependencyGrapher.Graph()
			if err != nil {
				log.FromContext(ctx).Errorf(err, "could not build graph to order deployment")
				continue
			}
			deploying = slices.Filter(deploying, func(d *pendingDeploy) bool {
				if d.changeset != optional.Some(key) {
					return true
				}
				if d.publishInSchema {
					// already in published schema
					updatedModules = []string{}
				}
				return false
			})
			for _, mod := range toDeploy {
				if c.tryDeployFromQueue(ctx, mod, toDeploy, graph) {
					if mod.changeset.Ok() {
						deploying = append(deploying, mod)
					}
					continue
				}
				tmp = append(tmp, mod)
			}
			toDeploy = tmp
			c.publishUpdatedSchema(ctx, updatedModules, toDeploy, deploying)
		}
	}
}

func (c *DeployCoordinator) tryDeployFromQueue(ctx context.Context, deployment *pendingDeploy, toDeploy []*pendingDeploy, depGraph map[string][]string) bool {
	logger := log.FromContext(ctx)
	if len(deployment.waitingForModules) > 0 {
		return false
	}
	sets := c.schemaSource.ActiveChangesets()
	modules := map[string]bool{}
	depModules := map[string]bool{}
	for _, module := range deployment.modules {
		modules[module.moduleName()] = true
		for _, dep := range depGraph[module.moduleName()] {
			depModules[dep] = true
		}
	}
	for _, cs := range sets {
		if cs.State >= schema.ChangesetStateCommitted {
			continue
		}
		for _, mod := range cs.InternalRealm().Modules {
			if modules[mod.Name] || depModules[mod.Name] {
				return false
			}
		}
	}
	for _, queued := range toDeploy {
		for _, mod := range queued.modules {
			// We only check for dependencies here, as we already have de-duped modules
			// And we have not been removed from toDeploy at this point so we would find ourself
			if depModules[mod.moduleName()] {
				return false
			}
		}
	}

	// No conflicts, lets deploy
	moduleNames := slices.Sort(slices.Map(stdslices.Collect(maps.Values(deployment.modules)), func(m *pendingModule) string { return m.moduleName() }))
	logger.Debugf("Deploying %s", strings.Join(moduleNames, ","))
	for _, module := range deployment.modules {
		c.engineUpdates <- &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
				ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
					Module: module.moduleName(),
				},
			},
		}
		if repo, ok := deployment.replicas.Get(); ok {
			module.schema.ModRuntime().ModScaling().MinReplicas = repo
		}
	}

	keyChan := make(chan result.Result[key.Changeset], 1)
	go func() {
		err := deploy(ctx, c.projectConfig.Name, slices.Map(stdslices.Collect(maps.Values(deployment.modules)), func(m *pendingModule) *schema.Module { return m.schema }), c.adminClient, keyChan, c.externalRealms)
		if err != nil {
			// Handle deployment failure
			for _, module := range deployment.modules {
				c.engineUpdates <- &buildenginepb.EngineEvent{
					Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
						ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
							Module: module.moduleName(),
							Errors: &langpb.ErrorList{
								Errors: errorToLangError(err),
							},
						},
					},
				}
			}
		} else {
			// Handle deployment success
			for _, module := range deployment.modules {
				c.engineUpdates <- &buildenginepb.EngineEvent{
					Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
						ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
							Module: module.moduleName(),
						},
					},
				}
			}
		}
		deployment.err <- err
		for _, sup := range deployment.supercededModules {
			sup.err <- err
		}
	}()
	if key, ok := (<-keyChan).Get(); ok {
		logger := log.FromContext(ctx)
		deployment.changeset = optional.Some(key)
		logger.Debugf("Created changeset %s [%s]", key, strings.Join(moduleNames, ","))
		if c.logChanges {
			go c.runChangeLogger(ctx, key)
		}
	}
	return true
}

func (c *DeployCoordinator) runChangeLogger(ctx context.Context, key key.Changeset) {
	logger := log.FromContext(ctx)
	stream, err := c.adminClient.StreamLogs(ctx, connect.NewRequest(&adminpb.StreamLogsRequest{
		Query: &timelinepb.TimelineQuery{
			Limit: 100,
			Filters: []*timelinepb.TimelineQuery_Filter{
				{
					Filter: &timelinepb.TimelineQuery_Filter_Changesets{
						Changesets: &timelinepb.TimelineQuery_ChangesetFilter{
							Changesets: []string{key.String()},
						},
					},
				},
			},
		},
	}))
	if errors.Is(err, context.Canceled) {
		return
	}
	if err != nil {
		logger.Errorf(err, "failed to stream changeset logs")
		return
	}
	for stream.Receive() {
		for _, logpb := range stream.Msg().Logs {
			logger.Log(log.Entry{
				Attributes: logpb.Attributes,
				Level:      log.Level(logpb.LogLevel),
				Message:    logpb.Message,
			})
		}
	}
}

func (c *DeployCoordinator) mergePendingDeployment(d *pendingDeploy, old *pendingDeploy) (*pendingDeploy, error) {
	if d.replicas != old.replicas {
		return nil, errors.Errorf("could not deploy %v with pending deployment of %v: replicas were different %d != %d", maps.Keys(d.modules), maps.Keys(old.modules), d.replicas.Default(-1), old.replicas.Default(-1))
	}
	out := reflect.DeepCopy(d)
	addedModules := []string{}
	for _, module := range old.modules {
		if _, exists := d.modules[module.moduleName()]; exists {
			continue
		}
		out.modules[module.moduleName()] = old.modules[module.moduleName()]
		addedModules = append(addedModules, module.moduleName())
	}
	if len(addedModules) > 0 {
		if invalid := c.invalidModulesForDeployment(c.schemaSource.CanonicalView(), out, addedModules); len(invalid) > 0 {
			return nil, errors.Errorf("could not deploy %v with pending deployment of %v: modules were incompatible %v", maps.Keys(d.modules), maps.Keys(old.modules), maps.Keys(invalid))
		}
	}
	out.publishInSchema = out.publishInSchema || old.publishInSchema
	out.supercededModules = append([]*pendingDeploy{}, d.supercededModules...)
	out.supercededModules = append(out.supercededModules, old)
	out.supercededModules = append(out.supercededModules, old.supercededModules...)

	return out, nil
}

func (c *DeployCoordinator) invalidModulesForDeployment(originalSch *schema.Schema, deployment *pendingDeploy, modulesToCheck []string) map[string]bool {
	out := map[string]bool{}
	sch := &schema.Schema{}
	for _, realm := range originalSch.Realms {
		newRealm := &schema.Realm{
			Name:     realm.Name,
			External: realm.External,
		}
		sch.Realms = append(sch.Realms, newRealm)
		for _, module := range realm.Modules {
			if _, ok := deployment.modules[module.Name]; ok {
				continue
			}
			newRealm.Modules = append(newRealm.Modules, reflect.DeepCopy(module))
		}
	}
	for _, m := range deployment.modules {
		for _, realm := range sch.Realms {
			if realm.External {
				continue
			}
			realm.Modules = append(realm.Modules, m.schema)
			break
		}
	}
	for _, mod := range modulesToCheck {
		depSch, ok := slices.Find(sch.InternalModules(), func(m *schema.Module) bool {
			return m.Name == mod
		})
		if !ok {
			continue
		}
		if _, err := schema.ValidateModuleInSchema(sch, optional.Some(depSch)); err != nil {
			out[mod] = true
		}
	}
	return out
}

func (c *DeployCoordinator) publishUpdatedSchema(ctx context.Context, updatedModules []string, toDeploy, deploying []*pendingDeploy) {
	logger := log.FromContext(ctx)
	overridden := map[string]bool{}
	toRemove := map[string]bool{}
	realm := &schema.Realm{Name: c.projectConfig.Name}
	sch := &schema.Schema{
		Realms: append([]*schema.Realm{realm}, c.externalRealms...),
	}
	for _, d := range append(toDeploy, deploying...) {
		if !d.publishInSchema {
			continue
		}
		for _, mod := range d.modules {
			if _, ok := overridden[mod.moduleName()]; ok {
				continue
			}
			overridden[mod.moduleName()] = true
			realm.Modules = append(realm.Modules, mod.schema)
		}
		for mod := range d.waitingForModules {
			toRemove[mod] = true
		}
	}
	for _, mod := range c.schemaSource.CanonicalView().InternalModules() {
		if _, ok := overridden[mod.Name]; ok {
			continue
		}
		realm.Modules = append(realm.Modules, reflect.DeepCopy(mod))
	}
	// remove modules that we need to rebuild so that the schema is valid
	for {
		foundMoreToRemove := false
		for _, mod := range sch.InternalModules() {
			if toRemove[mod.Name] {
				continue
			}
			for _, im := range mod.Imports() {
				if _, ok := toRemove[im]; ok {
					toRemove[mod.Name] = true
					foundMoreToRemove = true
					break
				}
			}
		}
		if !foundMoreToRemove {
			break
		}
	}
	for _, realm := range sch.Realms {
		if realm.External {
			continue
		}
		realm.Modules = slices.Filter(realm.Modules, func(m *schema.Module) bool {
			return !toRemove[m.Name]
		})
		break
	}

	sch, err := sch.Validate()
	if err != nil {
		logger.Errorf(err, "Deploy coordinator could not publish invalid schema")
		return
	}
	updated := map[string]bool{}
	for _, m := range updatedModules {
		updated[m] = true
	}
	c.SchemaUpdates <- SchemaUpdatedEvent{
		schema:         sch,
		updatedModules: updated,
	}
}

func (c *DeployCoordinator) terminateModuleDeployment(ctx context.Context, module string) error {
	logger := log.FromContext(ctx).Module(module).Scope("terminate")

	mod, ok := c.schemaSource.CanonicalView().Module(c.projectConfig.Name, module).Get()

	if !ok {
		return errors.Errorf("deployment for module %s not found", module)
	}
	key := mod.Runtime.Deployment.DeploymentKey

	logger.Infof("Terminating deployment %s", key) //nolint:forbidigo
	stream, err := c.adminClient.ApplyChangeset(ctx, connect.NewRequest(&adminpb.ApplyChangesetRequest{
		RealmChanges: []*adminpb.RealmChange{{
			Name:     c.projectConfig.Name,
			ToRemove: []string{key.String()},
		}},
	}))
	if err != nil {
		return errors.Wrap(err, "failed to terminate deployment")
	}
	for stream.Receive() {
		// Not interested in progress
	}
	if err := stream.Err(); err != nil {
		return errors.Wrap(err, "failed to terminate deployment")
	}
	return nil
}

func prepareForDeploy(ctx context.Context, modules map[string]*pendingModule, adminClient AdminClient) (err error) {
	for _, module := range modules {
		if module.schema.Runtime == nil {
			module.schema.Runtime = &schema.ModuleRuntime{
				Base: schema.ModuleRuntimeBase{
					CreateTime: time.Now(),
				},
			}
		}
		module.schema.Runtime.Base.Language = module.module.Config.Language
	}

	uploadGroup := errgroup.Group{}
	for _, module := range modules {
		uploadGroup.Go(func() error {
			err := uploadArtefacts(ctx, module, adminClient)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		})
	}
	if err := uploadGroup.Wait(); err != nil {
		return errors.Wrap(err, "failed to upload artefacts")
	}
	return nil
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func deploy(ctx context.Context, realm string, modules []*schema.Module, adminClient AdminClient, receivedKey chan result.Result[key.Changeset], externalRealms []*schema.Realm) (err error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Deploying %v", strings.Join(slices.Map(modules, func(m *schema.Module) string { return m.Name }), ", "))
	changesetKey := optional.Option[key.Changeset]{}
	defer func() {
		if !changesetKey.Ok() {
			receivedKey <- result.Err[key.Changeset](err)
		}
	}()

	ctx, closeStream := context.WithCancelCause(ctx)
	defer closeStream(errors.Wrap(context.Canceled, "function is complete"))

	realmChanges := []*adminpb.RealmChange{{
		Name: realm,
		Modules: slices.Map(modules, func(m *schema.Module) *schemapb.Module {
			return m.ToProto()
		}),
	}}
	for _, r := range externalRealms {
		realmChanges = append(realmChanges, &adminpb.RealmChange{
			Name:     r.Name,
			External: true,
			Modules:  slices.Map(r.Modules, func(m *schema.Module) *schemapb.Module { return m.ToProto() }),
		})
	}

	stream, err := adminClient.ApplyChangeset(ctx, connect.NewRequest(&adminpb.ApplyChangesetRequest{
		RealmChanges: realmChanges,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to deploy changeset")
	}

	for stream.Receive() {
		if !changesetKey.Ok() {
			k, err := key.ParseChangesetKey(stream.Msg().Changeset.Key)
			if err != nil {
				return errors.Wrap(err, "failed to parse changeset key")
			}
			changesetKey = optional.Some(k)
			receivedKey <- result.Ok(k)
		}
	}

	if err := stream.Err(); err != nil {
		return errors.Wrap(err, "failed to deploy changeset")
	}
	return nil
}

type deploymentArtefact struct {
	*adminpb.DeploymentArtefact
	localPath string
}

func uploadArtefacts(ctx context.Context, module *pendingModule, client AdminClient) error {
	logger := log.FromContext(ctx).Module(module.moduleName()).Scope("deploy")
	ctx = log.ContextWithLogger(ctx, logger)

	filesByHash, err := hashFiles(module.tmpDeployDir, module.deployPaths)
	if err != nil {
		return errors.WithStack(err)
	}

	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&adminpb.GetArtefactDiffsRequest{ClientDigests: stdslices.Collect(maps.Keys(filesByHash))}))
	if err != nil {
		return errors.Wrap(err, "failed to get artefact diffs")
	}

	logger.Debugf("Uploading %d/%d files", len(gadResp.Msg.MissingDigests), len(module.deployPaths))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		if err := uploadDeploymentArtefact(ctx, client, file); err != nil {
			return errors.Wrap(err, "failed to upload deployment artefact")
		}
	}

	for _, artefact := range filesByHash {
		digest, err := sha256.ParseSHA256(hex.EncodeToString(artefact.Digest))
		if err != nil {
			return errors.Wrap(err, "failed to parse SHA256 digest")
		}
		module.schema.Metadata = append(module.schema.Metadata, &schema.MetadataArtefact{
			Path:       artefact.Path,
			Digest:     digest,
			Executable: artefact.Executable,
		})
	}
	return nil
}

func uploadDeploymentArtefact(ctx context.Context, client AdminClient, file deploymentArtefact) error {
	logger := log.FromContext(ctx).Scope("upload:" + hex.EncodeToString(file.Digest))
	f, err := os.Open(file.localPath)
	if err != nil {
		return errors.Wrap(err, "failed to read fil")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, "failed to get file info")
	}
	digest, err := sha256.ParseBytes(file.Digest)
	if err != nil {
		return errors.Wrap(err, "failed to parse SHA256 digest")
	}
	logger.Debugf("Uploading %s", relToCWD(file.localPath))
	stream := client.UploadArtefact(ctx)

	// 4KB chunks
	data := make([]byte, 4096)
	for {
		n, err := f.Read(data)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read fil")
		}
		err = stream.Send(&adminpb.UploadArtefactRequest{
			Chunk:  data[:n],
			Digest: digest[:],
			Size:   info.Size(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to upload artefact")
		}
	}
	_, err = stream.CloseAndReceive()
	if err != nil {
		return errors.Wrap(err, "failed to upload artefact")
	}
	logger.Debugf("Uploaded %s as %s:%s", relToCWD(file.localPath), digest, file.Path)
	return nil
}

func hashFiles(base string, files []string) (filesByHash map[string]deploymentArtefact, err error) {
	filesByHash = map[string]deploymentArtefact{}
	for _, file := range files {
		r, err := os.Open(file)
		if err != nil {
			return nil, errors.Wrapf(err, "could not open %q", file)
		}
		defer r.Close() //nolint:gosec
		hash, err := sha256.SumReader(r)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		info, err := r.Stat()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		isExecutable := info.Mode()&0111 != 0
		path, err := filepath.Rel(base, file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		filesByHash[hash.String()] = deploymentArtefact{
			DeploymentArtefact: &adminpb.DeploymentArtefact{
				Digest:     hash[:],
				Path:       path,
				Executable: isExecutable,
			},
			localPath: file,
		}
	}
	return filesByHash, nil
}

func relToCWD(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return rel
}
