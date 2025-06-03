package common

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/beevik/etree"
	"github.com/go-viper/mapstructure/v2"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl"
	hotreloadpb "github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1/hotreloadpbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	langconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/flock"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
)

const BuildLockTimeout = time.Minute
const SchemaFile = "schema.pb"
const ErrorFile = "errors.pb"

var ErrInvalidateDependencies = errors.New("dependencies need to be updated")

type buildContextUpdatedEvent struct {
	buildCtx      buildContext
	schemaChanged bool
}

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
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
	}, nil
}

type Service struct {
	updatesTopic      *pubsub.Topic[buildContextUpdatedEvent]
	devModeRunning    atomic.Int32
	hotReloadClient   hotreloadpbconnect.HotReloadServiceClient
	devModeEndpoint   string
	hotReloadEndpoint string
	debugPort32       int32
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New() *Service {
	return &Service{
		updatesTopic:   pubsub.New[buildContextUpdatedEvent](),
		devModeRunning: atomic.NewInt32(0),
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {

	if req.Msg.Schema == nil {
		return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
	}
	sch, err := schema.FromProto(req.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse schema from proto")
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	_, err = s.writeGenericSchemaFiles(ctx, sch, config, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
}

// Build the module and stream back build events.
//
// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
// end of the build.
//
// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
// calls.
func (s *Service) Build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (*connect.Response[langpb.BuildResponse], error) {
	logger := log.FromContext(ctx)
	_ = os.Setenv("QUARKUS_ANALYTICS_DISABLED", "true") //nolint:errcheck
	logger = logger.Module(req.Msg.BuildContext.ModuleConfig.Name)
	ctx = log.ContextWithLogger(ctx, logger)
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	changed, err := s.writeGenericSchemaFiles(ctx, buildCtx.Schema, buildCtx.Config, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write generic schema files")
	}
	projectConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)
	running := s.devModeRunning.Load()
	if running == 1 {
		return s.reloadDevMode(ctx, buildCtx, changed)
	}
	if req.Msg.DevModeBuild {
		ensureCorrectFTLVersion(ctx, buildCtx)
		return s.runQuarkusDev(ctx, projectConfig, buildCtx)
	}

	// Initial build
	return buildAndSend(ctx, projectConfig, buildCtx)
}

func (s *Service) runQuarkusDev(ctx context.Context, projectConfig projectconfig.Config, buildCtx buildContext) (*connect.Response[langpb.BuildResponse], error) {
	logger := log.FromContext(ctx)

	output := &errorDetector{
		logger: logger,
	}

	events := make(chan buildContextUpdatedEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)
	release, err := flock.Acquire(ctx, buildCtx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not acquire build lock")
	}
	defer release() //nolint:errcheck
	address, err := plugin.AllocatePort()
	if err != nil {
		return nil, errors.Wrap(err, "could not allocate port")
	}
	ctx = log.ContextWithLogger(ctx, logger)
	s.devModeEndpoint = fmt.Sprintf("http://localhost:%d", address.Port)
	devModeBuild := buildCtx.Config.DevModeBuild
	debugPort, err := plugin.AllocatePort()
	s.debugPort32 = int32(debugPort.Port) //nolint

	if err == nil {
		devModeBuild = fmt.Sprintf("%s -Ddebug=%d", devModeBuild, debugPort.Port)
	}
	hotReloadPort, err := plugin.AllocatePort()
	if err != nil {
		return nil, errors.Wrap(err, "could not allocate port")
	}
	devModeBuild = fmt.Sprintf("%s -Dftl.language.port=%d", devModeBuild, hotReloadPort.Port)

	if os.Getenv("FTL_SUSPEND") == "true" {
		devModeBuild += " -Dsuspend "
	}
	s.launchQuarkusProcessAsync(ctx, devModeBuild, projectConfig, buildCtx, output)

	responses := make(chan *connect.Response[langpb.BuildResponse], 2)
	errorChan := make(chan error, 1)

	go func() {
		// Wait for the plugin to start.
		s.hotReloadEndpoint = fmt.Sprintf("http://localhost:%d", hotReloadPort.Port)
		client := s.connectReloadClient(ctx, s.hotReloadEndpoint, output)
		if err != nil || client == nil {
			errorChan <- errors.WithStack(err)
			return
		}
		logger.Debugf("Dev mode process started")
		s.hotReloadClient = client
		res, err := client.Watch(ctx, connect.NewRequest(&hotreloadpb.WatchRequest{}))
		if err != nil {
			errorChan <- errors.Wrap(err, "could not get initial hot reload state")
			return
		}
		logger.Debugf("watch client")
		resp, err := s.handleState(ctx, res.Msg.State, buildCtx)
		if err != nil {
			errorChan <- errors.Wrap(err, "could not handle state")
			return
		}
		responses <- resp
	}()
	select {
	case <-ctx.Done():
		// If the parent context is done we just return
		// the context is done before we notified the build engine
		// we need to send a build failure event

		ers := langpb.ErrorsToProto(output.FinalizeCapture(true))
		ers.Errors = append(ers.Errors, &langpb.Error{Msg: "The dev mode process exited", Level: langpb.Error_ERROR_LEVEL_ERROR, Type: langpb.Error_ERROR_TYPE_COMPILER})
		return connect.NewResponse(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				Errors: ers,
			}}}), nil
	case err := <-errorChan:
		return nil, err
	case resp := <-responses:
		return resp, nil
	}

}

func (s *Service) reloadDevMode(ctx context.Context, buildCtx buildContext, schemaChanged bool) (*connect.Response[langpb.BuildResponse], error) {
	logger := log.FromContext(ctx)
	newDeps, err := extractDependencies(buildCtx.Config.Module, buildCtx.Config.Dir)
	if err != nil {
		logger.Errorf(err, "could not extract dependencies")
	} else if !slices.Equal(islices.Sort(newDeps), islices.Sort(buildCtx.Dependencies)) {
		return connect.NewResponse(&langpb.BuildResponse{
			Event: &langpb.BuildResponse_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					InvalidateDependencies: true,
				},
			},
		}), nil
	}

	result, err := s.hotReloadClient.Reload(ctx, connect.NewRequest(&hotreloadpb.ReloadRequest{SchemaChanged: schemaChanged}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to invoke reload")
	}
	handleReloadResponse(result)
	return s.handleState(ctx, result.Msg.State, buildCtx)
}

