from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class AddMemberRequest(_message.Message):
    __slots__ = ("address", "replica_id", "shard_ids")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    REPLICA_ID_FIELD_NUMBER: _ClassVar[int]
    SHARD_IDS_FIELD_NUMBER: _ClassVar[int]
    address: str
    replica_id: int
    shard_ids: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, address: _Optional[str] = ..., replica_id: _Optional[int] = ..., shard_ids: _Optional[_Iterable[int]] = ...) -> None: ...

class AddMemberResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
