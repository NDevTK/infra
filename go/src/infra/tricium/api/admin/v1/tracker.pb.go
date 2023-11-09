// Copyright 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.1
// source: infra/tricium/api/admin/v1/tracker.proto

package admin

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	v1 "infra/tricium/api/v1"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// WorkflowLaunchedRequest specified details need to mark a workflow as
// launched.
//
// This message should be sent by the Launcher after a workflow has been launched.
type WorkflowLaunchedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunId int64 `protobuf:"varint,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
}

func (x *WorkflowLaunchedRequest) Reset() {
	*x = WorkflowLaunchedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowLaunchedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowLaunchedRequest) ProtoMessage() {}

func (x *WorkflowLaunchedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowLaunchedRequest.ProtoReflect.Descriptor instead.
func (*WorkflowLaunchedRequest) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{0}
}

func (x *WorkflowLaunchedRequest) GetRunId() int64 {
	if x != nil {
		return x.RunId
	}
	return 0
}

type WorkflowLaunchedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WorkflowLaunchedResponse) Reset() {
	*x = WorkflowLaunchedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowLaunchedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowLaunchedResponse) ProtoMessage() {}

func (x *WorkflowLaunchedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowLaunchedResponse.ProtoReflect.Descriptor instead.
func (*WorkflowLaunchedResponse) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{1}
}

// WorkerLaunchedRequest specifies details needed to mark a worker as launched.
// This includes details useful for the tracking UI.
//
// This message should be sent by the Driver after a Buildbucket build for the
// worker has been triggered.
type WorkerLaunchedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunId              int64  `protobuf:"varint,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	Worker             string `protobuf:"bytes,2,opt,name=worker,proto3" json:"worker,omitempty"`
	BuildbucketBuildId int64  `protobuf:"varint,5,opt,name=buildbucket_build_id,json=buildbucketBuildId,proto3" json:"buildbucket_build_id,omitempty"`
}

func (x *WorkerLaunchedRequest) Reset() {
	*x = WorkerLaunchedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkerLaunchedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerLaunchedRequest) ProtoMessage() {}

func (x *WorkerLaunchedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerLaunchedRequest.ProtoReflect.Descriptor instead.
func (*WorkerLaunchedRequest) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{2}
}

func (x *WorkerLaunchedRequest) GetRunId() int64 {
	if x != nil {
		return x.RunId
	}
	return 0
}

func (x *WorkerLaunchedRequest) GetWorker() string {
	if x != nil {
		return x.Worker
	}
	return ""
}

func (x *WorkerLaunchedRequest) GetBuildbucketBuildId() int64 {
	if x != nil {
		return x.BuildbucketBuildId
	}
	return 0
}

type WorkerLaunchedResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WorkerLaunchedResponse) Reset() {
	*x = WorkerLaunchedResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkerLaunchedResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerLaunchedResponse) ProtoMessage() {}

func (x *WorkerLaunchedResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerLaunchedResponse.ProtoReflect.Descriptor instead.
func (*WorkerLaunchedResponse) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{3}
}

