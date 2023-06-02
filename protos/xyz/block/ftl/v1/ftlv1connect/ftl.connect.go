// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/ftl.proto

package ftlv1connect

import (
	context "context"
	errors "errors"
	v1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	connect_go "github.com/bufbuild/connect-go"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// VerbServiceName is the fully-qualified name of the VerbService service.
	VerbServiceName = "xyz.block.ftl.v1.VerbService"
	// DevelServiceName is the fully-qualified name of the DevelService service.
	DevelServiceName = "xyz.block.ftl.v1.DevelService"
	// ControlPlaneServiceName is the fully-qualified name of the ControlPlaneService service.
	ControlPlaneServiceName = "xyz.block.ftl.v1.ControlPlaneService"
	// RunnerServiceName is the fully-qualified name of the RunnerService service.
	RunnerServiceName = "xyz.block.ftl.v1.RunnerService"
	// ObservabilityServiceName is the fully-qualified name of the ObservabilityService service.
	ObservabilityServiceName = "xyz.block.ftl.v1.ObservabilityService"
)

// VerbServiceClient is a client for the xyz.block.ftl.v1.VerbService service.
type VerbServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Issue a synchronous call to a Verb.
	Call(context.Context, *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error)
	// List the available Verbs.
	List(context.Context, *connect_go.Request[v1.ListRequest]) (*connect_go.Response[v1.ListResponse], error)
}

// NewVerbServiceClient constructs a client for the xyz.block.ftl.v1.VerbService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewVerbServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) VerbServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &verbServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.VerbService/Ping",
			opts...,
		),
		call: connect_go.NewClient[v1.CallRequest, v1.CallResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.VerbService/Call",
			opts...,
		),
		list: connect_go.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.VerbService/List",
			opts...,
		),
	}
}

// verbServiceClient implements VerbServiceClient.
type verbServiceClient struct {
	ping *connect_go.Client[v1.PingRequest, v1.PingResponse]
	call *connect_go.Client[v1.CallRequest, v1.CallResponse]
	list *connect_go.Client[v1.ListRequest, v1.ListResponse]
}

// Ping calls xyz.block.ftl.v1.VerbService.Ping.
func (c *verbServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Call calls xyz.block.ftl.v1.VerbService.Call.
func (c *verbServiceClient) Call(ctx context.Context, req *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error) {
	return c.call.CallUnary(ctx, req)
}

// List calls xyz.block.ftl.v1.VerbService.List.
func (c *verbServiceClient) List(ctx context.Context, req *connect_go.Request[v1.ListRequest]) (*connect_go.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// VerbServiceHandler is an implementation of the xyz.block.ftl.v1.VerbService service.
type VerbServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Issue a synchronous call to a Verb.
	Call(context.Context, *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error)
	// List the available Verbs.
	List(context.Context, *connect_go.Request[v1.ListRequest]) (*connect_go.Response[v1.ListResponse], error)
}

// NewVerbServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewVerbServiceHandler(svc VerbServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/xyz.block.ftl.v1.VerbService/Ping", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.VerbService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.VerbService/Call", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.VerbService/Call",
		svc.Call,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.VerbService/List", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.VerbService/List",
		svc.List,
		opts...,
	))
	return "/xyz.block.ftl.v1.VerbService/", mux
}

// UnimplementedVerbServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedVerbServiceHandler struct{}

func (UnimplementedVerbServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.VerbService.Ping is not implemented"))
}

func (UnimplementedVerbServiceHandler) Call(context.Context, *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.VerbService.Call is not implemented"))
}

func (UnimplementedVerbServiceHandler) List(context.Context, *connect_go.Request[v1.ListRequest]) (*connect_go.Response[v1.ListResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.VerbService.List is not implemented"))
}

// DevelServiceClient is a client for the xyz.block.ftl.v1.DevelService service.
type DevelServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Push schema changes to the server.
	PushSchema(context.Context) *connect_go.ClientStreamForClient[v1.PushSchemaRequest, v1.PushSchemaResponse]
	// Pull schema changes from the server.
	PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest]) (*connect_go.ServerStreamForClient[v1.PullSchemaResponse], error)
}

// NewDevelServiceClient constructs a client for the xyz.block.ftl.v1.DevelService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewDevelServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) DevelServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &develServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.DevelService/Ping",
			opts...,
		),
		pushSchema: connect_go.NewClient[v1.PushSchemaRequest, v1.PushSchemaResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.DevelService/PushSchema",
			opts...,
		),
		pullSchema: connect_go.NewClient[v1.PullSchemaRequest, v1.PullSchemaResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.DevelService/PullSchema",
			opts...,
		),
	}
}

