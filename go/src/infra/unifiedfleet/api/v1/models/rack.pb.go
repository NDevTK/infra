// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.21.7
// source: infra/unifiedfleet/api/v1/models/rack.proto

package ufspb

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// Rack refers to the racks which are placed in
// Chrome Browser lab and Chrome OS lab. Machines and Pheripherals
// are placed in the Racks.
type Rack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique (fake probably) asset tag
	// The format will be racks/XXX
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The Breadboard Number for rack.
	// Reference: https://moma.corp.google.com/glossary?q=bbnum
	Bbnum    int32     `protobuf:"varint,13,opt,name=bbnum,proto3" json:"bbnum,omitempty"`
	Location *Location `protobuf:"bytes,2,opt,name=location,proto3" json:"location,omitempty"`
	// Indicates the Rack Unit capacity of the rack.
	CapacityRu int32 `protobuf:"varint,3,opt,name=capacity_ru,json=capacityRu,proto3" json:"capacity_ru,omitempty"`
	// Types that are assignable to Rack:
	//
	//	*Rack_ChromeBrowserRack
	//	*Rack_ChromeosRack
	Rack isRack_Rack `protobuf_oneof:"rack"`
	// Record the last update timestamp of this Rack (In UTC timezone)
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	// Record the ACL info of the rack
	Realm string `protobuf:"bytes,7,opt,name=realm,proto3" json:"realm,omitempty"`
	// tags user can attach for easy querying/searching
	Tags []string `protobuf:"bytes,8,rep,name=tags,proto3" json:"tags,omitempty"`
	// A description of this rack.
	Description string `protobuf:"bytes,11,opt,name=description,proto3" json:"description,omitempty"`
	// State of the rack
	ResourceState State `protobuf:"varint,12,opt,name=resource_state,json=resourceState,proto3,enum=unifiedfleet.api.v1.models.State" json:"resource_state,omitempty"`
}

func (x *Rack) Reset() {
	*x = Rack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rack) ProtoMessage() {}

func (x *Rack) ProtoReflect() protoreflect.Message {
	mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rack.ProtoReflect.Descriptor instead.
func (*Rack) Descriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescGZIP(), []int{0}
}

func (x *Rack) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Rack) GetBbnum() int32 {
	if x != nil {
		return x.Bbnum
	}
	return 0
}

func (x *Rack) GetLocation() *Location {
	if x != nil {
		return x.Location
	}
	return nil
}

func (x *Rack) GetCapacityRu() int32 {
	if x != nil {
		return x.CapacityRu
	}
	return 0
}

func (m *Rack) GetRack() isRack_Rack {
	if m != nil {
		return m.Rack
	}
	return nil
}

func (x *Rack) GetChromeBrowserRack() *ChromeBrowserRack {
	if x, ok := x.GetRack().(*Rack_ChromeBrowserRack); ok {
		return x.ChromeBrowserRack
	}
	return nil
}

func (x *Rack) GetChromeosRack() *ChromeOSRack {
	if x, ok := x.GetRack().(*Rack_ChromeosRack); ok {
		return x.ChromeosRack
	}
	return nil
}

func (x *Rack) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *Rack) GetRealm() string {
	if x != nil {
		return x.Realm
	}
	return ""
}

func (x *Rack) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Rack) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Rack) GetResourceState() State {
	if x != nil {
		return x.ResourceState
	}
	return State_STATE_UNSPECIFIED
}

type isRack_Rack interface {
	isRack_Rack()
}

type Rack_ChromeBrowserRack struct {
	ChromeBrowserRack *ChromeBrowserRack `protobuf:"bytes,4,opt,name=chrome_browser_rack,json=chromeBrowserRack,proto3,oneof"`
}

type Rack_ChromeosRack struct {
	ChromeosRack *ChromeOSRack `protobuf:"bytes,5,opt,name=chromeos_rack,json=chromeosRack,proto3,oneof"`
}

func (*Rack_ChromeBrowserRack) isRack_Rack() {}

func (*Rack_ChromeosRack) isRack_Rack() {}

// ChromeBrowserRack refers to the rack in Chrome Browser lab
type ChromeBrowserRack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RpmObjects    []*RPM    `protobuf:"bytes,4,rep,name=rpm_objects,json=rpmObjects,proto3" json:"rpm_objects,omitempty"`
	KvmObjects    []*KVM    `protobuf:"bytes,5,rep,name=kvm_objects,json=kvmObjects,proto3" json:"kvm_objects,omitempty"`
	SwitchObjects []*Switch `protobuf:"bytes,6,rep,name=switch_objects,json=switchObjects,proto3" json:"switch_objects,omitempty"`
}

func (x *ChromeBrowserRack) Reset() {
	*x = ChromeBrowserRack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChromeBrowserRack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChromeBrowserRack) ProtoMessage() {}

