// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/unifiedfleet/api/v1/proto/network.proto

package ufspb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type Nic struct {
	// Unique serial_number or asset tag
	// The format will be nics/XXX
	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	MacAddress string `protobuf:"bytes,2,opt,name=mac_address,json=macAddress,proto3" json:"mac_address,omitempty"`
	// Record the last update timestamp of this machine (In UTC timezone)
	UpdateTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	// Refers to machine name
	Machine              string   `protobuf:"bytes,4,opt,name=machine,proto3" json:"machine,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nic) Reset()         { *m = Nic{} }
func (m *Nic) String() string { return proto.CompactTextString(m) }
func (*Nic) ProtoMessage()    {}
func (*Nic) Descriptor() ([]byte, []int) {
	return fileDescriptor_05c66b9144f80972, []int{0}
}

func (m *Nic) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nic.Unmarshal(m, b)
}
func (m *Nic) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nic.Marshal(b, m, deterministic)
}
func (m *Nic) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nic.Merge(m, src)
}
func (m *Nic) XXX_Size() int {
	return xxx_messageInfo_Nic.Size(m)
}
func (m *Nic) XXX_DiscardUnknown() {
	xxx_messageInfo_Nic.DiscardUnknown(m)
}

var xxx_messageInfo_Nic proto.InternalMessageInfo

func (m *Nic) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Nic) GetMacAddress() string {
	if m != nil {
		return m.MacAddress
	}
	return ""
}

func (m *Nic) GetUpdateTime() *timestamp.Timestamp {
	if m != nil {
		return m.UpdateTime
	}
	return nil
}

func (m *Nic) GetMachine() string {
	if m != nil {
		return m.Machine
	}
	return ""
}

func init() {
	proto.RegisterType((*Nic)(nil), "unifiedfleet.api.v1.proto.Nic")
}

func init() {
	proto.RegisterFile("infra/unifiedfleet/api/v1/proto/network.proto", fileDescriptor_05c66b9144f80972)
}

var fileDescriptor_05c66b9144f80972 = []byte{
	// 330 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0x3f, 0x4f, 0xc3, 0x30,
	0x10, 0xc5, 0x15, 0x52, 0x81, 0x70, 0xb7, 0x4c, 0xa1, 0x4b, 0x2b, 0x04, 0xa2, 0x42, 0x8a, 0xad,
	0x16, 0x31, 0x00, 0x03, 0xa4, 0x0b, 0x13, 0x0c, 0x15, 0x13, 0x4b, 0xe5, 0x38, 0x97, 0xf4, 0x44,
	0x1c, 0x5b, 0xb6, 0x53, 0x84, 0x10, 0xdf, 0x95, 0x99, 0x8f, 0x80, 0x18, 0x50, 0xe2, 0x54, 0x62,
	0xab, 0xd8, 0x7c, 0x7f, 0x7e, 0xef, 0xe9, 0x9d, 0x49, 0x82, 0x75, 0x61, 0x38, 0x6b, 0x6a, 0x2c,
	0x10, 0xf2, 0xa2, 0x02, 0x70, 0x8c, 0x6b, 0x64, 0x9b, 0x19, 0xd3, 0x46, 0x39, 0xc5, 0x6a, 0x70,
	0xaf, 0xca, 0xbc, 0xd0, 0xae, 0x8a, 0x8e, 0xfe, 0x2e, 0x52, 0xae, 0x91, 0x6e, 0x66, 0x7e, 0x34,
	0x1a, 0x97, 0x4a, 0x95, 0x15, 0x78, 0x2c, 0x6b, 0x0a, 0xe6, 0x50, 0x82, 0x75, 0x5c, 0xea, 0x7e,
	0xe1, 0xb6, 0x54, 0x54, 0xac, 0x8d, 0x92, 0xd8, 0x48, 0xaa, 0x4c, 0xc9, 0xaa, 0x46, 0x20, 0x2b,
	0x8d, 0x16, 0xbd, 0x55, 0x2f, 0xd0, 0xda, 0x17, 0x08, 0x55, 0xbe, 0xca, 0x60, 0xcd, 0x37, 0xa8,
	0x4c, 0x2f, 0x70, 0xf5, 0x0f, 0x01, 0x03, 0x56, 0x35, 0x46, 0x80, 0x47, 0x8f, 0x7f, 0x02, 0x12,
	0x3e, 0xa2, 0x88, 0x22, 0x32, 0xa8, 0xb9, 0x84, 0x38, 0x98, 0x04, 0xd3, 0xc3, 0x65, 0xf7, 0x8e,
	0xc6, 0x64, 0x28, 0xb9, 0x58, 0xf1, 0x3c, 0x37, 0x60, 0x6d, 0xbc, 0xd7, 0x8d, 0x88, 0xe4, 0x22,
	0xf5, 0x9d, 0xe8, 0x8e, 0x0c, 0x1b, 0x9d, 0x73, 0x07, 0xab, 0x36, 0x52, 0x1c, 0x4e, 0x82, 0xe9,
	0x70, 0x3e, 0xa2, 0xde, 0x8d, 0x6e, 0xf3, 0xd2, 0xa7, 0x6d, 0xde, 0x45, 0xf8, 0x99, 0x86, 0x4b,
	0xe2, 0x99, 0xb6, 0x1b, 0xdd, 0x93, 0x03, 0xc9, 0xc5, 0x1a, 0x6b, 0x88, 0x07, 0xad, 0xfc, 0x22,
	0xf9, 0x4e, 0xcf, 0xc9, 0xb4, 0x3f, 0x66, 0xd2, 0x5d, 0x33, 0xb1, 0x6f, 0xd6, 0x81, 0xa4, 0x5c,
	0x6b, 0xab, 0x95, 0xa3, 0x42, 0x49, 0xf6, 0xe0, 0xa1, 0xe5, 0x96, 0xbe, 0xbe, 0xfc, 0x4a, 0xe7,
	0xe4, 0x64, 0x27, 0xd7, 0x46, 0x25, 0x35, 0x0a, 0xcb, 0xde, 0x6b, 0x14, 0x1f, 0x8b, 0xb3, 0xe7,
	0xd3, 0x1d, 0xff, 0x7c, 0xd3, 0x14, 0x56, 0x67, 0xd9, 0x7e, 0x57, 0x5c, 0xfc, 0x06, 0x00, 0x00,
	0xff, 0xff, 0x92, 0x00, 0x5e, 0x4f, 0x17, 0x02, 0x00, 0x00,
}
