package buildengine

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type AdminClient interface {
	ApplyChangeset(context.Context, *connect.Request[ftlv1.ApplyChangesetRequest]) (*connect.ServerStreamForClient[ftlv1.ApplyChangesetResponse], error)
	ClusterInfo(ctx context.Context, req *connect.Request[ftlv1.ClusterInfoRequest]) (*connect.Response[ftlv1.ClusterInfoResponse], error)
	GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error)
	UploadArtefact(ctx context.Context) *connect.ClientStreamForClient[ftlv1.UploadArtefactRequest, ftlv1.UploadArtefactResponse]
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type DependencyGrapher interface {
	Graph(...string) (map[string][]string, error)
}

type pendingModule struct {
	name   string
	module Module

	schemaPath string
	schema     *schema.Module
}

type pendingDeploy struct {
	modules  map[string]*pendingModule
	replicas int32

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
}

func NewDeployCoordinator(ctx context.Context, adminClient AdminClient, schemaSource *schemaeventsource.EventSource,
	dependencyGrapher DependencyGrapher, engineUpdates chan *buildenginepb.EngineEvent) *DeployCoordinator {
	c := &DeployCoordinator{
		adminClient:       adminClient,
		schemaSource:      schemaSource,
		dependencyGrapher: dependencyGrapher,
		engineUpdates:     engineUpdates,
		deploymentQueue:   make(chan pendingDeploy, 128),
		SchemaUpdates:     make(chan SchemaUpdatedEvent, 128),
	}
	// Start the deployment queue processor
	go c.processEvents(ctx)

	return c
}

func (c *DeployCoordinator) deploy(ctx context.Context, projConfig projectconfig.Config, modules []Module, replicas int32) error {
	for _, module := range modules {
		c.engineUpdates <- &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
				ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
					Module: module.Config.Module,
				},
			},
		}
	}
	pendingModules := make(map[string]*pendingModule, len(modules))
	for _, m := range modules {
		pendingModules[m.Config.Module] = &pendingModule{
			name:       m.Config.Module,
			module:     m,
			schemaPath: projConfig.SchemaPath(m.Config.Module),
		}
	}

	errChan := make(chan error, 1)
	c.deploymentQueue <- pendingDeploy{
		modules:  pendingModules,
		replicas: replicas,
		err:      errChan}
	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	case err := <-errChan:
		return err
	}
}

