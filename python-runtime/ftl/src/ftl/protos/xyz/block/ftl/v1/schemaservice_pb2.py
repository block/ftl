# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/schemaservice.proto
# Protobuf Python Version: 5.29.1
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
    1,
    '',
    'xyz/block/ftl/v1/schemaservice.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v1/schemaservice.proto\x12\x10xyz.block.ftl.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x12\n\x10GetSchemaRequest\"L\n\x11GetSchemaResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\"\x13\n\x11PullSchemaRequest\"\xc1\x02\n\x12PullSchemaResponse\x12*\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tH\x00R\rdeploymentKey\x88\x01\x01\x12\x1f\n\x0bmodule_name\x18\x02 \x01(\tR\nmoduleName\x12<\n\x06schema\x18\x04 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleH\x01R\x06schema\x88\x01\x01\x12\x12\n\x04more\x18\x03 \x01(\x08R\x04more\x12G\n\x0b\x63hange_type\x18\x05 \x01(\x0e\x32&.xyz.block.ftl.v1.DeploymentChangeTypeR\nchangeType\x12%\n\x0emodule_removed\x18\x06 \x01(\x08R\rmoduleRemovedB\x11\n\x0f_deployment_keyB\t\n\x07_schema\"\x83\x01\n\x1eUpdateDeploymentRuntimeRequest\x12\x1e\n\ndeployment\x18\x01 \x01(\tR\ndeployment\x12\x41\n\x05\x65vent\x18\x02 \x01(\x0b\x32+.xyz.block.ftl.schema.v1.ModuleRuntimeEventR\x05\x65vent\"!\n\x1fUpdateDeploymentRuntimeResponse*\xa8\x01\n\x14\x44\x65ploymentChangeType\x12&\n\"DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED\x10\x00\x12 \n\x1c\x44\x45PLOYMENT_CHANGE_TYPE_ADDED\x10\x01\x12\"\n\x1e\x44\x45PLOYMENT_CHANGE_TYPE_REMOVED\x10\x02\x12\"\n\x1e\x44\x45PLOYMENT_CHANGE_TYPE_CHANGED\x10\x03\x32\x96\x03\n\rSchemaService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Y\n\tGetSchema\x12\".xyz.block.ftl.v1.GetSchemaRequest\x1a#.xyz.block.ftl.v1.GetSchemaResponse\"\x03\x90\x02\x01\x12^\n\nPullSchema\x12#.xyz.block.ftl.v1.PullSchemaRequest\x1a$.xyz.block.ftl.v1.PullSchemaResponse\"\x03\x90\x02\x01\x30\x01\x12~\n\x17UpdateDeploymentRuntime\x12\x30.xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest\x1a\x31.xyz.block.ftl.v1.UpdateDeploymentRuntimeResponseB>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.schemaservice_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_SCHEMASERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_SCHEMASERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_SCHEMASERVICE'].methods_by_name['GetSchema']._loaded_options = None
  _globals['_SCHEMASERVICE'].methods_by_name['GetSchema']._serialized_options = b'\220\002\001'
  _globals['_SCHEMASERVICE'].methods_by_name['PullSchema']._loaded_options = None
  _globals['_SCHEMASERVICE'].methods_by_name['PullSchema']._serialized_options = b'\220\002\001'
  _globals['_DEPLOYMENTCHANGETYPE']._serialized_start=737
  _globals['_DEPLOYMENTCHANGETYPE']._serialized_end=905
  _globals['_GETSCHEMAREQUEST']._serialized_start=124
  _globals['_GETSCHEMAREQUEST']._serialized_end=142
  _globals['_GETSCHEMARESPONSE']._serialized_start=144
  _globals['_GETSCHEMARESPONSE']._serialized_end=220
  _globals['_PULLSCHEMAREQUEST']._serialized_start=222
  _globals['_PULLSCHEMAREQUEST']._serialized_end=241
  _globals['_PULLSCHEMARESPONSE']._serialized_start=244
  _globals['_PULLSCHEMARESPONSE']._serialized_end=565
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_start=568
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_end=699
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_start=701
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_end=734
  _globals['_SCHEMASERVICE']._serialized_start=908
  _globals['_SCHEMASERVICE']._serialized_end=1314
# @@protoc_insertion_point(module_scope)
