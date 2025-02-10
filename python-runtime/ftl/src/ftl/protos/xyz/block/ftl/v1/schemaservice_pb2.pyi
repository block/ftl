from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetSchemaRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetSchemaResponse(_message.Message):
    __slots__ = ("schema", "changesets")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    CHANGESETS_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Schema
    changesets: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Changeset]
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Schema, _Mapping]] = ..., changesets: _Optional[_Iterable[_Union[_schema_pb2.Changeset, _Mapping]]] = ...) -> None: ...

class PullSchemaRequest(_message.Message):
    __slots__ = ("subscription_id",)
    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    subscription_id: str
    def __init__(self, subscription_id: _Optional[str] = ...) -> None: ...

class PullSchemaResponse(_message.Message):
    __slots__ = ("event",)
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _schema_pb2.Notification
    def __init__(self, event: _Optional[_Union[_schema_pb2.Notification, _Mapping]] = ...) -> None: ...

class UpdateDeploymentRuntimeRequest(_message.Message):
    __slots__ = ("changeset", "update")
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    UPDATE_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    update: _schema_pb2.RuntimeElement
    def __init__(self, changeset: _Optional[str] = ..., update: _Optional[_Union[_schema_pb2.RuntimeElement, _Mapping]] = ...) -> None: ...

class UpdateDeploymentRuntimeResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetDeploymentsRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetDeploymentsResponse(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _containers.RepeatedCompositeFieldContainer[DeployedSchema]
    def __init__(self, schema: _Optional[_Iterable[_Union[DeployedSchema, _Mapping]]] = ...) -> None: ...

class CreateChangesetRequest(_message.Message):
    __slots__ = ("modules", "removed_deployments")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    REMOVED_DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Module]
    removed_deployments: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, modules: _Optional[_Iterable[_Union[_schema_pb2.Module, _Mapping]]] = ..., removed_deployments: _Optional[_Iterable[str]] = ...) -> None: ...

class CreateChangesetResponse(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class DeployedSchema(_message.Message):
    __slots__ = ("deployment_key", "schema", "is_active")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    IS_ACTIVE_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    schema: _schema_pb2.Module
    is_active: bool
    def __init__(self, deployment_key: _Optional[str] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., is_active: bool = ...) -> None: ...

class PrepareChangesetRequest(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class PrepareChangesetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CommitChangesetRequest(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class CommitChangesetResponse(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: _schema_pb2.Changeset
    def __init__(self, changeset: _Optional[_Union[_schema_pb2.Changeset, _Mapping]] = ...) -> None: ...

class DrainChangesetRequest(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class DrainChangesetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class FinalizeChangesetRequest(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class FinalizeChangesetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class FailChangesetRequest(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    def __init__(self, changeset: _Optional[str] = ...) -> None: ...

class FailChangesetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class RollbackChangesetRequest(_message.Message):
    __slots__ = ("changeset", "error")
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    error: str
    def __init__(self, changeset: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class RollbackChangesetResponse(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: _schema_pb2.Changeset
    def __init__(self, changeset: _Optional[_Union[_schema_pb2.Changeset, _Mapping]] = ...) -> None: ...

class GetDeploymentRequest(_message.Message):
    __slots__ = ("deployment_key",)
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...

class GetDeploymentResponse(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Module
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
