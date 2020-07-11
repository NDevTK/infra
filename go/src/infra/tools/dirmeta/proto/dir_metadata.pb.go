// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: infra/tools/dirmeta/proto/dir_metadata.proto

package dirmetapb

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Supported OS to tag a directory.
type OS int32

const (
	OS_OS_UNSPECIFIED OS = 0
	OS_LINUX          OS = 1
	OS_WINDOWS        OS = 2
	OS_MAC            OS = 3
	OS_ANDROID        OS = 4
	OS_IOS            OS = 5
	OS_CHROME_OS      OS = 6
	OS_FUCHSIA        OS = 7
)

// Enum value maps for OS.
var (
	OS_name = map[int32]string{
		0: "OS_UNSPECIFIED",
		1: "LINUX",
		2: "WINDOWS",
		3: "MAC",
		4: "ANDROID",
		5: "IOS",
		6: "CHROME_OS",
		7: "FUCHSIA",
	}
	OS_value = map[string]int32{
		"OS_UNSPECIFIED": 0,
		"LINUX":          1,
		"WINDOWS":        2,
		"MAC":            3,
		"ANDROID":        4,
		"IOS":            5,
		"CHROME_OS":      6,
		"FUCHSIA":        7,
	}
)

func (x OS) Enum() *OS {
	p := new(OS)
	*p = x
	return p
}

func (x OS) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OS) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes[0].Descriptor()
}

func (OS) Type() protoreflect.EnumType {
	return &file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes[0]
}

func (x OS) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OS.Descriptor instead.
func (OS) EnumDescriptor() ([]byte, []int) {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP(), []int{0}
}

// Trinary is a tri-state boolean.
//
// It is convenient for conceptually-boolean fields that need inheritance.
// Plain bool does not work well because proto3 does not distinguish between
// false and absence of the value. Unlike bool, the default value of Trinary
// is TRINARY_UNSPECIFIED which is different from NO.
//
// Unfortunately the JSON representation are strings "YES" and "NO", as opposed
// to more natural true and false.
// The alternative to Trinary is BoolValue
// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/wrappers.proto#L101
// which has JSON representation as `true` and `false`, but then the text-proto
// representation is verbose: `wpt { notify { value: true }}`.
// Given that we will have more DIR_METADATA files than code that interpretes
// the files, we make a choice in favor of clean DIR_METADATA files.
type Trinary int32

const (
	Trinary_TRINARY_UNSPECIFIED Trinary = 0
	Trinary_YES                 Trinary = 1
	Trinary_NO                  Trinary = 2
)

// Enum value maps for Trinary.
var (
	Trinary_name = map[int32]string{
		0: "TRINARY_UNSPECIFIED",
		1: "YES",
		2: "NO",
	}
	Trinary_value = map[string]int32{
		"TRINARY_UNSPECIFIED": 0,
		"YES":                 1,
		"NO":                  2,
	}
)

func (x Trinary) Enum() *Trinary {
	p := new(Trinary)
	*p = x
	return p
}

func (x Trinary) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Trinary) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes[1].Descriptor()
}

func (Trinary) Type() protoreflect.EnumType {
	return &file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes[1]
}

func (x Trinary) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Trinary.Descriptor instead.
func (Trinary) EnumDescriptor() ([]byte, []int) {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP(), []int{1}
}

