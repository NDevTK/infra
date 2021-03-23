// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.1
// source: infra/appengine/statsui/api/service.proto

package api

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type Period int32

const (
	Period_WEEK Period = 0
	Period_DAY  Period = 1
)

// Enum value maps for Period.
var (
	Period_name = map[int32]string{
		0: "WEEK",
		1: "DAY",
	}
	Period_value = map[string]int32{
		"WEEK": 0,
		"DAY":  1,
	}
)

func (x Period) Enum() *Period {
	p := new(Period)
	*p = x
	return p
}

func (x Period) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Period) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_appengine_statsui_api_service_proto_enumTypes[0].Descriptor()
}

func (Period) Type() protoreflect.EnumType {
	return &file_infra_appengine_statsui_api_service_proto_enumTypes[0]
}

func (x Period) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Period.Descriptor instead.
func (Period) EnumDescriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{0}
}

type FetchMetricsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Datasource string   `protobuf:"bytes,1,opt,name=datasource,proto3" json:"datasource,omitempty"`
	Period     Period   `protobuf:"varint,2,opt,name=period,proto3,enum=statsui.Period" json:"period,omitempty"`
	Dates      []string `protobuf:"bytes,3,rep,name=dates,proto3" json:"dates,omitempty"`
	Metrics    []string `protobuf:"bytes,4,rep,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *FetchMetricsRequest) Reset() {
	*x = FetchMetricsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FetchMetricsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchMetricsRequest) ProtoMessage() {}

func (x *FetchMetricsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchMetricsRequest.ProtoReflect.Descriptor instead.
func (*FetchMetricsRequest) Descriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{0}
}

func (x *FetchMetricsRequest) GetDatasource() string {
	if x != nil {
		return x.Datasource
	}
	return ""
}

func (x *FetchMetricsRequest) GetPeriod() Period {
	if x != nil {
		return x.Period
	}
	return Period_WEEK
}

func (x *FetchMetricsRequest) GetDates() []string {
	if x != nil {
		return x.Dates
	}
	return nil
}

func (x *FetchMetricsRequest) GetMetrics() []string {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type Section struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name    string    `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Metrics []*Metric `protobuf:"bytes,2,rep,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *Section) Reset() {
	*x = Section{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Section) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Section) ProtoMessage() {}

func (x *Section) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Section.ProtoReflect.Descriptor instead.
func (*Section) Descriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{1}
}

func (x *Section) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Section) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type DataSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data map[string]float32 `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed32,2,opt,name=value,proto3"`
}

func (x *DataSet) Reset() {
	*x = DataSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataSet) ProtoMessage() {}

func (x *DataSet) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataSet.ProtoReflect.Descriptor instead.
func (*DataSet) Descriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{2}
}

func (x *DataSet) GetData() map[string]float32 {
	if x != nil {
		return x.Data
	}
	return nil
}

