// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/language/v1/language.proto

package languagepbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/language/v1"
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_7_0

const (
	// LanguageServiceName is the fully-qualified name of the LanguageService service.
	LanguageServiceName = "xyz.block.ftl.language.v1.LanguageService"
	// HotReloadServiceName is the fully-qualified name of the HotReloadService service.
	HotReloadServiceName = "xyz.block.ftl.language.v1.HotReloadService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// LanguageServicePingProcedure is the fully-qualified name of the LanguageService's Ping RPC.
	LanguageServicePingProcedure = "/xyz.block.ftl.language.v1.LanguageService/Ping"
	// LanguageServiceGetCreateModuleFlagsProcedure is the fully-qualified name of the LanguageService's
	// GetCreateModuleFlags RPC.
	LanguageServiceGetCreateModuleFlagsProcedure = "/xyz.block.ftl.language.v1.LanguageService/GetCreateModuleFlags"
	// LanguageServiceCreateModuleProcedure is the fully-qualified name of the LanguageService's
	// CreateModule RPC.
	LanguageServiceCreateModuleProcedure = "/xyz.block.ftl.language.v1.LanguageService/CreateModule"
	// LanguageServiceModuleConfigDefaultsProcedure is the fully-qualified name of the LanguageService's
	// ModuleConfigDefaults RPC.
	LanguageServiceModuleConfigDefaultsProcedure = "/xyz.block.ftl.language.v1.LanguageService/ModuleConfigDefaults"
	// LanguageServiceGetDependenciesProcedure is the fully-qualified name of the LanguageService's
	// GetDependencies RPC.
	LanguageServiceGetDependenciesProcedure = "/xyz.block.ftl.language.v1.LanguageService/GetDependencies"
	// LanguageServiceBuildProcedure is the fully-qualified name of the LanguageService's Build RPC.
	LanguageServiceBuildProcedure = "/xyz.block.ftl.language.v1.LanguageService/Build"
	// LanguageServiceBuildContextUpdatedProcedure is the fully-qualified name of the LanguageService's
	// BuildContextUpdated RPC.
	LanguageServiceBuildContextUpdatedProcedure = "/xyz.block.ftl.language.v1.LanguageService/BuildContextUpdated"
	// LanguageServiceGenerateStubsProcedure is the fully-qualified name of the LanguageService's
	// GenerateStubs RPC.
	LanguageServiceGenerateStubsProcedure = "/xyz.block.ftl.language.v1.LanguageService/GenerateStubs"
	// LanguageServiceSyncStubReferencesProcedure is the fully-qualified name of the LanguageService's
	// SyncStubReferences RPC.
	LanguageServiceSyncStubReferencesProcedure = "/xyz.block.ftl.language.v1.LanguageService/SyncStubReferences"
	// HotReloadServicePingProcedure is the fully-qualified name of the HotReloadService's Ping RPC.
	HotReloadServicePingProcedure = "/xyz.block.ftl.language.v1.HotReloadService/Ping"
	// HotReloadServiceRunnerStartedProcedure is the fully-qualified name of the HotReloadService's
	// RunnerStarted RPC.
	HotReloadServiceRunnerStartedProcedure = "/xyz.block.ftl.language.v1.HotReloadService/RunnerStarted"
)