func (x *ChromeBrowserRack) ProtoReflect() protoreflect.Message {
	mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChromeBrowserRack.ProtoReflect.Descriptor instead.
func (*ChromeBrowserRack) Descriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescGZIP(), []int{1}
}

func (x *ChromeBrowserRack) GetRpmObjects() []*RPM {
	if x != nil {
		return x.RpmObjects
	}
	return nil
}

func (x *ChromeBrowserRack) GetKvmObjects() []*KVM {
	if x != nil {
		return x.KvmObjects
	}
	return nil
}

func (x *ChromeBrowserRack) GetSwitchObjects() []*Switch {
	if x != nil {
		return x.SwitchObjects
	}
	return nil
}

// ChromeOSRack refers to the rack in Chrome Browser lab
type ChromeOSRack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SwitchObjects []*Switch `protobuf:"bytes,1,rep,name=switch_objects,json=switchObjects,proto3" json:"switch_objects,omitempty"`
}

func (x *ChromeOSRack) Reset() {
	*x = ChromeOSRack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChromeOSRack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChromeOSRack) ProtoMessage() {}

func (x *ChromeOSRack) ProtoReflect() protoreflect.Message {
	mi := &file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChromeOSRack.ProtoReflect.Descriptor instead.
func (*ChromeOSRack) Descriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescGZIP(), []int{2}
}

func (x *ChromeOSRack) GetSwitchObjects() []*Switch {
	if x != nil {
		return x.SwitchObjects
	}
	return nil
}

var File_infra_unifiedfleet_api_v1_models_rack_proto protoreflect.FileDescriptor

var file_infra_unifiedfleet_api_v1_models_rack_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2f, 0x72, 0x61, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x75,
	0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x32, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75,
	0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x70, 0x65, 0x72, 0x69, 0x70, 0x68,
	0x65, 0x72, 0x61, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c, 0x69, 0x6e, 0x66,
	0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xec, 0x04, 0x0a, 0x04, 0x52, 0x61,
	0x63, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x62, 0x6e, 0x75, 0x6d, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x62, 0x62, 0x6e, 0x75, 0x6d, 0x12, 0x40, 0x0a, 0x08,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24,
	0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x4c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f,
	0x0a, 0x0b, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x5f, 0x72, 0x75, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0a, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x52, 0x75, 0x12,
	0x5f, 0x0a, 0x13, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x5f, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65,
	0x72, 0x5f, 0x72, 0x61, 0x63, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x75,
	0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x43, 0x68, 0x72, 0x6f, 0x6d, 0x65,
	0x42, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x52, 0x61, 0x63, 0x6b, 0x48, 0x00, 0x52, 0x11, 0x63,
	0x68, 0x72, 0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x52, 0x61, 0x63, 0x6b,
	0x12, 0x4f, 0x0a, 0x0d, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x5f, 0x72, 0x61, 0x63,
	0x6b, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x43, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x4f, 0x53, 0x52, 0x61, 0x63,
	0x6b, 0x48, 0x00, 0x52, 0x0c, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x52, 0x61, 0x63,
	0x6b, 0x12, 0x40, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x61, 0x67,
	0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x20, 0x0a,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x48, 0x0a, 0x0e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x0d, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x3a, 0x38, 0xea, 0x41, 0x35, 0x0a, 0x25,
	0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x2d, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2d, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x2e, 0x61, 0x70, 0x70, 0x73, 0x70, 0x6f, 0x74, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x52, 0x61, 0x63, 0x6b, 0x12, 0x0c, 0x72, 0x61, 0x63, 0x6b, 0x73, 0x2f, 0x7b, 0x72, 0x61,
	0x63, 0x6b, 0x7d, 0x42, 0x06, 0x0a, 0x04, 0x72, 0x61, 0x63, 0x6b, 0x4a, 0x04, 0x08, 0x09, 0x10,
	0x0a, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x8a, 0x02, 0x0a, 0x11, 0x43, 0x68, 0x72,
	0x6f, 0x6d, 0x65, 0x42, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x52, 0x61, 0x63, 0x6b, 0x12, 0x40,
	0x0a, 0x0b, 0x72, 0x70, 0x6d, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65,
	0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x52, 0x50, 0x4d, 0x52, 0x0a, 0x72, 0x70, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x12, 0x40, 0x0a, 0x0b, 0x6b, 0x76, 0x6d, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18,
	0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2e, 0x4b, 0x56, 0x4d, 0x52, 0x0a, 0x6b, 0x76, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x12, 0x49, 0x0a, 0x0e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x5f, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x75, 0x6e, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x53, 0x77, 0x69, 0x74, 0x63, 0x68, 0x52, 0x0d,
	0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x4a, 0x04, 0x08,
	0x01, 0x10, 0x02, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x4a, 0x04, 0x08, 0x03, 0x10, 0x04, 0x52,
	0x04, 0x72, 0x70, 0x6d, 0x73, 0x52, 0x04, 0x6b, 0x76, 0x6d, 0x73, 0x52, 0x08, 0x73, 0x77, 0x69,
	0x74, 0x63, 0x68, 0x65, 0x73, 0x22, 0x59, 0x0a, 0x0c, 0x43, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x4f,
	0x53, 0x52, 0x61, 0x63, 0x6b, 0x12, 0x49, 0x0a, 0x0e, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x5f,
	0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x53, 0x77, 0x69, 0x74, 0x63,
	0x68, 0x52, 0x0d, 0x73, 0x77, 0x69, 0x74, 0x63, 0x68, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x42, 0x28, 0x5a, 0x26, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x3b, 0x75, 0x66, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescOnce sync.Once
	file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescData = file_infra_unifiedfleet_api_v1_models_rack_proto_rawDesc
)

