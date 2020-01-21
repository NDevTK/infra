// Code generated by protoc-gen-go. DO NOT EDIT.
// source: lab.proto

/*
Package inventory is a generated protocol buffer package.

It is generated from these files:
	lab.proto

It has these top-level messages:
	Lab
	Infrastructure
*/
package inventory

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import chrome_chromeos_infra_skylab_proto_inventory2 "."
import chrome_chromeos_infra_skylab_proto_inventory1 "."
import chrome_chromeos_infra_skylab_proto_inventory3 "."

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// NEXT TAG: 6
type Lab struct {
	Duts                 []*chrome_chromeos_infra_skylab_proto_inventory1.DeviceUnderTest     `protobuf:"bytes,1,rep,name=duts" json:"duts,omitempty"`
	ServoHosts           []*chrome_chromeos_infra_skylab_proto_inventory1.ServoHostDevice     `protobuf:"bytes,2,rep,name=servo_hosts,json=servoHosts" json:"servo_hosts,omitempty"`
	Chamelons            []*chrome_chromeos_infra_skylab_proto_inventory1.ChameleonDevice     `protobuf:"bytes,3,rep,name=chamelons" json:"chamelons,omitempty"`
	ServoHostConnections []*chrome_chromeos_infra_skylab_proto_inventory2.ServoHostConnection `protobuf:"bytes,4,rep,name=servo_host_connections,json=servoHostConnections" json:"servo_host_connections,omitempty"`
	ChameleonConnections []*chrome_chromeos_infra_skylab_proto_inventory2.ChameleonConnection `protobuf:"bytes,5,rep,name=chameleon_connections,json=chameleonConnections" json:"chameleon_connections,omitempty"`
	XXX_unrecognized     []byte                                                               `json:"-"`
}

func (m *Lab) Reset()                    { *m = Lab{} }
func (m *Lab) String() string            { return proto.CompactTextString(m) }
func (*Lab) ProtoMessage()               {}
func (*Lab) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Lab) GetDuts() []*chrome_chromeos_infra_skylab_proto_inventory1.DeviceUnderTest {
	if m != nil {
		return m.Duts
	}
	return nil
}

func (m *Lab) GetServoHosts() []*chrome_chromeos_infra_skylab_proto_inventory1.ServoHostDevice {
	if m != nil {
		return m.ServoHosts
	}
	return nil
}

func (m *Lab) GetChamelons() []*chrome_chromeos_infra_skylab_proto_inventory1.ChameleonDevice {
	if m != nil {
		return m.Chamelons
	}
	return nil
}

func (m *Lab) GetServoHostConnections() []*chrome_chromeos_infra_skylab_proto_inventory2.ServoHostConnection {
	if m != nil {
		return m.ServoHostConnections
	}
	return nil
}

func (m *Lab) GetChameleonConnections() []*chrome_chromeos_infra_skylab_proto_inventory2.ChameleonConnection {
	if m != nil {
		return m.ChameleonConnections
	}
	return nil
}

type Infrastructure struct {
	Servers          []*chrome_chromeos_infra_skylab_proto_inventory3.Server `protobuf:"bytes,1,rep,name=servers" json:"servers,omitempty"`
	XXX_unrecognized []byte                                                  `json:"-"`
}

func (m *Infrastructure) Reset()                    { *m = Infrastructure{} }
func (m *Infrastructure) String() string            { return proto.CompactTextString(m) }
func (*Infrastructure) ProtoMessage()               {}
func (*Infrastructure) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Infrastructure) GetServers() []*chrome_chromeos_infra_skylab_proto_inventory3.Server {
	if m != nil {
		return m.Servers
	}
	return nil
}

func init() {
	proto.RegisterType((*Lab)(nil), "chrome.chromeos_infra.skylab.proto.inventory.Lab")
	proto.RegisterType((*Infrastructure)(nil), "chrome.chromeos_infra.skylab.proto.inventory.Infrastructure")
}

