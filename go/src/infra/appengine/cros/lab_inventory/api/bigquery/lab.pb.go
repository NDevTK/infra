// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/cros/lab_inventory/api/bigquery/lab.proto

package apibq

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	lab "go.chromium.org/chromiumos/infra/proto/go/lab"
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

type LabInventory struct {
	Id                   string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Hostname             string               `protobuf:"bytes,2,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Dut                  *lab.DeviceUnderTest `protobuf:"bytes,3,opt,name=dut,proto3" json:"dut,omitempty"`
	UpdatedTime          *timestamp.Timestamp `protobuf:"bytes,4,opt,name=updated_time,json=updatedTime,proto3" json:"updated_time,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *LabInventory) Reset()         { *m = LabInventory{} }
func (m *LabInventory) String() string { return proto.CompactTextString(m) }
func (*LabInventory) ProtoMessage()    {}
func (*LabInventory) Descriptor() ([]byte, []int) {
	return fileDescriptor_fbc6db176d7c629b, []int{0}
}

func (m *LabInventory) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LabInventory.Unmarshal(m, b)
}
func (m *LabInventory) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LabInventory.Marshal(b, m, deterministic)
}
func (m *LabInventory) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LabInventory.Merge(m, src)
}
func (m *LabInventory) XXX_Size() int {
	return xxx_messageInfo_LabInventory.Size(m)
}
func (m *LabInventory) XXX_DiscardUnknown() {
	xxx_messageInfo_LabInventory.DiscardUnknown(m)
}

var xxx_messageInfo_LabInventory proto.InternalMessageInfo

func (m *LabInventory) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *LabInventory) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *LabInventory) GetDut() *lab.DeviceUnderTest {
	if m != nil {
		return m.Dut
	}
	return nil
}

func (m *LabInventory) GetUpdatedTime() *timestamp.Timestamp {
	if m != nil {
		return m.UpdatedTime
	}
	return nil
}

func init() {
	proto.RegisterType((*LabInventory)(nil), "apibq.LabInventory")
}

func init() {
	proto.RegisterFile("infra/appengine/cros/lab_inventory/api/bigquery/lab.proto", fileDescriptor_fbc6db176d7c629b)
}

var fileDescriptor_fbc6db176d7c629b = []byte{
	// 257 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8f, 0xb1, 0x4e, 0x33, 0x31,
	0x10, 0x84, 0x75, 0xc9, 0xff, 0x23, 0x70, 0x22, 0x8a, 0x13, 0xc5, 0xe9, 0x1a, 0x22, 0x0a, 0x94,
	0xca, 0x2b, 0x41, 0x85, 0x10, 0x1d, 0x0d, 0x12, 0x55, 0x14, 0xea, 0xc8, 0x3e, 0x6f, 0x9c, 0x95,
	0x72, 0x5e, 0xc7, 0xf6, 0x45, 0xca, 0xe3, 0xf0, 0xa6, 0xc8, 0x3e, 0x8e, 0x72, 0xc7, 0xdf, 0x8c,
	0x67, 0xc4, 0x0b, 0xb9, 0x7d, 0x50, 0xa0, 0xbc, 0x47, 0x67, 0xc9, 0x21, 0x74, 0x81, 0x23, 0x1c,
	0x95, 0xde, 0x91, 0x3b, 0xa3, 0x4b, 0x1c, 0x2e, 0xa0, 0x3c, 0x81, 0x26, 0x7b, 0x1a, 0x30, 0x5c,
	0xf2, 0x93, 0xf4, 0x81, 0x13, 0xd7, 0xff, 0x95, 0x27, 0x7d, 0x6a, 0xef, 0x2d, 0xb3, 0x3d, 0x22,
	0x14, 0x51, 0x0f, 0x7b, 0x48, 0xd4, 0x63, 0x4c, 0xaa, 0xf7, 0x23, 0xd7, 0xbe, 0x5a, 0x96, 0xdd,
	0x21, 0x70, 0x4f, 0x43, 0x2f, 0x39, 0x58, 0x98, 0x0e, 0x8e, 0x30, 0xfe, 0x5e, 0x38, 0x88, 0xa1,
	0xcb, 0xe9, 0x60, 0xf0, 0x4c, 0x1d, 0x8e, 0xe6, 0x87, 0xef, 0x4a, 0x2c, 0x3f, 0x95, 0xfe, 0x98,
	0xca, 0xd4, 0xb7, 0x62, 0x46, 0xa6, 0xa9, 0x56, 0xd5, 0xfa, 0x66, 0x33, 0x23, 0x53, 0xb7, 0xe2,
	0xfa, 0xc0, 0x31, 0x39, 0xd5, 0x63, 0x33, 0x2b, 0xea, 0xdf, 0x5d, 0x3f, 0x8a, 0xb9, 0x19, 0x52,
	0x33, 0x5f, 0x55, 0xeb, 0xc5, 0xd3, 0x9d, 0xcc, 0xd5, 0xdf, 0x4b, 0xf8, 0x97, 0x33, 0x18, 0xb6,
	0x18, 0xd3, 0x26, 0x03, 0xf5, 0x9b, 0x58, 0x0e, 0xde, 0xa8, 0x84, 0x66, 0x97, 0xcb, 0x37, 0xff,
	0x8a, 0xa1, 0x95, 0xe3, 0x32, 0x39, 0x2d, 0x93, 0xdb, 0x69, 0xd9, 0x66, 0xf1, 0xcb, 0x67, 0x45,
	0x5f, 0x15, 0xe0, 0xf9, 0x27, 0x00, 0x00, 0xff, 0xff, 0xaa, 0x94, 0x35, 0x8f, 0x4c, 0x01, 0x00,
	0x00,
}