// develServiceClient implements DevelServiceClient.
type develServiceClient struct {
	ping       *connect_go.Client[v1.PingRequest, v1.PingResponse]
	pushSchema *connect_go.Client[v1.PushSchemaRequest, v1.PushSchemaResponse]
	pullSchema *connect_go.Client[v1.PullSchemaRequest, v1.PullSchemaResponse]
}

// Ping calls xyz.block.ftl.v1.DevelService.Ping.
func (c *develServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// PushSchema calls xyz.block.ftl.v1.DevelService.PushSchema.
func (c *develServiceClient) PushSchema(ctx context.Context) *connect_go.ClientStreamForClient[v1.PushSchemaRequest, v1.PushSchemaResponse] {
	return c.pushSchema.CallClientStream(ctx)
}

// PullSchema calls xyz.block.ftl.v1.DevelService.PullSchema.
func (c *develServiceClient) PullSchema(ctx context.Context, req *connect_go.Request[v1.PullSchemaRequest]) (*connect_go.ServerStreamForClient[v1.PullSchemaResponse], error) {
	return c.pullSchema.CallServerStream(ctx, req)
}

// DevelServiceHandler is an implementation of the xyz.block.ftl.v1.DevelService service.
type DevelServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Push schema changes to the server.
	PushSchema(context.Context, *connect_go.ClientStream[v1.PushSchemaRequest]) (*connect_go.Response[v1.PushSchemaResponse], error)
	// Pull schema changes from the server.
	PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest], *connect_go.ServerStream[v1.PullSchemaResponse]) error
}

// NewDevelServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewDevelServiceHandler(svc DevelServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/xyz.block.ftl.v1.DevelService/Ping", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.DevelService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.DevelService/PushSchema", connect_go.NewClientStreamHandler(
		"/xyz.block.ftl.v1.DevelService/PushSchema",
		svc.PushSchema,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.DevelService/PullSchema", connect_go.NewServerStreamHandler(
		"/xyz.block.ftl.v1.DevelService/PullSchema",
		svc.PullSchema,
		opts...,
	))
	return "/xyz.block.ftl.v1.DevelService/", mux
}

// UnimplementedDevelServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedDevelServiceHandler struct{}

func (UnimplementedDevelServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.DevelService.Ping is not implemented"))
}

func (UnimplementedDevelServiceHandler) PushSchema(context.Context, *connect_go.ClientStream[v1.PushSchemaRequest]) (*connect_go.Response[v1.PushSchemaResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.DevelService.PushSchema is not implemented"))
}

func (UnimplementedDevelServiceHandler) PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest], *connect_go.ServerStream[v1.PullSchemaResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.DevelService.PullSchema is not implemented"))
}

// ControlPlaneServiceClient is a client for the xyz.block.ftl.v1.ControlPlaneService service.
type ControlPlaneServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error)
	// Upload an artefact to the server.
	UploadArtefact(context.Context, *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error)
	// Create a deployment.
	CreateDeployment(context.Context, *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error)
	// Get the schema and artefact metadata for a deployment.
	GetDeployment(context.Context, *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect_go.Request[v1.GetDeploymentArtefactsRequest]) (*connect_go.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error)
	// Register a Runner with the ControlPlane.
	//
	// Each runner MUST stream a RegisterRunnerRequest to the ControlPlaneService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context) *connect_go.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	// Starts a deployment.
	Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error)
	// Stream logs from a deployment
	StreamDeploymentLogs(context.Context) *connect_go.ClientStreamForClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse]
}

// NewControlPlaneServiceClient constructs a client for the xyz.block.ftl.v1.ControlPlaneService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewControlPlaneServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ControlPlaneServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &controlPlaneServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/Ping",
			opts...,
		),
		getArtefactDiffs: connect_go.NewClient[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/GetArtefactDiffs",
			opts...,
		),
		uploadArtefact: connect_go.NewClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/UploadArtefact",
			opts...,
		),
		createDeployment: connect_go.NewClient[v1.CreateDeploymentRequest, v1.CreateDeploymentResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/CreateDeployment",
			opts...,
		),
		getDeployment: connect_go.NewClient[v1.GetDeploymentRequest, v1.GetDeploymentResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/GetDeployment",
			opts...,
		),
		getDeploymentArtefacts: connect_go.NewClient[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/GetDeploymentArtefacts",
			opts...,
		),
		registerRunner: connect_go.NewClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/RegisterRunner",
			opts...,
		),
		deploy: connect_go.NewClient[v1.DeployRequest, v1.DeployResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/Deploy",
			opts...,
		),
		streamDeploymentLogs: connect_go.NewClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ControlPlaneService/StreamDeploymentLogs",
			opts...,
		),
	}
}

