// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/language/language.proto

package languagepbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	language "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/language"
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
	LanguageServiceName = "xyz.block.ftl.v1.language.LanguageService"
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
	LanguageServicePingProcedure = "/xyz.block.ftl.v1.language.LanguageService/Ping"
	// LanguageServiceGetCreateModuleFlagsProcedure is the fully-qualified name of the LanguageService's
	// GetCreateModuleFlags RPC.
	LanguageServiceGetCreateModuleFlagsProcedure = "/xyz.block.ftl.v1.language.LanguageService/GetCreateModuleFlags"
	// LanguageServiceCreateModuleProcedure is the fully-qualified name of the LanguageService's
	// CreateModule RPC.
	LanguageServiceCreateModuleProcedure = "/xyz.block.ftl.v1.language.LanguageService/CreateModule"
	// LanguageServiceModuleConfigDefaultsProcedure is the fully-qualified name of the LanguageService's
	// ModuleConfigDefaults RPC.
	LanguageServiceModuleConfigDefaultsProcedure = "/xyz.block.ftl.v1.language.LanguageService/ModuleConfigDefaults"
	// LanguageServiceGetDependenciesProcedure is the fully-qualified name of the LanguageService's
	// GetDependencies RPC.
	LanguageServiceGetDependenciesProcedure = "/xyz.block.ftl.v1.language.LanguageService/GetDependencies"
	// LanguageServiceBuildProcedure is the fully-qualified name of the LanguageService's Build RPC.
	LanguageServiceBuildProcedure = "/xyz.block.ftl.v1.language.LanguageService/Build"
	// LanguageServiceBuildContextUpdatedProcedure is the fully-qualified name of the LanguageService's
	// BuildContextUpdated RPC.
	LanguageServiceBuildContextUpdatedProcedure = "/xyz.block.ftl.v1.language.LanguageService/BuildContextUpdated"
	// LanguageServiceGenerateStubsProcedure is the fully-qualified name of the LanguageService's
	// GenerateStubs RPC.
	LanguageServiceGenerateStubsProcedure = "/xyz.block.ftl.v1.language.LanguageService/GenerateStubs"
	// LanguageServiceSyncStubReferencesProcedure is the fully-qualified name of the LanguageService's
	// SyncStubReferences RPC.
	LanguageServiceSyncStubReferencesProcedure = "/xyz.block.ftl.v1.language.LanguageService/SyncStubReferences"
)

// LanguageServiceClient is a client for the xyz.block.ftl.v1.language.LanguageService service.
type LanguageServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get language specific flags that can be used to create a new module.
	GetCreateModuleFlags(context.Context, *connect.Request[language.GetCreateModuleFlagsRequest]) (*connect.Response[language.GetCreateModuleFlagsResponse], error)
	// Generates files for a new module with the requested name
	CreateModule(context.Context, *connect.Request[language.CreateModuleRequest]) (*connect.Response[language.CreateModuleResponse], error)
	// Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
	ModuleConfigDefaults(context.Context, *connect.Request[language.ModuleConfigDefaultsRequest]) (*connect.Response[language.ModuleConfigDefaultsResponse], error)
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[language.DependenciesRequest]) (*connect.Response[language.DependenciesResponse], error)
	// Build the module and stream back build events.
	//
	// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
	// end of the build.
	//
	// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
	// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
	// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
	// calls.
	Build(context.Context, *connect.Request[language.BuildRequest]) (*connect.ServerStreamForClient[language.BuildEvent], error)
	// While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
	// build context is updated.
	//
	// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
	// event with the updated build context id with "is_automatic_rebuild" as false.
	//
	// If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
	// build stream, it must fail the BuildContextUpdated call.
	BuildContextUpdated(context.Context, *connect.Request[language.BuildContextUpdatedRequest]) (*connect.Response[language.BuildContextUpdatedResponse], error)
	// Generate stubs for a module.
	//
	// Stubs allow modules to import other module's exported interface. If a language does not need this step,
	// then it is not required to do anything in this call.
	//
	// This call is not tied to the module that this plugin is responsible for. A plugin of each language will
	// be chosen to generate stubs for each module.
	GenerateStubs(context.Context, *connect.Request[language.GenerateStubsRequest]) (*connect.Response[language.GenerateStubsResponse], error)
	// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
	// references to external modules, regardless of whether they are dependencies.
	//
	// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
	// import the modules when users start reference them.
	//
	// It is optional to do anything with this call.
	SyncStubReferences(context.Context, *connect.Request[language.SyncStubReferencesRequest]) (*connect.Response[language.SyncStubReferencesResponse], error)
}

