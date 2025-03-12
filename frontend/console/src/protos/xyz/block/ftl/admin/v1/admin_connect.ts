// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/admin/v1/admin.proto (package xyz.block.ftl.admin.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../../v1/ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { ApplyChangesetRequest, ApplyChangesetResponse, ClusterInfoRequest, ClusterInfoResponse, ConfigGetRequest, ConfigGetResponse, ConfigListRequest, ConfigListResponse, ConfigSetRequest, ConfigSetResponse, ConfigUnsetRequest, ConfigUnsetResponse, GetArtefactDiffsRequest, GetArtefactDiffsResponse, GetDeploymentArtefactsRequest, GetDeploymentArtefactsResponse, MapConfigsForModuleRequest, MapConfigsForModuleResponse, MapSecretsForModuleRequest, MapSecretsForModuleResponse, ResetSubscriptionRequest, ResetSubscriptionResponse, SecretGetRequest, SecretGetResponse, SecretSetRequest, SecretSetResponse, SecretsListRequest, SecretsListResponse, SecretUnsetRequest, SecretUnsetResponse, StreamChangesetLogsRequest, StreamChangesetLogsResponse, UpdateDeploymentRuntimeRequest, UpdateDeploymentRuntimeResponse, UploadArtefactRequest, UploadArtefactResponse } from "./admin_pb.js";
import { FailChangesetRequest, FailChangesetResponse, GetSchemaRequest, GetSchemaResponse, PullSchemaRequest, PullSchemaResponse, RollbackChangesetRequest, RollbackChangesetResponse } from "../../v1/schemaservice_pb.js";

/**
 * AdminService is the service that provides and updates admin data. For example,
 * it is used to encapsulate configuration and secrets.
 *
 * @generated from service xyz.block.ftl.admin.v1.AdminService
 */
export const AdminService = {
  typeName: "xyz.block.ftl.admin.v1.AdminService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * List configuration.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ConfigList
     */
    configList: {
      name: "ConfigList",
      I: ConfigListRequest,
      O: ConfigListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get a config value.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ConfigGet
     */
    configGet: {
      name: "ConfigGet",
      I: ConfigGetRequest,
      O: ConfigGetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Set a config value.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ConfigSet
     */
    configSet: {
      name: "ConfigSet",
      I: ConfigSetRequest,
      O: ConfigSetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Unset a config value.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ConfigUnset
     */
    configUnset: {
      name: "ConfigUnset",
      I: ConfigUnsetRequest,
      O: ConfigUnsetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * List secrets.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.SecretsList
     */
    secretsList: {
      name: "SecretsList",
      I: SecretsListRequest,
      O: SecretsListResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get a secret.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.SecretGet
     */
    secretGet: {
      name: "SecretGet",
      I: SecretGetRequest,
      O: SecretGetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Set a secret.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.SecretSet
     */
    secretSet: {
      name: "SecretSet",
      I: SecretSetRequest,
      O: SecretSetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Unset a secret.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.SecretUnset
     */
    secretUnset: {
      name: "SecretUnset",
      I: SecretUnsetRequest,
      O: SecretUnsetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * MapForModule combines all configuration values visible to the module.
     * Local values take precedence.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.MapConfigsForModule
     */
    mapConfigsForModule: {
      name: "MapConfigsForModule",
      I: MapConfigsForModuleRequest,
      O: MapConfigsForModuleResponse,
      kind: MethodKind.Unary,
    },
    /**
     * MapSecretsForModule combines all secrets visible to the module.
     * Local values take precedence.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.MapSecretsForModule
     */
    mapSecretsForModule: {
      name: "MapSecretsForModule",
      I: MapSecretsForModuleRequest,
      O: MapSecretsForModuleResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Reset the offset for a subscription to the latest of each partition.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ResetSubscription
     */
    resetSubscription: {
      name: "ResetSubscription",
      I: ResetSubscriptionRequest,
      O: ResetSubscriptionResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Creates and applies a changeset, returning the result
     * This blocks until the changeset has completed
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ApplyChangeset
     */
    applyChangeset: {
      name: "ApplyChangeset",
      I: ApplyChangesetRequest,
      O: ApplyChangesetResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Updates a runtime deployment
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.UpdateDeploymentRuntime
     */
    updateDeploymentRuntime: {
      name: "UpdateDeploymentRuntime",
      I: UpdateDeploymentRuntimeRequest,
      O: UpdateDeploymentRuntimeResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get the full schema.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.GetSchema
     */
    getSchema: {
      name: "GetSchema",
      I: GetSchemaRequest,
      O: GetSchemaResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Pull schema changes from the Schema Service.
     *
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.PullSchema
     */
    pullSchema: {
      name: "PullSchema",
      I: PullSchemaRequest,
      O: PullSchemaResponse,
      kind: MethodKind.ServerStreaming,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * RollbackChangeset Rolls back a failing changeset
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.RollbackChangeset
     */
    rollbackChangeset: {
      name: "RollbackChangeset",
      I: RollbackChangesetRequest,
      O: RollbackChangesetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * FailChangeset fails an active changeset.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.FailChangeset
     */
    failChangeset: {
      name: "FailChangeset",
      I: FailChangesetRequest,
      O: FailChangesetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.ClusterInfo
     */
    clusterInfo: {
      name: "ClusterInfo",
      I: ClusterInfoRequest,
      O: ClusterInfoResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get list of artefacts that differ between the server and client.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.GetArtefactDiffs
     */
    getArtefactDiffs: {
      name: "GetArtefactDiffs",
      I: GetArtefactDiffsRequest,
      O: GetArtefactDiffsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stream deployment artefacts from the server.
     *
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.GetDeploymentArtefacts
     */
    getDeploymentArtefacts: {
      name: "GetDeploymentArtefacts",
      I: GetDeploymentArtefactsRequest,
      O: GetDeploymentArtefactsResponse,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Upload an artefact to the server.
     *
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.UploadArtefact
     */
    uploadArtefact: {
      name: "UploadArtefact",
      I: UploadArtefactRequest,
      O: UploadArtefactResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * @generated from rpc xyz.block.ftl.admin.v1.AdminService.StreamChangesetLogs
     */
    streamChangesetLogs: {
      name: "StreamChangesetLogs",
      I: StreamChangesetLogsRequest,
      O: StreamChangesetLogsResponse,
      kind: MethodKind.ServerStreaming,
    },
  }
} as const;

