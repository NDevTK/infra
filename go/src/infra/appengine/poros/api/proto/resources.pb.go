// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.1
// source: poros/api/proto/resources.proto

package proto

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

type AssetModel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier of the asset
	AssetId string `protobuf:"bytes,1,opt,name=asset_id,json=assetId,proto3" json:"asset_id,omitempty"`
	// Name of the asset
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	//Type of the Asset (active_directory, etc)
	AssetType string `protobuf:"bytes,3,opt,name=asset_type,json=assetType,proto3" json:"asset_type,omitempty"`
	// Description of the asset
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	// User who created the record.
	CreatedBy string `protobuf:"bytes,5,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	// Timestamp for the creation of the record.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Timestamp for the last update of the record.
	ModifiedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
	// User who modified the record.
	ModifiedBy string `protobuf:"bytes,8,opt,name=modified_by,json=modifiedBy,proto3" json:"modified_by,omitempty"`
}

func (x *AssetModel) Reset() {
	*x = AssetModel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_poros_api_proto_resources_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssetModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssetModel) ProtoMessage() {}

func (x *AssetModel) ProtoReflect() protoreflect.Message {
	mi := &file_poros_api_proto_resources_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssetModel.ProtoReflect.Descriptor instead.
func (*AssetModel) Descriptor() ([]byte, []int) {
	return file_poros_api_proto_resources_proto_rawDescGZIP(), []int{0}
}

func (x *AssetModel) GetAssetId() string {
	if x != nil {
		return x.AssetId
	}
	return ""
}

func (x *AssetModel) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AssetModel) GetAssetType() string {
	if x != nil {
		return x.AssetType
	}
	return ""
}

func (x *AssetModel) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *AssetModel) GetCreatedBy() string {
	if x != nil {
		return x.CreatedBy
	}
	return ""
}

func (x *AssetModel) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *AssetModel) GetModifiedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ModifiedAt
	}
	return nil
}

func (x *AssetModel) GetModifiedBy() string {
	if x != nil {
		return x.ModifiedBy
	}
	return ""
}

type ResourceModel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier of the resource
	ResourceId string `protobuf:"bytes,1,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	// Name of the resource
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Description of the resource
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// Type of the resource
	Type string `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	// Operating system of the machine (If Type is machine)
	OperatingSystem string `protobuf:"bytes,5,opt,name=operating_system,json=operatingSystem,proto3" json:"operating_system,omitempty"`
	// TODO: crbug/1328854 move the ImageProject and ImageFamily as part of property later phases
	// Gcp project where image resides
	ImageProject string `protobuf:"bytes,6,opt,name=image_project,json=imageProject,proto3" json:"image_project,omitempty"`
	// Family accociated with the gcp project where image resides
	ImageFamily string `protobuf:"bytes,7,opt,name=image_family,json=imageFamily,proto3" json:"image_family,omitempty"`
	// User who created the resource.
	CreatedBy string `protobuf:"bytes,8,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	// Timestamp for the creation of the record.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Timestamp for the last update of the record.
	ModifiedAt *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
	// User who modified the record.
	ModifiedBy string `protobuf:"bytes,11,opt,name=modified_by,json=modifiedBy,proto3" json:"modified_by,omitempty"`
}

func (x *ResourceModel) Reset() {
	*x = ResourceModel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_poros_api_proto_resources_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceModel) ProtoMessage() {}

func (x *ResourceModel) ProtoReflect() protoreflect.Message {
	mi := &file_poros_api_proto_resources_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceModel.ProtoReflect.Descriptor instead.
func (*ResourceModel) Descriptor() ([]byte, []int) {
	return file_poros_api_proto_resources_proto_rawDescGZIP(), []int{1}
}

func (x *ResourceModel) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (x *ResourceModel) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ResourceModel) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ResourceModel) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *ResourceModel) GetOperatingSystem() string {
	if x != nil {
		return x.OperatingSystem
	}
	return ""
}

func (x *ResourceModel) GetImageProject() string {
	if x != nil {
		return x.ImageProject
	}
	return ""
}

func (x *ResourceModel) GetImageFamily() string {
	if x != nil {
		return x.ImageFamily
	}
	return ""
}

func (x *ResourceModel) GetCreatedBy() string {
	if x != nil {
		return x.CreatedBy
	}
	return ""
}

func (x *ResourceModel) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *ResourceModel) GetModifiedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ModifiedAt
	}
	return nil
}

func (x *ResourceModel) GetModifiedBy() string {
	if x != nil {
		return x.ModifiedBy
	}
	return ""
}

