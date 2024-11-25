// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1beta1/provisioner/resource.proto (package xyz.block.ftl.v1beta1.provisioner, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Struct } from "@bufbuild/protobuf";
import { Module, Ref } from "../../v1/schema/schema_pb.js";
import { DeploymentArtefact } from "../../v1/controller_pb.js";

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
   * @generated from oneof xyz.block.ftl.v1beta1.provisioner.Resource.resource
   */
  resource: {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.PostgresResource postgres = 102;
     */
    value: PostgresResource;
    case: "postgres";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.MysqlResource mysql = 103;
     */
    value: MysqlResource;
    case: "mysql";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.ModuleResource module = 104;
     */
    value: ModuleResource;
    case: "module";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource sql_migration = 105;
     */
    value: SqlMigrationResource;
    case: "sqlMigration";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.TopicResource topic = 106;
     */
    value: TopicResource;
    case: "topic";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.SubscriptionResource subscription = 107;
     */
    value: SubscriptionResource;
    case: "subscription";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Resource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 102, name: "postgres", kind: "message", T: PostgresResource, oneof: "resource" },
    { no: 103, name: "mysql", kind: "message", T: MysqlResource, oneof: "resource" },
    { no: 104, name: "module", kind: "message", T: ModuleResource, oneof: "resource" },
    { no: 105, name: "sql_migration", kind: "message", T: SqlMigrationResource, oneof: "resource" },
    { no: 106, name: "topic", kind: "message", T: TopicResource, oneof: "resource" },
    { no: 107, name: "subscription", kind: "message", T: SubscriptionResource, oneof: "resource" },
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
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PostgresResource
 */
