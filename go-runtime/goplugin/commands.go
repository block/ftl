package goplugin

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/block/scaffolder"

	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	goruntime "github.com/block/ftl/go-runtime"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
)

type CmdService struct{}

var _ languagepbconnect.LanguageCommandServiceHandler = CmdService{}

func (CmdService) GetNewModuleFlags(ctx context.Context, req *connect.Request[langpb.GetNewModuleFlagsRequest]) (*connect.Response[langpb.GetNewModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetNewModuleFlagsResponse{
		Flags: []*langpb.GetNewModuleFlagsResponse_Flag{
			{
				Name:        "replace",
				Help:        "Replace a module import path with a local path in the initialised FTL module.",
				Envar:       optional.Some("FTL_INIT_GO_REPLACE").Ptr(),
				Short:       optional.Some("r").Ptr(),
				Placeholder: optional.Some("OLD=NEW,...").Ptr(),
			},
		},
	}), nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

// CreateModule generates files for a new module with the requested name
func (CmdService) NewModule(ctx context.Context, req *connect.Request[langpb.NewModuleRequest]) (*connect.Response[langpb.NewModuleResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	flags := req.Msg.Flags.AsMap()
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := scaffoldingContext{
		Name:      req.Msg.Name,
		GoVersion: runtime.Version()[2:],
		Replace:   map[string]string{},
	}
	if replaceValue, ok := flags["replace"]; ok && replaceValue != "" {
		replaceStr, ok := replaceValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid replace flag is not a string: %v", replaceValue)
		}
		for _, replace := range strings.Split(replaceStr, ",") {
			parts := strings.Split(replace, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid replace flag (format: A=B,C=D): %q", replace)
			}
			sctx.Replace[parts[0]] = parts[1]
		}
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy: %s", req.Msg.Dir)
	if err := exec.Command(ctx, log.Debug, req.Msg.Dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return nil, fmt.Errorf("could not tidy: %w", err)
	}
	return connect.NewResponse(&langpb.NewModuleResponse{}), nil
}

// GetModuleConfigDefaults provides default values for ModuleConfig for values that are not configured in the ftl.toml file.
func (CmdService) GetModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.GetModuleConfigDefaultsRequest]) (*connect.Response[langpb.GetModuleConfigDefaultsResponse], error) {
	deployDir := ".ftl"
	watch := []string{"**/*.go", "go.mod", "go.sum"}
	additionalWatch, err := replacementWatches(req.Msg.Dir, deployDir)
	watch = append(watch, additionalWatch...)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&langpb.GetModuleConfigDefaultsResponse{
		Watch:      watch,
		DeployDir:  deployDir,
		SqlRootDir: "db",
	}), nil
}
