// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: xyz/block/ftl/v1/admin.proto

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
	// AdminServiceName is the fully-qualified name of the AdminService service.
	AdminServiceName = "xyz.block.ftl.v1.AdminService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// AdminServicePingProcedure is the fully-qualified name of the AdminService's Ping RPC.
	AdminServicePingProcedure = "/xyz.block.ftl.v1.AdminService/Ping"
	// AdminServiceConfigListProcedure is the fully-qualified name of the AdminService's ConfigList RPC.
	AdminServiceConfigListProcedure = "/xyz.block.ftl.v1.AdminService/ConfigList"
	// AdminServiceConfigGetProcedure is the fully-qualified name of the AdminService's ConfigGet RPC.
	AdminServiceConfigGetProcedure = "/xyz.block.ftl.v1.AdminService/ConfigGet"
	// AdminServiceConfigSetProcedure is the fully-qualified name of the AdminService's ConfigSet RPC.
	AdminServiceConfigSetProcedure = "/xyz.block.ftl.v1.AdminService/ConfigSet"
	// AdminServiceConfigUnsetProcedure is the fully-qualified name of the AdminService's ConfigUnset
	// RPC.
	AdminServiceConfigUnsetProcedure = "/xyz.block.ftl.v1.AdminService/ConfigUnset"
	// AdminServiceSecretsListProcedure is the fully-qualified name of the AdminService's SecretsList
	// RPC.
	AdminServiceSecretsListProcedure = "/xyz.block.ftl.v1.AdminService/SecretsList"
	// AdminServiceSecretGetProcedure is the fully-qualified name of the AdminService's SecretGet RPC.
	AdminServiceSecretGetProcedure = "/xyz.block.ftl.v1.AdminService/SecretGet"
	// AdminServiceSecretSetProcedure is the fully-qualified name of the AdminService's SecretSet RPC.
	AdminServiceSecretSetProcedure = "/xyz.block.ftl.v1.AdminService/SecretSet"
	// AdminServiceSecretUnsetProcedure is the fully-qualified name of the AdminService's SecretUnset
	// RPC.
	AdminServiceSecretUnsetProcedure = "/xyz.block.ftl.v1.AdminService/SecretUnset"
	// AdminServiceMapConfigsForModuleProcedure is the fully-qualified name of the AdminService's
	// MapConfigsForModule RPC.
	AdminServiceMapConfigsForModuleProcedure = "/xyz.block.ftl.v1.AdminService/MapConfigsForModule"
	// AdminServiceMapSecretsForModuleProcedure is the fully-qualified name of the AdminService's
	// MapSecretsForModule RPC.
	AdminServiceMapSecretsForModuleProcedure = "/xyz.block.ftl.v1.AdminService/MapSecretsForModule"
	// AdminServiceResetSubscriptionProcedure is the fully-qualified name of the AdminService's
	// ResetSubscription RPC.
	AdminServiceResetSubscriptionProcedure = "/xyz.block.ftl.v1.AdminService/ResetSubscription"
	// AdminServiceApplyChangesetProcedure is the fully-qualified name of the AdminService's
	// ApplyChangeset RPC.
	AdminServiceApplyChangesetProcedure = "/xyz.block.ftl.v1.AdminService/ApplyChangeset"
	// AdminServiceGetSchemaProcedure is the fully-qualified name of the AdminService's GetSchema RPC.
	AdminServiceGetSchemaProcedure = "/xyz.block.ftl.v1.AdminService/GetSchema"
	// AdminServicePullSchemaProcedure is the fully-qualified name of the AdminService's PullSchema RPC.
	AdminServicePullSchemaProcedure = "/xyz.block.ftl.v1.AdminService/PullSchema"
	// AdminServiceClusterInfoProcedure is the fully-qualified name of the AdminService's ClusterInfo
	// RPC.
	AdminServiceClusterInfoProcedure = "/xyz.block.ftl.v1.AdminService/ClusterInfo"
	// AdminServiceGetArtefactDiffsProcedure is the fully-qualified name of the AdminService's
	// GetArtefactDiffs RPC.
	AdminServiceGetArtefactDiffsProcedure = "/xyz.block.ftl.v1.AdminService/GetArtefactDiffs"
	// AdminServiceGetDeploymentArtefactsProcedure is the fully-qualified name of the AdminService's
	// GetDeploymentArtefacts RPC.
	AdminServiceGetDeploymentArtefactsProcedure = "/xyz.block.ftl.v1.AdminService/GetDeploymentArtefacts"
	// AdminServiceUploadArtefactProcedure is the fully-qualified name of the AdminService's
	// UploadArtefact RPC.
	AdminServiceUploadArtefactProcedure = "/xyz.block.ftl.v1.AdminService/UploadArtefact"
)

