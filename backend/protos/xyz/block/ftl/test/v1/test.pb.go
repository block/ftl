// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: xyz/block/ftl/test/v1/test.proto

package testpb

import (
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
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

type StartTestRunRequest struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	ModuleName        string                 `protobuf:"bytes,1,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
	Endpoint          string                 `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	HotReloadEndpoint string                 `protobuf:"bytes,3,opt,name=hot_reload_endpoint,json=hotReloadEndpoint,proto3" json:"hot_reload_endpoint,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *StartTestRunRequest) Reset() {
	*x = StartTestRunRequest{}
	mi := &file_xyz_block_ftl_test_v1_test_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartTestRunRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartTestRunRequest) ProtoMessage() {}

func (x *StartTestRunRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_test_v1_test_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartTestRunRequest.ProtoReflect.Descriptor instead.
func (*StartTestRunRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_test_v1_test_proto_rawDescGZIP(), []int{0}
}

func (x *StartTestRunRequest) GetModuleName() string {
	if x != nil {
		return x.ModuleName
	}
	return ""
}

func (x *StartTestRunRequest) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *StartTestRunRequest) GetHotReloadEndpoint() string {
	if x != nil {
		return x.HotReloadEndpoint
	}
	return ""
}

type StartTestRunResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DeploymentKey string                 `protobuf:"bytes,1,opt,name=deployment_key,json=deploymentKey,proto3" json:"deployment_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StartTestRunResponse) Reset() {
	*x = StartTestRunResponse{}
	mi := &file_xyz_block_ftl_test_v1_test_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StartTestRunResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartTestRunResponse) ProtoMessage() {}

func (x *StartTestRunResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_test_v1_test_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartTestRunResponse.ProtoReflect.Descriptor instead.
func (*StartTestRunResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_test_v1_test_proto_rawDescGZIP(), []int{1}
}

func (x *StartTestRunResponse) GetDeploymentKey() string {
	if x != nil {
		return x.DeploymentKey
	}
	return ""
}

var File_xyz_block_ftl_test_v1_test_proto protoreflect.FileDescriptor

const file_xyz_block_ftl_test_v1_test_proto_rawDesc = "" +
	"\n" +
	" xyz/block/ftl/test/v1/test.proto\x12\x15xyz.block.ftl.test.v1\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x82\x01\n" +
	"\x13StartTestRunRequest\x12\x1f\n" +
	"\vmodule_name\x18\x01 \x01(\tR\n" +
	"moduleName\x12\x1a\n" +
	"\bendpoint\x18\x02 \x01(\tR\bendpoint\x12.\n" +
	"\x13hot_reload_endpoint\x18\x03 \x01(\tR\x11hotReloadEndpoint\"=\n" +
	"\x14StartTestRunResponse\x12%\n" +
	"\x0edeployment_key\x18\x01 \x01(\tR\rdeploymentKey2\xc2\x01\n" +
	"\vTestService\x12J\n" +
	"\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12g\n" +
	"\fStartTestRun\x12*.xyz.block.ftl.test.v1.StartTestRunRequest\x1a+.xyz.block.ftl.test.v1.StartTestRunResponseBDP\x01Z@github.com/block/ftl/backend/protos/xyz/block/ftl/test/v1;testpbb\x06proto3"

var (
	file_xyz_block_ftl_test_v1_test_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_test_v1_test_proto_rawDescData []byte
)

func file_xyz_block_ftl_test_v1_test_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_test_v1_test_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_test_v1_test_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_test_v1_test_proto_rawDesc), len(file_xyz_block_ftl_test_v1_test_proto_rawDesc)))
	})
	return file_xyz_block_ftl_test_v1_test_proto_rawDescData
}

var file_xyz_block_ftl_test_v1_test_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_xyz_block_ftl_test_v1_test_proto_goTypes = []any{
	(*StartTestRunRequest)(nil),  // 0: xyz.block.ftl.test.v1.StartTestRunRequest
	(*StartTestRunResponse)(nil), // 1: xyz.block.ftl.test.v1.StartTestRunResponse
	(*v1.PingRequest)(nil),       // 2: xyz.block.ftl.v1.PingRequest
	(*v1.PingResponse)(nil),      // 3: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_test_v1_test_proto_depIdxs = []int32{
	2, // 0: xyz.block.ftl.test.v1.TestService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	0, // 1: xyz.block.ftl.test.v1.TestService.StartTestRun:input_type -> xyz.block.ftl.test.v1.StartTestRunRequest
	3, // 2: xyz.block.ftl.test.v1.TestService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	1, // 3: xyz.block.ftl.test.v1.TestService.StartTestRun:output_type -> xyz.block.ftl.test.v1.StartTestRunResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_test_v1_test_proto_init() }
func file_xyz_block_ftl_test_v1_test_proto_init() {
	if File_xyz_block_ftl_test_v1_test_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_test_v1_test_proto_rawDesc), len(file_xyz_block_ftl_test_v1_test_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_test_v1_test_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_test_v1_test_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_test_v1_test_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_test_v1_test_proto = out.File
	file_xyz_block_ftl_test_v1_test_proto_goTypes = nil
	file_xyz_block_ftl_test_v1_test_proto_depIdxs = nil
}
