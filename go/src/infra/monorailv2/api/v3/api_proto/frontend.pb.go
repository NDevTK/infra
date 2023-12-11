// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.7
// source: api/v3/api_proto/frontend.proto

package api_proto

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// Request message for GatherProjectEnvironment
// Next available tag: 2
type GatherProjectEnvironmentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the project these config environments belong to.
	Parent string `protobuf:"bytes,1,opt,name=parent,proto3" json:"parent,omitempty"`
}

func (x *GatherProjectEnvironmentRequest) Reset() {
	*x = GatherProjectEnvironmentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v3_api_proto_frontend_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GatherProjectEnvironmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatherProjectEnvironmentRequest) ProtoMessage() {}

func (x *GatherProjectEnvironmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v3_api_proto_frontend_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatherProjectEnvironmentRequest.ProtoReflect.Descriptor instead.
func (*GatherProjectEnvironmentRequest) Descriptor() ([]byte, []int) {
	return file_api_v3_api_proto_frontend_proto_rawDescGZIP(), []int{0}
}

func (x *GatherProjectEnvironmentRequest) GetParent() string {
	if x != nil {
		return x.Parent
	}
	return ""
}

// Response message for GatherProjectEnvironment
// Next available tag: 9
type GatherProjectEnvironmentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Project definitions such as display_name and summary.
	Project *Project `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// Configurations of this project such as default search term,
	// default templates for members and non members.
	ProjectConfig *ProjectConfig `protobuf:"bytes,2,opt,name=project_config,json=projectConfig,proto3" json:"project_config,omitempty"`
	// List of statuses that belong to this project.
	Statuses []*StatusDef `protobuf:"bytes,3,rep,name=statuses,proto3" json:"statuses,omitempty"`
	// List of well known labels that belong to this project.
	WellKnownLabels []*LabelDef `protobuf:"bytes,4,rep,name=well_known_labels,json=wellKnownLabels,proto3" json:"well_known_labels,omitempty"`
	// List of components that belong to this project.
	Components []*ComponentDef `protobuf:"bytes,5,rep,name=components,proto3" json:"components,omitempty"`
	// List of custom fields that belong to this project.
	Fields []*FieldDef `protobuf:"bytes,6,rep,name=fields,proto3" json:"fields,omitempty"`
	// List of approval fields that belong to this project.
	ApprovalFields []*ApprovalDef `protobuf:"bytes,7,rep,name=approval_fields,json=approvalFields,proto3" json:"approval_fields,omitempty"`
	// Saved search queries that admins defined for this project.
	SavedQueries []*ProjectSavedQuery `protobuf:"bytes,8,rep,name=saved_queries,json=savedQueries,proto3" json:"saved_queries,omitempty"`
}

func (x *GatherProjectEnvironmentResponse) Reset() {
	*x = GatherProjectEnvironmentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v3_api_proto_frontend_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GatherProjectEnvironmentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatherProjectEnvironmentResponse) ProtoMessage() {}

func (x *GatherProjectEnvironmentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v3_api_proto_frontend_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatherProjectEnvironmentResponse.ProtoReflect.Descriptor instead.
func (*GatherProjectEnvironmentResponse) Descriptor() ([]byte, []int) {
	return file_api_v3_api_proto_frontend_proto_rawDescGZIP(), []int{1}
}

func (x *GatherProjectEnvironmentResponse) GetProject() *Project {
	if x != nil {
		return x.Project
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetProjectConfig() *ProjectConfig {
	if x != nil {
		return x.ProjectConfig
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetStatuses() []*StatusDef {
	if x != nil {
		return x.Statuses
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetWellKnownLabels() []*LabelDef {
	if x != nil {
		return x.WellKnownLabels
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetComponents() []*ComponentDef {
	if x != nil {
		return x.Components
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetFields() []*FieldDef {
	if x != nil {
		return x.Fields
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetApprovalFields() []*ApprovalDef {
	if x != nil {
		return x.ApprovalFields
	}
	return nil
}

func (x *GatherProjectEnvironmentResponse) GetSavedQueries() []*ProjectSavedQuery {
	if x != nil {
		return x.SavedQueries
	}
	return nil
}

// The request message for Frontend.GatherProjectMembershipsForUser.
// Next available tag: 2
type GatherProjectMembershipsForUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the user to request.
	User string `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
}

