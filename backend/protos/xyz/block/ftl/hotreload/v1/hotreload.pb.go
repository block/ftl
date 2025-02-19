// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: xyz/block/ftl/hotreload/v1/hotreload.proto

package hotreloadpb

import (
	v11 "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	v12 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
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

type ReloadRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Force         bool                   `protobuf:"varint,1,opt,name=force,proto3" json:"force,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReloadRequest) Reset() {
	*x = ReloadRequest{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReloadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReloadRequest) ProtoMessage() {}

func (x *ReloadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReloadRequest.ProtoReflect.Descriptor instead.
func (*ReloadRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{0}
}

func (x *ReloadRequest) GetForce() bool {
	if x != nil {
		return x.Force
	}
	return false
}

type ReloadResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	State         *SchemaState           `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	Failed        bool                   `protobuf:"varint,2,opt,name=failed,proto3" json:"failed,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReloadResponse) Reset() {
	*x = ReloadResponse{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReloadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReloadResponse) ProtoMessage() {}

func (x *ReloadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReloadResponse.ProtoReflect.Descriptor instead.
func (*ReloadResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{1}
}

func (x *ReloadResponse) GetState() *SchemaState {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *ReloadResponse) GetFailed() bool {
	if x != nil {
		return x.Failed
	}
	return false
}

type WatchRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WatchRequest) Reset() {
	*x = WatchRequest{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchRequest) ProtoMessage() {}

func (x *WatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchRequest.ProtoReflect.Descriptor instead.
func (*WatchRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{2}
}

type WatchResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	State         *SchemaState           `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *WatchResponse) Reset() {
	*x = WatchResponse{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WatchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchResponse) ProtoMessage() {}

func (x *WatchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchResponse.ProtoReflect.Descriptor instead.
func (*WatchResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{3}
}

func (x *WatchResponse) GetState() *SchemaState {
	if x != nil {
		return x.State
	}
	return nil
}

type RunnerInfoRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Address       string                 `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Deployment    string                 `protobuf:"bytes,2,opt,name=deployment,proto3" json:"deployment,omitempty"`
	Databases     []*Database            `protobuf:"bytes,3,rep,name=databases,proto3" json:"databases,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RunnerInfoRequest) Reset() {
	*x = RunnerInfoRequest{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RunnerInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerInfoRequest) ProtoMessage() {}

func (x *RunnerInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunnerInfoRequest.ProtoReflect.Descriptor instead.
func (*RunnerInfoRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{4}
}

func (x *RunnerInfoRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *RunnerInfoRequest) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

func (x *RunnerInfoRequest) GetDatabases() []*Database {
	if x != nil {
		return x.Databases
	}
	return nil
}

type Database struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Address       string                 `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Database) Reset() {
	*x = Database{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Database) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Database) ProtoMessage() {}

func (x *Database) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Database.ProtoReflect.Descriptor instead.
func (*Database) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{5}
}

func (x *Database) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Database) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type RunnerInfoResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RunnerInfoResponse) Reset() {
	*x = RunnerInfoResponse{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RunnerInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerInfoResponse) ProtoMessage() {}

func (x *RunnerInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunnerInfoResponse.ProtoReflect.Descriptor instead.
func (*RunnerInfoResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{6}
}

type ReloadNotRequired struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReloadNotRequired) Reset() {
	*x = ReloadNotRequired{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReloadNotRequired) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReloadNotRequired) ProtoMessage() {}

func (x *ReloadNotRequired) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReloadNotRequired.ProtoReflect.Descriptor instead.
func (*ReloadNotRequired) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{7}
}

type ReloadSuccess struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	State         *SchemaState           `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReloadSuccess) Reset() {
	*x = ReloadSuccess{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReloadSuccess) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReloadSuccess) ProtoMessage() {}

func (x *ReloadSuccess) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReloadSuccess.ProtoReflect.Descriptor instead.
func (*ReloadSuccess) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{8}
}

func (x *ReloadSuccess) GetState() *SchemaState {
	if x != nil {
		return x.State
	}
	return nil
}

type ReloadFailed struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Module schema for the built module
	State         *SchemaState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReloadFailed) Reset() {
	*x = ReloadFailed{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReloadFailed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReloadFailed) ProtoMessage() {}

func (x *ReloadFailed) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReloadFailed.ProtoReflect.Descriptor instead.
func (*ReloadFailed) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{9}
}

