// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: infra/appengine/cros/lab_inventory/api/v1/repair_record.proto

package api

import (
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// The triggering device that led you to work on this repair.
//
// ie. If the DUT repairs dashboard led you to work on this device, then it is
// a DUT repair. If the Servo or Labstation dashboard lead to you work on this
// device, then it is a Servo or Labstation repair.
// Next tag: 3
type DeviceManualRepairRecord_RepairTargetType int32

const (
	DeviceManualRepairRecord_TYPE_DUT        DeviceManualRepairRecord_RepairTargetType = 0
	DeviceManualRepairRecord_TYPE_LABSTATION DeviceManualRepairRecord_RepairTargetType = 1
	DeviceManualRepairRecord_TYPE_SERVO      DeviceManualRepairRecord_RepairTargetType = 2
)

// Enum value maps for DeviceManualRepairRecord_RepairTargetType.
var (
	DeviceManualRepairRecord_RepairTargetType_name = map[int32]string{
		0: "TYPE_DUT",
		1: "TYPE_LABSTATION",
		2: "TYPE_SERVO",
	}
	DeviceManualRepairRecord_RepairTargetType_value = map[string]int32{
		"TYPE_DUT":        0,
		"TYPE_LABSTATION": 1,
		"TYPE_SERVO":      2,
	}
)

func (x DeviceManualRepairRecord_RepairTargetType) Enum() *DeviceManualRepairRecord_RepairTargetType {
	p := new(DeviceManualRepairRecord_RepairTargetType)
	*p = x
	return p
}

func (x DeviceManualRepairRecord_RepairTargetType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeviceManualRepairRecord_RepairTargetType) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[0].Descriptor()
}

func (DeviceManualRepairRecord_RepairTargetType) Type() protoreflect.EnumType {
	return &file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[0]
}

func (x DeviceManualRepairRecord_RepairTargetType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeviceManualRepairRecord_RepairTargetType.Descriptor instead.
func (DeviceManualRepairRecord_RepairTargetType) EnumDescriptor() ([]byte, []int) {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescGZIP(), []int{0, 0}
}

// State for tracking manual repair progress.
// Next tag: 4
type DeviceManualRepairRecord_RepairState int32

const (
	DeviceManualRepairRecord_STATE_INVALID     DeviceManualRepairRecord_RepairState = 0
	DeviceManualRepairRecord_STATE_NOT_STARTED DeviceManualRepairRecord_RepairState = 1
	DeviceManualRepairRecord_STATE_IN_PROGRESS DeviceManualRepairRecord_RepairState = 2
	DeviceManualRepairRecord_STATE_COMPLETED   DeviceManualRepairRecord_RepairState = 3
)

// Enum value maps for DeviceManualRepairRecord_RepairState.
var (
	DeviceManualRepairRecord_RepairState_name = map[int32]string{
		0: "STATE_INVALID",
		1: "STATE_NOT_STARTED",
		2: "STATE_IN_PROGRESS",
		3: "STATE_COMPLETED",
	}
	DeviceManualRepairRecord_RepairState_value = map[string]int32{
		"STATE_INVALID":     0,
		"STATE_NOT_STARTED": 1,
		"STATE_IN_PROGRESS": 2,
		"STATE_COMPLETED":   3,
	}
)

func (x DeviceManualRepairRecord_RepairState) Enum() *DeviceManualRepairRecord_RepairState {
	p := new(DeviceManualRepairRecord_RepairState)
	*p = x
	return p
}

func (x DeviceManualRepairRecord_RepairState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeviceManualRepairRecord_RepairState) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[1].Descriptor()
}

func (DeviceManualRepairRecord_RepairState) Type() protoreflect.EnumType {
	return &file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[1]
}

