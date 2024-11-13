// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: xyz/block/ftl/v1/verb.proto

package ftlv1

import (
	schema "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
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

type CallRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *Metadata   `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Verb     *schema.Ref `protobuf:"bytes,2,opt,name=verb,proto3" json:"verb,omitempty"`
	Body     []byte      `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
}

func (x *CallRequest) Reset() {
	*x = CallRequest{}
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CallRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallRequest) ProtoMessage() {}

func (x *CallRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallRequest.ProtoReflect.Descriptor instead.
func (*CallRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_verb_proto_rawDescGZIP(), []int{0}
}

func (x *CallRequest) GetMetadata() *Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *CallRequest) GetVerb() *schema.Ref {
	if x != nil {
		return x.Verb
	}
	return nil
}

func (x *CallRequest) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

type CallResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*CallResponse_Body
	//	*CallResponse_Error_
	Response isCallResponse_Response `protobuf_oneof:"response"`
}

func (x *CallResponse) Reset() {
	*x = CallResponse{}
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CallResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallResponse) ProtoMessage() {}

func (x *CallResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallResponse.ProtoReflect.Descriptor instead.
func (*CallResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_verb_proto_rawDescGZIP(), []int{1}
}

func (m *CallResponse) GetResponse() isCallResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *CallResponse) GetBody() []byte {
	if x, ok := x.GetResponse().(*CallResponse_Body); ok {
		return x.Body
	}
	return nil
}

func (x *CallResponse) GetError() *CallResponse_Error {
	if x, ok := x.GetResponse().(*CallResponse_Error_); ok {
		return x.Error
	}
	return nil
}

type isCallResponse_Response interface {
	isCallResponse_Response()
}

type CallResponse_Body struct {
	Body []byte `protobuf:"bytes,1,opt,name=body,proto3,oneof"`
}

type CallResponse_Error_ struct {
	Error *CallResponse_Error `protobuf:"bytes,2,opt,name=error,proto3,oneof"`
}

func (*CallResponse_Body) isCallResponse_Response() {}

func (*CallResponse_Error_) isCallResponse_Response() {}

type CallResponse_Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string  `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Stack   *string `protobuf:"bytes,2,opt,name=stack,proto3,oneof" json:"stack,omitempty"` // TODO: Richer error type.
}

func (x *CallResponse_Error) Reset() {
	*x = CallResponse_Error{}
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CallResponse_Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallResponse_Error) ProtoMessage() {}

func (x *CallResponse_Error) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_verb_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallResponse_Error.ProtoReflect.Descriptor instead.
func (*CallResponse_Error) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_verb_proto_rawDescGZIP(), []int{1, 0}
}

func (x *CallResponse_Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *CallResponse_Error) GetStack() string {
	if x != nil && x.Stack != nil {
		return *x.Stack
	}
	return ""
}

var File_xyz_block_ftl_v1_verb_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1_verb_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x76, 0x65, 0x72, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x1a,
	0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76,
	0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x78, 0x79, 0x7a,
	0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x8b, 0x01, 0x0a, 0x0b, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x36, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x30, 0x0a, 0x04, 0x76, 0x65, 0x72,
	0x62, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x2e, 0x52, 0x65, 0x66, 0x52, 0x04, 0x76, 0x65, 0x72, 0x62, 0x12, 0x12, 0x0a, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x22,
	0xb6, 0x01, 0x0a, 0x0c, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x14, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00,
	0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x3c, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x1a, 0x46, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x88,
	0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x42, 0x0a, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xa0, 0x01, 0x0a, 0x0b, 0x56, 0x65, 0x72,
	0x62, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67,
	0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x03, 0x90, 0x02, 0x01, 0x12, 0x45, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1d, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43,
	0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x44, 0x50, 0x01, 0x5a,
	0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42, 0x44, 0x35,
	0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x66, 0x74, 0x6c, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_v1_verb_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1_verb_proto_rawDescData = file_xyz_block_ftl_v1_verb_proto_rawDesc
)

func file_xyz_block_ftl_v1_verb_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1_verb_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1_verb_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_v1_verb_proto_rawDescData)
	})
	return file_xyz_block_ftl_v1_verb_proto_rawDescData
}

var file_xyz_block_ftl_v1_verb_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_xyz_block_ftl_v1_verb_proto_goTypes = []any{
	(*CallRequest)(nil),        // 0: xyz.block.ftl.v1.CallRequest
	(*CallResponse)(nil),       // 1: xyz.block.ftl.v1.CallResponse
	(*CallResponse_Error)(nil), // 2: xyz.block.ftl.v1.CallResponse.Error
	(*Metadata)(nil),           // 3: xyz.block.ftl.v1.Metadata
	(*schema.Ref)(nil),         // 4: xyz.block.ftl.v1.schema.Ref
	(*PingRequest)(nil),        // 5: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),       // 6: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_v1_verb_proto_depIdxs = []int32{
	3, // 0: xyz.block.ftl.v1.CallRequest.metadata:type_name -> xyz.block.ftl.v1.Metadata
	4, // 1: xyz.block.ftl.v1.CallRequest.verb:type_name -> xyz.block.ftl.v1.schema.Ref
	2, // 2: xyz.block.ftl.v1.CallResponse.error:type_name -> xyz.block.ftl.v1.CallResponse.Error
	5, // 3: xyz.block.ftl.v1.VerbService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	0, // 4: xyz.block.ftl.v1.VerbService.Call:input_type -> xyz.block.ftl.v1.CallRequest
	6, // 5: xyz.block.ftl.v1.VerbService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	1, // 6: xyz.block.ftl.v1.VerbService.Call:output_type -> xyz.block.ftl.v1.CallResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_verb_proto_init() }
func file_xyz_block_ftl_v1_verb_proto_init() {
	if File_xyz_block_ftl_v1_verb_proto != nil {
		return
	}
	file_xyz_block_ftl_v1_ftl_proto_init()
	file_xyz_block_ftl_v1_verb_proto_msgTypes[1].OneofWrappers = []any{
		(*CallResponse_Body)(nil),
		(*CallResponse_Error_)(nil),
	}
	file_xyz_block_ftl_v1_verb_proto_msgTypes[2].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1_verb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_v1_verb_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1_verb_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_v1_verb_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1_verb_proto = out.File
	file_xyz_block_ftl_v1_verb_proto_rawDesc = nil
	file_xyz_block_ftl_v1_verb_proto_goTypes = nil
	file_xyz_block_ftl_v1_verb_proto_depIdxs = nil
}
