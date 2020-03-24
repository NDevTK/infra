// Code generated by protoc-gen-go. DO NOT EDIT.
// source: network.proto

package fleet

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type Vlan struct {
	Id                   *VlanID  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	VlanAddress          string   `protobuf:"bytes,2,opt,name=vlan_address,json=vlanAddress,proto3" json:"vlan_address,omitempty"`
	Capacity             int32    `protobuf:"varint,3,opt,name=capacity,proto3" json:"capacity,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Vlan) Reset()         { *m = Vlan{} }
func (m *Vlan) String() string { return proto.CompactTextString(m) }
func (*Vlan) ProtoMessage()    {}
func (*Vlan) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{0}
}

func (m *Vlan) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Vlan.Unmarshal(m, b)
}
func (m *Vlan) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Vlan.Marshal(b, m, deterministic)
}
func (m *Vlan) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Vlan.Merge(m, src)
}
func (m *Vlan) XXX_Size() int {
	return xxx_messageInfo_Vlan.Size(m)
}
func (m *Vlan) XXX_DiscardUnknown() {
	xxx_messageInfo_Vlan.DiscardUnknown(m)
}

var xxx_messageInfo_Vlan proto.InternalMessageInfo

func (m *Vlan) GetId() *VlanID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Vlan) GetVlanAddress() string {
	if m != nil {
		return m.VlanAddress
	}
	return ""
}

func (m *Vlan) GetCapacity() int32 {
	if m != nil {
		return m.Capacity
	}
	return 0
}

func init() {
	proto.RegisterType((*Vlan)(nil), "fleet.Vlan")
}

func init() { proto.RegisterFile("network.proto", fileDescriptor_8571034d60397816) }

var fileDescriptor_8571034d60397816 = []byte{
	// 151 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0x4b, 0x2d, 0x29,
	0xcf, 0x2f, 0xca, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4d, 0xcb, 0x49, 0x4d, 0x2d,
	0x91, 0x12, 0x80, 0x8a, 0xc6, 0x67, 0xa6, 0x40, 0x24, 0x94, 0x52, 0xb8, 0x58, 0xc2, 0x72, 0x12,
	0xf3, 0x84, 0x64, 0xb9, 0x98, 0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0xb8, 0x8d, 0x78, 0xf5,
	0xc0, 0xaa, 0xf5, 0x40, 0x12, 0x9e, 0x2e, 0x41, 0x4c, 0x99, 0x29, 0x42, 0x8a, 0x5c, 0x3c, 0x65,
	0x39, 0x89, 0x79, 0xf1, 0x89, 0x29, 0x29, 0x45, 0xa9, 0xc5, 0xc5, 0x12, 0x4c, 0x0a, 0x8c, 0x1a,
	0x9c, 0x41, 0xdc, 0x20, 0x31, 0x47, 0x88, 0x90, 0x90, 0x14, 0x17, 0x47, 0x72, 0x62, 0x41, 0x62,
	0x72, 0x66, 0x49, 0xa5, 0x04, 0xb3, 0x02, 0xa3, 0x06, 0x6b, 0x10, 0x9c, 0xef, 0xc4, 0x19, 0xc5,
	0xae, 0x67, 0x0d, 0x36, 0x34, 0x89, 0x0d, 0x6c, 0xaf, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x3d,
	0xd8, 0x5f, 0xde, 0xa1, 0x00, 0x00, 0x00,
}