func (s *Service) doReload(ctx context.Context, client hotreloadpbconnect.HotReloadServiceClient, request *hotreloadpb.ReloadRequest, reloadEvents chan *buildResult, buildContextUpdated bool, newKey key.Deployment) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Sending hot reload request")
	result, err := client.Reload(ctx, connect.NewRequest(request))

	if err != nil {
		logger.Debugf("Reload failed, attempting to reconnect")
		// If the connection has failed we try again
		err = s.connectReloadClient(ctx, client)

		if err != nil {
			logger.Debugf("Reconnect failed, unable to connect to client")
			return err
		}
		result, err = client.Reload(ctx, connect.NewRequest(&hotreloadpb.ReloadRequest{}))
	}
	if err != nil {
		logger.Debugf("Unable to invoke reload on the JVM") //TODO: restart
		return errors.Wrap(err, "unable to invoke hot reload")
	}
	handleReloadResponse(result)
	reloadEvents <- &buildResult{state: result.Msg.GetState(), failed: result.Msg.Failed, bctx: s.buildContext.Load(), buildContextUpdated: buildContextUpdated}
	return nil
}

func handleReloadResponse(result *connect.Response[hotreloadpb.ReloadResponse]) {
	if result.Msg.State.Module != nil {
		if result.Msg.State.Module.Runtime == nil {
			result.Msg.State.Module.Runtime = &schemapb.ModuleRuntime{}
		}
		if result.Msg.State.Module.Runtime.Deployment == nil {
			result.Msg.State.Module.Runtime.Deployment = &schemapb.ModuleRuntimeDeployment{}
		}
	}
}

func (s *Service) handleState(ctx context.Context, state *hotreloadpb.SchemaState, buildCtx buildContext) (*connect.Response[langpb.BuildResponse], error) {
	logger := log.FromContext(ctx)
	changed := state.GetNewRunnerRequired()
	errorList := state.GetErrors()
	logger.Debugf("Checking for schema changes: changed: %v", changed)

	if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(errorList)) {
		// skip reading schema
		logger.Warnf("Build failed, skipping schema, sending build failure")
		return connect.NewResponse(&langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				Errors: errorList,
			}}}), nil
	}
	moduleProto := state.GetModule()
	moduleSch, err := schema.ModuleFromProto(moduleProto)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse module proto")
	}
	if _, validationErr := schema.ValidateModuleInSchema(buildCtx.Schema, optional.Some(moduleSch)); validationErr != nil {
		return connect.NewResponse(buildFailure(builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   validationErr.Error(),
		})), nil
	}

	return connect.NewResponse(&langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				Module:               moduleSch.ToProto(),
				DevEndpoint:          ptr(s.devModeEndpoint),
				DevHotReloadEndpoint: ptr(s.hotReloadEndpoint),
				DebugPort:            &s.debugPort32,
				Deploy:               []string{SchemaFile},
				DevHotReloadVersion:  &state.Version,
			},
		},
	}), nil
}

