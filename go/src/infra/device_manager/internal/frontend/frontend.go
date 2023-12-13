// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Prove that Server implements pb.DeviceLeaseServiceServer by instantiating a Server.
var _ api.DeviceLeaseServiceServer = (*Server)(nil)

// Server is a struct implements the pb.DeviceLeaseServiceServer.
type Server struct {
	api.UnimplementedDeviceLeaseServiceServer

	// retry defaults
	initialRetryBackoff time.Duration
	maxRetries          int
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

func (s *Server) LeaseDevice(ctx context.Context, r *api.LeaseDeviceRequest) (*api.LeaseDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "LeaseDevice is not implemented")
}

func (s *Server) ReleaseDevice(ctx context.Context, r *api.ReleaseDeviceRequest) (*api.ReleaseDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReleaseDevice is not implemented")
}

func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ExtendLease is not implemented")
}

func (s *Server) GetDevice(ctx context.Context, r *api.GetDeviceRequest) (*api.Device, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetDevice is not implemented")
}

func (s *Server) ListDevices(ctx context.Context, r *api.ListDevicesRequest) (*api.ListDevicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ListDevices is not implemented")
}
