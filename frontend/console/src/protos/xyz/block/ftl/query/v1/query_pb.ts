// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/query/v1/query.proto (package xyz.block.ftl.query.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from enum xyz.block.ftl.query.v1.TransactionStatus
 */
export enum TransactionStatus {
  /**
   * @generated from enum value: TRANSACTION_STATUS_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: TRANSACTION_STATUS_SUCCESS = 1;
   */
  SUCCESS = 1,

  /**
   * @generated from enum value: TRANSACTION_STATUS_FAILED = 2;
   */
  FAILED = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(TransactionStatus)
proto3.util.setEnumType(TransactionStatus, "xyz.block.ftl.query.v1.TransactionStatus", [
  { no: 0, name: "TRANSACTION_STATUS_UNSPECIFIED" },
  { no: 1, name: "TRANSACTION_STATUS_SUCCESS" },
  { no: 2, name: "TRANSACTION_STATUS_FAILED" },
]);

/**
 * @generated from enum xyz.block.ftl.query.v1.CommandType
 */
export enum CommandType {
  /**
   * @generated from enum value: COMMAND_TYPE_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: COMMAND_TYPE_EXEC = 1;
   */
  EXEC = 1,

  /**
   * @generated from enum value: COMMAND_TYPE_ONE = 2;
   */
  ONE = 2,

