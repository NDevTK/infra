// Copyright 2021 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v5.26.1
// source: infra/cros/karte/api/cron.proto

package kartepb

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// PersistToBigqueryRequest does not contain any info, since
// PersistToBigquery is intended to be called as a cron job.
type PersistToBigqueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PersistToBigqueryRequest) Reset() {
	*x = PersistToBigqueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_cron_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PersistToBigqueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PersistToBigqueryRequest) ProtoMessage() {}

func (x *PersistToBigqueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_cron_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PersistToBigqueryRequest.ProtoReflect.Descriptor instead.
func (*PersistToBigqueryRequest) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_cron_proto_rawDescGZIP(), []int{0}
}

type PersistToBigqueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Created actions is a count of the actions that were created.
	CreatedActions int32 `protobuf:"varint,1,opt,name=created_actions,json=createdActions,proto3" json:"created_actions,omitempty"`
	// Created observations is a count of the observations that were created.
	CreatedObservations int32 `protobuf:"varint,2,opt,name=created_observations,json=createdObservations,proto3" json:"created_observations,omitempty"`
	// Succeeded is true if and only if no errors at all were encountered during persistence.
	Succeeded bool `protobuf:"varint,3,opt,name=succeeded,proto3" json:"succeeded,omitempty"`
}

func (x *PersistToBigqueryResponse) Reset() {
	*x = PersistToBigqueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_cros_karte_api_cron_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PersistToBigqueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PersistToBigqueryResponse) ProtoMessage() {}

func (x *PersistToBigqueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_infra_cros_karte_api_cron_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PersistToBigqueryResponse.ProtoReflect.Descriptor instead.
func (*PersistToBigqueryResponse) Descriptor() ([]byte, []int) {
	return file_infra_cros_karte_api_cron_proto_rawDescGZIP(), []int{1}
}

func (x *PersistToBigqueryResponse) GetCreatedActions() int32 {
	if x != nil {
		return x.CreatedActions
	}
	return 0
}

func (x *PersistToBigqueryResponse) GetCreatedObservations() int32 {
	if x != nil {
		return x.CreatedObservations
	}
	return 0
}

func (x *PersistToBigqueryResponse) GetSucceeded() bool {
	if x != nil {
		return x.Succeeded
	}
	return false
}

var File_infra_cros_karte_api_cron_proto protoreflect.FileDescriptor

var file_infra_cros_karte_api_cron_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f, 0x73, 0x2f, 0x6b, 0x61, 0x72,
	0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x72, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74,
	0x65, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x1a, 0x0a, 0x18, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x54, 0x6f, 0x42, 0x69, 0x67, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x95, 0x01, 0x0a, 0x19,
	0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x54, 0x6f, 0x42, 0x69, 0x67, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0e, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x31, 0x0a, 0x14, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x6f, 0x62,
	0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x13, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64,
	0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x73, 0x75, 0x63, 0x63, 0x65, 0x65,
	0x64, 0x65, 0x64, 0x32, 0x9d, 0x01, 0x0a, 0x09, 0x4b, 0x61, 0x72, 0x74, 0x65, 0x43, 0x72, 0x6f,
	0x6e, 0x12, 0x8f, 0x01, 0x0a, 0x11, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x54, 0x6f, 0x42,
	0x69, 0x67, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x28, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65,
	0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74,
	0x54, 0x6f, 0x42, 0x69, 0x67, 0x71, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x29, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x6f, 0x73, 0x2e, 0x6b, 0x61, 0x72,
	0x74, 0x65, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x54, 0x6f, 0x42, 0x69, 0x67, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x25, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x1f, 0x12, 0x1d, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x2f,
	0x70, 0x65, 0x72, 0x73, 0x69, 0x73, 0x74, 0x2d, 0x74, 0x6f, 0x2d, 0x62, 0x69, 0x67, 0x71, 0x75,
	0x65, 0x72, 0x79, 0x42, 0x1e, 0x5a, 0x1c, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x72, 0x6f,
	0x73, 0x2f, 0x6b, 0x61, 0x72, 0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x3b, 0x6b, 0x61, 0x72, 0x74,
	0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_cros_karte_api_cron_proto_rawDescOnce sync.Once
	file_infra_cros_karte_api_cron_proto_rawDescData = file_infra_cros_karte_api_cron_proto_rawDesc
)