func (x *GatherProjectMembershipsForUserRequest) Reset() {
	*x = GatherProjectMembershipsForUserRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v3_api_proto_frontend_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GatherProjectMembershipsForUserRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatherProjectMembershipsForUserRequest) ProtoMessage() {}

func (x *GatherProjectMembershipsForUserRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v3_api_proto_frontend_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatherProjectMembershipsForUserRequest.ProtoReflect.Descriptor instead.
func (*GatherProjectMembershipsForUserRequest) Descriptor() ([]byte, []int) {
	return file_api_v3_api_proto_frontend_proto_rawDescGZIP(), []int{2}
}

func (x *GatherProjectMembershipsForUserRequest) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

// The response message for Frontend.GatherProjectMembershipsForUser.
// Next available tag: 2
type GatherProjectMembershipsForUserResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The projects that the user is a member of.
	ProjectMemberships []*ProjectMember `protobuf:"bytes,1,rep,name=project_memberships,json=projectMemberships,proto3" json:"project_memberships,omitempty"`
}

func (x *GatherProjectMembershipsForUserResponse) Reset() {
	*x = GatherProjectMembershipsForUserResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v3_api_proto_frontend_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GatherProjectMembershipsForUserResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GatherProjectMembershipsForUserResponse) ProtoMessage() {}

func (x *GatherProjectMembershipsForUserResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v3_api_proto_frontend_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GatherProjectMembershipsForUserResponse.ProtoReflect.Descriptor instead.
func (*GatherProjectMembershipsForUserResponse) Descriptor() ([]byte, []int) {
	return file_api_v3_api_proto_frontend_proto_rawDescGZIP(), []int{3}
}

func (x *GatherProjectMembershipsForUserResponse) GetProjectMemberships() []*ProjectMember {
	if x != nil {
		return x.ProjectMemberships
	}
	return nil
}

var File_api_v3_api_proto_frontend_proto protoreflect.FileDescriptor

var file_api_v3_api_proto_frontend_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x33, 0x2f, 0x61, 0x70, 0x69, 0x5f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0b, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x26, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x33, 0x2f, 0x61, 0x70, 0x69, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x58, 0x0a, 0x1f, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x35, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1d, 0xe0, 0x41, 0x02, 0xfa, 0x41, 0x17, 0x0a, 0x15, 0x61,
	0x70, 0x69, 0x2e, 0x63, 0x72, 0x62, 0x75, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x50, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x22, 0xfe, 0x03, 0x0a,
	0x20, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x45, 0x6e,
	0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2e, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33,
	0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x12, 0x41, 0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f,
	0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x32, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69,
	0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x44, 0x65, 0x66, 0x52, 0x08,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x12, 0x41, 0x0a, 0x11, 0x77, 0x65, 0x6c, 0x6c,
	0x5f, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x5f, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76,
	0x33, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x44, 0x65, 0x66, 0x52, 0x0f, 0x77, 0x65, 0x6c, 0x6c,
	0x4b, 0x6e, 0x6f, 0x77, 0x6e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x39, 0x0a, 0x0a, 0x63,
	0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x43, 0x6f,
	0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x44, 0x65, 0x66, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x70,
	0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x2d, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69,
	0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x44, 0x65, 0x66, 0x52, 0x06, 0x66,
	0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x41, 0x0a, 0x0f, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x61,
	0x6c, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18,
	0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x41, 0x70, 0x70,
	0x72, 0x6f, 0x76, 0x61, 0x6c, 0x44, 0x65, 0x66, 0x52, 0x0e, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x76,
	0x61, 0x6c, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x43, 0x0a, 0x0d, 0x73, 0x61, 0x76, 0x65,
	0x64, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1e, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x50, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x53, 0x61, 0x76, 0x65, 0x64, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52,
	0x0c, 0x73, 0x61, 0x76, 0x65, 0x64, 0x51, 0x75, 0x65, 0x72, 0x69, 0x65, 0x73, 0x22, 0x55, 0x0a,
	0x26, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x73, 0x46, 0x6f, 0x72, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2b, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x17, 0xfa, 0x41, 0x14, 0x0a, 0x12, 0x61, 0x70, 0x69, 0x2e,
	0x63, 0x72, 0x62, 0x75, 0x67, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04,
	0x75, 0x73, 0x65, 0x72, 0x22, 0x76, 0x0a, 0x27, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x73,
	0x46, 0x6f, 0x72, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x4b, 0x0a, 0x13, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x68, 0x69, 0x70, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d,
	0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x12, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x73, 0x32, 0x96, 0x02, 0x0a,
	0x08, 0x46, 0x72, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x64, 0x12, 0x79, 0x0a, 0x18, 0x47, 0x61, 0x74,
	0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f,
	0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x2c, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c,
	0x2e, 0x76, 0x33, 0x2e, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x45, 0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76,
	0x33, 0x2e, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x45,
	0x6e, 0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x8e, 0x01, 0x0a, 0x1f, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70,
	0x73, 0x46, 0x6f, 0x72, 0x55, 0x73, 0x65, 0x72, 0x12, 0x33, 0x2e, 0x6d, 0x6f, 0x6e, 0x6f, 0x72,
	0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x47, 0x61, 0x74, 0x68, 0x65, 0x72, 0x50, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x73, 0x46,
	0x6f, 0x72, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e,
	0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x2e, 0x76, 0x33, 0x2e, 0x47, 0x61, 0x74, 0x68,
	0x65, 0x72, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73,
	0x68, 0x69, 0x70, 0x73, 0x46, 0x6f, 0x72, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x23, 0x5a, 0x21, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x6d,
	0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x76, 0x32, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x33,
	0x2f, 0x61, 0x70, 0x69, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_v3_api_proto_frontend_proto_rawDescOnce sync.Once
	file_api_v3_api_proto_frontend_proto_rawDescData = file_api_v3_api_proto_frontend_proto_rawDesc
)

