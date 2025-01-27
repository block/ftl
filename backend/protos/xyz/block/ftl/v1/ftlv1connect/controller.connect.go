// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/controller.proto

package ftlv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
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
	// ControllerServiceName is the fully-qualified name of the ControllerService service.
	ControllerServiceName = "xyz.block.ftl.v1.ControllerService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ControllerServicePingProcedure is the fully-qualified name of the ControllerService's Ping RPC.
	ControllerServicePingProcedure = "/xyz.block.ftl.v1.ControllerService/Ping"
	// ControllerServiceProcessListProcedure is the fully-qualified name of the ControllerService's
	// ProcessList RPC.
	ControllerServiceProcessListProcedure = "/xyz.block.ftl.v1.ControllerService/ProcessList"
	// ControllerServiceStatusProcedure is the fully-qualified name of the ControllerService's Status
	// RPC.
	ControllerServiceStatusProcedure = "/xyz.block.ftl.v1.ControllerService/Status"
	// ControllerServiceGetArtefactDiffsProcedure is the fully-qualified name of the ControllerService's
	// GetArtefactDiffs RPC.
	ControllerServiceGetArtefactDiffsProcedure = "/xyz.block.ftl.v1.ControllerService/GetArtefactDiffs"
	// ControllerServiceUploadArtefactProcedure is the fully-qualified name of the ControllerService's
	// UploadArtefact RPC.
	ControllerServiceUploadArtefactProcedure = "/xyz.block.ftl.v1.ControllerService/UploadArtefact"
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
)

// ControllerServiceClient is a client for the xyz.block.ftl.v1.ControllerService service.
type ControllerServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// List "processes" running on the cluster.
	ProcessList(context.Context, *connect.Request[v1.ProcessListRequest]) (*connect.Response[v1.ProcessListResponse], error)
	Status(context.Context, *connect.Request[v1.StatusRequest]) (*connect.Response[v1.StatusResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error)
	// Upload an artefact to the server.
	UploadArtefact(context.Context, *connect.Request[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error)
	// Get the schema and artefact metadata for a deployment.
	GetDeployment(context.Context, *connect.Request[v1.GetDeploymentRequest]) (*connect.Response[v1.GetDeploymentResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest]) (*connect.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error)
	// Register a Runner with the Controller.
	//
	// Each runner issue a RegisterRunnerRequest to the ControllerService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context) *connect.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	// Update an existing deployment.
	UpdateDeploy(context.Context, *connect.Request[v1.UpdateDeployRequest]) (*connect.Response[v1.UpdateDeployResponse], error)
}

// NewControllerServiceClient constructs a client for the xyz.block.ftl.v1.ControllerService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewControllerServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ControllerServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &controllerServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+ControllerServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		processList: connect.NewClient[v1.ProcessListRequest, v1.ProcessListResponse](
			httpClient,
			baseURL+ControllerServiceProcessListProcedure,
			opts...,
		),
		status: connect.NewClient[v1.StatusRequest, v1.StatusResponse](
			httpClient,
			baseURL+ControllerServiceStatusProcedure,
			opts...,
		),
		getArtefactDiffs: connect.NewClient[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse](
			httpClient,
			baseURL+ControllerServiceGetArtefactDiffsProcedure,
			opts...,
		),
		uploadArtefact: connect.NewClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse](
			httpClient,
			baseURL+ControllerServiceUploadArtefactProcedure,
			opts...,
		),
		getDeployment: connect.NewClient[v1.GetDeploymentRequest, v1.GetDeploymentResponse](
			httpClient,
			baseURL+ControllerServiceGetDeploymentProcedure,
			opts...,
		),
		getDeploymentArtefacts: connect.NewClient[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse](
			httpClient,
			baseURL+ControllerServiceGetDeploymentArtefactsProcedure,
			opts...,
		),
		registerRunner: connect.NewClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse](
			httpClient,
			baseURL+ControllerServiceRegisterRunnerProcedure,
			opts...,
		),
		updateDeploy: connect.NewClient[v1.UpdateDeployRequest, v1.UpdateDeployResponse](
			httpClient,
			baseURL+ControllerServiceUpdateDeployProcedure,
			opts...,
		),
	}
}