// AdminServiceClient is a client for the xyz.block.ftl.v1.AdminService service.
type AdminServiceClient interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// List configuration.
	ConfigList(context.Context, *connect.Request[v1.ConfigListRequest]) (*connect.Response[v1.ConfigListResponse], error)
	// Get a config value.
	ConfigGet(context.Context, *connect.Request[v1.ConfigGetRequest]) (*connect.Response[v1.ConfigGetResponse], error)
	// Set a config value.
	ConfigSet(context.Context, *connect.Request[v1.ConfigSetRequest]) (*connect.Response[v1.ConfigSetResponse], error)
	// Unset a config value.
	ConfigUnset(context.Context, *connect.Request[v1.ConfigUnsetRequest]) (*connect.Response[v1.ConfigUnsetResponse], error)
	// List secrets.
	SecretsList(context.Context, *connect.Request[v1.SecretsListRequest]) (*connect.Response[v1.SecretsListResponse], error)
	// Get a secret.
	SecretGet(context.Context, *connect.Request[v1.SecretGetRequest]) (*connect.Response[v1.SecretGetResponse], error)
	// Set a secret.
	SecretSet(context.Context, *connect.Request[v1.SecretSetRequest]) (*connect.Response[v1.SecretSetResponse], error)
	// Unset a secret.
	SecretUnset(context.Context, *connect.Request[v1.SecretUnsetRequest]) (*connect.Response[v1.SecretUnsetResponse], error)
	// MapForModule combines all configuration values visible to the module.
	// Local values take precedence.
	MapConfigsForModule(context.Context, *connect.Request[v1.MapConfigsForModuleRequest]) (*connect.Response[v1.MapConfigsForModuleResponse], error)
	// MapSecretsForModule combines all secrets visible to the module.
	// Local values take precedence.
	MapSecretsForModule(context.Context, *connect.Request[v1.MapSecretsForModuleRequest]) (*connect.Response[v1.MapSecretsForModuleResponse], error)
	// Reset the offset for a subscription to the latest of each partition.
	ResetSubscription(context.Context, *connect.Request[v1.ResetSubscriptionRequest]) (*connect.Response[v1.ResetSubscriptionResponse], error)
	// Creates and applies a changeset, returning the result
	// This blocks until the changeset has completed
	ApplyChangeset(context.Context, *connect.Request[v1.ApplyChangesetRequest]) (*connect.Response[v1.ApplyChangesetResponse], error)
	// Get the full schema.
	GetSchema(context.Context, *connect.Request[v1.GetSchemaRequest]) (*connect.Response[v1.GetSchemaResponse], error)
	// Pull schema changes from the Schema Service.
	//
	// Note that if there are no deployments this will block indefinitely, making it unsuitable for
	// just retrieving the schema. Use GetSchema for that.
	PullSchema(context.Context, *connect.Request[v1.PullSchemaRequest]) (*connect.ServerStreamForClient[v1.PullSchemaResponse], error)
	ClusterInfo(context.Context, *connect.Request[v1.ClusterInfoRequest]) (*connect.Response[v1.ClusterInfoResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest]) (*connect.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error)
	// Upload an artefact to the server.
	UploadArtefact(context.Context) *connect.ClientStreamForClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse]
}