// processEvents handles the deployment queue and groups pending deployments into changesets
// It also maintains schema updates and constructing a target view of the schema for the build engine.
func (c *DeployCoordinator) processEvents(ctx context.Context) {
	logger := log.FromContext(ctx)
	events := c.schemaSource.Subscribe(ctx)
	if !c.schemaSource.Live() {
		logger.Debugf("Schema source is not live, skipping initial sync.")
		c.SchemaUpdates <- SchemaUpdatedEvent{
			schema: &schema.Schema{
				Modules: []*schema.Module{
					schema.Builtins(),
				},
			},
		}
	} else {
		c.schemaSource.WaitForInitialSync(ctx)
		c.SchemaUpdates <- SchemaUpdatedEvent{
			schema: c.schemaSource.CanonicalView(),
		}
	}

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
					if _, ok := deployment.modules[mod.name]; ok {
						existing.superseded = true
					}
				}
				for mod := range existing.waitingForModules {
					if _, ok := deployment.modules[mod]; ok {
						existing.superseded = true
					}
				}
				if existing.superseded {
					if deployment, err = c.mergePendingDeployment(deployment, existing); err != nil {
						// Fail new deployment attempt as it is incompitable with a dependency that is already in the queue
						deployment.err <- err
						continue
					}
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
				c.publishUpdatedSchema(ctx, maps.Keys(deployment.modules), toDeploy, deploying)
			}
		case notification := <-events:
			var key key.Changeset
			var updatedModules []string
			switch e := notification.(type) {
			case *schema.ChangesetCommittedNotification:
				key = e.Changeset.Key
				updatedModules = slices.Map(e.Changeset.Modules, func(m *schema.Module) string { return m.Name })

				for _, m := range e.Changeset.RemovingModules {
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
				updatedModules = slices.Map(e.Changeset.Modules, func(m *schema.Module) string { return m.Name })
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
	if len(deployment.waitingForModules) > 0 {
		return false
	}
	sets := c.schemaSource.ActiveChangesets()
	modules := map[string]bool{}
	depModules := map[string]bool{}
	for _, module := range deployment.modules {
		modules[module.name] = true
		for _, dep := range depGraph[module.name] {
			depModules[dep] = true
		}
	}
	for _, cs := range sets {
		if cs.State >= schema.ChangesetStateCommitted {
			continue
		}
		for _, mod := range cs.Modules {
			if modules[mod.Name] || depModules[mod.Name] {
				return false
			}
		}
	}
	for _, queued := range toDeploy {
		for _, mod := range queued.modules {
			// We only check for dependencies here, as we already have de-duped modules
			// And we have not been removed from toDeploy at this point so we would find ourself
			if depModules[mod.name] {
				return false
			}
		}
	}

	// No conflicts, lets deploy

	// Deploy all collected modules
	for _, module := range deployment.modules {
		c.engineUpdates <- &buildenginepb.EngineEvent{
			Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
				ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
					Module: module.name,
				},
			},
		}
	}

	keyChan := make(chan result.Result[key.Changeset], 1)
	go func() {
		err := deploy(ctx, slices.Map(maps.Values(deployment.modules), func(m *pendingModule) *schema.Module { return m.schema }), deployment.replicas, c.adminClient, keyChan)
		if err != nil {
			// Handle deployment failure
			for _, module := range deployment.modules {
				c.engineUpdates <- &buildenginepb.EngineEvent{
					Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
						ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
							Module: module.name,
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
							Module: module.name,
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
		deployment.changeset = optional.Some(key)
	}
	return true
}

func (c *DeployCoordinator) mergePendingDeployment(d *pendingDeploy, old *pendingDeploy) (*pendingDeploy, error) {
	if d.replicas != old.replicas {
		return nil, fmt.Errorf("could not deploy %v with pending deployment of %v: replicas were different %d != %d", maps.Keys(d.modules), maps.Keys(old.modules), d.replicas, old.replicas)
	}
	out := reflect.DeepCopy(d)
	addedModules := []string{}
	for _, module := range old.modules {
		if _, exists := d.modules[module.name]; exists {
			continue
		}
		out.modules[module.name] = old.modules[module.name]
		addedModules = append(addedModules, module.name)
	}
	if len(addedModules) > 0 {
		if invalid := c.invalidModulesForDeployment(c.schemaSource.CanonicalView(), out, addedModules); len(invalid) > 0 {
			return nil, fmt.Errorf("could not deploy %v with pending deployment of %v: modules were incompatible %v", maps.Keys(d.modules), maps.Keys(old.modules), maps.Keys(invalid))
		}
	}
	out.publishInSchema = out.publishInSchema || old.publishInSchema
	out.supercededModules = append(d.supercededModules, old)
	out.supercededModules = append(d.supercededModules, old.supercededModules...)
	return out, nil
}

func (c *DeployCoordinator) invalidModulesForDeployment(originalSch *schema.Schema, deployment *pendingDeploy, modulesToCheck []string) map[string]bool {
	out := map[string]bool{}
	sch := &schema.Schema{}
	for _, moduleSch := range originalSch.Modules {
		if _, ok := deployment.modules[moduleSch.Name]; ok {
			continue
		}
		sch.Modules = append(sch.Modules, reflect.DeepCopy(moduleSch))
	}
	for _, m := range deployment.modules {
		sch.Modules = append(sch.Modules, m.schema)
	}
	for _, mod := range modulesToCheck {
		depSch, ok := slices.Find(sch.Modules, func(m *schema.Module) bool {
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
	sch := &schema.Schema{}
	for _, d := range append(toDeploy, deploying...) {
		if !d.publishInSchema {
			continue
		}
		for _, mod := range d.modules {
			if _, ok := overridden[mod.name]; ok {
				continue
			}
			overridden[mod.name] = true
			sch.Modules = append(sch.Modules, mod.schema)
		}
		for mod := range d.waitingForModules {
			toRemove[mod] = true
		}
	}
	for _, mod := range c.schemaSource.CanonicalView().Modules {
		if _, ok := overridden[mod.Name]; ok {
			continue
		}
		sch.Modules = append(sch.Modules, reflect.DeepCopy(mod))
	}
	// remove modules that we need to rebuild so that the schema is valid
	for {
		foundMoreToRemove := false
		for _, mod := range sch.Modules {
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
	sch.Modules = slices.Filter(sch.Modules, func(m *schema.Module) bool {
		return !toRemove[m.Name]
	})

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

func (d *DeployCoordinator) terminateModuleDeployment(ctx context.Context, module string) error {
	logger := log.FromContext(ctx).Module(module).Scope("terminate")

	mod, ok := d.schemaSource.CanonicalView().Module(module).Get()

	if !ok {
		return fmt.Errorf("deployment for module %s not found", module)
	}
	key := mod.Runtime.Deployment.DeploymentKey

	logger.Infof("Terminating deployment %s", key)
	stream, err := d.adminClient.ApplyChangeset(ctx, connect.NewRequest(&ftlv1.ApplyChangesetRequest{
		ToRemove: []string{key.String()},
	}))
	if err != nil {
		return fmt.Errorf("failed to terminate deployment: %w", err)
	}
	for stream.Receive() {
		// Not interested in progress
	}
	if err := stream.Err(); err != nil {
		return fmt.Errorf("failed to terminate deployment: %w", err)
	}
	return nil
}

func prepareForDeploy(ctx context.Context, modules map[string]*pendingModule, adminClient AdminClient) (err error) {
	uploadGroup := errgroup.Group{}
	for _, module := range modules {
		uploadGroup.Go(func() error {
			sch, err := uploadArtefacts(ctx, module, adminClient)
			if err != nil {
				return err
			}
			module.schema = sch
			return nil
		})
	}
	if err := uploadGroup.Wait(); err != nil {
		return fmt.Errorf("failed to upload artefacts: %w", err)
	}
	return nil
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func deploy(ctx context.Context, modules []*schema.Module, replicas int32, adminClient AdminClient, receivedKey chan result.Result[key.Changeset]) (err error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Deploying %v", strings.Join(slices.Map(modules, func(m *schema.Module) string { return m.Name }), ", "))
	changesetKey := optional.Option[key.Changeset]{}
	defer func() {
		if !changesetKey.Ok() {
			receivedKey <- result.Err[key.Changeset](err)
		}
		if err != nil {
			logger.Errorf(err, "Failed to deploy %s", strings.Join(slices.Map(modules, func(m *schema.Module) string { return m.Name }), ", "))
		}
	}()

	ctx, closeStream := context.WithCancelCause(ctx)
	defer closeStream(fmt.Errorf("function is complete: %w", context.Canceled))

	stream, err := adminClient.ApplyChangeset(ctx, connect.NewRequest(&ftlv1.ApplyChangesetRequest{
		Modules: slices.Map(modules, func(m *schema.Module) *schemapb.Module {
			return m.ToProto()
		}),
	}))
	if err != nil {
		return fmt.Errorf("failed to deploy changeset: %w", err)
	}

	for stream.Receive() {
		if !changesetKey.Ok() {
			k, err := key.ParseChangesetKey(stream.Msg().Changeset.Key)
			if err != nil {
				return fmt.Errorf("failed to parse changeset key: %w", err)
			}
			changesetKey = optional.Some(k)
			receivedKey <- result.Ok(k)
		}
	}

	return stream.Err()
}

type deploymentArtefact struct {
	*ftlv1.DeploymentArtefact
	localPath string
}

func uploadArtefacts(ctx context.Context, module *pendingModule, client AdminClient) (*schema.Module, error) {
	logger := log.FromContext(ctx).Module(module.name).Scope("deploy")
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Deploying module")

	moduleConfig := module.module.Config.Abs()
	files, err := findFilesToDeploy(moduleConfig, module.module.Deploy)
	if err != nil {
		logger.Errorf(err, "failed to find files in %s", moduleConfig)
		return nil, err
	}

	filesByHash, err := hashFiles(moduleConfig.DeployDir, files)
	if err != nil {
		return nil, err
	}

	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
	if err != nil {
		return nil, fmt.Errorf("failed to get artefact diffs: %w", err)
	}

	moduleSchema, err := loadProtoSchema(moduleConfig, module.schemaPath)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Uploading %d/%d files", len(gadResp.Msg.MissingDigests), len(files))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		if err := uploadDeploymentArtefact(ctx, client, file); err != nil {
			return nil, fmt.Errorf("failed to upload deployment artefact: %w", err)
		}
	}

	for _, artefact := range filesByHash {
		moduleSchema.Metadata = append(moduleSchema.Metadata, &schemapb.Metadata{
			Value: &schemapb.Metadata_Artefact{
				Artefact: &schemapb.MetadataArtefact{
					Path:       artefact.Path,
					Digest:     hex.EncodeToString(artefact.Digest),
					Executable: artefact.Executable,
				},
			},
		})
	}
	parsedSchema, err := schema.ModuleFromProto(moduleSchema)
	if err != nil {
		return nil, fmt.Errorf("could not parse schema to upload: %w", err)
	}
	return parsedSchema, nil
}

func uploadDeploymentArtefact(ctx context.Context, client AdminClient, file deploymentArtefact) error {
	logger := log.FromContext(ctx).Scope("upload:" + hex.EncodeToString(file.Digest))
	f, err := os.Open(file.localPath)
	if err != nil {
		return fmt.Errorf("failed to read file %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	digest, err := sha256.ParseBytes(file.Digest)
	if err != nil {
		return fmt.Errorf("failed to parse SHA256 digest: %w", err)
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
			return fmt.Errorf("failed to read file %w", err)
		}
		err = stream.Send(&ftlv1.UploadArtefactRequest{
			Chunk:  data[:n],
			Digest: digest[:],
			Size:   info.Size(),
		})
		if err != nil {
			return fmt.Errorf("failed to upload artefact: %w", err)
		}
	}
	_, err = stream.CloseAndReceive()
	if err != nil {
		return fmt.Errorf("failed to upload artefact: %w", err)
	}
	logger.Debugf("Uploaded %s as %s:%s", relToCWD(file.localPath), digest, file.Path)
	return nil
}

func loadProtoSchema(config moduleconfig.AbsModuleConfig, schPath string) (*schemapb.Module, error) {
	content, err := os.ReadFile(schPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf schema from %q: %w", schPath, err)
	}
	module := &schemapb.Module{}
	err = proto.Unmarshal(content, module)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf schema from %q: %w", schPath, err)
	}
	runtime := module.Runtime
	if runtime == nil {
		runtime = &schemapb.ModuleRuntime{}
		module.Runtime = runtime
	}
	module.Runtime = runtime
	if runtime.Base == nil {
		runtime.Base = &schemapb.ModuleRuntimeBase{}
	}
	if runtime.Base.CreateTime == nil {
		runtime.Base.CreateTime = timestamppb.Now()
	}
	runtime.Base.Language = config.Language
	return module, nil
}

// findFilesToDeploy returns a list of files to deploy for the given module.
func findFilesToDeploy(config moduleconfig.AbsModuleConfig, deploy []string) ([]string, error) {
	var out []string
	for _, f := range deploy {
		file := filepath.Clean(filepath.Join(config.DeployDir, f))
		if !strings.HasPrefix(file, config.DeployDir) {
			return nil, fmt.Errorf("deploy path %q is not beneath deploy directory %q", file, config.DeployDir)
		}
		info, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			dirFiles, err := findFilesInDir(file)
			if err != nil {
				return nil, err
			}
			out = append(out, dirFiles...)
		} else {
			out = append(out, file)
		}
	}
	return out, nil
}

func findFilesInDir(dir string) ([]string, error) {
	var out []string
	return out, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	})
}

func hashFiles(base string, files []string) (filesByHash map[string]deploymentArtefact, err error) {
	filesByHash = map[string]deploymentArtefact{}
	for _, file := range files {
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer r.Close() //nolint:gosec
		hash, err := sha256.SumReader(r)
		if err != nil {
			return nil, err
		}
		info, err := r.Stat()
		if err != nil {
			return nil, err
		}
		isExecutable := info.Mode()&0111 != 0
		path, err := filepath.Rel(base, file)
		if err != nil {
			return nil, err
		}
		filesByHash[hash.String()] = deploymentArtefact{
			DeploymentArtefact: &ftlv1.DeploymentArtefact{
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