// WorkerDoneRequest specifies details needed to mark a worker as done.
// This includes details useful for the tracking UI.
type WorkerDoneRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunId             int64        `protobuf:"varint,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	Worker            string       `protobuf:"bytes,2,opt,name=worker,proto3" json:"worker,omitempty"`
	Provides          v1.Data_Type `protobuf:"varint,4,opt,name=provides,proto3,enum=tricium.Data_Type" json:"provides,omitempty"`
	State             v1.State     `protobuf:"varint,5,opt,name=state,proto3,enum=tricium.State" json:"state,omitempty"`
	BuildbucketOutput string       `protobuf:"bytes,6,opt,name=buildbucket_output,json=buildbucketOutput,proto3" json:"buildbucket_output,omitempty"`
}

func (x *WorkerDoneRequest) Reset() {
	*x = WorkerDoneRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkerDoneRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerDoneRequest) ProtoMessage() {}

func (x *WorkerDoneRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerDoneRequest.ProtoReflect.Descriptor instead.
func (*WorkerDoneRequest) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{4}
}

func (x *WorkerDoneRequest) GetRunId() int64 {
	if x != nil {
		return x.RunId
	}
	return 0
}

func (x *WorkerDoneRequest) GetWorker() string {
	if x != nil {
		return x.Worker
	}
	return ""
}

func (x *WorkerDoneRequest) GetProvides() v1.Data_Type {
	if x != nil {
		return x.Provides
	}
	return v1.Data_NONE
}

func (x *WorkerDoneRequest) GetState() v1.State {
	if x != nil {
		return x.State
	}
	return v1.State_PENDING
}

func (x *WorkerDoneRequest) GetBuildbucketOutput() string {
	if x != nil {
		return x.BuildbucketOutput
	}
	return ""
}

type WorkerDoneResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WorkerDoneResponse) Reset() {
	*x = WorkerDoneResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkerDoneResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerDoneResponse) ProtoMessage() {}

func (x *WorkerDoneResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerDoneResponse.ProtoReflect.Descriptor instead.
func (*WorkerDoneResponse) Descriptor() ([]byte, []int) {
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP(), []int{5}
}

var File_infra_tricium_api_admin_v1_tracker_proto protoreflect.FileDescriptor

var file_infra_tricium_api_admin_v1_tracker_proto_rawDesc = []byte{
	0x0a, 0x28, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x72, 0x69, 0x63, 0x69, 0x75, 0x6d, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x61,
	0x63, 0x6b, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x1a, 0x1f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x72, 0x69, 0x63, 0x69, 0x75, 0x6d,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x22, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x72, 0x69, 0x63, 0x69, 0x75,
	0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x69, 0x63, 0x69, 0x75, 0x6d,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x30, 0x0a, 0x17, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x72, 0x75, 0x6e, 0x49, 0x64, 0x22, 0x1a, 0x0a, 0x18, 0x57, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x84, 0x01, 0x0a, 0x15, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4c,
	0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15,
	0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x72, 0x75, 0x6e, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x30, 0x0a,
	0x14, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x12, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x64, 0x4a,
	0x04, 0x08, 0x03, 0x10, 0x04, 0x4a, 0x04, 0x08, 0x04, 0x10, 0x05, 0x22, 0x18, 0x0a, 0x16, 0x57,
	0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xd3, 0x01, 0x0a, 0x11, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72,
	0x44, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x72,
	0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x72, 0x75, 0x6e,
	0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x2e, 0x0a, 0x08, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x74,
	0x72, 0x69, 0x63, 0x69, 0x75, 0x6d, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x74, 0x72, 0x69, 0x63,
	0x69, 0x75, 0x6d, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x12, 0x2d, 0x0a, 0x12, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f,
	0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x4a,
	0x04, 0x08, 0x03, 0x10, 0x04, 0x4a, 0x04, 0x08, 0x07, 0x10, 0x08, 0x22, 0x14, 0x0a, 0x12, 0x57,
	0x6f, 0x72, 0x6b, 0x65, 0x72, 0x44, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x32, 0xf0, 0x01, 0x0a, 0x07, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x53, 0x0a,
	0x10, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65,
	0x64, 0x12, 0x1e, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1f, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x4d, 0x0a, 0x0e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4c, 0x61, 0x75, 0x6e,
	0x63, 0x68, 0x65, 0x64, 0x12, 0x1c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x57, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65,
	0x72, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x41, 0x0a, 0x0a, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x44, 0x6f, 0x6e, 0x65, 0x12,
	0x18, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x44, 0x6f,
	0x6e, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x44, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x22, 0x5a, 0x20, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x72,
	0x69, 0x63, 0x69, 0x75, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f,
	0x76, 0x31, 0x3b, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_tricium_api_admin_v1_tracker_proto_rawDescOnce sync.Once
	file_infra_tricium_api_admin_v1_tracker_proto_rawDescData = file_infra_tricium_api_admin_v1_tracker_proto_rawDesc
)

func file_infra_tricium_api_admin_v1_tracker_proto_rawDescGZIP() []byte {
	file_infra_tricium_api_admin_v1_tracker_proto_rawDescOnce.Do(func() {
		file_infra_tricium_api_admin_v1_tracker_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_tricium_api_admin_v1_tracker_proto_rawDescData)
	})
	return file_infra_tricium_api_admin_v1_tracker_proto_rawDescData
}

var file_infra_tricium_api_admin_v1_tracker_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_infra_tricium_api_admin_v1_tracker_proto_goTypes = []interface{}{
	(*WorkflowLaunchedRequest)(nil),  // 0: admin.WorkflowLaunchedRequest
	(*WorkflowLaunchedResponse)(nil), // 1: admin.WorkflowLaunchedResponse
	(*WorkerLaunchedRequest)(nil),    // 2: admin.WorkerLaunchedRequest
	(*WorkerLaunchedResponse)(nil),   // 3: admin.WorkerLaunchedResponse
	(*WorkerDoneRequest)(nil),        // 4: admin.WorkerDoneRequest
	(*WorkerDoneResponse)(nil),       // 5: admin.WorkerDoneResponse
	(v1.Data_Type)(0),                // 6: tricium.Data.Type
	(v1.State)(0),                    // 7: tricium.State
}
var file_infra_tricium_api_admin_v1_tracker_proto_depIdxs = []int32{
	6, // 0: admin.WorkerDoneRequest.provides:type_name -> tricium.Data.Type
	7, // 1: admin.WorkerDoneRequest.state:type_name -> tricium.State
	0, // 2: admin.Tracker.WorkflowLaunched:input_type -> admin.WorkflowLaunchedRequest
	2, // 3: admin.Tracker.WorkerLaunched:input_type -> admin.WorkerLaunchedRequest
	4, // 4: admin.Tracker.WorkerDone:input_type -> admin.WorkerDoneRequest
	1, // 5: admin.Tracker.WorkflowLaunched:output_type -> admin.WorkflowLaunchedResponse
	3, // 6: admin.Tracker.WorkerLaunched:output_type -> admin.WorkerLaunchedResponse
	5, // 7: admin.Tracker.WorkerDone:output_type -> admin.WorkerDoneResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_infra_tricium_api_admin_v1_tracker_proto_init() }
func file_infra_tricium_api_admin_v1_tracker_proto_init() {
	if File_infra_tricium_api_admin_v1_tracker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowLaunchedRequest); i {
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
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowLaunchedResponse); i {
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
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkerLaunchedRequest); i {
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
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkerLaunchedResponse); i {
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
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkerDoneRequest); i {
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
		file_infra_tricium_api_admin_v1_tracker_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkerDoneResponse); i {
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
			RawDescriptor: file_infra_tricium_api_admin_v1_tracker_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_infra_tricium_api_admin_v1_tracker_proto_goTypes,
		DependencyIndexes: file_infra_tricium_api_admin_v1_tracker_proto_depIdxs,
		MessageInfos:      file_infra_tricium_api_admin_v1_tracker_proto_msgTypes,
	}.Build()
	File_infra_tricium_api_admin_v1_tracker_proto = out.File
	file_infra_tricium_api_admin_v1_tracker_proto_rawDesc = nil
	file_infra_tricium_api_admin_v1_tracker_proto_goTypes = nil
	file_infra_tricium_api_admin_v1_tracker_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TrackerClient is the client API for Tracker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TrackerClient interface {
	// WorkflowLaunched marks the workflow as launched for a specified run.
	WorkflowLaunched(ctx context.Context, in *WorkflowLaunchedRequest, opts ...grpc.CallOption) (*WorkflowLaunchedResponse, error)
	// WorkerLaunched marks the specified worker as launched.
	WorkerLaunched(ctx context.Context, in *WorkerLaunchedRequest, opts ...grpc.CallOption) (*WorkerLaunchedResponse, error)
	// WorkerDone marks the specified worker as done.
	WorkerDone(ctx context.Context, in *WorkerDoneRequest, opts ...grpc.CallOption) (*WorkerDoneResponse, error)
}
type trackerPRPCClient struct {
	client *prpc.Client
}

func NewTrackerPRPCClient(client *prpc.Client) TrackerClient {
	return &trackerPRPCClient{client}
}

func (c *trackerPRPCClient) WorkflowLaunched(ctx context.Context, in *WorkflowLaunchedRequest, opts ...grpc.CallOption) (*WorkflowLaunchedResponse, error) {
	out := new(WorkflowLaunchedResponse)
	err := c.client.Call(ctx, "admin.Tracker", "WorkflowLaunched", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerPRPCClient) WorkerLaunched(ctx context.Context, in *WorkerLaunchedRequest, opts ...grpc.CallOption) (*WorkerLaunchedResponse, error) {
	out := new(WorkerLaunchedResponse)
	err := c.client.Call(ctx, "admin.Tracker", "WorkerLaunched", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerPRPCClient) WorkerDone(ctx context.Context, in *WorkerDoneRequest, opts ...grpc.CallOption) (*WorkerDoneResponse, error) {
	out := new(WorkerDoneResponse)
	err := c.client.Call(ctx, "admin.Tracker", "WorkerDone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type trackerClient struct {
	cc grpc.ClientConnInterface
}

func NewTrackerClient(cc grpc.ClientConnInterface) TrackerClient {
	return &trackerClient{cc}
}

func (c *trackerClient) WorkflowLaunched(ctx context.Context, in *WorkflowLaunchedRequest, opts ...grpc.CallOption) (*WorkflowLaunchedResponse, error) {
	out := new(WorkflowLaunchedResponse)
	err := c.cc.Invoke(ctx, "/admin.Tracker/WorkflowLaunched", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerClient) WorkerLaunched(ctx context.Context, in *WorkerLaunchedRequest, opts ...grpc.CallOption) (*WorkerLaunchedResponse, error) {
	out := new(WorkerLaunchedResponse)
	err := c.cc.Invoke(ctx, "/admin.Tracker/WorkerLaunched", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerClient) WorkerDone(ctx context.Context, in *WorkerDoneRequest, opts ...grpc.CallOption) (*WorkerDoneResponse, error) {
	out := new(WorkerDoneResponse)
	err := c.cc.Invoke(ctx, "/admin.Tracker/WorkerDone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrackerServer is the server API for Tracker service.
type TrackerServer interface {
	// WorkflowLaunched marks the workflow as launched for a specified run.
	WorkflowLaunched(context.Context, *WorkflowLaunchedRequest) (*WorkflowLaunchedResponse, error)
	// WorkerLaunched marks the specified worker as launched.
	WorkerLaunched(context.Context, *WorkerLaunchedRequest) (*WorkerLaunchedResponse, error)
	// WorkerDone marks the specified worker as done.
	WorkerDone(context.Context, *WorkerDoneRequest) (*WorkerDoneResponse, error)
}

// UnimplementedTrackerServer can be embedded to have forward compatible implementations.
type UnimplementedTrackerServer struct {
}

func (*UnimplementedTrackerServer) WorkflowLaunched(context.Context, *WorkflowLaunchedRequest) (*WorkflowLaunchedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkflowLaunched not implemented")
}
func (*UnimplementedTrackerServer) WorkerLaunched(context.Context, *WorkerLaunchedRequest) (*WorkerLaunchedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkerLaunched not implemented")
}
func (*UnimplementedTrackerServer) WorkerDone(context.Context, *WorkerDoneRequest) (*WorkerDoneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkerDone not implemented")
}

func RegisterTrackerServer(s prpc.Registrar, srv TrackerServer) {
	s.RegisterService(&_Tracker_serviceDesc, srv)
}

func _Tracker_WorkflowLaunched_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkflowLaunchedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).WorkflowLaunched(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.Tracker/WorkflowLaunched",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).WorkflowLaunched(ctx, req.(*WorkflowLaunchedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tracker_WorkerLaunched_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkerLaunchedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).WorkerLaunched(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.Tracker/WorkerLaunched",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).WorkerLaunched(ctx, req.(*WorkerLaunchedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tracker_WorkerDone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkerDoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).WorkerDone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.Tracker/WorkerDone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).WorkerDone(ctx, req.(*WorkerDoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Tracker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "admin.Tracker",
	HandlerType: (*TrackerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WorkflowLaunched",
			Handler:    _Tracker_WorkflowLaunched_Handler,
		},
		{
			MethodName: "WorkerLaunched",
			Handler:    _Tracker_WorkerLaunched_Handler,
		},
		{
			MethodName: "WorkerDone",
			Handler:    _Tracker_WorkerDone_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/tricium/api/admin/v1/tracker.proto",
}