// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.12.1
// source: infra/cros/karte/api/searchquery.proto

package kartepb

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

// An ActionSearchQuery contains all possible criteria that can
// be used to search for actions.
type ActionSearchQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The action filter is a collection of all search criteria that are not
	// intrinsically ordered.
	CategoricalFilter *ActionCategoricalFilter `protobuf:"bytes,1,opt,name=categorical_filter,json=categoricalFilter,proto3" json:"categorical_filter,omitempty"`
	// The minimum stop time, if provided, is a lower bound on the stop time of
	// actions.
	MinimumStopTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=minimum_stop_time,json=minimumStopTime,proto3" json:"minimum_stop_time,omitempty"`
	// The maximum stop time, if provided, is an upper bound on the stop time of
	// actions.
	MaximumStopTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=maximum_stop_time,json=maximumStopTime,proto3" json:"maximum_stop_time,omitempty"`
}

func (x *ActionSearchQuery) Reset() {
	*x = ActionSearchQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActionSearchQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActionSearchQuery) ProtoMessage() {}

func (x *ActionSearchQuery) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActionSearchQuery.ProtoReflect.Descriptor instead.
func (*ActionSearchQuery) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_searchquery_proto_rawDescGZIP(), []int{0}
}

func (x *ActionSearchQuery) GetCategoricalFilter() *ActionCategoricalFilter {
	if x != nil {
		return x.CategoricalFilter
	}
	return nil
}

func (x *ActionSearchQuery) GetMinimumStopTime() *timestamppb.Timestamp {
	if x != nil {
		return x.MinimumStopTime
	}
	return nil
}

func (x *ActionSearchQuery) GetMaximumStopTime() *timestamppb.Timestamp {
	if x != nil {
		return x.MaximumStopTime
	}
	return nil
}

// An ObservationSearchQuery contains all possible criteria that can
// be used to search for observations.
type ObservationSearchQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The observation filter is a collection of all search criteria for
	// observations that are not intrinsically ordered.
	CategoricalFilter *ObservationCategoricalFilter `protobuf:"bytes,1,opt,name=categorical_filter,json=categoricalFilter,proto3" json:"categorical_filter,omitempty"`
	// The value number filter contains maximum and minimum times for values, if
	// any exist.
	ValueNumberFilter *ValueNumberOrderableObservationFilter `protobuf:"bytes,2,opt,name=value_number_filter,json=valueNumberFilter,proto3" json:"value_number_filter,omitempty"`
}

func (x *ObservationSearchQuery) Reset() {
	*x = ObservationSearchQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObservationSearchQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObservationSearchQuery) ProtoMessage() {}

func (x *ObservationSearchQuery) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObservationSearchQuery.ProtoReflect.Descriptor instead.
func (*ObservationSearchQuery) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_searchquery_proto_rawDescGZIP(), []int{1}
}

func (x *ObservationSearchQuery) GetCategoricalFilter() *ObservationCategoricalFilter {
	if x != nil {
		return x.CategoricalFilter
	}
	return nil
}

func (x *ObservationSearchQuery) GetValueNumberFilter() *ValueNumberOrderableObservationFilter {
	if x != nil {
		return x.ValueNumberFilter
	}
	return nil
}

// An ActionCategoricalFilter contains search criteria that are categorical,
// i.e. they are not considered intrinsically ordered.
type ActionCategoricalFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique name of the action.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The kind of the action. This is a course-grained classification.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// The ID of the associated swarming task.
	SwarmingTaskId string `protobuf:"bytes,3,opt,name=swarming_task_id,json=swarmingTaskId,proto3" json:"swarming_task_id,omitempty"`
	// The component that failed and is "blamed" by the event.
	FailComponent string `protobuf:"bytes,4,opt,name=fail_component,json=failComponent,proto3" json:"fail_component,omitempty"`
	// Whether to include actions with an unspecified status.
	IsStatusUnspecified bool `protobuf:"varint,5,opt,name=is_status_unspecified,json=isStatusUnspecified,proto3" json:"is_status_unspecified,omitempty"`
	// Whether to include successful actions.
	IsSuccess bool `protobuf:"varint,6,opt,name=is_success,json=isSuccess,proto3" json:"is_success,omitempty"`
	// Whether to include failed actions.
	IsFail bool `protobuf:"varint,7,opt,name=is_fail,json=isFail,proto3" json:"is_fail,omitempty"`
}

