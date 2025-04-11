from xyz.block.ftl.buildengine.v1 import buildengine_pb2 as _buildengine_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.timeline.v1 import timeline_pb2 as _timeline_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1 import verb_pb2 as _verb_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Edges(_message.Message):
    __slots__ = ("out",)
    IN_FIELD_NUMBER: _ClassVar[int]
    OUT_FIELD_NUMBER: _ClassVar[int]
    out: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, out: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ..., **kwargs) -> None: ...

class Config(_message.Message):
    __slots__ = ("config", "edges", "schema")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    config: _schema_pb2.Config
    edges: Edges
    schema: str
    def __init__(self, config: _Optional[_Union[_schema_pb2.Config, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class Data(_message.Message):
    __slots__ = ("data", "schema", "edges")
    DATA_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    data: _schema_pb2.Data
    schema: str
    edges: Edges
    def __init__(self, data: _Optional[_Union[_schema_pb2.Data, _Mapping]] = ..., schema: _Optional[str] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ...) -> None: ...

class Database(_message.Message):
    __slots__ = ("database", "edges", "schema")
    DATABASE_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    database: _schema_pb2.Database
    edges: Edges
    schema: str
    def __init__(self, database: _Optional[_Union[_schema_pb2.Database, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class Enum(_message.Message):
    __slots__ = ("enum", "edges", "schema")
    ENUM_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    enum: _schema_pb2.Enum
    edges: Edges
    schema: str
    def __init__(self, enum: _Optional[_Union[_schema_pb2.Enum, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class Topic(_message.Message):
    __slots__ = ("topic", "edges", "schema")
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    topic: _schema_pb2.Topic
    edges: Edges
    schema: str
    def __init__(self, topic: _Optional[_Union[_schema_pb2.Topic, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class TypeAlias(_message.Message):
    __slots__ = ("typealias", "edges", "schema")
    TYPEALIAS_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    typealias: _schema_pb2.TypeAlias
    edges: Edges
    schema: str
    def __init__(self, typealias: _Optional[_Union[_schema_pb2.TypeAlias, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class Secret(_message.Message):
    __slots__ = ("secret", "edges", "schema")
    SECRET_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    secret: _schema_pb2.Secret
    edges: Edges
    schema: str
    def __init__(self, secret: _Optional[_Union[_schema_pb2.Secret, _Mapping]] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ..., schema: _Optional[str] = ...) -> None: ...

class Verb(_message.Message):
    __slots__ = ("verb", "schema", "json_request_schema", "edges")
    VERB_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    JSON_REQUEST_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    EDGES_FIELD_NUMBER: _ClassVar[int]
    verb: _schema_pb2.Verb
    schema: str
    json_request_schema: str
    edges: Edges
    def __init__(self, verb: _Optional[_Union[_schema_pb2.Verb, _Mapping]] = ..., schema: _Optional[str] = ..., json_request_schema: _Optional[str] = ..., edges: _Optional[_Union[Edges, _Mapping]] = ...) -> None: ...

class Module(_message.Message):
    __slots__ = ("name", "schema", "runtime", "verbs", "data", "secrets", "configs", "databases", "enums", "topics", "typealiases")
    NAME_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    VERBS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ENUMS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    TYPEALIASES_FIELD_NUMBER: _ClassVar[int]
    name: str
    schema: str
    runtime: _schema_pb2.ModuleRuntime
    verbs: _containers.RepeatedCompositeFieldContainer[Verb]
    data: _containers.RepeatedCompositeFieldContainer[Data]
    secrets: _containers.RepeatedCompositeFieldContainer[Secret]
    configs: _containers.RepeatedCompositeFieldContainer[Config]
    databases: _containers.RepeatedCompositeFieldContainer[Database]
    enums: _containers.RepeatedCompositeFieldContainer[Enum]
    topics: _containers.RepeatedCompositeFieldContainer[Topic]
    typealiases: _containers.RepeatedCompositeFieldContainer[TypeAlias]
    def __init__(self, name: _Optional[str] = ..., schema: _Optional[str] = ..., runtime: _Optional[_Union[_schema_pb2.ModuleRuntime, _Mapping]] = ..., verbs: _Optional[_Iterable[_Union[Verb, _Mapping]]] = ..., data: _Optional[_Iterable[_Union[Data, _Mapping]]] = ..., secrets: _Optional[_Iterable[_Union[Secret, _Mapping]]] = ..., configs: _Optional[_Iterable[_Union[Config, _Mapping]]] = ..., databases: _Optional[_Iterable[_Union[Database, _Mapping]]] = ..., enums: _Optional[_Iterable[_Union[Enum, _Mapping]]] = ..., topics: _Optional[_Iterable[_Union[Topic, _Mapping]]] = ..., typealiases: _Optional[_Iterable[_Union[TypeAlias, _Mapping]]] = ...) -> None: ...

class TopologyGroup(_message.Message):
    __slots__ = ("modules",)
    MODULES_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, modules: _Optional[_Iterable[str]] = ...) -> None: ...

class Topology(_message.Message):
    __slots__ = ("levels",)
    LEVELS_FIELD_NUMBER: _ClassVar[int]
    levels: _containers.RepeatedCompositeFieldContainer[TopologyGroup]
    def __init__(self, levels: _Optional[_Iterable[_Union[TopologyGroup, _Mapping]]] = ...) -> None: ...

class GetModulesRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetModulesResponse(_message.Message):
    __slots__ = ("modules", "topology")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TOPOLOGY_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    topology: Topology
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., topology: _Optional[_Union[Topology, _Mapping]] = ...) -> None: ...

class StreamModulesRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StreamModulesResponse(_message.Message):
    __slots__ = ("modules", "topology")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TOPOLOGY_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    topology: Topology
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., topology: _Optional[_Union[Topology, _Mapping]] = ...) -> None: ...

class GetConfigRequest(_message.Message):
    __slots__ = ("name", "module")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ...) -> None: ...

class GetConfigResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetConfigRequest(_message.Message):
    __slots__ = ("name", "module", "value")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    value: bytes
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetConfigResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class GetSecretRequest(_message.Message):
    __slots__ = ("name", "module")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ...) -> None: ...

class GetSecretResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetSecretRequest(_message.Message):
    __slots__ = ("name", "module", "value")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    value: bytes
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetSecretResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class GetInfoRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetInfoResponse(_message.Message):
    __slots__ = ("version", "build_time")
    VERSION_FIELD_NUMBER: _ClassVar[int]
    BUILD_TIME_FIELD_NUMBER: _ClassVar[int]
    version: str
    build_time: str
    def __init__(self, version: _Optional[str] = ..., build_time: _Optional[str] = ...) -> None: ...

class ExecuteGooseRequest(_message.Message):
    __slots__ = ("prompt",)
    PROMPT_FIELD_NUMBER: _ClassVar[int]
    prompt: str
    def __init__(self, prompt: _Optional[str] = ...) -> None: ...

class ExecuteGooseResponse(_message.Message):
    __slots__ = ("response", "source")
    class Source(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        SOURCE_UNSPECIFIED: _ClassVar[ExecuteGooseResponse.Source]
        SOURCE_STDOUT: _ClassVar[ExecuteGooseResponse.Source]
        SOURCE_STDERR: _ClassVar[ExecuteGooseResponse.Source]
        SOURCE_COMPLETION: _ClassVar[ExecuteGooseResponse.Source]
    SOURCE_UNSPECIFIED: ExecuteGooseResponse.Source
    SOURCE_STDOUT: ExecuteGooseResponse.Source
    SOURCE_STDERR: ExecuteGooseResponse.Source
    SOURCE_COMPLETION: ExecuteGooseResponse.Source
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    response: str
    source: ExecuteGooseResponse.Source
    def __init__(self, response: _Optional[str] = ..., source: _Optional[_Union[ExecuteGooseResponse.Source, str]] = ...) -> None: ...