// NewAdminServiceClient constructs a client for the xyz.block.ftl.v1.AdminService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAdminServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) AdminServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &adminServiceClient{
		ping: connect.NewClient[v1.PingRequest, v1.PingResponse](
			httpClient,
			baseURL+AdminServicePingProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		configList: connect.NewClient[v1.ConfigListRequest, v1.ConfigListResponse](
			httpClient,
			baseURL+AdminServiceConfigListProcedure,
			opts...,
		),
		configGet: connect.NewClient[v1.ConfigGetRequest, v1.ConfigGetResponse](
			httpClient,
			baseURL+AdminServiceConfigGetProcedure,
			opts...,
		),
		configSet: connect.NewClient[v1.ConfigSetRequest, v1.ConfigSetResponse](
			httpClient,
			baseURL+AdminServiceConfigSetProcedure,
			opts...,
		),
		configUnset: connect.NewClient[v1.ConfigUnsetRequest, v1.ConfigUnsetResponse](
			httpClient,
			baseURL+AdminServiceConfigUnsetProcedure,
			opts...,
		),
		secretsList: connect.NewClient[v1.SecretsListRequest, v1.SecretsListResponse](
			httpClient,
			baseURL+AdminServiceSecretsListProcedure,
			opts...,
		),
		secretGet: connect.NewClient[v1.SecretGetRequest, v1.SecretGetResponse](
			httpClient,
			baseURL+AdminServiceSecretGetProcedure,
			opts...,
		),
		secretSet: connect.NewClient[v1.SecretSetRequest, v1.SecretSetResponse](
			httpClient,
			baseURL+AdminServiceSecretSetProcedure,
			opts...,
		),
		secretUnset: connect.NewClient[v1.SecretUnsetRequest, v1.SecretUnsetResponse](
			httpClient,
			baseURL+AdminServiceSecretUnsetProcedure,
			opts...,
		),
		mapConfigsForModule: connect.NewClient[v1.MapConfigsForModuleRequest, v1.MapConfigsForModuleResponse](
			httpClient,
			baseURL+AdminServiceMapConfigsForModuleProcedure,
			opts...,
		),
		mapSecretsForModule: connect.NewClient[v1.MapSecretsForModuleRequest, v1.MapSecretsForModuleResponse](
			httpClient,
			baseURL+AdminServiceMapSecretsForModuleProcedure,
			opts...,
		),
		resetSubscription: connect.NewClient[v1.ResetSubscriptionRequest, v1.ResetSubscriptionResponse](
			httpClient,
			baseURL+AdminServiceResetSubscriptionProcedure,
			opts...,
		),
		applyChangeset: connect.NewClient[v1.ApplyChangesetRequest, v1.ApplyChangesetResponse](
			httpClient,
			baseURL+AdminServiceApplyChangesetProcedure,
			opts...,
		),
		getSchema: connect.NewClient[v1.GetSchemaRequest, v1.GetSchemaResponse](
			httpClient,
			baseURL+AdminServiceGetSchemaProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		pullSchema: connect.NewClient[v1.PullSchemaRequest, v1.PullSchemaResponse](
			httpClient,
			baseURL+AdminServicePullSchemaProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		clusterInfo: connect.NewClient[v1.ClusterInfoRequest, v1.ClusterInfoResponse](
			httpClient,
			baseURL+AdminServiceClusterInfoProcedure,
			opts...,
		),
		getArtefactDiffs: connect.NewClient[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse](
			httpClient,
			baseURL+AdminServiceGetArtefactDiffsProcedure,
			opts...,
		),
		getDeploymentArtefacts: connect.NewClient[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse](
			httpClient,
			baseURL+AdminServiceGetDeploymentArtefactsProcedure,
			opts...,
		),
		uploadArtefact: connect.NewClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse](
			httpClient,
			baseURL+AdminServiceUploadArtefactProcedure,
			opts...,
		),
	}
}

