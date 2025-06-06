// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: xyz/block/ftl/v1/deploymentcontext.proto

package ftlv1

import (
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

type GetDeploymentContextResponse_DbType int32

const (
	GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED GetDeploymentContextResponse_DbType = 0
	GetDeploymentContextResponse_DB_TYPE_POSTGRES    GetDeploymentContextResponse_DbType = 1
	GetDeploymentContextResponse_DB_TYPE_MYSQL       GetDeploymentContextResponse_DbType = 2
)

// Enum value maps for GetDeploymentContextResponse_DbType.
var (
	GetDeploymentContextResponse_DbType_name = map[int32]string{
		0: "DB_TYPE_UNSPECIFIED",
		1: "DB_TYPE_POSTGRES",
		2: "DB_TYPE_MYSQL",
	}
	GetDeploymentContextResponse_DbType_value = map[string]int32{
		"DB_TYPE_UNSPECIFIED": 0,
		"DB_TYPE_POSTGRES":    1,
		"DB_TYPE_MYSQL":       2,
	}
)

func (x GetDeploymentContextResponse_DbType) Enum() *GetDeploymentContextResponse_DbType {
	p := new(GetDeploymentContextResponse_DbType)
	*p = x
	return p
}

func (x GetDeploymentContextResponse_DbType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GetDeploymentContextResponse_DbType) Descriptor() protoreflect.EnumDescriptor {
	return file_xyz_block_ftl_v1_deploymentcontext_proto_enumTypes[0].Descriptor()
}

func (GetDeploymentContextResponse_DbType) Type() protoreflect.EnumType {
	return &file_xyz_block_ftl_v1_deploymentcontext_proto_enumTypes[0]
}

func (x GetDeploymentContextResponse_DbType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GetDeploymentContextResponse_DbType.Descriptor instead.
func (GetDeploymentContextResponse_DbType) EnumDescriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescGZIP(), []int{1, 0}
}

type GetDeploymentContextRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Deployment    string                 `protobuf:"bytes,1,opt,name=deployment,proto3" json:"deployment,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDeploymentContextRequest) Reset() {
	*x = GetDeploymentContextRequest{}
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextRequest) ProtoMessage() {}

func (x *GetDeploymentContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeploymentContextRequest.ProtoReflect.Descriptor instead.
func (*GetDeploymentContextRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescGZIP(), []int{0}
}

func (x *GetDeploymentContextRequest) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

type GetDeploymentContextResponse struct {
	state         protoimpl.MessageState              `protogen:"open.v1"`
	Module        string                              `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Deployment    string                              `protobuf:"bytes,2,opt,name=deployment,proto3" json:"deployment,omitempty"`
	Configs       map[string][]byte                   `protobuf:"bytes,3,rep,name=configs,proto3" json:"configs,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Secrets       map[string][]byte                   `protobuf:"bytes,4,rep,name=secrets,proto3" json:"secrets,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Databases     []*GetDeploymentContextResponse_DSN `protobuf:"bytes,5,rep,name=databases,proto3" json:"databases,omitempty"`
	Egress        map[string]string                   `protobuf:"bytes,7,rep,name=egress,proto3" json:"egress,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDeploymentContextResponse) Reset() {
	*x = GetDeploymentContextResponse{}
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextResponse) ProtoMessage() {}

func (x *GetDeploymentContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeploymentContextResponse.ProtoReflect.Descriptor instead.
func (*GetDeploymentContextResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescGZIP(), []int{1}
}

func (x *GetDeploymentContextResponse) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *GetDeploymentContextResponse) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

func (x *GetDeploymentContextResponse) GetConfigs() map[string][]byte {
	if x != nil {
		return x.Configs
	}
	return nil
}

func (x *GetDeploymentContextResponse) GetSecrets() map[string][]byte {
	if x != nil {
		return x.Secrets
	}
	return nil
}

func (x *GetDeploymentContextResponse) GetDatabases() []*GetDeploymentContextResponse_DSN {
	if x != nil {
		return x.Databases
	}
	return nil
}

func (x *GetDeploymentContextResponse) GetEgress() map[string]string {
	if x != nil {
		return x.Egress
	}
	return nil
}

type GetDeploymentContextResponse_DSN struct {
	state         protoimpl.MessageState              `protogen:"open.v1"`
	Name          string                              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type          GetDeploymentContextResponse_DbType `protobuf:"varint,2,opt,name=type,proto3,enum=xyz.block.ftl.v1.GetDeploymentContextResponse_DbType" json:"type,omitempty"`
	Dsn           string                              `protobuf:"bytes,3,opt,name=dsn,proto3" json:"dsn,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDeploymentContextResponse_DSN) Reset() {
	*x = GetDeploymentContextResponse_DSN{}
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextResponse_DSN) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextResponse_DSN) ProtoMessage() {}

