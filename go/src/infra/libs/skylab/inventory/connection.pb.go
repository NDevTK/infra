// Copyright 2018 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: connection.proto

package inventory

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// NEXT TAG: 6
type ServoHostConnection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServoHostId *string `protobuf:"bytes,1,req,name=servo_host_id,json=servoHostId" json:"servo_host_id,omitempty"`
	DutId       *string `protobuf:"bytes,2,req,name=dut_id,json=dutId" json:"dut_id,omitempty"`
	ServoPort   *int32  `protobuf:"varint,3,req,name=servo_port,json=servoPort" json:"servo_port,omitempty"`
	ServoSerial *string `protobuf:"bytes,4,opt,name=servo_serial,json=servoSerial" json:"servo_serial,omitempty"`
	ServoType   *string `protobuf:"bytes,5,opt,name=servo_type,json=servoType" json:"servo_type,omitempty"`
}

func (x *ServoHostConnection) Reset() {
	*x = ServoHostConnection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connection_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServoHostConnection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServoHostConnection) ProtoMessage() {}

func (x *ServoHostConnection) ProtoReflect() protoreflect.Message {
	mi := &file_connection_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServoHostConnection.ProtoReflect.Descriptor instead.
func (*ServoHostConnection) Descriptor() ([]byte, []int) {
	return file_connection_proto_rawDescGZIP(), []int{0}
}

func (x *ServoHostConnection) GetServoHostId() string {
	if x != nil && x.ServoHostId != nil {
		return *x.ServoHostId
	}
	return ""
}

func (x *ServoHostConnection) GetDutId() string {
	if x != nil && x.DutId != nil {
		return *x.DutId
	}
	return ""
}

func (x *ServoHostConnection) GetServoPort() int32 {
	if x != nil && x.ServoPort != nil {
		return *x.ServoPort
	}
	return 0
}

func (x *ServoHostConnection) GetServoSerial() string {
	if x != nil && x.ServoSerial != nil {
		return *x.ServoSerial
	}
	return ""
}

func (x *ServoHostConnection) GetServoType() string {
	if x != nil && x.ServoType != nil {
		return *x.ServoType
	}
	return ""
}

// NEXT TAG: 3
type ChameleonConnection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chameleon        *ChameleonDevice `protobuf:"bytes,1,req,name=chameleon" json:"chameleon,omitempty"`
	ControlledDevice *Device          `protobuf:"bytes,2,req,name=controlled_device,json=controlledDevice" json:"controlled_device,omitempty"`
}

func (x *ChameleonConnection) Reset() {
	*x = ChameleonConnection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connection_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChameleonConnection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChameleonConnection) ProtoMessage() {}

func (x *ChameleonConnection) ProtoReflect() protoreflect.Message {
	mi := &file_connection_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChameleonConnection.ProtoReflect.Descriptor instead.
func (*ChameleonConnection) Descriptor() ([]byte, []int) {
	return file_connection_proto_rawDescGZIP(), []int{1}
}

func (x *ChameleonConnection) GetChameleon() *ChameleonDevice {
	if x != nil {
		return x.Chameleon
	}
	return nil
}

func (x *ChameleonConnection) GetControlledDevice() *Device {
	if x != nil {
		return x.ControlledDevice
	}
	return nil
}

var File_connection_proto protoreflect.FileDescriptor

var file_connection_proto_rawDesc = []byte{
	0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x2c, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x65, 0x6f, 0x73, 0x5f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2e, 0x73, 0x6b, 0x79, 0x6c, 0x61, 0x62,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
	0x1a, 0x0c, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb1,
	0x01, 0x0a, 0x13, 0x53, 0x65, 0x72, 0x76, 0x6f, 0x48, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f,
	0x68, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x0b, 0x73,
	0x65, 0x72, 0x76, 0x6f, 0x48, 0x6f, 0x73, 0x74, 0x49, 0x64, 0x12, 0x15, 0x0a, 0x06, 0x64, 0x75,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x05, 0x64, 0x75, 0x74, 0x49,
	0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18,
	0x03, 0x20, 0x02, 0x28, 0x05, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x50, 0x6f, 0x72, 0x74,
	0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x53, 0x65, 0x72,
	0x69, 0x61, 0x6c, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x54, 0x79,
	0x70, 0x65, 0x22, 0xd5, 0x01, 0x0a, 0x13, 0x43, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x5b, 0x0a, 0x09, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x3d, 0x2e,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x5f,
	0x69, 0x6e, 0x66, 0x72, 0x61, 0x2e, 0x73, 0x6b, 0x79, 0x6c, 0x61, 0x62, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x43, 0x68, 0x61,
	0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x09, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x12, 0x61, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x64, 0x5f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x0b, 0x32, 0x34, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x6f, 0x73, 0x5f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2e, 0x73, 0x6b, 0x79, 0x6c, 0x61,
	0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72,
	0x79, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x64, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x3b,
	0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79,
}

var (
	file_connection_proto_rawDescOnce sync.Once
	file_connection_proto_rawDescData = file_connection_proto_rawDesc
)

func file_connection_proto_rawDescGZIP() []byte {
	file_connection_proto_rawDescOnce.Do(func() {
		file_connection_proto_rawDescData = protoimpl.X.CompressGZIP(file_connection_proto_rawDescData)
	})
	return file_connection_proto_rawDescData
}

var file_connection_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_connection_proto_goTypes = []interface{}{
	(*ServoHostConnection)(nil), // 0: chrome.chromeos_infra.skylab.proto.inventory.ServoHostConnection
	(*ChameleonConnection)(nil), // 1: chrome.chromeos_infra.skylab.proto.inventory.ChameleonConnection
	(*ChameleonDevice)(nil),     // 2: chrome.chromeos_infra.skylab.proto.inventory.ChameleonDevice
	(*Device)(nil),              // 3: chrome.chromeos_infra.skylab.proto.inventory.Device
}
var file_connection_proto_depIdxs = []int32{
	2, // 0: chrome.chromeos_infra.skylab.proto.inventory.ChameleonConnection.chameleon:type_name -> chrome.chromeos_infra.skylab.proto.inventory.ChameleonDevice
	3, // 1: chrome.chromeos_infra.skylab.proto.inventory.ChameleonConnection.controlled_device:type_name -> chrome.chromeos_infra.skylab.proto.inventory.Device
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_connection_proto_init() }
func file_connection_proto_init() {
	if File_connection_proto != nil {
		return
	}
	file_device_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_connection_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServoHostConnection); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_connection_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChameleonConnection); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connection_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_connection_proto_goTypes,
		DependencyIndexes: file_connection_proto_depIdxs,
		MessageInfos:      file_connection_proto_msgTypes,
	}.Build()
	File_connection_proto = out.File
	file_connection_proto_rawDesc = nil
	file_connection_proto_goTypes = nil
	file_connection_proto_depIdxs = nil
}
