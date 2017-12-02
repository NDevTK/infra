// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/luci-migration/config/config.proto

/*
Package config is a generated protocol buffer package.

It is generated from these files:
	infra/appengine/luci-migration/config/config.proto

It has these top-level messages:
	Config
	Master
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
	// Buildbot masters that we want to migrate to LUCI.
	Masters []*Master `protobuf:"bytes,1,rep,name=masters" json:"masters,omitempty"`
	// New bugs for discovered builders are filed on this Monorail instance.
	MonorailHostname string `protobuf:"bytes,2,opt,name=monorail_hostname,json=monorailHostname" json:"monorail_hostname,omitempty"`
	// Buildbot master information is fetched from this instance.
	BuildbotServiceHostname string `protobuf:"bytes,3,opt,name=buildbot_service_hostname,json=buildbotServiceHostname" json:"buildbot_service_hostname,omitempty"`
	// Builds will be searched and scheduled on this instance of buildbucket.
	BuildbucketHostname string `protobuf:"bytes,4,opt,name=buildbucket_hostname,json=buildbucketHostname" json:"buildbucket_hostname,omitempty"`
}

func (m *Config) Reset()                    { *m = Config{} }
func (m *Config) String() string            { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()               {}
func (*Config) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Config) GetMasters() []*Master {
	if m != nil {
		return m.Masters
	}
	return nil
}

func (m *Config) GetMonorailHostname() string {
	if m != nil {
		return m.MonorailHostname
	}
	return ""
}

func (m *Config) GetBuildbotServiceHostname() string {
	if m != nil {
		return m.BuildbotServiceHostname
	}
	return ""
}

func (m *Config) GetBuildbucketHostname() string {
	if m != nil {
		return m.BuildbucketHostname
	}
	return ""
}

// A single buildbot master.
type Master struct {
	// Name of the master without "master." prefix, e.g.
	// "tryserver.chromium.linux".
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// SchedulingType defines how builders on this mastter will be analyzed.
	SchedulingType SchedulingType `protobuf:"varint,2,opt,name=scheduling_type,json=schedulingType,enum=luci.migration.SchedulingType" json:"scheduling_type,omitempty"`
	// OS defines "OS" Monorail label.
	Os OS `protobuf:"varint,3,opt,name=os,enum=luci.migration.OS" json:"os,omitempty"`
	// LuciBucket is the equivalent LUCI buildbucket bucket.
	// It is assumed to have "LUCI <buildbot_builder_name>" builders for each
	// Buildbot builder.
	LuciBucket string `protobuf:"bytes,4,opt,name=luci_bucket,json=luciBucket" json:"luci_bucket,omitempty"`
	// If public, access is not controlled for read-only requests.
	Public bool `protobuf:"varint,5,opt,name=public" json:"public,omitempty"`
}

func (m *Master) Reset()                    { *m = Master{} }
func (m *Master) String() string            { return proto.CompactTextString(m) }
func (*Master) ProtoMessage()               {}
func (*Master) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Master) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Master) GetSchedulingType() SchedulingType {
	if m != nil {
		return m.SchedulingType
	}
	return SchedulingType_UNSET_SCHEDULING_TYPE
}

func (m *Master) GetOs() OS {
	if m != nil {
		return m.Os
	}
	return OS_UNSET_OS
}

func (m *Master) GetLuciBucket() string {
	if m != nil {
		return m.LuciBucket
	}
	return ""
}

func (m *Master) GetPublic() bool {
	if m != nil {
		return m.Public
	}
	return false
}

func init() {
	proto.RegisterType((*Config)(nil), "luci.migration.Config")
	proto.RegisterType((*Master)(nil), "luci.migration.Master")
	proto.RegisterEnum("luci.migration.SchedulingType", SchedulingType_name, SchedulingType_value)
	proto.RegisterEnum("luci.migration.OS", OS_name, OS_value)
}

func init() { proto.RegisterFile("infra/appengine/luci-migration/config/config.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 434 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x92, 0xcd, 0x6e, 0x9b, 0x40,
	0x14, 0x85, 0x03, 0xd8, 0xd8, 0xb9, 0xae, 0xe8, 0x74, 0xda, 0xa6, 0xce, 0xa6, 0xb5, 0xbc, 0xb2,
	0x52, 0xd5, 0x6e, 0xdd, 0x5d, 0x77, 0x31, 0x58, 0xc9, 0x54, 0x09, 0x13, 0x31, 0xd0, 0x34, 0xdd,
	0x20, 0x4c, 0x26, 0xce, 0xa8, 0x98, 0x41, 0xfc, 0x54, 0xca, 0xeb, 0xf5, 0x21, 0xfa, 0x3c, 0x15,
	0x83, 0x71, 0xe2, 0xac, 0xe0, 0x9e, 0xf3, 0x9d, 0xcb, 0x41, 0xba, 0x30, 0x17, 0xe9, 0x5d, 0x1e,
	0xcd, 0xa2, 0x2c, 0xe3, 0xe9, 0x5a, 0xa4, 0x7c, 0x96, 0x54, 0xb1, 0xf8, 0xb4, 0x11, 0xeb, 0x3c,
	0x2a, 0x85, 0x4c, 0x67, 0xb1, 0x4c, 0xef, 0xc4, 0x7a, 0xfb, 0x98, 0x66, 0xb9, 0x2c, 0x25, 0xb6,
	0x6a, 0x66, 0xba, 0x63, 0xc6, 0xff, 0x34, 0x30, 0x6d, 0x05, 0xe0, 0xcf, 0xd0, 0xdb, 0x44, 0x45,
	0xc9, 0xf3, 0x62, 0xa8, 0x8d, 0x8c, 0xc9, 0x60, 0x7e, 0x34, 0xdd, 0x87, 0xa7, 0x97, 0xca, 0xf6,
	0x5a, 0x0c, 0x7f, 0x84, 0x57, 0x1b, 0x99, 0xca, 0x3c, 0x12, 0x49, 0x78, 0x2f, 0x8b, 0x32, 0x8d,
	0x36, 0x7c, 0xa8, 0x8f, 0xb4, 0xc9, 0xa1, 0x87, 0x5a, 0xe3, 0x7c, 0xab, 0xe3, 0x6f, 0x70, 0xbc,
	0xaa, 0x44, 0x72, 0xbb, 0x92, 0x65, 0x58, 0xf0, 0xfc, 0x8f, 0x88, 0xf9, 0x63, 0xc8, 0x50, 0xa1,
	0x77, 0x2d, 0xc0, 0x1a, 0x7f, 0x97, 0xfd, 0x02, 0x6f, 0x1a, 0xab, 0x8a, 0x7f, 0xf3, 0xf2, 0x31,
	0xd6, 0x51, 0xb1, 0xd7, 0x4f, 0xbc, 0x36, 0x32, 0xfe, 0xab, 0x81, 0xd9, 0xf4, 0xc5, 0x18, 0x3a,
	0x8a, 0xd6, 0x14, 0xad, 0xde, 0xf1, 0x19, 0xbc, 0x2c, 0xe2, 0x7b, 0x7e, 0x5b, 0x25, 0x22, 0x5d,
	0x87, 0xe5, 0x43, 0xd6, 0x14, 0xb7, 0xe6, 0xef, 0x9f, 0xff, 0x34, 0xdb, 0x61, 0xfe, 0x43, 0xc6,
	0x3d, 0xab, 0xd8, 0x9b, 0xf1, 0x18, 0x74, 0x59, 0xa8, 0xfe, 0xd6, 0x1c, 0x3f, 0xcf, 0x52, 0xe6,
	0xe9, 0xb2, 0xc0, 0x1f, 0x60, 0x50, 0x1b, 0x61, 0x53, 0x71, 0xdb, 0x1a, 0x6a, 0x69, 0xa1, 0x14,
	0x7c, 0x04, 0x66, 0x56, 0xad, 0x12, 0x11, 0x0f, 0xbb, 0x23, 0x6d, 0xd2, 0xf7, 0xb6, 0xd3, 0xc9,
	0x0f, 0xb0, 0xf6, 0x3f, 0x8f, 0x8f, 0xe1, 0x6d, 0xe0, 0xb2, 0xa5, 0x1f, 0x32, 0xfb, 0x7c, 0xe9,
	0x04, 0x17, 0xc4, 0x3d, 0x0b, 0xfd, 0x9b, 0xab, 0x25, 0x3a, 0xc0, 0x03, 0xe8, 0xf9, 0xde, 0xcd,
	0x77, 0xba, 0x60, 0x48, 0xc3, 0x16, 0x80, 0x4d, 0x5d, 0x9f, 0xb8, 0x01, 0x0d, 0x18, 0xd2, 0xf1,
	0x0b, 0xe8, 0x5f, 0x2d, 0x3d, 0x42, 0x1d, 0x62, 0x23, 0xe3, 0x84, 0x80, 0x4e, 0x59, 0xad, 0x35,
	0xbb, 0x28, 0x43, 0x07, 0xf8, 0x10, 0xba, 0x17, 0xc4, 0x0d, 0x7e, 0x22, 0x0d, 0xf7, 0xc0, 0xb8,
	0x3c, 0xb5, 0x91, 0x5e, 0xaf, 0xbc, 0x26, 0xae, 0x43, 0xaf, 0x19, 0x32, 0xea, 0xe1, 0xd4, 0x75,
	0x3c, 0x4a, 0x1c, 0xd4, 0xa9, 0x11, 0x42, 0x19, 0xea, 0x2e, 0xfa, 0xbf, 0xcc, 0xe6, 0xc0, 0x56,
	0xa6, 0xba, 0xb0, 0xaf, 0xff, 0x03, 0x00, 0x00, 0xff, 0xff, 0x91, 0x2f, 0x04, 0xef, 0x97, 0x02,
	0x00, 0x00,
}