func (x *GetDeploymentContextResponse_DSN) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeploymentContextResponse_DSN.ProtoReflect.Descriptor instead.
func (*GetDeploymentContextResponse_DSN) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescGZIP(), []int{1, 0}
}

func (x *GetDeploymentContextResponse_DSN) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetDeploymentContextResponse_DSN) GetType() GetDeploymentContextResponse_DbType {
	if x != nil {
		return x.Type
	}
	return GetDeploymentContextResponse_DB_TYPE_UNSPECIFIED
}

func (x *GetDeploymentContextResponse_DSN) GetDsn() string {
	if x != nil {
		return x.Dsn
	}
	return ""
}

var File_xyz_block_ftl_v1_deploymentcontext_proto protoreflect.FileDescriptor

const file_xyz_block_ftl_v1_deploymentcontext_proto_rawDesc = "" +
	"\n" +
	"(xyz/block/ftl/v1/deploymentcontext.proto\x12\x10xyz.block.ftl.v1\x1a\x1axyz/block/ftl/v1/ftl.proto\"=\n" +
	"\x1bGetDeploymentContextRequest\x12\x1e\n" +
	"\n" +
	"deployment\x18\x01 \x01(\tR\n" +
	"deployment\"\xa1\x06\n" +
	"\x1cGetDeploymentContextResponse\x12\x16\n" +
	"\x06module\x18\x01 \x01(\tR\x06module\x12\x1e\n" +
	"\n" +
	"deployment\x18\x02 \x01(\tR\n" +
	"deployment\x12U\n" +
	"\aconfigs\x18\x03 \x03(\v2;.xyz.block.ftl.v1.GetDeploymentContextResponse.ConfigsEntryR\aconfigs\x12U\n" +
	"\asecrets\x18\x04 \x03(\v2;.xyz.block.ftl.v1.GetDeploymentContextResponse.SecretsEntryR\asecrets\x12P\n" +
	"\tdatabases\x18\x05 \x03(\v22.xyz.block.ftl.v1.GetDeploymentContextResponse.DSNR\tdatabases\x12R\n" +
	"\x06egress\x18\a \x03(\v2:.xyz.block.ftl.v1.GetDeploymentContextResponse.EgressEntryR\x06egress\x1av\n" +
	"\x03DSN\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12I\n" +
	"\x04type\x18\x02 \x01(\x0e25.xyz.block.ftl.v1.GetDeploymentContextResponse.DbTypeR\x04type\x12\x10\n" +
	"\x03dsn\x18\x03 \x01(\tR\x03dsn\x1a:\n" +
	"\fConfigsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\fR\x05value:\x028\x01\x1a:\n" +
	"\fSecretsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\fR\x05value:\x028\x01\x1a9\n" +
	"\vEgressEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"J\n" +
	"\x06DbType\x12\x17\n" +
	"\x13DB_TYPE_UNSPECIFIED\x10\x00\x12\x14\n" +
	"\x10DB_TYPE_POSTGRES\x10\x01\x12\x11\n" +
	"\rDB_TYPE_MYSQL\x10\x022\xdf\x01\n" +
	"\x18DeploymentContextService\x12J\n" +
	"\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12w\n" +
	"\x14GetDeploymentContext\x12-.xyz.block.ftl.v1.GetDeploymentContextRequest\x1a..xyz.block.ftl.v1.GetDeploymentContextResponse0\x01B>P\x01Z:github.com/block/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3"

