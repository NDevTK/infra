// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.17.0
// source: infra/chromeperf/alert_groups/alert_groups.proto

package proto

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

// AlertGroupsClient is the client API for AlertGroups service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AlertGroupsClient interface {
	// NOTE: An upstream project depends on this API, please contact
	// webrtc-infra@google.com if making not backwards compatible changes.
	MergeAlertGroups(ctx context.Context, in *MergeAlertGroupsRequest, opts ...grpc.CallOption) (*AlertGroup, error)
}

type alertGroupsClient struct {
	cc grpc.ClientConnInterface
}

func NewAlertGroupsClient(cc grpc.ClientConnInterface) AlertGroupsClient {
	return &alertGroupsClient{cc}
}

func (c *alertGroupsClient) MergeAlertGroups(ctx context.Context, in *MergeAlertGroupsRequest, opts ...grpc.CallOption) (*AlertGroup, error) {
	out := new(AlertGroup)
	err := c.cc.Invoke(ctx, "/alert_groups.AlertGroups/MergeAlertGroups", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AlertGroupsServer is the server API for AlertGroups service.
// All implementations must embed UnimplementedAlertGroupsServer
// for forward compatibility
type AlertGroupsServer interface {
	// NOTE: An upstream project depends on this API, please contact
	// webrtc-infra@google.com if making not backwards compatible changes.
	MergeAlertGroups(context.Context, *MergeAlertGroupsRequest) (*AlertGroup, error)
	mustEmbedUnimplementedAlertGroupsServer()
}

// UnimplementedAlertGroupsServer must be embedded to have forward compatible implementations.
type UnimplementedAlertGroupsServer struct {
}

func (UnimplementedAlertGroupsServer) MergeAlertGroups(context.Context, *MergeAlertGroupsRequest) (*AlertGroup, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MergeAlertGroups not implemented")
}
func (UnimplementedAlertGroupsServer) mustEmbedUnimplementedAlertGroupsServer() {}

// UnsafeAlertGroupsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AlertGroupsServer will
// result in compilation errors.
type UnsafeAlertGroupsServer interface {
	mustEmbedUnimplementedAlertGroupsServer()
}

func RegisterAlertGroupsServer(s grpc.ServiceRegistrar, srv AlertGroupsServer) {
	s.RegisterService(&AlertGroups_ServiceDesc, srv)
}

func _AlertGroups_MergeAlertGroups_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MergeAlertGroupsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AlertGroupsServer).MergeAlertGroups(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/alert_groups.AlertGroups/MergeAlertGroups",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AlertGroupsServer).MergeAlertGroups(ctx, req.(*MergeAlertGroupsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AlertGroups_ServiceDesc is the grpc.ServiceDesc for AlertGroups service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AlertGroups_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "alert_groups.AlertGroups",
	HandlerType: (*AlertGroupsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MergeAlertGroups",
			Handler:    _AlertGroups_MergeAlertGroups_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "infra/chromeperf/alert_groups/alert_groups.proto",
}