func file_api_v3_api_proto_frontend_proto_rawDescGZIP() []byte {
	file_api_v3_api_proto_frontend_proto_rawDescOnce.Do(func() {
		file_api_v3_api_proto_frontend_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v3_api_proto_frontend_proto_rawDescData)
	})
	return file_api_v3_api_proto_frontend_proto_rawDescData
}

var file_api_v3_api_proto_frontend_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_api_v3_api_proto_frontend_proto_goTypes = []interface{}{
	(*GatherProjectEnvironmentRequest)(nil),         // 0: monorail.v3.GatherProjectEnvironmentRequest
	(*GatherProjectEnvironmentResponse)(nil),        // 1: monorail.v3.GatherProjectEnvironmentResponse
	(*GatherProjectMembershipsForUserRequest)(nil),  // 2: monorail.v3.GatherProjectMembershipsForUserRequest
	(*GatherProjectMembershipsForUserResponse)(nil), // 3: monorail.v3.GatherProjectMembershipsForUserResponse
	(*Project)(nil),           // 4: monorail.v3.Project
	(*ProjectConfig)(nil),     // 5: monorail.v3.ProjectConfig
	(*StatusDef)(nil),         // 6: monorail.v3.StatusDef
	(*LabelDef)(nil),          // 7: monorail.v3.LabelDef
	(*ComponentDef)(nil),      // 8: monorail.v3.ComponentDef
	(*FieldDef)(nil),          // 9: monorail.v3.FieldDef
	(*ApprovalDef)(nil),       // 10: monorail.v3.ApprovalDef
	(*ProjectSavedQuery)(nil), // 11: monorail.v3.ProjectSavedQuery
	(*ProjectMember)(nil),     // 12: monorail.v3.ProjectMember
}
var file_api_v3_api_proto_frontend_proto_depIdxs = []int32{
	4,  // 0: monorail.v3.GatherProjectEnvironmentResponse.project:type_name -> monorail.v3.Project
	5,  // 1: monorail.v3.GatherProjectEnvironmentResponse.project_config:type_name -> monorail.v3.ProjectConfig
	6,  // 2: monorail.v3.GatherProjectEnvironmentResponse.statuses:type_name -> monorail.v3.StatusDef
	7,  // 3: monorail.v3.GatherProjectEnvironmentResponse.well_known_labels:type_name -> monorail.v3.LabelDef
	8,  // 4: monorail.v3.GatherProjectEnvironmentResponse.components:type_name -> monorail.v3.ComponentDef
	9,  // 5: monorail.v3.GatherProjectEnvironmentResponse.fields:type_name -> monorail.v3.FieldDef
	10, // 6: monorail.v3.GatherProjectEnvironmentResponse.approval_fields:type_name -> monorail.v3.ApprovalDef
	11, // 7: monorail.v3.GatherProjectEnvironmentResponse.saved_queries:type_name -> monorail.v3.ProjectSavedQuery
	12, // 8: monorail.v3.GatherProjectMembershipsForUserResponse.project_memberships:type_name -> monorail.v3.ProjectMember
	0,  // 9: monorail.v3.Frontend.GatherProjectEnvironment:input_type -> monorail.v3.GatherProjectEnvironmentRequest
	2,  // 10: monorail.v3.Frontend.GatherProjectMembershipsForUser:input_type -> monorail.v3.GatherProjectMembershipsForUserRequest
	1,  // 11: monorail.v3.Frontend.GatherProjectEnvironment:output_type -> monorail.v3.GatherProjectEnvironmentResponse
	3,  // 12: monorail.v3.Frontend.GatherProjectMembershipsForUser:output_type -> monorail.v3.GatherProjectMembershipsForUserResponse
	11, // [11:13] is the sub-list for method output_type
	9,  // [9:11] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_api_v3_api_proto_frontend_proto_init() }
func file_api_v3_api_proto_frontend_proto_init() {
	if File_api_v3_api_proto_frontend_proto != nil {
		return
	}
	file_api_v3_api_proto_project_objects_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_v3_api_proto_frontend_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GatherProjectEnvironmentRequest); i {
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
		file_api_v3_api_proto_frontend_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GatherProjectEnvironmentResponse); i {
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
		file_api_v3_api_proto_frontend_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GatherProjectMembershipsForUserRequest); i {
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
		file_api_v3_api_proto_frontend_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GatherProjectMembershipsForUserResponse); i {
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
			RawDescriptor: file_api_v3_api_proto_frontend_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v3_api_proto_frontend_proto_goTypes,
		DependencyIndexes: file_api_v3_api_proto_frontend_proto_depIdxs,
		MessageInfos:      file_api_v3_api_proto_frontend_proto_msgTypes,
	}.Build()
	File_api_v3_api_proto_frontend_proto = out.File
	file_api_v3_api_proto_frontend_proto_rawDesc = nil
	file_api_v3_api_proto_frontend_proto_goTypes = nil
	file_api_v3_api_proto_frontend_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// FrontendClient is the client API for Frontend service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FrontendClient interface {
	// status: DO NOT USE
	// Returns all project specific configurations needed for the SPA client.
	//
	// Raises:
	//
	//	INVALID_ARGUMENT if the project resource name provided is invalid.
	//	NOT_FOUND if the parent project is not found.
	//	PERMISSION_DENIED if user is not allowed to view this project.
	GatherProjectEnvironment(ctx context.Context, in *GatherProjectEnvironmentRequest, opts ...grpc.CallOption) (*GatherProjectEnvironmentResponse, error)
	// status: DO NOT USE
	// Returns all of a given user's project memberships.
	//
	// Raises:
	//
	//	NOT_FOUND if the user is not found.
	//	INVALID_ARGUMENT if the user resource name provided is invalid.
	GatherProjectMembershipsForUser(ctx context.Context, in *GatherProjectMembershipsForUserRequest, opts ...grpc.CallOption) (*GatherProjectMembershipsForUserResponse, error)
}
type frontendPRPCClient struct {
	client *prpc.Client
}

