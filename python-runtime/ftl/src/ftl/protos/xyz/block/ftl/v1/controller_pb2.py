# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/controller.proto
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
    'xyz/block/ftl/v1/controller.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n!xyz/block/ftl/v1/controller.proto\x12\x10xyz.block.ftl.v1\x1a\x1cgoogle/protobuf/struct.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"w\n\x17GetCertificationRequest\x12>\n\x07request\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.CertificateContentR\x07request\x12\x1c\n\tsignature\x18\x02 \x01(\x0cR\tsignature\"[\n\x18GetCertificationResponse\x12?\n\x0b\x63\x65rtificate\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.CertificateR\x0b\x63\x65rtificate\"O\n\x12\x43\x65rtificateContent\x12\x1a\n\x08identity\x18\x01 \x01(\tR\x08identity\x12\x1d\n\npublic_key\x18\x02 \x01(\x0cR\tpublicKey\"\x80\x01\n\x0b\x43\x65rtificate\x12>\n\x07\x63ontent\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.CertificateContentR\x07\x63ontent\x12\x31\n\x14\x63ontroller_signature\x18\x03 \x01(\x0cR\x13\x63ontrollerSignature\"@\n\x17GetArtefactDiffsRequest\x12%\n\x0e\x63lient_digests\x18\x01 \x03(\tR\rclientDigests\"\x94\x01\n\x18GetArtefactDiffsResponse\x12\'\n\x0fmissing_digests\x18\x01 \x03(\tR\x0emissingDigests\x12O\n\x10\x63lient_artefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\x0f\x63lientArtefacts\"1\n\x15UploadArtefactRequest\x12\x18\n\x07\x63ontent\x18\x01 \x01(\x0cR\x07\x63ontent\"0\n\x16UploadArtefactResponse\x12\x16\n\x06\x64igest\x18\x02 \x01(\x0cR\x06\x64igest\"`\n\x12\x44\x65ploymentArtefact\x12\x16\n\x06\x64igest\x18\x01 \x01(\tR\x06\x64igest\x12\x12\n\x04path\x18\x02 \x01(\tR\x04path\x12\x1e\n\nexecutable\x18\x03 \x01(\x08R\nexecutable\"\x96\x01\n\x17\x43reateDeploymentRequest\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x42\n\tartefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\tartefacts\"\x94\x01\n\x18\x43reateDeploymentResponse\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x37\n\x15\x61\x63tive_deployment_key\x18\x02 \x01(\tH\x00R\x13\x61\x63tiveDeploymentKey\x88\x01\x01\x42\x18\n\x16_active_deployment_key\"\x93\x01\n\x1dGetDeploymentArtefactsRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12K\n\x0ehave_artefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\rhaveArtefacts\"x\n\x1eGetDeploymentArtefactsResponse\x12@\n\x08\x61rtefact\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\x08\x61rtefact\x12\x14\n\x05\x63hunk\x18\x02 \x01(\x0cR\x05\x63hunk\"=\n\x14GetDeploymentRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\"\x94\x01\n\x15GetDeploymentResponse\x12\x37\n\x06schema\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x42\n\tartefacts\x18\x02 \x03(\x0b\x32$.xyz.block.ftl.v1.DeploymentArtefactR\tartefacts\"\x96\x01\n\x15RegisterRunnerRequest\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12\x1e\n\ndeployment\x18\x03 \x01(\tR\ndeployment\x12/\n\x06labels\x18\x05 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\"\x18\n\x16RegisterRunnerResponse\"\xa3\x01\n\x13UpdateDeployRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12&\n\x0cmin_replicas\x18\x02 \x01(\x05H\x00R\x0bminReplicas\x88\x01\x01\x12\x1f\n\x08\x65ndpoint\x18\x03 \x01(\tH\x01R\x08\x65ndpoint\x88\x01\x01\x42\x0f\n\r_min_replicasB\x0b\n\t_endpoint\"\x16\n\x14UpdateDeployResponse\"`\n\x14ReplaceDeployRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12!\n\x0cmin_replicas\x18\x02 \x01(\x05R\x0bminReplicas\"\x17\n\x15ReplaceDeployResponse\"\xaf\x03\n\x1bStreamDeploymentLogsRequest\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12\x39\n\ntime_stamp\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x1b\n\tlog_level\x18\x04 \x01(\x05R\x08logLevel\x12]\n\nattributes\x18\x05 \x03(\x0b\x32=.xyz.block.ftl.v1.StreamDeploymentLogsRequest.AttributesEntryR\nattributes\x12\x18\n\x07message\x18\x06 \x01(\tR\x07message\x12\x19\n\x05\x65rror\x18\x07 \x01(\tH\x01R\x05\x65rror\x88\x01\x01\x1a=\n\x0f\x41ttributesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x42\x0e\n\x0c_request_keyB\x08\n\x06_error\"\x1e\n\x1cStreamDeploymentLogsResponse\"\x0f\n\rStatusRequest\"\xfc\x06\n\x0eStatusResponse\x12M\n\x0b\x63ontrollers\x18\x01 \x03(\x0b\x32+.xyz.block.ftl.v1.StatusResponse.ControllerR\x0b\x63ontrollers\x12\x41\n\x07runners\x18\x02 \x03(\x0b\x32\'.xyz.block.ftl.v1.StatusResponse.RunnerR\x07runners\x12M\n\x0b\x64\x65ployments\x18\x03 \x03(\x0b\x32+.xyz.block.ftl.v1.StatusResponse.DeploymentR\x0b\x64\x65ployments\x12>\n\x06routes\x18\x05 \x03(\x0b\x32&.xyz.block.ftl.v1.StatusResponse.RouteR\x06routes\x1aT\n\nController\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12\x18\n\x07version\x18\x03 \x01(\tR\x07version\x1a\x9b\x01\n\x06Runner\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12#\n\ndeployment\x18\x03 \x01(\tH\x00R\ndeployment\x88\x01\x01\x12/\n\x06labels\x18\x04 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labelsB\r\n\x0b_deployment\x1a\xf7\x01\n\nDeployment\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08language\x18\x02 \x01(\tR\x08language\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12!\n\x0cmin_replicas\x18\x04 \x01(\x05R\x0bminReplicas\x12\x1a\n\x08replicas\x18\x07 \x01(\x05R\x08replicas\x12/\n\x06labels\x18\x05 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x12\x37\n\x06schema\x18\x06 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x1a[\n\x05Route\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1e\n\ndeployment\x18\x02 \x01(\tR\ndeployment\x12\x1a\n\x08\x65ndpoint\x18\x03 \x01(\tR\x08\x65ndpoint\"\x14\n\x12ProcessListRequest\"\xaf\x03\n\x13ProcessListResponse\x12K\n\tprocesses\x18\x01 \x03(\x0b\x32-.xyz.block.ftl.v1.ProcessListResponse.ProcessR\tprocesses\x1an\n\rProcessRunner\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12/\n\x06labels\x18\x03 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x1a\xda\x01\n\x07Process\x12\x1e\n\ndeployment\x18\x01 \x01(\tR\ndeployment\x12!\n\x0cmin_replicas\x18\x02 \x01(\x05R\x0bminReplicas\x12/\n\x06labels\x18\x03 \x01(\x0b\x32\x17.google.protobuf.StructR\x06labels\x12P\n\x06runner\x18\x04 \x01(\x0b\x32\x33.xyz.block.ftl.v1.ProcessListResponse.ProcessRunnerH\x00R\x06runner\x88\x01\x01\x42\t\n\x07_runner\"\\\n\x18ResetSubscriptionRequest\x12@\n\x0csubscription\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x0csubscription\"\x1b\n\x19ResetSubscriptionResponse2\x9e\x0b\n\x11\x43ontrollerService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12Z\n\x0bProcessList\x12$.xyz.block.ftl.v1.ProcessListRequest\x1a%.xyz.block.ftl.v1.ProcessListResponse\x12K\n\x06Status\x12\x1f.xyz.block.ftl.v1.StatusRequest\x1a .xyz.block.ftl.v1.StatusResponse\x12i\n\x10GetCertification\x12).xyz.block.ftl.v1.GetCertificationRequest\x1a*.xyz.block.ftl.v1.GetCertificationResponse\x12i\n\x10GetArtefactDiffs\x12).xyz.block.ftl.v1.GetArtefactDiffsRequest\x1a*.xyz.block.ftl.v1.GetArtefactDiffsResponse\x12\x63\n\x0eUploadArtefact\x12\'.xyz.block.ftl.v1.UploadArtefactRequest\x1a(.xyz.block.ftl.v1.UploadArtefactResponse\x12i\n\x10\x43reateDeployment\x12).xyz.block.ftl.v1.CreateDeploymentRequest\x1a*.xyz.block.ftl.v1.CreateDeploymentResponse\x12`\n\rGetDeployment\x12&.xyz.block.ftl.v1.GetDeploymentRequest\x1a\'.xyz.block.ftl.v1.GetDeploymentResponse\x12}\n\x16GetDeploymentArtefacts\x12/.xyz.block.ftl.v1.GetDeploymentArtefactsRequest\x1a\x30.xyz.block.ftl.v1.GetDeploymentArtefactsResponse0\x01\x12\x65\n\x0eRegisterRunner\x12\'.xyz.block.ftl.v1.RegisterRunnerRequest\x1a(.xyz.block.ftl.v1.RegisterRunnerResponse(\x01\x12]\n\x0cUpdateDeploy\x12%.xyz.block.ftl.v1.UpdateDeployRequest\x1a&.xyz.block.ftl.v1.UpdateDeployResponse\x12`\n\rReplaceDeploy\x12&.xyz.block.ftl.v1.ReplaceDeployRequest\x1a\'.xyz.block.ftl.v1.ReplaceDeployResponse\x12w\n\x14StreamDeploymentLogs\x12-.xyz.block.ftl.v1.StreamDeploymentLogsRequest\x1a..xyz.block.ftl.v1.StreamDeploymentLogsResponse(\x01\x12l\n\x11ResetSubscription\x12*.xyz.block.ftl.v1.ResetSubscriptionRequest\x1a+.xyz.block.ftl.v1.ResetSubscriptionResponseBDP\x01Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.controller_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_STREAMDEPLOYMENTLOGSREQUEST_ATTRIBUTESENTRY']._loaded_options = None
  _globals['_STREAMDEPLOYMENTLOGSREQUEST_ATTRIBUTESENTRY']._serialized_options = b'8\001'
  _globals['_CONTROLLERSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_CONTROLLERSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_GETCERTIFICATIONREQUEST']._serialized_start=184
  _globals['_GETCERTIFICATIONREQUEST']._serialized_end=303
  _globals['_GETCERTIFICATIONRESPONSE']._serialized_start=305
  _globals['_GETCERTIFICATIONRESPONSE']._serialized_end=396
  _globals['_CERTIFICATECONTENT']._serialized_start=398
  _globals['_CERTIFICATECONTENT']._serialized_end=477
  _globals['_CERTIFICATE']._serialized_start=480
  _globals['_CERTIFICATE']._serialized_end=608
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_start=610
  _globals['_GETARTEFACTDIFFSREQUEST']._serialized_end=674
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_start=677
  _globals['_GETARTEFACTDIFFSRESPONSE']._serialized_end=825
  _globals['_UPLOADARTEFACTREQUEST']._serialized_start=827
  _globals['_UPLOADARTEFACTREQUEST']._serialized_end=876
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_start=878
  _globals['_UPLOADARTEFACTRESPONSE']._serialized_end=926
  _globals['_DEPLOYMENTARTEFACT']._serialized_start=928
  _globals['_DEPLOYMENTARTEFACT']._serialized_end=1024
  _globals['_CREATEDEPLOYMENTREQUEST']._serialized_start=1027
  _globals['_CREATEDEPLOYMENTREQUEST']._serialized_end=1177
  _globals['_CREATEDEPLOYMENTRESPONSE']._serialized_start=1180
  _globals['_CREATEDEPLOYMENTRESPONSE']._serialized_end=1328
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_start=1331
  _globals['_GETDEPLOYMENTARTEFACTSREQUEST']._serialized_end=1478
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_start=1480
  _globals['_GETDEPLOYMENTARTEFACTSRESPONSE']._serialized_end=1600
  _globals['_GETDEPLOYMENTREQUEST']._serialized_start=1602
  _globals['_GETDEPLOYMENTREQUEST']._serialized_end=1663
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_start=1666
  _globals['_GETDEPLOYMENTRESPONSE']._serialized_end=1814
  _globals['_REGISTERRUNNERREQUEST']._serialized_start=1817
  _globals['_REGISTERRUNNERREQUEST']._serialized_end=1967
  _globals['_REGISTERRUNNERRESPONSE']._serialized_start=1969
  _globals['_REGISTERRUNNERRESPONSE']._serialized_end=1993
  _globals['_UPDATEDEPLOYREQUEST']._serialized_start=1996
  _globals['_UPDATEDEPLOYREQUEST']._serialized_end=2159
  _globals['_UPDATEDEPLOYRESPONSE']._serialized_start=2161
  _globals['_UPDATEDEPLOYRESPONSE']._serialized_end=2183
  _globals['_REPLACEDEPLOYREQUEST']._serialized_start=2185
  _globals['_REPLACEDEPLOYREQUEST']._serialized_end=2281
  _globals['_REPLACEDEPLOYRESPONSE']._serialized_start=2283
  _globals['_REPLACEDEPLOYRESPONSE']._serialized_end=2306
  _globals['_STREAMDEPLOYMENTLOGSREQUEST']._serialized_start=2309
  _globals['_STREAMDEPLOYMENTLOGSREQUEST']._serialized_end=2740
  _globals['_STREAMDEPLOYMENTLOGSREQUEST_ATTRIBUTESENTRY']._serialized_start=2653
  _globals['_STREAMDEPLOYMENTLOGSREQUEST_ATTRIBUTESENTRY']._serialized_end=2714
  _globals['_STREAMDEPLOYMENTLOGSRESPONSE']._serialized_start=2742
  _globals['_STREAMDEPLOYMENTLOGSRESPONSE']._serialized_end=2772
  _globals['_STATUSREQUEST']._serialized_start=2774
  _globals['_STATUSREQUEST']._serialized_end=2789
  _globals['_STATUSRESPONSE']._serialized_start=2792
  _globals['_STATUSRESPONSE']._serialized_end=3684
  _globals['_STATUSRESPONSE_CONTROLLER']._serialized_start=3099
  _globals['_STATUSRESPONSE_CONTROLLER']._serialized_end=3183
  _globals['_STATUSRESPONSE_RUNNER']._serialized_start=3186
  _globals['_STATUSRESPONSE_RUNNER']._serialized_end=3341
  _globals['_STATUSRESPONSE_DEPLOYMENT']._serialized_start=3344
  _globals['_STATUSRESPONSE_DEPLOYMENT']._serialized_end=3591
  _globals['_STATUSRESPONSE_ROUTE']._serialized_start=3593
  _globals['_STATUSRESPONSE_ROUTE']._serialized_end=3684
  _globals['_PROCESSLISTREQUEST']._serialized_start=3686
  _globals['_PROCESSLISTREQUEST']._serialized_end=3706
  _globals['_PROCESSLISTRESPONSE']._serialized_start=3709
  _globals['_PROCESSLISTRESPONSE']._serialized_end=4140
  _globals['_PROCESSLISTRESPONSE_PROCESSRUNNER']._serialized_start=3809
  _globals['_PROCESSLISTRESPONSE_PROCESSRUNNER']._serialized_end=3919
  _globals['_PROCESSLISTRESPONSE_PROCESS']._serialized_start=3922
  _globals['_PROCESSLISTRESPONSE_PROCESS']._serialized_end=4140
  _globals['_RESETSUBSCRIPTIONREQUEST']._serialized_start=4142
  _globals['_RESETSUBSCRIPTIONREQUEST']._serialized_end=4234
  _globals['_RESETSUBSCRIPTIONRESPONSE']._serialized_start=4236
  _globals['_RESETSUBSCRIPTIONRESPONSE']._serialized_end=4263
  _globals['_CONTROLLERSERVICE']._serialized_start=4266
  _globals['_CONTROLLERSERVICE']._serialized_end=5704
# @@protoc_insertion_point(module_scope)
