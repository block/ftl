# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/schemaservice.proto
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
    'xyz/block/ftl/v1/schemaservice.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v1/schemaservice.proto\x12\x10xyz.block.ftl.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x12\n\x10GetSchemaRequest\"\x90\x01\n\x11GetSchemaResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\x12\x42\n\nchangesets\x18\x02 \x03(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\nchangesets\"\x13\n\x11PullSchemaRequest\"\xc2\t\n\x12PullSchemaResponse\x12\x64\n\x11\x63hangeset_created\x18\x04 \x01(\x0b\x32\x35.xyz.block.ftl.v1.PullSchemaResponse.ChangesetCreatedH\x00R\x10\x63hangesetCreated\x12\x61\n\x10\x63hangeset_failed\x18\x05 \x01(\x0b\x32\x34.xyz.block.ftl.v1.PullSchemaResponse.ChangesetFailedH\x00R\x0f\x63hangesetFailed\x12g\n\x12\x63hangeset_commited\x18\x06 \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.ChangesetCommitedH\x00R\x11\x63hangesetCommited\x12g\n\x12\x64\x65ployment_created\x18\x07 \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentCreatedH\x00R\x11\x64\x65ploymentCreated\x12g\n\x12\x64\x65ployment_updated\x18\x08 \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentUpdatedH\x00R\x11\x64\x65ploymentUpdated\x12g\n\x12\x64\x65ployment_removed\x18\t \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentRemovedH\x00R\x11\x64\x65ploymentRemoved\x12\x14\n\x04more\x18\x92\xf7\x01 \x01(\x08R\x04more\x1aT\n\x10\x43hangesetCreated\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\x1a\x39\n\x0f\x43hangesetFailed\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\x1a%\n\x11\x43hangesetCommited\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x1a\\\n\x11\x44\x65ploymentCreated\x12<\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleH\x00R\x06schema\x88\x01\x01\x42\t\n\x07_schema\x1a\x8d\x01\n\x11\x44\x65ploymentUpdated\x12!\n\tchangeset\x18\x01 \x01(\tH\x00R\tchangeset\x88\x01\x01\x12<\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleH\x01R\x06schema\x88\x01\x01\x42\x0c\n\n_changesetB\t\n\x07_schema\x1az\n\x11\x44\x65ploymentRemoved\x12\x15\n\x03key\x18\x01 \x01(\tH\x00R\x03key\x88\x01\x01\x12\x1f\n\x0bmodule_name\x18\x02 \x01(\tR\nmoduleName\x12%\n\x0emodule_removed\x18\x03 \x01(\x08R\rmoduleRemovedB\x06\n\x04_keyB\x07\n\x05\x65vent\"\xb4\x01\n\x1eUpdateDeploymentRuntimeRequest\x12\x1e\n\ndeployment\x18\x01 \x01(\tR\ndeployment\x12!\n\tchangeset\x18\x02 \x01(\tH\x00R\tchangeset\x88\x01\x01\x12\x41\n\x05\x65vent\x18\x03 \x01(\x0b\x32+.xyz.block.ftl.schema.v1.ModuleRuntimeEventR\x05\x65ventB\x0c\n\n_changeset\"!\n\x1fUpdateDeploymentRuntimeResponse\"K\n\x13UpdateSchemaRequest\x12\x34\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.EventR\x05\x65vent\"\x16\n\x14UpdateSchemaResponse\"\x17\n\x15GetDeploymentsRequest\"R\n\x16GetDeploymentsResponse\x12\x38\n\x06schema\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.v1.DeployedSchemaR\x06schema\"S\n\x16\x43reateChangesetRequest\x12\x39\n\x07modules\x18\x01 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\"7\n\x17\x43reateChangesetResponse\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x8d\x01\n\x0e\x44\x65ployedSchema\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x37\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x1b\n\tis_active\x18\x03 \x01(\x08R\x08isActive\"6\n\x16\x43ommitChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x19\n\x17\x43ommitChangesetResponse\"J\n\x14\x46\x61ilChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"\x17\n\x15\x46\x61ilChangesetResponse2\x8c\x07\n\rSchemaService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Y\n\tGetSchema\x12\".xyz.block.ftl.v1.GetSchemaRequest\x1a#.xyz.block.ftl.v1.GetSchemaResponse\"\x03\x90\x02\x01\x12^\n\nPullSchema\x12#.xyz.block.ftl.v1.PullSchemaRequest\x1a$.xyz.block.ftl.v1.PullSchemaResponse\"\x03\x90\x02\x01\x30\x01\x12~\n\x17UpdateDeploymentRuntime\x12\x30.xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest\x1a\x31.xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse\x12]\n\x0cUpdateSchema\x12%.xyz.block.ftl.v1.UpdateSchemaRequest\x1a&.xyz.block.ftl.v1.UpdateSchemaResponse\x12\x63\n\x0eGetDeployments\x12\'.xyz.block.ftl.v1.GetDeploymentsRequest\x1a(.xyz.block.ftl.v1.GetDeploymentsResponse\x12\x66\n\x0f\x43reateChangeset\x12(.xyz.block.ftl.v1.CreateChangesetRequest\x1a).xyz.block.ftl.v1.CreateChangesetResponse\x12\x66\n\x0f\x43ommitChangeset\x12(.xyz.block.ftl.v1.CommitChangesetRequest\x1a).xyz.block.ftl.v1.CommitChangesetResponse\x12`\n\rFailChangeset\x12&.xyz.block.ftl.v1.FailChangesetRequest\x1a\'.xyz.block.ftl.v1.FailChangesetResponseB>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

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
  _globals['_GETSCHEMAREQUEST']._serialized_start=124
  _globals['_GETSCHEMAREQUEST']._serialized_end=142
  _globals['_GETSCHEMARESPONSE']._serialized_start=145
  _globals['_GETSCHEMARESPONSE']._serialized_end=289
  _globals['_PULLSCHEMAREQUEST']._serialized_start=291
  _globals['_PULLSCHEMAREQUEST']._serialized_end=310
  _globals['_PULLSCHEMARESPONSE']._serialized_start=313
  _globals['_PULLSCHEMARESPONSE']._serialized_end=1531
  _globals['_PULLSCHEMARESPONSE_CHANGESETCREATED']._serialized_start=978
  _globals['_PULLSCHEMARESPONSE_CHANGESETCREATED']._serialized_end=1062
  _globals['_PULLSCHEMARESPONSE_CHANGESETFAILED']._serialized_start=1064
  _globals['_PULLSCHEMARESPONSE_CHANGESETFAILED']._serialized_end=1121
  _globals['_PULLSCHEMARESPONSE_CHANGESETCOMMITED']._serialized_start=1123
  _globals['_PULLSCHEMARESPONSE_CHANGESETCOMMITED']._serialized_end=1160
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTCREATED']._serialized_start=1162
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTCREATED']._serialized_end=1254
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTUPDATED']._serialized_start=1257
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTUPDATED']._serialized_end=1398
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTREMOVED']._serialized_start=1400
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTREMOVED']._serialized_end=1522
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_start=1534
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_end=1714
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_start=1716
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_end=1749
  _globals['_UPDATESCHEMAREQUEST']._serialized_start=1751
  _globals['_UPDATESCHEMAREQUEST']._serialized_end=1826
  _globals['_UPDATESCHEMARESPONSE']._serialized_start=1828
  _globals['_UPDATESCHEMARESPONSE']._serialized_end=1850
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_start=1852
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_end=1875
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_start=1877
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_end=1959
  _globals['_CREATECHANGESETREQUEST']._serialized_start=1961
  _globals['_CREATECHANGESETREQUEST']._serialized_end=2044
  _globals['_CREATECHANGESETRESPONSE']._serialized_start=2046
  _globals['_CREATECHANGESETRESPONSE']._serialized_end=2101
  _globals['_DEPLOYEDSCHEMA']._serialized_start=2104
  _globals['_DEPLOYEDSCHEMA']._serialized_end=2245
  _globals['_COMMITCHANGESETREQUEST']._serialized_start=2247
  _globals['_COMMITCHANGESETREQUEST']._serialized_end=2301
  _globals['_COMMITCHANGESETRESPONSE']._serialized_start=2303
  _globals['_COMMITCHANGESETRESPONSE']._serialized_end=2328
  _globals['_FAILCHANGESETREQUEST']._serialized_start=2330
  _globals['_FAILCHANGESETREQUEST']._serialized_end=2404
  _globals['_FAILCHANGESETRESPONSE']._serialized_start=2406
  _globals['_FAILCHANGESETRESPONSE']._serialized_end=2429
  _globals['_SCHEMASERVICE']._serialized_start=2432
  _globals['_SCHEMASERVICE']._serialized_end=3340
# @@protoc_insertion_point(module_scope)