// controlPlaneServiceClient implements ControlPlaneServiceClient.
type controlPlaneServiceClient struct {
	ping                   *connect_go.Client[v1.PingRequest, v1.PingResponse]
	getArtefactDiffs       *connect_go.Client[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse]
	uploadArtefact         *connect_go.Client[v1.UploadArtefactRequest, v1.UploadArtefactResponse]
	createDeployment       *connect_go.Client[v1.CreateDeploymentRequest, v1.CreateDeploymentResponse]
	getDeployment          *connect_go.Client[v1.GetDeploymentRequest, v1.GetDeploymentResponse]
	getDeploymentArtefacts *connect_go.Client[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse]
	registerRunner         *connect_go.Client[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	deploy                 *connect_go.Client[v1.DeployRequest, v1.DeployResponse]
	streamDeploymentLogs   *connect_go.Client[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse]
}

// Ping calls xyz.block.ftl.v1.ControlPlaneService.Ping.
func (c *controlPlaneServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// GetArtefactDiffs calls xyz.block.ftl.v1.ControlPlaneService.GetArtefactDiffs.
func (c *controlPlaneServiceClient) GetArtefactDiffs(ctx context.Context, req *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error) {
	return c.getArtefactDiffs.CallUnary(ctx, req)
}

// UploadArtefact calls xyz.block.ftl.v1.ControlPlaneService.UploadArtefact.
func (c *controlPlaneServiceClient) UploadArtefact(ctx context.Context, req *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error) {
	return c.uploadArtefact.CallUnary(ctx, req)
}

// CreateDeployment calls xyz.block.ftl.v1.ControlPlaneService.CreateDeployment.
func (c *controlPlaneServiceClient) CreateDeployment(ctx context.Context, req *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error) {
	return c.createDeployment.CallUnary(ctx, req)
}

// GetDeployment calls xyz.block.ftl.v1.ControlPlaneService.GetDeployment.
func (c *controlPlaneServiceClient) GetDeployment(ctx context.Context, req *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error) {
	return c.getDeployment.CallUnary(ctx, req)
}

// GetDeploymentArtefacts calls xyz.block.ftl.v1.ControlPlaneService.GetDeploymentArtefacts.
func (c *controlPlaneServiceClient) GetDeploymentArtefacts(ctx context.Context, req *connect_go.Request[v1.GetDeploymentArtefactsRequest]) (*connect_go.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error) {
	return c.getDeploymentArtefacts.CallServerStream(ctx, req)
}

// RegisterRunner calls xyz.block.ftl.v1.ControlPlaneService.RegisterRunner.
func (c *controlPlaneServiceClient) RegisterRunner(ctx context.Context) *connect_go.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse] {
	return c.registerRunner.CallClientStream(ctx)
}

// Deploy calls xyz.block.ftl.v1.ControlPlaneService.Deploy.
func (c *controlPlaneServiceClient) Deploy(ctx context.Context, req *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error) {
	return c.deploy.CallUnary(ctx, req)
}

// StreamDeploymentLogs calls xyz.block.ftl.v1.ControlPlaneService.StreamDeploymentLogs.
func (c *controlPlaneServiceClient) StreamDeploymentLogs(ctx context.Context) *connect_go.ClientStreamForClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse] {
	return c.streamDeploymentLogs.CallClientStream(ctx)
}

// ControlPlaneServiceHandler is an implementation of the xyz.block.ftl.v1.ControlPlaneService
// service.
type ControlPlaneServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error)
	// Upload an artefact to the server.
	UploadArtefact(context.Context, *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error)
	// Create a deployment.
	CreateDeployment(context.Context, *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error)
	// Get the schema and artefact metadata for a deployment.
	GetDeployment(context.Context, *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect_go.Request[v1.GetDeploymentArtefactsRequest], *connect_go.ServerStream[v1.GetDeploymentArtefactsResponse]) error
	// Register a Runner with the ControlPlane.
	//
	// Each runner MUST stream a RegisterRunnerRequest to the ControlPlaneService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context, *connect_go.ClientStream[v1.RegisterRunnerRequest]) (*connect_go.Response[v1.RegisterRunnerResponse], error)
	// Starts a deployment.
	Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error)
	// Stream logs from a deployment
	StreamDeploymentLogs(context.Context, *connect_go.ClientStream[v1.StreamDeploymentLogsRequest]) (*connect_go.Response[v1.StreamDeploymentLogsResponse], error)
}

// NewControlPlaneServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewControlPlaneServiceHandler(svc ControlPlaneServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/Ping", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/GetArtefactDiffs", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/GetArtefactDiffs",
		svc.GetArtefactDiffs,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/UploadArtefact", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/UploadArtefact",
		svc.UploadArtefact,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/CreateDeployment", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/CreateDeployment",
		svc.CreateDeployment,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/GetDeployment", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/GetDeployment",
		svc.GetDeployment,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/GetDeploymentArtefacts", connect_go.NewServerStreamHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/GetDeploymentArtefacts",
		svc.GetDeploymentArtefacts,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/RegisterRunner", connect_go.NewClientStreamHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/RegisterRunner",
		svc.RegisterRunner,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/Deploy", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/Deploy",
		svc.Deploy,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ControlPlaneService/StreamDeploymentLogs", connect_go.NewClientStreamHandler(
		"/xyz.block.ftl.v1.ControlPlaneService/StreamDeploymentLogs",
		svc.StreamDeploymentLogs,
		opts...,
	))
	return "/xyz.block.ftl.v1.ControlPlaneService/", mux
}

// UnimplementedControlPlaneServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedControlPlaneServiceHandler struct{}

func (UnimplementedControlPlaneServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.Ping is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) GetArtefactDiffs(context.Context, *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.GetArtefactDiffs is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) UploadArtefact(context.Context, *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.UploadArtefact is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) CreateDeployment(context.Context, *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.CreateDeployment is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) GetDeployment(context.Context, *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.GetDeployment is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) GetDeploymentArtefacts(context.Context, *connect_go.Request[v1.GetDeploymentArtefactsRequest], *connect_go.ServerStream[v1.GetDeploymentArtefactsResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.GetDeploymentArtefacts is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) RegisterRunner(context.Context, *connect_go.ClientStream[v1.RegisterRunnerRequest]) (*connect_go.Response[v1.RegisterRunnerResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.RegisterRunner is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.Deploy is not implemented"))
}

func (UnimplementedControlPlaneServiceHandler) StreamDeploymentLogs(context.Context, *connect_go.ClientStream[v1.StreamDeploymentLogsRequest]) (*connect_go.Response[v1.StreamDeploymentLogsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControlPlaneService.StreamDeploymentLogs is not implemented"))
}

// RunnerServiceClient is a client for the xyz.block.ftl.v1.RunnerService service.
type RunnerServiceClient interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Initiate a deployment on this Runner.
	DeployToRunner(context.Context, *connect_go.Request[v1.DeployToRunnerRequest]) (*connect_go.Response[v1.DeployToRunnerResponse], error)
}

// NewRunnerServiceClient constructs a client for the xyz.block.ftl.v1.RunnerService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewRunnerServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) RunnerServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &runnerServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.RunnerService/Ping",
			opts...,
		),
		deployToRunner: connect_go.NewClient[v1.DeployToRunnerRequest, v1.DeployToRunnerResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.RunnerService/DeployToRunner",
			opts...,
		),
	}
}

// runnerServiceClient implements RunnerServiceClient.
type runnerServiceClient struct {
	ping           *connect_go.Client[v1.PingRequest, v1.PingResponse]
	deployToRunner *connect_go.Client[v1.DeployToRunnerRequest, v1.DeployToRunnerResponse]
}

// Ping calls xyz.block.ftl.v1.RunnerService.Ping.
func (c *runnerServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// DeployToRunner calls xyz.block.ftl.v1.RunnerService.DeployToRunner.
func (c *runnerServiceClient) DeployToRunner(ctx context.Context, req *connect_go.Request[v1.DeployToRunnerRequest]) (*connect_go.Response[v1.DeployToRunnerResponse], error) {
	return c.deployToRunner.CallUnary(ctx, req)
}

// RunnerServiceHandler is an implementation of the xyz.block.ftl.v1.RunnerService service.
type RunnerServiceHandler interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Initiate a deployment on this Runner.
	DeployToRunner(context.Context, *connect_go.Request[v1.DeployToRunnerRequest]) (*connect_go.Response[v1.DeployToRunnerResponse], error)
}

// NewRunnerServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewRunnerServiceHandler(svc RunnerServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/xyz.block.ftl.v1.RunnerService/Ping", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.RunnerService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.RunnerService/DeployToRunner", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.RunnerService/DeployToRunner",
		svc.DeployToRunner,
		opts...,
	))
	return "/xyz.block.ftl.v1.RunnerService/", mux
}

// UnimplementedRunnerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedRunnerServiceHandler struct{}

func (UnimplementedRunnerServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.Ping is not implemented"))
}