func (s *Service) launchQuarkusProcessAsync(ctx context.Context, devModeBuild string, projectConfig projectconfig.Config, buildCtx buildContext, stdout *errorDetector) {
	go func() {
		logger := log.FromContext(ctx)
		ctx, cancel := context.WithCancelCause(log.ContextWithLogger(context.Background(), logger))
		s.devModeRunning.Store(1)
		defer func() {
			s.devModeRunning.Store(0)
			cancel(nil)
		}()
		logger.Infof("Using dev mode build command '%s'", devModeBuild)
		command := exec.Command(ctx, log.Debug, buildCtx.Config.Dir, "bash", "-c", devModeBuild)
		if os.Getenv("MAVEN_OPTS") == "" {
			command.Env = append(command.Env, "MAVEN_OPTS=-Xmx2048m")
		}
		command.Env = append(command.Env, fmt.Sprintf("FTL_BIND=%s", s.devModeEndpoint), "FTL_MODULE_NAME="+buildCtx.Config.Module, "FTL_PROJECT_ROOT="+projectConfig.Root())
		command.Stdout = stdout
		command.Stderr = os.Stderr
		err := command.Run()
		if err != nil {
			stdout.FinalizeCapture(true)
			logger.Errorf(err, "Dev mode process exited with error")
			cancel(errors.Wrap(errors.Join(err, context.Canceled), "dev mode process exited with error"))
		} else {
			logger.Infof("Dev mode process exited")
			cancel(errors.Wrap(context.Canceled, "dev mode process exited"))
		}
	}()
}

func (s *Service) connectReloadClient(ctx context.Context, client hotreloadpbconnect.HotReloadServiceClient) error {
	logger := log.FromContext(ctx)
	err := rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute*100, client)
	if err != nil {
		logger.Infof("Dev mode process failed to start")
		select {
		case <-ctx.Done():
			return errors.Errorf("dev mode process exited")
		default:
		}
		return errors.Wrap(err, "timed out waiting for start")
	}
	return nil
}

func build(ctx context.Context, projectConfig projectconfig.Config, bctx buildContext) (*langpb.BuildResponse, error) {
	logger := log.FromContext(ctx)
	release, err := flock.Acquire(ctx, bctx.Config.BuildLock, BuildLockTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not acquire build lock")
	}
	defer release() //nolint:errcheck

	deps, err := extractDependencies(bctx.Config.Module, bctx.Config.Dir)
	if err != nil {
		return nil, errors.Wrap(err, "could not extract dependencies")
	}

	if !slices.Equal(islices.Sort(deps), islices.Sort(bctx.Dependencies)) {
		// dependencies have changed
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				InvalidateDependencies: true,
			},
		}}, nil
	}
	ensureCorrectFTLVersion(ctx, bctx)
	output := &errorDetector{
		logger: logger,
	}
	config := bctx.Config
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	command.Env = append(command.Env, "FTL_MODULE_NAME="+bctx.Config.Module, "FTL_PROJECT_ROOT="+projectConfig.Root())
	command.Stdout = output
	command.Stderr = os.Stderr
	err = command.Run()

	if err != nil {
		buildErrs := output.FinalizeCapture(true)
		if len(buildErrs) == 0 {
			buildErrs = []builderrors.Error{{Msg: "Compile process unexpectedly exited without reporting any errors", Level: builderrors.ERROR, Type: builderrors.COMPILER}}
		}
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{BuildFailure: &langpb.BuildFailure{
			Errors: langpb.ErrorsToProto(buildErrs),
		}}}, nil
	}

	buildErrs, err := loadProtoErrors(config)
	capturedErrors := output.FinalizeCapture(false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load build errors")
	}
	buildErrs.Errors = append(buildErrs.Errors, langpb.ErrorsToProto(capturedErrors).Errors...)
	if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(buildErrs)) {
		// skip reading schema
		return &langpb.BuildResponse{Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				Errors: buildErrs,
			}}}, nil
	}

	moduleProto, err := readSchema(bctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				Errors: buildErrs,
				Module: moduleProto,
				Deploy: []string{"launch", "quarkus-app"},
			},
		},
	}, nil
}

