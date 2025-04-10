# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/hotreload/v1/hotreload.proto
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
    'xyz/block/ftl/hotreload/v1/hotreload.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.language.v1 import service_pb2 as xyz_dot_block_dot_ftl_dot_language_dot_v1_dot_service__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n*xyz/block/ftl/hotreload/v1/hotreload.proto\x12\x1axyz.block.ftl.hotreload.v1\x1a\'xyz/block/ftl/language/v1/service.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x8e\x01\n\rReloadRequest\x12(\n\x10\x66orce_new_runner\x18\x01 \x01(\x08R\x0e\x66orceNewRunner\x12,\n\x12new_deployment_key\x18\x02 \x01(\tR\x10newDeploymentKey\x12%\n\x0eschema_changed\x18\x03 \x01(\x08R\rschemaChanged\"g\n\x0eReloadResponse\x12=\n\x05state\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.hotreload.v1.SchemaStateR\x05state\x12\x16\n\x06\x66\x61iled\x18\x02 \x01(\x08R\x06\x66\x61iled\"\x0e\n\x0cWatchRequest\"N\n\rWatchResponse\x12=\n\x05state\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.hotreload.v1.SchemaStateR\x05state\"\x91\x01\n\x11RunnerInfoRequest\x12\x18\n\x07\x61\x64\x64ress\x18\x01 \x01(\tR\x07\x61\x64\x64ress\x12\x1e\n\ndeployment\x18\x02 \x01(\tR\ndeployment\x12\x42\n\tdatabases\x18\x03 \x03(\x0b\x32$.xyz.block.ftl.hotreload.v1.DatabaseR\tdatabases\"8\n\x08\x44\x61tabase\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x18\n\x07\x61\x64\x64ress\x18\x02 \x01(\tR\x07\x61\x64\x64ress\"0\n\x12RunnerInfoResponse\x12\x1a\n\x08outdated\x18\x01 \x01(\x08R\x08outdated\"\x13\n\x11ReloadNotRequired\"N\n\rReloadSuccess\x12=\n\x05state\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.hotreload.v1.SchemaStateR\x05state\"M\n\x0cReloadFailed\x12=\n\x05state\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.hotreload.v1.SchemaStateR\x05state\"\xce\x01\n\x0bSchemaState\x12\x37\n\x06module\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06module\x12<\n\x06\x65rrors\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\x12.\n\x13new_runner_required\x18\x03 \x01(\x08R\x11newRunnerRequired\x12\x18\n\x07version\x18\x04 \x01(\x03R\x07version2\x8c\x03\n\x10HotReloadService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12_\n\x06Reload\x12).xyz.block.ftl.hotreload.v1.ReloadRequest\x1a*.xyz.block.ftl.hotreload.v1.ReloadResponse\x12^\n\x05Watch\x12(.xyz.block.ftl.hotreload.v1.WatchRequest\x1a).xyz.block.ftl.hotreload.v1.WatchResponse0\x01\x12k\n\nRunnerInfo\x12-.xyz.block.ftl.hotreload.v1.RunnerInfoRequest\x1a..xyz.block.ftl.hotreload.v1.RunnerInfoResponseBNP\x01ZJgithub.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1;hotreloadpbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.hotreload.v1.hotreload_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZJgithub.com/block/ftl/backend/protos/xyz/block/ftl/hotreload/v1;hotreloadpb'
  _globals['_HOTRELOADSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_HOTRELOADSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_RELOADREQUEST']._serialized_start=182
  _globals['_RELOADREQUEST']._serialized_end=324
  _globals['_RELOADRESPONSE']._serialized_start=326
  _globals['_RELOADRESPONSE']._serialized_end=429
  _globals['_WATCHREQUEST']._serialized_start=431
  _globals['_WATCHREQUEST']._serialized_end=445
  _globals['_WATCHRESPONSE']._serialized_start=447
  _globals['_WATCHRESPONSE']._serialized_end=525
  _globals['_RUNNERINFOREQUEST']._serialized_start=528
  _globals['_RUNNERINFOREQUEST']._serialized_end=673
  _globals['_DATABASE']._serialized_start=675
  _globals['_DATABASE']._serialized_end=731
  _globals['_RUNNERINFORESPONSE']._serialized_start=733
  _globals['_RUNNERINFORESPONSE']._serialized_end=781
  _globals['_RELOADNOTREQUIRED']._serialized_start=783
  _globals['_RELOADNOTREQUIRED']._serialized_end=802
  _globals['_RELOADSUCCESS']._serialized_start=804
  _globals['_RELOADSUCCESS']._serialized_end=882
  _globals['_RELOADFAILED']._serialized_start=884
  _globals['_RELOADFAILED']._serialized_end=961
  _globals['_SCHEMASTATE']._serialized_start=964
  _globals['_SCHEMASTATE']._serialized_end=1170
  _globals['_HOTRELOADSERVICE']._serialized_start=1173
  _globals['_HOTRELOADSERVICE']._serialized_end=1569
# @@protoc_insertion_point(module_scope)
