// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/crosskylabadmin/api/fleet/v1/tracker.proto

package fleet

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
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

type DutState int32

const (
	DutState_DutStateInvalid DutState = 0
	DutState_Ready           DutState = 1
	DutState_NeedsCleanup    DutState = 2
	DutState_NeedsRepair     DutState = 3
	DutState_NeedsReset      DutState = 4
	DutState_RepairFailed    DutState = 5
)

var DutState_name = map[int32]string{
	0: "DutStateInvalid",
	1: "Ready",
	2: "NeedsCleanup",
	3: "NeedsRepair",
	4: "NeedsReset",
	5: "RepairFailed",
}

var DutState_value = map[string]int32{
	"DutStateInvalid": 0,
	"Ready":           1,
	"NeedsCleanup":    2,
	"NeedsRepair":     3,
	"NeedsReset":      4,
	"RepairFailed":    5,
}

func (x DutState) String() string {
	return proto.EnumName(DutState_name, int32(x))
}

func (DutState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{0}
}

type Health int32

const (
	Health_HealthInvalid Health = 0
	// A Healthy bot may be used for external workload.
	Health_Healthy Health = 1
	// An Unhealthy bot is not usable for external workload.
	// Further classification of the problem is not available.
	Health_Unhealthy Health = 2
)

var Health_name = map[int32]string{
	0: "HealthInvalid",
	1: "Healthy",
	2: "Unhealthy",
}

var Health_value = map[string]int32{
	"HealthInvalid": 0,
	"Healthy":       1,
	"Unhealthy":     2,
}

func (x Health) String() string {
	return proto.EnumName(Health_name, int32(x))
}

func (Health) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{1}
}

type TaskType int32

const (
	TaskType_Invalid TaskType = 0
	TaskType_Reset   TaskType = 1
	TaskType_Cleanup TaskType = 2
	TaskType_Repair  TaskType = 3
)

var TaskType_name = map[int32]string{
	0: "Invalid",
	1: "Reset",
	2: "Cleanup",
	3: "Repair",
}

var TaskType_value = map[string]int32{
	"Invalid": 0,
	"Reset":   1,
	"Cleanup": 2,
	"Repair":  3,
}

func (x TaskType) String() string {
	return proto.EnumName(TaskType_name, int32(x))
}

func (TaskType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{2}
}

type PushBotsForAdminTasksRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushBotsForAdminTasksRequest) Reset()         { *m = PushBotsForAdminTasksRequest{} }
func (m *PushBotsForAdminTasksRequest) String() string { return proto.CompactTextString(m) }
func (*PushBotsForAdminTasksRequest) ProtoMessage()    {}
func (*PushBotsForAdminTasksRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{0}
}

func (m *PushBotsForAdminTasksRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushBotsForAdminTasksRequest.Unmarshal(m, b)
}
func (m *PushBotsForAdminTasksRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushBotsForAdminTasksRequest.Marshal(b, m, deterministic)
}
func (m *PushBotsForAdminTasksRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushBotsForAdminTasksRequest.Merge(m, src)
}
func (m *PushBotsForAdminTasksRequest) XXX_Size() int {
	return xxx_messageInfo_PushBotsForAdminTasksRequest.Size(m)
}
func (m *PushBotsForAdminTasksRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PushBotsForAdminTasksRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PushBotsForAdminTasksRequest proto.InternalMessageInfo

type PushBotsForAdminTasksResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushBotsForAdminTasksResponse) Reset()         { *m = PushBotsForAdminTasksResponse{} }
func (m *PushBotsForAdminTasksResponse) String() string { return proto.CompactTextString(m) }
func (*PushBotsForAdminTasksResponse) ProtoMessage()    {}
func (*PushBotsForAdminTasksResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{1}
}

func (m *PushBotsForAdminTasksResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushBotsForAdminTasksResponse.Unmarshal(m, b)
}
func (m *PushBotsForAdminTasksResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushBotsForAdminTasksResponse.Marshal(b, m, deterministic)
}
func (m *PushBotsForAdminTasksResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushBotsForAdminTasksResponse.Merge(m, src)
}
func (m *PushBotsForAdminTasksResponse) XXX_Size() int {
	return xxx_messageInfo_PushBotsForAdminTasksResponse.Size(m)
}
func (m *PushBotsForAdminTasksResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PushBotsForAdminTasksResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PushBotsForAdminTasksResponse proto.InternalMessageInfo

type PushRepairJobsForLabstationsRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushRepairJobsForLabstationsRequest) Reset()         { *m = PushRepairJobsForLabstationsRequest{} }
func (m *PushRepairJobsForLabstationsRequest) String() string { return proto.CompactTextString(m) }
func (*PushRepairJobsForLabstationsRequest) ProtoMessage()    {}
func (*PushRepairJobsForLabstationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{2}
}

func (m *PushRepairJobsForLabstationsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushRepairJobsForLabstationsRequest.Unmarshal(m, b)
}
func (m *PushRepairJobsForLabstationsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushRepairJobsForLabstationsRequest.Marshal(b, m, deterministic)
}
func (m *PushRepairJobsForLabstationsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushRepairJobsForLabstationsRequest.Merge(m, src)
}
func (m *PushRepairJobsForLabstationsRequest) XXX_Size() int {
	return xxx_messageInfo_PushRepairJobsForLabstationsRequest.Size(m)
}
func (m *PushRepairJobsForLabstationsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PushRepairJobsForLabstationsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PushRepairJobsForLabstationsRequest proto.InternalMessageInfo

type PushRepairJobsForLabstationsResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PushRepairJobsForLabstationsResponse) Reset()         { *m = PushRepairJobsForLabstationsResponse{} }
func (m *PushRepairJobsForLabstationsResponse) String() string { return proto.CompactTextString(m) }
func (*PushRepairJobsForLabstationsResponse) ProtoMessage()    {}
func (*PushRepairJobsForLabstationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{3}
}

func (m *PushRepairJobsForLabstationsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PushRepairJobsForLabstationsResponse.Unmarshal(m, b)
}
func (m *PushRepairJobsForLabstationsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PushRepairJobsForLabstationsResponse.Marshal(b, m, deterministic)
}
func (m *PushRepairJobsForLabstationsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PushRepairJobsForLabstationsResponse.Merge(m, src)
}
func (m *PushRepairJobsForLabstationsResponse) XXX_Size() int {
	return xxx_messageInfo_PushRepairJobsForLabstationsResponse.Size(m)
}
func (m *PushRepairJobsForLabstationsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PushRepairJobsForLabstationsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PushRepairJobsForLabstationsResponse proto.InternalMessageInfo

type ReportBotsRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportBotsRequest) Reset()         { *m = ReportBotsRequest{} }
func (m *ReportBotsRequest) String() string { return proto.CompactTextString(m) }
func (*ReportBotsRequest) ProtoMessage()    {}
func (*ReportBotsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{4}
}

func (m *ReportBotsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportBotsRequest.Unmarshal(m, b)
}
func (m *ReportBotsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportBotsRequest.Marshal(b, m, deterministic)
}
func (m *ReportBotsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportBotsRequest.Merge(m, src)
}
func (m *ReportBotsRequest) XXX_Size() int {
	return xxx_messageInfo_ReportBotsRequest.Size(m)
}
func (m *ReportBotsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportBotsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReportBotsRequest proto.InternalMessageInfo

type ReportBotsResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportBotsResponse) Reset()         { *m = ReportBotsResponse{} }
func (m *ReportBotsResponse) String() string { return proto.CompactTextString(m) }
func (*ReportBotsResponse) ProtoMessage()    {}
func (*ReportBotsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_474af594abe23e82, []int{5}
}

