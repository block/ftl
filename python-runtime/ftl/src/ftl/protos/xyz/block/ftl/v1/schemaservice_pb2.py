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


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v1/schemaservice.proto\x12\x10xyz.block.ftl.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x12\n\x10GetSchemaRequest\"\x90\x01\n\x11GetSchemaResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\x12\x42\n\nchangesets\x18\x02 \x03(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\nchangesets\"\x13\n\x11PullSchemaRequest\"\xf8\t\n\x12PullSchemaResponse\x12\x64\n\x11\x63hangeset_created\x18\x04 \x01(\x0b\x32\x35.xyz.block.ftl.v1.PullSchemaResponse.ChangesetCreatedH\x00R\x10\x63hangesetCreated\x12\x61\n\x10\x63hangeset_failed\x18\x05 \x01(\x0b\x32\x34.xyz.block.ftl.v1.PullSchemaResponse.ChangesetFailedH\x00R\x0f\x63hangesetFailed\x12j\n\x13\x63hangeset_committed\x18\x06 \x01(\x0b\x32\x37.xyz.block.ftl.v1.PullSchemaResponse.ChangesetCommittedH\x00R\x12\x63hangesetCommitted\x12g\n\x12\x64\x65ployment_created\x18\x07 \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentCreatedH\x00R\x11\x64\x65ploymentCreated\x12g\n\x12\x64\x65ployment_updated\x18\x08 \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentUpdatedH\x00R\x11\x64\x65ploymentUpdated\x12g\n\x12\x64\x65ployment_removed\x18\t \x01(\x0b\x32\x36.xyz.block.ftl.v1.PullSchemaResponse.DeploymentRemovedH\x00R\x11\x64\x65ploymentRemoved\x12\x14\n\x04more\x18\x92\xf7\x01 \x01(\x08R\x04more\x1aT\n\x10\x43hangesetCreated\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\x1a\x39\n\x0f\x43hangesetFailed\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\x1a&\n\x12\x43hangesetCommitted\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x1a\x8d\x01\n\x11\x44\x65ploymentCreated\x12!\n\tchangeset\x18\x01 \x01(\tH\x00R\tchangeset\x88\x01\x01\x12<\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleH\x01R\x06schema\x88\x01\x01\x42\x0c\n\n_changesetB\t\n\x07_schema\x1a\x8d\x01\n\x11\x44\x65ploymentUpdated\x12!\n\tchangeset\x18\x01 \x01(\tH\x00R\tchangeset\x88\x01\x01\x12<\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleH\x01R\x06schema\x88\x01\x01\x42\x0c\n\n_changesetB\t\n\x07_schema\x1az\n\x11\x44\x65ploymentRemoved\x12\x15\n\x03key\x18\x01 \x01(\tH\x00R\x03key\x88\x01\x01\x12\x1f\n\x0bmodule_name\x18\x02 \x01(\tR\nmoduleName\x12%\n\x0emodule_removed\x18\x03 \x01(\x08R\rmoduleRemovedB\x06\n\x04_keyB\x07\n\x05\x65vent\"c\n\x1eUpdateDeploymentRuntimeRequest\x12\x41\n\x05\x65vent\x18\x01 \x01(\x0b\x32+.xyz.block.ftl.schema.v1.ModuleRuntimeEventR\x05\x65vent\"!\n\x1fUpdateDeploymentRuntimeResponse\"K\n\x13UpdateSchemaRequest\x12\x34\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.EventR\x05\x65vent\"\x16\n\x14UpdateSchemaResponse\"\x17\n\x15GetDeploymentsRequest\"R\n\x16GetDeploymentsResponse\x12\x38\n\x06schema\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.v1.DeployedSchemaR\x06schema\"\x84\x01\n\x16\x43reateChangesetRequest\x12\x39\n\x07modules\x18\x01 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\x12/\n\x13removed_deployments\x18\x02 \x03(\tR\x12removedDeployments\"7\n\x17\x43reateChangesetResponse\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x8d\x01\n\x0e\x44\x65ployedSchema\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x37\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x1b\n\tis_active\x18\x03 \x01(\x08R\x08isActive\"7\n\x17PrepareChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x1a\n\x18PrepareChangesetResponse\"6\n\x16\x43ommitChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"[\n\x17\x43ommitChangesetResponse\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\"5\n\x15\x44rainChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x18\n\x16\x44rainChangesetResponse\"8\n\x18\x46inalizeChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\"\x1b\n\x19\x46inalizeChangesetResponse\"J\n\x14\x46\x61ilChangesetRequest\x12\x1c\n\tchangeset\x18\x01 \x01(\tR\tchangeset\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"\x17\n\x15\x46\x61ilChangesetResponse\"=\n\x14GetDeploymentRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\"P\n\x15GetDeploymentResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema2\xac\n\n\rSchemaService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Y\n\tGetSchema\x12\".xyz.block.ftl.v1.GetSchemaRequest\x1a#.xyz.block.ftl.v1.GetSchemaResponse\"\x03\x90\x02\x01\x12^\n\nPullSchema\x12#.xyz.block.ftl.v1.PullSchemaRequest\x1a$.xyz.block.ftl.v1.PullSchemaResponse\"\x03\x90\x02\x01\x30\x01\x12~\n\x17UpdateDeploymentRuntime\x12\x30.xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest\x1a\x31.xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse\x12]\n\x0cUpdateSchema\x12%.xyz.block.ftl.v1.UpdateSchemaRequest\x1a&.xyz.block.ftl.v1.UpdateSchemaResponse\x12\x63\n\x0eGetDeployments\x12\'.xyz.block.ftl.v1.GetDeploymentsRequest\x1a(.xyz.block.ftl.v1.GetDeploymentsResponse\x12\x66\n\x0f\x43reateChangeset\x12(.xyz.block.ftl.v1.CreateChangesetRequest\x1a).xyz.block.ftl.v1.CreateChangesetResponse\x12i\n\x10PrepareChangeset\x12).xyz.block.ftl.v1.PrepareChangesetRequest\x1a*.xyz.block.ftl.v1.PrepareChangesetResponse\x12\x66\n\x0f\x43ommitChangeset\x12(.xyz.block.ftl.v1.CommitChangesetRequest\x1a).xyz.block.ftl.v1.CommitChangesetResponse\x12\x63\n\x0e\x44rainChangeset\x12\'.xyz.block.ftl.v1.DrainChangesetRequest\x1a(.xyz.block.ftl.v1.DrainChangesetResponse\x12l\n\x11\x46inalizeChangeset\x12*.xyz.block.ftl.v1.FinalizeChangesetRequest\x1a+.xyz.block.ftl.v1.FinalizeChangesetResponse\x12`\n\rFailChangeset\x12&.xyz.block.ftl.v1.FailChangesetRequest\x1a\'.xyz.block.ftl.v1.FailChangesetResponse\x12`\n\rGetDeployment\x12&.xyz.block.ftl.v1.GetDeploymentRequest\x1a\'.xyz.block.ftl.v1.GetDeploymentResponseB>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

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
  _globals['_PULLSCHEMARESPONSE']._serialized_end=1585
  _globals['_PULLSCHEMARESPONSE_CHANGESETCREATED']._serialized_start=981
  _globals['_PULLSCHEMARESPONSE_CHANGESETCREATED']._serialized_end=1065
  _globals['_PULLSCHEMARESPONSE_CHANGESETFAILED']._serialized_start=1067
  _globals['_PULLSCHEMARESPONSE_CHANGESETFAILED']._serialized_end=1124
  _globals['_PULLSCHEMARESPONSE_CHANGESETCOMMITTED']._serialized_start=1126
  _globals['_PULLSCHEMARESPONSE_CHANGESETCOMMITTED']._serialized_end=1164
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTCREATED']._serialized_start=1167
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTCREATED']._serialized_end=1308
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTUPDATED']._serialized_start=1311
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTUPDATED']._serialized_end=1452
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTREMOVED']._serialized_start=1454
  _globals['_PULLSCHEMARESPONSE_DEPLOYMENTREMOVED']._serialized_end=1576
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_start=1587
  _globals['_UPDATEDEPLOYMENTRUNTIMEREQUEST']._serialized_end=1686
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_start=1688
  _globals['_UPDATEDEPLOYMENTRUNTIMERESPONSE']._serialized_end=1721
  _globals['_UPDATESCHEMAREQUEST']._serialized_start=1723
  _globals['_UPDATESCHEMAREQUEST']._serialized_end=1798
  _globals['_UPDATESCHEMARESPONSE']._serialized_start=1800
  _globals['_UPDATESCHEMARESPONSE']._serialized_end=1822
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_start=1824
  _globals['_GETDEPLOYMENTSREQUEST']._serialized_end=1847
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_start=1849
  _globals['_GETDEPLOYMENTSRESPONSE']._serialized_end=1931
  _globals['_CREATECHANGESETREQUEST']._serialized_start=1934
  _globals['_CREATECHANGESETREQUEST']._serialized_end=2066
  _globals['_CREATECHANGESETRESPONSE']._serialized_start=2068
  _globals['_CREATECHANGESETRESPONSE']._serialized_end=2123
  _globals['_DEPLOYEDSCHEMA']._serialized_start=2126
  _globals['_DEPLOYEDSCHEMA']._serialized_end=2267
  _globals['_PREPARECHANGESETREQUEST']._serialized_start=2269
  _globals['_PREPARECHANGESETREQUEST']._serialized_end=2324
  _globals['_PREPARECHANGESETRESPONSE']._serialized_start=2326
  _globals['_PREPARECHANGESETRESPONSE']._serialized_end=2352
  _globals['_COMMITCHANGESETREQUEST']._serialized_start=2354
  _globals['_COMMITCHANGESETREQUEST']._serialized_end=2408
  _globals['_COMMITCHANGESETRESPONSE']._serialized_start=2410
  _globals['_COMMITCHANGESETRESPONSE']._serialized_end=2501
  _globals['_DRAINCHANGESETREQUEST']._serialized_start=2503
  _globals['_DRAINCHANGESETREQUEST']._serialized_end=2556
  _globals['_DRAINCHANGESETRESPONSE']._serialized_start=2558
  _globals['_DRAINCHANGESETRESPONSE']._serialized_end=2582
  _globals['_FINALIZECHANGESETREQUEST']._serialized_start=2584
  _globals['_FINALIZECHANGESETREQUEST']._serialized_end=2640
  _globals['_FINALIZECHANGESETRESPONSE']._serialized_start=2642
  _globals['_FINALIZECHANGESETRESPONSE']._serialized_end=2669
  _globals['_FAILCHANGESETREQUEST']._serialized_start=2671
  _globals['_FAILCHANGESETREQUEST']._serialized_end=2745
  _globals['_FAILCHANGESETRESPONSE']._serialized_start=2747
  _globals['_FAILCHANGESETRESPONSE']._serialized_end=2770
  _globals['_GETDEPLOYMENTREQUEST']._serialized_start=2772
  _globals['_GETDEPLOYMENTREQUEST']._serialized_end=2833
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_start=2835
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_end=2915
  _globals['_SCHEMASERVICE']._serialized_start=2918
  _globals['_SCHEMASERVICE']._serialized_end=4242
# @@protoc_insertion_point(module_scope)
