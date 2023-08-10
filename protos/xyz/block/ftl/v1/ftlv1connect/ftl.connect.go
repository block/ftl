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
const _ = connect_go.IsAtLeastVersion1_7_0

const (
	// VerbServiceName is the fully-qualified name of the VerbService service.
	VerbServiceName = "xyz.block.ftl.v1.VerbService"
	// ControllerServiceName is the fully-qualified name of the ControllerService service.
	ControllerServiceName = "xyz.block.ftl.v1.ControllerService"
	// RunnerServiceName is the fully-qualified name of the RunnerService service.
	RunnerServiceName = "xyz.block.ftl.v1.RunnerService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// VerbServicePingProcedure is the fully-qualified name of the VerbService's Ping RPC.
	VerbServicePingProcedure = "/xyz.block.ftl.v1.VerbService/Ping"
	// VerbServiceCallProcedure is the fully-qualified name of the VerbService's Call RPC.
	VerbServiceCallProcedure = "/xyz.block.ftl.v1.VerbService/Call"
	// ControllerServicePingProcedure is the fully-qualified name of the ControllerService's Ping RPC.
	ControllerServicePingProcedure = "/xyz.block.ftl.v1.ControllerService/Ping"
	// ControllerServiceStatusProcedure is the fully-qualified name of the ControllerService's Status
	// RPC.
	ControllerServiceStatusProcedure = "/xyz.block.ftl.v1.ControllerService/Status"
	// ControllerServiceGetArtefactDiffsProcedure is the fully-qualified name of the ControllerService's
	// GetArtefactDiffs RPC.
	ControllerServiceGetArtefactDiffsProcedure = "/xyz.block.ftl.v1.ControllerService/GetArtefactDiffs"
	// ControllerServiceUploadArtefactProcedure is the fully-qualified name of the ControllerService's
	// UploadArtefact RPC.
	ControllerServiceUploadArtefactProcedure = "/xyz.block.ftl.v1.ControllerService/UploadArtefact"
	// ControllerServiceCreateDeploymentProcedure is the fully-qualified name of the ControllerService's
	// CreateDeployment RPC.
	ControllerServiceCreateDeploymentProcedure = "/xyz.block.ftl.v1.ControllerService/CreateDeployment"
	// ControllerServiceGetDeploymentProcedure is the fully-qualified name of the ControllerService's
	// GetDeployment RPC.
	ControllerServiceGetDeploymentProcedure = "/xyz.block.ftl.v1.ControllerService/GetDeployment"
	// ControllerServiceGetDeploymentArtefactsProcedure is the fully-qualified name of the
	// ControllerService's GetDeploymentArtefacts RPC.
	ControllerServiceGetDeploymentArtefactsProcedure = "/xyz.block.ftl.v1.ControllerService/GetDeploymentArtefacts"
	// ControllerServiceRegisterRunnerProcedure is the fully-qualified name of the ControllerService's
	// RegisterRunner RPC.
	ControllerServiceRegisterRunnerProcedure = "/xyz.block.ftl.v1.ControllerService/RegisterRunner"
	// ControllerServiceUpdateDeployProcedure is the fully-qualified name of the ControllerService's
	// UpdateDeploy RPC.
	ControllerServiceUpdateDeployProcedure = "/xyz.block.ftl.v1.ControllerService/UpdateDeploy"
	// ControllerServiceReplaceDeployProcedure is the fully-qualified name of the ControllerService's
	// ReplaceDeploy RPC.
	ControllerServiceReplaceDeployProcedure = "/xyz.block.ftl.v1.ControllerService/ReplaceDeploy"
	// ControllerServiceStreamDeploymentLogsProcedure is the fully-qualified name of the
	// ControllerService's StreamDeploymentLogs RPC.
	ControllerServiceStreamDeploymentLogsProcedure = "/xyz.block.ftl.v1.ControllerService/StreamDeploymentLogs"
	// ControllerServicePullSchemaProcedure is the fully-qualified name of the ControllerService's
	// PullSchema RPC.
	ControllerServicePullSchemaProcedure = "/xyz.block.ftl.v1.ControllerService/PullSchema"
	// RunnerServicePingProcedure is the fully-qualified name of the RunnerService's Ping RPC.
	RunnerServicePingProcedure = "/xyz.block.ftl.v1.RunnerService/Ping"
	// RunnerServiceReserveProcedure is the fully-qualified name of the RunnerService's Reserve RPC.
	RunnerServiceReserveProcedure = "/xyz.block.ftl.v1.RunnerService/Reserve"
	// RunnerServiceDeployProcedure is the fully-qualified name of the RunnerService's Deploy RPC.
	RunnerServiceDeployProcedure = "/xyz.block.ftl.v1.RunnerService/Deploy"
	// RunnerServiceTerminateProcedure is the fully-qualified name of the RunnerService's Terminate RPC.
	RunnerServiceTerminateProcedure = "/xyz.block.ftl.v1.RunnerService/Terminate"
)

