// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: xyz/block/ftl/v1/ftl.proto

package ftlv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{0}
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If present, the service is not ready to accept requests and this is the
	// reason.
	NotReady *string `protobuf:"bytes,1,opt,name=not_ready,json=notReady,proto3,oneof" json:"not_ready,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetNotReady() string {
	if x != nil && x.NotReady != nil {
		return *x.NotReady
	}
	return ""
}

type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []*Metadata_Pair `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{2}
}

func (x *Metadata) GetValues() []*Metadata_Pair {
	if x != nil {
		return x.Values
	}
	return nil
}

type Metadata_Pair struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Metadata_Pair) Reset() {
	*x = Metadata_Pair{}
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metadata_Pair) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata_Pair) ProtoMessage() {}

func (x *Metadata_Pair) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata_Pair.ProtoReflect.Descriptor instead.
func (*Metadata_Pair) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{2, 0}
}

func (x *Metadata_Pair) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Metadata_Pair) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_xyz_block_ftl_v1_ftl_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1_ftl_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x22, 0x0d,
	0x0a, 0x0b, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x3e, 0x0a,
	0x0c, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a,
	0x09, 0x6e, 0x6f, 0x74, 0x5f, 0x72, 0x65, 0x61, 0x64, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x08, 0x6e, 0x6f, 0x74, 0x52, 0x65, 0x61, 0x64, 0x79, 0x88, 0x01, 0x01, 0x42,
	0x0c, 0x0a, 0x0a, 0x5f, 0x6e, 0x6f, 0x74, 0x5f, 0x72, 0x65, 0x61, 0x64, 0x79, 0x22, 0x73, 0x0a,
	0x08, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x37, 0x0a, 0x06, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x50, 0x61, 0x69, 0x72, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x1a, 0x2e, 0x0a, 0x04, 0x50, 0x61, 0x69, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x42, 0x3e, 0x50, 0x01, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x66, 0x74, 0x6c,
	0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_v1_ftl_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1_ftl_proto_rawDescData = file_xyz_block_ftl_v1_ftl_proto_rawDesc
)

func file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1_ftl_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1_ftl_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_v1_ftl_proto_rawDescData)
	})
	return file_xyz_block_ftl_v1_ftl_proto_rawDescData
}

var file_xyz_block_ftl_v1_ftl_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_xyz_block_ftl_v1_ftl_proto_goTypes = []any{
	(*PingRequest)(nil),   // 0: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),  // 1: xyz.block.ftl.v1.PingResponse
	(*Metadata)(nil),      // 2: xyz.block.ftl.v1.Metadata
	(*Metadata_Pair)(nil), // 3: xyz.block.ftl.v1.Metadata.Pair
}
var file_xyz_block_ftl_v1_ftl_proto_depIdxs = []int32{
	3, // 0: xyz.block.ftl.v1.Metadata.values:type_name -> xyz.block.ftl.v1.Metadata.Pair
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_ftl_proto_init() }
func file_xyz_block_ftl_v1_ftl_proto_init() {
	if File_xyz_block_ftl_v1_ftl_proto != nil {
		return
	}
	file_xyz_block_ftl_v1_ftl_proto_msgTypes[1].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1_ftl_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_xyz_block_ftl_v1_ftl_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1_ftl_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_v1_ftl_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1_ftl_proto = out.File
	file_xyz_block_ftl_v1_ftl_proto_rawDesc = nil
	file_xyz_block_ftl_v1_ftl_proto_goTypes = nil
	file_xyz_block_ftl_v1_ftl_proto_depIdxs = nil
}