func (x DeviceManualRepairRecord_RepairState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeviceManualRepairRecord_RepairState.Descriptor instead.
func (DeviceManualRepairRecord_RepairState) EnumDescriptor() ([]byte, []int) {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescGZIP(), []int{0, 1}
}

// Standard manual repair actions taken to fix the device.
// Next tag: 7
type DeviceManualRepairRecord_ManualRepairAction int32

const (
	// Fix Labstation
	DeviceManualRepairRecord_ACTION_FIX_LABSTATION DeviceManualRepairRecord_ManualRepairAction = 0
	// Fix Servo
	DeviceManualRepairRecord_ACTION_FIX_SERVO DeviceManualRepairRecord_ManualRepairAction = 1
	// Fix Yoshi cable / servo_micro
	DeviceManualRepairRecord_ACTION_FIX_YOSHI_CABLE DeviceManualRepairRecord_ManualRepairAction = 2
	// Visual Inspection
	DeviceManualRepairRecord_ACTION_VISUAL_INSPECTION DeviceManualRepairRecord_ManualRepairAction = 3
	// Check / Fix Power for DUT
	DeviceManualRepairRecord_ACTION_DUT_POWER DeviceManualRepairRecord_ManualRepairAction = 4
	// Troubleshoot DUT
	DeviceManualRepairRecord_ACTION_TROUBLESHOOT_DUT DeviceManualRepairRecord_ManualRepairAction = 5
	// Reimage / Reflash DUT
	DeviceManualRepairRecord_ACTION_REIMAGE_DUT DeviceManualRepairRecord_ManualRepairAction = 6
)

// Enum value maps for DeviceManualRepairRecord_ManualRepairAction.
var (
	DeviceManualRepairRecord_ManualRepairAction_name = map[int32]string{
		0: "ACTION_FIX_LABSTATION",
		1: "ACTION_FIX_SERVO",
		2: "ACTION_FIX_YOSHI_CABLE",
		3: "ACTION_VISUAL_INSPECTION",
		4: "ACTION_DUT_POWER",
		5: "ACTION_TROUBLESHOOT_DUT",
		6: "ACTION_REIMAGE_DUT",
	}
	DeviceManualRepairRecord_ManualRepairAction_value = map[string]int32{
		"ACTION_FIX_LABSTATION":    0,
		"ACTION_FIX_SERVO":         1,
		"ACTION_FIX_YOSHI_CABLE":   2,
		"ACTION_VISUAL_INSPECTION": 3,
		"ACTION_DUT_POWER":         4,
		"ACTION_TROUBLESHOOT_DUT":  5,
		"ACTION_REIMAGE_DUT":       6,
	}
)

func (x DeviceManualRepairRecord_ManualRepairAction) Enum() *DeviceManualRepairRecord_ManualRepairAction {
	p := new(DeviceManualRepairRecord_ManualRepairAction)
	*p = x
	return p
}

func (x DeviceManualRepairRecord_ManualRepairAction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeviceManualRepairRecord_ManualRepairAction) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[2].Descriptor()
}

func (DeviceManualRepairRecord_ManualRepairAction) Type() protoreflect.EnumType {
	return &file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes[2]
}

