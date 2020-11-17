// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: infra/unifiedfleet/api/v1/models/chromeos/lab/dut_state.proto

package ufspb

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

// Next Tag: 16
type PeripheralState int32

const (
	// Please keep for all unknown states.
	PeripheralState_UNKNOWN PeripheralState = 0
	// Device and software on it is working as expected.
	PeripheralState_WORKING PeripheralState = 1
	// Configuration for device is not provided.
	PeripheralState_MISSING_CONFIG PeripheralState = 5
	// Configuration contains incorrect information.
	PeripheralState_WRONG_CONFIG PeripheralState = 4
	// Device is not connected/plugged.
	PeripheralState_NOT_CONNECTED PeripheralState = 2
	// Device is not reachable over ssh.
	PeripheralState_NO_SSH PeripheralState = 6
	// Device is broken or not working as expected. the state used if no specified state for the issue.
	PeripheralState_BROKEN PeripheralState = 3
	// Device cannot be repaired or required manual attention to fix/replace it.
	PeripheralState_NEED_REPLACEMENT PeripheralState = 7
	// Servo specific states.
	// cr50 console missing or unresponsive.
	PeripheralState_CR50_CONSOLE_MISSING PeripheralState = 13
	// Servod daemon cannot start on servo-host because cr50 testlab not enabled.
	PeripheralState_CCD_TESTLAB_ISSUE PeripheralState = 8
	// Servod daemon cannot start on servo-host.
	PeripheralState_SERVOD_ISSUE PeripheralState = 9
	// device lid is not open.
	PeripheralState_LID_OPEN_FAILED PeripheralState = 10
	// the ribbon cable between servo and DUT is broken or not connected.
	PeripheralState_BAD_RIBBON_CABLE PeripheralState = 11
	// the EC on the DUT has issue.
	PeripheralState_EC_BROKEN PeripheralState = 12
	// Servo is not connected to the DUT.
	PeripheralState_DUT_NOT_CONNECTED PeripheralState = 14
	// Some component in servo-topology missed or not detected.
	PeripheralState_TOPOLOGY_ISSUE PeripheralState = 15
)

// Enum value maps for PeripheralState.
var (
	PeripheralState_name = map[int32]string{
		0:  "UNKNOWN",
		1:  "WORKING",
		5:  "MISSING_CONFIG",
		4:  "WRONG_CONFIG",
		2:  "NOT_CONNECTED",
		6:  "NO_SSH",
		3:  "BROKEN",
		7:  "NEED_REPLACEMENT",
		13: "CR50_CONSOLE_MISSING",
		8:  "CCD_TESTLAB_ISSUE",
		9:  "SERVOD_ISSUE",
		10: "LID_OPEN_FAILED",
		11: "BAD_RIBBON_CABLE",
		12: "EC_BROKEN",
		14: "DUT_NOT_CONNECTED",
		15: "TOPOLOGY_ISSUE",
	}
	PeripheralState_value = map[string]int32{
		"UNKNOWN":              0,
		"WORKING":              1,
		"MISSING_CONFIG":       5,
		"WRONG_CONFIG":         4,
		"NOT_CONNECTED":        2,
		"NO_SSH":               6,
		"BROKEN":               3,
		"NEED_REPLACEMENT":     7,
		"CR50_CONSOLE_MISSING": 13,
		"CCD_TESTLAB_ISSUE":    8,
		"SERVOD_ISSUE":         9,
		"LID_OPEN_FAILED":      10,
		"BAD_RIBBON_CABLE":     11,
		"EC_BROKEN":            12,
		"DUT_NOT_CONNECTED":    14,
		"TOPOLOGY_ISSUE":       15,
	}
)

func (x PeripheralState) Enum() *PeripheralState {
	p := new(PeripheralState)
	*p = x
	return p
}

func (x PeripheralState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PeripheralState) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[0].Descriptor()
}

func (PeripheralState) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[0]
}

func (x PeripheralState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PeripheralState.Descriptor instead.
func (PeripheralState) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP(), []int{0}
}

// The states are using for DUT storage and USB-drive on servo.
// Next Tag: 5
type HardwareState int32