// controllerServiceClient implements ControllerServiceClient.
type controllerServiceClient struct {
	ping                   *connect.Client[v1.PingRequest, v1.PingResponse]
	processList            *connect.Client[v1.ProcessListRequest, v1.ProcessListResponse]
	status                 *connect.Client[v1.StatusRequest, v1.StatusResponse]
	getArtefactDiffs       *connect.Client[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse]
	uploadArtefact         *connect.Client[v1.UploadArtefactRequest, v1.UploadArtefactResponse]
	getDeployment          *connect.Client[v1.GetDeploymentRequest, v1.GetDeploymentResponse]
	getDeploymentArtefacts *connect.Client[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse]
	registerRunner         *connect.Client[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse]
	updateDeploy           *connect.Client[v1.UpdateDeployRequest, v1.UpdateDeployResponse]
}

// Ping calls xyz.block.ftl.v1.ControllerService.Ping.
func (c *controllerServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// ProcessList calls xyz.block.ftl.v1.ControllerService.ProcessList.
func (c *controllerServiceClient) ProcessList(ctx context.Context, req *connect.Request[v1.ProcessListRequest]) (*connect.Response[v1.ProcessListResponse], error) {
	return c.processList.CallUnary(ctx, req)
}

// Status calls xyz.block.ftl.v1.ControllerService.Status.
func (c *controllerServiceClient) Status(ctx context.Context, req *connect.Request[v1.StatusRequest]) (*connect.Response[v1.StatusResponse], error) {
	return c.status.CallUnary(ctx, req)
}

// GetArtefactDiffs calls xyz.block.ftl.v1.ControllerService.GetArtefactDiffs.
func (c *controllerServiceClient) GetArtefactDiffs(ctx context.Context, req *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error) {
	return c.getArtefactDiffs.CallUnary(ctx, req)
}

// UploadArtefact calls xyz.block.ftl.v1.ControllerService.UploadArtefact.
func (c *controllerServiceClient) UploadArtefact(ctx context.Context, req *connect.Request[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error) {
	return c.uploadArtefact.CallUnary(ctx, req)
}

// GetDeployment calls xyz.block.ftl.v1.ControllerService.GetDeployment.
func (c *controllerServiceClient) GetDeployment(ctx context.Context, req *connect.Request[v1.GetDeploymentRequest]) (*connect.Response[v1.GetDeploymentResponse], error) {
	return c.getDeployment.CallUnary(ctx, req)
}

// GetDeploymentArtefacts calls xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts.
func (c *controllerServiceClient) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[v1.GetDeploymentArtefactsRequest]) (*connect.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error) {
	return c.getDeploymentArtefacts.CallServerStream(ctx, req)
}

// RegisterRunner calls xyz.block.ftl.v1.ControllerService.RegisterRunner.
func (c *controllerServiceClient) RegisterRunner(ctx context.Context) *connect.ClientStreamForClient[v1.RegisterRunnerRequest, v1.RegisterRunnerResponse] {
	return c.registerRunner.CallClientStream(ctx)
}

// UpdateDeploy calls xyz.block.ftl.v1.ControllerService.UpdateDeploy.
func (c *controllerServiceClient) UpdateDeploy(ctx context.Context, req *connect.Request[v1.UpdateDeployRequest]) (*connect.Response[v1.UpdateDeployResponse], error) {
	return c.updateDeploy.CallUnary(ctx, req)
}

// ControllerServiceHandler is an implementation of the xyz.block.ftl.v1.ControllerService service.
type ControllerServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// List "processes" running on the cluster.
	ProcessList(context.Context, *connect.Request[v1.ProcessListRequest]) (*connect.Response[v1.ProcessListResponse], error)
	Status(context.Context, *connect.Request[v1.StatusRequest]) (*connect.Response[v1.StatusResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error)
	// Upload an artefact to the server.
	UploadArtefact(context.Context, *connect.Request[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error)
	// Get the schema and artefact metadata for a deployment.
	GetDeployment(context.Context, *connect.Request[v1.GetDeploymentRequest]) (*connect.Response[v1.GetDeploymentResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest], *connect.ServerStream[v1.GetDeploymentArtefactsResponse]) error
	// Register a Runner with the Controller.
	//
	// Each runner issue a RegisterRunnerRequest to the ControllerService
	// every 10 seconds to maintain its heartbeat.
	RegisterRunner(context.Context, *connect.ClientStream[v1.RegisterRunnerRequest]) (*connect.Response[v1.RegisterRunnerResponse], error)
	// Update an existing deployment.
	UpdateDeploy(context.Context, *connect.Request[v1.UpdateDeployRequest]) (*connect.Response[v1.UpdateDeployResponse], error)
}

// NewControllerServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewControllerServiceHandler(svc ControllerServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	controllerServicePingHandler := connect.NewUnaryHandler(
		ControllerServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	controllerServiceProcessListHandler := connect.NewUnaryHandler(
		ControllerServiceProcessListProcedure,
		svc.ProcessList,
		opts...,
	)
	controllerServiceStatusHandler := connect.NewUnaryHandler(
		ControllerServiceStatusProcedure,
		svc.Status,
		opts...,
	)
	controllerServiceGetArtefactDiffsHandler := connect.NewUnaryHandler(
		ControllerServiceGetArtefactDiffsProcedure,
		svc.GetArtefactDiffs,
		opts...,
	)
	controllerServiceUploadArtefactHandler := connect.NewUnaryHandler(
		ControllerServiceUploadArtefactProcedure,
		svc.UploadArtefact,
		opts...,
	)
	controllerServiceGetDeploymentHandler := connect.NewUnaryHandler(
		ControllerServiceGetDeploymentProcedure,
		svc.GetDeployment,
		opts...,
	)
	controllerServiceGetDeploymentArtefactsHandler := connect.NewServerStreamHandler(
		ControllerServiceGetDeploymentArtefactsProcedure,
		svc.GetDeploymentArtefacts,
		opts...,
	)
	controllerServiceRegisterRunnerHandler := connect.NewClientStreamHandler(
		ControllerServiceRegisterRunnerProcedure,
		svc.RegisterRunner,
		opts...,
	)
	controllerServiceUpdateDeployHandler := connect.NewUnaryHandler(
		ControllerServiceUpdateDeployProcedure,
		svc.UpdateDeploy,
		opts...,
	)
	return "/xyz.block.ftl.v1.ControllerService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ControllerServicePingProcedure:
			controllerServicePingHandler.ServeHTTP(w, r)
		case ControllerServiceProcessListProcedure:
			controllerServiceProcessListHandler.ServeHTTP(w, r)
		case ControllerServiceStatusProcedure:
			controllerServiceStatusHandler.ServeHTTP(w, r)
		case ControllerServiceGetArtefactDiffsProcedure:
			controllerServiceGetArtefactDiffsHandler.ServeHTTP(w, r)
		case ControllerServiceUploadArtefactProcedure:
			controllerServiceUploadArtefactHandler.ServeHTTP(w, r)
		case ControllerServiceGetDeploymentProcedure:
			controllerServiceGetDeploymentHandler.ServeHTTP(w, r)
		case ControllerServiceGetDeploymentArtefactsProcedure:
			controllerServiceGetDeploymentArtefactsHandler.ServeHTTP(w, r)
		case ControllerServiceRegisterRunnerProcedure:
			controllerServiceRegisterRunnerHandler.ServeHTTP(w, r)
		case ControllerServiceUpdateDeployProcedure:
			controllerServiceUpdateDeployHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedControllerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedControllerServiceHandler struct{}

func (UnimplementedControllerServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.Ping is not implemented"))
}

func (UnimplementedControllerServiceHandler) ProcessList(context.Context, *connect.Request[v1.ProcessListRequest]) (*connect.Response[v1.ProcessListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.ProcessList is not implemented"))
}

func (UnimplementedControllerServiceHandler) Status(context.Context, *connect.Request[v1.StatusRequest]) (*connect.Response[v1.StatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.Status is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetArtefactDiffs is not implemented"))
}

func (UnimplementedControllerServiceHandler) UploadArtefact(context.Context, *connect.Request[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.UploadArtefact is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetDeployment(context.Context, *connect.Request[v1.GetDeploymentRequest]) (*connect.Response[v1.GetDeploymentResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetDeployment is not implemented"))
}

func (UnimplementedControllerServiceHandler) GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest], *connect.ServerStream[v1.GetDeploymentArtefactsResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.GetDeploymentArtefacts is not implemented"))
}

func (UnimplementedControllerServiceHandler) RegisterRunner(context.Context, *connect.ClientStream[v1.RegisterRunnerRequest]) (*connect.Response[v1.RegisterRunnerResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.RegisterRunner is not implemented"))
}

func (UnimplementedControllerServiceHandler) UpdateDeploy(context.Context, *connect.Request[v1.UpdateDeployRequest]) (*connect.Response[v1.UpdateDeployResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.ControllerService.UpdateDeploy is not implemented"))
}
