from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ProvisionResponseStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PROVISION_RESPONSE_STATUS_UNSPECIFIED: _ClassVar[ProvisionResponseStatus]
    PROVISION_RESPONSE_STATUS_SUBMITTED: _ClassVar[ProvisionResponseStatus]
PROVISION_RESPONSE_STATUS_UNSPECIFIED: ProvisionResponseStatus
PROVISION_RESPONSE_STATUS_SUBMITTED: ProvisionResponseStatus

class ProvisionRequest(_message.Message):
    __slots__ = ("ftl_cluster_id", "desired_module", "previous_module", "changeset", "kinds")
    FTL_CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    DESIRED_MODULE_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_MODULE_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    KINDS_FIELD_NUMBER: _ClassVar[int]
    ftl_cluster_id: str
    desired_module: _schema_pb2.Module
    previous_module: _schema_pb2.Module
    changeset: str
    kinds: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, ftl_cluster_id: _Optional[str] = ..., desired_module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., previous_module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., changeset: _Optional[str] = ..., kinds: _Optional[_Iterable[str]] = ...) -> None: ...

class ProvisionResponse(_message.Message):
    __slots__ = ("provisioning_token", "status")
    PROVISIONING_TOKEN_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    provisioning_token: str
    status: ProvisionResponseStatus
    def __init__(self, provisioning_token: _Optional[str] = ..., status: _Optional[_Union[ProvisionResponseStatus, str]] = ...) -> None: ...

class DeProvisionRequest(_message.Message):
    __slots__ = ("ftl_cluster_id", "module", "replacement_module", "changeset", "kinds")
    FTL_CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    REPLACEMENT_MODULE_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    KINDS_FIELD_NUMBER: _ClassVar[int]
    ftl_cluster_id: str
    module: _schema_pb2.Module
    replacement_module: _schema_pb2.Module
    changeset: str
    kinds: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, ftl_cluster_id: _Optional[str] = ..., module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., replacement_module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., changeset: _Optional[str] = ..., kinds: _Optional[_Iterable[str]] = ...) -> None: ...

class DeProvisionResponse(_message.Message):
    __slots__ = ("provisioning_token", "status")
    PROVISIONING_TOKEN_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    provisioning_token: str
    status: ProvisionResponseStatus
    def __init__(self, provisioning_token: _Optional[str] = ..., status: _Optional[_Union[ProvisionResponseStatus, str]] = ...) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ("provisioning_token", "desired_module")
    PROVISIONING_TOKEN_FIELD_NUMBER: _ClassVar[int]
    DESIRED_MODULE_FIELD_NUMBER: _ClassVar[int]
    provisioning_token: str
    desired_module: _schema_pb2.Module
    def __init__(self, provisioning_token: _Optional[str] = ..., desired_module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("running", "success")
    class ProvisioningRunning(_message.Message):
        __slots__ = ()
        def __init__(self) -> None: ...
    class ProvisioningFailed(_message.Message):
        __slots__ = ("error_message",)
        ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
        error_message: str
        def __init__(self, error_message: _Optional[str] = ...) -> None: ...
    class ProvisioningSuccess(_message.Message):
        __slots__ = ("outputs",)
        OUTPUTS_FIELD_NUMBER: _ClassVar[int]
        outputs: _containers.RepeatedCompositeFieldContainer[_schema_pb2.RuntimeElement]
        def __init__(self, outputs: _Optional[_Iterable[_Union[_schema_pb2.RuntimeElement, _Mapping]]] = ...) -> None: ...
    RUNNING_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    running: StatusResponse.ProvisioningRunning
    success: StatusResponse.ProvisioningSuccess
    def __init__(self, running: _Optional[_Union[StatusResponse.ProvisioningRunning, _Mapping]] = ..., success: _Optional[_Union[StatusResponse.ProvisioningSuccess, _Mapping]] = ...) -> None: ...
