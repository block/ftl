from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import schemaservice_pb2 as _schemaservice_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ConfigProvider(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CONFIG_PROVIDER_UNSPECIFIED: _ClassVar[ConfigProvider]
    CONFIG_PROVIDER_INLINE: _ClassVar[ConfigProvider]
    CONFIG_PROVIDER_ENVAR: _ClassVar[ConfigProvider]

class SecretProvider(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SECRET_PROVIDER_UNSPECIFIED: _ClassVar[SecretProvider]
    SECRET_PROVIDER_INLINE: _ClassVar[SecretProvider]
    SECRET_PROVIDER_ENVAR: _ClassVar[SecretProvider]
    SECRET_PROVIDER_KEYCHAIN: _ClassVar[SecretProvider]
    SECRET_PROVIDER_OP: _ClassVar[SecretProvider]
    SECRET_PROVIDER_ASM: _ClassVar[SecretProvider]

class SubscriptionOffset(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SUBSCRIPTION_OFFSET_UNSPECIFIED: _ClassVar[SubscriptionOffset]
    SUBSCRIPTION_OFFSET_EARLIEST: _ClassVar[SubscriptionOffset]
    SUBSCRIPTION_OFFSET_LATEST: _ClassVar[SubscriptionOffset]
CONFIG_PROVIDER_UNSPECIFIED: ConfigProvider
CONFIG_PROVIDER_INLINE: ConfigProvider
CONFIG_PROVIDER_ENVAR: ConfigProvider
SECRET_PROVIDER_UNSPECIFIED: SecretProvider
SECRET_PROVIDER_INLINE: SecretProvider
SECRET_PROVIDER_ENVAR: SecretProvider
SECRET_PROVIDER_KEYCHAIN: SecretProvider
SECRET_PROVIDER_OP: SecretProvider
SECRET_PROVIDER_ASM: SecretProvider
SUBSCRIPTION_OFFSET_UNSPECIFIED: SubscriptionOffset
SUBSCRIPTION_OFFSET_EARLIEST: SubscriptionOffset
SUBSCRIPTION_OFFSET_LATEST: SubscriptionOffset

class ConfigRef(_message.Message):
    __slots__ = ("module", "name")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    module: str
    name: str
    def __init__(self, module: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ConfigListRequest(_message.Message):
    __slots__ = ("module", "include_values", "provider")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_VALUES_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    module: str
    include_values: bool
    provider: ConfigProvider
    def __init__(self, module: _Optional[str] = ..., include_values: bool = ..., provider: _Optional[_Union[ConfigProvider, str]] = ...) -> None: ...

class ConfigListResponse(_message.Message):
    __slots__ = ("configs",)
    class Config(_message.Message):
        __slots__ = ("ref_path", "value")
        REF_PATH_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        ref_path: str
        value: bytes
        def __init__(self, ref_path: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    configs: _containers.RepeatedCompositeFieldContainer[ConfigListResponse.Config]
    def __init__(self, configs: _Optional[_Iterable[_Union[ConfigListResponse.Config, _Mapping]]] = ...) -> None: ...

class ConfigGetRequest(_message.Message):
    __slots__ = ("ref",)
    REF_FIELD_NUMBER: _ClassVar[int]
    ref: ConfigRef
    def __init__(self, ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class ConfigGetResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class ConfigSetRequest(_message.Message):
    __slots__ = ("provider", "ref", "value")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    provider: ConfigProvider
    ref: ConfigRef
    value: bytes
    def __init__(self, provider: _Optional[_Union[ConfigProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ..., value: _Optional[bytes] = ...) -> None: ...

class ConfigSetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ConfigUnsetRequest(_message.Message):
    __slots__ = ("provider", "ref")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    provider: ConfigProvider
    ref: ConfigRef
    def __init__(self, provider: _Optional[_Union[ConfigProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class ConfigUnsetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SecretsListRequest(_message.Message):
    __slots__ = ("module", "include_values", "provider")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_VALUES_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    module: str
    include_values: bool
    provider: SecretProvider
    def __init__(self, module: _Optional[str] = ..., include_values: bool = ..., provider: _Optional[_Union[SecretProvider, str]] = ...) -> None: ...

class SecretsListResponse(_message.Message):
    __slots__ = ("secrets",)
    class Secret(_message.Message):
        __slots__ = ("ref_path", "value")
        REF_PATH_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        ref_path: str
        value: bytes
        def __init__(self, ref_path: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    secrets: _containers.RepeatedCompositeFieldContainer[SecretsListResponse.Secret]
    def __init__(self, secrets: _Optional[_Iterable[_Union[SecretsListResponse.Secret, _Mapping]]] = ...) -> None: ...

class SecretGetRequest(_message.Message):
    __slots__ = ("ref",)
    REF_FIELD_NUMBER: _ClassVar[int]
    ref: ConfigRef
    def __init__(self, ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class SecretGetResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SecretSetRequest(_message.Message):
    __slots__ = ("provider", "ref", "value")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    provider: SecretProvider
    ref: ConfigRef
    value: bytes
    def __init__(self, provider: _Optional[_Union[SecretProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ..., value: _Optional[bytes] = ...) -> None: ...

class SecretSetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SecretUnsetRequest(_message.Message):
    __slots__ = ("provider", "ref")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    provider: SecretProvider
    ref: ConfigRef
    def __init__(self, provider: _Optional[_Union[SecretProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class SecretUnsetResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class MapConfigsForModuleRequest(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class MapConfigsForModuleResponse(_message.Message):
    __slots__ = ("values",)
    class ValuesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, bytes]
    def __init__(self, values: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class MapSecretsForModuleRequest(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class MapSecretsForModuleResponse(_message.Message):
    __slots__ = ("values",)
    class ValuesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, bytes]
    def __init__(self, values: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class ResetSubscriptionRequest(_message.Message):
    __slots__ = ("subscription", "offset")
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    subscription: _schema_pb2.Ref
    offset: SubscriptionOffset
    def __init__(self, subscription: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., offset: _Optional[_Union[SubscriptionOffset, str]] = ...) -> None: ...

class ResetSubscriptionResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ApplyChangesetRequest(_message.Message):
    __slots__ = ("modules", "to_remove")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TO_REMOVE_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Module]
    to_remove: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, modules: _Optional[_Iterable[_Union[_schema_pb2.Module, _Mapping]]] = ..., to_remove: _Optional[_Iterable[str]] = ...) -> None: ...

class ApplyChangesetResponse(_message.Message):
    __slots__ = ("changeset",)
    CHANGESET_FIELD_NUMBER: _ClassVar[int]
    changeset: _schema_pb2.Changeset
    def __init__(self, changeset: _Optional[_Union[_schema_pb2.Changeset, _Mapping]] = ...) -> None: ...

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

class DeploymentArtefact(_message.Message):
    __slots__ = ("digest", "path", "executable")
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    EXECUTABLE_FIELD_NUMBER: _ClassVar[int]
    digest: bytes
    path: str
    executable: bool
    def __init__(self, digest: _Optional[bytes] = ..., path: _Optional[str] = ..., executable: bool = ...) -> None: ...

class UploadArtefactRequest(_message.Message):
    __slots__ = ("digest", "size", "chunk")
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    SIZE_FIELD_NUMBER: _ClassVar[int]
    CHUNK_FIELD_NUMBER: _ClassVar[int]
    digest: bytes
    size: int
    chunk: bytes
    def __init__(self, digest: _Optional[bytes] = ..., size: _Optional[int] = ..., chunk: _Optional[bytes] = ...) -> None: ...

class UploadArtefactResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ClusterInfoRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ClusterInfoResponse(_message.Message):
    __slots__ = ("os", "arch")
    OS_FIELD_NUMBER: _ClassVar[int]
    ARCH_FIELD_NUMBER: _ClassVar[int]
    os: str
    arch: str
    def __init__(self, os: _Optional[str] = ..., arch: _Optional[str] = ...) -> None: ...