export class PostgresResource extends Message<PostgresResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput output = 1;
   */
  output?: PostgresResource_PostgresResourceOutput;

  constructor(data?: PartialMessage<PostgresResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PostgresResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: PostgresResource_PostgresResourceOutput },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresResource {
    return new PostgresResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresResource {
    return new PostgresResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresResource {
    return new PostgresResource().fromJsonString(jsonString, options);
  }

  static equals(a: PostgresResource | PlainMessage<PostgresResource> | undefined, b: PostgresResource | PlainMessage<PostgresResource> | undefined): boolean {
    return proto3.util.equals(PostgresResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput
 */
export class PostgresResource_PostgresResourceOutput extends Message<PostgresResource_PostgresResourceOutput> {
  /**
   * @generated from field: string read_dsn = 1;
   */
  readDsn = "";

  /**
   * @generated from field: string write_dsn = 2;
   */
  writeDsn = "";

  constructor(data?: PartialMessage<PostgresResource_PostgresResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "read_dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "write_dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: PostgresResource_PostgresResourceOutput | PlainMessage<PostgresResource_PostgresResourceOutput> | undefined, b: PostgresResource_PostgresResourceOutput | PlainMessage<PostgresResource_PostgresResourceOutput> | undefined): boolean {
    return proto3.util.equals(PostgresResource_PostgresResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.MysqlResource
 */
export class MysqlResource extends Message<MysqlResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput output = 1;
   */
  output?: MysqlResource_MysqlResourceOutput;

  constructor(data?: PartialMessage<MysqlResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.MysqlResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: MysqlResource_MysqlResourceOutput },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlResource {
    return new MysqlResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlResource {
    return new MysqlResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlResource {
    return new MysqlResource().fromJsonString(jsonString, options);
  }

  static equals(a: MysqlResource | PlainMessage<MysqlResource> | undefined, b: MysqlResource | PlainMessage<MysqlResource> | undefined): boolean {
    return proto3.util.equals(MysqlResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput
 */
export class MysqlResource_MysqlResourceOutput extends Message<MysqlResource_MysqlResourceOutput> {
  /**
   * @generated from field: string read_dsn = 1;
   */
  readDsn = "";

  /**
   * @generated from field: string write_dsn = 2;
   */
  writeDsn = "";

  constructor(data?: PartialMessage<MysqlResource_MysqlResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "read_dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "write_dsn", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: MysqlResource_MysqlResourceOutput | PlainMessage<MysqlResource_MysqlResourceOutput> | undefined, b: MysqlResource_MysqlResourceOutput | PlainMessage<MysqlResource_MysqlResourceOutput> | undefined): boolean {
    return proto3.util.equals(MysqlResource_MysqlResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource
 */
export class SqlMigrationResource extends Message<SqlMigrationResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource.SqlMigrationResourceOutput output = 1;
   */
  output?: SqlMigrationResource_SqlMigrationResourceOutput;

  /**
   * @generated from field: string digest = 2;
   */
  digest = "";

  constructor(data?: PartialMessage<SqlMigrationResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: SqlMigrationResource_SqlMigrationResourceOutput },
    { no: 2, name: "digest", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SqlMigrationResource {
    return new SqlMigrationResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SqlMigrationResource {
    return new SqlMigrationResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SqlMigrationResource {
    return new SqlMigrationResource().fromJsonString(jsonString, options);
  }

  static equals(a: SqlMigrationResource | PlainMessage<SqlMigrationResource> | undefined, b: SqlMigrationResource | PlainMessage<SqlMigrationResource> | undefined): boolean {
    return proto3.util.equals(SqlMigrationResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource.SqlMigrationResourceOutput
 */
export class SqlMigrationResource_SqlMigrationResourceOutput extends Message<SqlMigrationResource_SqlMigrationResourceOutput> {
  constructor(data?: PartialMessage<SqlMigrationResource_SqlMigrationResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.SqlMigrationResource.SqlMigrationResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SqlMigrationResource_SqlMigrationResourceOutput {
    return new SqlMigrationResource_SqlMigrationResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SqlMigrationResource_SqlMigrationResourceOutput {
    return new SqlMigrationResource_SqlMigrationResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SqlMigrationResource_SqlMigrationResourceOutput {
    return new SqlMigrationResource_SqlMigrationResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: SqlMigrationResource_SqlMigrationResourceOutput | PlainMessage<SqlMigrationResource_SqlMigrationResourceOutput> | undefined, b: SqlMigrationResource_SqlMigrationResourceOutput | PlainMessage<SqlMigrationResource_SqlMigrationResourceOutput> | undefined): boolean {
    return proto3.util.equals(SqlMigrationResource_SqlMigrationResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.ModuleResource
 */
export class ModuleResource extends Message<ModuleResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.ModuleResource.ModuleResourceOutput output = 1;
   */
  output?: ModuleResource_ModuleResourceOutput;

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Module schema = 2;
   */
  schema?: Module;

  /**
   * @generated from field: repeated xyz.block.ftl.v1.DeploymentArtefact artefacts = 3;
   */
  artefacts: DeploymentArtefact[] = [];

  /**
   * Runner labels required to run this deployment.
   *
   * @generated from field: optional google.protobuf.Struct labels = 4;
   */
  labels?: Struct;

  constructor(data?: PartialMessage<ModuleResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.ModuleResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: ModuleResource_ModuleResourceOutput },
    { no: 2, name: "schema", kind: "message", T: Module },
    { no: 3, name: "artefacts", kind: "message", T: DeploymentArtefact, repeated: true },
    { no: 4, name: "labels", kind: "message", T: Struct, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleResource {
    return new ModuleResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleResource {
    return new ModuleResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleResource {
    return new ModuleResource().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleResource | PlainMessage<ModuleResource> | undefined, b: ModuleResource | PlainMessage<ModuleResource> | undefined): boolean {
    return proto3.util.equals(ModuleResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.ModuleResource.ModuleResourceOutput
 */
export class ModuleResource_ModuleResourceOutput extends Message<ModuleResource_ModuleResourceOutput> {
  /**
   * @generated from field: string deployment_key = 1;
   */
  deploymentKey = "";

  constructor(data?: PartialMessage<ModuleResource_ModuleResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.ModuleResource.ModuleResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployment_key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleResource_ModuleResourceOutput {
    return new ModuleResource_ModuleResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleResource_ModuleResourceOutput {
    return new ModuleResource_ModuleResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleResource_ModuleResourceOutput {
    return new ModuleResource_ModuleResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleResource_ModuleResourceOutput | PlainMessage<ModuleResource_ModuleResourceOutput> | undefined, b: ModuleResource_ModuleResourceOutput | PlainMessage<ModuleResource_ModuleResourceOutput> | undefined): boolean {
    return proto3.util.equals(ModuleResource_ModuleResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.TopicResource
 */
export class TopicResource extends Message<TopicResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.TopicResource.TopicResourceOutput output = 1;
   */
  output?: TopicResource_TopicResourceOutput;

  constructor(data?: PartialMessage<TopicResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.TopicResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: TopicResource_TopicResourceOutput },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TopicResource {
    return new TopicResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TopicResource {
    return new TopicResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TopicResource {
    return new TopicResource().fromJsonString(jsonString, options);
  }

  static equals(a: TopicResource | PlainMessage<TopicResource> | undefined, b: TopicResource | PlainMessage<TopicResource> | undefined): boolean {
    return proto3.util.equals(TopicResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.TopicResource.TopicResourceOutput
 */
export class TopicResource_TopicResourceOutput extends Message<TopicResource_TopicResourceOutput> {
  /**
   * @generated from field: repeated string kafka_brokers = 1;
   */
  kafkaBrokers: string[] = [];

  /**
   * @generated from field: string topic_id = 2;
   */
  topicId = "";

  constructor(data?: PartialMessage<TopicResource_TopicResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.TopicResource.TopicResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "kafka_brokers", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 2, name: "topic_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TopicResource_TopicResourceOutput {
    return new TopicResource_TopicResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TopicResource_TopicResourceOutput {
    return new TopicResource_TopicResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TopicResource_TopicResourceOutput {
    return new TopicResource_TopicResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: TopicResource_TopicResourceOutput | PlainMessage<TopicResource_TopicResourceOutput> | undefined, b: TopicResource_TopicResourceOutput | PlainMessage<TopicResource_TopicResourceOutput> | undefined): boolean {
    return proto3.util.equals(TopicResource_TopicResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.SubscriptionResource
 */
export class SubscriptionResource extends Message<SubscriptionResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.SubscriptionResource.SubscriptionResourceOutput output = 1;
   */
  output?: SubscriptionResource_SubscriptionResourceOutput;

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Ref topic = 2;
   */
  topic?: Ref;

  constructor(data?: PartialMessage<SubscriptionResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.SubscriptionResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: SubscriptionResource_SubscriptionResourceOutput },
    { no: 2, name: "topic", kind: "message", T: Ref },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SubscriptionResource {
    return new SubscriptionResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SubscriptionResource {
    return new SubscriptionResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SubscriptionResource {
    return new SubscriptionResource().fromJsonString(jsonString, options);
  }

  static equals(a: SubscriptionResource | PlainMessage<SubscriptionResource> | undefined, b: SubscriptionResource | PlainMessage<SubscriptionResource> | undefined): boolean {
    return proto3.util.equals(SubscriptionResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.SubscriptionResource.SubscriptionResourceOutput
 */
export class SubscriptionResource_SubscriptionResourceOutput extends Message<SubscriptionResource_SubscriptionResourceOutput> {
  /**
   * @generated from field: repeated string kafka_brokers = 1;
   */
  kafkaBrokers: string[] = [];

  /**
   * @generated from field: string topic_id = 2;
   */
  topicId = "";

  /**
   * @generated from field: string consumer_group_id = 3;
   */
  consumerGroupId = "";

  constructor(data?: PartialMessage<SubscriptionResource_SubscriptionResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.SubscriptionResource.SubscriptionResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "kafka_brokers", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 2, name: "topic_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "consumer_group_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SubscriptionResource_SubscriptionResourceOutput {
    return new SubscriptionResource_SubscriptionResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SubscriptionResource_SubscriptionResourceOutput {
    return new SubscriptionResource_SubscriptionResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SubscriptionResource_SubscriptionResourceOutput {
    return new SubscriptionResource_SubscriptionResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: SubscriptionResource_SubscriptionResourceOutput | PlainMessage<SubscriptionResource_SubscriptionResourceOutput> | undefined, b: SubscriptionResource_SubscriptionResourceOutput | PlainMessage<SubscriptionResource_SubscriptionResourceOutput> | undefined): boolean {
    return proto3.util.equals(SubscriptionResource_SubscriptionResourceOutput, a, b);
  }
}

