// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/unifiedfleet/api/v1/proto/rack.proto

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

// Rack refers to the racks which are placed in
// Chrome Browser lab and Chrome OS lab. Machines and Pheripherals
// are placed in the Racks.
type Rack struct {
	// Unique (fake probably) asset tag
	// The format will be racks/XXX
	Name     string    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Location *Location `protobuf:"bytes,2,opt,name=location,proto3" json:"location,omitempty"`
	// Indicates the Rack Unit capacity of the rack.
	CapacityRu int32 `protobuf:"varint,3,opt,name=capacity_ru,json=capacityRu,proto3" json:"capacity_ru,omitempty"`
	// Types that are valid to be assigned to Rack:
	//	*Rack_ChromeBrowserRack
	//	*Rack_ChromeosRack
	Rack isRack_Rack `protobuf_oneof:"rack"`
	// Record the last update timestamp of this Rack (In UTC timezone)
	UpdateTime *timestamp.Timestamp `protobuf:"bytes,6,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	// Record the ACL info of the rack
	Realm                string   `protobuf:"bytes,7,opt,name=realm,proto3" json:"realm,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Rack) Reset()         { *m = Rack{} }
func (m *Rack) String() string { return proto.CompactTextString(m) }
func (*Rack) ProtoMessage()    {}
func (*Rack) Descriptor() ([]byte, []int) {
	return fileDescriptor_4efc02fbdf306edb, []int{0}
}

func (m *Rack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rack.Unmarshal(m, b)
}
func (m *Rack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rack.Marshal(b, m, deterministic)
}
func (m *Rack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rack.Merge(m, src)
}
func (m *Rack) XXX_Size() int {
	return xxx_messageInfo_Rack.Size(m)
}
func (m *Rack) XXX_DiscardUnknown() {
	xxx_messageInfo_Rack.DiscardUnknown(m)
}

var xxx_messageInfo_Rack proto.InternalMessageInfo

func (m *Rack) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Rack) GetLocation() *Location {
	if m != nil {
		return m.Location
	}
	return nil
}

func (m *Rack) GetCapacityRu() int32 {
	if m != nil {
		return m.CapacityRu
	}
	return 0
}

type isRack_Rack interface {
	isRack_Rack()
}

type Rack_ChromeBrowserRack struct {
	ChromeBrowserRack *ChromeBrowserRack `protobuf:"bytes,4,opt,name=chrome_browser_rack,json=chromeBrowserRack,proto3,oneof"`
}

type Rack_ChromeosRack struct {
	ChromeosRack *ChromeOSRack `protobuf:"bytes,5,opt,name=chromeos_rack,json=chromeosRack,proto3,oneof"`
}

func (*Rack_ChromeBrowserRack) isRack_Rack() {}

func (*Rack_ChromeosRack) isRack_Rack() {}

func (m *Rack) GetRack() isRack_Rack {
	if m != nil {
		return m.Rack
	}
	return nil
}

func (m *Rack) GetChromeBrowserRack() *ChromeBrowserRack {
	if x, ok := m.GetRack().(*Rack_ChromeBrowserRack); ok {
		return x.ChromeBrowserRack
	}
	return nil
}

func (m *Rack) GetChromeosRack() *ChromeOSRack {
	if x, ok := m.GetRack().(*Rack_ChromeosRack); ok {
		return x.ChromeosRack
	}
	return nil
}

func (m *Rack) GetUpdateTime() *timestamp.Timestamp {
	if m != nil {
		return m.UpdateTime
	}
	return nil
}

func (m *Rack) GetRealm() string {
	if m != nil {
		return m.Realm
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Rack) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Rack_ChromeBrowserRack)(nil),
		(*Rack_ChromeosRack)(nil),
	}
}

