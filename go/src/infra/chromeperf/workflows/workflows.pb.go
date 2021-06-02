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
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.0
// source: infra/chromeperf/workflows/workflows.proto

package workflows

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
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

// The state of the workflow.
type Workflow_State int32

const (
	Workflow_STATE_UNSPECIFIED Workflow_State = 0
	Workflow_PENDING           Workflow_State = 1
	Workflow_ONGOING           Workflow_State = 2
	Workflow_COMPLETED         Workflow_State = 3
	Workflow_CANCELLED         Workflow_State = 4
)

// Enum value maps for Workflow_State.
var (
	Workflow_State_name = map[int32]string{
		0: "STATE_UNSPECIFIED",
		1: "PENDING",
		2: "ONGOING",
		3: "COMPLETED",
		4: "CANCELLED",
	}
	Workflow_State_value = map[string]int32{
		"STATE_UNSPECIFIED": 0,
		"PENDING":           1,
		"ONGOING":           2,
		"COMPLETED":         3,
		"CANCELLED":         4,
	}
)

func (x Workflow_State) Enum() *Workflow_State {
	p := new(Workflow_State)
	*p = x
	return p
}

func (x Workflow_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Workflow_State) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_chromeperf_workflows_workflows_proto_enumTypes[0].Descriptor()
}

func (Workflow_State) Type() protoreflect.EnumType {
	return &file_infra_chromeperf_workflows_workflows_proto_enumTypes[0]
}

func (x Workflow_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Workflow_State.Descriptor instead.
func (Workflow_State) EnumDescriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_workflows_proto_rawDescGZIP(), []int{0, 0}
}

type Task_State int32

const (
	Task_STATE_UNSPECIFIED Task_State = 0
	Task_PENDING           Task_State = 1
	Task_ONGOING           Task_State = 2
	Task_FAILED            Task_State = 3
	Task_COMPLETED         Task_State = 4
	Task_CANCELLED         Task_State = 5
)

// Enum value maps for Task_State.
var (
	Task_State_name = map[int32]string{
		0: "STATE_UNSPECIFIED",
		1: "PENDING",
		2: "ONGOING",
		3: "FAILED",
		4: "COMPLETED",
		5: "CANCELLED",
	}
	Task_State_value = map[string]int32{
		"STATE_UNSPECIFIED": 0,
		"PENDING":           1,
		"ONGOING":           2,
		"FAILED":            3,
		"COMPLETED":         4,
		"CANCELLED":         5,
	}
)

func (x Task_State) Enum() *Task_State {
	p := new(Task_State)
	*p = x
	return p
}

func (x Task_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Task_State) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_chromeperf_workflows_workflows_proto_enumTypes[1].Descriptor()
}

func (Task_State) Type() protoreflect.EnumType {
	return &file_infra_chromeperf_workflows_workflows_proto_enumTypes[1]
}

func (x Task_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Task_State.Descriptor instead.
func (Task_State) EnumDescriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_workflows_proto_rawDescGZIP(), []int{1, 0}
}

