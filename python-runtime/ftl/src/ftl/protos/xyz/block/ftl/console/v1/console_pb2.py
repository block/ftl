# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/console/v1/console.proto
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
    'xyz/block/ftl/console/v1/console.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.buildengine.v1 import buildengine_pb2 as xyz_dot_block_dot_ftl_dot_buildengine_dot_v1_dot_buildengine__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.timeline.v1 import timeline_pb2 as xyz_dot_block_dot_ftl_dot_timeline_dot_v1_dot_timeline__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2
from xyz.block.ftl.v1 import verb_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_verb__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n&xyz/block/ftl/console/v1/console.proto\x12\x18xyz.block.ftl.console.v1\x1a.xyz/block/ftl/buildengine/v1/buildengine.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a(xyz/block/ftl/timeline/v1/timeline.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\x1a\x1bxyz/block/ftl/v1/verb.proto\"e\n\x05\x45\x64ges\x12,\n\x02in\x18\x01 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x02in\x12.\n\x03out\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x03out\"\x90\x01\n\x06\x43onfig\x12\x37\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ConfigR\x06\x63onfig\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\x88\x01\n\x04\x44\x61ta\x12\x31\n\x04\x64\x61ta\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.DataR\x04\x64\x61ta\x12\x16\n\x06schema\x18\x02 \x01(\tR\x06schema\x12\x35\n\x05\x65\x64ges\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\"\x98\x01\n\x08\x44\x61tabase\x12=\n\x08\x64\x61tabase\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.DatabaseR\x08\x64\x61tabase\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\x88\x01\n\x04\x45num\x12\x31\n\x04\x65num\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.EnumR\x04\x65num\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\x8c\x01\n\x05Topic\x12\x34\n\x05topic\x18\x01 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.TopicR\x05topic\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\x9c\x01\n\tTypeAlias\x12@\n\ttypealias\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.TypeAliasR\ttypealias\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\x90\x01\n\x06Secret\x12\x37\n\x06secret\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SecretR\x06secret\x12\x35\n\x05\x65\x64ges\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\x12\x16\n\x06schema\x18\x03 \x01(\tR\x06schema\"\xb8\x01\n\x04Verb\x12\x31\n\x04verb\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.VerbR\x04verb\x12\x16\n\x06schema\x18\x02 \x01(\tR\x06schema\x12.\n\x13json_request_schema\x18\x03 \x01(\tR\x11jsonRequestSchema\x12\x35\n\x05\x65\x64ges\x18\x04 \x01(\x0b\x32\x1f.xyz.block.ftl.console.v1.EdgesR\x05\x65\x64ges\"\xd0\x04\n\x06Module\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x16\n\x06schema\x18\x02 \x01(\tR\x06schema\x12@\n\x07runtime\x18\x03 \x01(\x0b\x32&.xyz.block.ftl.schema.v1.ModuleRuntimeR\x07runtime\x12\x34\n\x05verbs\x18\x04 \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.VerbR\x05verbs\x12\x32\n\x04\x64\x61ta\x18\x05 \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.DataR\x04\x64\x61ta\x12:\n\x07secrets\x18\x06 \x03(\x0b\x32 .xyz.block.ftl.console.v1.SecretR\x07secrets\x12:\n\x07\x63onfigs\x18\x07 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ConfigR\x07\x63onfigs\x12@\n\tdatabases\x18\x08 \x03(\x0b\x32\".xyz.block.ftl.console.v1.DatabaseR\tdatabases\x12\x34\n\x05\x65nums\x18\t \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.EnumR\x05\x65nums\x12\x37\n\x06topics\x18\n \x03(\x0b\x32\x1f.xyz.block.ftl.console.v1.TopicR\x06topics\x12\x45\n\x0btypealiases\x18\x0b \x03(\x0b\x32#.xyz.block.ftl.console.v1.TypeAliasR\x0btypealiases\")\n\rTopologyGroup\x12\x18\n\x07modules\x18\x01 \x03(\tR\x07modules\"K\n\x08Topology\x12?\n\x06levels\x18\x01 \x03(\x0b\x32\'.xyz.block.ftl.console.v1.TopologyGroupR\x06levels\"\x13\n\x11GetModulesRequest\"\x90\x01\n\x12GetModulesResponse\x12:\n\x07modules\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ModuleR\x07modules\x12>\n\x08topology\x18\x02 \x01(\x0b\x32\".xyz.block.ftl.console.v1.TopologyR\x08topology\"\x16\n\x14StreamModulesRequest\"\x93\x01\n\x15StreamModulesResponse\x12:\n\x07modules\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ModuleR\x07modules\x12>\n\x08topology\x18\x02 \x01(\x0b\x32\".xyz.block.ftl.console.v1.TopologyR\x08topology\"N\n\x10GetConfigRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x42\t\n\x07_module\")\n\x11GetConfigResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"d\n\x10SetConfigRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\t\n\x07_module\")\n\x11SetConfigResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"N\n\x10GetSecretRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x42\t\n\x07_module\")\n\x11GetSecretResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"d\n\x10SetSecretRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\t\n\x07_module\")\n\x11SetSecretResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"\x10\n\x0eGetInfoRequest\"J\n\x0fGetInfoResponse\x12\x18\n\x07version\x18\x01 \x01(\tR\x07version\x12\x1d\n\nbuild_time\x18\x04 \x01(\tR\tbuildTime\"-\n\x13\x45xecuteGooseRequest\x12\x16\n\x06prompt\x18\x01 \x01(\tR\x06prompt\"\xe0\x01\n\x14\x45xecuteGooseResponse\x12\x1a\n\x08response\x18\x01 \x01(\tR\x08response\x12M\n\x06source\x18\x02 \x01(\x0e\x32\x35.xyz.block.ftl.console.v1.ExecuteGooseResponse.SourceR\x06source\"]\n\x06Source\x12\x16\n\x12SOURCE_UNSPECIFIED\x10\x00\x12\x11\n\rSOURCE_STDOUT\x10\x01\x12\x11\n\rSOURCE_STDERR\x10\x02\x12\x15\n\x11SOURCE_COMPLETION\x10\x03\x32\xdc\n\n\x0e\x43onsoleService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12g\n\nGetModules\x12+.xyz.block.ftl.console.v1.GetModulesRequest\x1a,.xyz.block.ftl.console.v1.GetModulesResponse\x12r\n\rStreamModules\x12..xyz.block.ftl.console.v1.StreamModulesRequest\x1a/.xyz.block.ftl.console.v1.StreamModulesResponse0\x01\x12l\n\x0bGetTimeline\x12-.xyz.block.ftl.timeline.v1.GetTimelineRequest\x1a..xyz.block.ftl.timeline.v1.GetTimelineResponse\x12w\n\x0eStreamTimeline\x12\x30.xyz.block.ftl.timeline.v1.StreamTimelineRequest\x1a\x31.xyz.block.ftl.timeline.v1.StreamTimelineResponse0\x01\x12\x64\n\tGetConfig\x12*.xyz.block.ftl.console.v1.GetConfigRequest\x1a+.xyz.block.ftl.console.v1.GetConfigResponse\x12\x64\n\tSetConfig\x12*.xyz.block.ftl.console.v1.SetConfigRequest\x1a+.xyz.block.ftl.console.v1.SetConfigResponse\x12\x64\n\tGetSecret\x12*.xyz.block.ftl.console.v1.GetSecretRequest\x1a+.xyz.block.ftl.console.v1.GetSecretResponse\x12\x64\n\tSetSecret\x12*.xyz.block.ftl.console.v1.SetSecretRequest\x1a+.xyz.block.ftl.console.v1.SetSecretResponse\x12\x45\n\x04\x43\x61ll\x12\x1d.xyz.block.ftl.v1.CallRequest\x1a\x1e.xyz.block.ftl.v1.CallResponse\x12\x89\x01\n\x12StreamEngineEvents\x12\x37.xyz.block.ftl.buildengine.v1.StreamEngineEventsRequest\x1a\x38.xyz.block.ftl.buildengine.v1.StreamEngineEventsResponse0\x01\x12^\n\x07GetInfo\x12(.xyz.block.ftl.console.v1.GetInfoRequest\x1a).xyz.block.ftl.console.v1.GetInfoResponse\x12o\n\x0c\x45xecuteGoose\x12-.xyz.block.ftl.console.v1.ExecuteGooseRequest\x1a..xyz.block.ftl.console.v1.ExecuteGooseResponse0\x01\x42JP\x01ZFgithub.com/block/ftl/backend/protos/xyz/block/ftl/console/v1;consolepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.console.v1.console_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZFgithub.com/block/ftl/backend/protos/xyz/block/ftl/console/v1;consolepb'
  _globals['_CONSOLESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_CONSOLESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_EDGES']._serialized_start=253
  _globals['_EDGES']._serialized_end=354
  _globals['_CONFIG']._serialized_start=357
  _globals['_CONFIG']._serialized_end=501
  _globals['_DATA']._serialized_start=504
  _globals['_DATA']._serialized_end=640
  _globals['_DATABASE']._serialized_start=643
  _globals['_DATABASE']._serialized_end=795
  _globals['_ENUM']._serialized_start=798
  _globals['_ENUM']._serialized_end=934
  _globals['_TOPIC']._serialized_start=937
  _globals['_TOPIC']._serialized_end=1077
  _globals['_TYPEALIAS']._serialized_start=1080
  _globals['_TYPEALIAS']._serialized_end=1236
  _globals['_SECRET']._serialized_start=1239
  _globals['_SECRET']._serialized_end=1383
  _globals['_VERB']._serialized_start=1386
  _globals['_VERB']._serialized_end=1570
  _globals['_MODULE']._serialized_start=1573
  _globals['_MODULE']._serialized_end=2165
  _globals['_TOPOLOGYGROUP']._serialized_start=2167
  _globals['_TOPOLOGYGROUP']._serialized_end=2208
  _globals['_TOPOLOGY']._serialized_start=2210
  _globals['_TOPOLOGY']._serialized_end=2285
  _globals['_GETMODULESREQUEST']._serialized_start=2287
  _globals['_GETMODULESREQUEST']._serialized_end=2306
  _globals['_GETMODULESRESPONSE']._serialized_start=2309
  _globals['_GETMODULESRESPONSE']._serialized_end=2453
  _globals['_STREAMMODULESREQUEST']._serialized_start=2455
  _globals['_STREAMMODULESREQUEST']._serialized_end=2477
  _globals['_STREAMMODULESRESPONSE']._serialized_start=2480
  _globals['_STREAMMODULESRESPONSE']._serialized_end=2627
  _globals['_GETCONFIGREQUEST']._serialized_start=2629
  _globals['_GETCONFIGREQUEST']._serialized_end=2707
  _globals['_GETCONFIGRESPONSE']._serialized_start=2709
  _globals['_GETCONFIGRESPONSE']._serialized_end=2750
  _globals['_SETCONFIGREQUEST']._serialized_start=2752
  _globals['_SETCONFIGREQUEST']._serialized_end=2852
  _globals['_SETCONFIGRESPONSE']._serialized_start=2854
  _globals['_SETCONFIGRESPONSE']._serialized_end=2895
  _globals['_GETSECRETREQUEST']._serialized_start=2897
  _globals['_GETSECRETREQUEST']._serialized_end=2975
  _globals['_GETSECRETRESPONSE']._serialized_start=2977
  _globals['_GETSECRETRESPONSE']._serialized_end=3018
  _globals['_SETSECRETREQUEST']._serialized_start=3020
  _globals['_SETSECRETREQUEST']._serialized_end=3120
  _globals['_SETSECRETRESPONSE']._serialized_start=3122
  _globals['_SETSECRETRESPONSE']._serialized_end=3163
  _globals['_GETINFOREQUEST']._serialized_start=3165
  _globals['_GETINFOREQUEST']._serialized_end=3181
  _globals['_GETINFORESPONSE']._serialized_start=3183
  _globals['_GETINFORESPONSE']._serialized_end=3257
  _globals['_EXECUTEGOOSEREQUEST']._serialized_start=3259
  _globals['_EXECUTEGOOSEREQUEST']._serialized_end=3304
  _globals['_EXECUTEGOOSERESPONSE']._serialized_start=3307
  _globals['_EXECUTEGOOSERESPONSE']._serialized_end=3531
  _globals['_EXECUTEGOOSERESPONSE_SOURCE']._serialized_start=3438
  _globals['_EXECUTEGOOSERESPONSE_SOURCE']._serialized_end=3531
  _globals['_CONSOLESERVICE']._serialized_start=3534
  _globals['_CONSOLESERVICE']._serialized_end=4906
# @@protoc_insertion_point(module_scope)
