// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
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
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
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
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
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

type FileChangeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *FileChangeRequest) Reset() {
	*x = FileChangeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileChangeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChangeRequest) ProtoMessage() {}

func (x *FileChangeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChangeRequest.ProtoReflect.Descriptor instead.
func (*FileChangeRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{2}
}

func (x *FileChangeRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type FileChangeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FileChangeResponse) Reset() {
	*x = FileChangeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileChangeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileChangeResponse) ProtoMessage() {}

func (x *FileChangeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileChangeResponse.ProtoReflect.Descriptor instead.
func (*FileChangeResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{3}
}

type CallRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Verb string `protobuf:"bytes,1,opt,name=verb,proto3" json:"verb,omitempty"`
	Body []byte `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
}

func (x *CallRequest) Reset() {
	*x = CallRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallRequest) ProtoMessage() {}

func (x *CallRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{4}
}

func (x *CallRequest) GetVerb() string {
	if x != nil {
		return x.Verb
	}
	return ""
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
	//	*CallResponse_Body
	//	*CallResponse_Error_
	Response isCallResponse_Response `protobuf_oneof:"response"`
}

func (x *CallResponse) Reset() {
	*x = CallResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallResponse) ProtoMessage() {}

func (x *CallResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{5}
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

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRequest.ProtoReflect.Descriptor instead.
func (*ListRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{6}
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Verbs []string `protobuf:"bytes,1,rep,name=verbs,proto3" json:"verbs,omitempty"`
}

func (x *ListResponse) Reset() {
	*x = ListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListResponse) ProtoMessage() {}

func (x *ListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListResponse.ProtoReflect.Descriptor instead.
func (*ListResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{7}
}

func (x *ListResponse) GetVerbs() []string {
	if x != nil {
		return x.Verbs
	}
	return nil
}

type ServeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to local module to serve.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *ServeRequest) Reset() {
	*x = ServeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServeRequest) ProtoMessage() {}

func (x *ServeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServeRequest.ProtoReflect.Descriptor instead.
func (*ServeRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{8}
}

func (x *ServeRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type ServeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ServeResponse) Reset() {
	*x = ServeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServeResponse) ProtoMessage() {}

func (x *ServeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServeResponse.ProtoReflect.Descriptor instead.
func (*ServeResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{9}
}

type CallResponse_Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"` // TODO: Richer error type.
}

func (x *CallResponse_Error) Reset() {
	*x = CallResponse_Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallResponse_Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallResponse_Error) ProtoMessage() {}

func (x *CallResponse_Error) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_ftl_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
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
	return file_xyz_block_ftl_v1_ftl_proto_rawDescGZIP(), []int{5, 0}
}

