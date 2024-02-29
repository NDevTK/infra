// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/device_manager/internal/controller"
	"infra/device_manager/internal/database"
)

// Prove that Server implements pb.DeviceLeaseServiceServer by instantiating a Server.
var _ api.DeviceLeaseServiceServer = (*Server)(nil)

// Server is a struct implements the pb.DeviceLeaseServiceServer.
type Server struct {
	api.UnimplementedDeviceLeaseServiceServer

	// server options
	cloudProject string

	// database config
	dbConfig database.DatabaseConfig

	// retry defaults
	initialRetryBackoff time.Duration
	maxRetries          int
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// InstallServices takes a DeviceLeaseServiceServer and exposes it to a LUCI
// prpc.Server.
func InstallServices(s *Server, srv grpc.ServiceRegistrar) {
	api.RegisterDeviceLeaseServiceServer(srv, s)
}

// SetDBConfig sets the database password location string for the server.
func SetDBConfig(server *Server, dbconf database.DatabaseConfig) {
	server.dbConfig = dbconf
}

// SetCloudProject sets the cloud project string value for the server.
func SetCloudProject(server *Server, cp string) {
	if cp == "" {
		server.cloudProject = "fleet-device-manager-dev"
	} else {
		server.cloudProject = cp
	}
}

// LeaseDevice takes a LeaseDeviceRequest and leases a corresponding device.
func (s *Server) LeaseDevice(ctx context.Context, r *api.LeaseDeviceRequest) (*api.LeaseDeviceResponse, error) {
	logging.Debugf(ctx, "LeaseDevice: received LeaseDeviceRequest %v", r)

	opts := controller.RequestOpts{
		CloudProject: s.cloudProject,
	}
	db := database.ConnectDB(ctx, s.dbConfig)

	// Check idempotency of lease. Return if there is an existing unexpired lease.
	rsp, err := controller.CheckLeaseIdempotency(ctx, db, r.GetIdempotencyKey())
	if err != nil {
		return nil, err
	}
	if rsp.GetDeviceLease() != nil {
		return rsp, nil
	}

	// Parse hardware requirements. Initial iteration will take a deviceID and
	// search for the device to lease.
	deviceLabels := r.GetHardwareDeviceReqs().GetSchedulableLabels()
	deviceID := deviceLabels["device_id"].GetValues()[0] // assumes only leasing one device
	device, err := controller.GetDevice(ctx, db, deviceID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "LeaseDevice: failed to find Device %s: %s", deviceID, err)
	}
	logging.Debugf(ctx, "LeaseDevice: found Device %s: %v", deviceID, device)

	if !controller.IsDeviceAvailable(ctx, device.GetState()) {
		return nil, status.Errorf(codes.Unavailable, "LeaseDevice: device %s is unavailable for lease", deviceID)
	}
	return controller.LeaseDevice(ctx, db, opts, r, device)
}

func (s *Server) ReleaseDevice(ctx context.Context, r *api.ReleaseDeviceRequest) (*api.ReleaseDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReleaseDevice is not implemented")
}

func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ExtendLease is not implemented")
}

// GetDevice takes a GetDeviceRequest and returns a corresponding device.
func (s *Server) GetDevice(ctx context.Context, r *api.GetDeviceRequest) (*api.Device, error) {
	logging.Debugf(ctx, "GetDevice: received GetDeviceRequest %v", r)
	db := database.ConnectDB(ctx, s.dbConfig)
	if r.Name == "" {
		return nil, status.Errorf(codes.Internal, "GetDevice: request has no device name")
	}

	device, err := controller.GetDevice(ctx, db, r.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetDevice: failed to get Device %s: %s", r.Name, err)
	}
	logging.Debugf(ctx, "GetDevice: received Device %v", device)
	return device, nil
}

func (s *Server) ListDevices(ctx context.Context, r *api.ListDevicesRequest) (*api.ListDevicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ListDevices is not implemented")
}
