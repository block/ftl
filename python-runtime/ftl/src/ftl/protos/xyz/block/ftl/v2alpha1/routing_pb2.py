# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v2alpha1/routing.proto
# Protobuf Python Version: 5.28.3
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    3,
    '',
    'xyz/block/ftl/v2alpha1/routing.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2
from xyz.block.ftl.v1 import verb_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_verb__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v2alpha1/routing.proto\x12\x16xyz.block.ftl.v2alpha1\x1a\x1axyz/block/ftl/v1/ftl.proto\x1a\x1bxyz/block/ftl/v1/verb.proto2\xa3\x01\n\x0eRoutingService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x45\n\x04\x43\x61ll\x12\x1d.xyz.block.ftl.v1.CallRequest\x1a\x1e.xyz.block.ftl.v1.CallResponseBDP\x01Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v2alpha1.routing_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1'
  _globals['_ROUTINGSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_ROUTINGSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ROUTINGSERVICE']._serialized_start=122
  _globals['_ROUTINGSERVICE']._serialized_end=285
# @@protoc_insertion_point(module_scope)
