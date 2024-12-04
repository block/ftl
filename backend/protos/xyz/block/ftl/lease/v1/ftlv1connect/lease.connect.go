// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/lease/v1/lease.proto

package ftlv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1"
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
	// LeaseServiceName is the fully-qualified name of the LeaseService service.
	LeaseServiceName = "xyz.block.ftl.lease.v1.LeaseService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// LeaseServicePingProcedure is the fully-qualified name of the LeaseService's Ping RPC.
	LeaseServicePingProcedure = "/xyz.block.ftl.lease.v1.LeaseService/Ping"
	// LeaseServiceAcquireLeaseProcedure is the fully-qualified name of the LeaseService's AcquireLease
	// RPC.
	LeaseServiceAcquireLeaseProcedure = "/xyz.block.ftl.lease.v1.LeaseService/AcquireLease"
)

// LeaseServiceClient is a client for the xyz.block.ftl.lease.v1.LeaseService service.
type LeaseServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Acquire (and renew) a lease for a deployment.
	//
	// Returns ResourceExhausted if the lease is held.
	AcquireLease(context.Context) *connect.BidiStreamForClient[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse]
}

// NewLeaseServiceClient constructs a client for the xyz.block.ftl.lease.v1.LeaseService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewLeaseServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) LeaseServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &leaseServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+LeaseServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		acquireLease: connect.NewClient[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse](
			httpClient,
			baseURL+LeaseServiceAcquireLeaseProcedure,
			opts...,
		),
	}
}

// leaseServiceClient implements LeaseServiceClient.
type leaseServiceClient struct {
	ping         *connect.Client[v1.PingRequest, v1.PingResponse]
	acquireLease *connect.Client[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse]
}

// Ping calls xyz.block.ftl.lease.v1.LeaseService.Ping.
func (c *leaseServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// AcquireLease calls xyz.block.ftl.lease.v1.LeaseService.AcquireLease.
func (c *leaseServiceClient) AcquireLease(ctx context.Context) *connect.BidiStreamForClient[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse] {
	return c.acquireLease.CallBidiStream(ctx)
}

// LeaseServiceHandler is an implementation of the xyz.block.ftl.lease.v1.LeaseService service.
type LeaseServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Acquire (and renew) a lease for a deployment.
	//
	// Returns ResourceExhausted if the lease is held.
	AcquireLease(context.Context, *connect.BidiStream[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse]) error
}

// NewLeaseServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewLeaseServiceHandler(svc LeaseServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	leaseServicePingHandler := connect.NewUnaryHandler(
		LeaseServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	leaseServiceAcquireLeaseHandler := connect.NewBidiStreamHandler(
		LeaseServiceAcquireLeaseProcedure,
		svc.AcquireLease,
		opts...,
	)
	return "/xyz.block.ftl.lease.v1.LeaseService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LeaseServicePingProcedure:
			leaseServicePingHandler.ServeHTTP(w, r)
		case LeaseServiceAcquireLeaseProcedure:
			leaseServiceAcquireLeaseHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedLeaseServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedLeaseServiceHandler struct{}

func (UnimplementedLeaseServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.lease.v1.LeaseService.Ping is not implemented"))
}

func (UnimplementedLeaseServiceHandler) AcquireLease(context.Context, *connect.BidiStream[v11.AcquireLeaseRequest, v11.AcquireLeaseResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.lease.v1.LeaseService.AcquireLease is not implemented"))
}