func ensureCorrectFTLVersion(ctx context.Context, bctx buildContext) {
	logger := log.FromContext(ctx)
	javaConfig, err := loadJavaConfig(bctx.Config.LanguageConfig, bctx.Config.Language)
	if err != nil {
		logger.Errorf(err, "unable to read JVM config %s", bctx.Config.Dir)
		return
	}
	if javaConfig.BuildTool == JavaBuildToolMaven {
		if err := setPOMProperties(ctx, bctx.Config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Errorf(err, "unable to update ftl.version in %s", bctx.Config.Dir)
		}
	}
}

func readSchema(bctx buildContext) (*schemapb.Module, error) {
	path := filepath.Join(bctx.Config.DeployDir, SchemaFile)
	moduleSchema, err := schema.ModuleFromProtoFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read schema for module: %s from %s", bctx.Config.Module, path)
	}

	moduleSchema.Runtime = &schema.ModuleRuntime{
		Base: schema.ModuleRuntimeBase{
			CreateTime: time.Now(),
			Language:   bctx.Config.Language,
			Image:      "ftl0/ftl-runner-jvm",
		},
		Scaling: &schema.ModuleRuntimeScaling{
			MinReplicas: 1,
		},
	}

	moduleProto := moduleSchema.ToProto()
	return moduleProto, nil
}

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, projectConfig projectconfig.Config, buildCtx buildContext) (*connect.Response[langpb.BuildResponse], error) {
	buildEvent, err := build(ctx, projectConfig, buildCtx)
	if err != nil {
		buildEvent = buildFailure(builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		})
	}
	return connect.NewResponse(buildEvent), nil
}

// buildFailure creates a BuildFailure event based on build errors.
func buildFailure(errs ...builderrors.Error) *langpb.BuildResponse {
	return &langpb.BuildResponse{
		Event: &langpb.BuildResponse_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				Errors:                 langpb.ErrorsToProto(errs),
				InvalidateDependencies: false,
			},
		},
	}
}

const JavaBuildToolMaven string = "maven"
const JavaBuildToolGradle string = "gradle"

type JavaConfig struct {
	BuildTool string `mapstructure:"build-tool"`
}

