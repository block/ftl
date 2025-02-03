// Code generated by go2proto. DO NOT EDIT.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: xyz/block/ftl/cron/v1/cron.proto

package cronpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
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

// CronState is the state of scheduled cron jobs
type CronState struct {
	state          protoimpl.MessageState            `protogen:"open.v1"`
	LastExecutions map[string]*timestamppb.Timestamp `protobuf:"bytes,1,rep,name=last_executions,json=lastExecutions,proto3" json:"last_executions,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	NextExecutions map[string]*timestamppb.Timestamp `protobuf:"bytes,2,rep,name=next_executions,json=nextExecutions,proto3" json:"next_executions,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *CronState) Reset() {
	*x = CronState{}
	mi := &file_xyz_block_ftl_cron_v1_cron_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CronState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CronState) ProtoMessage() {}

func (x *CronState) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_cron_v1_cron_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CronState.ProtoReflect.Descriptor instead.
func (*CronState) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_cron_v1_cron_proto_rawDescGZIP(), []int{0}
}

func (x *CronState) GetLastExecutions() map[string]*timestamppb.Timestamp {
	if x != nil {
		return x.LastExecutions
	}
	return nil
}

func (x *CronState) GetNextExecutions() map[string]*timestamppb.Timestamp {
	if x != nil {
		return x.NextExecutions
	}
	return nil
}

var File_xyz_block_ftl_cron_v1_cron_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_cron_v1_cron_proto_rawDesc = string([]byte{
	0x0a, 0x20, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x63, 0x72, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x15, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x63, 0x72, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x87, 0x03, 0x0a, 0x09, 0x43,
	0x72, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x5d, 0x0a, 0x0f, 0x6c, 0x61, 0x73, 0x74,
	0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x34, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x63, 0x72, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x2e, 0x4c, 0x61, 0x73, 0x74, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x5d, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f,
	0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x34, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x63, 0x72, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x6f, 0x6e, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x4e, 0x65, 0x78, 0x74, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0e, 0x6e, 0x65, 0x78, 0x74, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x5d, 0x0a, 0x13, 0x4c, 0x61, 0x73, 0x74, 0x45, 0x78,
	0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x30, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x5d, 0x0a, 0x13, 0x4e, 0x65, 0x78, 0x74, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x30,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x63, 0x72, 0x6f, 0x6e, 0x2f, 0x76,
	0x31, 0x3b, 0x63, 0x72, 0x6f, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_xyz_block_ftl_cron_v1_cron_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_cron_v1_cron_proto_rawDescData []byte
)

func file_xyz_block_ftl_cron_v1_cron_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_cron_v1_cron_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_cron_v1_cron_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_cron_v1_cron_proto_rawDesc), len(file_xyz_block_ftl_cron_v1_cron_proto_rawDesc)))
	})
	return file_xyz_block_ftl_cron_v1_cron_proto_rawDescData
}

var file_xyz_block_ftl_cron_v1_cron_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_xyz_block_ftl_cron_v1_cron_proto_goTypes = []any{
	(*CronState)(nil),             // 0: xyz.block.ftl.cron.v1.CronState
	nil,                           // 1: xyz.block.ftl.cron.v1.CronState.LastExecutionsEntry
	nil,                           // 2: xyz.block.ftl.cron.v1.CronState.NextExecutionsEntry
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_xyz_block_ftl_cron_v1_cron_proto_depIdxs = []int32{
	1, // 0: xyz.block.ftl.cron.v1.CronState.last_executions:type_name -> xyz.block.ftl.cron.v1.CronState.LastExecutionsEntry
	2, // 1: xyz.block.ftl.cron.v1.CronState.next_executions:type_name -> xyz.block.ftl.cron.v1.CronState.NextExecutionsEntry
	3, // 2: xyz.block.ftl.cron.v1.CronState.LastExecutionsEntry.value:type_name -> google.protobuf.Timestamp
	3, // 3: xyz.block.ftl.cron.v1.CronState.NextExecutionsEntry.value:type_name -> google.protobuf.Timestamp
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_cron_v1_cron_proto_init() }
func file_xyz_block_ftl_cron_v1_cron_proto_init() {
	if File_xyz_block_ftl_cron_v1_cron_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_cron_v1_cron_proto_rawDesc), len(file_xyz_block_ftl_cron_v1_cron_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_xyz_block_ftl_cron_v1_cron_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_cron_v1_cron_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_cron_v1_cron_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_cron_v1_cron_proto = out.File
	file_xyz_block_ftl_cron_v1_cron_proto_goTypes = nil
	file_xyz_block_ftl_cron_v1_cron_proto_depIdxs = nil
}