func (m *ReportBotsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportBotsResponse.Unmarshal(m, b)
}
func (m *ReportBotsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportBotsResponse.Marshal(b, m, deterministic)
}
func (m *ReportBotsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportBotsResponse.Merge(m, src)
}
func (m *ReportBotsResponse) XXX_Size() int {
	return xxx_messageInfo_ReportBotsResponse.Size(m)
}
func (m *ReportBotsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportBotsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ReportBotsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterEnum("crosskylabadmin.fleet.DutState", DutState_name, DutState_value)
	proto.RegisterEnum("crosskylabadmin.fleet.Health", Health_name, Health_value)
	proto.RegisterEnum("crosskylabadmin.fleet.TaskType", TaskType_name, TaskType_value)
	proto.RegisterType((*PushBotsForAdminTasksRequest)(nil), "crosskylabadmin.fleet.PushBotsForAdminTasksRequest")
	proto.RegisterType((*PushBotsForAdminTasksResponse)(nil), "crosskylabadmin.fleet.PushBotsForAdminTasksResponse")
	proto.RegisterType((*PushRepairJobsForLabstationsRequest)(nil), "crosskylabadmin.fleet.PushRepairJobsForLabstationsRequest")
	proto.RegisterType((*PushRepairJobsForLabstationsResponse)(nil), "crosskylabadmin.fleet.PushRepairJobsForLabstationsResponse")
	proto.RegisterType((*ReportBotsRequest)(nil), "crosskylabadmin.fleet.ReportBotsRequest")
	proto.RegisterType((*ReportBotsResponse)(nil), "crosskylabadmin.fleet.ReportBotsResponse")
}

func init() {
	proto.RegisterFile("infra/appengine/crosskylabadmin/api/fleet/v1/tracker.proto", fileDescriptor_474af594abe23e82)
}