// adminServiceClient implements AdminServiceClient.
type adminServiceClient struct {
	ping                   *connect.Client[v1.PingRequest, v1.PingResponse]
	configList             *connect.Client[v1.ConfigListRequest, v1.ConfigListResponse]
	configGet              *connect.Client[v1.ConfigGetRequest, v1.ConfigGetResponse]
	configSet              *connect.Client[v1.ConfigSetRequest, v1.ConfigSetResponse]
	configUnset            *connect.Client[v1.ConfigUnsetRequest, v1.ConfigUnsetResponse]
	secretsList            *connect.Client[v1.SecretsListRequest, v1.SecretsListResponse]
	secretGet              *connect.Client[v1.SecretGetRequest, v1.SecretGetResponse]
	secretSet              *connect.Client[v1.SecretSetRequest, v1.SecretSetResponse]
	secretUnset            *connect.Client[v1.SecretUnsetRequest, v1.SecretUnsetResponse]
	mapConfigsForModule    *connect.Client[v1.MapConfigsForModuleRequest, v1.MapConfigsForModuleResponse]
	mapSecretsForModule    *connect.Client[v1.MapSecretsForModuleRequest, v1.MapSecretsForModuleResponse]
	resetSubscription      *connect.Client[v1.ResetSubscriptionRequest, v1.ResetSubscriptionResponse]
	applyChangeset         *connect.Client[v1.ApplyChangesetRequest, v1.ApplyChangesetResponse]
	getSchema              *connect.Client[v1.GetSchemaRequest, v1.GetSchemaResponse]
	pullSchema             *connect.Client[v1.PullSchemaRequest, v1.PullSchemaResponse]
	clusterInfo            *connect.Client[v1.ClusterInfoRequest, v1.ClusterInfoResponse]
	getArtefactDiffs       *connect.Client[v1.GetArtefactDiffsRequest, v1.GetArtefactDiffsResponse]
	getDeploymentArtefacts *connect.Client[v1.GetDeploymentArtefactsRequest, v1.GetDeploymentArtefactsResponse]
	uploadArtefact         *connect.Client[v1.UploadArtefactRequest, v1.UploadArtefactResponse]
}

// Ping calls xyz.block.ftl.v1.AdminService.Ping.
func (c *adminServiceClient) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return c.ping.CallUnary(ctx, req)
}

// ConfigList calls xyz.block.ftl.v1.AdminService.ConfigList.
func (c *adminServiceClient) ConfigList(ctx context.Context, req *connect.Request[v1.ConfigListRequest]) (*connect.Response[v1.ConfigListResponse], error) {
	return c.configList.CallUnary(ctx, req)
}

// ConfigGet calls xyz.block.ftl.v1.AdminService.ConfigGet.
func (c *adminServiceClient) ConfigGet(ctx context.Context, req *connect.Request[v1.ConfigGetRequest]) (*connect.Response[v1.ConfigGetResponse], error) {
	return c.configGet.CallUnary(ctx, req)
}

// ConfigSet calls xyz.block.ftl.v1.AdminService.ConfigSet.
func (c *adminServiceClient) ConfigSet(ctx context.Context, req *connect.Request[v1.ConfigSetRequest]) (*connect.Response[v1.ConfigSetResponse], error) {
	return c.configSet.CallUnary(ctx, req)
}

// ConfigUnset calls xyz.block.ftl.v1.AdminService.ConfigUnset.
func (c *adminServiceClient) ConfigUnset(ctx context.Context, req *connect.Request[v1.ConfigUnsetRequest]) (*connect.Response[v1.ConfigUnsetResponse], error) {
	return c.configUnset.CallUnary(ctx, req)
}

// SecretsList calls xyz.block.ftl.v1.AdminService.SecretsList.
func (c *adminServiceClient) SecretsList(ctx context.Context, req *connect.Request[v1.SecretsListRequest]) (*connect.Response[v1.SecretsListResponse], error) {
	return c.secretsList.CallUnary(ctx, req)
}

// SecretGet calls xyz.block.ftl.v1.AdminService.SecretGet.
func (c *adminServiceClient) SecretGet(ctx context.Context, req *connect.Request[v1.SecretGetRequest]) (*connect.Response[v1.SecretGetResponse], error) {
	return c.secretGet.CallUnary(ctx, req)
}

// SecretSet calls xyz.block.ftl.v1.AdminService.SecretSet.
func (c *adminServiceClient) SecretSet(ctx context.Context, req *connect.Request[v1.SecretSetRequest]) (*connect.Response[v1.SecretSetResponse], error) {
	return c.secretSet.CallUnary(ctx, req)
}

// SecretUnset calls xyz.block.ftl.v1.AdminService.SecretUnset.
func (c *adminServiceClient) SecretUnset(ctx context.Context, req *connect.Request[v1.SecretUnsetRequest]) (*connect.Response[v1.SecretUnsetResponse], error) {
	return c.secretUnset.CallUnary(ctx, req)
}

