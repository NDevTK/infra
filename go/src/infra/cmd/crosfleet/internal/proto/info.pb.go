// Copyright 2021 The Chromium Authors
// Use of this source code is governed by the Apache v2.0 license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.1
// source: infra/cmd/crosfleet/internal/proto/info.proto

package crosfleetpb

import (
	proto "go.chromium.org/luci/buildbucket/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	models "infra/unifiedfleet/api/v1/models"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// DUTInfo contains details about the physical lab setup and machine of a
// particular DUT.
type DUTInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The DUT's hostname.
	Hostname string `protobuf:"bytes,1,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	// The DUT's lab setup.
	LabSetup *models.MachineLSE `protobuf:"bytes,2,opt,name=LabSetup,proto3" json:"LabSetup,omitempty"`
	// The DUT's physical machine.
	Machine *models.Machine `protobuf:"bytes,3,opt,name=Machine,proto3" json:"Machine,omitempty"`
}

func (x *DUTInfo) Reset() {
	*x = DUTInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DUTInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DUTInfo) ProtoMessage() {}

func (x *DUTInfo) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DUTInfo.ProtoReflect.Descriptor instead.
func (*DUTInfo) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{0}
}

func (x *DUTInfo) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *DUTInfo) GetLabSetup() *models.MachineLSE {
	if x != nil {
		return x.LabSetup
	}
	return nil
}

func (x *DUTInfo) GetMachine() *models.Machine {
	if x != nil {
		return x.Machine
	}
	return nil
}

type DUTInfoList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DUTs []*DUTInfo `protobuf:"bytes,1,rep,name=DUTs,proto3" json:"DUTs,omitempty"`
}

func (x *DUTInfoList) Reset() {
	*x = DUTInfoList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DUTInfoList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DUTInfoList) ProtoMessage() {}

func (x *DUTInfoList) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DUTInfoList.ProtoReflect.Descriptor instead.
func (*DUTInfoList) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{1}
}

func (x *DUTInfoList) GetDUTs() []*DUTInfo {
	if x != nil {
		return x.DUTs
	}
	return nil
}

type LeaseInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The dut_leaser Buildbucket build.
	Build *proto.Build `protobuf:"bytes,2,opt,name=Build,proto3" json:"Build,omitempty"`
	// Information about the leased DUT.
	DUT *DUTInfo `protobuf:"bytes,3,opt,name=DUT,proto3" json:"DUT,omitempty"`
}

func (x *LeaseInfo) Reset() {
	*x = LeaseInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaseInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaseInfo) ProtoMessage() {}

func (x *LeaseInfo) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaseInfo.ProtoReflect.Descriptor instead.
func (*LeaseInfo) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{2}
}

func (x *LeaseInfo) GetBuild() *proto.Build {
	if x != nil {
		return x.Build
	}
	return nil
}

func (x *LeaseInfo) GetDUT() *DUTInfo {
	if x != nil {
		return x.DUT
	}
	return nil
}

type LeaseInfoList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Leases []*LeaseInfo `protobuf:"bytes,1,rep,name=Leases,proto3" json:"Leases,omitempty"`
}

func (x *LeaseInfoList) Reset() {
	*x = LeaseInfoList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaseInfoList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaseInfoList) ProtoMessage() {}

func (x *LeaseInfoList) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaseInfoList.ProtoReflect.Descriptor instead.
func (*LeaseInfoList) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{3}
}

func (x *LeaseInfoList) GetLeases() []*LeaseInfo {
	if x != nil {
		return x.Leases
	}
	return nil
}

type BuildIdList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ids []int64 `protobuf:"varint,1,rep,packed,name=Ids,proto3" json:"Ids,omitempty"`
}

func (x *BuildIdList) Reset() {
	*x = BuildIdList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildIdList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildIdList) ProtoMessage() {}

func (x *BuildIdList) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildIdList.ProtoReflect.Descriptor instead.
func (*BuildIdList) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{4}
}

func (x *BuildIdList) GetIds() []int64 {
	if x != nil {
		return x.Ids
	}
	return nil
}

type BuildLaunchList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Launches []*BuildLaunch `protobuf:"bytes,1,rep,name=Launches,proto3" json:"Launches,omitempty"`
}

func (x *BuildLaunchList) Reset() {
	*x = BuildLaunchList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildLaunchList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildLaunchList) ProtoMessage() {}

func (x *BuildLaunchList) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildLaunchList.ProtoReflect.Descriptor instead.
func (*BuildLaunchList) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{5}
}

func (x *BuildLaunchList) GetLaunches() []*BuildLaunch {
	if x != nil {
		return x.Launches
	}
	return nil
}

type BuildLaunch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Build      *proto.Build `protobuf:"bytes,1,opt,name=Build,proto3" json:"Build,omitempty"`
	BuildError string       `protobuf:"bytes,2,opt,name=BuildError,proto3" json:"BuildError,omitempty"`
}

func (x *BuildLaunch) Reset() {
	*x = BuildLaunch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildLaunch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildLaunch) ProtoMessage() {}

func (x *BuildLaunch) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildLaunch.ProtoReflect.Descriptor instead.
func (*BuildLaunch) Descriptor() ([]byte, []int) {
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP(), []int{6}
}

func (x *BuildLaunch) GetBuild() *proto.Build {
	if x != nil {
		return x.Build
	}
	return nil
}

func (x *BuildLaunch) GetBuildError() string {
	if x != nil {
		return x.BuildError
	}
	return ""
}

var File_infra_cmd_crosfleet_internal_proto_info_proto protoreflect.FileDescriptor

var file_infra_cmd_crosfleet_internal_proto_info_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x63, 0x72, 0x6f, 0x73,
	0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x15, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x70, 0x62, 0x1a, 0x32, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x32, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x5f, 0x6c, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa8,
	0x01, 0x0a, 0x07, 0x44, 0x55, 0x54, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x48, 0x6f,
	0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x48, 0x6f,
	0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x42, 0x0a, 0x08, 0x4c, 0x61, 0x62, 0x53, 0x65, 0x74,
	0x75, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x4c, 0x53, 0x45,
	0x52, 0x08, 0x4c, 0x61, 0x62, 0x53, 0x65, 0x74, 0x75, 0x70, 0x12, 0x3d, 0x0a, 0x07, 0x4d, 0x61,
	0x63, 0x68, 0x69, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x75, 0x6e,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65,
	0x52, 0x07, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x22, 0x41, 0x0a, 0x0b, 0x44, 0x55, 0x54,
	0x49, 0x6e, 0x66, 0x6f, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x32, 0x0a, 0x04, 0x44, 0x55, 0x54, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65,
	0x65, 0x74, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x70, 0x62, 0x2e, 0x44,
	0x55, 0x54, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x44, 0x55, 0x54, 0x73, 0x22, 0x6a, 0x0a, 0x09,
	0x4c, 0x65, 0x61, 0x73, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2b, 0x0a, 0x05, 0x42, 0x75, 0x69,
	0x6c, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52,
	0x05, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x12, 0x30, 0x0a, 0x03, 0x44, 0x55, 0x54, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e,
	0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x70, 0x62, 0x2e, 0x44, 0x55, 0x54, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x03, 0x44, 0x55, 0x54, 0x22, 0x49, 0x0a, 0x0d, 0x4c, 0x65, 0x61, 0x73,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x06, 0x4c, 0x65, 0x61,
	0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x72, 0x6f, 0x73,
	0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x70,
	0x62, 0x2e, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x4c, 0x65, 0x61,
	0x73, 0x65, 0x73, 0x22, 0x1f, 0x0a, 0x0b, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x64, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x03, 0x52,
	0x03, 0x49, 0x64, 0x73, 0x22, 0x51, 0x0a, 0x0f, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x4c, 0x61, 0x75,
	0x6e, 0x63, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x3e, 0x0a, 0x08, 0x4c, 0x61, 0x75, 0x6e, 0x63,
	0x68, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x72, 0x6f, 0x73,
	0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x70,
	0x62, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x52, 0x08, 0x4c,
	0x61, 0x75, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x22, 0x5a, 0x0a, 0x0b, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x12, 0x2b, 0x0a, 0x05, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x05, 0x42, 0x75,
	0x69, 0x6c, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x42, 0x30, 0x5a, 0x2e, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x6d, 0x64,
	0x2f, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x63, 0x72, 0x6f, 0x73, 0x66, 0x6c,
	0x65, 0x65, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescOnce sync.Once
	file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescData = file_infra_cmd_crosfleet_internal_proto_info_proto_rawDesc
)

func file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescGZIP() []byte {
	file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescOnce.Do(func() {
		file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescData)
	})
	return file_infra_cmd_crosfleet_internal_proto_info_proto_rawDescData
}

var file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_infra_cmd_crosfleet_internal_proto_info_proto_goTypes = []interface{}{
	(*DUTInfo)(nil),           // 0: crosfleet.crosfleetpb.DUTInfo
	(*DUTInfoList)(nil),       // 1: crosfleet.crosfleetpb.DUTInfoList
	(*LeaseInfo)(nil),         // 2: crosfleet.crosfleetpb.LeaseInfo
	(*LeaseInfoList)(nil),     // 3: crosfleet.crosfleetpb.LeaseInfoList
	(*BuildIdList)(nil),       // 4: crosfleet.crosfleetpb.BuildIdList
	(*BuildLaunchList)(nil),   // 5: crosfleet.crosfleetpb.BuildLaunchList
	(*BuildLaunch)(nil),       // 6: crosfleet.crosfleetpb.BuildLaunch
	(*models.MachineLSE)(nil), // 7: unifiedfleet.api.v1.models.MachineLSE
	(*models.Machine)(nil),    // 8: unifiedfleet.api.v1.models.Machine
	(*proto.Build)(nil),       // 9: buildbucket.v2.Build
}
var file_infra_cmd_crosfleet_internal_proto_info_proto_depIdxs = []int32{
	7, // 0: crosfleet.crosfleetpb.DUTInfo.LabSetup:type_name -> unifiedfleet.api.v1.models.MachineLSE
	8, // 1: crosfleet.crosfleetpb.DUTInfo.Machine:type_name -> unifiedfleet.api.v1.models.Machine
	0, // 2: crosfleet.crosfleetpb.DUTInfoList.DUTs:type_name -> crosfleet.crosfleetpb.DUTInfo
	9, // 3: crosfleet.crosfleetpb.LeaseInfo.Build:type_name -> buildbucket.v2.Build
	0, // 4: crosfleet.crosfleetpb.LeaseInfo.DUT:type_name -> crosfleet.crosfleetpb.DUTInfo
	2, // 5: crosfleet.crosfleetpb.LeaseInfoList.Leases:type_name -> crosfleet.crosfleetpb.LeaseInfo
	6, // 6: crosfleet.crosfleetpb.BuildLaunchList.Launches:type_name -> crosfleet.crosfleetpb.BuildLaunch
	9, // 7: crosfleet.crosfleetpb.BuildLaunch.Build:type_name -> buildbucket.v2.Build
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_infra_cmd_crosfleet_internal_proto_info_proto_init() }
func file_infra_cmd_crosfleet_internal_proto_info_proto_init() {
	if File_infra_cmd_crosfleet_internal_proto_info_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DUTInfo); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DUTInfoList); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaseInfo); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaseInfoList); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildIdList); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildLaunchList); i {
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
		file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildLaunch); i {
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
			RawDescriptor: file_infra_cmd_crosfleet_internal_proto_info_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_cmd_crosfleet_internal_proto_info_proto_goTypes,
		DependencyIndexes: file_infra_cmd_crosfleet_internal_proto_info_proto_depIdxs,
		MessageInfos:      file_infra_cmd_crosfleet_internal_proto_info_proto_msgTypes,
	}.Build()
	File_infra_cmd_crosfleet_internal_proto_info_proto = out.File
	file_infra_cmd_crosfleet_internal_proto_info_proto_rawDesc = nil
	file_infra_cmd_crosfleet_internal_proto_info_proto_goTypes = nil
	file_infra_cmd_crosfleet_internal_proto_info_proto_depIdxs = nil
}