func (x *ActionCategoricalFilter) Reset() {
	*x = ActionCategoricalFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActionCategoricalFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActionCategoricalFilter) ProtoMessage() {}

func (x *ActionCategoricalFilter) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActionCategoricalFilter.ProtoReflect.Descriptor instead.
func (*ActionCategoricalFilter) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_searchquery_proto_rawDescGZIP(), []int{2}
}

func (x *ActionCategoricalFilter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ActionCategoricalFilter) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *ActionCategoricalFilter) GetSwarmingTaskId() string {
	if x != nil {
		return x.SwarmingTaskId
	}
	return ""
}

func (x *ActionCategoricalFilter) GetFailComponent() string {
	if x != nil {
		return x.FailComponent
	}
	return ""
}

func (x *ActionCategoricalFilter) GetIsStatusUnspecified() bool {
	if x != nil {
		return x.IsStatusUnspecified
	}
	return false
}

func (x *ActionCategoricalFilter) GetIsSuccess() bool {
	if x != nil {
		return x.IsSuccess
	}
	return false
}

func (x *ActionCategoricalFilter) GetIsFail() bool {
	if x != nil {
		return x.IsFail
	}
	return false
}

// An ObservationCategoricalFilter contains search criteria that are
// categorical, i.e. they are not considered intrinsically ordered.
type ObservationCategoricalFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique name of the observation.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The name of the associated action.
	ActionName string `protobuf:"bytes,2,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
	// The kind of action.
	ActionKind string `protobuf:"bytes,3,opt,name=action_kind,json=actionKind,proto3" json:"action_kind,omitempty"`
	// The swarming task ID of the associated action.
	SwarmingTaskId string `protobuf:"bytes,4,opt,name=swarming_task_id,json=swarmingTaskId,proto3" json:"swarming_task_id,omitempty"`
	// The high-level component that failed.
	FailComponent string `protobuf:"bytes,5,opt,name=fail_component,json=failComponent,proto3" json:"fail_component,omitempty"`
	// Whether to include actions with an unspecified status.
	IsActionStatusUnspecified bool `protobuf:"varint,6,opt,name=is_action_status_unspecified,json=isActionStatusUnspecified,proto3" json:"is_action_status_unspecified,omitempty"`
	// Whether to include successful actions.
	IsActionSuccess bool `protobuf:"varint,7,opt,name=is_action_success,json=isActionSuccess,proto3" json:"is_action_success,omitempty"`
	// Whether to include failed actions.
	IsActionFail bool `protobuf:"varint,8,opt,name=is_action_fail,json=isActionFail,proto3" json:"is_action_fail,omitempty"`
	// This is the kind of measurement (e.g. "disk percentage")
	MetricKind string `protobuf:"bytes,9,opt,name=metric_kind,json=metricKind,proto3" json:"metric_kind,omitempty"`
	// This is a string measurement.
	ValueString string `protobuf:"bytes,10,opt,name=value_string,json=valueString,proto3" json:"value_string,omitempty"`
}

func (x *ObservationCategoricalFilter) Reset() {
	*x = ObservationCategoricalFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObservationCategoricalFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObservationCategoricalFilter) ProtoMessage() {}

func (x *ObservationCategoricalFilter) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObservationCategoricalFilter.ProtoReflect.Descriptor instead.
func (*ObservationCategoricalFilter) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_searchquery_proto_rawDescGZIP(), []int{3}
}

func (x *ObservationCategoricalFilter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetActionKind() string {
	if x != nil {
		return x.ActionKind
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetSwarmingTaskId() string {
	if x != nil {
		return x.SwarmingTaskId
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetFailComponent() string {
	if x != nil {
		return x.FailComponent
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetIsActionStatusUnspecified() bool {
	if x != nil {
		return x.IsActionStatusUnspecified
	}
	return false
}

func (x *ObservationCategoricalFilter) GetIsActionSuccess() bool {
	if x != nil {
		return x.IsActionSuccess
	}
	return false
}

func (x *ObservationCategoricalFilter) GetIsActionFail() bool {
	if x != nil {
		return x.IsActionFail
	}
	return false
}

func (x *ObservationCategoricalFilter) GetMetricKind() string {
	if x != nil {
		return x.MetricKind
	}
	return ""
}

func (x *ObservationCategoricalFilter) GetValueString() string {
	if x != nil {
		return x.ValueString
	}
	return ""
}

type ValueNumberOrderableObservationFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The minimum stop time, if provided, is a lower bound on the stop time of
	// actions.
	MinimumStopTime *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=minimum_stop_time,json=minimumStopTime,proto3" json:"minimum_stop_time,omitempty"`
	// The maximum stop time, if provided, is an upper bound on the stop time of
	// actions.
	MaximumStopTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=maximum_stop_time,json=maximumStopTime,proto3" json:"maximum_stop_time,omitempty"`
	// The minimum value number, if provided, is a lower bound on the measurement
	// value.
	MinimumValueNumber float64 `protobuf:"fixed64,3,opt,name=minimum_value_number,json=minimumValueNumber,proto3" json:"minimum_value_number,omitempty"`
	// If minimum_value_is_zero is set to true, then an explicit minimum of zero
	// is intended.
	MinimumValueIsZero bool `protobuf:"varint,4,opt,name=minimum_value_is_zero,json=minimumValueIsZero,proto3" json:"minimum_value_is_zero,omitempty"`
	// The maximum value number, if provided, is an upper bound bound on the
	// measurement value.
	MaximumValueNumber float64 `protobuf:"fixed64,5,opt,name=maximum_value_number,json=maximumValueNumber,proto3" json:"maximum_value_number,omitempty"`
	// If maximum_value_is_zero is set to true, then an explicit maximum of zero
	// is intended.
	MaximumValueIsZero bool `protobuf:"varint,6,opt,name=maximum_value_is_zero,json=maximumValueIsZero,proto3" json:"maximum_value_is_zero,omitempty"`
}