type Workflow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The resource name of the workflow.
	//
	// NOTE: This is system-generated and ignored when provided in a
	// CreateWorkflowRequest.
	//
	// Format: workflows/{name}
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The template name to use to seed the workflow graph.
	//
	// Must be of the form /workflow-templates/<name>.
	Template string `protobuf:"bytes,2,opt,name=template,proto3" json:"template,omitempty"`
	// A mapping between a key and values. This is provided to the template when
	// seeding the workflow graph. Validation for the inputs is determined by
	// the field descriptors provided by the workflow template definition.
	Inputs map[string]*structpb.Value `protobuf:"bytes,3,rep,name=inputs,proto3" json:"inputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	State  Workflow_State             `protobuf:"varint,4,opt,name=state,proto3,enum=workflows.Workflow_State" json:"state,omitempty"`
	// The creation timestamp for the workflow.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// The most recent update time for the
	LastUpdateTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	// Each task in the workflow.
	Tasks []*Task `protobuf:"bytes,7,rep,name=tasks,proto3" json:"tasks,omitempty"`
}

func (x *Workflow) Reset() {
	*x = Workflow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_chromeperf_workflows_workflows_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Workflow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Workflow) ProtoMessage() {}

func (x *Workflow) ProtoReflect() protoreflect.Message {
	mi := &file_infra_chromeperf_workflows_workflows_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Workflow.ProtoReflect.Descriptor instead.
func (*Workflow) Descriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_workflows_proto_rawDescGZIP(), []int{0}
}

func (x *Workflow) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Workflow) GetTemplate() string {
	if x != nil {
		return x.Template
	}
	return ""
}

func (x *Workflow) GetInputs() map[string]*structpb.Value {
	if x != nil {
		return x.Inputs
	}
	return nil
}

func (x *Workflow) GetState() Workflow_State {
	if x != nil {
		return x.State
	}
	return Workflow_STATE_UNSPECIFIED
}

func (x *Workflow) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Workflow) GetLastUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastUpdateTime
	}
	return nil
}

func (x *Workflow) GetTasks() []*Task {
	if x != nil {
		return x.Tasks
	}
	return nil
}

type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name for the task.
	//
	// Format: workflow/{workflow}/task/{task}
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The state of this particular task.
	State Task_State `protobuf:"varint,2,opt,name=state,proto3,enum=workflows.Task_State" json:"state,omitempty"`
	// The creation timestamp for this task.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// The timestamp for the most recent update on the task.
	LastUpdateTime *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	// This is the structured input provided to the Task at creation time.
	Input *structpb.Struct `protobuf:"bytes,5,opt,name=input,proto3" json:"input,omitempty"`
	// This is the structured output at last update time.
	Output *structpb.Struct `protobuf:"bytes,6,opt,name=output,proto3" json:"output,omitempty"`
	// This is the list of Task names which this Task is dependent on.
	Dependencies []string `protobuf:"bytes,7,rep,name=dependencies,proto3" json:"dependencies,omitempty"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_chromeperf_workflows_workflows_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_infra_chromeperf_workflows_workflows_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_workflows_proto_rawDescGZIP(), []int{1}
}

func (x *Task) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Task) GetState() Task_State {
	if x != nil {
		return x.State
	}
	return Task_STATE_UNSPECIFIED
}

func (x *Task) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Task) GetLastUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastUpdateTime
	}
	return nil
}

func (x *Task) GetInput() *structpb.Struct {
	if x != nil {
		return x.Input
	}
	return nil
}

func (x *Task) GetOutput() *structpb.Struct {
	if x != nil {
		return x.Output
	}
	return nil
}

func (x *Task) GetDependencies() []string {
	if x != nil {
		return x.Dependencies
	}
	return nil
}

var File_infra_chromeperf_workflows_workflows_proto protoreflect.FileDescriptor

var file_infra_chromeperf_workflows_workflows_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65,
	0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69,
	0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xd9, 0x04, 0x0a, 0x08, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12,
	0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0,
	0x41, 0x03, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x08, 0x74, 0x65, 0x6d, 0x70,
	0x6c, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x08, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x3c, 0x0a, 0x06, 0x69, 0x6e, 0x70,
	0x75, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x12, 0x34, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x73, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x40, 0x0a,
	0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03,
	0xe0, 0x41, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x49, 0x0a, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x2a, 0x0a, 0x05, 0x74, 0x61,
	0x73, 0x6b, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52,
	0x05, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x1a, 0x51, 0x0a, 0x0b, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x56, 0x0a, 0x05, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x45, 0x4e,
	0x44, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x4f, 0x4e, 0x47, 0x4f, 0x49, 0x4e,
	0x47, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45, 0x44,
	0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x4c, 0x45, 0x44, 0x10,
	0x04, 0x3a, 0x3b, 0xea, 0x41, 0x38, 0x0a, 0x20, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x70,
	0x65, 0x72, 0x66, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x72, 0x2e, 0x64, 0x65, 0x76, 0x2f,
	0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x14, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x73, 0x2f, 0x7b, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x7d, 0x22, 0xbc,
	0x04, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x15, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x54, 0x61, 0x73, 0x6b,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x40, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x49, 0x0a, 0x10, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52,
	0x0e, 0x6c, 0x61, 0x73, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x32, 0x0a, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x05, 0x69, 0x6e,
	0x70, 0x75, 0x74, 0x12, 0x34, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x42, 0x03, 0xe0, 0x41,
	0x03, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x48, 0x0a, 0x0c, 0x64, 0x65, 0x70,
	0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x42,
	0x24, 0xe0, 0x41, 0x03, 0xfa, 0x41, 0x1e, 0x0a, 0x1c, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e,
	0x70, 0x65, 0x72, 0x66, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x72, 0x2e, 0x64, 0x65, 0x76,
	0x2f, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63,
	0x69, 0x65, 0x73, 0x22, 0x62, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x15, 0x0a, 0x11,
	0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45,
	0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x01,
	0x12, 0x0b, 0x0a, 0x07, 0x4f, 0x4e, 0x47, 0x4f, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x0a, 0x0a,
	0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d,
	0x50, 0x4c, 0x45, 0x54, 0x45, 0x44, 0x10, 0x04, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x41, 0x4e, 0x43,
	0x45, 0x4c, 0x4c, 0x45, 0x44, 0x10, 0x05, 0x3a, 0x44, 0xea, 0x41, 0x41, 0x0a, 0x1c, 0x65, 0x6e,
	0x67, 0x69, 0x6e, 0x65, 0x2e, 0x70, 0x65, 0x72, 0x66, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63,
	0x72, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x21, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x7b, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x7d,
	0x2f, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x7d, 0x42, 0x1c, 0x5a,
	0x1a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65, 0x72,
	0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_infra_chromeperf_workflows_workflows_proto_rawDescOnce sync.Once
	file_infra_chromeperf_workflows_workflows_proto_rawDescData = file_infra_chromeperf_workflows_workflows_proto_rawDesc
)

