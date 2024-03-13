// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.7
// source: infra/cros/fleetcost/api/indicator.proto

package fleetcostpb

import (
	money "google.golang.org/genproto/googleapis/type/money"
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

// Each Indicator refers to cost from a particular item, e.g. DUT, servo, servers, operation, space.
type IndicatorType int32

const (
	IndicatorType_INDICATOR_TYPE_UNKNOWN    IndicatorType = 0
	IndicatorType_INDICATOR_TYPE_DUT        IndicatorType = 1
	IndicatorType_INDICATOR_TYPE_SERVO      IndicatorType = 2
	IndicatorType_INDICATOR_TYPE_USBHUB     IndicatorType = 3
	IndicatorType_INDICATOR_TYPE_SERVER     IndicatorType = 4
	IndicatorType_INDICATOR_TYPE_USB_DRIVE  IndicatorType = 5
	IndicatorType_INDICATOR_TYPE_LABSTATION IndicatorType = 6
	IndicatorType_INDICATOR_TYPE_SPACE      IndicatorType = 7
	IndicatorType_INDICATOR_TYPE_OPERATION  IndicatorType = 8
	IndicatorType_INDICATOR_TYPE_CLOUD      IndicatorType = 9
	IndicatorType_INDICATOR_TYPE_POWER      IndicatorType = 10
)

// Enum value maps for IndicatorType.
var (
	IndicatorType_name = map[int32]string{
		0:  "INDICATOR_TYPE_UNKNOWN",
		1:  "INDICATOR_TYPE_DUT",
		2:  "INDICATOR_TYPE_SERVO",
		3:  "INDICATOR_TYPE_USBHUB",
		4:  "INDICATOR_TYPE_SERVER",
		5:  "INDICATOR_TYPE_USB_DRIVE",
		6:  "INDICATOR_TYPE_LABSTATION",
		7:  "INDICATOR_TYPE_SPACE",
		8:  "INDICATOR_TYPE_OPERATION",
		9:  "INDICATOR_TYPE_CLOUD",
		10: "INDICATOR_TYPE_POWER",
	}
	IndicatorType_value = map[string]int32{
		"INDICATOR_TYPE_UNKNOWN":    0,
		"INDICATOR_TYPE_DUT":        1,
		"INDICATOR_TYPE_SERVO":      2,
		"INDICATOR_TYPE_USBHUB":     3,
		"INDICATOR_TYPE_SERVER":     4,
		"INDICATOR_TYPE_USB_DRIVE":  5,
		"INDICATOR_TYPE_LABSTATION": 6,
		"INDICATOR_TYPE_SPACE":      7,
		"INDICATOR_TYPE_OPERATION":  8,
		"INDICATOR_TYPE_CLOUD":      9,
		"INDICATOR_TYPE_POWER":      10,
	}
)

func (x IndicatorType) Enum() *IndicatorType {
	p := new(IndicatorType)
	*p = x
	return p
}

func (x IndicatorType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (IndicatorType) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_cros_fleetcost_api_indicator_proto_enumTypes[0].Descriptor()
}

func (IndicatorType) Type() protoreflect.EnumType {
	return &file_infra_cros_fleetcost_api_indicator_proto_enumTypes[0]
}

func (x IndicatorType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use IndicatorType.Descriptor instead.
func (IndicatorType) EnumDescriptor() ([]byte, []int) {
	return file_infra_cros_fleetcost_api_indicator_proto_rawDescGZIP(), []int{0}
}

// Location indicates location scope for costs that may vary in different sites.
type Location int32

const (
	Location_LOCATION_UNKNOWN Location = 0
	Location_LOCATION_ALL     Location = 1
	Location_LOCATION_SFO36   Location = 2
	Location_LOCATION_IAD65   Location = 3
	Location_LOCATION_ACS     Location = 4
)

// Enum value maps for Location.
var (
	Location_name = map[int32]string{
		0: "LOCATION_UNKNOWN",
		1: "LOCATION_ALL",
		2: "LOCATION_SFO36",
		3: "LOCATION_IAD65",
		4: "LOCATION_ACS",
	}
	Location_value = map[string]int32{
		"LOCATION_UNKNOWN": 0,
		"LOCATION_ALL":     1,
		"LOCATION_SFO36":   2,
		"LOCATION_IAD65":   3,
		"LOCATION_ACS":     4,
	}
)

func (x Location) Enum() *Location {
	p := new(Location)
	*p = x
	return p
}

func (x Location) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Location) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_cros_fleetcost_api_indicator_proto_enumTypes[1].Descriptor()
}

