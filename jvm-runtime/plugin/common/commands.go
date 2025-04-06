package common

import (
	"archive/zip"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/block/scaffolder"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/block/ftl"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/log"
)

type CmdService struct {
	ScaffoldFiles *zip.Reader
}

func NewCmdService(scaffoldFiles *zip.Reader) *CmdService {
	return &CmdService{
		ScaffoldFiles: scaffoldFiles,
	}
}

var _ languagepbconnect.LanguageCommandServiceHandler = CmdService{}

func (CmdService) GetNewModuleFlags(ctx context.Context, req *connect.Request[langpb.GetNewModuleFlagsRequest]) (*connect.Response[langpb.GetNewModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetNewModuleFlagsResponse{
		Flags: []*langpb.GetNewModuleFlagsResponse_Flag{
			{
				Name: "group",
				Help: "The Maven groupId of the project.",
			},
		},
	}), nil
}

// NewModule generates files for a new module with the requested name
func (c CmdService) NewModule(ctx context.Context, req *connect.Request[langpb.NewModuleRequest]) (*connect.Response[langpb.NewModuleResponse], error) {
	logger := log.FromContext(ctx)
	logger = logger.Module(req.Msg.Name)
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)
	groupAny, ok := req.Msg.Flags.AsMap()["group"]
	if !ok {
		groupAny = ""
	}
	group, ok := groupAny.(string)
	if !ok {
		return nil, fmt.Errorf("group not a string")
	}
	if group == "" {
		group = "ftl." + req.Msg.Name
	}

	packageDir := strings.ReplaceAll(group, ".", "/")
	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"), // This is still needed, as there is an 'ignore' module in the scaffold dir
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	version := ftl.BaseVersion(ftl.Version)
	if !ftl.IsRelease(version) {
		version = "1.0-SNAPSHOT"
	}
	sctx := struct {
		Dir        string
		Name       string
		Group      string
		PackageDir string
		Version    string
	}{
		Dir:        projConfig.Path,
		Name:       req.Msg.Name,
		Group:      group,
		PackageDir: packageDir,
		Version:    version,
	}
	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(c.ScaffoldFiles, parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	return connect.NewResponse(&langpb.NewModuleResponse{}), nil
}

func (CmdService) GetModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.GetModuleConfigDefaultsRequest]) (*connect.Response[langpb.GetModuleConfigDefaultsResponse], error) {
	defaults := &langpb.GetModuleConfigDefaultsResponse{
		LanguageConfig: &structpb.Struct{Fields: map[string]*structpb.Value{}},
		Watch:          []string{"src/**", "build/generated", "target/generated-sources", "src/main/resources/db"},
		SqlRootDir:     "src/main/resources/db",
	}
	dir := req.Msg.Dir
	pom := filepath.Join(dir, "pom.xml")
	buildGradle := filepath.Join(dir, "build.gradle")
	buildGradleKts := filepath.Join(dir, "build.gradle.kts")
	if fileExists(pom) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolMaven)
		defaults.DevModeBuild = ptr("mvn -Dquarkus.console.enabled=false clean quarkus:dev")
		defaults.Build = ptr("mvn -B clean package")
		defaults.DeployDir = "target"
		defaults.Watch = append(defaults.Watch, "pom.xml")
	} else if fileExists(buildGradle) || fileExists(buildGradleKts) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolGradle)
		defaults.DevModeBuild = ptr("gradle clean quarkusDev -Dquarkus.console.enabled=false")
		defaults.Build = ptr("gradle clean build")
		defaults.DeployDir = "build"
		defaults.Watch = append(defaults.Watch, "build.gradle", "build.gradle.kts", "settings.gradle", "gradle.properties")
	} else {
		return nil, fmt.Errorf("could not find JVM build file in %s", dir)
	}

	return connect.NewResponse(defaults), nil
}
