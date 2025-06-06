// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/hotreload/v1/hotreload.proto

package hotreloadpbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1"
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
	// HotReloadServiceName is the fully-qualified name of the HotReloadService service.
	HotReloadServiceName = "xyz.block.ftl.hotreload.v1.HotReloadService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// HotReloadServicePingProcedure is the fully-qualified name of the HotReloadService's Ping RPC.
	HotReloadServicePingProcedure = "/xyz.block.ftl.hotreload.v1.HotReloadService/Ping"
	// HotReloadServiceReloadProcedure is the fully-qualified name of the HotReloadService's Reload RPC.
	HotReloadServiceReloadProcedure = "/xyz.block.ftl.hotreload.v1.HotReloadService/Reload"
	// HotReloadServiceWatchProcedure is the fully-qualified name of the HotReloadService's Watch RPC.
	HotReloadServiceWatchProcedure = "/xyz.block.ftl.hotreload.v1.HotReloadService/Watch"
	// HotReloadServiceRunnerInfoProcedure is the fully-qualified name of the HotReloadService's
	// RunnerInfo RPC.
	HotReloadServiceRunnerInfoProcedure = "/xyz.block.ftl.hotreload.v1.HotReloadService/RunnerInfo"
)

// HotReloadServiceClient is a client for the xyz.block.ftl.hotreload.v1.HotReloadService service.
type HotReloadServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Forces an explicit Reload from the plugin. This is useful for when the plugin needs to trigger a Reload,
	// such as when the Reload context changes.
	Reload(context.Context, *connect.Request[v11.ReloadRequest]) (*connect.Response[v11.ReloadResponse], error)
	// Watch for a reload not initiated by an explicit Reload call.
	// This is generally used to get the initial state of the runner.
	Watch(context.Context, *connect.Request[v11.WatchRequest]) (*connect.Response[v11.WatchResponse], error)
	// Invoked by the runner to provide runner information to the plugin.
	RunnerInfo(context.Context, *connect.Request[v11.RunnerInfoRequest]) (*connect.Response[v11.RunnerInfoResponse], error)
}

// NewHotReloadServiceClient constructs a client for the xyz.block.ftl.hotreload.v1.HotReloadService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewHotReloadServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) HotReloadServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	hotReloadServiceMethods := v11.File_xyz_block_ftl_hotreload_v1_hotreload_proto.Services().ByName("HotReloadService").Methods()
	return &hotReloadServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+HotReloadServicePingProcedure,
			connect.WithSchema(hotReloadServiceMethods.ByName("Ping")),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		reload: connect.NewClient[v11.ReloadRequest, v11.ReloadResponse](
			httpClient,
			baseURL+HotReloadServiceReloadProcedure,
			connect.WithSchema(hotReloadServiceMethods.ByName("Reload")),
			connect.WithClientOptions(opts...),
		),
		watch: connect.NewClient[v11.WatchRequest, v11.WatchResponse](
			httpClient,
			baseURL+HotReloadServiceWatchProcedure,
			connect.WithSchema(hotReloadServiceMethods.ByName("Watch")),
			connect.WithClientOptions(opts...),
		),
		runnerInfo: connect.NewClient[v11.RunnerInfoRequest, v11.RunnerInfoResponse](
			httpClient,
			baseURL+HotReloadServiceRunnerInfoProcedure,
			connect.WithSchema(hotReloadServiceMethods.ByName("RunnerInfo")),
			connect.WithClientOptions(opts...),
		),
	}
}

// hotReloadServiceClient implements HotReloadServiceClient.
type hotReloadServiceClient struct {
	ping       *connect.Client[v1.PingRequest, v1.PingResponse]
	reload     *connect.Client[v11.ReloadRequest, v11.ReloadResponse]
	watch      *connect.Client[v11.WatchRequest, v11.WatchResponse]
	runnerInfo *connect.Client[v11.RunnerInfoRequest, v11.RunnerInfoResponse]
}

