package pythonplugin

import (
	"context"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/python-runtime/compile"
)

// buildContext contains contextual information needed to build.
type buildContext struct {
	ID           string
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, errors.Wrap(err, "could not parse schema from proto")
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	return buildContext{
		ID:           proto.Id,
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
	}, nil
}

type Service struct{}

var _ langconnect.LanguageServiceHandler = &Service{}

func New() *Service {
	return &Service{}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	return connect.NewResponse(&langpb.GetDependenciesResponse{}), nil
}

func (s *Service) Build(ctx context.Context, req *connect.Request[langpb.BuildRequest], stream *connect.ServerStream[langpb.BuildResponse]) error {
	logger := log.FromContext(ctx)
	logger.Infof("Do python build")

	projectConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return errors.WithStack(err)
	}

	_, _, err = compile.Build(ctx, projectConfig.Root(), req.Msg.StubsRoot, buildCtx.Config, nil, nil, false)
	logger.Errorf(err, "build failed")

	// TODO: Actually build the module instead of just returning an error.
	buildEvent := &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				ContextId:          req.Msg.BuildContext.Id,
				IsAutomaticRebuild: false,
				Errors: langpb.ErrorsToProto([]builderrors.Error{
					{
						Level: builderrors.ERROR,
						Msg:   "not implemented",
					},
				}),
				InvalidateDependencies: false,
			},
		},
	}

	if err := stream.Send(buildEvent); err != nil {
		return errors.Wrap(err, "could not send build event")
	}
	return nil
}

func (s *Service) BuildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	moduleSchema, err := schema.ValidatedModuleFromProto(req.Msg.Module)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	var nativeConfig optional.Option[moduleconfig.AbsModuleConfig]
	if req.Msg.NativeModuleConfig != nil {
		nativeConfig = optional.Some(langpb.ModuleConfigFromProto(req.Msg.NativeModuleConfig))
	}

	err = compile.GenerateStubs(ctx, req.Msg.Dir, moduleSchema, config, nativeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate stubs")
	}
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
}