var fileDescriptor_474af594abe23e82 = []byte{
	// 403 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0x4d, 0x8f, 0x12, 0x41,
	0x10, 0x75, 0x76, 0x05, 0x76, 0x6b, 0x5d, 0x77, 0x68, 0xe4, 0x42, 0xfc, 0x48, 0xf0, 0x23, 0xca,
	0x61, 0x26, 0x8a, 0x89, 0x09, 0x9c, 0x44, 0x43, 0xd4, 0x18, 0x63, 0x46, 0xbc, 0x78, 0xab, 0x61,
	0x0a, 0xe9, 0x30, 0x76, 0xb7, 0xdd, 0x3d, 0x24, 0x5c, 0xbd, 0xfa, 0x03, 0xfc, 0xbb, 0xa6, 0xa7,
	0x07, 0x31, 0x08, 0xa8, 0x7b, 0xeb, 0x79, 0xf5, 0xaa, 0xde, 0x4b, 0xbd, 0x1a, 0x18, 0x70, 0x31,
	0xd3, 0x18, 0xa3, 0x52, 0x24, 0x3e, 0x73, 0x41, 0xf1, 0x54, 0x4b, 0x63, 0x16, 0xab, 0x1c, 0x53,
	0xcc, 0xbe, 0x70, 0x11, 0xa3, 0xe2, 0xf1, 0x2c, 0x27, 0xb2, 0xf1, 0xf2, 0x71, 0x6c, 0x35, 0x4e,
	0x17, 0xa4, 0x23, 0xa5, 0xa5, 0x95, 0xac, 0xbd, 0xc5, 0x8d, 0x4a, 0x5e, 0xf7, 0x36, 0xdc, 0x7c,
	0x5f, 0x98, 0xf9, 0x48, 0x5a, 0x33, 0x96, 0xfa, 0xb9, 0xab, 0x4c, 0xd0, 0x2c, 0x4c, 0x42, 0x5f,
	0x0b, 0x32, 0xb6, 0x7b, 0x07, 0x6e, 0xed, 0xa9, 0x1b, 0x25, 0x85, 0xa1, 0xee, 0x7d, 0xb8, 0xeb,
	0x08, 0x09, 0x29, 0xe4, 0xfa, 0x8d, 0x4c, 0x1d, 0xed, 0x2d, 0xa6, 0xc6, 0xa2, 0xe5, 0x52, 0xfc,
	0x9a, 0xf3, 0x00, 0xee, 0x1d, 0xa6, 0x55, 0xe3, 0x5a, 0xd0, 0x4c, 0x48, 0x49, 0x6d, 0x9d, 0xe2,
	0xba, 0xf9, 0x06, 0xb0, 0xdf, 0x41, 0x4f, 0xed, 0x49, 0x38, 0x79, 0x59, 0xd8, 0x0f, 0x16, 0x2d,
	0xb1, 0x16, 0x5c, 0xac, 0xdf, 0xaf, 0xc5, 0x12, 0x73, 0x9e, 0x85, 0x57, 0xd8, 0x29, 0xd4, 0x12,
	0xc2, 0x6c, 0x15, 0x06, 0x2c, 0x84, 0x6b, 0xef, 0x88, 0x32, 0xf3, 0x22, 0x27, 0x14, 0x85, 0x0a,
	0x8f, 0xd8, 0x05, 0x9c, 0x95, 0x88, 0x77, 0x14, 0x1e, 0xb3, 0xeb, 0x00, 0x15, 0x60, 0xc8, 0x86,
	0x57, 0x5d, 0x8b, 0xaf, 0x8d, 0x91, 0xe7, 0x94, 0x85, 0xb5, 0xde, 0x33, 0xa8, 0xbf, 0x22, 0xcc,
	0xed, 0x9c, 0x35, 0xe1, 0xdc, 0xbf, 0x36, 0x62, 0x67, 0xd0, 0xf0, 0x90, 0x93, 0x3b, 0x87, 0xd3,
	0x8f, 0x62, 0x5e, 0x7d, 0x1e, 0xf5, 0x86, 0x70, 0xe2, 0x96, 0x36, 0x59, 0x29, 0x72, 0xbc, 0x2d,
	0x87, 0x4e, 0x2e, 0x70, 0xf8, 0xc6, 0x1c, 0x40, 0x7d, 0xed, 0xeb, 0xc9, 0xf7, 0x63, 0x68, 0x4c,
	0x7c, 0x94, 0xec, 0x5b, 0x00, 0xed, 0x9d, 0x71, 0xb0, 0x7e, 0xb4, 0x33, 0xdf, 0xe8, 0x50, 0xb8,
	0x9d, 0xa7, 0xff, 0xd7, 0xe4, 0xf7, 0xce, 0x7e, 0x04, 0xfe, 0x66, 0xf6, 0x65, 0xc9, 0x06, 0x07,
	0xc6, 0xfe, 0xe5, 0x4e, 0x3a, 0xc3, 0x4b, 0xf5, 0x56, 0xce, 0x10, 0x60, 0x73, 0x27, 0xec, 0xe1,
	0x9e, 0x51, 0x7f, 0xdc, 0x57, 0xe7, 0xd1, 0x3f, 0x30, 0xbd, 0xc4, 0xa8, 0xf1, 0xa9, 0x56, 0xd6,
	0xd2, 0x7a, 0xf9, 0x5b, 0xf5, 0x7f, 0x06, 0x00, 0x00, 0xff, 0xff, 0x6c, 0x97, 0xce, 0xda, 0x94,
	0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TrackerClient is the client API for Tracker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TrackerClient interface {
	// Filter out and queue the CrOS bots that require admin tasks Repair and Reset.
	PushBotsForAdminTasks(ctx context.Context, in *PushBotsForAdminTasksRequest, opts ...grpc.CallOption) (*PushBotsForAdminTasksResponse, error)
	// Filter out and queue the labstation bots that require admin tasks Repair.
	PushRepairJobsForLabstations(ctx context.Context, in *PushRepairJobsForLabstationsRequest, opts ...grpc.CallOption) (*PushRepairJobsForLabstationsResponse, error)
	// Report bots metrics.
	ReportBots(ctx context.Context, in *ReportBotsRequest, opts ...grpc.CallOption) (*ReportBotsResponse, error)
}
type trackerPRPCClient struct {
	client *prpc.Client
}

func NewTrackerPRPCClient(client *prpc.Client) TrackerClient {
	return &trackerPRPCClient{client}
}

func (c *trackerPRPCClient) PushBotsForAdminTasks(ctx context.Context, in *PushBotsForAdminTasksRequest, opts ...grpc.CallOption) (*PushBotsForAdminTasksResponse, error) {
	out := new(PushBotsForAdminTasksResponse)
	err := c.client.Call(ctx, "crosskylabadmin.fleet.Tracker", "PushBotsForAdminTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerPRPCClient) PushRepairJobsForLabstations(ctx context.Context, in *PushRepairJobsForLabstationsRequest, opts ...grpc.CallOption) (*PushRepairJobsForLabstationsResponse, error) {
	out := new(PushRepairJobsForLabstationsResponse)
	err := c.client.Call(ctx, "crosskylabadmin.fleet.Tracker", "PushRepairJobsForLabstations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerPRPCClient) ReportBots(ctx context.Context, in *ReportBotsRequest, opts ...grpc.CallOption) (*ReportBotsResponse, error) {
	out := new(ReportBotsResponse)
	err := c.client.Call(ctx, "crosskylabadmin.fleet.Tracker", "ReportBots", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type trackerClient struct {
	cc *grpc.ClientConn
}

func NewTrackerClient(cc *grpc.ClientConn) TrackerClient {
	return &trackerClient{cc}
}

func (c *trackerClient) PushBotsForAdminTasks(ctx context.Context, in *PushBotsForAdminTasksRequest, opts ...grpc.CallOption) (*PushBotsForAdminTasksResponse, error) {
	out := new(PushBotsForAdminTasksResponse)
	err := c.cc.Invoke(ctx, "/crosskylabadmin.fleet.Tracker/PushBotsForAdminTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerClient) PushRepairJobsForLabstations(ctx context.Context, in *PushRepairJobsForLabstationsRequest, opts ...grpc.CallOption) (*PushRepairJobsForLabstationsResponse, error) {
	out := new(PushRepairJobsForLabstationsResponse)
	err := c.cc.Invoke(ctx, "/crosskylabadmin.fleet.Tracker/PushRepairJobsForLabstations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trackerClient) ReportBots(ctx context.Context, in *ReportBotsRequest, opts ...grpc.CallOption) (*ReportBotsResponse, error) {
	out := new(ReportBotsResponse)
	err := c.cc.Invoke(ctx, "/crosskylabadmin.fleet.Tracker/ReportBots", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrackerServer is the server API for Tracker service.
type TrackerServer interface {
	// Filter out and queue the CrOS bots that require admin tasks Repair and Reset.
	PushBotsForAdminTasks(context.Context, *PushBotsForAdminTasksRequest) (*PushBotsForAdminTasksResponse, error)
	// Filter out and queue the labstation bots that require admin tasks Repair.
	PushRepairJobsForLabstations(context.Context, *PushRepairJobsForLabstationsRequest) (*PushRepairJobsForLabstationsResponse, error)
	// Report bots metrics.
	ReportBots(context.Context, *ReportBotsRequest) (*ReportBotsResponse, error)
}

// UnimplementedTrackerServer can be embedded to have forward compatible implementations.
type UnimplementedTrackerServer struct {
}

func (*UnimplementedTrackerServer) PushBotsForAdminTasks(ctx context.Context, req *PushBotsForAdminTasksRequest) (*PushBotsForAdminTasksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushBotsForAdminTasks not implemented")
}
func (*UnimplementedTrackerServer) PushRepairJobsForLabstations(ctx context.Context, req *PushRepairJobsForLabstationsRequest) (*PushRepairJobsForLabstationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushRepairJobsForLabstations not implemented")
}
func (*UnimplementedTrackerServer) ReportBots(ctx context.Context, req *ReportBotsRequest) (*ReportBotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportBots not implemented")
}

func RegisterTrackerServer(s prpc.Registrar, srv TrackerServer) {
	s.RegisterService(&_Tracker_serviceDesc, srv)
}

func _Tracker_PushBotsForAdminTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushBotsForAdminTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).PushBotsForAdminTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crosskylabadmin.fleet.Tracker/PushBotsForAdminTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).PushBotsForAdminTasks(ctx, req.(*PushBotsForAdminTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tracker_PushRepairJobsForLabstations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRepairJobsForLabstationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).PushRepairJobsForLabstations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crosskylabadmin.fleet.Tracker/PushRepairJobsForLabstations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).PushRepairJobsForLabstations(ctx, req.(*PushRepairJobsForLabstationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tracker_ReportBots_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportBotsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).ReportBots(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/crosskylabadmin.fleet.Tracker/ReportBots",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).ReportBots(ctx, req.(*ReportBotsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Tracker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "crosskylabadmin.fleet.Tracker",
	HandlerType: (*TrackerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PushBotsForAdminTasks",
			Handler:    _Tracker_PushBotsForAdminTasks_Handler,
		},
		{
			MethodName: "PushRepairJobsForLabstations",
			Handler:    _Tracker_PushRepairJobsForLabstations_Handler,
		},
		{
			MethodName: "ReportBots",
			Handler:    _Tracker_ReportBots_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/appengine/crosskylabadmin/api/fleet/v1/tracker.proto",
}
