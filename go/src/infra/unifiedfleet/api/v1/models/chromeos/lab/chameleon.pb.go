// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.7
// source: infra/unifiedfleet/api/v1/models/chromeos/lab/chameleon.proto

package ufspb

import (
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

// Next Tag: 12
type ChameleonType int32

const (
	ChameleonType_CHAMELEON_TYPE_INVALID ChameleonType = 0
	ChameleonType_CHAMELEON_TYPE_DP      ChameleonType = 2
	// Deprecated: Marked as deprecated in infra/unifiedfleet/api/v1/models/chromeos/lab/chameleon.proto.
	ChameleonType_CHAMELEON_TYPE_DP_HDMI ChameleonType = 3
	// Deprecated: Marked as deprecated in infra/unifiedfleet/api/v1/models/chromeos/lab/chameleon.proto.
	ChameleonType_CHAMELEON_TYPE_VGA  ChameleonType = 4
	ChameleonType_CHAMELEON_TYPE_HDMI ChameleonType = 5
	ChameleonType_CHAMELEON_TYPE_V2   ChameleonType = 9
	ChameleonType_CHAMELEON_TYPE_V3   ChameleonType = 10
	ChameleonType_CHAMELEON_TYPE_RPI  ChameleonType = 11
)

// Enum value maps for ChameleonType.
var (
	ChameleonType_name = map[int32]string{
		0:  "CHAMELEON_TYPE_INVALID",
		2:  "CHAMELEON_TYPE_DP",
		3:  "CHAMELEON_TYPE_DP_HDMI",
		4:  "CHAMELEON_TYPE_VGA",
		5:  "CHAMELEON_TYPE_HDMI",
		9:  "CHAMELEON_TYPE_V2",
		10: "CHAMELEON_TYPE_V3",
		11: "CHAMELEON_TYPE_RPI",
	}
	ChameleonType_value = map[string]int32{
		"CHAMELEON_TYPE_INVALID": 0,
		"CHAMELEON_TYPE_DP":      2,
		"CHAMELEON_TYPE_DP_HDMI": 3,
		"CHAMELEON_TYPE_VGA":     4,
		"CHAMELEON_TYPE_HDMI":    5,
		"CHAMELEON_TYPE_V2":      9,
		"CHAMELEON_TYPE_V3":      10,
		"CHAMELEON_TYPE_RPI":     11,
	}
)

func (x ChameleonType) Enum() *ChameleonType {
	p := new(ChameleonType)
	*p = x
	return p
}

func (x ChameleonType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ChameleonType) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[0].Descriptor()
}

func (ChameleonType) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[0]
}

func (x ChameleonType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ChameleonType.Descriptor instead.
func (ChameleonType) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescGZIP(), []int{0}
}

// Indicate the audio box jack plugger state
// Next Tag: 4
type Chameleon_AudioBoxJackPlugger int32

const (
	Chameleon_AUDIOBOX_JACKPLUGGER_UNSPECIFIED    Chameleon_AudioBoxJackPlugger = 0
	Chameleon_AUDIOBOX_JACKPLUGGER_WORKING        Chameleon_AudioBoxJackPlugger = 1
	Chameleon_AUDIOBOX_JACKPLUGGER_BROKEN         Chameleon_AudioBoxJackPlugger = 2
	Chameleon_AUDIOBOX_JACKPLUGGER_NOT_APPLICABLE Chameleon_AudioBoxJackPlugger = 3
)

// Enum value maps for Chameleon_AudioBoxJackPlugger.
var (
	Chameleon_AudioBoxJackPlugger_name = map[int32]string{
		0: "AUDIOBOX_JACKPLUGGER_UNSPECIFIED",
		1: "AUDIOBOX_JACKPLUGGER_WORKING",
		2: "AUDIOBOX_JACKPLUGGER_BROKEN",
		3: "AUDIOBOX_JACKPLUGGER_NOT_APPLICABLE",
	}
	Chameleon_AudioBoxJackPlugger_value = map[string]int32{
		"AUDIOBOX_JACKPLUGGER_UNSPECIFIED":    0,
		"AUDIOBOX_JACKPLUGGER_WORKING":        1,
		"AUDIOBOX_JACKPLUGGER_BROKEN":         2,
		"AUDIOBOX_JACKPLUGGER_NOT_APPLICABLE": 3,
	}
)

