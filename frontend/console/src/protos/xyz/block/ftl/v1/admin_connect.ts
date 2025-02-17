// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/admin.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "./ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { ApplyChangesetRequest, ApplyChangesetResponse, ConfigGetRequest, ConfigGetResponse, ConfigListRequest, ConfigListResponse, ConfigSetRequest, ConfigSetResponse, ConfigUnsetRequest, ConfigUnsetResponse, MapConfigsForModuleRequest, MapConfigsForModuleResponse, MapSecretsForModuleRequest, MapSecretsForModuleResponse, ResetSubscriptionRequest, ResetSubscriptionResponse, SecretGetRequest, SecretGetResponse, SecretSetRequest, SecretSetResponse, SecretsListRequest, SecretsListResponse, SecretUnsetRequest, SecretUnsetResponse } from "./admin_pb.js";
import { GetSchemaRequest, GetSchemaResponse } from "./schemaservice_pb.js";

/**
 * AdminService is the service that provides and updates admin data. For example,
 * it is used to encapsulate configuration and secrets.
 *
 * @generated from service xyz.block.ftl.v1.AdminService
 */
export const AdminService = {
  typeName: "xyz.block.ftl.v1.AdminService",
  methods: {
    /**
     * @generated from rpc xyz.block.ftl.v1.AdminService.Ping
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigList
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigGet
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigSet
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ConfigUnset
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretsList
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretGet
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretSet
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.SecretUnset
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.MapConfigsForModule
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.MapSecretsForModule
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ResetSubscription
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
     * @generated from rpc xyz.block.ftl.v1.AdminService.ApplyChangeset
     */
    applyChangeset: {
      name: "ApplyChangeset",
      I: ApplyChangesetRequest,
      O: ApplyChangesetResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get the full schema.
     *
     * @generated from rpc xyz.block.ftl.v1.AdminService.GetSchema
     */
    getSchema: {
      name: "GetSchema",
      I: GetSchemaRequest,
      O: GetSchemaResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
  }
} as const;

