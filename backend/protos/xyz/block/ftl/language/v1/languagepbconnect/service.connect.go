// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/language/v1/service.proto

package languagepbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// LanguageServiceName is the fully-qualified name of the LanguageService service.
	LanguageServiceName = "xyz.block.ftl.language.v1.LanguageService"
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
	// LanguageServiceGetDependenciesProcedure is the fully-qualified name of the LanguageService's
	// GetDependencies RPC.
	LanguageServiceGetDependenciesProcedure = "/xyz.block.ftl.language.v1.LanguageService/GetDependencies"
	// LanguageServiceBuildProcedure is the fully-qualified name of the LanguageService's Build RPC.
	LanguageServiceBuildProcedure = "/xyz.block.ftl.language.v1.LanguageService/Build"
	// LanguageServiceGenerateStubsProcedure is the fully-qualified name of the LanguageService's
	// GenerateStubs RPC.
	LanguageServiceGenerateStubsProcedure = "/xyz.block.ftl.language.v1.LanguageService/GenerateStubs"
	// LanguageServiceSyncStubReferencesProcedure is the fully-qualified name of the LanguageService's
	// SyncStubReferences RPC.
	LanguageServiceSyncStubReferencesProcedure = "/xyz.block.ftl.language.v1.LanguageService/SyncStubReferences"
)

// LanguageServiceClient is a client for the xyz.block.ftl.language.v1.LanguageService service.
type LanguageServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error)
	// Build the module and receive the result
	Build(context.Context, *connect.Request[v11.BuildRequest]) (*connect.Response[v11.BuildResponse], error)
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
	languageServiceMethods := v11.File_xyz_block_ftl_language_v1_service_proto.Services().ByName("LanguageService").Methods()
	return &languageServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+LanguageServicePingProcedure,
			connect.WithSchema(languageServiceMethods.ByName("Ping")),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		getDependencies: connect.NewClient[v11.GetDependenciesRequest, v11.GetDependenciesResponse](
			httpClient,
			baseURL+LanguageServiceGetDependenciesProcedure,
			connect.WithSchema(languageServiceMethods.ByName("GetDependencies")),
			connect.WithClientOptions(opts...),
		),
		build: connect.NewClient[v11.BuildRequest, v11.BuildResponse](
			httpClient,
			baseURL+LanguageServiceBuildProcedure,
			connect.WithSchema(languageServiceMethods.ByName("Build")),
			connect.WithClientOptions(opts...),
		),
		generateStubs: connect.NewClient[v11.GenerateStubsRequest, v11.GenerateStubsResponse](
			httpClient,
			baseURL+LanguageServiceGenerateStubsProcedure,
			connect.WithSchema(languageServiceMethods.ByName("GenerateStubs")),
			connect.WithClientOptions(opts...),
		),
		syncStubReferences: connect.NewClient[v11.SyncStubReferencesRequest, v11.SyncStubReferencesResponse](
			httpClient,
			baseURL+LanguageServiceSyncStubReferencesProcedure,
			connect.WithSchema(languageServiceMethods.ByName("SyncStubReferences")),
			connect.WithClientOptions(opts...),
		),
	}
}

// languageServiceClient implements LanguageServiceClient.
type languageServiceClient struct {
	ping               *connect.Client[v1.PingRequest, v1.PingResponse]
	getDependencies    *connect.Client[v11.GetDependenciesRequest, v11.GetDependenciesResponse]
	build              *connect.Client[v11.BuildRequest, v11.BuildResponse]
	generateStubs      *connect.Client[v11.GenerateStubsRequest, v11.GenerateStubsResponse]
	syncStubReferences *connect.Client[v11.SyncStubReferencesRequest, v11.SyncStubReferencesResponse]
}

// Ping calls xyz.block.ftl.language.v1.LanguageService.Ping.
func (c *languageServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetDependencies calls xyz.block.ftl.language.v1.LanguageService.GetDependencies.
func (c *languageServiceClient) GetDependencies(ctx context.Context, req *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error) {
	return c.getDependencies.CallUnary(ctx, req)
}

// Build calls xyz.block.ftl.language.v1.LanguageService.Build.
func (c *languageServiceClient) Build(ctx context.Context, req *connect.Request[v11.BuildRequest]) (*connect.Response[v11.BuildResponse], error) {
	return c.build.CallUnary(ctx, req)
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
	// Extract dependencies for a module
	// FTL will ensure that these dependencies are built before requesting a build for this module.
	GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error)
	// Build the module and receive the result
	Build(context.Context, *connect.Request[v11.BuildRequest]) (*connect.Response[v11.BuildResponse], error)
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
	languageServiceMethods := v11.File_xyz_block_ftl_language_v1_service_proto.Services().ByName("LanguageService").Methods()
	languageServicePingHandler := connect.NewUnaryHandler(
		LanguageServicePingProcedure,
		svc.Ping,
		connect.WithSchema(languageServiceMethods.ByName("Ping")),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	languageServiceGetDependenciesHandler := connect.NewUnaryHandler(
		LanguageServiceGetDependenciesProcedure,
		svc.GetDependencies,
		connect.WithSchema(languageServiceMethods.ByName("GetDependencies")),
		connect.WithHandlerOptions(opts...),
	)
	languageServiceBuildHandler := connect.NewUnaryHandler(
		LanguageServiceBuildProcedure,
		svc.Build,
		connect.WithSchema(languageServiceMethods.ByName("Build")),
		connect.WithHandlerOptions(opts...),
	)
	languageServiceGenerateStubsHandler := connect.NewUnaryHandler(
		LanguageServiceGenerateStubsProcedure,
		svc.GenerateStubs,
		connect.WithSchema(languageServiceMethods.ByName("GenerateStubs")),
		connect.WithHandlerOptions(opts...),
	)
	languageServiceSyncStubReferencesHandler := connect.NewUnaryHandler(
		LanguageServiceSyncStubReferencesProcedure,
		svc.SyncStubReferences,
		connect.WithSchema(languageServiceMethods.ByName("SyncStubReferences")),
		connect.WithHandlerOptions(opts...),
	)
	return "/xyz.block.ftl.language.v1.LanguageService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LanguageServicePingProcedure:
			languageServicePingHandler.ServeHTTP(w, r)
		case LanguageServiceGetDependenciesProcedure:
			languageServiceGetDependenciesHandler.ServeHTTP(w, r)
		case LanguageServiceBuildProcedure:
			languageServiceBuildHandler.ServeHTTP(w, r)
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

func (UnimplementedLanguageServiceHandler) GetDependencies(context.Context, *connect.Request[v11.GetDependenciesRequest]) (*connect.Response[v11.GetDependenciesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.GetDependencies is not implemented"))
}

func (UnimplementedLanguageServiceHandler) Build(context.Context, *connect.Request[v11.BuildRequest]) (*connect.Response[v11.BuildResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.Build is not implemented"))
}

func (UnimplementedLanguageServiceHandler) GenerateStubs(context.Context, *connect.Request[v11.GenerateStubsRequest]) (*connect.Response[v11.GenerateStubsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.GenerateStubs is not implemented"))
}

func (UnimplementedLanguageServiceHandler) SyncStubReferences(context.Context, *connect.Request[v11.SyncStubReferencesRequest]) (*connect.Response[v11.SyncStubReferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.language.v1.LanguageService.SyncStubReferences is not implemented"))
}