// VerbServiceClient is a client for the xyz.block.ftl.v1.VerbService service.
type VerbServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Issue a synchronous call to a Verb.
	Call(context.Context, *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error)
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
			baseURL+VerbServicePingProcedure,
			connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
			connect_go.WithClientOptions(opts...),
		),
		call: connect_go.NewClient[v1.CallRequest, v1.CallResponse](
			httpClient,
			baseURL+VerbServiceCallProcedure,
			opts...,
		),
	}
}

// verbServiceClient implements VerbServiceClient.
type verbServiceClient struct {
	ping *connect_go.Client[v1.PingRequest, v1.PingResponse]
	call *connect_go.Client[v1.CallRequest, v1.CallResponse]
}

// Ping calls xyz.block.ftl.v1.VerbService.Ping.
func (c *verbServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Call calls xyz.block.ftl.v1.VerbService.Call.
func (c *verbServiceClient) Call(ctx context.Context, req *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error) {
	return c.call.CallUnary(ctx, req)
}

// VerbServiceHandler is an implementation of the xyz.block.ftl.v1.VerbService service.
type VerbServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Issue a synchronous call to a Verb.
	Call(context.Context, *connect_go.Request[v1.CallRequest]) (*connect_go.Response[v1.CallResponse], error)
}

// NewVerbServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewVerbServiceHandler(svc VerbServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(VerbServicePingProcedure, connect_go.NewUnaryHandler(
		VerbServicePingProcedure,
		svc.Ping,
		connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
		connect_go.WithHandlerOptions(opts...),
	))
	mux.Handle(VerbServiceCallProcedure, connect_go.NewUnaryHandler(
		VerbServiceCallProcedure,
		svc.Call,
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

// ControllerServiceClient is a client for the xyz.block.ftl.v1.ControllerService service.
type ControllerServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	Status(context.Context, *connect_go.Request[v1.StatusRequest]) (*connect_go.Response[v1.StatusResponse], error)
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
	// Register a Runner with the Controller.
	//
	// Each runner issue a RegisterRunnerRequest to the ControllerService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context) *connect_go.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	// Update an existing deployment.
	UpdateDeploy(context.Context, *connect_go.Request[v1.UpdateDeployRequest]) (*connect_go.Response[v1.UpdateDeployResponse], error)
	// Gradually replace an existing deployment with a new one.
	//
	// If a deployment already exists for the module of the new deployment,
	// it will be scaled down and replaced by the new one.
	ReplaceDeploy(context.Context, *connect_go.Request[v1.ReplaceDeployRequest]) (*connect_go.Response[v1.ReplaceDeployResponse], error)
	// Stream logs from a deployment
	StreamDeploymentLogs(context.Context) *connect_go.ClientStreamForClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse]
	// Pull schema changes from the Controller.
	PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest]) (*connect_go.ServerStreamForClient[v1.PullSchemaResponse], error)
}

