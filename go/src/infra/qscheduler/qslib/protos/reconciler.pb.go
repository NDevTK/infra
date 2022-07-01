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
// source: infra/qscheduler/qslib/protos/reconciler.proto

package protos

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// WorkerQueue represents a task request that is pending assignment to a given
// worker and optionally the expected task on the worker to preempt.
//
// Note: the name WorkerQueue is a legacy name, which is why it isn't a great
// match for what it represents.
type WorkerQueue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// EnqueueTime is the time at which the pending assignment was created
	// by the scheduler.
	EnqueueTime *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=enqueue_time,json=enqueueTime,proto3" json:"enqueue_time,omitempty"`
	// TaskToAssign is the id of the task that should be assigned to this worker.
	TaskToAssign string `protobuf:"bytes,2,opt,name=task_to_assign,json=taskToAssign,proto3" json:"task_to_assign,omitempty"`
	// TaskToAbort is the id of the task that should be aborted on this worker.
	//
	// An empty string indicates that there is no task to abort.
	TaskToAbort string `protobuf:"bytes,3,opt,name=task_to_abort,json=taskToAbort,proto3" json:"task_to_abort,omitempty"`
}

func (x *WorkerQueue) Reset() {
	*x = WorkerQueue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkerQueue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkerQueue) ProtoMessage() {}

func (x *WorkerQueue) ProtoReflect() protoreflect.Message {
	mi := &file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkerQueue.ProtoReflect.Descriptor instead.
func (*WorkerQueue) Descriptor() ([]byte, []int) {
	return file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescGZIP(), []int{0}
}

func (x *WorkerQueue) GetEnqueueTime() *timestamppb.Timestamp {
	if x != nil {
		return x.EnqueueTime
	}
	return nil
}

func (x *WorkerQueue) GetTaskToAssign() string {
	if x != nil {
		return x.TaskToAssign
	}
	return ""
}

func (x *WorkerQueue) GetTaskToAbort() string {
	if x != nil {
		return x.TaskToAbort
	}
	return ""
}

// ReconcilerState represents a reconciler. It holds tasks that are pending
// assignment to workers and tasks that have errored out.
type Reconciler struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// WorkerQueues holds pending assignments for workers.
	//
	// An assignment remains pending until a notification from Swarming
	// acknowledges that it has taken place.
	WorkerQueues map[string]*WorkerQueue `protobuf:"bytes,1,rep,name=worker_queues,json=workerQueues,proto3" json:"worker_queues,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// TaskErrors is a map from task ids that had an error to the error description.
	//
	// Task errors remain pending until a notification from Swarming acknowledges
	// that the task is no longer pending.
	TaskErrors map[string]string `protobuf:"bytes,2,rep,name=task_errors,json=taskErrors,proto3" json:"task_errors,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Reconciler) Reset() {
	*x = Reconciler{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reconciler) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reconciler) ProtoMessage() {}

func (x *Reconciler) ProtoReflect() protoreflect.Message {
	mi := &file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reconciler.ProtoReflect.Descriptor instead.
func (*Reconciler) Descriptor() ([]byte, []int) {
	return file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescGZIP(), []int{1}
}

func (x *Reconciler) GetWorkerQueues() map[string]*WorkerQueue {
	if x != nil {
		return x.WorkerQueues
	}
	return nil
}

func (x *Reconciler) GetTaskErrors() map[string]string {
	if x != nil {
		return x.TaskErrors
	}
	return nil
}

var File_infra_qscheduler_qslib_protos_reconciler_proto protoreflect.FileDescriptor

var file_infra_qscheduler_qslib_protos_reconciler_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x71, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x65, 0x72, 0x2f, 0x71, 0x73, 0x6c, 0x69, 0x62, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f,
	0x72, 0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x96, 0x01, 0x0a, 0x0b, 0x57, 0x6f,
	0x72, 0x6b, 0x65, 0x72, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x65, 0x6e, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x65, 0x6e, 0x71,
	0x75, 0x65, 0x75, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x61, 0x73, 0x6b,
	0x5f, 0x74, 0x6f, 0x5f, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x74, 0x61, 0x73, 0x6b, 0x54, 0x6f, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x12, 0x22,
	0x0a, 0x0d, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x74, 0x6f, 0x5f, 0x61, 0x62, 0x6f, 0x72, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x74, 0x61, 0x73, 0x6b, 0x54, 0x6f, 0x41, 0x62, 0x6f,
	0x72, 0x74, 0x22, 0xb1, 0x02, 0x0a, 0x0a, 0x52, 0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65,
	0x72, 0x12, 0x49, 0x0a, 0x0d, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x2e, 0x57, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x51, 0x75, 0x65, 0x75, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0c,
	0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x51, 0x75, 0x65, 0x75, 0x65, 0x73, 0x12, 0x43, 0x0a, 0x0b,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x22, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6e,
	0x63, 0x69, 0x6c, 0x65, 0x72, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x74, 0x61, 0x73, 0x6b, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x73, 0x1a, 0x54, 0x0a, 0x11, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x51, 0x75, 0x65, 0x75, 0x65,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x29, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x51, 0x75, 0x65, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3d, 0x0a, 0x0f, 0x54, 0x61, 0x73, 0x6b, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x1f, 0x5a, 0x1d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f,
	0x71, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2f, 0x71, 0x73, 0x6c, 0x69, 0x62,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescOnce sync.Once
	file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescData = file_infra_qscheduler_qslib_protos_reconciler_proto_rawDesc
)

func file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescGZIP() []byte {
	file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescOnce.Do(func() {
		file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescData)
	})
	return file_infra_qscheduler_qslib_protos_reconciler_proto_rawDescData
}

var file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_infra_qscheduler_qslib_protos_reconciler_proto_goTypes = []interface{}{
	(*WorkerQueue)(nil),           // 0: protos.WorkerQueue
	(*Reconciler)(nil),            // 1: protos.Reconciler
	nil,                           // 2: protos.Reconciler.WorkerQueuesEntry
	nil,                           // 3: protos.Reconciler.TaskErrorsEntry
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_infra_qscheduler_qslib_protos_reconciler_proto_depIdxs = []int32{
	4, // 0: protos.WorkerQueue.enqueue_time:type_name -> google.protobuf.Timestamp
	2, // 1: protos.Reconciler.worker_queues:type_name -> protos.Reconciler.WorkerQueuesEntry
	3, // 2: protos.Reconciler.task_errors:type_name -> protos.Reconciler.TaskErrorsEntry
	0, // 3: protos.Reconciler.WorkerQueuesEntry.value:type_name -> protos.WorkerQueue
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_infra_qscheduler_qslib_protos_reconciler_proto_init() }
func file_infra_qscheduler_qslib_protos_reconciler_proto_init() {
	if File_infra_qscheduler_qslib_protos_reconciler_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkerQueue); i {
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
		file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reconciler); i {
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
			RawDescriptor: file_infra_qscheduler_qslib_protos_reconciler_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_qscheduler_qslib_protos_reconciler_proto_goTypes,
		DependencyIndexes: file_infra_qscheduler_qslib_protos_reconciler_proto_depIdxs,
		MessageInfos:      file_infra_qscheduler_qslib_protos_reconciler_proto_msgTypes,
	}.Build()
	File_infra_qscheduler_qslib_protos_reconciler_proto = out.File
	file_infra_qscheduler_qslib_protos_reconciler_proto_rawDesc = nil
	file_infra_qscheduler_qslib_protos_reconciler_proto_goTypes = nil
	file_infra_qscheduler_qslib_protos_reconciler_proto_depIdxs = nil
}
