from xyz.block.ftl.language.v1 import language_pb2 as _language_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ReloadRequest(_message.Message):
    __slots__ = ("force",)
    FORCE_FIELD_NUMBER: _ClassVar[int]
    force: bool
    def __init__(self, force: bool = ...) -> None: ...

class ReloadResponse(_message.Message):
    __slots__ = ("reload_not_required", "reload_success", "reload_failed")
    RELOAD_NOT_REQUIRED_FIELD_NUMBER: _ClassVar[int]
    RELOAD_SUCCESS_FIELD_NUMBER: _ClassVar[int]
    RELOAD_FAILED_FIELD_NUMBER: _ClassVar[int]
    reload_not_required: ReloadNotRequired
    reload_success: ReloadSuccess
    reload_failed: ReloadFailed
    def __init__(self, reload_not_required: _Optional[_Union[ReloadNotRequired, _Mapping]] = ..., reload_success: _Optional[_Union[ReloadSuccess, _Mapping]] = ..., reload_failed: _Optional[_Union[ReloadFailed, _Mapping]] = ...) -> None: ...

class WatchRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class WatchResponse(_message.Message):
    __slots__ = ("reload_success", "reload_failed")
    RELOAD_SUCCESS_FIELD_NUMBER: _ClassVar[int]
    RELOAD_FAILED_FIELD_NUMBER: _ClassVar[int]
    reload_success: ReloadSuccess
    reload_failed: ReloadFailed
    def __init__(self, reload_success: _Optional[_Union[ReloadSuccess, _Mapping]] = ..., reload_failed: _Optional[_Union[ReloadFailed, _Mapping]] = ...) -> None: ...

class ReloadNotRequired(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ReloadSuccess(_message.Message):
    __slots__ = ("module", "errors")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    module: _schema_pb2.Module
    errors: _language_pb2.ErrorList
    def __init__(self, module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., errors: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ...) -> None: ...

class ReloadFailed(_message.Message):
    __slots__ = ("errors",)
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    errors: _language_pb2.ErrorList
    def __init__(self, errors: _Optional[_Union[_language_pb2.ErrorList, _Mapping]] = ...) -> None: ...