func (x Chameleon_AudioBoxJackPlugger) Enum() *Chameleon_AudioBoxJackPlugger {
	p := new(Chameleon_AudioBoxJackPlugger)
	*p = x
	return p
}

func (x Chameleon_AudioBoxJackPlugger) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Chameleon_AudioBoxJackPlugger) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[1].Descriptor()
}

func (Chameleon_AudioBoxJackPlugger) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[1]
}

func (x Chameleon_AudioBoxJackPlugger) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Chameleon_AudioBoxJackPlugger.Descriptor instead.
func (Chameleon_AudioBoxJackPlugger) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescGZIP(), []int{0, 0}
}

// Indicate the trrs types
// Next Tag: 3
type Chameleon_TRRSType int32

const (
	Chameleon_TRRS_TYPE_UNSPECIFIED Chameleon_TRRSType = 0
	Chameleon_TRRS_TYPE_CTIA        Chameleon_TRRSType = 1
	Chameleon_TRRS_TYPE_OMTP        Chameleon_TRRSType = 2 // Refer "go/wiki/Phone_connector_(audio)#TRRS_standards" for more types
)

// Enum value maps for Chameleon_TRRSType.
var (
	Chameleon_TRRSType_name = map[int32]string{
		0: "TRRS_TYPE_UNSPECIFIED",
		1: "TRRS_TYPE_CTIA",
		2: "TRRS_TYPE_OMTP",
	}
	Chameleon_TRRSType_value = map[string]int32{
		"TRRS_TYPE_UNSPECIFIED": 0,
		"TRRS_TYPE_CTIA":        1,
		"TRRS_TYPE_OMTP":        2,
	}
)

func (x Chameleon_TRRSType) Enum() *Chameleon_TRRSType {
	p := new(Chameleon_TRRSType)
	*p = x
	return p
}

func (x Chameleon_TRRSType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Chameleon_TRRSType) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[2].Descriptor()
}

func (Chameleon_TRRSType) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes[2]
}

func (x Chameleon_TRRSType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Chameleon_TRRSType.Descriptor instead.
func (Chameleon_TRRSType) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescGZIP(), []int{0, 1}
}

// Next Tag: 8
type Chameleon struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChameleonPeripherals []ChameleonType `protobuf:"varint,3,rep,packed,name=chameleon_peripherals,json=chameleonPeripherals,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.ChameleonType" json:"chameleon_peripherals,omitempty"`
	// Indicate if there's audio_board in the chameleon.
	AudioBoard bool   `protobuf:"varint,2,opt,name=audio_board,json=audioBoard,proto3" json:"audio_board,omitempty"`
	Hostname   string `protobuf:"bytes,4,opt,name=hostname,proto3" json:"hostname,omitempty"`
	// Remote Power Management for chameleon device.
	Rpm                 *OSRPM                        `protobuf:"bytes,5,opt,name=rpm,proto3" json:"rpm,omitempty"`
	AudioboxJackplugger Chameleon_AudioBoxJackPlugger `protobuf:"varint,6,opt,name=audiobox_jackplugger,json=audioboxJackplugger,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.Chameleon_AudioBoxJackPlugger" json:"audiobox_jackplugger,omitempty"`
	// Indicate the type of audio cable
	TrrsType Chameleon_TRRSType `protobuf:"varint,7,opt,name=trrs_type,json=trrsType,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.Chameleon_TRRSType" json:"trrs_type,omitempty"`
}

func (x *Chameleon) Reset() {
	*x = Chameleon{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Chameleon) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Chameleon) ProtoMessage() {}

func (x *Chameleon) ProtoReflect() protoreflect.Message {
	mi := &file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Chameleon.ProtoReflect.Descriptor instead.
func (*Chameleon) Descriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescGZIP(), []int{0}
}

func (x *Chameleon) GetChameleonPeripherals() []ChameleonType {
	if x != nil {
		return x.ChameleonPeripherals
	}
	return nil
}

func (x *Chameleon) GetAudioBoard() bool {
	if x != nil {
		return x.AudioBoard
	}
	return false
}

func (x *Chameleon) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *Chameleon) GetRpm() *OSRPM {
	if x != nil {
		return x.Rpm
	}
	return nil
}

