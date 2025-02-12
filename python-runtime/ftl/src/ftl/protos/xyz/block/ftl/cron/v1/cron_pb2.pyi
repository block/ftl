from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CronState(_message.Message):
    __slots__ = ("last_executions", "next_executions")
    class LastExecutionsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _timestamp_pb2.Timestamp
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    class NextExecutionsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _timestamp_pb2.Timestamp
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    LAST_EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    NEXT_EXECUTIONS_FIELD_NUMBER: _ClassVar[int]
    last_executions: _containers.MessageMap[str, _timestamp_pb2.Timestamp]
    next_executions: _containers.MessageMap[str, _timestamp_pb2.Timestamp]
    def __init__(self, last_executions: _Optional[_Mapping[str, _timestamp_pb2.Timestamp]] = ..., next_executions: _Optional[_Mapping[str, _timestamp_pb2.Timestamp]] = ...) -> None: ...
