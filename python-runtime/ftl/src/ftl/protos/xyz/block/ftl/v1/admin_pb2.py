# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/admin.proto
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
    'xyz/block/ftl/v1/admin.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1cxyz/block/ftl/v1/admin.proto\x12\x10xyz.block.ftl.v1\x1a\x1axyz/block/ftl/v1/ftl.proto\"G\n\tConfigRef\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\t\n\x07_module\"\xca\x01\n\x11ListConfigRequest\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12*\n\x0einclude_values\x18\x02 \x01(\x08H\x01R\rincludeValues\x88\x01\x01\x12\x41\n\x08provider\x18\x03 \x01(\x0e\x32 .xyz.block.ftl.v1.ConfigProviderH\x02R\x08provider\x88\x01\x01\x42\t\n\x07_moduleB\x11\n\x0f_include_valuesB\x0b\n\t_provider\"\xa4\x01\n\x12ListConfigResponse\x12\x45\n\x07\x63onfigs\x18\x01 \x03(\x0b\x32+.xyz.block.ftl.v1.ListConfigResponse.ConfigR\x07\x63onfigs\x1aG\n\x06\x43onfig\x12\x18\n\x07refPath\x18\x01 \x01(\tR\x07refPath\x12\x19\n\x05value\x18\x02 \x01(\x0cH\x00R\x05value\x88\x01\x01\x42\x08\n\x06_value\"A\n\x10GetConfigRequest\x12-\n\x03ref\x18\x01 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03ref\")\n\x11GetConfigResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"\xa7\x01\n\x10SetConfigRequest\x12\x41\n\x08provider\x18\x01 \x01(\x0e\x32 .xyz.block.ftl.v1.ConfigProviderH\x00R\x08provider\x88\x01\x01\x12-\n\x03ref\x18\x02 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03ref\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\x0b\n\t_provider\"\x13\n\x11SetConfigResponse\"\x93\x01\n\x12UnsetConfigRequest\x12\x41\n\x08provider\x18\x01 \x01(\x0e\x32 .xyz.block.ftl.v1.ConfigProviderH\x00R\x08provider\x88\x01\x01\x12-\n\x03ref\x18\x02 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03refB\x0b\n\t_provider\"\x15\n\x13UnsetConfigResponse\"\xcb\x01\n\x12ListSecretsRequest\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12*\n\x0einclude_values\x18\x02 \x01(\x08H\x01R\rincludeValues\x88\x01\x01\x12\x41\n\x08provider\x18\x03 \x01(\x0e\x32 .xyz.block.ftl.v1.SecretProviderH\x02R\x08provider\x88\x01\x01\x42\t\n\x07_moduleB\x11\n\x0f_include_valuesB\x0b\n\t_provider\"\xa6\x01\n\x13ListSecretsResponse\x12\x46\n\x07secrets\x18\x01 \x03(\x0b\x32,.xyz.block.ftl.v1.ListSecretsResponse.SecretR\x07secrets\x1aG\n\x06Secret\x12\x18\n\x07refPath\x18\x01 \x01(\tR\x07refPath\x12\x19\n\x05value\x18\x02 \x01(\x0cH\x00R\x05value\x88\x01\x01\x42\x08\n\x06_value\"A\n\x10GetSecretRequest\x12-\n\x03ref\x18\x01 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03ref\")\n\x11GetSecretResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"\xa7\x01\n\x10SetSecretRequest\x12\x41\n\x08provider\x18\x01 \x01(\x0e\x32 .xyz.block.ftl.v1.SecretProviderH\x00R\x08provider\x88\x01\x01\x12-\n\x03ref\x18\x02 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03ref\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\x0b\n\t_provider\"\x13\n\x11SetSecretResponse\"\x93\x01\n\x12UnsetSecretRequest\x12\x41\n\x08provider\x18\x01 \x01(\x0e\x32 .xyz.block.ftl.v1.SecretProviderH\x00R\x08provider\x88\x01\x01\x12-\n\x03ref\x18\x02 \x01(\x0b\x32\x1b.xyz.block.ftl.v1.ConfigRefR\x03refB\x0b\n\t_provider\"\x15\n\x13UnsetSecretResponse*D\n\x0e\x43onfigProvider\x12\x11\n\rCONFIG_INLINE\x10\x00\x12\x10\n\x0c\x43ONFIG_ENVAR\x10\x01\x12\r\n\tCONFIG_DB\x10\x02*i\n\x0eSecretProvider\x12\x11\n\rSECRET_INLINE\x10\x00\x12\x10\n\x0cSECRET_ENVAR\x10\x01\x12\x13\n\x0fSECRET_KEYCHAIN\x10\x02\x12\r\n\tSECRET_OP\x10\x03\x12\x0e\n\nSECRET_ASM\x10\x04\x32\x9f\x06\n\x0c\x41\x64minService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12W\n\nConfigList\x12#.xyz.block.ftl.v1.ListConfigRequest\x1a$.xyz.block.ftl.v1.ListConfigResponse\x12T\n\tConfigGet\x12\".xyz.block.ftl.v1.GetConfigRequest\x1a#.xyz.block.ftl.v1.GetConfigResponse\x12T\n\tConfigSet\x12\".xyz.block.ftl.v1.SetConfigRequest\x1a#.xyz.block.ftl.v1.SetConfigResponse\x12Z\n\x0b\x43onfigUnset\x12$.xyz.block.ftl.v1.UnsetConfigRequest\x1a%.xyz.block.ftl.v1.UnsetConfigResponse\x12Z\n\x0bSecretsList\x12$.xyz.block.ftl.v1.ListSecretsRequest\x1a%.xyz.block.ftl.v1.ListSecretsResponse\x12T\n\tSecretGet\x12\".xyz.block.ftl.v1.GetSecretRequest\x1a#.xyz.block.ftl.v1.GetSecretResponse\x12T\n\tSecretSet\x12\".xyz.block.ftl.v1.SetSecretRequest\x1a#.xyz.block.ftl.v1.SetSecretResponse\x12Z\n\x0bSecretUnset\x12$.xyz.block.ftl.v1.UnsetSecretRequest\x1a%.xyz.block.ftl.v1.UnsetSecretResponseBDP\x01Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.admin_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_ADMINSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_ADMINSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_CONFIGPROVIDER']._serialized_start=1846
  _globals['_CONFIGPROVIDER']._serialized_end=1914
  _globals['_SECRETPROVIDER']._serialized_start=1916
  _globals['_SECRETPROVIDER']._serialized_end=2021
  _globals['_CONFIGREF']._serialized_start=78
  _globals['_CONFIGREF']._serialized_end=149
  _globals['_LISTCONFIGREQUEST']._serialized_start=152
  _globals['_LISTCONFIGREQUEST']._serialized_end=354
  _globals['_LISTCONFIGRESPONSE']._serialized_start=357
  _globals['_LISTCONFIGRESPONSE']._serialized_end=521
  _globals['_LISTCONFIGRESPONSE_CONFIG']._serialized_start=450
  _globals['_LISTCONFIGRESPONSE_CONFIG']._serialized_end=521
  _globals['_GETCONFIGREQUEST']._serialized_start=523
  _globals['_GETCONFIGREQUEST']._serialized_end=588
  _globals['_GETCONFIGRESPONSE']._serialized_start=590
  _globals['_GETCONFIGRESPONSE']._serialized_end=631
  _globals['_SETCONFIGREQUEST']._serialized_start=634
  _globals['_SETCONFIGREQUEST']._serialized_end=801
  _globals['_SETCONFIGRESPONSE']._serialized_start=803
  _globals['_SETCONFIGRESPONSE']._serialized_end=822
  _globals['_UNSETCONFIGREQUEST']._serialized_start=825
  _globals['_UNSETCONFIGREQUEST']._serialized_end=972
  _globals['_UNSETCONFIGRESPONSE']._serialized_start=974
  _globals['_UNSETCONFIGRESPONSE']._serialized_end=995
  _globals['_LISTSECRETSREQUEST']._serialized_start=998
  _globals['_LISTSECRETSREQUEST']._serialized_end=1201
  _globals['_LISTSECRETSRESPONSE']._serialized_start=1204
  _globals['_LISTSECRETSRESPONSE']._serialized_end=1370
  _globals['_LISTSECRETSRESPONSE_SECRET']._serialized_start=1299
  _globals['_LISTSECRETSRESPONSE_SECRET']._serialized_end=1370
  _globals['_GETSECRETREQUEST']._serialized_start=1372
  _globals['_GETSECRETREQUEST']._serialized_end=1437
  _globals['_GETSECRETRESPONSE']._serialized_start=1439
  _globals['_GETSECRETRESPONSE']._serialized_end=1480
  _globals['_SETSECRETREQUEST']._serialized_start=1483
  _globals['_SETSECRETREQUEST']._serialized_end=1650
  _globals['_SETSECRETRESPONSE']._serialized_start=1652
  _globals['_SETSECRETRESPONSE']._serialized_end=1671
  _globals['_UNSETSECRETREQUEST']._serialized_start=1674
  _globals['_UNSETSECRETREQUEST']._serialized_end=1821
  _globals['_UNSETSECRETRESPONSE']._serialized_start=1823
  _globals['_UNSETSECRETRESPONSE']._serialized_end=1844
  _globals['_ADMINSERVICE']._serialized_start=2024
  _globals['_ADMINSERVICE']._serialized_end=2823
# @@protoc_insertion_point(module_scope)
