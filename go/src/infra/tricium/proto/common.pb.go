// Code generated by protoc-gen-go.
// source: infra/tricium/proto/common.proto
// DO NOT EDIT!

/*
Package tricium is a generated protocol buffer package.

It is generated from these files:
	infra/tricium/proto/common.proto
	infra/tricium/proto/data.proto
	infra/tricium/proto/tricium.proto
	infra/tricium/proto/workflow.proto

It has these top-level messages:
	Cmd
	CipdPackage
	Data
	ServiceConfig
	ProjectConfig
	RepoDetails
	GitRepoDetails
	Acl
	Platform
	Selection
	Analyzer
	ConfigDef
	Impl
	Recipe
	Config
	Workflow
	Worker
*/
package tricium

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

// Specification of a command.
type Cmd struct {
	// Executable binary.
	Exec string `protobuf:"bytes,1,opt,name=exec" json:"exec,omitempty"`
	// Arguments in order.
	Arg []string `protobuf:"bytes,2,rep,name=arg" json:"arg,omitempty"`
}

func (m *Cmd) Reset()                    { *m = Cmd{} }
func (m *Cmd) String() string            { return proto.CompactTextString(m) }
func (*Cmd) ProtoMessage()               {}
func (*Cmd) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// CIPD package.
type CipdPackage struct {
	// CIPD package name.
	PackageName string `protobuf:"bytes,1,opt,name=package_name,json=packageName" json:"package_name,omitempty"`
	// Path to directory, relative to the working directory, where to install
	// package. Cannot be empty or start with a slash.
	Path string `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	// Package version.
	Version string `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
}

func (m *CipdPackage) Reset()                    { *m = CipdPackage{} }
func (m *CipdPackage) String() string            { return proto.CompactTextString(m) }
func (*CipdPackage) ProtoMessage()               {}
func (*CipdPackage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*Cmd)(nil), "tricium.Cmd")
	proto.RegisterType((*CipdPackage)(nil), "tricium.CipdPackage")
}

func init() { proto.RegisterFile("infra/tricium/proto/common.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 167 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x34, 0x8e, 0x41, 0x0a, 0xc2, 0x30,
	0x10, 0x45, 0x69, 0x23, 0x96, 0x4e, 0x5d, 0x48, 0x56, 0x59, 0xd6, 0xae, 0x0a, 0x82, 0x5d, 0x78,
	0x84, 0xee, 0x45, 0x7a, 0x00, 0x65, 0x4c, 0xc7, 0x1a, 0x64, 0x92, 0x10, 0xab, 0x78, 0x7c, 0x69,
	0x5a, 0x77, 0xef, 0xff, 0x61, 0x1e, 0x1f, 0x4a, 0x63, 0xef, 0x01, 0x9b, 0x31, 0x18, 0x6d, 0xde,
	0xdc, 0xf8, 0xe0, 0x46, 0xd7, 0x68, 0xc7, 0xec, 0xec, 0x21, 0x06, 0x99, 0x2d, 0xb7, 0x6a, 0x0f,
	0xa2, 0xe5, 0x5e, 0x4a, 0x58, 0xd1, 0x97, 0xb4, 0x4a, 0xca, 0xa4, 0xce, 0xbb, 0xc8, 0x72, 0x0b,
	0x02, 0xc3, 0xa0, 0xd2, 0x52, 0xd4, 0x79, 0x37, 0x61, 0x75, 0x81, 0xa2, 0x35, 0xbe, 0x3f, 0xa3,
	0x7e, 0xe2, 0x40, 0x72, 0x07, 0x1b, 0x3f, 0xe3, 0xd5, 0x22, 0xd3, 0xf2, 0x5c, 0x2c, 0xdd, 0x09,
	0x99, 0x26, 0xaf, 0xc7, 0xf1, 0xa1, 0xd2, 0xd9, 0x3b, 0xb1, 0x54, 0x90, 0x7d, 0x28, 0xbc, 0x8c,
	0xb3, 0x4a, 0xc4, 0xfa, 0x1f, 0x6f, 0xeb, 0x38, 0xee, 0xf8, 0x0b, 0x00, 0x00, 0xff, 0xff, 0x69,
	0x68, 0xd5, 0x38, 0xc0, 0x00, 0x00, 0x00,
}