func file_infra_chromeperf_workflows_workflows_proto_rawDescGZIP() []byte {
	file_infra_chromeperf_workflows_workflows_proto_rawDescOnce.Do(func() {
		file_infra_chromeperf_workflows_workflows_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_chromeperf_workflows_workflows_proto_rawDescData)
	})
	return file_infra_chromeperf_workflows_workflows_proto_rawDescData
}

var file_infra_chromeperf_workflows_workflows_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_infra_chromeperf_workflows_workflows_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_infra_chromeperf_workflows_workflows_proto_goTypes = []interface{}{
	(Workflow_State)(0),           // 0: workflows.Workflow.State
	(Task_State)(0),               // 1: workflows.Task.State
	(*Workflow)(nil),              // 2: workflows.Workflow
	(*Task)(nil),                  // 3: workflows.Task
	nil,                           // 4: workflows.Workflow.InputsEntry
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
	(*structpb.Struct)(nil),       // 6: google.protobuf.Struct
	(*structpb.Value)(nil),        // 7: google.protobuf.Value
}
var file_infra_chromeperf_workflows_workflows_proto_depIdxs = []int32{
	4,  // 0: workflows.Workflow.inputs:type_name -> workflows.Workflow.InputsEntry
	0,  // 1: workflows.Workflow.state:type_name -> workflows.Workflow.State
	5,  // 2: workflows.Workflow.create_time:type_name -> google.protobuf.Timestamp
	5,  // 3: workflows.Workflow.last_update_time:type_name -> google.protobuf.Timestamp
	3,  // 4: workflows.Workflow.tasks:type_name -> workflows.Task
	1,  // 5: workflows.Task.state:type_name -> workflows.Task.State
	5,  // 6: workflows.Task.create_time:type_name -> google.protobuf.Timestamp
	5,  // 7: workflows.Task.last_update_time:type_name -> google.protobuf.Timestamp
	6,  // 8: workflows.Task.input:type_name -> google.protobuf.Struct
	6,  // 9: workflows.Task.output:type_name -> google.protobuf.Struct
	7,  // 10: workflows.Workflow.InputsEntry.value:type_name -> google.protobuf.Value
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_infra_chromeperf_workflows_workflows_proto_init() }
func file_infra_chromeperf_workflows_workflows_proto_init() {
	if File_infra_chromeperf_workflows_workflows_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_chromeperf_workflows_workflows_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Workflow); i {
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
		file_infra_chromeperf_workflows_workflows_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
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
			RawDescriptor: file_infra_chromeperf_workflows_workflows_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_chromeperf_workflows_workflows_proto_goTypes,
		DependencyIndexes: file_infra_chromeperf_workflows_workflows_proto_depIdxs,
		EnumInfos:         file_infra_chromeperf_workflows_workflows_proto_enumTypes,
		MessageInfos:      file_infra_chromeperf_workflows_workflows_proto_msgTypes,
	}.Build()
	File_infra_chromeperf_workflows_workflows_proto = out.File
	file_infra_chromeperf_workflows_workflows_proto_rawDesc = nil
	file_infra_chromeperf_workflows_workflows_proto_goTypes = nil
	file_infra_chromeperf_workflows_workflows_proto_depIdxs = nil
}
