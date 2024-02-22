// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// copy from google3/devtools/cider/services/build/companion/extractor.proto
// remove unused messages.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: extractor.proto

package proto

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

type Status_Code int32

const (
	Status_OK      Status_Code = 0
	Status_FAILURE Status_Code = 1
)

// Enum value maps for Status_Code.
var (
	Status_Code_name = map[int32]string{
		0: "OK",
		1: "FAILURE",
	}
	Status_Code_value = map[string]int32{
		"OK":      0,
		"FAILURE": 1,
	}
)

func (x Status_Code) Enum() *Status_Code {
	p := new(Status_Code)
	*p = x
	return p
}

func (x Status_Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status_Code) Descriptor() protoreflect.EnumDescriptor {
	return file_extractor_proto_enumTypes[0].Descriptor()
}

func (Status_Code) Type() protoreflect.EnumType {
	return &file_extractor_proto_enumTypes[0]
}

func (x Status_Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status_Code.Descriptor instead.
func (Status_Code) EnumDescriptor() ([]byte, []int) {
	return file_extractor_proto_rawDescGZIP(), []int{0, 0}
}

// Indicates the success/failure for analysis.
type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code Status_Code `protobuf:"varint,1,opt,name=code,proto3,enum=cider.services.build.companion.Status_Code" json:"code,omitempty"`
	// Details about the status, might be displayed to user.
	Message *string `protobuf:"bytes,2,opt,name=message,proto3,oneof" json:"message,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_extractor_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_extractor_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_extractor_proto_rawDescGZIP(), []int{0}
}

func (x *Status) GetCode() Status_Code {
	if x != nil {
		return x.Code
	}
	return Status_OK
}

func (x *Status) GetMessage() string {
	if x != nil && x.Message != nil {
		return *x.Message
	}
	return ""
}

type GeneratedFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to the file relative to IdeAnalysis.build_artifact_root.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// The text of the generated file, if not provided contents will be read
	// from the path above in user's workstation.
	Contents []byte `protobuf:"bytes,2,opt,name=contents,proto3,oneof" json:"contents,omitempty"`
}

func (x *GeneratedFile) Reset() {
	*x = GeneratedFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_extractor_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeneratedFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeneratedFile) ProtoMessage() {}

func (x *GeneratedFile) ProtoReflect() protoreflect.Message {
	mi := &file_extractor_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeneratedFile.ProtoReflect.Descriptor instead.
func (*GeneratedFile) Descriptor() ([]byte, []int) {
	return file_extractor_proto_rawDescGZIP(), []int{1}
}

func (x *GeneratedFile) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *GeneratedFile) GetContents() []byte {
	if x != nil {
		return x.Contents
	}
	return nil
}

type SourceFile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to the source file relative to the repo root.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// Working directory used by the build system. All the relative
	// paths in compiler_arguments should be relative to this path.
	// Relative to workspace root.
	WorkingDir string `protobuf:"bytes,2,opt,name=working_dir,json=workingDir,proto3" json:"working_dir,omitempty"`
	// Compiler arguments to compile the source file. If multiple variants
	// of the module being compiled are possible, the query script will choose
	// one.
	CompilerArguments []string `protobuf:"bytes,3,rep,name=compiler_arguments,json=compilerArguments,proto3" json:"compiler_arguments,omitempty"`
	// Any generated files that are used in compiling the file.
	Generated []*GeneratedFile `protobuf:"bytes,4,rep,name=generated,proto3" json:"generated,omitempty"`
	// Paths to all of the sources, like build files, code generators,
	// proto files etc. that were used  during analysis. Used to figure
	// out when a set of build artifacts are stale and the query tool
	// must be re-run.
	// Relative to workspace root.
	Deps []string `protobuf:"bytes,5,rep,name=deps,proto3" json:"deps,omitempty"`
	// Represensts analysis status for this particular file. e.g. not part
	// of the build graph.
	Status *Status `protobuf:"bytes,6,opt,name=status,proto3,oneof" json:"status,omitempty"`
}

func (x *SourceFile) Reset() {
	*x = SourceFile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_extractor_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SourceFile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SourceFile) ProtoMessage() {}

func (x *SourceFile) ProtoReflect() protoreflect.Message {
	mi := &file_extractor_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SourceFile.ProtoReflect.Descriptor instead.
func (*SourceFile) Descriptor() ([]byte, []int) {
	return file_extractor_proto_rawDescGZIP(), []int{2}
}

func (x *SourceFile) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *SourceFile) GetWorkingDir() string {
	if x != nil {
		return x.WorkingDir
	}
	return ""
}

func (x *SourceFile) GetCompilerArguments() []string {
	if x != nil {
		return x.CompilerArguments
	}
	return nil
}

func (x *SourceFile) GetGenerated() []*GeneratedFile {
	if x != nil {
		return x.Generated
	}
	return nil
}

func (x *SourceFile) GetDeps() []string {
	if x != nil {
		return x.Deps
	}
	return nil
}

func (x *SourceFile) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

type IdeAnalysis struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path relative to workspace root, containing all the artifacts
	// generated by the build system. GeneratedFile.path are always
	// relative to this directory.
	BuildArtifactRoot string        `protobuf:"bytes,1,opt,name=build_artifact_root,json=buildArtifactRoot,proto3" json:"build_artifact_root,omitempty"`
	Sources           []*SourceFile `protobuf:"bytes,2,rep,name=sources,proto3" json:"sources,omitempty"`
	// Status representing overall analysis.
	// Should fail only when no analysis can be performed, e.g. workspace
	// isn't setup.
	Status *Status `protobuf:"bytes,3,opt,name=status,proto3,oneof" json:"status,omitempty"`
}

func (x *IdeAnalysis) Reset() {
	*x = IdeAnalysis{}
	if protoimpl.UnsafeEnabled {
		mi := &file_extractor_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IdeAnalysis) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdeAnalysis) ProtoMessage() {}

func (x *IdeAnalysis) ProtoReflect() protoreflect.Message {
	mi := &file_extractor_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdeAnalysis.ProtoReflect.Descriptor instead.
func (*IdeAnalysis) Descriptor() ([]byte, []int) {
	return file_extractor_proto_rawDescGZIP(), []int{3}
}

func (x *IdeAnalysis) GetBuildArtifactRoot() string {
	if x != nil {
		return x.BuildArtifactRoot
	}
	return ""
}

func (x *IdeAnalysis) GetSources() []*SourceFile {
	if x != nil {
		return x.Sources
	}
	return nil
}

func (x *IdeAnalysis) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

var File_extractor_proto protoreflect.FileDescriptor

var file_extractor_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x65, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x1e, 0x63, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x69, 0x6f,
	0x6e, 0x22, 0x91, 0x01, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x3f, 0x0a, 0x04,
	0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x63, 0x69, 0x64,
	0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x1d, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x22, 0x1b, 0x0a, 0x04,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x06, 0x0a, 0x02, 0x4f, 0x4b, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07,
	0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x51, 0x0a, 0x0d, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1f, 0x0a, 0x08, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x08,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x88, 0x01, 0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xa1, 0x02, 0x0a, 0x0a, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1f, 0x0a, 0x0b, 0x77,
	0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x64, 0x69, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x44, 0x69, 0x72, 0x12, 0x2d, 0x0a, 0x12,
	0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c, 0x65, 0x72, 0x5f, 0x61, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x11, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c,
	0x65, 0x72, 0x41, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x4b, 0x0a, 0x09, 0x67,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2d,
	0x2e, 0x63, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x69, 0x6f, 0x6e, 0x2e,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x09, 0x67,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x70, 0x73,
	0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x70, 0x73, 0x12, 0x43, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63,
	0x69, 0x64, 0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x48, 0x00, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x88, 0x01,
	0x01, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0xd3, 0x01, 0x0a,
	0x0b, 0x49, 0x64, 0x65, 0x41, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x12, 0x2e, 0x0a, 0x13,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x5f, 0x72,
	0x6f, 0x6f, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x44, 0x0a, 0x07,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e,
	0x63, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x69, 0x6f, 0x6e, 0x2e, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x07, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0x12, 0x43, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x2e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x6e,
	0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x48, 0x00, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x42, 0x2e, 0x5a, 0x2c, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x2f, 0x73, 0x69, 0x73, 0x6f, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x75, 0x70, 0x70, 0x6f,
	0x72, 0x74, 0x2f, 0x63, 0x69, 0x64, 0x65, 0x72, 0x75, 0x74, 0x69, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_extractor_proto_rawDescOnce sync.Once
	file_extractor_proto_rawDescData = file_extractor_proto_rawDesc
)

func file_extractor_proto_rawDescGZIP() []byte {
	file_extractor_proto_rawDescOnce.Do(func() {
		file_extractor_proto_rawDescData = protoimpl.X.CompressGZIP(file_extractor_proto_rawDescData)
	})
	return file_extractor_proto_rawDescData
}

var file_extractor_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_extractor_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_extractor_proto_goTypes = []interface{}{
	(Status_Code)(0),      // 0: cider.services.build.companion.Status.Code
	(*Status)(nil),        // 1: cider.services.build.companion.Status
	(*GeneratedFile)(nil), // 2: cider.services.build.companion.GeneratedFile
	(*SourceFile)(nil),    // 3: cider.services.build.companion.SourceFile
	(*IdeAnalysis)(nil),   // 4: cider.services.build.companion.IdeAnalysis
}
var file_extractor_proto_depIdxs = []int32{
	0, // 0: cider.services.build.companion.Status.code:type_name -> cider.services.build.companion.Status.Code
	2, // 1: cider.services.build.companion.SourceFile.generated:type_name -> cider.services.build.companion.GeneratedFile
	1, // 2: cider.services.build.companion.SourceFile.status:type_name -> cider.services.build.companion.Status
	3, // 3: cider.services.build.companion.IdeAnalysis.sources:type_name -> cider.services.build.companion.SourceFile
	1, // 4: cider.services.build.companion.IdeAnalysis.status:type_name -> cider.services.build.companion.Status
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_extractor_proto_init() }
func file_extractor_proto_init() {
	if File_extractor_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_extractor_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
		file_extractor_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeneratedFile); i {
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
		file_extractor_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SourceFile); i {
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
		file_extractor_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IdeAnalysis); i {
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
	file_extractor_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_extractor_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_extractor_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_extractor_proto_msgTypes[3].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_extractor_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_extractor_proto_goTypes,
		DependencyIndexes: file_extractor_proto_depIdxs,
		EnumInfos:         file_extractor_proto_enumTypes,
		MessageInfos:      file_extractor_proto_msgTypes,
	}.Build()
	File_extractor_proto = out.File
	file_extractor_proto_rawDesc = nil
	file_extractor_proto_goTypes = nil
	file_extractor_proto_depIdxs = nil
}