// ChromeBrowserRack refers to the rack in Chrome Browser lab
type ChromeBrowserRack struct {
	// RPMs in the rack
	Rpms []string `protobuf:"bytes,1,rep,name=rpms,proto3" json:"rpms,omitempty"`
	// KVMs in the rack
	Kvms []string `protobuf:"bytes,2,rep,name=kvms,proto3" json:"kvms,omitempty"`
	// Switches in the rack
	Switches             []string `protobuf:"bytes,3,rep,name=switches,proto3" json:"switches,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChromeBrowserRack) Reset()         { *m = ChromeBrowserRack{} }
func (m *ChromeBrowserRack) String() string { return proto.CompactTextString(m) }
func (*ChromeBrowserRack) ProtoMessage()    {}
func (*ChromeBrowserRack) Descriptor() ([]byte, []int) {
	return fileDescriptor_4efc02fbdf306edb, []int{1}
}

func (m *ChromeBrowserRack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChromeBrowserRack.Unmarshal(m, b)
}
func (m *ChromeBrowserRack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChromeBrowserRack.Marshal(b, m, deterministic)
}
func (m *ChromeBrowserRack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChromeBrowserRack.Merge(m, src)
}
func (m *ChromeBrowserRack) XXX_Size() int {
	return xxx_messageInfo_ChromeBrowserRack.Size(m)
}
func (m *ChromeBrowserRack) XXX_DiscardUnknown() {
	xxx_messageInfo_ChromeBrowserRack.DiscardUnknown(m)
}

var xxx_messageInfo_ChromeBrowserRack proto.InternalMessageInfo

func (m *ChromeBrowserRack) GetRpms() []string {
	if m != nil {
		return m.Rpms
	}
	return nil
}

func (m *ChromeBrowserRack) GetKvms() []string {
	if m != nil {
		return m.Kvms
	}
	return nil
}

func (m *ChromeBrowserRack) GetSwitches() []string {
	if m != nil {
		return m.Switches
	}
	return nil
}

// ChromeOSRack refers to the rack in Chrome Browser lab
type ChromeOSRack struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChromeOSRack) Reset()         { *m = ChromeOSRack{} }
func (m *ChromeOSRack) String() string { return proto.CompactTextString(m) }
func (*ChromeOSRack) ProtoMessage()    {}
func (*ChromeOSRack) Descriptor() ([]byte, []int) {
	return fileDescriptor_4efc02fbdf306edb, []int{2}
}

func (m *ChromeOSRack) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChromeOSRack.Unmarshal(m, b)
}
func (m *ChromeOSRack) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChromeOSRack.Marshal(b, m, deterministic)
}
func (m *ChromeOSRack) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChromeOSRack.Merge(m, src)
}
func (m *ChromeOSRack) XXX_Size() int {
	return xxx_messageInfo_ChromeOSRack.Size(m)
}
func (m *ChromeOSRack) XXX_DiscardUnknown() {
	xxx_messageInfo_ChromeOSRack.DiscardUnknown(m)
}

var xxx_messageInfo_ChromeOSRack proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Rack)(nil), "unifiedfleet.api.v1.proto.Rack")
	proto.RegisterType((*ChromeBrowserRack)(nil), "unifiedfleet.api.v1.proto.ChromeBrowserRack")
	proto.RegisterType((*ChromeOSRack)(nil), "unifiedfleet.api.v1.proto.ChromeOSRack")
}

func init() {
	proto.RegisterFile("infra/unifiedfleet/api/v1/proto/rack.proto", fileDescriptor_4efc02fbdf306edb)
}

var fileDescriptor_4efc02fbdf306edb = []byte{
	// 500 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x4d, 0x6f, 0xd3, 0x30,
	0x18, 0xc7, 0xc9, 0xd2, 0x96, 0xcd, 0x2d, 0x48, 0x33, 0x1c, 0xc2, 0x2e, 0xab, 0x0a, 0xa3, 0x1d,
	0xda, 0x6c, 0x0d, 0x84, 0xc4, 0x8b, 0xd0, 0x68, 0xb9, 0x4c, 0x82, 0x01, 0xf2, 0x10, 0x07, 0x0e,
	0x54, 0x8e, 0xeb, 0xa4, 0x56, 0xe3, 0xda, 0xb2, 0x9d, 0x4e, 0x13, 0xe2, 0xeb, 0xf1, 0x25, 0xb8,
	0x70, 0xe6, 0x23, 0xec, 0x84, 0xe2, 0x24, 0xd3, 0x04, 0x62, 0xd1, 0x4e, 0x79, 0xfc, 0xe4, 0xf9,
	0xfd, 0x9f, 0x57, 0xf0, 0x48, 0x2c, 0x13, 0x43, 0x71, 0xbe, 0x14, 0x89, 0xe0, 0xb3, 0x24, 0xe3,
	0xdc, 0x61, 0xaa, 0x05, 0x5e, 0x1d, 0x60, 0x6d, 0x94, 0x53, 0xd8, 0x50, 0xb6, 0x40, 0xde, 0x84,
	0xf7, 0x2e, 0x47, 0x21, 0xaa, 0x05, 0x5a, 0x1d, 0x94, 0xbf, 0xb6, 0xb6, 0x53, 0xa5, 0xd2, 0x8c,
	0x97, 0x4c, 0x9c, 0x27, 0xd8, 0x09, 0xc9, 0xad, 0xa3, 0x52, 0x57, 0x01, 0xcf, 0x53, 0x85, 0xd8,
	0xdc, 0x28, 0x29, 0x72, 0x89, 0x94, 0x49, 0x71, 0x96, 0x33, 0x81, 0x53, 0xa3, 0x59, 0x95, 0xa7,
	0x12, 0x28, 0x72, 0x1b, 0x6e, 0x55, 0x6e, 0x18, 0xaf, 0xd0, 0xc3, 0x6b, 0xa0, 0x89, 0xe0, 0xd9,
	0x6c, 0x1a, 0xf3, 0x39, 0x5d, 0x09, 0x65, 0x2a, 0x01, 0xd4, 0xd4, 0x63, 0xa6, 0x18, 0x75, 0x42,
	0x2d, 0xcb, 0xf8, 0xc1, 0x8f, 0x10, 0xb4, 0x08, 0x65, 0x0b, 0x08, 0x41, 0x6b, 0x49, 0x25, 0x8f,
	0x82, 0x7e, 0x30, 0xda, 0x20, 0xde, 0x86, 0x87, 0x60, 0xbd, 0x0e, 0x8f, 0xd6, 0xfa, 0xc1, 0xa8,
	0xfb, 0xf8, 0x3e, 0xfa, 0xef, 0x5c, 0xd0, 0xbb, 0x2a, 0x94, 0x5c, 0x40, 0x70, 0x1b, 0x74, 0x19,
	0xd5, 0x94, 0x09, 0x77, 0x36, 0x35, 0x79, 0x14, 0xf6, 0x83, 0x51, 0x9b, 0x80, 0xda, 0x45, 0x72,
	0xf8, 0x15, 0xdc, 0xf1, 0xed, 0xf2, 0x69, 0x6c, 0xd4, 0xa9, 0xe5, 0x66, 0x5a, 0xec, 0x20, 0x6a,
	0xf9, 0x64, 0x7b, 0x57, 0x24, 0x7b, 0xe3, 0xa9, 0x49, 0x09, 0x15, 0x0d, 0x1c, 0xdd, 0x20, 0x9b,
	0xec, 0x6f, 0x27, 0x7c, 0x0f, 0x6e, 0x95, 0x4e, 0x65, 0x4b, 0xe5, 0xb6, 0x57, 0x1e, 0x36, 0x2a,
	0x7f, 0x38, 0xa9, 0x44, 0x7b, 0x35, 0xef, 0xf5, 0x5e, 0x83, 0x6e, 0xae, 0x67, 0xd4, 0xf1, 0x69,
	0xb1, 0xf4, 0xa8, 0xe3, 0xd5, 0xb6, 0x50, 0xb9, 0x15, 0x54, 0x5f, 0x04, 0xfa, 0x54, 0x5f, 0xc4,
	0x24, 0xfc, 0x35, 0x0e, 0x09, 0x28, 0x99, 0xc2, 0x0b, 0xef, 0x82, 0xb6, 0xe1, 0x34, 0x93, 0xd1,
	0x4d, 0x3f, 0xe8, 0xf2, 0xf1, 0xe2, 0xd9, 0xef, 0xf1, 0x53, 0xb0, 0x53, 0x15, 0xb5, 0xef, 0xab,
	0xda, 0xb7, 0x67, 0xd6, 0x71, 0x89, 0xa8, 0xd6, 0x56, 0x2b, 0x87, 0x98, 0x92, 0xd8, 0xd7, 0xd0,
	0x2b, 0x1a, 0xb1, 0xf8, 0x5b, 0xf1, 0xf9, 0x3e, 0xe9, 0x80, 0x56, 0x61, 0x0c, 0x7e, 0x06, 0x60,
	0xf3, 0x9f, 0xa1, 0xc0, 0x57, 0xa0, 0x65, 0xb4, 0xb4, 0x51, 0xd0, 0x0f, 0x47, 0x1b, 0x93, 0xdd,
	0xf3, 0xf1, 0x43, 0xf0, 0xa0, 0x39, 0xcb, 0xc7, 0x63, 0xe2, 0xb1, 0x02, 0x5f, 0xac, 0xa4, 0x8d,
	0xd6, 0xae, 0x83, 0xbf, 0xfd, 0x7c, 0x4c, 0x3c, 0x06, 0x8f, 0xc0, 0xba, 0x3d, 0x15, 0x8e, 0xcd,
	0xb9, 0x8d, 0x42, 0x2f, 0xb1, 0x77, 0x3e, 0xde, 0x05, 0xc3, 0x46, 0x89, 0x13, 0x4f, 0x91, 0x0b,
	0x7a, 0x70, 0x1b, 0xf4, 0x2e, 0xef, 0x65, 0x32, 0xfc, 0xb2, 0xd3, 0x70, 0xe8, 0x2f, 0xf3, 0xc4,
	0xea, 0x38, 0xee, 0xf8, 0xc7, 0x93, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xeb, 0x0b, 0xff, 0x41,
	0xfc, 0x03, 0x00, 0x00,
}
