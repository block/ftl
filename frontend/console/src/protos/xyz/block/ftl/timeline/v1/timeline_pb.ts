// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/timeline/v1/timeline.proto (package xyz.block.ftl.timeline.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Duration, Message, proto3, protoInt64, Timestamp } from "@bufbuild/protobuf";
import { CallEvent, ChangesetCreatedEvent, ChangesetStateChangedEvent, CronScheduledEvent, DeploymentRuntimeEvent, Event, EventType, IngressEvent, LogEvent, LogLevel, PubSubConsumeEvent, PubSubPublishEvent } from "./event_pb.js";

/**
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery
 */
export class TimelineQuery extends Message<TimelineQuery> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.TimelineQuery.Filter filters = 1;
   */
  filters: TimelineQuery_Filter[] = [];

  /**
   * @generated from field: int32 limit = 2;
   */
  limit = 0;

  /**
   * Ordering is done by id which matches publication order.
   * This roughly corresponds to the time of the event, but not strictly.
   *
   * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.Order order = 3;
   */
  order = TimelineQuery_Order.UNSPECIFIED;

  constructor(data?: PartialMessage<TimelineQuery>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "filters", kind: "message", T: TimelineQuery_Filter, repeated: true },
    { no: 2, name: "limit", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "order", kind: "enum", T: proto3.getEnumType(TimelineQuery_Order) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery {
    return new TimelineQuery().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery {
    return new TimelineQuery().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery {
    return new TimelineQuery().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery | PlainMessage<TimelineQuery> | undefined, b: TimelineQuery | PlainMessage<TimelineQuery> | undefined): boolean {
    return proto3.util.equals(TimelineQuery, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.timeline.v1.TimelineQuery.Order
 */
export enum TimelineQuery_Order {
  /**
   * @generated from enum value: ORDER_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: ORDER_ASC = 1;
   */
  ASC = 1,

  /**
   * @generated from enum value: ORDER_DESC = 2;
   */
  DESC = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(TimelineQuery_Order)
proto3.util.setEnumType(TimelineQuery_Order, "xyz.block.ftl.timeline.v1.TimelineQuery.Order", [
  { no: 0, name: "ORDER_UNSPECIFIED" },
  { no: 1, name: "ORDER_ASC" },
  { no: 2, name: "ORDER_DESC" },
]);

/**
 * Filters events by log level.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.LogLevelFilter
 */
export class TimelineQuery_LogLevelFilter extends Message<TimelineQuery_LogLevelFilter> {
  /**
   * @generated from field: xyz.block.ftl.timeline.v1.LogLevel log_level = 1;
   */
  logLevel = LogLevel.UNSPECIFIED;

  constructor(data?: PartialMessage<TimelineQuery_LogLevelFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.LogLevelFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "log_level", kind: "enum", T: proto3.getEnumType(LogLevel) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_LogLevelFilter {
    return new TimelineQuery_LogLevelFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_LogLevelFilter {
    return new TimelineQuery_LogLevelFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_LogLevelFilter {
    return new TimelineQuery_LogLevelFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_LogLevelFilter | PlainMessage<TimelineQuery_LogLevelFilter> | undefined, b: TimelineQuery_LogLevelFilter | PlainMessage<TimelineQuery_LogLevelFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_LogLevelFilter, a, b);
  }
}

/**
 * Filters events by deployment key.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.DeploymentFilter
 */
export class TimelineQuery_DeploymentFilter extends Message<TimelineQuery_DeploymentFilter> {
  /**
   * @generated from field: repeated string deployments = 1;
   */
  deployments: string[] = [];

  constructor(data?: PartialMessage<TimelineQuery_DeploymentFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.DeploymentFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_DeploymentFilter {
    return new TimelineQuery_DeploymentFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_DeploymentFilter {
    return new TimelineQuery_DeploymentFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_DeploymentFilter {
    return new TimelineQuery_DeploymentFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_DeploymentFilter | PlainMessage<TimelineQuery_DeploymentFilter> | undefined, b: TimelineQuery_DeploymentFilter | PlainMessage<TimelineQuery_DeploymentFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_DeploymentFilter, a, b);
  }
}

/**
 * Filters events by changeset key.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.ChangesetFilter
 */
export class TimelineQuery_ChangesetFilter extends Message<TimelineQuery_ChangesetFilter> {
  /**
   * @generated from field: repeated string changesets = 1;
   */
  changesets: string[] = [];

  constructor(data?: PartialMessage<TimelineQuery_ChangesetFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.ChangesetFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "changesets", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_ChangesetFilter {
    return new TimelineQuery_ChangesetFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_ChangesetFilter {
    return new TimelineQuery_ChangesetFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_ChangesetFilter {
    return new TimelineQuery_ChangesetFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_ChangesetFilter | PlainMessage<TimelineQuery_ChangesetFilter> | undefined, b: TimelineQuery_ChangesetFilter | PlainMessage<TimelineQuery_ChangesetFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_ChangesetFilter, a, b);
  }
}

/**
 * Filters events by request key.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.RequestFilter
 */
export class TimelineQuery_RequestFilter extends Message<TimelineQuery_RequestFilter> {
  /**
   * @generated from field: repeated string requests = 1;
   */
  requests: string[] = [];

  constructor(data?: PartialMessage<TimelineQuery_RequestFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.RequestFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "requests", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_RequestFilter {
    return new TimelineQuery_RequestFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_RequestFilter {
    return new TimelineQuery_RequestFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_RequestFilter {
    return new TimelineQuery_RequestFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_RequestFilter | PlainMessage<TimelineQuery_RequestFilter> | undefined, b: TimelineQuery_RequestFilter | PlainMessage<TimelineQuery_RequestFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_RequestFilter, a, b);
  }
}

/**
 * Filters events by event type.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.EventTypeFilter
 */
export class TimelineQuery_EventTypeFilter extends Message<TimelineQuery_EventTypeFilter> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.EventType event_types = 1;
   */
  eventTypes: EventType[] = [];

  constructor(data?: PartialMessage<TimelineQuery_EventTypeFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.EventTypeFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_types", kind: "enum", T: proto3.getEnumType(EventType), repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_EventTypeFilter {
    return new TimelineQuery_EventTypeFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_EventTypeFilter {
    return new TimelineQuery_EventTypeFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_EventTypeFilter {
    return new TimelineQuery_EventTypeFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_EventTypeFilter | PlainMessage<TimelineQuery_EventTypeFilter> | undefined, b: TimelineQuery_EventTypeFilter | PlainMessage<TimelineQuery_EventTypeFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_EventTypeFilter, a, b);
  }
}

/**
 * Filters events by time.
 *
 * Either end of the time range can be omitted to indicate no bound.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.TimeFilter
 */
export class TimelineQuery_TimeFilter extends Message<TimelineQuery_TimeFilter> {
  /**
   * @generated from field: optional google.protobuf.Timestamp older_than = 1;
   */
  olderThan?: Timestamp;

  /**
   * @generated from field: optional google.protobuf.Timestamp newer_than = 2;
   */
  newerThan?: Timestamp;

  constructor(data?: PartialMessage<TimelineQuery_TimeFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.TimeFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "older_than", kind: "message", T: Timestamp, opt: true },
    { no: 2, name: "newer_than", kind: "message", T: Timestamp, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_TimeFilter {
    return new TimelineQuery_TimeFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_TimeFilter {
    return new TimelineQuery_TimeFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_TimeFilter {
    return new TimelineQuery_TimeFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_TimeFilter | PlainMessage<TimelineQuery_TimeFilter> | undefined, b: TimelineQuery_TimeFilter | PlainMessage<TimelineQuery_TimeFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_TimeFilter, a, b);
  }
}

/**
 * Filters events by ID.
 *
 * Either end of the ID range can be omitted to indicate no bound.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.IDFilter
 */
export class TimelineQuery_IDFilter extends Message<TimelineQuery_IDFilter> {
  /**
   * @generated from field: optional int64 lower_than = 1;
   */
  lowerThan?: bigint;

  /**
   * @generated from field: optional int64 higher_than = 2;
   */
  higherThan?: bigint;

  constructor(data?: PartialMessage<TimelineQuery_IDFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.IDFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "lower_than", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
    { no: 2, name: "higher_than", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_IDFilter {
    return new TimelineQuery_IDFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_IDFilter {
    return new TimelineQuery_IDFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_IDFilter {
    return new TimelineQuery_IDFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_IDFilter | PlainMessage<TimelineQuery_IDFilter> | undefined, b: TimelineQuery_IDFilter | PlainMessage<TimelineQuery_IDFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_IDFilter, a, b);
  }
}

/**
 * Filters events by call.
 *
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.CallFilter
 */
export class TimelineQuery_CallFilter extends Message<TimelineQuery_CallFilter> {
  /**
   * @generated from field: string dest_module = 1;
   */
  destModule = "";

  /**
   * @generated from field: optional string dest_verb = 2;
   */
  destVerb?: string;

  /**
   * @generated from field: optional string source_module = 3;
   */
  sourceModule?: string;

  constructor(data?: PartialMessage<TimelineQuery_CallFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.CallFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "dest_module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "dest_verb", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "source_module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_CallFilter {
    return new TimelineQuery_CallFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_CallFilter {
    return new TimelineQuery_CallFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_CallFilter {
    return new TimelineQuery_CallFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_CallFilter | PlainMessage<TimelineQuery_CallFilter> | undefined, b: TimelineQuery_CallFilter | PlainMessage<TimelineQuery_CallFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_CallFilter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.ModuleFilter
 */
export class TimelineQuery_ModuleFilter extends Message<TimelineQuery_ModuleFilter> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  /**
   * @generated from field: optional string verb = 2;
   */
  verb?: string;

  constructor(data?: PartialMessage<TimelineQuery_ModuleFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.ModuleFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "verb", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_ModuleFilter {
    return new TimelineQuery_ModuleFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_ModuleFilter {
    return new TimelineQuery_ModuleFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_ModuleFilter {
    return new TimelineQuery_ModuleFilter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_ModuleFilter | PlainMessage<TimelineQuery_ModuleFilter> | undefined, b: TimelineQuery_ModuleFilter | PlainMessage<TimelineQuery_ModuleFilter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_ModuleFilter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.TimelineQuery.Filter
 */
export class TimelineQuery_Filter extends Message<TimelineQuery_Filter> {
  /**
   * These map 1:1 with filters in backend/timeline/filters.go
   *
   * @generated from oneof xyz.block.ftl.timeline.v1.TimelineQuery.Filter.filter
   */
  filter: {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.LogLevelFilter log_level = 1;
     */
    value: TimelineQuery_LogLevelFilter;
    case: "logLevel";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.DeploymentFilter deployments = 2;
     */
    value: TimelineQuery_DeploymentFilter;
    case: "deployments";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.RequestFilter requests = 3;
     */
    value: TimelineQuery_RequestFilter;
    case: "requests";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.EventTypeFilter event_types = 4;
     */
    value: TimelineQuery_EventTypeFilter;
    case: "eventTypes";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.TimeFilter time = 5;
     */
    value: TimelineQuery_TimeFilter;
    case: "time";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.IDFilter id = 6;
     */
    value: TimelineQuery_IDFilter;
    case: "id";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.CallFilter call = 7;
     */
    value: TimelineQuery_CallFilter;
    case: "call";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.ModuleFilter module = 8;
     */
    value: TimelineQuery_ModuleFilter;
    case: "module";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery.ChangesetFilter changesets = 9;
     */
    value: TimelineQuery_ChangesetFilter;
    case: "changesets";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<TimelineQuery_Filter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.TimelineQuery.Filter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "log_level", kind: "message", T: TimelineQuery_LogLevelFilter, oneof: "filter" },
    { no: 2, name: "deployments", kind: "message", T: TimelineQuery_DeploymentFilter, oneof: "filter" },
    { no: 3, name: "requests", kind: "message", T: TimelineQuery_RequestFilter, oneof: "filter" },
    { no: 4, name: "event_types", kind: "message", T: TimelineQuery_EventTypeFilter, oneof: "filter" },
    { no: 5, name: "time", kind: "message", T: TimelineQuery_TimeFilter, oneof: "filter" },
    { no: 6, name: "id", kind: "message", T: TimelineQuery_IDFilter, oneof: "filter" },
    { no: 7, name: "call", kind: "message", T: TimelineQuery_CallFilter, oneof: "filter" },
    { no: 8, name: "module", kind: "message", T: TimelineQuery_ModuleFilter, oneof: "filter" },
    { no: 9, name: "changesets", kind: "message", T: TimelineQuery_ChangesetFilter, oneof: "filter" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TimelineQuery_Filter {
    return new TimelineQuery_Filter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TimelineQuery_Filter {
    return new TimelineQuery_Filter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TimelineQuery_Filter {
    return new TimelineQuery_Filter().fromJsonString(jsonString, options);
  }

  static equals(a: TimelineQuery_Filter | PlainMessage<TimelineQuery_Filter> | undefined, b: TimelineQuery_Filter | PlainMessage<TimelineQuery_Filter> | undefined): boolean {
    return proto3.util.equals(TimelineQuery_Filter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest
 */
export class GetTimelineRequest extends Message<GetTimelineRequest> {
  /**
   * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery query = 1;
   */
  query?: TimelineQuery;

  constructor(data?: PartialMessage<GetTimelineRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "query", kind: "message", T: TimelineQuery },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest | PlainMessage<GetTimelineRequest> | undefined, b: GetTimelineRequest | PlainMessage<GetTimelineRequest> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineResponse
 */
export class GetTimelineResponse extends Message<GetTimelineResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.Event events = 1;
   */
  events: Event[] = [];

  /**
   * For pagination, this cursor is where we should start our next query
   *
   * @generated from field: optional int64 cursor = 2;
   */
  cursor?: bigint;

  constructor(data?: PartialMessage<GetTimelineResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: Event, repeated: true },
    { no: 2, name: "cursor", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineResponse | PlainMessage<GetTimelineResponse> | undefined, b: GetTimelineResponse | PlainMessage<GetTimelineResponse> | undefined): boolean {
    return proto3.util.equals(GetTimelineResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.StreamTimelineRequest
 */
export class StreamTimelineRequest extends Message<StreamTimelineRequest> {
  /**
   * @generated from field: optional google.protobuf.Duration update_interval = 1;
   */
  updateInterval?: Duration;

  /**
   * @generated from field: xyz.block.ftl.timeline.v1.TimelineQuery query = 2;
   */
  query?: TimelineQuery;

  constructor(data?: PartialMessage<StreamTimelineRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.StreamTimelineRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "update_interval", kind: "message", T: Duration, opt: true },
    { no: 2, name: "query", kind: "message", T: TimelineQuery },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamTimelineRequest {
    return new StreamTimelineRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamTimelineRequest {
    return new StreamTimelineRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamTimelineRequest {
    return new StreamTimelineRequest().fromJsonString(jsonString, options);
  }

  static equals(a: StreamTimelineRequest | PlainMessage<StreamTimelineRequest> | undefined, b: StreamTimelineRequest | PlainMessage<StreamTimelineRequest> | undefined): boolean {
    return proto3.util.equals(StreamTimelineRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.StreamTimelineResponse
 */
export class StreamTimelineResponse extends Message<StreamTimelineResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.Event events = 1;
   */
  events: Event[] = [];

  constructor(data?: PartialMessage<StreamTimelineResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.StreamTimelineResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: Event, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): StreamTimelineResponse {
    return new StreamTimelineResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): StreamTimelineResponse {
    return new StreamTimelineResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): StreamTimelineResponse {
    return new StreamTimelineResponse().fromJsonString(jsonString, options);
  }

  static equals(a: StreamTimelineResponse | PlainMessage<StreamTimelineResponse> | undefined, b: StreamTimelineResponse | PlainMessage<StreamTimelineResponse> | undefined): boolean {
    return proto3.util.equals(StreamTimelineResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.CreateEventsRequest
 */
export class CreateEventsRequest extends Message<CreateEventsRequest> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.CreateEventsRequest.EventEntry entries = 1;
   */
  entries: CreateEventsRequest_EventEntry[] = [];

  constructor(data?: PartialMessage<CreateEventsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.CreateEventsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "entries", kind: "message", T: CreateEventsRequest_EventEntry, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEventsRequest {
    return new CreateEventsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEventsRequest {
    return new CreateEventsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEventsRequest {
    return new CreateEventsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateEventsRequest | PlainMessage<CreateEventsRequest> | undefined, b: CreateEventsRequest | PlainMessage<CreateEventsRequest> | undefined): boolean {
    return proto3.util.equals(CreateEventsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.CreateEventsRequest.EventEntry
 */
export class CreateEventsRequest_EventEntry extends Message<CreateEventsRequest_EventEntry> {
  /**
   * @generated from field: google.protobuf.Timestamp timestamp = 1;
   */
  timestamp?: Timestamp;

  /**
   * @generated from oneof xyz.block.ftl.timeline.v1.CreateEventsRequest.EventEntry.entry
   */
  entry: {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.LogEvent log = 2;
     */
    value: LogEvent;
    case: "log";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.CallEvent call = 3;
     */
    value: CallEvent;
    case: "call";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.IngressEvent ingress = 4;
     */
    value: IngressEvent;
    case: "ingress";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.CronScheduledEvent cron_scheduled = 5;
     */
    value: CronScheduledEvent;
    case: "cronScheduled";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.PubSubPublishEvent pubsub_publish = 6;
     */
    value: PubSubPublishEvent;
    case: "pubsubPublish";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.PubSubConsumeEvent pubsub_consume = 7;
     */
    value: PubSubConsumeEvent;
    case: "pubsubConsume";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.ChangesetCreatedEvent changeset_created = 8;
     */
    value: ChangesetCreatedEvent;
    case: "changesetCreated";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.ChangesetStateChangedEvent changeset_state_changed = 9;
     */
    value: ChangesetStateChangedEvent;
    case: "changesetStateChanged";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.DeploymentRuntimeEvent deployment_runtime = 10;
     */
    value: DeploymentRuntimeEvent;
    case: "deploymentRuntime";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<CreateEventsRequest_EventEntry>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.CreateEventsRequest.EventEntry";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "timestamp", kind: "message", T: Timestamp },
    { no: 2, name: "log", kind: "message", T: LogEvent, oneof: "entry" },
    { no: 3, name: "call", kind: "message", T: CallEvent, oneof: "entry" },
    { no: 4, name: "ingress", kind: "message", T: IngressEvent, oneof: "entry" },
    { no: 5, name: "cron_scheduled", kind: "message", T: CronScheduledEvent, oneof: "entry" },
    { no: 6, name: "pubsub_publish", kind: "message", T: PubSubPublishEvent, oneof: "entry" },
    { no: 7, name: "pubsub_consume", kind: "message", T: PubSubConsumeEvent, oneof: "entry" },
    { no: 8, name: "changeset_created", kind: "message", T: ChangesetCreatedEvent, oneof: "entry" },
    { no: 9, name: "changeset_state_changed", kind: "message", T: ChangesetStateChangedEvent, oneof: "entry" },
    { no: 10, name: "deployment_runtime", kind: "message", T: DeploymentRuntimeEvent, oneof: "entry" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEventsRequest_EventEntry {
    return new CreateEventsRequest_EventEntry().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEventsRequest_EventEntry {
    return new CreateEventsRequest_EventEntry().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEventsRequest_EventEntry {
    return new CreateEventsRequest_EventEntry().fromJsonString(jsonString, options);
  }

  static equals(a: CreateEventsRequest_EventEntry | PlainMessage<CreateEventsRequest_EventEntry> | undefined, b: CreateEventsRequest_EventEntry | PlainMessage<CreateEventsRequest_EventEntry> | undefined): boolean {
    return proto3.util.equals(CreateEventsRequest_EventEntry, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.CreateEventsResponse
 */
export class CreateEventsResponse extends Message<CreateEventsResponse> {
  constructor(data?: PartialMessage<CreateEventsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.CreateEventsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEventsResponse {
    return new CreateEventsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEventsResponse {
    return new CreateEventsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEventsResponse {
    return new CreateEventsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CreateEventsResponse | PlainMessage<CreateEventsResponse> | undefined, b: CreateEventsResponse | PlainMessage<CreateEventsResponse> | undefined): boolean {
    return proto3.util.equals(CreateEventsResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.DeleteOldEventsRequest
 */
export class DeleteOldEventsRequest extends Message<DeleteOldEventsRequest> {
  /**
   * @generated from field: xyz.block.ftl.timeline.v1.EventType event_type = 1;
   */
  eventType = EventType.UNSPECIFIED;

  /**
   * @generated from field: int64 age_seconds = 2;
   */
  ageSeconds = protoInt64.zero;

  constructor(data?: PartialMessage<DeleteOldEventsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.DeleteOldEventsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_type", kind: "enum", T: proto3.getEnumType(EventType) },
    { no: 2, name: "age_seconds", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteOldEventsRequest | PlainMessage<DeleteOldEventsRequest> | undefined, b: DeleteOldEventsRequest | PlainMessage<DeleteOldEventsRequest> | undefined): boolean {
    return proto3.util.equals(DeleteOldEventsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.DeleteOldEventsResponse
 */
export class DeleteOldEventsResponse extends Message<DeleteOldEventsResponse> {
  /**
   * @generated from field: int64 deleted_count = 1;
   */
  deletedCount = protoInt64.zero;

  constructor(data?: PartialMessage<DeleteOldEventsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.DeleteOldEventsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deleted_count", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteOldEventsResponse | PlainMessage<DeleteOldEventsResponse> | undefined, b: DeleteOldEventsResponse | PlainMessage<DeleteOldEventsResponse> | undefined): boolean {
    return proto3.util.equals(DeleteOldEventsResponse, a, b);
  }
}

