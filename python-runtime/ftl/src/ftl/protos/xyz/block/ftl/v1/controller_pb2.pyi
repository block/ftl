from google.protobuf import struct_pb2 as _struct_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetArtefactDiffsRequest(_message.Message):
    __slots__ = ("client_digests",)
    CLIENT_DIGESTS_FIELD_NUMBER: _ClassVar[int]
    client_digests: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, client_digests: _Optional[_Iterable[str]] = ...) -> None: ...

class GetArtefactDiffsResponse(_message.Message):
    __slots__ = ("missing_digests", "client_artefacts")
    MISSING_DIGESTS_FIELD_NUMBER: _ClassVar[int]
    CLIENT_ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    missing_digests: _containers.RepeatedScalarFieldContainer[str]
    client_artefacts: _containers.RepeatedCompositeFieldContainer[DeploymentArtefact]
    def __init__(self, missing_digests: _Optional[_Iterable[str]] = ..., client_artefacts: _Optional[_Iterable[_Union[DeploymentArtefact, _Mapping]]] = ...) -> None: ...

class UploadArtefactRequest(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: bytes
    def __init__(self, content: _Optional[bytes] = ...) -> None: ...

class UploadArtefactResponse(_message.Message):
    __slots__ = ("digest",)
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    digest: bytes
    def __init__(self, digest: _Optional[bytes] = ...) -> None: ...

class DeploymentArtefact(_message.Message):
    __slots__ = ("digest", "path", "executable")
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    EXECUTABLE_FIELD_NUMBER: _ClassVar[int]
    digest: str
    path: str
    executable: bool
    def __init__(self, digest: _Optional[str] = ..., path: _Optional[str] = ..., executable: bool = ...) -> None: ...

class CreateDeploymentRequest(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Module
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...

class CreateDeploymentResponse(_message.Message):
    __slots__ = ("deployment_key", "active_deployment_key")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    active_deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ..., active_deployment_key: _Optional[str] = ...) -> None: ...

class GetDeploymentArtefactsRequest(_message.Message):
    __slots__ = ("deployment_key", "have_artefacts")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    HAVE_ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    have_artefacts: _containers.RepeatedCompositeFieldContainer[DeploymentArtefact]
    def __init__(self, deployment_key: _Optional[str] = ..., have_artefacts: _Optional[_Iterable[_Union[DeploymentArtefact, _Mapping]]] = ...) -> None: ...

class GetDeploymentArtefactsResponse(_message.Message):
    __slots__ = ("artefact", "chunk")
    ARTEFACT_FIELD_NUMBER: _ClassVar[int]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    artefact: DeploymentArtefact
    chunk: bytes
    def __init__(self, artefact: _Optional[_Union[DeploymentArtefact, _Mapping]] = ..., chunk: _Optional[bytes] = ...) -> None: ...

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

class RegisterRunnerRequest(_message.Message):
    __slots__ = ("key", "endpoint", "deployment", "labels")
    KEY_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    key: str
    endpoint: str
    deployment: str
    labels: _struct_pb2.Struct
    def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class RegisterRunnerResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UpdateDeployRequest(_message.Message):
    __slots__ = ("deployment_key", "min_replicas")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    min_replicas: int
    def __init__(self, deployment_key: _Optional[str] = ..., min_replicas: _Optional[int] = ...) -> None: ...

class UpdateDeployResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ReplaceDeployRequest(_message.Message):
    __slots__ = ("deployment_key", "min_replicas")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    min_replicas: int
    def __init__(self, deployment_key: _Optional[str] = ..., min_replicas: _Optional[int] = ...) -> None: ...

class ReplaceDeployResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("controllers", "runners", "deployments", "routes")
    class Controller(_message.Message):
        __slots__ = ("key", "endpoint", "version")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        VERSION_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        version: str
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...
    class Runner(_message.Message):
        __slots__ = ("key", "endpoint", "deployment", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        deployment: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Deployment(_message.Message):
        __slots__ = ("key", "language", "name", "min_replicas", "replicas", "labels", "schema")
        KEY_FIELD_NUMBER: _ClassVar[int]
        LANGUAGE_FIELD_NUMBER: _ClassVar[int]
        NAME_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        SCHEMA_FIELD_NUMBER: _ClassVar[int]
        key: str
        language: str
        name: str
        min_replicas: int
        replicas: int
        labels: _struct_pb2.Struct
        schema: _schema_pb2.Module
        def __init__(self, key: _Optional[str] = ..., language: _Optional[str] = ..., name: _Optional[str] = ..., min_replicas: _Optional[int] = ..., replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
    class Route(_message.Message):
        __slots__ = ("module", "deployment", "endpoint")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        module: str
        deployment: str
        endpoint: str
        def __init__(self, module: _Optional[str] = ..., deployment: _Optional[str] = ..., endpoint: _Optional[str] = ...) -> None: ...
    CONTROLLERS_FIELD_NUMBER: _ClassVar[int]
    RUNNERS_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
    ROUTES_FIELD_NUMBER: _ClassVar[int]
    controllers: _containers.RepeatedCompositeFieldContainer[StatusResponse.Controller]
    runners: _containers.RepeatedCompositeFieldContainer[StatusResponse.Runner]
    deployments: _containers.RepeatedCompositeFieldContainer[StatusResponse.Deployment]
    routes: _containers.RepeatedCompositeFieldContainer[StatusResponse.Route]
    def __init__(self, controllers: _Optional[_Iterable[_Union[StatusResponse.Controller, _Mapping]]] = ..., runners: _Optional[_Iterable[_Union[StatusResponse.Runner, _Mapping]]] = ..., deployments: _Optional[_Iterable[_Union[StatusResponse.Deployment, _Mapping]]] = ..., routes: _Optional[_Iterable[_Union[StatusResponse.Route, _Mapping]]] = ...) -> None: ...

class ProcessListRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ProcessListResponse(_message.Message):
    __slots__ = ("processes",)
    class ProcessRunner(_message.Message):
        __slots__ = ("key", "endpoint", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Process(_message.Message):
        __slots__ = ("deployment", "min_replicas", "labels", "runner")
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        RUNNER_FIELD_NUMBER: _ClassVar[int]
        deployment: str
        min_replicas: int
        labels: _struct_pb2.Struct
        runner: ProcessListResponse.ProcessRunner
        def __init__(self, deployment: _Optional[str] = ..., min_replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., runner: _Optional[_Union[ProcessListResponse.ProcessRunner, _Mapping]] = ...) -> None: ...
    PROCESSES_FIELD_NUMBER: _ClassVar[int]
    processes: _containers.RepeatedCompositeFieldContainer[ProcessListResponse.Process]
    def __init__(self, processes: _Optional[_Iterable[_Union[ProcessListResponse.Process, _Mapping]]] = ...) -> None: ...
