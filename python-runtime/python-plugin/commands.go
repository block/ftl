package pythonplugin

import (
	"context"
	"path/filepath"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/block/scaffolder"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/log"
	pythonruntime "github.com/block/ftl/python-runtime"
)

type CmdService struct{}

var _ languagepbconnect.LanguageCommandServiceHandler = CmdService{}

func (CmdService) GetNewModuleFlags(ctx context.Context, req *connect.Request[langpb.GetNewModuleFlagsRequest]) (*connect.Response[langpb.GetNewModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetNewModuleFlagsResponse{}), nil
}

type scaffoldingContext struct {
	Name string
}

func (CmdService) NewModule(ctx context.Context, req *connect.Request[langpb.NewModuleRequest]) (*connect.Response[langpb.NewModuleResponse], error) {
	logger := log.FromContext(ctx)
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	opts := []scaffolder.Option{}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := scaffoldingContext{
		Name: req.Msg.Name,
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(pythonruntime.Files(), parentPath, sctx, opts...); err != nil {
		return nil, errors.Wrap(err, "failed to scaffold")
	}
	return connect.NewResponse(&langpb.NewModuleResponse{}), nil
}

func (CmdService) GetModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.GetModuleConfigDefaultsRequest]) (*connect.Response[langpb.GetModuleConfigDefaultsResponse], error) {
	return connect.NewResponse(&langpb.GetModuleConfigDefaultsResponse{
		Watch:     []string{"**/*.py"},
		DeployDir: ".ftl",
	}), nil
}