// Ping calls xyz.block.ftl.hotreload.v1.HotReloadService.Ping.
func (c *hotReloadServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Reload calls xyz.block.ftl.hotreload.v1.HotReloadService.Reload.
func (c *hotReloadServiceClient) Reload(ctx context.Context, req *connect.Request[v11.ReloadRequest]) (*connect.Response[v11.ReloadResponse], error) {
	return c.reload.CallUnary(ctx, req)
}

// Watch calls xyz.block.ftl.hotreload.v1.HotReloadService.Watch.
func (c *hotReloadServiceClient) Watch(ctx context.Context, req *connect.Request[v11.WatchRequest]) (*connect.Response[v11.WatchResponse], error) {
	return c.watch.CallUnary(ctx, req)
}

// RunnerInfo calls xyz.block.ftl.hotreload.v1.HotReloadService.RunnerInfo.
func (c *hotReloadServiceClient) RunnerInfo(ctx context.Context, req *connect.Request[v11.RunnerInfoRequest]) (*connect.Response[v11.RunnerInfoResponse], error) {
	return c.runnerInfo.CallUnary(ctx, req)
}

// HotReloadServiceHandler is an implementation of the xyz.block.ftl.hotreload.v1.HotReloadService
// service.
type HotReloadServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Forces an explicit Reload from the plugin. This is useful for when the plugin needs to trigger a Reload,
	// such as when the Reload context changes.
	Reload(context.Context, *connect.Request[v11.ReloadRequest]) (*connect.Response[v11.ReloadResponse], error)
	// Watch for a reload not initiated by an explicit Reload call.
	// This is generally used to get the initial state of the runner.
	Watch(context.Context, *connect.Request[v11.WatchRequest]) (*connect.Response[v11.WatchResponse], error)
	// Invoked by the runner to provide runner information to the plugin.
	RunnerInfo(context.Context, *connect.Request[v11.RunnerInfoRequest]) (*connect.Response[v11.RunnerInfoResponse], error)
}

// NewHotReloadServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewHotReloadServiceHandler(svc HotReloadServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	hotReloadServiceMethods := v11.File_xyz_block_ftl_hotreload_v1_hotreload_proto.Services().ByName("HotReloadService").Methods()
	hotReloadServicePingHandler := connect.NewUnaryHandler(
		HotReloadServicePingProcedure,
		svc.Ping,
		connect.WithSchema(hotReloadServiceMethods.ByName("Ping")),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	hotReloadServiceReloadHandler := connect.NewUnaryHandler(
		HotReloadServiceReloadProcedure,
		svc.Reload,
		connect.WithSchema(hotReloadServiceMethods.ByName("Reload")),
		connect.WithHandlerOptions(opts...),
	)
	hotReloadServiceWatchHandler := connect.NewUnaryHandler(
		HotReloadServiceWatchProcedure,
		svc.Watch,
		connect.WithSchema(hotReloadServiceMethods.ByName("Watch")),
		connect.WithHandlerOptions(opts...),
	)
	hotReloadServiceRunnerInfoHandler := connect.NewUnaryHandler(
		HotReloadServiceRunnerInfoProcedure,
		svc.RunnerInfo,
		connect.WithSchema(hotReloadServiceMethods.ByName("RunnerInfo")),
		connect.WithHandlerOptions(opts...),
	)
	return "/xyz.block.ftl.hotreload.v1.HotReloadService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case HotReloadServicePingProcedure:
			hotReloadServicePingHandler.ServeHTTP(w, r)
		case HotReloadServiceReloadProcedure:
			hotReloadServiceReloadHandler.ServeHTTP(w, r)
		case HotReloadServiceWatchProcedure:
			hotReloadServiceWatchHandler.ServeHTTP(w, r)
		case HotReloadServiceRunnerInfoProcedure:
			hotReloadServiceRunnerInfoHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedHotReloadServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedHotReloadServiceHandler struct{}

func (UnimplementedHotReloadServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.hotreload.v1.HotReloadService.Ping is not implemented"))
}

func (UnimplementedHotReloadServiceHandler) Reload(context.Context, *connect.Request[v11.ReloadRequest]) (*connect.Response[v11.ReloadResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.hotreload.v1.HotReloadService.Reload is not implemented"))
}

func (UnimplementedHotReloadServiceHandler) Watch(context.Context, *connect.Request[v11.WatchRequest]) (*connect.Response[v11.WatchResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.hotreload.v1.HotReloadService.Watch is not implemented"))
}

func (UnimplementedHotReloadServiceHandler) RunnerInfo(context.Context, *connect.Request[v11.RunnerInfoRequest]) (*connect.Response[v11.RunnerInfoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.hotreload.v1.HotReloadService.RunnerInfo is not implemented"))
}
