from xyz.block.ftl.language.v1 import language_pb2 as _language_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ReloadRequest(_message.Message):
    __slots__ = ("force", "new_deployment_key")
    FORCE_FIELD_NUMBER: _ClassVar[int]
    NEW_DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    force: bool
    new_deployment_key: str
    def __init__(self, force: bool = ..., new_deployment_key: _Optional[str] = ...) -> None: ...

class ReloadResponse(_message.Message):
    __slots__ = ("state", "failed")
    STATE_FIELD_NUMBER: _ClassVar[int]
    FAILED_FIELD_NUMBER: _ClassVar[int]
    state: SchemaState
    failed: bool
    def __init__(self, state: _Optional[_Union[SchemaState, _Mapping]] = ..., failed: bool = ...) -> None: ...

class WatchRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class WatchResponse(_message.Message):
    __slots__ = ("state",)
    STATE_FIELD_NUMBER: _ClassVar[int]
    state: SchemaState
    def __init__(self, state: _Optional[_Union[SchemaState, _Mapping]] = ...) -> None: ...

class RunnerInfoRequest(_message.Message):
    __slots__ = ("address", "deployment", "databases")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    address: str
    deployment: str
    databases: _containers.RepeatedCompositeFieldContainer[Database]
    def __init__(self, address: _Optional[str] = ..., deployment: _Optional[str] = ..., databases: _Optional[_Iterable[_Union[Database, _Mapping]]] = ...) -> None: ...

class Database(_message.Message):
    __slots__ = ("name", "address")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    name: str
    address: str
    def __init__(self, name: _Optional[str] = ..., address: _Optional[str] = ...) -> None: ...

class RunnerInfoResponse(_message.Message):
    __slots__ = ("outdated",)
    OUTDATED_FIELD_NUMBER: _ClassVar[int]
    outdated: bool
    def __init__(self, outdated: bool = ...) -> None: ...

class ReloadNotRequired(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ReloadSuccess(_message.Message):
    __slots__ = ("state",)
    STATE_FIELD_NUMBER: _ClassVar[int]
    state: SchemaState
    def __init__(self, state: _Optional[_Union[SchemaState, _Mapping]] = ...) -> None: ...

class ReloadFailed(_message.Message):
    __slots__ = ("state",)
    STATE_FIELD_NUMBER: _ClassVar[int]
    state: SchemaState
    def __init__(self, state: _Optional[_Union[SchemaState, _Mapping]] = ...) -> None: ...

class SchemaState(_message.Message):
    __slots__ = ("module", "errors", "new_runner_required", "version")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    NEW_RUNNER_REQUIRED_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    module: _schema_pb2.Module
    errors: _language_pb2.ErrorList
    new_runner_required: bool
    version: int
    def __init__(self, module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., errors: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ..., new_runner_required: bool = ..., version: _Optional[int] = ...) -> None: ...
