// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/qscheduler/qslib/scheduler/state.proto

package scheduler

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// StateProto represents the overall state of a quota scheduler worker pool,
// account set, and task queue. This is represented separately from
// configuration information. The state is expected to be updated frequently,
// on each scheduler tick.
type StateProto struct {
	// QueuedRequests is the set of Requests that are waiting to be assigned to a
	// worker, keyed by request id.
	QueuedRequests map[string]*TaskRequestProto `protobuf:"bytes,1,rep,name=queued_requests,json=queuedRequests,proto3" json:"queued_requests,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Balance of all quota accounts for this pool, keyed by account id.
	Balances map[string]*StateProto_Balance `protobuf:"bytes,2,rep,name=balances,proto3" json:"balances,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Workers that may run tasks, and their states, keyed by worker id.
	Workers map[string]*Worker `protobuf:"bytes,3,rep,name=workers,proto3" json:"workers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// LastUpdateTime is the last time at which UpdateTime was called on a scheduler,
	// and corresponds to the when the quota account balances were updated.
	LastUpdateTime *timestamp.Timestamp `protobuf:"bytes,4,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	// LabelMap maps label IDs to their string values.
	//
	// Requests and workers store labels by IDs.
	LabelMap             map[uint64]string `protobuf:"bytes,5,rep,name=label_map,json=labelMap,proto3" json:"label_map,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *StateProto) Reset()         { *m = StateProto{} }
func (m *StateProto) String() string { return proto.CompactTextString(m) }
func (*StateProto) ProtoMessage()    {}
func (*StateProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_d6b2685915185dd1, []int{0}
}

func (m *StateProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateProto.Unmarshal(m, b)
}
func (m *StateProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateProto.Marshal(b, m, deterministic)
}
func (m *StateProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateProto.Merge(m, src)
}
func (m *StateProto) XXX_Size() int {
	return xxx_messageInfo_StateProto.Size(m)
}
func (m *StateProto) XXX_DiscardUnknown() {
	xxx_messageInfo_StateProto.DiscardUnknown(m)
}

var xxx_messageInfo_StateProto proto.InternalMessageInfo

func (m *StateProto) GetQueuedRequests() map[string]*TaskRequestProto {
	if m != nil {
		return m.QueuedRequests
	}
	return nil
}

func (m *StateProto) GetBalances() map[string]*StateProto_Balance {
	if m != nil {
		return m.Balances
	}
	return nil
}

func (m *StateProto) GetWorkers() map[string]*Worker {
	if m != nil {
		return m.Workers
	}
	return nil
}

func (m *StateProto) GetLastUpdateTime() *timestamp.Timestamp {
	if m != nil {
		return m.LastUpdateTime
	}
	return nil
}

func (m *StateProto) GetLabelMap() map[uint64]string {
	if m != nil {
		return m.LabelMap
	}
	return nil
}

type StateProto_Balance struct {
	Value                []float64 `protobuf:"fixed64,1,rep,packed,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *StateProto_Balance) Reset()         { *m = StateProto_Balance{} }
func (m *StateProto_Balance) String() string { return proto.CompactTextString(m) }
func (*StateProto_Balance) ProtoMessage()    {}
func (*StateProto_Balance) Descriptor() ([]byte, []int) {
	return fileDescriptor_d6b2685915185dd1, []int{0, 1}
}

func (m *StateProto_Balance) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateProto_Balance.Unmarshal(m, b)
}
func (m *StateProto_Balance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateProto_Balance.Marshal(b, m, deterministic)
}
func (m *StateProto_Balance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateProto_Balance.Merge(m, src)
}
func (m *StateProto_Balance) XXX_Size() int {
	return xxx_messageInfo_StateProto_Balance.Size(m)
}
func (m *StateProto_Balance) XXX_DiscardUnknown() {
	xxx_messageInfo_StateProto_Balance.DiscardUnknown(m)
}

var xxx_messageInfo_StateProto_Balance proto.InternalMessageInfo

func (m *StateProto_Balance) GetValue() []float64 {
	if m != nil {
		return m.Value
	}
	return nil
}

