// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/console/v1/console.proto (package xyz.block.ftl.console.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Config as Config$1, Data as Data$1, Database as Database$1, Enum as Enum$1, Ref, Secret as Secret$1, Topic as Topic$1, TypeAlias as TypeAlias$1, Verb as Verb$1 } from "../../schema/v1/schema_pb.js";

/**
 * @generated from message xyz.block.ftl.console.v1.Config
 */
export class Config extends Message<Config> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Config config = 1;
   */
  config?: Config$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<Config>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Config";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: Config$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Config {
    return new Config().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Config {
    return new Config().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Config {
    return new Config().fromJsonString(jsonString, options);
  }

  static equals(a: Config | PlainMessage<Config> | undefined, b: Config | PlainMessage<Config> | undefined): boolean {
    return proto3.util.equals(Config, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Data
 */
export class Data extends Message<Data> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Data data = 1;
   */
  data?: Data$1;

  /**
   * @generated from field: string schema = 2;
   */
  schema = "";

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 3;
   */
  references: Ref[] = [];

  constructor(data?: PartialMessage<Data>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Data";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "data", kind: "message", T: Data$1 },
    { no: 2, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "references", kind: "message", T: Ref, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Data {
    return new Data().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Data {
    return new Data().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Data {
    return new Data().fromJsonString(jsonString, options);
  }

  static equals(a: Data | PlainMessage<Data> | undefined, b: Data | PlainMessage<Data> | undefined): boolean {
    return proto3.util.equals(Data, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Database
 */
export class Database extends Message<Database> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Database database = 1;
   */
  database?: Database$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<Database>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Database";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "database", kind: "message", T: Database$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Database {
    return new Database().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Database {
    return new Database().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Database {
    return new Database().fromJsonString(jsonString, options);
  }

  static equals(a: Database | PlainMessage<Database> | undefined, b: Database | PlainMessage<Database> | undefined): boolean {
    return proto3.util.equals(Database, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Enum
 */
export class Enum extends Message<Enum> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Enum enum = 1;
   */
  enum?: Enum$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<Enum>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Enum";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "enum", kind: "message", T: Enum$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Enum {
    return new Enum().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Enum {
    return new Enum().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Enum {
    return new Enum().fromJsonString(jsonString, options);
  }

  static equals(a: Enum | PlainMessage<Enum> | undefined, b: Enum | PlainMessage<Enum> | undefined): boolean {
    return proto3.util.equals(Enum, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Topic
 */
export class Topic extends Message<Topic> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Topic topic = 1;
   */
  topic?: Topic$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<Topic>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Topic";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "topic", kind: "message", T: Topic$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Topic {
    return new Topic().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Topic {
    return new Topic().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Topic {
    return new Topic().fromJsonString(jsonString, options);
  }

  static equals(a: Topic | PlainMessage<Topic> | undefined, b: Topic | PlainMessage<Topic> | undefined): boolean {
    return proto3.util.equals(Topic, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.TypeAlias
 */
export class TypeAlias extends Message<TypeAlias> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.TypeAlias typealias = 1;
   */
  typealias?: TypeAlias$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<TypeAlias>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.TypeAlias";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "typealias", kind: "message", T: TypeAlias$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TypeAlias {
    return new TypeAlias().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TypeAlias {
    return new TypeAlias().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TypeAlias {
    return new TypeAlias().fromJsonString(jsonString, options);
  }

  static equals(a: TypeAlias | PlainMessage<TypeAlias> | undefined, b: TypeAlias | PlainMessage<TypeAlias> | undefined): boolean {
    return proto3.util.equals(TypeAlias, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Secret
 */
export class Secret extends Message<Secret> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Secret secret = 1;
   */
  secret?: Secret$1;

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 2;
   */
  references: Ref[] = [];

  /**
   * @generated from field: string schema = 3;
   */
  schema = "";

  constructor(data?: PartialMessage<Secret>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Secret";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "secret", kind: "message", T: Secret$1 },
    { no: 2, name: "references", kind: "message", T: Ref, repeated: true },
    { no: 3, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Secret {
    return new Secret().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Secret {
    return new Secret().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Secret {
    return new Secret().fromJsonString(jsonString, options);
  }

  static equals(a: Secret | PlainMessage<Secret> | undefined, b: Secret | PlainMessage<Secret> | undefined): boolean {
    return proto3.util.equals(Secret, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Verb
 */
export class Verb extends Message<Verb> {
  /**
   * @generated from field: xyz.block.ftl.schema.v1.Verb verb = 1;
   */
  verb?: Verb$1;

  /**
   * @generated from field: string schema = 2;
   */
  schema = "";

  /**
   * @generated from field: string json_request_schema = 3;
   */
  jsonRequestSchema = "";

  /**
   * @generated from field: repeated xyz.block.ftl.schema.v1.Ref references = 4;
   */
  references: Ref[] = [];

  constructor(data?: PartialMessage<Verb>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Verb";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "verb", kind: "message", T: Verb$1 },
    { no: 2, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "json_request_schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "references", kind: "message", T: Ref, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Verb {
    return new Verb().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Verb {
    return new Verb().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Verb {
    return new Verb().fromJsonString(jsonString, options);
  }

  static equals(a: Verb | PlainMessage<Verb> | undefined, b: Verb | PlainMessage<Verb> | undefined): boolean {
    return proto3.util.equals(Verb, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Module
 */
export class Module extends Message<Module> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: string deployment_key = 2;
   */
  deploymentKey = "";

  /**
   * @generated from field: string language = 3;
   */
  language = "";

  /**
   * @generated from field: string schema = 4;
   */
  schema = "";

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Verb verbs = 5;
   */
  verbs: Verb[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Data data = 6;
   */
  data: Data[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Secret secrets = 7;
   */
  secrets: Secret[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Config configs = 8;
   */
  configs: Config[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Database databases = 9;
   */
  databases: Database[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Enum enums = 10;
   */
  enums: Enum[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Topic topics = 11;
   */
  topics: Topic[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.TypeAlias typealiases = 12;
   */
  typealiases: TypeAlias[] = [];

  constructor(data?: PartialMessage<Module>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Module";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "deployment_key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "language", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "schema", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "verbs", kind: "message", T: Verb, repeated: true },
    { no: 6, name: "data", kind: "message", T: Data, repeated: true },
    { no: 7, name: "secrets", kind: "message", T: Secret, repeated: true },
    { no: 8, name: "configs", kind: "message", T: Config, repeated: true },
    { no: 9, name: "databases", kind: "message", T: Database, repeated: true },
    { no: 10, name: "enums", kind: "message", T: Enum, repeated: true },
    { no: 11, name: "topics", kind: "message", T: Topic, repeated: true },
    { no: 12, name: "typealiases", kind: "message", T: TypeAlias, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Module {
    return new Module().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Module {
    return new Module().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Module {
    return new Module().fromJsonString(jsonString, options);
  }

  static equals(a: Module | PlainMessage<Module> | undefined, b: Module | PlainMessage<Module> | undefined): boolean {
    return proto3.util.equals(Module, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.TopologyGroup
 */
export class TopologyGroup extends Message<TopologyGroup> {
  /**
   * @generated from field: repeated string modules = 1;
   */
  modules: string[] = [];

  constructor(data?: PartialMessage<TopologyGroup>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.TopologyGroup";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "modules", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TopologyGroup {
    return new TopologyGroup().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TopologyGroup {
    return new TopologyGroup().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TopologyGroup {
    return new TopologyGroup().fromJsonString(jsonString, options);
  }

  static equals(a: TopologyGroup | PlainMessage<TopologyGroup> | undefined, b: TopologyGroup | PlainMessage<TopologyGroup> | undefined): boolean {
    return proto3.util.equals(TopologyGroup, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.Topology
 */
export class Topology extends Message<Topology> {
  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.TopologyGroup levels = 1;
   */
  levels: TopologyGroup[] = [];

  constructor(data?: PartialMessage<Topology>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.Topology";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "levels", kind: "message", T: TopologyGroup, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Topology {
    return new Topology().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Topology {
    return new Topology().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Topology {
    return new Topology().fromJsonString(jsonString, options);
  }

  static equals(a: Topology | PlainMessage<Topology> | undefined, b: Topology | PlainMessage<Topology> | undefined): boolean {
    return proto3.util.equals(Topology, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetModulesRequest
 */
export class GetModulesRequest extends Message<GetModulesRequest> {
  constructor(data?: PartialMessage<GetModulesRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetModulesRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModulesRequest {
    return new GetModulesRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModulesRequest {
    return new GetModulesRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModulesRequest {
    return new GetModulesRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetModulesRequest | PlainMessage<GetModulesRequest> | undefined, b: GetModulesRequest | PlainMessage<GetModulesRequest> | undefined): boolean {
    return proto3.util.equals(GetModulesRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetModulesResponse
 */
export class GetModulesResponse extends Message<GetModulesResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Module modules = 1;
   */
  modules: Module[] = [];

  /**
   * @generated from field: xyz.block.ftl.console.v1.Topology topology = 2;
   */
  topology?: Topology;

  constructor(data?: PartialMessage<GetModulesResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetModulesResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "modules", kind: "message", T: Module, repeated: true },
    { no: 2, name: "topology", kind: "message", T: Topology },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetModulesResponse {
    return new GetModulesResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetModulesResponse {
    return new GetModulesResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetModulesResponse {
    return new GetModulesResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetModulesResponse | PlainMessage<GetModulesResponse> | undefined, b: GetModulesResponse | PlainMessage<GetModulesResponse> | undefined): boolean {
    return proto3.util.equals(GetModulesResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.StreamModulesRequest
 */
export class StreamModulesRequest extends Message<StreamModulesRequest> {
  constructor(data?: PartialMessage<StreamModulesRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.StreamModulesRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamModulesRequest {
    return new StreamModulesRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamModulesRequest {
    return new StreamModulesRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamModulesRequest {
    return new StreamModulesRequest().fromJsonString(jsonString, options);
  }

  static equals(a: StreamModulesRequest | PlainMessage<StreamModulesRequest> | undefined, b: StreamModulesRequest | PlainMessage<StreamModulesRequest> | undefined): boolean {
    return proto3.util.equals(StreamModulesRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.StreamModulesResponse
 */
export class StreamModulesResponse extends Message<StreamModulesResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.console.v1.Module modules = 1;
   */
  modules: Module[] = [];

  /**
   * @generated from field: xyz.block.ftl.console.v1.Topology topology = 2;
   */
  topology?: Topology;

  constructor(data?: PartialMessage<StreamModulesResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.StreamModulesResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "modules", kind: "message", T: Module, repeated: true },
    { no: 2, name: "topology", kind: "message", T: Topology },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamModulesResponse {
    return new StreamModulesResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamModulesResponse {
    return new StreamModulesResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamModulesResponse {
    return new StreamModulesResponse().fromJsonString(jsonString, options);
  }

  static equals(a: StreamModulesResponse | PlainMessage<StreamModulesResponse> | undefined, b: StreamModulesResponse | PlainMessage<StreamModulesResponse> | undefined): boolean {
    return proto3.util.equals(StreamModulesResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetConfigRequest
 */
export class GetConfigRequest extends Message<GetConfigRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: optional string module = 2;
   */
  module?: string;

  constructor(data?: PartialMessage<GetConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetConfigRequest | PlainMessage<GetConfigRequest> | undefined, b: GetConfigRequest | PlainMessage<GetConfigRequest> | undefined): boolean {
    return proto3.util.equals(GetConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetConfigResponse
 */
export class GetConfigResponse extends Message<GetConfigResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<GetConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetConfigResponse | PlainMessage<GetConfigResponse> | undefined, b: GetConfigResponse | PlainMessage<GetConfigResponse> | undefined): boolean {
    return proto3.util.equals(GetConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.SetConfigRequest
 */
export class SetConfigRequest extends Message<SetConfigRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: optional string module = 2;
   */
  module?: string;

  /**
   * @generated from field: bytes value = 3;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.SetConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SetConfigRequest | PlainMessage<SetConfigRequest> | undefined, b: SetConfigRequest | PlainMessage<SetConfigRequest> | undefined): boolean {
    return proto3.util.equals(SetConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.SetConfigResponse
 */
export class SetConfigResponse extends Message<SetConfigResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.SetConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SetConfigResponse | PlainMessage<SetConfigResponse> | undefined, b: SetConfigResponse | PlainMessage<SetConfigResponse> | undefined): boolean {
    return proto3.util.equals(SetConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetSecretRequest
 */
export class GetSecretRequest extends Message<GetSecretRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: optional string module = 2;
   */
  module?: string;

  constructor(data?: PartialMessage<GetSecretRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetSecretRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetSecretRequest | PlainMessage<GetSecretRequest> | undefined, b: GetSecretRequest | PlainMessage<GetSecretRequest> | undefined): boolean {
    return proto3.util.equals(GetSecretRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.GetSecretResponse
 */
export class GetSecretResponse extends Message<GetSecretResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<GetSecretResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.GetSecretResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetSecretResponse | PlainMessage<GetSecretResponse> | undefined, b: GetSecretResponse | PlainMessage<GetSecretResponse> | undefined): boolean {
    return proto3.util.equals(GetSecretResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.SetSecretRequest
 */
export class SetSecretRequest extends Message<SetSecretRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: optional string module = 2;
   */
  module?: string;

  /**
   * @generated from field: bytes value = 3;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetSecretRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.SetSecretRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SetSecretRequest | PlainMessage<SetSecretRequest> | undefined, b: SetSecretRequest | PlainMessage<SetSecretRequest> | undefined): boolean {
    return proto3.util.equals(SetSecretRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.console.v1.SetSecretResponse
 */
export class SetSecretResponse extends Message<SetSecretResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetSecretResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.console.v1.SetSecretResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SetSecretResponse | PlainMessage<SetSecretResponse> | undefined, b: SetSecretResponse | PlainMessage<SetSecretResponse> | undefined): boolean {
    return proto3.util.equals(SetSecretResponse, a, b);
  }
}

