// Code generated by protoc-gen-go.
// source: infra/appengine/luci-migration/config/config.proto
// DO NOT EDIT!

/*
Package config is a generated protocol buffer package.

It is generated from these files:
	infra/appengine/luci-migration/config/config.proto

It has these top-level messages:
	Config
	Monorail
	Milo
	Buildbot
*/
package config

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SchedulingType int32

const (
	SchedulingType_UNSET_SCHEDULING_TYPE SchedulingType = 0
	// TRYJOBS builds are scheduled for uncommitted CLs.
	SchedulingType_TRYJOBS SchedulingType = 1
	// CONTINUOUS builds are scheduled for landed CLs.
	SchedulingType_CONTINUOUS SchedulingType = 2
	// PERIODIC builds are scheduled every X time-units.
	SchedulingType_PERIODIC SchedulingType = 3
)

var SchedulingType_name = map[int32]string{
	0: "UNSET_SCHEDULING_TYPE",
	1: "TRYJOBS",
	2: "CONTINUOUS",
	3: "PERIODIC",
}
var SchedulingType_value = map[string]int32{
	"UNSET_SCHEDULING_TYPE": 0,
	"TRYJOBS":               1,
	"CONTINUOUS":            2,
	"PERIODIC":              3,
}

func (x SchedulingType) String() string {
	return proto.EnumName(SchedulingType_name, int32(x))
}
func (SchedulingType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// OS is an Operating System.
// OS names must match built-in "OS-<value>" Monorail labels.
type OS int32

const (
	OS_UNSET_OS OS = 0
	OS_LINUX    OS = 1
	OS_MAC      OS = 2
	OS_WINDOWS  OS = 3
	OS_ANDROID  OS = 4
	OS_IOS      OS = 5
)

var OS_name = map[int32]string{
	0: "UNSET_OS",
	1: "LINUX",
	2: "MAC",
	3: "WINDOWS",
	4: "ANDROID",
	5: "IOS",
}
var OS_value = map[string]int32{
	"UNSET_OS": 0,
	"LINUX":    1,
	"MAC":      2,
	"WINDOWS":  3,
	"ANDROID":  4,
	"IOS":      5,
}

func (x OS) String() string {
	return proto.EnumName(OS_name, int32(x))
}
func (OS) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Config struct {
	Monorail *Monorail `protobuf:"bytes,1,opt,name=monorail" json:"monorail,omitempty"`
	Milo     *Milo     `protobuf:"bytes,2,opt,name=milo" json:"milo,omitempty"`
	Buildbot *Buildbot `protobuf:"bytes,3,opt,name=buildbot" json:"buildbot,omitempty"`
}

func (m *Config) Reset()                    { *m = Config{} }
func (m *Config) String() string            { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()               {}
func (*Config) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Config) GetMonorail() *Monorail {
	if m != nil {
		return m.Monorail
	}
	return nil
}

func (m *Config) GetMilo() *Milo {
	if m != nil {
		return m.Milo
	}
	return nil
}

func (m *Config) GetBuildbot() *Buildbot {
	if m != nil {
		return m.Buildbot
	}
	return nil
}

type Monorail struct {
	Hostname string `protobuf:"bytes,1,opt,name=hostname" json:"hostname,omitempty"`
}

func (m *Monorail) Reset()                    { *m = Monorail{} }
func (m *Monorail) String() string            { return proto.CompactTextString(m) }
func (*Monorail) ProtoMessage()               {}
func (*Monorail) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Monorail) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

type Milo struct {
	Hostname string `protobuf:"bytes,1,opt,name=hostname" json:"hostname,omitempty"`
}

func (m *Milo) Reset()                    { *m = Milo{} }
func (m *Milo) String() string            { return proto.CompactTextString(m) }
func (*Milo) ProtoMessage()               {}
func (*Milo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Milo) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

type Buildbot struct {
	Masters []*Buildbot_Master `protobuf:"bytes,1,rep,name=masters" json:"masters,omitempty"`
}

func (m *Buildbot) Reset()                    { *m = Buildbot{} }
func (m *Buildbot) String() string            { return proto.CompactTextString(m) }
func (*Buildbot) ProtoMessage()               {}
func (*Buildbot) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Buildbot) GetMasters() []*Buildbot_Master {
	if m != nil {
		return m.Masters
	}
	return nil
}