func init() { proto.RegisterFile("lab.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 284 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xc1, 0x4a, 0xc4, 0x40,
	0x0c, 0x40, 0xd1, 0xad, 0x48, 0x53, 0x11, 0x19, 0x54, 0xca, 0x9e, 0x64, 0x4f, 0x1e, 0x64, 0x0e,
	0xe2, 0xd5, 0x83, 0xae, 0x07, 0x05, 0x11, 0xac, 0x7a, 0x51, 0xb0, 0x76, 0xa7, 0x91, 0x16, 0x77,
	0x27, 0x92, 0x4c, 0x2b, 0xfb, 0x03, 0x7e, 0xb7, 0xb4, 0xb5, 0xad, 0xb2, 0xa7, 0xed, 0x69, 0x48,
	0x32, 0xf3, 0x5e, 0x32, 0x04, 0xfc, 0x79, 0x32, 0xd3, 0x9f, 0x4c, 0x8e, 0xd4, 0x89, 0xc9, 0x98,
	0x16, 0xa8, 0x9b, 0x83, 0x24, 0xce, 0xed, 0x3b, 0x27, 0x5a, 0x3e, 0x96, 0xdd, 0x1d, 0x9d, 0xdb,
	0x12, 0xad, 0x23, 0x5e, 0x8e, 0xf7, 0x0c, 0x59, 0x8b, 0xc6, 0xe5, 0x64, 0x9b, 0xda, 0x78, 0x27,
	0xc5, 0x32, 0x37, 0xd8, 0x46, 0x82, 0x5c, 0x22, 0x37, 0xd1, 0xe4, 0xdb, 0x83, 0xd1, 0x6d, 0x32,
	0x53, 0xf7, 0xe0, 0xa5, 0x85, 0x93, 0x70, 0xe3, 0x68, 0x74, 0x1c, 0x9c, 0x9e, 0xeb, 0x75, 0x94,
	0xfa, 0xaa, 0xe6, 0x3f, 0xd9, 0x14, 0xf9, 0x11, 0xc5, 0x45, 0x35, 0x4a, 0xbd, 0x42, 0x50, 0xa9,
	0x28, 0xce, 0x48, 0x9c, 0x84, 0x9b, 0x43, 0xc8, 0x0f, 0x15, 0xe0, 0x9a, 0xc4, 0x35, 0x8a, 0x08,
	0xa4, 0x4d, 0x88, 0x7a, 0x01, 0xdf, 0x64, 0xc9, 0x02, 0xe7, 0x64, 0x25, 0x1c, 0x0d, 0xa1, 0x4f,
	0xeb, 0xe7, 0x48, 0xf6, 0x97, 0xde, 0xf3, 0xd4, 0x17, 0x1c, 0xf6, 0xcd, 0xc7, 0xfd, 0x97, 0x4a,
	0xe8, 0xd5, 0xa6, 0x8b, 0x81, 0x73, 0x4c, 0x3b, 0x52, 0xb4, 0x2f, 0xab, 0x49, 0x51, 0x25, 0x1c,
	0x98, 0xb6, 0xad, 0x7f, 0xde, 0xad, 0x21, 0xde, 0x6e, 0xc2, 0xbf, 0x5e, 0xb3, 0x9a, 0x94, 0xc9,
	0x1b, 0xec, 0xde, 0x54, 0x24, 0x71, 0x5c, 0x18, 0x57, 0x30, 0xaa, 0x3b, 0xd8, 0x6e, 0x56, 0xa5,
	0xdd, 0x8a, 0xb3, 0xf5, 0x67, 0x46, 0x8e, 0x5a, 0xc8, 0x65, 0xf0, 0xec, 0x77, 0xc5, 0x9f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x97, 0x9c, 0x14, 0x81, 0xdf, 0x02, 0x00, 0x00,
}