func (x *CallResponse_Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_xyz_block_ftl_v1_ftl_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1_ftl_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x22, 0x0d,
	0x0a, 0x0b, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0e, 0x0a,
	0x0c, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x27, 0x0a,
	0x11, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0x14, 0x0a, 0x12, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x35, 0x0a, 0x0b,
	0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x76,
	0x65, 0x72, 0x62, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x76, 0x65, 0x72, 0x62, 0x12,
	0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x22, 0x91, 0x01, 0x0a, 0x0c, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x48, 0x00, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x3c, 0x0a, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6c,
	0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x48,
	0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x1a, 0x21, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x0d, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x24, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x65, 0x72, 0x62, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x76, 0x65, 0x72, 0x62, 0x73, 0x22, 0x22, 0x0a, 0x0c,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x22, 0x0f, 0x0a, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x32, 0xad, 0x02, 0x0a, 0x0c, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x48, 0x0a, 0x05, 0x53, 0x65, 0x72, 0x76, 0x65, 0x12, 0x1e, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x04,
	0x50, 0x69, 0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1d, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43,
	0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61,
	0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x32, 0xbc, 0x02, 0x0a, 0x0c, 0x44, 0x72, 0x69, 0x76, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x57, 0x0a, 0x0a, 0x46, 0x69, 0x6c,
	0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x23, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x43,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61,
	0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6c,
	0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x4c, 0x69, 0x73,
	0x74, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x40, 0x50, 0x01, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x54, 0x42, 0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x78, 0x79, 0x7a, 0x2f,
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

var file_xyz_block_ftl_v1_ftl_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_xyz_block_ftl_v1_ftl_proto_goTypes = []interface{}{
	(*PingRequest)(nil),        // 0: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),       // 1: xyz.block.ftl.v1.PingResponse
	(*FileChangeRequest)(nil),  // 2: xyz.block.ftl.v1.FileChangeRequest
	(*FileChangeResponse)(nil), // 3: xyz.block.ftl.v1.FileChangeResponse
	(*CallRequest)(nil),        // 4: xyz.block.ftl.v1.CallRequest
	(*CallResponse)(nil),       // 5: xyz.block.ftl.v1.CallResponse
	(*ListRequest)(nil),        // 6: xyz.block.ftl.v1.ListRequest
	(*ListResponse)(nil),       // 7: xyz.block.ftl.v1.ListResponse
	(*ServeRequest)(nil),       // 8: xyz.block.ftl.v1.ServeRequest
	(*ServeResponse)(nil),      // 9: xyz.block.ftl.v1.ServeResponse
	(*CallResponse_Error)(nil), // 10: xyz.block.ftl.v1.CallResponse.Error
}
var file_xyz_block_ftl_v1_ftl_proto_depIdxs = []int32{
	10, // 0: xyz.block.ftl.v1.CallResponse.error:type_name -> xyz.block.ftl.v1.CallResponse.Error
	8,  // 1: xyz.block.ftl.v1.AgentService.Serve:input_type -> xyz.block.ftl.v1.ServeRequest
	0,  // 2: xyz.block.ftl.v1.AgentService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	4,  // 3: xyz.block.ftl.v1.AgentService.Call:input_type -> xyz.block.ftl.v1.CallRequest
	6,  // 4: xyz.block.ftl.v1.AgentService.List:input_type -> xyz.block.ftl.v1.ListRequest
	0,  // 5: xyz.block.ftl.v1.DriveService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	2,  // 6: xyz.block.ftl.v1.DriveService.FileChange:input_type -> xyz.block.ftl.v1.FileChangeRequest
	4,  // 7: xyz.block.ftl.v1.DriveService.Call:input_type -> xyz.block.ftl.v1.CallRequest
	6,  // 8: xyz.block.ftl.v1.DriveService.List:input_type -> xyz.block.ftl.v1.ListRequest
	9,  // 9: xyz.block.ftl.v1.AgentService.Serve:output_type -> xyz.block.ftl.v1.ServeResponse
	1,  // 10: xyz.block.ftl.v1.AgentService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	5,  // 11: xyz.block.ftl.v1.AgentService.Call:output_type -> xyz.block.ftl.v1.CallResponse
	7,  // 12: xyz.block.ftl.v1.AgentService.List:output_type -> xyz.block.ftl.v1.ListResponse
	1,  // 13: xyz.block.ftl.v1.DriveService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	3,  // 14: xyz.block.ftl.v1.DriveService.FileChange:output_type -> xyz.block.ftl.v1.FileChangeResponse
	5,  // 15: xyz.block.ftl.v1.DriveService.Call:output_type -> xyz.block.ftl.v1.CallResponse
	7,  // 16: xyz.block.ftl.v1.DriveService.List:output_type -> xyz.block.ftl.v1.ListResponse
	9,  // [9:17] is the sub-list for method output_type
	1,  // [1:9] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_ftl_proto_init() }
func file_xyz_block_ftl_v1_ftl_proto_init() {
	if File_xyz_block_ftl_v1_ftl_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileChangeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileChangeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_xyz_block_ftl_v1_ftl_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallResponse_Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_xyz_block_ftl_v1_ftl_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*CallResponse_Body)(nil),
		(*CallResponse_Error_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1_ftl_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   2,
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
