package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

type deploymentArtefact struct {
	*ftlv1.DeploymentArtefact
	localPath string
}

type DeployClient interface {
	GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error)
	UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error)
	Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error)
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type SchemaServiceClient interface {
	CreateChangeset(ctx context.Context, req *connect.Request[ftlv1.CreateChangesetRequest]) (*connect.Response[ftlv1.CreateChangesetResponse], error)
	PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
	UpdateDeploymentRuntime(ctx context.Context, req *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error)
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func Deploy(ctx context.Context, projectConfig projectconfig.Config, modules []Module, replicas int32, waitForDeployOnline bool, deployClient DeployClient, schemaserviceClient SchemaServiceClient) error {
	logger := log.FromContext(ctx)
	uploadGroup := errgroup.Group{}
	moduleSchemas := make(chan *schemapb.Module, len(modules))
	for _, module := range modules {
		uploadGroup.Go(func() error {
			sch, err := uploadArtefacts(ctx, projectConfig, module, deployClient)
			if err != nil {
				return err
			}
			moduleSchemas <- sch
			return nil
		})
	}
	if err := uploadGroup.Wait(); err != nil {
		return fmt.Errorf("failed to upload artefacts: %w", err)
	}
	close(moduleSchemas)
	collectedSchemas := []*schemapb.Module{}
	for {
		sch, ok := <-moduleSchemas
		if !ok {
			break
		}
		collectedSchemas = append(collectedSchemas, sch)
	}
	resp, err := schemaserviceClient.CreateChangeset(ctx, connect.NewRequest(&ftlv1.CreateChangesetRequest{
		Modules: collectedSchemas,
	}))
	logger.Debugf("Created changeset with %d modules", len(collectedSchemas))
	if err != nil {
		return fmt.Errorf("failed to create changeset: %w", err)
	}
	key := resp.Msg.Changeset

	ctx, closeStream := context.WithCancelCause(ctx)
	stream, err := schemaserviceClient.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to pull schema: %w", err)
	}
	defer closeStream(fmt.Errorf("function is complete"))
	for {
		if !stream.Receive() {
			return fmt.Errorf("failed to pull schema: %w", stream.Err())
		}
		msg := stream.Msg()
		switch msg := msg.Event.(type) {
		case *ftlv1.PullSchemaResponse_ChangesetCommitted_:
			if msg.ChangesetCommitted.Changeset.Key != key {
				logger.Warnf("Expecting changeset %s to complete but got commit for %s", key, msg.ChangesetCommitted.Changeset.Key)
				continue
			}
			logger.Infof("Changeset %s deployed and ready", key)
			return nil
		case *ftlv1.PullSchemaResponse_ChangesetFailed_:
			if msg.ChangesetFailed.Key != key {
				logger.Warnf("Expecting changeset %s to complete but got failure for %s", key, msg.ChangesetFailed.Key)
				continue
			}
			return fmt.Errorf("changeset %s failed: %s", key, msg.ChangesetFailed.Error)

		case *ftlv1.PullSchemaResponse_ChangesetCreated_:
			// TODO: handle this case where stream starts after changeset ends. Or reconnects when changeset has ended
		case *ftlv1.PullSchemaResponse_DeploymentCreated_,
			*ftlv1.PullSchemaResponse_DeploymentUpdated_,
			*ftlv1.PullSchemaResponse_DeploymentRemoved_:
		}
	}
}

func uploadArtefacts(ctx context.Context, projectConfig projectconfig.Config, module Module, client DeployClient) (*schemapb.Module, error) {
	logger := log.FromContext(ctx).Module(module.Config.Module).Scope("deploy")
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Deploying module")

	moduleConfig := module.Config.Abs()
	files, err := FindFilesToDeploy(moduleConfig, module.Deploy)
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

	moduleSchema, err := loadProtoSchema(projectConfig, moduleConfig)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Uploading %d/%d files", len(gadResp.Msg.MissingDigests), len(files))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		content, err := os.ReadFile(file.localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %w", err)
		}
		logger.Debugf("Uploading %s", relToCWD(file.localPath))
		resp, err := client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{
			Content: content,
		}))
		if err != nil {
			return nil, fmt.Errorf("failed to upload artefact: %w", err)
		}
		logger.Debugf("Uploaded %s as %s:%s", relToCWD(file.localPath), sha256.FromBytes(resp.Msg.Digest), file.Path)
	}

	for _, artefact := range filesByHash {
		moduleSchema.Metadata = append(moduleSchema.Metadata, &schemapb.Metadata{
			Value: &schemapb.Metadata_Artefact{
				Artefact: &schemapb.MetadataArtefact{
					Path:       artefact.Path,
					Digest:     artefact.Digest,
					Executable: artefact.Executable,
				},
			},
		})
	}
	return moduleSchema, nil
}

func terminateModuleDeployment(ctx context.Context, client DeployClient, schemaClient SchemaServiceClient, module string) error {
	logger := log.FromContext(ctx).Module(module).Scope("terminate")

	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
	if err != nil {
		return err
	}

	var key string
	for _, deployment := range status.Msg.Deployments {
		if deployment.Name == module {
			key = deployment.Key
			continue
		}
	}

	if key == "" {
		return fmt.Errorf("deployment for module %s not found: %v", module, status.Msg.Deployments)
	}

	logger.Infof("Terminating deployment %s", key)
	_, err = schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
		Event: &schemapb.ModuleRuntimeEvent{
			Key:     key,
			Scaling: &schemapb.ModuleRuntimeScaling{MinReplicas: 0},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to kill deployment: %w", err)
	}
	return nil
}

func loadProtoSchema(projectConfig projectconfig.Config, config moduleconfig.AbsModuleConfig) (*schemapb.Module, error) {
	schPath := projectConfig.SchemaPath(config.Module)
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

// FindFilesToDeploy returns a list of files to deploy for the given module.
func FindFilesToDeploy(config moduleconfig.AbsModuleConfig, deploy []string) ([]string, error) {
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
				Digest:     hash.String(),
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

func checkReadiness(ctx context.Context, client DeployClient, deploymentKey string, replicas int32, schema *schemapb.Module) error {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	hasVerbs := false
	for _, dec := range schema.Decls {
		if dec.GetVerb() != nil {
			hasVerbs = true
			break
		}
	}
	for range channels.IterContext(ctx, ticker.C) {
		status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		for _, deployment := range status.Msg.Deployments {
			if deployment.Key == deploymentKey {
				if deployment.Replicas >= replicas {
					if hasVerbs {
						// Also verify the routing table is ready
						for _, route := range status.Msg.Routes {
							if route.Deployment == deploymentKey {
								return nil
							}
						}

					} else {
						return nil
					}
				}
			}
		}
	}
	return nil
}
