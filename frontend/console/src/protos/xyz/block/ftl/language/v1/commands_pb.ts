// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/language/v1/commands.proto (package xyz.block.ftl.language.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Struct } from "@bufbuild/protobuf";
import { ProjectConfig } from "./service_pb.js";

/**
 * @generated from message xyz.block.ftl.language.v1.GetNewModuleFlagsRequest
 */
export class GetNewModuleFlagsRequest extends Message<GetNewModuleFlagsRequest> {
  /**
   * @generated from field: string language = 1;
   */
  language = "";

  constructor(data?: PartialMessage<GetNewModuleFlagsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.GetNewModuleFlagsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "language", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetNewModuleFlagsRequest {
    return new GetNewModuleFlagsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetNewModuleFlagsRequest {
    return new GetNewModuleFlagsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetNewModuleFlagsRequest {
    return new GetNewModuleFlagsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetNewModuleFlagsRequest | PlainMessage<GetNewModuleFlagsRequest> | undefined, b: GetNewModuleFlagsRequest | PlainMessage<GetNewModuleFlagsRequest> | undefined): boolean {
    return proto3.util.equals(GetNewModuleFlagsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.language.v1.GetNewModuleFlagsResponse
 */
export class GetNewModuleFlagsResponse extends Message<GetNewModuleFlagsResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.Flag flags = 1;
   */
  flags: GetNewModuleFlagsResponse_Flag[] = [];

  constructor(data?: PartialMessage<GetNewModuleFlagsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.GetNewModuleFlagsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "flags", kind: "message", T: GetNewModuleFlagsResponse_Flag, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetNewModuleFlagsResponse {
    return new GetNewModuleFlagsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetNewModuleFlagsResponse {
    return new GetNewModuleFlagsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetNewModuleFlagsResponse {
    return new GetNewModuleFlagsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetNewModuleFlagsResponse | PlainMessage<GetNewModuleFlagsResponse> | undefined, b: GetNewModuleFlagsResponse | PlainMessage<GetNewModuleFlagsResponse> | undefined): boolean {
    return proto3.util.equals(GetNewModuleFlagsResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.Flag
 */
export class GetNewModuleFlagsResponse_Flag extends Message<GetNewModuleFlagsResponse_Flag> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: string help = 2;
   */
  help = "";

  /**
   * @generated from field: optional string envar = 3;
   */
  envar?: string;

  /**
   * short must be a single character
   *
   * @generated from field: optional string short = 4;
   */
  short?: string;

  /**
   * @generated from field: optional string placeholder = 5;
   */
  placeholder?: string;

  /**
   * @generated from field: optional string default = 6;
   */
  default?: string;

  constructor(data?: PartialMessage<GetNewModuleFlagsResponse_Flag>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.Flag";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "help", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "envar", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 4, name: "short", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 5, name: "placeholder", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 6, name: "default", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetNewModuleFlagsResponse_Flag {
    return new GetNewModuleFlagsResponse_Flag().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetNewModuleFlagsResponse_Flag {
    return new GetNewModuleFlagsResponse_Flag().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetNewModuleFlagsResponse_Flag {
    return new GetNewModuleFlagsResponse_Flag().fromJsonString(jsonString, options);
  }

  static equals(a: GetNewModuleFlagsResponse_Flag | PlainMessage<GetNewModuleFlagsResponse_Flag> | undefined, b: GetNewModuleFlagsResponse_Flag | PlainMessage<GetNewModuleFlagsResponse_Flag> | undefined): boolean {
    return proto3.util.equals(GetNewModuleFlagsResponse_Flag, a, b);
  }
}

/**
 * Request to create a new module.
 *
 * @generated from message xyz.block.ftl.language.v1.NewModuleRequest
 */
export class NewModuleRequest extends Message<NewModuleRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * The root directory for the module, which does not yet exist.
   * The plugin should create the directory.
   *
   * @generated from field: string dir = 2;
   */
  dir = "";

  /**
   * The project configuration
   *
   * @generated from field: xyz.block.ftl.language.v1.ProjectConfig project_config = 3;
   */
  projectConfig?: ProjectConfig;

  /**
   * Flags contains any values set for those configured in the GetCreateModuleFlags call
   *
   * @generated from field: google.protobuf.Struct flags = 4;
   */
  flags?: Struct;

  constructor(data?: PartialMessage<NewModuleRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.NewModuleRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "dir", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "project_config", kind: "message", T: ProjectConfig },
    { no: 4, name: "flags", kind: "message", T: Struct },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): NewModuleRequest {
    return new NewModuleRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): NewModuleRequest {
    return new NewModuleRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): NewModuleRequest {
    return new NewModuleRequest().fromJsonString(jsonString, options);
  }

  static equals(a: NewModuleRequest | PlainMessage<NewModuleRequest> | undefined, b: NewModuleRequest | PlainMessage<NewModuleRequest> | undefined): boolean {
    return proto3.util.equals(NewModuleRequest, a, b);
  }
}

/**
 * Response to a create module request.
 *
 * @generated from message xyz.block.ftl.language.v1.NewModuleResponse
 */
export class NewModuleResponse extends Message<NewModuleResponse> {
  constructor(data?: PartialMessage<NewModuleResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.NewModuleResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): NewModuleResponse {
    return new NewModuleResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): NewModuleResponse {
    return new NewModuleResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): NewModuleResponse {
    return new NewModuleResponse().fromJsonString(jsonString, options);
  }

  static equals(a: NewModuleResponse | PlainMessage<NewModuleResponse> | undefined, b: NewModuleResponse | PlainMessage<NewModuleResponse> | undefined): boolean {
    return proto3.util.equals(NewModuleResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.language.v1.GetModuleConfigDefaultsRequest
 */
export class GetModuleConfigDefaultsRequest extends Message<GetModuleConfigDefaultsRequest> {
  /**
   * @generated from field: string dir = 1;
   */
  dir = "";

  constructor(data?: PartialMessage<GetModuleConfigDefaultsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.GetModuleConfigDefaultsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "dir", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleConfigDefaultsRequest {
    return new GetModuleConfigDefaultsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleConfigDefaultsRequest {
    return new GetModuleConfigDefaultsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleConfigDefaultsRequest {
    return new GetModuleConfigDefaultsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleConfigDefaultsRequest | PlainMessage<GetModuleConfigDefaultsRequest> | undefined, b: GetModuleConfigDefaultsRequest | PlainMessage<GetModuleConfigDefaultsRequest> | undefined): boolean {
    return proto3.util.equals(GetModuleConfigDefaultsRequest, a, b);
  }
}

/**
 * GetModuleConfigDefaultsResponse provides defaults for ModuleConfig.
 *
 * The result may be cached by FTL, so defaulting logic should not be changing due to normal module changes.
 * For example, it is valid to return defaults based on which build tool is configured within the module directory,
 * as that is not expected to change during normal operation.
 * It is not recommended to read the module's toml file to determine defaults, as when the toml file is updated,
 * the module defaults will not be recalculated.
 *
 * @generated from message xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse
 */
export class GetModuleConfigDefaultsResponse extends Message<GetModuleConfigDefaultsResponse> {
  /**
   * Default relative path to the directory containing all build artifacts for deployments
   *
   * @generated from field: string deploy_dir = 1;
   */
  deployDir = "";

  /**
   * Default build command
   *
   * @generated from field: optional string build = 2;
   */
  build?: string;

  /**
   * Dev mode build command, if different from the regular build command
   *
   * @generated from field: optional string dev_mode_build = 3;
   */
  devModeBuild?: string;

  /**
   * Build lock path to prevent concurrent builds
   *
   * @generated from field: optional string build_lock = 4;
   */
  buildLock?: string;

  /**
   * Default patterns to watch for file changes, relative to the module directory
   *
   * @generated from field: repeated string watch = 6;
   */
  watch: string[] = [];

  /**
   * Default language specific configuration.
   * These defaults are filled in by looking at each root key only. If the key is not present, the default is used.
   *
   * @generated from field: google.protobuf.Struct language_config = 7;
   */
  languageConfig?: Struct;

  /**
   * Root directory containing SQL files.
   *
   * @generated from field: string sql_root_dir = 8;
   */
  sqlRootDir = "";

  constructor(data?: PartialMessage<GetModuleConfigDefaultsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deploy_dir", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "build", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "dev_mode_build", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 4, name: "build_lock", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 6, name: "watch", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 7, name: "language_config", kind: "message", T: Struct },
    { no: 8, name: "sql_root_dir", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModuleConfigDefaultsResponse {
    return new GetModuleConfigDefaultsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModuleConfigDefaultsResponse {
    return new GetModuleConfigDefaultsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModuleConfigDefaultsResponse {
    return new GetModuleConfigDefaultsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetModuleConfigDefaultsResponse | PlainMessage<GetModuleConfigDefaultsResponse> | undefined, b: GetModuleConfigDefaultsResponse | PlainMessage<GetModuleConfigDefaultsResponse> | undefined): boolean {
    return proto3.util.equals(GetModuleConfigDefaultsResponse, a, b);
  }
}

