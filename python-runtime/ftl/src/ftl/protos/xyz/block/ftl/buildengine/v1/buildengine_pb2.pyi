from xyz.block.ftl.language.v1 import language_pb2 as _language_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EngineStarted(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class EngineEnded(_message.Message):
    __slots__ = ("module_errors",)
    class ModuleErrorsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _language_pb2.ErrorList
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ...) -> None: ...
    MODULE_ERRORS_FIELD_NUMBER: _ClassVar[int]
    module_errors: _containers.MessageMap[str, _language_pb2.ErrorList]
    def __init__(self, module_errors: _Optional[_Mapping[str, _language_pb2.ErrorList]] = ...) -> None: ...

class ModuleAdded(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class ModuleRemoved(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class ModuleBuildWaiting(_message.Message):
    __slots__ = ("config",)
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config: _language_pb2.ModuleConfig
    def __init__(self, config: _Optional[_Union[_language_pb2.ModuleConfig, _Mapping]] = ...) -> None: ...

class ModuleBuildStarted(_message.Message):
    __slots__ = ("config", "is_auto_rebuild")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    IS_AUTO_REBUILD_FIELD_NUMBER: _ClassVar[int]
    config: _language_pb2.ModuleConfig
    is_auto_rebuild: bool
    def __init__(self, config: _Optional[_Union[_language_pb2.ModuleConfig, _Mapping]] = ..., is_auto_rebuild: bool = ...) -> None: ...

class ModuleBuildFailed(_message.Message):
    __slots__ = ("config", "errors", "is_auto_rebuild")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    IS_AUTO_REBUILD_FIELD_NUMBER: _ClassVar[int]
    config: _language_pb2.ModuleConfig
    errors: _language_pb2.ErrorList
    is_auto_rebuild: bool
    def __init__(self, config: _Optional[_Union[_language_pb2.ModuleConfig, _Mapping]] = ..., errors: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ..., is_auto_rebuild: bool = ...) -> None: ...

class ModuleBuildSuccess(_message.Message):
    __slots__ = ("config", "is_auto_rebuild")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    IS_AUTO_REBUILD_FIELD_NUMBER: _ClassVar[int]
    config: _language_pb2.ModuleConfig
    is_auto_rebuild: bool
    def __init__(self, config: _Optional[_Union[_language_pb2.ModuleConfig, _Mapping]] = ..., is_auto_rebuild: bool = ...) -> None: ...

class ModuleDeployStarted(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class ModuleDeployFailed(_message.Message):
    __slots__ = ("module", "errors")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    module: str
    errors: _language_pb2.ErrorList
    def __init__(self, module: _Optional[str] = ..., errors: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ...) -> None: ...

class ModuleDeploySuccess(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class EngineEvent(_message.Message):
    __slots__ = ("engine_started", "engine_ended", "module_added", "module_removed", "module_build_waiting", "module_build_started", "module_build_failed", "module_build_success", "module_deploy_started", "module_deploy_failed", "module_deploy_success")
    ENGINE_STARTED_FIELD_NUMBER: _ClassVar[int]
    ENGINE_ENDED_FIELD_NUMBER: _ClassVar[int]
    MODULE_ADDED_FIELD_NUMBER: _ClassVar[int]
    MODULE_REMOVED_FIELD_NUMBER: _ClassVar[int]
    MODULE_BUILD_WAITING_FIELD_NUMBER: _ClassVar[int]
    MODULE_BUILD_STARTED_FIELD_NUMBER: _ClassVar[int]
    MODULE_BUILD_FAILED_FIELD_NUMBER: _ClassVar[int]
    MODULE_BUILD_SUCCESS_FIELD_NUMBER: _ClassVar[int]
    MODULE_DEPLOY_STARTED_FIELD_NUMBER: _ClassVar[int]
    MODULE_DEPLOY_FAILED_FIELD_NUMBER: _ClassVar[int]
    MODULE_DEPLOY_SUCCESS_FIELD_NUMBER: _ClassVar[int]
    engine_started: EngineStarted
    engine_ended: EngineEnded
    module_added: ModuleAdded
    module_removed: ModuleRemoved
    module_build_waiting: ModuleBuildWaiting
    module_build_started: ModuleBuildStarted
    module_build_failed: ModuleBuildFailed
    module_build_success: ModuleBuildSuccess
    module_deploy_started: ModuleDeployStarted
    module_deploy_failed: ModuleDeployFailed
    module_deploy_success: ModuleDeploySuccess
    def __init__(self, engine_started: _Optional[_Union[EngineStarted, _Mapping]] = ..., engine_ended: _Optional[_Union[EngineEnded, _Mapping]] = ..., module_added: _Optional[_Union[ModuleAdded, _Mapping]] = ..., module_removed: _Optional[_Union[ModuleRemoved, _Mapping]] = ..., module_build_waiting: _Optional[_Union[ModuleBuildWaiting, _Mapping]] = ..., module_build_started: _Optional[_Union[ModuleBuildStarted, _Mapping]] = ..., module_build_failed: _Optional[_Union[ModuleBuildFailed, _Mapping]] = ..., module_build_success: _Optional[_Union[ModuleBuildSuccess, _Mapping]] = ..., module_deploy_started: _Optional[_Union[ModuleDeployStarted, _Mapping]] = ..., module_deploy_failed: _Optional[_Union[ModuleDeployFailed, _Mapping]] = ..., module_deploy_success: _Optional[_Union[ModuleDeploySuccess, _Mapping]] = ...) -> None: ...

class StreamEngineEventsRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StreamEngineEventsResponse(_message.Message):
    __slots__ = ("event",)
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: EngineEvent
    def __init__(self, event: _Optional[_Union[EngineEvent, _Mapping]] = ...) -> None: ...
