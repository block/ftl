# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/schema/v1/schema.proto
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
    'xyz/block/ftl/schema/v1/schema.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/schema/v1/schema.proto\x12\x17xyz.block.ftl.schema.v1\x1a\x1fgoogle/protobuf/timestamp.proto\"\xb3\x01\n\x1b\x41WSIAMAuthDatabaseConnector\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08username\x18\x02 \x01(\tR\x08username\x12\x1a\n\x08\x65ndpoint\x18\x03 \x01(\tR\x08\x65ndpoint\x12\x1a\n\x08\x64\x61tabase\x18\x04 \x01(\tR\x08\x64\x61tabaseB\x06\n\x04_pos\"G\n\x03\x41ny\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\x82\x01\n\x05\x41rray\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x37\n\x07\x65lement\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x07\x65lementB\x06\n\x04_pos\"H\n\x04\x42ool\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"I\n\x05\x42ytes\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xe0\x02\n\tChangeset\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x39\n\ncreated_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\x07modules\x18\x03 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\x12\x1b\n\tto_remove\x18\x04 \x03(\tR\x08toRemove\x12J\n\x10removing_modules\x18\x05 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x0fremovingModules\x12=\n\x05state\x18\x06 \x01(\x0e\x32\'.xyz.block.ftl.schema.v1.ChangesetStateR\x05state\x12\x19\n\x05\x65rror\x18\x07 \x01(\tH\x00R\x05\x65rror\x88\x01\x01\x42\x08\n\x06_error\"+\n\x17\x43hangesetCommittedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\"Y\n\x15\x43hangesetCreatedEvent\x12@\n\tchangeset\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\tchangeset\")\n\x15\x43hangesetDrainedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\">\n\x14\x43hangesetFailedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"+\n\x17\x43hangesetFinalizedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\"*\n\x16\x43hangesetPreparedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\"\xad\x01\n\x06\x43onfig\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x04typeB\x06\n\x04_pos\"j\n\x14\x44SNDatabaseConnector\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x10\n\x03\x64sn\x18\x02 \x01(\tR\x03\x64snB\x06\n\x04_pos\"\xd8\x02\n\x04\x44\x61ta\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12O\n\x0ftype_parameters\x18\x05 \x03(\x0b\x32&.xyz.block.ftl.schema.v1.TypeParameterR\x0etypeParameters\x12\x36\n\x06\x66ields\x18\x06 \x03(\x0b\x32\x1e.xyz.block.ftl.schema.v1.FieldR\x06\x66ields\x12=\n\x08metadata\x18\x07 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadataB\x06\n\x04_pos\"\xa6\x02\n\x08\x44\x61tabase\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12I\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32(.xyz.block.ftl.schema.v1.DatabaseRuntimeH\x01R\x07runtime\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04type\x18\x04 \x01(\tR\x04type\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12=\n\x08metadata\x18\x05 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadataB\x06\n\x04_posB\n\n\x08_runtime\"\x80\x02\n\x11\x44\x61tabaseConnector\x12{\n\x1e\x61wsiam_auth_database_connector\x18\x02 \x01(\x0b\x32\x34.xyz.block.ftl.schema.v1.AWSIAMAuthDatabaseConnectorH\x00R\x1b\x61wsiamAuthDatabaseConnector\x12\x65\n\x16\x64sn_database_connector\x18\x01 \x01(\x0b\x32-.xyz.block.ftl.schema.v1.DSNDatabaseConnectorH\x00R\x14\x64snDatabaseConnectorB\x07\n\x05value\"}\n\x0f\x44\x61tabaseRuntime\x12Z\n\x0b\x63onnections\x18\x01 \x01(\x0b\x32\x33.xyz.block.ftl.schema.v1.DatabaseRuntimeConnectionsH\x00R\x0b\x63onnections\x88\x01\x01\x42\x0e\n\x0c_connections\"\x9e\x01\n\x1a\x44\x61tabaseRuntimeConnections\x12>\n\x04read\x18\x01 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.DatabaseConnectorR\x04read\x12@\n\x05write\x18\x02 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.DatabaseConnectorR\x05write\"\xb3\x01\n\x14\x44\x61tabaseRuntimeEvent\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1c\n\tchangeset\x18\x02 \x01(\tR\tchangeset\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12U\n\x0b\x63onnections\x18\x04 \x01(\x0b\x32\x33.xyz.block.ftl.schema.v1.DatabaseRuntimeConnectionsR\x0b\x63onnections\"\xe2\x03\n\x04\x44\x65\x63l\x12\x39\n\x06\x63onfig\x18\x06 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ConfigH\x00R\x06\x63onfig\x12\x33\n\x04\x64\x61ta\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.DataH\x00R\x04\x64\x61ta\x12?\n\x08\x64\x61tabase\x18\x03 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.DatabaseH\x00R\x08\x64\x61tabase\x12\x33\n\x04\x65num\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.EnumH\x00R\x04\x65num\x12\x39\n\x06secret\x18\x07 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SecretH\x00R\x06secret\x12\x36\n\x05topic\x18\t \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.TopicH\x00R\x05topic\x12\x43\n\ntype_alias\x18\x05 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.TypeAliasH\x00R\ttypeAlias\x12\x33\n\x04verb\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.VerbH\x00R\x04verbB\x07\n\x05value\"\xac\x01\n\x18\x44\x65ploymentActivatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12=\n\x0c\x61\x63tivated_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x0b\x61\x63tivatedAt\x12!\n\x0cmin_replicas\x18\x03 \x01(\x03R\x0bminReplicas\x12\x1c\n\tchangeset\x18\x04 \x01(\tR\tchangeset\"\x81\x01\n\x16\x44\x65ploymentCreatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x37\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x1c\n\tchangeset\x18\x03 \x01(\tR\tchangeset\"s\n\x1a\x44\x65ploymentDeactivatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12%\n\x0emodule_removed\x18\x02 \x01(\x08R\rmoduleRemoved\x12\x1c\n\tchangeset\x18\x03 \x01(\tR\tchangeset\"l\n\x1e\x44\x65ploymentReplicasUpdatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08replicas\x18\x02 \x01(\x03R\x08replicas\x12\x1c\n\tchangeset\x18\x03 \x01(\tR\tchangeset\"\x87\x01\n\x1c\x44\x65ploymentSchemaUpdatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x37\n\x06schema\x18\x02 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x06schema\x12\x1c\n\tchangeset\x18\x03 \x01(\tR\tchangeset\"\x93\x02\n\x04\x45num\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x36\n\x04type\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeH\x01R\x04type\x88\x01\x01\x12@\n\x08variants\x18\x06 \x03(\x0b\x32$.xyz.block.ftl.schema.v1.EnumVariantR\x08variantsB\x06\n\x04_posB\x07\n\x05_type\"\xb5\x01\n\x0b\x45numVariant\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x34\n\x05value\x18\x04 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.ValueR\x05valueB\x06\n\x04_pos\"\xf9\x0c\n\x05\x45vent\x12n\n\x19\x63hangeset_committed_event\x18\x0c \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.ChangesetCommittedEventH\x00R\x17\x63hangesetCommittedEvent\x12h\n\x17\x63hangeset_created_event\x18\n \x01(\x0b\x32..xyz.block.ftl.schema.v1.ChangesetCreatedEventH\x00R\x15\x63hangesetCreatedEvent\x12h\n\x17\x63hangeset_drained_event\x18\r \x01(\x0b\x32..xyz.block.ftl.schema.v1.ChangesetDrainedEventH\x00R\x15\x63hangesetDrainedEvent\x12\x65\n\x16\x63hangeset_failed_event\x18\x0f \x01(\x0b\x32-.xyz.block.ftl.schema.v1.ChangesetFailedEventH\x00R\x14\x63hangesetFailedEvent\x12n\n\x19\x63hangeset_finalized_event\x18\x0e \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.ChangesetFinalizedEventH\x00R\x17\x63hangesetFinalizedEvent\x12k\n\x18\x63hangeset_prepared_event\x18\x0b \x01(\x0b\x32/.xyz.block.ftl.schema.v1.ChangesetPreparedEventH\x00R\x16\x63hangesetPreparedEvent\x12\x65\n\x16\x64\x61tabase_runtime_event\x18\x08 \x01(\x0b\x32-.xyz.block.ftl.schema.v1.DatabaseRuntimeEventH\x00R\x14\x64\x61tabaseRuntimeEvent\x12q\n\x1a\x64\x65ployment_activated_event\x18\x04 \x01(\x0b\x32\x31.xyz.block.ftl.schema.v1.DeploymentActivatedEventH\x00R\x18\x64\x65ploymentActivatedEvent\x12k\n\x18\x64\x65ployment_created_event\x18\x01 \x01(\x0b\x32/.xyz.block.ftl.schema.v1.DeploymentCreatedEventH\x00R\x16\x64\x65ploymentCreatedEvent\x12w\n\x1c\x64\x65ployment_deactivated_event\x18\x05 \x01(\x0b\x32\x33.xyz.block.ftl.schema.v1.DeploymentDeactivatedEventH\x00R\x1a\x64\x65ploymentDeactivatedEvent\x12\x84\x01\n!deployment_replicas_updated_event\x18\x03 \x01(\x0b\x32\x37.xyz.block.ftl.schema.v1.DeploymentReplicasUpdatedEventH\x00R\x1e\x64\x65ploymentReplicasUpdatedEvent\x12~\n\x1f\x64\x65ployment_schema_updated_event\x18\x02 \x01(\x0b\x32\x35.xyz.block.ftl.schema.v1.DeploymentSchemaUpdatedEventH\x00R\x1c\x64\x65ploymentSchemaUpdatedEvent\x12_\n\x14module_runtime_event\x18\t \x01(\x0b\x32+.xyz.block.ftl.schema.v1.ModuleRuntimeEventH\x00R\x12moduleRuntimeEvent\x12\\\n\x13topic_runtime_event\x18\x07 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.TopicRuntimeEventH\x00R\x11topicRuntimeEvent\x12Y\n\x12verb_runtime_event\x18\x06 \x01(\x0b\x32).xyz.block.ftl.schema.v1.VerbRuntimeEventH\x00R\x10verbRuntimeEventB\x07\n\x05value\"\xeb\x01\n\x05\x46ield\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x03 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x04type\x12=\n\x08metadata\x18\x05 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadataB\x06\n\x04_pos\"I\n\x05\x46loat\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xe7\x01\n\x14IngressPathComponent\x12_\n\x14ingress_path_literal\x18\x01 \x01(\x0b\x32+.xyz.block.ftl.schema.v1.IngressPathLiteralH\x00R\x12ingressPathLiteral\x12\x65\n\x16ingress_path_parameter\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.schema.v1.IngressPathParameterH\x00R\x14ingressPathParameterB\x07\n\x05value\"j\n\x12IngressPathLiteral\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04text\x18\x02 \x01(\tR\x04textB\x06\n\x04_pos\"l\n\x14IngressPathParameter\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\x06\n\x04_pos\"G\n\x03Int\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"b\n\x08IntValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x14\n\x05value\x18\x02 \x01(\x03R\x05valueB\x06\n\x04_pos\"\xad\x01\n\x03Map\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12/\n\x03key\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x03key\x12\x33\n\x05value\x18\x03 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x05valueB\x06\n\x04_pos\"\xe5\t\n\x08Metadata\x12>\n\x05\x61lias\x18\x05 \x01(\x0b\x32&.xyz.block.ftl.schema.v1.MetadataAliasH\x00R\x05\x61lias\x12G\n\x08\x61rtefact\x18\x0e \x01(\x0b\x32).xyz.block.ftl.schema.v1.MetadataArtefactH\x00R\x08\x61rtefact\x12>\n\x05\x63\x61lls\x18\x01 \x01(\x0b\x32&.xyz.block.ftl.schema.v1.MetadataCallsH\x00R\x05\x63\x61lls\x12\x41\n\x06\x63onfig\x18\n \x01(\x0b\x32\'.xyz.block.ftl.schema.v1.MetadataConfigH\x00R\x06\x63onfig\x12\x45\n\x08\x63ron_job\x18\x03 \x01(\x0b\x32(.xyz.block.ftl.schema.v1.MetadataCronJobH\x00R\x07\x63ronJob\x12J\n\tdatabases\x18\x04 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.MetadataDatabasesH\x00R\tdatabases\x12G\n\x08\x65ncoding\x18\t \x01(\x0b\x32).xyz.block.ftl.schema.v1.MetadataEncodingH\x00R\x08\x65ncoding\x12\x44\n\x07ingress\x18\x02 \x01(\x0b\x32(.xyz.block.ftl.schema.v1.MetadataIngressH\x00R\x07ingress\x12M\n\npartitions\x18\x0f \x01(\x0b\x32+.xyz.block.ftl.schema.v1.MetadataPartitionsH\x00R\npartitions\x12J\n\tpublisher\x18\x0c \x01(\x0b\x32*.xyz.block.ftl.schema.v1.MetadataPublisherH\x00R\tpublisher\x12>\n\x05retry\x18\x06 \x01(\x0b\x32&.xyz.block.ftl.schema.v1.MetadataRetryH\x00R\x05retry\x12K\n\nsql_column\x18\x11 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.MetadataSQLColumnH\x00R\tsqlColumn\x12T\n\rsql_migration\x18\r \x01(\x0b\x32-.xyz.block.ftl.schema.v1.MetadataSQLMigrationH\x00R\x0csqlMigration\x12H\n\tsql_query\x18\x10 \x01(\x0b\x32).xyz.block.ftl.schema.v1.MetadataSQLQueryH\x00R\x08sqlQuery\x12\x44\n\x07secrets\x18\x0b \x01(\x0b\x32(.xyz.block.ftl.schema.v1.MetadataSecretsH\x00R\x07secrets\x12M\n\nsubscriber\x18\x07 \x01(\x0b\x32+.xyz.block.ftl.schema.v1.MetadataSubscriberH\x00R\nsubscriber\x12\x45\n\x08type_map\x18\x08 \x01(\x0b\x32(.xyz.block.ftl.schema.v1.MetadataTypeMapH\x00R\x07typeMapB\x07\n\x05value\"\x9f\x01\n\rMetadataAlias\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x04kind\x18\x02 \x01(\x0e\x32\".xyz.block.ftl.schema.v1.AliasKindR\x04kind\x12\x14\n\x05\x61lias\x18\x03 \x01(\tR\x05\x61liasB\x06\n\x04_pos\"\xa0\x01\n\x10MetadataArtefact\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04path\x18\x02 \x01(\tR\x04path\x12\x16\n\x06\x64igest\x18\x03 \x01(\tR\x06\x64igest\x12\x1e\n\nexecutable\x18\x04 \x01(\x08R\nexecutableB\x06\n\x04_pos\"\x85\x01\n\rMetadataCalls\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x32\n\x05\x63\x61lls\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x05\x63\x61llsB\x06\n\x04_pos\"\x88\x01\n\x0eMetadataConfig\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x34\n\x06\x63onfig\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x06\x63onfigB\x06\n\x04_pos\"g\n\x0fMetadataCronJob\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04\x63ron\x18\x02 \x01(\tR\x04\x63ronB\x06\n\x04_pos\"\x89\x01\n\x11MetadataDatabases\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x32\n\x05\x63\x61lls\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x05\x63\x61llsB\x06\n\x04_pos\"\x82\x01\n\x10MetadataEncoding\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04type\x18\x02 \x01(\tR\x04type\x12\x18\n\x07lenient\x18\x03 \x01(\x08R\x07lenientB\x06\n\x04_pos\"\xc2\x01\n\x0fMetadataIngress\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04type\x18\x02 \x01(\tR\x04type\x12\x16\n\x06method\x18\x03 \x01(\tR\x06method\x12\x41\n\x04path\x18\x04 \x03(\x0b\x32-.xyz.block.ftl.schema.v1.IngressPathComponentR\x04pathB\x06\n\x04_pos\"v\n\x12MetadataPartitions\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1e\n\npartitions\x18\x02 \x01(\x03R\npartitionsB\x06\n\x04_pos\"\x8b\x01\n\x11MetadataPublisher\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x34\n\x06topics\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x06topicsB\x06\n\x04_pos\"\xfb\x01\n\rMetadataRetry\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x19\n\x05\x63ount\x18\x02 \x01(\x03H\x01R\x05\x63ount\x88\x01\x01\x12\x1f\n\x0bmin_backoff\x18\x03 \x01(\tR\nminBackoff\x12\x1f\n\x0bmax_backoff\x18\x04 \x01(\tR\nmaxBackoff\x12\x37\n\x05\x63\x61tch\x18\x05 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefH\x02R\x05\x63\x61tch\x88\x01\x01\x42\x06\n\x04_posB\x08\n\x06_countB\x08\n\x06_catch\"\x7f\n\x11MetadataSQLColumn\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x14\n\x05table\x18\x02 \x01(\tR\x05table\x12\x12\n\x04name\x18\x03 \x01(\tR\x04nameB\x06\n\x04_pos\"p\n\x14MetadataSQLMigration\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x16\n\x06\x64igest\x18\x02 \x01(\tR\x06\x64igestB\x06\n\x04_pos\"\x84\x01\n\x10MetadataSQLQuery\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x18\n\x07\x63ommand\x18\x02 \x01(\tR\x07\x63ommand\x12\x14\n\x05query\x18\x03 \x01(\tR\x05queryB\x06\n\x04_pos\"\x8b\x01\n\x0fMetadataSecrets\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x07secrets\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x07secretsB\x06\n\x04_pos\"\xf1\x01\n\x12MetadataSubscriber\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x32\n\x05topic\x18\x02 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x05topic\x12\x44\n\x0b\x66rom_offset\x18\x03 \x01(\x0e\x32#.xyz.block.ftl.schema.v1.FromOffsetR\nfromOffset\x12\x1f\n\x0b\x64\x65\x61\x64_letter\x18\x04 \x01(\x08R\ndeadLetterB\x06\n\x04_pos\"\x8e\x01\n\x0fMetadataTypeMap\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x18\n\x07runtime\x18\x02 \x01(\tR\x07runtime\x12\x1f\n\x0bnative_name\x18\x03 \x01(\tR\nnativeNameB\x06\n\x04_pos\"\xcc\x02\n\x06Module\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x18\n\x07\x62uiltin\x18\x03 \x01(\x08R\x07\x62uiltin\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12=\n\x08metadata\x18\x06 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadata\x12\x33\n\x05\x64\x65\x63ls\x18\x05 \x03(\x0b\x32\x1d.xyz.block.ftl.schema.v1.DeclR\x05\x64\x65\x63ls\x12\x42\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32&.xyz.block.ftl.schema.v1.ModuleRuntimeR\x07runtimeB\x06\n\x04_pos\"\x8f\x02\n\rModuleRuntime\x12>\n\x04\x62\x61se\x18\x01 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.ModuleRuntimeBaseR\x04\x62\x61se\x12L\n\x07scaling\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.schema.v1.ModuleRuntimeScalingH\x00R\x07scaling\x88\x01\x01\x12U\n\ndeployment\x18\x03 \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.ModuleRuntimeDeploymentH\x01R\ndeployment\x88\x01\x01\x42\n\n\x08_scalingB\r\n\x0b_deployment\"\xcf\x01\n\x11ModuleRuntimeBase\x12;\n\x0b\x63reate_time\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ncreateTime\x12\x1a\n\x08language\x18\x02 \x01(\tR\x08language\x12\x13\n\x02os\x18\x03 \x01(\tH\x00R\x02os\x88\x01\x01\x12\x17\n\x04\x61rch\x18\x04 \x01(\tH\x01R\x04\x61rch\x88\x01\x01\x12\x19\n\x05image\x18\x05 \x01(\tH\x02R\x05image\x88\x01\x01\x42\x05\n\x03_osB\x07\n\x05_archB\x08\n\x06_image\"\xac\x02\n\x17ModuleRuntimeDeployment\x12\x1a\n\x08\x65ndpoint\x18\x01 \x01(\tR\x08\x65ndpoint\x12%\n\x0e\x64\x65ployment_key\x18\x02 \x01(\tR\rdeploymentKey\x12\x39\n\ncreated_at\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x42\n\x0c\x61\x63tivated_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampH\x00R\x0b\x61\x63tivatedAt\x88\x01\x01\x12>\n\x05state\x18\x05 \x01(\x0e\x32(.xyz.block.ftl.schema.v1.DeploymentStateR\x05stateB\x0f\n\r_activated_at\"\xe7\x02\n\x12ModuleRuntimeEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x1c\n\tchangeset\x18\x02 \x01(\tR\tchangeset\x12\x43\n\x04\x62\x61se\x18\x03 \x01(\x0b\x32*.xyz.block.ftl.schema.v1.ModuleRuntimeBaseH\x00R\x04\x62\x61se\x88\x01\x01\x12L\n\x07scaling\x18\x04 \x01(\x0b\x32-.xyz.block.ftl.schema.v1.ModuleRuntimeScalingH\x01R\x07scaling\x88\x01\x01\x12U\n\ndeployment\x18\x05 \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.ModuleRuntimeDeploymentH\x02R\ndeployment\x88\x01\x01\x42\x07\n\x05_baseB\n\n\x08_scalingB\r\n\x0b_deployment\"9\n\x14ModuleRuntimeScaling\x12!\n\x0cmin_replicas\x18\x01 \x01(\x05R\x0bminReplicas\"\x8d\x01\n\x08Optional\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x04type\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeH\x01R\x04type\x88\x01\x01\x42\x06\n\x04_posB\x07\n\x05_type\"R\n\x08Position\x12\x1a\n\x08\x66ilename\x18\x01 \x01(\tR\x08\x66ilename\x12\x12\n\x04line\x18\x02 \x01(\x03R\x04line\x12\x16\n\x06\x63olumn\x18\x03 \x01(\x03R\x06\x63olumn\"\xbb\x01\n\x03Ref\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x16\n\x06module\x18\x03 \x01(\tR\x06module\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x46\n\x0ftype_parameters\x18\x04 \x03(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x0etypeParametersB\x06\n\x04_pos\"\x85\x01\n\x06Schema\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x39\n\x07modules\x18\x02 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modulesB\x06\n\x04_pos\"\x8c\x01\n\x0bSchemaState\x12\x39\n\x07modules\x18\x01 \x03(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ModuleR\x07modules\x12\x42\n\nchangesets\x18\x02 \x03(\x0b\x32\".xyz.block.ftl.schema.v1.ChangesetR\nchangesets\"\xad\x01\n\x06Secret\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x04typeB\x06\n\x04_pos\"J\n\x06String\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"e\n\x0bStringValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x14\n\x05value\x18\x02 \x01(\tR\x05valueB\x06\n\x04_pos\"H\n\x04Time\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xd9\x02\n\x05Topic\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x46\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32%.xyz.block.ftl.schema.v1.TopicRuntimeH\x01R\x07runtime\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x33\n\x05\x65vent\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x05\x65vent\x12=\n\x08metadata\x18\x06 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadataB\x06\n\x04_posB\n\n\x08_runtime\"N\n\x0cTopicRuntime\x12#\n\rkafka_brokers\x18\x01 \x03(\tR\x0ckafkaBrokers\x12\x19\n\x08topic_id\x18\x02 \x01(\tR\x07topicId\"\x9a\x01\n\x11TopicRuntimeEvent\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1c\n\tchangeset\x18\x02 \x01(\tR\tchangeset\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12?\n\x07payload\x18\x04 \x01(\x0b\x32%.xyz.block.ftl.schema.v1.TopicRuntimeR\x07payload\"\x9a\x05\n\x04Type\x12\x30\n\x03\x61ny\x18\t \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.AnyH\x00R\x03\x61ny\x12\x36\n\x05\x61rray\x18\x07 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.ArrayH\x00R\x05\x61rray\x12\x33\n\x04\x62ool\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.BoolH\x00R\x04\x62ool\x12\x36\n\x05\x62ytes\x18\x04 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.BytesH\x00R\x05\x62ytes\x12\x36\n\x05\x66loat\x18\x02 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.FloatH\x00R\x05\x66loat\x12\x30\n\x03int\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.IntH\x00R\x03int\x12\x30\n\x03map\x18\x08 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.MapH\x00R\x03map\x12?\n\x08optional\x18\x0c \x01(\x0b\x32!.xyz.block.ftl.schema.v1.OptionalH\x00R\x08optional\x12\x30\n\x03ref\x18\x0b \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefH\x00R\x03ref\x12\x39\n\x06string\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.StringH\x00R\x06string\x12\x33\n\x04time\x18\x06 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TimeH\x00R\x04time\x12\x33\n\x04unit\x18\n \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.UnitH\x00R\x04unitB\x07\n\x05value\"\x87\x02\n\tTypeAlias\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x31\n\x04type\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x04type\x12=\n\x08metadata\x18\x06 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadataB\x06\n\x04_pos\"e\n\rTypeParameter\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\x06\n\x04_pos\"\x82\x01\n\tTypeValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x33\n\x05value\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x05valueB\x06\n\x04_pos\"H\n\x04Unit\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xe2\x01\n\x05Value\x12@\n\tint_value\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.IntValueH\x00R\x08intValue\x12I\n\x0cstring_value\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.schema.v1.StringValueH\x00R\x0bstringValue\x12\x43\n\ntype_value\x18\x03 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.TypeValueH\x00R\ttypeValueB\x07\n\x05value\"\x96\x03\n\x04Verb\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x37\n\x07request\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x07request\x12\x39\n\x08response\x18\x06 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.TypeR\x08response\x12=\n\x08metadata\x18\x07 \x03(\x0b\x32!.xyz.block.ftl.schema.v1.MetadataR\x08metadata\x12\x45\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32$.xyz.block.ftl.schema.v1.VerbRuntimeH\x01R\x07runtime\x88\x01\x01\x42\x06\n\x04_posB\n\n\x08_runtime\"y\n\x0bVerbRuntime\x12Y\n\x0csubscription\x18\x01 \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.VerbRuntimeSubscriptionH\x00R\x0csubscription\x88\x01\x01\x42\x0f\n\r_subscription\"\xc4\x01\n\x10VerbRuntimeEvent\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x1c\n\tchangeset\x18\x02 \x01(\tR\tchangeset\x12\x0e\n\x02id\x18\x03 \x01(\tR\x02id\x12Y\n\x0csubscription\x18\x04 \x01(\x0b\x32\x30.xyz.block.ftl.schema.v1.VerbRuntimeSubscriptionH\x00R\x0csubscription\x88\x01\x01\x42\x0f\n\r_subscription\">\n\x17VerbRuntimeSubscription\x12#\n\rkafka_brokers\x18\x01 \x03(\tR\x0ckafkaBrokers*<\n\tAliasKind\x12\x1a\n\x16\x41LIAS_KIND_UNSPECIFIED\x10\x00\x12\x13\n\x0f\x41LIAS_KIND_JSON\x10\x01*\x87\x02\n\x0e\x43hangesetState\x12\x1f\n\x1b\x43HANGESET_STATE_UNSPECIFIED\x10\x00\x12\x1d\n\x19\x43HANGESET_STATE_PREPARING\x10\x01\x12\x1c\n\x18\x43HANGESET_STATE_PREPARED\x10\x02\x12\x1d\n\x19\x43HANGESET_STATE_COMMITTED\x10\x03\x12\x1b\n\x17\x43HANGESET_STATE_DRAINED\x10\x04\x12\x1d\n\x19\x43HANGESET_STATE_FINALIZED\x10\x05\x12 \n\x1c\x43HANGESET_STATE_ROLLING_BACK\x10\x06\x12\x1a\n\x16\x43HANGESET_STATE_FAILED\x10\x07*\xaf\x02\n\x0f\x44\x65ploymentState\x12 \n\x1c\x44\x45PLOYMENT_STATE_UNSPECIFIED\x10\x00\x12!\n\x1d\x44\x45PLOYMENT_STATE_PROVISIONING\x10\x01\x12\x1a\n\x16\x44\x45PLOYMENT_STATE_READY\x10\x02\x12\x1b\n\x17\x44\x45PLOYMENT_STATE_CANARY\x10\x03\x12\x1e\n\x1a\x44\x45PLOYMENT_STATE_CANONICAL\x10\x04\x12\x1d\n\x19\x44\x45PLOYMENT_STATE_DRAINING\x10\x05\x12$\n DEPLOYMENT_STATE_DE_PROVISIONING\x10\x06\x12\x1c\n\x18\x44\x45PLOYMENT_STATE_DELETED\x10\x07\x12\x1b\n\x17\x44\x45PLOYMENT_STATE_FAILED\x10\x08*\\\n\nFromOffset\x12\x1b\n\x17\x46ROM_OFFSET_UNSPECIFIED\x10\x00\x12\x19\n\x15\x46ROM_OFFSET_BEGINNING\x10\x01\x12\x16\n\x12\x46ROM_OFFSET_LATEST\x10\x02\x42GP\x01ZCgithub.com/block/ftl/common/protos/xyz/block/ftl/schema/v1;schemapbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.schema.v1.schema_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZCgithub.com/block/ftl/common/protos/xyz/block/ftl/schema/v1;schemapb'
  _globals['_ALIASKIND']._serialized_start=16751
  _globals['_ALIASKIND']._serialized_end=16811
  _globals['_CHANGESETSTATE']._serialized_start=16814
  _globals['_CHANGESETSTATE']._serialized_end=17077
  _globals['_DEPLOYMENTSTATE']._serialized_start=17080
  _globals['_DEPLOYMENTSTATE']._serialized_end=17383
  _globals['_FROMOFFSET']._serialized_start=17385
  _globals['_FROMOFFSET']._serialized_end=17477
  _globals['_AWSIAMAUTHDATABASECONNECTOR']._serialized_start=99
  _globals['_AWSIAMAUTHDATABASECONNECTOR']._serialized_end=278
  _globals['_ANY']._serialized_start=280
  _globals['_ANY']._serialized_end=351
  _globals['_ARRAY']._serialized_start=354
  _globals['_ARRAY']._serialized_end=484
  _globals['_BOOL']._serialized_start=486
  _globals['_BOOL']._serialized_end=558
  _globals['_BYTES']._serialized_start=560
  _globals['_BYTES']._serialized_end=633
  _globals['_CHANGESET']._serialized_start=636
  _globals['_CHANGESET']._serialized_end=988
  _globals['_CHANGESETCOMMITTEDEVENT']._serialized_start=990
  _globals['_CHANGESETCOMMITTEDEVENT']._serialized_end=1033
  _globals['_CHANGESETCREATEDEVENT']._serialized_start=1035
  _globals['_CHANGESETCREATEDEVENT']._serialized_end=1124
  _globals['_CHANGESETDRAINEDEVENT']._serialized_start=1126
  _globals['_CHANGESETDRAINEDEVENT']._serialized_end=1167
  _globals['_CHANGESETFAILEDEVENT']._serialized_start=1169
  _globals['_CHANGESETFAILEDEVENT']._serialized_end=1231
  _globals['_CHANGESETFINALIZEDEVENT']._serialized_start=1233
  _globals['_CHANGESETFINALIZEDEVENT']._serialized_end=1276
  _globals['_CHANGESETPREPAREDEVENT']._serialized_start=1278
  _globals['_CHANGESETPREPAREDEVENT']._serialized_end=1320
  _globals['_CONFIG']._serialized_start=1323
  _globals['_CONFIG']._serialized_end=1496
  _globals['_DSNDATABASECONNECTOR']._serialized_start=1498
  _globals['_DSNDATABASECONNECTOR']._serialized_end=1604
  _globals['_DATA']._serialized_start=1607
  _globals['_DATA']._serialized_end=1951
  _globals['_DATABASE']._serialized_start=1954
  _globals['_DATABASE']._serialized_end=2248
  _globals['_DATABASECONNECTOR']._serialized_start=2251
  _globals['_DATABASECONNECTOR']._serialized_end=2507
  _globals['_DATABASERUNTIME']._serialized_start=2509
  _globals['_DATABASERUNTIME']._serialized_end=2634
  _globals['_DATABASERUNTIMECONNECTIONS']._serialized_start=2637
  _globals['_DATABASERUNTIMECONNECTIONS']._serialized_end=2795
  _globals['_DATABASERUNTIMEEVENT']._serialized_start=2798
  _globals['_DATABASERUNTIMEEVENT']._serialized_end=2977
  _globals['_DECL']._serialized_start=2980
  _globals['_DECL']._serialized_end=3462
  _globals['_DEPLOYMENTACTIVATEDEVENT']._serialized_start=3465
  _globals['_DEPLOYMENTACTIVATEDEVENT']._serialized_end=3637
  _globals['_DEPLOYMENTCREATEDEVENT']._serialized_start=3640
  _globals['_DEPLOYMENTCREATEDEVENT']._serialized_end=3769
  _globals['_DEPLOYMENTDEACTIVATEDEVENT']._serialized_start=3771
  _globals['_DEPLOYMENTDEACTIVATEDEVENT']._serialized_end=3886
  _globals['_DEPLOYMENTREPLICASUPDATEDEVENT']._serialized_start=3888
  _globals['_DEPLOYMENTREPLICASUPDATEDEVENT']._serialized_end=3996
  _globals['_DEPLOYMENTSCHEMAUPDATEDEVENT']._serialized_start=3999
  _globals['_DEPLOYMENTSCHEMAUPDATEDEVENT']._serialized_end=4134
  _globals['_ENUM']._serialized_start=4137
  _globals['_ENUM']._serialized_end=4412
  _globals['_ENUMVARIANT']._serialized_start=4415
  _globals['_ENUMVARIANT']._serialized_end=4596
  _globals['_EVENT']._serialized_start=4599
  _globals['_EVENT']._serialized_end=6256
  _globals['_FIELD']._serialized_start=6259
  _globals['_FIELD']._serialized_end=6494
  _globals['_FLOAT']._serialized_start=6496
  _globals['_FLOAT']._serialized_end=6569
  _globals['_INGRESSPATHCOMPONENT']._serialized_start=6572
  _globals['_INGRESSPATHCOMPONENT']._serialized_end=6803
  _globals['_INGRESSPATHLITERAL']._serialized_start=6805
  _globals['_INGRESSPATHLITERAL']._serialized_end=6911
  _globals['_INGRESSPATHPARAMETER']._serialized_start=6913
  _globals['_INGRESSPATHPARAMETER']._serialized_end=7021
  _globals['_INT']._serialized_start=7023
  _globals['_INT']._serialized_end=7094
  _globals['_INTVALUE']._serialized_start=7096
  _globals['_INTVALUE']._serialized_end=7194
  _globals['_MAP']._serialized_start=7197
  _globals['_MAP']._serialized_end=7370
  _globals['_METADATA']._serialized_start=7373
  _globals['_METADATA']._serialized_end=8626
  _globals['_METADATAALIAS']._serialized_start=8629
  _globals['_METADATAALIAS']._serialized_end=8788
  _globals['_METADATAARTEFACT']._serialized_start=8791
  _globals['_METADATAARTEFACT']._serialized_end=8951
  _globals['_METADATACALLS']._serialized_start=8954
  _globals['_METADATACALLS']._serialized_end=9087
  _globals['_METADATACONFIG']._serialized_start=9090
  _globals['_METADATACONFIG']._serialized_end=9226
  _globals['_METADATACRONJOB']._serialized_start=9228
  _globals['_METADATACRONJOB']._serialized_end=9331
  _globals['_METADATADATABASES']._serialized_start=9334
  _globals['_METADATADATABASES']._serialized_end=9471
  _globals['_METADATAENCODING']._serialized_start=9474
  _globals['_METADATAENCODING']._serialized_end=9604
  _globals['_METADATAINGRESS']._serialized_start=9607
  _globals['_METADATAINGRESS']._serialized_end=9801
  _globals['_METADATAPARTITIONS']._serialized_start=9803
  _globals['_METADATAPARTITIONS']._serialized_end=9921
  _globals['_METADATAPUBLISHER']._serialized_start=9924
  _globals['_METADATAPUBLISHER']._serialized_end=10063
  _globals['_METADATARETRY']._serialized_start=10066
  _globals['_METADATARETRY']._serialized_end=10317
  _globals['_METADATASQLCOLUMN']._serialized_start=10319
  _globals['_METADATASQLCOLUMN']._serialized_end=10446
  _globals['_METADATASQLMIGRATION']._serialized_start=10448
  _globals['_METADATASQLMIGRATION']._serialized_end=10560
  _globals['_METADATASQLQUERY']._serialized_start=10563
  _globals['_METADATASQLQUERY']._serialized_end=10695
  _globals['_METADATASECRETS']._serialized_start=10698
  _globals['_METADATASECRETS']._serialized_end=10837
  _globals['_METADATASUBSCRIBER']._serialized_start=10840
  _globals['_METADATASUBSCRIBER']._serialized_end=11081
  _globals['_METADATATYPEMAP']._serialized_start=11084
  _globals['_METADATATYPEMAP']._serialized_end=11226
  _globals['_MODULE']._serialized_start=11229
  _globals['_MODULE']._serialized_end=11561
  _globals['_MODULERUNTIME']._serialized_start=11564
  _globals['_MODULERUNTIME']._serialized_end=11835
  _globals['_MODULERUNTIMEBASE']._serialized_start=11838
  _globals['_MODULERUNTIMEBASE']._serialized_end=12045
  _globals['_MODULERUNTIMEDEPLOYMENT']._serialized_start=12048
  _globals['_MODULERUNTIMEDEPLOYMENT']._serialized_end=12348
  _globals['_MODULERUNTIMEEVENT']._serialized_start=12351
  _globals['_MODULERUNTIMEEVENT']._serialized_end=12710
  _globals['_MODULERUNTIMESCALING']._serialized_start=12712
  _globals['_MODULERUNTIMESCALING']._serialized_end=12769
  _globals['_OPTIONAL']._serialized_start=12772
  _globals['_OPTIONAL']._serialized_end=12913
  _globals['_POSITION']._serialized_start=12915
  _globals['_POSITION']._serialized_end=12997
  _globals['_REF']._serialized_start=13000
  _globals['_REF']._serialized_end=13187
  _globals['_SCHEMA']._serialized_start=13190
  _globals['_SCHEMA']._serialized_end=13323
  _globals['_SCHEMASTATE']._serialized_start=13326
  _globals['_SCHEMASTATE']._serialized_end=13466
  _globals['_SECRET']._serialized_start=13469
  _globals['_SECRET']._serialized_end=13642
  _globals['_STRING']._serialized_start=13644
  _globals['_STRING']._serialized_end=13718
  _globals['_STRINGVALUE']._serialized_start=13720
  _globals['_STRINGVALUE']._serialized_end=13821
  _globals['_TIME']._serialized_start=13823
  _globals['_TIME']._serialized_end=13895
  _globals['_TOPIC']._serialized_start=13898
  _globals['_TOPIC']._serialized_end=14243
  _globals['_TOPICRUNTIME']._serialized_start=14245
  _globals['_TOPICRUNTIME']._serialized_end=14323
  _globals['_TOPICRUNTIMEEVENT']._serialized_start=14326
  _globals['_TOPICRUNTIMEEVENT']._serialized_end=14480
  _globals['_TYPE']._serialized_start=14483
  _globals['_TYPE']._serialized_end=15149
  _globals['_TYPEALIAS']._serialized_start=15152
  _globals['_TYPEALIAS']._serialized_end=15415
  _globals['_TYPEPARAMETER']._serialized_start=15417
  _globals['_TYPEPARAMETER']._serialized_end=15518
  _globals['_TYPEVALUE']._serialized_start=15521
  _globals['_TYPEVALUE']._serialized_end=15651
  _globals['_UNIT']._serialized_start=15653
  _globals['_UNIT']._serialized_end=15725
  _globals['_VALUE']._serialized_start=15728
  _globals['_VALUE']._serialized_end=15954
  _globals['_VERB']._serialized_start=15957
  _globals['_VERB']._serialized_end=16363
  _globals['_VERBRUNTIME']._serialized_start=16365
  _globals['_VERBRUNTIME']._serialized_end=16486
  _globals['_VERBRUNTIMEEVENT']._serialized_start=16489
  _globals['_VERBRUNTIMEEVENT']._serialized_end=16685
  _globals['_VERBRUNTIMESUBSCRIPTION']._serialized_start=16687
  _globals['_VERBRUNTIMESUBSCRIPTION']._serialized_end=16749
# @@protoc_insertion_point(module_scope)
