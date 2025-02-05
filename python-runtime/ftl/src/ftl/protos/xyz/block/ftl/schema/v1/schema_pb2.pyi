from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AliasKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ALIAS_KIND_UNSPECIFIED: _ClassVar[AliasKind]
    ALIAS_KIND_JSON: _ClassVar[AliasKind]

class ChangesetState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CHANGESET_STATE_UNSPECIFIED: _ClassVar[ChangesetState]
    CHANGESET_STATE_PREPARING: _ClassVar[ChangesetState]
    CHANGESET_STATE_PREPARED: _ClassVar[ChangesetState]
    CHANGESET_STATE_COMMITTED: _ClassVar[ChangesetState]
    CHANGESET_STATE_DRAINED: _ClassVar[ChangesetState]
    CHANGESET_STATE_FINALIZED: _ClassVar[ChangesetState]
    CHANGESET_STATE_ROLLING_BACK: _ClassVar[ChangesetState]
    CHANGESET_STATE_FAILED: _ClassVar[ChangesetState]

class DeploymentState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DEPLOYMENT_STATE_UNSPECIFIED: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_PROVISIONING: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_READY: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_CANARY: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_CANONICAL: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_DRAINING: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_DE_PROVISIONING: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_DELETED: _ClassVar[DeploymentState]
    DEPLOYMENT_STATE_FAILED: _ClassVar[DeploymentState]

class FromOffset(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FROM_OFFSET_UNSPECIFIED: _ClassVar[FromOffset]
    FROM_OFFSET_BEGINNING: _ClassVar[FromOffset]
    FROM_OFFSET_LATEST: _ClassVar[FromOffset]
ALIAS_KIND_UNSPECIFIED: AliasKind
ALIAS_KIND_JSON: AliasKind
CHANGESET_STATE_UNSPECIFIED: ChangesetState
CHANGESET_STATE_PREPARING: ChangesetState
CHANGESET_STATE_PREPARED: ChangesetState
CHANGESET_STATE_COMMITTED: ChangesetState
CHANGESET_STATE_DRAINED: ChangesetState
CHANGESET_STATE_FINALIZED: ChangesetState
CHANGESET_STATE_ROLLING_BACK: ChangesetState
CHANGESET_STATE_FAILED: ChangesetState
DEPLOYMENT_STATE_UNSPECIFIED: DeploymentState
DEPLOYMENT_STATE_PROVISIONING: DeploymentState
DEPLOYMENT_STATE_READY: DeploymentState
DEPLOYMENT_STATE_CANARY: DeploymentState
DEPLOYMENT_STATE_CANONICAL: DeploymentState
DEPLOYMENT_STATE_DRAINING: DeploymentState
DEPLOYMENT_STATE_DE_PROVISIONING: DeploymentState
DEPLOYMENT_STATE_DELETED: DeploymentState
DEPLOYMENT_STATE_FAILED: DeploymentState
FROM_OFFSET_UNSPECIFIED: FromOffset
FROM_OFFSET_BEGINNING: FromOffset
FROM_OFFSET_LATEST: FromOffset

class AWSIAMAuthDatabaseConnector(_message.Message):
    __slots__ = ("pos", "username", "endpoint", "database")
    POS_FIELD_NUMBER: _ClassVar[int]
    USERNAME_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    DATABASE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    username: str
    endpoint: str
    database: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., username: _Optional[str] = ..., endpoint: _Optional[str] = ..., database: _Optional[str] = ...) -> None: ...

