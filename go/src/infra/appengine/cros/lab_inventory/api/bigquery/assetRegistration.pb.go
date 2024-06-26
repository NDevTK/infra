// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.7
// source: infra/appengine/cros/lab_inventory/api/bigquery/assetRegistration.proto

package apibq

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	protos "infra/libs/fleet/protos"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RegisteredAsset struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Asset       *protos.ChopsAsset     `protobuf:"bytes,2,opt,name=asset,proto3" json:"asset,omitempty"`
	UpdatedTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=updated_time,json=updatedTime,proto3" json:"updated_time,omitempty"`
}

func (x *RegisteredAsset) Reset() {
	*x = RegisteredAsset{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisteredAsset) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisteredAsset) ProtoMessage() {}

func (x *RegisteredAsset) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisteredAsset.ProtoReflect.Descriptor instead.
func (*RegisteredAsset) Descriptor() ([]byte, []int) {
	return file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescGZIP(), []int{0}
}

func (x *RegisteredAsset) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RegisteredAsset) GetAsset() *protos.ChopsAsset {
	if x != nil {
		return x.Asset
	}
	return nil
}

func (x *RegisteredAsset) GetUpdatedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedTime
	}
	return nil
}

var File_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto protoreflect.FileDescriptor

var file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDesc = []byte{
	0x0a, 0x47, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x5f, 0x69, 0x6e, 0x76, 0x65, 0x6e,
	0x74, 0x6f, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x69, 0x67, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x61, 0x70, 0x69, 0x62, 0x71,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x23, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6c, 0x69, 0x62, 0x73, 0x2f, 0x66, 0x6c,
	0x65, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x89, 0x01, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x65, 0x64, 0x41, 0x73, 0x73, 0x65, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x27, 0x0a, 0x05, 0x61, 0x73,
	0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x66, 0x6c, 0x65, 0x65,
	0x74, 0x2e, 0x43, 0x68, 0x6f, 0x70, 0x73, 0x41, 0x73, 0x73, 0x65, 0x74, 0x52, 0x05, 0x61, 0x73,
	0x73, 0x65, 0x74, 0x12, 0x3d, 0x0a, 0x0c, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x54, 0x69,
	0x6d, 0x65, 0x42, 0x37, 0x5a, 0x35, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65,
	0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x5f, 0x69,
	0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x69, 0x67,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x3b, 0x61, 0x70, 0x69, 0x62, 0x71, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescOnce sync.Once
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescData = file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDesc
)

func file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescGZIP() []byte {
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescOnce.Do(func() {
		file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescData)
	})
	return file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDescData
}

var file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_goTypes = []interface{}{
	(*RegisteredAsset)(nil),       // 0: apibq.RegisteredAsset
	(*protos.ChopsAsset)(nil),     // 1: fleet.ChopsAsset
	(*timestamppb.Timestamp)(nil), // 2: google.protobuf.Timestamp
}
var file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_depIdxs = []int32{
	1, // 0: apibq.RegisteredAsset.asset:type_name -> fleet.ChopsAsset
	2, // 1: apibq.RegisteredAsset.updated_time:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_init() }
func file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_init() {
	if File_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisteredAsset); i {
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
			RawDescriptor: file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_goTypes,
		DependencyIndexes: file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_depIdxs,
		MessageInfos:      file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_msgTypes,
	}.Build()
	File_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto = out.File
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_rawDesc = nil
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_goTypes = nil
	file_infra_appengine_cros_lab_inventory_api_bigquery_assetRegistration_proto_depIdxs = nil
}