func (Location) Type() protoreflect.EnumType {
	return &file_infra_cros_fleetcost_api_indicator_proto_enumTypes[1]
}

func (x Location) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Location.Descriptor instead.
func (Location) EnumDescriptor() ([]byte, []int) {
	return file_infra_cros_fleetcost_api_indicator_proto_rawDescGZIP(), []int{1}
}

type CostCadence int32

const (
	CostCadence_COST_CADENCE_UNKNOWN  CostCadence = 0
	CostCadence_COST_CADENCE_ONE_TIME CostCadence = 1
	CostCadence_COST_CADENCE_ANNUALLY CostCadence = 2
	CostCadence_COST_CADENCE_MONTHLY  CostCadence = 3
	CostCadence_COST_CADENCE_DAILY    CostCadence = 4
	CostCadence_COST_CADENCE_HOURLY   CostCadence = 5
)

// Enum value maps for CostCadence.
var (
	CostCadence_name = map[int32]string{
		0: "COST_CADENCE_UNKNOWN",
		1: "COST_CADENCE_ONE_TIME",
		2: "COST_CADENCE_ANNUALLY",
		3: "COST_CADENCE_MONTHLY",
		4: "COST_CADENCE_DAILY",
		5: "COST_CADENCE_HOURLY",
	}
	CostCadence_value = map[string]int32{
		"COST_CADENCE_UNKNOWN":  0,
		"COST_CADENCE_ONE_TIME": 1,
		"COST_CADENCE_ANNUALLY": 2,
		"COST_CADENCE_MONTHLY":  3,
		"COST_CADENCE_DAILY":    4,
		"COST_CADENCE_HOURLY":   5,
	}
)

func (x CostCadence) Enum() *CostCadence {
	p := new(CostCadence)
	*p = x
	return p
}

func (x CostCadence) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CostCadence) Descriptor() protoreflect.EnumDescriptor {
	return file_infra_cros_fleetcost_api_indicator_proto_enumTypes[2].Descriptor()
}

func (CostCadence) Type() protoreflect.EnumType {
	return &file_infra_cros_fleetcost_api_indicator_proto_enumTypes[2]
}

func (x CostCadence) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CostCadence.Descriptor instead.
func (CostCadence) EnumDescriptor() ([]byte, []int) {
	return file_infra_cros_fleetcost_api_indicator_proto_rawDescGZIP(), []int{2}
}

// Any combination of type/board/model/sku/location will be unique.
type CostIndicator struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string        `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type  IndicatorType `protobuf:"varint,2,opt,name=type,proto3,enum=IndicatorType" json:"type,omitempty"`
	Board string        `protobuf:"bytes,3,opt,name=board,proto3" json:"board,omitempty"`
	Model string        `protobuf:"bytes,4,opt,name=model,proto3" json:"model,omitempty"`
	Sku   string        `protobuf:"bytes,5,opt,name=sku,proto3" json:"sku,omitempty"`
	Cost  *money.Money  `protobuf:"bytes,6,opt,name=cost,proto3" json:"cost,omitempty"`
	// How frequently will the cost occur.
	CostCadence CostCadence `protobuf:"varint,7,opt,name=cost_cadence,json=costCadence,proto3,enum=CostCadence" json:"cost_cadence,omitempty"`
	// Annual burnout rate, e.g. 0.1 meaning 10% of devices need to be replaced per
	// year, note this should apply to the cost associated with one time cadence.
	BurnoutRate float64  `protobuf:"fixed64,8,opt,name=burnout_rate,json=burnoutRate,proto3" json:"burnout_rate,omitempty"`
	Location    Location `protobuf:"varint,9,opt,name=location,proto3,enum=Location" json:"location,omitempty"`
	Description string   `protobuf:"bytes,10,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *CostIndicator) Reset() {
	*x = CostIndicator{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_fleetcost_api_indicator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CostIndicator) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CostIndicator) ProtoMessage() {}

func (x *CostIndicator) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_fleetcost_api_indicator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CostIndicator.ProtoReflect.Descriptor instead.
func (*CostIndicator) Descriptor() ([]byte, []int) {
	return file_infra_cros_fleetcost_api_indicator_proto_rawDescGZIP(), []int{0}
}

func (x *CostIndicator) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CostIndicator) GetType() IndicatorType {
	if x != nil {
		return x.Type
	}
	return IndicatorType_INDICATOR_TYPE_UNKNOWN
}

func (x *CostIndicator) GetBoard() string {
	if x != nil {
		return x.Board
	}
	return ""
}

func (x *CostIndicator) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *CostIndicator) GetSku() string {
	if x != nil {
		return x.Sku
	}
	return ""
}