func file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescGZIP() []byte {
	file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescOnce.Do(func() {
		file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescData)
	})
	return file_infra_unifiedfleet_api_v1_models_rack_proto_rawDescData
}

var file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_infra_unifiedfleet_api_v1_models_rack_proto_goTypes = []interface{}{
	(*Rack)(nil),                  // 0: unifiedfleet.api.v1.models.Rack
	(*ChromeBrowserRack)(nil),     // 1: unifiedfleet.api.v1.models.ChromeBrowserRack
	(*ChromeOSRack)(nil),          // 2: unifiedfleet.api.v1.models.ChromeOSRack
	(*Location)(nil),              // 3: unifiedfleet.api.v1.models.Location
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
	(State)(0),                    // 5: unifiedfleet.api.v1.models.State
	(*RPM)(nil),                   // 6: unifiedfleet.api.v1.models.RPM
	(*KVM)(nil),                   // 7: unifiedfleet.api.v1.models.KVM
	(*Switch)(nil),                // 8: unifiedfleet.api.v1.models.Switch
}
var file_infra_unifiedfleet_api_v1_models_rack_proto_depIdxs = []int32{
	3, // 0: unifiedfleet.api.v1.models.Rack.location:type_name -> unifiedfleet.api.v1.models.Location
	1, // 1: unifiedfleet.api.v1.models.Rack.chrome_browser_rack:type_name -> unifiedfleet.api.v1.models.ChromeBrowserRack
	2, // 2: unifiedfleet.api.v1.models.Rack.chromeos_rack:type_name -> unifiedfleet.api.v1.models.ChromeOSRack
	4, // 3: unifiedfleet.api.v1.models.Rack.update_time:type_name -> google.protobuf.Timestamp
	5, // 4: unifiedfleet.api.v1.models.Rack.resource_state:type_name -> unifiedfleet.api.v1.models.State
	6, // 5: unifiedfleet.api.v1.models.ChromeBrowserRack.rpm_objects:type_name -> unifiedfleet.api.v1.models.RPM
	7, // 6: unifiedfleet.api.v1.models.ChromeBrowserRack.kvm_objects:type_name -> unifiedfleet.api.v1.models.KVM
	8, // 7: unifiedfleet.api.v1.models.ChromeBrowserRack.switch_objects:type_name -> unifiedfleet.api.v1.models.Switch
	8, // 8: unifiedfleet.api.v1.models.ChromeOSRack.switch_objects:type_name -> unifiedfleet.api.v1.models.Switch
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_infra_unifiedfleet_api_v1_models_rack_proto_init() }
func file_infra_unifiedfleet_api_v1_models_rack_proto_init() {
	if File_infra_unifiedfleet_api_v1_models_rack_proto != nil {
		return
	}
	file_infra_unifiedfleet_api_v1_models_location_proto_init()
	file_infra_unifiedfleet_api_v1_models_peripherals_proto_init()
	file_infra_unifiedfleet_api_v1_models_state_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rack); i {
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
		file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChromeBrowserRack); i {
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
		file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChromeOSRack); i {
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
	file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Rack_ChromeBrowserRack)(nil),
		(*Rack_ChromeosRack)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_infra_unifiedfleet_api_v1_models_rack_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_unifiedfleet_api_v1_models_rack_proto_goTypes,
		DependencyIndexes: file_infra_unifiedfleet_api_v1_models_rack_proto_depIdxs,
		MessageInfos:      file_infra_unifiedfleet_api_v1_models_rack_proto_msgTypes,
	}.Build()
	File_infra_unifiedfleet_api_v1_models_rack_proto = out.File
	file_infra_unifiedfleet_api_v1_models_rack_proto_rawDesc = nil
	file_infra_unifiedfleet_api_v1_models_rack_proto_goTypes = nil
	file_infra_unifiedfleet_api_v1_models_rack_proto_depIdxs = nil
}
