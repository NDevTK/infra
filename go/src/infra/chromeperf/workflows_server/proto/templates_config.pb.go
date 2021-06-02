// Copyright 2020 The Chromium Authors.
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
// limitations under the License

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.0
// source: infra/chromeperf/workflows_server/templates_config.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	workflows "infra/chromeperf/workflows"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WorkflowTemplatesConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Templates []*workflows.WorkflowTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates,omitempty"`
}

func (x *WorkflowTemplatesConfig) Reset() {
	*x = WorkflowTemplatesConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_chromeperf_workflows_server_templates_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowTemplatesConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowTemplatesConfig) ProtoMessage() {}

func (x *WorkflowTemplatesConfig) ProtoReflect() protoreflect.Message {
	mi := &file_infra_chromeperf_workflows_server_templates_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowTemplatesConfig.ProtoReflect.Descriptor instead.
func (*WorkflowTemplatesConfig) Descriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_server_templates_config_proto_rawDescGZIP(), []int{0}
}

func (x *WorkflowTemplatesConfig) GetTemplates() []*workflows.WorkflowTemplate {
	if x != nil {
		return x.Templates
	}
	return nil
}

var File_infra_chromeperf_workflows_server_templates_config_proto protoreflect.FileDescriptor

var file_infra_chromeperf_workflows_server_templates_config_proto_rawDesc = []byte{
	0x0a, 0x38, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65,
	0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x5f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x33, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65, 0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f,
	0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x54, 0x0a, 0x17, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x73, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x39, 0x0a, 0x09, 0x74, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x09, 0x74, 0x65, 0x6d, 0x70,
	0x6c, 0x61, 0x74, 0x65, 0x73, 0x42, 0x29, 0x5a, 0x27, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63,
	0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65, 0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_chromeperf_workflows_server_templates_config_proto_rawDescOnce sync.Once
	file_infra_chromeperf_workflows_server_templates_config_proto_rawDescData = file_infra_chromeperf_workflows_server_templates_config_proto_rawDesc
)

func file_infra_chromeperf_workflows_server_templates_config_proto_rawDescGZIP() []byte {
	file_infra_chromeperf_workflows_server_templates_config_proto_rawDescOnce.Do(func() {
		file_infra_chromeperf_workflows_server_templates_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_chromeperf_workflows_server_templates_config_proto_rawDescData)
	})
	return file_infra_chromeperf_workflows_server_templates_config_proto_rawDescData
}

var file_infra_chromeperf_workflows_server_templates_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_chromeperf_workflows_server_templates_config_proto_goTypes = []interface{}{
	(*WorkflowTemplatesConfig)(nil),    // 0: WorkflowTemplatesConfig
	(*workflows.WorkflowTemplate)(nil), // 1: workflows.WorkflowTemplate
}
var file_infra_chromeperf_workflows_server_templates_config_proto_depIdxs = []int32{
	1, // 0: WorkflowTemplatesConfig.templates:type_name -> workflows.WorkflowTemplate
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_infra_chromeperf_workflows_server_templates_config_proto_init() }
func file_infra_chromeperf_workflows_server_templates_config_proto_init() {
	if File_infra_chromeperf_workflows_server_templates_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_chromeperf_workflows_server_templates_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowTemplatesConfig); i {
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
			RawDescriptor: file_infra_chromeperf_workflows_server_templates_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_chromeperf_workflows_server_templates_config_proto_goTypes,
		DependencyIndexes: file_infra_chromeperf_workflows_server_templates_config_proto_depIdxs,
		MessageInfos:      file_infra_chromeperf_workflows_server_templates_config_proto_msgTypes,
	}.Build()
	File_infra_chromeperf_workflows_server_templates_config_proto = out.File
	file_infra_chromeperf_workflows_server_templates_config_proto_rawDesc = nil
	file_infra_chromeperf_workflows_server_templates_config_proto_goTypes = nil
	file_infra_chromeperf_workflows_server_templates_config_proto_depIdxs = nil
}
