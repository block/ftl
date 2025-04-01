from google.protobuf import struct_pb2 as _struct_pb2
from xyz.block.ftl.language.v1 import service_pb2 as _service_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetNewModuleFlagsRequest(_message.Message):
    __slots__ = ("language",)
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    language: str
    def __init__(self, language: _Optional[str] = ...) -> None: ...

class GetNewModuleFlagsResponse(_message.Message):
    __slots__ = ("flags",)
    class Flag(_message.Message):
        __slots__ = ("name", "help", "envar", "short", "placeholder", "default")
        NAME_FIELD_NUMBER: _ClassVar[int]
        HELP_FIELD_NUMBER: _ClassVar[int]
        ENVAR_FIELD_NUMBER: _ClassVar[int]
        SHORT_FIELD_NUMBER: _ClassVar[int]
        PLACEHOLDER_FIELD_NUMBER: _ClassVar[int]
        DEFAULT_FIELD_NUMBER: _ClassVar[int]
        name: str
        help: str
        envar: str
        short: str
        placeholder: str
        default: str
        def __init__(self, name: _Optional[str] = ..., help: _Optional[str] = ..., envar: _Optional[str] = ..., short: _Optional[str] = ..., placeholder: _Optional[str] = ..., default: _Optional[str] = ...) -> None: ...
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    flags: _containers.RepeatedCompositeFieldContainer[GetNewModuleFlagsResponse.Flag]
    def __init__(self, flags: _Optional[_Iterable[_Union[GetNewModuleFlagsResponse.Flag, _Mapping]]] = ...) -> None: ...

class NewModuleRequest(_message.Message):
    __slots__ = ("name", "dir", "project_config", "flags")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DIR_FIELD_NUMBER: _ClassVar[int]
    PROJECT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    name: str
    dir: str
    project_config: _service_pb2.ProjectConfig
    flags: _struct_pb2.Struct
    def __init__(self, name: _Optional[str] = ..., dir: _Optional[str] = ..., project_config: _Optional[_Union[_service_pb2.ProjectConfig, _Mapping]] = ..., flags: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class NewModuleResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetModuleConfigDefaultsRequest(_message.Message):
    __slots__ = ("dir",)
    DIR_FIELD_NUMBER: _ClassVar[int]
    dir: str
    def __init__(self, dir: _Optional[str] = ...) -> None: ...

class GetModuleConfigDefaultsResponse(_message.Message):
    __slots__ = ("deploy_dir", "build", "dev_mode_build", "build_lock", "watch", "language_config", "sql_root_dir")
    DEPLOY_DIR_FIELD_NUMBER: _ClassVar[int]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    DEV_MODE_BUILD_FIELD_NUMBER: _ClassVar[int]
    BUILD_LOCK_FIELD_NUMBER: _ClassVar[int]
    WATCH_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    SQL_ROOT_DIR_FIELD_NUMBER: _ClassVar[int]
    deploy_dir: str
    build: str
    dev_mode_build: str
    build_lock: str
    watch: _containers.RepeatedScalarFieldContainer[str]
    language_config: _struct_pb2.Struct
    sql_root_dir: str
    def __init__(self, deploy_dir: _Optional[str] = ..., build: _Optional[str] = ..., dev_mode_build: _Optional[str] = ..., build_lock: _Optional[str] = ..., watch: _Optional[_Iterable[str]] = ..., language_config: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., sql_root_dir: _Optional[str] = ...) -> None: ...
