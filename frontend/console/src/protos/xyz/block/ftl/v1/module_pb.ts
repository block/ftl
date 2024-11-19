// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/module.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Duration, Message, proto3 } from "@bufbuild/protobuf";
import { Ref } from "./schema/schema_pb.js";

/**
 * @generated from message xyz.block.ftl.v1.AcquireLeaseRequest
 */
export class AcquireLeaseRequest extends Message<AcquireLeaseRequest> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  /**
   * @generated from field: repeated string key = 2;
   */
  key: string[] = [];

  /**
   * @generated from field: google.protobuf.Duration ttl = 3;
   */
  ttl?: Duration;

  constructor(data?: PartialMessage<AcquireLeaseRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.AcquireLeaseRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "key", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 3, name: "ttl", kind: "message", T: Duration },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AcquireLeaseRequest {
    return new AcquireLeaseRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AcquireLeaseRequest {
    return new AcquireLeaseRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AcquireLeaseRequest {
    return new AcquireLeaseRequest().fromJsonString(jsonString, options);
  }

  static equals(a: AcquireLeaseRequest | PlainMessage<AcquireLeaseRequest> | undefined, b: AcquireLeaseRequest | PlainMessage<AcquireLeaseRequest> | undefined): boolean {
    return proto3.util.equals(AcquireLeaseRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.AcquireLeaseResponse
 */
export class AcquireLeaseResponse extends Message<AcquireLeaseResponse> {
  constructor(data?: PartialMessage<AcquireLeaseResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.AcquireLeaseResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AcquireLeaseResponse {
    return new AcquireLeaseResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AcquireLeaseResponse {
    return new AcquireLeaseResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AcquireLeaseResponse {
    return new AcquireLeaseResponse().fromJsonString(jsonString, options);
  }

  static equals(a: AcquireLeaseResponse | PlainMessage<AcquireLeaseResponse> | undefined, b: AcquireLeaseResponse | PlainMessage<AcquireLeaseResponse> | undefined): boolean {
    return proto3.util.equals(AcquireLeaseResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.PublishEventRequest
 */
export class PublishEventRequest extends Message<PublishEventRequest> {
  /**
   * @generated from field: xyz.block.ftl.v1.schema.Ref topic = 1;
   */
  topic?: Ref;

  /**
   * @generated from field: bytes body = 2;
   */
  body = new Uint8Array(0);

  /**
   * Only verb name is included because this verb will be in the same module as topic
   *
   * @generated from field: string caller = 3;
   */
  caller = "";

  constructor(data?: PartialMessage<PublishEventRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.PublishEventRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "topic", kind: "message", T: Ref },
    { no: 2, name: "body", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
    { no: 3, name: "caller", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PublishEventRequest {
    return new PublishEventRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PublishEventRequest {
    return new PublishEventRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PublishEventRequest {
    return new PublishEventRequest().fromJsonString(jsonString, options);
  }

  static equals(a: PublishEventRequest | PlainMessage<PublishEventRequest> | undefined, b: PublishEventRequest | PlainMessage<PublishEventRequest> | undefined): boolean {
    return proto3.util.equals(PublishEventRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.PublishEventResponse
 */
export class PublishEventResponse extends Message<PublishEventResponse> {
  constructor(data?: PartialMessage<PublishEventResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.PublishEventResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PublishEventResponse {
    return new PublishEventResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PublishEventResponse {
    return new PublishEventResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PublishEventResponse {
    return new PublishEventResponse().fromJsonString(jsonString, options);
  }

  static equals(a: PublishEventResponse | PlainMessage<PublishEventResponse> | undefined, b: PublishEventResponse | PlainMessage<PublishEventResponse> | undefined): boolean {
    return proto3.util.equals(PublishEventResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ModuleContextRequest
 */
export class ModuleContextRequest extends Message<ModuleContextRequest> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<ModuleContextRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ModuleContextRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleContextRequest {
    return new ModuleContextRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleContextRequest {
    return new ModuleContextRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleContextRequest {
    return new ModuleContextRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleContextRequest | PlainMessage<ModuleContextRequest> | undefined, b: ModuleContextRequest | PlainMessage<ModuleContextRequest> | undefined): boolean {
    return proto3.util.equals(ModuleContextRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ModuleContextResponse
 */
export class ModuleContextResponse extends Message<ModuleContextResponse> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  /**
   * @generated from field: map<string, bytes> configs = 2;
   */
  configs: { [key: string]: Uint8Array } = {};

  /**
   * @generated from field: map<string, bytes> secrets = 3;
   */
  secrets: { [key: string]: Uint8Array } = {};

  /**
   * @generated from field: repeated xyz.block.ftl.v1.ModuleContextResponse.DSN databases = 4;
   */
  databases: ModuleContextResponse_DSN[] = [];

  constructor(data?: PartialMessage<ModuleContextResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ModuleContextResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "configs", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 12 /* ScalarType.BYTES */} },
    { no: 3, name: "secrets", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 12 /* ScalarType.BYTES */} },
    { no: 4, name: "databases", kind: "message", T: ModuleContextResponse_DSN, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleContextResponse {
    return new ModuleContextResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleContextResponse {
    return new ModuleContextResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleContextResponse {
    return new ModuleContextResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleContextResponse | PlainMessage<ModuleContextResponse> | undefined, b: ModuleContextResponse | PlainMessage<ModuleContextResponse> | undefined): boolean {
    return proto3.util.equals(ModuleContextResponse, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.v1.ModuleContextResponse.DBType
 */
export enum ModuleContextResponse_DBType {
  /**
   * @generated from enum value: POSTGRES = 0;
   */
  POSTGRES = 0,

  /**
   * @generated from enum value: MYSQL = 1;
   */
  MYSQL = 1,
}
// Retrieve enum metadata with: proto3.getEnumType(ModuleContextResponse_DBType)
proto3.util.setEnumType(ModuleContextResponse_DBType, "xyz.block.ftl.v1.ModuleContextResponse.DBType", [
  { no: 0, name: "POSTGRES" },
  { no: 1, name: "MYSQL" },
]);

/**
 * @generated from message xyz.block.ftl.v1.ModuleContextResponse.Ref
 */
export class ModuleContextResponse_Ref extends Message<ModuleContextResponse_Ref> {
  /**
   * @generated from field: optional string module = 1;
   */
  module?: string;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  constructor(data?: PartialMessage<ModuleContextResponse_Ref>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ModuleContextResponse.Ref";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleContextResponse_Ref {
    return new ModuleContextResponse_Ref().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleContextResponse_Ref {
    return new ModuleContextResponse_Ref().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleContextResponse_Ref {
    return new ModuleContextResponse_Ref().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleContextResponse_Ref | PlainMessage<ModuleContextResponse_Ref> | undefined, b: ModuleContextResponse_Ref | PlainMessage<ModuleContextResponse_Ref> | undefined): boolean {
    return proto3.util.equals(ModuleContextResponse_Ref, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ModuleContextResponse.DSN
 */
export class ModuleContextResponse_DSN extends Message<ModuleContextResponse_DSN> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: xyz.block.ftl.v1.ModuleContextResponse.DBType type = 2;
   */
  type = ModuleContextResponse_DBType.POSTGRES;

  /**
   * @generated from field: string dsn = 3;
   */
  dsn = "";

  constructor(data?: PartialMessage<ModuleContextResponse_DSN>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ModuleContextResponse.DSN";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "type", kind: "enum", T: proto3.getEnumType(ModuleContextResponse_DBType) },
    { no: 3, name: "dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleContextResponse_DSN {
    return new ModuleContextResponse_DSN().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleContextResponse_DSN {
    return new ModuleContextResponse_DSN().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleContextResponse_DSN {
    return new ModuleContextResponse_DSN().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleContextResponse_DSN | PlainMessage<ModuleContextResponse_DSN> | undefined, b: ModuleContextResponse_DSN | PlainMessage<ModuleContextResponse_DSN> | undefined): boolean {
    return proto3.util.equals(ModuleContextResponse_DSN, a, b);
  }
}