type Buildbot_Master struct {
	Name           string         `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Public         bool           `protobuf:"varint,2,opt,name=public" json:"public,omitempty"`
	SchedulingType SchedulingType `protobuf:"varint,3,opt,name=scheduling_type,json=schedulingType,enum=luci.migration.SchedulingType" json:"scheduling_type,omitempty"`
	Os             OS             `protobuf:"varint,4,opt,name=os,enum=luci.migration.OS" json:"os,omitempty"`
	LuciBucket     string         `protobuf:"bytes,5,opt,name=luci_bucket,json=luciBucket" json:"luci_bucket,omitempty"`
}

func (m *Buildbot_Master) Reset()                    { *m = Buildbot_Master{} }
func (m *Buildbot_Master) String() string            { return proto.CompactTextString(m) }
func (*Buildbot_Master) ProtoMessage()               {}
func (*Buildbot_Master) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3, 0} }

func (m *Buildbot_Master) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Buildbot_Master) GetPublic() bool {
	if m != nil {
		return m.Public
	}
	return false
}

func (m *Buildbot_Master) GetSchedulingType() SchedulingType {
	if m != nil {
		return m.SchedulingType
	}
	return SchedulingType_UNSET_SCHEDULING_TYPE
}

func (m *Buildbot_Master) GetOs() OS {
	if m != nil {
		return m.Os
	}
	return OS_UNSET_OS
}

func (m *Buildbot_Master) GetLuciBucket() string {
	if m != nil {
		return m.LuciBucket
	}
	return ""
}

func init() {
	proto.RegisterType((*Config)(nil), "luci.migration.Config")
	proto.RegisterType((*Monorail)(nil), "luci.migration.Monorail")
	proto.RegisterType((*Milo)(nil), "luci.migration.Milo")
	proto.RegisterType((*Buildbot)(nil), "luci.migration.Buildbot")
	proto.RegisterType((*Buildbot_Master)(nil), "luci.migration.Buildbot.Master")
	proto.RegisterEnum("luci.migration.SchedulingType", SchedulingType_name, SchedulingType_value)
	proto.RegisterEnum("luci.migration.OS", OS_name, OS_value)
}

func init() { proto.RegisterFile("infra/appengine/luci-migration/config/config.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 458 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xcd, 0x6e, 0x9b, 0x4e,
	0x14, 0xc5, 0x03, 0xb6, 0x31, 0xb9, 0xfe, 0x8b, 0xff, 0x68, 0xd4, 0x56, 0x34, 0x8b, 0x26, 0x62,
	0x51, 0x59, 0x91, 0x8a, 0x25, 0xb7, 0x9b, 0x2e, 0x63, 0xb0, 0x52, 0xaa, 0x98, 0x89, 0x66, 0xa0,
	0x69, 0xba, 0xb1, 0x80, 0x10, 0x67, 0x54, 0x60, 0x10, 0x1f, 0x8b, 0xbc, 0x42, 0x1f, 0xa2, 0x0f,
	0xd3, 0x27, 0xab, 0x18, 0xdb, 0xa8, 0x4e, 0x3f, 0x56, 0x70, 0xef, 0xfd, 0x9d, 0x73, 0xae, 0x46,
	0x17, 0xe6, 0xbc, 0xb8, 0xaf, 0xa2, 0x59, 0x54, 0x96, 0x69, 0xb1, 0xe1, 0x45, 0x3a, 0xcb, 0xda,
	0x84, 0xbf, 0xc9, 0xf9, 0xa6, 0x8a, 0x1a, 0x2e, 0x8a, 0x59, 0x22, 0x8a, 0x7b, 0xbe, 0xd9, 0x7d,
	0xec, 0xb2, 0x12, 0x8d, 0xc0, 0x46, 0xc7, 0xd8, 0x3d, 0x63, 0x7d, 0x57, 0x40, 0x73, 0x24, 0x80,
	0xdf, 0x81, 0x9e, 0x8b, 0x42, 0x54, 0x11, 0xcf, 0x4c, 0xe5, 0x4c, 0x99, 0x4e, 0xe6, 0xa6, 0x7d,
	0x48, 0xdb, 0xab, 0xdd, 0x9c, 0xf6, 0x24, 0x9e, 0xc2, 0x30, 0xe7, 0x99, 0x30, 0x55, 0xa9, 0x78,
	0xf6, 0x9b, 0x82, 0x67, 0x82, 0x4a, 0xa2, 0xf3, 0x8f, 0x5b, 0x9e, 0xdd, 0xc5, 0xa2, 0x31, 0x07,
	0x7f, 0xf6, 0x5f, 0xec, 0xe6, 0xb4, 0x27, 0xad, 0xd7, 0xa0, 0xef, 0x53, 0xf1, 0x09, 0xe8, 0x0f,
	0xa2, 0x6e, 0x8a, 0x28, 0x4f, 0xe5, 0x86, 0xc7, 0xb4, 0xaf, 0x2d, 0x0b, 0x86, 0x5d, 0xd6, 0x3f,
	0x99, 0x6f, 0x2a, 0xe8, 0xfb, 0x08, 0xfc, 0x1e, 0xc6, 0x79, 0x54, 0x37, 0x69, 0x55, 0x9b, 0xca,
	0xd9, 0x60, 0x3a, 0x99, 0x9f, 0xfe, 0x6d, 0x1b, 0x7b, 0x25, 0x39, 0xba, 0xe7, 0x4f, 0x7e, 0x28,
	0xa0, 0x6d, 0x7b, 0x18, 0xc3, 0xf0, 0x97, 0x28, 0xf9, 0x8f, 0x5f, 0x80, 0x56, 0xb6, 0x71, 0xc6,
	0x13, 0xf9, 0x28, 0x3a, 0xdd, 0x55, 0xf8, 0x12, 0xfe, 0xaf, 0x93, 0x87, 0xf4, 0xae, 0xcd, 0x78,
	0xb1, 0x59, 0x37, 0x8f, 0x65, 0x2a, 0xdf, 0xc1, 0x98, 0xbf, 0x7a, 0x9a, 0xcc, 0x7a, 0x2c, 0x78,
	0x2c, 0x53, 0x6a, 0xd4, 0x07, 0x35, 0xb6, 0x40, 0x15, 0xb5, 0x39, 0x94, 0x5a, 0xfc, 0x54, 0x4b,
	0x18, 0x55, 0x45, 0x8d, 0x4f, 0x61, 0xd2, 0x0d, 0xd6, 0x71, 0x9b, 0x7c, 0x4d, 0x1b, 0x73, 0x24,
	0xf7, 0x83, 0xae, 0xb5, 0x90, 0x9d, 0xf3, 0x4f, 0x60, 0x1c, 0xc6, 0xe0, 0x97, 0xf0, 0x3c, 0xf4,
	0xd9, 0x32, 0x58, 0x33, 0xe7, 0xc3, 0xd2, 0x0d, 0xaf, 0x3c, 0xff, 0x72, 0x1d, 0xdc, 0x5e, 0x2f,
	0xd1, 0x11, 0x9e, 0xc0, 0x38, 0xa0, 0xb7, 0x1f, 0xc9, 0x82, 0x21, 0x05, 0x1b, 0x00, 0x0e, 0xf1,
	0x03, 0xcf, 0x0f, 0x49, 0xc8, 0x90, 0x8a, 0xff, 0x03, 0xfd, 0x7a, 0x49, 0x3d, 0xe2, 0x7a, 0x0e,
	0x1a, 0x9c, 0x7b, 0xa0, 0x12, 0xd6, 0xf5, 0xb6, 0x5e, 0x84, 0xa1, 0x23, 0x7c, 0x0c, 0xa3, 0x2b,
	0xcf, 0x0f, 0x3f, 0x23, 0x05, 0x8f, 0x61, 0xb0, 0xba, 0x70, 0x90, 0xda, 0x59, 0xde, 0x78, 0xbe,
	0x4b, 0x6e, 0x18, 0x1a, 0x74, 0xc5, 0x85, 0xef, 0x52, 0xe2, 0xb9, 0x68, 0xd8, 0x21, 0x1e, 0x61,
	0x68, 0xb4, 0xd0, 0xbf, 0x68, 0xdb, 0xe3, 0x8d, 0x35, 0x79, 0xbd, 0x6f, 0x7f, 0x06, 0x00, 0x00,
	0xff, 0xff, 0x09, 0xf9, 0x32, 0x6f, 0xf3, 0x02, 0x00, 0x00,
}