const (
	// keep for all unknown state by default.
	HardwareState_HARDWARE_UNKNOWN HardwareState = 0
	// Hardware is in good shape and pass all verifiers.
	HardwareState_HARDWARE_NORMAL HardwareState = 1
	// Hardware is still good but some not critical verifiers did not pass or provided border values.
	// (used for DUT storage when usage reached 98%)
	HardwareState_HARDWARE_ACCEPTABLE HardwareState = 2
	// Hardware is broken or bad (did not pass verifiers).
	HardwareState_HARDWARE_NEED_REPLACEMENT HardwareState = 3
	// Hardware is not detected to run verifiers.
	// (used for USB-drive when it expected but not detected on the device)
	HardwareState_HARDWARE_NOT_DETECTED HardwareState = 4
)

// Enum value maps for HardwareState.
var (
	HardwareState_name = map[int32]string{
		0: "HARDWARE_UNKNOWN",
		1: "HARDWARE_NORMAL",
		2: "HARDWARE_ACCEPTABLE",
		3: "HARDWARE_NEED_REPLACEMENT",
		4: "HARDWARE_NOT_DETECTED",
	}
	HardwareState_value = map[string]int32{
		"HARDWARE_UNKNOWN":          0,
		"HARDWARE_NORMAL":           1,
		"HARDWARE_ACCEPTABLE":       2,
		"HARDWARE_NEED_REPLACEMENT": 3,
		"HARDWARE_NOT_DETECTED":     4,
	}
)

func (x HardwareState) Enum() *HardwareState {
	p := new(HardwareState)
	*p = x
	return p
}

func (x HardwareState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HardwareState) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[1].Descriptor()
}

func (HardwareState) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[1]
}

func (x HardwareState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HardwareState.Descriptor instead.
func (HardwareState) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP(), []int{1}
}

// CR50-related configs by definition shouldn't be a state config, but a build config.
// However, we don't have a way to source it from any external configuration system,
// and it's changed frequently enough to handle cr50 tests, which makes
// it basically impossible for manual updatings: See crbug.com/1057145 for the
// troubles it causes.
//
// So we temporarily set it in state config so that repair job can update it.
// For further changes of it, please see tracking bug crbug.com/1057719.
//
// phases for cr50 module. Next Tag: 3
type DutState_CR50Phase int32

const (
	DutState_CR50_PHASE_INVALID DutState_CR50Phase = 0
	DutState_CR50_PHASE_PREPVT  DutState_CR50Phase = 1
	DutState_CR50_PHASE_PVT     DutState_CR50Phase = 2
)

// Enum value maps for DutState_CR50Phase.
var (
	DutState_CR50Phase_name = map[int32]string{
		0: "CR50_PHASE_INVALID",
		1: "CR50_PHASE_PREPVT",
		2: "CR50_PHASE_PVT",
	}
	DutState_CR50Phase_value = map[string]int32{
		"CR50_PHASE_INVALID": 0,
		"CR50_PHASE_PREPVT":  1,
		"CR50_PHASE_PVT":     2,
	}
)

func (x DutState_CR50Phase) Enum() *DutState_CR50Phase {
	p := new(DutState_CR50Phase)
	*p = x
	return p
}

func (x DutState_CR50Phase) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DutState_CR50Phase) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[2].Descriptor()
}

func (DutState_CR50Phase) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[2]
}

func (x DutState_CR50Phase) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DutState_CR50Phase.Descriptor instead.
func (DutState_CR50Phase) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP(), []int{0, 0}
}

// key env for cr50 RW version. Next Tag: 3
type DutState_CR50KeyEnv int32

const (
	DutState_CR50_KEYENV_INVALID DutState_CR50KeyEnv = 0
	DutState_CR50_KEYENV_PROD    DutState_CR50KeyEnv = 1
	DutState_CR50_KEYENV_DEV     DutState_CR50KeyEnv = 2
)

// Enum value maps for DutState_CR50KeyEnv.
var (
	DutState_CR50KeyEnv_name = map[int32]string{
		0: "CR50_KEYENV_INVALID",
		1: "CR50_KEYENV_PROD",
		2: "CR50_KEYENV_DEV",
	}
	DutState_CR50KeyEnv_value = map[string]int32{
		"CR50_KEYENV_INVALID": 0,
		"CR50_KEYENV_PROD":    1,
		"CR50_KEYENV_DEV":     2,
	}
)

func (x DutState_CR50KeyEnv) Enum() *DutState_CR50KeyEnv {
	p := new(DutState_CR50KeyEnv)
	*p = x
	return p
}