// TaskRequestProto represents a requested task in the queue, and refers to the
// quota account to run it against. This representation intentionally
// excludes most of the details of a Swarming task request.
type TaskRequestProto struct {
	// AccountId is the id of the account that this request charges to.
	AccountId string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	// EnqueueTime is the time at which the request was enqueued.
	EnqueueTime *timestamp.Timestamp `protobuf:"bytes,2,opt,name=enqueue_time,json=enqueueTime,proto3" json:"enqueue_time,omitempty"`
	// ConfirmedTime is the most recent time at which the Request state was
	// provided or confirmed by external authority (via a call to Enforce or
	// AddRequest).
	ConfirmedTime *timestamp.Timestamp `protobuf:"bytes,4,opt,name=confirmed_time,json=confirmedTime,proto3" json:"confirmed_time,omitempty"`
	// ProvisionableLabelIds represents the task's provisionable labels.
	ProvisionableLabelIds []uint64 `protobuf:"varint,6,rep,packed,name=provisionable_label_ids,json=provisionableLabelIds,proto3" json:"provisionable_label_ids,omitempty"`
	// BaseLabelIds represents the task's base labels.
	BaseLabelIds         []uint64 `protobuf:"varint,7,rep,packed,name=base_label_ids,json=baseLabelIds,proto3" json:"base_label_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TaskRequestProto) Reset()         { *m = TaskRequestProto{} }
func (m *TaskRequestProto) String() string { return proto.CompactTextString(m) }
func (*TaskRequestProto) ProtoMessage()    {}
func (*TaskRequestProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_d6b2685915185dd1, []int{1}
}

func (m *TaskRequestProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TaskRequestProto.Unmarshal(m, b)
}
func (m *TaskRequestProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TaskRequestProto.Marshal(b, m, deterministic)
}
func (m *TaskRequestProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskRequestProto.Merge(m, src)
}
func (m *TaskRequestProto) XXX_Size() int {
	return xxx_messageInfo_TaskRequestProto.Size(m)
}
func (m *TaskRequestProto) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskRequestProto.DiscardUnknown(m)
}

var xxx_messageInfo_TaskRequestProto proto.InternalMessageInfo

func (m *TaskRequestProto) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *TaskRequestProto) GetEnqueueTime() *timestamp.Timestamp {
	if m != nil {
		return m.EnqueueTime
	}
	return nil
}

func (m *TaskRequestProto) GetConfirmedTime() *timestamp.Timestamp {
	if m != nil {
		return m.ConfirmedTime
	}
	return nil
}

func (m *TaskRequestProto) GetProvisionableLabelIds() []uint64 {
	if m != nil {
		return m.ProvisionableLabelIds
	}
	return nil
}

func (m *TaskRequestProto) GetBaseLabelIds() []uint64 {
	if m != nil {
		return m.BaseLabelIds
	}
	return nil
}

// TaskRun represents a task that has been assigned to a worker and is
// now running.
type TaskRun struct {
	// Cost is the total cost that has been incurred on this task while running.
	Cost []float64 `protobuf:"fixed64,1,rep,packed,name=cost,proto3" json:"cost,omitempty"`
	// Request is the request that this running task corresponds to.
	Request *TaskRequestProto `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	// RequestId is the request id of the request that this running task
	// corresponds to.
	RequestId string `protobuf:"bytes,3,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Priority is the current priority level of the running task.
	Priority             int32    `protobuf:"varint,4,opt,name=priority,proto3" json:"priority,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TaskRun) Reset()         { *m = TaskRun{} }
func (m *TaskRun) String() string { return proto.CompactTextString(m) }
func (*TaskRun) ProtoMessage()    {}
func (*TaskRun) Descriptor() ([]byte, []int) {
	return fileDescriptor_d6b2685915185dd1, []int{2}
}

func (m *TaskRun) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TaskRun.Unmarshal(m, b)
}
func (m *TaskRun) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TaskRun.Marshal(b, m, deterministic)
}
func (m *TaskRun) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskRun.Merge(m, src)
}
func (m *TaskRun) XXX_Size() int {
	return xxx_messageInfo_TaskRun.Size(m)
}
func (m *TaskRun) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskRun.DiscardUnknown(m)
}

