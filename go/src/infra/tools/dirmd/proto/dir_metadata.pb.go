// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: infra/tools/dirmd/proto/dir_metadata.proto

package dirmdpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	chromeos "infra/tools/dirmd/proto/chromeos"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

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
	OS_LACROS         OS = 8
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
		8: "LACROS",
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
		"LACROS":         8,
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
	return file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes[0].Descriptor()
}

func (OS) Type() protoreflect.EnumType {
	return &file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes[0]
}

func (x OS) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OS.Descriptor instead.
func (OS) EnumDescriptor() ([]byte, []int) {
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{0}
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
// Given that we will have more DIR_METADATA files than code that interprets
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
	return file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes[1].Descriptor()
}

func (Trinary) Type() protoreflect.EnumType {
	return &file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes[1]
}

func (x Trinary) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Trinary.Descriptor instead.
func (Trinary) EnumDescriptor() ([]byte, []int) {
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{1}
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

	// Paths to Metadata files to import into this file.
	//
	// A mixin is imported AS IS, without inheritance from its ancestors.
	// To prevent confusion, a mixin's base name may not be DIR_METADATA.
	// Instead, the file name must entail its scope, e.g. the software component
	// that it is describing.
	//
	// Mixins can be useful to reduce duplication of metadata across the repo.
	// For example in Java, implementation and tests of the same feature live in
	// different directories. Common metadata can be placed in a mixin, and then
	// imported by DIR_METADATA files in both directories.
	//
	// A path must be source-absolute, e.g. "//components/autofill/COMMON_METADATA"
	// Here the root is the root of the repository which contains the current
	// file.
	//
	// Precedence:
	//   - Mixins override values in the parent directory.
	//   - The order of mixins is significant: if the same attribute is defined in
	//     two mixins, then the later one wins.
	//   - Values defined in the importee override values in the mixins.
	//
	// Importing mixins in mixins is not supported.
	// Only DIR_METADATA files are allowed to use this field.
	Mixins []string `protobuf:"bytes,10,rep,name=mixins,proto3" json:"mixins,omitempty"`
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
	// Private buganizer metadata.
	Buganizer *Buganizer `protobuf:"bytes,5,opt,name=buganizer,proto3" json:"buganizer,omitempty"`
	// ResultDB tags.
	Resultdb *ResultDB `protobuf:"bytes,6,opt,name=resultdb,proto3" json:"resultdb,omitempty"`
	// Public buganizer metadata.
	BuganizerPublic *Buganizer `protobuf:"bytes,7,opt,name=buganizer_public,json=buganizerPublic,proto3" json:"buganizer_public,omitempty"`
	// ChromeOS specific metadata.
	Chromeos *chromeos.ChromeOS `protobuf:"bytes,9,opt,name=chromeos,proto3" json:"chromeos,omitempty"`
	// Metadata Overrides
	//
	// Used for files that require different Metadata from the one specified in
	// the same dir or inherited from levels above. Override metadata should only
	// apply to files within the directory specified, and cannot be inherited,
	// meaning that it only applies to the files directly specified in the
	// directory. For example, overrides in a/*.txt will not apply to files in
	// a/b/*.txt. Values for Metadata specified in these ovrrides should follow
	// the same rules. Mixins defined under overrides are not supported.
	Overrides []*MetadataOverride `protobuf:"bytes,11,rep,name=overrides,proto3" json:"overrides,omitempty"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[0]
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
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *Metadata) GetMixins() []string {
	if x != nil {
		return x.Mixins
	}
	return nil
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

func (x *Metadata) GetBuganizer() *Buganizer {
	if x != nil {
		return x.Buganizer
	}
	return nil
}

func (x *Metadata) GetResultdb() *ResultDB {
	if x != nil {
		return x.Resultdb
	}
	return nil
}

func (x *Metadata) GetBuganizerPublic() *Buganizer {
	if x != nil {
		return x.BuganizerPublic
	}
	return nil
}

func (x *Metadata) GetChromeos() *chromeos.ChromeOS {
	if x != nil {
		return x.Chromeos
	}
	return nil
}

func (x *Metadata) GetOverrides() []*MetadataOverride {
	if x != nil {
		return x.Overrides
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
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WPT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WPT) ProtoMessage() {}

func (x *WPT) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[1]
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
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{1}
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
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Monorail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Monorail) ProtoMessage() {}

func (x *Monorail) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[2]
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
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{2}
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

// Buganizer specific metadata.
type Buganizer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The component id. For example, 690873.
	ComponentId int64 `protobuf:"varint,1,opt,name=component_id,json=componentId,proto3" json:"component_id,omitempty"`
}

func (x *Buganizer) Reset() {
	*x = Buganizer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Buganizer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Buganizer) ProtoMessage() {}

func (x *Buganizer) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Buganizer.ProtoReflect.Descriptor instead.
func (*Buganizer) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{3}
}

func (x *Buganizer) GetComponentId() int64 {
	if x != nil {
		return x.ComponentId
	}
	return 0
}

// Metadata override definition
type MetadataOverride struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// a list of regex strings defining file_patterns.
	//
	// all file_patterns listed here are relative to the location of the
	// DIR_METADATA file. file patterns listed here are directly passed to git
	// ls-files to find all files associated with that pattern. for example, if
	// DIR_METADATA is specified at a/b/DIR_METADATA, and file_pattern is
	// \*.proto, `git ls-files --full-name a/b/\*.proto` will be run to identify
	// files.
	FilePatterns []string `protobuf:"bytes,1,rep,name=file_patterns,json=filePatterns,proto3" json:"file_patterns,omitempty"`
	// Note that mixin imports are not supported for override definitions.
	Metadata *Metadata `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *MetadataOverride) Reset() {
	*x = MetadataOverride{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetadataOverride) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetadataOverride) ProtoMessage() {}