// NewLanguageServiceClient constructs a client for the xyz.block.ftl.v1.language.LanguageService
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
		getCreateModuleFlags: connect.NewClient[language.GetCreateModuleFlagsRequest, language.GetCreateModuleFlagsResponse](
			httpClient,
			baseURL+LanguageServiceGetCreateModuleFlagsProcedure,
			opts...,
		),
		createModule: connect.NewClient[language.CreateModuleRequest, language.CreateModuleResponse](
			httpClient,
			baseURL+LanguageServiceCreateModuleProcedure,
			opts...,
		),
		moduleConfigDefaults: connect.NewClient[language.ModuleConfigDefaultsRequest, language.ModuleConfigDefaultsResponse](
			httpClient,
			baseURL+LanguageServiceModuleConfigDefaultsProcedure,
			opts...,
		),
		getDependencies: connect.NewClient[language.DependenciesRequest, language.DependenciesResponse](
			httpClient,
			baseURL+LanguageServiceGetDependenciesProcedure,
			opts...,
		),
		build: connect.NewClient[language.BuildRequest, language.BuildEvent](
			httpClient,
			baseURL+LanguageServiceBuildProcedure,
			opts...,
		),
		buildContextUpdated: connect.NewClient[language.BuildContextUpdatedRequest, language.BuildContextUpdatedResponse](
			httpClient,
			baseURL+LanguageServiceBuildContextUpdatedProcedure,
			opts...,
		),
		generateStubs: connect.NewClient[language.GenerateStubsRequest, language.GenerateStubsResponse](
			httpClient,
			baseURL+LanguageServiceGenerateStubsProcedure,
			opts...,
		),
		syncStubReferences: connect.NewClient[language.SyncStubReferencesRequest, language.SyncStubReferencesResponse](
			httpClient,
			baseURL+LanguageServiceSyncStubReferencesProcedure,
			opts...,
		),
	}
}

// languageServiceClient implements LanguageServiceClient.
type languageServiceClient struct {
	ping                 *connect.Client[v1.PingRequest, v1.PingResponse]
	getCreateModuleFlags *connect.Client[language.GetCreateModuleFlagsRequest, language.GetCreateModuleFlagsResponse]
	createModule         *connect.Client[language.CreateModuleRequest, language.CreateModuleResponse]
	moduleConfigDefaults *connect.Client[language.ModuleConfigDefaultsRequest, language.ModuleConfigDefaultsResponse]
	getDependencies      *connect.Client[language.DependenciesRequest, language.DependenciesResponse]
	build                *connect.Client[language.BuildRequest, language.BuildEvent]
	buildContextUpdated  *connect.Client[language.BuildContextUpdatedRequest, language.BuildContextUpdatedResponse]
	generateStubs        *connect.Client[language.GenerateStubsRequest, language.GenerateStubsResponse]
	syncStubReferences   *connect.Client[language.SyncStubReferencesRequest, language.SyncStubReferencesResponse]
}

