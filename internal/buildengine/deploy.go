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
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type deploymentArtefact struct {
	*ftlv1.DeploymentArtefact
	localPath string
}

type AdminClient interface {
	ApplyChangeset(ctx context.Context, req *connect.Request[ftlv1.ApplyChangesetRequest]) (*connect.Response[ftlv1.ApplyChangesetResponse], error)
	ClusterInfo(ctx context.Context, req *connect.Request[ftlv1.ClusterInfoRequest]) (*connect.Response[ftlv1.ClusterInfoResponse], error)
	GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error)
	UploadArtefact(ctx context.Context) *connect.ClientStreamForClient[ftlv1.UploadArtefactRequest, ftlv1.UploadArtefactResponse]
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func Deploy(ctx context.Context, projectConfig projectconfig.Config, modules []Module, replicas int32, waitForDeployOnline bool, adminClient AdminClient) (err error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Deploying %v", strings.Join(slices.Map(modules, func(m Module) string { return m.Config.Module }), ", "))
	defer func() {
		if err != nil {
			logger.Errorf(err, "Failed to deploy %s", strings.Join(slices.Map(modules, func(m Module) string { return m.Config.Module }), ", "))
		}
	}()
	uploadGroup := errgroup.Group{}
	moduleSchemas := make(chan *schemapb.Module, len(modules))
	for _, module := range modules {
		uploadGroup.Go(func() error {
			sch, err := uploadArtefacts(ctx, projectConfig, module, adminClient)
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

	ctx, closeStream := context.WithCancelCause(ctx)
	defer closeStream(fmt.Errorf("function is complete: %w", context.Canceled))

	_, err = adminClient.ApplyChangeset(ctx, connect.NewRequest(&ftlv1.ApplyChangesetRequest{
		Modules: collectedSchemas,
	}))
	if err != nil {
		return fmt.Errorf("failed to deploy changeset: %w", err)
	}
	return nil
}

func uploadArtefacts(ctx context.Context, projectConfig projectconfig.Config, module Module, client AdminClient) (*schemapb.Module, error) {
	logger := log.FromContext(ctx).Module(module.Config.Module).Scope("deploy")
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Deploying module")

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
	return moduleSchema, nil
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

func terminateModuleDeployment(ctx context.Context, events *schemaeventsource.EventSource, client AdminClient, module string) error {
	logger := log.FromContext(ctx).Module(module).Scope("terminate")

	mod, ok := events.CanonicalView().Module(module).Get()

	if !ok {
		return fmt.Errorf("deployment for module %s not found", module)
	}
	key := mod.Runtime.Deployment.DeploymentKey

	logger.Infof("Terminating deployment %s", key)
	_, err := client.ApplyChangeset(ctx, connect.NewRequest(&ftlv1.ApplyChangesetRequest{
		ToRemove: []string{key.String()},
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