func (x *MetadataOverride) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetadataOverride.ProtoReflect.Descriptor instead.
func (*MetadataOverride) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{4}
}

func (x *MetadataOverride) GetFilePatterns() []string {
	if x != nil {
		return x.FilePatterns
	}
	return nil
}

func (x *MetadataOverride) GetMetadata() *Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// ResultDB contains a list of tags to be attached to results of tests
// defined in a directory.
type ResultDB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A list of tags to be attached to results of tests defined in a directory.
	// Each tag must be formatted as "key:value", and keys can be repeated.
	Tags []string `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *ResultDB) Reset() {
	*x = ResultDB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResultDB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResultDB) ProtoMessage() {}

func (x *ResultDB) ProtoReflect() protoreflect.Message {
	mi := &file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResultDB.ProtoReflect.Descriptor instead.
func (*ResultDB) Descriptor() ([]byte, []int) {
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP(), []int{5}
}

func (x *ResultDB) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

var File_infra_tools_dirmd_proto_dir_metadata_proto protoreflect.FileDescriptor

var file_infra_tools_dirmd_proto_dir_metadata_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64, 0x69,
	0x72, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63, 0x68,
	0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x1a, 0x2f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64,
	0x69, 0x72, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x65, 0x6f, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xc5, 0x04, 0x0a, 0x08, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x16, 0x0a, 0x06, 0x6d, 0x69, 0x78, 0x69, 0x6e, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x06, 0x6d, 0x69, 0x78, 0x69, 0x6e, 0x73, 0x12, 0x39, 0x0a, 0x08, 0x6d, 0x6f, 0x6e, 0x6f, 0x72,
	0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x4d, 0x6f, 0x6e, 0x6f, 0x72, 0x61, 0x69, 0x6c, 0x52, 0x08, 0x6d, 0x6f, 0x6e, 0x6f, 0x72, 0x61,
	0x69, 0x6c, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x65, 0x61, 0x6d, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x65, 0x61, 0x6d, 0x45, 0x6d, 0x61, 0x69,
	0x6c, 0x12, 0x27, 0x0a, 0x02, 0x6f, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x2e, 0x4f, 0x53, 0x52, 0x02, 0x6f, 0x73, 0x12, 0x2a, 0x0a, 0x03, 0x77, 0x70,
	0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65,
	0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x57, 0x50,
	0x54, 0x52, 0x03, 0x77, 0x70, 0x74, 0x12, 0x3c, 0x0a, 0x09, 0x62, 0x75, 0x67, 0x61, 0x6e, 0x69,
	0x7a, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x42, 0x75, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x52, 0x09, 0x62, 0x75, 0x67, 0x61, 0x6e,
	0x69, 0x7a, 0x65, 0x72, 0x12, 0x39, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x64, 0x62,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e,
	0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x44, 0x42, 0x52, 0x08, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x64, 0x62, 0x12,
	0x49, 0x0a, 0x10, 0x62, 0x75, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x5f, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x42, 0x75, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x52, 0x0f, 0x62, 0x75, 0x67, 0x61, 0x6e,
	0x69, 0x7a, 0x65, 0x72, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x12, 0x42, 0x0a, 0x08, 0x63, 0x68,
	0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63,
	0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x43, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x4f, 0x53, 0x52, 0x08, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x12, 0x43,
	0x0a, 0x09, 0x6f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x73, 0x18, 0x0b, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x25, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x4f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x52, 0x09, 0x6f, 0x76, 0x65, 0x72, 0x72, 0x69,
	0x64, 0x65, 0x73, 0x4a, 0x04, 0x08, 0x08, 0x10, 0x09, 0x4a, 0x04, 0x08, 0x0d, 0x10, 0x0e, 0x52,
	0x0c, 0x69, 0x6e, 0x68, 0x65, 0x72, 0x69, 0x74, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x52, 0x0b, 0x74,
	0x68, 0x69, 0x72, 0x64, 0x5f, 0x70, 0x61, 0x72, 0x74, 0x79, 0x22, 0x3b, 0x0a, 0x03, 0x57, 0x50,
	0x54, 0x12, 0x34, 0x0a, 0x06, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1c, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x64, 0x69, 0x72, 0x5f, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x54, 0x72, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x52,
	0x06, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x22, 0x42, 0x0a, 0x08, 0x4d, 0x6f, 0x6e, 0x6f, 0x72,
	0x61, 0x69, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x22, 0x2e, 0x0a, 0x09, 0x42,
	0x75, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x70,
	0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b,
	0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x22, 0x72, 0x0a, 0x10, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x4f, 0x76, 0x65, 0x72, 0x72, 0x69, 0x64, 0x65, 0x12,
	0x23, 0x0a, 0x0d, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x69, 0x6c, 0x65, 0x50, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x73, 0x12, 0x39, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e,
	0x64, 0x69, 0x72, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22,
	0x1e, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x44, 0x42, 0x12, 0x12, 0x0a, 0x04, 0x74,
	0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x2a,
	0x77, 0x0a, 0x02, 0x4f, 0x53, 0x12, 0x12, 0x0a, 0x0e, 0x4f, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x4c, 0x49, 0x4e,
	0x55, 0x58, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x57, 0x49, 0x4e, 0x44, 0x4f, 0x57, 0x53, 0x10,
	0x02, 0x12, 0x07, 0x0a, 0x03, 0x4d, 0x41, 0x43, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x4e,
	0x44, 0x52, 0x4f, 0x49, 0x44, 0x10, 0x04, 0x12, 0x07, 0x0a, 0x03, 0x49, 0x4f, 0x53, 0x10, 0x05,
	0x12, 0x0d, 0x0a, 0x09, 0x43, 0x48, 0x52, 0x4f, 0x4d, 0x45, 0x5f, 0x4f, 0x53, 0x10, 0x06, 0x12,
	0x0b, 0x0a, 0x07, 0x46, 0x55, 0x43, 0x48, 0x53, 0x49, 0x41, 0x10, 0x07, 0x12, 0x0a, 0x0a, 0x06,
	0x4c, 0x41, 0x43, 0x52, 0x4f, 0x53, 0x10, 0x08, 0x2a, 0x33, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x6e,
	0x61, 0x72, 0x79, 0x12, 0x17, 0x0a, 0x13, 0x54, 0x52, 0x49, 0x4e, 0x41, 0x52, 0x59, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03,
	0x59, 0x45, 0x53, 0x10, 0x01, 0x12, 0x06, 0x0a, 0x02, 0x4e, 0x4f, 0x10, 0x02, 0x42, 0x21, 0x5a,
	0x1f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x64, 0x69, 0x72,
	0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x64, 0x69, 0x72, 0x6d, 0x64, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescOnce sync.Once
	file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescData = file_infra_tools_dirmd_proto_dir_metadata_proto_rawDesc
)

func file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescGZIP() []byte {
	file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescOnce.Do(func() {
		file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescData)
	})
	return file_infra_tools_dirmd_proto_dir_metadata_proto_rawDescData
}

var file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_infra_tools_dirmd_proto_dir_metadata_proto_goTypes = []interface{}{
	(OS)(0),                   // 0: chrome.dir_metadata.OS
	(Trinary)(0),              // 1: chrome.dir_metadata.Trinary
	(*Metadata)(nil),          // 2: chrome.dir_metadata.Metadata
	(*WPT)(nil),               // 3: chrome.dir_metadata.WPT
	(*Monorail)(nil),          // 4: chrome.dir_metadata.Monorail
	(*Buganizer)(nil),         // 5: chrome.dir_metadata.Buganizer
	(*MetadataOverride)(nil),  // 6: chrome.dir_metadata.MetadataOverride
	(*ResultDB)(nil),          // 7: chrome.dir_metadata.ResultDB
	(*chromeos.ChromeOS)(nil), // 8: chrome.dir_metadata.chromeos.ChromeOS
}
var file_infra_tools_dirmd_proto_dir_metadata_proto_depIdxs = []int32{
	4,  // 0: chrome.dir_metadata.Metadata.monorail:type_name -> chrome.dir_metadata.Monorail
	0,  // 1: chrome.dir_metadata.Metadata.os:type_name -> chrome.dir_metadata.OS
	3,  // 2: chrome.dir_metadata.Metadata.wpt:type_name -> chrome.dir_metadata.WPT
	5,  // 3: chrome.dir_metadata.Metadata.buganizer:type_name -> chrome.dir_metadata.Buganizer
	7,  // 4: chrome.dir_metadata.Metadata.resultdb:type_name -> chrome.dir_metadata.ResultDB
	5,  // 5: chrome.dir_metadata.Metadata.buganizer_public:type_name -> chrome.dir_metadata.Buganizer
	8,  // 6: chrome.dir_metadata.Metadata.chromeos:type_name -> chrome.dir_metadata.chromeos.ChromeOS
	6,  // 7: chrome.dir_metadata.Metadata.overrides:type_name -> chrome.dir_metadata.MetadataOverride
	1,  // 8: chrome.dir_metadata.WPT.notify:type_name -> chrome.dir_metadata.Trinary
	2,  // 9: chrome.dir_metadata.MetadataOverride.metadata:type_name -> chrome.dir_metadata.Metadata
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_infra_tools_dirmd_proto_dir_metadata_proto_init() }
func file_infra_tools_dirmd_proto_dir_metadata_proto_init() {
	if File_infra_tools_dirmd_proto_dir_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Buganizer); i {
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
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetadataOverride); i {
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
		file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResultDB); i {
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
			RawDescriptor: file_infra_tools_dirmd_proto_dir_metadata_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_tools_dirmd_proto_dir_metadata_proto_goTypes,
		DependencyIndexes: file_infra_tools_dirmd_proto_dir_metadata_proto_depIdxs,
		EnumInfos:         file_infra_tools_dirmd_proto_dir_metadata_proto_enumTypes,
		MessageInfos:      file_infra_tools_dirmd_proto_dir_metadata_proto_msgTypes,
	}.Build()
	File_infra_tools_dirmd_proto_dir_metadata_proto = out.File
	file_infra_tools_dirmd_proto_dir_metadata_proto_rawDesc = nil
	file_infra_tools_dirmd_proto_dir_metadata_proto_goTypes = nil
	file_infra_tools_dirmd_proto_dir_metadata_proto_depIdxs = nil
}