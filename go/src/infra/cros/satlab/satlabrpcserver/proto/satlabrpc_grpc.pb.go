// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.7
// source: satlabrpc.proto

package satlabrpcserver

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SatlabRpcService_ListBuildTargets_FullMethodName          = "/satlabrpcserver.SatlabRpcService/list_build_targets"
	SatlabRpcService_ListMilestones_FullMethodName            = "/satlabrpcserver.SatlabRpcService/list_milestones"
	SatlabRpcService_ListAccessibleModels_FullMethodName      = "/satlabrpcserver.SatlabRpcService/list_accessible_models"
	SatlabRpcService_ListBuildVersions_FullMethodName         = "/satlabrpcserver.SatlabRpcService/list_build_versions"
	SatlabRpcService_StageBuild_FullMethodName                = "/satlabrpcserver.SatlabRpcService/stage_build"
	SatlabRpcService_ListConnectedDutsFirmware_FullMethodName = "/satlabrpcserver.SatlabRpcService/list_connected_duts_firmware"
	SatlabRpcService_GetSystemInfo_FullMethodName             = "/satlabrpcserver.SatlabRpcService/get_system_info"
	SatlabRpcService_GetVersionInfo_FullMethodName            = "/satlabrpcserver.SatlabRpcService/get_version_info"
	SatlabRpcService_GetPeripheralInformation_FullMethodName  = "/satlabrpcserver.SatlabRpcService/get_peripheral_information"
	SatlabRpcService_UpdateDutsFirmware_FullMethodName        = "/satlabrpcserver.SatlabRpcService/update_duts_firmware"
	SatlabRpcService_RunSuite_FullMethodName                  = "/satlabrpcserver.SatlabRpcService/run_suite"
)

// SatlabRpcServiceClient is the client API for SatlabRpcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SatlabRpcServiceClient interface {
	ListBuildTargets(ctx context.Context, in *ListBuildTargetsRequest, opts ...grpc.CallOption) (*ListBuildTargetsResponse, error)
	ListMilestones(ctx context.Context, in *ListMilestonesRequest, opts ...grpc.CallOption) (*ListMilestonesResponse, error)
	ListAccessibleModels(ctx context.Context, in *ListAccessibleModelsRequest, opts ...grpc.CallOption) (*ListAccessibleModelsResponse, error)
	ListBuildVersions(ctx context.Context, in *ListBuildVersionsRequest, opts ...grpc.CallOption) (*ListBuildVersionsResponse, error)
	StageBuild(ctx context.Context, in *StageBuildRequest, opts ...grpc.CallOption) (*StageBuildResponse, error)
	ListConnectedDutsFirmware(ctx context.Context, in *ListConnectedDutsFirmwareRequest, opts ...grpc.CallOption) (*ListConnectedDutsFirmwareResponse, error)
	GetSystemInfo(ctx context.Context, in *GetSystemInfoRequest, opts ...grpc.CallOption) (*GetSystemInfoResponse, error)
	GetVersionInfo(ctx context.Context, in *GetVersionInfoRequest, opts ...grpc.CallOption) (*GetVersionInfoResponse, error)
	GetPeripheralInformation(ctx context.Context, in *GetPeripheralInformationRequest, opts ...grpc.CallOption) (*GetPeripheralInformationResponse, error)
	UpdateDutsFirmware(ctx context.Context, in *UpdateDutsFirmwareRequest, opts ...grpc.CallOption) (*UpdateDutsFirmwareResponse, error)
	// services to run different types of test suites
	RunSuite(ctx context.Context, in *RunSuiteRequest, opts ...grpc.CallOption) (*RunSuiteResponse, error)
}

type satlabRpcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSatlabRpcServiceClient(cc grpc.ClientConnInterface) SatlabRpcServiceClient {
	return &satlabRpcServiceClient{cc}
}

