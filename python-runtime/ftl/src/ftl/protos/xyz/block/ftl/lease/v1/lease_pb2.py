# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/lease/v1/lease.proto
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
    'xyz/block/ftl/lease/v1/lease.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\"xyz/block/ftl/lease/v1/lease.proto\x12\x16xyz.block.ftl.lease.v1\x1a\x1egoogle/protobuf/duration.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"T\n\x13\x41\x63quireLeaseRequest\x12\x10\n\x03key\x18\x01 \x03(\tR\x03key\x12+\n\x03ttl\x18\x03 \x01(\x0b\x32\x19.google.protobuf.DurationR\x03ttl\"\x16\n\x14\x41\x63quireLeaseResponse2\xc9\x01\n\x0cLeaseService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12m\n\x0c\x41\x63quireLease\x12+.xyz.block.ftl.lease.v1.AcquireLeaseRequest\x1a,.xyz.block.ftl.lease.v1.AcquireLeaseResponse(\x01\x30\x01\x42\x46P\x01ZBgithub.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1;leasepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.lease.v1.lease_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZBgithub.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1;leasepb'
  _globals['_LEASESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_LEASESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ACQUIRELEASEREQUEST']._serialized_start=122
  _globals['_ACQUIRELEASEREQUEST']._serialized_end=206
  _globals['_ACQUIRELEASERESPONSE']._serialized_start=208
  _globals['_ACQUIRELEASERESPONSE']._serialized_end=230
  _globals['_LEASESERVICE']._serialized_start=233
  _globals['_LEASESERVICE']._serialized_end=434
# @@protoc_insertion_point(module_scope)