func (x *ValueNumberOrderableObservationFilter) Reset() {
	*x = ValueNumberOrderableObservationFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValueNumberOrderableObservationFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValueNumberOrderableObservationFilter) ProtoMessage() {}

func (x *ValueNumberOrderableObservationFilter) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_searchquery_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValueNumberOrderableObservationFilter.ProtoReflect.Descriptor instead.
func (*ValueNumberOrderableObservationFilter) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_searchquery_proto_rawDescGZIP(), []int{4}
}

func (x *ValueNumberOrderableObservationFilter) GetMinimumStopTime() *timestamppb.Timestamp {
	if x != nil {
		return x.MinimumStopTime
	}
	return nil
}

func (x *ValueNumberOrderableObservationFilter) GetMaximumStopTime() *timestamppb.Timestamp {
	if x != nil {
		return x.MaximumStopTime
	}
	return nil
}

func (x *ValueNumberOrderableObservationFilter) GetMinimumValueNumber() float64 {
	if x != nil {
		return x.MinimumValueNumber
	}
	return 0
}

func (x *ValueNumberOrderableObservationFilter) GetMinimumValueIsZero() bool {
	if x != nil {
		return x.MinimumValueIsZero
	}
	return false
}

func (x *ValueNumberOrderableObservationFilter) GetMaximumValueNumber() float64 {
	if x != nil {
		return x.MaximumValueNumber
	}
	return 0
}

func (x *ValueNumberOrderableObservationFilter) GetMaximumValueIsZero() bool {
	if x != nil {
		return x.MaximumValueIsZero
	}
	return false
}

var File_infra_cros_karte_api_searchquery_proto protoreflect.FileDescriptor

