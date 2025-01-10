from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ResetOffsetsOfSubscriptionRequest(_message.Message):
    __slots__ = ("subscription",)
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    subscription: _schema_pb2.Ref
    def __init__(self, subscription: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ...) -> None: ...

class ResetOffsetsOfSubscriptionResponse(_message.Message):
    __slots__ = ("partitions",)
    PARTITIONS_FIELD_NUMBER: _ClassVar[int]
    partitions: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, partitions: _Optional[_Iterable[int]] = ...) -> None: ...