func (c *satlabRpcServiceClient) ListBuildTargets(ctx context.Context, in *ListBuildTargetsRequest, opts ...grpc.CallOption) (*ListBuildTargetsResponse, error) {
	out := new(ListBuildTargetsResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_ListBuildTargets_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) ListMilestones(ctx context.Context, in *ListMilestonesRequest, opts ...grpc.CallOption) (*ListMilestonesResponse, error) {
	out := new(ListMilestonesResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_ListMilestones_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) ListAccessibleModels(ctx context.Context, in *ListAccessibleModelsRequest, opts ...grpc.CallOption) (*ListAccessibleModelsResponse, error) {
	out := new(ListAccessibleModelsResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_ListAccessibleModels_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) ListBuildVersions(ctx context.Context, in *ListBuildVersionsRequest, opts ...grpc.CallOption) (*ListBuildVersionsResponse, error) {
	out := new(ListBuildVersionsResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_ListBuildVersions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) StageBuild(ctx context.Context, in *StageBuildRequest, opts ...grpc.CallOption) (*StageBuildResponse, error) {
	out := new(StageBuildResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_StageBuild_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) ListConnectedDutsFirmware(ctx context.Context, in *ListConnectedDutsFirmwareRequest, opts ...grpc.CallOption) (*ListConnectedDutsFirmwareResponse, error) {
	out := new(ListConnectedDutsFirmwareResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_ListConnectedDutsFirmware_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) GetSystemInfo(ctx context.Context, in *GetSystemInfoRequest, opts ...grpc.CallOption) (*GetSystemInfoResponse, error) {
	out := new(GetSystemInfoResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_GetSystemInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) GetVersionInfo(ctx context.Context, in *GetVersionInfoRequest, opts ...grpc.CallOption) (*GetVersionInfoResponse, error) {
	out := new(GetVersionInfoResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_GetVersionInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) GetPeripheralInformation(ctx context.Context, in *GetPeripheralInformationRequest, opts ...grpc.CallOption) (*GetPeripheralInformationResponse, error) {
	out := new(GetPeripheralInformationResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_GetPeripheralInformation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) UpdateDutsFirmware(ctx context.Context, in *UpdateDutsFirmwareRequest, opts ...grpc.CallOption) (*UpdateDutsFirmwareResponse, error) {
	out := new(UpdateDutsFirmwareResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_UpdateDutsFirmware_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *satlabRpcServiceClient) RunSuite(ctx context.Context, in *RunSuiteRequest, opts ...grpc.CallOption) (*RunSuiteResponse, error) {
	out := new(RunSuiteResponse)
	err := c.cc.Invoke(ctx, SatlabRpcService_RunSuite_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SatlabRpcServiceServer is the server API for SatlabRpcService service.
// All implementations must embed UnimplementedSatlabRpcServiceServer
// for forward compatibility
type SatlabRpcServiceServer interface {
	ListBuildTargets(context.Context, *ListBuildTargetsRequest) (*ListBuildTargetsResponse, error)
	ListMilestones(context.Context, *ListMilestonesRequest) (*ListMilestonesResponse, error)
	ListAccessibleModels(context.Context, *ListAccessibleModelsRequest) (*ListAccessibleModelsResponse, error)
	ListBuildVersions(context.Context, *ListBuildVersionsRequest) (*ListBuildVersionsResponse, error)
	StageBuild(context.Context, *StageBuildRequest) (*StageBuildResponse, error)
	ListConnectedDutsFirmware(context.Context, *ListConnectedDutsFirmwareRequest) (*ListConnectedDutsFirmwareResponse, error)
	GetSystemInfo(context.Context, *GetSystemInfoRequest) (*GetSystemInfoResponse, error)
	GetVersionInfo(context.Context, *GetVersionInfoRequest) (*GetVersionInfoResponse, error)
	GetPeripheralInformation(context.Context, *GetPeripheralInformationRequest) (*GetPeripheralInformationResponse, error)
	UpdateDutsFirmware(context.Context, *UpdateDutsFirmwareRequest) (*UpdateDutsFirmwareResponse, error)
	// services to run different types of test suites
	RunSuite(context.Context, *RunSuiteRequest) (*RunSuiteResponse, error)
	mustEmbedUnimplementedSatlabRpcServiceServer()
}

// UnimplementedSatlabRpcServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSatlabRpcServiceServer struct {
}

func (UnimplementedSatlabRpcServiceServer) ListBuildTargets(context.Context, *ListBuildTargetsRequest) (*ListBuildTargetsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBuildTargets not implemented")
}
func (UnimplementedSatlabRpcServiceServer) ListMilestones(context.Context, *ListMilestonesRequest) (*ListMilestonesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMilestones not implemented")
}
func (UnimplementedSatlabRpcServiceServer) ListAccessibleModels(context.Context, *ListAccessibleModelsRequest) (*ListAccessibleModelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccessibleModels not implemented")
}
func (UnimplementedSatlabRpcServiceServer) ListBuildVersions(context.Context, *ListBuildVersionsRequest) (*ListBuildVersionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBuildVersions not implemented")
}
func (UnimplementedSatlabRpcServiceServer) StageBuild(context.Context, *StageBuildRequest) (*StageBuildResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StageBuild not implemented")
}
func (UnimplementedSatlabRpcServiceServer) ListConnectedDutsFirmware(context.Context, *ListConnectedDutsFirmwareRequest) (*ListConnectedDutsFirmwareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListConnectedDutsFirmware not implemented")
}
func (UnimplementedSatlabRpcServiceServer) GetSystemInfo(context.Context, *GetSystemInfoRequest) (*GetSystemInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSystemInfo not implemented")
}
func (UnimplementedSatlabRpcServiceServer) GetVersionInfo(context.Context, *GetVersionInfoRequest) (*GetVersionInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersionInfo not implemented")
}
func (UnimplementedSatlabRpcServiceServer) GetPeripheralInformation(context.Context, *GetPeripheralInformationRequest) (*GetPeripheralInformationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPeripheralInformation not implemented")
}
func (UnimplementedSatlabRpcServiceServer) UpdateDutsFirmware(context.Context, *UpdateDutsFirmwareRequest) (*UpdateDutsFirmwareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDutsFirmware not implemented")
}
func (UnimplementedSatlabRpcServiceServer) RunSuite(context.Context, *RunSuiteRequest) (*RunSuiteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunSuite not implemented")
}
func (UnimplementedSatlabRpcServiceServer) mustEmbedUnimplementedSatlabRpcServiceServer() {}

// UnsafeSatlabRpcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SatlabRpcServiceServer will
// result in compilation errors.
type UnsafeSatlabRpcServiceServer interface {
	mustEmbedUnimplementedSatlabRpcServiceServer()
}

func RegisterSatlabRpcServiceServer(s grpc.ServiceRegistrar, srv SatlabRpcServiceServer) {
	s.RegisterService(&SatlabRpcService_ServiceDesc, srv)
}

func _SatlabRpcService_ListBuildTargets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListBuildTargetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).ListBuildTargets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_ListBuildTargets_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).ListBuildTargets(ctx, req.(*ListBuildTargetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_ListMilestones_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMilestonesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).ListMilestones(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_ListMilestones_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).ListMilestones(ctx, req.(*ListMilestonesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_ListAccessibleModels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAccessibleModelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).ListAccessibleModels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_ListAccessibleModels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).ListAccessibleModels(ctx, req.(*ListAccessibleModelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_ListBuildVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListBuildVersionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).ListBuildVersions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_ListBuildVersions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).ListBuildVersions(ctx, req.(*ListBuildVersionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_StageBuild_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StageBuildRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).StageBuild(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_StageBuild_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).StageBuild(ctx, req.(*StageBuildRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_ListConnectedDutsFirmware_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListConnectedDutsFirmwareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).ListConnectedDutsFirmware(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_ListConnectedDutsFirmware_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).ListConnectedDutsFirmware(ctx, req.(*ListConnectedDutsFirmwareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_GetSystemInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSystemInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).GetSystemInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_GetSystemInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).GetSystemInfo(ctx, req.(*GetSystemInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_GetVersionInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).GetVersionInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_GetVersionInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).GetVersionInfo(ctx, req.(*GetVersionInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_GetPeripheralInformation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPeripheralInformationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).GetPeripheralInformation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_GetPeripheralInformation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).GetPeripheralInformation(ctx, req.(*GetPeripheralInformationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_UpdateDutsFirmware_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDutsFirmwareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).UpdateDutsFirmware(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_UpdateDutsFirmware_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).UpdateDutsFirmware(ctx, req.(*UpdateDutsFirmwareRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SatlabRpcService_RunSuite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunSuiteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SatlabRpcServiceServer).RunSuite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SatlabRpcService_RunSuite_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SatlabRpcServiceServer).RunSuite(ctx, req.(*RunSuiteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SatlabRpcService_ServiceDesc is the grpc.ServiceDesc for SatlabRpcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SatlabRpcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "satlabrpcserver.SatlabRpcService",
	HandlerType: (*SatlabRpcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "list_build_targets",
			Handler:    _SatlabRpcService_ListBuildTargets_Handler,
		},
		{
			MethodName: "list_milestones",
			Handler:    _SatlabRpcService_ListMilestones_Handler,
		},
		{
			MethodName: "list_accessible_models",
			Handler:    _SatlabRpcService_ListAccessibleModels_Handler,
		},
		{
			MethodName: "list_build_versions",
			Handler:    _SatlabRpcService_ListBuildVersions_Handler,
		},
		{
			MethodName: "stage_build",
			Handler:    _SatlabRpcService_StageBuild_Handler,
		},
		{
			MethodName: "list_connected_duts_firmware",
			Handler:    _SatlabRpcService_ListConnectedDutsFirmware_Handler,
		},
		{
			MethodName: "get_system_info",
			Handler:    _SatlabRpcService_GetSystemInfo_Handler,
		},
		{
			MethodName: "get_version_info",
			Handler:    _SatlabRpcService_GetVersionInfo_Handler,
		},
		{
			MethodName: "get_peripheral_information",
			Handler:    _SatlabRpcService_GetPeripheralInformation_Handler,
		},
		{
			MethodName: "update_duts_firmware",
			Handler:    _SatlabRpcService_UpdateDutsFirmware_Handler,
		},
		{
			MethodName: "run_suite",
			Handler:    _SatlabRpcService_RunSuite_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "satlabrpc.proto",
}
