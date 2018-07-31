// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/qscheduler/qslib/types/task/task.proto

package task

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"
import vector "infra/qscheduler/qslib/types/vector"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Request represents a requested task in the queue, and refers to the
// quota account to run it against. This representation intentionally
// excludes most of the details of a Swarming task request.
type Request struct {
	Id          string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	AccountId   string               `protobuf:"bytes,2,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	EnqueueTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=enqueue_time,json=enqueueTime,proto3" json:"enqueue_time,omitempty"`
	// The set of Provisionable Labels for this task.
	Labels               []string `protobuf:"bytes,4,rep,name=labels,proto3" json:"labels,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_task_e7c7f9a62484acbf, []int{0}
}
func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (dst *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(dst, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Request) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *Request) GetEnqueueTime() *timestamp.Timestamp {
	if m != nil {
		return m.EnqueueTime
	}
	return nil
}

func (m *Request) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

// Running represents a task that has been assigned to a worker and is
// now running.
type Run struct {
	// The total cost that has been incurred on this task while running.
	Cost *vector.Vector `protobuf:"bytes,1,opt,name=cost,proto3" json:"cost,omitempty"`
	// The request that this running task corresponds to.
	Request *Request `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	// The current priority level of the running task.
	Priority             int32    `protobuf:"varint,3,opt,name=priority,proto3" json:"priority,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Run) Reset()         { *m = Run{} }
func (m *Run) String() string { return proto.CompactTextString(m) }
func (*Run) ProtoMessage()    {}
func (*Run) Descriptor() ([]byte, []int) {
	return fileDescriptor_task_e7c7f9a62484acbf, []int{1}
}
func (m *Run) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Run.Unmarshal(m, b)
}
func (m *Run) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Run.Marshal(b, m, deterministic)
}
func (dst *Run) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Run.Merge(dst, src)
}
func (m *Run) XXX_Size() int {
	return xxx_messageInfo_Run.Size(m)
}
func (m *Run) XXX_DiscardUnknown() {
	xxx_messageInfo_Run.DiscardUnknown(m)
}

var xxx_messageInfo_Run proto.InternalMessageInfo

func (m *Run) GetCost() *vector.Vector {
	if m != nil {
		return m.Cost
	}
	return nil
}

func (m *Run) GetRequest() *Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *Run) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func init() {
	proto.RegisterType((*Request)(nil), "task.Request")
	proto.RegisterType((*Run)(nil), "task.Run")
}

func init() {
	proto.RegisterFile("infra/qscheduler/qslib/types/task/task.proto", fileDescriptor_task_e7c7f9a62484acbf)
}

var fileDescriptor_task_e7c7f9a62484acbf = []byte{
	// 274 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0xd0, 0xcd, 0x4e, 0x84, 0x30,
	0x10, 0x07, 0xf0, 0xb0, 0xe0, 0xae, 0x0c, 0xba, 0x87, 0x1e, 0x0c, 0x21, 0x31, 0x12, 0x2e, 0x72,
	0x30, 0xad, 0xc1, 0xb3, 0x0f, 0xe0, 0xb5, 0x31, 0x5e, 0x37, 0x7c, 0xcc, 0xae, 0x8d, 0x2c, 0x85,
	0x7e, 0x98, 0xec, 0x53, 0xf8, 0xca, 0x86, 0x52, 0x3c, 0x7a, 0x61, 0x98, 0xff, 0x0c, 0xe4, 0xd7,
	0xc2, 0x93, 0x18, 0x8e, 0xaa, 0x66, 0x93, 0x6e, 0x3f, 0xb1, 0xb3, 0x3d, 0x2a, 0x36, 0xe9, 0x5e,
	0x34, 0xcc, 0x5c, 0x46, 0xd4, 0xcc, 0xd4, 0xfa, 0xcb, 0x3d, 0xe8, 0xa8, 0xa4, 0x91, 0x24, 0x9a,
	0xdf, 0xb3, 0x87, 0x93, 0x94, 0xa7, 0x1e, 0x99, 0xcb, 0x1a, 0x7b, 0x64, 0x46, 0x9c, 0x51, 0x9b,
	0xfa, 0x3c, 0x2e, 0x6b, 0xd9, 0xf3, 0xbf, 0x3f, 0xfd, 0xc6, 0xd6, 0x48, 0xe5, 0xcb, 0xf2, 0x45,
	0xf1, 0x13, 0xc0, 0x8e, 0xe3, 0x64, 0x51, 0x1b, 0xb2, 0x87, 0x8d, 0xe8, 0xd2, 0x20, 0x0f, 0xca,
	0x98, 0x6f, 0x44, 0x47, 0xee, 0x01, 0xea, 0xb6, 0x95, 0x76, 0x30, 0x07, 0xd1, 0xa5, 0x1b, 0x97,
	0xc7, 0x3e, 0x79, 0xeb, 0xc8, 0x2b, 0xdc, 0xe0, 0x30, 0x59, 0xb4, 0x78, 0x98, 0x1d, 0x69, 0x98,
	0x07, 0x65, 0x52, 0x65, 0x74, 0x41, 0xd2, 0x15, 0x49, 0xdf, 0x57, 0x24, 0x4f, 0xfc, 0xfe, 0x9c,
	0x90, 0x3b, 0xd8, 0xf6, 0x75, 0x83, 0xbd, 0x4e, 0xa3, 0x3c, 0x2c, 0x63, 0xee, 0xbb, 0x62, 0x80,
	0x90, 0xdb, 0x81, 0x14, 0x10, 0xb5, 0x52, 0x1b, 0xc7, 0x49, 0xaa, 0x3d, 0xf5, 0xea, 0x0f, 0x57,
	0xb8, 0x9b, 0x91, 0x47, 0xd8, 0xa9, 0xc5, 0xee, 0x74, 0x49, 0x75, 0x4b, 0xdd, 0x9d, 0xf9, 0x03,
	0xf1, 0x75, 0x4a, 0x32, 0xb8, 0x1e, 0x95, 0x90, 0x4a, 0x98, 0x8b, 0x63, 0x5e, 0xf1, 0xbf, 0xbe,
	0xd9, 0x3a, 0xe8, 0xcb, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x80, 0x31, 0x71, 0xdf, 0x91, 0x01,
	0x00, 0x00,
}