func (UnimplementedRunnerServiceHandler) DeployToRunner(context.Context, *connect_go.Request[v1.DeployToRunnerRequest]) (*connect_go.Response[v1.DeployToRunnerResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.DeployToRunner is not implemented"))
}

// ObservabilityServiceClient is a client for the xyz.block.ftl.v1.ObservabilityService service.
type ObservabilityServiceClient interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Send OTEL metrics from the Deployment to the ControlPlane via the Runner.
	SendMetrics(context.Context) *connect_go.ClientStreamForClient[v1.SendMetricsRequest, v1.SendMetricsResponse]
	// Send OTEL traces from the Deployment to the ControlPlane via the Runner.
	SendTraces(context.Context) *connect_go.ClientStreamForClient[v1.SendTracesRequest, v1.SendTracesResponse]
}

// NewObservabilityServiceClient constructs a client for the xyz.block.ftl.v1.ObservabilityService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewObservabilityServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ObservabilityServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &observabilityServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ObservabilityService/Ping",
			opts...,
		),
		sendMetrics: connect_go.NewClient[v1.SendMetricsRequest, v1.SendMetricsResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ObservabilityService/SendMetrics",
			opts...,
		),
		sendTraces: connect_go.NewClient[v1.SendTracesRequest, v1.SendTracesResponse](
			httpClient,
			baseURL+"/xyz.block.ftl.v1.ObservabilityService/SendTraces",
			opts...,
		),
	}
}

// observabilityServiceClient implements ObservabilityServiceClient.
type observabilityServiceClient struct {
	ping        *connect_go.Client[v1.PingRequest, v1.PingResponse]
	sendMetrics *connect_go.Client[v1.SendMetricsRequest, v1.SendMetricsResponse]
	sendTraces  *connect_go.Client[v1.SendTracesRequest, v1.SendTracesResponse]
}

// Ping calls xyz.block.ftl.v1.ObservabilityService.Ping.
func (c *observabilityServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// SendMetrics calls xyz.block.ftl.v1.ObservabilityService.SendMetrics.
func (c *observabilityServiceClient) SendMetrics(ctx context.Context) *connect_go.ClientStreamForClient[v1.SendMetricsRequest, v1.SendMetricsResponse] {
	return c.sendMetrics.CallClientStream(ctx)
}

// SendTraces calls xyz.block.ftl.v1.ObservabilityService.SendTraces.
func (c *observabilityServiceClient) SendTraces(ctx context.Context) *connect_go.ClientStreamForClient[v1.SendTracesRequest, v1.SendTracesResponse] {
	return c.sendTraces.CallClientStream(ctx)
}

// ObservabilityServiceHandler is an implementation of the xyz.block.ftl.v1.ObservabilityService
// service.
type ObservabilityServiceHandler interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Send OTEL metrics from the Deployment to the ControlPlane via the Runner.
	SendMetrics(context.Context, *connect_go.ClientStream[v1.SendMetricsRequest]) (*connect_go.Response[v1.SendMetricsResponse], error)
	// Send OTEL traces from the Deployment to the ControlPlane via the Runner.
	SendTraces(context.Context, *connect_go.ClientStream[v1.SendTracesRequest]) (*connect_go.Response[v1.SendTracesResponse], error)
}

// NewObservabilityServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewObservabilityServiceHandler(svc ObservabilityServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle("/xyz.block.ftl.v1.ObservabilityService/Ping", connect_go.NewUnaryHandler(
		"/xyz.block.ftl.v1.ObservabilityService/Ping",
		svc.Ping,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ObservabilityService/SendMetrics", connect_go.NewClientStreamHandler(
		"/xyz.block.ftl.v1.ObservabilityService/SendMetrics",
		svc.SendMetrics,
		opts...,
	))
	mux.Handle("/xyz.block.ftl.v1.ObservabilityService/SendTraces", connect_go.NewClientStreamHandler(
		"/xyz.block.ftl.v1.ObservabilityService/SendTraces",
		svc.SendTraces,
		opts...,
	))
	return "/xyz.block.ftl.v1.ObservabilityService/", mux
}

// UnimplementedObservabilityServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedObservabilityServiceHandler struct{}

func (UnimplementedObservabilityServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ObservabilityService.Ping is not implemented"))
}

func (UnimplementedObservabilityServiceHandler) SendMetrics(context.Context, *connect_go.ClientStream[v1.SendMetricsRequest]) (*connect_go.Response[v1.SendMetricsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ObservabilityService.SendMetrics is not implemented"))
}

func (UnimplementedObservabilityServiceHandler) SendTraces(context.Context, *connect_go.ClientStream[v1.SendTracesRequest]) (*connect_go.Response[v1.SendTracesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ObservabilityService.SendTraces is not implemented"))
}
