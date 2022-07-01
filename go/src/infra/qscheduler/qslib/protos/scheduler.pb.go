// Copyright 2021 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.17.3
// source: infra/qscheduler/qslib/protos/scheduler.proto

package protos

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Scheduler encapsulates the state and configuration of a running
// quotascheduler for a single pool.
type Scheduler struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SchedulerState is the state of the scheduler.
	State *SchedulerState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	// SchedulerConfig is the config of the scheduler.
	Config *SchedulerConfig `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *Scheduler) Reset() {
	*x = Scheduler{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_qscheduler_qslib_protos_scheduler_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Scheduler) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Scheduler) ProtoMessage() {}

func (x *Scheduler) ProtoReflect() protoreflect.Message {
	mi := &file_infra_qscheduler_qslib_protos_scheduler_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Scheduler.ProtoReflect.Descriptor instead.
func (*Scheduler) Descriptor() ([]byte, []int) {
	return file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescGZIP(), []int{0}
}

func (x *Scheduler) GetState() *SchedulerState {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *Scheduler) GetConfig() *SchedulerConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

var File_infra_qscheduler_qslib_protos_scheduler_proto protoreflect.FileDescriptor

var file_infra_qscheduler_qslib_protos_scheduler_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x71, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x65, 0x72, 0x2f, 0x71, 0x73, 0x6c, 0x69, 0x62, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x1a, 0x29, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x71,
	0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2f, 0x71, 0x73, 0x6c, 0x69, 0x62, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x2a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x71, 0x73, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x72, 0x2f, 0x71, 0x73, 0x6c, 0x69, 0x62, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6a,
	0x0a, 0x09, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x12, 0x2c, 0x0a, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x2f, 0x0a, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x42, 0x1f, 0x5a, 0x1d, 0x69, 0x6e,
	0x66, 0x72, 0x61, 0x2f, 0x71, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2f, 0x71,
	0x73, 0x6c, 0x69, 0x62, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescOnce sync.Once
	file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescData = file_infra_qscheduler_qslib_protos_scheduler_proto_rawDesc
)

func file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescGZIP() []byte {
	file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescOnce.Do(func() {
		file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescData)
	})
	return file_infra_qscheduler_qslib_protos_scheduler_proto_rawDescData
}

var file_infra_qscheduler_qslib_protos_scheduler_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_qscheduler_qslib_protos_scheduler_proto_goTypes = []interface{}{
	(*Scheduler)(nil),       // 0: protos.Scheduler
	(*SchedulerState)(nil),  // 1: protos.SchedulerState
	(*SchedulerConfig)(nil), // 2: protos.SchedulerConfig
}
var file_infra_qscheduler_qslib_protos_scheduler_proto_depIdxs = []int32{
	1, // 0: protos.Scheduler.state:type_name -> protos.SchedulerState
	2, // 1: protos.Scheduler.config:type_name -> protos.SchedulerConfig
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_infra_qscheduler_qslib_protos_scheduler_proto_init() }
func file_infra_qscheduler_qslib_protos_scheduler_proto_init() {
	if File_infra_qscheduler_qslib_protos_scheduler_proto != nil {
		return
	}
	file_infra_qscheduler_qslib_protos_state_proto_init()
	file_infra_qscheduler_qslib_protos_config_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_qscheduler_qslib_protos_scheduler_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Scheduler); i {
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
			RawDescriptor: file_infra_qscheduler_qslib_protos_scheduler_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_qscheduler_qslib_protos_scheduler_proto_goTypes,
		DependencyIndexes: file_infra_qscheduler_qslib_protos_scheduler_proto_depIdxs,
		MessageInfos:      file_infra_qscheduler_qslib_protos_scheduler_proto_msgTypes,
	}.Build()
	File_infra_qscheduler_qslib_protos_scheduler_proto = out.File
	file_infra_qscheduler_qslib_protos_scheduler_proto_rawDesc = nil
	file_infra_qscheduler_qslib_protos_scheduler_proto_goTypes = nil
	file_infra_qscheduler_qslib_protos_scheduler_proto_depIdxs = nil
}