// LanguageServiceClient is a client for the xyz.block.ftl.language.v1.LanguageService service.
type LanguageServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get language specific flags that can be used to create a new module.
	GetCreateModuleFlags(context.Context, *connect.Request[v11.GetCreateModuleFlagsRequest]) (*connect.Response[v11.GetCreateModuleFlagsResponse], error)
	// Generates files for a new module with the requested name
	CreateModule(context.Context, *connect.Request[v11.CreateModuleRequest]) (*connect.Response[v11.CreateModuleResponse], error)
	// Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
	ModuleConfigDefaults(context.Context, *connect.Request[v11.ModuleConfigDefaultsRequest]) (*connect.Response[v11.ModuleConfigDefaultsResponse], error)
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error)
	// Build the module and stream back build events.
	//
	// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
	// end of the build.
	//
	// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
	// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
	// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
	// calls.
	Build(context.Context, *connect.Request[v11.BuildRequest]) (*connect.ServerStreamForClient[v11.BuildResponse], error)
	// While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
	// build context is updated.
	//
	// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
	// event with the updated build context id with "is_automatic_rebuild" as false.
	//
	// If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
	// build stream, it must fail the BuildContextUpdated call.
	BuildContextUpdated(context.Context, *connect.Request[v11.BuildContextUpdatedRequest]) (*connect.Response[v11.BuildContextUpdatedResponse], error)
	// Generate stubs for a module.
	//
	// Stubs allow modules to import other module's exported interface. If a language does not need this step,
	// then it is not required to do anything in this call.
	//
	// This call is not tied to the module that this plugin is responsible for. A plugin of each language will
	// be chosen to generate stubs for each module.
	GenerateStubs(context.Context, *connect.Request[v11.GenerateStubsRequest]) (*connect.Response[v11.GenerateStubsResponse], error)
	// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
	// references to external modules, regardless of whether they are dependencies.
	//
	// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
	// import the modules when users start reference them.
	//
	// It is optional to do anything with this call.
	SyncStubReferences(context.Context, *connect.Request[v11.SyncStubReferencesRequest]) (*connect.Response[v11.SyncStubReferencesResponse], error)
}

// NewLanguageServiceClient constructs a client for the xyz.block.ftl.language.v1.LanguageService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewLanguageServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) LanguageServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &languageServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+LanguageServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		getCreateModuleFlags: connect.NewClient[v11.GetCreateModuleFlagsRequest, v11.GetCreateModuleFlagsResponse](
			httpClient,
			baseURL+LanguageServiceGetCreateModuleFlagsProcedure,
			opts...,
		),
		createModule: connect.NewClient[v11.CreateModuleRequest, v11.CreateModuleResponse](
			httpClient,
			baseURL+LanguageServiceCreateModuleProcedure,
			opts...,
		),
		moduleConfigDefaults: connect.NewClient[v11.ModuleConfigDefaultsRequest, v11.ModuleConfigDefaultsResponse](
			httpClient,
			baseURL+LanguageServiceModuleConfigDefaultsProcedure,
			opts...,
		),
		getDependencies: connect.NewClient[v11.GetDependenciesRequest, v11.GetDependenciesResponse](
			httpClient,
			baseURL+LanguageServiceGetDependenciesProcedure,
			opts...,
		),
		build: connect.NewClient[v11.BuildRequest, v11.BuildResponse](
			httpClient,
			baseURL+LanguageServiceBuildProcedure,
			opts...,
		),
		buildContextUpdated: connect.NewClient[v11.BuildContextUpdatedRequest, v11.BuildContextUpdatedResponse](
			httpClient,
			baseURL+LanguageServiceBuildContextUpdatedProcedure,
			opts...,
		),
		generateStubs: connect.NewClient[v11.GenerateStubsRequest, v11.GenerateStubsResponse](
			httpClient,
			baseURL+LanguageServiceGenerateStubsProcedure,
			opts...,
		),
		syncStubReferences: connect.NewClient[v11.SyncStubReferencesRequest, v11.SyncStubReferencesResponse](
			httpClient,
			baseURL+LanguageServiceSyncStubReferencesProcedure,
			opts...,
		),
	}
}

// languageServiceClient implements LanguageServiceClient.
type languageServiceClient struct {
	ping                 *connect.Client[v1.PingRequest, v1.PingResponse]
	getCreateModuleFlags *connect.Client[v11.GetCreateModuleFlagsRequest, v11.GetCreateModuleFlagsResponse]
	createModule         *connect.Client[v11.CreateModuleRequest, v11.CreateModuleResponse]
	moduleConfigDefaults *connect.Client[v11.ModuleConfigDefaultsRequest, v11.ModuleConfigDefaultsResponse]
	getDependencies      *connect.Client[v11.GetDependenciesRequest, v11.GetDependenciesResponse]
	build                *connect.Client[v11.BuildRequest, v11.BuildResponse]
	buildContextUpdated  *connect.Client[v11.BuildContextUpdatedRequest, v11.BuildContextUpdatedResponse]
	generateStubs        *connect.Client[v11.GenerateStubsRequest, v11.GenerateStubsResponse]
	syncStubReferences   *connect.Client[v11.SyncStubReferencesRequest, v11.SyncStubReferencesResponse]
}

