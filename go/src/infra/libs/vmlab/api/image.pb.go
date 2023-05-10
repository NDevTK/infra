// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.7
// source: infra/libs/vmlab/api/image.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Status tracks the relevant status of VM lab use case: READY, PENDING,
// FAILED, DELETING, NOT_FOUND. Unknown statuses are mapped to UNKNOWN.
type GceImage_Status int32

const (
	GceImage_UNKNOWN   GceImage_Status = 0
	GceImage_READY     GceImage_Status = 1
	GceImage_PENDING   GceImage_Status = 2
	GceImage_FAILED    GceImage_Status = 3
	GceImage_DELETING  GceImage_Status = 4
	GceImage_NOT_FOUND GceImage_Status = 5
)

// Enum value maps for GceImage_Status.
var (
	GceImage_Status_name = map[int32]string{
		0: "UNKNOWN",
		1: "READY",
		2: "PENDING",
		3: "FAILED",
		4: "DELETING",
		5: "NOT_FOUND",
	}
	GceImage_Status_value = map[string]int32{
		"UNKNOWN":   0,
		"READY":     1,
		"PENDING":   2,
		"FAILED":    3,
		"DELETING":  4,
		"NOT_FOUND": 5,
	}
)

func (x GceImage_Status) Enum() *GceImage_Status {
	p := new(GceImage_Status)
	*p = x
	return p
}

func (x GceImage_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GceImage_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_libs_vmlab_api_image_proto_enumTypes[0].Descriptor()
}

func (GceImage_Status) Type() protoreflect.EnumType {
	return &file_infra_libs_vmlab_api_image_proto_enumTypes[0]
}

func (x GceImage_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GceImage_Status.Descriptor instead.
func (GceImage_Status) EnumDescriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_image_proto_rawDescGZIP(), []int{0, 0}
}

// GceImage represents an existing GCE image.
type GceImage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of GCP project. E.g. betty-cloud-prototype
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// Name of GCE image.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Readiness of GCE image.
	Status GceImage_Status `protobuf:"varint,3,opt,name=status,proto3,enum=vmlab.api.GceImage_Status" json:"status,omitempty"`
	// GCS URL of source image to be imported.
	Source string `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
	// Labels of the image, including build-type, board, milestone.
	Labels map[string]string `protobuf:"bytes,5,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Description of the image which contains source image in GS bucket.
	Description string `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	// Timestamp when the image is created in UTC.
	TimeCreated *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=timeCreated,proto3" json:"timeCreated,omitempty"`
}

func (x *GceImage) Reset() {
	*x = GceImage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_libs_vmlab_api_image_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GceImage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GceImage) ProtoMessage() {}

func (x *GceImage) ProtoReflect() protoreflect.Message {
	mi := &file_infra_libs_vmlab_api_image_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GceImage.ProtoReflect.Descriptor instead.
func (*GceImage) Descriptor() ([]byte, []int) {
	return file_infra_libs_vmlab_api_image_proto_rawDescGZIP(), []int{0}
}

func (x *GceImage) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *GceImage) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GceImage) GetStatus() GceImage_Status {
	if x != nil {
		return x.Status
	}
	return GceImage_UNKNOWN
}

func (x *GceImage) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *GceImage) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *GceImage) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *GceImage) GetTimeCreated() *timestamppb.Timestamp {
	if x != nil {
		return x.TimeCreated
	}
	return nil
}

var File_infra_libs_vmlab_api_image_proto protoreflect.FileDescriptor

var file_infra_libs_vmlab_api_image_proto_rawDesc = []byte{
	0x0a, 0x20, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6c, 0x69, 0x62, 0x73, 0x2f, 0x76, 0x6d, 0x6c,
	0x61, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x09, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb0,
	0x03, 0x0a, 0x08, 0x47, 0x63, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x32, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x76, 0x6d, 0x6c, 0x61,
	0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x63, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18,
	0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x47, 0x63, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x3c, 0x0a, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x1a, 0x39,
	0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x56, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00,
	0x12, 0x09, 0x0a, 0x05, 0x52, 0x45, 0x41, 0x44, 0x59, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x50,
	0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c,
	0x45, 0x44, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x49, 0x4e, 0x47,
	0x10, 0x04, 0x12, 0x0d, 0x0a, 0x09, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10,
	0x05, 0x42, 0x16, 0x5a, 0x14, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6c, 0x69, 0x62, 0x73, 0x2f,
	0x76, 0x6d, 0x6c, 0x61, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_infra_libs_vmlab_api_image_proto_rawDescOnce sync.Once
	file_infra_libs_vmlab_api_image_proto_rawDescData = file_infra_libs_vmlab_api_image_proto_rawDesc
)

func file_infra_libs_vmlab_api_image_proto_rawDescGZIP() []byte {
	file_infra_libs_vmlab_api_image_proto_rawDescOnce.Do(func() {
		file_infra_libs_vmlab_api_image_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_libs_vmlab_api_image_proto_rawDescData)
	})
	return file_infra_libs_vmlab_api_image_proto_rawDescData
}

var file_infra_libs_vmlab_api_image_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_infra_libs_vmlab_api_image_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_infra_libs_vmlab_api_image_proto_goTypes = []interface{}{
	(GceImage_Status)(0),          // 0: vmlab.api.GceImage.Status
	(*GceImage)(nil),              // 1: vmlab.api.GceImage
	nil,                           // 2: vmlab.api.GceImage.LabelsEntry
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_infra_libs_vmlab_api_image_proto_depIdxs = []int32{
	0, // 0: vmlab.api.GceImage.status:type_name -> vmlab.api.GceImage.Status
	2, // 1: vmlab.api.GceImage.labels:type_name -> vmlab.api.GceImage.LabelsEntry
	3, // 2: vmlab.api.GceImage.timeCreated:type_name -> google.protobuf.Timestamp
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_infra_libs_vmlab_api_image_proto_init() }
func file_infra_libs_vmlab_api_image_proto_init() {
	if File_infra_libs_vmlab_api_image_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_libs_vmlab_api_image_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GceImage); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_infra_libs_vmlab_api_image_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_libs_vmlab_api_image_proto_goTypes,
		DependencyIndexes: file_infra_libs_vmlab_api_image_proto_depIdxs,
		EnumInfos:         file_infra_libs_vmlab_api_image_proto_enumTypes,
		MessageInfos:      file_infra_libs_vmlab_api_image_proto_msgTypes,
	}.Build()
	File_infra_libs_vmlab_api_image_proto = out.File
	file_infra_libs_vmlab_api_image_proto_rawDesc = nil
	file_infra_libs_vmlab_api_image_proto_goTypes = nil
	file_infra_libs_vmlab_api_image_proto_depIdxs = nil
}