func (x *ReloadFailed) GetState() *SchemaState {
	if x != nil {
		return x.State
	}
	return nil
}

type SchemaState struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Module            *v1.Module             `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Errors            *v11.ErrorList         `protobuf:"bytes,2,opt,name=errors,proto3" json:"errors,omitempty"`
	NewRunnerRequired bool                   `protobuf:"varint,3,opt,name=new_runner_required,json=newRunnerRequired,proto3" json:"new_runner_required,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *SchemaState) Reset() {
	*x = SchemaState{}
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SchemaState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchemaState) ProtoMessage() {}

func (x *SchemaState) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SchemaState.ProtoReflect.Descriptor instead.
func (*SchemaState) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP(), []int{10}
}

func (x *SchemaState) GetModule() *v1.Module {
	if x != nil {
		return x.Module
	}
	return nil
}

func (x *SchemaState) GetErrors() *v11.ErrorList {
	if x != nil {
		return x.Errors
	}
	return nil
}

func (x *SchemaState) GetNewRunnerRequired() bool {
	if x != nil {
		return x.NewRunnerRequired
	}
	return false
}

var File_xyz_block_ftl_hotreload_v1_hotreload_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDesc = string([]byte{
	0x0a, 0x2a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x68, 0x6f, 0x74,
	0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72,
	0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x1a, 0x28, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x24, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74,
	0x6c, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65,
	0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x25, 0x0a, 0x0d, 0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x22, 0x67, 0x0a, 0x0e, 0x52,
	0x65, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74,
	0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x66, 0x61,
	0x69, 0x6c, 0x65, 0x64, 0x22, 0x0e, 0x0a, 0x0c, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x4e, 0x0a, 0x0d, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x22, 0x91, 0x01, 0x0a, 0x11, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65,
	0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x12, 0x42, 0x0a, 0x09, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61,
	0x64, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x52, 0x09, 0x64,
	0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73, 0x22, 0x38, 0x0a, 0x08, 0x44, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x22, 0x14, 0x0a, 0x12, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x0a, 0x11, 0x52, 0x65, 0x6c, 0x6f,
	0x61, 0x64, 0x4e, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x22, 0x4e, 0x0a,
	0x0d, 0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x3d,
	0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f,
	0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x4d, 0x0a,
	0x0c, 0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x12, 0x3d, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74,
	0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0xb4, 0x01, 0x0a,
	0x0b, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x37, 0x0a, 0x06,
	0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x06, 0x6d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x3c, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x06, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x73, 0x12, 0x2e, 0x0a, 0x13, 0x6e, 0x65, 0x77, 0x5f, 0x72, 0x75, 0x6e, 0x6e, 0x65,
	0x72, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x11, 0x6e, 0x65, 0x77, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x64, 0x32, 0x8c, 0x03, 0x0a, 0x10, 0x48, 0x6f, 0x74, 0x52, 0x65, 0x6c, 0x6f, 0x61,
	0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67,
	0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x03, 0x90, 0x02, 0x01, 0x12, 0x5f, 0x0a, 0x06, 0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x29,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68,
	0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x6c, 0x6f,
	0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c,
	0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5e, 0x0a, 0x05, 0x57, 0x61, 0x74, 0x63, 0x68, 0x12, 0x28,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68,
	0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x61, 0x74, 0x63,
	0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f,
	0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x6b, 0x0a, 0x0a, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x2d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x76, 0x31, 0x2e,
	0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x4e, 0x50, 0x01, 0x5a, 0x4a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c,
	0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31, 0x3b, 0x68, 0x6f, 0x74, 0x72, 0x65, 0x6c, 0x6f, 0x61, 0x64,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescData []byte
)

func file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDesc), len(file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDesc)))
	})
	return file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDescData
}

