// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: infra/cros/karte/api/bigquery/action.proto

package kbqpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Action is the type a chromeos.karte.action that has been exported to BigQuery.
type Action struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The resource name of the action. Names are generated
	// automatically when a new action is created.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// A kind is a coarse-grained type of an action, such as
	// ssh-attempt. New action_kinds will be created frequently so this field
	// is a string; see https://google.aip.dev/126 for details.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// A swarming task ID is the ID of a single swarming task.
	// The swarming task of an action is the swarming task that invoked the
	// action.
	// For example, "4f6c0ba2ef3fc610" is a swarming task ID.
	SwarmingTaskId string `protobuf:"bytes,3,opt,name=swarming_task_id,json=swarmingTaskId,proto3" json:"swarming_task_id,omitempty"`
	// An asset tag is the tag of a given asset in UFS.
	// An asset tag may be a short number such as C444444 printed on a device,
	// or it may be a UUID in some circumstances.
	AssetTag string `protobuf:"bytes,4,opt,name=asset_tag,json=assetTag,proto3" json:"asset_tag,omitempty"`
	// The start time is the time that an action started.
	StartTime *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	// The stop time is the time that an action finished.
	StopTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=stop_time,json=stopTime,proto3" json:"stop_time,omitempty"`
	// The create time is the time that an action was created by Karte.
	// This is the time that the event was first received, since events are
	// immutable outside of rare cases.
	// This field is managed by Karte itself.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// The status of an action is whether the action succeeded or failed.
	Status string `protobuf:"bytes,8,opt,name=status,proto3" json:"status,omitempty"`
	// The fail reason of an event is a diagnostic message that is emitted when
	// the action in question has failed.
	FailReason string `protobuf:"bytes,9,opt,name=fail_reason,json=failReason,proto3" json:"fail_reason,omitempty"`
	// The seal time is when the particular Karte record is sealed and no further changes can be made.
	SealTime *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=seal_time,json=sealTime,proto3" json:"seal_time,omitempty"`
	// This is the last time that the particular Karte record was updated on the Karte side.
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,11,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	// The client name is the name of the entity creating the Action entry, e.g. "paris".
	ClientName string `protobuf:"bytes,12,opt,name=client_name,json=clientName,proto3" json:"client_name,omitempty"`
	// The client version is the version of the entity creating the Action entry, e.g. "0.0.1".
	ClientVersion string `protobuf:"bytes,13,opt,name=client_version,json=clientVersion,proto3" json:"client_version,omitempty"`
	// The buildbucket ID is the ID of the buildbucket build associated with the event in question.
	BuildbucketId string `protobuf:"bytes,14,opt,name=buildbucket_id,json=buildbucketId,proto3" json:"buildbucket_id,omitempty"`
	// The hostname is the hostname of the DUT in question.
	Hostname string `protobuf:"bytes,15,opt,name=hostname,proto3" json:"hostname,omitempty"`
	// model is the model of the DUT this event applies to.
	Model string `protobuf:"bytes,16,opt,name=model,proto3" json:"model,omitempty"`
	// board is the board of the DUT this event applies to.
	Board string `protobuf:"bytes,17,opt,name=board,proto3" json:"board,omitempty"`
	// The modification count is the number of times that the record has been updated.
	// A complete action should be written to twice: once when it was created and once when it completed.
	ModificationCount int32 `protobuf:"varint,18,opt,name=modification_count,json=modificationCount,proto3" json:"modification_count,omitempty"`
	// recovered by is the name of the task that recovered the current action, if one exists.
	RecoveredBy string `protobuf:"bytes,19,opt,name=recovered_by,json=recoveredBy,proto3" json:"recovered_by,omitempty"`
	// restarts is the number of times that the current plan was restarted.
	Restarts int32 `protobuf:"varint,20,opt,name=restarts,proto3" json:"restarts,omitempty"`
	// plan_name is the name of the plan that we're currently executing.
	PlanName string `protobuf:"bytes,21,opt,name=plan_name,json=planName,proto3" json:"plan_name,omitempty"`
	// allow_fail captures whether the action is allowed to fail or not.
	AllowFail string `protobuf:"bytes,22,opt,name=allow_fail,json=allowFail,proto3" json:"allow_fail,omitempty"`
	// Action type is the type of action "ACTION_TYPE_" + {"UNSPECIFIED", "VERIFIER", "CONDITION", "RECOVERY"}.
	ActionType string `protobuf:"bytes,23,opt,name=action_type,json=actionType,proto3" json:"action_type,omitempty"`
}

func (x *Action) Reset() {
	*x = Action{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_bigquery_action_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Action) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Action) ProtoMessage() {}

func (x *Action) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_bigquery_action_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Action.ProtoReflect.Descriptor instead.
func (*Action) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_bigquery_action_proto_rawDescGZIP(), []int{0}
}

func (x *Action) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Action) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Action) GetSwarmingTaskId() string {
	if x != nil {
		return x.SwarmingTaskId
	}
	return ""
}

func (x *Action) GetAssetTag() string {
	if x != nil {
		return x.AssetTag
	}
	return ""
}

func (x *Action) GetStartTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *Action) GetStopTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StopTime
	}
	return nil
}

func (x *Action) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Action) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Action) GetFailReason() string {
	if x != nil {
		return x.FailReason
	}
	return ""
}

func (x *Action) GetSealTime() *timestamppb.Timestamp {
	if x != nil {
		return x.SealTime
	}
	return nil
}