func NewFrontendPRPCClient(client *prpc.Client) FrontendClient {
	return &frontendPRPCClient{client}
}

func (c *frontendPRPCClient) GatherProjectEnvironment(ctx context.Context, in *GatherProjectEnvironmentRequest, opts ...grpc.CallOption) (*GatherProjectEnvironmentResponse, error) {
	out := new(GatherProjectEnvironmentResponse)
	err := c.client.Call(ctx, "monorail.v3.Frontend", "GatherProjectEnvironment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *frontendPRPCClient) GatherProjectMembershipsForUser(ctx context.Context, in *GatherProjectMembershipsForUserRequest, opts ...grpc.CallOption) (*GatherProjectMembershipsForUserResponse, error) {
	out := new(GatherProjectMembershipsForUserResponse)
	err := c.client.Call(ctx, "monorail.v3.Frontend", "GatherProjectMembershipsForUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type frontendClient struct {
	cc grpc.ClientConnInterface
}

func NewFrontendClient(cc grpc.ClientConnInterface) FrontendClient {
	return &frontendClient{cc}
}

func (c *frontendClient) GatherProjectEnvironment(ctx context.Context, in *GatherProjectEnvironmentRequest, opts ...grpc.CallOption) (*GatherProjectEnvironmentResponse, error) {
	out := new(GatherProjectEnvironmentResponse)
	err := c.cc.Invoke(ctx, "/monorail.v3.Frontend/GatherProjectEnvironment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *frontendClient) GatherProjectMembershipsForUser(ctx context.Context, in *GatherProjectMembershipsForUserRequest, opts ...grpc.CallOption) (*GatherProjectMembershipsForUserResponse, error) {
	out := new(GatherProjectMembershipsForUserResponse)
	err := c.cc.Invoke(ctx, "/monorail.v3.Frontend/GatherProjectMembershipsForUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FrontendServer is the server API for Frontend service.
type FrontendServer interface {
	// status: DO NOT USE
	// Returns all project specific configurations needed for the SPA client.
	//
	// Raises:
	//
	//	INVALID_ARGUMENT if the project resource name provided is invalid.
	//	NOT_FOUND if the parent project is not found.
	//	PERMISSION_DENIED if user is not allowed to view this project.
	GatherProjectEnvironment(context.Context, *GatherProjectEnvironmentRequest) (*GatherProjectEnvironmentResponse, error)
	// status: DO NOT USE
	// Returns all of a given user's project memberships.
	//
	// Raises:
	//
	//	NOT_FOUND if the user is not found.
	//	INVALID_ARGUMENT if the user resource name provided is invalid.
	GatherProjectMembershipsForUser(context.Context, *GatherProjectMembershipsForUserRequest) (*GatherProjectMembershipsForUserResponse, error)
}

// UnimplementedFrontendServer can be embedded to have forward compatible implementations.
type UnimplementedFrontendServer struct {
}

func (*UnimplementedFrontendServer) GatherProjectEnvironment(context.Context, *GatherProjectEnvironmentRequest) (*GatherProjectEnvironmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GatherProjectEnvironment not implemented")
}
func (*UnimplementedFrontendServer) GatherProjectMembershipsForUser(context.Context, *GatherProjectMembershipsForUserRequest) (*GatherProjectMembershipsForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GatherProjectMembershipsForUser not implemented")
}

func RegisterFrontendServer(s prpc.Registrar, srv FrontendServer) {
	s.RegisterService(&_Frontend_serviceDesc, srv)
}

func _Frontend_GatherProjectEnvironment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GatherProjectEnvironmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FrontendServer).GatherProjectEnvironment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monorail.v3.Frontend/GatherProjectEnvironment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FrontendServer).GatherProjectEnvironment(ctx, req.(*GatherProjectEnvironmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Frontend_GatherProjectMembershipsForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GatherProjectMembershipsForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FrontendServer).GatherProjectMembershipsForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monorail.v3.Frontend/GatherProjectMembershipsForUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FrontendServer).GatherProjectMembershipsForUser(ctx, req.(*GatherProjectMembershipsForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Frontend_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monorail.v3.Frontend",
	HandlerType: (*FrontendServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GatherProjectEnvironment",
			Handler:    _Frontend_GatherProjectEnvironment_Handler,
		},
		{
			MethodName: "GatherProjectMembershipsForUser",
			Handler:    _Frontend_GatherProjectMembershipsForUser_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v3/api_proto/frontend.proto",
}