// MapConfigsForModule calls xyz.block.ftl.v1.AdminService.MapConfigsForModule.
func (c *adminServiceClient) MapConfigsForModule(ctx context.Context, req *connect.Request[v1.MapConfigsForModuleRequest]) (*connect.Response[v1.MapConfigsForModuleResponse], error) {
	return c.mapConfigsForModule.CallUnary(ctx, req)
}

// MapSecretsForModule calls xyz.block.ftl.v1.AdminService.MapSecretsForModule.
func (c *adminServiceClient) MapSecretsForModule(ctx context.Context, req *connect.Request[v1.MapSecretsForModuleRequest]) (*connect.Response[v1.MapSecretsForModuleResponse], error) {
	return c.mapSecretsForModule.CallUnary(ctx, req)
}

// ResetSubscription calls xyz.block.ftl.v1.AdminService.ResetSubscription.
func (c *adminServiceClient) ResetSubscription(ctx context.Context, req *connect.Request[v1.ResetSubscriptionRequest]) (*connect.Response[v1.ResetSubscriptionResponse], error) {
	return c.resetSubscription.CallUnary(ctx, req)
}

// ApplyChangeset calls xyz.block.ftl.v1.AdminService.ApplyChangeset.
func (c *adminServiceClient) ApplyChangeset(ctx context.Context, req *connect.Request[v1.ApplyChangesetRequest]) (*connect.Response[v1.ApplyChangesetResponse], error) {
	return c.applyChangeset.CallUnary(ctx, req)
}

// GetSchema calls xyz.block.ftl.v1.AdminService.GetSchema.
func (c *adminServiceClient) GetSchema(ctx context.Context, req *connect.Request[v1.GetSchemaRequest]) (*connect.Response[v1.GetSchemaResponse], error) {
	return c.getSchema.CallUnary(ctx, req)
}

// PullSchema calls xyz.block.ftl.v1.AdminService.PullSchema.
func (c *adminServiceClient) PullSchema(ctx context.Context, req *connect.Request[v1.PullSchemaRequest]) (*connect.ServerStreamForClient[v1.PullSchemaResponse], error) {
	return c.pullSchema.CallServerStream(ctx, req)
}

// ClusterInfo calls xyz.block.ftl.v1.AdminService.ClusterInfo.
func (c *adminServiceClient) ClusterInfo(ctx context.Context, req *connect.Request[v1.ClusterInfoRequest]) (*connect.Response[v1.ClusterInfoResponse], error) {
	return c.clusterInfo.CallUnary(ctx, req)
}

// GetArtefactDiffs calls xyz.block.ftl.v1.AdminService.GetArtefactDiffs.
func (c *adminServiceClient) GetArtefactDiffs(ctx context.Context, req *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error) {
	return c.getArtefactDiffs.CallUnary(ctx, req)
}

// GetDeploymentArtefacts calls xyz.block.ftl.v1.AdminService.GetDeploymentArtefacts.
func (c *adminServiceClient) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[v1.GetDeploymentArtefactsRequest]) (*connect.ServerStreamForClient[v1.GetDeploymentArtefactsResponse], error) {
	return c.getDeploymentArtefacts.CallServerStream(ctx, req)
}

// UploadArtefact calls xyz.block.ftl.v1.AdminService.UploadArtefact.
func (c *adminServiceClient) UploadArtefact(ctx context.Context) *connect.ClientStreamForClient[v1.UploadArtefactRequest, v1.UploadArtefactResponse] {
	return c.uploadArtefact.CallClientStream(ctx)
}

