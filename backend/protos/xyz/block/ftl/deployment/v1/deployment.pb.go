// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: xyz/block/ftl/deployment/v1/deployment.proto

package ftlv1

import (
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
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
	return file_xyz_block_ftl_deployment_v1_deployment_proto_enumTypes[0].Descriptor()
}

func (GetDeploymentContextResponse_DbType) Type() protoreflect.EnumType {
	return &file_xyz_block_ftl_deployment_v1_deployment_proto_enumTypes[0]
}

func (x GetDeploymentContextResponse_DbType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GetDeploymentContextResponse_DbType.Descriptor instead.
func (GetDeploymentContextResponse_DbType) EnumDescriptor() ([]byte, []int) {
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP(), []int{1, 0}
}

type GetDeploymentContextRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Deployment string `protobuf:"bytes,1,opt,name=deployment,proto3" json:"deployment,omitempty"`
}

func (x *GetDeploymentContextRequest) Reset() {
	*x = GetDeploymentContextRequest{}
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextRequest) ProtoMessage() {}

func (x *GetDeploymentContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[0]
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
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP(), []int{0}
}

func (x *GetDeploymentContextRequest) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

type GetDeploymentContextResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module     string                                `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Deployment string                                `protobuf:"bytes,2,opt,name=deployment,proto3" json:"deployment,omitempty"`
	Configs    map[string][]byte                     `protobuf:"bytes,3,rep,name=configs,proto3" json:"configs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Secrets    map[string][]byte                     `protobuf:"bytes,4,rep,name=secrets,proto3" json:"secrets,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Databases  []*GetDeploymentContextResponse_DSN   `protobuf:"bytes,5,rep,name=databases,proto3" json:"databases,omitempty"`
	Routes     []*GetDeploymentContextResponse_Route `protobuf:"bytes,6,rep,name=routes,proto3" json:"routes,omitempty"`
}

func (x *GetDeploymentContextResponse) Reset() {
	*x = GetDeploymentContextResponse{}
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextResponse) ProtoMessage() {}

func (x *GetDeploymentContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[1]
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
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP(), []int{1}
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

func (x *GetDeploymentContextResponse) GetRoutes() []*GetDeploymentContextResponse_Route {
	if x != nil {
		return x.Routes
	}
	return nil
}

type GetDeploymentContextResponse_DSN struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string                              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type GetDeploymentContextResponse_DbType `protobuf:"varint,2,opt,name=type,proto3,enum=xyz.block.ftl.deployment.v1.GetDeploymentContextResponse_DbType" json:"type,omitempty"`
	Dsn  string                              `protobuf:"bytes,3,opt,name=dsn,proto3" json:"dsn,omitempty"`
}

func (x *GetDeploymentContextResponse_DSN) Reset() {
	*x = GetDeploymentContextResponse_DSN{}
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextResponse_DSN) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextResponse_DSN) ProtoMessage() {}

func (x *GetDeploymentContextResponse_DSN) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[2]
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
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP(), []int{1, 0}
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

type GetDeploymentContextResponse_Route struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Uri    string `protobuf:"bytes,2,opt,name=uri,proto3" json:"uri,omitempty"`
}

func (x *GetDeploymentContextResponse_Route) Reset() {
	*x = GetDeploymentContextResponse_Route{}
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDeploymentContextResponse_Route) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeploymentContextResponse_Route) ProtoMessage() {}

func (x *GetDeploymentContextResponse_Route) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeploymentContextResponse_Route.ProtoReflect.Descriptor instead.
func (*GetDeploymentContextResponse_Route) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP(), []int{1, 1}
}

func (x *GetDeploymentContextResponse_Route) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *GetDeploymentContextResponse_Route) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