func (x *Chameleon) GetAudioboxJackplugger() Chameleon_AudioBoxJackPlugger {
	if x != nil {
		return x.AudioboxJackplugger
	}
	return Chameleon_AUDIOBOX_JACKPLUGGER_UNSPECIFIED
}

func (x *Chameleon) GetTrrsType() Chameleon_TRRSType {
	if x != nil {
		return x.TrrsType
	}
	return Chameleon_TRRS_TYPE_UNSPECIFIED
}

var File_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto protoreflect.FileDescriptor

var file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDesc = []byte{
	0x0a, 0x3d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x2f,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x27, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62, 0x1a, 0x37, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f,
	0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x2f, 0x72, 0x70, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xcb, 0x05, 0x0a, 0x09, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x12,
	0x6b, 0x0a, 0x15, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x5f, 0x70, 0x65, 0x72,
	0x69, 0x70, 0x68, 0x65, 0x72, 0x61, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x36,
	0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65,
	0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x14, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x50, 0x65, 0x72, 0x69, 0x70, 0x68, 0x65, 0x72, 0x61, 0x6c, 0x73, 0x12, 0x1f, 0x0a, 0x0b,
	0x61, 0x75, 0x64, 0x69, 0x6f, 0x5f, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0a, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x40, 0x0a, 0x03, 0x72, 0x70, 0x6d,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62,
	0x2e, 0x4f, 0x53, 0x52, 0x50, 0x4d, 0x52, 0x03, 0x72, 0x70, 0x6d, 0x12, 0x79, 0x0a, 0x14, 0x61,
	0x75, 0x64, 0x69, 0x6f, 0x62, 0x6f, 0x78, 0x5f, 0x6a, 0x61, 0x63, 0x6b, 0x70, 0x6c, 0x75, 0x67,
	0x67, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x46, 0x2e, 0x75, 0x6e, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e,
	0x6c, 0x61, 0x62, 0x2e, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x41, 0x75,
	0x64, 0x69, 0x6f, 0x42, 0x6f, 0x78, 0x4a, 0x61, 0x63, 0x6b, 0x50, 0x6c, 0x75, 0x67, 0x67, 0x65,
	0x72, 0x52, 0x13, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x62, 0x6f, 0x78, 0x4a, 0x61, 0x63, 0x6b, 0x70,
	0x6c, 0x75, 0x67, 0x67, 0x65, 0x72, 0x12, 0x58, 0x0a, 0x09, 0x74, 0x72, 0x72, 0x73, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x3b, 0x2e, 0x75, 0x6e, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e,
	0x6c, 0x61, 0x62, 0x2e, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x54, 0x52,
	0x52, 0x53, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x74, 0x72, 0x72, 0x73, 0x54, 0x79, 0x70, 0x65,
	0x22, 0xa7, 0x01, 0x0a, 0x13, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x42, 0x6f, 0x78, 0x4a, 0x61, 0x63,
	0x6b, 0x50, 0x6c, 0x75, 0x67, 0x67, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x20, 0x41, 0x55, 0x44, 0x49,
	0x4f, 0x42, 0x4f, 0x58, 0x5f, 0x4a, 0x41, 0x43, 0x4b, 0x50, 0x4c, 0x55, 0x47, 0x47, 0x45, 0x52,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x20,
	0x0a, 0x1c, 0x41, 0x55, 0x44, 0x49, 0x4f, 0x42, 0x4f, 0x58, 0x5f, 0x4a, 0x41, 0x43, 0x4b, 0x50,
	0x4c, 0x55, 0x47, 0x47, 0x45, 0x52, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x49, 0x4e, 0x47, 0x10, 0x01,
	0x12, 0x1f, 0x0a, 0x1b, 0x41, 0x55, 0x44, 0x49, 0x4f, 0x42, 0x4f, 0x58, 0x5f, 0x4a, 0x41, 0x43,
	0x4b, 0x50, 0x4c, 0x55, 0x47, 0x47, 0x45, 0x52, 0x5f, 0x42, 0x52, 0x4f, 0x4b, 0x45, 0x4e, 0x10,
	0x02, 0x12, 0x27, 0x0a, 0x23, 0x41, 0x55, 0x44, 0x49, 0x4f, 0x42, 0x4f, 0x58, 0x5f, 0x4a, 0x41,
	0x43, 0x4b, 0x50, 0x4c, 0x55, 0x47, 0x47, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x41, 0x50,
	0x50, 0x4c, 0x49, 0x43, 0x41, 0x42, 0x4c, 0x45, 0x10, 0x03, 0x22, 0x4d, 0x0a, 0x08, 0x54, 0x52,
	0x52, 0x53, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x15, 0x54, 0x52, 0x52, 0x53, 0x5f, 0x54,
	0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x52, 0x52, 0x53, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43,
	0x54, 0x49, 0x41, 0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x52, 0x52, 0x53, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x4f, 0x4d, 0x54, 0x50, 0x10, 0x02, 0x4a, 0x04, 0x08, 0x01, 0x10, 0x02, 0x2a,
	0xe9, 0x01, 0x0a, 0x0d, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x1a, 0x0a, 0x16, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54,
	0x59, 0x50, 0x45, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a,
	0x11, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x44, 0x50, 0x10, 0x02, 0x12, 0x1e, 0x0a, 0x16, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f,
	0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x44, 0x50, 0x5f, 0x48, 0x44, 0x4d, 0x49, 0x10, 0x03,
	0x1a, 0x02, 0x08, 0x01, 0x12, 0x1a, 0x0a, 0x12, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f,
	0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x56, 0x47, 0x41, 0x10, 0x04, 0x1a, 0x02, 0x08, 0x01,
	0x12, 0x17, 0x0a, 0x13, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x48, 0x44, 0x4d, 0x49, 0x10, 0x05, 0x12, 0x15, 0x0a, 0x11, 0x43, 0x48, 0x41,
	0x4d, 0x45, 0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x56, 0x32, 0x10, 0x09,
	0x12, 0x15, 0x0a, 0x11, 0x43, 0x48, 0x41, 0x4d, 0x45, 0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x56, 0x33, 0x10, 0x0a, 0x12, 0x16, 0x0a, 0x12, 0x43, 0x48, 0x41, 0x4d, 0x45,
	0x4c, 0x45, 0x4f, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x52, 0x50, 0x49, 0x10, 0x0b, 0x22,
	0x04, 0x08, 0x01, 0x10, 0x01, 0x22, 0x04, 0x08, 0x06, 0x10, 0x08, 0x42, 0x35, 0x5a, 0x33, 0x69,
	0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65,
	0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x3b, 0x75, 0x66, 0x73,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescOnce sync.Once
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescData = file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDesc
)

