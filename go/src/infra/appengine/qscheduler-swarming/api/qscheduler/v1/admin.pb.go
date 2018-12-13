// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/qscheduler-swarming/api/qscheduler/v1/admin.proto

package qscheduler

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	fmt "fmt"
	account "infra/qscheduler/qslib/types/account"
	vector "infra/qscheduler/qslib/types/vector"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
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

type CreateSchedulerPoolRequest struct {
	// TODO(akeshet): The client shouldn't be creating this id. It should be
	// creating a scheduler pool for a given name, and the id should be generated
	// by admin service and returned. Punting on this for now because the naming
	// and organization scheme for scheduler pools is not yet established. (i.e.
	// will we have some hierarchical structure to these pools? will there be
	// sub-pools?).
	PoolId string `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// Config is the scheduler configuration for the scheduler to create.
	Config               *SchedulerPoolConfig `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *CreateSchedulerPoolRequest) Reset()         { *m = CreateSchedulerPoolRequest{} }
func (m *CreateSchedulerPoolRequest) String() string { return proto.CompactTextString(m) }
func (*CreateSchedulerPoolRequest) ProtoMessage()    {}
func (*CreateSchedulerPoolRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{0}
}

func (m *CreateSchedulerPoolRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSchedulerPoolRequest.Unmarshal(m, b)
}
func (m *CreateSchedulerPoolRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSchedulerPoolRequest.Marshal(b, m, deterministic)
}
func (m *CreateSchedulerPoolRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSchedulerPoolRequest.Merge(m, src)
}
func (m *CreateSchedulerPoolRequest) XXX_Size() int {
	return xxx_messageInfo_CreateSchedulerPoolRequest.Size(m)
}
func (m *CreateSchedulerPoolRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSchedulerPoolRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSchedulerPoolRequest proto.InternalMessageInfo

func (m *CreateSchedulerPoolRequest) GetPoolId() string {
	if m != nil {
		return m.PoolId
	}
	return ""
}

func (m *CreateSchedulerPoolRequest) GetConfig() *SchedulerPoolConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

type CreateSchedulerPoolResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateSchedulerPoolResponse) Reset()         { *m = CreateSchedulerPoolResponse{} }
func (m *CreateSchedulerPoolResponse) String() string { return proto.CompactTextString(m) }
func (*CreateSchedulerPoolResponse) ProtoMessage()    {}
func (*CreateSchedulerPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{1}
}

