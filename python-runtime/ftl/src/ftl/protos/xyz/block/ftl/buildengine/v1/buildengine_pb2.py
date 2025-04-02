# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/buildengine/v1/buildengine.proto
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
    'xyz/block/ftl/buildengine/v1/buildengine.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from xyz.block.ftl.language.v1 import service_pb2 as xyz_dot_block_dot_ftl_dot_language_dot_v1_dot_service__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n.xyz/block/ftl/buildengine/v1/buildengine.proto\x12\x1cxyz.block.ftl.buildengine.v1\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\'xyz/block/ftl/language/v1/service.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x0f\n\rEngineStarted\"\xde\x01\n\x0b\x45ngineEnded\x12J\n\x07modules\x18\x01 \x03(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.EngineEnded.ModuleR\x07modules\x1a\x82\x01\n\x06Module\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x12\n\x04path\x18\x02 \x01(\tR\x04path\x12\x41\n\x06\x65rrors\x18\x03 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListH\x00R\x06\x65rrors\x88\x01\x01\x42\t\n\x07_errors\"%\n\x0bModuleAdded\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\'\n\rModuleRemoved\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"U\n\x12ModuleBuildWaiting\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\"}\n\x12ModuleBuildStarted\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12&\n\x0fis_auto_rebuild\x18\x02 \x01(\x08R\risAutoRebuild\"\xba\x01\n\x11ModuleBuildFailed\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12<\n\x06\x65rrors\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\x12&\n\x0fis_auto_rebuild\x18\x03 \x01(\x08R\risAutoRebuild\"}\n\x12ModuleBuildSuccess\x12?\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x06\x63onfig\x12&\n\x0fis_auto_rebuild\x18\x02 \x01(\x08R\risAutoRebuild\"-\n\x13ModuleDeployWaiting\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"-\n\x13ModuleDeployStarted\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"j\n\x12ModuleDeployFailed\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12<\n\x06\x65rrors\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\"-\n\x13ModuleDeploySuccess\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\xd2\t\n\x0b\x45ngineEvent\x12\x38\n\ttimestamp\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimestamp\x12T\n\x0e\x65ngine_started\x18\x02 \x01(\x0b\x32+.xyz.block.ftl.buildengine.v1.EngineStartedH\x00R\rengineStarted\x12N\n\x0c\x65ngine_ended\x18\x03 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.EngineEndedH\x00R\x0b\x65ngineEnded\x12N\n\x0cmodule_added\x18\x04 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.ModuleAddedH\x00R\x0bmoduleAdded\x12T\n\x0emodule_removed\x18\x05 \x01(\x0b\x32+.xyz.block.ftl.buildengine.v1.ModuleRemovedH\x00R\rmoduleRemoved\x12\x64\n\x14module_build_waiting\x18\x06 \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildWaitingH\x00R\x12moduleBuildWaiting\x12\x64\n\x14module_build_started\x18\x07 \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildStartedH\x00R\x12moduleBuildStarted\x12\x61\n\x13module_build_failed\x18\x08 \x01(\x0b\x32/.xyz.block.ftl.buildengine.v1.ModuleBuildFailedH\x00R\x11moduleBuildFailed\x12\x64\n\x14module_build_success\x18\t \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleBuildSuccessH\x00R\x12moduleBuildSuccess\x12g\n\x15module_deploy_waiting\x18\n \x01(\x0b\x32\x31.xyz.block.ftl.buildengine.v1.ModuleDeployWaitingH\x00R\x13moduleDeployWaiting\x12g\n\x15module_deploy_started\x18\x0b \x01(\x0b\x32\x31.xyz.block.ftl.buildengine.v1.ModuleDeployStartedH\x00R\x13moduleDeployStarted\x12\x64\n\x14module_deploy_failed\x18\x0c \x01(\x0b\x32\x30.xyz.block.ftl.buildengine.v1.ModuleDeployFailedH\x00R\x12moduleDeployFailed\x12g\n\x15module_deploy_success\x18\r \x01(\x0b\x32\x31.xyz.block.ftl.buildengine.v1.ModuleDeploySuccessH\x00R\x13moduleDeploySuccessB\x07\n\x05\x65vent\"B\n\x19StreamEngineEventsRequest\x12%\n\x0ereplay_history\x18\x01 \x01(\x08R\rreplayHistory\"]\n\x1aStreamEngineEventsResponse\x12?\n\x05\x65vent\x18\x01 \x01(\x0b\x32).xyz.block.ftl.buildengine.v1.EngineEventR\x05\x65vent2\xee\x01\n\x12\x42uildEngineService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x8b\x01\n\x12StreamEngineEvents\x12\x37.xyz.block.ftl.buildengine.v1.StreamEngineEventsRequest\x1a\x38.xyz.block.ftl.buildengine.v1.StreamEngineEventsResponse\"\x00\x30\x01\x42PZNgithub.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1;buildenginepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.buildengine.v1.buildengine_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'ZNgithub.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1;buildenginepb'
  _globals['_BUILDENGINESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_BUILDENGINESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ENGINESTARTED']._serialized_start=182
  _globals['_ENGINESTARTED']._serialized_end=197
  _globals['_ENGINEENDED']._serialized_start=200
  _globals['_ENGINEENDED']._serialized_end=422
  _globals['_ENGINEENDED_MODULE']._serialized_start=292
  _globals['_ENGINEENDED_MODULE']._serialized_end=422
  _globals['_MODULEADDED']._serialized_start=424
  _globals['_MODULEADDED']._serialized_end=461
  _globals['_MODULEREMOVED']._serialized_start=463
  _globals['_MODULEREMOVED']._serialized_end=502
  _globals['_MODULEBUILDWAITING']._serialized_start=504
  _globals['_MODULEBUILDWAITING']._serialized_end=589
  _globals['_MODULEBUILDSTARTED']._serialized_start=591
  _globals['_MODULEBUILDSTARTED']._serialized_end=716
  _globals['_MODULEBUILDFAILED']._serialized_start=719
  _globals['_MODULEBUILDFAILED']._serialized_end=905
  _globals['_MODULEBUILDSUCCESS']._serialized_start=907
  _globals['_MODULEBUILDSUCCESS']._serialized_end=1032
  _globals['_MODULEDEPLOYWAITING']._serialized_start=1034
  _globals['_MODULEDEPLOYWAITING']._serialized_end=1079
  _globals['_MODULEDEPLOYSTARTED']._serialized_start=1081
  _globals['_MODULEDEPLOYSTARTED']._serialized_end=1126
  _globals['_MODULEDEPLOYFAILED']._serialized_start=1128
  _globals['_MODULEDEPLOYFAILED']._serialized_end=1234
  _globals['_MODULEDEPLOYSUCCESS']._serialized_start=1236
  _globals['_MODULEDEPLOYSUCCESS']._serialized_end=1281
  _globals['_ENGINEEVENT']._serialized_start=1284
  _globals['_ENGINEEVENT']._serialized_end=2518
  _globals['_STREAMENGINEEVENTSREQUEST']._serialized_start=2520
  _globals['_STREAMENGINEEVENTSREQUEST']._serialized_end=2586
  _globals['_STREAMENGINEEVENTSRESPONSE']._serialized_start=2588
  _globals['_STREAMENGINEEVENTSRESPONSE']._serialized_end=2681
  _globals['_BUILDENGINESERVICE']._serialized_start=2684
  _globals['_BUILDENGINESERVICE']._serialized_end=2922
# @@protoc_insertion_point(module_scope)
