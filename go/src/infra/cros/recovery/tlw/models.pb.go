// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: infra/cros/recovery/tlw/models.proto

package tlw

import (
	xmlrpc "go.chromium.org/chromiumos/config/go/api/test/xmlrpc"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// RunRequest represents result of executed command.
type RunRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource name
	Resource string `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	// Command executed on the resource.
	Command string `protobuf:"bytes,2,opt,name=command,proto3" json:"command,omitempty"`
	// Command arguments.
	Args    []string             `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
	Timeout *durationpb.Duration `protobuf:"bytes,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
}

func (x *RunRequest) Reset() {
	*x = RunRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunRequest) ProtoMessage() {}

func (x *RunRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunRequest.ProtoReflect.Descriptor instead.
func (*RunRequest) Descriptor() ([]byte, []int) {
	return file_infra_cros_recovery_tlw_models_proto_rawDescGZIP(), []int{0}
}

func (x *RunRequest) GetResource() string {
	if x != nil {
		return x.Resource
	}
	return ""
}

func (x *RunRequest) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *RunRequest) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *RunRequest) GetTimeout() *durationpb.Duration {
	if x != nil {
		return x.Timeout
	}
	return nil
}

// ProvisionRequest provides data to perform provisioning of the device.
type ProvisionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource name
	Resource string `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	// Path to system image.
	// Path to the GS file.
	// Example: gs://bucket/file_name
	SystemImagePath string `protobuf:"bytes,2,opt,name=system_image_path,json=systemImagePath,proto3" json:"system_image_path,omitempty"`
	// Prevent reboot during provision OS.
	PreventReboot bool `protobuf:"varint,3,opt,name=prevent_reboot,json=preventReboot,proto3" json:"prevent_reboot,omitempty"`
}

func (x *ProvisionRequest) Reset() {
	*x = ProvisionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProvisionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProvisionRequest) ProtoMessage() {}

func (x *ProvisionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProvisionRequest.ProtoReflect.Descriptor instead.
func (*ProvisionRequest) Descriptor() ([]byte, []int) {
	return file_infra_cros_recovery_tlw_models_proto_rawDescGZIP(), []int{1}
}

func (x *ProvisionRequest) GetResource() string {
	if x != nil {
		return x.Resource
	}
	return ""
}

func (x *ProvisionRequest) GetSystemImagePath() string {
	if x != nil {
		return x.SystemImagePath
	}
	return ""
}

func (x *ProvisionRequest) GetPreventReboot() bool {
	if x != nil {
		return x.PreventReboot
	}
	return false
}

// CallBluetoothPeerRequest represents data to run command on bluetooth peer.
type CallBluetoothPeerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource string          `protobuf:"bytes,1,opt,name=Resource,proto3" json:"Resource,omitempty"`
	Method   string          `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Args     []*xmlrpc.Value `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
}

func (x *CallBluetoothPeerRequest) Reset() {
	*x = CallBluetoothPeerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallBluetoothPeerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallBluetoothPeerRequest) ProtoMessage() {}

func (x *CallBluetoothPeerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallBluetoothPeerRequest.ProtoReflect.Descriptor instead.
func (*CallBluetoothPeerRequest) Descriptor() ([]byte, []int) {
	return file_infra_cros_recovery_tlw_models_proto_rawDescGZIP(), []int{2}
}

func (x *CallBluetoothPeerRequest) GetResource() string {
	if x != nil {
		return x.Resource
	}
	return ""
}

func (x *CallBluetoothPeerRequest) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *CallBluetoothPeerRequest) GetArgs() []*xmlrpc.Value {
	if x != nil {
		return x.Args
	}
	return nil
}

