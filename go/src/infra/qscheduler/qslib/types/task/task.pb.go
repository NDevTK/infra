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
// excludes most of the details of a true Swarming task request.
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
	return fileDescriptor_task_1f3b282ab4cc7b15, []int{0}
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
type Running struct {
	// The total cost that has been spent on this task while running.
	Cost *vector.Vector `protobuf:"bytes,1,opt,name=cost,proto3" json:"cost,omitempty"`
	// The request that this running task corresponds to.
	Request *Request `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	// The current priority level of the running task.
	Priority             int32    `protobuf:"varint,3,opt,name=priority,proto3" json:"priority,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Running) Reset()         { *m = Running{} }
func (m *Running) String() string { return proto.CompactTextString(m) }
func (*Running) ProtoMessage()    {}
func (*Running) Descriptor() ([]byte, []int) {
	return fileDescriptor_task_1f3b282ab4cc7b15, []int{1}
}
func (m *Running) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Running.Unmarshal(m, b)
}
func (m *Running) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Running.Marshal(b, m, deterministic)
}
func (dst *Running) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Running.Merge(dst, src)
}
func (m *Running) XXX_Size() int {
	return xxx_messageInfo_Running.Size(m)
}
func (m *Running) XXX_DiscardUnknown() {
	xxx_messageInfo_Running.DiscardUnknown(m)
}

var xxx_messageInfo_Running proto.InternalMessageInfo

func (m *Running) GetCost() *vector.Vector {
	if m != nil {
		return m.Cost
	}
	return nil
}

func (m *Running) GetRequest() *Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *Running) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func init() {
	proto.RegisterType((*Request)(nil), "task.Request")
	proto.RegisterType((*Running)(nil), "task.Running")
}

func init() {
	proto.RegisterFile("infra/qscheduler/qslib/types/task/task.proto", fileDescriptor_task_1f3b282ab4cc7b15)
}

var fileDescriptor_task_1f3b282ab4cc7b15 = []byte{
	// 276 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0xd0, 0xcd, 0x4e, 0x84, 0x30,
	0x10, 0x07, 0xf0, 0xc0, 0xe2, 0x22, 0x45, 0xf7, 0xd0, 0x83, 0x21, 0x24, 0x46, 0xc2, 0x45, 0x0e,
	0xa6, 0x35, 0x78, 0xf6, 0x01, 0xbc, 0x36, 0xc6, 0xeb, 0x86, 0x8f, 0x59, 0x6c, 0x64, 0x5b, 0xe8,
	0x87, 0xc9, 0x3e, 0x85, 0xaf, 0x6c, 0x28, 0xc5, 0xa3, 0x17, 0x86, 0xf9, 0xcf, 0x40, 0x7e, 0x2d,
	0x7a, 0xe2, 0xe2, 0xa4, 0x1a, 0x3a, 0xeb, 0xee, 0x13, 0x7a, 0x3b, 0x82, 0xa2, 0xb3, 0x1e, 0x79,
	0x4b, 0xcd, 0x65, 0x02, 0x4d, 0x4d, 0xa3, 0xbf, 0xdc, 0x83, 0x4c, 0x4a, 0x1a, 0x89, 0xa3, 0xe5,
	0x3d, 0x7f, 0x18, 0xa4, 0x1c, 0x46, 0xa0, 0x2e, 0x6b, 0xed, 0x89, 0x1a, 0x7e, 0x06, 0x6d, 0x9a,
	0xf3, 0xb4, 0xae, 0xe5, 0xcf, 0xff, 0xfe, 0xf4, 0x1b, 0x3a, 0x23, 0x95, 0x2f, 0xeb, 0x17, 0xe5,
	0x4f, 0x80, 0x62, 0x06, 0xb3, 0x05, 0x6d, 0xf0, 0x01, 0x85, 0xbc, 0xcf, 0x82, 0x22, 0xa8, 0x12,
	0x16, 0xf2, 0x1e, 0xdf, 0x23, 0xd4, 0x74, 0x9d, 0xb4, 0xc2, 0x1c, 0x79, 0x9f, 0x85, 0x2e, 0x4f,
	0x7c, 0xf2, 0xd6, 0xe3, 0x57, 0x74, 0x03, 0x62, 0xb6, 0x60, 0xe1, 0xb8, 0x38, 0xb2, 0x5d, 0x11,
	0x54, 0x69, 0x9d, 0x93, 0x15, 0x49, 0x36, 0x24, 0x79, 0xdf, 0x90, 0x2c, 0xf5, 0xfb, 0x4b, 0x82,
	0xef, 0xd0, 0x7e, 0x6c, 0x5a, 0x18, 0x75, 0x16, 0x15, 0xbb, 0x2a, 0x61, 0xbe, 0x2b, 0x15, 0x8a,
	0x99, 0x15, 0x82, 0x8b, 0x01, 0x97, 0x28, 0xea, 0xa4, 0x36, 0x8e, 0x94, 0xd6, 0x07, 0xe2, 0xe5,
	0x1f, 0xae, 0x30, 0x37, 0xc3, 0x8f, 0x28, 0x56, 0xab, 0xdf, 0x09, 0xd3, 0xfa, 0x96, 0xb8, 0x7b,
	0xf3, 0x87, 0x62, 0xdb, 0x14, 0xe7, 0xe8, 0x7a, 0x52, 0x5c, 0x2a, 0x6e, 0x2e, 0x8e, 0x7a, 0xc5,
	0xfe, 0xfa, 0x76, 0xef, 0xb0, 0x2f, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x13, 0xa9, 0x92, 0xb9,
	0x95, 0x01, 0x00, 0x00,
}
