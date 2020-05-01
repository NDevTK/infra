// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/unifiedfleet/api/v1/proto/lse_prototype.proto

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

// The supported peripheral type in LSE definition. The list is not completed
// as we have many special setups in ChromeOS high-touch labs. Will add them later
// when it comes to use.
type PeripheralType int32

const (
	PeripheralType_PERIPHERAL_TYPE_UNSPECIFIED       PeripheralType = 0
	PeripheralType_PERIPHERAL_TYPE_SERVO             PeripheralType = 1
	PeripheralType_PERIPHERAL_TYPE_LABSTATION        PeripheralType = 2
	PeripheralType_PERIPHERAL_TYPE_RPM               PeripheralType = 3
	PeripheralType_PERIPHERAL_TYPE_KVM               PeripheralType = 4
	PeripheralType_PERIPHERAL_TYPE_SWITCH            PeripheralType = 5
	PeripheralType_PERIPHERAL_TYPE_BLUETOOTH_BTPEERS PeripheralType = 6
	PeripheralType_PERIPHERAL_TYPE_WIFICELL          PeripheralType = 7
	PeripheralType_PERIPHERAL_TYPE_CAMERA            PeripheralType = 8
)

var PeripheralType_name = map[int32]string{
	0: "PERIPHERAL_TYPE_UNSPECIFIED",
	1: "PERIPHERAL_TYPE_SERVO",
	2: "PERIPHERAL_TYPE_LABSTATION",
	3: "PERIPHERAL_TYPE_RPM",
	4: "PERIPHERAL_TYPE_KVM",
	5: "PERIPHERAL_TYPE_SWITCH",
	6: "PERIPHERAL_TYPE_BLUETOOTH_BTPEERS",
	7: "PERIPHERAL_TYPE_WIFICELL",
	8: "PERIPHERAL_TYPE_CAMERA",
}

var PeripheralType_value = map[string]int32{
	"PERIPHERAL_TYPE_UNSPECIFIED":       0,
	"PERIPHERAL_TYPE_SERVO":             1,
	"PERIPHERAL_TYPE_LABSTATION":        2,
	"PERIPHERAL_TYPE_RPM":               3,
	"PERIPHERAL_TYPE_KVM":               4,
	"PERIPHERAL_TYPE_SWITCH":            5,
	"PERIPHERAL_TYPE_BLUETOOTH_BTPEERS": 6,
	"PERIPHERAL_TYPE_WIFICELL":          7,
	"PERIPHERAL_TYPE_CAMERA":            8,
}

func (x PeripheralType) String() string {
	return proto.EnumName(PeripheralType_name, int32(x))
}

func (PeripheralType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_febad07b5f2297d1, []int{0}
}

type RackLSEPrototype struct {
	// A unique name for the RackLSEPrototype.
	// The format will be rackLSEPrototypes/XXX
	Name                  string                   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	PeripheralRequirement []*PeripheralRequirement `protobuf:"bytes,2,rep,name=peripheral_requirement,json=peripheralRequirement,proto3" json:"peripheral_requirement,omitempty"`
	// Record the last update timestamp of this RackLSEPrototype (In UTC timezone)
	UpdateTime           *timestamp.Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *RackLSEPrototype) Reset()         { *m = RackLSEPrototype{} }
func (m *RackLSEPrototype) String() string { return proto.CompactTextString(m) }
func (*RackLSEPrototype) ProtoMessage()    {}
func (*RackLSEPrototype) Descriptor() ([]byte, []int) {
	return fileDescriptor_febad07b5f2297d1, []int{0}
}

func (m *RackLSEPrototype) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RackLSEPrototype.Unmarshal(m, b)
}
func (m *RackLSEPrototype) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RackLSEPrototype.Marshal(b, m, deterministic)
}
func (m *RackLSEPrototype) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RackLSEPrototype.Merge(m, src)
}
func (m *RackLSEPrototype) XXX_Size() int {
	return xxx_messageInfo_RackLSEPrototype.Size(m)
}
func (m *RackLSEPrototype) XXX_DiscardUnknown() {
	xxx_messageInfo_RackLSEPrototype.DiscardUnknown(m)
}

var xxx_messageInfo_RackLSEPrototype proto.InternalMessageInfo

func (m *RackLSEPrototype) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *RackLSEPrototype) GetPeripheralRequirement() []*PeripheralRequirement {
	if m != nil {
		return m.PeripheralRequirement
	}
	return nil
}

func (m *RackLSEPrototype) GetUpdateTime() *timestamp.Timestamp {
	if m != nil {
		return m.UpdateTime
	}
	return nil
}

