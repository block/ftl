// @generated by protoc-gen-connect-es v1.6.1 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/schemaservice.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "./ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { GetSchemaRequest, GetSchemaResponse, PullSchemaRequest, PullSchemaResponse } from "./schemaservice_pb.js";

/**
 * @generated from service xyz.block.ftl.v1.SchemaService
 */
export const SchemaService = {
  typeName: "xyz.block.ftl.v1.SchemaService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.SchemaService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Get the full schema.
     *
     * @generated from rpc xyz.block.ftl.v1.SchemaService.GetSchema
     */
    getSchema: {
      name: "GetSchema",
      I: GetSchemaRequest,
      O: GetSchemaResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Pull schema changes from the Controller.
     *
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     *
     * @generated from rpc xyz.block.ftl.v1.SchemaService.PullSchema
     */
    pullSchema: {
      name: "PullSchema",
      I: PullSchemaRequest,
      O: PullSchemaResponse,
      kind: MethodKind.ServerStreaming,
      idempotency: MethodIdempotency.NoSideEffects,
    },
  }
} as const;