var file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_xyz_block_ftl_hotreload_v1_hotreload_proto_goTypes = []any{
	(*ReloadRequest)(nil),      // 0: xyz.block.ftl.hotreload.v1.ReloadRequest
	(*ReloadResponse)(nil),     // 1: xyz.block.ftl.hotreload.v1.ReloadResponse
	(*WatchRequest)(nil),       // 2: xyz.block.ftl.hotreload.v1.WatchRequest
	(*WatchResponse)(nil),      // 3: xyz.block.ftl.hotreload.v1.WatchResponse
	(*RunnerInfoRequest)(nil),  // 4: xyz.block.ftl.hotreload.v1.RunnerInfoRequest
	(*Database)(nil),           // 5: xyz.block.ftl.hotreload.v1.Database
	(*RunnerInfoResponse)(nil), // 6: xyz.block.ftl.hotreload.v1.RunnerInfoResponse
	(*ReloadNotRequired)(nil),  // 7: xyz.block.ftl.hotreload.v1.ReloadNotRequired
	(*ReloadSuccess)(nil),      // 8: xyz.block.ftl.hotreload.v1.ReloadSuccess
	(*ReloadFailed)(nil),       // 9: xyz.block.ftl.hotreload.v1.ReloadFailed
	(*SchemaState)(nil),        // 10: xyz.block.ftl.hotreload.v1.SchemaState
	(*v1.Module)(nil),          // 11: xyz.block.ftl.schema.v1.Module
	(*v11.ErrorList)(nil),      // 12: xyz.block.ftl.language.v1.ErrorList
	(*v12.PingRequest)(nil),    // 13: xyz.block.ftl.v1.PingRequest
	(*v12.PingResponse)(nil),   // 14: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_hotreload_v1_hotreload_proto_depIdxs = []int32{
	10, // 0: xyz.block.ftl.hotreload.v1.ReloadResponse.state:type_name -> xyz.block.ftl.hotreload.v1.SchemaState
	10, // 1: xyz.block.ftl.hotreload.v1.WatchResponse.state:type_name -> xyz.block.ftl.hotreload.v1.SchemaState
	5,  // 2: xyz.block.ftl.hotreload.v1.RunnerInfoRequest.databases:type_name -> xyz.block.ftl.hotreload.v1.Database
	10, // 3: xyz.block.ftl.hotreload.v1.ReloadSuccess.state:type_name -> xyz.block.ftl.hotreload.v1.SchemaState
	10, // 4: xyz.block.ftl.hotreload.v1.ReloadFailed.state:type_name -> xyz.block.ftl.hotreload.v1.SchemaState
	11, // 5: xyz.block.ftl.hotreload.v1.SchemaState.module:type_name -> xyz.block.ftl.schema.v1.Module
	12, // 6: xyz.block.ftl.hotreload.v1.SchemaState.errors:type_name -> xyz.block.ftl.language.v1.ErrorList
	13, // 7: xyz.block.ftl.hotreload.v1.HotReloadService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	0,  // 8: xyz.block.ftl.hotreload.v1.HotReloadService.Reload:input_type -> xyz.block.ftl.hotreload.v1.ReloadRequest
	2,  // 9: xyz.block.ftl.hotreload.v1.HotReloadService.Watch:input_type -> xyz.block.ftl.hotreload.v1.WatchRequest
	4,  // 10: xyz.block.ftl.hotreload.v1.HotReloadService.RunnerInfo:input_type -> xyz.block.ftl.hotreload.v1.RunnerInfoRequest
	14, // 11: xyz.block.ftl.hotreload.v1.HotReloadService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	1,  // 12: xyz.block.ftl.hotreload.v1.HotReloadService.Reload:output_type -> xyz.block.ftl.hotreload.v1.ReloadResponse
	3,  // 13: xyz.block.ftl.hotreload.v1.HotReloadService.Watch:output_type -> xyz.block.ftl.hotreload.v1.WatchResponse
	6,  // 14: xyz.block.ftl.hotreload.v1.HotReloadService.RunnerInfo:output_type -> xyz.block.ftl.hotreload.v1.RunnerInfoResponse
	11, // [11:15] is the sub-list for method output_type
	7,  // [7:11] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_hotreload_v1_hotreload_proto_init() }
func file_xyz_block_ftl_hotreload_v1_hotreload_proto_init() {
	if File_xyz_block_ftl_hotreload_v1_hotreload_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDesc), len(file_xyz_block_ftl_hotreload_v1_hotreload_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_hotreload_v1_hotreload_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_hotreload_v1_hotreload_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_hotreload_v1_hotreload_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_hotreload_v1_hotreload_proto = out.File
	file_xyz_block_ftl_hotreload_v1_hotreload_proto_goTypes = nil
	file_xyz_block_ftl_hotreload_v1_hotreload_proto_depIdxs = nil
}
