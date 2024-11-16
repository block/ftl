// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v2alpha1/ftl.proto (package xyz.block.ftl.v2alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Module, Schema } from "../v1/schema/schema_pb.js";

/**
 * @generated from message xyz.block.ftl.v2alpha1.GetSchemaRequest
 */
export class GetSchemaRequest extends Message<GetSchemaRequest> {
  constructor(data?: PartialMessage<GetSchemaRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.GetSchemaRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetSchemaRequest | PlainMessage<GetSchemaRequest> | undefined, b: GetSchemaRequest | PlainMessage<GetSchemaRequest> | undefined): boolean {
    return proto3.util.equals(GetSchemaRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.GetSchemaResponse
 */
export class GetSchemaResponse extends Message<GetSchemaResponse> {
  /**
   * The current full schema for the cluster.
   *
   * The active changesets.
   * repeated Changeset changesets = 2;
   *
   * @generated from field: xyz.block.ftl.v1.schema.Schema schema = 1;
   */
  schema?: Schema;

  constructor(data?: PartialMessage<GetSchemaResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.GetSchemaResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "schema", kind: "message", T: Schema },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetSchemaResponse | PlainMessage<GetSchemaResponse> | undefined, b: GetSchemaResponse | PlainMessage<GetSchemaResponse> | undefined): boolean {
    return proto3.util.equals(GetSchemaResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.PullSchemaRequest
 */
export class PullSchemaRequest extends Message<PullSchemaRequest> {
  constructor(data?: PartialMessage<PullSchemaRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.PullSchemaRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromJsonString(jsonString, options);
  }

  static equals(a: PullSchemaRequest | PlainMessage<PullSchemaRequest> | undefined, b: PullSchemaRequest | PlainMessage<PullSchemaRequest> | undefined): boolean {
    return proto3.util.equals(PullSchemaRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.PullSchemaResponse
 */
export class PullSchemaResponse extends Message<PullSchemaResponse> {
  /**
   * True if this schema change is a delete, false if it is an upsert.
   *
   * @generated from field: bool deleted = 1;
   */
  deleted = false;

  /**
   * The schema being changed.
   *
   * @generated from field: xyz.block.ftl.v1.schema.Module schema = 2;
   */
  schema?: Module;

  /**
   * If true there are more schema changes immediately following this one as part of the initial batch.
   * If false this is the last schema change in the initial batch, but others may follow later.
   *
   * @generated from field: bool initial_batch = 3;
   */
  initialBatch = false;

  constructor(data?: PartialMessage<PullSchemaResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.PullSchemaResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deleted", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 2, name: "schema", kind: "message", T: Module },
    { no: 3, name: "initial_batch", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromJsonString(jsonString, options);
  }

  static equals(a: PullSchemaResponse | PlainMessage<PullSchemaResponse> | undefined, b: PullSchemaResponse | PlainMessage<PullSchemaResponse> | undefined): boolean {
    return proto3.util.equals(PullSchemaResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.UpsertModuleRequest
 */
export class UpsertModuleRequest extends Message<UpsertModuleRequest> {
  /**
   * @generated from field: xyz.block.ftl.v1.schema.Module schema = 1;
   */
  schema?: Module;

  constructor(data?: PartialMessage<UpsertModuleRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.UpsertModuleRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "schema", kind: "message", T: Module },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpsertModuleRequest {
    return new UpsertModuleRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpsertModuleRequest {
    return new UpsertModuleRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpsertModuleRequest {
    return new UpsertModuleRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UpsertModuleRequest | PlainMessage<UpsertModuleRequest> | undefined, b: UpsertModuleRequest | PlainMessage<UpsertModuleRequest> | undefined): boolean {
    return proto3.util.equals(UpsertModuleRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.UpsertModuleResponse
 */
export class UpsertModuleResponse extends Message<UpsertModuleResponse> {
  /**
   * @generated from field: string deployment_key = 1;
   */
  deploymentKey = "";

  constructor(data?: PartialMessage<UpsertModuleResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.UpsertModuleResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployment_key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpsertModuleResponse {
    return new UpsertModuleResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpsertModuleResponse {
    return new UpsertModuleResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpsertModuleResponse {
    return new UpsertModuleResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UpsertModuleResponse | PlainMessage<UpsertModuleResponse> | undefined, b: UpsertModuleResponse | PlainMessage<UpsertModuleResponse> | undefined): boolean {
    return proto3.util.equals(UpsertModuleResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.DeleteDeploymentRequest
 */
export class DeleteDeploymentRequest extends Message<DeleteDeploymentRequest> {
  /**
   * @generated from field: string deployment_key = 1;
   */
  deploymentKey = "";

  constructor(data?: PartialMessage<DeleteDeploymentRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.DeleteDeploymentRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployment_key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteDeploymentRequest {
    return new DeleteDeploymentRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteDeploymentRequest {
    return new DeleteDeploymentRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteDeploymentRequest {
    return new DeleteDeploymentRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteDeploymentRequest | PlainMessage<DeleteDeploymentRequest> | undefined, b: DeleteDeploymentRequest | PlainMessage<DeleteDeploymentRequest> | undefined): boolean {
    return proto3.util.equals(DeleteDeploymentRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.DeleteDeploymentResponse
 */
export class DeleteDeploymentResponse extends Message<DeleteDeploymentResponse> {
  constructor(data?: PartialMessage<DeleteDeploymentResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.DeleteDeploymentResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteDeploymentResponse {
    return new DeleteDeploymentResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteDeploymentResponse {
    return new DeleteDeploymentResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteDeploymentResponse {
    return new DeleteDeploymentResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteDeploymentResponse | PlainMessage<DeleteDeploymentResponse> | undefined, b: DeleteDeploymentResponse | PlainMessage<DeleteDeploymentResponse> | undefined): boolean {
    return proto3.util.equals(DeleteDeploymentResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.DeleteModuleRequest
 */
export class DeleteModuleRequest extends Message<DeleteModuleRequest> {
  /**
   * @generated from field: string module_name = 1;
   */
  moduleName = "";

  constructor(data?: PartialMessage<DeleteModuleRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.DeleteModuleRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteModuleRequest {
    return new DeleteModuleRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteModuleRequest {
    return new DeleteModuleRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteModuleRequest {
    return new DeleteModuleRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteModuleRequest | PlainMessage<DeleteModuleRequest> | undefined, b: DeleteModuleRequest | PlainMessage<DeleteModuleRequest> | undefined): boolean {
    return proto3.util.equals(DeleteModuleRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v2alpha1.DeleteModuleResponse
 */
export class DeleteModuleResponse extends Message<DeleteModuleResponse> {
  constructor(data?: PartialMessage<DeleteModuleResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v2alpha1.DeleteModuleResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteModuleResponse {
    return new DeleteModuleResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteModuleResponse {
    return new DeleteModuleResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteModuleResponse {
    return new DeleteModuleResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteModuleResponse | PlainMessage<DeleteModuleResponse> | undefined, b: DeleteModuleResponse | PlainMessage<DeleteModuleResponse> | undefined): boolean {
    return proto3.util.equals(DeleteModuleResponse, a, b);
  }
}

