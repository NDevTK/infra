// Copyright 2020 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.0
// source: infra/chromeperf/workflows/workflows_service.proto

package workflows

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

type CreateWorkflowRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The Workflow to create.
	Workflow *Workflow `protobuf:"bytes,1,opt,name=workflow,proto3" json:"workflow,omitempty"`
	// A unique identifier for the request. A random UUID is recommended.
	// This is used to ensure request idempotency. If empty, idempotency is not
	// guaranteed.
	RequestId string `protobuf:"bytes,2,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *CreateWorkflowRequest) Reset() {
	*x = CreateWorkflowRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_infra_chromeperf_workflows_workflows_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateWorkflowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWorkflowRequest) ProtoMessage() {}

func (x *CreateWorkflowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_infra_chromeperf_workflows_workflows_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWorkflowRequest.ProtoReflect.Descriptor instead.
func (*CreateWorkflowRequest) Descriptor() ([]byte, []int) {
	return file_infra_chromeperf_workflows_workflows_service_proto_rawDescGZIP(), []int{0}
}

func (x *CreateWorkflowRequest) GetWorkflow() *Workflow {
	if x != nil {
		return x.Workflow
	}
	return nil
}

func (x *CreateWorkflowRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

var File_infra_chromeperf_workflows_workflows_service_proto protoreflect.FileDescriptor

var file_infra_chromeperf_workflows_workflows_service_proto_rawDesc = []byte{
	0x0a, 0x32, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65,
	0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x1a,
	0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f,
	0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2a,
	0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70, 0x65, 0x72, 0x66,
	0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66,
	0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6c, 0x0a, 0x15, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x34, 0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x73, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x32, 0x75, 0x0a, 0x09, 0x57, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x12, 0x68, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x20, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c,
	0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x22, 0x1f,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x22, 0x0d, 0x2f, 0x76, 0x31, 0x2f, 0x77, 0x6f, 0x72, 0x6b,
	0x66, 0x6c, 0x6f, 0x77, 0x73, 0x3a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x42,
	0x1c, 0x5a, 0x1a, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x70,
	0x65, 0x72, 0x66, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x73, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_infra_chromeperf_workflows_workflows_service_proto_rawDescOnce sync.Once
	file_infra_chromeperf_workflows_workflows_service_proto_rawDescData = file_infra_chromeperf_workflows_workflows_service_proto_rawDesc
)

func file_infra_chromeperf_workflows_workflows_service_proto_rawDescGZIP() []byte {
	file_infra_chromeperf_workflows_workflows_service_proto_rawDescOnce.Do(func() {
		file_infra_chromeperf_workflows_workflows_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_infra_chromeperf_workflows_workflows_service_proto_rawDescData)
	})
	return file_infra_chromeperf_workflows_workflows_service_proto_rawDescData
}

var file_infra_chromeperf_workflows_workflows_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_infra_chromeperf_workflows_workflows_service_proto_goTypes = []interface{}{
	(*CreateWorkflowRequest)(nil), // 0: workflows.CreateWorkflowRequest
	(*Workflow)(nil),              // 1: workflows.Workflow
}
var file_infra_chromeperf_workflows_workflows_service_proto_depIdxs = []int32{
	1, // 0: workflows.CreateWorkflowRequest.workflow:type_name -> workflows.Workflow
	0, // 1: workflows.Workflows.CreateWorkflow:input_type -> workflows.CreateWorkflowRequest
	1, // 2: workflows.Workflows.CreateWorkflow:output_type -> workflows.Workflow
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_infra_chromeperf_workflows_workflows_service_proto_init() }
func file_infra_chromeperf_workflows_workflows_service_proto_init() {
	if File_infra_chromeperf_workflows_workflows_service_proto != nil {
		return
	}
	file_infra_chromeperf_workflows_workflows_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_infra_chromeperf_workflows_workflows_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateWorkflowRequest); i {
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
			RawDescriptor: file_infra_chromeperf_workflows_workflows_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_infra_chromeperf_workflows_workflows_service_proto_goTypes,
		DependencyIndexes: file_infra_chromeperf_workflows_workflows_service_proto_depIdxs,
		MessageInfos:      file_infra_chromeperf_workflows_workflows_service_proto_msgTypes,
	}.Build()
	File_infra_chromeperf_workflows_workflows_service_proto = out.File
	file_infra_chromeperf_workflows_workflows_service_proto_rawDesc = nil
	file_infra_chromeperf_workflows_workflows_service_proto_goTypes = nil
	file_infra_chromeperf_workflows_workflows_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// WorkflowsClient is the client API for Workflows service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type WorkflowsClient interface {
	CreateWorkflow(ctx context.Context, in *CreateWorkflowRequest, opts ...grpc.CallOption) (*Workflow, error)
}
type workflowsPRPCClient struct {
	client *prpc.Client
}

func NewWorkflowsPRPCClient(client *prpc.Client) WorkflowsClient {
	return &workflowsPRPCClient{client}
}

func (c *workflowsPRPCClient) CreateWorkflow(ctx context.Context, in *CreateWorkflowRequest, opts ...grpc.CallOption) (*Workflow, error) {
	out := new(Workflow)
	err := c.client.Call(ctx, "workflows.Workflows", "CreateWorkflow", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type workflowsClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkflowsClient(cc grpc.ClientConnInterface) WorkflowsClient {
	return &workflowsClient{cc}
}

func (c *workflowsClient) CreateWorkflow(ctx context.Context, in *CreateWorkflowRequest, opts ...grpc.CallOption) (*Workflow, error) {
	out := new(Workflow)
	err := c.cc.Invoke(ctx, "/workflows.Workflows/CreateWorkflow", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkflowsServer is the server API for Workflows service.
type WorkflowsServer interface {
	CreateWorkflow(context.Context, *CreateWorkflowRequest) (*Workflow, error)
}

// UnimplementedWorkflowsServer can be embedded to have forward compatible implementations.
type UnimplementedWorkflowsServer struct {
}

func (*UnimplementedWorkflowsServer) CreateWorkflow(context.Context, *CreateWorkflowRequest) (*Workflow, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWorkflow not implemented")
}

func RegisterWorkflowsServer(s prpc.Registrar, srv WorkflowsServer) {
	s.RegisterService(&_Workflows_serviceDesc, srv)
}

func _Workflows_CreateWorkflow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWorkflowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkflowsServer).CreateWorkflow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/workflows.Workflows/CreateWorkflow",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkflowsServer).CreateWorkflow(ctx, req.(*CreateWorkflowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Workflows_serviceDesc = grpc.ServiceDesc{
	ServiceName: "workflows.Workflows",
	HandlerType: (*WorkflowsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateWorkflow",
			Handler:    _Workflows_CreateWorkflow_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/chromeperf/workflows/workflows_service.proto",
}