// CallBluetoothPeerResponse represents result data from running command on
// bluetooth peer.
type CallBluetoothPeerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value *xmlrpc.Value `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Fault bool          `protobuf:"varint,2,opt,name=fault,proto3" json:"fault,omitempty"`
}

func (x *CallBluetoothPeerResponse) Reset() {
	*x = CallBluetoothPeerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallBluetoothPeerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallBluetoothPeerResponse) ProtoMessage() {}

func (x *CallBluetoothPeerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_recovery_tlw_models_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallBluetoothPeerResponse.ProtoReflect.Descriptor instead.
func (*CallBluetoothPeerResponse) Descriptor() ([]byte, []int) {
	return file_infra_cros_recovery_tlw_models_proto_rawDescGZIP(), []int{3}
}

func (x *CallBluetoothPeerResponse) GetValue() *xmlrpc.Value {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *CallBluetoothPeerResponse) GetFault() bool {
	if x != nil {
		return x.Fault
	}
	return false
}

var File_infra_cros_recovery_tlw_models_proto protoreflect.FileDescriptor

var file_infra_cros_recovery_tlw_models_proto_rawDesc = []byte{
	0x0a, 0x24, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x72, 0x65, 0x63,
	0x6f, 0x76, 0x65, 0x72, 0x79, 0x2f, 0x74, 0x6c, 0x77, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73,
	0x2e, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x56, 0x67, 0x6f, 0x2e, 0x63, 0x68,
	0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x78, 0x6d,
	0x6c, 0x72, 0x70, 0x63, 0x2f, 0x78, 0x6d, 0x6c, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x8b, 0x01, 0x0a, 0x0a, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x33, 0x0a, 0x07, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x22,
	0x81, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x12, 0x2a, 0x0a, 0x11, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x50, 0x61, 0x74, 0x68, 0x12, 0x25, 0x0a, 0x0e,
	0x70, 0x72, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x62, 0x6f, 0x6f, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x62,
	0x6f, 0x6f, 0x74, 0x22, 0x8c, 0x01, 0x0a, 0x18, 0x43, 0x61, 0x6c, 0x6c, 0x42, 0x6c, 0x75, 0x65,
	0x74, 0x6f, 0x6f, 0x74, 0x68, 0x50, 0x65, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x12, 0x3c, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x28, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2e,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x78, 0x6d, 0x6c, 0x72, 0x70, 0x63, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x04, 0x61, 0x72,
	0x67, 0x73, 0x22, 0x71, 0x0a, 0x19, 0x43, 0x61, 0x6c, 0x6c, 0x42, 0x6c, 0x75, 0x65, 0x74, 0x6f,
	0x6f, 0x74, 0x68, 0x50, 0x65, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x3e, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28,
	0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x6f, 0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x78, 0x6d, 0x6c, 0x72,
	0x70, 0x63, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x42, 0x1d, 0x5a, 0x1b, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63,
	0x72, 0x6f, 0x73, 0x2f, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2f, 0x74, 0x6c, 0x77,
	0x3b, 0x74, 0x6c, 0x77, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_cros_recovery_tlw_models_proto_rawDescOnce sync.Once
	file_infra_cros_recovery_tlw_models_proto_rawDescData = file_infra_cros_recovery_tlw_models_proto_rawDesc
)

func file_infra_cros_recovery_tlw_models_proto_rawDescGZIP() []byte {
	file_infra_cros_recovery_tlw_models_proto_rawDescOnce.Do(func() {
		file_infra_cros_recovery_tlw_models_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cros_recovery_tlw_models_proto_rawDescData)
	})
	return file_infra_cros_recovery_tlw_models_proto_rawDescData
}

var file_infra_cros_recovery_tlw_models_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_infra_cros_recovery_tlw_models_proto_goTypes = []interface{}{
	(*RunRequest)(nil),                // 0: chromeos.recovery.RunRequest
	(*ProvisionRequest)(nil),          // 1: chromeos.recovery.ProvisionRequest
	(*CallBluetoothPeerRequest)(nil),  // 2: chromeos.recovery.CallBluetoothPeerRequest
	(*CallBluetoothPeerResponse)(nil), // 3: chromeos.recovery.CallBluetoothPeerResponse
	(*durationpb.Duration)(nil),       // 4: google.protobuf.Duration
	(*xmlrpc.Value)(nil),              // 5: chromiumos.config.api.test.xmlrpc.Value
}
var file_infra_cros_recovery_tlw_models_proto_depIdxs = []int32{
	4, // 0: chromeos.recovery.RunRequest.timeout:type_name -> google.protobuf.Duration
	5, // 1: chromeos.recovery.CallBluetoothPeerRequest.args:type_name -> chromiumos.config.api.test.xmlrpc.Value
	5, // 2: chromeos.recovery.CallBluetoothPeerResponse.value:type_name -> chromiumos.config.api.test.xmlrpc.Value
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_infra_cros_recovery_tlw_models_proto_init() }
func file_infra_cros_recovery_tlw_models_proto_init() {
	if File_infra_cros_recovery_tlw_models_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cros_recovery_tlw_models_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunRequest); i {
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
		file_infra_cros_recovery_tlw_models_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProvisionRequest); i {
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
		file_infra_cros_recovery_tlw_models_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallBluetoothPeerRequest); i {
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
		file_infra_cros_recovery_tlw_models_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallBluetoothPeerResponse); i {
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
			RawDescriptor: file_infra_cros_recovery_tlw_models_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_cros_recovery_tlw_models_proto_goTypes,
		DependencyIndexes: file_infra_cros_recovery_tlw_models_proto_depIdxs,
		MessageInfos:      file_infra_cros_recovery_tlw_models_proto_msgTypes,
	}.Build()
	File_infra_cros_recovery_tlw_models_proto = out.File
	file_infra_cros_recovery_tlw_models_proto_rawDesc = nil
	file_infra_cros_recovery_tlw_models_proto_goTypes = nil
	file_infra_cros_recovery_tlw_models_proto_depIdxs = nil
}