class Any(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Array(_message.Message):
    __slots__ = ("pos", "element")
    POS_FIELD_NUMBER: _ClassVar[int]
    ELEMENT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    element: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., element: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Bool(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Bytes(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Changeset(_message.Message):
    __slots__ = ("key", "created_at", "modules", "to_remove", "removing_modules", "state", "error")
    KEY_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TO_REMOVE_FIELD_NUMBER: _ClassVar[int]
    REMOVING_MODULES_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    key: str
    created_at: _timestamp_pb2.Timestamp
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    to_remove: _containers.RepeatedScalarFieldContainer[str]
    removing_modules: _containers.RepeatedCompositeFieldContainer[Module]
    state: ChangesetState
    error: str
    def __init__(self, key: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., to_remove: _Optional[_Iterable[str]] = ..., removing_modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., state: _Optional[_Union[ChangesetState, str]] = ..., error: _Optional[str] = ...) -> None: ...

class ChangesetCommittedEvent(_message.Message):
    __slots__ = ("key",)
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class ChangesetCreatedEvent(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: Changeset
    def __init__(self, changeset: _Optional[_Union[Changeset, _Mapping]] = ...) -> None: ...

class ChangesetDrainedEvent(_message.Message):
    __slots__ = ("key",)
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class ChangesetFailedEvent(_message.Message):
    __slots__ = ("key", "error")
    KEY_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    key: str
    error: str
    def __init__(self, key: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class ChangesetFinalizedEvent(_message.Message):
    __slots__ = ("key",)
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class ChangesetPreparedEvent(_message.Message):
    __slots__ = ("key",)
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: str
    def __init__(self, key: _Optional[str] = ...) -> None: ...

class Config(_message.Message):
    __slots__ = ("pos", "comments", "name", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class DSNDatabaseConnector(_message.Message):
    __slots__ = ("pos", "dsn")
    POS_FIELD_NUMBER: _ClassVar[int]
    DSN_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    dsn: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., dsn: _Optional[str] = ...) -> None: ...

class Data(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type_parameters", "fields", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type_parameters: _containers.RepeatedCompositeFieldContainer[TypeParameter]
    fields: _containers.RepeatedCompositeFieldContainer[Field]
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type_parameters: _Optional[_Iterable[_Union[TypeParameter, _Mapping]]] = ..., fields: _Optional[_Iterable[_Union[Field, _Mapping]]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class Database(_message.Message):
    __slots__ = ("pos", "runtime", "comments", "type", "name", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    runtime: DatabaseRuntime
    comments: _containers.RepeatedScalarFieldContainer[str]
    type: str
    name: str
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., runtime: _Optional[_Union[DatabaseRuntime, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., type: _Optional[str] = ..., name: _Optional[str] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class DatabaseConnector(_message.Message):
    __slots__ = ("awsiam_auth_database_connector", "dsn_database_connector")
    AWSIAM_AUTH_DATABASE_CONNECTOR_FIELD_NUMBER: _ClassVar[int]
    DSN_DATABASE_CONNECTOR_FIELD_NUMBER: _ClassVar[int]
    awsiam_auth_database_connector: AWSIAMAuthDatabaseConnector
    dsn_database_connector: DSNDatabaseConnector
    def __init__(self, awsiam_auth_database_connector: _Optional[_Union[AWSIAMAuthDatabaseConnector, _Mapping]] = ..., dsn_database_connector: _Optional[_Union[DSNDatabaseConnector, _Mapping]] = ...) -> None: ...

class DatabaseRuntime(_message.Message):
    __slots__ = ("connections",)
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: DatabaseRuntimeConnections
    def __init__(self, connections: _Optional[_Union[DatabaseRuntimeConnections, _Mapping]] = ...) -> None: ...

class DatabaseRuntimeConnections(_message.Message):
    __slots__ = ("read", "write")
    READ_FIELD_NUMBER: _ClassVar[int]
    WRITE_FIELD_NUMBER: _ClassVar[int]
    read: DatabaseConnector
    write: DatabaseConnector
    def __init__(self, read: _Optional[_Union[DatabaseConnector, _Mapping]] = ..., write: _Optional[_Union[DatabaseConnector, _Mapping]] = ...) -> None: ...

class DatabaseRuntimeEvent(_message.Message):
    __slots__ = ("deployment", "changeset", "id", "connections")
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    deployment: str
    changeset: str
    id: str
    connections: DatabaseRuntimeConnections
    def __init__(self, deployment: _Optional[str] = ..., changeset: _Optional[str] = ..., id: _Optional[str] = ..., connections: _Optional[_Union[DatabaseRuntimeConnections, _Mapping]] = ...) -> None: ...

class Decl(_message.Message):
    __slots__ = ("config", "data", "database", "enum", "secret", "topic", "type_alias", "verb")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    DATABASE_FIELD_NUMBER: _ClassVar[int]
    ENUM_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    TYPE_ALIAS_FIELD_NUMBER: _ClassVar[int]
    VERB_FIELD_NUMBER: _ClassVar[int]
    config: Config
    data: Data
    database: Database
    enum: Enum
    secret: Secret
    topic: Topic
    type_alias: TypeAlias
    verb: Verb
    def __init__(self, config: _Optional[_Union[Config, _Mapping]] = ..., data: _Optional[_Union[Data, _Mapping]] = ..., database: _Optional[_Union[Database, _Mapping]] = ..., enum: _Optional[_Union[Enum, _Mapping]] = ..., secret: _Optional[_Union[Secret, _Mapping]] = ..., topic: _Optional[_Union[Topic, _Mapping]] = ..., type_alias: _Optional[_Union[TypeAlias, _Mapping]] = ..., verb: _Optional[_Union[Verb, _Mapping]] = ...) -> None: ...

class DeploymentActivatedEvent(_message.Message):
    __slots__ = ("key", "activated_at", "min_replicas", "changeset")
    KEY_FIELD_NUMBER: _ClassVar[int]
    ACTIVATED_AT_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    key: str
    activated_at: _timestamp_pb2.Timestamp
    min_replicas: int
    changeset: str
    def __init__(self, key: _Optional[str] = ..., activated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., min_replicas: _Optional[int] = ..., changeset: _Optional[str] = ...) -> None: ...

class DeploymentCreatedEvent(_message.Message):
    __slots__ = ("key", "schema", "changeset")
    KEY_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    key: str
    schema: Module
    changeset: str
    def __init__(self, key: _Optional[str] = ..., schema: _Optional[_Union[Module, _Mapping]] = ..., changeset: _Optional[str] = ...) -> None: ...

class DeploymentDeactivatedEvent(_message.Message):
    __slots__ = ("key", "module_removed", "changeset")
    KEY_FIELD_NUMBER: _ClassVar[int]
    MODULE_REMOVED_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    key: str
    module_removed: bool
    changeset: str
    def __init__(self, key: _Optional[str] = ..., module_removed: bool = ..., changeset: _Optional[str] = ...) -> None: ...

class DeploymentReplicasUpdatedEvent(_message.Message):
    __slots__ = ("key", "replicas", "changeset")
    KEY_FIELD_NUMBER: _ClassVar[int]
    REPLICAS_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    key: str
    replicas: int
    changeset: str
    def __init__(self, key: _Optional[str] = ..., replicas: _Optional[int] = ..., changeset: _Optional[str] = ...) -> None: ...

class DeploymentSchemaUpdatedEvent(_message.Message):
    __slots__ = ("key", "schema", "changeset")
    KEY_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    key: str
    schema: Module
    changeset: str
    def __init__(self, key: _Optional[str] = ..., schema: _Optional[_Union[Module, _Mapping]] = ..., changeset: _Optional[str] = ...) -> None: ...

class Enum(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type", "variants")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VARIANTS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type: Type
    variants: _containers.RepeatedCompositeFieldContainer[EnumVariant]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., variants: _Optional[_Iterable[_Union[EnumVariant, _Mapping]]] = ...) -> None: ...

class EnumVariant(_message.Message):
    __slots__ = ("pos", "comments", "name", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    value: Value
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., value: _Optional[_Union[Value, _Mapping]] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ("changeset_committed_event", "changeset_created_event", "changeset_drained_event", "changeset_failed_event", "changeset_finalized_event", "changeset_prepared_event", "database_runtime_event", "deployment_activated_event", "deployment_created_event", "deployment_deactivated_event", "deployment_replicas_updated_event", "deployment_schema_updated_event", "module_runtime_event", "topic_runtime_event", "verb_runtime_event")
    CHANGESET_COMMITTED_EVENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_CREATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_DRAINED_EVENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FAILED_EVENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FINALIZED_EVENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_PREPARED_EVENT_FIELD_NUMBER: _ClassVar[int]
    DATABASE_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_ACTIVATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_CREATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_DEACTIVATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_REPLICAS_UPDATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_SCHEMA_UPDATED_EVENT_FIELD_NUMBER: _ClassVar[int]
    MODULE_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    TOPIC_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    VERB_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    changeset_committed_event: ChangesetCommittedEvent
    changeset_created_event: ChangesetCreatedEvent
    changeset_drained_event: ChangesetDrainedEvent
    changeset_failed_event: ChangesetFailedEvent
    changeset_finalized_event: ChangesetFinalizedEvent
    changeset_prepared_event: ChangesetPreparedEvent
    database_runtime_event: DatabaseRuntimeEvent
    deployment_activated_event: DeploymentActivatedEvent
    deployment_created_event: DeploymentCreatedEvent
    deployment_deactivated_event: DeploymentDeactivatedEvent
    deployment_replicas_updated_event: DeploymentReplicasUpdatedEvent
    deployment_schema_updated_event: DeploymentSchemaUpdatedEvent
    module_runtime_event: ModuleRuntimeEvent
    topic_runtime_event: TopicRuntimeEvent
    verb_runtime_event: VerbRuntimeEvent
    def __init__(self, changeset_committed_event: _Optional[_Union[ChangesetCommittedEvent, _Mapping]] = ..., changeset_created_event: _Optional[_Union[ChangesetCreatedEvent, _Mapping]] = ..., changeset_drained_event: _Optional[_Union[ChangesetDrainedEvent, _Mapping]] = ..., changeset_failed_event: _Optional[_Union[ChangesetFailedEvent, _Mapping]] = ..., changeset_finalized_event: _Optional[_Union[ChangesetFinalizedEvent, _Mapping]] = ..., changeset_prepared_event: _Optional[_Union[ChangesetPreparedEvent, _Mapping]] = ..., database_runtime_event: _Optional[_Union[DatabaseRuntimeEvent, _Mapping]] = ..., deployment_activated_event: _Optional[_Union[DeploymentActivatedEvent, _Mapping]] = ..., deployment_created_event: _Optional[_Union[DeploymentCreatedEvent, _Mapping]] = ..., deployment_deactivated_event: _Optional[_Union[DeploymentDeactivatedEvent, _Mapping]] = ..., deployment_replicas_updated_event: _Optional[_Union[DeploymentReplicasUpdatedEvent, _Mapping]] = ..., deployment_schema_updated_event: _Optional[_Union[DeploymentSchemaUpdatedEvent, _Mapping]] = ..., module_runtime_event: _Optional[_Union[ModuleRuntimeEvent, _Mapping]] = ..., topic_runtime_event: _Optional[_Union[TopicRuntimeEvent, _Mapping]] = ..., verb_runtime_event: _Optional[_Union[VerbRuntimeEvent, _Mapping]] = ...) -> None: ...

class Field(_message.Message):
    __slots__ = ("pos", "comments", "name", "type", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class Float(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class IngressPathComponent(_message.Message):
    __slots__ = ("ingress_path_literal", "ingress_path_parameter")
    INGRESS_PATH_LITERAL_FIELD_NUMBER: _ClassVar[int]
    INGRESS_PATH_PARAMETER_FIELD_NUMBER: _ClassVar[int]
    ingress_path_literal: IngressPathLiteral
    ingress_path_parameter: IngressPathParameter
    def __init__(self, ingress_path_literal: _Optional[_Union[IngressPathLiteral, _Mapping]] = ..., ingress_path_parameter: _Optional[_Union[IngressPathParameter, _Mapping]] = ...) -> None: ...

class IngressPathLiteral(_message.Message):
    __slots__ = ("pos", "text")
    POS_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    text: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., text: _Optional[str] = ...) -> None: ...

class IngressPathParameter(_message.Message):
    __slots__ = ("pos", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class Int(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class IntValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: int
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[int] = ...) -> None: ...

class Map(_message.Message):
    __slots__ = ("pos", "key", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    key: Type
    value: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., key: _Optional[_Union[Type, _Mapping]] = ..., value: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ("alias", "artefact", "calls", "config", "cron_job", "databases", "encoding", "ingress", "partitions", "publisher", "retry", "sql_column", "sql_migration", "sql_query", "secrets", "subscriber", "type_map")
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    ARTEFACT_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    CRON_JOB_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ENCODING_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    PUBLISHER_FIELD_NUMBER: _ClassVar[int]
    RETRY_FIELD_NUMBER: _ClassVar[int]
    SQL_COLUMN_FIELD_NUMBER: _ClassVar[int]
    SQL_MIGRATION_FIELD_NUMBER: _ClassVar[int]
    SQL_QUERY_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIBER_FIELD_NUMBER: _ClassVar[int]
    TYPE_MAP_FIELD_NUMBER: _ClassVar[int]
    alias: MetadataAlias
    artefact: MetadataArtefact
    calls: MetadataCalls
    config: MetadataConfig
    cron_job: MetadataCronJob
    databases: MetadataDatabases
    encoding: MetadataEncoding
    ingress: MetadataIngress
    partitions: MetadataPartitions
    publisher: MetadataPublisher
    retry: MetadataRetry
    sql_column: MetadataSQLColumn
    sql_migration: MetadataSQLMigration
    sql_query: MetadataSQLQuery
    secrets: MetadataSecrets
    subscriber: MetadataSubscriber
    type_map: MetadataTypeMap
    def __init__(self, alias: _Optional[_Union[MetadataAlias, _Mapping]] = ..., artefact: _Optional[_Union[MetadataArtefact, _Mapping]] = ..., calls: _Optional[_Union[MetadataCalls, _Mapping]] = ..., config: _Optional[_Union[MetadataConfig, _Mapping]] = ..., cron_job: _Optional[_Union[MetadataCronJob, _Mapping]] = ..., databases: _Optional[_Union[MetadataDatabases, _Mapping]] = ..., encoding: _Optional[_Union[MetadataEncoding, _Mapping]] = ..., ingress: _Optional[_Union[MetadataIngress, _Mapping]] = ..., partitions: _Optional[_Union[MetadataPartitions, _Mapping]] = ..., publisher: _Optional[_Union[MetadataPublisher, _Mapping]] = ..., retry: _Optional[_Union[MetadataRetry, _Mapping]] = ..., sql_column: _Optional[_Union[MetadataSQLColumn, _Mapping]] = ..., sql_migration: _Optional[_Union[MetadataSQLMigration, _Mapping]] = ..., sql_query: _Optional[_Union[MetadataSQLQuery, _Mapping]] = ..., secrets: _Optional[_Union[MetadataSecrets, _Mapping]] = ..., subscriber: _Optional[_Union[MetadataSubscriber, _Mapping]] = ..., type_map: _Optional[_Union[MetadataTypeMap, _Mapping]] = ...) -> None: ...

class MetadataAlias(_message.Message):
    __slots__ = ("pos", "kind", "alias")
    POS_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    kind: AliasKind
    alias: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., kind: _Optional[_Union[AliasKind, str]] = ..., alias: _Optional[str] = ...) -> None: ...

class MetadataArtefact(_message.Message):
    __slots__ = ("pos", "path", "digest", "executable")
    POS_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    EXECUTABLE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    path: str
    digest: str
    executable: bool
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., path: _Optional[str] = ..., digest: _Optional[str] = ..., executable: bool = ...) -> None: ...

class MetadataCalls(_message.Message):
    __slots__ = ("pos", "calls")
    POS_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    calls: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., calls: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataConfig(_message.Message):
    __slots__ = ("pos", "config")
    POS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    config: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., config: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataCronJob(_message.Message):
    __slots__ = ("pos", "cron")
    POS_FIELD_NUMBER: _ClassVar[int]
    CRON_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    cron: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., cron: _Optional[str] = ...) -> None: ...

class MetadataDatabases(_message.Message):
    __slots__ = ("pos", "calls")
    POS_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    calls: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., calls: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataEncoding(_message.Message):
    __slots__ = ("pos", "type", "lenient")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    LENIENT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: str
    lenient: bool
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[str] = ..., lenient: bool = ...) -> None: ...

class MetadataIngress(_message.Message):
    __slots__ = ("pos", "type", "method", "path")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METHOD_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: str
    method: str
    path: _containers.RepeatedCompositeFieldContainer[IngressPathComponent]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[str] = ..., method: _Optional[str] = ..., path: _Optional[_Iterable[_Union[IngressPathComponent, _Mapping]]] = ...) -> None: ...

class MetadataPartitions(_message.Message):
    __slots__ = ("pos", "partitions")
    POS_FIELD_NUMBER: _ClassVar[int]
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    partitions: int
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., partitions: _Optional[int] = ...) -> None: ...

class MetadataPublisher(_message.Message):
    __slots__ = ("pos", "topics")
    POS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    topics: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., topics: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataRetry(_message.Message):
    __slots__ = ("pos", "count", "min_backoff", "max_backoff", "catch")
    POS_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    MIN_BACKOFF_FIELD_NUMBER: _ClassVar[int]
    MAX_BACKOFF_FIELD_NUMBER: _ClassVar[int]
    CATCH_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    count: int
    min_backoff: str
    max_backoff: str
    catch: Ref
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., count: _Optional[int] = ..., min_backoff: _Optional[str] = ..., max_backoff: _Optional[str] = ..., catch: _Optional[_Union[Ref, _Mapping]] = ...) -> None: ...

class MetadataSQLColumn(_message.Message):
    __slots__ = ("pos", "table", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    table: str
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., table: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class MetadataSQLMigration(_message.Message):
    __slots__ = ("pos", "digest")
    POS_FIELD_NUMBER: _ClassVar[int]
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    digest: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., digest: _Optional[str] = ...) -> None: ...

class MetadataSQLQuery(_message.Message):
    __slots__ = ("pos", "command", "query")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    command: str
    query: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., command: _Optional[str] = ..., query: _Optional[str] = ...) -> None: ...

class MetadataSecrets(_message.Message):
    __slots__ = ("pos", "secrets")
    POS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    secrets: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., secrets: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataSubscriber(_message.Message):
    __slots__ = ("pos", "topic", "from_offset", "dead_letter")
    POS_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    FROM_OFFSET_FIELD_NUMBER: _ClassVar[int]
    DEAD_LETTER_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    topic: Ref
    from_offset: FromOffset
    dead_letter: bool
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., topic: _Optional[_Union[Ref, _Mapping]] = ..., from_offset: _Optional[_Union[FromOffset, str]] = ..., dead_letter: bool = ...) -> None: ...

class MetadataTypeMap(_message.Message):
    __slots__ = ("pos", "runtime", "native_name")
    POS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    NATIVE_NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    runtime: str
    native_name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., runtime: _Optional[str] = ..., native_name: _Optional[str] = ...) -> None: ...

class Module(_message.Message):
    __slots__ = ("pos", "comments", "builtin", "name", "metadata", "decls", "runtime")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    BUILTIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    DECLS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    builtin: bool
    name: str
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    decls: _containers.RepeatedCompositeFieldContainer[Decl]
    runtime: ModuleRuntime
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., builtin: bool = ..., name: _Optional[str] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ..., decls: _Optional[_Iterable[_Union[Decl, _Mapping]]] = ..., runtime: _Optional[_Union[ModuleRuntime, _Mapping]] = ...) -> None: ...

class ModuleRuntime(_message.Message):
    __slots__ = ("base", "scaling", "deployment", "runner")
    BASE_FIELD_NUMBER: _ClassVar[int]
    SCALING_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    RUNNER_FIELD_NUMBER: _ClassVar[int]
    base: ModuleRuntimeBase
    scaling: ModuleRuntimeScaling
    deployment: ModuleRuntimeDeployment
    runner: ModuleRuntimeRunner
    def __init__(self, base: _Optional[_Union[ModuleRuntimeBase, _Mapping]] = ..., scaling: _Optional[_Union[ModuleRuntimeScaling, _Mapping]] = ..., deployment: _Optional[_Union[ModuleRuntimeDeployment, _Mapping]] = ..., runner: _Optional[_Union[ModuleRuntimeRunner, _Mapping]] = ...) -> None: ...

class ModuleRuntimeBase(_message.Message):
    __slots__ = ("create_time", "language", "os", "arch", "image")
    CREATE_TIME_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    OS_FIELD_NUMBER: _ClassVar[int]
    ARCH_FIELD_NUMBER: _ClassVar[int]
    IMAGE_FIELD_NUMBER: _ClassVar[int]
    create_time: _timestamp_pb2.Timestamp
    language: str
    os: str
    arch: str
    image: str
    def __init__(self, create_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., language: _Optional[str] = ..., os: _Optional[str] = ..., arch: _Optional[str] = ..., image: _Optional[str] = ...) -> None: ...

class ModuleRuntimeDeployment(_message.Message):
    __slots__ = ("deployment_key", "created_at", "activated_at", "state")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    ACTIVATED_AT_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    created_at: _timestamp_pb2.Timestamp
    activated_at: _timestamp_pb2.Timestamp
    state: DeploymentState
    def __init__(self, deployment_key: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., activated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., state: _Optional[_Union[DeploymentState, str]] = ...) -> None: ...

class ModuleRuntimeEvent(_message.Message):
    __slots__ = ("key", "changeset", "base", "scaling", "deployment", "runner")
    KEY_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    BASE_FIELD_NUMBER: _ClassVar[int]
    SCALING_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    RUNNER_FIELD_NUMBER: _ClassVar[int]
    key: str
    changeset: str
    base: ModuleRuntimeBase
    scaling: ModuleRuntimeScaling
    deployment: ModuleRuntimeDeployment
    runner: ModuleRuntimeRunner
    def __init__(self, key: _Optional[str] = ..., changeset: _Optional[str] = ..., base: _Optional[_Union[ModuleRuntimeBase, _Mapping]] = ..., scaling: _Optional[_Union[ModuleRuntimeScaling, _Mapping]] = ..., deployment: _Optional[_Union[ModuleRuntimeDeployment, _Mapping]] = ..., runner: _Optional[_Union[ModuleRuntimeRunner, _Mapping]] = ...) -> None: ...

class ModuleRuntimeRunner(_message.Message):
    __slots__ = ("endpoint",)
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    endpoint: str
    def __init__(self, endpoint: _Optional[str] = ...) -> None: ...

class ModuleRuntimeScaling(_message.Message):
    __slots__ = ("min_replicas",)
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    min_replicas: int
    def __init__(self, min_replicas: _Optional[int] = ...) -> None: ...

class Optional(_message.Message):
    __slots__ = ("pos", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Position(_message.Message):
    __slots__ = ("filename", "line", "column")
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    LINE_FIELD_NUMBER: _ClassVar[int]
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    filename: str
    line: int
    column: int
    def __init__(self, filename: _Optional[str] = ..., line: _Optional[int] = ..., column: _Optional[int] = ...) -> None: ...

class Ref(_message.Message):
    __slots__ = ("pos", "module", "name", "type_parameters")
    POS_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    module: str
    name: str
    type_parameters: _containers.RepeatedCompositeFieldContainer[Type]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., module: _Optional[str] = ..., name: _Optional[str] = ..., type_parameters: _Optional[_Iterable[_Union[Type, _Mapping]]] = ...) -> None: ...

class RuntimeEvent(_message.Message):
    __slots__ = ("database_runtime_event", "module_runtime_event", "topic_runtime_event", "verb_runtime_event")
    DATABASE_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    MODULE_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    TOPIC_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    VERB_RUNTIME_EVENT_FIELD_NUMBER: _ClassVar[int]
    database_runtime_event: DatabaseRuntimeEvent
    module_runtime_event: ModuleRuntimeEvent
    topic_runtime_event: TopicRuntimeEvent
    verb_runtime_event: VerbRuntimeEvent
    def __init__(self, database_runtime_event: _Optional[_Union[DatabaseRuntimeEvent, _Mapping]] = ..., module_runtime_event: _Optional[_Union[ModuleRuntimeEvent, _Mapping]] = ..., topic_runtime_event: _Optional[_Union[TopicRuntimeEvent, _Mapping]] = ..., verb_runtime_event: _Optional[_Union[VerbRuntimeEvent, _Mapping]] = ...) -> None: ...

class Schema(_message.Message):
    __slots__ = ("pos", "modules")
    POS_FIELD_NUMBER: _ClassVar[int]
    MODULES_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ...) -> None: ...

class SchemaState(_message.Message):
    __slots__ = ("modules", "changesets", "runtime_events")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    CHANGESETS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_EVENTS_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    changesets: _containers.RepeatedCompositeFieldContainer[Changeset]
    runtime_events: _containers.RepeatedCompositeFieldContainer[RuntimeEvent]
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., changesets: _Optional[_Iterable[_Union[Changeset, _Mapping]]] = ..., runtime_events: _Optional[_Iterable[_Union[RuntimeEvent, _Mapping]]] = ...) -> None: ...

class Secret(_message.Message):
    __slots__ = ("pos", "comments", "name", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class String(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class StringValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[str] = ...) -> None: ...

class Time(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Topic(_message.Message):
    __slots__ = ("pos", "runtime", "comments", "export", "name", "event", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    runtime: TopicRuntime
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    event: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., runtime: _Optional[_Union[TopicRuntime, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., event: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class TopicRuntime(_message.Message):
    __slots__ = ("kafka_brokers", "topic_id")
    KAFKA_BROKERS_FIELD_NUMBER: _ClassVar[int]
    TOPIC_ID_FIELD_NUMBER: _ClassVar[int]
    kafka_brokers: _containers.RepeatedScalarFieldContainer[str]
    topic_id: str
    def __init__(self, kafka_brokers: _Optional[_Iterable[str]] = ..., topic_id: _Optional[str] = ...) -> None: ...

class TopicRuntimeEvent(_message.Message):
    __slots__ = ("deployment", "changeset", "id", "payload")
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    deployment: str
    changeset: str
    id: str
    payload: TopicRuntime
    def __init__(self, deployment: _Optional[str] = ..., changeset: _Optional[str] = ..., id: _Optional[str] = ..., payload: _Optional[_Union[TopicRuntime, _Mapping]] = ...) -> None: ...

class Type(_message.Message):
    __slots__ = ("any", "array", "bool", "bytes", "float", "int", "map", "optional", "ref", "string", "time", "unit")
    ANY_FIELD_NUMBER: _ClassVar[int]
    ARRAY_FIELD_NUMBER: _ClassVar[int]
    BOOL_FIELD_NUMBER: _ClassVar[int]
    BYTES_FIELD_NUMBER: _ClassVar[int]
    FLOAT_FIELD_NUMBER: _ClassVar[int]
    INT_FIELD_NUMBER: _ClassVar[int]
    MAP_FIELD_NUMBER: _ClassVar[int]
    OPTIONAL_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    UNIT_FIELD_NUMBER: _ClassVar[int]
    any: Any
    array: Array
    bool: Bool
    bytes: Bytes
    float: Float
    int: Int
    map: Map
    optional: Optional
    ref: Ref
    string: String
    time: Time
    unit: Unit
    def __init__(self, any: _Optional[_Union[Any, _Mapping]] = ..., array: _Optional[_Union[Array, _Mapping]] = ..., bool: _Optional[_Union[Bool, _Mapping]] = ..., bytes: _Optional[_Union[Bytes, _Mapping]] = ..., float: _Optional[_Union[Float, _Mapping]] = ..., int: _Optional[_Union[Int, _Mapping]] = ..., map: _Optional[_Union[Map, _Mapping]] = ..., optional: _Optional[_Union[Optional, _Mapping]] = ..., ref: _Optional[_Union[Ref, _Mapping]] = ..., string: _Optional[_Union[String, _Mapping]] = ..., time: _Optional[_Union[Time, _Mapping]] = ..., unit: _Optional[_Union[Unit, _Mapping]] = ...) -> None: ...

class TypeAlias(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class TypeParameter(_message.Message):
    __slots__ = ("pos", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class TypeValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Unit(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Value(_message.Message):
    __slots__ = ("int_value", "string_value", "type_value")
    INT_VALUE_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_VALUE_FIELD_NUMBER: _ClassVar[int]
    int_value: IntValue
    string_value: StringValue
    type_value: TypeValue
    def __init__(self, int_value: _Optional[_Union[IntValue, _Mapping]] = ..., string_value: _Optional[_Union[StringValue, _Mapping]] = ..., type_value: _Optional[_Union[TypeValue, _Mapping]] = ...) -> None: ...

class Verb(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "request", "response", "metadata", "runtime")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    request: Type
    response: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    runtime: VerbRuntime
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., request: _Optional[_Union[Type, _Mapping]] = ..., response: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ..., runtime: _Optional[_Union[VerbRuntime, _Mapping]] = ...) -> None: ...

class VerbRuntime(_message.Message):
    __slots__ = ("subscription",)
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    subscription: VerbRuntimeSubscription
    def __init__(self, subscription: _Optional[_Union[VerbRuntimeSubscription, _Mapping]] = ...) -> None: ...

class VerbRuntimeEvent(_message.Message):
    __slots__ = ("deployment", "changeset", "id", "subscription")
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    deployment: str
    changeset: str
    id: str
    subscription: VerbRuntimeSubscription
    def __init__(self, deployment: _Optional[str] = ..., changeset: _Optional[str] = ..., id: _Optional[str] = ..., subscription: _Optional[_Union[VerbRuntimeSubscription, _Mapping]] = ...) -> None: ...

class VerbRuntimeSubscription(_message.Message):
    __slots__ = ("kafka_brokers",)
    KAFKA_BROKERS_FIELD_NUMBER: _ClassVar[int]
    kafka_brokers: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, kafka_brokers: _Optional[_Iterable[str]] = ...) -> None: ...
