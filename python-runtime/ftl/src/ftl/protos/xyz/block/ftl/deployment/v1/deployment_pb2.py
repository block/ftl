# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/deployment/v1/deployment.proto
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
    'xyz/block/ftl/deployment/v1/deployment.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n,xyz/block/ftl/deployment/v1/deployment.proto\x12\x1bxyz.block.ftl.deployment.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"u\n\x13PublishEventRequest\x12\x32\n\x05topic\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x05topic\x12\x12\n\x04\x62ody\x18\x02 \x01(\x0cR\x04\x62ody\x12\x16\n\x06\x63\x61ller\x18\x03 \x01(\tR\x06\x63\x61ller\"\x16\n\x14PublishEventResponse\"=\n\x1bGetDeploymentContextRequest\x12\x1e\n\ndeployment\x18\x01 \x01(\tR\ndeployment\"\xcb\x06\n\x1cGetDeploymentContextResponse\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1e\n\ndeployment\x18\x02 \x01(\tR\ndeployment\x12`\n\x07\x63onfigs\x18\x03 \x03(\x0b\x32\x46.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.ConfigsEntryR\x07\x63onfigs\x12`\n\x07secrets\x18\x04 \x03(\x0b\x32\x46.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.SecretsEntryR\x07secrets\x12[\n\tdatabases\x18\x05 \x03(\x0b\x32=.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DSNR\tdatabases\x12W\n\x06routes\x18\x06 \x03(\x0b\x32?.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.RouteR\x06routes\x1a\x81\x01\n\x03\x44SN\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12T\n\x04type\x18\x02 \x01(\x0e\x32@.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DbTypeR\x04type\x12\x10\n\x03\x64sn\x18\x03 \x01(\tR\x03\x64sn\x1a\x31\n\x05Route\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x10\n\x03uri\x18\x02 \x01(\tR\x03uri\x1a:\n\x0c\x43onfigsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\x1a:\n\x0cSecretsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"J\n\x06\x44\x62Type\x12\x17\n\x13\x44\x42_TYPE_UNSPECIFIED\x10\x00\x12\x14\n\x10\x44\x42_TYPE_POSTGRES\x10\x01\x12\x11\n\rDB_TYPE_MYSQL\x10\x02\x32\xe4\x02\n\x11\x44\x65ploymentService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x8d\x01\n\x14GetDeploymentContext\x12\x38.xyz.block.ftl.deployment.v1.GetDeploymentContextRequest\x1a\x39.xyz.block.ftl.deployment.v1.GetDeploymentContextResponse0\x01\x12s\n\x0cPublishEvent\x12\x30.xyz.block.ftl.deployment.v1.PublishEventRequest\x1a\x31.xyz.block.ftl.deployment.v1.PublishEventResponseBOP\x01ZKgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.deployment.v1.deployment_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZKgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1;ftlv1'
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_CONFIGSENTRY']._loaded_options = None
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_CONFIGSENTRY']._serialized_options = b'8\001'
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_SECRETSENTRY']._loaded_options = None
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_SECRETSENTRY']._serialized_options = b'8\001'
  _globals['_DEPLOYMENTSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_DEPLOYMENTSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_PUBLISHEVENTREQUEST']._serialized_start=143
  _globals['_PUBLISHEVENTREQUEST']._serialized_end=260
  _globals['_PUBLISHEVENTRESPONSE']._serialized_start=262
  _globals['_PUBLISHEVENTRESPONSE']._serialized_end=284
  _globals['_GETDEPLOYMENTCONTEXTREQUEST']._serialized_start=286
  _globals['_GETDEPLOYMENTCONTEXTREQUEST']._serialized_end=347
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE']._serialized_start=350
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE']._serialized_end=1193
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_DSN']._serialized_start=817
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_DSN']._serialized_end=946
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_ROUTE']._serialized_start=948
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_ROUTE']._serialized_end=997
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_CONFIGSENTRY']._serialized_start=999
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_CONFIGSENTRY']._serialized_end=1057
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_SECRETSENTRY']._serialized_start=1059
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_SECRETSENTRY']._serialized_end=1117
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_DBTYPE']._serialized_start=1119
  _globals['_GETDEPLOYMENTCONTEXTRESPONSE_DBTYPE']._serialized_end=1193
  _globals['_DEPLOYMENTSERVICE']._serialized_start=1196
  _globals['_DEPLOYMENTSERVICE']._serialized_end=1552
# @@protoc_insertion_point(module_scope)