func (x *CostIndicator) GetCost() *money.Money {
	if x != nil {
		return x.Cost
	}
	return nil
}

func (x *CostIndicator) GetCostCadence() CostCadence {
	if x != nil {
		return x.CostCadence
	}
	return CostCadence_COST_CADENCE_UNKNOWN
}

func (x *CostIndicator) GetBurnoutRate() float64 {
	if x != nil {
		return x.BurnoutRate
	}
	return 0
}

func (x *CostIndicator) GetLocation() Location {
	if x != nil {
		return x.Location
	}
	return Location_LOCATION_UNKNOWN
}

func (x *CostIndicator) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

var File_infra_cros_fleetcost_api_indicator_proto protoreflect.FileDescriptor

var file_infra_cros_fleetcost_api_indicator_proto_rawDesc = []byte{
	0x0a, 0x28, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x66, 0x6c, 0x65,
	0x65, 0x74, 0x63, 0x6f, 0x73, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6e, 0x64, 0x69, 0x63,
	0x61, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x6f, 0x6e, 0x65, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xca, 0x02, 0x0a, 0x0d, 0x43, 0x6f, 0x73, 0x74, 0x49, 0x6e, 0x64, 0x69,
	0x63, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x49, 0x6e, 0x64, 0x69, 0x63, 0x61,
	0x74, 0x6f, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x62, 0x6f,
	0x61, 0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x6b, 0x75,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x6b, 0x75, 0x12, 0x26, 0x0a, 0x04, 0x63,
	0x6f, 0x73, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x4d, 0x6f, 0x6e, 0x65, 0x79, 0x52, 0x04, 0x63,
	0x6f, 0x73, 0x74, 0x12, 0x2f, 0x0a, 0x0c, 0x63, 0x6f, 0x73, 0x74, 0x5f, 0x63, 0x61, 0x64, 0x65,
	0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x43, 0x6f, 0x73, 0x74,
	0x43, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x73, 0x74, 0x43, 0x61, 0x64,
	0x65, 0x6e, 0x63, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x75, 0x72, 0x6e, 0x6f, 0x75, 0x74, 0x5f,
	0x72, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0b, 0x62, 0x75, 0x72, 0x6e,
	0x6f, 0x75, 0x74, 0x52, 0x61, 0x74, 0x65, 0x12, 0x25, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x09, 0x2e, 0x4c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x2a, 0xbc, 0x02, 0x0a, 0x0d, 0x49, 0x6e, 0x64, 0x69, 0x63, 0x61, 0x74, 0x6f, 0x72, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f,
	0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x16,
	0x0a, 0x12, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x44, 0x55, 0x54, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41,
	0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x4f, 0x10, 0x02,
	0x12, 0x19, 0x0a, 0x15, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x55, 0x53, 0x42, 0x48, 0x55, 0x42, 0x10, 0x03, 0x12, 0x19, 0x0a, 0x15, 0x49,
	0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x45,
	0x52, 0x56, 0x45, 0x52, 0x10, 0x04, 0x12, 0x1c, 0x0a, 0x18, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41,
	0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x53, 0x42, 0x5f, 0x44, 0x52, 0x49,
	0x56, 0x45, 0x10, 0x05, 0x12, 0x1d, 0x0a, 0x19, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f,
	0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4c, 0x41, 0x42, 0x53, 0x54, 0x41, 0x54, 0x49, 0x4f,
	0x4e, 0x10, 0x06, 0x12, 0x18, 0x0a, 0x14, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52,
	0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x50, 0x41, 0x43, 0x45, 0x10, 0x07, 0x12, 0x1c, 0x0a,
	0x18, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x08, 0x12, 0x18, 0x0a, 0x14, 0x49,
	0x4e, 0x44, 0x49, 0x43, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x43, 0x4c,
	0x4f, 0x55, 0x44, 0x10, 0x09, 0x12, 0x18, 0x0a, 0x14, 0x49, 0x4e, 0x44, 0x49, 0x43, 0x41, 0x54,
	0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x4f, 0x57, 0x45, 0x52, 0x10, 0x0a, 0x2a,
	0x6c, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x10, 0x4c,
	0x4f, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x10, 0x0a, 0x0c, 0x4c, 0x4f, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x41, 0x4c,
	0x4c, 0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x4c, 0x4f, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f,
	0x53, 0x46, 0x4f, 0x33, 0x36, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x4c, 0x4f, 0x43, 0x41, 0x54,
	0x49, 0x4f, 0x4e, 0x5f, 0x49, 0x41, 0x44, 0x36, 0x35, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x4c,
	0x4f, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x41, 0x43, 0x53, 0x10, 0x04, 0x2a, 0xa8, 0x01,
	0x0a, 0x0b, 0x43, 0x6f, 0x73, 0x74, 0x43, 0x61, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x18, 0x0a,
	0x14, 0x43, 0x4f, 0x53, 0x54, 0x5f, 0x43, 0x41, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f, 0x55, 0x4e,
	0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x19, 0x0a, 0x15, 0x43, 0x4f, 0x53, 0x54, 0x5f,
	0x43, 0x41, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f, 0x4f, 0x4e, 0x45, 0x5f, 0x54, 0x49, 0x4d, 0x45,
	0x10, 0x01, 0x12, 0x19, 0x0a, 0x15, 0x43, 0x4f, 0x53, 0x54, 0x5f, 0x43, 0x41, 0x44, 0x45, 0x4e,
	0x43, 0x45, 0x5f, 0x41, 0x4e, 0x4e, 0x55, 0x41, 0x4c, 0x4c, 0x59, 0x10, 0x02, 0x12, 0x18, 0x0a,
	0x14, 0x43, 0x4f, 0x53, 0x54, 0x5f, 0x43, 0x41, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f, 0x4d, 0x4f,
	0x4e, 0x54, 0x48, 0x4c, 0x59, 0x10, 0x03, 0x12, 0x16, 0x0a, 0x12, 0x43, 0x4f, 0x53, 0x54, 0x5f,
	0x43, 0x41, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f, 0x44, 0x41, 0x49, 0x4c, 0x59, 0x10, 0x04, 0x12,
	0x17, 0x0a, 0x13, 0x43, 0x4f, 0x53, 0x54, 0x5f, 0x43, 0x41, 0x44, 0x45, 0x4e, 0x43, 0x45, 0x5f,
	0x48, 0x4f, 0x55, 0x52, 0x4c, 0x59, 0x10, 0x05, 0x42, 0x26, 0x5a, 0x24, 0x69, 0x6e, 0x66, 0x72,
	0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x63, 0x6f, 0x73, 0x74,
	0x2f, 0x61, 0x70, 0x69, 0x3b, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x63, 0x6f, 0x73, 0x74, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_cros_fleetcost_api_indicator_proto_rawDescOnce sync.Once
	file_infra_cros_fleetcost_api_indicator_proto_rawDescData = file_infra_cros_fleetcost_api_indicator_proto_rawDesc
)

