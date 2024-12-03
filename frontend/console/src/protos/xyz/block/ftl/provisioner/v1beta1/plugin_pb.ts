// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/provisioner/v1beta1/plugin.proto (package xyz.block.ftl.provisioner.v1beta1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { DatabaseRuntimeEvent, Module, ModuleRuntimeEvent } from "../../schema/v1/schema_pb.js";

/**
 * @generated from message xyz.block.ftl.provisioner.v1beta1.ProvisionRequest
 */
export class ProvisionRequest extends Message<ProvisionRequest> {
  /**
   * @generated from field: string ftl_cluster_id = 1;
   */
  ftlClusterId = "";

  /**
   * @generated from field: xyz.block.ftl.schema.v1.Module desired_module = 2;
   */
  desiredModule?: Module;

  /**
   * @generated from field: xyz.block.ftl.schema.v1.Module previous_module = 3;
   */
  previousModule?: Module;

  /**
   * @generated from field: repeated string kinds = 4;
   */
  kinds: string[] = [];

  constructor(data?: PartialMessage<ProvisionRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.ProvisionRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ftl_cluster_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "desired_module", kind: "message", T: Module },
    { no: 3, name: "previous_module", kind: "message", T: Module },
    { no: 4, name: "kinds", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
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
 * @generated from message xyz.block.ftl.provisioner.v1beta1.ProvisionResponse
 */
export class ProvisionResponse extends Message<ProvisionResponse> {
  /**
   * @generated from field: string provisioning_token = 1;
   */
  provisioningToken = "";

  /**
   * @generated from field: xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatus status = 2;
   */
  status = ProvisionResponse_ProvisionResponseStatus.UNSPECIFIED;

  constructor(data?: PartialMessage<ProvisionResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.ProvisionResponse";
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
 * @generated from enum xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatus
 */
export enum ProvisionResponse_ProvisionResponseStatus {
  /**
   * @generated from enum value: PROVISION_RESPONSE_STATUS_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: PROVISION_RESPONSE_STATUS_SUBMITTED = 1;
   */
  SUBMITTED = 1,
}
// Retrieve enum metadata with: proto3.getEnumType(ProvisionResponse_ProvisionResponseStatus)
proto3.util.setEnumType(ProvisionResponse_ProvisionResponseStatus, "xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatus", [
  { no: 0, name: "PROVISION_RESPONSE_STATUS_UNSPECIFIED" },
  { no: 1, name: "PROVISION_RESPONSE_STATUS_SUBMITTED" },
]);

/**
 * @generated from message xyz.block.ftl.provisioner.v1beta1.StatusRequest
 */
export class StatusRequest extends Message<StatusRequest> {
  /**
   * @generated from field: string provisioning_token = 1;
   */
  provisioningToken = "";

  /**
   * The outputs of this module are updated if the the status is a success
   *
   * @generated from field: xyz.block.ftl.schema.v1.Module desired_module = 2;
   */
  desiredModule?: Module;

  constructor(data?: PartialMessage<StatusRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.StatusRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provisioning_token", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "desired_module", kind: "message", T: Module },
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
 * @generated from message xyz.block.ftl.provisioner.v1beta1.ProvisioningEvent
 */
export class ProvisioningEvent extends Message<ProvisioningEvent> {
  /**
   * @generated from oneof xyz.block.ftl.provisioner.v1beta1.ProvisioningEvent.value
   */
  value: {
    /**
     * @generated from field: xyz.block.ftl.schema.v1.ModuleRuntimeEvent module_runtime_event = 1;
     */
    value: ModuleRuntimeEvent;
    case: "moduleRuntimeEvent";
  } | {
    /**
     * @generated from field: xyz.block.ftl.schema.v1.DatabaseRuntimeEvent database_runtime_event = 2;
     */
    value: DatabaseRuntimeEvent;
    case: "databaseRuntimeEvent";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<ProvisioningEvent>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.ProvisioningEvent";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module_runtime_event", kind: "message", T: ModuleRuntimeEvent, oneof: "value" },
    { no: 2, name: "database_runtime_event", kind: "message", T: DatabaseRuntimeEvent, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProvisioningEvent {
    return new ProvisioningEvent().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProvisioningEvent {
    return new ProvisioningEvent().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProvisioningEvent {
    return new ProvisioningEvent().fromJsonString(jsonString, options);
  }

  static equals(a: ProvisioningEvent | PlainMessage<ProvisioningEvent> | undefined, b: ProvisioningEvent | PlainMessage<ProvisioningEvent> | undefined): boolean {
    return proto3.util.equals(ProvisioningEvent, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.provisioner.v1beta1.StatusResponse
 */
export class StatusResponse extends Message<StatusResponse> {
  /**
   * @generated from oneof xyz.block.ftl.provisioner.v1beta1.StatusResponse.status
   */
  status: {
    /**
     * @generated from field: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunning running = 1;
     */
    value: StatusResponse_ProvisioningRunning;
    case: "running";
  } | {
    /**
     * @generated from field: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess success = 2;
     */
    value: StatusResponse_ProvisioningSuccess;
    case: "success";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<StatusResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.StatusResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "running", kind: "message", T: StatusResponse_ProvisioningRunning, oneof: "status" },
    { no: 2, name: "success", kind: "message", T: StatusResponse_ProvisioningSuccess, oneof: "status" },
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
 * @generated from message xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunning
 */
export class StatusResponse_ProvisioningRunning extends Message<StatusResponse_ProvisioningRunning> {
  constructor(data?: PartialMessage<StatusResponse_ProvisioningRunning>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunning";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StatusResponse_ProvisioningRunning {
    return new StatusResponse_ProvisioningRunning().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningRunning {
    return new StatusResponse_ProvisioningRunning().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningRunning {
    return new StatusResponse_ProvisioningRunning().fromJsonString(jsonString, options);
  }

  static equals(a: StatusResponse_ProvisioningRunning | PlainMessage<StatusResponse_ProvisioningRunning> | undefined, b: StatusResponse_ProvisioningRunning | PlainMessage<StatusResponse_ProvisioningRunning> | undefined): boolean {
    return proto3.util.equals(StatusResponse_ProvisioningRunning, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningFailed
 */
export class StatusResponse_ProvisioningFailed extends Message<StatusResponse_ProvisioningFailed> {
  /**
   * @generated from field: string error_message = 1;
   */
  errorMessage = "";

  constructor(data?: PartialMessage<StatusResponse_ProvisioningFailed>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningFailed";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StatusResponse_ProvisioningFailed {
    return new StatusResponse_ProvisioningFailed().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningFailed {
    return new StatusResponse_ProvisioningFailed().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningFailed {
    return new StatusResponse_ProvisioningFailed().fromJsonString(jsonString, options);
  }

  static equals(a: StatusResponse_ProvisioningFailed | PlainMessage<StatusResponse_ProvisioningFailed> | undefined, b: StatusResponse_ProvisioningFailed | PlainMessage<StatusResponse_ProvisioningFailed> | undefined): boolean {
    return proto3.util.equals(StatusResponse_ProvisioningFailed, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess
 */
export class StatusResponse_ProvisioningSuccess extends Message<StatusResponse_ProvisioningSuccess> {
  /**
   * @generated from field: repeated xyz.block.ftl.provisioner.v1beta1.ProvisioningEvent events = 1;
   */
  events: ProvisioningEvent[] = [];

  constructor(data?: PartialMessage<StatusResponse_ProvisioningSuccess>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: ProvisioningEvent, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StatusResponse_ProvisioningSuccess {
    return new StatusResponse_ProvisioningSuccess().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningSuccess {
    return new StatusResponse_ProvisioningSuccess().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StatusResponse_ProvisioningSuccess {
    return new StatusResponse_ProvisioningSuccess().fromJsonString(jsonString, options);
  }

  static equals(a: StatusResponse_ProvisioningSuccess | PlainMessage<StatusResponse_ProvisioningSuccess> | undefined, b: StatusResponse_ProvisioningSuccess | PlainMessage<StatusResponse_ProvisioningSuccess> | undefined): boolean {
    return proto3.util.equals(StatusResponse_ProvisioningSuccess, a, b);
  }
}

