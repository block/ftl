# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/provisioner/v1beta1/plugin.proto
# Protobuf Python Version: 5.29.4
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
    4,
    '',
    'xyz/block/ftl/provisioner/v1beta1/plugin.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n.xyz/block/ftl/provisioner/v1beta1/plugin.proto\x12!xyz.block.ftl.provisioner.v1beta1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\xfe\x01\n\x10ProvisionRequest\x12$\n\x0e\x66tl_cluster_id\x18\x01 \x01(\tR\x0c\x66tlClusterId\x12\x46\n\x0e\x64\x65sired_module\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\rdesiredModule\x12H\n\x0fprevious_module\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x0epreviousModule\x12\x1c\n\tchangeset\x18\x04 \x01(\tR\tchangeset\x12\x14\n\x05kinds\x18\x05 \x03(\tR\x05kinds\"\x97\x02\n\x11ProvisionResponse\x12-\n\x12provisioning_token\x18\x01 \x01(\tR\x11provisioningToken\x12\x64\n\x06status\x18\x02 \x01(\x0e\x32L.xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatusR\x06status\"m\n\x17ProvisionResponseStatus\x12)\n%PROVISION_RESPONSE_STATUS_UNSPECIFIED\x10\x00\x12\'\n#PROVISION_RESPONSE_STATUS_SUBMITTED\x10\x01\"\x86\x01\n\rStatusRequest\x12-\n\x12provisioning_token\x18\x01 \x01(\tR\x11provisioningToken\x12\x46\n\x0e\x64\x65sired_module\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\rdesiredModule\"\xec\x03\n\x0eStatusResponse\x12\x61\n\x07running\x18\x01 \x01(\x0b\x32\x45.xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunningH\x00R\x07running\x12\x61\n\x07success\x18\x02 \x01(\x0b\x32\x45.xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccessH\x00R\x07success\x12^\n\x06\x66\x61iled\x18\x03 \x01(\x0b\x32\x44.xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningFailedH\x00R\x06\x66\x61iled\x1a\x15\n\x13ProvisioningRunning\x1a\x39\n\x12ProvisioningFailed\x12#\n\rerror_message\x18\x01 \x01(\tR\x0c\x65rrorMessage\x1aX\n\x13ProvisioningSuccess\x12\x41\n\x07outputs\x18\x01 \x03(\x0b\x32\'.xyz.block.ftl.schema.v1.RuntimeElementR\x07outputsB\x08\n\x06status2\xc8\x02\n\x18ProvisionerPluginService\x12\x45\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\x12v\n\tProvision\x12\x33.xyz.block.ftl.provisioner.v1beta1.ProvisionRequest\x1a\x34.xyz.block.ftl.provisioner.v1beta1.ProvisionResponse\x12m\n\x06Status\x12\x30.xyz.block.ftl.provisioner.v1beta1.StatusRequest\x1a\x31.xyz.block.ftl.provisioner.v1beta1.StatusResponseBWP\x01ZSgithub.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1;provisionerpbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.provisioner.v1beta1.plugin_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZSgithub.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1;provisionerpb'
  _globals['_PROVISIONREQUEST']._serialized_start=152
  _globals['_PROVISIONREQUEST']._serialized_end=406
  _globals['_PROVISIONRESPONSE']._serialized_start=409
  _globals['_PROVISIONRESPONSE']._serialized_end=688
  _globals['_PROVISIONRESPONSE_PROVISIONRESPONSESTATUS']._serialized_start=579
  _globals['_PROVISIONRESPONSE_PROVISIONRESPONSESTATUS']._serialized_end=688
  _globals['_STATUSREQUEST']._serialized_start=691
  _globals['_STATUSREQUEST']._serialized_end=825
  _globals['_STATUSRESPONSE']._serialized_start=828
  _globals['_STATUSRESPONSE']._serialized_end=1320
  _globals['_STATUSRESPONSE_PROVISIONINGRUNNING']._serialized_start=1140
  _globals['_STATUSRESPONSE_PROVISIONINGRUNNING']._serialized_end=1161
  _globals['_STATUSRESPONSE_PROVISIONINGFAILED']._serialized_start=1163
  _globals['_STATUSRESPONSE_PROVISIONINGFAILED']._serialized_end=1220
  _globals['_STATUSRESPONSE_PROVISIONINGSUCCESS']._serialized_start=1222
  _globals['_STATUSRESPONSE_PROVISIONINGSUCCESS']._serialized_end=1310
  _globals['_PROVISIONERPLUGINSERVICE']._serialized_start=1323
  _globals['_PROVISIONERPLUGINSERVICE']._serialized_end=1651
# @@protoc_insertion_point(module_scope)
