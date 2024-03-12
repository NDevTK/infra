// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/device_manager/internal/controller"
	"infra/device_manager/internal/database"
	"infra/device_manager/internal/external"
)

// Prove that Server implements pb.DeviceLeaseServiceServer by instantiating a Server.
var _ api.DeviceLeaseServiceServer = (*Server)(nil)

// Server is a struct implements the pb.DeviceLeaseServiceServer.
type Server struct {
	api.UnimplementedDeviceLeaseServiceServer

	// service clients
	dbClient     database.Client
	pubSubClient *pubsub.Client

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

// SetUpDBClient sets up a reusable database client for the server
func SetUpDBClient(ctx context.Context, server *Server, dbconf database.DatabaseConfig) error {
	db, err := database.ConnectDB(ctx, dbconf)
	if err != nil {
		return status.Errorf(codes.Internal, "SetUpDBClient: could not set up DB client: %s", err)
	}

	server.dbClient.Conn = db
	server.dbClient.Config = dbconf

	return nil
}

// SetUpPubSubClient sets up a reusable PubSub client for the server
func SetUpPubSubClient(ctx context.Context, server *Server, cloudProject string) error {
	var cp string
	if cloudProject == "" {
		cp = "fleet-device-manager-dev"
	} else {
		cp = cloudProject
	}

	client, err := external.NewPubSubClient(ctx, cp)
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: cannot set up PubSub client: %s", err)
		return err
	}
	server.pubSubClient = client
	return nil
}

// LeaseDevice takes a LeaseDeviceRequest and leases a corresponding device.
func (s *Server) LeaseDevice(ctx context.Context, r *api.LeaseDeviceRequest) (*api.LeaseDeviceResponse, error) {
	logging.Debugf(ctx, "LeaseDevice: received LeaseDeviceRequest %v", r)

	// Check idempotency of lease. Return if there is an existing unexpired lease.
	rsp, err := controller.CheckLeaseIdempotency(ctx, s.dbClient.Conn, r.GetIdempotencyKey())
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
	device, err := controller.GetDevice(ctx, s.dbClient.Conn, deviceID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "LeaseDevice: failed to find Device %s: %s", deviceID, err)
	}
	logging.Debugf(ctx, "LeaseDevice: found Device %s: %v", deviceID, device)

	if !controller.IsDeviceAvailable(ctx, device.GetState()) {
		return nil, status.Errorf(codes.Unavailable, "LeaseDevice: device %s is unavailable for lease", deviceID)
	}
	return controller.LeaseDevice(ctx, s.dbClient.Conn, s.pubSubClient, r, device)
}

func (s *Server) ReleaseDevice(ctx context.Context, r *api.ReleaseDeviceRequest) (*api.ReleaseDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReleaseDevice is not implemented")
}

// ExtendLease attempts to extend the lease on a device by ExtendLeaseRequest.
func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	logging.Debugf(ctx, "ExtendLease: received ExtendLeaseRequest %v", r)

	// Check idempotency of ExtendLeaseRequest. Return request if it is a
	// duplicate.
	rsp, err := controller.CheckExtensionIdempotency(ctx, s.dbClient.Conn, r.GetIdempotencyKey())
	if err != nil {
		return nil, err
	}
	if rsp.GetLeaseId() != "" {
		return rsp, nil
	}

	return controller.ExtendLease(ctx, s.dbClient.Conn, r)
}

// GetDevice takes a GetDeviceRequest and returns a corresponding device.
func (s *Server) GetDevice(ctx context.Context, r *api.GetDeviceRequest) (*api.Device, error) {
	logging.Debugf(ctx, "GetDevice: received GetDeviceRequest %v", r)
	if r.Name == "" {
		return nil, status.Errorf(codes.Internal, "GetDevice: request has no device name")
	}

	device, err := controller.GetDevice(ctx, s.dbClient.Conn, r.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetDevice: failed to get Device %s: %s", r.Name, err)
	}
	logging.Debugf(ctx, "GetDevice: received Device %v", device)
	return device, nil
}

func (s *Server) ListDevices(ctx context.Context, r *api.ListDevicesRequest) (*api.ListDevicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ListDevices is not implemented")
}
