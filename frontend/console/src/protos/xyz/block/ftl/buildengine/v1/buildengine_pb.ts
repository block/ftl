// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/buildengine/v1/buildengine.proto (package xyz.block.ftl.buildengine.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";
import { ErrorList, ModuleConfig } from "../../language/v1/language_pb.js";

/**
 * EngineStarted is published when the engine becomes busy building and deploying modules.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.EngineStarted
 */
export class EngineStarted extends Message<EngineStarted> {
  constructor(data?: PartialMessage<EngineStarted>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.EngineStarted";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EngineStarted {
    return new EngineStarted().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EngineStarted {
    return new EngineStarted().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EngineStarted {
    return new EngineStarted().fromJsonString(jsonString, options);
  }

  static equals(a: EngineStarted | PlainMessage<EngineStarted> | undefined, b: EngineStarted | PlainMessage<EngineStarted> | undefined): boolean {
    return proto3.util.equals(EngineStarted, a, b);
  }
}

/**
 * EngineEnded is published when the engine is no longer building or deploying any modules.
 * If there are any remaining errors, they will be included in the ModuleErrors map.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.EngineEnded
 */
export class EngineEnded extends Message<EngineEnded> {
  /**
   * module name -> error
   *
   * @generated from field: map<string, xyz.block.ftl.language.v1.ErrorList> module_errors = 1;
   */
  moduleErrors: { [key: string]: ErrorList } = {};

  constructor(data?: PartialMessage<EngineEnded>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.EngineEnded";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module_errors", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: ErrorList} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EngineEnded {
    return new EngineEnded().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EngineEnded {
    return new EngineEnded().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EngineEnded {
    return new EngineEnded().fromJsonString(jsonString, options);
  }

  static equals(a: EngineEnded | PlainMessage<EngineEnded> | undefined, b: EngineEnded | PlainMessage<EngineEnded> | undefined): boolean {
    return proto3.util.equals(EngineEnded, a, b);
  }
}

/**
 * ModuleAdded is published when the engine discovers a module.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleAdded
 */
export class ModuleAdded extends Message<ModuleAdded> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<ModuleAdded>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleAdded";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleAdded {
    return new ModuleAdded().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleAdded {
    return new ModuleAdded().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleAdded {
    return new ModuleAdded().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleAdded | PlainMessage<ModuleAdded> | undefined, b: ModuleAdded | PlainMessage<ModuleAdded> | undefined): boolean {
    return proto3.util.equals(ModuleAdded, a, b);
  }
}

/**
 * ModuleRemoved is published when the engine discovers a module has been removed.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleRemoved
 */
export class ModuleRemoved extends Message<ModuleRemoved> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<ModuleRemoved>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleRemoved";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleRemoved {
    return new ModuleRemoved().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleRemoved {
    return new ModuleRemoved().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleRemoved {
    return new ModuleRemoved().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleRemoved | PlainMessage<ModuleRemoved> | undefined, b: ModuleRemoved | PlainMessage<ModuleRemoved> | undefined): boolean {
    return proto3.util.equals(ModuleRemoved, a, b);
  }
}

/**
 * ModuleBuildWaiting is published when a build is waiting for dependencies to build
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleBuildWaiting
 */
export class ModuleBuildWaiting extends Message<ModuleBuildWaiting> {
  /**
   * @generated from field: xyz.block.ftl.language.v1.ModuleConfig config = 1;
   */
  config?: ModuleConfig;

  constructor(data?: PartialMessage<ModuleBuildWaiting>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleBuildWaiting";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: ModuleConfig },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleBuildWaiting {
    return new ModuleBuildWaiting().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleBuildWaiting {
    return new ModuleBuildWaiting().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleBuildWaiting {
    return new ModuleBuildWaiting().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleBuildWaiting | PlainMessage<ModuleBuildWaiting> | undefined, b: ModuleBuildWaiting | PlainMessage<ModuleBuildWaiting> | undefined): boolean {
    return proto3.util.equals(ModuleBuildWaiting, a, b);
  }
}

/**
 * ModuleBuildStarted is published when a build has started for a module.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleBuildStarted
 */
export class ModuleBuildStarted extends Message<ModuleBuildStarted> {
  /**
   * @generated from field: xyz.block.ftl.language.v1.ModuleConfig config = 1;
   */
  config?: ModuleConfig;

  /**
   * @generated from field: bool is_auto_rebuild = 2;
   */
  isAutoRebuild = false;

  constructor(data?: PartialMessage<ModuleBuildStarted>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleBuildStarted";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: ModuleConfig },
    { no: 2, name: "is_auto_rebuild", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleBuildStarted {
    return new ModuleBuildStarted().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleBuildStarted {
    return new ModuleBuildStarted().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleBuildStarted {
    return new ModuleBuildStarted().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleBuildStarted | PlainMessage<ModuleBuildStarted> | undefined, b: ModuleBuildStarted | PlainMessage<ModuleBuildStarted> | undefined): boolean {
    return proto3.util.equals(ModuleBuildStarted, a, b);
  }
}

/**
 * ModuleBuildFailed is published for any build failures.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleBuildFailed
 */
export class ModuleBuildFailed extends Message<ModuleBuildFailed> {
  /**
   * @generated from field: xyz.block.ftl.language.v1.ModuleConfig config = 1;
   */
  config?: ModuleConfig;

  /**
   * @generated from field: xyz.block.ftl.language.v1.ErrorList errors = 2;
   */
  errors?: ErrorList;

  /**
   * @generated from field: bool is_auto_rebuild = 3;
   */
  isAutoRebuild = false;

  constructor(data?: PartialMessage<ModuleBuildFailed>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleBuildFailed";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: ModuleConfig },
    { no: 2, name: "errors", kind: "message", T: ErrorList },
    { no: 3, name: "is_auto_rebuild", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleBuildFailed {
    return new ModuleBuildFailed().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleBuildFailed {
    return new ModuleBuildFailed().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleBuildFailed {
    return new ModuleBuildFailed().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleBuildFailed | PlainMessage<ModuleBuildFailed> | undefined, b: ModuleBuildFailed | PlainMessage<ModuleBuildFailed> | undefined): boolean {
    return proto3.util.equals(ModuleBuildFailed, a, b);
  }
}

/**
 * ModuleBuildSuccess is published when all modules have been built successfully built.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleBuildSuccess
 */
export class ModuleBuildSuccess extends Message<ModuleBuildSuccess> {
  /**
   * @generated from field: xyz.block.ftl.language.v1.ModuleConfig config = 1;
   */
  config?: ModuleConfig;

  /**
   * @generated from field: bool is_auto_rebuild = 2;
   */
  isAutoRebuild = false;

  constructor(data?: PartialMessage<ModuleBuildSuccess>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleBuildSuccess";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: ModuleConfig },
    { no: 2, name: "is_auto_rebuild", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleBuildSuccess {
    return new ModuleBuildSuccess().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleBuildSuccess {
    return new ModuleBuildSuccess().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleBuildSuccess {
    return new ModuleBuildSuccess().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleBuildSuccess | PlainMessage<ModuleBuildSuccess> | undefined, b: ModuleBuildSuccess | PlainMessage<ModuleBuildSuccess> | undefined): boolean {
    return proto3.util.equals(ModuleBuildSuccess, a, b);
  }
}

/**
 * ModuleDeployStarted is published when a deploy has begun for a module.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleDeployStarted
 */
export class ModuleDeployStarted extends Message<ModuleDeployStarted> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<ModuleDeployStarted>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleDeployStarted";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleDeployStarted {
    return new ModuleDeployStarted().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleDeployStarted {
    return new ModuleDeployStarted().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleDeployStarted {
    return new ModuleDeployStarted().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleDeployStarted | PlainMessage<ModuleDeployStarted> | undefined, b: ModuleDeployStarted | PlainMessage<ModuleDeployStarted> | undefined): boolean {
    return proto3.util.equals(ModuleDeployStarted, a, b);
  }
}

/**
 * ModuleDeployFailed is published for any deploy failures.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleDeployFailed
 */
export class ModuleDeployFailed extends Message<ModuleDeployFailed> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  /**
   * @generated from field: xyz.block.ftl.language.v1.ErrorList errors = 2;
   */
  errors?: ErrorList;

  constructor(data?: PartialMessage<ModuleDeployFailed>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleDeployFailed";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "errors", kind: "message", T: ErrorList },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleDeployFailed {
    return new ModuleDeployFailed().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleDeployFailed {
    return new ModuleDeployFailed().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleDeployFailed {
    return new ModuleDeployFailed().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleDeployFailed | PlainMessage<ModuleDeployFailed> | undefined, b: ModuleDeployFailed | PlainMessage<ModuleDeployFailed> | undefined): boolean {
    return proto3.util.equals(ModuleDeployFailed, a, b);
  }
}

/**
 * ModuleDeploySuccess is published when all modules have been built successfully deployed.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.ModuleDeploySuccess
 */
export class ModuleDeploySuccess extends Message<ModuleDeploySuccess> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  constructor(data?: PartialMessage<ModuleDeploySuccess>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.ModuleDeploySuccess";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleDeploySuccess {
    return new ModuleDeploySuccess().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleDeploySuccess {
    return new ModuleDeploySuccess().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleDeploySuccess {
    return new ModuleDeploySuccess().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleDeploySuccess | PlainMessage<ModuleDeploySuccess> | undefined, b: ModuleDeploySuccess | PlainMessage<ModuleDeploySuccess> | undefined): boolean {
    return proto3.util.equals(ModuleDeploySuccess, a, b);
  }
}

/**
 * EngineEvent is an event published by the engine as modules get built and deployed.
 *
 * @generated from message xyz.block.ftl.buildengine.v1.EngineEvent
 */
export class EngineEvent extends Message<EngineEvent> {
  /**
   * @generated from field: google.protobuf.Timestamp timestamp = 1;
   */
  timestamp?: Timestamp;

  /**
   * @generated from oneof xyz.block.ftl.buildengine.v1.EngineEvent.event
   */
  event: {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.EngineStarted engine_started = 2;
     */
    value: EngineStarted;
    case: "engineStarted";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.EngineEnded engine_ended = 3;
     */
    value: EngineEnded;
    case: "engineEnded";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleAdded module_added = 4;
     */
    value: ModuleAdded;
    case: "moduleAdded";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleRemoved module_removed = 5;
     */
    value: ModuleRemoved;
    case: "moduleRemoved";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleBuildWaiting module_build_waiting = 6;
     */
    value: ModuleBuildWaiting;
    case: "moduleBuildWaiting";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleBuildStarted module_build_started = 7;
     */
    value: ModuleBuildStarted;
    case: "moduleBuildStarted";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleBuildFailed module_build_failed = 8;
     */
    value: ModuleBuildFailed;
    case: "moduleBuildFailed";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleBuildSuccess module_build_success = 9;
     */
    value: ModuleBuildSuccess;
    case: "moduleBuildSuccess";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleDeployStarted module_deploy_started = 10;
     */
    value: ModuleDeployStarted;
    case: "moduleDeployStarted";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleDeployFailed module_deploy_failed = 11;
     */
    value: ModuleDeployFailed;
    case: "moduleDeployFailed";
  } | {
    /**
     * @generated from field: xyz.block.ftl.buildengine.v1.ModuleDeploySuccess module_deploy_success = 12;
     */
    value: ModuleDeploySuccess;
    case: "moduleDeploySuccess";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<EngineEvent>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.EngineEvent";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "timestamp", kind: "message", T: Timestamp },
    { no: 2, name: "engine_started", kind: "message", T: EngineStarted, oneof: "event" },
    { no: 3, name: "engine_ended", kind: "message", T: EngineEnded, oneof: "event" },
    { no: 4, name: "module_added", kind: "message", T: ModuleAdded, oneof: "event" },
    { no: 5, name: "module_removed", kind: "message", T: ModuleRemoved, oneof: "event" },
    { no: 6, name: "module_build_waiting", kind: "message", T: ModuleBuildWaiting, oneof: "event" },
    { no: 7, name: "module_build_started", kind: "message", T: ModuleBuildStarted, oneof: "event" },
    { no: 8, name: "module_build_failed", kind: "message", T: ModuleBuildFailed, oneof: "event" },
    { no: 9, name: "module_build_success", kind: "message", T: ModuleBuildSuccess, oneof: "event" },
    { no: 10, name: "module_deploy_started", kind: "message", T: ModuleDeployStarted, oneof: "event" },
    { no: 11, name: "module_deploy_failed", kind: "message", T: ModuleDeployFailed, oneof: "event" },
    { no: 12, name: "module_deploy_success", kind: "message", T: ModuleDeploySuccess, oneof: "event" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EngineEvent {
    return new EngineEvent().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EngineEvent {
    return new EngineEvent().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EngineEvent {
    return new EngineEvent().fromJsonString(jsonString, options);
  }

  static equals(a: EngineEvent | PlainMessage<EngineEvent> | undefined, b: EngineEvent | PlainMessage<EngineEvent> | undefined): boolean {
    return proto3.util.equals(EngineEvent, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.buildengine.v1.StreamEngineEventsRequest
 */
export class StreamEngineEventsRequest extends Message<StreamEngineEventsRequest> {
  /**
   * If true, cached events will be replayed before streaming new events.
   * If false, only new events will be streamed.
   *
   * @generated from field: bool replay_history = 1;
   */
  replayHistory = false;

  constructor(data?: PartialMessage<StreamEngineEventsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.StreamEngineEventsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "replay_history", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamEngineEventsRequest {
    return new StreamEngineEventsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamEngineEventsRequest {
    return new StreamEngineEventsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamEngineEventsRequest {
    return new StreamEngineEventsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: StreamEngineEventsRequest | PlainMessage<StreamEngineEventsRequest> | undefined, b: StreamEngineEventsRequest | PlainMessage<StreamEngineEventsRequest> | undefined): boolean {
    return proto3.util.equals(StreamEngineEventsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.buildengine.v1.StreamEngineEventsResponse
 */
export class StreamEngineEventsResponse extends Message<StreamEngineEventsResponse> {
  /**
   * @generated from field: xyz.block.ftl.buildengine.v1.EngineEvent event = 1;
   */
  event?: EngineEvent;

  constructor(data?: PartialMessage<StreamEngineEventsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.buildengine.v1.StreamEngineEventsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event", kind: "message", T: EngineEvent },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamEngineEventsResponse {
    return new StreamEngineEventsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamEngineEventsResponse {
    return new StreamEngineEventsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamEngineEventsResponse {
    return new StreamEngineEventsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: StreamEngineEventsResponse | PlainMessage<StreamEngineEventsResponse> | undefined, b: StreamEngineEventsResponse | PlainMessage<StreamEngineEventsResponse> | undefined): boolean {
    return proto3.util.equals(StreamEngineEventsResponse, a, b);
  }
}