func file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescGZIP() []byte {
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescOnce.Do(func() {
		file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescData)
	})
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDescData
}

var file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_goTypes = []interface{}{
	(ChameleonType)(0),                 // 0: unifiedfleet.api.v1.models.chromeos.lab.ChameleonType
	(Chameleon_AudioBoxJackPlugger)(0), // 1: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.AudioBoxJackPlugger
	(Chameleon_TRRSType)(0),            // 2: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.TRRSType
	(*Chameleon)(nil),                  // 3: unifiedfleet.api.v1.models.chromeos.lab.Chameleon
	(*OSRPM)(nil),                      // 4: unifiedfleet.api.v1.models.chromeos.lab.OSRPM
}
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_depIdxs = []int32{
	0, // 0: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.chameleon_peripherals:type_name -> unifiedfleet.api.v1.models.chromeos.lab.ChameleonType
	4, // 1: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.rpm:type_name -> unifiedfleet.api.v1.models.chromeos.lab.OSRPM
	1, // 2: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.audiobox_jackplugger:type_name -> unifiedfleet.api.v1.models.chromeos.lab.Chameleon.AudioBoxJackPlugger
	2, // 3: unifiedfleet.api.v1.models.chromeos.lab.Chameleon.trrs_type:type_name -> unifiedfleet.api.v1.models.chromeos.lab.Chameleon.TRRSType
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_init() }
func file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_init() {
	if File_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto != nil {
		return
	}
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_rpm_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Chameleon); i {
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
			RawDescriptor: file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_goTypes,
		DependencyIndexes: file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_depIdxs,
		EnumInfos:         file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_enumTypes,
		MessageInfos:      file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_msgTypes,
	}.Build()
	File_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto = out.File
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_rawDesc = nil
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_goTypes = nil
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chameleon_proto_depIdxs = nil
}