var (
	file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescData []byte
)

func file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_v1_deploymentcontext_proto_rawDesc), len(file_xyz_block_ftl_v1_deploymentcontext_proto_rawDesc)))
	})
	return file_xyz_block_ftl_v1_deploymentcontext_proto_rawDescData
}

var file_xyz_block_ftl_v1_deploymentcontext_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_xyz_block_ftl_v1_deploymentcontext_proto_goTypes = []any{
	(GetDeploymentContextResponse_DbType)(0), // 0: xyz.block.ftl.v1.GetDeploymentContextResponse.DbType
	(*GetDeploymentContextRequest)(nil),      // 1: xyz.block.ftl.v1.GetDeploymentContextRequest
	(*GetDeploymentContextResponse)(nil),     // 2: xyz.block.ftl.v1.GetDeploymentContextResponse
	(*GetDeploymentContextResponse_DSN)(nil), // 3: xyz.block.ftl.v1.GetDeploymentContextResponse.DSN
	nil,                                      // 4: xyz.block.ftl.v1.GetDeploymentContextResponse.ConfigsEntry
	nil,                                      // 5: xyz.block.ftl.v1.GetDeploymentContextResponse.SecretsEntry
	nil,                                      // 6: xyz.block.ftl.v1.GetDeploymentContextResponse.EgressEntry
	(*PingRequest)(nil),                      // 7: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),                     // 8: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_v1_deploymentcontext_proto_depIdxs = []int32{
	4, // 0: xyz.block.ftl.v1.GetDeploymentContextResponse.configs:type_name -> xyz.block.ftl.v1.GetDeploymentContextResponse.ConfigsEntry
	5, // 1: xyz.block.ftl.v1.GetDeploymentContextResponse.secrets:type_name -> xyz.block.ftl.v1.GetDeploymentContextResponse.SecretsEntry
	3, // 2: xyz.block.ftl.v1.GetDeploymentContextResponse.databases:type_name -> xyz.block.ftl.v1.GetDeploymentContextResponse.DSN
	6, // 3: xyz.block.ftl.v1.GetDeploymentContextResponse.egress:type_name -> xyz.block.ftl.v1.GetDeploymentContextResponse.EgressEntry
	0, // 4: xyz.block.ftl.v1.GetDeploymentContextResponse.DSN.type:type_name -> xyz.block.ftl.v1.GetDeploymentContextResponse.DbType
	7, // 5: xyz.block.ftl.v1.DeploymentContextService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	1, // 6: xyz.block.ftl.v1.DeploymentContextService.GetDeploymentContext:input_type -> xyz.block.ftl.v1.GetDeploymentContextRequest
	8, // 7: xyz.block.ftl.v1.DeploymentContextService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	2, // 8: xyz.block.ftl.v1.DeploymentContextService.GetDeploymentContext:output_type -> xyz.block.ftl.v1.GetDeploymentContextResponse
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_deploymentcontext_proto_init() }
func file_xyz_block_ftl_v1_deploymentcontext_proto_init() {
	if File_xyz_block_ftl_v1_deploymentcontext_proto != nil {
		return
	}
	file_xyz_block_ftl_v1_ftl_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_v1_deploymentcontext_proto_rawDesc), len(file_xyz_block_ftl_v1_deploymentcontext_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_v1_deploymentcontext_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1_deploymentcontext_proto_depIdxs,
		EnumInfos:         file_xyz_block_ftl_v1_deploymentcontext_proto_enumTypes,
		MessageInfos:      file_xyz_block_ftl_v1_deploymentcontext_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1_deploymentcontext_proto = out.File
	file_xyz_block_ftl_v1_deploymentcontext_proto_goTypes = nil
	file_xyz_block_ftl_v1_deploymentcontext_proto_depIdxs = nil
}
