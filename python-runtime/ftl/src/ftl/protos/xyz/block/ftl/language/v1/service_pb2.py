# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/language/v1/service.proto
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
    'xyz/block/ftl/language/v1/service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\'xyz/block/ftl/language/v1/service.proto\x12\x19xyz.block.ftl.language.v1\x1a\x1cgoogle/protobuf/struct.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x81\x03\n\x0cModuleConfig\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x10\n\x03\x64ir\x18\x02 \x01(\tR\x03\x64ir\x12\x1a\n\x08language\x18\x03 \x01(\tR\x08language\x12\x1d\n\ndeploy_dir\x18\x04 \x01(\tR\tdeployDir\x12\x19\n\x05\x62uild\x18\x05 \x01(\tH\x00R\x05\x62uild\x88\x01\x01\x12)\n\x0e\x64\x65v_mode_build\x18\x06 \x01(\tH\x01R\x0c\x64\x65vModeBuild\x88\x01\x01\x12\x1d\n\nbuild_lock\x18\x07 \x01(\tR\tbuildLock\x12\x14\n\x05watch\x18\t \x03(\tR\x05watch\x12@\n\x0flanguage_config\x18\n \x01(\x0b\x32\x17.google.protobuf.StructR\x0elanguageConfig\x12 \n\x0csql_root_dir\x18\x0b \x01(\tR\nsqlRootDir\x12\x14\n\x05realm\x18\x0c \x01(\tR\x05realmB\x08\n\x06_buildB\x11\n\x0f_dev_mode_build\"d\n\rProjectConfig\x12\x10\n\x03\x64ir\x18\x01 \x01(\tR\x03\x64ir\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x15\n\x06no_git\x18\x03 \x01(\x08R\x05noGit\x12\x16\n\x06hermit\x18\x04 \x01(\x08R\x06hermit\"f\n\x16GetDependenciesRequest\x12L\n\rmodule_config\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x0cmoduleConfig\"3\n\x17GetDependenciesResponse\x12\x18\n\x07modules\x18\x01 \x03(\tR\x07modules\"\x8a\x02\n\x0c\x42uildContext\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\x12L\n\rmodule_config\x18\x02 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x0cmoduleConfig\x12\x37\n\x06schema\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\x12\"\n\x0c\x64\x65pendencies\x18\x04 \x03(\tR\x0c\x64\x65pendencies\x12\x1b\n\tbuild_env\x18\x05 \x03(\tR\x08\x62uildEnv\x12\x0e\n\x02os\x18\x06 \x01(\tR\x02os\x12\x12\n\x04\x61rch\x18\x07 \x01(\tR\x04\x61rch\"j\n\x1a\x42uildContextUpdatedRequest\x12L\n\rbuild_context\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.BuildContextR\x0c\x62uildContext\"\x1d\n\x1b\x42uildContextUpdatedResponse\"\xa4\x03\n\x05\x45rror\x12\x10\n\x03msg\x18\x01 \x01(\tR\x03msg\x12\x41\n\x05level\x18\x04 \x01(\x0e\x32+.xyz.block.ftl.language.v1.Error.ErrorLevelR\x05level\x12:\n\x03pos\x18\x05 \x01(\x0b\x32#.xyz.block.ftl.language.v1.PositionH\x00R\x03pos\x88\x01\x01\x12>\n\x04type\x18\x06 \x01(\x0e\x32*.xyz.block.ftl.language.v1.Error.ErrorTypeR\x04type\"l\n\nErrorLevel\x12\x1b\n\x17\x45RROR_LEVEL_UNSPECIFIED\x10\x00\x12\x14\n\x10\x45RROR_LEVEL_INFO\x10\x01\x12\x14\n\x10\x45RROR_LEVEL_WARN\x10\x02\x12\x15\n\x11\x45RROR_LEVEL_ERROR\x10\x03\"T\n\tErrorType\x12\x1a\n\x16\x45RROR_TYPE_UNSPECIFIED\x10\x00\x12\x12\n\x0e\x45RROR_TYPE_FTL\x10\x01\x12\x17\n\x13\x45RROR_TYPE_COMPILER\x10\x02\x42\x06\n\x04_pos\"|\n\x08Position\x12\x1a\n\x08\x66ilename\x18\x01 \x01(\tR\x08\x66ilename\x12\x12\n\x04line\x18\x02 \x01(\x03R\x04line\x12!\n\x0cstart_column\x18\x03 \x01(\x03R\x0bstartColumn\x12\x1d\n\nend_column\x18\x04 \x01(\x03R\tendColumn\"E\n\tErrorList\x12\x38\n\x06\x65rrors\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.language.v1.ErrorR\x06\x65rrors\"\x81\x02\n\x0c\x42uildRequest\x12O\n\x0eproject_config\x18\x01 \x01(\x0b\x32(.xyz.block.ftl.language.v1.ProjectConfigR\rprojectConfig\x12\x1d\n\nstubs_root\x18\x02 \x01(\tR\tstubsRoot\x12\x33\n\x15rebuild_automatically\x18\x03 \x01(\x08R\x14rebuildAutomatically\x12L\n\rbuild_context\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.BuildContextR\x0c\x62uildContext\"3\n\x12\x41utoRebuildStarted\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId\"\xaa\x04\n\x0c\x42uildSuccess\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId\x12\x30\n\x14is_automatic_rebuild\x18\x02 \x01(\x08R\x12isAutomaticRebuild\x12\x37\n\x06module\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06module\x12\x16\n\x06\x64\x65ploy\x18\x04 \x03(\tR\x06\x64\x65ploy\x12!\n\x0c\x64ocker_image\x18\x05 \x01(\tR\x0b\x64ockerImage\x12<\n\x06\x65rrors\x18\x06 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\x12&\n\x0c\x64\x65v_endpoint\x18\x07 \x01(\tH\x00R\x0b\x64\x65vEndpoint\x88\x01\x01\x12\"\n\ndebug_port\x18\x08 \x01(\x05H\x01R\tdebugPort\x88\x01\x01\x12:\n\x17\x64\x65v_hot_reload_endpoint\x18\t \x01(\tH\x02R\x14\x64\x65vHotReloadEndpoint\x88\x01\x01\x12\x38\n\x16\x64\x65v_hot_reload_version\x18\n \x01(\x03H\x03R\x13\x64\x65vHotReloadVersion\x88\x01\x01\x42\x0f\n\r_dev_endpointB\r\n\x0b_debug_portB\x1a\n\x18_dev_hot_reload_endpointB\x19\n\x17_dev_hot_reload_version\"\xd6\x01\n\x0c\x42uildFailure\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId\x12\x30\n\x14is_automatic_rebuild\x18\x02 \x01(\x08R\x12isAutomaticRebuild\x12<\n\x06\x65rrors\x18\x03 \x01(\x0b\x32$.xyz.block.ftl.language.v1.ErrorListR\x06\x65rrors\x12\x37\n\x17invalidate_dependencies\x18\x04 \x01(\x08R\x16invalidateDependencies\"\x9b\x02\n\rBuildResponse\x12\x61\n\x14\x61uto_rebuild_started\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.language.v1.AutoRebuildStartedH\x00R\x12\x61utoRebuildStarted\x12N\n\rbuild_success\x18\x03 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.BuildSuccessH\x00R\x0c\x62uildSuccess\x12N\n\rbuild_failure\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.BuildFailureH\x00R\x0c\x62uildFailureB\x07\n\x05\x65vent\"\xa8\x02\n\x14GenerateStubsRequest\x12\x10\n\x03\x64ir\x18\x01 \x01(\tR\x03\x64ir\x12\x37\n\x06module\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06module\x12L\n\rmodule_config\x18\x03 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x0cmoduleConfig\x12^\n\x14native_module_config\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigH\x00R\x12nativeModuleConfig\x88\x01\x01\x42\x17\n\x15_native_module_config\"\x17\n\x15GenerateStubsResponse\"\xdb\x01\n\x19SyncStubReferencesRequest\x12L\n\rmodule_config\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.language.v1.ModuleConfigR\x0cmoduleConfig\x12\x1d\n\nstubs_root\x18\x02 \x01(\tR\tstubsRoot\x12\x18\n\x07modules\x18\x03 \x03(\tR\x07modules\x12\x37\n\x06schema\x18\x04 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SchemaR\x06schema\"\x1c\n\x1aSyncStubReferencesResponse2\xb4\x05\n\x0fLanguageService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12x\n\x0fGetDependencies\x12\x31.xyz.block.ftl.language.v1.GetDependenciesRequest\x1a\x32.xyz.block.ftl.language.v1.GetDependenciesResponse\x12\\\n\x05\x42uild\x12\'.xyz.block.ftl.language.v1.BuildRequest\x1a(.xyz.block.ftl.language.v1.BuildResponse0\x01\x12\x84\x01\n\x13\x42uildContextUpdated\x12\x35.xyz.block.ftl.language.v1.BuildContextUpdatedRequest\x1a\x36.xyz.block.ftl.language.v1.BuildContextUpdatedResponse\x12r\n\rGenerateStubs\x12/.xyz.block.ftl.language.v1.GenerateStubsRequest\x1a\x30.xyz.block.ftl.language.v1.GenerateStubsResponse\x12\x81\x01\n\x12SyncStubReferences\x12\x34.xyz.block.ftl.language.v1.SyncStubReferencesRequest\x1a\x35.xyz.block.ftl.language.v1.SyncStubReferencesResponseBLP\x01ZHgithub.com/block/ftl/backend/protos/xyz/block/ftl/language/v1;languagepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.language.v1.service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZHgithub.com/block/ftl/backend/protos/xyz/block/ftl/language/v1;languagepb'
  _globals['_LANGUAGESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_LANGUAGESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_MODULECONFIG']._serialized_start=167
  _globals['_MODULECONFIG']._serialized_end=552
  _globals['_PROJECTCONFIG']._serialized_start=554
  _globals['_PROJECTCONFIG']._serialized_end=654
  _globals['_GETDEPENDENCIESREQUEST']._serialized_start=656
  _globals['_GETDEPENDENCIESREQUEST']._serialized_end=758
  _globals['_GETDEPENDENCIESRESPONSE']._serialized_start=760
  _globals['_GETDEPENDENCIESRESPONSE']._serialized_end=811
  _globals['_BUILDCONTEXT']._serialized_start=814
  _globals['_BUILDCONTEXT']._serialized_end=1080
  _globals['_BUILDCONTEXTUPDATEDREQUEST']._serialized_start=1082
  _globals['_BUILDCONTEXTUPDATEDREQUEST']._serialized_end=1188
  _globals['_BUILDCONTEXTUPDATEDRESPONSE']._serialized_start=1190
  _globals['_BUILDCONTEXTUPDATEDRESPONSE']._serialized_end=1219
  _globals['_ERROR']._serialized_start=1222
  _globals['_ERROR']._serialized_end=1642
  _globals['_ERROR_ERRORLEVEL']._serialized_start=1440
  _globals['_ERROR_ERRORLEVEL']._serialized_end=1548
  _globals['_ERROR_ERRORTYPE']._serialized_start=1550
  _globals['_ERROR_ERRORTYPE']._serialized_end=1634
  _globals['_POSITION']._serialized_start=1644
  _globals['_POSITION']._serialized_end=1768
  _globals['_ERRORLIST']._serialized_start=1770
  _globals['_ERRORLIST']._serialized_end=1839
  _globals['_BUILDREQUEST']._serialized_start=1842
  _globals['_BUILDREQUEST']._serialized_end=2099
  _globals['_AUTOREBUILDSTARTED']._serialized_start=2101
  _globals['_AUTOREBUILDSTARTED']._serialized_end=2152
  _globals['_BUILDSUCCESS']._serialized_start=2155
  _globals['_BUILDSUCCESS']._serialized_end=2709
  _globals['_BUILDFAILURE']._serialized_start=2712
  _globals['_BUILDFAILURE']._serialized_end=2926
  _globals['_BUILDRESPONSE']._serialized_start=2929
  _globals['_BUILDRESPONSE']._serialized_end=3212
  _globals['_GENERATESTUBSREQUEST']._serialized_start=3215
  _globals['_GENERATESTUBSREQUEST']._serialized_end=3511
  _globals['_GENERATESTUBSRESPONSE']._serialized_start=3513
  _globals['_GENERATESTUBSRESPONSE']._serialized_end=3536
  _globals['_SYNCSTUBREFERENCESREQUEST']._serialized_start=3539
  _globals['_SYNCSTUBREFERENCESREQUEST']._serialized_end=3758
  _globals['_SYNCSTUBREFERENCESRESPONSE']._serialized_start=3760
  _globals['_SYNCSTUBREFERENCESRESPONSE']._serialized_end=3788
  _globals['_LANGUAGESERVICE']._serialized_start=3791
  _globals['_LANGUAGESERVICE']._serialized_end=4483
# @@protoc_insertion_point(module_scope)
