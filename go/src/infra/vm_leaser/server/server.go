// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/protobuf/proto"

	pb "infra/vm_leaser/api/v1"
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

const (
	projectId string = "chrome-fleet-vm-leaser-dev"
)

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *pb.LeaseVMRequest) (*pb.LeaseVMResponse, error) {
	log.Println("[server:LeaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.LeaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	leaseId := fmt.Sprintf("test-vm-%s", strconv.FormatInt(time.Now().UnixMilli(), 10))

	err := createInstance(leaseId, r.GetHostReqs())
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
	log.Println("[server:ExtendLease] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ExtendLeaseResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	return &pb.ExtendLeaseResponse{}, nil
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *pb.ReleaseVMRequest) (*pb.ReleaseVMResponse, error) {
	log.Println("[server:ReleaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &pb.ReleaseVMResponse{}, fmt.Errorf("client cancelled: abandoning")
	}

	leaseId := r.GetLeaseId()
	// TODO (justinsuen): add zone as an argument
	zone := "us-central1-a"

	err := deleteInstance(leaseId, zone)
	if err != nil {
		return nil, err
	}

	return &pb.ReleaseVMResponse{
		LeaseId: leaseId,
	}, nil
}

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(leaseId string, hostReqs *pb.VMRequirements) error {
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient error: %v", err)
	}
	defer instancesClient.Close()

	zone := hostReqs.GetGceRegion()
	req := &computepb.InsertInstanceRequest{
		Project: projectId,
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
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: proto.String(hostReqs.GetGceNetwork()),
				},
			},
		},
	}

	log.Println("instance request params: ", req)
	op, err := instancesClient.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	log.Println("instance created")
	return nil
}

// deleteInstance sends an instance deletion request to the Compute Engine API and waits for it to complete.
func deleteInstance(leaseId, zone string) error {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient error: %v", err)
	}
	defer c.Close()

	req := &computepb.DeleteInstanceRequest{
		Instance: leaseId,
		Project:  projectId,
		Zone:     zone,
	}

	log.Println("instance request params: ", req)
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

	log.Println("instance deleted")
	return nil
}
