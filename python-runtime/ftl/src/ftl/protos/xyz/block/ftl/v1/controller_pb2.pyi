from google.protobuf import struct_pb2 as _struct_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RegisterRunnerRequest(_message.Message):
    __slots__ = ("key", "endpoint", "deployment", "labels")
    KEY_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    key: str
    endpoint: str
    deployment: str
    labels: _struct_pb2.Struct
    def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class RegisterRunnerResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("controllers", "runners", "deployments", "routes")
    class Controller(_message.Message):
        __slots__ = ("key", "endpoint", "version")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        VERSION_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        version: str
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...
    class Runner(_message.Message):
        __slots__ = ("key", "endpoint", "deployment", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        deployment: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Deployment(_message.Message):
        __slots__ = ("key", "language", "name", "min_replicas", "replicas", "labels", "schema")
        KEY_FIELD_NUMBER: _ClassVar[int]
        LANGUAGE_FIELD_NUMBER: _ClassVar[int]
        NAME_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        SCHEMA_FIELD_NUMBER: _ClassVar[int]
        key: str
        language: str
        name: str
        min_replicas: int
        replicas: int
        labels: _struct_pb2.Struct
        schema: _schema_pb2.Module
        def __init__(self, key: _Optional[str] = ..., language: _Optional[str] = ..., name: _Optional[str] = ..., min_replicas: _Optional[int] = ..., replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
    class Route(_message.Message):
        __slots__ = ("module", "deployment", "endpoint")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        module: str
        deployment: str
        endpoint: str
        def __init__(self, module: _Optional[str] = ..., deployment: _Optional[str] = ..., endpoint: _Optional[str] = ...) -> None: ...
    CONTROLLERS_FIELD_NUMBER: _ClassVar[int]
    RUNNERS_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
    ROUTES_FIELD_NUMBER: _ClassVar[int]
    controllers: _containers.RepeatedCompositeFieldContainer[StatusResponse.Controller]
    runners: _containers.RepeatedCompositeFieldContainer[StatusResponse.Runner]
    deployments: _containers.RepeatedCompositeFieldContainer[StatusResponse.Deployment]
    routes: _containers.RepeatedCompositeFieldContainer[StatusResponse.Route]
    def __init__(self, controllers: _Optional[_Iterable[_Union[StatusResponse.Controller, _Mapping]]] = ..., runners: _Optional[_Iterable[_Union[StatusResponse.Runner, _Mapping]]] = ..., deployments: _Optional[_Iterable[_Union[StatusResponse.Deployment, _Mapping]]] = ..., routes: _Optional[_Iterable[_Union[StatusResponse.Route, _Mapping]]] = ...) -> None: ...

class ProcessListRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ProcessListResponse(_message.Message):
    __slots__ = ("processes",)
    class ProcessRunner(_message.Message):
        __slots__ = ("key", "endpoint", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Process(_message.Message):
        __slots__ = ("deployment", "min_replicas", "labels", "runner")
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        RUNNER_FIELD_NUMBER: _ClassVar[int]
        deployment: str
        min_replicas: int
        labels: _struct_pb2.Struct
        runner: ProcessListResponse.ProcessRunner
        def __init__(self, deployment: _Optional[str] = ..., min_replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., runner: _Optional[_Union[ProcessListResponse.ProcessRunner, _Mapping]] = ...) -> None: ...
    PROCESSES_FIELD_NUMBER: _ClassVar[int]
    processes: _containers.RepeatedCompositeFieldContainer[ProcessListResponse.Process]
    def __init__(self, processes: _Optional[_Iterable[_Union[ProcessListResponse.Process, _Mapping]]] = ...) -> None: ...

class GetDeploymentContextRequest(_message.Message):
    __slots__ = ("deployment",)
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    deployment: str
    def __init__(self, deployment: _Optional[str] = ...) -> None: ...

class GetDeploymentContextResponse(_message.Message):
    __slots__ = ("module", "deployment", "configs", "secrets", "databases", "routes")
    class DbType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        DB_TYPE_UNSPECIFIED: _ClassVar[GetDeploymentContextResponse.DbType]
        DB_TYPE_POSTGRES: _ClassVar[GetDeploymentContextResponse.DbType]
        DB_TYPE_MYSQL: _ClassVar[GetDeploymentContextResponse.DbType]
    DB_TYPE_UNSPECIFIED: GetDeploymentContextResponse.DbType
    DB_TYPE_POSTGRES: GetDeploymentContextResponse.DbType
    DB_TYPE_MYSQL: GetDeploymentContextResponse.DbType
    class DSN(_message.Message):
        __slots__ = ("name", "type", "dsn")
        NAME_FIELD_NUMBER: _ClassVar[int]
        TYPE_FIELD_NUMBER: _ClassVar[int]
        DSN_FIELD_NUMBER: _ClassVar[int]
        name: str
        type: GetDeploymentContextResponse.DbType
        dsn: str
        def __init__(self, name: _Optional[str] = ..., type: _Optional[_Union[GetDeploymentContextResponse.DbType, str]] = ..., dsn: _Optional[str] = ...) -> None: ...
    class Route(_message.Message):
        __slots__ = ("deployment", "uri")
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        URI_FIELD_NUMBER: _ClassVar[int]
        deployment: str
        uri: str
        def __init__(self, deployment: _Optional[str] = ..., uri: _Optional[str] = ...) -> None: ...
    class ConfigsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    class SecretsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    MODULE_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ROUTES_FIELD_NUMBER: _ClassVar[int]
    module: str
    deployment: str
    configs: _containers.ScalarMap[str, bytes]
    secrets: _containers.ScalarMap[str, bytes]
    databases: _containers.RepeatedCompositeFieldContainer[GetDeploymentContextResponse.DSN]
    routes: _containers.RepeatedCompositeFieldContainer[GetDeploymentContextResponse.Route]
    def __init__(self, module: _Optional[str] = ..., deployment: _Optional[str] = ..., configs: _Optional[_Mapping[str, bytes]] = ..., secrets: _Optional[_Mapping[str, bytes]] = ..., databases: _Optional[_Iterable[_Union[GetDeploymentContextResponse.DSN, _Mapping]]] = ..., routes: _Optional[_Iterable[_Union[GetDeploymentContextResponse.Route, _Mapping]]] = ...) -> None: ...
