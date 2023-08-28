// Copyright 2019 The LUCI Authors.
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
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.7
// source: infra/appengine/arquebus/app/backend/model/model.proto

package model

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

// TaskStatus represents the status of an Assigner task.
type TaskStatus int32

const (
	// TaskStatusScheduled means that the task has been scheduled, but
	// has not yet started.
	TaskStatus_Scheduled TaskStatus = 0
	// Running means that the task is running.
	TaskStatus_Running TaskStatus = 1
	// Succeeded means that the task completed successfully.
	TaskStatus_Succeeded TaskStatus = 2
	// Failed means that the task failed.
	TaskStatus_Failed TaskStatus = 3
	// Cancelled means that the task was cancelled before starting.
	TaskStatus_Cancelled TaskStatus = 4
	// Aborted means that the task started, but was aborted manually or
	// due to hard deadline.
	TaskStatus_Aborted TaskStatus = 5
)

// Enum value maps for TaskStatus.
var (
	TaskStatus_name = map[int32]string{
		0: "Scheduled",
		1: "Running",
		2: "Succeeded",
		3: "Failed",
		4: "Cancelled",
		5: "Aborted",
	}
	TaskStatus_value = map[string]int32{
		"Scheduled": 0,
		"Running":   1,
		"Succeeded": 2,
		"Failed":    3,
		"Cancelled": 4,
		"Aborted":   5,
	}
)

func (x TaskStatus) Enum() *TaskStatus {
	p := new(TaskStatus)
	*p = x
	return p
}

func (x TaskStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_appengine_arquebus_app_backend_model_model_proto_enumTypes[0].Descriptor()
}

func (TaskStatus) Type() protoreflect.EnumType {
	return &file_infra_appengine_arquebus_app_backend_model_model_proto_enumTypes[0]
}

func (x TaskStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskStatus.Descriptor instead.
func (TaskStatus) EnumDescriptor() ([]byte, []int) {
	return file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescGZIP(), []int{0}
}

var File_infra_appengine_arquebus_app_backend_model_model_proto protoreflect.FileDescriptor

var file_infra_appengine_arquebus_app_backend_model_model_proto_rawDesc = []byte{
	0x0a, 0x36, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x61, 0x72, 0x71, 0x75, 0x65, 0x62, 0x75, 0x73, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x62,
	0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x6d, 0x6f, 0x64,
	0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2a, 0x5f, 0x0a, 0x0a, 0x54, 0x61, 0x73, 0x6b, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x65, 0x64, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x10,
	0x01, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09,
	0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x65, 0x64, 0x10, 0x04, 0x12, 0x0b, 0x0a, 0x07, 0x41,
	0x62, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x10, 0x05, 0x42, 0x2c, 0x5a, 0x2a, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x61, 0x72, 0x71, 0x75,
	0x65, 0x62, 0x75, 0x73, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescOnce sync.Once
	file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescData = file_infra_appengine_arquebus_app_backend_model_model_proto_rawDesc
)

func file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescGZIP() []byte {
	file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescOnce.Do(func() {
		file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescData)
	})
	return file_infra_appengine_arquebus_app_backend_model_model_proto_rawDescData
}

var file_infra_appengine_arquebus_app_backend_model_model_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_infra_appengine_arquebus_app_backend_model_model_proto_goTypes = []interface{}{
	(TaskStatus)(0), // 0: backend.model.TaskStatus
}
var file_infra_appengine_arquebus_app_backend_model_model_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_infra_appengine_arquebus_app_backend_model_model_proto_init() }
func file_infra_appengine_arquebus_app_backend_model_model_proto_init() {
	if File_infra_appengine_arquebus_app_backend_model_model_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_infra_appengine_arquebus_app_backend_model_model_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_appengine_arquebus_app_backend_model_model_proto_goTypes,
		DependencyIndexes: file_infra_appengine_arquebus_app_backend_model_model_proto_depIdxs,
		EnumInfos:         file_infra_appengine_arquebus_app_backend_model_model_proto_enumTypes,
	}.Build()
	File_infra_appengine_arquebus_app_backend_model_model_proto = out.File
	file_infra_appengine_arquebus_app_backend_model_model_proto_rawDesc = nil
	file_infra_appengine_arquebus_app_backend_model_model_proto_goTypes = nil
	file_infra_appengine_arquebus_app_backend_model_model_proto_depIdxs = nil
}
