// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/module.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Duration, Message, proto3 } from "@bufbuild/protobuf";
import { Ref } from "../schema/v1/schema_pb.js";

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
   * @generated from field: xyz.block.ftl.schema.v1.Ref topic = 1;
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
 * @generated from message xyz.block.ftl.v1.GetModuleContextRequest
 */
export class GetModuleContextRequest extends Message<GetModuleContextRequest> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<GetModuleContextRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetModuleContextRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleContextRequest {
    return new GetModuleContextRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleContextRequest {
    return new GetModuleContextRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleContextRequest {
    return new GetModuleContextRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleContextRequest | PlainMessage<GetModuleContextRequest> | undefined, b: GetModuleContextRequest | PlainMessage<GetModuleContextRequest> | undefined): boolean {
    return proto3.util.equals(GetModuleContextRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetModuleContextResponse
 */
export class GetModuleContextResponse extends Message<GetModuleContextResponse> {
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
   * @generated from field: repeated xyz.block.ftl.v1.GetModuleContextResponse.DSN databases = 4;
   */
  databases: GetModuleContextResponse_DSN[] = [];

  constructor(data?: PartialMessage<GetModuleContextResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetModuleContextResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "configs", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 12 /* ScalarType.BYTES */} },
    { no: 3, name: "secrets", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 12 /* ScalarType.BYTES */} },
    { no: 4, name: "databases", kind: "message", T: GetModuleContextResponse_DSN, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleContextResponse {
    return new GetModuleContextResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleContextResponse {
    return new GetModuleContextResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleContextResponse {
    return new GetModuleContextResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleContextResponse | PlainMessage<GetModuleContextResponse> | undefined, b: GetModuleContextResponse | PlainMessage<GetModuleContextResponse> | undefined): boolean {
    return proto3.util.equals(GetModuleContextResponse, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.v1.GetModuleContextResponse.DbType
 */
export enum GetModuleContextResponse_DbType {
  /**
   * @generated from enum value: DB_TYPE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: DB_TYPE_POSTGRES = 1;
   */
  POSTGRES = 1,

  /**
   * @generated from enum value: DB_TYPE_MYSQL = 2;
   */
  MYSQL = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(GetModuleContextResponse_DbType)
proto3.util.setEnumType(GetModuleContextResponse_DbType, "xyz.block.ftl.v1.GetModuleContextResponse.DbType", [
  { no: 0, name: "DB_TYPE_UNSPECIFIED" },
  { no: 1, name: "DB_TYPE_POSTGRES" },
  { no: 2, name: "DB_TYPE_MYSQL" },
]);

/**
 * @generated from message xyz.block.ftl.v1.GetModuleContextResponse.Ref
 */
export class GetModuleContextResponse_Ref extends Message<GetModuleContextResponse_Ref> {
  /**
   * @generated from field: optional string module = 1;
   */
  module?: string;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  constructor(data?: PartialMessage<GetModuleContextResponse_Ref>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetModuleContextResponse.Ref";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleContextResponse_Ref {
    return new GetModuleContextResponse_Ref().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleContextResponse_Ref {
    return new GetModuleContextResponse_Ref().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleContextResponse_Ref {
    return new GetModuleContextResponse_Ref().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleContextResponse_Ref | PlainMessage<GetModuleContextResponse_Ref> | undefined, b: GetModuleContextResponse_Ref | PlainMessage<GetModuleContextResponse_Ref> | undefined): boolean {
    return proto3.util.equals(GetModuleContextResponse_Ref, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetModuleContextResponse.DSN
 */
export class GetModuleContextResponse_DSN extends Message<GetModuleContextResponse_DSN> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: xyz.block.ftl.v1.GetModuleContextResponse.DbType type = 2;
   */
  type = GetModuleContextResponse_DbType.UNSPECIFIED;

  /**
   * @generated from field: string dsn = 3;
   */
  dsn = "";

  constructor(data?: PartialMessage<GetModuleContextResponse_DSN>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetModuleContextResponse.DSN";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "type", kind: "enum", T: proto3.getEnumType(GetModuleContextResponse_DbType) },
    { no: 3, name: "dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleContextResponse_DSN {
    return new GetModuleContextResponse_DSN().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleContextResponse_DSN {
    return new GetModuleContextResponse_DSN().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleContextResponse_DSN {
    return new GetModuleContextResponse_DSN().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleContextResponse_DSN | PlainMessage<GetModuleContextResponse_DSN> | undefined, b: GetModuleContextResponse_DSN | PlainMessage<GetModuleContextResponse_DSN> | undefined): boolean {
    return proto3.util.equals(GetModuleContextResponse_DSN, a, b);
  }
}