type MachineLSEPrototype struct {
	// A unique name for the MachineLSEPrototype.
	// The format will be machineLSEPrototypes/XXX
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// peripheral_requirements.peripheral_type must be unique.
	PeripheralRequirements []*PeripheralRequirement `protobuf:"bytes,2,rep,name=peripheral_requirements,json=peripheralRequirements,proto3" json:"peripheral_requirements,omitempty"`
	// Indicates the Rack Unit capacity of this setup, corresponding
	// to a Rack’s Rack Unit capacity.
	OccupiedCapacityRu int32 `protobuf:"varint,3,opt,name=occupied_capacity_ru,json=occupiedCapacityRu,proto3" json:"occupied_capacity_ru,omitempty"`
	// Record the last update timestamp of this MachineLSEPrototype (In UTC timezone)
	UpdateTime           *timestamp.Timestamp `protobuf:"bytes,4,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *MachineLSEPrototype) Reset()         { *m = MachineLSEPrototype{} }
func (m *MachineLSEPrototype) String() string { return proto.CompactTextString(m) }
func (*MachineLSEPrototype) ProtoMessage()    {}
func (*MachineLSEPrototype) Descriptor() ([]byte, []int) {
	return fileDescriptor_febad07b5f2297d1, []int{1}
}

func (m *MachineLSEPrototype) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MachineLSEPrototype.Unmarshal(m, b)
}
func (m *MachineLSEPrototype) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MachineLSEPrototype.Marshal(b, m, deterministic)
}
func (m *MachineLSEPrototype) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MachineLSEPrototype.Merge(m, src)
}
func (m *MachineLSEPrototype) XXX_Size() int {
	return xxx_messageInfo_MachineLSEPrototype.Size(m)
}
func (m *MachineLSEPrototype) XXX_DiscardUnknown() {
	xxx_messageInfo_MachineLSEPrototype.DiscardUnknown(m)
}

var xxx_messageInfo_MachineLSEPrototype proto.InternalMessageInfo

func (m *MachineLSEPrototype) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MachineLSEPrototype) GetPeripheralRequirements() []*PeripheralRequirement {
	if m != nil {
		return m.PeripheralRequirements
	}
	return nil
}

func (m *MachineLSEPrototype) GetOccupiedCapacityRu() int32 {
	if m != nil {
		return m.OccupiedCapacityRu
	}
	return 0
}

func (m *MachineLSEPrototype) GetUpdateTime() *timestamp.Timestamp {
	if m != nil {
		return m.UpdateTime
	}
	return nil
}

// The requirement for peripherals of a LSE. Usually it’s predefined
// by the designer of the test and lab, e.g. a test needs 2 cameras, 1 rpm,
// 1 servo, and a labstation.
// We probably also record cables as ChromeOS ACS lab wants to track the cable
// usage also.
type PeripheralRequirement struct {
	// It refers to the peripheral type that a LSE needs. The common use cases
	// include: kvm, switch, servo, rpm, labstation, camera, ...
	PeripheralType PeripheralType `protobuf:"varint,1,opt,name=peripheral_type,json=peripheralType,proto3,enum=unifiedfleet.api.v1.proto.PeripheralType" json:"peripheral_type,omitempty"`
	// The minimum/maximum number of the peripherals that needed by a LSE, e.g.
	// A test needs 1-3 bluetooth bt peers to be set up.
	Min                  int32    `protobuf:"varint,2,opt,name=min,proto3" json:"min,omitempty"`
	Max                  int32    `protobuf:"varint,3,opt,name=max,proto3" json:"max,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeripheralRequirement) Reset()         { *m = PeripheralRequirement{} }
func (m *PeripheralRequirement) String() string { return proto.CompactTextString(m) }
func (*PeripheralRequirement) ProtoMessage()    {}
func (*PeripheralRequirement) Descriptor() ([]byte, []int) {
	return fileDescriptor_febad07b5f2297d1, []int{2}
}

func (m *PeripheralRequirement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeripheralRequirement.Unmarshal(m, b)
}
func (m *PeripheralRequirement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeripheralRequirement.Marshal(b, m, deterministic)
}
func (m *PeripheralRequirement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeripheralRequirement.Merge(m, src)
}
func (m *PeripheralRequirement) XXX_Size() int {
	return xxx_messageInfo_PeripheralRequirement.Size(m)
}
func (m *PeripheralRequirement) XXX_DiscardUnknown() {
	xxx_messageInfo_PeripheralRequirement.DiscardUnknown(m)
}

var xxx_messageInfo_PeripheralRequirement proto.InternalMessageInfo

func (m *PeripheralRequirement) GetPeripheralType() PeripheralType {
	if m != nil {
		return m.PeripheralType
	}
	return PeripheralType_PERIPHERAL_TYPE_UNSPECIFIED
}

func (m *PeripheralRequirement) GetMin() int32 {
	if m != nil {
		return m.Min
	}
	return 0
}

func (m *PeripheralRequirement) GetMax() int32 {
	if m != nil {
		return m.Max
	}
	return 0
}

func init() {
	proto.RegisterEnum("unifiedfleet.api.v1.proto.PeripheralType", PeripheralType_name, PeripheralType_value)
	proto.RegisterType((*RackLSEPrototype)(nil), "unifiedfleet.api.v1.proto.RackLSEPrototype")
	proto.RegisterType((*MachineLSEPrototype)(nil), "unifiedfleet.api.v1.proto.MachineLSEPrototype")
	proto.RegisterType((*PeripheralRequirement)(nil), "unifiedfleet.api.v1.proto.PeripheralRequirement")
}

func init() {
	proto.RegisterFile("infra/unifiedfleet/api/v1/proto/lse_prototype.proto", fileDescriptor_febad07b5f2297d1)
}

var fileDescriptor_febad07b5f2297d1 = []byte{
	// 617 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0xdd, 0x4e, 0x13, 0x4d,
	0x18, 0xc7, 0xdf, 0xb6, 0xc0, 0xab, 0x43, 0x82, 0x9b, 0x41, 0xa0, 0x54, 0x23, 0x48, 0x24, 0x22,
	0x09, 0xb3, 0x7c, 0x78, 0xa2, 0x1e, 0xe8, 0xb6, 0x0e, 0x61, 0x63, 0x4b, 0x37, 0xd3, 0x05, 0x82,
	0x31, 0xd9, 0x4c, 0xb7, 0xd3, 0x76, 0x62, 0x77, 0x67, 0x9c, 0xdd, 0x25, 0x36, 0x84, 0x6b, 0xf0,
	0x02, 0xbc, 0x08, 0x6f, 0xc9, 0x63, 0x4e, 0xbc, 0x05, 0xb3, 0xdb, 0xad, 0xc2, 0xb2, 0xa6, 0xe2,
	0xd9, 0x33, 0xff, 0xff, 0xf3, 0xf9, 0x3b, 0x18, 0xb0, 0xc7, 0xfd, 0xae, 0xa2, 0x7a, 0xe4, 0xf3,
	0x2e, 0x67, 0x9d, 0xee, 0x80, 0xb1, 0x50, 0xa7, 0x92, 0xeb, 0x67, 0x3b, 0xba, 0x54, 0x22, 0x14,
	0xfa, 0x20, 0x60, 0x4e, 0x12, 0x85, 0x43, 0xc9, 0x50, 0x12, 0xc1, 0xe5, 0xab, 0xe9, 0x88, 0x4a,
	0x8e, 0xce, 0x76, 0x46, 0x56, 0x65, 0xa5, 0x27, 0x44, 0x6f, 0xc0, 0x46, 0xc5, 0xed, 0xa8, 0xab,
	0x87, 0xdc, 0x63, 0x41, 0x48, 0x3d, 0x99, 0x26, 0xbc, 0xe8, 0x09, 0xe4, 0xf6, 0x95, 0xf0, 0x78,
	0xe4, 0x21, 0xa1, 0x7a, 0xfa, 0x20, 0x72, 0xb9, 0xde, 0x53, 0xd2, 0x4d, 0x07, 0xa6, 0x0d, 0xe2,
	0x25, 0x14, 0x0b, 0x44, 0xa4, 0xdc, 0x74, 0x6c, 0xe5, 0xf5, 0x2d, 0x4a, 0xbb, 0x9c, 0x0d, 0x3a,
	0x4e, 0x9b, 0xf5, 0xe9, 0x19, 0x17, 0x6a, 0xd4, 0x60, 0xed, 0x5b, 0x11, 0x68, 0x84, 0xba, 0x1f,
	0xeb, 0x2d, 0x6c, 0x8d, 0x4f, 0x82, 0x10, 0x4c, 0xf9, 0xd4, 0x63, 0xe5, 0xc2, 0x6a, 0x61, 0xe3,
	0x2e, 0x49, 0x62, 0xd8, 0x03, 0x8b, 0x92, 0x29, 0x2e, 0xfb, 0x4c, 0xd1, 0x81, 0xa3, 0xd8, 0xa7,
	0x88, 0x2b, 0xe6, 0x31, 0x3f, 0x2c, 0x17, 0x57, 0x4b, 0x1b, 0xb3, 0xbb, 0xdb, 0xe8, 0x8f, 0x04,
	0x90, 0xf5, 0xab, 0x90, 0xfc, 0xae, 0x23, 0x0b, 0x32, 0x4f, 0x86, 0x6f, 0xc0, 0x6c, 0x24, 0x3b,
	0x34, 0x64, 0x4e, 0xcc, 0xa9, 0x5c, 0x5a, 0x2d, 0x6c, 0xcc, 0xee, 0x56, 0xd0, 0xe8, 0x10, 0x34,
	0x86, 0x88, 0xec, 0x31, 0xc4, 0x6a, 0xe9, 0xbb, 0x51, 0x22, 0x60, 0x54, 0x13, 0xab, 0x2f, 0x3f,
	0x5c, 0x1a, 0xa7, 0x60, 0x27, 0x5d, 0x67, 0x2b, 0xd9, 0x67, 0x2b, 0x18, 0x06, 0x21, 0xf3, 0x10,
	0x95, 0x32, 0x90, 0x22, 0x44, 0xae, 0xf0, 0xf4, 0x1b, 0x67, 0x3f, 0x51, 0x19, 0x25, 0xd0, 0xcf,
	0xb3, 0xd2, 0xc5, 0xda, 0x8f, 0x22, 0x98, 0x6f, 0x50, 0xb7, 0xcf, 0x7d, 0x36, 0x11, 0x1a, 0x07,
	0x4b, 0xf9, 0xd0, 0x82, 0x7f, 0xa6, 0xb6, 0x98, 0x4b, 0x2d, 0x80, 0xdb, 0xe0, 0xbe, 0x70, 0xdd,
	0x48, 0x72, 0xd6, 0x71, 0x5c, 0x2a, 0xa9, 0xcb, 0xc3, 0xa1, 0xa3, 0xa2, 0x84, 0xdf, 0x34, 0x81,
	0x63, 0xaf, 0x96, 0x5a, 0x24, 0xca, 0x82, 0x9e, 0xba, 0x3d, 0x68, 0x76, 0x69, 0xb4, 0xc1, 0xf3,
	0x89, 0xa0, 0xf3, 0x68, 0x6d, 0x7a, 0x37, 0xc5, 0x40, 0x3f, 0xcf, 0x51, 0x2f, 0xd6, 0xbe, 0x14,
	0xc0, 0x42, 0x2e, 0x0c, 0x48, 0xc0, 0xbd, 0x2b, 0x7c, 0xe3, 0xec, 0x04, 0xff, 0xdc, 0xee, 0xb3,
	0xbf, 0xe2, 0x6a, 0x0f, 0x25, 0x23, 0x73, 0xf2, 0xda, 0x1b, 0x6a, 0xa0, 0xe4, 0x71, 0xbf, 0x5c,
	0x4c, 0xb8, 0xc5, 0x61, 0xa2, 0xd0, 0xcf, 0x29, 0xc9, 0x38, 0xdc, 0xfc, 0x5a, 0x04, 0x73, 0xd7,
	0xdb, 0xc0, 0x15, 0xf0, 0xc0, 0xc2, 0xc4, 0xb4, 0x0e, 0x30, 0x31, 0xea, 0x8e, 0x7d, 0x6a, 0x61,
	0xe7, 0xe8, 0xb0, 0x65, 0xe1, 0x9a, 0xb9, 0x6f, 0xe2, 0xb7, 0xda, 0x7f, 0x70, 0x19, 0x2c, 0x64,
	0x13, 0x5a, 0x98, 0x1c, 0x37, 0xb5, 0x02, 0x7c, 0x04, 0x2a, 0x59, 0xab, 0x6e, 0x54, 0x5b, 0xb6,
	0x61, 0x9b, 0xcd, 0x43, 0xad, 0x08, 0x97, 0xc0, 0x7c, 0xd6, 0x27, 0x56, 0x43, 0x2b, 0xe5, 0x19,
	0xef, 0x8e, 0x1b, 0xda, 0x14, 0xac, 0x80, 0xc5, 0x1b, 0xc3, 0x4e, 0x4c, 0xbb, 0x76, 0xa0, 0x4d,
	0xc3, 0x75, 0xf0, 0x38, 0xeb, 0x55, 0xeb, 0x47, 0xd8, 0x6e, 0x36, 0xed, 0x03, 0xa7, 0x6a, 0x5b,
	0x18, 0x93, 0x96, 0x36, 0x03, 0x1f, 0x82, 0x72, 0x36, 0xed, 0xc4, 0xdc, 0x37, 0x6b, 0xb8, 0x5e,
	0xd7, 0xfe, 0xcf, 0x1b, 0x50, 0x33, 0x1a, 0x98, 0x18, 0xda, 0x9d, 0xea, 0xd3, 0xf7, 0xeb, 0x13,
	0xbe, 0xd0, 0x57, 0x51, 0x37, 0x90, 0xed, 0xf6, 0x4c, 0xf2, 0xd8, 0xfb, 0x19, 0x00, 0x00, 0xff,
	0xff, 0xb7, 0x52, 0x8c, 0x67, 0x72, 0x05, 0x00, 0x00,
}