// Ping calls xyz.block.ftl.v1.language.LanguageService.Ping.
func (c *languageServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetCreateModuleFlags calls xyz.block.ftl.v1.language.LanguageService.GetCreateModuleFlags.
func (c *languageServiceClient) GetCreateModuleFlags(ctx context.Context, req *connect.Request[language.GetCreateModuleFlagsRequest]) (*connect.Response[language.GetCreateModuleFlagsResponse], error) {
	return c.getCreateModuleFlags.CallUnary(ctx, req)
}

// CreateModule calls xyz.block.ftl.v1.language.LanguageService.CreateModule.
func (c *languageServiceClient) CreateModule(ctx context.Context, req *connect.Request[language.CreateModuleRequest]) (*connect.Response[language.CreateModuleResponse], error) {
	return c.createModule.CallUnary(ctx, req)
}

// ModuleConfigDefaults calls xyz.block.ftl.v1.language.LanguageService.ModuleConfigDefaults.
func (c *languageServiceClient) ModuleConfigDefaults(ctx context.Context, req *connect.Request[language.ModuleConfigDefaultsRequest]) (*connect.Response[language.ModuleConfigDefaultsResponse], error) {
	return c.moduleConfigDefaults.CallUnary(ctx, req)
}

// GetDependencies calls xyz.block.ftl.v1.language.LanguageService.GetDependencies.
func (c *languageServiceClient) GetDependencies(ctx context.Context, req *connect.Request[language.DependenciesRequest]) (*connect.Response[language.DependenciesResponse], error) {
	return c.getDependencies.CallUnary(ctx, req)
}

// Build calls xyz.block.ftl.v1.language.LanguageService.Build.
func (c *languageServiceClient) Build(ctx context.Context, req *connect.Request[language.BuildRequest]) (*connect.ServerStreamForClient[language.BuildEvent], error) {
	return c.build.CallServerStream(ctx, req)
}

// BuildContextUpdated calls xyz.block.ftl.v1.language.LanguageService.BuildContextUpdated.
func (c *languageServiceClient) BuildContextUpdated(ctx context.Context, req *connect.Request[language.BuildContextUpdatedRequest]) (*connect.Response[language.BuildContextUpdatedResponse], error) {
	return c.buildContextUpdated.CallUnary(ctx, req)
}

// GenerateStubs calls xyz.block.ftl.v1.language.LanguageService.GenerateStubs.
func (c *languageServiceClient) GenerateStubs(ctx context.Context, req *connect.Request[language.GenerateStubsRequest]) (*connect.Response[language.GenerateStubsResponse], error) {
	return c.generateStubs.CallUnary(ctx, req)
}

// SyncStubReferences calls xyz.block.ftl.v1.language.LanguageService.SyncStubReferences.
func (c *languageServiceClient) SyncStubReferences(ctx context.Context, req *connect.Request[language.SyncStubReferencesRequest]) (*connect.Response[language.SyncStubReferencesResponse], error) {
	return c.syncStubReferences.CallUnary(ctx, req)
}

// LanguageServiceHandler is an implementation of the xyz.block.ftl.v1.language.LanguageService
// service.
type LanguageServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Get language specific flags that can be used to create a new module.
	GetCreateModuleFlags(context.Context, *connect.Request[language.GetCreateModuleFlagsRequest]) (*connect.Response[language.GetCreateModuleFlagsResponse], error)
	// Generates files for a new module with the requested name
	CreateModule(context.Context, *connect.Request[language.CreateModuleRequest]) (*connect.Response[language.CreateModuleResponse], error)
	// Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
	ModuleConfigDefaults(context.Context, *connect.Request[language.ModuleConfigDefaultsRequest]) (*connect.Response[language.ModuleConfigDefaultsResponse], error)
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[language.DependenciesRequest]) (*connect.Response[language.DependenciesResponse], error)
	// Build the module and stream back build events.
	//
	// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
	// end of the build.
	//
	// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
	// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
	// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
	// calls.
	Build(context.Context, *connect.Request[language.BuildRequest], *connect.ServerStream[language.BuildEvent]) error
	// While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
	// build context is updated.
	//
	// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
	// event with the updated build context id with "is_automatic_rebuild" as false.
	//
	// If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
	// build stream, it must fail the BuildContextUpdated call.
	BuildContextUpdated(context.Context, *connect.Request[language.BuildContextUpdatedRequest]) (*connect.Response[language.BuildContextUpdatedResponse], error)
	// Generate stubs for a module.
	//
	// Stubs allow modules to import other module's exported interface. If a language does not need this step,
	// then it is not required to do anything in this call.
	//
	// This call is not tied to the module that this plugin is responsible for. A plugin of each language will
	// be chosen to generate stubs for each module.
	GenerateStubs(context.Context, *connect.Request[language.GenerateStubsRequest]) (*connect.Response[language.GenerateStubsResponse], error)
	// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
	// references to external modules, regardless of whether they are dependencies.
	//
	// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
	// import the modules when users start reference them.
	//
	// It is optional to do anything with this call.
	SyncStubReferences(context.Context, *connect.Request[language.SyncStubReferencesRequest]) (*connect.Response[language.SyncStubReferencesResponse], error)
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
	return "/xyz.block.ftl.v1.language.LanguageService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.Ping is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GetCreateModuleFlags(context.Context, *connect.Request[language.GetCreateModuleFlagsRequest]) (*connect.Response[language.GetCreateModuleFlagsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.GetCreateModuleFlags is not implemented"))
}

func (UnimplementedLanguageServiceHandler) CreateModule(context.Context, *connect.Request[language.CreateModuleRequest]) (*connect.Response[language.CreateModuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.CreateModule is not implemented"))
}

func (UnimplementedLanguageServiceHandler) ModuleConfigDefaults(context.Context, *connect.Request[language.ModuleConfigDefaultsRequest]) (*connect.Response[language.ModuleConfigDefaultsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.ModuleConfigDefaults is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GetDependencies(context.Context, *connect.Request[language.DependenciesRequest]) (*connect.Response[language.DependenciesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.GetDependencies is not implemented"))
}

func (UnimplementedLanguageServiceHandler) Build(context.Context, *connect.Request[language.BuildRequest], *connect.ServerStream[language.BuildEvent]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.Build is not implemented"))
}

func (UnimplementedLanguageServiceHandler) BuildContextUpdated(context.Context, *connect.Request[language.BuildContextUpdatedRequest]) (*connect.Response[language.BuildContextUpdatedResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.BuildContextUpdated is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GenerateStubs(context.Context, *connect.Request[language.GenerateStubsRequest]) (*connect.Response[language.GenerateStubsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.GenerateStubs is not implemented"))
}

func (UnimplementedLanguageServiceHandler) SyncStubReferences(context.Context, *connect.Request[language.SyncStubReferencesRequest]) (*connect.Response[language.SyncStubReferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.language.LanguageService.SyncStubReferences is not implemented"))
}