func (m *CreateSchedulerPoolResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSchedulerPoolResponse.Unmarshal(m, b)
}
func (m *CreateSchedulerPoolResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSchedulerPoolResponse.Marshal(b, m, deterministic)
}
func (m *CreateSchedulerPoolResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSchedulerPoolResponse.Merge(m, src)
}
func (m *CreateSchedulerPoolResponse) XXX_Size() int {
	return xxx_messageInfo_CreateSchedulerPoolResponse.Size(m)
}
func (m *CreateSchedulerPoolResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSchedulerPoolResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSchedulerPoolResponse proto.InternalMessageInfo

type CreateAccountRequest struct {
	// PoolID is the id of the scheduler to create an account within.
	PoolId string `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// TODO(akeshet): Similar to pool_id above, account_id should be generated
	// on the server, not client. Instead, pass in some kind of path or
	// hierarchical account name. Punting until this is figured out better.
	AccountId string `protobuf:"bytes,2,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	// Config is the quota account config for the quota account to create.
	Config               *account.Config `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *CreateAccountRequest) Reset()         { *m = CreateAccountRequest{} }
func (m *CreateAccountRequest) String() string { return proto.CompactTextString(m) }
func (*CreateAccountRequest) ProtoMessage()    {}
func (*CreateAccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{2}
}

func (m *CreateAccountRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateAccountRequest.Unmarshal(m, b)
}
func (m *CreateAccountRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateAccountRequest.Marshal(b, m, deterministic)
}
func (m *CreateAccountRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateAccountRequest.Merge(m, src)
}
func (m *CreateAccountRequest) XXX_Size() int {
	return xxx_messageInfo_CreateAccountRequest.Size(m)
}
func (m *CreateAccountRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateAccountRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateAccountRequest proto.InternalMessageInfo

func (m *CreateAccountRequest) GetPoolId() string {
	if m != nil {
		return m.PoolId
	}
	return ""
}

func (m *CreateAccountRequest) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *CreateAccountRequest) GetConfig() *account.Config {
	if m != nil {
		return m.Config
	}
	return nil
}

type CreateAccountResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateAccountResponse) Reset()         { *m = CreateAccountResponse{} }
func (m *CreateAccountResponse) String() string { return proto.CompactTextString(m) }
func (*CreateAccountResponse) ProtoMessage()    {}
func (*CreateAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{3}
}

func (m *CreateAccountResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateAccountResponse.Unmarshal(m, b)
}
func (m *CreateAccountResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateAccountResponse.Marshal(b, m, deterministic)
}
func (m *CreateAccountResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateAccountResponse.Merge(m, src)
}
func (m *CreateAccountResponse) XXX_Size() int {
	return xxx_messageInfo_CreateAccountResponse.Size(m)
}
func (m *CreateAccountResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateAccountResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CreateAccountResponse proto.InternalMessageInfo

type ListAccountsRequest struct {
	// PoolID is the id of the scheduler to list accounts from.
	PoolId               string   `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListAccountsRequest) Reset()         { *m = ListAccountsRequest{} }
func (m *ListAccountsRequest) String() string { return proto.CompactTextString(m) }
func (*ListAccountsRequest) ProtoMessage()    {}
func (*ListAccountsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{4}
}

func (m *ListAccountsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListAccountsRequest.Unmarshal(m, b)
}
func (m *ListAccountsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListAccountsRequest.Marshal(b, m, deterministic)
}
func (m *ListAccountsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListAccountsRequest.Merge(m, src)
}
func (m *ListAccountsRequest) XXX_Size() int {
	return xxx_messageInfo_ListAccountsRequest.Size(m)
}
func (m *ListAccountsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListAccountsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListAccountsRequest proto.InternalMessageInfo

func (m *ListAccountsRequest) GetPoolId() string {
	if m != nil {
		return m.PoolId
	}
	return ""
}

type ListAccountsResponse struct {
	Accounts             map[string]*account.Config `protobuf:"bytes,1,rep,name=accounts,proto3" json:"accounts,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                   `json:"-"`
	XXX_unrecognized     []byte                     `json:"-"`
	XXX_sizecache        int32                      `json:"-"`
}

func (m *ListAccountsResponse) Reset()         { *m = ListAccountsResponse{} }
func (m *ListAccountsResponse) String() string { return proto.CompactTextString(m) }
func (*ListAccountsResponse) ProtoMessage()    {}
func (*ListAccountsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{5}
}

func (m *ListAccountsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListAccountsResponse.Unmarshal(m, b)
}
func (m *ListAccountsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListAccountsResponse.Marshal(b, m, deterministic)
}
func (m *ListAccountsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListAccountsResponse.Merge(m, src)
}
func (m *ListAccountsResponse) XXX_Size() int {
	return xxx_messageInfo_ListAccountsResponse.Size(m)
}
func (m *ListAccountsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListAccountsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListAccountsResponse proto.InternalMessageInfo

func (m *ListAccountsResponse) GetAccounts() map[string]*account.Config {
	if m != nil {
		return m.Accounts
	}
	return nil
}

type SchedulerPoolConfig struct {
	// Labels is a list of swarming dimensions in "key:value" form. This corresponds
	// to swarming.ExternalSchedulerConfig.dimensionsions; it is the minimal set
	// of dimensions for tasks or bots that will use this scheduler.
	Labels               []string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SchedulerPoolConfig) Reset()         { *m = SchedulerPoolConfig{} }
func (m *SchedulerPoolConfig) String() string { return proto.CompactTextString(m) }
func (*SchedulerPoolConfig) ProtoMessage()    {}
func (*SchedulerPoolConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{6}
}

func (m *SchedulerPoolConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SchedulerPoolConfig.Unmarshal(m, b)
}
func (m *SchedulerPoolConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SchedulerPoolConfig.Marshal(b, m, deterministic)
}
func (m *SchedulerPoolConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SchedulerPoolConfig.Merge(m, src)
}
func (m *SchedulerPoolConfig) XXX_Size() int {
	return xxx_messageInfo_SchedulerPoolConfig.Size(m)
}
func (m *SchedulerPoolConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_SchedulerPoolConfig.DiscardUnknown(m)
}

var xxx_messageInfo_SchedulerPoolConfig proto.InternalMessageInfo

func (m *SchedulerPoolConfig) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

type InspectPoolRequest struct {
	PoolId               string   `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InspectPoolRequest) Reset()         { *m = InspectPoolRequest{} }
func (m *InspectPoolRequest) String() string { return proto.CompactTextString(m) }
func (*InspectPoolRequest) ProtoMessage()    {}
func (*InspectPoolRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{7}
}

func (m *InspectPoolRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InspectPoolRequest.Unmarshal(m, b)
}
func (m *InspectPoolRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InspectPoolRequest.Marshal(b, m, deterministic)
}
func (m *InspectPoolRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InspectPoolRequest.Merge(m, src)
}
func (m *InspectPoolRequest) XXX_Size() int {
	return xxx_messageInfo_InspectPoolRequest.Size(m)
}
func (m *InspectPoolRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InspectPoolRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InspectPoolRequest proto.InternalMessageInfo

func (m *InspectPoolRequest) GetPoolId() string {
	if m != nil {
		return m.PoolId
	}
	return ""
}

type InspectPoolResponse struct {
	// NumWaitingTasks is the number of waiting tasks.
	NumWaitingTasks int32 `protobuf:"varint,1,opt,name=num_waiting_tasks,json=numWaitingTasks,proto3" json:"num_waiting_tasks,omitempty"`
	// NumRunningTasks is the number of running tasks.
	NumRunningTasks int32 `protobuf:"varint,2,opt,name=num_running_tasks,json=numRunningTasks,proto3" json:"num_running_tasks,omitempty"`
	// IdleBots is the number of idle bots.
	NumIdleBots int32 `protobuf:"varint,3,opt,name=num_idle_bots,json=numIdleBots,proto3" json:"num_idle_bots,omitempty"`
	// AccountBalances is the set of balances for all accounts.
	AccountBalances map[string]*vector.Vector `protobuf:"bytes,4,rep,name=account_balances,json=accountBalances,proto3" json:"account_balances,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// RunningTasks is a description of the running tasks according to
	// quotascheduler.
	RunningTasks         []*InspectPoolResponse_RunningTask `protobuf:"bytes,5,rep,name=running_tasks,json=runningTasks,proto3" json:"running_tasks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                           `json:"-"`
	XXX_unrecognized     []byte                             `json:"-"`
	XXX_sizecache        int32                              `json:"-"`
}

func (m *InspectPoolResponse) Reset()         { *m = InspectPoolResponse{} }
func (m *InspectPoolResponse) String() string { return proto.CompactTextString(m) }
func (*InspectPoolResponse) ProtoMessage()    {}
func (*InspectPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{8}
}

func (m *InspectPoolResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InspectPoolResponse.Unmarshal(m, b)
}
func (m *InspectPoolResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InspectPoolResponse.Marshal(b, m, deterministic)
}
func (m *InspectPoolResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InspectPoolResponse.Merge(m, src)
}
func (m *InspectPoolResponse) XXX_Size() int {
	return xxx_messageInfo_InspectPoolResponse.Size(m)
}
func (m *InspectPoolResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_InspectPoolResponse.DiscardUnknown(m)
}

var xxx_messageInfo_InspectPoolResponse proto.InternalMessageInfo

func (m *InspectPoolResponse) GetNumWaitingTasks() int32 {
	if m != nil {
		return m.NumWaitingTasks
	}
	return 0
}

func (m *InspectPoolResponse) GetNumRunningTasks() int32 {
	if m != nil {
		return m.NumRunningTasks
	}
	return 0
}

func (m *InspectPoolResponse) GetNumIdleBots() int32 {
	if m != nil {
		return m.NumIdleBots
	}
	return 0
}

func (m *InspectPoolResponse) GetAccountBalances() map[string]*vector.Vector {
	if m != nil {
		return m.AccountBalances
	}
	return nil
}

func (m *InspectPoolResponse) GetRunningTasks() []*InspectPoolResponse_RunningTask {
	if m != nil {
		return m.RunningTasks
	}
	return nil
}

type InspectPoolResponse_RunningTask struct {
	// Id is the id of the request.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// BotId is the id of the bot running the request.
	BotId string `protobuf:"bytes,2,opt,name=bot_id,json=botId,proto3" json:"bot_id,omitempty"`
	// Priority is the current quotascheduler priority that the task is
	// running at.
	Priority             int32    `protobuf:"varint,3,opt,name=priority,proto3" json:"priority,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InspectPoolResponse_RunningTask) Reset()         { *m = InspectPoolResponse_RunningTask{} }
func (m *InspectPoolResponse_RunningTask) String() string { return proto.CompactTextString(m) }
func (*InspectPoolResponse_RunningTask) ProtoMessage()    {}
func (*InspectPoolResponse_RunningTask) Descriptor() ([]byte, []int) {
	return fileDescriptor_19eb132fa8b54f73, []int{8, 1}
}

func (m *InspectPoolResponse_RunningTask) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InspectPoolResponse_RunningTask.Unmarshal(m, b)
}
func (m *InspectPoolResponse_RunningTask) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InspectPoolResponse_RunningTask.Marshal(b, m, deterministic)
}
func (m *InspectPoolResponse_RunningTask) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InspectPoolResponse_RunningTask.Merge(m, src)
}
func (m *InspectPoolResponse_RunningTask) XXX_Size() int {
	return xxx_messageInfo_InspectPoolResponse_RunningTask.Size(m)
}
func (m *InspectPoolResponse_RunningTask) XXX_DiscardUnknown() {
	xxx_messageInfo_InspectPoolResponse_RunningTask.DiscardUnknown(m)
}

var xxx_messageInfo_InspectPoolResponse_RunningTask proto.InternalMessageInfo

func (m *InspectPoolResponse_RunningTask) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *InspectPoolResponse_RunningTask) GetBotId() string {
	if m != nil {
		return m.BotId
	}
	return ""
}

func (m *InspectPoolResponse_RunningTask) GetPriority() int32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func init() {
	proto.RegisterType((*CreateSchedulerPoolRequest)(nil), "qscheduler.CreateSchedulerPoolRequest")
	proto.RegisterType((*CreateSchedulerPoolResponse)(nil), "qscheduler.CreateSchedulerPoolResponse")
	proto.RegisterType((*CreateAccountRequest)(nil), "qscheduler.CreateAccountRequest")
	proto.RegisterType((*CreateAccountResponse)(nil), "qscheduler.CreateAccountResponse")
	proto.RegisterType((*ListAccountsRequest)(nil), "qscheduler.ListAccountsRequest")
	proto.RegisterType((*ListAccountsResponse)(nil), "qscheduler.ListAccountsResponse")
	proto.RegisterMapType((map[string]*account.Config)(nil), "qscheduler.ListAccountsResponse.AccountsEntry")
	proto.RegisterType((*SchedulerPoolConfig)(nil), "qscheduler.SchedulerPoolConfig")
	proto.RegisterType((*InspectPoolRequest)(nil), "qscheduler.InspectPoolRequest")
	proto.RegisterType((*InspectPoolResponse)(nil), "qscheduler.InspectPoolResponse")
	proto.RegisterMapType((map[string]*vector.Vector)(nil), "qscheduler.InspectPoolResponse.AccountBalancesEntry")
	proto.RegisterType((*InspectPoolResponse_RunningTask)(nil), "qscheduler.InspectPoolResponse.RunningTask")
}

func init() {
	proto.RegisterFile("infra/appengine/qscheduler-swarming/api/qscheduler/v1/admin.proto", fileDescriptor_19eb132fa8b54f73)
}

var fileDescriptor_19eb132fa8b54f73 = []byte{
	// 640 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x51, 0x6f, 0xd3, 0x3e,
	0x10, 0x57, 0xd3, 0x7f, 0xfb, 0xdf, 0xae, 0xeb, 0x3a, 0xdc, 0x8d, 0x55, 0x41, 0x63, 0x25, 0x02,
	0x36, 0x81, 0x96, 0x40, 0x41, 0x02, 0xf1, 0xb6, 0x4d, 0x3c, 0x14, 0x4d, 0x68, 0x0b, 0x13, 0x3c,
	0x56, 0x4e, 0xe2, 0x15, 0x6b, 0xa9, 0x9d, 0xc5, 0xce, 0xa6, 0x7e, 0x2b, 0x3e, 0x17, 0x8f, 0x7c,
	0x02, 0x94, 0xd8, 0xe9, 0x92, 0x2d, 0x5d, 0x78, 0x4a, 0x7c, 0xf7, 0xbb, 0xbb, 0x9f, 0xef, 0x77,
	0x67, 0x38, 0xa4, 0xec, 0x22, 0xc6, 0x0e, 0x8e, 0x22, 0xc2, 0xa6, 0x94, 0x11, 0xe7, 0x4a, 0xf8,
	0x3f, 0x49, 0x90, 0x84, 0x24, 0x3e, 0x10, 0x37, 0x38, 0x9e, 0x51, 0x36, 0x75, 0x70, 0x44, 0x0b,
	0x76, 0xe7, 0xfa, 0xad, 0x83, 0x83, 0x19, 0x65, 0x76, 0x14, 0x73, 0xc9, 0x11, 0xdc, 0xba, 0xcc,
	0x91, 0x4a, 0x57, 0x00, 0x5f, 0x89, 0x90, 0x7a, 0x8e, 0x9c, 0x47, 0x44, 0x38, 0xd8, 0xf7, 0x79,
	0xc2, 0x64, 0xfe, 0x55, 0xf1, 0xe6, 0x9b, 0x07, 0x63, 0xae, 0x89, 0x2f, 0x79, 0xac, 0x3f, 0x2a,
	0xc2, 0x62, 0x60, 0x1e, 0xc7, 0x04, 0x4b, 0xf2, 0x2d, 0x0f, 0x39, 0xe5, 0x3c, 0x74, 0xc9, 0x55,
	0x42, 0x84, 0x44, 0xdb, 0xf0, 0x7f, 0xc4, 0x79, 0x38, 0xa1, 0xc1, 0xa0, 0x31, 0x6c, 0xec, 0xaf,
	0xba, 0xed, 0xf4, 0x38, 0x0e, 0xd0, 0x07, 0x68, 0xfb, 0x9c, 0x5d, 0xd0, 0xe9, 0xc0, 0x18, 0x36,
	0xf6, 0x3b, 0xa3, 0x5d, 0xfb, 0xb6, 0xa6, 0x5d, 0x4a, 0x75, 0x9c, 0xc1, 0x5c, 0x0d, 0xb7, 0x76,
	0xe0, 0x49, 0x65, 0x3d, 0x11, 0x71, 0x26, 0x88, 0x75, 0x03, 0x9b, 0xca, 0x7d, 0xa8, 0xee, 0x55,
	0x4b, 0x64, 0x07, 0x40, 0xb7, 0x20, 0xf5, 0x19, 0x99, 0x6f, 0x55, 0x5b, 0xc6, 0x01, 0xda, 0x5b,
	0xf0, 0x6c, 0x66, 0x3c, 0x7b, 0x76, 0xde, 0xb0, 0x3b, 0xbc, 0xb6, 0x61, 0xeb, 0x4e, 0x61, 0xcd,
	0xc8, 0x86, 0xfe, 0x09, 0x15, 0x52, 0x9b, 0x45, 0x1d, 0x21, 0xeb, 0x57, 0x03, 0x36, 0xcb, 0x01,
	0x2a, 0x11, 0xfa, 0x02, 0x2b, 0xba, 0xb6, 0x18, 0x34, 0x86, 0xcd, 0xfd, 0xce, 0xc8, 0x2e, 0x36,
	0xad, 0x2a, 0xc6, 0xce, 0x0d, 0x9f, 0x99, 0x8c, 0xe7, 0xee, 0x22, 0xde, 0x3c, 0x81, 0x6e, 0xc9,
	0x85, 0x36, 0xa0, 0x79, 0x49, 0xe6, 0x9a, 0x4a, 0xfa, 0x8b, 0x5e, 0x40, 0xeb, 0x1a, 0x87, 0x09,
	0xd1, 0x02, 0xdd, 0xbb, 0xb8, 0xf2, 0x7e, 0x32, 0x3e, 0x36, 0xac, 0x03, 0xe8, 0x57, 0x48, 0x86,
	0x1e, 0x43, 0x3b, 0xc4, 0x1e, 0x09, 0x15, 0xdd, 0x55, 0x57, 0x9f, 0xac, 0x03, 0x40, 0x63, 0x26,
	0x22, 0xe2, 0xcb, 0x7f, 0x19, 0x15, 0xeb, 0x77, 0x13, 0xfa, 0x25, 0xbc, 0xee, 0xc7, 0x2b, 0x78,
	0xc4, 0x92, 0xd9, 0xe4, 0x06, 0x53, 0x49, 0xd9, 0x74, 0x22, 0xb1, 0xb8, 0x14, 0x59, 0x68, 0xcb,
	0xed, 0xb1, 0x64, 0xf6, 0x43, 0xd9, 0xcf, 0x53, 0x73, 0x8e, 0x8d, 0x13, 0xc6, 0x6e, 0xb1, 0xc6,
	0x02, 0xeb, 0x2a, 0xbb, 0xc2, 0x5a, 0xd0, 0x4d, 0xb1, 0x34, 0x08, 0xc9, 0xc4, 0xe3, 0x52, 0x64,
	0xca, 0xb7, 0xdc, 0x0e, 0x4b, 0x66, 0xe3, 0x20, 0x24, 0x47, 0x5c, 0x0a, 0x34, 0x81, 0x8d, 0x7c,
	0x6a, 0x3c, 0x1c, 0x62, 0xe6, 0x13, 0x31, 0xf8, 0x2f, 0xd3, 0xe4, 0x7d, 0x51, 0x93, 0x0a, 0xda,
	0xb9, 0x24, 0x47, 0x3a, 0x4c, 0x29, 0xd3, 0xc3, 0x65, 0x2b, 0x3a, 0x85, 0x6e, 0x99, 0x6c, 0x2b,
	0xcb, 0xfe, 0xba, 0x2e, 0x7b, 0xe1, 0x26, 0xee, 0x5a, 0x5c, 0xb8, 0x96, 0xe9, 0xc2, 0x66, 0x55,
	0xe9, 0x0a, 0xe5, 0x9f, 0x97, 0x95, 0x5f, 0xb7, 0xf5, 0xc2, 0x7f, 0xcf, 0x3e, 0x05, 0xe1, 0xcd,
	0x53, 0xe8, 0x14, 0x0a, 0xa2, 0x75, 0x30, 0x16, 0xea, 0x19, 0x34, 0x40, 0x5b, 0xd0, 0xf6, 0x78,
	0x61, 0xaf, 0x5a, 0x1e, 0x4f, 0x77, 0xca, 0x84, 0x95, 0x28, 0xa6, 0x3c, 0xa6, 0x72, 0xae, 0x7b,
	0xbb, 0x38, 0x8f, 0xfe, 0x18, 0xd0, 0x3b, 0x5b, 0x0c, 0xd3, 0x61, 0xfa, 0xb4, 0xa1, 0x0b, 0xe8,
	0x57, 0xac, 0x3c, 0x7a, 0x59, 0xec, 0xc5, 0xf2, 0x37, 0xc8, 0xdc, 0xab, 0xc5, 0xe9, 0x81, 0x3a,
	0x87, 0x6e, 0x69, 0x85, 0xd1, 0xf0, 0x7e, 0x64, 0xf9, 0x59, 0x31, 0x9f, 0x3d, 0x80, 0xd0, 0x59,
	0xcf, 0x60, 0xad, 0xb8, 0x9a, 0x68, 0x77, 0xf9, 0xd2, 0xaa, 0x9c, 0xc3, 0xba, 0xad, 0x46, 0x5f,
	0xa1, 0x53, 0xd0, 0x1e, 0x3d, 0x5d, 0x3a, 0x14, 0x2a, 0xe1, 0x6e, 0xcd, 0xd0, 0x78, 0xed, 0xec,
	0x29, 0x7f, 0xf7, 0x37, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x01, 0xb8, 0xf2, 0x81, 0x06, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QSchedulerAdminClient is the client API for QSchedulerAdmin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QSchedulerAdminClient interface {
	// CreateSchedulerPool creates a scheduler, with the given configuration
	// options.
	CreateSchedulerPool(ctx context.Context, in *CreateSchedulerPoolRequest, opts ...grpc.CallOption) (*CreateSchedulerPoolResponse, error)
	// CreateAccount creates a quota account within a scheduler, with the
	// given configuration options.
	CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error)
	// ListAccounts returns the set of accounts for a given scheduler.
	ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*ListAccountsResponse, error)
	// InspectPool returns a description of the state of a scheduler, for debugging
	// or diagnostic purposes.
	InspectPool(ctx context.Context, in *InspectPoolRequest, opts ...grpc.CallOption) (*InspectPoolResponse, error)
}
type qSchedulerAdminPRPCClient struct {
	client *prpc.Client
}

func NewQSchedulerAdminPRPCClient(client *prpc.Client) QSchedulerAdminClient {
	return &qSchedulerAdminPRPCClient{client}
}

func (c *qSchedulerAdminPRPCClient) CreateSchedulerPool(ctx context.Context, in *CreateSchedulerPoolRequest, opts ...grpc.CallOption) (*CreateSchedulerPoolResponse, error) {
	out := new(CreateSchedulerPoolResponse)
	err := c.client.Call(ctx, "qscheduler.QSchedulerAdmin", "CreateSchedulerPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminPRPCClient) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error) {
	out := new(CreateAccountResponse)
	err := c.client.Call(ctx, "qscheduler.QSchedulerAdmin", "CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminPRPCClient) ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*ListAccountsResponse, error) {
	out := new(ListAccountsResponse)
	err := c.client.Call(ctx, "qscheduler.QSchedulerAdmin", "ListAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminPRPCClient) InspectPool(ctx context.Context, in *InspectPoolRequest, opts ...grpc.CallOption) (*InspectPoolResponse, error) {
	out := new(InspectPoolResponse)
	err := c.client.Call(ctx, "qscheduler.QSchedulerAdmin", "InspectPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type qSchedulerAdminClient struct {
	cc *grpc.ClientConn
}

func NewQSchedulerAdminClient(cc *grpc.ClientConn) QSchedulerAdminClient {
	return &qSchedulerAdminClient{cc}
}

func (c *qSchedulerAdminClient) CreateSchedulerPool(ctx context.Context, in *CreateSchedulerPoolRequest, opts ...grpc.CallOption) (*CreateSchedulerPoolResponse, error) {
	out := new(CreateSchedulerPoolResponse)
	err := c.cc.Invoke(ctx, "/qscheduler.QSchedulerAdmin/CreateSchedulerPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminClient) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error) {
	out := new(CreateAccountResponse)
	err := c.cc.Invoke(ctx, "/qscheduler.QSchedulerAdmin/CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminClient) ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*ListAccountsResponse, error) {
	out := new(ListAccountsResponse)
	err := c.cc.Invoke(ctx, "/qscheduler.QSchedulerAdmin/ListAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *qSchedulerAdminClient) InspectPool(ctx context.Context, in *InspectPoolRequest, opts ...grpc.CallOption) (*InspectPoolResponse, error) {
	out := new(InspectPoolResponse)
	err := c.cc.Invoke(ctx, "/qscheduler.QSchedulerAdmin/InspectPool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QSchedulerAdminServer is the server API for QSchedulerAdmin service.
type QSchedulerAdminServer interface {
	// CreateSchedulerPool creates a scheduler, with the given configuration
	// options.
	CreateSchedulerPool(context.Context, *CreateSchedulerPoolRequest) (*CreateSchedulerPoolResponse, error)
	// CreateAccount creates a quota account within a scheduler, with the
	// given configuration options.
	CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error)
	// ListAccounts returns the set of accounts for a given scheduler.
	ListAccounts(context.Context, *ListAccountsRequest) (*ListAccountsResponse, error)
	// InspectPool returns a description of the state of a scheduler, for debugging
	// or diagnostic purposes.
	InspectPool(context.Context, *InspectPoolRequest) (*InspectPoolResponse, error)
}

func RegisterQSchedulerAdminServer(s prpc.Registrar, srv QSchedulerAdminServer) {
	s.RegisterService(&_QSchedulerAdmin_serviceDesc, srv)
}

func _QSchedulerAdmin_CreateSchedulerPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSchedulerPoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QSchedulerAdminServer).CreateSchedulerPool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qscheduler.QSchedulerAdmin/CreateSchedulerPool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QSchedulerAdminServer).CreateSchedulerPool(ctx, req.(*CreateSchedulerPoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QSchedulerAdmin_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QSchedulerAdminServer).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qscheduler.QSchedulerAdmin/CreateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QSchedulerAdminServer).CreateAccount(ctx, req.(*CreateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QSchedulerAdmin_ListAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QSchedulerAdminServer).ListAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qscheduler.QSchedulerAdmin/ListAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QSchedulerAdminServer).ListAccounts(ctx, req.(*ListAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QSchedulerAdmin_InspectPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InspectPoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QSchedulerAdminServer).InspectPool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qscheduler.QSchedulerAdmin/InspectPool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QSchedulerAdminServer).InspectPool(ctx, req.(*InspectPoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _QSchedulerAdmin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "qscheduler.QSchedulerAdmin",
	HandlerType: (*QSchedulerAdminServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateSchedulerPool",
			Handler:    _QSchedulerAdmin_CreateSchedulerPool_Handler,
		},
		{
			MethodName: "CreateAccount",
			Handler:    _QSchedulerAdmin_CreateAccount_Handler,
		},
		{
			MethodName: "ListAccounts",
			Handler:    _QSchedulerAdmin_ListAccounts_Handler,
		},
		{
			MethodName: "InspectPool",
			Handler:    _QSchedulerAdmin_InspectPool_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/appengine/qscheduler-swarming/api/qscheduler/v1/admin.proto",
}