// AdminServiceHandler is an implementation of the xyz.block.ftl.v1.AdminService service.
type AdminServiceHandler interface {
	Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error)
	// List configuration.
	ConfigList(context.Context, *connect.Request[v1.ConfigListRequest]) (*connect.Response[v1.ConfigListResponse], error)
	// Get a config value.
	ConfigGet(context.Context, *connect.Request[v1.ConfigGetRequest]) (*connect.Response[v1.ConfigGetResponse], error)
	// Set a config value.
	ConfigSet(context.Context, *connect.Request[v1.ConfigSetRequest]) (*connect.Response[v1.ConfigSetResponse], error)
	// Unset a config value.
	ConfigUnset(context.Context, *connect.Request[v1.ConfigUnsetRequest]) (*connect.Response[v1.ConfigUnsetResponse], error)
	// List secrets.
	SecretsList(context.Context, *connect.Request[v1.SecretsListRequest]) (*connect.Response[v1.SecretsListResponse], error)
	// Get a secret.
	SecretGet(context.Context, *connect.Request[v1.SecretGetRequest]) (*connect.Response[v1.SecretGetResponse], error)
	// Set a secret.
	SecretSet(context.Context, *connect.Request[v1.SecretSetRequest]) (*connect.Response[v1.SecretSetResponse], error)
	// Unset a secret.
	SecretUnset(context.Context, *connect.Request[v1.SecretUnsetRequest]) (*connect.Response[v1.SecretUnsetResponse], error)
	// MapForModule combines all configuration values visible to the module.
	// Local values take precedence.
	MapConfigsForModule(context.Context, *connect.Request[v1.MapConfigsForModuleRequest]) (*connect.Response[v1.MapConfigsForModuleResponse], error)
	// MapSecretsForModule combines all secrets visible to the module.
	// Local values take precedence.
	MapSecretsForModule(context.Context, *connect.Request[v1.MapSecretsForModuleRequest]) (*connect.Response[v1.MapSecretsForModuleResponse], error)
	// Reset the offset for a subscription to the latest of each partition.
	ResetSubscription(context.Context, *connect.Request[v1.ResetSubscriptionRequest]) (*connect.Response[v1.ResetSubscriptionResponse], error)
	// Creates and applies a changeset, returning the result
	// This blocks until the changeset has completed
	ApplyChangeset(context.Context, *connect.Request[v1.ApplyChangesetRequest]) (*connect.Response[v1.ApplyChangesetResponse], error)
	// Get the full schema.
	GetSchema(context.Context, *connect.Request[v1.GetSchemaRequest]) (*connect.Response[v1.GetSchemaResponse], error)
	// Pull schema changes from the Schema Service.
	//
	// Note that if there are no deployments this will block indefinitely, making it unsuitable for
	// just retrieving the schema. Use GetSchema for that.
	PullSchema(context.Context, *connect.Request[v1.PullSchemaRequest], *connect.ServerStream[v1.PullSchemaResponse]) error
	ClusterInfo(context.Context, *connect.Request[v1.ClusterInfoRequest]) (*connect.Response[v1.ClusterInfoResponse], error)
	// Get list of artefacts that differ between the server and client.
	GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error)
	// Stream deployment artefacts from the server.
	//
	// Each artefact is streamed one after the other as a sequence of max 1MB
	// chunks.
	GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest], *connect.ServerStream[v1.GetDeploymentArtefactsResponse]) error
	// Upload an artefact to the server.
	UploadArtefact(context.Context, *connect.ClientStream[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error)
}

// NewAdminServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAdminServiceHandler(svc AdminServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	adminServicePingHandler := connect.NewUnaryHandler(
		AdminServicePingProcedure,
		svc.Ping,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	adminServiceConfigListHandler := connect.NewUnaryHandler(
		AdminServiceConfigListProcedure,
		svc.ConfigList,
		opts...,
	)
	adminServiceConfigGetHandler := connect.NewUnaryHandler(
		AdminServiceConfigGetProcedure,
		svc.ConfigGet,
		opts...,
	)
	adminServiceConfigSetHandler := connect.NewUnaryHandler(
		AdminServiceConfigSetProcedure,
		svc.ConfigSet,
		opts...,
	)
	adminServiceConfigUnsetHandler := connect.NewUnaryHandler(
		AdminServiceConfigUnsetProcedure,
		svc.ConfigUnset,
		opts...,
	)
	adminServiceSecretsListHandler := connect.NewUnaryHandler(
		AdminServiceSecretsListProcedure,
		svc.SecretsList,
		opts...,
	)
	adminServiceSecretGetHandler := connect.NewUnaryHandler(
		AdminServiceSecretGetProcedure,
		svc.SecretGet,
		opts...,
	)
	adminServiceSecretSetHandler := connect.NewUnaryHandler(
		AdminServiceSecretSetProcedure,
		svc.SecretSet,
		opts...,
	)
	adminServiceSecretUnsetHandler := connect.NewUnaryHandler(
		AdminServiceSecretUnsetProcedure,
		svc.SecretUnset,
		opts...,
	)
	adminServiceMapConfigsForModuleHandler := connect.NewUnaryHandler(
		AdminServiceMapConfigsForModuleProcedure,
		svc.MapConfigsForModule,
		opts...,
	)
	adminServiceMapSecretsForModuleHandler := connect.NewUnaryHandler(
		AdminServiceMapSecretsForModuleProcedure,
		svc.MapSecretsForModule,
		opts...,
	)
	adminServiceResetSubscriptionHandler := connect.NewUnaryHandler(
		AdminServiceResetSubscriptionProcedure,
		svc.ResetSubscription,
		opts...,
	)
	adminServiceApplyChangesetHandler := connect.NewUnaryHandler(
		AdminServiceApplyChangesetProcedure,
		svc.ApplyChangeset,
		opts...,
	)
	adminServiceGetSchemaHandler := connect.NewUnaryHandler(
		AdminServiceGetSchemaProcedure,
		svc.GetSchema,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	adminServicePullSchemaHandler := connect.NewServerStreamHandler(
		AdminServicePullSchemaProcedure,
		svc.PullSchema,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	adminServiceClusterInfoHandler := connect.NewUnaryHandler(
		AdminServiceClusterInfoProcedure,
		svc.ClusterInfo,
		opts...,
	)
	adminServiceGetArtefactDiffsHandler := connect.NewUnaryHandler(
		AdminServiceGetArtefactDiffsProcedure,
		svc.GetArtefactDiffs,
		opts...,
	)
	adminServiceGetDeploymentArtefactsHandler := connect.NewServerStreamHandler(
		AdminServiceGetDeploymentArtefactsProcedure,
		svc.GetDeploymentArtefacts,
		opts...,
	)
	adminServiceUploadArtefactHandler := connect.NewClientStreamHandler(
		AdminServiceUploadArtefactProcedure,
		svc.UploadArtefact,
		opts...,
	)
	return "/xyz.block.ftl.v1.AdminService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AdminServicePingProcedure:
			adminServicePingHandler.ServeHTTP(w, r)
		case AdminServiceConfigListProcedure:
			adminServiceConfigListHandler.ServeHTTP(w, r)
		case AdminServiceConfigGetProcedure:
			adminServiceConfigGetHandler.ServeHTTP(w, r)
		case AdminServiceConfigSetProcedure:
			adminServiceConfigSetHandler.ServeHTTP(w, r)
		case AdminServiceConfigUnsetProcedure:
			adminServiceConfigUnsetHandler.ServeHTTP(w, r)
		case AdminServiceSecretsListProcedure:
			adminServiceSecretsListHandler.ServeHTTP(w, r)
		case AdminServiceSecretGetProcedure:
			adminServiceSecretGetHandler.ServeHTTP(w, r)
		case AdminServiceSecretSetProcedure:
			adminServiceSecretSetHandler.ServeHTTP(w, r)
		case AdminServiceSecretUnsetProcedure:
			adminServiceSecretUnsetHandler.ServeHTTP(w, r)
		case AdminServiceMapConfigsForModuleProcedure:
			adminServiceMapConfigsForModuleHandler.ServeHTTP(w, r)
		case AdminServiceMapSecretsForModuleProcedure:
			adminServiceMapSecretsForModuleHandler.ServeHTTP(w, r)
		case AdminServiceResetSubscriptionProcedure:
			adminServiceResetSubscriptionHandler.ServeHTTP(w, r)
		case AdminServiceApplyChangesetProcedure:
			adminServiceApplyChangesetHandler.ServeHTTP(w, r)
		case AdminServiceGetSchemaProcedure:
			adminServiceGetSchemaHandler.ServeHTTP(w, r)
		case AdminServicePullSchemaProcedure:
			adminServicePullSchemaHandler.ServeHTTP(w, r)
		case AdminServiceClusterInfoProcedure:
			adminServiceClusterInfoHandler.ServeHTTP(w, r)
		case AdminServiceGetArtefactDiffsProcedure:
			adminServiceGetArtefactDiffsHandler.ServeHTTP(w, r)
		case AdminServiceGetDeploymentArtefactsProcedure:
			adminServiceGetDeploymentArtefactsHandler.ServeHTTP(w, r)
		case AdminServiceUploadArtefactProcedure:
			adminServiceUploadArtefactHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAdminServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedAdminServiceHandler struct{}

func (UnimplementedAdminServiceHandler) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.Ping is not implemented"))
}

func (UnimplementedAdminServiceHandler) ConfigList(context.Context, *connect.Request[v1.ConfigListRequest]) (*connect.Response[v1.ConfigListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ConfigList is not implemented"))
}

func (UnimplementedAdminServiceHandler) ConfigGet(context.Context, *connect.Request[v1.ConfigGetRequest]) (*connect.Response[v1.ConfigGetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ConfigGet is not implemented"))
}

func (UnimplementedAdminServiceHandler) ConfigSet(context.Context, *connect.Request[v1.ConfigSetRequest]) (*connect.Response[v1.ConfigSetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ConfigSet is not implemented"))
}

func (UnimplementedAdminServiceHandler) ConfigUnset(context.Context, *connect.Request[v1.ConfigUnsetRequest]) (*connect.Response[v1.ConfigUnsetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ConfigUnset is not implemented"))
}

func (UnimplementedAdminServiceHandler) SecretsList(context.Context, *connect.Request[v1.SecretsListRequest]) (*connect.Response[v1.SecretsListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.SecretsList is not implemented"))
}

func (UnimplementedAdminServiceHandler) SecretGet(context.Context, *connect.Request[v1.SecretGetRequest]) (*connect.Response[v1.SecretGetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.SecretGet is not implemented"))
}

func (UnimplementedAdminServiceHandler) SecretSet(context.Context, *connect.Request[v1.SecretSetRequest]) (*connect.Response[v1.SecretSetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.SecretSet is not implemented"))
}

func (UnimplementedAdminServiceHandler) SecretUnset(context.Context, *connect.Request[v1.SecretUnsetRequest]) (*connect.Response[v1.SecretUnsetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.SecretUnset is not implemented"))
}

func (UnimplementedAdminServiceHandler) MapConfigsForModule(context.Context, *connect.Request[v1.MapConfigsForModuleRequest]) (*connect.Response[v1.MapConfigsForModuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.MapConfigsForModule is not implemented"))
}

func (UnimplementedAdminServiceHandler) MapSecretsForModule(context.Context, *connect.Request[v1.MapSecretsForModuleRequest]) (*connect.Response[v1.MapSecretsForModuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.MapSecretsForModule is not implemented"))
}

func (UnimplementedAdminServiceHandler) ResetSubscription(context.Context, *connect.Request[v1.ResetSubscriptionRequest]) (*connect.Response[v1.ResetSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ResetSubscription is not implemented"))
}

func (UnimplementedAdminServiceHandler) ApplyChangeset(context.Context, *connect.Request[v1.ApplyChangesetRequest]) (*connect.Response[v1.ApplyChangesetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ApplyChangeset is not implemented"))
}

func (UnimplementedAdminServiceHandler) GetSchema(context.Context, *connect.Request[v1.GetSchemaRequest]) (*connect.Response[v1.GetSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.GetSchema is not implemented"))
}

func (UnimplementedAdminServiceHandler) PullSchema(context.Context, *connect.Request[v1.PullSchemaRequest], *connect.ServerStream[v1.PullSchemaResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.PullSchema is not implemented"))
}

func (UnimplementedAdminServiceHandler) ClusterInfo(context.Context, *connect.Request[v1.ClusterInfoRequest]) (*connect.Response[v1.ClusterInfoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.ClusterInfo is not implemented"))
}

func (UnimplementedAdminServiceHandler) GetArtefactDiffs(context.Context, *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.GetArtefactDiffs is not implemented"))
}

func (UnimplementedAdminServiceHandler) GetDeploymentArtefacts(context.Context, *connect.Request[v1.GetDeploymentArtefactsRequest], *connect.ServerStream[v1.GetDeploymentArtefactsResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.GetDeploymentArtefacts is not implemented"))
}

func (UnimplementedAdminServiceHandler) UploadArtefact(context.Context, *connect.ClientStream[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("xyz.block.ftl.v1.AdminService.UploadArtefact is not implemented"))
}