// Ping calls xyz.block.ftl.language.v1.LanguageService.Ping.
func (c *languageServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetCreateModuleFlags calls xyz.block.ftl.language.v1.LanguageService.GetCreateModuleFlags.
func (c *languageServiceClient) GetCreateModuleFlags(ctx context.Context, req *connect.Request[v11.GetCreateModuleFlagsRequest]) (*connect.Response[v11.GetCreateModuleFlagsResponse], error) {
	return c.getCreateModuleFlags.CallUnary(ctx, req)
}

// CreateModule calls xyz.block.ftl.language.v1.LanguageService.CreateModule.
func (c *languageServiceClient) CreateModule(ctx context.Context, req *connect.Request[v11.CreateModuleRequest]) (*connect.Response[v11.CreateModuleResponse], error) {
	return c.createModule.CallUnary(ctx, req)
}

// ModuleConfigDefaults calls xyz.block.ftl.language.v1.LanguageService.ModuleConfigDefaults.
func (c *languageServiceClient) ModuleConfigDefaults(ctx context.Context, req *connect.Request[v11.ModuleConfigDefaultsRequest]) (*connect.Response[v11.ModuleConfigDefaultsResponse], error) {
	return c.moduleConfigDefaults.CallUnary(ctx, req)
}

// GetDependencies calls xyz.block.ftl.language.v1.LanguageService.GetDependencies.
func (c *languageServiceClient) GetDependencies(ctx context.Context, req *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error) {
	return c.getDependencies.CallUnary(ctx, req)
}

// Build calls xyz.block.ftl.language.v1.LanguageService.Build.
func (c *languageServiceClient) Build(ctx context.Context, req *connect.Request[v11.BuildRequest]) (*connect.ServerStreamForClient[v11.BuildResponse], error) {
	return c.build.CallServerStream(ctx, req)
}

// BuildContextUpdated calls xyz.block.ftl.language.v1.LanguageService.BuildContextUpdated.
func (c *languageServiceClient) BuildContextUpdated(ctx context.Context, req *connect.Request[v11.BuildContextUpdatedRequest]) (*connect.Response[v11.BuildContextUpdatedResponse], error) {
	return c.buildContextUpdated.CallUnary(ctx, req)
}

// GenerateStubs calls xyz.block.ftl.language.v1.LanguageService.GenerateStubs.
func (c *languageServiceClient) GenerateStubs(ctx context.Context, req *connect.Request[v11.GenerateStubsRequest]) (*connect.Response[v11.GenerateStubsResponse], error) {
	return c.generateStubs.CallUnary(ctx, req)
}

// SyncStubReferences calls xyz.block.ftl.language.v1.LanguageService.SyncStubReferences.
func (c *languageServiceClient) SyncStubReferences(ctx context.Context, req *connect.Request[v11.SyncStubReferencesRequest]) (*connect.Response[v11.SyncStubReferencesResponse], error) {
	return c.syncStubReferences.CallUnary(ctx, req)
}

// LanguageServiceHandler is an implementation of the xyz.block.ftl.language.v1.LanguageService
// service.
type LanguageServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get language specific flags that can be used to create a new module.
	GetCreateModuleFlags(context.Context, *connect.Request[v11.GetCreateModuleFlagsRequest]) (*connect.Response[v11.GetCreateModuleFlagsResponse], error)
	// Generates files for a new module with the requested name
	CreateModule(context.Context, *connect.Request[v11.CreateModuleRequest]) (*connect.Response[v11.CreateModuleResponse], error)
	// Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
	ModuleConfigDefaults(context.Context, *connect.Request[v11.ModuleConfigDefaultsRequest]) (*connect.Response[v11.ModuleConfigDefaultsResponse], error)
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error)
	// Build the module and stream back build events.
	//
	// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
	// end of the build.
	//
	// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
	// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
	// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
	// calls.
	Build(context.Context, *connect.Request[v11.BuildRequest], *connect.ServerStream[v11.BuildResponse]) error
	// While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
	// build context is updated.
	//
	// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
	// event with the updated build context id with "is_automatic_rebuild" as false.
	//
	// If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
	// build stream, it must fail the BuildContextUpdated call.
	BuildContextUpdated(context.Context, *connect.Request[v11.BuildContextUpdatedRequest]) (*connect.Response[v11.BuildContextUpdatedResponse], error)
	// Generate stubs for a module.
	//
	// Stubs allow modules to import other module's exported interface. If a language does not need this step,
	// then it is not required to do anything in this call.
	//
	// This call is not tied to the module that this plugin is responsible for. A plugin of each language will
	// be chosen to generate stubs for each module.
	GenerateStubs(context.Context, *connect.Request[v11.GenerateStubsRequest]) (*connect.Response[v11.GenerateStubsResponse], error)
	// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
	// references to external modules, regardless of whether they are dependencies.
	//
	// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
	// import the modules when users start reference them.
	//
	// It is optional to do anything with this call.
	SyncStubReferences(context.Context, *connect.Request[v11.SyncStubReferencesRequest]) (*connect.Response[v11.SyncStubReferencesResponse], error)
}

// NewLanguageServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewLanguageServiceHandler(svc LanguageServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	languageServicePingHandler := connect.NewUnaryHandler(
		LanguageServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	languageServiceGetCreateModuleFlagsHandler := connect.NewUnaryHandler(
		LanguageServiceGetCreateModuleFlagsProcedure,
		svc.GetCreateModuleFlags,
		opts...,
	)
	languageServiceCreateModuleHandler := connect.NewUnaryHandler(
		LanguageServiceCreateModuleProcedure,
		svc.CreateModule,
		opts...,
	)
	languageServiceModuleConfigDefaultsHandler := connect.NewUnaryHandler(
		LanguageServiceModuleConfigDefaultsProcedure,
		svc.ModuleConfigDefaults,
		opts...,
	)
	languageServiceGetDependenciesHandler := connect.NewUnaryHandler(
		LanguageServiceGetDependenciesProcedure,
		svc.GetDependencies,
		opts...,
	)
	languageServiceBuildHandler := connect.NewServerStreamHandler(
		LanguageServiceBuildProcedure,
		svc.Build,
		opts...,
	)
	languageServiceBuildContextUpdatedHandler := connect.NewUnaryHandler(
		LanguageServiceBuildContextUpdatedProcedure,
		svc.BuildContextUpdated,
		opts...,
	)
	languageServiceGenerateStubsHandler := connect.NewUnaryHandler(
		LanguageServiceGenerateStubsProcedure,
		svc.GenerateStubs,
		opts...,
	)
	languageServiceSyncStubReferencesHandler := connect.NewUnaryHandler(
		LanguageServiceSyncStubReferencesProcedure,
		svc.SyncStubReferences,
		opts...,
	)
	return "/xyz.block.ftl.language.v1.LanguageService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LanguageServicePingProcedure:
			languageServicePingHandler.ServeHTTP(w, r)
		case LanguageServiceGetCreateModuleFlagsProcedure:
			languageServiceGetCreateModuleFlagsHandler.ServeHTTP(w, r)
		case LanguageServiceCreateModuleProcedure:
			languageServiceCreateModuleHandler.ServeHTTP(w, r)
		case LanguageServiceModuleConfigDefaultsProcedure:
			languageServiceModuleConfigDefaultsHandler.ServeHTTP(w, r)
		case LanguageServiceGetDependenciesProcedure:
			languageServiceGetDependenciesHandler.ServeHTTP(w, r)
		case LanguageServiceBuildProcedure:
			languageServiceBuildHandler.ServeHTTP(w, r)
		case LanguageServiceBuildContextUpdatedProcedure:
			languageServiceBuildContextUpdatedHandler.ServeHTTP(w, r)
		case LanguageServiceGenerateStubsProcedure:
			languageServiceGenerateStubsHandler.ServeHTTP(w, r)
		case LanguageServiceSyncStubReferencesProcedure:
			languageServiceSyncStubReferencesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedLanguageServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedLanguageServiceHandler struct{}

