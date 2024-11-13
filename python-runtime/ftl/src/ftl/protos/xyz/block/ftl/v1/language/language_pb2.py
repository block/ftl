# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/language/v1/language.proto
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
    "",
    "xyz/block/ftl/language/v1/language.proto",
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2
from xyz.block.ftl.v1.schema import (
    schema_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_schema_dot_schema__pb2,
)

DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n(xyz/block/ftl/language/v1/language.proto\x12\x19xyz.block.ftl.v1.language\x1a\x1cgoogle/protobuf/struct.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\x1a$xyz/block/ftl/schema/v1/schema.proto"\xdb\x02\n\x0cModuleConfig\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x10\n\x03\x64ir\x18\x02 \x01(\tR\x03\x64ir\x12\x1a\n\x08language\x18\x03 \x01(\tR\x08language\x12\x1d\n\ndeploy_dir\x18\x04 \x01(\tR\tdeployDir\x12\x19\n\x05\x62uild\x18\x05 \x01(\tH\x00R\x05\x62uild\x88\x01\x01\x12\x1d\n\nbuild_lock\x18\x06 \x01(\tR\tbuildLock\x12\x35\n\x14generated_schema_dir\x18\x07 \x01(\tH\x01R\x12generatedSchemaDir\x88\x01\x01\x12\x14\n\x05watch\x18\x08 \x03(\tR\x05watch\x12@\n\x0flanguage_config\x18\t \x01(\x0b\x32\x17.google.protobuf.StructR\x0elanguageConfigB\x08\n\x06_buildB\x17\n\x15_generated_schema_dir"d\n\rProjectConfig\x12\x10\n\x03\x64ir\x18\x01 \x01(\tR\x03\x64ir\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x15\n\x06no_git\x18\x03 \x01(\x08R\x05noGit\x12\x16\n\x06hermit\x18\x04 \x01(\x08R\x06hermit"9\n\x1bGetCreateModuleFlagsRequest\x12\x1a\n\x08language\x18\x01 \x01(\tR\x08language"\xcf\x02\n\x1cGetCreateModuleFlagsResponse\x12R\n\x05\x66lags\x18\x01 \x03(\x0b\x32<.xyz.block.ftl.v1.language.GetCreateModuleFlagsResponse.FlagR\x05\x66lags\x1a\xda\x01\n\x04\x46lag\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x12\n\x04help\x18\x02 \x01(\tR\x04help\x12\x19\n\x05\x65nvar\x18\x03 \x01(\tH\x00R\x05\x65nvar\x88\x01\x01\x12\x19\n\x05short\x18\x04 \x01(\tH\x01R\x05short\x88\x01\x01\x12%\n\x0bplaceholder\x18\x05 \x01(\tH\x02R\x0bplaceholder\x88\x01\x01\x12\x1d\n\x07\x64\x65\x66\x61ult\x18\x06 \x01(\tH\x03R\x07\x64\x65\x66\x61ult\x88\x01\x01\x42\x08\n\x06_envarB\x08\n\x06_shortB\x0e\n\x0c_placeholderB\n\n\x08_default"\xbb\x01\n\x13\x43reateModuleRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x10\n\x03\x64ir\x18\x02 \x01(\tR\x03\x64ir\x12O\n\x0eproject_config\x18\x03 \x01(\x0b\x32(.xyz.block.ftl.v1.language.ProjectConfigR\rprojectConfig\x12-\n\x05\x46lags\x18\x04 \x01(\x0b\x32\x17.google.protobuf.StructR\x05\x46lags"\x16\n\x14\x43reateModuleResponse"/\n\x1bModuleConfigDefaultsRequest\x12\x10\n\x03\x64ir\x18\x01 \x01(\tR\x03\x64ir"\xbd\x02\n\x1cModuleConfigDefaultsResponse\x12\x1d\n\ndeploy_dir\x18\x01 \x01(\tR\tdeployDir\x12\x19\n\x05\x62uild\x18\x02 \x01(\tH\x00R\x05\x62uild\x88\x01\x01\x12"\n\nbuild_lock\x18\x03 \x01(\tH\x01R\tbuildLock\x88\x01\x01\x12\x35\n\x14generated_schema_dir\x18\x04 \x01(\tH\x02R\x12generatedSchemaDir\x88\x01\x01\x12\x14\n\x05watch\x18\x05 \x03(\tR\x05watch\x12@\n\x0flanguage_config\x18\x06 \x01(\x0b\x32\x17.google.protobuf.StructR\x0elanguageConfigB\x08\n\x06_buildB\r\n\x0b_build_lockB\x17\n\x15_generated_schema_dir"c\n\x13\x44\x65pendenciesRequest\x12L\n\rmodule_config\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.ModuleConfigR\x0cmoduleConfig"0\n\x14\x44\x65pendenciesResponse\x12\x18\n\x07modules\x18\x01 \x03(\tR\x07modules"\xe6\x01\n\x0c\x42uildContext\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\x12L\n\rmodule_config\x18\x02 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.ModuleConfigR\x0cmoduleConfig\x12\x37\n\x06schema\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.SchemaR\x06schema\x12"\n\x0c\x64\x65pendencies\x18\x04 \x03(\tR\x0c\x64\x65pendencies\x12\x1b\n\tbuild_env\x18\x05 \x03(\tR\x08\x62uildEnv"i\n\x1a\x42uildContextUpdatedRequest\x12K\n\x0c\x62uildContext\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.BuildContextR\x0c\x62uildContext"\x1d\n\x1b\x42uildContextUpdatedResponse"\xb1\x02\n\x05\x45rror\x12\x10\n\x03msg\x18\x01 \x01(\tR\x03msg\x12\x41\n\x05level\x18\x04 \x01(\x0e\x32+.xyz.block.ftl.v1.language.Error.ErrorLevelR\x05level\x12:\n\x03pos\x18\x05 \x01(\x0b\x32#.xyz.block.ftl.v1.language.PositionH\x00R\x03pos\x88\x01\x01\x12>\n\x04type\x18\x06 \x01(\x0e\x32*.xyz.block.ftl.v1.language.Error.ErrorTypeR\x04type"+\n\nErrorLevel\x12\x08\n\x04INFO\x10\x00\x12\x08\n\x04WARN\x10\x01\x12\t\n\x05\x45RROR\x10\x02""\n\tErrorType\x12\x07\n\x03\x46TL\x10\x00\x12\x0c\n\x08\x43OMPILER\x10\x01\x42\x06\n\x04_pos"z\n\x08Position\x12\x1a\n\x08\x66ilename\x18\x01 \x01(\tR\x08\x66ilename\x12\x12\n\x04line\x18\x02 \x01(\x03R\x04line\x12 \n\x0bstartColumn\x18\x03 \x01(\x03R\x0bstartColumn\x12\x1c\n\tendColumn\x18\x04 \x01(\x03R\tendColumn"E\n\tErrorList\x12\x38\n\x06\x65rrors\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.v1.language.ErrorR\x06\x65rrors"\xd3\x01\n\x0c\x42uildRequest\x12!\n\x0cproject_root\x18\x01 \x01(\tR\x0bprojectRoot\x12\x1d\n\nstubs_root\x18\x02 \x01(\tR\tstubsRoot\x12\x33\n\x15rebuild_automatically\x18\x03 \x01(\x08R\x14rebuildAutomatically\x12L\n\rbuild_context\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.BuildContextR\x0c\x62uildContext"3\n\x12\x41utoRebuildStarted\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId"\xee\x01\n\x0c\x42uildSuccess\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId\x12\x30\n\x14is_automatic_rebuild\x18\x02 \x01(\x08R\x12isAutomaticRebuild\x12\x37\n\x06module\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.ModuleR\x06module\x12\x16\n\x06\x64\x65ploy\x18\x04 \x03(\tR\x06\x64\x65ploy\x12<\n\x06\x65rrors\x18\x05 \x01(\x0b\x32$.xyz.block.ftl.v1.language.ErrorListR\x06\x65rrors"\xd6\x01\n\x0c\x42uildFailure\x12\x1d\n\ncontext_id\x18\x01 \x01(\tR\tcontextId\x12\x30\n\x14is_automatic_rebuild\x18\x02 \x01(\x08R\x12isAutomaticRebuild\x12<\n\x06\x65rrors\x18\x03 \x01(\x0b\x32$.xyz.block.ftl.v1.language.ErrorListR\x06\x65rrors\x12\x37\n\x17invalidate_dependencies\x18\x04 \x01(\x08R\x16invalidateDependencies"\x98\x02\n\nBuildEvent\x12\x61\n\x14\x61uto_rebuild_started\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.v1.language.AutoRebuildStartedH\x00R\x12\x61utoRebuildStarted\x12N\n\rbuild_success\x18\x03 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.BuildSuccessH\x00R\x0c\x62uildSuccess\x12N\n\rbuild_failure\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.BuildFailureH\x00R\x0c\x62uildFailureB\x07\n\x05\x65vent"\xa8\x02\n\x14GenerateStubsRequest\x12\x10\n\x03\x64ir\x18\x01 \x01(\tR\x03\x64ir\x12\x37\n\x06module\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.ModuleR\x06module\x12L\n\rmodule_config\x18\x03 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.ModuleConfigR\x0cmoduleConfig\x12^\n\x14native_module_config\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.ModuleConfigH\x00R\x12nativeModuleConfig\x88\x01\x01\x42\x17\n\x15_native_module_config"\x17\n\x15GenerateStubsResponse"\xa2\x01\n\x19SyncStubReferencesRequest\x12L\n\rmodule_config\x18\x01 \x01(\x0b\x32\'.xyz.block.ftl.v1.language.ModuleConfigR\x0cmoduleConfig\x12\x1d\n\nstubs_root\x18\x02 \x01(\tR\tstubsRoot\x12\x18\n\x07modules\x18\x03 \x03(\tR\x07modules"\x1c\n\x1aSyncStubReferencesResponse2\xb0\x08\n\x0fLanguageService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse"\x03\x90\x02\x01\x12\x87\x01\n\x14GetCreateModuleFlags\x12\x36.xyz.block.ftl.v1.language.GetCreateModuleFlagsRequest\x1a\x37.xyz.block.ftl.v1.language.GetCreateModuleFlagsResponse\x12o\n\x0c\x43reateModule\x12..xyz.block.ftl.v1.language.CreateModuleRequest\x1a/.xyz.block.ftl.v1.language.CreateModuleResponse\x12\x87\x01\n\x14ModuleConfigDefaults\x12\x36.xyz.block.ftl.v1.language.ModuleConfigDefaultsRequest\x1a\x37.xyz.block.ftl.v1.language.ModuleConfigDefaultsResponse\x12r\n\x0fGetDependencies\x12..xyz.block.ftl.v1.language.DependenciesRequest\x1a/.xyz.block.ftl.v1.language.DependenciesResponse\x12Y\n\x05\x42uild\x12\'.xyz.block.ftl.v1.language.BuildRequest\x1a%.xyz.block.ftl.v1.language.BuildEvent0\x01\x12\x84\x01\n\x13\x42uildContextUpdated\x12\x35.xyz.block.ftl.v1.language.BuildContextUpdatedRequest\x1a\x36.xyz.block.ftl.v1.language.BuildContextUpdatedResponse\x12r\n\rGenerateStubs\x12/.xyz.block.ftl.v1.language.GenerateStubsRequest\x1a\x30.xyz.block.ftl.v1.language.GenerateStubsResponse\x12\x81\x01\n\x12SyncStubReferences\x12\x34.xyz.block.ftl.v1.language.SyncStubReferencesRequest\x1a\x35.xyz.block.ftl.v1.language.SyncStubReferencesResponseBRP\x01ZNgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/language/v1;languagepbb\x06proto3'
)

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(
    DESCRIPTOR, "xyz.block.ftl.v1.language.language_pb2", _globals
)
if not _descriptor._USE_C_DESCRIPTORS:
    _globals["DESCRIPTOR"]._loaded_options = None
    _globals[
        "DESCRIPTOR"
    ]._serialized_options = b"P\001ZNgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/language/v1;languagepb"
    _globals["_LANGUAGESERVICE"].methods_by_name["Ping"]._loaded_options = None
    _globals["_LANGUAGESERVICE"].methods_by_name[
        "Ping"
    ]._serialized_options = b"\220\002\001"
    _globals["_MODULECONFIG"]._serialized_start = 168
    _globals["_MODULECONFIG"]._serialized_end = 515
    _globals["_PROJECTCONFIG"]._serialized_start = 517
    _globals["_PROJECTCONFIG"]._serialized_end = 617
    _globals["_GETCREATEMODULEFLAGSREQUEST"]._serialized_start = 619
    _globals["_GETCREATEMODULEFLAGSREQUEST"]._serialized_end = 676
    _globals["_GETCREATEMODULEFLAGSRESPONSE"]._serialized_start = 679
    _globals["_GETCREATEMODULEFLAGSRESPONSE"]._serialized_end = 1014
    _globals["_GETCREATEMODULEFLAGSRESPONSE_FLAG"]._serialized_start = 796
    _globals["_GETCREATEMODULEFLAGSRESPONSE_FLAG"]._serialized_end = 1014
    _globals["_CREATEMODULEREQUEST"]._serialized_start = 1017
    _globals["_CREATEMODULEREQUEST"]._serialized_end = 1204
    _globals["_CREATEMODULERESPONSE"]._serialized_start = 1206
    _globals["_CREATEMODULERESPONSE"]._serialized_end = 1228
    _globals["_MODULECONFIGDEFAULTSREQUEST"]._serialized_start = 1230
    _globals["_MODULECONFIGDEFAULTSREQUEST"]._serialized_end = 1277
    _globals["_MODULECONFIGDEFAULTSRESPONSE"]._serialized_start = 1280
    _globals["_MODULECONFIGDEFAULTSRESPONSE"]._serialized_end = 1597
    _globals["_DEPENDENCIESREQUEST"]._serialized_start = 1599
    _globals["_DEPENDENCIESREQUEST"]._serialized_end = 1698
    _globals["_DEPENDENCIESRESPONSE"]._serialized_start = 1700
    _globals["_DEPENDENCIESRESPONSE"]._serialized_end = 1748
    _globals["_BUILDCONTEXT"]._serialized_start = 1751
    _globals["_BUILDCONTEXT"]._serialized_end = 1981
    _globals["_BUILDCONTEXTUPDATEDREQUEST"]._serialized_start = 1983
    _globals["_BUILDCONTEXTUPDATEDREQUEST"]._serialized_end = 2088
    _globals["_BUILDCONTEXTUPDATEDRESPONSE"]._serialized_start = 2090
    _globals["_BUILDCONTEXTUPDATEDRESPONSE"]._serialized_end = 2119
    _globals["_ERROR"]._serialized_start = 2122
    _globals["_ERROR"]._serialized_end = 2427
    _globals["_ERROR_ERRORLEVEL"]._serialized_start = 2340
    _globals["_ERROR_ERRORLEVEL"]._serialized_end = 2383
    _globals["_ERROR_ERRORTYPE"]._serialized_start = 2385
    _globals["_ERROR_ERRORTYPE"]._serialized_end = 2419
    _globals["_POSITION"]._serialized_start = 2429
    _globals["_POSITION"]._serialized_end = 2551
    _globals["_ERRORLIST"]._serialized_start = 2553
    _globals["_ERRORLIST"]._serialized_end = 2622
    _globals["_BUILDREQUEST"]._serialized_start = 2625
    _globals["_BUILDREQUEST"]._serialized_end = 2836
    _globals["_AUTOREBUILDSTARTED"]._serialized_start = 2838
    _globals["_AUTOREBUILDSTARTED"]._serialized_end = 2889
    _globals["_BUILDSUCCESS"]._serialized_start = 2892
    _globals["_BUILDSUCCESS"]._serialized_end = 3130
    _globals["_BUILDFAILURE"]._serialized_start = 3133
    _globals["_BUILDFAILURE"]._serialized_end = 3347
    _globals["_BUILDEVENT"]._serialized_start = 3350
    _globals["_BUILDEVENT"]._serialized_end = 3630
    _globals["_GENERATESTUBSREQUEST"]._serialized_start = 3633
    _globals["_GENERATESTUBSREQUEST"]._serialized_end = 3929
    _globals["_GENERATESTUBSRESPONSE"]._serialized_start = 3931
    _globals["_GENERATESTUBSRESPONSE"]._serialized_end = 3954
    _globals["_SYNCSTUBREFERENCESREQUEST"]._serialized_start = 3957
    _globals["_SYNCSTUBREFERENCESREQUEST"]._serialized_end = 4119
    _globals["_SYNCSTUBREFERENCESRESPONSE"]._serialized_start = 4121
    _globals["_SYNCSTUBREFERENCESRESPONSE"]._serialized_end = 4149
    _globals["_LANGUAGESERVICE"]._serialized_start = 4152
    _globals["_LANGUAGESERVICE"]._serialized_end = 5224
# @@protoc_insertion_point(module_scope)
