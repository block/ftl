// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: xyz/block/ftl/pubsub/v1/publish.proto

package pubsubpb

import (
	v11 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	v1 "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PublishEventRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Topic *v1.Ref                `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Body  []byte                 `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	Key   string                 `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	// Only verb name is included because this verb will be in the same module as topic
	Caller        string `protobuf:"bytes,4,opt,name=caller,proto3" json:"caller,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PublishEventRequest) Reset() {
	*x = PublishEventRequest{}
	mi := &file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventRequest) ProtoMessage() {}

func (x *PublishEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventRequest.ProtoReflect.Descriptor instead.
func (*PublishEventRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescGZIP(), []int{0}
}

func (x *PublishEventRequest) GetTopic() *v1.Ref {
	if x != nil {
		return x.Topic
	}
	return nil
}

func (x *PublishEventRequest) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *PublishEventRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *PublishEventRequest) GetCaller() string {
	if x != nil {
		return x.Caller
	}
	return ""
}

type PublishEventResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PublishEventResponse) Reset() {
	*x = PublishEventResponse{}
	mi := &file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventResponse) ProtoMessage() {}

func (x *PublishEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventResponse.ProtoReflect.Descriptor instead.
func (*PublishEventResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescGZIP(), []int{1}
}

var File_xyz_block_ftl_pubsub_v1_publish_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_pubsub_v1_publish_proto_rawDesc = string([]byte{
	0x0a, 0x25, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73,
	0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x76, 0x31,
	0x1a, 0x24, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x87, 0x01, 0x0a, 0x13, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x32, 0x0a, 0x05, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x66, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x12,
	0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f,
	0x64, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x22, 0x16, 0x0a, 0x14,
	0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x32, 0xc9, 0x01, 0x0a, 0x0e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12,
	0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03,
	0x90, 0x02, 0x01, 0x12, 0x6b, 0x0a, 0x0c, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x2c, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75,
	0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x2d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x48, 0x50, 0x01, 0x5a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2f, 0x76,
	0x31, 0x3b, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var (
	file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescData []byte
)

func file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_pubsub_v1_publish_proto_rawDesc), len(file_xyz_block_ftl_pubsub_v1_publish_proto_rawDesc)))
	})
	return file_xyz_block_ftl_pubsub_v1_publish_proto_rawDescData
}

var file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_xyz_block_ftl_pubsub_v1_publish_proto_goTypes = []any{
	(*PublishEventRequest)(nil),  // 0: xyz.block.ftl.pubsub.v1.PublishEventRequest
	(*PublishEventResponse)(nil), // 1: xyz.block.ftl.pubsub.v1.PublishEventResponse
	(*v1.Ref)(nil),               // 2: xyz.block.ftl.schema.v1.Ref
	(*v11.PingRequest)(nil),      // 3: xyz.block.ftl.v1.PingRequest
	(*v11.PingResponse)(nil),     // 4: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_pubsub_v1_publish_proto_depIdxs = []int32{
	2, // 0: xyz.block.ftl.pubsub.v1.PublishEventRequest.topic:type_name -> xyz.block.ftl.schema.v1.Ref
	3, // 1: xyz.block.ftl.pubsub.v1.PublishService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	0, // 2: xyz.block.ftl.pubsub.v1.PublishService.PublishEvent:input_type -> xyz.block.ftl.pubsub.v1.PublishEventRequest
	4, // 3: xyz.block.ftl.pubsub.v1.PublishService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	1, // 4: xyz.block.ftl.pubsub.v1.PublishService.PublishEvent:output_type -> xyz.block.ftl.pubsub.v1.PublishEventResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_pubsub_v1_publish_proto_init() }
func file_xyz_block_ftl_pubsub_v1_publish_proto_init() {
	if File_xyz_block_ftl_pubsub_v1_publish_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_pubsub_v1_publish_proto_rawDesc), len(file_xyz_block_ftl_pubsub_v1_publish_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_pubsub_v1_publish_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_pubsub_v1_publish_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_pubsub_v1_publish_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_pubsub_v1_publish_proto = out.File
	file_xyz_block_ftl_pubsub_v1_publish_proto_goTypes = nil
	file_xyz_block_ftl_pubsub_v1_publish_proto_depIdxs = nil
}