func (x DeviceManualRepairRecord_ManualRepairAction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeviceManualRepairRecord_ManualRepairAction.Descriptor instead.
func (DeviceManualRepairRecord_ManualRepairAction) EnumDescriptor() ([]byte, []int) {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescGZIP(), []int{0, 2}
}

// Next tag: 20
type DeviceManualRepairRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hostname         string                                    `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	AssetTag         string                                    `protobuf:"bytes,2,opt,name=asset_tag,json=assetTag,proto3" json:"asset_tag,omitempty"`
	RepairTargetType DeviceManualRepairRecord_RepairTargetType `protobuf:"varint,3,opt,name=repair_target_type,json=repairTargetType,proto3,enum=inventory.DeviceManualRepairRecord_RepairTargetType" json:"repair_target_type,omitempty"`
	RepairState      DeviceManualRepairRecord_RepairState      `protobuf:"varint,4,opt,name=repair_state,json=repairState,proto3,enum=inventory.DeviceManualRepairRecord_RepairState" json:"repair_state,omitempty"`
	// Buganizer bug tracking https://b/XXXXXXXXX.
	BuganizerBugUrl string `protobuf:"bytes,5,opt,name=buganizer_bug_url,json=buganizerBugUrl,proto3" json:"buganizer_bug_url,omitempty"`
	// Chromium bug tracking https://crbug.com/XXXXXXX.
	ChromiumBugUrl string `protobuf:"bytes,6,opt,name=chromium_bug_url,json=chromiumBugUrl,proto3" json:"chromium_bug_url,omitempty"`
	// DUT repair failure description.
	DutRepairFailureDescription string `protobuf:"bytes,7,opt,name=dut_repair_failure_description,json=dutRepairFailureDescription,proto3" json:"dut_repair_failure_description,omitempty"`
	// The last DUT repair verifier that failed.
	DutVerifierFailureDescription string `protobuf:"bytes,8,opt,name=dut_verifier_failure_description,json=dutVerifierFailureDescription,proto3" json:"dut_verifier_failure_description,omitempty"`
	// Servo repair failure description.
	ServoRepairFailureDescription string `protobuf:"bytes,9,opt,name=servo_repair_failure_description,json=servoRepairFailureDescription,proto3" json:"servo_repair_failure_description,omitempty"`
	// The last Servo repair verifier that failed.
	ServoVerifierFailureDescription string `protobuf:"bytes,10,opt,name=servo_verifier_failure_description,json=servoVerifierFailureDescription,proto3" json:"servo_verifier_failure_description,omitempty"`
	// Diagnosis of what is wrong with the device.
	Diagnosis string `protobuf:"bytes,11,opt,name=diagnosis,proto3" json:"diagnosis,omitempty"`
	// The procedure that fixed the device. This can be a best guess. Assumption
	// is that admin/skylab repairs will run to verify the repair post fix.
	RepairProcedure     string                                        `protobuf:"bytes,12,opt,name=repair_procedure,json=repairProcedure,proto3" json:"repair_procedure,omitempty"`
	ManualRepairActions []DeviceManualRepairRecord_ManualRepairAction `protobuf:"varint,13,rep,packed,name=manual_repair_actions,json=manualRepairActions,proto3,enum=inventory.DeviceManualRepairRecord_ManualRepairAction" json:"manual_repair_actions,omitempty"`
	// Boolean value of whether the primary issue has been fixed or not.
	IssueFixed bool `protobuf:"varint,18,opt,name=issue_fixed,json=issueFixed,proto3" json:"issue_fixed,omitempty"`
	// User ldap of who started the device repair
	UserLdap string `protobuf:"bytes,19,opt,name=user_ldap,json=userLdap,proto3" json:"user_ldap,omitempty"`
	// Input by Lab Tech as a best guess on how much time (in minutes) it took to
	// investigate and resolve the repair. This is to give a better repair time
	// estimate that excludes time spent waiting. Record updations may not happen
	// right after a repair is completed so calculating from other timestamps may
	// not accurately portray the time a Lab Tech spent actually investigating and
	// repairing the device.
	TimeTaken int32 `protobuf:"varint,14,opt,name=time_taken,json=timeTaken,proto3" json:"time_taken,omitempty"`
	// Timestamp when repair record was created.
	CreatedTime *timestamp.Timestamp `protobuf:"bytes,15,opt,name=created_time,json=createdTime,proto3" json:"created_time,omitempty"`
	// Timestamp when repair record was last updated.
	UpdatedTime *timestamp.Timestamp `protobuf:"bytes,16,opt,name=updated_time,json=updatedTime,proto3" json:"updated_time,omitempty"`
	// Timestamp when repair record was marked completed.
	CompletedTime *timestamp.Timestamp `protobuf:"bytes,17,opt,name=completed_time,json=completedTime,proto3" json:"completed_time,omitempty"`
}

func (x *DeviceManualRepairRecord) Reset() {
	*x = DeviceManualRepairRecord{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceManualRepairRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceManualRepairRecord) ProtoMessage() {}

func (x *DeviceManualRepairRecord) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceManualRepairRecord.ProtoReflect.Descriptor instead.
func (*DeviceManualRepairRecord) Descriptor() ([]byte, []int) {
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescGZIP(), []int{0}
}

func (x *DeviceManualRepairRecord) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetAssetTag() string {
	if x != nil {
		return x.AssetTag
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetRepairTargetType() DeviceManualRepairRecord_RepairTargetType {
	if x != nil {
		return x.RepairTargetType
	}
	return DeviceManualRepairRecord_TYPE_DUT
}

func (x *DeviceManualRepairRecord) GetRepairState() DeviceManualRepairRecord_RepairState {
	if x != nil {
		return x.RepairState
	}
	return DeviceManualRepairRecord_STATE_INVALID
}

func (x *DeviceManualRepairRecord) GetBuganizerBugUrl() string {
	if x != nil {
		return x.BuganizerBugUrl
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetChromiumBugUrl() string {
	if x != nil {
		return x.ChromiumBugUrl
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetDutRepairFailureDescription() string {
	if x != nil {
		return x.DutRepairFailureDescription
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetDutVerifierFailureDescription() string {
	if x != nil {
		return x.DutVerifierFailureDescription
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetServoRepairFailureDescription() string {
	if x != nil {
		return x.ServoRepairFailureDescription
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetServoVerifierFailureDescription() string {
	if x != nil {
		return x.ServoVerifierFailureDescription
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetDiagnosis() string {
	if x != nil {
		return x.Diagnosis
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetRepairProcedure() string {
	if x != nil {
		return x.RepairProcedure
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetManualRepairActions() []DeviceManualRepairRecord_ManualRepairAction {
	if x != nil {
		return x.ManualRepairActions
	}
	return nil
}

func (x *DeviceManualRepairRecord) GetIssueFixed() bool {
	if x != nil {
		return x.IssueFixed
	}
	return false
}

func (x *DeviceManualRepairRecord) GetUserLdap() string {
	if x != nil {
		return x.UserLdap
	}
	return ""
}

func (x *DeviceManualRepairRecord) GetTimeTaken() int32 {
	if x != nil {
		return x.TimeTaken
	}
	return 0
}

func (x *DeviceManualRepairRecord) GetCreatedTime() *timestamp.Timestamp {
	if x != nil {
		return x.CreatedTime
	}
	return nil
}

func (x *DeviceManualRepairRecord) GetUpdatedTime() *timestamp.Timestamp {
	if x != nil {
		return x.UpdatedTime
	}
	return nil
}

func (x *DeviceManualRepairRecord) GetCompletedTime() *timestamp.Timestamp {
	if x != nil {
		return x.CompletedTime
	}
	return nil
}

var File_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto protoreflect.FileDescriptor

var file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDesc = []byte{
	0x0a, 0x3d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x5f, 0x69, 0x6e, 0x76, 0x65, 0x6e,
	0x74, 0x6f, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x70, 0x61,
	0x69, 0x72, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd1, 0x0b, 0x0a, 0x18,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61,
	0x69, 0x72, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x74, 0x61,
	0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x73, 0x73, 0x65, 0x74, 0x54, 0x61,
	0x67, 0x12, 0x62, 0x0a, 0x12, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x34, 0x2e,
	0x69, 0x6e, 0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x4d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x52, 0x65, 0x63, 0x6f,
	0x72, 0x64, 0x2e, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x10, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x54, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x52, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2f, 0x2e, 0x69, 0x6e,
	0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x61,
	0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64,
	0x2e, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x0b, 0x72, 0x65,
	0x70, 0x61, 0x69, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x2a, 0x0a, 0x11, 0x62, 0x75, 0x67,
	0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x5f, 0x62, 0x75, 0x67, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x62, 0x75, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x42,
	0x75, 0x67, 0x55, 0x72, 0x6c, 0x12, 0x28, 0x0a, 0x10, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75,
	0x6d, 0x5f, 0x62, 0x75, 0x67, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x42, 0x75, 0x67, 0x55, 0x72, 0x6c, 0x12,
	0x43, 0x0a, 0x1e, 0x64, 0x75, 0x74, 0x5f, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x66, 0x61,
	0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1b, 0x64, 0x75, 0x74, 0x52, 0x65, 0x70, 0x61,
	0x69, 0x72, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x20, 0x64, 0x75, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x69,
	0x66, 0x69, 0x65, 0x72, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1d,
	0x64, 0x75, 0x74, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x72, 0x46, 0x61, 0x69, 0x6c, 0x75,
	0x72, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x47, 0x0a,
	0x20, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x66, 0x61,
	0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x1d, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x52, 0x65,
	0x70, 0x61, 0x69, 0x72, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4b, 0x0a, 0x22, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x5f,
	0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x72, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65,
	0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x1f, 0x73, 0x65, 0x72, 0x76, 0x6f, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65,
	0x72, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x69, 0x73,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x69, 0x61, 0x67, 0x6e, 0x6f, 0x73, 0x69,
	0x73, 0x12, 0x29, 0x0a, 0x10, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x70, 0x72, 0x6f, 0x63,
	0x65, 0x64, 0x75, 0x72, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x65, 0x70,
	0x61, 0x69, 0x72, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x64, 0x75, 0x72, 0x65, 0x12, 0x6a, 0x0a, 0x15,
	0x6d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x5f, 0x72, 0x65, 0x70, 0x61, 0x69, 0x72, 0x5f, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x36, 0x2e, 0x69, 0x6e,
	0x76, 0x65, 0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x61,
	0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64,
	0x2e, 0x4d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x41, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x13, 0x6d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70, 0x61, 0x69,
	0x72, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x73, 0x73, 0x75,
	0x65, 0x5f, 0x66, 0x69, 0x78, 0x65, 0x64, 0x18, 0x12, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x69,
	0x73, 0x73, 0x75, 0x65, 0x46, 0x69, 0x78, 0x65, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x75, 0x73, 0x65,
	0x72, 0x5f, 0x6c, 0x64, 0x61, 0x70, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73,
	0x65, 0x72, 0x4c, 0x64, 0x61, 0x70, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x74,
	0x61, 0x6b, 0x65, 0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x54, 0x61, 0x6b, 0x65, 0x6e, 0x12, 0x3d, 0x0a, 0x0c, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x41, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x64,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0d, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74,
	0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x45, 0x0a, 0x10, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72,
	0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x44, 0x55, 0x54, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x4c, 0x41, 0x42, 0x53, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x01, 0x12, 0x0e, 0x0a,
	0x0a, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x4f, 0x10, 0x02, 0x22, 0x63, 0x0a,
	0x0b, 0x52, 0x65, 0x70, 0x61, 0x69, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x11, 0x0a, 0x0d,
	0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12,
	0x15, 0x0a, 0x11, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x53, 0x54, 0x41,
	0x52, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f,
	0x49, 0x4e, 0x5f, 0x50, 0x52, 0x4f, 0x47, 0x52, 0x45, 0x53, 0x53, 0x10, 0x02, 0x12, 0x13, 0x0a,
	0x0f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45, 0x44,
	0x10, 0x03, 0x22, 0xca, 0x01, 0x0a, 0x12, 0x4d, 0x61, 0x6e, 0x75, 0x61, 0x6c, 0x52, 0x65, 0x70,
	0x61, 0x69, 0x72, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x15, 0x41, 0x43, 0x54,
	0x49, 0x4f, 0x4e, 0x5f, 0x46, 0x49, 0x58, 0x5f, 0x4c, 0x41, 0x42, 0x53, 0x54, 0x41, 0x54, 0x49,
	0x4f, 0x4e, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x46,
	0x49, 0x58, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x4f, 0x10, 0x01, 0x12, 0x1a, 0x0a, 0x16, 0x41, 0x43,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x46, 0x49, 0x58, 0x5f, 0x59, 0x4f, 0x53, 0x48, 0x49, 0x5f, 0x43,
	0x41, 0x42, 0x4c, 0x45, 0x10, 0x02, 0x12, 0x1c, 0x0a, 0x18, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e,
	0x5f, 0x56, 0x49, 0x53, 0x55, 0x41, 0x4c, 0x5f, 0x49, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x54, 0x49,
	0x4f, 0x4e, 0x10, 0x03, 0x12, 0x14, 0x0a, 0x10, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x44,
	0x55, 0x54, 0x5f, 0x50, 0x4f, 0x57, 0x45, 0x52, 0x10, 0x04, 0x12, 0x1b, 0x0a, 0x17, 0x41, 0x43,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x54, 0x52, 0x4f, 0x55, 0x42, 0x4c, 0x45, 0x53, 0x48, 0x4f, 0x4f,
	0x54, 0x5f, 0x44, 0x55, 0x54, 0x10, 0x05, 0x12, 0x16, 0x0a, 0x12, 0x41, 0x43, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x52, 0x45, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f, 0x44, 0x55, 0x54, 0x10, 0x06, 0x42,
	0x2f, 0x5a, 0x2d, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6c, 0x61, 0x62, 0x5f, 0x69, 0x6e, 0x76, 0x65,
	0x6e, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x70, 0x69,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescOnce sync.Once
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescData = file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDesc
)

func file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescGZIP() []byte {
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescOnce.Do(func() {
		file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescData)
	})
	return file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDescData
}

var file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_goTypes = []interface{}{
	(DeviceManualRepairRecord_RepairTargetType)(0),   // 0: inventory.DeviceManualRepairRecord.RepairTargetType
	(DeviceManualRepairRecord_RepairState)(0),        // 1: inventory.DeviceManualRepairRecord.RepairState
	(DeviceManualRepairRecord_ManualRepairAction)(0), // 2: inventory.DeviceManualRepairRecord.ManualRepairAction
	(*DeviceManualRepairRecord)(nil),                 // 3: inventory.DeviceManualRepairRecord
	(*timestamp.Timestamp)(nil),                      // 4: google.protobuf.Timestamp
}
var file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_depIdxs = []int32{
	0, // 0: inventory.DeviceManualRepairRecord.repair_target_type:type_name -> inventory.DeviceManualRepairRecord.RepairTargetType
	1, // 1: inventory.DeviceManualRepairRecord.repair_state:type_name -> inventory.DeviceManualRepairRecord.RepairState
	2, // 2: inventory.DeviceManualRepairRecord.manual_repair_actions:type_name -> inventory.DeviceManualRepairRecord.ManualRepairAction
	4, // 3: inventory.DeviceManualRepairRecord.created_time:type_name -> google.protobuf.Timestamp
	4, // 4: inventory.DeviceManualRepairRecord.updated_time:type_name -> google.protobuf.Timestamp
	4, // 5: inventory.DeviceManualRepairRecord.completed_time:type_name -> google.protobuf.Timestamp
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_init() }
func file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_init() {
	if File_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceManualRepairRecord); i {
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
			RawDescriptor: file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_goTypes,
		DependencyIndexes: file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_depIdxs,
		EnumInfos:         file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_enumTypes,
		MessageInfos:      file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_msgTypes,
	}.Build()
	File_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto = out.File
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_rawDesc = nil
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_goTypes = nil
	file_infra_appengine_cros_lab_inventory_api_v1_repair_record_proto_depIdxs = nil
}