var file_infra_cros_karte_api_searchquery_proto_rawDesc = []byte{
	0x0a, 0x26, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6b, 0x61, 0x72,
	0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x71, 0x75, 0x65,
	0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65,
	0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfb, 0x01, 0x0a, 0x11, 0x41, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x56, 0x0a, 0x12, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x5f, 0x66,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x63, 0x68,
	0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2e, 0x41, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x52, 0x11, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61,
	0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x46, 0x0a, 0x11, 0x6d, 0x69, 0x6e, 0x69, 0x6d,
	0x75, 0x6d, 0x5f, 0x73, 0x74, 0x6f, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0f,
	0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x53, 0x74, 0x6f, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x46, 0x0a, 0x11, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x73, 0x74, 0x6f, 0x70, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0f, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x53,
	0x74, 0x6f, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x22, 0xdc, 0x01, 0x0a, 0x16, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x12, 0x5b, 0x0a, 0x12, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61,
	0x6c, 0x5f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c,
	0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2e,
	0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x61, 0x74, 0x65, 0x67,
	0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x11, 0x63, 0x61,
	0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12,
	0x65, 0x0a, 0x13, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x5f,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x63,
	0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2e, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x61, 0x62,
	0x6c, 0x65, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c,
	0x74, 0x65, 0x72, 0x52, 0x11, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0xfe, 0x01, 0x0a, 0x17, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x28, 0x0a, 0x10, 0x73, 0x77,
	0x61, 0x72, 0x6d, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x77, 0x61, 0x72, 0x6d, 0x69, 0x6e, 0x67, 0x54, 0x61,
	0x73, 0x6b, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c, 0x5f, 0x63, 0x6f, 0x6d,
	0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x66, 0x61,
	0x69, 0x6c, 0x43, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x12, 0x32, 0x0a, 0x15, 0x69,
	0x73, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x75, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x13, 0x69, 0x73, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x64, 0x12,
	0x1d, 0x0a, 0x0a, 0x69, 0x73, 0x5f, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x09, 0x69, 0x73, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x17,
	0x0a, 0x07, 0x69, 0x73, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x06, 0x69, 0x73, 0x46, 0x61, 0x69, 0x6c, 0x22, 0x9c, 0x03, 0x0a, 0x1c, 0x4f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x63,
	0x61, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a,
	0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4b, 0x69, 0x6e, 0x64, 0x12, 0x28,
	0x0a, 0x10, 0x73, 0x77, 0x61, 0x72, 0x6d, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x5f,
	0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x77, 0x61, 0x72, 0x6d, 0x69,
	0x6e, 0x67, 0x54, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x66, 0x61, 0x69, 0x6c,
	0x5f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x66, 0x61, 0x69, 0x6c, 0x43, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x12,
	0x3f, 0x0a, 0x1c, 0x69, 0x73, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x5f, 0x75, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x19, 0x69, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x12, 0x2a, 0x0a, 0x11, 0x69, 0x73, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x69, 0x73, 0x41,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x24, 0x0a, 0x0e,
	0x69, 0x73, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x66, 0x61, 0x69, 0x6c, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x69, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x61,
	0x69, 0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f, 0x6b, 0x69, 0x6e,
	0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x4b,
	0x69, 0x6e, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x73, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22, 0x81, 0x03, 0x0a, 0x25, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x61, 0x62, 0x6c, 0x65, 0x4f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x12, 0x46, 0x0a, 0x11, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x73, 0x74, 0x6f, 0x70,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0f, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d,
	0x53, 0x74, 0x6f, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x46, 0x0a, 0x11, 0x6d, 0x61, 0x78, 0x69,
	0x6d, 0x75, 0x6d, 0x5f, 0x73, 0x74, 0x6f, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x0f, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x53, 0x74, 0x6f, 0x70, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x30, 0x0a, 0x14, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x12,
	0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x12, 0x31, 0x0a, 0x15, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x5f, 0x69, 0x73, 0x5f, 0x7a, 0x65, 0x72, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x12, 0x6d, 0x69, 0x6e, 0x69, 0x6d, 0x75, 0x6d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x49,
	0x73, 0x5a, 0x65, 0x72, 0x6f, 0x12, 0x30, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d,
	0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x12, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x31, 0x0a, 0x15, 0x6d, 0x61, 0x78, 0x69, 0x6d,
	0x75, 0x6d, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69, 0x73, 0x5f, 0x7a, 0x65, 0x72, 0x6f,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x12, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x49, 0x73, 0x5a, 0x65, 0x72, 0x6f, 0x42, 0x1e, 0x5a, 0x1c, 0x69, 0x6e,
	0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x3b, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_infra_cros_karte_api_searchquery_proto_rawDescOnce sync.Once
	file_infra_cros_karte_api_searchquery_proto_rawDescData = file_infra_cros_karte_api_searchquery_proto_rawDesc
)