// NewControllerServiceClient constructs a client for the xyz.block.ftl.v1.ControllerService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewControllerServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) ControllerServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &controllerServiceClient{
		ping: connect_go.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ControllerServicePingProcedure,
			connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
			connect_go.WithClientOptions(opts...),
		),
		status: connect_go.NewClient[v1.StatusRequest, v1.StatusResponse](
			httpClient,
			baseURL+ControllerServiceStatusProcedure,
			opts...,
		),
		getArtefactDiffs: connect_go.NewClient[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse](
			httpClient,
			baseURL+ControllerServiceGetArtefactDiffsProcedure,
			opts...,
		),
		uploadArtefact: connect_go.NewClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse](
			httpClient,
			baseURL+ControllerServiceUploadArtefactProcedure,
			opts...,
		),
		createDeployment: connect_go.NewClient[v1.CreateDeploymentRequest, v1.CreateDeploymentResponse](
			httpClient,
			baseURL+ControllerServiceCreateDeploymentProcedure,
			opts...,
		),
		getDeployment: connect_go.NewClient[v1.GetDeploymentRequest, v1.GetDeploymentResponse](
			httpClient,
			baseURL+ControllerServiceGetDeploymentProcedure,
			opts...,
		),
		getDeploymentArtefacts: connect_go.NewClient[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse](
			httpClient,
			baseURL+ControllerServiceGetDeploymentArtefactsProcedure,
			opts...,
		),
		registerRunner: connect_go.NewClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse](
			httpClient,
			baseURL+ControllerServiceRegisterRunnerProcedure,
			opts...,
		),
		updateDeploy: connect_go.NewClient[v1.UpdateDeployRequest, v1.UpdateDeployResponse](
			httpClient,
			baseURL+ControllerServiceUpdateDeployProcedure,
			opts...,
		),
		replaceDeploy: connect_go.NewClient[v1.ReplaceDeployRequest, v1.ReplaceDeployResponse](
			httpClient,
			baseURL+ControllerServiceReplaceDeployProcedure,
			opts...,
		),
		streamDeploymentLogs: connect_go.NewClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse](
			httpClient,
			baseURL+ControllerServiceStreamDeploymentLogsProcedure,
			opts...,
		),
		pullSchema: connect_go.NewClient[v1.PullSchemaRequest, v1.PullSchemaResponse](
			httpClient,
			baseURL+ControllerServicePullSchemaProcedure,
			opts...,
		),
	}
}

