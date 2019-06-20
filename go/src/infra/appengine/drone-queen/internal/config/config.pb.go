// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/appengine/drone-queen/internal/config/config.proto

package config

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
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

// Config is the configuration data served by luci-config for this app.
type Config struct {
	// access_groups are the luci-auth groups controlling access to RPC endpoints.
	AccessGroups *AccessGroups `protobuf:"bytes,1,opt,name=access_groups,json=accessGroups,proto3" json:"access_groups,omitempty"`
	// assignment_duration is the duration before expiration for drone
	// assignments.
	AssignmentDuration   *duration.Duration `protobuf:"bytes,2,opt,name=assignment_duration,json=assignmentDuration,proto3" json:"assignment_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bcfef40975c8024, []int{0}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetAccessGroups() *AccessGroups {
	if m != nil {
		return m.AccessGroups
	}
	return nil
}

func (m *Config) GetAssignmentDuration() *duration.Duration {
	if m != nil {
		return m.AssignmentDuration
	}
	return nil
}

// AccessGroups holds access group configuration
type AccessGroups struct {
	// drones is the group for calling drone RPCs.
	Drones string `protobuf:"bytes,1,opt,name=drones,proto3" json:"drones,omitempty"`
	// inventory_providers is the group for calling inventory RPCs.
	InventoryProviders   string   `protobuf:"bytes,2,opt,name=inventory_providers,json=inventoryProviders,proto3" json:"inventory_providers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccessGroups) Reset()         { *m = AccessGroups{} }
func (m *AccessGroups) String() string { return proto.CompactTextString(m) }
func (*AccessGroups) ProtoMessage()    {}
func (*AccessGroups) Descriptor() ([]byte, []int) {
	return fileDescriptor_9bcfef40975c8024, []int{1}
}

func (m *AccessGroups) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccessGroups.Unmarshal(m, b)
}
func (m *AccessGroups) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccessGroups.Marshal(b, m, deterministic)
}
func (m *AccessGroups) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccessGroups.Merge(m, src)
}
func (m *AccessGroups) XXX_Size() int {
	return xxx_messageInfo_AccessGroups.Size(m)
}
func (m *AccessGroups) XXX_DiscardUnknown() {
	xxx_messageInfo_AccessGroups.DiscardUnknown(m)
}

var xxx_messageInfo_AccessGroups proto.InternalMessageInfo

func (m *AccessGroups) GetDrones() string {
	if m != nil {
		return m.Drones
	}
	return ""
}

func (m *AccessGroups) GetInventoryProviders() string {
	if m != nil {
		return m.InventoryProviders
	}
	return ""
}

func init() {
	proto.RegisterType((*Config)(nil), "drone_queen.config.Config")
	proto.RegisterType((*AccessGroups)(nil), "drone_queen.config.AccessGroups")
}

func init() {
	proto.RegisterFile("infra/appengine/drone-queen/internal/config/config.proto", fileDescriptor_9bcfef40975c8024)
}

var fileDescriptor_9bcfef40975c8024 = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x8f, 0x3d, 0x4b, 0xc4, 0x40,
	0x10, 0x86, 0x89, 0x45, 0xf0, 0xd6, 0xb3, 0xd9, 0x03, 0x39, 0x2d, 0xe4, 0xb8, 0xca, 0xc6, 0x5d,
	0xd0, 0xc6, 0xd6, 0x2f, 0x04, 0x2b, 0x49, 0x23, 0xd8, 0x84, 0xbd, 0x64, 0xb2, 0x2c, 0x9c, 0x33,
	0x71, 0x76, 0x73, 0xe0, 0x6f, 0xf1, 0xcf, 0x0a, 0xb3, 0x89, 0x1e, 0x5c, 0x15, 0x26, 0xef, 0x33,
	0xf3, 0x3e, 0xab, 0xee, 0x02, 0x76, 0xec, 0xac, 0xeb, 0x7b, 0x40, 0x1f, 0x10, 0x6c, 0xcb, 0x84,
	0x70, 0xfd, 0x35, 0x00, 0xa0, 0x0d, 0x98, 0x80, 0xd1, 0x6d, 0x6d, 0x43, 0xd8, 0x05, 0x3f, 0x7e,
	0x4c, 0xcf, 0x94, 0x48, 0x6b, 0x21, 0x6b, 0x21, 0x4d, 0x4e, 0x2e, 0x2e, 0x3d, 0x91, 0xdf, 0x82,
	0x15, 0x62, 0x33, 0x74, 0xb6, 0x1d, 0xd8, 0xa5, 0x40, 0x98, 0x77, 0xd6, 0x3f, 0x85, 0x2a, 0x1f,
	0x05, 0xd5, 0xcf, 0xea, 0xd4, 0x35, 0x0d, 0xc4, 0x58, 0x7b, 0xa6, 0xa1, 0x8f, 0xcb, 0x62, 0x55,
	0x5c, 0x9d, 0xdc, 0xac, 0xcc, 0xe1, 0x59, 0x73, 0x2f, 0xe0, 0x8b, 0x70, 0xd5, 0xdc, 0xed, 0x4d,
	0xfa, 0x55, 0x2d, 0x5c, 0x8c, 0xc1, 0xe3, 0x27, 0x60, 0xaa, 0xa7, 0xba, 0xe5, 0x91, 0x1c, 0x3b,
	0x37, 0xd9, 0xc7, 0x4c, 0x3e, 0xe6, 0x69, 0x04, 0x2a, 0xfd, 0xbf, 0x35, 0xfd, 0x5b, 0xbf, 0xab,
	0xf9, 0x7e, 0x93, 0x3e, 0x53, 0xa5, 0xc8, 0x64, 0xb7, 0x59, 0x35, 0x4e, 0xda, 0xaa, 0x45, 0xc0,
	0x1d, 0x60, 0x22, 0xfe, 0xae, 0x7b, 0xa6, 0x5d, 0x68, 0x81, 0xa3, 0x74, 0xce, 0x2a, 0xfd, 0x17,
	0xbd, 0x4d, 0xc9, 0xc3, 0xf1, 0x47, 0x99, 0x5f, 0xb2, 0x29, 0xc5, 0xe4, 0xf6, 0x37, 0x00, 0x00,
	0xff, 0xff, 0xf4, 0x09, 0x4d, 0xb6, 0x77, 0x01, 0x00, 0x00,
}
