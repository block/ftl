// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: xyz/block/ftl/provisioner/v1beta1/service.proto

package provisionerpb

import (
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_xyz_block_ftl_provisioner_v1beta1_service_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_provisioner_v1beta1_service_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x21, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x1a, 0x21, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f,
	0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x32, 0xa9, 0x05, 0x0a, 0x12, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69,
	0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x12, 0x4b, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x1f, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x20, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x69, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61,
	0x63, 0x74, 0x44, 0x69, 0x66, 0x66, 0x73, 0x12, 0x29, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72,
	0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x44, 0x69, 0x66, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63,
	0x74, 0x44, 0x69, 0x66, 0x66, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x63,
	0x0a, 0x0e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74,
	0x12, 0x27, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61,
	0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x69, 0x0a, 0x10, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70,
	0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x29, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d,
	0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x12, 0x25,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44,
	0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x60, 0x0a,
	0x0d, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x12, 0x26,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63,
	0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x57, 0x50, 0x01, 0x5a, 0x53, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_xyz_block_ftl_provisioner_v1beta1_service_proto_goTypes = []any{
	(*v1.PingRequest)(nil),              // 0: xyz.block.ftl.v1.PingRequest
	(*v1.StatusRequest)(nil),            // 1: xyz.block.ftl.v1.StatusRequest
	(*v1.GetArtefactDiffsRequest)(nil),  // 2: xyz.block.ftl.v1.GetArtefactDiffsRequest
	(*v1.UploadArtefactRequest)(nil),    // 3: xyz.block.ftl.v1.UploadArtefactRequest
	(*v1.CreateDeploymentRequest)(nil),  // 4: xyz.block.ftl.v1.CreateDeploymentRequest
	(*v1.UpdateDeployRequest)(nil),      // 5: xyz.block.ftl.v1.UpdateDeployRequest
	(*v1.ReplaceDeployRequest)(nil),     // 6: xyz.block.ftl.v1.ReplaceDeployRequest
	(*v1.PingResponse)(nil),             // 7: xyz.block.ftl.v1.PingResponse
	(*v1.StatusResponse)(nil),           // 8: xyz.block.ftl.v1.StatusResponse
	(*v1.GetArtefactDiffsResponse)(nil), // 9: xyz.block.ftl.v1.GetArtefactDiffsResponse
	(*v1.UploadArtefactResponse)(nil),   // 10: xyz.block.ftl.v1.UploadArtefactResponse
	(*v1.CreateDeploymentResponse)(nil), // 11: xyz.block.ftl.v1.CreateDeploymentResponse
	(*v1.UpdateDeployResponse)(nil),     // 12: xyz.block.ftl.v1.UpdateDeployResponse
	(*v1.ReplaceDeployResponse)(nil),    // 13: xyz.block.ftl.v1.ReplaceDeployResponse
}
var file_xyz_block_ftl_provisioner_v1beta1_service_proto_depIdxs = []int32{
	0,  // 0: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	1,  // 1: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.Status:input_type -> xyz.block.ftl.v1.StatusRequest
	2,  // 2: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.GetArtefactDiffs:input_type -> xyz.block.ftl.v1.GetArtefactDiffsRequest
	3,  // 3: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.UploadArtefact:input_type -> xyz.block.ftl.v1.UploadArtefactRequest
	4,  // 4: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.CreateDeployment:input_type -> xyz.block.ftl.v1.CreateDeploymentRequest
	5,  // 5: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.UpdateDeploy:input_type -> xyz.block.ftl.v1.UpdateDeployRequest
	6,  // 6: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.ReplaceDeploy:input_type -> xyz.block.ftl.v1.ReplaceDeployRequest
	7,  // 7: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	8,  // 8: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.Status:output_type -> xyz.block.ftl.v1.StatusResponse
	9,  // 9: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.GetArtefactDiffs:output_type -> xyz.block.ftl.v1.GetArtefactDiffsResponse
	10, // 10: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.UploadArtefact:output_type -> xyz.block.ftl.v1.UploadArtefactResponse
	11, // 11: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.CreateDeployment:output_type -> xyz.block.ftl.v1.CreateDeploymentResponse
	12, // 12: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.UpdateDeploy:output_type -> xyz.block.ftl.v1.UpdateDeployResponse
	13, // 13: xyz.block.ftl.provisioner.v1beta1.ProvisionerService.ReplaceDeploy:output_type -> xyz.block.ftl.v1.ReplaceDeployResponse
	7,  // [7:14] is the sub-list for method output_type
	0,  // [0:7] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_provisioner_v1beta1_service_proto_init() }
func file_xyz_block_ftl_provisioner_v1beta1_service_proto_init() {
	if File_xyz_block_ftl_provisioner_v1beta1_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_provisioner_v1beta1_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_provisioner_v1beta1_service_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_provisioner_v1beta1_service_proto_depIdxs,
	}.Build()
	File_xyz_block_ftl_provisioner_v1beta1_service_proto = out.File
	file_xyz_block_ftl_provisioner_v1beta1_service_proto_rawDesc = nil
	file_xyz_block_ftl_provisioner_v1beta1_service_proto_goTypes = nil
	file_xyz_block_ftl_provisioner_v1beta1_service_proto_depIdxs = nil
}
