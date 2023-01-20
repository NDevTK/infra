// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: infra/experimental/golangbuild/golangbuildpb/params.proto

package golangbuildpb

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

type Inputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the project this is supposed to build against.
	//
	// Used for both validation, and necessary in order to identify when a
	// subrepo builder is triggered by a different repository, like the main
	// Go repository.
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// Indicates that this build should run with the race detector enabled.
	RaceMode bool `protobuf:"varint,2,opt,name=race_mode,json=raceMode,proto3" json:"race_mode,omitempty"`
	// Extra environment variables to set for building and testing.
	Env map[string]string `protobuf:"bytes,3,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Inputs) Reset() {
	*x = Inputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Inputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Inputs) ProtoMessage() {}

func (x *Inputs) ProtoReflect() protoreflect.Message {
	mi := &file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Inputs.ProtoReflect.Descriptor instead.
func (*Inputs) Descriptor() ([]byte, []int) {
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP(), []int{0}
}

func (x *Inputs) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Inputs) GetRaceMode() bool {
	if x != nil {
		return x.RaceMode
	}
	return false
}

func (x *Inputs) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

var File_infra_experimental_golangbuild_golangbuildpb_params_proto protoreflect.FileDescriptor

var file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc = []byte{
	0x0a, 0x39, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65,
	0x6e, 0x74, 0x61, 0x6c, 0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64,
	0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2f, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x67, 0x6f, 0x6c,
	0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x22, 0xa9, 0x01, 0x0a, 0x06, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12,
	0x1b, 0x0a, 0x09, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x72, 0x61, 0x63, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x30, 0x0a, 0x03,
	0x65, 0x6e, 0x76, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x67, 0x6f, 0x6c, 0x61,
	0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x2e, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65, 0x6e, 0x76, 0x1a, 0x36,
	0x0a, 0x08, 0x45, 0x6e, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x2e, 0x5a, 0x2c, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f,
	0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x6c, 0x2f, 0x67, 0x6f, 0x6c,
	0x61, 0x6e, 0x67, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescOnce sync.Once
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData = file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc
)

func file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescGZIP() []byte {
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescOnce.Do(func() {
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData)
	})
	return file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDescData
}

var file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes = []interface{}{
	(*Inputs)(nil), // 0: golangbuildpb.Inputs
	nil,            // 1: golangbuildpb.Inputs.EnvEntry
}
var file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs = []int32{
	1, // 0: golangbuildpb.Inputs.env:type_name -> golangbuildpb.Inputs.EnvEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_infra_experimental_golangbuild_golangbuildpb_params_proto_init() }
func file_infra_experimental_golangbuild_golangbuildpb_params_proto_init() {
	if File_infra_experimental_golangbuild_golangbuildpb_params_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Inputs); i {
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
			RawDescriptor: file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes,
		DependencyIndexes: file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs,
		MessageInfos:      file_infra_experimental_golangbuild_golangbuildpb_params_proto_msgTypes,
	}.Build()
	File_infra_experimental_golangbuild_golangbuildpb_params_proto = out.File
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_rawDesc = nil
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_goTypes = nil
	file_infra_experimental_golangbuild_golangbuildpb_params_proto_depIdxs = nil
}
