# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/buildengine/v1/buildengine.proto
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
    'xyz/block/ftl/buildengine/v1/buildengine.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.language.v1 import language_pb2 as xyz_dot_block_dot_ftl_dot_language_dot_v1_dot_language__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n.xyz/block/ftl/buildengine/v1/buildengine.proto\x12\x1cxyz.block.ftl.buildengine.v1\x1a(xyz/block/ftl/language/v1/language.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x0f\n\rEngineStarted\"\xd6\x01\n\x0b\x45ngineEnded\x12`\n\rmodule_errors\x18\x01 \x03(\x0b\x32;.xyz.block.ftl.buildengine.v1.EngineEnded.ModuleErrorsEntryR\x0cmoduleErrors\x1a\x65\n\x11ModuleErrorsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12:\n\x05value\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x05value:\x02\x38\x01\"%\n\x0bModuleAdded\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\'\n\rModuleRemoved\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"U\n\x12ModuleBuildWaiting\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\"}\n\x12ModuleBuildStarted\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12&\n\x0fis_auto_rebuild\x18\x02 \x01(\x08R\risAutoRebuild\"\xba\x01\n\x11ModuleBuildFailed\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12<\n\x06\x65rrors\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\x12&\n\x0fis_auto_rebuild\x18\x03 \x01(\x08R\risAutoRebuild\"}\n\x12ModuleBuildSuccess\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12&\n\x0fis_auto_rebuild\x18\x02 \x01(\x08R\risAutoRebuild\"-\n\x13ModuleDeployStarted\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"j\n\x12ModuleDeployFailed\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12<\n\x06\x65rrors\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\"-\n\x13ModuleDeploySuccess\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\xaf\x08\n\x0b\x45ngineEvent\x12T\n\x0e\x65ngine_started\x18\x01 \x01(\x0b\x32+.xyz.block.ftl.buildengine.v1.EngineStartedH\x00R\rengineStarted\x12N\n\x0c\x65ngine_ended\x18\x02 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.EngineEndedH\x00R\x0b\x65ngineEnded\x12N\n\x0cmodule_added\x18\x03 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.ModuleAddedH\x00R\x0bmoduleAdded\x12T\n\x0emodule_removed\x18\x04 \x01(\x0b\x32+.xyz.block.ftl.buildengine.v1.ModuleRemovedH\x00R\rmoduleRemoved\x12\x64\n\x14module_build_waiting\x18\x05 \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildWaitingH\x00R\x12moduleBuildWaiting\x12\x64\n\x14module_build_started\x18\x06 \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildStartedH\x00R\x12moduleBuildStarted\x12\x61\n\x13module_build_failed\x18\x07 \x01(\x0b\x32/.xyz.block.ftl.buildengine.v1.ModuleBuildFailedH\x00R\x11moduleBuildFailed\x12\x64\n\x14module_build_success\x18\x08 \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildSuccessH\x00R\x12moduleBuildSuccess\x12g\n\x15module_deploy_started\x18\t \x01(\x0b\x32\x31.xyz.block.ftl.buildengine.v1.ModuleDeployStartedH\x00R\x13moduleDeployStarted\x12\x64\n\x14module_deploy_failed\x18\n \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleDeployFailedH\x00R\x12moduleDeployFailed\x12g\n\x15module_deploy_success\x18\x0b \x01(\x0b\x32\x31.xyz.block.ftl.buildengine.v1.ModuleDeploySuccessH\x00R\x13moduleDeploySuccessB\x07\n\x05\x65vent\"\x1b\n\x19StreamEngineEventsRequest\"]\n\x1aStreamEngineEventsResponse\x12?\n\x05\x65vent\x18\x01 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.EngineEventR\x05\x65vent2\xee\x01\n\x12\x42uildEngineService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x8b\x01\n\x12StreamEngineEvents\x12\x37.xyz.block.ftl.buildengine.v1.StreamEngineEventsRequest\x1a\x38.xyz.block.ftl.buildengine.v1.StreamEngineEventsResponse\"\x00\x30\x01\x42PZNgithub.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1;buildenginepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.buildengine.v1.buildengine_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'ZNgithub.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1;buildenginepb'
  _globals['_ENGINEENDED_MODULEERRORSENTRY']._loaded_options = None
  _globals['_ENGINEENDED_MODULEERRORSENTRY']._serialized_options = b'8\001'
  _globals['_BUILDENGINESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_BUILDENGINESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ENGINESTARTED']._serialized_start=150
  _globals['_ENGINESTARTED']._serialized_end=165
  _globals['_ENGINEENDED']._serialized_start=168
  _globals['_ENGINEENDED']._serialized_end=382
  _globals['_ENGINEENDED_MODULEERRORSENTRY']._serialized_start=281
  _globals['_ENGINEENDED_MODULEERRORSENTRY']._serialized_end=382
  _globals['_MODULEADDED']._serialized_start=384
  _globals['_MODULEADDED']._serialized_end=421
  _globals['_MODULEREMOVED']._serialized_start=423
  _globals['_MODULEREMOVED']._serialized_end=462
  _globals['_MODULEBUILDWAITING']._serialized_start=464
  _globals['_MODULEBUILDWAITING']._serialized_end=549
  _globals['_MODULEBUILDSTARTED']._serialized_start=551
  _globals['_MODULEBUILDSTARTED']._serialized_end=676
  _globals['_MODULEBUILDFAILED']._serialized_start=679
  _globals['_MODULEBUILDFAILED']._serialized_end=865
  _globals['_MODULEBUILDSUCCESS']._serialized_start=867
  _globals['_MODULEBUILDSUCCESS']._serialized_end=992
  _globals['_MODULEDEPLOYSTARTED']._serialized_start=994
  _globals['_MODULEDEPLOYSTARTED']._serialized_end=1039
  _globals['_MODULEDEPLOYFAILED']._serialized_start=1041
  _globals['_MODULEDEPLOYFAILED']._serialized_end=1147
  _globals['_MODULEDEPLOYSUCCESS']._serialized_start=1149
  _globals['_MODULEDEPLOYSUCCESS']._serialized_end=1194
  _globals['_ENGINEEVENT']._serialized_start=1197
  _globals['_ENGINEEVENT']._serialized_end=2268
  _globals['_STREAMENGINEEVENTSREQUEST']._serialized_start=2270
  _globals['_STREAMENGINEEVENTSREQUEST']._serialized_end=2297
  _globals['_STREAMENGINEEVENTSRESPONSE']._serialized_start=2299
  _globals['_STREAMENGINEEVENTSRESPONSE']._serialized_end=2392
  _globals['_BUILDENGINESERVICE']._serialized_start=2395
  _globals['_BUILDENGINESERVICE']._serialized_end=2633
# @@protoc_insertion_point(module_scope)