type AssetResourceModel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier of the entity
	AssetResourceId string `protobuf:"bytes,1,opt,name=asset_resource_id,json=assetResourceId,proto3" json:"asset_resource_id,omitempty"`
	// Identifier of the asset associated with the entity
	AssetId string `protobuf:"bytes,2,opt,name=asset_id,json=assetId,proto3" json:"asset_id,omitempty"`
	// Identifier of the resource associated with the entity
	ResourceId string `protobuf:"bytes,3,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	// Alias name of the entity
	AliasName string `protobuf:"bytes,4,opt,name=alias_name,json=aliasName,proto3" json:"alias_name,omitempty"`
	// User who created the record.
	CreatedBy string `protobuf:"bytes,5,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	// Timestamp for the creation of the record.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Timestamp for the last update of the record.
	ModifiedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
	// User who modified the record.
	ModifiedBy string `protobuf:"bytes,8,opt,name=modified_by,json=modifiedBy,proto3" json:"modified_by,omitempty"`
	// Flag to denote whether this AssetResource is default
	Default bool `protobuf:"varint,9,opt,name=default,proto3" json:"default,omitempty"`
}

func (x *AssetResourceModel) Reset() {
	*x = AssetResourceModel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_poros_api_proto_resources_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssetResourceModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssetResourceModel) ProtoMessage() {}

func (x *AssetResourceModel) ProtoReflect() protoreflect.Message {
	mi := &file_poros_api_proto_resources_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssetResourceModel.ProtoReflect.Descriptor instead.
func (*AssetResourceModel) Descriptor() ([]byte, []int) {
	return file_poros_api_proto_resources_proto_rawDescGZIP(), []int{2}
}

func (x *AssetResourceModel) GetAssetResourceId() string {
	if x != nil {
		return x.AssetResourceId
	}
	return ""
}

func (x *AssetResourceModel) GetAssetId() string {
	if x != nil {
		return x.AssetId
	}
	return ""
}

func (x *AssetResourceModel) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (x *AssetResourceModel) GetAliasName() string {
	if x != nil {
		return x.AliasName
	}
	return ""
}

func (x *AssetResourceModel) GetCreatedBy() string {
	if x != nil {
		return x.CreatedBy
	}
	return ""
}

func (x *AssetResourceModel) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *AssetResourceModel) GetModifiedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ModifiedAt
	}
	return nil
}

func (x *AssetResourceModel) GetModifiedBy() string {
	if x != nil {
		return x.ModifiedBy
	}
	return ""
}

func (x *AssetResourceModel) GetDefault() bool {
	if x != nil {
		return x.Default
	}
	return false
}

type AssetInstanceModel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier of the AssetInstance
	AssetInstanceId string `protobuf:"bytes,1,opt,name=asset_instance_id,json=assetInstanceId,proto3" json:"asset_instance_id,omitempty"`
	// AssetId associated with the AssetInstance
	AssetId string `protobuf:"bytes,2,opt,name=asset_id,json=assetId,proto3" json:"asset_id,omitempty"`
	// Status of the AssetInstance
	Status string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	// Project Id associated with the asset instance
	ProjectId string `protobuf:"bytes,4,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	// User who created the record.
	CreatedBy string `protobuf:"bytes,5,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	// Timestamp for the creation of the record.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// Timestamp for the last update of the record.
	ModifiedAt *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=modified_at,json=modifiedAt,proto3" json:"modified_at,omitempty"`
	// User who modified the record.
	ModifiedBy string `protobuf:"bytes,8,opt,name=modified_by,json=modifiedBy,proto3" json:"modified_by,omitempty"`
	// Timestamp to delete the machines
	DeleteAt *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=delete_at,json=deleteAt,proto3" json:"delete_at,omitempty"`
}

func (x *AssetInstanceModel) Reset() {
	*x = AssetInstanceModel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_poros_api_proto_resources_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssetInstanceModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssetInstanceModel) ProtoMessage() {}

func (x *AssetInstanceModel) ProtoReflect() protoreflect.Message {
	mi := &file_poros_api_proto_resources_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssetInstanceModel.ProtoReflect.Descriptor instead.
func (*AssetInstanceModel) Descriptor() ([]byte, []int) {
	return file_poros_api_proto_resources_proto_rawDescGZIP(), []int{3}
}

func (x *AssetInstanceModel) GetAssetInstanceId() string {
	if x != nil {
		return x.AssetInstanceId
	}
	return ""
}

func (x *AssetInstanceModel) GetAssetId() string {
	if x != nil {
		return x.AssetId
	}
	return ""
}

func (x *AssetInstanceModel) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *AssetInstanceModel) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *AssetInstanceModel) GetCreatedBy() string {
	if x != nil {
		return x.CreatedBy
	}
	return ""
}

func (x *AssetInstanceModel) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *AssetInstanceModel) GetModifiedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ModifiedAt
	}
	return nil
}

func (x *AssetInstanceModel) GetModifiedBy() string {
	if x != nil {
		return x.ModifiedBy
	}
	return ""
}

func (x *AssetInstanceModel) GetDeleteAt() *timestamppb.Timestamp {
	if x != nil {
		return x.DeleteAt
	}
	return nil
}

var File_poros_api_proto_resources_proto protoreflect.FileDescriptor

