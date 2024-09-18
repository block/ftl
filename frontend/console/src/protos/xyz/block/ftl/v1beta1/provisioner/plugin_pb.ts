// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1beta1/provisioner/plugin.proto (package xyz.block.ftl.v1beta1.provisioner, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * ResourceProperty is an implementation specific property of the resource populated at creation time
 *
 * @generated from message xyz.block.ftl.v1beta1.provisioner.ResourceProperty
 */
export class ResourceProperty extends Message<ResourceProperty> {
  /**
   * @generated from field: string resource_id = 1;
   */
  resourceId = "";

  /**
   * @generated from field: string key = 2;
   */
  key = "";

  /**
   * @generated from field: string value = 3;
   */
  value = "";

  constructor(data?: PartialMessage<ResourceProperty>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.ResourceProperty";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "value", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ResourceProperty {
    return new ResourceProperty().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ResourceProperty {
    return new ResourceProperty().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ResourceProperty {
    return new ResourceProperty().fromJsonString(jsonString, options);
  }

  static equals(a: ResourceProperty | PlainMessage<ResourceProperty> | undefined, b: ResourceProperty | PlainMessage<ResourceProperty> | undefined): boolean {
    return proto3.util.equals(ResourceProperty, a, b);
  }
}

/**
 * Resource is an abstract resource extracted from FTL Schema.
 *
 * @generated from message xyz.block.ftl.v1beta1.provisioner.Resource
 */
export class Resource extends Message<Resource> {
  /**
   * id unique within the module
   *
   * @generated from field: string resource_id = 1;
   */
  resourceId = "";

  /**
   * potential implementation specific properties populated
   * when the resource was created
   *
   * @generated from field: repeated xyz.block.ftl.v1beta1.provisioner.ResourceProperty properties = 2;
   */
  properties: ResourceProperty[] = [];

  /**
   * direct downstream dependencies this Resource depends on
   *
   * @generated from field: repeated xyz.block.ftl.v1beta1.provisioner.Resource dependencies = 3;
   */
  dependencies: Resource[] = [];

  /**
   * @generated from oneof xyz.block.ftl.v1beta1.provisioner.Resource.resource
   */
  resource: {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.Resource.FtlCluster ftl = 101;
     */
    value: Resource_FtlCluster;
    case: "ftl";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.Resource.PostgresResource postgres = 102;
     */
    value: Resource_PostgresResource;
    case: "postgres";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.Resource.MysqlResource mysql = 103;
     */
    value: Resource_MysqlResource;
    case: "mysql";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Resource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "properties", kind: "message", T: ResourceProperty, repeated: true },
    { no: 3, name: "dependencies", kind: "message", T: Resource, repeated: true },
    { no: 101, name: "ftl", kind: "message", T: Resource_FtlCluster, oneof: "resource" },
    { no: 102, name: "postgres", kind: "message", T: Resource_PostgresResource, oneof: "resource" },
    { no: 103, name: "mysql", kind: "message", T: Resource_MysqlResource, oneof: "resource" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Resource {
    return new Resource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Resource {
    return new Resource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Resource {
    return new Resource().fromJsonString(jsonString, options);
  }

  static equals(a: Resource | PlainMessage<Resource> | undefined, b: Resource | PlainMessage<Resource> | undefined): boolean {
    return proto3.util.equals(Resource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.Resource.FtlCluster
 */
export class Resource_FtlCluster extends Message<Resource_FtlCluster> {
  constructor(data?: PartialMessage<Resource_FtlCluster>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource.FtlCluster";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Resource_FtlCluster {
    return new Resource_FtlCluster().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Resource_FtlCluster {
    return new Resource_FtlCluster().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Resource_FtlCluster {
    return new Resource_FtlCluster().fromJsonString(jsonString, options);
  }

  static equals(a: Resource_FtlCluster | PlainMessage<Resource_FtlCluster> | undefined, b: Resource_FtlCluster | PlainMessage<Resource_FtlCluster> | undefined): boolean {
    return proto3.util.equals(Resource_FtlCluster, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.Resource.PostgresResource
 */
export class Resource_PostgresResource extends Message<Resource_PostgresResource> {
  constructor(data?: PartialMessage<Resource_PostgresResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource.PostgresResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Resource_PostgresResource {
    return new Resource_PostgresResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Resource_PostgresResource {
    return new Resource_PostgresResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Resource_PostgresResource {
    return new Resource_PostgresResource().fromJsonString(jsonString, options);
  }

  static equals(a: Resource_PostgresResource | PlainMessage<Resource_PostgresResource> | undefined, b: Resource_PostgresResource | PlainMessage<Resource_PostgresResource> | undefined): boolean {
    return proto3.util.equals(Resource_PostgresResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.Resource.MysqlResource
 */
export class Resource_MysqlResource extends Message<Resource_MysqlResource> {
  constructor(data?: PartialMessage<Resource_MysqlResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource.MysqlResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Resource_MysqlResource {
    return new Resource_MysqlResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Resource_MysqlResource {
    return new Resource_MysqlResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Resource_MysqlResource {
    return new Resource_MysqlResource().fromJsonString(jsonString, options);
  }

  static equals(a: Resource_MysqlResource | PlainMessage<Resource_MysqlResource> | undefined, b: Resource_MysqlResource | PlainMessage<Resource_MysqlResource> | undefined): boolean {
    return proto3.util.equals(Resource_MysqlResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.ProvisionRequest
 */
export class ProvisionRequest extends Message<ProvisionRequest> {
  /**
   * @generated from field: string ftl_cluster_id = 1;
   */
  ftlClusterId = "";

  /**
   * @generated from field: string module = 2;
   */
  module = "";

  /**
   * The resource FTL thinks exists currently
   *
   * @generated from field: repeated xyz.block.ftl.v1beta1.provisioner.Resource existing_resources = 3;
   */
  existingResources: Resource[] = [];

  /**
   * The resource FTL would like to exist after this provisioning run.
   * This includes all new, existing, and changes resources in this change.
   *
   * @generated from field: repeated xyz.block.ftl.v1beta1.provisioner.Resource desired_resources = 4;
   */
  desiredResources: Resource[] = [];

  constructor(data?: PartialMessage<ProvisionRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.ProvisionRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ftl_cluster_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "existing_resources", kind: "message", T: Resource, repeated: true },
    { no: 4, name: "desired_resources", kind: "message", T: Resource, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProvisionRequest {
    return new ProvisionRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProvisionRequest {
    return new ProvisionRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProvisionRequest {
    return new ProvisionRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ProvisionRequest | PlainMessage<ProvisionRequest> | undefined, b: ProvisionRequest | PlainMessage<ProvisionRequest> | undefined): boolean {
    return proto3.util.equals(ProvisionRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.ProvisionResponse
 */
export class ProvisionResponse extends Message<ProvisionResponse> {
  /**
   * @generated from field: string provisioning_token = 1;
   */
  provisioningToken = "";

  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.ProvisionResponse.ProvisionResponseStatus status = 2;
   */
  status = ProvisionResponse_ProvisionResponseStatus.UNKNOWN;

  constructor(data?: PartialMessage<ProvisionResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.ProvisionResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provisioning_token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "status", kind: "enum", T: proto3.getEnumType(ProvisionResponse_ProvisionResponseStatus) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProvisionResponse {
    return new ProvisionResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProvisionResponse {
    return new ProvisionResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProvisionResponse {
    return new ProvisionResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ProvisionResponse | PlainMessage<ProvisionResponse> | undefined, b: ProvisionResponse | PlainMessage<ProvisionResponse> | undefined): boolean {
    return proto3.util.equals(ProvisionResponse, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.v1beta1.provisioner.ProvisionResponse.ProvisionResponseStatus
 */
export enum ProvisionResponse_ProvisionResponseStatus {
  /**
   * @generated from enum value: UNKNOWN = 0;
   */
  UNKNOWN = 0,

  /**
   * @generated from enum value: SUBMITTED = 1;
   */
  SUBMITTED = 1,

  /**
   * @generated from enum value: NO_CHANGES = 2;
   */
  NO_CHANGES = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(ProvisionResponse_ProvisionResponseStatus)
proto3.util.setEnumType(ProvisionResponse_ProvisionResponseStatus, "xyz.block.ftl.v1beta1.provisioner.ProvisionResponse.ProvisionResponseStatus", [
  { no: 0, name: "UNKNOWN" },
  { no: 1, name: "SUBMITTED" },
  { no: 2, name: "NO_CHANGES" },
]);

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.StatusRequest
 */
export class StatusRequest extends Message<StatusRequest> {
  /**
   * @generated from field: string provisioning_token = 1;
   */
  provisioningToken = "";

  constructor(data?: PartialMessage<StatusRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.StatusRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provisioning_token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StatusRequest {
    return new StatusRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StatusRequest {
    return new StatusRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StatusRequest {
    return new StatusRequest().fromJsonString(jsonString, options);
  }

  static equals(a: StatusRequest | PlainMessage<StatusRequest> | undefined, b: StatusRequest | PlainMessage<StatusRequest> | undefined): boolean {
    return proto3.util.equals(StatusRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.StatusResponse
 */
export class StatusResponse extends Message<StatusResponse> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.StatusResponse.ProvisioningStatus status = 1;
   */
  status = StatusResponse_ProvisioningStatus.UNKNOWN;

  /**
   * @generated from field: string error_message = 2;
   */
  errorMessage = "";

  /**
   * @generated from field: repeated xyz.block.ftl.v1beta1.provisioner.ResourceProperty properties = 3;
   */
  properties: ResourceProperty[] = [];

  constructor(data?: PartialMessage<StatusResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.StatusResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "status", kind: "enum", T: proto3.getEnumType(StatusResponse_ProvisioningStatus) },
    { no: 2, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "properties", kind: "message", T: ResourceProperty, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StatusResponse {
    return new StatusResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StatusResponse {
    return new StatusResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StatusResponse {
    return new StatusResponse().fromJsonString(jsonString, options);
  }

  static equals(a: StatusResponse | PlainMessage<StatusResponse> | undefined, b: StatusResponse | PlainMessage<StatusResponse> | undefined): boolean {
    return proto3.util.equals(StatusResponse, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.v1beta1.provisioner.StatusResponse.ProvisioningStatus
 */
export enum StatusResponse_ProvisioningStatus {
  /**
   * @generated from enum value: UNKNOWN = 0;
   */
  UNKNOWN = 0,

  /**
   * @generated from enum value: RUNNING = 1;
   */
  RUNNING = 1,

  /**
   * @generated from enum value: SUCCEEDED = 2;
   */
  SUCCEEDED = 2,

  /**
   * @generated from enum value: FAILED = 3;
   */
  FAILED = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(StatusResponse_ProvisioningStatus)
proto3.util.setEnumType(StatusResponse_ProvisioningStatus, "xyz.block.ftl.v1beta1.provisioner.StatusResponse.ProvisioningStatus", [
  { no: 0, name: "UNKNOWN" },
  { no: 1, name: "RUNNING" },
  { no: 2, name: "SUCCEEDED" },
  { no: 3, name: "FAILED" },
]);

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PlanRequest
 */
export class PlanRequest extends Message<PlanRequest> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.ProvisionRequest provisioning = 1;
   */
  provisioning?: ProvisionRequest;

  constructor(data?: PartialMessage<PlanRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PlanRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provisioning", kind: "message", T: ProvisionRequest },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PlanRequest {
    return new PlanRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PlanRequest {
    return new PlanRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PlanRequest {
    return new PlanRequest().fromJsonString(jsonString, options);
  }

  static equals(a: PlanRequest | PlainMessage<PlanRequest> | undefined, b: PlanRequest | PlainMessage<PlanRequest> | undefined): boolean {
    return proto3.util.equals(PlanRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PlanResponse
 */
export class PlanResponse extends Message<PlanResponse> {
  /**
   * a detailed, implementation specific, plan of changes this deployment would do
   *
   * @generated from field: string plan = 1;
   */
  plan = "";

  constructor(data?: PartialMessage<PlanResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PlanResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "plan", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PlanResponse {
    return new PlanResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PlanResponse {
    return new PlanResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PlanResponse {
    return new PlanResponse().fromJsonString(jsonString, options);
  }

  static equals(a: PlanResponse | PlainMessage<PlanResponse> | undefined, b: PlanResponse | PlainMessage<PlanResponse> | undefined): boolean {
    return proto3.util.equals(PlanResponse, a, b);
  }
}

