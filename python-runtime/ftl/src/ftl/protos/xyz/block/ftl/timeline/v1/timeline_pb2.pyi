from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.timeline.v1 import event_pb2 as _event_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TimelineQuery(_message.Message):
    __slots__ = ("filters", "limit", "order")
    class Order(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ORDER_UNSPECIFIED: _ClassVar[TimelineQuery.Order]
        ORDER_ASC: _ClassVar[TimelineQuery.Order]
        ORDER_DESC: _ClassVar[TimelineQuery.Order]
    ORDER_UNSPECIFIED: TimelineQuery.Order
    ORDER_ASC: TimelineQuery.Order
    ORDER_DESC: TimelineQuery.Order
    class LogLevelFilter(_message.Message):
        __slots__ = ("log_level",)
        LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
        log_level: _event_pb2.LogLevel
        def __init__(self, log_level: _Optional[_Union[_event_pb2.LogLevel, str]] = ...) -> None: ...
    class DeploymentFilter(_message.Message):
        __slots__ = ("deployments",)
        DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
        deployments: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, deployments: _Optional[_Iterable[str]] = ...) -> None: ...
    class ChangesetFilter(_message.Message):
        __slots__ = ("changesets",)
        CHANGESETS_FIELD_NUMBER: _ClassVar[int]
        changesets: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, changesets: _Optional[_Iterable[str]] = ...) -> None: ...
    class RequestFilter(_message.Message):
        __slots__ = ("requests",)
        REQUESTS_FIELD_NUMBER: _ClassVar[int]
        requests: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, requests: _Optional[_Iterable[str]] = ...) -> None: ...
    class EventTypeFilter(_message.Message):
        __slots__ = ("event_types",)
        EVENT_TYPES_FIELD_NUMBER: _ClassVar[int]
        event_types: _containers.RepeatedScalarFieldContainer[_event_pb2.EventType]
        def __init__(self, event_types: _Optional[_Iterable[_Union[_event_pb2.EventType, str]]] = ...) -> None: ...
    class TimeFilter(_message.Message):
        __slots__ = ("older_than", "newer_than")
        OLDER_THAN_FIELD_NUMBER: _ClassVar[int]
        NEWER_THAN_FIELD_NUMBER: _ClassVar[int]
        older_than: _timestamp_pb2.Timestamp
        newer_than: _timestamp_pb2.Timestamp
        def __init__(self, older_than: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., newer_than: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    class IDFilter(_message.Message):
        __slots__ = ("lower_than", "higher_than")
        LOWER_THAN_FIELD_NUMBER: _ClassVar[int]
        HIGHER_THAN_FIELD_NUMBER: _ClassVar[int]
        lower_than: int
        higher_than: int
        def __init__(self, lower_than: _Optional[int] = ..., higher_than: _Optional[int] = ...) -> None: ...
    class CallFilter(_message.Message):
        __slots__ = ("dest_module", "dest_verb", "source_module")
        DEST_MODULE_FIELD_NUMBER: _ClassVar[int]
        DEST_VERB_FIELD_NUMBER: _ClassVar[int]
        SOURCE_MODULE_FIELD_NUMBER: _ClassVar[int]
        dest_module: str
        dest_verb: str
        source_module: str
        def __init__(self, dest_module: _Optional[str] = ..., dest_verb: _Optional[str] = ..., source_module: _Optional[str] = ...) -> None: ...
    class ModuleFilter(_message.Message):
        __slots__ = ("module", "verb")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        VERB_FIELD_NUMBER: _ClassVar[int]
        module: str
        verb: str
        def __init__(self, module: _Optional[str] = ..., verb: _Optional[str] = ...) -> None: ...
    class Filter(_message.Message):
        __slots__ = ("log_level", "deployments", "requests", "event_types", "time", "id", "call", "module", "changesets")
        LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
        REQUESTS_FIELD_NUMBER: _ClassVar[int]
        EVENT_TYPES_FIELD_NUMBER: _ClassVar[int]
        TIME_FIELD_NUMBER: _ClassVar[int]
        ID_FIELD_NUMBER: _ClassVar[int]
        CALL_FIELD_NUMBER: _ClassVar[int]
        MODULE_FIELD_NUMBER: _ClassVar[int]
        CHANGESETS_FIELD_NUMBER: _ClassVar[int]
        log_level: TimelineQuery.LogLevelFilter
        deployments: TimelineQuery.DeploymentFilter
        requests: TimelineQuery.RequestFilter
        event_types: TimelineQuery.EventTypeFilter
        time: TimelineQuery.TimeFilter
        id: TimelineQuery.IDFilter
        call: TimelineQuery.CallFilter
        module: TimelineQuery.ModuleFilter
        changesets: TimelineQuery.ChangesetFilter
        def __init__(self, log_level: _Optional[_Union[TimelineQuery.LogLevelFilter, _Mapping]] = ..., deployments: _Optional[_Union[TimelineQuery.DeploymentFilter, _Mapping]] = ..., requests: _Optional[_Union[TimelineQuery.RequestFilter, _Mapping]] = ..., event_types: _Optional[_Union[TimelineQuery.EventTypeFilter, _Mapping]] = ..., time: _Optional[_Union[TimelineQuery.TimeFilter, _Mapping]] = ..., id: _Optional[_Union[TimelineQuery.IDFilter, _Mapping]] = ..., call: _Optional[_Union[TimelineQuery.CallFilter, _Mapping]] = ..., module: _Optional[_Union[TimelineQuery.ModuleFilter, _Mapping]] = ..., changesets: _Optional[_Union[TimelineQuery.ChangesetFilter, _Mapping]] = ...) -> None: ...
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    filters: _containers.RepeatedCompositeFieldContainer[TimelineQuery.Filter]
    limit: int
    order: TimelineQuery.Order
    def __init__(self, filters: _Optional[_Iterable[_Union[TimelineQuery.Filter, _Mapping]]] = ..., limit: _Optional[int] = ..., order: _Optional[_Union[TimelineQuery.Order, str]] = ...) -> None: ...

class GetTimelineRequest(_message.Message):
    __slots__ = ("query",)
    QUERY_FIELD_NUMBER: _ClassVar[int]
    query: TimelineQuery
    def __init__(self, query: _Optional[_Union[TimelineQuery, _Mapping]] = ...) -> None: ...

class GetTimelineResponse(_message.Message):
    __slots__ = ("events", "cursor")
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    cursor: int
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ..., cursor: _Optional[int] = ...) -> None: ...

