# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/controller.proto
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
    'xyz/block/ftl/v1/controller.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n!xyz/block/ftl/v1/controller.proto\x12\x10xyz.block.ftl.v1\x1a\x1cgoogle/protobuf/struct.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"@\n\x17GetArtefactDiffsRequest\x12%\n\x0e\x63lient_digests\x18\x01 \x03(\tR\rclientDigests\"\x94\x01\n\x18GetArtefactDiffsResponse\x12\'\n\x0fmissing_digests\x18\x01 \x03(\tR\x0emissingDigests\x12O\n\x10\x63lient_artefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\x0f\x63lientArtefacts\"Y\n\x15UploadArtefactRequest\x12\x16\n\x06\x64igest\x18\x01 \x01(\x0cR\x06\x64igest\x12\x12\n\x04size\x18\x02 \x01(\x03R\x04size\x12\x14\n\x05\x63hunk\x18\x03 \x01(\x0cR\x05\x63hunk\"\x18\n\x16UploadArtefactResponse\"`\n\x12\x44\x65ploymentArtefact\x12\x16\n\x06\x64igest\x18\x01 \x01(\x0cR\x06\x64igest\x12\x12\n\x04path\x18\x02 \x01(\tR\x04path\x12\x1e\n\nexecutable\x18\x03 \x01(\x08R\nexecutable\"\x93\x01\n\x1dGetDeploymentArtefactsRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12K\n\x0ehave_artefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\rhaveArtefacts\"x\n\x1eGetDeploymentArtefactsResponse\x12@\n\x08\x61rtefact\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\x08\x61rtefact\x12\x14\n\x05\x63hunk\x18\x02 \x01(\x0cR\x05\x63hunk\"\x96\x01\n\x15RegisterRunnerRequest\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12\x1e\n\ndeployment\x18\x03 \x01(\tR\ndeployment\x12/\n\x06labels\x18\x05 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\"\x18\n\x16RegisterRunnerResponse\"\x14\n\x12\x43lusterInfoRequest\"9\n\x13\x43lusterInfoResponse\x12\x0e\n\x02os\x18\x01 \x01(\tR\x02os\x12\x12\n\x04\x61rch\x18\x02 \x01(\tR\x04\x61rch\"\x0f\n\rStatusRequest\"\xfc\x06\n\x0eStatusResponse\x12M\n\x0b\x63ontrollers\x18\x01 \x03(\x0b\x32+.xyz.block.ftl.v1.StatusResponse.ControllerR\x0b\x63ontrollers\x12\x41\n\x07runners\x18\x02 \x03(\x0b\x32\'.xyz.block.ftl.v1.StatusResponse.RunnerR\x07runners\x12M\n\x0b\x64\x65ployments\x18\x03 \x03(\x0b\x32+.xyz.block.ftl.v1.StatusResponse.DeploymentR\x0b\x64\x65ployments\x12>\n\x06routes\x18\x05 \x03(\x0b\x32&.xyz.block.ftl.v1.StatusResponse.RouteR\x06routes\x1aT\n\nController\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12\x18\n\x07version\x18\x03 \x01(\tR\x07version\x1a\x9b\x01\n\x06Runner\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12#\n\ndeployment\x18\x03 \x01(\tH\x00R\ndeployment\x88\x01\x01\x12/\n\x06labels\x18\x04 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labelsB\r\n\x0b_deployment\x1a\xf7\x01\n\nDeployment\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08language\x18\x02 \x01(\tR\x08language\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12!\n\x0cmin_replicas\x18\x04 \x01(\x05R\x0bminReplicas\x12\x1a\n\x08replicas\x18\x07 \x01(\x05R\x08replicas\x12/\n\x06labels\x18\x05 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x12\x37\n\x06schema\x18\x06 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x1a[\n\x05Route\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1e\n\ndeployment\x18\x02 \x01(\tR\ndeployment\x12\x1a\n\x08\x65ndpoint\x18\x03 \x01(\tR\x08\x65ndpoint\"\x14\n\x12ProcessListRequest\"\xaf\x03\n\x13ProcessListResponse\x12K\n\tprocesses\x18\x01 \x03(\x0b\x32-.xyz.block.ftl.v1.ProcessListResponse.ProcessR\tprocesses\x1an\n\rProcessRunner\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12/\n\x06labels\x18\x03 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x1a\xda\x01\n\x07Process\x12\x1e\n\ndeployment\x18\x01 \x01(\tR\ndeployment\x12!\n\x0cmin_replicas\x18\x02 \x01(\x05R\x0bminReplicas\x12/\n\x06labels\x18\x03 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x12P\n\x06runner\x18\x04 \x01(\x0b\x32\x33.xyz.block.ftl.v1.ProcessListResponse.ProcessRunnerH\x00R\x06runner\x88\x01\x01\x42\t\n\x07_runner2\x9c\x06\n\x11\x43ontrollerService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Z\n\x0b\x43lusterInfo\x12$.xyz.block.ftl.v1.ClusterInfoRequest\x1a%.xyz.block.ftl.v1.ClusterInfoResponse\x12Z\n\x0bProcessList\x12$.xyz.block.ftl.v1.ProcessListRequest\x1a%.xyz.block.ftl.v1.ProcessListResponse\x12K\n\x06Status\x12\x1f.xyz.block.ftl.v1.StatusRequest\x1a .xyz.block.ftl.v1.StatusResponse\x12i\n\x10GetArtefactDiffs\x12).xyz.block.ftl.v1.GetArtefactDiffsRequest\x1a*.xyz.block.ftl.v1.GetArtefactDiffsResponse\x12\x65\n\x0eUploadArtefact\x12\'.xyz.block.ftl.v1.UploadArtefactRequest\x1a(.xyz.block.ftl.v1.UploadArtefactResponse(\x01\x12}\n\x16GetDeploymentArtefacts\x12/.xyz.block.ftl.v1.GetDeploymentArtefactsRequest\x1a\x30.xyz.block.ftl.v1.GetDeploymentArtefactsResponse0\x01\x12\x65\n\x0eRegisterRunner\x12\'.xyz.block.ftl.v1.RegisterRunnerRequest\x1a(.xyz.block.ftl.v1.RegisterRunnerResponse(\x01\x42>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.controller_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_CONTROLLERSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_CONTROLLERSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_start=151
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_end=215
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_start=218
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_end=366
  _globals['_UPLOADARTEFACTREQUEST']._serialized_start=368
  _globals['_UPLOADARTEFACTREQUEST']._serialized_end=457
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_start=459
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_end=483
  _globals['_DEPLOYMENTARTEFACT']._serialized_start=485
  _globals['_DEPLOYMENTARTEFACT']._serialized_end=581
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_start=584
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_end=731
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_start=733
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_end=853
  _globals['_REGISTERRUNNERREQUEST']._serialized_start=856
  _globals['_REGISTERRUNNERREQUEST']._serialized_end=1006
  _globals['_REGISTERRUNNERRESPONSE']._serialized_start=1008
  _globals['_REGISTERRUNNERRESPONSE']._serialized_end=1032
  _globals['_CLUSTERINFOREQUEST']._serialized_start=1034
  _globals['_CLUSTERINFOREQUEST']._serialized_end=1054
  _globals['_CLUSTERINFORESPONSE']._serialized_start=1056
  _globals['_CLUSTERINFORESPONSE']._serialized_end=1113
  _globals['_STATUSREQUEST']._serialized_start=1115
  _globals['_STATUSREQUEST']._serialized_end=1130
  _globals['_STATUSRESPONSE']._serialized_start=1133
  _globals['_STATUSRESPONSE']._serialized_end=2025
  _globals['_STATUSRESPONSE_CONTROLLER']._serialized_start=1440
  _globals['_STATUSRESPONSE_CONTROLLER']._serialized_end=1524
  _globals['_STATUSRESPONSE_RUNNER']._serialized_start=1527
  _globals['_STATUSRESPONSE_RUNNER']._serialized_end=1682
  _globals['_STATUSRESPONSE_DEPLOYMENT']._serialized_start=1685
  _globals['_STATUSRESPONSE_DEPLOYMENT']._serialized_end=1932
  _globals['_STATUSRESPONSE_ROUTE']._serialized_start=1934
  _globals['_STATUSRESPONSE_ROUTE']._serialized_end=2025
  _globals['_PROCESSLISTREQUEST']._serialized_start=2027
  _globals['_PROCESSLISTREQUEST']._serialized_end=2047
  _globals['_PROCESSLISTRESPONSE']._serialized_start=2050
  _globals['_PROCESSLISTRESPONSE']._serialized_end=2481
  _globals['_PROCESSLISTRESPONSE_PROCESSRUNNER']._serialized_start=2150
  _globals['_PROCESSLISTRESPONSE_PROCESSRUNNER']._serialized_end=2260
  _globals['_PROCESSLISTRESPONSE_PROCESS']._serialized_start=2263
  _globals['_PROCESSLISTRESPONSE_PROCESS']._serialized_end=2481
  _globals['_CONTROLLERSERVICE']._serialized_start=2484
  _globals['_CONTROLLERSERVICE']._serialized_end=3280
# @@protoc_insertion_point(module_scope)