func file_infra_cros_karte_api_cron_proto_rawDescGZIP() []byte {
	file_infra_cros_karte_api_cron_proto_rawDescOnce.Do(func() {
		file_infra_cros_karte_api_cron_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_cros_karte_api_cron_proto_rawDescData)
	})
	return file_infra_cros_karte_api_cron_proto_rawDescData
}

var file_infra_cros_karte_api_cron_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_infra_cros_karte_api_cron_proto_goTypes = []interface{}{
	(*PersistToBigqueryRequest)(nil),  // 0: chromeos.karte.PersistToBigqueryRequest
	(*PersistToBigqueryResponse)(nil), // 1: chromeos.karte.PersistToBigqueryResponse
}
var file_infra_cros_karte_api_cron_proto_depIdxs = []int32{
	0, // 0: chromeos.karte.KarteCron.PersistToBigquery:input_type -> chromeos.karte.PersistToBigqueryRequest
	1, // 1: chromeos.karte.KarteCron.PersistToBigquery:output_type -> chromeos.karte.PersistToBigqueryResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_infra_cros_karte_api_cron_proto_init() }
func file_infra_cros_karte_api_cron_proto_init() {
	if File_infra_cros_karte_api_cron_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_infra_cros_karte_api_cron_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PersistToBigqueryRequest); i {
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
		file_infra_cros_karte_api_cron_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PersistToBigqueryResponse); i {
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
			RawDescriptor: file_infra_cros_karte_api_cron_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_infra_cros_karte_api_cron_proto_goTypes,
		DependencyIndexes: file_infra_cros_karte_api_cron_proto_depIdxs,
		MessageInfos:      file_infra_cros_karte_api_cron_proto_msgTypes,
	}.Build()
	File_infra_cros_karte_api_cron_proto = out.File
	file_infra_cros_karte_api_cron_proto_rawDesc = nil
	file_infra_cros_karte_api_cron_proto_goTypes = nil
	file_infra_cros_karte_api_cron_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// KarteCronClient is the client API for KarteCron service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KarteCronClient interface {
	PersistToBigquery(ctx context.Context, in *PersistToBigqueryRequest, opts ...grpc.CallOption) (*PersistToBigqueryResponse, error)
}
type karteCronPRPCClient struct {
	client *prpc.Client
}

func NewKarteCronPRPCClient(client *prpc.Client) KarteCronClient {
	return &karteCronPRPCClient{client}
}

func (c *karteCronPRPCClient) PersistToBigquery(ctx context.Context, in *PersistToBigqueryRequest, opts ...grpc.CallOption) (*PersistToBigqueryResponse, error) {
	out := new(PersistToBigqueryResponse)
	err := c.client.Call(ctx, "chromeos.karte.KarteCron", "PersistToBigquery", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type karteCronClient struct {
	cc grpc.ClientConnInterface
}

func NewKarteCronClient(cc grpc.ClientConnInterface) KarteCronClient {
	return &karteCronClient{cc}
}

func (c *karteCronClient) PersistToBigquery(ctx context.Context, in *PersistToBigqueryRequest, opts ...grpc.CallOption) (*PersistToBigqueryResponse, error) {
	out := new(PersistToBigqueryResponse)
	err := c.cc.Invoke(ctx, "/chromeos.karte.KarteCron/PersistToBigquery", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KarteCronServer is the server API for KarteCron service.
type KarteCronServer interface {
	PersistToBigquery(context.Context, *PersistToBigqueryRequest) (*PersistToBigqueryResponse, error)
}

// UnimplementedKarteCronServer can be embedded to have forward compatible implementations.
type UnimplementedKarteCronServer struct {
}

func (*UnimplementedKarteCronServer) PersistToBigquery(context.Context, *PersistToBigqueryRequest) (*PersistToBigqueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PersistToBigquery not implemented")
}

func RegisterKarteCronServer(s prpc.Registrar, srv KarteCronServer) {
	s.RegisterService(&_KarteCron_serviceDesc, srv)
}

func _KarteCron_PersistToBigquery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PersistToBigqueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KarteCronServer).PersistToBigquery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chromeos.karte.KarteCron/PersistToBigquery",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KarteCronServer).PersistToBigquery(ctx, req.(*PersistToBigqueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KarteCron_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chromeos.karte.KarteCron",
	HandlerType: (*KarteCronServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PersistToBigquery",
			Handler:    _KarteCron_PersistToBigquery_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/cros/karte/api/cron.proto",
}