var file_poros_api_proto_resources_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x70, 0x6f, 0x72, 0x6f, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x6f, 0x72, 0x6f, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb4, 0x02, 0x0a, 0x0a, 0x41, 0x73,
	0x73, 0x65, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x19, 0x0a, 0x08, 0x61, 0x73, 0x73, 0x65,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x73, 0x73, 0x65,
	0x74, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x73, 0x73, 0x65, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x73, 0x73,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x42, 0x79, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x41, 0x74, 0x12, 0x3b, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61,
	0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x74, 0x12,
	0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42, 0x79,
	0x22, 0xa5, 0x03, 0x0a, 0x0d, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4d, 0x6f, 0x64,
	0x65, 0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x29, 0x0a,
	0x10, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6e, 0x67, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x12, 0x23, 0x0a, 0x0d, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x21, 0x0a,
	0x0c, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x5f, 0x66, 0x61, 0x6d, 0x69, 0x6c, 0x79, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x46, 0x61, 0x6d, 0x69, 0x6c, 0x79,
	0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x42, 0x79, 0x12,
	0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x3b, 0x0a, 0x0b, 0x6d, 0x6f,
	0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x6d, 0x6f, 0x64,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x6f,
	0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42, 0x79, 0x22, 0xed, 0x02, 0x0a, 0x12, 0x41, 0x73, 0x73,
	0x65, 0x74, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12,
	0x2a, 0x0a, 0x11, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x61, 0x73, 0x73, 0x65,
	0x74, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x6c, 0x69, 0x61, 0x73,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x6c, 0x69,
	0x61, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x5f, 0x62, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x42, 0x79, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x3b, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1f, 0x0a,
	0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42, 0x79, 0x12, 0x18,
	0x0a, 0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x22, 0x83, 0x03, 0x0a, 0x12, 0x41, 0x73, 0x73,
	0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12,
	0x2a, 0x0a, 0x11, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x61, 0x73, 0x73, 0x65,
	0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x1d, 0x0a,
	0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x42, 0x79, 0x12, 0x39, 0x0a, 0x0a,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x3b, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x41, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x5f, 0x62, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x42, 0x79, 0x12, 0x37, 0x0a, 0x09, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x5f,
	0x61, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x74, 0x42, 0x27,
	0x5a, 0x25, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x70, 0x6f, 0x72, 0x6f, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_poros_api_proto_resources_proto_rawDescOnce sync.Once
	file_poros_api_proto_resources_proto_rawDescData = file_poros_api_proto_resources_proto_rawDesc
)

func file_poros_api_proto_resources_proto_rawDescGZIP() []byte {
	file_poros_api_proto_resources_proto_rawDescOnce.Do(func() {
		file_poros_api_proto_resources_proto_rawDescData = protoimpl.X.CompressGZIP(file_poros_api_proto_resources_proto_rawDescData)
	})
	return file_poros_api_proto_resources_proto_rawDescData
}

var file_poros_api_proto_resources_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_poros_api_proto_resources_proto_goTypes = []interface{}{
	(*AssetModel)(nil),            // 0: poros.AssetModel
	(*ResourceModel)(nil),         // 1: poros.ResourceModel
	(*AssetResourceModel)(nil),    // 2: poros.AssetResourceModel
	(*AssetInstanceModel)(nil),    // 3: poros.AssetInstanceModel
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_poros_api_proto_resources_proto_depIdxs = []int32{
	4, // 0: poros.AssetModel.created_at:type_name -> google.protobuf.Timestamp
	4, // 1: poros.AssetModel.modified_at:type_name -> google.protobuf.Timestamp
	4, // 2: poros.ResourceModel.created_at:type_name -> google.protobuf.Timestamp
	4, // 3: poros.ResourceModel.modified_at:type_name -> google.protobuf.Timestamp
	4, // 4: poros.AssetResourceModel.created_at:type_name -> google.protobuf.Timestamp
	4, // 5: poros.AssetResourceModel.modified_at:type_name -> google.protobuf.Timestamp
	4, // 6: poros.AssetInstanceModel.created_at:type_name -> google.protobuf.Timestamp
	4, // 7: poros.AssetInstanceModel.modified_at:type_name -> google.protobuf.Timestamp
	4, // 8: poros.AssetInstanceModel.delete_at:type_name -> google.protobuf.Timestamp
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_poros_api_proto_resources_proto_init() }
func file_poros_api_proto_resources_proto_init() {
	if File_poros_api_proto_resources_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_poros_api_proto_resources_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssetModel); i {
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
		file_poros_api_proto_resources_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceModel); i {
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
		file_poros_api_proto_resources_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssetResourceModel); i {
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
		file_poros_api_proto_resources_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssetInstanceModel); i {
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
			RawDescriptor: file_poros_api_proto_resources_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_poros_api_proto_resources_proto_goTypes,
		DependencyIndexes: file_poros_api_proto_resources_proto_depIdxs,
		MessageInfos:      file_poros_api_proto_resources_proto_msgTypes,
	}.Build()
	File_poros_api_proto_resources_proto = out.File
	file_poros_api_proto_resources_proto_rawDesc = nil
	file_poros_api_proto_resources_proto_goTypes = nil
	file_poros_api_proto_resources_proto_depIdxs = nil
}
