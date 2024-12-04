# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/timeline/v1/event.proto
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
    'xyz/block/ftl/timeline/v1/event.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n%xyz/block/ftl/timeline/v1/event.proto\x12\x19xyz.block.ftl.timeline.v1\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\"\xb7\x03\n\x08LogEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12\x39\n\ntime_stamp\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x1b\n\tlog_level\x18\x04 \x01(\x05R\x08logLevel\x12S\n\nattributes\x18\x05 \x03(\x0b\x32\x33.xyz.block.ftl.timeline.v1.LogEvent.AttributesEntryR\nattributes\x12\x18\n\x07message\x18\x06 \x01(\tR\x07message\x12\x19\n\x05\x65rror\x18\x07 \x01(\tH\x01R\x05\x65rror\x88\x01\x01\x12\x19\n\x05stack\x18\x08 \x01(\tH\x02R\x05stack\x88\x01\x01\x1a=\n\x0f\x41ttributesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x42\x0e\n\x0c_request_keyB\x08\n\x06_errorB\x08\n\x06_stack\"\x95\x04\n\tCallEvent\x12$\n\x0brequest_key\x18\x01 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12%\n\x0e\x64\x65ployment_key\x18\x02 \x01(\tR\rdeploymentKey\x12\x39\n\ntime_stamp\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12I\n\x0fsource_verb_ref\x18\x0b \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefH\x01R\rsourceVerbRef\x88\x01\x01\x12N\n\x14\x64\x65stination_verb_ref\x18\x0c \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x12\x64\x65stinationVerbRef\x12\x35\n\x08\x64uration\x18\x06 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12\x18\n\x07request\x18\x07 \x01(\tR\x07request\x12\x1a\n\x08response\x18\x08 \x01(\tR\x08response\x12\x19\n\x05\x65rror\x18\t \x01(\tH\x02R\x05\x65rror\x88\x01\x01\x12\x19\n\x05stack\x18\n \x01(\tH\x03R\x05stack\x88\x01\x01\x42\x0e\n\x0c_request_keyB\x12\n\x10_source_verb_refB\x08\n\x06_errorB\x08\n\x06_stackJ\x04\x08\x04\x10\x05J\x04\x08\x05\x10\x06\"\xb8\x01\n\x16\x44\x65ploymentCreatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x1a\n\x08language\x18\x02 \x01(\tR\x08language\x12\x1f\n\x0bmodule_name\x18\x03 \x01(\tR\nmoduleName\x12!\n\x0cmin_replicas\x18\x04 \x01(\x05R\x0bminReplicas\x12\x1f\n\x08replaced\x18\x05 \x01(\tH\x00R\x08replaced\x88\x01\x01\x42\x0b\n\t_replaced\"y\n\x16\x44\x65ploymentUpdatedEvent\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12!\n\x0cmin_replicas\x18\x02 \x01(\x05R\x0bminReplicas\x12*\n\x11prev_min_replicas\x18\x03 \x01(\x05R\x0fprevMinReplicas\"\x8e\x04\n\x0cIngressEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12\x37\n\x08verb_ref\x18\x03 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x07verbRef\x12\x16\n\x06method\x18\x04 \x01(\tR\x06method\x12\x12\n\x04path\x18\x05 \x01(\tR\x04path\x12\x1f\n\x0bstatus_code\x18\x07 \x01(\x05R\nstatusCode\x12\x39\n\ntime_stamp\x18\x08 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x35\n\x08\x64uration\x18\t \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12\x18\n\x07request\x18\n \x01(\tR\x07request\x12%\n\x0erequest_header\x18\x0b \x01(\tR\rrequestHeader\x12\x1a\n\x08response\x18\x0c \x01(\tR\x08response\x12\'\n\x0fresponse_header\x18\r \x01(\tR\x0eresponseHeader\x12\x19\n\x05\x65rror\x18\x0e \x01(\tH\x01R\x05\x65rror\x88\x01\x01\x42\x0e\n\x0c_request_keyB\x08\n\x06_error\"\xe6\x02\n\x12\x43ronScheduledEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12\x37\n\x08verb_ref\x18\x02 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x07verbRef\x12\x39\n\ntime_stamp\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x35\n\x08\x64uration\x18\x04 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12=\n\x0cscheduled_at\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x0bscheduledAt\x12\x1a\n\x08schedule\x18\x06 \x01(\tR\x08schedule\x12\x19\n\x05\x65rror\x18\x07 \x01(\tH\x00R\x05\x65rror\x88\x01\x01\x42\x08\n\x06_error\"\x9c\x03\n\x11\x41syncExecuteEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12\x37\n\x08verb_ref\x18\x03 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x07verbRef\x12\x39\n\ntime_stamp\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x35\n\x08\x64uration\x18\x05 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12Z\n\x10\x61sync_event_type\x18\x06 \x01(\x0e\x32\x30.xyz.block.ftl.timeline.v1.AsyncExecuteEventTypeR\x0e\x61syncEventType\x12\x19\n\x05\x65rror\x18\x07 \x01(\tH\x01R\x05\x65rror\x88\x01\x01\x42\x0e\n\x0c_request_keyB\x08\n\x06_error\"\xf1\x02\n\x12PubSubPublishEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12\x37\n\x08verb_ref\x18\x03 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x07verbRef\x12\x39\n\ntime_stamp\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x35\n\x08\x64uration\x18\x05 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12\x14\n\x05topic\x18\x06 \x01(\tR\x05topic\x12\x18\n\x07request\x18\x07 \x01(\tR\x07request\x12\x19\n\x05\x65rror\x18\x08 \x01(\tH\x01R\x05\x65rror\x88\x01\x01\x42\x0e\n\x0c_request_keyB\x08\n\x06_error\"\xa0\x03\n\x12PubSubConsumeEvent\x12%\n\x0e\x64\x65ployment_key\x18\x01 \x01(\tR\rdeploymentKey\x12$\n\x0brequest_key\x18\x02 \x01(\tH\x00R\nrequestKey\x88\x01\x01\x12-\n\x10\x64\x65st_verb_module\x18\x03 \x01(\tH\x01R\x0e\x64\x65stVerbModule\x88\x01\x01\x12)\n\x0e\x64\x65st_verb_name\x18\x04 \x01(\tH\x02R\x0c\x64\x65stVerbName\x88\x01\x01\x12\x39\n\ntime_stamp\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x35\n\x08\x64uration\x18\x06 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12\x14\n\x05topic\x18\x07 \x01(\tR\x05topic\x12\x19\n\x05\x65rror\x18\x08 \x01(\tH\x03R\x05\x65rror\x88\x01\x01\x42\x0e\n\x0c_request_keyB\x13\n\x11_dest_verb_moduleB\x11\n\x0f_dest_verb_nameB\x08\n\x06_error\"\xba\x06\n\x05\x45vent\x12\x39\n\ntime_stamp\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimeStamp\x12\x0e\n\x02id\x18\x02 \x01(\x03R\x02id\x12\x37\n\x03log\x18\x03 \x01(\x0b\x32#.xyz.block.ftl.timeline.v1.LogEventH\x00R\x03log\x12:\n\x04\x63\x61ll\x18\x04 \x01(\x0b\x32$.xyz.block.ftl.timeline.v1.CallEventH\x00R\x04\x63\x61ll\x12\x62\n\x12\x64\x65ployment_created\x18\x05 \x01(\x0b\x32\x31.xyz.block.ftl.timeline.v1.DeploymentCreatedEventH\x00R\x11\x64\x65ploymentCreated\x12\x62\n\x12\x64\x65ployment_updated\x18\x06 \x01(\x0b\x32\x31.xyz.block.ftl.timeline.v1.DeploymentUpdatedEventH\x00R\x11\x64\x65ploymentUpdated\x12\x43\n\x07ingress\x18\x07 \x01(\x0b\x32\'.xyz.block.ftl.timeline.v1.IngressEventH\x00R\x07ingress\x12V\n\x0e\x63ron_scheduled\x18\x08 \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.CronScheduledEventH\x00R\rcronScheduled\x12S\n\rasync_execute\x18\t \x01(\x0b\x32,.xyz.block.ftl.timeline.v1.AsyncExecuteEventH\x00R\x0c\x61syncExecute\x12V\n\x0epubsub_publish\x18\n \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.PubSubPublishEventH\x00R\rpubsubPublish\x12V\n\x0epubsub_consume\x18\x0b \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.PubSubConsumeEventH\x00R\rpubsubConsumeB\x07\n\x05\x65ntry*\xa9\x02\n\tEventType\x12\x1a\n\x16\x45VENT_TYPE_UNSPECIFIED\x10\x00\x12\x12\n\x0e\x45VENT_TYPE_LOG\x10\x01\x12\x13\n\x0f\x45VENT_TYPE_CALL\x10\x02\x12!\n\x1d\x45VENT_TYPE_DEPLOYMENT_CREATED\x10\x03\x12!\n\x1d\x45VENT_TYPE_DEPLOYMENT_UPDATED\x10\x04\x12\x16\n\x12\x45VENT_TYPE_INGRESS\x10\x05\x12\x1d\n\x19\x45VENT_TYPE_CRON_SCHEDULED\x10\x06\x12\x1c\n\x18\x45VENT_TYPE_ASYNC_EXECUTE\x10\x07\x12\x1d\n\x19\x45VENT_TYPE_PUBSUB_PUBLISH\x10\x08\x12\x1d\n\x19\x45VENT_TYPE_PUBSUB_CONSUME\x10\t*\x89\x01\n\x15\x41syncExecuteEventType\x12(\n$ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED\x10\x00\x12!\n\x1d\x41SYNC_EXECUTE_EVENT_TYPE_CRON\x10\x01\x12#\n\x1f\x41SYNC_EXECUTE_EVENT_TYPE_PUBSUB\x10\x02*\x8c\x01\n\x08LogLevel\x12\x19\n\x15LOG_LEVEL_UNSPECIFIED\x10\x00\x12\x13\n\x0fLOG_LEVEL_TRACE\x10\x01\x12\x13\n\x0fLOG_LEVEL_DEBUG\x10\x05\x12\x12\n\x0eLOG_LEVEL_INFO\x10\t\x12\x12\n\x0eLOG_LEVEL_WARN\x10\r\x12\x13\n\x0fLOG_LEVEL_ERROR\x10\x11\x42RP\x01ZNgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1;timelinev1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.timeline.v1.event_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZNgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1;timelinev1'
  _globals['_LOGEVENT_ATTRIBUTESENTRY']._loaded_options = None
  _globals['_LOGEVENT_ATTRIBUTESENTRY']._serialized_options = b'8\001'
  _globals['_EVENTTYPE']._serialized_start=4385
  _globals['_EVENTTYPE']._serialized_end=4682
  _globals['_ASYNCEXECUTEEVENTTYPE']._serialized_start=4685
  _globals['_ASYNCEXECUTEEVENTTYPE']._serialized_end=4822
  _globals['_LOGLEVEL']._serialized_start=4825
  _globals['_LOGLEVEL']._serialized_end=4965
  _globals['_LOGEVENT']._serialized_start=172
  _globals['_LOGEVENT']._serialized_end=611
  _globals['_LOGEVENT_ATTRIBUTESENTRY']._serialized_start=514
  _globals['_LOGEVENT_ATTRIBUTESENTRY']._serialized_end=575
  _globals['_CALLEVENT']._serialized_start=614
  _globals['_CALLEVENT']._serialized_end=1147
  _globals['_DEPLOYMENTCREATEDEVENT']._serialized_start=1150
  _globals['_DEPLOYMENTCREATEDEVENT']._serialized_end=1334
  _globals['_DEPLOYMENTUPDATEDEVENT']._serialized_start=1336
  _globals['_DEPLOYMENTUPDATEDEVENT']._serialized_end=1457
  _globals['_INGRESSEVENT']._serialized_start=1460
  _globals['_INGRESSEVENT']._serialized_end=1986
  _globals['_CRONSCHEDULEDEVENT']._serialized_start=1989
  _globals['_CRONSCHEDULEDEVENT']._serialized_end=2347
  _globals['_ASYNCEXECUTEEVENT']._serialized_start=2350
  _globals['_ASYNCEXECUTEEVENT']._serialized_end=2762
  _globals['_PUBSUBPUBLISHEVENT']._serialized_start=2765
  _globals['_PUBSUBPUBLISHEVENT']._serialized_end=3134
  _globals['_PUBSUBCONSUMEEVENT']._serialized_start=3137
  _globals['_PUBSUBCONSUMEEVENT']._serialized_end=3553
  _globals['_EVENT']._serialized_start=3556
  _globals['_EVENT']._serialized_end=4382
# @@protoc_insertion_point(module_scope)
