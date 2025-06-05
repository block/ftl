package common

import (
	"archive/zip"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/block/scaffolder"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/block/ftl"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/watch"
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
		return nil, errors.Errorf("group not a string")
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
		return nil, errors.Wrap(err, "failed to scaffold")
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
		defaults.DevModeBuild = ptr("mvn clean quarkus:dev -Ddev")
		defaults.Build = ptr("mvn -B clean package")
		defaults.DeployDir = "target"
		defaults.Watch = append(defaults.Watch, "pom.xml")
	} else if fileExists(buildGradle) || fileExists(buildGradleKts) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolGradle)
		defaults.DevModeBuild = ptr("gradle clean quarkusDev")
		defaults.Build = ptr("gradle clean build")
		defaults.DeployDir = "build"
		defaults.Watch = append(defaults.Watch, "build.gradle", "build.gradle.kts", "settings.gradle", "gradle.properties")
	} else {
		return nil, errors.Errorf("could not find JVM build file in %s", dir)
	}

	return connect.NewResponse(defaults), nil
}

func (CmdService) GetSQLInterfaces(ctx context.Context, req *connect.Request[langpb.GetSQLInterfacesRequest]) (*connect.Response[langpb.GetSQLInterfacesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.Config)

	interfaces, err := interfacesForGeneratedFiles(config.Dir, config.Module)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get interfaces for generated files")
	}
	return connect.NewResponse(&langpb.GetSQLInterfacesResponse{
		Interfaces: interfaces,
	}), nil
}

func interfacesForGeneratedFiles(moduleDir, name string) ([]*langpb.GetSQLInterfacesResponse_Interface, error) {
	interfaces := []*langpb.GetSQLInterfacesResponse_Interface{}
	generatedDir := filepath.Join(moduleDir, "target", "generated-sources", "ftl-clients", "ftl", name)
	err := watch.WalkDir(generatedDir, false, func(path string, d os.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read generated file at %s", path)
		}
		_, filename := filepath.Split(path)
		newInterfaces, err := interfacesForGeneratedFile(filename, string(content))
		if err != nil {
			return errors.Wrapf(err, "failed to parse generated file at %s", path)
		}
		interfaces = append(interfaces, newInterfaces...)
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, errors.Wrap(err, "failed to walk generated directory")
	}
	return interfaces, nil
}

var kotlinVerbInterfaceRegex = regexp.MustCompile(`SQL query verb(\n|.)*?(public fun interface ([a-zA-Z0-9_]+) : \w+.*? {\n)(.|\n)*?( *override fun call.*?\n})`)
var kotlinDataInterfaceRegex = regexp.MustCompile(`Generated data type for use with SQL query verbs(\n|.)*?(public data class ([a-zA-Z0-9_]+)\(\n(.|\n)*?\n\))`)
var javaVerbInterfaceRegex = regexp.MustCompile(`SQL query verb(\n|.)*?(public interface ([a-zA-Z0-9_]+) extends \w+.*? {\n)(.|\n)*?( *.*? call.*?\n})`)
var javaDataInterfaceRegex = regexp.MustCompile(`SQL query verbs(.|\n)*?public class ([a-zA-Z_]+) {((.|\n)*?)\n}`)
var javaDataFuncInterfaceRegex = regexp.MustCompile(`(public [^\n]*?) {\n`)

func interfacesForGeneratedFile(filename, fileContent string) ([]*langpb.GetSQLInterfacesResponse_Interface, error) {
	interfaces := []*langpb.GetSQLInterfacesResponse_Interface{}
	if strings.HasSuffix(filename, ".kt") {
		for _, match := range kotlinVerbInterfaceRegex.FindAllStringSubmatch(fileContent, -1) {
			if len(match) < 6 {
				return nil, errors.New("unexpected components in verb interface regex result")
			}
			interfaces = append(interfaces, &langpb.GetSQLInterfacesResponse_Interface{
				Name:      match[3],
				Interface: match[2] + match[5],
			})
		}
		for _, match := range kotlinDataInterfaceRegex.FindAllStringSubmatch(fileContent, -1) {
			if len(match) < 5 {
				return nil, errors.Errorf("unexpected components in data interface regex result")
			}
			interfaces = append(interfaces, &langpb.GetSQLInterfacesResponse_Interface{
				Name:      match[3],
				Interface: match[2],
			})
		}
	} else if strings.HasSuffix(filename, ".java") {
		for _, match := range javaVerbInterfaceRegex.FindAllStringSubmatch(fileContent, -1) {
			if len(match) < 6 {
				return nil, errors.New("unexpected components in verb interface regex result")
			}
			interfaces = append(interfaces, &langpb.GetSQLInterfacesResponse_Interface{
				Name:      match[3],
				Interface: match[2] + match[5],
			})
		}
		for _, match := range javaDataInterfaceRegex.FindAllStringSubmatch(fileContent, -1) {
			if len(match) < 5 {
				return nil, errors.New("unexpected components in data interface regex result")
			}
			name := match[2]

			funcs := []string{}
			for _, match := range javaDataFuncInterfaceRegex.FindAllStringSubmatch(match[3], -1) {
				if len(match) < 2 {
					return nil, errors.New("unexpected components in data accessor interface regex result")
				}
				funcs = append(funcs, match[1])
			}

			finalInterface := "public class " + name + " {" + strings.Join(slices.Map(funcs, func(f string) string { return "\n  " + f + ";" }), "") + "\n}"

			interfaces = append(interfaces, &langpb.GetSQLInterfacesResponse_Interface{
				Name:      name,
				Interface: finalInterface,
			})
		}
	}
	return interfaces, nil
}