func file_infra_cros_fleetcost_api_indicator_proto_rawDescGZIP() []byte {
	file_infra_cros_fleetcost_api_indicator_proto_rawDescOnce.Do(func() {
		file_infra_cros_fleetcost_api_indicator_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cros_fleetcost_api_indicator_proto_rawDescData)
	})
	return file_infra_cros_fleetcost_api_indicator_proto_rawDescData
}

var file_infra_cros_fleetcost_api_indicator_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_infra_cros_fleetcost_api_indicator_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_cros_fleetcost_api_indicator_proto_goTypes = []interface{}{
	(IndicatorType)(0),    // 0: IndicatorType
	(Location)(0),         // 1: Location
	(CostCadence)(0),      // 2: CostCadence
	(*CostIndicator)(nil), // 3: CostIndicator
	(*money.Money)(nil),   // 4: google.type.Money
}
var file_infra_cros_fleetcost_api_indicator_proto_depIdxs = []int32{
	0, // 0: CostIndicator.type:type_name -> IndicatorType
	4, // 1: CostIndicator.cost:type_name -> google.type.Money
	2, // 2: CostIndicator.cost_cadence:type_name -> CostCadence
	1, // 3: CostIndicator.location:type_name -> Location
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_infra_cros_fleetcost_api_indicator_proto_init() }
func file_infra_cros_fleetcost_api_indicator_proto_init() {
	if File_infra_cros_fleetcost_api_indicator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cros_fleetcost_api_indicator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CostIndicator); i {
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
			RawDescriptor: file_infra_cros_fleetcost_api_indicator_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_infra_cros_fleetcost_api_indicator_proto_goTypes,
		DependencyIndexes: file_infra_cros_fleetcost_api_indicator_proto_depIdxs,
		EnumInfos:         file_infra_cros_fleetcost_api_indicator_proto_enumTypes,
		MessageInfos:      file_infra_cros_fleetcost_api_indicator_proto_msgTypes,
	}.Build()
	File_infra_cros_fleetcost_api_indicator_proto = out.File
	file_infra_cros_fleetcost_api_indicator_proto_rawDesc = nil
	file_infra_cros_fleetcost_api_indicator_proto_goTypes = nil
	file_infra_cros_fleetcost_api_indicator_proto_depIdxs = nil
}
