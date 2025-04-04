# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/schemaservice.proto
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
    'xyz/block/ftl/v1/schemaservice.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v1/schemaservice.proto\x12\x10xyz.block.ftl.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x12\n\x10GetSchemaRequest\"\x90\x01\n\x11GetSchemaResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\x12\x42\n\nchangesets\x18\x02 \x03(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\nchangesets\"<\n\x11PullSchemaRequest\x12\'\n\x0fsubscription_id\x18\x01 \x01(\tR\x0esubscriptionId\"Q\n\x12PullSchemaResponse\x12;\n\x05\x65vent\x18\x01 \x01(\x0b\x32%.xyz.block.ftl.schema.v1.NotificationR\x05\x65vent\"\xa8\x01\n\x1eUpdateDeploymentRuntimeRequest\x12!\n\tchangeset\x18\x01 \x01(\tH\x00R\tchangeset\x88\x01\x01\x12\x14\n\x05realm\x18\x02 \x01(\tR\x05realm\x12?\n\x06update\x18\x03 \x01(\x0b\x32\'.xyz.block.ftl.schema.v1.RuntimeElementR\x06updateB\x0c\n\n_changeset\"!\n\x1fUpdateDeploymentRuntimeResponse\"\x17\n\x15GetDeploymentsRequest\"R\n\x16GetDeploymentsResponse\x12\x38\n\x06schema\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.v1.DeployedSchemaR\x06schema\"e\n\x0bRealmChange\x12\x39\n\x07modules\x18\x01 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\x12\x1b\n\tto_remove\x18\x02 \x03(\tR\x08toRemove\"\\\n\x16\x43reateChangesetRequest\x12\x42\n\rrealm_changes\x18\x01 \x03(\x0b\x32\x1d.xyz.block.ftl.v1.RealmChangeR\x0crealmChanges\"7\n\x17\x43reateChangesetResponse\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x8d\x01\n\x0e\x44\x65ployedSchema\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x37\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x1b\n\tis_active\x18\x03 \x01(\x08R\x08isActive\"7\n\x17PrepareChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x1a\n\x18PrepareChangesetResponse\"6\n\x16\x43ommitChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"[\n\x17\x43ommitChangesetResponse\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\"5\n\x15\x44rainChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x18\n\x16\x44rainChangesetResponse\"8\n\x18\x46inalizeChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x1b\n\x19\x46inalizeChangesetResponse\"4\n\x14\x46\x61ilChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x17\n\x15\x46\x61ilChangesetResponse\"N\n\x18RollbackChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"]\n\x19RollbackChangesetResponse\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\"=\n\x14GetDeploymentRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\"P\n\x15GetDeploymentResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema2\xbb\n\n\rSchemaService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Y\n\tGetSchema\x12\".xyz.block.ftl.v1.GetSchemaRequest\x1a#.xyz.block.ftl.v1.GetSchemaResponse\"\x03\x90\x02\x01\x12^\n\nPullSchema\x12#.xyz.block.ftl.v1.PullSchemaRequest\x1a$.xyz.block.ftl.v1.PullSchemaResponse\"\x03\x90\x02\x01\x30\x01\x12~\n\x17UpdateDeploymentRuntime\x12\x30.xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest\x1a\x31.xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse\x12\x63\n\x0eGetDeployments\x12\'.xyz.block.ftl.v1.GetDeploymentsRequest\x1a(.xyz.block.ftl.v1.GetDeploymentsResponse\x12\x66\n\x0f\x43reateChangeset\x12(.xyz.block.ftl.v1.CreateChangesetRequest\x1a).xyz.block.ftl.v1.CreateChangesetResponse\x12i\n\x10PrepareChangeset\x12).xyz.block.ftl.v1.PrepareChangesetRequest\x1a*.xyz.block.ftl.v1.PrepareChangesetResponse\x12\x66\n\x0f\x43ommitChangeset\x12(.xyz.block.ftl.v1.CommitChangesetRequest\x1a).xyz.block.ftl.v1.CommitChangesetResponse\x12\x63\n\x0e\x44rainChangeset\x12\'.xyz.block.ftl.v1.DrainChangesetRequest\x1a(.xyz.block.ftl.v1.DrainChangesetResponse\x12l\n\x11\x46inalizeChangeset\x12*.xyz.block.ftl.v1.FinalizeChangesetRequest\x1a+.xyz.block.ftl.v1.FinalizeChangesetResponse\x12l\n\x11RollbackChangeset\x12*.xyz.block.ftl.v1.RollbackChangesetRequest\x1a+.xyz.block.ftl.v1.RollbackChangesetResponse\x12`\n\rFailChangeset\x12&.xyz.block.ftl.v1.FailChangesetRequest\x1a\'.xyz.block.ftl.v1.FailChangesetResponse\x12`\n\rGetDeployment\x12&.xyz.block.ftl.v1.GetDeploymentRequest\x1a\'.xyz.block.ftl.v1.GetDeploymentResponseB>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

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
  _globals['_PULLSCHEMAREQUEST']._serialized_end=351
  _globals['_PULLSCHEMARESPONSE']._serialized_start=353
  _globals['_PULLSCHEMARESPONSE']._serialized_end=434
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_start=437
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_end=605
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_start=607
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_end=640
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_start=642
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_end=665
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_start=667
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_end=749
  _globals['_REALMCHANGE']._serialized_start=751
  _globals['_REALMCHANGE']._serialized_end=852
  _globals['_CREATECHANGESETREQUEST']._serialized_start=854
  _globals['_CREATECHANGESETREQUEST']._serialized_end=946
  _globals['_CREATECHANGESETRESPONSE']._serialized_start=948
  _globals['_CREATECHANGESETRESPONSE']._serialized_end=1003
  _globals['_DEPLOYEDSCHEMA']._serialized_start=1006
  _globals['_DEPLOYEDSCHEMA']._serialized_end=1147
  _globals['_PREPARECHANGESETREQUEST']._serialized_start=1149
  _globals['_PREPARECHANGESETREQUEST']._serialized_end=1204
  _globals['_PREPARECHANGESETRESPONSE']._serialized_start=1206
  _globals['_PREPARECHANGESETRESPONSE']._serialized_end=1232
  _globals['_COMMITCHANGESETREQUEST']._serialized_start=1234
  _globals['_COMMITCHANGESETREQUEST']._serialized_end=1288
  _globals['_COMMITCHANGESETRESPONSE']._serialized_start=1290
  _globals['_COMMITCHANGESETRESPONSE']._serialized_end=1381
  _globals['_DRAINCHANGESETREQUEST']._serialized_start=1383
  _globals['_DRAINCHANGESETREQUEST']._serialized_end=1436
  _globals['_DRAINCHANGESETRESPONSE']._serialized_start=1438
  _globals['_DRAINCHANGESETRESPONSE']._serialized_end=1462
  _globals['_FINALIZECHANGESETREQUEST']._serialized_start=1464
  _globals['_FINALIZECHANGESETREQUEST']._serialized_end=1520
  _globals['_FINALIZECHANGESETRESPONSE']._serialized_start=1522
  _globals['_FINALIZECHANGESETRESPONSE']._serialized_end=1549
  _globals['_FAILCHANGESETREQUEST']._serialized_start=1551
  _globals['_FAILCHANGESETREQUEST']._serialized_end=1603
  _globals['_FAILCHANGESETRESPONSE']._serialized_start=1605
  _globals['_FAILCHANGESETRESPONSE']._serialized_end=1628
  _globals['_ROLLBACKCHANGESETREQUEST']._serialized_start=1630
  _globals['_ROLLBACKCHANGESETREQUEST']._serialized_end=1708
  _globals['_ROLLBACKCHANGESETRESPONSE']._serialized_start=1710
  _globals['_ROLLBACKCHANGESETRESPONSE']._serialized_end=1803
  _globals['_GETDEPLOYMENTREQUEST']._serialized_start=1805
  _globals['_GETDEPLOYMENTREQUEST']._serialized_end=1866
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_start=1868
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_end=1948
  _globals['_SCHEMASERVICE']._serialized_start=1951
  _globals['_SCHEMASERVICE']._serialized_end=3290
# @@protoc_insertion_point(module_scope)