// controllerServiceClient implements ControllerServiceClient.
type controllerServiceClient struct {
	ping                   *connect_go.Client[v1.PingRequest, v1.PingResponse]
	status                 *connect_go.Client[v1.StatusRequest, v1.StatusResponse]
	getArtefactDiffs       *connect_go.Client[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse]
	uploadArtefact         *connect_go.Client[v1.UploadArtefactRequest, v1.UploadArtefactResponse]
	createDeployment       *connect_go.Client[v1.CreateDeploymentRequest, v1.CreateDeploymentResponse]
	getDeployment          *connect_go.Client[v1.GetDeploymentRequest, v1.GetDeploymentResponse]
	getDeploymentArtefacts *connect_go.Client[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse]
	registerRunner         *connect_go.Client[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	updateDeploy           *connect_go.Client[v1.UpdateDeployRequest, v1.UpdateDeployResponse]
	replaceDeploy          *connect_go.Client[v1.ReplaceDeployRequest, v1.ReplaceDeployResponse]
	streamDeploymentLogs   *connect_go.Client[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse]
	pullSchema             *connect_go.Client[v1.PullSchemaRequest, v1.PullSchemaResponse]
}

// Ping calls xyz.block.ftl.v1.ControllerService.Ping.
func (c *controllerServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Status calls xyz.block.ftl.v1.ControllerService.Status.
func (c *controllerServiceClient) Status(ctx context.Context, req *connect_go.Request[v1.StatusRequest]) (*connect_go.Response[v1.StatusResponse], error) {
	return c.status.CallUnary(ctx, req)
}

// GetArtefactDiffs calls xyz.block.ftl.v1.ControllerService.GetArtefactDiffs.
func (c *controllerServiceClient) GetArtefactDiffs(ctx context.Context, req *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error) {
	return c.getArtefactDiffs.CallUnary(ctx, req)
}

// UploadArtefact calls xyz.block.ftl.v1.ControllerService.UploadArtefact.
func (c *controllerServiceClient) UploadArtefact(ctx context.Context, req *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error) {
	return c.uploadArtefact.CallUnary(ctx, req)
}

// CreateDeployment calls xyz.block.ftl.v1.ControllerService.CreateDeployment.
func (c *controllerServiceClient) CreateDeployment(ctx context.Context, req *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error) {
	return c.createDeployment.CallUnary(ctx, req)
}

// GetDeployment calls xyz.block.ftl.v1.ControllerService.GetDeployment.
func (c *controllerServiceClient) GetDeployment(ctx context.Context, req *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error) {
	return c.getDeployment.CallUnary(ctx, req)
}

// GetDeploymentArtefacts calls xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts.
func (c *controllerServiceClient) GetDeploymentArtefacts(ctx context.Context, req *connect_go.Request[v1.GetDeploymentArtefactsRequest]) (*connect_go.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error) {
	return c.getDeploymentArtefacts.CallServerStream(ctx, req)
}

// RegisterRunner calls xyz.block.ftl.v1.ControllerService.RegisterRunner.
func (c *controllerServiceClient) RegisterRunner(ctx context.Context) *connect_go.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse] {
	return c.registerRunner.CallClientStream(ctx)
}

// UpdateDeploy calls xyz.block.ftl.v1.ControllerService.UpdateDeploy.
func (c *controllerServiceClient) UpdateDeploy(ctx context.Context, req *connect_go.Request[v1.UpdateDeployRequest]) (*connect_go.Response[v1.UpdateDeployResponse], error) {
	return c.updateDeploy.CallUnary(ctx, req)
}

// ReplaceDeploy calls xyz.block.ftl.v1.ControllerService.ReplaceDeploy.
func (c *controllerServiceClient) ReplaceDeploy(ctx context.Context, req *connect_go.Request[v1.ReplaceDeployRequest]) (*connect_go.Response[v1.ReplaceDeployResponse], error) {
	return c.replaceDeploy.CallUnary(ctx, req)
}

// StreamDeploymentLogs calls xyz.block.ftl.v1.ControllerService.StreamDeploymentLogs.
func (c *controllerServiceClient) StreamDeploymentLogs(ctx context.Context) *connect_go.ClientStreamForClient[v1.StreamDeploymentLogsRequest, v1.StreamDeploymentLogsResponse] {
	return c.streamDeploymentLogs.CallClientStream(ctx)
}

// PullSchema calls xyz.block.ftl.v1.ControllerService.PullSchema.
func (c *controllerServiceClient) PullSchema(ctx context.Context, req *connect_go.Request[v1.PullSchemaRequest]) (*connect_go.ServerStreamForClient[v1.PullSchemaResponse], error) {
	return c.pullSchema.CallServerStream(ctx, req)
}

// ControllerServiceHandler is an implementation of the xyz.block.ftl.v1.ControllerService service.
type ControllerServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	Status(context.Context, *connect_go.Request[v1.StatusRequest]) (*connect_go.Response[v1.StatusResponse], error)
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
	// Register a Runner with the Controller.
	//
	// Each runner issue a RegisterRunnerRequest to the ControllerService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context, *connect_go.ClientStream[v1.RegisterRunnerRequest]) (*connect_go.Response[v1.RegisterRunnerResponse], error)
	// Update an existing deployment.
	UpdateDeploy(context.Context, *connect_go.Request[v1.UpdateDeployRequest]) (*connect_go.Response[v1.UpdateDeployResponse], error)
	// Gradually replace an existing deployment with a new one.
	//
	// If a deployment already exists for the module of the new deployment,
	// it will be scaled down and replaced by the new one.
	ReplaceDeploy(context.Context, *connect_go.Request[v1.ReplaceDeployRequest]) (*connect_go.Response[v1.ReplaceDeployResponse], error)
	// Stream logs from a deployment
	StreamDeploymentLogs(context.Context, *connect_go.ClientStream[v1.StreamDeploymentLogsRequest]) (*connect_go.Response[v1.StreamDeploymentLogsResponse], error)
	// Pull schema changes from the Controller.
	PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest], *connect_go.ServerStream[v1.PullSchemaResponse]) error
}

// NewControllerServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewControllerServiceHandler(svc ControllerServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(ControllerServicePingProcedure, connect_go.NewUnaryHandler(
		ControllerServicePingProcedure,
		svc.Ping,
		connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
		connect_go.WithHandlerOptions(opts...),
	))
	mux.Handle(ControllerServiceStatusProcedure, connect_go.NewUnaryHandler(
		ControllerServiceStatusProcedure,
		svc.Status,
		opts...,
	))
	mux.Handle(ControllerServiceGetArtefactDiffsProcedure, connect_go.NewUnaryHandler(
		ControllerServiceGetArtefactDiffsProcedure,
		svc.GetArtefactDiffs,
		opts...,
	))
	mux.Handle(ControllerServiceUploadArtefactProcedure, connect_go.NewUnaryHandler(
		ControllerServiceUploadArtefactProcedure,
		svc.UploadArtefact,
		opts...,
	))
	mux.Handle(ControllerServiceCreateDeploymentProcedure, connect_go.NewUnaryHandler(
		ControllerServiceCreateDeploymentProcedure,
		svc.CreateDeployment,
		opts...,
	))
	mux.Handle(ControllerServiceGetDeploymentProcedure, connect_go.NewUnaryHandler(
		ControllerServiceGetDeploymentProcedure,
		svc.GetDeployment,
		opts...,
	))
	mux.Handle(ControllerServiceGetDeploymentArtefactsProcedure, connect_go.NewServerStreamHandler(
		ControllerServiceGetDeploymentArtefactsProcedure,
		svc.GetDeploymentArtefacts,
		opts...,
	))
	mux.Handle(ControllerServiceRegisterRunnerProcedure, connect_go.NewClientStreamHandler(
		ControllerServiceRegisterRunnerProcedure,
		svc.RegisterRunner,
		opts...,
	))
	mux.Handle(ControllerServiceUpdateDeployProcedure, connect_go.NewUnaryHandler(
		ControllerServiceUpdateDeployProcedure,
		svc.UpdateDeploy,
		opts...,
	))
	mux.Handle(ControllerServiceReplaceDeployProcedure, connect_go.NewUnaryHandler(
		ControllerServiceReplaceDeployProcedure,
		svc.ReplaceDeploy,
		opts...,
	))
	mux.Handle(ControllerServiceStreamDeploymentLogsProcedure, connect_go.NewClientStreamHandler(
		ControllerServiceStreamDeploymentLogsProcedure,
		svc.StreamDeploymentLogs,
		opts...,
	))
	mux.Handle(ControllerServicePullSchemaProcedure, connect_go.NewServerStreamHandler(
		ControllerServicePullSchemaProcedure,
		svc.PullSchema,
		opts...,
	))
	return "/xyz.block.ftl.v1.ControllerService/", mux
}

// UnimplementedControllerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedControllerServiceHandler struct{}

func (UnimplementedControllerServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.Ping is not implemented"))
}

