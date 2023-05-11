// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	vmleaserpb "infra/vm_leaser/api/v1"
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

// computeInstancesClient interfaces the GCE instance client API.
type computeInstancesClient interface {
	Get(ctx context.Context, r *computepb.GetInstanceRequest, opts ...gax.CallOption) (*computepb.Instance, error)
	Insert(ctx context.Context, r *computepb.InsertInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error)
}

// Prove that Server implements pb.VMLeaserServiceServer by instantiating a Server
var _ api.VMLeaserServiceServer = (*Server)(nil)

// Server is a struct implements the pb.VMLeaserServiceServer
type Server struct {
	api.UnimplementedVMLeaserServiceServer
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *api.LeaseVMRequest) (*api.LeaseVMResponse, error) {
	logging.Infof(ctx, "[server:LeaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &api.LeaseVMResponse{}, status.Errorf(codes.Internal, "context canceled")
	}

	// Set defaults for LeaseVMRequest if needed.
	r = setDefaultLeaseVMRequest(r)

	if err := vmleaserpb.ValidateLeaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate lease request: %s", err)
	}

	// Appending "vm-" to satisfy GCE regex
	leaseID := fmt.Sprintf("vm-%s", uuid.New().String())
	expirationTime, err := computeExpirationTime(ctx, r.GetLeaseDuration())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to compute expiration time: %s", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create NewInstancesRESTClient: %s", err)
	}
	defer instancesClient.Close()

	err = createInstance(ctx, instancesClient, leaseID, expirationTime, r.GetHostReqs())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create instance: %s", err)
	}

	ins, err := getInstance(ctx, instancesClient, leaseID, r.GetHostReqs(), true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get instance: %s", err)
	}

	return &api.LeaseVMResponse{
		LeaseId: leaseID,
		Vm: &api.VM{
			Id: leaseID,
			Address: &api.VMAddress{
				// Only one NetworkInterface and AccessConfig should be available
				Host: ins.GetNetworkInterfaces()[0].GetAccessConfigs()[0].GetNatIP(),
				// Temporarily hardcode as port 22
				Port: 22,
			},
		},
	}, nil
}

// ExtendLease extends a VM lease
func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	logging.Infof(ctx, "[server:ExtendLease] Started")
	if ctx.Err() == context.Canceled {
		return &api.ExtendLeaseResponse{}, status.Errorf(codes.Internal, "context canceled")
	}

	return &api.ExtendLeaseResponse{}, nil
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *api.ReleaseVMRequest) (*api.ReleaseVMResponse, error) {
	logging.Infof(ctx, "[server:ReleaseVM] Started")
	if ctx.Err() == context.Canceled {
		return &api.ReleaseVMResponse{}, status.Errorf(codes.Internal, "context canceled")
	}

	// Set default values for ReleaseVMRequest if needed.
	r = setDefaultReleaseVMRequest(r)

	if err := vmleaserpb.ValidateReleaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate release request: %s", err)
	}

	err := deleteInstance(ctx, r)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete instance: %s", err)
	}

	return &api.ReleaseVMResponse{
		LeaseId: r.GetLeaseId(),
	}, nil
}

// getInstance gets a GCE instance based on lease id and GCE configs.
//
// getInstance returns a GCE instance with valid network interface and network
// IP. If no network is available, it does not return the instance.
func getInstance(ctx context.Context, client computeInstancesClient, leaseID string, hostReqs *api.VMRequirements, shouldPoll bool) (*computepb.Instance, error) {
	getReq := &computepb.GetInstanceRequest{
		Instance: leaseID,
		Project:  hostReqs.GetGceProject(),
		Zone:     hostReqs.GetGceRegion(),
	}

	var ins *computepb.Instance
	var err error
	if shouldPoll {
		// Implement a 30 second deadline for polling for the instance
		d := time.Now().Add(30 * time.Second)
		ctx, cancel := context.WithDeadline(ctx, d)
		defer cancel()

		err = poll(ctx, func(ctx context.Context) (bool, error) {
			ins, err = client.Get(ctx, getReq)
			if err != nil {
				return false, err
			}
			return true, nil
		}, 2*time.Second)
		if err != nil {
			return nil, err
		}
	} else {
		ins, err = client.Get(ctx, getReq)
		if err != nil {
			return nil, err
		}
	}

	if ins.GetNetworkInterfaces() == nil || ins.GetNetworkInterfaces()[0] == nil {
		return nil, errors.New("instance does not have a network interface")
	}
	if ins.GetNetworkInterfaces()[0].GetAccessConfigs() == nil || ins.GetNetworkInterfaces()[0].GetAccessConfigs()[0] == nil {
		return nil, errors.New("instance does not have an access config")
	}
	if ins.GetNetworkInterfaces()[0].GetAccessConfigs()[0].GetNatIP() == "" {
		return nil, errors.New("instance does not have a nat ip")
	}
	return ins, nil
}

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(ctx context.Context, client computeInstancesClient, leaseID string, expirationTime int64, hostReqs *api.VMRequirements) error {
	zone := hostReqs.GetGceRegion()
	req := &computepb.InsertInstanceRequest{
		Project: hostReqs.GetGceProject(),
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Name: proto.String(leaseID),
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
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name: proto.String("External NAT"),
						},
					},
					Network: proto.String(hostReqs.GetGceNetwork()),
				},
			},
		},
	}

	logging.Debugf(ctx, "instance request params: %v", req)
	op, err := client.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %v", err)
	}
	if op == nil {
		return errors.New("no operation returned for waiting")
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	logging.Infof(ctx, "instance created")
	return nil
}

// deleteInstance sends an instance deletion request to the Compute Engine API and waits for it to complete.
func deleteInstance(ctx context.Context, r *api.ReleaseVMRequest) error {
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
func setDefaultLeaseVMRequest(r *api.LeaseVMRequest) *api.LeaseVMRequest {
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
func setDefaultReleaseVMRequest(r *api.ReleaseVMRequest) *api.ReleaseVMRequest {
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

// poll is a generic polling function that polls by interval
//
// poll provides a generic implementation of calling f at interval, exits on
// error or ctx timeout. f return true to end poll early.
func poll(ctx context.Context, f func(context.Context) (bool, error), interval time.Duration) error {
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("context must have a deadline to avoid infinite polling")
	}
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			success, err := f(ctx)
			if err != nil {
				return err
			}
			if success {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
