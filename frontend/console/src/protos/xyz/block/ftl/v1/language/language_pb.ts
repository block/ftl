// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/language/language.proto (package xyz.block.ftl.v1.language, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64 } from "@bufbuild/protobuf";
import { Position } from "../schema/schema_pb.js";

/**
 * @generated from message xyz.block.ftl.v1.language.Error
 */
export class Error extends Message<Error> {
  /**
   * @generated from field: string msg = 1;
   */
  msg = "";

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Position pos = 2;
   */
  pos?: Position;

  /**
   * @generated from field: int64 endColumn = 3;
   */
  endColumn = protoInt64.zero;

  /**
   * @generated from field: xyz.block.ftl.v1.language.Error.ErrorLevel level = 4;
   */
  level = Error_ErrorLevel.INFO;

  constructor(data?: PartialMessage<Error>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.language.Error";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "msg", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "pos", kind: "message", T: Position },
    { no: 3, name: "endColumn", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
    { no: 4, name: "level", kind: "enum", T: proto3.getEnumType(Error_ErrorLevel) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Error {
    return new Error().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Error {
    return new Error().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Error {
    return new Error().fromJsonString(jsonString, options);
  }

  static equals(a: Error | PlainMessage<Error> | undefined, b: Error | PlainMessage<Error> | undefined): boolean {
    return proto3.util.equals(Error, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.v1.language.Error.ErrorLevel
 */
export enum Error_ErrorLevel {
  /**
   * @generated from enum value: INFO = 0;
   */
  INFO = 0,

  /**
   * @generated from enum value: WARN = 1;
   */
  WARN = 1,

  /**
   * @generated from enum value: ERROR = 2;
   */
  ERROR = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(Error_ErrorLevel)
proto3.util.setEnumType(Error_ErrorLevel, "xyz.block.ftl.v1.language.Error.ErrorLevel", [
  { no: 0, name: "INFO" },
  { no: 1, name: "WARN" },
  { no: 2, name: "ERROR" },
]);

/**
 * @generated from message xyz.block.ftl.v1.language.ErrorList
 */
export class ErrorList extends Message<ErrorList> {
  /**
   * @generated from field: repeated xyz.block.ftl.v1.language.Error errors = 1;
   */
  errors: Error[] = [];

  constructor(data?: PartialMessage<ErrorList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.language.ErrorList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "errors", kind: "message", T: Error, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ErrorList {
    return new ErrorList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ErrorList {
    return new ErrorList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ErrorList {
    return new ErrorList().fromJsonString(jsonString, options);
  }

  static equals(a: ErrorList | PlainMessage<ErrorList> | undefined, b: ErrorList | PlainMessage<ErrorList> | undefined): boolean {
    return proto3.util.equals(ErrorList, a, b);
  }
}