  /**
   * @generated from enum value: COMMAND_TYPE_MANY = 3;
   */
  MANY = 3,
}
// Retrieve enum metadata with: proto3.getEnumType(CommandType)
proto3.util.setEnumType(CommandType, "xyz.block.ftl.query.v1.CommandType", [
  { no: 0, name: "COMMAND_TYPE_UNSPECIFIED" },
  { no: 1, name: "COMMAND_TYPE_EXEC" },
  { no: 2, name: "COMMAND_TYPE_ONE" },
  { no: 3, name: "COMMAND_TYPE_MANY" },
]);

/**
 * @generated from message xyz.block.ftl.query.v1.BeginTransactionRequest
 */
export class BeginTransactionRequest extends Message<BeginTransactionRequest> {
  constructor(data?: PartialMessage<BeginTransactionRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.BeginTransactionRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BeginTransactionRequest {
    return new BeginTransactionRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BeginTransactionRequest {
    return new BeginTransactionRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BeginTransactionRequest {
    return new BeginTransactionRequest().fromJsonString(jsonString, options);
  }

  static equals(a: BeginTransactionRequest | PlainMessage<BeginTransactionRequest> | undefined, b: BeginTransactionRequest | PlainMessage<BeginTransactionRequest> | undefined): boolean {
    return proto3.util.equals(BeginTransactionRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.BeginTransactionResponse
 */
export class BeginTransactionResponse extends Message<BeginTransactionResponse> {
  /**
   * @generated from field: string transaction_id = 1;
   */
  transactionId = "";

  /**
   * @generated from field: xyz.block.ftl.query.v1.TransactionStatus status = 2;
   */
  status = TransactionStatus.UNSPECIFIED;

  constructor(data?: PartialMessage<BeginTransactionResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.BeginTransactionResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "transaction_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "status", kind: "enum", T: proto3.getEnumType(TransactionStatus) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): BeginTransactionResponse {
    return new BeginTransactionResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): BeginTransactionResponse {
    return new BeginTransactionResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): BeginTransactionResponse {
    return new BeginTransactionResponse().fromJsonString(jsonString, options);
  }

  static equals(a: BeginTransactionResponse | PlainMessage<BeginTransactionResponse> | undefined, b: BeginTransactionResponse | PlainMessage<BeginTransactionResponse> | undefined): boolean {
    return proto3.util.equals(BeginTransactionResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.CommitTransactionRequest
 */
export class CommitTransactionRequest extends Message<CommitTransactionRequest> {
  /**
   * @generated from field: string transaction_id = 1;
   */
  transactionId = "";

  constructor(data?: PartialMessage<CommitTransactionRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.CommitTransactionRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "transaction_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CommitTransactionRequest {
    return new CommitTransactionRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CommitTransactionRequest {
    return new CommitTransactionRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CommitTransactionRequest {
    return new CommitTransactionRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CommitTransactionRequest | PlainMessage<CommitTransactionRequest> | undefined, b: CommitTransactionRequest | PlainMessage<CommitTransactionRequest> | undefined): boolean {
    return proto3.util.equals(CommitTransactionRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.CommitTransactionResponse
 */
export class CommitTransactionResponse extends Message<CommitTransactionResponse> {
  /**
   * @generated from field: xyz.block.ftl.query.v1.TransactionStatus status = 1;
   */
  status = TransactionStatus.UNSPECIFIED;

  constructor(data?: PartialMessage<CommitTransactionResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.CommitTransactionResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "status", kind: "enum", T: proto3.getEnumType(TransactionStatus) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CommitTransactionResponse {
    return new CommitTransactionResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CommitTransactionResponse {
    return new CommitTransactionResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CommitTransactionResponse {
    return new CommitTransactionResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CommitTransactionResponse | PlainMessage<CommitTransactionResponse> | undefined, b: CommitTransactionResponse | PlainMessage<CommitTransactionResponse> | undefined): boolean {
    return proto3.util.equals(CommitTransactionResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.RollbackTransactionRequest
 */
export class RollbackTransactionRequest extends Message<RollbackTransactionRequest> {
  /**
   * @generated from field: string transaction_id = 1;
   */
  transactionId = "";

  constructor(data?: PartialMessage<RollbackTransactionRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.RollbackTransactionRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "transaction_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RollbackTransactionRequest {
    return new RollbackTransactionRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RollbackTransactionRequest {
    return new RollbackTransactionRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RollbackTransactionRequest {
    return new RollbackTransactionRequest().fromJsonString(jsonString, options);
  }

  static equals(a: RollbackTransactionRequest | PlainMessage<RollbackTransactionRequest> | undefined, b: RollbackTransactionRequest | PlainMessage<RollbackTransactionRequest> | undefined): boolean {
    return proto3.util.equals(RollbackTransactionRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.RollbackTransactionResponse
 */
export class RollbackTransactionResponse extends Message<RollbackTransactionResponse> {
  /**
   * @generated from field: xyz.block.ftl.query.v1.TransactionStatus status = 1;
   */
  status = TransactionStatus.UNSPECIFIED;

  constructor(data?: PartialMessage<RollbackTransactionResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.RollbackTransactionResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "status", kind: "enum", T: proto3.getEnumType(TransactionStatus) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RollbackTransactionResponse {
    return new RollbackTransactionResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RollbackTransactionResponse {
    return new RollbackTransactionResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RollbackTransactionResponse {
    return new RollbackTransactionResponse().fromJsonString(jsonString, options);
  }

  static equals(a: RollbackTransactionResponse | PlainMessage<RollbackTransactionResponse> | undefined, b: RollbackTransactionResponse | PlainMessage<RollbackTransactionResponse> | undefined): boolean {
    return proto3.util.equals(RollbackTransactionResponse, a, b);
  }
}

/**
 * A value that can be used as a SQL parameter
 *
 * @generated from message xyz.block.ftl.query.v1.SQLValue
 */
export class SQLValue extends Message<SQLValue> {
  /**
   * @generated from oneof xyz.block.ftl.query.v1.SQLValue.value
   */
  value: {
    /**
     * @generated from field: string string_value = 1;
     */
    value: string;
    case: "stringValue";
  } | {
    /**
     * @generated from field: int64 int_value = 2;
     */
    value: bigint;
    case: "intValue";
  } | {
    /**
     * @generated from field: double float_value = 3;
     */
    value: number;
    case: "floatValue";
  } | {
    /**
     * @generated from field: bool bool_value = 4;
     */
    value: boolean;
    case: "boolValue";
  } | {
    /**
     * @generated from field: bytes bytes_value = 5;
     */
    value: Uint8Array;
    case: "bytesValue";
  } | {
    /**
     * @generated from field: google.protobuf.Timestamp timestamp_value = 6;
     */
    value: Timestamp;
    case: "timestampValue";
  } | {
    /**
     * Set to true to represent NULL
     *
     * @generated from field: bool null_value = 7;
     */
    value: boolean;
    case: "nullValue";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<SQLValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.SQLValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "string_value", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "value" },
    { no: 2, name: "int_value", kind: "scalar", T: 3 /* ScalarType.INT64 */, oneof: "value" },
    { no: 3, name: "float_value", kind: "scalar", T: 1 /* ScalarType.DOUBLE */, oneof: "value" },
    { no: 4, name: "bool_value", kind: "scalar", T: 8 /* ScalarType.BOOL */, oneof: "value" },
    { no: 5, name: "bytes_value", kind: "scalar", T: 12 /* ScalarType.BYTES */, oneof: "value" },
    { no: 6, name: "timestamp_value", kind: "message", T: Timestamp, oneof: "value" },
    { no: 7, name: "null_value", kind: "scalar", T: 8 /* ScalarType.BOOL */, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SQLValue {
    return new SQLValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SQLValue {
    return new SQLValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SQLValue {
    return new SQLValue().fromJsonString(jsonString, options);
  }

  static equals(a: SQLValue | PlainMessage<SQLValue> | undefined, b: SQLValue | PlainMessage<SQLValue> | undefined): boolean {
    return proto3.util.equals(SQLValue, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.ExecuteQueryRequest
 */
export class ExecuteQueryRequest extends Message<ExecuteQueryRequest> {
  /**
   * @generated from field: string raw_sql = 1;
   */
  rawSql = "";

  /**
   * @generated from field: xyz.block.ftl.query.v1.CommandType command_type = 2;
   */
  commandType = CommandType.UNSPECIFIED;

  /**
   * SQL parameter values in order
   *
   * @generated from field: repeated xyz.block.ftl.query.v1.SQLValue parameters = 3;
   */
  parameters: SQLValue[] = [];

  /**
   * Column names to scan for the result type
   *
   * @generated from field: repeated string result_columns = 6;
   */
  resultColumns: string[] = [];

  /**
   * @generated from field: optional string transaction_id = 4;
   */
  transactionId?: string;

  /**
   * Default 100 if not set
   *
   * @generated from field: optional int32 batch_size = 5;
   */
  batchSize?: number;

  constructor(data?: PartialMessage<ExecuteQueryRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.ExecuteQueryRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "raw_sql", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "command_type", kind: "enum", T: proto3.getEnumType(CommandType) },
    { no: 3, name: "parameters", kind: "message", T: SQLValue, repeated: true },
    { no: 6, name: "result_columns", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "transaction_id", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 5, name: "batch_size", kind: "scalar", T: 5 /* ScalarType.INT32 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecuteQueryRequest {
    return new ExecuteQueryRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecuteQueryRequest {
    return new ExecuteQueryRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecuteQueryRequest {
    return new ExecuteQueryRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ExecuteQueryRequest | PlainMessage<ExecuteQueryRequest> | undefined, b: ExecuteQueryRequest | PlainMessage<ExecuteQueryRequest> | undefined): boolean {
    return proto3.util.equals(ExecuteQueryRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.ExecuteQueryResponse
 */
export class ExecuteQueryResponse extends Message<ExecuteQueryResponse> {
  /**
   * @generated from oneof xyz.block.ftl.query.v1.ExecuteQueryResponse.result
   */
  result: {
    /**
     * For EXEC commands
     *
     * @generated from field: xyz.block.ftl.query.v1.ExecResult exec_result = 1;
     */
    value: ExecResult;
    case: "execResult";
  } | {
    /**
     * For ONE/MANY commands
     *
     * @generated from field: xyz.block.ftl.query.v1.RowResults row_results = 2;
     */
    value: RowResults;
    case: "rowResults";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<ExecuteQueryResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.ExecuteQueryResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "exec_result", kind: "message", T: ExecResult, oneof: "result" },
    { no: 2, name: "row_results", kind: "message", T: RowResults, oneof: "result" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecuteQueryResponse {
    return new ExecuteQueryResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecuteQueryResponse {
    return new ExecuteQueryResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecuteQueryResponse {
    return new ExecuteQueryResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ExecuteQueryResponse | PlainMessage<ExecuteQueryResponse> | undefined, b: ExecuteQueryResponse | PlainMessage<ExecuteQueryResponse> | undefined): boolean {
    return proto3.util.equals(ExecuteQueryResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.ExecResult
 */
export class ExecResult extends Message<ExecResult> {
  /**
   * @generated from field: int64 rows_affected = 1;
   */
  rowsAffected = protoInt64.zero;

  /**
   * Only for some databases like MySQL
   *
   * @generated from field: optional int64 last_insert_id = 2;
   */
  lastInsertId?: bigint;

  constructor(data?: PartialMessage<ExecResult>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.ExecResult";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "rows_affected", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
    { no: 2, name: "last_insert_id", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExecResult {
    return new ExecResult().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExecResult {
    return new ExecResult().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExecResult {
    return new ExecResult().fromJsonString(jsonString, options);
  }

  static equals(a: ExecResult | PlainMessage<ExecResult> | undefined, b: ExecResult | PlainMessage<ExecResult> | undefined): boolean {
    return proto3.util.equals(ExecResult, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.query.v1.RowResults
 */
export class RowResults extends Message<RowResults> {
  /**
   * Each row is a map of column name to value
   *
   * @generated from field: map<string, xyz.block.ftl.query.v1.SQLValue> rows = 1;
   */
  rows: { [key: string]: SQLValue } = {};

  /**
   * Indicates if there are more rows to fetch
   *
   * @generated from field: bool has_more = 2;
   */
  hasMore = false;

  constructor(data?: PartialMessage<RowResults>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.query.v1.RowResults";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "rows", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: SQLValue} },
    { no: 2, name: "has_more", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RowResults {
    return new RowResults().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RowResults {
    return new RowResults().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RowResults {
    return new RowResults().fromJsonString(jsonString, options);
  }

  static equals(a: RowResults | PlainMessage<RowResults> | undefined, b: RowResults | PlainMessage<RowResults> | undefined): boolean {
    return proto3.util.equals(RowResults, a, b);
  }
}

