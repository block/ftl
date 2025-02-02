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
    __slots__ = ()
    def __init__(self) -> None: ...

class PullSchemaResponse(_message.Message):
    __slots__ = ("changeset_created", "changeset_failed", "changeset_committed", "deployment_created", "deployment_updated", "deployment_removed", "more")
    class ChangesetCreated(_message.Message):
        __slots__ = ("changeset",)
        CHANGESET_FIELD_NUMBER: _ClassVar[int]
        changeset: _schema_pb2.Changeset
        def __init__(self, changeset: _Optional[_Union[_schema_pb2.Changeset, _Mapping]] = ...) -> None: ...
    class ChangesetFailed(_message.Message):
        __slots__ = ("key", "error")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ERROR_FIELD_NUMBER: _ClassVar[int]
        key: str
        error: str
        def __init__(self, key: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...
    class ChangesetCommitted(_message.Message):
        __slots__ = ("key",)
        KEY_FIELD_NUMBER: _ClassVar[int]
        key: str
        def __init__(self, key: _Optional[str] = ...) -> None: ...
    class DeploymentCreated(_message.Message):
        __slots__ = ("changeset", "schema")
        CHANGESET_FIELD_NUMBER: _ClassVar[int]
        SCHEMA_FIELD_NUMBER: _ClassVar[int]
        changeset: str
        schema: _schema_pb2.Module
        def __init__(self, changeset: _Optional[str] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
    class DeploymentUpdated(_message.Message):
        __slots__ = ("changeset", "schema")
        CHANGESET_FIELD_NUMBER: _ClassVar[int]
        SCHEMA_FIELD_NUMBER: _ClassVar[int]
        changeset: str
        schema: _schema_pb2.Module
        def __init__(self, changeset: _Optional[str] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
    class DeploymentRemoved(_message.Message):
        __slots__ = ("key", "module_name", "module_removed")
        KEY_FIELD_NUMBER: _ClassVar[int]
        MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
        MODULE_REMOVED_FIELD_NUMBER: _ClassVar[int]
        key: str
        module_name: str
        module_removed: bool
        def __init__(self, key: _Optional[str] = ..., module_name: _Optional[str] = ..., module_removed: bool = ...) -> None: ...
    CHANGESET_CREATED_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_FAILED_FIELD_NUMBER: _ClassVar[int]
    CHANGESET_COMMITTED_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_CREATED_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_UPDATED_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_REMOVED_FIELD_NUMBER: _ClassVar[int]
    MORE_FIELD_NUMBER: _ClassVar[int]
    changeset_created: PullSchemaResponse.ChangesetCreated
    changeset_failed: PullSchemaResponse.ChangesetFailed
    changeset_committed: PullSchemaResponse.ChangesetCommitted
    deployment_created: PullSchemaResponse.DeploymentCreated
    deployment_updated: PullSchemaResponse.DeploymentUpdated
    deployment_removed: PullSchemaResponse.DeploymentRemoved
    more: bool
    def __init__(self, changeset_created: _Optional[_Union[PullSchemaResponse.ChangesetCreated, _Mapping]] = ..., changeset_failed: _Optional[_Union[PullSchemaResponse.ChangesetFailed, _Mapping]] = ..., changeset_committed: _Optional[_Union[PullSchemaResponse.ChangesetCommitted, _Mapping]] = ..., deployment_created: _Optional[_Union[PullSchemaResponse.DeploymentCreated, _Mapping]] = ..., deployment_updated: _Optional[_Union[PullSchemaResponse.DeploymentUpdated, _Mapping]] = ..., deployment_removed: _Optional[_Union[PullSchemaResponse.DeploymentRemoved, _Mapping]] = ..., more: bool = ...) -> None: ...

class UpdateDeploymentRuntimeRequest(_message.Message):
    __slots__ = ("event",)
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _schema_pb2.ModuleRuntimeEvent
    def __init__(self, event: _Optional[_Union[_schema_pb2.ModuleRuntimeEvent, _Mapping]] = ...) -> None: ...

class UpdateDeploymentRuntimeResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UpdateSchemaRequest(_message.Message):
    __slots__ = ("event",)
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: _schema_pb2.Event
    def __init__(self, event: _Optional[_Union[_schema_pb2.Event, _Mapping]] = ...) -> None: ...

class UpdateSchemaResponse(_message.Message):
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
    __slots__ = ()
    def __init__(self) -> None: ...

class FailChangesetRequest(_message.Message):
    __slots__ = ("changeset", "error")
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    changeset: str
    error: str
    def __init__(self, changeset: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class FailChangesetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
