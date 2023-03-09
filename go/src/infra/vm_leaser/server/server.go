// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/uuid"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	pb "infra/vm_leaser/api/v1"
)

// Default VM Leaser parameters
const (
	// Default disk size to use for VM creation
	DefaultDiskSize int64 = 20
	// Default machine type to use for VM creation
	DefaultMachineType string = "e2-medium"
	// Default network to use for VM creation
	DefaultNetwork string = "global/networks/default"
	// Default GCP Project to use
	DefaultProject string = "chrome-fleet-vm-leaser-dev"
	// Default region (zone) to use
	DefaultRegion string = "us-central1-a"
	// Default duration of lease (in minutes)
	DefaultLeaseDuration int64 = 60
)

// Prove that Server implements pb.VMLeaserServiceServer by instantiating a Server
var _ pb.VMLeaserServiceServer = (*Server)(nil)

// Server is a struct implements the pb.VMLeaserServiceServer
type Server struct {
	pb.UnimplementedVMLeaserServiceServer
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

func serviceContext(ctx context.Context) context.Context {
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Debug)
	return ctx
}

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *pb.LeaseVMRequest) (*pb.LeaseVMResponse, error) {
	ctx = serviceContext(ctx)
	logging.Infof(ctx, "[server:LeaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.LeaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	// Set defaults for LeaseVMRequest if needed.
	r = setDefaultLeaseVMRequest(r)

	if err := r.Validate(); err != nil {
		return nil, err
	}

	// Appending "vm-" to satisfy GCE regex
	leaseId := fmt.Sprintf("vm-%s", uuid.New().String())
	expirationTime, err := computeExpirationTime(ctx, r.GetLeaseDuration())
	if err != nil {
		return nil, err
	}

	err = createInstance(ctx, leaseId, expirationTime, r.GetHostReqs())
	if err != nil {
		return nil, err
	}

	return &pb.LeaseVMResponse{
		LeaseId: leaseId,
		Vm: &pb.VM{
			Id: leaseId,
		},
	}, nil
}

// ExtendLease extends a VM lease
func (s *Server) ExtendLease(ctx context.Context, r *pb.ExtendLeaseRequest) (*pb.ExtendLeaseResponse, error) {
	ctx = serviceContext(ctx)
	logging.Infof(ctx, "[server:ExtendLease] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ExtendLeaseResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.ExtendLeaseResponse{}, nil
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *pb.ReleaseVMRequest) (*pb.ReleaseVMResponse, error) {
	ctx = serviceContext(ctx)
	logging.Infof(ctx, "[server:ReleaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ReleaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	// Set default values for ReleaseVMRequest if needed.
	r = setDefaultReleaseVMRequest(r)

	if err := r.Validate(); err != nil {
		return nil, err
	}

	err := deleteInstance(ctx, r)
	if err != nil {
		return nil, err
	}

	return &pb.ReleaseVMResponse{
		LeaseId: r.GetLeaseId(),
	}, nil
}

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(ctx context.Context, leaseId string, expirationTime int64, hostReqs *pb.VMRequirements) error {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient error: %v", err)
	}
	defer instancesClient.Close()

	zone := hostReqs.GetGceRegion()
	req := &computepb.InsertInstanceRequest{
		Project: hostReqs.GetGceProject(),
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Name: proto.String(leaseId),
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  proto.Int64(hostReqs.GetGceDiskSize()),
						SourceImage: proto.String(hostReqs.GetGceImage()),
					},
					AutoDelete: proto.Bool(true),
					Boot:       proto.Bool(true),
					Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
				},
			},
			MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", zone, hostReqs.GetGceMachineType())),
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   proto.String("expiration_time"),
						Value: proto.String(strconv.FormatInt(expirationTime, 10)),
					},
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: proto.String(hostReqs.GetGceNetwork()),
				},
			},
		},
	}

	logging.Debugf(ctx, "instance request params: %v", req)
	op, err := instancesClient.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	logging.Infof(ctx, "instance created")
	return nil
}

// deleteInstance sends an instance deletion request to the Compute Engine API and waits for it to complete.
func deleteInstance(ctx context.Context, r *pb.ReleaseVMRequest) error {
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient error: %v", err)
	}
	defer c.Close()

	req := &computepb.DeleteInstanceRequest{
		Instance: r.GetLeaseId(),
		Project:  r.GetGceProject(),
		Zone:     r.GetGceRegion(),
	}

	logging.Debugf(ctx, "instance request params: %v", req)
	op, err := c.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete instance: %v", err)
	}

	// Duplicate requests will not error. Both requests will receive its own
	// response.
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	logging.Infof(ctx, "instance deleted")
	return nil
}

// setDefaultLeaseVMRequest sets default values for VMRequirements.
func setDefaultLeaseVMRequest(r *pb.LeaseVMRequest) *pb.LeaseVMRequest {
	hostReqs := r.GetHostReqs()
	if hostReqs.GetGceDiskSize() == 0 {
		hostReqs.GceDiskSize = DefaultDiskSize
	}
	if hostReqs.GetGceMachineType() == "" {
		hostReqs.GceMachineType = DefaultMachineType
	}
	if hostReqs.GetGceNetwork() == "" {
		hostReqs.GceNetwork = DefaultNetwork
	}
	if hostReqs.GetGceProject() == "" {
		hostReqs.GceProject = DefaultProject
	}
	if hostReqs.GetGceRegion() == "" {
		hostReqs.GceRegion = DefaultRegion
	}
	return r
}

// setDefaultReleaseVMRequest sets default values for ReleaseVMRequest.
func setDefaultReleaseVMRequest(r *pb.ReleaseVMRequest) *pb.ReleaseVMRequest {
	if r.GetGceProject() == "" {
		r.GceProject = DefaultProject
	}
	if r.GetGceRegion() == "" {
		r.GceRegion = DefaultRegion
	}
	return r
}

// computeExpirationTime calculates the expiration time of a VM
//
// computeExpirationTime return a future Unix time as an int64. The calculation
// is based on the specified lease duration.
func computeExpirationTime(ctx context.Context, leaseDuration *durationpb.Duration) (int64, error) {
	expirationTime := time.Now().Unix()
	if leaseDuration == nil {
		return expirationTime + (DefaultLeaseDuration * 60), nil
	}
	return expirationTime + leaseDuration.GetSeconds(), nil
}
