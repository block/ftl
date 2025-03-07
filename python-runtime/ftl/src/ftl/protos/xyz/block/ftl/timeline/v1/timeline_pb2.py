# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/timeline/v1/timeline.proto
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
    'xyz/block/ftl/timeline/v1/timeline.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from xyz.block.ftl.timeline.v1 import event_pb2 as xyz_dot_block_dot_ftl_dot_timeline_dot_v1_dot_event__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n(xyz/block/ftl/timeline/v1/timeline.proto\x12\x19xyz.block.ftl.timeline.v1\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a%xyz/block/ftl/timeline/v1/event.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x84\x0f\n\x12GetTimelineRequest\x12N\n\x07\x66ilters\x18\x01 \x03(\x0b\x32\x34.xyz.block.ftl.timeline.v1.GetTimelineRequest.FilterR\x07\x66ilters\x12\x14\n\x05limit\x18\x02 \x01(\x05R\x05limit\x12I\n\x05order\x18\x03 \x01(\x0e\x32\x33.xyz.block.ftl.timeline.v1.GetTimelineRequest.OrderR\x05order\x1aR\n\x0eLogLevelFilter\x12@\n\tlog_level\x18\x01 \x01(\x0e\x32#.xyz.block.ftl.timeline.v1.LogLevelR\x08logLevel\x1a\x34\n\x10\x44\x65ploymentFilter\x12 \n\x0b\x64\x65ployments\x18\x01 \x03(\tR\x0b\x64\x65ployments\x1a\x31\n\x0f\x43hangesetFilter\x12\x1e\n\nchangesets\x18\x01 \x03(\tR\nchangesets\x1a+\n\rRequestFilter\x12\x1a\n\x08requests\x18\x01 \x03(\tR\x08requests\x1aX\n\x0f\x45ventTypeFilter\x12\x45\n\x0b\x65vent_types\x18\x01 \x03(\x0e\x32$.xyz.block.ftl.timeline.v1.EventTypeR\neventTypes\x1a\xaa\x01\n\nTimeFilter\x12>\n\nolder_than\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampH\x00R\tolderThan\x88\x01\x01\x12>\n\nnewer_than\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampH\x01R\tnewerThan\x88\x01\x01\x42\r\n\x0b_older_thanB\r\n\x0b_newer_than\x1as\n\x08IDFilter\x12\"\n\nlower_than\x18\x01 \x01(\x03H\x00R\tlowerThan\x88\x01\x01\x12$\n\x0bhigher_than\x18\x02 \x01(\x03H\x01R\nhigherThan\x88\x01\x01\x42\r\n\x0b_lower_thanB\x0e\n\x0c_higher_than\x1a\x99\x01\n\nCallFilter\x12\x1f\n\x0b\x64\x65st_module\x18\x01 \x01(\tR\ndestModule\x12 \n\tdest_verb\x18\x02 \x01(\tH\x00R\x08\x64\x65stVerb\x88\x01\x01\x12(\n\rsource_module\x18\x03 \x01(\tH\x01R\x0csourceModule\x88\x01\x01\x42\x0c\n\n_dest_verbB\x10\n\x0e_source_module\x1aH\n\x0cModuleFilter\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x17\n\x04verb\x18\x02 \x01(\tH\x00R\x04verb\x88\x01\x01\x42\x07\n\x05_verb\x1a\xb1\x06\n\x06\x46ilter\x12[\n\tlog_level\x18\x01 \x01(\x0b\x32<.xyz.block.ftl.timeline.v1.GetTimelineRequest.LogLevelFilterH\x00R\x08logLevel\x12\x62\n\x0b\x64\x65ployments\x18\x02 \x01(\x0b\x32>.xyz.block.ftl.timeline.v1.GetTimelineRequest.DeploymentFilterH\x00R\x0b\x64\x65ployments\x12Y\n\x08requests\x18\x03 \x01(\x0b\x32;.xyz.block.ftl.timeline.v1.GetTimelineRequest.RequestFilterH\x00R\x08requests\x12`\n\x0b\x65vent_types\x18\x04 \x01(\x0b\x32=.xyz.block.ftl.timeline.v1.GetTimelineRequest.EventTypeFilterH\x00R\neventTypes\x12N\n\x04time\x18\x05 \x01(\x0b\x32\x38.xyz.block.ftl.timeline.v1.GetTimelineRequest.TimeFilterH\x00R\x04time\x12H\n\x02id\x18\x06 \x01(\x0b\x32\x36.xyz.block.ftl.timeline.v1.GetTimelineRequest.IDFilterH\x00R\x02id\x12N\n\x04\x63\x61ll\x18\x07 \x01(\x0b\x32\x38.xyz.block.ftl.timeline.v1.GetTimelineRequest.CallFilterH\x00R\x04\x63\x61ll\x12T\n\x06module\x18\x08 \x01(\x0b\x32:.xyz.block.ftl.timeline.v1.GetTimelineRequest.ModuleFilterH\x00R\x06module\x12_\n\nchangesets\x18\t \x01(\x0b\x32=.xyz.block.ftl.timeline.v1.GetTimelineRequest.ChangesetFilterH\x00R\nchangesetsB\x08\n\x06\x66ilter\"=\n\x05Order\x12\x15\n\x11ORDER_UNSPECIFIED\x10\x00\x12\r\n\tORDER_ASC\x10\x01\x12\x0e\n\nORDER_DESC\x10\x02\"w\n\x13GetTimelineResponse\x12\x38\n\x06\x65vents\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.timeline.v1.EventR\x06\x65vents\x12\x1b\n\x06\x63ursor\x18\x02 \x01(\x03H\x00R\x06\x63ursor\x88\x01\x01\x42\t\n\x07_cursor\"\xb9\x01\n\x15StreamTimelineRequest\x12G\n\x0fupdate_interval\x18\x01 \x01(\x0b\x32\x19.google.protobuf.DurationH\x00R\x0eupdateInterval\x88\x01\x01\x12\x43\n\x05query\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.GetTimelineRequestR\x05queryB\x12\n\x10_update_interval\"R\n\x16StreamTimelineResponse\x12\x38\n\x06\x65vents\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.timeline.v1.EventR\x06\x65vents\"\x89\x08\n\x13\x43reateEventsRequest\x12S\n\x07\x65ntries\x18\x01 \x03(\x0b\x32\x39.xyz.block.ftl.timeline.v1.CreateEventsRequest.EventEntryR\x07\x65ntries\x1a\x9c\x07\n\nEventEntry\x12\x38\n\ttimestamp\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ttimestamp\x12\x37\n\x03log\x18\x02 \x01(\x0b\x32#.xyz.block.ftl.timeline.v1.LogEventH\x00R\x03log\x12:\n\x04\x63\x61ll\x18\x03 \x01(\x0b\x32$.xyz.block.ftl.timeline.v1.CallEventH\x00R\x04\x63\x61ll\x12\x43\n\x07ingress\x18\x04 \x01(\x0b\x32\'.xyz.block.ftl.timeline.v1.IngressEventH\x00R\x07ingress\x12V\n\x0e\x63ron_scheduled\x18\x05 \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.CronScheduledEventH\x00R\rcronScheduled\x12S\n\rasync_execute\x18\x06 \x01(\x0b\x32,.xyz.block.ftl.timeline.v1.AsyncExecuteEventH\x00R\x0c\x61syncExecute\x12V\n\x0epubsub_publish\x18\x07 \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.PubSubPublishEventH\x00R\rpubsubPublish\x12V\n\x0epubsub_consume\x18\x08 \x01(\x0b\x32-.xyz.block.ftl.timeline.v1.PubSubConsumeEventH\x00R\rpubsubConsume\x12_\n\x11\x63hangeset_created\x18\t \x01(\x0b\x32\x30.xyz.block.ftl.timeline.v1.ChangesetCreatedEventH\x00R\x10\x63hangesetCreated\x12o\n\x17\x63hangeset_state_changed\x18\n \x01(\x0b\x32\x35.xyz.block.ftl.timeline.v1.ChangesetStateChangedEventH\x00R\x15\x63hangesetStateChanged\x12\x62\n\x12\x64\x65ployment_runtime\x18\x0b \x01(\x0b\x32\x31.xyz.block.ftl.timeline.v1.DeploymentRuntimeEventH\x00R\x11\x64\x65ploymentRuntimeB\x07\n\x05\x65ntry\"\x16\n\x14\x43reateEventsResponse\"~\n\x16\x44\x65leteOldEventsRequest\x12\x43\n\nevent_type\x18\x01 \x01(\x0e\x32$.xyz.block.ftl.timeline.v1.EventTypeR\teventType\x12\x1f\n\x0b\x61ge_seconds\x18\x02 \x01(\x03R\nageSeconds\">\n\x17\x44\x65leteOldEventsResponse\x12#\n\rdeleted_count\x18\x01 \x01(\x03R\x0c\x64\x65letedCount2\xb8\x04\n\x0fTimelineService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12q\n\x0bGetTimeline\x12-.xyz.block.ftl.timeline.v1.GetTimelineRequest\x1a..xyz.block.ftl.timeline.v1.GetTimelineResponse\"\x03\x90\x02\x01\x12w\n\x0eStreamTimeline\x12\x30.xyz.block.ftl.timeline.v1.StreamTimelineRequest\x1a\x31.xyz.block.ftl.timeline.v1.StreamTimelineResponse0\x01\x12q\n\x0c\x43reateEvents\x12..xyz.block.ftl.timeline.v1.CreateEventsRequest\x1a/.xyz.block.ftl.timeline.v1.CreateEventsResponse\"\x00\x12z\n\x0f\x44\x65leteOldEvents\x12\x31.xyz.block.ftl.timeline.v1.DeleteOldEventsRequest\x1a\x32.xyz.block.ftl.timeline.v1.DeleteOldEventsResponse\"\x00\x42LP\x01ZHgithub.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1;timelinepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.timeline.v1.timeline_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZHgithub.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1;timelinepb'
  _globals['_TIMELINESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_TIMELINESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_TIMELINESERVICE'].methods_by_name['GetTimeline']._loaded_options = None
  _globals['_TIMELINESERVICE'].methods_by_name['GetTimeline']._serialized_options = b'\220\002\001'
  _globals['_GETTIMELINEREQUEST']._serialized_start=204
  _globals['_GETTIMELINEREQUEST']._serialized_end=2128
  _globals['_GETTIMELINEREQUEST_LOGLEVELFILTER']._serialized_start=403
  _globals['_GETTIMELINEREQUEST_LOGLEVELFILTER']._serialized_end=485
  _globals['_GETTIMELINEREQUEST_DEPLOYMENTFILTER']._serialized_start=487
  _globals['_GETTIMELINEREQUEST_DEPLOYMENTFILTER']._serialized_end=539
  _globals['_GETTIMELINEREQUEST_CHANGESETFILTER']._serialized_start=541
  _globals['_GETTIMELINEREQUEST_CHANGESETFILTER']._serialized_end=590
  _globals['_GETTIMELINEREQUEST_REQUESTFILTER']._serialized_start=592
  _globals['_GETTIMELINEREQUEST_REQUESTFILTER']._serialized_end=635
  _globals['_GETTIMELINEREQUEST_EVENTTYPEFILTER']._serialized_start=637
  _globals['_GETTIMELINEREQUEST_EVENTTYPEFILTER']._serialized_end=725
  _globals['_GETTIMELINEREQUEST_TIMEFILTER']._serialized_start=728
  _globals['_GETTIMELINEREQUEST_TIMEFILTER']._serialized_end=898
  _globals['_GETTIMELINEREQUEST_IDFILTER']._serialized_start=900
  _globals['_GETTIMELINEREQUEST_IDFILTER']._serialized_end=1015
  _globals['_GETTIMELINEREQUEST_CALLFILTER']._serialized_start=1018
  _globals['_GETTIMELINEREQUEST_CALLFILTER']._serialized_end=1171
  _globals['_GETTIMELINEREQUEST_MODULEFILTER']._serialized_start=1173
  _globals['_GETTIMELINEREQUEST_MODULEFILTER']._serialized_end=1245
  _globals['_GETTIMELINEREQUEST_FILTER']._serialized_start=1248
  _globals['_GETTIMELINEREQUEST_FILTER']._serialized_end=2065
  _globals['_GETTIMELINEREQUEST_ORDER']._serialized_start=2067
  _globals['_GETTIMELINEREQUEST_ORDER']._serialized_end=2128
  _globals['_GETTIMELINERESPONSE']._serialized_start=2130
  _globals['_GETTIMELINERESPONSE']._serialized_end=2249
  _globals['_STREAMTIMELINEREQUEST']._serialized_start=2252
  _globals['_STREAMTIMELINEREQUEST']._serialized_end=2437
  _globals['_STREAMTIMELINERESPONSE']._serialized_start=2439
  _globals['_STREAMTIMELINERESPONSE']._serialized_end=2521
  _globals['_CREATEEVENTSREQUEST']._serialized_start=2524
  _globals['_CREATEEVENTSREQUEST']._serialized_end=3557
  _globals['_CREATEEVENTSREQUEST_EVENTENTRY']._serialized_start=2633
  _globals['_CREATEEVENTSREQUEST_EVENTENTRY']._serialized_end=3557
  _globals['_CREATEEVENTSRESPONSE']._serialized_start=3559
  _globals['_CREATEEVENTSRESPONSE']._serialized_end=3581
  _globals['_DELETEOLDEVENTSREQUEST']._serialized_start=3583
  _globals['_DELETEOLDEVENTSREQUEST']._serialized_end=3709
  _globals['_DELETEOLDEVENTSRESPONSE']._serialized_start=3711
  _globals['_DELETEOLDEVENTSRESPONSE']._serialized_end=3773
  _globals['_TIMELINESERVICE']._serialized_start=3776
  _globals['_TIMELINESERVICE']._serialized_end=4344
# @@protoc_insertion_point(module_scope)