func (x DutState_CR50KeyEnv) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DutState_CR50KeyEnv) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[3].Descriptor()
}

func (DutState_CR50KeyEnv) Type() protoreflect.EnumType {
	return &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes[3]
}

func (x DutState_CR50KeyEnv) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DutState_CR50KeyEnv.Descriptor instead.
func (DutState_CR50KeyEnv) EnumDescriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP(), []int{0, 1}
}

// Next Tag: 10
type DutState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                  *ChromeOSDeviceID `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Servo               PeripheralState   `protobuf:"varint,2,opt,name=servo,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.PeripheralState" json:"servo,omitempty"`
	Chameleon           PeripheralState   `protobuf:"varint,3,opt,name=chameleon,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.PeripheralState" json:"chameleon,omitempty"`
	AudioLoopbackDongle PeripheralState   `protobuf:"varint,4,opt,name=audio_loopback_dongle,json=audioLoopbackDongle,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.PeripheralState" json:"audio_loopback_dongle,omitempty"`
	// Indicate how many working bluetooth btpeer for a device.
	WorkingBluetoothBtpeer int32              `protobuf:"varint,5,opt,name=working_bluetooth_btpeer,json=workingBluetoothBtpeer,proto3" json:"working_bluetooth_btpeer,omitempty"`
	Cr50Phase              DutState_CR50Phase `protobuf:"varint,6,opt,name=cr50_phase,json=cr50Phase,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.DutState_CR50Phase" json:"cr50_phase,omitempty"`
	// Detected based on the cr50 RW version that the DUT is running on.
	Cr50KeyEnv DutState_CR50KeyEnv `protobuf:"varint,7,opt,name=cr50_key_env,json=cr50KeyEnv,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.DutState_CR50KeyEnv" json:"cr50_key_env,omitempty"`
	// Detected during running admin_audit task.
	StorageState  HardwareState `protobuf:"varint,8,opt,name=storage_state,json=storageState,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.HardwareState" json:"storage_state,omitempty"`
	ServoUsbState HardwareState `protobuf:"varint,9,opt,name=servo_usb_state,json=servoUsbState,proto3,enum=unifiedfleet.api.v1.models.chromeos.lab.HardwareState" json:"servo_usb_state,omitempty"`
}

func (x *DutState) Reset() {
	*x = DutState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DutState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DutState) ProtoMessage() {}

func (x *DutState) ProtoReflect() protoreflect.Message {
	mi := &file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DutState.ProtoReflect.Descriptor instead.
func (*DutState) Descriptor() ([]byte, []int) {
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP(), []int{0}
}

func (x *DutState) GetId() *ChromeOSDeviceID {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *DutState) GetServo() PeripheralState {
	if x != nil {
		return x.Servo
	}
	return PeripheralState_UNKNOWN
}

func (x *DutState) GetChameleon() PeripheralState {
	if x != nil {
		return x.Chameleon
	}
	return PeripheralState_UNKNOWN
}

func (x *DutState) GetAudioLoopbackDongle() PeripheralState {
	if x != nil {
		return x.AudioLoopbackDongle
	}
	return PeripheralState_UNKNOWN
}

func (x *DutState) GetWorkingBluetoothBtpeer() int32 {
	if x != nil {
		return x.WorkingBluetoothBtpeer
	}
	return 0
}

func (x *DutState) GetCr50Phase() DutState_CR50Phase {
	if x != nil {
		return x.Cr50Phase
	}
	return DutState_CR50_PHASE_INVALID
}

func (x *DutState) GetCr50KeyEnv() DutState_CR50KeyEnv {
	if x != nil {
		return x.Cr50KeyEnv
	}
	return DutState_CR50_KEYENV_INVALID
}

func (x *DutState) GetStorageState() HardwareState {
	if x != nil {
		return x.StorageState
	}
	return HardwareState_HARDWARE_UNKNOWN
}

func (x *DutState) GetServoUsbState() HardwareState {
	if x != nil {
		return x.ServoUsbState
	}
	return HardwareState_HARDWARE_UNKNOWN
}

var File_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto protoreflect.FileDescriptor

var file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDesc = []byte{
	0x0a, 0x3d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x2f,
	0x64, 0x75, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x27, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f,
	0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62, 0x1a, 0x46, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f,
	0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73,
	0x5f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xc0, 0x07, 0x0a, 0x08, 0x44, 0x75, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x49, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x75, 0x6e, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e,
	0x6c, 0x61, 0x62, 0x2e, 0x43, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x4f, 0x53, 0x44, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x49, 0x44, 0x52, 0x02, 0x69, 0x64, 0x12, 0x4e, 0x0a, 0x05, 0x73, 0x65, 0x72, 0x76,
	0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x38, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61,
	0x62, 0x2e, 0x50, 0x65, 0x72, 0x69, 0x70, 0x68, 0x65, 0x72, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x52, 0x05, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x12, 0x56, 0x0a, 0x09, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x38, 0x2e, 0x75, 0x6e,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f,
	0x73, 0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x50, 0x65, 0x72, 0x69, 0x70, 0x68, 0x65, 0x72, 0x61, 0x6c,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x09, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x12, 0x6c, 0x0a, 0x15, 0x61, 0x75, 0x64, 0x69, 0x6f, 0x5f, 0x6c, 0x6f, 0x6f, 0x70, 0x62, 0x61,
	0x63, 0x6b, 0x5f, 0x64, 0x6f, 0x6e, 0x67, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x38, 0x2e, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72,
	0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x50, 0x65, 0x72, 0x69, 0x70, 0x68,
	0x65, 0x72, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x13, 0x61, 0x75, 0x64, 0x69, 0x6f,
	0x4c, 0x6f, 0x6f, 0x70, 0x62, 0x61, 0x63, 0x6b, 0x44, 0x6f, 0x6e, 0x67, 0x6c, 0x65, 0x12, 0x38,
	0x0a, 0x18, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x62, 0x6c, 0x75, 0x65, 0x74, 0x6f,
	0x6f, 0x74, 0x68, 0x5f, 0x62, 0x74, 0x70, 0x65, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x16, 0x77, 0x6f, 0x72, 0x6b, 0x69, 0x6e, 0x67, 0x42, 0x6c, 0x75, 0x65, 0x74, 0x6f, 0x6f,
	0x74, 0x68, 0x42, 0x74, 0x70, 0x65, 0x65, 0x72, 0x12, 0x5a, 0x0a, 0x0a, 0x63, 0x72, 0x35, 0x30,
	0x5f, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x3b, 0x2e, 0x75,
	0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65,
	0x6f, 0x73, 0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x44, 0x75, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x2e,
	0x43, 0x52, 0x35, 0x30, 0x50, 0x68, 0x61, 0x73, 0x65, 0x52, 0x09, 0x63, 0x72, 0x35, 0x30, 0x50,
	0x68, 0x61, 0x73, 0x65, 0x12, 0x5e, 0x0a, 0x0c, 0x63, 0x72, 0x35, 0x30, 0x5f, 0x6b, 0x65, 0x79,
	0x5f, 0x65, 0x6e, 0x76, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x3c, 0x2e, 0x75, 0x6e, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73,
	0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x44, 0x75, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x43, 0x52,
	0x35, 0x30, 0x4b, 0x65, 0x79, 0x45, 0x6e, 0x76, 0x52, 0x0a, 0x63, 0x72, 0x35, 0x30, 0x4b, 0x65,
	0x79, 0x45, 0x6e, 0x76, 0x12, 0x5b, 0x0a, 0x0d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x36, 0x2e, 0x75, 0x6e,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f,
	0x73, 0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x48, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x52, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x5e, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f, 0x75, 0x73, 0x62, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x36, 0x2e, 0x75, 0x6e, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73,
	0x2e, 0x6c, 0x61, 0x62, 0x2e, 0x48, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x55, 0x73, 0x62, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x22, 0x4e, 0x0a, 0x09, 0x43, 0x52, 0x35, 0x30, 0x50, 0x68, 0x61, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x12, 0x43, 0x52, 0x35, 0x30, 0x5f, 0x50, 0x48, 0x41, 0x53, 0x45, 0x5f, 0x49, 0x4e, 0x56,
	0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x43, 0x52, 0x35, 0x30, 0x5f, 0x50,
	0x48, 0x41, 0x53, 0x45, 0x5f, 0x50, 0x52, 0x45, 0x50, 0x56, 0x54, 0x10, 0x01, 0x12, 0x12, 0x0a,
	0x0e, 0x43, 0x52, 0x35, 0x30, 0x5f, 0x50, 0x48, 0x41, 0x53, 0x45, 0x5f, 0x50, 0x56, 0x54, 0x10,
	0x02, 0x22, 0x50, 0x0a, 0x0a, 0x43, 0x52, 0x35, 0x30, 0x4b, 0x65, 0x79, 0x45, 0x6e, 0x76, 0x12,
	0x17, 0x0a, 0x13, 0x43, 0x52, 0x35, 0x30, 0x5f, 0x4b, 0x45, 0x59, 0x45, 0x4e, 0x56, 0x5f, 0x49,
	0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x52, 0x35, 0x30,
	0x5f, 0x4b, 0x45, 0x59, 0x45, 0x4e, 0x56, 0x5f, 0x50, 0x52, 0x4f, 0x44, 0x10, 0x01, 0x12, 0x13,
	0x0a, 0x0f, 0x43, 0x52, 0x35, 0x30, 0x5f, 0x4b, 0x45, 0x59, 0x45, 0x4e, 0x56, 0x5f, 0x44, 0x45,
	0x56, 0x10, 0x02, 0x2a, 0xba, 0x02, 0x0a, 0x0f, 0x50, 0x65, 0x72, 0x69, 0x70, 0x68, 0x65, 0x72,
	0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f,
	0x57, 0x4e, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x57, 0x4f, 0x52, 0x4b, 0x49, 0x4e, 0x47, 0x10,
	0x01, 0x12, 0x12, 0x0a, 0x0e, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4e, 0x47, 0x5f, 0x43, 0x4f, 0x4e,
	0x46, 0x49, 0x47, 0x10, 0x05, 0x12, 0x10, 0x0a, 0x0c, 0x57, 0x52, 0x4f, 0x4e, 0x47, 0x5f, 0x43,
	0x4f, 0x4e, 0x46, 0x49, 0x47, 0x10, 0x04, 0x12, 0x11, 0x0a, 0x0d, 0x4e, 0x4f, 0x54, 0x5f, 0x43,
	0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x4f,
	0x5f, 0x53, 0x53, 0x48, 0x10, 0x06, 0x12, 0x0a, 0x0a, 0x06, 0x42, 0x52, 0x4f, 0x4b, 0x45, 0x4e,
	0x10, 0x03, 0x12, 0x14, 0x0a, 0x10, 0x4e, 0x45, 0x45, 0x44, 0x5f, 0x52, 0x45, 0x50, 0x4c, 0x41,
	0x43, 0x45, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x07, 0x12, 0x18, 0x0a, 0x14, 0x43, 0x52, 0x35, 0x30,
	0x5f, 0x43, 0x4f, 0x4e, 0x53, 0x4f, 0x4c, 0x45, 0x5f, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4e, 0x47,
	0x10, 0x0d, 0x12, 0x15, 0x0a, 0x11, 0x43, 0x43, 0x44, 0x5f, 0x54, 0x45, 0x53, 0x54, 0x4c, 0x41,
	0x42, 0x5f, 0x49, 0x53, 0x53, 0x55, 0x45, 0x10, 0x08, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x45, 0x52,
	0x56, 0x4f, 0x44, 0x5f, 0x49, 0x53, 0x53, 0x55, 0x45, 0x10, 0x09, 0x12, 0x13, 0x0a, 0x0f, 0x4c,
	0x49, 0x44, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x0a,
	0x12, 0x14, 0x0a, 0x10, 0x42, 0x41, 0x44, 0x5f, 0x52, 0x49, 0x42, 0x42, 0x4f, 0x4e, 0x5f, 0x43,
	0x41, 0x42, 0x4c, 0x45, 0x10, 0x0b, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x43, 0x5f, 0x42, 0x52, 0x4f,
	0x4b, 0x45, 0x4e, 0x10, 0x0c, 0x12, 0x15, 0x0a, 0x11, 0x44, 0x55, 0x54, 0x5f, 0x4e, 0x4f, 0x54,
	0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x0e, 0x12, 0x12, 0x0a, 0x0e,
	0x54, 0x4f, 0x50, 0x4f, 0x4c, 0x4f, 0x47, 0x59, 0x5f, 0x49, 0x53, 0x53, 0x55, 0x45, 0x10, 0x0f,
	0x2a, 0x8d, 0x01, 0x0a, 0x0d, 0x48, 0x61, 0x72, 0x64, 0x77, 0x61, 0x72, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x48, 0x41, 0x52, 0x44, 0x57, 0x41, 0x52, 0x45, 0x5f, 0x55,
	0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f, 0x48, 0x41, 0x52, 0x44,
	0x57, 0x41, 0x52, 0x45, 0x5f, 0x4e, 0x4f, 0x52, 0x4d, 0x41, 0x4c, 0x10, 0x01, 0x12, 0x17, 0x0a,
	0x13, 0x48, 0x41, 0x52, 0x44, 0x57, 0x41, 0x52, 0x45, 0x5f, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54,
	0x41, 0x42, 0x4c, 0x45, 0x10, 0x02, 0x12, 0x1d, 0x0a, 0x19, 0x48, 0x41, 0x52, 0x44, 0x57, 0x41,
	0x52, 0x45, 0x5f, 0x4e, 0x45, 0x45, 0x44, 0x5f, 0x52, 0x45, 0x50, 0x4c, 0x41, 0x43, 0x45, 0x4d,
	0x45, 0x4e, 0x54, 0x10, 0x03, 0x12, 0x19, 0x0a, 0x15, 0x48, 0x41, 0x52, 0x44, 0x57, 0x41, 0x52,
	0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x44, 0x45, 0x54, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x04,
	0x42, 0x35, 0x5a, 0x33, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2f, 0x6c, 0x61,
	0x62, 0x3b, 0x75, 0x66, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescOnce sync.Once
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescData = file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDesc
)

func file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescGZIP() []byte {
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescOnce.Do(func() {
		file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescData)
	})
	return file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDescData
}

var file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_goTypes = []interface{}{
	(PeripheralState)(0),     // 0: unifiedfleet.api.v1.models.chromeos.lab.PeripheralState
	(HardwareState)(0),       // 1: unifiedfleet.api.v1.models.chromeos.lab.HardwareState
	(DutState_CR50Phase)(0),  // 2: unifiedfleet.api.v1.models.chromeos.lab.DutState.CR50Phase
	(DutState_CR50KeyEnv)(0), // 3: unifiedfleet.api.v1.models.chromeos.lab.DutState.CR50KeyEnv
	(*DutState)(nil),         // 4: unifiedfleet.api.v1.models.chromeos.lab.DutState
	(*ChromeOSDeviceID)(nil), // 5: unifiedfleet.api.v1.models.chromeos.lab.ChromeOSDeviceID
}
var file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_depIdxs = []int32{
	5, // 0: unifiedfleet.api.v1.models.chromeos.lab.DutState.id:type_name -> unifiedfleet.api.v1.models.chromeos.lab.ChromeOSDeviceID
	0, // 1: unifiedfleet.api.v1.models.chromeos.lab.DutState.servo:type_name -> unifiedfleet.api.v1.models.chromeos.lab.PeripheralState
	0, // 2: unifiedfleet.api.v1.models.chromeos.lab.DutState.chameleon:type_name -> unifiedfleet.api.v1.models.chromeos.lab.PeripheralState
	0, // 3: unifiedfleet.api.v1.models.chromeos.lab.DutState.audio_loopback_dongle:type_name -> unifiedfleet.api.v1.models.chromeos.lab.PeripheralState
	2, // 4: unifiedfleet.api.v1.models.chromeos.lab.DutState.cr50_phase:type_name -> unifiedfleet.api.v1.models.chromeos.lab.DutState.CR50Phase
	3, // 5: unifiedfleet.api.v1.models.chromeos.lab.DutState.cr50_key_env:type_name -> unifiedfleet.api.v1.models.chromeos.lab.DutState.CR50KeyEnv
	1, // 6: unifiedfleet.api.v1.models.chromeos.lab.DutState.storage_state:type_name -> unifiedfleet.api.v1.models.chromeos.lab.HardwareState
	1, // 7: unifiedfleet.api.v1.models.chromeos.lab.DutState.servo_usb_state:type_name -> unifiedfleet.api.v1.models.chromeos.lab.HardwareState
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_init() }
func file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_init() {
	if File_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto != nil {
		return
	}
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_chromeos_device_id_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DutState); i {
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
			RawDescriptor: file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_goTypes,
		DependencyIndexes: file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_depIdxs,
		EnumInfos:         file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_enumTypes,
		MessageInfos:      file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_msgTypes,
	}.Build()
	File_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto = out.File
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_rawDesc = nil
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_goTypes = nil
	file_infra_unifiedfleet_api_v1_models_chromeos_lab_dut_state_proto_depIdxs = nil
}