func (UnimplementedControllerServiceHandler) Status(context.Context, *connect_go.Request[v1.StatusRequest]) (*connect_go.Response[v1.StatusResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.Status is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetArtefactDiffs(context.Context, *connect_go.Request[v1.GetArtefactDiffsRequest]) (*connect_go.Response[v1.GetArtefactDiffsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetArtefactDiffs is not implemented"))
}

func (UnimplementedControllerServiceHandler) UploadArtefact(context.Context, *connect_go.Request[v1.UploadArtefactRequest]) (*connect_go.Response[v1.UploadArtefactResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.UploadArtefact is not implemented"))
}

func (UnimplementedControllerServiceHandler) CreateDeployment(context.Context, *connect_go.Request[v1.CreateDeploymentRequest]) (*connect_go.Response[v1.CreateDeploymentResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.CreateDeployment is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetDeployment(context.Context, *connect_go.Request[v1.GetDeploymentRequest]) (*connect_go.Response[v1.GetDeploymentResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetDeployment is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetDeploymentArtefacts(context.Context, *connect_go.Request[v1.GetDeploymentArtefactsRequest], *connect_go.ServerStream[v1.GetDeploymentArtefactsResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts is not implemented"))
}

func (UnimplementedControllerServiceHandler) RegisterRunner(context.Context, *connect_go.ClientStream[v1.RegisterRunnerRequest]) (*connect_go.Response[v1.RegisterRunnerResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.RegisterRunner is not implemented"))
}

func (UnimplementedControllerServiceHandler) UpdateDeploy(context.Context, *connect_go.Request[v1.UpdateDeployRequest]) (*connect_go.Response[v1.UpdateDeployResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.UpdateDeploy is not implemented"))
}

func (UnimplementedControllerServiceHandler) ReplaceDeploy(context.Context, *connect_go.Request[v1.ReplaceDeployRequest]) (*connect_go.Response[v1.ReplaceDeployResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.ReplaceDeploy is not implemented"))
}

func (UnimplementedControllerServiceHandler) StreamDeploymentLogs(context.Context, *connect_go.ClientStream[v1.StreamDeploymentLogsRequest]) (*connect_go.Response[v1.StreamDeploymentLogsResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.StreamDeploymentLogs is not implemented"))
}

func (UnimplementedControllerServiceHandler) PullSchema(context.Context, *connect_go.Request[v1.PullSchemaRequest], *connect_go.ServerStream[v1.PullSchemaResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.PullSchema is not implemented"))
}

// RunnerServiceClient is a client for the xyz.block.ftl.v1.RunnerService service.
type RunnerServiceClient interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Reserve synchronously reserves a Runner for a deployment but does nothing else.
	Reserve(context.Context, *connect_go.Request[v1.ReserveRequest]) (*connect_go.Response[v1.ReserveResponse], error)
	// Initiate a deployment on this Runner.
	Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error)
	// Terminate the deployment on this Runner.
	Terminate(context.Context, *connect_go.Request[v1.TerminateRequest]) (*connect_go.Response[v1.RegisterRunnerRequest], error)
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
			baseURL+RunnerServicePingProcedure,
			connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
			connect_go.WithClientOptions(opts...),
		),
		reserve: connect_go.NewClient[v1.ReserveRequest, v1.ReserveResponse](
			httpClient,
			baseURL+RunnerServiceReserveProcedure,
			opts...,
		),
		deploy: connect_go.NewClient[v1.DeployRequest, v1.DeployResponse](
			httpClient,
			baseURL+RunnerServiceDeployProcedure,
			opts...,
		),
		terminate: connect_go.NewClient[v1.TerminateRequest, v1.RegisterRunnerRequest](
			httpClient,
			baseURL+RunnerServiceTerminateProcedure,
			opts...,
		),
	}
}

// runnerServiceClient implements RunnerServiceClient.
type runnerServiceClient struct {
	ping      *connect_go.Client[v1.PingRequest, v1.PingResponse]
	reserve   *connect_go.Client[v1.ReserveRequest, v1.ReserveResponse]
	deploy    *connect_go.Client[v1.DeployRequest, v1.DeployResponse]
	terminate *connect_go.Client[v1.TerminateRequest, v1.RegisterRunnerRequest]
}

// Ping calls xyz.block.ftl.v1.RunnerService.Ping.
func (c *runnerServiceClient) Ping(ctx context.Context, req *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// Reserve calls xyz.block.ftl.v1.RunnerService.Reserve.
func (c *runnerServiceClient) Reserve(ctx context.Context, req *connect_go.Request[v1.ReserveRequest]) (*connect_go.Response[v1.ReserveResponse], error) {
	return c.reserve.CallUnary(ctx, req)
}

// Deploy calls xyz.block.ftl.v1.RunnerService.Deploy.
func (c *runnerServiceClient) Deploy(ctx context.Context, req *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error) {
	return c.deploy.CallUnary(ctx, req)
}

// Terminate calls xyz.block.ftl.v1.RunnerService.Terminate.
func (c *runnerServiceClient) Terminate(ctx context.Context, req *connect_go.Request[v1.TerminateRequest]) (*connect_go.Response[v1.RegisterRunnerRequest], error) {
	return c.terminate.CallUnary(ctx, req)
}

// RunnerServiceHandler is an implementation of the xyz.block.ftl.v1.RunnerService service.
type RunnerServiceHandler interface {
	Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error)
	// Reserve synchronously reserves a Runner for a deployment but does nothing else.
	Reserve(context.Context, *connect_go.Request[v1.ReserveRequest]) (*connect_go.Response[v1.ReserveResponse], error)
	// Initiate a deployment on this Runner.
	Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error)
	// Terminate the deployment on this Runner.
	Terminate(context.Context, *connect_go.Request[v1.TerminateRequest]) (*connect_go.Response[v1.RegisterRunnerRequest], error)
}

// NewRunnerServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewRunnerServiceHandler(svc RunnerServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(RunnerServicePingProcedure, connect_go.NewUnaryHandler(
		RunnerServicePingProcedure,
		svc.Ping,
		connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
		connect_go.WithHandlerOptions(opts...),
	))
	mux.Handle(RunnerServiceReserveProcedure, connect_go.NewUnaryHandler(
		RunnerServiceReserveProcedure,
		svc.Reserve,
		opts...,
	))
	mux.Handle(RunnerServiceDeployProcedure, connect_go.NewUnaryHandler(
		RunnerServiceDeployProcedure,
		svc.Deploy,
		opts...,
	))
	mux.Handle(RunnerServiceTerminateProcedure, connect_go.NewUnaryHandler(
		RunnerServiceTerminateProcedure,
		svc.Terminate,
		opts...,
	))
	return "/xyz.block.ftl.v1.RunnerService/", mux
}

// UnimplementedRunnerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedRunnerServiceHandler struct{}

func (UnimplementedRunnerServiceHandler) Ping(context.Context, *connect_go.Request[v1.PingRequest]) (*connect_go.Response[v1.PingResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.Ping is not implemented"))
}

func (UnimplementedRunnerServiceHandler) Reserve(context.Context, *connect_go.Request[v1.ReserveRequest]) (*connect_go.Response[v1.ReserveResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.Reserve is not implemented"))
}

func (UnimplementedRunnerServiceHandler) Deploy(context.Context, *connect_go.Request[v1.DeployRequest]) (*connect_go.Response[v1.DeployResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.Deploy is not implemented"))
}

func (UnimplementedRunnerServiceHandler) Terminate(context.Context, *connect_go.Request[v1.TerminateRequest]) (*connect_go.Response[v1.RegisterRunnerRequest], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("xyz.block.ftl.v1.RunnerService.Terminate is not implemented"))
}