func loadJavaConfig(languageConfig any, language string) (JavaConfig, error) {
	var javaConfig JavaConfig
	err := mapstructure.Decode(languageConfig, &javaConfig)
	if err != nil {
		return JavaConfig{}, errors.Wrapf(err, "failed to decode %s config", language)
	}
	return javaConfig, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.GetDependenciesRequest]) (*connect.Response[langpb.GetDependenciesResponse], error) {
	modules, err := extractDependencies(req.Msg.ModuleConfig.Name, req.Msg.ModuleConfig.Dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse[langpb.GetDependenciesResponse](&langpb.GetDependenciesResponse{Modules: modules}), nil
}

func extractDependencies(moduleName string, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	// We also attempt to look at kotlin files
	// As the Java module supports both
	kotin, kotlinErr := extractKotlinFTLImports(moduleName, dir)
	if kotlinErr == nil {
		// We don't really care about the error case, its probably a Java project
		for _, imp := range kotin {
			dependencies[imp] = true
		}
	}
	javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to walk directory")
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := javaImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == moduleName {
					continue
				}
				dependencies[module] = true
			}
		}
		return errors.WithStack(scanner.Err())
	})

	// We only error out if they both failed
	if err != nil && kotlinErr != nil {
		return nil, errors.Wrapf(err, "%s: failed to extract dependencies from Java module", moduleName)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

func extractKotlinFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "could not open file while extracting dependencies")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == self {
					continue
				}
				dependencies[module] = true
			}
		}
		return errors.WithStack(scanner.Err())
	})

	if err != nil {
		return nil, errors.Wrapf(err, "%s: failed to extract dependencies from Kotlin module", self)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

// setPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func setPOMProperties(ctx context.Context, baseDir string) error {
	changed := false
	logger := log.FromContext(ctx)
	ftlVersion := ftl.BaseVersion(ftl.Version)
	if !ftl.IsRelease(ftlVersion) || ftlVersion != ftl.BaseVersion(ftl.Version) {
		ftlVersion = "1.0-SNAPSHOT"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return errors.Wrapf(err, "unable to read %s", pomFile)
	}
	root := tree.Root()

	parent := root.SelectElement("parent")
	versionSet := false
	if parent != nil {
		// You can't use properties in the parent
		// If they are using our parent then we want to update the version
		group := parent.SelectElement("groupId")
		artifact := parent.SelectElement("artifactId")
		if group.Text() == "xyz.block.ftl" && (artifact.Text() == "ftl-build-parent-java" || artifact.Text() == "ftl-build-parent-kotlin") {
			version := parent.SelectElement("version")
			if version != nil {
				if version.Text() != ftlVersion {
					version.SetText(ftlVersion)
					changed = true
				}
				versionSet = true
			}
		}
	}

	propChanged, err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return errors.WithStack(err)
	}
	if propChanged || changed {
		err = tree.WriteToFile(pomFile)
		if err != nil {
			return errors.Wrapf(err, "unable to write %s", pomFile)
		}
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) (bool, error) {
	properties := root.SelectElement("properties")
	if properties == nil {
		return false, errors.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return false, errors.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	if version.Text() == ftlVersion {
		return false, nil
	}
	version.SetText(ftlVersion)
	return true, nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*langpb.ErrorList, error) {
	errorsPath := filepath.Join(config.DeployDir, "errors.pb")
	if _, err := os.Stat(errorsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(errorsPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not load build errors file")
	}

	errorspb := &langpb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal build error")
	}
	return errorspb, nil
}

func ptr(s string) *string {
	return &s
}

func (s *Service) writeGenericSchemaFiles(ctx context.Context, v *schema.Schema, config moduleconfig.AbsModuleConfig, containsGeneratedSchema bool) (bool, error) {
	if v == nil {
		return false, nil
	}
	logger := log.FromContext(ctx)
	modPath := filepath.Join(config.Dir, "src", "main", "ftl-module-schema")
	err := os.MkdirAll(modPath, 0750)
	if err != nil {
		return false, errors.Wrapf(err, "failed to create directory %s", modPath)
	}
	changed := false

	for _, mod := range v.InternalModules() {
		if mod.Name == config.Module {
			if containsGeneratedSchema {
				logger.Debugf("writing generated schema files for %s", mod.Name)
				sw, err := s.writeGeneratedSchemaFiles(mod, modPath)
				if err != nil {
					return false, errors.WithStack(err)
				}
				changed = changed || sw
			}
			continue
		}
		deps := v.ModuleDependencies(mod.Name)
		if deps[config.Module] != nil {
			continue
		}
		data, err := schema.ModuleToBytes(mod)
		if err != nil {
			return false, errors.Wrapf(err, "failed to export module schema for module %s", mod.Name)
		}
		schemaFile := filepath.Join(modPath, mod.Name+".pb")
		if fileExists(schemaFile) {
			existing, err := os.ReadFile(schemaFile)
			if err == nil {
				// We ignore errors, but if the read succeeded we need to check if the file has changed
				if bytes.Equal(existing, data) {
					continue
				}
			}
		}
		changed = true
		err = os.WriteFile(schemaFile, data, 0644) // #nosec
		logger.Debugf("writing schema files for %s to %s", mod.Name, schemaFile)
		if err != nil {
			return false, errors.Wrapf(err, "failed to write schema file for module %s", mod.Name)
		}
	}
	return changed, nil
}

func (s *Service) writeGeneratedSchemaFiles(m *schema.Module, schemaDir string) (bool, error) {
	if m == nil {
		return false, nil
	}
	generatedDir := filepath.Join(schemaDir, "generated")
	err := os.MkdirAll(generatedDir, 0750)
	if err != nil {
		return false, errors.Wrapf(err, "failed to create directory %s", generatedDir)
	}
	schemaFile := filepath.Join(generatedDir, m.Name+".pb")

	generatedModule := m.ToGeneratedModule()
	if len(generatedModule.Decls) == 0 {
		if fileExists(schemaFile) {
			err := os.Remove(schemaFile)
			if err != nil {
				return false, errors.Wrapf(err, "failed to remove generated schema file for module %s", m.Name)
			}
		}
		return false, nil
	}

	data, err := schema.ModuleToBytes(generatedModule)
	if err != nil {
		return false, errors.Wrapf(err, "failed to export generated module schema for module %s", m.Name)
	}

	if fileExists(schemaFile) {
		existing, err := os.ReadFile(schemaFile)
		if err == nil {
			// file has not changed, no need to write
			if bytes.Equal(existing, data) {
				return false, nil
			}
		}
	}

	err = os.WriteFile(schemaFile, data, 0644) // #nosec
	if err != nil {
		return false, errors.Wrapf(err, "failed to write generated schema file for module %s", m.Name)
	}
	return true, nil
}
