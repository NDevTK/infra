// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/unifiedfleet/api/v1/rpc/fleet.proto

package ufspb

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	proto1 "infra/unifiedfleet/api/v1/proto"
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

type ImportChromePlatformsRequest struct {
	// Support importing from local file.
	LocalFilepath        string   `protobuf:"bytes,1,opt,name=local_filepath,json=localFilepath,proto3" json:"local_filepath,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ImportChromePlatformsRequest) Reset()         { *m = ImportChromePlatformsRequest{} }
func (m *ImportChromePlatformsRequest) String() string { return proto.CompactTextString(m) }
func (*ImportChromePlatformsRequest) ProtoMessage()    {}
func (*ImportChromePlatformsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{0}
}

func (m *ImportChromePlatformsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImportChromePlatformsRequest.Unmarshal(m, b)
}
func (m *ImportChromePlatformsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImportChromePlatformsRequest.Marshal(b, m, deterministic)
}
func (m *ImportChromePlatformsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImportChromePlatformsRequest.Merge(m, src)
}
func (m *ImportChromePlatformsRequest) XXX_Size() int {
	return xxx_messageInfo_ImportChromePlatformsRequest.Size(m)
}
func (m *ImportChromePlatformsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ImportChromePlatformsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ImportChromePlatformsRequest proto.InternalMessageInfo

func (m *ImportChromePlatformsRequest) GetLocalFilepath() string {
	if m != nil {
		return m.LocalFilepath
	}
	return ""
}

type ImportChromePlatformsResponse struct {
	Passed               []*ChromePlatformResult `protobuf:"bytes,1,rep,name=passed,proto3" json:"passed,omitempty"`
	Failed               []*ChromePlatformResult `protobuf:"bytes,2,rep,name=failed,proto3" json:"failed,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ImportChromePlatformsResponse) Reset()         { *m = ImportChromePlatformsResponse{} }
func (m *ImportChromePlatformsResponse) String() string { return proto.CompactTextString(m) }
func (*ImportChromePlatformsResponse) ProtoMessage()    {}
func (*ImportChromePlatformsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{1}
}

func (m *ImportChromePlatformsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImportChromePlatformsResponse.Unmarshal(m, b)
}
func (m *ImportChromePlatformsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImportChromePlatformsResponse.Marshal(b, m, deterministic)
}
func (m *ImportChromePlatformsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImportChromePlatformsResponse.Merge(m, src)
}
func (m *ImportChromePlatformsResponse) XXX_Size() int {
	return xxx_messageInfo_ImportChromePlatformsResponse.Size(m)
}
func (m *ImportChromePlatformsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ImportChromePlatformsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ImportChromePlatformsResponse proto.InternalMessageInfo

func (m *ImportChromePlatformsResponse) GetPassed() []*ChromePlatformResult {
	if m != nil {
		return m.Passed
	}
	return nil
}

func (m *ImportChromePlatformsResponse) GetFailed() []*ChromePlatformResult {
	if m != nil {
		return m.Failed
	}
	return nil
}

type ChromePlatformResult struct {
	Platform             *proto1.ChromePlatform `protobuf:"bytes,1,opt,name=platform,proto3" json:"platform,omitempty"`
	ErrorMsg             string                 `protobuf:"bytes,2,opt,name=error_msg,json=errorMsg,proto3" json:"error_msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ChromePlatformResult) Reset()         { *m = ChromePlatformResult{} }
func (m *ChromePlatformResult) String() string { return proto.CompactTextString(m) }
func (*ChromePlatformResult) ProtoMessage()    {}
func (*ChromePlatformResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{2}
}

func (m *ChromePlatformResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChromePlatformResult.Unmarshal(m, b)
}
func (m *ChromePlatformResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChromePlatformResult.Marshal(b, m, deterministic)
}
func (m *ChromePlatformResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChromePlatformResult.Merge(m, src)
}
func (m *ChromePlatformResult) XXX_Size() int {
	return xxx_messageInfo_ChromePlatformResult.Size(m)
}
func (m *ChromePlatformResult) XXX_DiscardUnknown() {
	xxx_messageInfo_ChromePlatformResult.DiscardUnknown(m)
}

var xxx_messageInfo_ChromePlatformResult proto.InternalMessageInfo

func (m *ChromePlatformResult) GetPlatform() *proto1.ChromePlatform {
	if m != nil {
		return m.Platform
	}
	return nil
}

func (m *ChromePlatformResult) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

// Contains the required information for creating a Machine represented in
// the database.
type CreateMachineRequest struct {
	// The machine to create.
	Machine *proto1.Machine `protobuf:"bytes,1,opt,name=machine,proto3" json:"machine,omitempty"`
	// The ID to use for the Machine, which will become the final component of
	// the Machine's resource name.
	//
	// This value should follow the regex "^[a-zA-Z0-9-_]{4,63}$" (4-63 characters,
	// contains only ASCII letters, numbers, dash and underscore.
	MachineId            string   `protobuf:"bytes,2,opt,name=machine_id,json=machineId,proto3" json:"machine_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateMachineRequest) Reset()         { *m = CreateMachineRequest{} }
func (m *CreateMachineRequest) String() string { return proto.CompactTextString(m) }
func (*CreateMachineRequest) ProtoMessage()    {}
func (*CreateMachineRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{3}
}

func (m *CreateMachineRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateMachineRequest.Unmarshal(m, b)
}
func (m *CreateMachineRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateMachineRequest.Marshal(b, m, deterministic)
}
func (m *CreateMachineRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateMachineRequest.Merge(m, src)
}
func (m *CreateMachineRequest) XXX_Size() int {
	return xxx_messageInfo_CreateMachineRequest.Size(m)
}
func (m *CreateMachineRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateMachineRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateMachineRequest proto.InternalMessageInfo

func (m *CreateMachineRequest) GetMachine() *proto1.Machine {
	if m != nil {
		return m.Machine
	}
	return nil
}

func (m *CreateMachineRequest) GetMachineId() string {
	if m != nil {
		return m.MachineId
	}
	return ""
}

type UpdateMachineRequest struct {
	// The machine to update.
	Machine *proto1.Machine `protobuf:"bytes,1,opt,name=machine,proto3" json:"machine,omitempty"`
	// The list of fields to be updated.
	UpdateMask           *field_mask.FieldMask `protobuf:"bytes,2,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *UpdateMachineRequest) Reset()         { *m = UpdateMachineRequest{} }
func (m *UpdateMachineRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateMachineRequest) ProtoMessage()    {}
func (*UpdateMachineRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{4}
}

func (m *UpdateMachineRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateMachineRequest.Unmarshal(m, b)
}
func (m *UpdateMachineRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateMachineRequest.Marshal(b, m, deterministic)
}
func (m *UpdateMachineRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateMachineRequest.Merge(m, src)
}
func (m *UpdateMachineRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateMachineRequest.Size(m)
}
func (m *UpdateMachineRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateMachineRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateMachineRequest proto.InternalMessageInfo

func (m *UpdateMachineRequest) GetMachine() *proto1.Machine {
	if m != nil {
		return m.Machine
	}
	return nil
}

func (m *UpdateMachineRequest) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

type GetMachineRequest struct {
	// The name of the machine to retrieve.
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetMachineRequest) Reset()         { *m = GetMachineRequest{} }
func (m *GetMachineRequest) String() string { return proto.CompactTextString(m) }
func (*GetMachineRequest) ProtoMessage()    {}
func (*GetMachineRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{5}
}

func (m *GetMachineRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetMachineRequest.Unmarshal(m, b)
}
func (m *GetMachineRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetMachineRequest.Marshal(b, m, deterministic)
}
func (m *GetMachineRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetMachineRequest.Merge(m, src)
}
func (m *GetMachineRequest) XXX_Size() int {
	return xxx_messageInfo_GetMachineRequest.Size(m)
}
func (m *GetMachineRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetMachineRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetMachineRequest proto.InternalMessageInfo

func (m *GetMachineRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type ListMachinesRequest struct {
	// The maximum number of machines to return. The service may return fewer than
	// this value.
	// If unspecified, at most 100 machines will be returned.
	// The maximum value is 1000; values above 1000 will be coerced to 1000.
	PageSize int32 `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// A page token, received from a previous `ListMachines` call.
	// Provide this to retrieve the subsequent page.
	//
	// When paginating, all other parameters provided to `ListMachines` must match
	// the call that provided the page token.
	PageToken            string   `protobuf:"bytes,2,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListMachinesRequest) Reset()         { *m = ListMachinesRequest{} }
func (m *ListMachinesRequest) String() string { return proto.CompactTextString(m) }
func (*ListMachinesRequest) ProtoMessage()    {}
func (*ListMachinesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{6}
}

func (m *ListMachinesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListMachinesRequest.Unmarshal(m, b)
}
func (m *ListMachinesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListMachinesRequest.Marshal(b, m, deterministic)
}
func (m *ListMachinesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListMachinesRequest.Merge(m, src)
}
func (m *ListMachinesRequest) XXX_Size() int {
	return xxx_messageInfo_ListMachinesRequest.Size(m)
}
func (m *ListMachinesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListMachinesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListMachinesRequest proto.InternalMessageInfo

func (m *ListMachinesRequest) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *ListMachinesRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

type ListMachinesResponse struct {
	// The machines from datastore.
	Machines []*proto1.Machine `protobuf:"bytes,1,rep,name=machines,proto3" json:"machines,omitempty"`
	// A token, which can be sent as `page_token` to retrieve the next page.
	// If this field is omitted, there are no subsequent pages.
	NextPageToken        string   `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListMachinesResponse) Reset()         { *m = ListMachinesResponse{} }
func (m *ListMachinesResponse) String() string { return proto.CompactTextString(m) }
func (*ListMachinesResponse) ProtoMessage()    {}
func (*ListMachinesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{7}
}

func (m *ListMachinesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListMachinesResponse.Unmarshal(m, b)
}
func (m *ListMachinesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListMachinesResponse.Marshal(b, m, deterministic)
}
func (m *ListMachinesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListMachinesResponse.Merge(m, src)
}
func (m *ListMachinesResponse) XXX_Size() int {
	return xxx_messageInfo_ListMachinesResponse.Size(m)
}
func (m *ListMachinesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListMachinesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListMachinesResponse proto.InternalMessageInfo

func (m *ListMachinesResponse) GetMachines() []*proto1.Machine {
	if m != nil {
		return m.Machines
	}
	return nil
}

func (m *ListMachinesResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

type DeleteMachineRequest struct {
	// The name of the Machine to delete
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteMachineRequest) Reset()         { *m = DeleteMachineRequest{} }
func (m *DeleteMachineRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteMachineRequest) ProtoMessage()    {}
func (*DeleteMachineRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfc37a625f56a717, []int{8}
}

func (m *DeleteMachineRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteMachineRequest.Unmarshal(m, b)
}
func (m *DeleteMachineRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteMachineRequest.Marshal(b, m, deterministic)
}
func (m *DeleteMachineRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteMachineRequest.Merge(m, src)
}
func (m *DeleteMachineRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteMachineRequest.Size(m)
}
func (m *DeleteMachineRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteMachineRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteMachineRequest proto.InternalMessageInfo

func (m *DeleteMachineRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*ImportChromePlatformsRequest)(nil), "unifiedfleet.api.v1.rpc.ImportChromePlatformsRequest")
	proto.RegisterType((*ImportChromePlatformsResponse)(nil), "unifiedfleet.api.v1.rpc.ImportChromePlatformsResponse")
	proto.RegisterType((*ChromePlatformResult)(nil), "unifiedfleet.api.v1.rpc.ChromePlatformResult")
	proto.RegisterType((*CreateMachineRequest)(nil), "unifiedfleet.api.v1.rpc.CreateMachineRequest")
	proto.RegisterType((*UpdateMachineRequest)(nil), "unifiedfleet.api.v1.rpc.UpdateMachineRequest")
	proto.RegisterType((*GetMachineRequest)(nil), "unifiedfleet.api.v1.rpc.GetMachineRequest")
	proto.RegisterType((*ListMachinesRequest)(nil), "unifiedfleet.api.v1.rpc.ListMachinesRequest")
	proto.RegisterType((*ListMachinesResponse)(nil), "unifiedfleet.api.v1.rpc.ListMachinesResponse")
	proto.RegisterType((*DeleteMachineRequest)(nil), "unifiedfleet.api.v1.rpc.DeleteMachineRequest")
}

func init() {
	proto.RegisterFile("infra/unifiedfleet/api/v1/rpc/fleet.proto", fileDescriptor_bfc37a625f56a717)
}

var fileDescriptor_bfc37a625f56a717 = []byte{
	// 708 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x55, 0xef, 0x6a, 0x13, 0x4f,
	0x14, 0x25, 0xe9, 0x9f, 0x5f, 0x72, 0xf3, 0x8b, 0xe2, 0x1a, 0x35, 0xa4, 0x16, 0xca, 0x4a, 0xa5,
	0x2d, 0x66, 0xd7, 0x46, 0x2a, 0x48, 0x41, 0x49, 0xdb, 0x54, 0x0a, 0x16, 0xea, 0xfa, 0x07, 0x94,
	0x42, 0x98, 0x6c, 0x66, 0x37, 0x43, 0x76, 0x33, 0xe3, 0xcc, 0x6c, 0x69, 0xfb, 0xc1, 0x8f, 0x3e,
	0x82, 0x6f, 0xe1, 0x3b, 0xf5, 0x39, 0xf4, 0x8b, 0xec, 0xec, 0x6c, 0x6c, 0x9a, 0x4d, 0x4c, 0x45,
	0xbf, 0x2d, 0x77, 0xee, 0x39, 0xe7, 0xee, 0x99, 0x7b, 0x76, 0x61, 0x9d, 0x0c, 0x3c, 0x8e, 0xec,
	0x68, 0x40, 0x3c, 0x82, 0xbb, 0x5e, 0x80, 0xb1, 0xb4, 0x11, 0x23, 0xf6, 0xc9, 0xa6, 0xcd, 0x99,
	0x6b, 0xab, 0x82, 0xc5, 0x38, 0x95, 0xd4, 0xb8, 0x77, 0xb9, 0xc9, 0x42, 0x8c, 0x58, 0x27, 0x9b,
	0x16, 0x67, 0x6e, 0x6d, 0xc9, 0xa7, 0xd4, 0x0f, 0xb0, 0xad, 0xda, 0x3a, 0x91, 0x67, 0xe3, 0x90,
	0xc9, 0xb3, 0x04, 0x55, 0x5b, 0xb9, 0x7a, 0xe8, 0x11, 0x1c, 0x74, 0xdb, 0x21, 0x12, 0x7d, 0xdd,
	0xf1, 0xc2, 0xa7, 0x96, 0xdb, 0xe3, 0x34, 0x24, 0x51, 0x68, 0x51, 0xee, 0xdb, 0x41, 0xe4, 0x12,
	0xdb, 0x8f, 0xe5, 0x55, 0x83, 0xad, 0x19, 0xe2, 0xb1, 0x12, 0x70, 0x07, 0xf7, 0xd0, 0x09, 0xa1,
	0x5c, 0x13, 0x3c, 0xbb, 0x06, 0x01, 0xc7, 0x82, 0x46, 0xdc, 0xc5, 0x1a, 0x5a, 0x9f, 0xfc, 0xfa,
	0x09, 0x34, 0x44, 0x6e, 0x8f, 0x0c, 0xd2, 0xf6, 0xad, 0xdf, 0xb5, 0xab, 0x31, 0x70, 0x9b, 0x05,
	0x48, 0x7a, 0x94, 0x87, 0x09, 0xcc, 0x6c, 0xc1, 0xfd, 0x83, 0x90, 0x51, 0x2e, 0x77, 0xd5, 0xf1,
	0x91, 0x3e, 0x15, 0x0e, 0xfe, 0x14, 0x61, 0x21, 0x8d, 0x55, 0xb8, 0x11, 0x50, 0x17, 0x05, 0x6d,
	0x8f, 0x04, 0x98, 0x21, 0xd9, 0xab, 0xe6, 0x56, 0x72, 0x6b, 0x45, 0xa7, 0xac, 0xaa, 0xfb, 0xba,
	0x68, 0x7e, 0xcb, 0xc1, 0xf2, 0x04, 0x1e, 0xc1, 0xe8, 0x40, 0x60, 0xa3, 0x05, 0x8b, 0x0c, 0x09,
	0x81, 0xbb, 0xd5, 0xdc, 0xca, 0xdc, 0x5a, 0xa9, 0x51, 0xb7, 0x26, 0xdc, 0x99, 0x35, 0xca, 0xe0,
	0x60, 0x11, 0x05, 0xd2, 0xd1, 0xe0, 0x98, 0xc6, 0x43, 0x24, 0xc0, 0xdd, 0x6a, 0xfe, 0x8f, 0x68,
	0x12, 0xb0, 0x79, 0x0e, 0x95, 0xac, 0x73, 0xa3, 0x05, 0x85, 0xd4, 0x20, 0xf5, 0xa2, 0xa5, 0xc6,
	0x7a, 0xa6, 0x80, 0x32, 0xef, 0xaa, 0xc4, 0x10, 0x6a, 0x2c, 0x41, 0x11, 0x73, 0x4e, 0x79, 0x3b,
	0x14, 0x7e, 0x35, 0xaf, 0x0c, 0x2b, 0xa8, 0xc2, 0xa1, 0xf0, 0xcd, 0x53, 0xa8, 0xec, 0x72, 0x8c,
	0x24, 0x3e, 0x4c, 0x2e, 0x30, 0xb5, 0xba, 0x09, 0xff, 0xe9, 0x2b, 0xd5, 0xd2, 0xe6, 0x14, 0x69,
	0x8d, 0xdd, 0x99, 0xbb, 0x68, 0xe6, 0x9d, 0x14, 0x67, 0x2c, 0x03, 0xe8, 0xc7, 0x36, 0xe9, 0x6a,
	0xe1, 0xa2, 0xae, 0x1c, 0x74, 0xcd, 0xaf, 0x39, 0xa8, 0xbc, 0x63, 0xdd, 0x7f, 0x22, 0xbd, 0x0d,
	0xa5, 0x48, 0x51, 0xab, 0xfc, 0x28, 0xed, 0x52, 0xa3, 0x66, 0x25, 0xfb, 0x6d, 0xa5, 0x11, 0xb3,
	0xf6, 0xe3, 0x94, 0x1c, 0x22, 0xd1, 0x77, 0x20, 0xd2, 0x93, 0x88, 0xbe, 0xf9, 0x01, 0x6e, 0xbd,
	0xc4, 0xf2, 0xca, 0x50, 0x7b, 0x30, 0x3f, 0x40, 0x61, 0x32, 0x51, 0x71, 0xe7, 0xf1, 0x45, 0x33,
	0xff, 0xbd, 0xb9, 0x01, 0x6b, 0x7a, 0xb0, 0xba, 0x9a, 0xac, 0x2e, 0xce, 0x84, 0xc4, 0xa1, 0x85,
	0x18, 0x13, 0x8c, 0x4a, 0xcb, 0xa5, 0xa1, 0x9d, 0xd2, 0x28, 0xb4, 0xf9, 0x1a, 0x6e, 0xbf, 0x22,
	0x22, 0xe5, 0x1e, 0xee, 0xf5, 0x12, 0x14, 0x19, 0xf2, 0x71, 0x5b, 0x90, 0xf3, 0x44, 0x61, 0xc1,
	0x29, 0xc4, 0x85, 0x37, 0xe4, 0x5c, 0xd9, 0xa8, 0x0e, 0x25, 0xed, 0xe3, 0x41, 0x6a, 0x63, 0x5c,
	0x79, 0x1b, 0x17, 0xcc, 0xcf, 0x50, 0x19, 0xa5, 0xd4, 0x2b, 0xfe, 0x1c, 0x0a, 0xda, 0x0d, 0xa1,
	0x97, 0x7c, 0x06, 0x1b, 0x9d, 0x21, 0xc6, 0x78, 0x08, 0x37, 0x07, 0xf8, 0x54, 0xb6, 0xc7, 0xb4,
	0xcb, 0x71, 0xf9, 0x68, 0xa8, 0x7f, 0x0c, 0x95, 0x3d, 0x1c, 0xe0, 0xb1, 0x5b, 0xfc, 0x2b, 0x86,
	0x35, 0x7e, 0xcc, 0xc3, 0xc2, 0x7e, 0xdc, 0x6a, 0x7c, 0xc9, 0xc1, 0x9d, 0xcc, 0x50, 0x1b, 0x5b,
	0x13, 0x53, 0x37, 0xed, 0x63, 0x52, 0x7b, 0x7a, 0x5d, 0x98, 0x36, 0xb6, 0x03, 0xe5, 0x91, 0xc4,
	0x18, 0x53, 0x52, 0x9f, 0x91, 0xac, 0xda, 0x0c, 0xd7, 0x10, 0x6b, 0x8c, 0x44, 0x63, 0x8a, 0x46,
	0x56, 0x84, 0x66, 0xd2, 0x38, 0x06, 0xf8, 0xb5, 0xe6, 0xc6, 0xc6, 0x44, 0x81, 0xb1, 0x2c, 0xcc,
	0xc4, 0xde, 0x87, 0xff, 0x2f, 0xaf, 0xa5, 0xf1, 0x68, 0x22, 0x7f, 0x46, 0x20, 0x6a, 0xf5, 0x19,
	0xbb, 0xf5, 0x95, 0xbc, 0x87, 0xf2, 0xc8, 0x0e, 0x4e, 0xb1, 0x2b, 0x6b, 0x57, 0x6b, 0x77, 0xc7,
	0xbe, 0x0c, 0xad, 0xf8, 0xcf, 0xbc, 0xb3, 0xfa, 0xf1, 0xc1, 0xd4, 0xdf, 0xfe, 0x76, 0xe4, 0x09,
	0xd6, 0xe9, 0x2c, 0x2a, 0xd8, 0x93, 0x9f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x05, 0x5b, 0x71, 0x1e,
	0x24, 0x08, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// FleetClient is the client API for Fleet service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FleetClient interface {
	// ImportChromePlatforms imports chrome platforms.
	ImportChromePlatforms(ctx context.Context, in *ImportChromePlatformsRequest, opts ...grpc.CallOption) (*ImportChromePlatformsResponse, error)
	// CreateMachine creates a new machine.
	CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error)
	// Update updates the machine
	UpdateMachine(ctx context.Context, in *UpdateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error)
	// Get retrieves the details of the machine
	GetMachine(ctx context.Context, in *GetMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error)
	// List gets all the machines
	ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error)
	// Delete delete the machine
	DeleteMachine(ctx context.Context, in *DeleteMachineRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}
type fleetPRPCClient struct {
	client *prpc.Client
}

func NewFleetPRPCClient(client *prpc.Client) FleetClient {
	return &fleetPRPCClient{client}
}

func (c *fleetPRPCClient) ImportChromePlatforms(ctx context.Context, in *ImportChromePlatformsRequest, opts ...grpc.CallOption) (*ImportChromePlatformsResponse, error) {
	out := new(ImportChromePlatformsResponse)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "ImportChromePlatforms", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetPRPCClient) CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "CreateMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetPRPCClient) UpdateMachine(ctx context.Context, in *UpdateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "UpdateMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetPRPCClient) GetMachine(ctx context.Context, in *GetMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "GetMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetPRPCClient) ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error) {
	out := new(ListMachinesResponse)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "ListMachines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetPRPCClient) DeleteMachine(ctx context.Context, in *DeleteMachineRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.client.Call(ctx, "unifiedfleet.api.v1.rpc.Fleet", "DeleteMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type fleetClient struct {
	cc grpc.ClientConnInterface
}

func NewFleetClient(cc grpc.ClientConnInterface) FleetClient {
	return &fleetClient{cc}
}

func (c *fleetClient) ImportChromePlatforms(ctx context.Context, in *ImportChromePlatformsRequest, opts ...grpc.CallOption) (*ImportChromePlatformsResponse, error) {
	out := new(ImportChromePlatformsResponse)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/ImportChromePlatforms", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetClient) CreateMachine(ctx context.Context, in *CreateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/CreateMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetClient) UpdateMachine(ctx context.Context, in *UpdateMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/UpdateMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetClient) GetMachine(ctx context.Context, in *GetMachineRequest, opts ...grpc.CallOption) (*proto1.Machine, error) {
	out := new(proto1.Machine)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/GetMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetClient) ListMachines(ctx context.Context, in *ListMachinesRequest, opts ...grpc.CallOption) (*ListMachinesResponse, error) {
	out := new(ListMachinesResponse)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/ListMachines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetClient) DeleteMachine(ctx context.Context, in *DeleteMachineRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/unifiedfleet.api.v1.rpc.Fleet/DeleteMachine", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FleetServer is the server API for Fleet service.
type FleetServer interface {
	// ImportChromePlatforms imports chrome platforms.
	ImportChromePlatforms(context.Context, *ImportChromePlatformsRequest) (*ImportChromePlatformsResponse, error)
	// CreateMachine creates a new machine.
	CreateMachine(context.Context, *CreateMachineRequest) (*proto1.Machine, error)
	// Update updates the machine
	UpdateMachine(context.Context, *UpdateMachineRequest) (*proto1.Machine, error)
	// Get retrieves the details of the machine
	GetMachine(context.Context, *GetMachineRequest) (*proto1.Machine, error)
	// List gets all the machines
	ListMachines(context.Context, *ListMachinesRequest) (*ListMachinesResponse, error)
	// Delete delete the machine
	DeleteMachine(context.Context, *DeleteMachineRequest) (*empty.Empty, error)
}

// UnimplementedFleetServer can be embedded to have forward compatible implementations.
type UnimplementedFleetServer struct {
}

func (*UnimplementedFleetServer) ImportChromePlatforms(ctx context.Context, req *ImportChromePlatformsRequest) (*ImportChromePlatformsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportChromePlatforms not implemented")
}
func (*UnimplementedFleetServer) CreateMachine(ctx context.Context, req *CreateMachineRequest) (*proto1.Machine, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateMachine not implemented")
}
func (*UnimplementedFleetServer) UpdateMachine(ctx context.Context, req *UpdateMachineRequest) (*proto1.Machine, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMachine not implemented")
}
func (*UnimplementedFleetServer) GetMachine(ctx context.Context, req *GetMachineRequest) (*proto1.Machine, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMachine not implemented")
}
func (*UnimplementedFleetServer) ListMachines(ctx context.Context, req *ListMachinesRequest) (*ListMachinesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMachines not implemented")
}
func (*UnimplementedFleetServer) DeleteMachine(ctx context.Context, req *DeleteMachineRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteMachine not implemented")
}

func RegisterFleetServer(s prpc.Registrar, srv FleetServer) {
	s.RegisterService(&_Fleet_serviceDesc, srv)
}

func _Fleet_ImportChromePlatforms_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportChromePlatformsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).ImportChromePlatforms(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/ImportChromePlatforms",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).ImportChromePlatforms(ctx, req.(*ImportChromePlatformsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fleet_CreateMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateMachineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).CreateMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/CreateMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).CreateMachine(ctx, req.(*CreateMachineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fleet_UpdateMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMachineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).UpdateMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/UpdateMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).UpdateMachine(ctx, req.(*UpdateMachineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fleet_GetMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMachineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).GetMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/GetMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).GetMachine(ctx, req.(*GetMachineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fleet_ListMachines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMachinesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).ListMachines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/ListMachines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).ListMachines(ctx, req.(*ListMachinesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fleet_DeleteMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteMachineRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetServer).DeleteMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/unifiedfleet.api.v1.rpc.Fleet/DeleteMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetServer).DeleteMachine(ctx, req.(*DeleteMachineRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Fleet_serviceDesc = grpc.ServiceDesc{
	ServiceName: "unifiedfleet.api.v1.rpc.Fleet",
	HandlerType: (*FleetServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ImportChromePlatforms",
			Handler:    _Fleet_ImportChromePlatforms_Handler,
		},
		{
			MethodName: "CreateMachine",
			Handler:    _Fleet_CreateMachine_Handler,
		},
		{
			MethodName: "UpdateMachine",
			Handler:    _Fleet_UpdateMachine_Handler,
		},
		{
			MethodName: "GetMachine",
			Handler:    _Fleet_GetMachine_Handler,
		},
		{
			MethodName: "ListMachines",
			Handler:    _Fleet_ListMachines_Handler,
		},
		{
			MethodName: "DeleteMachine",
			Handler:    _Fleet_DeleteMachine_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/unifiedfleet/api/v1/rpc/fleet.proto",
}