func (UnimplementedLanguageServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.Ping is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GetCreateModuleFlags(context.Context, *connect.Request[v11.GetCreateModuleFlagsRequest]) (*connect.Response[v11.GetCreateModuleFlagsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.GetCreateModuleFlags is not implemented"))
}

func (UnimplementedLanguageServiceHandler) CreateModule(context.Context, *connect.Request[v11.CreateModuleRequest]) (*connect.Response[v11.CreateModuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.CreateModule is not implemented"))
}

func (UnimplementedLanguageServiceHandler) ModuleConfigDefaults(context.Context, *connect.Request[v11.ModuleConfigDefaultsRequest]) (*connect.Response[v11.ModuleConfigDefaultsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.ModuleConfigDefaults is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.GetDependencies is not implemented"))
}

func (UnimplementedLanguageServiceHandler) Build(context.Context, *connect.Request[v11.BuildRequest], *connect.ServerStream[v11.BuildResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.Build is not implemented"))
}

func (UnimplementedLanguageServiceHandler) BuildContextUpdated(context.Context, *connect.Request[v11.BuildContextUpdatedRequest]) (*connect.Response[v11.BuildContextUpdatedResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.BuildContextUpdated is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GenerateStubs(context.Context, *connect.Request[v11.GenerateStubsRequest]) (*connect.Response[v11.GenerateStubsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.GenerateStubs is not implemented"))
}

func (UnimplementedLanguageServiceHandler) SyncStubReferences(context.Context, *connect.Request[v11.SyncStubReferencesRequest]) (*connect.Response[v11.SyncStubReferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.SyncStubReferences is not implemented"))
}

// HotReloadServiceClient is a client for the xyz.block.ftl.language.v1.HotReloadService service.
type HotReloadServiceClient interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// RunnerStarted is called when a runner has started, to tell a live reload server the runner address
	//
	// As lots of functionality such as database connections rely on proxying through the runner, this
	// allows the long running live reload process to know where to proxy to
	RunnerStarted(context.Context, *connect.Request[v11.RunnerStartedRequest]) (*connect.Response[v11.RunnerStartedResponse], error)
}

// NewHotReloadServiceClient constructs a client for the xyz.block.ftl.language.v1.HotReloadService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewHotReloadServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) HotReloadServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &hotReloadServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+HotReloadServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		runnerStarted: connect.NewClient[v11.RunnerStartedRequest, v11.RunnerStartedResponse](
			httpClient,
			baseURL+HotReloadServiceRunnerStartedProcedure,
			opts...,
		),
	}
}

// hotReloadServiceClient implements HotReloadServiceClient.
type hotReloadServiceClient struct {
	ping          *connect.Client[v1.PingRequest, v1.PingResponse]
	runnerStarted *connect.Client[v11.RunnerStartedRequest, v11.RunnerStartedResponse]
}

// Ping calls xyz.block.ftl.language.v1.HotReloadService.Ping.
func (c *hotReloadServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// RunnerStarted calls xyz.block.ftl.language.v1.HotReloadService.RunnerStarted.
func (c *hotReloadServiceClient) RunnerStarted(ctx context.Context, req *connect.Request[v11.RunnerStartedRequest]) (*connect.Response[v11.RunnerStartedResponse], error) {
	return c.runnerStarted.CallUnary(ctx, req)
}

// HotReloadServiceHandler is an implementation of the xyz.block.ftl.language.v1.HotReloadService
// service.
type HotReloadServiceHandler interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// RunnerStarted is called when a runner has started, to tell a live reload server the runner address
	//
	// As lots of functionality such as database connections rely on proxying through the runner, this
	// allows the long running live reload process to know where to proxy to
	RunnerStarted(context.Context, *connect.Request[v11.RunnerStartedRequest]) (*connect.Response[v11.RunnerStartedResponse], error)
}

// NewHotReloadServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewHotReloadServiceHandler(svc HotReloadServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	hotReloadServicePingHandler := connect.NewUnaryHandler(
		HotReloadServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	hotReloadServiceRunnerStartedHandler := connect.NewUnaryHandler(
		HotReloadServiceRunnerStartedProcedure,
		svc.RunnerStarted,
		opts...,
	)
	return "/xyz.block.ftl.language.v1.HotReloadService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case HotReloadServicePingProcedure:
			hotReloadServicePingHandler.ServeHTTP(w, r)
		case HotReloadServiceRunnerStartedProcedure:
			hotReloadServiceRunnerStartedHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedHotReloadServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedHotReloadServiceHandler struct{}

func (UnimplementedHotReloadServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.HotReloadService.Ping is not implemented"))
}

func (UnimplementedHotReloadServiceHandler) RunnerStarted(context.Context, *connect.Request[v11.RunnerStartedRequest]) (*connect.Response[v11.RunnerStartedResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.HotReloadService.RunnerStarted is not implemented"))
}
