# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/admin/v1/admin.proto
# Protobuf Python Version: 5.29.3
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    3,
    '',
    'xyz/block/ftl/admin/v1/admin.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2
from xyz.block.ftl.v1 import schemaservice_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_schemaservice__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\"xyz/block/ftl/admin/v1/admin.proto\x12\x16xyz.block.ftl.admin.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\x1a$xyz/block/ftl/v1/schemaservice.proto\"G\n\tConfigRef\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\t\n\x07_module\"\xd0\x01\n\x11\x43onfigListRequest\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12*\n\x0einclude_values\x18\x02 \x01(\x08H\x01R\rincludeValues\x88\x01\x01\x12G\n\x08provider\x18\x03 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.ConfigProviderH\x02R\x08provider\x88\x01\x01\x42\t\n\x07_moduleB\x11\n\x0f_include_valuesB\x0b\n\t_provider\"\xab\x01\n\x12\x43onfigListResponse\x12K\n\x07\x63onfigs\x18\x01 \x03(\x0b\x32\x31.xyz.block.ftl.admin.v1.ConfigListResponse.ConfigR\x07\x63onfigs\x1aH\n\x06\x43onfig\x12\x19\n\x08ref_path\x18\x01 \x01(\tR\x07refPath\x12\x19\n\x05value\x18\x02 \x01(\x0cH\x00R\x05value\x88\x01\x01\x42\x08\n\x06_value\"G\n\x10\x43onfigGetRequest\x12\x33\n\x03ref\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03ref\")\n\x11\x43onfigGetResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"\xb3\x01\n\x10\x43onfigSetRequest\x12G\n\x08provider\x18\x01 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.ConfigProviderH\x00R\x08provider\x88\x01\x01\x12\x33\n\x03ref\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03ref\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\x0b\n\t_provider\"\x13\n\x11\x43onfigSetResponse\"\x9f\x01\n\x12\x43onfigUnsetRequest\x12G\n\x08provider\x18\x01 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.ConfigProviderH\x00R\x08provider\x88\x01\x01\x12\x33\n\x03ref\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03refB\x0b\n\t_provider\"\x15\n\x13\x43onfigUnsetResponse\"\xd1\x01\n\x12SecretsListRequest\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12*\n\x0einclude_values\x18\x02 \x01(\x08H\x01R\rincludeValues\x88\x01\x01\x12G\n\x08provider\x18\x03 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.SecretProviderH\x02R\x08provider\x88\x01\x01\x42\t\n\x07_moduleB\x11\n\x0f_include_valuesB\x0b\n\t_provider\"\xad\x01\n\x13SecretsListResponse\x12L\n\x07secrets\x18\x01 \x03(\x0b\x32\x32.xyz.block.ftl.admin.v1.SecretsListResponse.SecretR\x07secrets\x1aH\n\x06Secret\x12\x19\n\x08ref_path\x18\x01 \x01(\tR\x07refPath\x12\x19\n\x05value\x18\x02 \x01(\x0cH\x00R\x05value\x88\x01\x01\x42\x08\n\x06_value\"G\n\x10SecretGetRequest\x12\x33\n\x03ref\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03ref\")\n\x11SecretGetResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"\xb3\x01\n\x10SecretSetRequest\x12G\n\x08provider\x18\x01 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.SecretProviderH\x00R\x08provider\x88\x01\x01\x12\x33\n\x03ref\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03ref\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\x0b\n\t_provider\"\x13\n\x11SecretSetResponse\"\x9f\x01\n\x12SecretUnsetRequest\x12G\n\x08provider\x18\x01 \x01(\x0e\x32&.xyz.block.ftl.admin.v1.SecretProviderH\x00R\x08provider\x88\x01\x01\x12\x33\n\x03ref\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.admin.v1.ConfigRefR\x03refB\x0b\n\t_provider\"\x15\n\x13SecretUnsetResponse\"4\n\x1aMapConfigsForModuleRequest\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\xb1\x01\n\x1bMapConfigsForModuleResponse\x12W\n\x06values\x18\x01 \x03(\x0b\x32?.xyz.block.ftl.admin.v1.MapConfigsForModuleResponse.ValuesEntryR\x06values\x1a\x39\n\x0bValuesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"4\n\x1aMapSecretsForModuleRequest\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\xb1\x01\n\x1bMapSecretsForModuleResponse\x12W\n\x06values\x18\x01 \x03(\x0b\x32?.xyz.block.ftl.admin.v1.MapSecretsForModuleResponse.ValuesEntryR\x06values\x1a\x39\n\x0bValuesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"\xa0\x01\n\x18ResetSubscriptionRequest\x12@\n\x0csubscription\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x0csubscription\x12\x42\n\x06offset\x18\x02 \x01(\x0e\x32*.xyz.block.ftl.admin.v1.SubscriptionOffsetR\x06offset\"\x1b\n\x19ResetSubscriptionResponse\"o\n\x15\x41pplyChangesetRequest\x12\x39\n\x07modules\x18\x01 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\x12\x1b\n\tto_remove\x18\x02 \x03(\tR\x08toRemove\"Z\n\x16\x41pplyChangesetResponse\x12@\n\tchangeset\x18\x02 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\"@\n\x17GetArtefactDiffsRequest\x12%\n\x0e\x63lient_digests\x18\x01 \x03(\tR\rclientDigests\"\x9a\x01\n\x18GetArtefactDiffsResponse\x12\'\n\x0fmissing_digests\x18\x01 \x03(\tR\x0emissingDigests\x12U\n\x10\x63lient_artefacts\x18\x02 \x03(\x0b\x32*.xyz.block.ftl.admin.v1.DeploymentArtefactR\x0f\x63lientArtefacts\"\x99\x01\n\x1dGetDeploymentArtefactsRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12Q\n\x0ehave_artefacts\x18\x02 \x03(\x0b\x32*.xyz.block.ftl.admin.v1.DeploymentArtefactR\rhaveArtefacts\"~\n\x1eGetDeploymentArtefactsResponse\x12\x46\n\x08\x61rtefact\x18\x01 \x01(\x0b\x32*.xyz.block.ftl.admin.v1.DeploymentArtefactR\x08\x61rtefact\x12\x14\n\x05\x63hunk\x18\x02 \x01(\x0cR\x05\x63hunk\"`\n\x12\x44\x65ploymentArtefact\x12\x16\n\x06\x64igest\x18\x01 \x01(\x0cR\x06\x64igest\x12\x12\n\x04path\x18\x02 \x01(\tR\x04path\x12\x1e\n\nexecutable\x18\x03 \x01(\x08R\nexecutable\"Y\n\x15UploadArtefactRequest\x12\x16\n\x06\x64igest\x18\x01 \x01(\x0cR\x06\x64igest\x12\x12\n\x04size\x18\x02 \x01(\x03R\x04size\x12\x14\n\x05\x63hunk\x18\x03 \x01(\x0cR\x05\x63hunk\"\x18\n\x16UploadArtefactResponse\"\x14\n\x12\x43lusterInfoRequest\"9\n\x13\x43lusterInfoResponse\x12\x0e\n\x02os\x18\x01 \x01(\tR\x02os\x12\x12\n\x04\x61rch\x18\x02 \x01(\tR\x04\x61rch*h\n\x0e\x43onfigProvider\x12\x1f\n\x1b\x43ONFIG_PROVIDER_UNSPECIFIED\x10\x00\x12\x1a\n\x16\x43ONFIG_PROVIDER_INLINE\x10\x01\x12\x19\n\x15\x43ONFIG_PROVIDER_ENVAR\x10\x02*\xb7\x01\n\x0eSecretProvider\x12\x1f\n\x1bSECRET_PROVIDER_UNSPECIFIED\x10\x00\x12\x1a\n\x16SECRET_PROVIDER_INLINE\x10\x01\x12\x19\n\x15SECRET_PROVIDER_ENVAR\x10\x02\x12\x1c\n\x18SECRET_PROVIDER_KEYCHAIN\x10\x03\x12\x16\n\x12SECRET_PROVIDER_OP\x10\x04\x12\x17\n\x13SECRET_PROVIDER_ASM\x10\x05*{\n\x12SubscriptionOffset\x12#\n\x1fSUBSCRIPTION_OFFSET_UNSPECIFIED\x10\x00\x12 \n\x1cSUBSCRIPTION_OFFSET_EARLIEST\x10\x01\x12\x1e\n\x1aSUBSCRIPTION_OFFSET_LATEST\x10\x02\x32\xd5\x11\n\x0c\x41\x64minService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x63\n\nConfigList\x12).xyz.block.ftl.admin.v1.ConfigListRequest\x1a*.xyz.block.ftl.admin.v1.ConfigListResponse\x12`\n\tConfigGet\x12(.xyz.block.ftl.admin.v1.ConfigGetRequest\x1a).xyz.block.ftl.admin.v1.ConfigGetResponse\x12`\n\tConfigSet\x12(.xyz.block.ftl.admin.v1.ConfigSetRequest\x1a).xyz.block.ftl.admin.v1.ConfigSetResponse\x12\x66\n\x0b\x43onfigUnset\x12*.xyz.block.ftl.admin.v1.ConfigUnsetRequest\x1a+.xyz.block.ftl.admin.v1.ConfigUnsetResponse\x12\x66\n\x0bSecretsList\x12*.xyz.block.ftl.admin.v1.SecretsListRequest\x1a+.xyz.block.ftl.admin.v1.SecretsListResponse\x12`\n\tSecretGet\x12(.xyz.block.ftl.admin.v1.SecretGetRequest\x1a).xyz.block.ftl.admin.v1.SecretGetResponse\x12`\n\tSecretSet\x12(.xyz.block.ftl.admin.v1.SecretSetRequest\x1a).xyz.block.ftl.admin.v1.SecretSetResponse\x12\x66\n\x0bSecretUnset\x12*.xyz.block.ftl.admin.v1.SecretUnsetRequest\x1a+.xyz.block.ftl.admin.v1.SecretUnsetResponse\x12~\n\x13MapConfigsForModule\x12\x32.xyz.block.ftl.admin.v1.MapConfigsForModuleRequest\x1a\x33.xyz.block.ftl.admin.v1.MapConfigsForModuleResponse\x12~\n\x13MapSecretsForModule\x12\x32.xyz.block.ftl.admin.v1.MapSecretsForModuleRequest\x1a\x33.xyz.block.ftl.admin.v1.MapSecretsForModuleResponse\x12x\n\x11ResetSubscription\x12\x30.xyz.block.ftl.admin.v1.ResetSubscriptionRequest\x1a\x31.xyz.block.ftl.admin.v1.ResetSubscriptionResponse\x12q\n\x0e\x41pplyChangeset\x12-.xyz.block.ftl.admin.v1.ApplyChangesetRequest\x1a..xyz.block.ftl.admin.v1.ApplyChangesetResponse0\x01\x12Y\n\tGetSchema\x12\".xyz.block.ftl.v1.GetSchemaRequest\x1a#.xyz.block.ftl.v1.GetSchemaResponse\"\x03\x90\x02\x01\x12^\n\nPullSchema\x12#.xyz.block.ftl.v1.PullSchemaRequest\x1a$.xyz.block.ftl.v1.PullSchemaResponse\"\x03\x90\x02\x01\x30\x01\x12l\n\x11RollbackChangeset\x12*.xyz.block.ftl.v1.RollbackChangesetRequest\x1a+.xyz.block.ftl.v1.RollbackChangesetResponse\x12`\n\rFailChangeset\x12&.xyz.block.ftl.v1.FailChangesetRequest\x1a\'.xyz.block.ftl.v1.FailChangesetResponse\x12\x66\n\x0b\x43lusterInfo\x12*.xyz.block.ftl.admin.v1.ClusterInfoRequest\x1a+.xyz.block.ftl.admin.v1.ClusterInfoResponse\x12u\n\x10GetArtefactDiffs\x12/.xyz.block.ftl.admin.v1.GetArtefactDiffsRequest\x1a\x30.xyz.block.ftl.admin.v1.GetArtefactDiffsResponse\x12\x89\x01\n\x16GetDeploymentArtefacts\x12\x35.xyz.block.ftl.admin.v1.GetDeploymentArtefactsRequest\x1a\x36.xyz.block.ftl.admin.v1.GetDeploymentArtefactsResponse0\x01\x12q\n\x0eUploadArtefact\x12-.xyz.block.ftl.admin.v1.UploadArtefactRequest\x1a..xyz.block.ftl.admin.v1.UploadArtefactResponse(\x01\x42\x46P\x01ZBgithub.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1;adminpbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.admin.v1.admin_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZBgithub.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1;adminpb'
  _globals['_MAPCONFIGSFORMODULERESPONSE_VALUESENTRY']._loaded_options = None
  _globals['_MAPCONFIGSFORMODULERESPONSE_VALUESENTRY']._serialized_options = b'8\001'
  _globals['_MAPSECRETSFORMODULERESPONSE_VALUESENTRY']._loaded_options = None
  _globals['_MAPSECRETSFORMODULERESPONSE_VALUESENTRY']._serialized_options = b'8\001'
  _globals['_ADMINSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_ADMINSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ADMINSERVICE'].methods_by_name['GetSchema']._loaded_options = None
  _globals['_ADMINSERVICE'].methods_by_name['GetSchema']._serialized_options = b'\220\002\001'
  _globals['_ADMINSERVICE'].methods_by_name['PullSchema']._loaded_options = None
  _globals['_ADMINSERVICE'].methods_by_name['PullSchema']._serialized_options = b'\220\002\001'
  _globals['_CONFIGPROVIDER']._serialized_start=3688
  _globals['_CONFIGPROVIDER']._serialized_end=3792
  _globals['_SECRETPROVIDER']._serialized_start=3795
  _globals['_SECRETPROVIDER']._serialized_end=3978
  _globals['_SUBSCRIPTIONOFFSET']._serialized_start=3980
  _globals['_SUBSCRIPTIONOFFSET']._serialized_end=4103
  _globals['_CONFIGREF']._serialized_start=166
  _globals['_CONFIGREF']._serialized_end=237
  _globals['_CONFIGLISTREQUEST']._serialized_start=240
  _globals['_CONFIGLISTREQUEST']._serialized_end=448
  _globals['_CONFIGLISTRESPONSE']._serialized_start=451
  _globals['_CONFIGLISTRESPONSE']._serialized_end=622
  _globals['_CONFIGLISTRESPONSE_CONFIG']._serialized_start=550
  _globals['_CONFIGLISTRESPONSE_CONFIG']._serialized_end=622
  _globals['_CONFIGGETREQUEST']._serialized_start=624
  _globals['_CONFIGGETREQUEST']._serialized_end=695
  _globals['_CONFIGGETRESPONSE']._serialized_start=697
  _globals['_CONFIGGETRESPONSE']._serialized_end=738
  _globals['_CONFIGSETREQUEST']._serialized_start=741
  _globals['_CONFIGSETREQUEST']._serialized_end=920
  _globals['_CONFIGSETRESPONSE']._serialized_start=922
  _globals['_CONFIGSETRESPONSE']._serialized_end=941
  _globals['_CONFIGUNSETREQUEST']._serialized_start=944
  _globals['_CONFIGUNSETREQUEST']._serialized_end=1103
  _globals['_CONFIGUNSETRESPONSE']._serialized_start=1105
  _globals['_CONFIGUNSETRESPONSE']._serialized_end=1126
  _globals['_SECRETSLISTREQUEST']._serialized_start=1129
  _globals['_SECRETSLISTREQUEST']._serialized_end=1338
  _globals['_SECRETSLISTRESPONSE']._serialized_start=1341
  _globals['_SECRETSLISTRESPONSE']._serialized_end=1514
  _globals['_SECRETSLISTRESPONSE_SECRET']._serialized_start=1442
  _globals['_SECRETSLISTRESPONSE_SECRET']._serialized_end=1514
  _globals['_SECRETGETREQUEST']._serialized_start=1516
  _globals['_SECRETGETREQUEST']._serialized_end=1587
  _globals['_SECRETGETRESPONSE']._serialized_start=1589
  _globals['_SECRETGETRESPONSE']._serialized_end=1630
  _globals['_SECRETSETREQUEST']._serialized_start=1633
  _globals['_SECRETSETREQUEST']._serialized_end=1812
  _globals['_SECRETSETRESPONSE']._serialized_start=1814
  _globals['_SECRETSETRESPONSE']._serialized_end=1833
  _globals['_SECRETUNSETREQUEST']._serialized_start=1836
  _globals['_SECRETUNSETREQUEST']._serialized_end=1995
  _globals['_SECRETUNSETRESPONSE']._serialized_start=1997
  _globals['_SECRETUNSETRESPONSE']._serialized_end=2018
  _globals['_MAPCONFIGSFORMODULEREQUEST']._serialized_start=2020
  _globals['_MAPCONFIGSFORMODULEREQUEST']._serialized_end=2072
  _globals['_MAPCONFIGSFORMODULERESPONSE']._serialized_start=2075
  _globals['_MAPCONFIGSFORMODULERESPONSE']._serialized_end=2252
  _globals['_MAPCONFIGSFORMODULERESPONSE_VALUESENTRY']._serialized_start=2195
  _globals['_MAPCONFIGSFORMODULERESPONSE_VALUESENTRY']._serialized_end=2252
  _globals['_MAPSECRETSFORMODULEREQUEST']._serialized_start=2254
  _globals['_MAPSECRETSFORMODULEREQUEST']._serialized_end=2306
  _globals['_MAPSECRETSFORMODULERESPONSE']._serialized_start=2309
  _globals['_MAPSECRETSFORMODULERESPONSE']._serialized_end=2486
  _globals['_MAPSECRETSFORMODULERESPONSE_VALUESENTRY']._serialized_start=2195
  _globals['_MAPSECRETSFORMODULERESPONSE_VALUESENTRY']._serialized_end=2252
  _globals['_RESETSUBSCRIPTIONREQUEST']._serialized_start=2489
  _globals['_RESETSUBSCRIPTIONREQUEST']._serialized_end=2649
  _globals['_RESETSUBSCRIPTIONRESPONSE']._serialized_start=2651
  _globals['_RESETSUBSCRIPTIONRESPONSE']._serialized_end=2678
  _globals['_APPLYCHANGESETREQUEST']._serialized_start=2680
  _globals['_APPLYCHANGESETREQUEST']._serialized_end=2791
  _globals['_APPLYCHANGESETRESPONSE']._serialized_start=2793
  _globals['_APPLYCHANGESETRESPONSE']._serialized_end=2883
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_start=2885
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_end=2949
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_start=2952
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_end=3106
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_start=3109
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_end=3262
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_start=3264
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_end=3390
  _globals['_DEPLOYMENTARTEFACT']._serialized_start=3392
  _globals['_DEPLOYMENTARTEFACT']._serialized_end=3488
  _globals['_UPLOADARTEFACTREQUEST']._serialized_start=3490
  _globals['_UPLOADARTEFACTREQUEST']._serialized_end=3579
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_start=3581
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_end=3605
  _globals['_CLUSTERINFOREQUEST']._serialized_start=3607
  _globals['_CLUSTERINFOREQUEST']._serialized_end=3627
  _globals['_CLUSTERINFORESPONSE']._serialized_start=3629
  _globals['_CLUSTERINFORESPONSE']._serialized_end=3686
  _globals['_ADMINSERVICE']._serialized_start=4106
  _globals['_ADMINSERVICE']._serialized_end=6367
# @@protoc_insertion_point(module_scope)