var xxx_messageInfo_TaskRun proto.InternalMessageInfo

func (m *TaskRun) GetCost() []float64 {
	if m != nil {
		return m.Cost
	}
	return nil
}

func (m *TaskRun) GetRequest() *TaskRequestProto {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *TaskRun) GetRequestId() string {
	if m != nil {
		return m.RequestId
	}
	return ""
}

func (m *TaskRun) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

// Worker represents a resource that can run 1 task at a time. This corresponds
// to the swarming concept of a Bot. This representation considers only the
// subset of Labels that are Provisionable (can be changed by running a task),
// because the quota scheduler algorithm is expected to run against a pool of
// otherwise homogenous workers.
type Worker struct {
	// RunningTask is, if non-nil, the task that is currently running on the
	// worker.
	RunningTask *TaskRun `protobuf:"bytes,2,opt,name=running_task,json=runningTask,proto3" json:"running_task,omitempty"`
	// ConfirmedTime is the most recent time at which the Worker state was
	// directly confirmed as idle by external authority (via a call to MarkIdle or
	// NotifyRequest).
	ConfirmedTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=confirmed_time,json=confirmedTime,proto3" json:"confirmed_time,omitempty"`
	// LabelIds represents the worker's labels.
	LabelIds             []uint64 `protobuf:"varint,4,rep,packed,name=label_ids,json=labelIds,proto3" json:"label_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Worker) Reset()         { *m = Worker{} }
func (m *Worker) String() string { return proto.CompactTextString(m) }
func (*Worker) ProtoMessage()    {}
func (*Worker) Descriptor() ([]byte, []int) {
	return fileDescriptor_d6b2685915185dd1, []int{3}
}

func (m *Worker) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Worker.Unmarshal(m, b)
}
func (m *Worker) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Worker.Marshal(b, m, deterministic)
}
func (m *Worker) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Worker.Merge(m, src)
}
func (m *Worker) XXX_Size() int {
	return xxx_messageInfo_Worker.Size(m)
}
func (m *Worker) XXX_DiscardUnknown() {
	xxx_messageInfo_Worker.DiscardUnknown(m)
}

var xxx_messageInfo_Worker proto.InternalMessageInfo

func (m *Worker) GetRunningTask() *TaskRun {
	if m != nil {
		return m.RunningTask
	}
	return nil
}

func (m *Worker) GetConfirmedTime() *timestamp.Timestamp {
	if m != nil {
		return m.ConfirmedTime
	}
	return nil
}

func (m *Worker) GetLabelIds() []uint64 {
	if m != nil {
		return m.LabelIds
	}
	return nil
}

func init() {
	proto.RegisterType((*StateProto)(nil), "scheduler.StateProto")
	proto.RegisterMapType((map[string]*StateProto_Balance)(nil), "scheduler.StateProto.BalancesEntry")
	proto.RegisterMapType((map[uint64]string)(nil), "scheduler.StateProto.LabelMapEntry")
	proto.RegisterMapType((map[string]*TaskRequestProto)(nil), "scheduler.StateProto.QueuedRequestsEntry")
	proto.RegisterMapType((map[string]*Worker)(nil), "scheduler.StateProto.WorkersEntry")
	proto.RegisterType((*StateProto_Balance)(nil), "scheduler.StateProto.Balance")
	proto.RegisterType((*TaskRequestProto)(nil), "scheduler.TaskRequestProto")
	proto.RegisterType((*TaskRun)(nil), "scheduler.TaskRun")
	proto.RegisterType((*Worker)(nil), "scheduler.Worker")
}

func init() {
	proto.RegisterFile("infra/qscheduler/qslib/scheduler/state.proto", fileDescriptor_d6b2685915185dd1)
}

var fileDescriptor_d6b2685915185dd1 = []byte{
	// 615 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xdb, 0x6a, 0xd4, 0x40,
	0x18, 0xc7, 0xc9, 0x26, 0x7b, 0xfa, 0x76, 0xbb, 0xae, 0x63, 0xc5, 0xb0, 0xa5, 0x74, 0x59, 0x05,
	0x57, 0x90, 0x2c, 0xb6, 0x54, 0xc4, 0x03, 0x1e, 0xd0, 0x8b, 0x16, 0x0b, 0x1a, 0x2b, 0x82, 0x17,
	0x86, 0xc9, 0x66, 0x5a, 0xc3, 0x66, 0x27, 0xd9, 0x99, 0x49, 0xa5, 0x4f, 0xe1, 0x13, 0x78, 0xe9,
	0xfb, 0xf8, 0x48, 0x32, 0x87, 0x6c, 0x93, 0x12, 0x6a, 0xbd, 0x9b, 0xf9, 0xf2, 0xff, 0xff, 0xbf,
	0xc9, 0x6f, 0xbe, 0x81, 0x87, 0x31, 0x3d, 0x61, 0x78, 0xb6, 0xe2, 0xf3, 0xef, 0x24, 0xca, 0x13,
	0xc2, 0x66, 0x2b, 0x9e, 0xc4, 0xe1, 0xec, 0x62, 0xcf, 0x05, 0x16, 0xc4, 0xcb, 0x58, 0x2a, 0x52,
	0xd4, 0x5d, 0x97, 0x47, 0x3b, 0xa7, 0x69, 0x7a, 0x9a, 0x90, 0x99, 0xfa, 0x10, 0xe6, 0x27, 0x33,
	0x11, 0x2f, 0x09, 0x17, 0x78, 0x99, 0x69, 0xed, 0xe4, 0x4f, 0x13, 0xe0, 0x93, 0xf4, 0x7e, 0x50,
	0x56, 0x1f, 0x6e, 0xac, 0x72, 0x92, 0x93, 0x28, 0x60, 0x64, 0x95, 0x13, 0x2e, 0xb8, 0x6b, 0x8d,
	0xed, 0x69, 0x6f, 0xf7, 0x81, 0xb7, 0x0e, 0xf5, 0x2e, 0xf4, 0xde, 0x47, 0x25, 0xf6, 0x8d, 0xf6,
	0x1d, 0x15, 0xec, 0xdc, 0x1f, 0xac, 0x2a, 0x45, 0xf4, 0x12, 0x3a, 0x21, 0x4e, 0x30, 0x9d, 0x13,
	0xee, 0x36, 0x54, 0xd8, 0xdd, 0xfa, 0xb0, 0x37, 0x46, 0xa5, 0x63, 0xd6, 0x26, 0xf4, 0x1c, 0xda,
	0x3f, 0x52, 0xb6, 0x20, 0x8c, 0xbb, 0xb6, 0xf2, 0x4f, 0xea, 0xfd, 0x5f, 0xb4, 0x48, 0xdb, 0x0b,
	0x0b, 0x7a, 0x0b, 0xc3, 0x04, 0x73, 0x11, 0xe4, 0x59, 0x84, 0x05, 0x09, 0x24, 0x00, 0xd7, 0x19,
	0x5b, 0xd3, 0xde, 0xee, 0xc8, 0xd3, 0x74, 0xbc, 0x82, 0x8e, 0x77, 0x5c, 0xd0, 0xf1, 0x07, 0xd2,
	0xf3, 0x59, 0x59, 0x64, 0x11, 0xbd, 0x82, 0x6e, 0x82, 0x43, 0x92, 0x04, 0x4b, 0x9c, 0xb9, 0xcd,
	0xab, 0xfe, 0xe2, 0xbd, 0x94, 0x1d, 0xe1, 0xcc, 0xfc, 0x45, 0x62, 0xb6, 0xa3, 0x6f, 0x70, 0xab,
	0x86, 0x16, 0x1a, 0x82, 0xbd, 0x20, 0xe7, 0xae, 0x35, 0xb6, 0xa6, 0x5d, 0x5f, 0x2e, 0xd1, 0x23,
	0x68, 0x9e, 0xe1, 0x24, 0x27, 0x6e, 0x43, 0x9d, 0x72, 0xab, 0xd4, 0xe6, 0x18, 0xf3, 0x85, 0xb1,
	0xab, 0x66, 0xbe, 0x56, 0x3e, 0x6d, 0x3c, 0xb1, 0x46, 0x3b, 0xd0, 0x36, 0x00, 0xd1, 0x66, 0x91,
	0x20, 0xef, 0xce, 0x32, 0xa2, 0xd1, 0x57, 0xd8, 0xa8, 0x10, 0xae, 0x69, 0xbd, 0x57, 0x6d, 0xbd,
	0x7d, 0xe5, 0x3d, 0x95, 0x9b, 0x1f, 0x41, 0xbf, 0x4c, 0xbf, 0x26, 0xfa, 0x7e, 0x35, 0xfa, 0x66,
	0x29, 0x5a, 0x3b, 0xcb, 0x71, 0xcf, 0x60, 0xa3, 0x82, 0xb1, 0x9c, 0xe7, 0xe8, 0xbc, 0xcd, 0x72,
	0x5e, 0xb7, 0x64, 0x9e, 0xfc, 0x6a, 0xc0, 0xf0, 0x32, 0x28, 0xb4, 0x0d, 0x80, 0xe7, 0xf3, 0x34,
	0xa7, 0x22, 0x88, 0x23, 0x73, 0xae, 0xae, 0xa9, 0x1c, 0x44, 0xe8, 0x05, 0xf4, 0x09, 0x55, 0x73,
	0xab, 0x07, 0xa4, 0xf1, 0xcf, 0x01, 0xe9, 0x19, 0xbd, 0x9a, 0x8e, 0xd7, 0x30, 0x98, 0xa7, 0xf4,
	0x24, 0x66, 0x4b, 0x12, 0x5d, 0x77, 0xc2, 0x36, 0xd6, 0x0e, 0x15, 0xf1, 0x18, 0xee, 0x64, 0x2c,
	0x3d, 0x8b, 0x79, 0x9c, 0x52, 0x1c, 0x26, 0x24, 0xd0, 0xe3, 0x16, 0x47, 0xdc, 0x6d, 0x8d, 0xed,
	0xa9, 0xe3, 0xdf, 0xae, 0x7c, 0x56, 0x78, 0x0e, 0x22, 0x8e, 0xee, 0xc1, 0x20, 0xc4, 0xbc, 0x2c,
	0x6f, 0x2b, 0x79, 0x5f, 0x56, 0x0b, 0xd5, 0xa1, 0xd3, 0xb1, 0x87, 0xce, 0xa1, 0xd3, 0x69, 0x0e,
	0x5b, 0x93, 0x9f, 0x16, 0xb4, 0x15, 0x9f, 0x9c, 0x22, 0x04, 0xce, 0x3c, 0xe5, 0xc2, 0x0c, 0x8a,
	0x5a, 0xa3, 0x7d, 0x68, 0x9b, 0xc7, 0x7f, 0x9d, 0x09, 0x2c, 0xb4, 0x92, 0xb0, 0x59, 0x4a, 0xc2,
	0xb6, 0x26, 0x6c, 0x2a, 0x07, 0x11, 0x1a, 0x41, 0x27, 0x63, 0x71, 0xca, 0x62, 0x71, 0xae, 0xe0,
	0x34, 0xfd, 0xf5, 0x7e, 0xf2, 0xdb, 0x82, 0x96, 0x1e, 0x02, 0xb4, 0x0f, 0x7d, 0x96, 0x53, 0x1a,
	0xd3, 0xd3, 0x40, 0x60, 0xbe, 0x30, 0x27, 0x40, 0x97, 0x4f, 0x90, 0x53, 0xbf, 0x67, 0x74, 0x72,
	0x5f, 0x73, 0x01, 0xf6, 0xff, 0x5e, 0xc0, 0x56, 0xf1, 0xc2, 0x25, 0x43, 0x47, 0x31, 0xd4, 0x8f,
	0x57, 0xf3, 0xb3, 0x86, 0x8d, 0xb0, 0xa5, 0x52, 0xf6, 0xfe, 0x06, 0x00, 0x00, 0xff, 0xff, 0x67,
	0x5f, 0xf6, 0xe5, 0x8f, 0x05, 0x00, 0x00,
}