// Metadata information for a directory.
//
// Individual fields/subfields are inherited by subdirectories.
// For example, if a/DIR_METADATA specifies `monorail.project="chromium"` and
// a/b/DIR_METADATA file exists and does not specify `monorail.project`, then
// `a/b`'s Monorail project is "chromium".
// Note: this semantics can be extended to opt-out individual fields from
// inheritance.
type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Monorail metadata. Its component field should always map to the same
	// team_email field (see below) for every directory that it is listed in.
	Monorail *Monorail `protobuf:"bytes,1,opt,name=monorail,proto3" json:"monorail,omitempty"`
	// The team responsible for the directory. Only a single email is allowed.
	// For example, “team-email@chromium.org”.
	TeamEmail string `protobuf:"bytes,2,opt,name=team_email,json=teamEmail,proto3" json:"team_email,omitempty"`
	// The code in this directory is specific to this OS.
	Os OS `protobuf:"varint,3,opt,name=os,proto3,enum=chrome.dir_metadata.OS" json:"os,omitempty"`
	// WPT-specific metadata.
	Wpt *WPT `protobuf:"bytes,4,opt,name=wpt,proto3" json:"wpt,omitempty"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *Metadata) GetMonorail() *Monorail {
	if x != nil {
		return x.Monorail
	}
	return nil
}

func (x *Metadata) GetTeamEmail() string {
	if x != nil {
		return x.TeamEmail
	}
	return ""
}

func (x *Metadata) GetOs() OS {
	if x != nil {
		return x.Os
	}
	return OS_OS_UNSPECIFIED
}

func (x *Metadata) GetWpt() *WPT {
	if x != nil {
		return x.Wpt
	}
	return nil
}

// WPT is WebPlatformTest-specific metadata.
type WPT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether wpt-importer should notify the team about new failures.
	Notify Trinary `protobuf:"varint,1,opt,name=notify,proto3,enum=chrome.dir_metadata.Trinary" json:"notify,omitempty"`
}

func (x *WPT) Reset() {
	*x = WPT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WPT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WPT) ProtoMessage() {}

func (x *WPT) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WPT.ProtoReflect.Descriptor instead.
func (*WPT) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP(), []int{1}
}

func (x *WPT) GetNotify() Trinary {
	if x != nil {
		return x.Notify
	}
	return Trinary_TRINARY_UNSPECIFIED
}

// Monorail specific metadata.
type Monorail struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The Monorail project. For example, “chromium”.
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// The Monorail component. For example, “UI>Browser>Zoom”.
	// Only a single component is allowed.
	Component string `protobuf:"bytes,2,opt,name=component,proto3" json:"component,omitempty"`
}

func (x *Monorail) Reset() {
	*x = Monorail{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Monorail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Monorail) ProtoMessage() {}

func (x *Monorail) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Monorail.ProtoReflect.Descriptor instead.
func (*Monorail) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP(), []int{2}
}

func (x *Monorail) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Monorail) GetComponent() string {
	if x != nil {
		return x.Component
	}
	return ""
}

var File_infra_tools_dirmeta_proto_dir_metadata_proto protoreflect.FileDescriptor

var file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64, 0x69,
	0x72, 0x6d, 0x65, 0x74, 0x61, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x69, 0x72, 0x5f,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x22, 0xcc, 0x01, 0x0a, 0x08, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x39, 0x0a, 0x08, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69,
	0x6c, 0x52, 0x08, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x12, 0x1d, 0x0a, 0x0a, 0x74,
	0x65, 0x61, 0x6d, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x65, 0x61, 0x6d, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x27, 0x0a, 0x02, 0x6f, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e,
	0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4f, 0x53, 0x52,
	0x02, 0x6f, 0x73, 0x12, 0x2a, 0x0a, 0x03, 0x77, 0x70, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x57, 0x50, 0x54, 0x52, 0x03, 0x77, 0x70, 0x74, 0x4a,
	0x04, 0x08, 0x0d, 0x10, 0x0e, 0x52, 0x0b, 0x74, 0x68, 0x69, 0x72, 0x64, 0x5f, 0x70, 0x61, 0x72,
	0x74, 0x79, 0x22, 0x3b, 0x0a, 0x03, 0x57, 0x50, 0x54, 0x12, 0x34, 0x0a, 0x06, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x54, 0x72, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x52, 0x06, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x22,
	0x42, 0x0a, 0x08, 0x4d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65,
	0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e,
	0x65, 0x6e, 0x74, 0x2a, 0x6b, 0x0a, 0x02, 0x4f, 0x53, 0x12, 0x12, 0x0a, 0x0e, 0x4f, 0x53, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x09, 0x0a,
	0x05, 0x4c, 0x49, 0x4e, 0x55, 0x58, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x57, 0x49, 0x4e, 0x44,
	0x4f, 0x57, 0x53, 0x10, 0x02, 0x12, 0x07, 0x0a, 0x03, 0x4d, 0x41, 0x43, 0x10, 0x03, 0x12, 0x0b,
	0x0a, 0x07, 0x41, 0x4e, 0x44, 0x52, 0x4f, 0x49, 0x44, 0x10, 0x04, 0x12, 0x07, 0x0a, 0x03, 0x49,
	0x4f, 0x53, 0x10, 0x05, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x48, 0x52, 0x4f, 0x4d, 0x45, 0x5f, 0x4f,
	0x53, 0x10, 0x06, 0x12, 0x0b, 0x0a, 0x07, 0x46, 0x55, 0x43, 0x48, 0x53, 0x49, 0x41, 0x10, 0x07,
	0x2a, 0x33, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x12, 0x17, 0x0a, 0x13, 0x54,
	0x52, 0x49, 0x4e, 0x41, 0x52, 0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49,
	0x45, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x59, 0x45, 0x53, 0x10, 0x01, 0x12, 0x06, 0x0a,
	0x02, 0x4e, 0x4f, 0x10, 0x02, 0x42, 0x25, 0x5a, 0x23, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74,
	0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64, 0x69, 0x72, 0x6d, 0x65, 0x74, 0x61, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x3b, 0x64, 0x69, 0x72, 0x6d, 0x65, 0x74, 0x61, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescOnce sync.Once
	file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescData = file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDesc
)

func file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescGZIP() []byte {
	file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescOnce.Do(func() {
		file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescData)
	})
	return file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDescData
}

var file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_infra_tools_dirmeta_proto_dir_metadata_proto_goTypes = []interface{}{
	(OS)(0),          // 0: chrome.dir_metadata.OS
	(Trinary)(0),     // 1: chrome.dir_metadata.Trinary
	(*Metadata)(nil), // 2: chrome.dir_metadata.Metadata
	(*WPT)(nil),      // 3: chrome.dir_metadata.WPT
	(*Monorail)(nil), // 4: chrome.dir_metadata.Monorail
}
var file_infra_tools_dirmeta_proto_dir_metadata_proto_depIdxs = []int32{
	4, // 0: chrome.dir_metadata.Metadata.monorail:type_name -> chrome.dir_metadata.Monorail
	0, // 1: chrome.dir_metadata.Metadata.os:type_name -> chrome.dir_metadata.OS
	3, // 2: chrome.dir_metadata.Metadata.wpt:type_name -> chrome.dir_metadata.WPT
	1, // 3: chrome.dir_metadata.WPT.notify:type_name -> chrome.dir_metadata.Trinary
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_infra_tools_dirmeta_proto_dir_metadata_proto_init() }
func file_infra_tools_dirmeta_proto_dir_metadata_proto_init() {
	if File_infra_tools_dirmeta_proto_dir_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metadata); i {
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
		file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WPT); i {
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
		file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Monorail); i {
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
			RawDescriptor: file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_tools_dirmeta_proto_dir_metadata_proto_goTypes,
		DependencyIndexes: file_infra_tools_dirmeta_proto_dir_metadata_proto_depIdxs,
		EnumInfos:         file_infra_tools_dirmeta_proto_dir_metadata_proto_enumTypes,
		MessageInfos:      file_infra_tools_dirmeta_proto_dir_metadata_proto_msgTypes,
	}.Build()
	File_infra_tools_dirmeta_proto_dir_metadata_proto = out.File
	file_infra_tools_dirmeta_proto_dir_metadata_proto_rawDesc = nil
	file_infra_tools_dirmeta_proto_dir_metadata_proto_goTypes = nil
	file_infra_tools_dirmeta_proto_dir_metadata_proto_depIdxs = nil
}