class StreamTimelineRequest(_message.Message):
    __slots__ = ("update_interval", "query")
    UPDATE_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    update_interval: _duration_pb2.Duration
    query: TimelineQuery
    def __init__(self, update_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., query: _Optional[_Union[TimelineQuery, _Mapping]] = ...) -> None: ...

class StreamTimelineResponse(_message.Message):
    __slots__ = ("events",)
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ...) -> None: ...

class CreateEventsRequest(_message.Message):
    __slots__ = ("entries",)
    class EventEntry(_message.Message):
        __slots__ = ("timestamp", "log", "call", "ingress", "cron_scheduled", "pubsub_publish", "pubsub_consume", "changeset_created", "changeset_state_changed", "deployment_runtime")
        TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
        LOG_FIELD_NUMBER: _ClassVar[int]
        CALL_FIELD_NUMBER: _ClassVar[int]
        INGRESS_FIELD_NUMBER: _ClassVar[int]
        CRON_SCHEDULED_FIELD_NUMBER: _ClassVar[int]
        PUBSUB_PUBLISH_FIELD_NUMBER: _ClassVar[int]
        PUBSUB_CONSUME_FIELD_NUMBER: _ClassVar[int]
        CHANGESET_CREATED_FIELD_NUMBER: _ClassVar[int]
        CHANGESET_STATE_CHANGED_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_RUNTIME_FIELD_NUMBER: _ClassVar[int]
        timestamp: _timestamp_pb2.Timestamp
        log: _event_pb2.LogEvent
        call: _event_pb2.CallEvent
        ingress: _event_pb2.IngressEvent
        cron_scheduled: _event_pb2.CronScheduledEvent
        pubsub_publish: _event_pb2.PubSubPublishEvent
        pubsub_consume: _event_pb2.PubSubConsumeEvent
        changeset_created: _event_pb2.ChangesetCreatedEvent
        changeset_state_changed: _event_pb2.ChangesetStateChangedEvent
        deployment_runtime: _event_pb2.DeploymentRuntimeEvent
        def __init__(self, timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., log: _Optional[_Union[_event_pb2.LogEvent, _Mapping]] = ..., call: _Optional[_Union[_event_pb2.CallEvent, _Mapping]] = ..., ingress: _Optional[_Union[_event_pb2.IngressEvent, _Mapping]] = ..., cron_scheduled: _Optional[_Union[_event_pb2.CronScheduledEvent, _Mapping]] = ..., pubsub_publish: _Optional[_Union[_event_pb2.PubSubPublishEvent, _Mapping]] = ..., pubsub_consume: _Optional[_Union[_event_pb2.PubSubConsumeEvent, _Mapping]] = ..., changeset_created: _Optional[_Union[_event_pb2.ChangesetCreatedEvent, _Mapping]] = ..., changeset_state_changed: _Optional[_Union[_event_pb2.ChangesetStateChangedEvent, _Mapping]] = ..., deployment_runtime: _Optional[_Union[_event_pb2.DeploymentRuntimeEvent, _Mapping]] = ...) -> None: ...
    ENTRIES_FIELD_NUMBER: _ClassVar[int]
    entries: _containers.RepeatedCompositeFieldContainer[CreateEventsRequest.EventEntry]
    def __init__(self, entries: _Optional[_Iterable[_Union[CreateEventsRequest.EventEntry, _Mapping]]] = ...) -> None: ...

class CreateEventsResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class DeleteOldEventsRequest(_message.Message):
    __slots__ = ("event_type", "age_seconds")
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    AGE_SECONDS_FIELD_NUMBER: _ClassVar[int]
    event_type: _event_pb2.EventType
    age_seconds: int
    def __init__(self, event_type: _Optional[_Union[_event_pb2.EventType, str]] = ..., age_seconds: _Optional[int] = ...) -> None: ...

class DeleteOldEventsResponse(_message.Message):
    __slots__ = ("deleted_count",)
    DELETED_COUNT_FIELD_NUMBER: _ClassVar[int]
    deleted_count: int
    def __init__(self, deleted_count: _Optional[int] = ...) -> None: ...
