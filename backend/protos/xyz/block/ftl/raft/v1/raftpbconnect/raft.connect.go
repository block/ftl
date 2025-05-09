// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/raft/v1/raft.proto

package raftpbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v11 "github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1"
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
	// RaftServiceName is the fully-qualified name of the RaftService service.
	RaftServiceName = "xyz.block.ftl.raft.v1.RaftService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// RaftServicePingProcedure is the fully-qualified name of the RaftService's Ping RPC.
	RaftServicePingProcedure = "/xyz.block.ftl.raft.v1.RaftService/Ping"
	// RaftServiceAddMemberProcedure is the fully-qualified name of the RaftService's AddMember RPC.
	RaftServiceAddMemberProcedure = "/xyz.block.ftl.raft.v1.RaftService/AddMember"
	// RaftServiceRemoveMemberProcedure is the fully-qualified name of the RaftService's RemoveMember
	// RPC.
	RaftServiceRemoveMemberProcedure = "/xyz.block.ftl.raft.v1.RaftService/RemoveMember"
)

// RaftServiceClient is a client for the xyz.block.ftl.raft.v1.RaftService service.
type RaftServiceClient interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Add a new member to the cluster.
	AddMember(context.Context, *connect.Request[v11.AddMemberRequest]) (*connect.Response[v11.AddMemberResponse], error)
	// Remove a member from the cluster.
	RemoveMember(context.Context, *connect.Request[v11.RemoveMemberRequest]) (*connect.Response[v11.RemoveMemberResponse], error)
}

// NewRaftServiceClient constructs a client for the xyz.block.ftl.raft.v1.RaftService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewRaftServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) RaftServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	raftServiceMethods := v11.File_xyz_block_ftl_raft_v1_raft_proto.Services().ByName("RaftService").Methods()
	return &raftServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+RaftServicePingProcedure,
			connect.WithSchema(raftServiceMethods.ByName("Ping")),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		addMember: connect.NewClient[v11.AddMemberRequest, v11.AddMemberResponse](
			httpClient,
			baseURL+RaftServiceAddMemberProcedure,
			connect.WithSchema(raftServiceMethods.ByName("AddMember")),
			connect.WithClientOptions(opts...),
		),
		removeMember: connect.NewClient[v11.RemoveMemberRequest, v11.RemoveMemberResponse](
			httpClient,
			baseURL+RaftServiceRemoveMemberProcedure,
			connect.WithSchema(raftServiceMethods.ByName("RemoveMember")),
			connect.WithClientOptions(opts...),
		),
	}
}

// raftServiceClient implements RaftServiceClient.
type raftServiceClient struct {
	ping         *connect.Client[v1.PingRequest, v1.PingResponse]
	addMember    *connect.Client[v11.AddMemberRequest, v11.AddMemberResponse]
	removeMember *connect.Client[v11.RemoveMemberRequest, v11.RemoveMemberResponse]
}

// Ping calls xyz.block.ftl.raft.v1.RaftService.Ping.
func (c *raftServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// AddMember calls xyz.block.ftl.raft.v1.RaftService.AddMember.
func (c *raftServiceClient) AddMember(ctx context.Context, req *connect.Request[v11.AddMemberRequest]) (*connect.Response[v11.AddMemberResponse], error) {
	return c.addMember.CallUnary(ctx, req)
}

// RemoveMember calls xyz.block.ftl.raft.v1.RaftService.RemoveMember.
func (c *raftServiceClient) RemoveMember(ctx context.Context, req *connect.Request[v11.RemoveMemberRequest]) (*connect.Response[v11.RemoveMemberResponse], error) {
	return c.removeMember.CallUnary(ctx, req)
}

// RaftServiceHandler is an implementation of the xyz.block.ftl.raft.v1.RaftService service.
type RaftServiceHandler interface {
	// Ping service for readiness.
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// Add a new member to the cluster.
	AddMember(context.Context, *connect.Request[v11.AddMemberRequest]) (*connect.Response[v11.AddMemberResponse], error)
	// Remove a member from the cluster.
	RemoveMember(context.Context, *connect.Request[v11.RemoveMemberRequest]) (*connect.Response[v11.RemoveMemberResponse], error)
}

// NewRaftServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewRaftServiceHandler(svc RaftServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	raftServiceMethods := v11.File_xyz_block_ftl_raft_v1_raft_proto.Services().ByName("RaftService").Methods()
	raftServicePingHandler := connect.NewUnaryHandler(
		RaftServicePingProcedure,
		svc.Ping,
		connect.WithSchema(raftServiceMethods.ByName("Ping")),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	raftServiceAddMemberHandler := connect.NewUnaryHandler(
		RaftServiceAddMemberProcedure,
		svc.AddMember,
		connect.WithSchema(raftServiceMethods.ByName("AddMember")),
		connect.WithHandlerOptions(opts...),
	)
	raftServiceRemoveMemberHandler := connect.NewUnaryHandler(
		RaftServiceRemoveMemberProcedure,
		svc.RemoveMember,
		connect.WithSchema(raftServiceMethods.ByName("RemoveMember")),
		connect.WithHandlerOptions(opts...),
	)
	return "/xyz.block.ftl.raft.v1.RaftService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case RaftServicePingProcedure:
			raftServicePingHandler.ServeHTTP(w, r)
		case RaftServiceAddMemberProcedure:
			raftServiceAddMemberHandler.ServeHTTP(w, r)
		case RaftServiceRemoveMemberProcedure:
			raftServiceRemoveMemberHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedRaftServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedRaftServiceHandler struct{}

func (UnimplementedRaftServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.raft.v1.RaftService.Ping is not implemented"))
}

func (UnimplementedRaftServiceHandler) AddMember(context.Context, *connect.Request[v11.AddMemberRequest]) (*connect.Response[v11.AddMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.raft.v1.RaftService.AddMember is not implemented"))
}

func (UnimplementedRaftServiceHandler) RemoveMember(context.Context, *connect.Request[v11.RemoveMemberRequest]) (*connect.Response[v11.RemoveMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.raft.v1.RaftService.RemoveMember is not implemented"))
}
