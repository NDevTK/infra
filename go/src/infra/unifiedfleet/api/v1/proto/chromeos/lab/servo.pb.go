// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/unifiedfleet/api/v1/proto/chromeos/lab/servo.proto

package ufspb

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

// NEXT TAG: 6
type Servo struct {
	// Servo-specific configs
	ServoHostname string `protobuf:"bytes,2,opt,name=servo_hostname,json=servoHostname,proto3" json:"servo_hostname,omitempty"`
	ServoPort     int32  `protobuf:"varint,3,opt,name=servo_port,json=servoPort,proto3" json:"servo_port,omitempty"`
	ServoSerial   string `protobuf:"bytes,4,opt,name=servo_serial,json=servoSerial,proto3" json:"servo_serial,omitempty"`
	// Based on https://docs.google.com/document/d/1TPp7yp-uwFUh5xOnBLI4jPYtYD7IcdyQ1dgqFqtcJEU/edit?ts=5d8eafb7#heading=h.csdfk1i6g0l
	// servo_type will contain different setup of servos. So string is recommended than enum.
	ServoType            string   `protobuf:"bytes,5,opt,name=servo_type,json=servoType,proto3" json:"servo_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Servo) Reset()         { *m = Servo{} }
func (m *Servo) String() string { return proto.CompactTextString(m) }
func (*Servo) ProtoMessage()    {}
func (*Servo) Descriptor() ([]byte, []int) {
	return fileDescriptor_a84aa1af8ad613f1, []int{0}
}

func (m *Servo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Servo.Unmarshal(m, b)
}
func (m *Servo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Servo.Marshal(b, m, deterministic)
}
func (m *Servo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Servo.Merge(m, src)
}
func (m *Servo) XXX_Size() int {
	return xxx_messageInfo_Servo.Size(m)
}
func (m *Servo) XXX_DiscardUnknown() {
	xxx_messageInfo_Servo.DiscardUnknown(m)
}

var xxx_messageInfo_Servo proto.InternalMessageInfo

func (m *Servo) GetServoHostname() string {
	if m != nil {
		return m.ServoHostname
	}
	return ""
}

func (m *Servo) GetServoPort() int32 {
	if m != nil {
		return m.ServoPort
	}
	return 0
}

func (m *Servo) GetServoSerial() string {
	if m != nil {
		return m.ServoSerial
	}
	return ""
}

func (m *Servo) GetServoType() string {
	if m != nil {
		return m.ServoType
	}
	return ""
}

func init() {
	proto.RegisterType((*Servo)(nil), "unifiedfleet.api.v1.proto.chromeos.lab.Servo")
}

func init() {
	proto.RegisterFile("infra/unifiedfleet/api/v1/proto/chromeos/lab/servo.proto", fileDescriptor_a84aa1af8ad613f1)
}

var fileDescriptor_a84aa1af8ad613f1 = []byte{
	// 211 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x8f, 0xbf, 0x4b, 0xc7, 0x30,
	0x10, 0x47, 0xa9, 0x5a, 0xe1, 0x1b, 0x7f, 0x0c, 0x99, 0xb2, 0x08, 0x55, 0x50, 0x3a, 0x25, 0x54,
	0x1d, 0x04, 0x37, 0x27, 0x47, 0x69, 0x9d, 0x5c, 0x24, 0xd1, 0x0b, 0x0d, 0xa4, 0xbd, 0x90, 0xa4,
	0x85, 0xfe, 0x15, 0xfe, 0xcb, 0xe2, 0x45, 0xd1, 0xd5, 0xf5, 0xbd, 0xfb, 0x3c, 0x38, 0x76, 0xe7,
	0x66, 0x1b, 0xb5, 0x5a, 0x66, 0x67, 0x1d, 0xbc, 0x5b, 0x0f, 0x90, 0x95, 0x0e, 0x4e, 0xad, 0x9d,
	0x0a, 0x11, 0x33, 0xaa, 0xb7, 0x31, 0xe2, 0x04, 0x98, 0x94, 0xd7, 0x46, 0x25, 0x88, 0x2b, 0x4a,
	0x12, 0xfc, 0xea, 0xef, 0x46, 0xea, 0xe0, 0xe4, 0xda, 0x15, 0x25, 0x7f, 0x36, 0xd2, 0x6b, 0x73,
	0xf1, 0x51, 0xb1, 0x7a, 0xf8, 0xda, 0xf1, 0x4b, 0x76, 0x4a, 0x81, 0xd7, 0x11, 0x53, 0x9e, 0xf5,
	0x04, 0x62, 0xaf, 0xa9, 0xda, 0x5d, 0x7f, 0x42, 0xf4, 0xf1, 0x1b, 0xf2, 0x33, 0xc6, 0xca, 0x59,
	0xc0, 0x98, 0xc5, 0x7e, 0x53, 0xb5, 0x75, 0xbf, 0x23, 0xf2, 0x84, 0x31, 0xf3, 0x73, 0x76, 0x5c,
	0x74, 0x82, 0xe8, 0xb4, 0x17, 0x07, 0xd4, 0x38, 0x22, 0x36, 0x10, 0xfa, 0x2d, 0xe4, 0x2d, 0x80,
	0xa8, 0xe9, 0xa0, 0x14, 0x9e, 0xb7, 0x00, 0x0f, 0xb7, 0x2f, 0xd7, 0xff, 0xf9, 0xfa, 0x7e, 0xb1,
	0x29, 0x18, 0x73, 0x48, 0xe6, 0xe6, 0x33, 0x00, 0x00, 0xff, 0xff, 0x21, 0xa8, 0xf0, 0x02, 0x32,
	0x01, 0x00, 0x00,
}
