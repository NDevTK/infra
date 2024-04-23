// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: infra/appengine/drone-queen/internal/config/config.proto

package config

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config is the configuration data served by luci-config for this app.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// access_groups are the luci-auth groups controlling access to RPC endpoints.
	AccessGroups *AccessGroups `protobuf:"bytes,1,opt,name=access_groups,json=accessGroups,proto3" json:"access_groups,omitempty"`
	// assignment_duration is the duration before expiration for drone
	// assignments.
	AssignmentDuration *durationpb.Duration `protobuf:"bytes,2,opt,name=assignment_duration,json=assignmentDuration,proto3" json:"assignment_duration,omitempty"`
	// instance identifies which instance of the service this is.  For
	// example, this could be prod for the prod instance.
	Instance string `protobuf:"bytes,3,opt,name=instance,proto3" json:"instance,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_infra_appengine_drone_queen_internal_config_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetAccessGroups() *AccessGroups {
	if x != nil {
		return x.AccessGroups
	}
	return nil
}

func (x *Config) GetAssignmentDuration() *durationpb.Duration {
	if x != nil {
		return x.AssignmentDuration
	}
	return nil
}

func (x *Config) GetInstance() string {
	if x != nil {
		return x.Instance
	}
	return ""
}

// AccessGroups holds access group configuration
type AccessGroups struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// drones is the group for calling drone RPCs.
	Drones string `protobuf:"bytes,1,opt,name=drones,proto3" json:"drones,omitempty"`
	// inventory_providers is the group for calling inventory RPCs.
	InventoryProviders string `protobuf:"bytes,2,opt,name=inventory_providers,json=inventoryProviders,proto3" json:"inventory_providers,omitempty"`
	// inspectors is the group for calling inspection RPCs.
	Inspectors string `protobuf:"bytes,3,opt,name=inspectors,proto3" json:"inspectors,omitempty"`
}

func (x *AccessGroups) Reset() {
	*x = AccessGroups{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccessGroups) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccessGroups) ProtoMessage() {}

func (x *AccessGroups) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccessGroups.ProtoReflect.Descriptor instead.
func (*AccessGroups) Descriptor() ([]byte, []int) {
	return file_infra_appengine_drone_queen_internal_config_config_proto_rawDescGZIP(), []int{1}
}

func (x *AccessGroups) GetDrones() string {
	if x != nil {
		return x.Drones
	}
	return ""
}

func (x *AccessGroups) GetInventoryProviders() string {
	if x != nil {
		return x.InventoryProviders
	}
	return ""
}

func (x *AccessGroups) GetInspectors() string {
	if x != nil {
		return x.Inspectors
	}
	return ""
}

var File_infra_appengine_drone_queen_internal_config_config_proto protoreflect.FileDescriptor

var file_infra_appengine_drone_queen_internal_config_config_proto_rawDesc = []byte{
	0x0a, 0x38, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x64, 0x72, 0x6f, 0x6e, 0x65, 0x2d, 0x71, 0x75, 0x65, 0x65, 0x6e, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x64, 0x72, 0x6f, 0x6e,
	0x65, 0x5f, 0x71, 0x75, 0x65, 0x65, 0x6e, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x1a, 0x1e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb7,
	0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x45, 0x0a, 0x0d, 0x61, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x20, 0x2e, 0x64, 0x72, 0x6f, 0x6e, 0x65, 0x5f, 0x71, 0x75, 0x65, 0x65, 0x6e, 0x2e, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x73, 0x52, 0x0c, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73,
	0x12, 0x4a, 0x0a, 0x13, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x12, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e,
	0x6d, 0x65, 0x6e, 0x74, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x22, 0x77, 0x0a, 0x0c, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x72, 0x6f, 0x6e,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x72, 0x6f, 0x6e, 0x65, 0x73,
	0x12, 0x2f, 0x0a, 0x13, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x69,
	0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72,
	0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x73, 0x42, 0x2d, 0x5a, 0x2b, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e,
	0x67, 0x69, 0x6e, 0x65, 0x2f, 0x64, 0x72, 0x6f, 0x6e, 0x65, 0x2d, 0x71, 0x75, 0x65, 0x65, 0x6e,
	0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_appengine_drone_queen_internal_config_config_proto_rawDescOnce sync.Once
	file_infra_appengine_drone_queen_internal_config_config_proto_rawDescData = file_infra_appengine_drone_queen_internal_config_config_proto_rawDesc
)

func file_infra_appengine_drone_queen_internal_config_config_proto_rawDescGZIP() []byte {
	file_infra_appengine_drone_queen_internal_config_config_proto_rawDescOnce.Do(func() {
		file_infra_appengine_drone_queen_internal_config_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_drone_queen_internal_config_config_proto_rawDescData)
	})
	return file_infra_appengine_drone_queen_internal_config_config_proto_rawDescData
}

var file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_infra_appengine_drone_queen_internal_config_config_proto_goTypes = []interface{}{
	(*Config)(nil),              // 0: drone_queen.config.Config
	(*AccessGroups)(nil),        // 1: drone_queen.config.AccessGroups
	(*durationpb.Duration)(nil), // 2: google.protobuf.Duration
}
var file_infra_appengine_drone_queen_internal_config_config_proto_depIdxs = []int32{
	1, // 0: drone_queen.config.Config.access_groups:type_name -> drone_queen.config.AccessGroups
	2, // 1: drone_queen.config.Config.assignment_duration:type_name -> google.protobuf.Duration
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_infra_appengine_drone_queen_internal_config_config_proto_init() }
func file_infra_appengine_drone_queen_internal_config_config_proto_init() {
	if File_infra_appengine_drone_queen_internal_config_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccessGroups); i {
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
			RawDescriptor: file_infra_appengine_drone_queen_internal_config_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_appengine_drone_queen_internal_config_config_proto_goTypes,
		DependencyIndexes: file_infra_appengine_drone_queen_internal_config_config_proto_depIdxs,
		MessageInfos:      file_infra_appengine_drone_queen_internal_config_config_proto_msgTypes,
	}.Build()
	File_infra_appengine_drone_queen_internal_config_config_proto = out.File
	file_infra_appengine_drone_queen_internal_config_config_proto_rawDesc = nil
	file_infra_appengine_drone_queen_internal_config_config_proto_goTypes = nil
	file_infra_appengine_drone_queen_internal_config_config_proto_depIdxs = nil
}