func file_infra_cros_karte_api_searchquery_proto_rawDescGZIP() []byte {
	file_infra_cros_karte_api_searchquery_proto_rawDescOnce.Do(func() {
		file_infra_cros_karte_api_searchquery_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cros_karte_api_searchquery_proto_rawDescData)
	})
	return file_infra_cros_karte_api_searchquery_proto_rawDescData
}

var file_infra_cros_karte_api_searchquery_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_infra_cros_karte_api_searchquery_proto_goTypes = []interface{}{
	(*ActionSearchQuery)(nil),                     // 0: chromeos.karte.ActionSearchQuery
	(*ObservationSearchQuery)(nil),                // 1: chromeos.karte.ObservationSearchQuery
	(*ActionCategoricalFilter)(nil),               // 2: chromeos.karte.ActionCategoricalFilter
	(*ObservationCategoricalFilter)(nil),          // 3: chromeos.karte.ObservationCategoricalFilter
	(*ValueNumberOrderableObservationFilter)(nil), // 4: chromeos.karte.ValueNumberOrderableObservationFilter
	(*timestamppb.Timestamp)(nil),                 // 5: google.protobuf.Timestamp
}
var file_infra_cros_karte_api_searchquery_proto_depIdxs = []int32{
	2, // 0: chromeos.karte.ActionSearchQuery.categorical_filter:type_name -> chromeos.karte.ActionCategoricalFilter
	5, // 1: chromeos.karte.ActionSearchQuery.minimum_stop_time:type_name -> google.protobuf.Timestamp
	5, // 2: chromeos.karte.ActionSearchQuery.maximum_stop_time:type_name -> google.protobuf.Timestamp
	3, // 3: chromeos.karte.ObservationSearchQuery.categorical_filter:type_name -> chromeos.karte.ObservationCategoricalFilter
	4, // 4: chromeos.karte.ObservationSearchQuery.value_number_filter:type_name -> chromeos.karte.ValueNumberOrderableObservationFilter
	5, // 5: chromeos.karte.ValueNumberOrderableObservationFilter.minimum_stop_time:type_name -> google.protobuf.Timestamp
	5, // 6: chromeos.karte.ValueNumberOrderableObservationFilter.maximum_stop_time:type_name -> google.protobuf.Timestamp
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_infra_cros_karte_api_searchquery_proto_init() }
func file_infra_cros_karte_api_searchquery_proto_init() {
	if File_infra_cros_karte_api_searchquery_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cros_karte_api_searchquery_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActionSearchQuery); i {
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
		file_infra_cros_karte_api_searchquery_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObservationSearchQuery); i {
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
		file_infra_cros_karte_api_searchquery_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActionCategoricalFilter); i {
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
		file_infra_cros_karte_api_searchquery_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObservationCategoricalFilter); i {
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
		file_infra_cros_karte_api_searchquery_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValueNumberOrderableObservationFilter); i {
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
			RawDescriptor: file_infra_cros_karte_api_searchquery_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_cros_karte_api_searchquery_proto_goTypes,
		DependencyIndexes: file_infra_cros_karte_api_searchquery_proto_depIdxs,
		MessageInfos:      file_infra_cros_karte_api_searchquery_proto_msgTypes,
	}.Build()
	File_infra_cros_karte_api_searchquery_proto = out.File
	file_infra_cros_karte_api_searchquery_proto_rawDesc = nil
	file_infra_cros_karte_api_searchquery_proto_goTypes = nil
	file_infra_cros_karte_api_searchquery_proto_depIdxs = nil
}