type Metric struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string              `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Data     *DataSet            `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Sections map[string]*DataSet `protobuf:"bytes,3,rep,name=sections,proto3" json:"sections,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Metric) Reset() {
	*x = Metric{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{3}
}

func (x *Metric) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Metric) GetData() *DataSet {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Metric) GetSections() map[string]*DataSet {
	if x != nil {
		return x.Sections
	}
	return nil
}

type FetchMetricsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sections []*Section `protobuf:"bytes,1,rep,name=sections,proto3" json:"sections,omitempty"`
}

func (x *FetchMetricsResponse) Reset() {
	*x = FetchMetricsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FetchMetricsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FetchMetricsResponse) ProtoMessage() {}

func (x *FetchMetricsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_appengine_statsui_api_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FetchMetricsResponse.ProtoReflect.Descriptor instead.
func (*FetchMetricsResponse) Descriptor() ([]byte, []int) {
	return file_infra_appengine_statsui_api_service_proto_rawDescGZIP(), []int{4}
}

func (x *FetchMetricsResponse) GetSections() []*Section {
	if x != nil {
		return x.Sections
	}
	return nil
}

var File_infra_appengine_statsui_api_service_proto protoreflect.FileDescriptor

var file_infra_appengine_statsui_api_service_proto_rawDesc = []byte{
	0x0a, 0x29, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x73, 0x74, 0x61,
	0x74, 0x73, 0x75, 0x69, 0x22, 0x8e, 0x01, 0x0a, 0x13, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x64, 0x61, 0x74, 0x61, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x27, 0x0a, 0x06,
	0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f, 0x2e, 0x73,
	0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x52, 0x06, 0x70,
	0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x64, 0x61, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x22, 0x48, 0x0a, 0x07, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x29, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x22,
	0x72, 0x0a, 0x07, 0x44, 0x61, 0x74, 0x61, 0x53, 0x65, 0x74, 0x12, 0x2e, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73,
	0x75, 0x69, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x53, 0x65, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x37, 0x0a, 0x09, 0x44, 0x61,
	0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x22, 0xcc, 0x01, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x24, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x53,
	0x65, 0x74, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x39, 0x0a, 0x08, 0x73, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x73, 0x74, 0x61,
	0x74, 0x73, 0x75, 0x69, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x2e, 0x53, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x1a, 0x4d, 0x0a, 0x0d, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x26, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e,
	0x44, 0x61, 0x74, 0x61, 0x53, 0x65, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x44, 0x0a, 0x14, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x08, 0x73, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73,
	0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e, 0x53, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08,
	0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2a, 0x1b, 0x0a, 0x06, 0x50, 0x65, 0x72, 0x69,
	0x6f, 0x64, 0x12, 0x08, 0x0a, 0x04, 0x57, 0x45, 0x45, 0x4b, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03,
	0x44, 0x41, 0x59, 0x10, 0x01, 0x32, 0x54, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x4b,
	0x0a, 0x0c, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x1c,
	0x2e, 0x73, 0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73,
	0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x1d, 0x5a, 0x1b, 0x69,
	0x6e, 0x66, 0x72, 0x61, 0x2f, 0x61, 0x70, 0x70, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x73,
	0x74, 0x61, 0x74, 0x73, 0x75, 0x69, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_infra_appengine_statsui_api_service_proto_rawDescOnce sync.Once
	file_infra_appengine_statsui_api_service_proto_rawDescData = file_infra_appengine_statsui_api_service_proto_rawDesc
)

func file_infra_appengine_statsui_api_service_proto_rawDescGZIP() []byte {
	file_infra_appengine_statsui_api_service_proto_rawDescOnce.Do(func() {
		file_infra_appengine_statsui_api_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_appengine_statsui_api_service_proto_rawDescData)
	})
	return file_infra_appengine_statsui_api_service_proto_rawDescData
}

var file_infra_appengine_statsui_api_service_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_infra_appengine_statsui_api_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_infra_appengine_statsui_api_service_proto_goTypes = []interface{}{
	(Period)(0),                  // 0: statsui.Period
	(*FetchMetricsRequest)(nil),  // 1: statsui.FetchMetricsRequest
	(*Section)(nil),              // 2: statsui.Section
	(*DataSet)(nil),              // 3: statsui.DataSet
	(*Metric)(nil),               // 4: statsui.Metric
	(*FetchMetricsResponse)(nil), // 5: statsui.FetchMetricsResponse
	nil,                          // 6: statsui.DataSet.DataEntry
	nil,                          // 7: statsui.Metric.SectionsEntry
}
var file_infra_appengine_statsui_api_service_proto_depIdxs = []int32{
	0, // 0: statsui.FetchMetricsRequest.period:type_name -> statsui.Period
	4, // 1: statsui.Section.metrics:type_name -> statsui.Metric
	6, // 2: statsui.DataSet.data:type_name -> statsui.DataSet.DataEntry
	3, // 3: statsui.Metric.data:type_name -> statsui.DataSet
	7, // 4: statsui.Metric.sections:type_name -> statsui.Metric.SectionsEntry
	2, // 5: statsui.FetchMetricsResponse.sections:type_name -> statsui.Section
	3, // 6: statsui.Metric.SectionsEntry.value:type_name -> statsui.DataSet
	1, // 7: statsui.Stats.FetchMetrics:input_type -> statsui.FetchMetricsRequest
	5, // 8: statsui.Stats.FetchMetrics:output_type -> statsui.FetchMetricsResponse
	8, // [8:9] is the sub-list for method output_type
	7, // [7:8] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_infra_appengine_statsui_api_service_proto_init() }
func file_infra_appengine_statsui_api_service_proto_init() {
	if File_infra_appengine_statsui_api_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_appengine_statsui_api_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FetchMetricsRequest); i {
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
		file_infra_appengine_statsui_api_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Section); i {
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
		file_infra_appengine_statsui_api_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataSet); i {
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
		file_infra_appengine_statsui_api_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metric); i {
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
		file_infra_appengine_statsui_api_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FetchMetricsResponse); i {
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
			RawDescriptor: file_infra_appengine_statsui_api_service_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_infra_appengine_statsui_api_service_proto_goTypes,
		DependencyIndexes: file_infra_appengine_statsui_api_service_proto_depIdxs,
		EnumInfos:         file_infra_appengine_statsui_api_service_proto_enumTypes,
		MessageInfos:      file_infra_appengine_statsui_api_service_proto_msgTypes,
	}.Build()
	File_infra_appengine_statsui_api_service_proto = out.File
	file_infra_appengine_statsui_api_service_proto_rawDesc = nil
	file_infra_appengine_statsui_api_service_proto_goTypes = nil
	file_infra_appengine_statsui_api_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// StatsClient is the client API for Stats service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StatsClient interface {
	FetchMetrics(ctx context.Context, in *FetchMetricsRequest, opts ...grpc.CallOption) (*FetchMetricsResponse, error)
}
type statsPRPCClient struct {
	client *prpc.Client
}

func NewStatsPRPCClient(client *prpc.Client) StatsClient {
	return &statsPRPCClient{client}
}

func (c *statsPRPCClient) FetchMetrics(ctx context.Context, in *FetchMetricsRequest, opts ...grpc.CallOption) (*FetchMetricsResponse, error) {
	out := new(FetchMetricsResponse)
	err := c.client.Call(ctx, "statsui.Stats", "FetchMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type statsClient struct {
	cc grpc.ClientConnInterface
}

func NewStatsClient(cc grpc.ClientConnInterface) StatsClient {
	return &statsClient{cc}
}

func (c *statsClient) FetchMetrics(ctx context.Context, in *FetchMetricsRequest, opts ...grpc.CallOption) (*FetchMetricsResponse, error) {
	out := new(FetchMetricsResponse)
	err := c.cc.Invoke(ctx, "/statsui.Stats/FetchMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StatsServer is the server API for Stats service.
type StatsServer interface {
	FetchMetrics(context.Context, *FetchMetricsRequest) (*FetchMetricsResponse, error)
}

// UnimplementedStatsServer can be embedded to have forward compatible implementations.
type UnimplementedStatsServer struct {
}

func (*UnimplementedStatsServer) FetchMetrics(context.Context, *FetchMetricsRequest) (*FetchMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMetrics not implemented")
}

func RegisterStatsServer(s prpc.Registrar, srv StatsServer) {
	s.RegisterService(&_Stats_serviceDesc, srv)
}

func _Stats_FetchMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatsServer).FetchMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/statsui.Stats/FetchMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatsServer).FetchMetrics(ctx, req.(*FetchMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Stats_serviceDesc = grpc.ServiceDesc{
	ServiceName: "statsui.Stats",
	HandlerType: (*StatsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchMetrics",
			Handler:    _Stats_FetchMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/appengine/statsui/api/service.proto",
}
