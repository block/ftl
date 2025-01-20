from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TransactionStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRANSACTION_STATUS_UNSPECIFIED: _ClassVar[TransactionStatus]
    TRANSACTION_STATUS_SUCCESS: _ClassVar[TransactionStatus]
    TRANSACTION_STATUS_FAILED: _ClassVar[TransactionStatus]

class CommandType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    COMMAND_TYPE_UNSPECIFIED: _ClassVar[CommandType]
    COMMAND_TYPE_EXEC: _ClassVar[CommandType]
    COMMAND_TYPE_ONE: _ClassVar[CommandType]
    COMMAND_TYPE_MANY: _ClassVar[CommandType]
TRANSACTION_STATUS_UNSPECIFIED: TransactionStatus
TRANSACTION_STATUS_SUCCESS: TransactionStatus
TRANSACTION_STATUS_FAILED: TransactionStatus
COMMAND_TYPE_UNSPECIFIED: CommandType
COMMAND_TYPE_EXEC: CommandType
COMMAND_TYPE_ONE: CommandType
COMMAND_TYPE_MANY: CommandType

class BeginTransactionRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class BeginTransactionResponse(_message.Message):
    __slots__ = ("transaction_id", "status")
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    transaction_id: str
    status: TransactionStatus
    def __init__(self, transaction_id: _Optional[str] = ..., status: _Optional[_Union[TransactionStatus, str]] = ...) -> None: ...

class CommitTransactionRequest(_message.Message):
    __slots__ = ("transaction_id",)
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    transaction_id: str
    def __init__(self, transaction_id: _Optional[str] = ...) -> None: ...

class CommitTransactionResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: TransactionStatus
    def __init__(self, status: _Optional[_Union[TransactionStatus, str]] = ...) -> None: ...

class RollbackTransactionRequest(_message.Message):
    __slots__ = ("transaction_id",)
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    transaction_id: str
    def __init__(self, transaction_id: _Optional[str] = ...) -> None: ...

class RollbackTransactionResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: TransactionStatus
    def __init__(self, status: _Optional[_Union[TransactionStatus, str]] = ...) -> None: ...

class SQLValue(_message.Message):
    __slots__ = ("string_value", "int_value", "float_value", "bool_value", "bytes_value", "timestamp_value", "null_value")
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    INT_VALUE_FIELD_NUMBER: _ClassVar[int]
    FLOAT_VALUE_FIELD_NUMBER: _ClassVar[int]
    BOOL_VALUE_FIELD_NUMBER: _ClassVar[int]
    BYTES_VALUE_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_VALUE_FIELD_NUMBER: _ClassVar[int]
    NULL_VALUE_FIELD_NUMBER: _ClassVar[int]
    string_value: str
    int_value: int
    float_value: float
    bool_value: bool
    bytes_value: bytes
    timestamp_value: _timestamp_pb2.Timestamp
    null_value: bool
    def __init__(self, string_value: _Optional[str] = ..., int_value: _Optional[int] = ..., float_value: _Optional[float] = ..., bool_value: bool = ..., bytes_value: _Optional[bytes] = ..., timestamp_value: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., null_value: bool = ...) -> None: ...

class ExecuteQueryRequest(_message.Message):
    __slots__ = ("raw_sql", "command_type", "parameters", "result_columns", "transaction_id", "batch_size")
    RAW_SQL_FIELD_NUMBER: _ClassVar[int]
    COMMAND_TYPE_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    RESULT_COLUMNS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTION_ID_FIELD_NUMBER: _ClassVar[int]
    BATCH_SIZE_FIELD_NUMBER: _ClassVar[int]
    raw_sql: str
    command_type: CommandType
    parameters: _containers.RepeatedCompositeFieldContainer[SQLValue]
    result_columns: _containers.RepeatedScalarFieldContainer[str]
    transaction_id: str
    batch_size: int
    def __init__(self, raw_sql: _Optional[str] = ..., command_type: _Optional[_Union[CommandType, str]] = ..., parameters: _Optional[_Iterable[_Union[SQLValue, _Mapping]]] = ..., result_columns: _Optional[_Iterable[str]] = ..., transaction_id: _Optional[str] = ..., batch_size: _Optional[int] = ...) -> None: ...

class ExecuteQueryResponse(_message.Message):
    __slots__ = ("exec_result", "row_results")
    EXEC_RESULT_FIELD_NUMBER: _ClassVar[int]
    ROW_RESULTS_FIELD_NUMBER: _ClassVar[int]
    exec_result: ExecResult
    row_results: RowResults
    def __init__(self, exec_result: _Optional[_Union[ExecResult, _Mapping]] = ..., row_results: _Optional[_Union[RowResults, _Mapping]] = ...) -> None: ...

class ExecResult(_message.Message):
    __slots__ = ("rows_affected", "last_insert_id")
    ROWS_AFFECTED_FIELD_NUMBER: _ClassVar[int]
    LAST_INSERT_ID_FIELD_NUMBER: _ClassVar[int]
    rows_affected: int
    last_insert_id: int
    def __init__(self, rows_affected: _Optional[int] = ..., last_insert_id: _Optional[int] = ...) -> None: ...

class RowResults(_message.Message):
    __slots__ = ("rows", "has_more")
    class RowsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: SQLValue
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[SQLValue, _Mapping]] = ...) -> None: ...
    ROWS_FIELD_NUMBER: _ClassVar[int]
    HAS_MORE_FIELD_NUMBER: _ClassVar[int]
    rows: _containers.MessageMap[str, SQLValue]
    has_more: bool
    def __init__(self, rows: _Optional[_Mapping[str, SQLValue]] = ..., has_more: bool = ...) -> None: ...
