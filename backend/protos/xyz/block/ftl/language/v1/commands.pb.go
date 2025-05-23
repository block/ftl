// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: xyz/block/ftl/language/v1/commands.proto

package languagepb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
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

type GetNewModuleFlagsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Language      string                 `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetNewModuleFlagsRequest) Reset() {
	*x = GetNewModuleFlagsRequest{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetNewModuleFlagsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNewModuleFlagsRequest) ProtoMessage() {}

func (x *GetNewModuleFlagsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNewModuleFlagsRequest.ProtoReflect.Descriptor instead.
func (*GetNewModuleFlagsRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{0}
}

func (x *GetNewModuleFlagsRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

type GetNewModuleFlagsResponse struct {
	state         protoimpl.MessageState            `protogen:"open.v1"`
	Flags         []*GetNewModuleFlagsResponse_Flag `protobuf:"bytes,1,rep,name=flags,proto3" json:"flags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetNewModuleFlagsResponse) Reset() {
	*x = GetNewModuleFlagsResponse{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetNewModuleFlagsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNewModuleFlagsResponse) ProtoMessage() {}

func (x *GetNewModuleFlagsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNewModuleFlagsResponse.ProtoReflect.Descriptor instead.
func (*GetNewModuleFlagsResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{1}
}

func (x *GetNewModuleFlagsResponse) GetFlags() []*GetNewModuleFlagsResponse_Flag {
	if x != nil {
		return x.Flags
	}
	return nil
}

// Request to create a new module.
type NewModuleRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Name  string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The root directory for the module, which does not yet exist.
	// The plugin should create the directory.
	Dir string `protobuf:"bytes,2,opt,name=dir,proto3" json:"dir,omitempty"`
	// The project configuration
	ProjectConfig *ProjectConfig `protobuf:"bytes,3,opt,name=project_config,json=projectConfig,proto3" json:"project_config,omitempty"`
	// Flags contains any values set for those configured in the GetCreateModuleFlags call
	Flags         *structpb.Struct `protobuf:"bytes,4,opt,name=flags,proto3" json:"flags,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *NewModuleRequest) Reset() {
	*x = NewModuleRequest{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NewModuleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewModuleRequest) ProtoMessage() {}

func (x *NewModuleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewModuleRequest.ProtoReflect.Descriptor instead.
func (*NewModuleRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{2}
}

func (x *NewModuleRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *NewModuleRequest) GetDir() string {
	if x != nil {
		return x.Dir
	}
	return ""
}

func (x *NewModuleRequest) GetProjectConfig() *ProjectConfig {
	if x != nil {
		return x.ProjectConfig
	}
	return nil
}

func (x *NewModuleRequest) GetFlags() *structpb.Struct {
	if x != nil {
		return x.Flags
	}
	return nil
}

// Response to a create module request.
type NewModuleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *NewModuleResponse) Reset() {
	*x = NewModuleResponse{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *NewModuleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewModuleResponse) ProtoMessage() {}

func (x *NewModuleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewModuleResponse.ProtoReflect.Descriptor instead.
func (*NewModuleResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{3}
}

type GetModuleConfigDefaultsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Dir           string                 `protobuf:"bytes,1,opt,name=dir,proto3" json:"dir,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetModuleConfigDefaultsRequest) Reset() {
	*x = GetModuleConfigDefaultsRequest{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetModuleConfigDefaultsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetModuleConfigDefaultsRequest) ProtoMessage() {}

func (x *GetModuleConfigDefaultsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetModuleConfigDefaultsRequest.ProtoReflect.Descriptor instead.
func (*GetModuleConfigDefaultsRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{4}
}

func (x *GetModuleConfigDefaultsRequest) GetDir() string {
	if x != nil {
		return x.Dir
	}
	return ""
}

// GetModuleConfigDefaultsResponse provides defaults for ModuleConfig.
//
// The result may be cached by FTL, so defaulting logic should not be changing due to normal module changes.
// For example, it is valid to return defaults based on which build tool is configured within the module directory,
// as that is not expected to change during normal operation.
// It is not recommended to read the module's toml file to determine defaults, as when the toml file is updated,
// the module defaults will not be recalculated.
type GetModuleConfigDefaultsResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Default relative path to the directory containing all build artifacts for deployments
	DeployDir string `protobuf:"bytes,1,opt,name=deploy_dir,json=deployDir,proto3" json:"deploy_dir,omitempty"`
	// Default build command
	Build *string `protobuf:"bytes,2,opt,name=build,proto3,oneof" json:"build,omitempty"`
	// Dev mode build command, if different from the regular build command
	DevModeBuild *string `protobuf:"bytes,3,opt,name=dev_mode_build,json=devModeBuild,proto3,oneof" json:"dev_mode_build,omitempty"`
	// Build lock path to prevent concurrent builds
	BuildLock *string `protobuf:"bytes,4,opt,name=build_lock,json=buildLock,proto3,oneof" json:"build_lock,omitempty"`
	// Default patterns to watch for file changes, relative to the module directory
	Watch []string `protobuf:"bytes,6,rep,name=watch,proto3" json:"watch,omitempty"`
	// Default language specific configuration.
	// These defaults are filled in by looking at each root key only. If the key is not present, the default is used.
	LanguageConfig *structpb.Struct `protobuf:"bytes,7,opt,name=language_config,json=languageConfig,proto3" json:"language_config,omitempty"`
	// Root directory containing SQL files.
	SqlRootDir    string `protobuf:"bytes,8,opt,name=sql_root_dir,json=sqlRootDir,proto3" json:"sql_root_dir,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetModuleConfigDefaultsResponse) Reset() {
	*x = GetModuleConfigDefaultsResponse{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetModuleConfigDefaultsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetModuleConfigDefaultsResponse) ProtoMessage() {}

func (x *GetModuleConfigDefaultsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetModuleConfigDefaultsResponse.ProtoReflect.Descriptor instead.
func (*GetModuleConfigDefaultsResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{5}
}

func (x *GetModuleConfigDefaultsResponse) GetDeployDir() string {
	if x != nil {
		return x.DeployDir
	}
	return ""
}

func (x *GetModuleConfigDefaultsResponse) GetBuild() string {
	if x != nil && x.Build != nil {
		return *x.Build
	}
	return ""
}

func (x *GetModuleConfigDefaultsResponse) GetDevModeBuild() string {
	if x != nil && x.DevModeBuild != nil {
		return *x.DevModeBuild
	}
	return ""
}

func (x *GetModuleConfigDefaultsResponse) GetBuildLock() string {
	if x != nil && x.BuildLock != nil {
		return *x.BuildLock
	}
	return ""
}

func (x *GetModuleConfigDefaultsResponse) GetWatch() []string {
	if x != nil {
		return x.Watch
	}
	return nil
}

func (x *GetModuleConfigDefaultsResponse) GetLanguageConfig() *structpb.Struct {
	if x != nil {
		return x.LanguageConfig
	}
	return nil
}

func (x *GetModuleConfigDefaultsResponse) GetSqlRootDir() string {
	if x != nil {
		return x.SqlRootDir
	}
	return ""
}

type GetSQLInterfacesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *ModuleConfig          `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetSQLInterfacesRequest) Reset() {
	*x = GetSQLInterfacesRequest{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSQLInterfacesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSQLInterfacesRequest) ProtoMessage() {}

func (x *GetSQLInterfacesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSQLInterfacesRequest.ProtoReflect.Descriptor instead.
func (*GetSQLInterfacesRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{6}
}

func (x *GetSQLInterfacesRequest) GetConfig() *ModuleConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type GetSQLInterfacesResponse struct {
	state         protoimpl.MessageState                `protogen:"open.v1"`
	Interfaces    []*GetSQLInterfacesResponse_Interface `protobuf:"bytes,1,rep,name=interfaces,proto3" json:"interfaces,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetSQLInterfacesResponse) Reset() {
	*x = GetSQLInterfacesResponse{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSQLInterfacesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSQLInterfacesResponse) ProtoMessage() {}

func (x *GetSQLInterfacesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSQLInterfacesResponse.ProtoReflect.Descriptor instead.
func (*GetSQLInterfacesResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{7}
}

func (x *GetSQLInterfacesResponse) GetInterfaces() []*GetSQLInterfacesResponse_Interface {
	if x != nil {
		return x.Interfaces
	}
	return nil
}

type GetNewModuleFlagsResponse_Flag struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Name  string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Help  string                 `protobuf:"bytes,2,opt,name=help,proto3" json:"help,omitempty"`
	Envar *string                `protobuf:"bytes,3,opt,name=envar,proto3,oneof" json:"envar,omitempty"`
	// short must be a single character
	Short         *string `protobuf:"bytes,4,opt,name=short,proto3,oneof" json:"short,omitempty"`
	Placeholder   *string `protobuf:"bytes,5,opt,name=placeholder,proto3,oneof" json:"placeholder,omitempty"`
	Default       *string `protobuf:"bytes,6,opt,name=default,proto3,oneof" json:"default,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetNewModuleFlagsResponse_Flag) Reset() {
	*x = GetNewModuleFlagsResponse_Flag{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetNewModuleFlagsResponse_Flag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNewModuleFlagsResponse_Flag) ProtoMessage() {}

func (x *GetNewModuleFlagsResponse_Flag) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNewModuleFlagsResponse_Flag.ProtoReflect.Descriptor instead.
func (*GetNewModuleFlagsResponse_Flag) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{1, 0}
}

func (x *GetNewModuleFlagsResponse_Flag) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetNewModuleFlagsResponse_Flag) GetHelp() string {
	if x != nil {
		return x.Help
	}
	return ""
}

func (x *GetNewModuleFlagsResponse_Flag) GetEnvar() string {
	if x != nil && x.Envar != nil {
		return *x.Envar
	}
	return ""
}

func (x *GetNewModuleFlagsResponse_Flag) GetShort() string {
	if x != nil && x.Short != nil {
		return *x.Short
	}
	return ""
}

func (x *GetNewModuleFlagsResponse_Flag) GetPlaceholder() string {
	if x != nil && x.Placeholder != nil {
		return *x.Placeholder
	}
	return ""
}

func (x *GetNewModuleFlagsResponse_Flag) GetDefault() string {
	if x != nil && x.Default != nil {
		return *x.Default
	}
	return ""
}

type GetSQLInterfacesResponse_Interface struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Interface     string                 `protobuf:"bytes,2,opt,name=interface,proto3" json:"interface,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetSQLInterfacesResponse_Interface) Reset() {
	*x = GetSQLInterfacesResponse_Interface{}
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSQLInterfacesResponse_Interface) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSQLInterfacesResponse_Interface) ProtoMessage() {}

func (x *GetSQLInterfacesResponse_Interface) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_language_v1_commands_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSQLInterfacesResponse_Interface.ProtoReflect.Descriptor instead.
func (*GetSQLInterfacesResponse_Interface) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP(), []int{7, 0}
}

func (x *GetSQLInterfacesResponse_Interface) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetSQLInterfacesResponse_Interface) GetInterface() string {
	if x != nil {
		return x.Interface
	}
	return ""
}

var File_xyz_block_ftl_language_v1_commands_proto protoreflect.FileDescriptor

const file_xyz_block_ftl_language_v1_commands_proto_rawDesc = "" +
	"\n" +
	"(xyz/block/ftl/language/v1/commands.proto\x12\x19xyz.block.ftl.language.v1\x1a\x1cgoogle/protobuf/struct.proto\x1a'xyz/block/ftl/language/v1/service.proto\"6\n" +
	"\x18GetNewModuleFlagsRequest\x12\x1a\n" +
	"\blanguage\x18\x01 \x01(\tR\blanguage\"\xc9\x02\n" +
	"\x19GetNewModuleFlagsResponse\x12O\n" +
	"\x05flags\x18\x01 \x03(\v29.xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.FlagR\x05flags\x1a\xda\x01\n" +
	"\x04Flag\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x12\n" +
	"\x04help\x18\x02 \x01(\tR\x04help\x12\x19\n" +
	"\x05envar\x18\x03 \x01(\tH\x00R\x05envar\x88\x01\x01\x12\x19\n" +
	"\x05short\x18\x04 \x01(\tH\x01R\x05short\x88\x01\x01\x12%\n" +
	"\vplaceholder\x18\x05 \x01(\tH\x02R\vplaceholder\x88\x01\x01\x12\x1d\n" +
	"\adefault\x18\x06 \x01(\tH\x03R\adefault\x88\x01\x01B\b\n" +
	"\x06_envarB\b\n" +
	"\x06_shortB\x0e\n" +
	"\f_placeholderB\n" +
	"\n" +
	"\b_default\"\xb8\x01\n" +
	"\x10NewModuleRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x10\n" +
	"\x03dir\x18\x02 \x01(\tR\x03dir\x12O\n" +
	"\x0eproject_config\x18\x03 \x01(\v2(.xyz.block.ftl.language.v1.ProjectConfigR\rprojectConfig\x12-\n" +
	"\x05flags\x18\x04 \x01(\v2\x17.google.protobuf.StructR\x05flags\"\x13\n" +
	"\x11NewModuleResponse\"2\n" +
	"\x1eGetModuleConfigDefaultsRequest\x12\x10\n" +
	"\x03dir\x18\x01 \x01(\tR\x03dir\"\xd0\x02\n" +
	"\x1fGetModuleConfigDefaultsResponse\x12\x1d\n" +
	"\n" +
	"deploy_dir\x18\x01 \x01(\tR\tdeployDir\x12\x19\n" +
	"\x05build\x18\x02 \x01(\tH\x00R\x05build\x88\x01\x01\x12)\n" +
	"\x0edev_mode_build\x18\x03 \x01(\tH\x01R\fdevModeBuild\x88\x01\x01\x12\"\n" +
	"\n" +
	"build_lock\x18\x04 \x01(\tH\x02R\tbuildLock\x88\x01\x01\x12\x14\n" +
	"\x05watch\x18\x06 \x03(\tR\x05watch\x12@\n" +
	"\x0flanguage_config\x18\a \x01(\v2\x17.google.protobuf.StructR\x0elanguageConfig\x12 \n" +
	"\fsql_root_dir\x18\b \x01(\tR\n" +
	"sqlRootDirB\b\n" +
	"\x06_buildB\x11\n" +
	"\x0f_dev_mode_buildB\r\n" +
	"\v_build_lock\"Z\n" +
	"\x17GetSQLInterfacesRequest\x12?\n" +
	"\x06config\x18\x01 \x01(\v2'.xyz.block.ftl.language.v1.ModuleConfigR\x06config\"\xb8\x01\n" +
	"\x18GetSQLInterfacesResponse\x12]\n" +
	"\n" +
	"interfaces\x18\x01 \x03(\v2=.xyz.block.ftl.language.v1.GetSQLInterfacesResponse.InterfaceR\n" +
	"interfaces\x1a=\n" +
	"\tInterface\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x1c\n" +
	"\tinterface\x18\x02 \x01(\tR\tinterface2\x90\x04\n" +
	"\x16LanguageCommandService\x12~\n" +
	"\x11GetNewModuleFlags\x123.xyz.block.ftl.language.v1.GetNewModuleFlagsRequest\x1a4.xyz.block.ftl.language.v1.GetNewModuleFlagsResponse\x12f\n" +
	"\tNewModule\x12+.xyz.block.ftl.language.v1.NewModuleRequest\x1a,.xyz.block.ftl.language.v1.NewModuleResponse\x12\x90\x01\n" +
	"\x17GetModuleConfigDefaults\x129.xyz.block.ftl.language.v1.GetModuleConfigDefaultsRequest\x1a:.xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse\x12{\n" +
	"\x10GetSQLInterfaces\x122.xyz.block.ftl.language.v1.GetSQLInterfacesRequest\x1a3.xyz.block.ftl.language.v1.GetSQLInterfacesResponseBLP\x01ZHgithub.com/block/ftl/backend/protos/xyz/block/ftl/language/v1;languagepbb\x06proto3"

var (
	file_xyz_block_ftl_language_v1_commands_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_language_v1_commands_proto_rawDescData []byte
)

func file_xyz_block_ftl_language_v1_commands_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_language_v1_commands_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_language_v1_commands_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_language_v1_commands_proto_rawDesc), len(file_xyz_block_ftl_language_v1_commands_proto_rawDesc)))
	})
	return file_xyz_block_ftl_language_v1_commands_proto_rawDescData
}

var file_xyz_block_ftl_language_v1_commands_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_xyz_block_ftl_language_v1_commands_proto_goTypes = []any{
	(*GetNewModuleFlagsRequest)(nil),           // 0: xyz.block.ftl.language.v1.GetNewModuleFlagsRequest
	(*GetNewModuleFlagsResponse)(nil),          // 1: xyz.block.ftl.language.v1.GetNewModuleFlagsResponse
	(*NewModuleRequest)(nil),                   // 2: xyz.block.ftl.language.v1.NewModuleRequest
	(*NewModuleResponse)(nil),                  // 3: xyz.block.ftl.language.v1.NewModuleResponse
	(*GetModuleConfigDefaultsRequest)(nil),     // 4: xyz.block.ftl.language.v1.GetModuleConfigDefaultsRequest
	(*GetModuleConfigDefaultsResponse)(nil),    // 5: xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse
	(*GetSQLInterfacesRequest)(nil),            // 6: xyz.block.ftl.language.v1.GetSQLInterfacesRequest
	(*GetSQLInterfacesResponse)(nil),           // 7: xyz.block.ftl.language.v1.GetSQLInterfacesResponse
	(*GetNewModuleFlagsResponse_Flag)(nil),     // 8: xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.Flag
	(*GetSQLInterfacesResponse_Interface)(nil), // 9: xyz.block.ftl.language.v1.GetSQLInterfacesResponse.Interface
	(*ProjectConfig)(nil),                      // 10: xyz.block.ftl.language.v1.ProjectConfig
	(*structpb.Struct)(nil),                    // 11: google.protobuf.Struct
	(*ModuleConfig)(nil),                       // 12: xyz.block.ftl.language.v1.ModuleConfig
}
var file_xyz_block_ftl_language_v1_commands_proto_depIdxs = []int32{
	8,  // 0: xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.flags:type_name -> xyz.block.ftl.language.v1.GetNewModuleFlagsResponse.Flag
	10, // 1: xyz.block.ftl.language.v1.NewModuleRequest.project_config:type_name -> xyz.block.ftl.language.v1.ProjectConfig
	11, // 2: xyz.block.ftl.language.v1.NewModuleRequest.flags:type_name -> google.protobuf.Struct
	11, // 3: xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse.language_config:type_name -> google.protobuf.Struct
	12, // 4: xyz.block.ftl.language.v1.GetSQLInterfacesRequest.config:type_name -> xyz.block.ftl.language.v1.ModuleConfig
	9,  // 5: xyz.block.ftl.language.v1.GetSQLInterfacesResponse.interfaces:type_name -> xyz.block.ftl.language.v1.GetSQLInterfacesResponse.Interface
	0,  // 6: xyz.block.ftl.language.v1.LanguageCommandService.GetNewModuleFlags:input_type -> xyz.block.ftl.language.v1.GetNewModuleFlagsRequest
	2,  // 7: xyz.block.ftl.language.v1.LanguageCommandService.NewModule:input_type -> xyz.block.ftl.language.v1.NewModuleRequest
	4,  // 8: xyz.block.ftl.language.v1.LanguageCommandService.GetModuleConfigDefaults:input_type -> xyz.block.ftl.language.v1.GetModuleConfigDefaultsRequest
	6,  // 9: xyz.block.ftl.language.v1.LanguageCommandService.GetSQLInterfaces:input_type -> xyz.block.ftl.language.v1.GetSQLInterfacesRequest
	1,  // 10: xyz.block.ftl.language.v1.LanguageCommandService.GetNewModuleFlags:output_type -> xyz.block.ftl.language.v1.GetNewModuleFlagsResponse
	3,  // 11: xyz.block.ftl.language.v1.LanguageCommandService.NewModule:output_type -> xyz.block.ftl.language.v1.NewModuleResponse
	5,  // 12: xyz.block.ftl.language.v1.LanguageCommandService.GetModuleConfigDefaults:output_type -> xyz.block.ftl.language.v1.GetModuleConfigDefaultsResponse
	7,  // 13: xyz.block.ftl.language.v1.LanguageCommandService.GetSQLInterfaces:output_type -> xyz.block.ftl.language.v1.GetSQLInterfacesResponse
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_language_v1_commands_proto_init() }
func file_xyz_block_ftl_language_v1_commands_proto_init() {
	if File_xyz_block_ftl_language_v1_commands_proto != nil {
		return
	}
	file_xyz_block_ftl_language_v1_service_proto_init()
	file_xyz_block_ftl_language_v1_commands_proto_msgTypes[5].OneofWrappers = []any{}
	file_xyz_block_ftl_language_v1_commands_proto_msgTypes[8].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_xyz_block_ftl_language_v1_commands_proto_rawDesc), len(file_xyz_block_ftl_language_v1_commands_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_language_v1_commands_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_language_v1_commands_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_language_v1_commands_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_language_v1_commands_proto = out.File
	file_xyz_block_ftl_language_v1_commands_proto_goTypes = nil
	file_xyz_block_ftl_language_v1_commands_proto_depIdxs = nil
}