var File_xyz_block_ftl_deployment_v1_deployment_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_deployment_v1_deployment_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x64, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1a, 0x78, 0x79, 0x7a,
	0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3d, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0xcb, 0x06, 0x0a, 0x1c, 0x47, 0x65, 0x74, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12,
	0x1e, 0x0a, 0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12,
	0x60, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x46, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x73, 0x12, 0x60, 0x0a, 0x07, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x46, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x65,
	0x63, 0x72, 0x65, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x73, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x73, 0x12, 0x5b, 0x0a, 0x09, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65,
	0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x44, 0x53, 0x4e, 0x52, 0x09, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73,
	0x12, 0x57, 0x0a, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x3f, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x52, 0x6f, 0x75, 0x74,
	0x65, 0x52, 0x06, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x73, 0x1a, 0x81, 0x01, 0x0a, 0x03, 0x44, 0x53,
	0x4e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x54, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x40, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x44,
	0x62, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x64,
	0x73, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x73, 0x6e, 0x1a, 0x31, 0x0a,
	0x05, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x69,
	0x1a, 0x3a, 0x0a, 0x0c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3a, 0x0a, 0x0c,
	0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4a, 0x0a, 0x06, 0x44, 0x62, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x17, 0x0a, 0x13, 0x44, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x44,
	0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x4f, 0x53, 0x54, 0x47, 0x52, 0x45, 0x53, 0x10,
	0x01, 0x12, 0x11, 0x0a, 0x0d, 0x44, 0x42, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d, 0x59, 0x53,
	0x51, 0x4c, 0x10, 0x02, 0x32, 0xef, 0x01, 0x0a, 0x11, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69,
	0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x12, 0x8d, 0x01, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12,
	0x38, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x39, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x42, 0x4f, 0x50, 0x01, 0x5a, 0x4b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42, 0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39,
	0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f,
	0x66, 0x74, 0x6c, 0x2f, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2f, 0x76,
	0x31, 0x3b, 0x66, 0x74, 0x6c, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescData = file_xyz_block_ftl_deployment_v1_deployment_proto_rawDesc
)

func file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescData)
	})
	return file_xyz_block_ftl_deployment_v1_deployment_proto_rawDescData
}

var file_xyz_block_ftl_deployment_v1_deployment_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_xyz_block_ftl_deployment_v1_deployment_proto_goTypes = []any{
	(GetDeploymentContextResponse_DbType)(0),   // 0: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DbType
	(*GetDeploymentContextRequest)(nil),        // 1: xyz.block.ftl.deployment.v1.GetDeploymentContextRequest
	(*GetDeploymentContextResponse)(nil),       // 2: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse
	(*GetDeploymentContextResponse_DSN)(nil),   // 3: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DSN
	(*GetDeploymentContextResponse_Route)(nil), // 4: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.Route
	nil,                     // 5: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.ConfigsEntry
	nil,                     // 6: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.SecretsEntry
	(*v1.PingRequest)(nil),  // 7: xyz.block.ftl.v1.PingRequest
	(*v1.PingResponse)(nil), // 8: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_deployment_v1_deployment_proto_depIdxs = []int32{
	5, // 0: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.configs:type_name -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.ConfigsEntry
	6, // 1: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.secrets:type_name -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.SecretsEntry
	3, // 2: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.databases:type_name -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DSN
	4, // 3: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.routes:type_name -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.Route
	0, // 4: xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DSN.type:type_name -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.DbType
	7, // 5: xyz.block.ftl.deployment.v1.DeploymentService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	1, // 6: xyz.block.ftl.deployment.v1.DeploymentService.GetDeploymentContext:input_type -> xyz.block.ftl.deployment.v1.GetDeploymentContextRequest
	8, // 7: xyz.block.ftl.deployment.v1.DeploymentService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	2, // 8: xyz.block.ftl.deployment.v1.DeploymentService.GetDeploymentContext:output_type -> xyz.block.ftl.deployment.v1.GetDeploymentContextResponse
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_deployment_v1_deployment_proto_init() }
func file_xyz_block_ftl_deployment_v1_deployment_proto_init() {
	if File_xyz_block_ftl_deployment_v1_deployment_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_deployment_v1_deployment_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_deployment_v1_deployment_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_deployment_v1_deployment_proto_depIdxs,
		EnumInfos:         file_xyz_block_ftl_deployment_v1_deployment_proto_enumTypes,
		MessageInfos:      file_xyz_block_ftl_deployment_v1_deployment_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_deployment_v1_deployment_proto = out.File
	file_xyz_block_ftl_deployment_v1_deployment_proto_rawDesc = nil
	file_xyz_block_ftl_deployment_v1_deployment_proto_goTypes = nil
	file_xyz_block_ftl_deployment_v1_deployment_proto_depIdxs = nil
}