func (x *Action) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *Action) GetClientName() string {
	if x != nil {
		return x.ClientName
	}
	return ""
}

func (x *Action) GetClientVersion() string {
	if x != nil {
		return x.ClientVersion
	}
	return ""
}

func (x *Action) GetBuildbucketId() string {
	if x != nil {
		return x.BuildbucketId
	}
	return ""
}

func (x *Action) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *Action) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *Action) GetBoard() string {
	if x != nil {
		return x.Board
	}
	return ""
}

func (x *Action) GetModificationCount() int32 {
	if x != nil {
		return x.ModificationCount
	}
	return 0
}

func (x *Action) GetRecoveredBy() string {
	if x != nil {
		return x.RecoveredBy
	}
	return ""
}

func (x *Action) GetRestarts() int32 {
	if x != nil {
		return x.Restarts
	}
	return 0
}

func (x *Action) GetPlanName() string {
	if x != nil {
		return x.PlanName
	}
	return ""
}

func (x *Action) GetAllowFail() string {
	if x != nil {
		return x.AllowFail
	}
	return ""
}

func (x *Action) GetActionType() string {
	if x != nil {
		return x.ActionType
	}
	return ""
}

var File_infra_cros_karte_api_bigquery_action_proto protoreflect.FileDescriptor

var file_infra_cros_karte_api_bigquery_action_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6b, 0x61, 0x72,
	0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x69, 0x67, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x63, 0x68,
	0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2e, 0x62, 0x69, 0x67,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x06, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x28, 0x0a, 0x10, 0x73, 0x77, 0x61,
	0x72, 0x6d, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x77, 0x61, 0x72, 0x6d, 0x69, 0x6e, 0x67, 0x54, 0x61, 0x73,
	0x6b, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x74, 0x61, 0x67,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x73, 0x73, 0x65, 0x74, 0x54, 0x61, 0x67,
	0x12, 0x39, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x37, 0x0a, 0x09, 0x73,
	0x74, 0x6f, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x73, 0x74, 0x6f, 0x70,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x61, 0x69,
	0x6c, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x66, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x37, 0x0a, 0x09, 0x73, 0x65,
	0x61, 0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x73, 0x65, 0x61, 0x6c, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x49, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0f, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d,
	0x6f, 0x64, 0x65, 0x6c, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x2d, 0x0a, 0x12, 0x6d, 0x6f, 0x64, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x12, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x11, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x21, 0x0a, 0x0c, 0x72, 0x65, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x72, 0x65,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x42, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x73, 0x18, 0x14, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x72, 0x65, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x6e, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x15, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x66, 0x61, 0x69, 0x6c,
	0x18, 0x16, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x46, 0x61, 0x69,
	0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x42, 0x25, 0x5a, 0x23, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73,
	0x2f, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x62, 0x69, 0x67, 0x71, 0x75,
	0x65, 0x72, 0x79, 0x3b, 0x6b, 0x62, 0x71, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_infra_cros_karte_api_bigquery_action_proto_rawDescOnce sync.Once
	file_infra_cros_karte_api_bigquery_action_proto_rawDescData = file_infra_cros_karte_api_bigquery_action_proto_rawDesc
)

func file_infra_cros_karte_api_bigquery_action_proto_rawDescGZIP() []byte {
	file_infra_cros_karte_api_bigquery_action_proto_rawDescOnce.Do(func() {
		file_infra_cros_karte_api_bigquery_action_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cros_karte_api_bigquery_action_proto_rawDescData)
	})
	return file_infra_cros_karte_api_bigquery_action_proto_rawDescData
}

var file_infra_cros_karte_api_bigquery_action_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_cros_karte_api_bigquery_action_proto_goTypes = []interface{}{
	(*Action)(nil),                // 0: chromeos.karte.bigquery.Action
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_infra_cros_karte_api_bigquery_action_proto_depIdxs = []int32{
	1, // 0: chromeos.karte.bigquery.Action.start_time:type_name -> google.protobuf.Timestamp
	1, // 1: chromeos.karte.bigquery.Action.stop_time:type_name -> google.protobuf.Timestamp
	1, // 2: chromeos.karte.bigquery.Action.create_time:type_name -> google.protobuf.Timestamp
	1, // 3: chromeos.karte.bigquery.Action.seal_time:type_name -> google.protobuf.Timestamp
	1, // 4: chromeos.karte.bigquery.Action.update_time:type_name -> google.protobuf.Timestamp
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_infra_cros_karte_api_bigquery_action_proto_init() }
func file_infra_cros_karte_api_bigquery_action_proto_init() {
	if File_infra_cros_karte_api_bigquery_action_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cros_karte_api_bigquery_action_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Action); i {
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
			RawDescriptor: file_infra_cros_karte_api_bigquery_action_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_cros_karte_api_bigquery_action_proto_goTypes,
		DependencyIndexes: file_infra_cros_karte_api_bigquery_action_proto_depIdxs,
		MessageInfos:      file_infra_cros_karte_api_bigquery_action_proto_msgTypes,
	}.Build()
	File_infra_cros_karte_api_bigquery_action_proto = out.File
	file_infra_cros_karte_api_bigquery_action_proto_rawDesc = nil
	file_infra_cros_karte_api_bigquery_action_proto_goTypes = nil
	file_infra_cros_karte_api_bigquery_action_proto_depIdxs = nil
}