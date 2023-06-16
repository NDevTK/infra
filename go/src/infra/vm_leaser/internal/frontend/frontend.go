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
	"infra/vm_leaser/internal/constants"
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
	Env string
}

// NewServer returns a new Server
func NewServer(env string) *Server {
	return &Server{
		Env: env,
	}
}

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *api.LeaseVMRequest) (*api.LeaseVMResponse, error) {
	logging.Infof(ctx, "[server:LeaseVM] Started")

	// Set defaults for LeaseVMRequest if needed.
	r = setDefaultLeaseVMRequest(r, s.Env)

	if err := vmleaserpb.ValidateLeaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate lease request: %s", err)
	}

	// Appending "vm-" to satisfy GCE regex
	leaseID := fmt.Sprintf("vm-%s", uuid.New().String())
	expirationTime, err := computeExpirationTime(ctx, r.GetLeaseDuration(), s.Env)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to compute expiration time: %s", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create NewInstancesRESTClient: %s", err)
	}
	defer instancesClient.Close()

	err = createInstance(ctx, instancesClient, leaseID, expirationTime, r.GetHostReqs())
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.Errorf(codes.DeadlineExceeded, "when creating instance: %s", ctx.Err())
		}
		return nil, status.Errorf(codes.Internal, "failed to create instance: %s", err)
	}

	ins, err := getInstance(ctx, instancesClient, leaseID, r.GetHostReqs(), true)
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.Errorf(codes.DeadlineExceeded, "when getting instance: %s", ctx.Err())
		}
		return nil, status.Errorf(codes.Internal, "failed to get instance: %s", err)
	}

	return &api.LeaseVMResponse{
		LeaseId: leaseID,
		Vm: &api.VM{
			Id: leaseID,
			Address: &api.VMAddress{
				// Internal IP. Only one NetworkInterface should be available.
				Host: ins.GetNetworkInterfaces()[0].GetNetworkIP(),
				// Temporarily hardcode as port 22
				Port: 22,
			},
		},
	}, nil
}

// ExtendLease extends a VM lease
func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	logging.Infof(ctx, "[server:ExtendLease] Started")
	return nil, status.Errorf(codes.Unimplemented, "ExtendLease is not implemented")
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *api.ReleaseVMRequest) (*api.ReleaseVMResponse, error) {
	logging.Infof(ctx, "[server:ReleaseVM] Started")

	// Set default values for ReleaseVMRequest if needed.
	r = setDefaultReleaseVMRequest(r, s.Env)

	if err := vmleaserpb.ValidateReleaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate release request: %s", err)
	}

	err := deleteInstance(ctx, r)
	if err != nil {
		if ctx.Err() != nil {
			return nil, status.Errorf(codes.DeadlineExceeded, "when deleting instance: %s", ctx.Err())
		}
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

		logging.Debugf(ctx, "polling for instance")
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
		logging.Debugf(ctx, "getting instance without polling")
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
	networkInterfaces, err := getInstanceNetworkInterfaces(ctx, hostReqs)
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %v", err)
	}
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
			NetworkInterfaces: networkInterfaces,
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

	logging.Infof(ctx, "instance created: %s", leaseID)
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

// getDefaultParams returns the default params of a prod/dev environment
func getDefaultParams(env string) constants.DefaultLeaseParams {
	switch env {
	case "dev":
		return constants.DevDefaultParams
	case "prod":
		return constants.ProdDefaultParams
	default:
		return constants.DevDefaultParams
	}
}

// setDefaultLeaseVMRequest sets default values for VMRequirements.
func setDefaultLeaseVMRequest(r *api.LeaseVMRequest, env string) *api.LeaseVMRequest {
	defaultParams := getDefaultParams(env)
	hostReqs := r.GetHostReqs()
	if hostReqs.GetGceDiskSize() == 0 {
		hostReqs.GceDiskSize = defaultParams.DefaultDiskSize
	}
	if hostReqs.GetGceMachineType() == "" {
		hostReqs.GceMachineType = defaultParams.DefaultMachineType
	}
	if hostReqs.GetGceNetwork() == "" {
		hostReqs.GceNetwork = defaultParams.DefaultNetwork
	}
	if hostReqs.GetGceProject() == "" {
		hostReqs.GceProject = defaultParams.DefaultProject
	}
	if hostReqs.GetGceRegion() == "" {
		hostReqs.GceRegion = defaultParams.DefaultRegion
	}
	return r
}

// setDefaultReleaseVMRequest sets default values for ReleaseVMRequest.
func setDefaultReleaseVMRequest(r *api.ReleaseVMRequest, env string) *api.ReleaseVMRequest {
	defaultParams := getDefaultParams(env)
	if r.GetGceProject() == "" {
		r.GceProject = defaultParams.DefaultProject
	}
	if r.GetGceRegion() == "" {
		r.GceRegion = defaultParams.DefaultRegion
	}
	return r
}

// getInstanceNetworkInterfaces gets the NetworkInterfaces based on VM reqs.
func getInstanceNetworkInterfaces(ctx context.Context, hostReqs *api.VMRequirements) ([]*computepb.NetworkInterface, error) {
	if hostReqs.GetGceNetwork() == "" {
		return nil, errors.New("gce network cannot be empty")
	}

	netInts := []*computepb.NetworkInterface{
		{
			AccessConfigs: []*computepb.AccessConfig{
				{
					Name: proto.String("External NAT"),
				},
			},
			Network: proto.String(hostReqs.GetGceNetwork()),
		},
	}
	if hostReqs.GetGceSubnet() != "" {
		netInts[0].Subnetwork = proto.String(hostReqs.GetGceSubnet())
	}

	return netInts, nil
}

// computeExpirationTime calculates the expiration time of a VM
//
// computeExpirationTime return a future Unix time as an int64. The calculation
// is based on the specified lease duration.
func computeExpirationTime(ctx context.Context, leaseDuration *durationpb.Duration, env string) (int64, error) {
	defaultParams := getDefaultParams(env)
	expirationTime := time.Now().Unix()
	if leaseDuration == nil {
		return expirationTime + (defaultParams.DefaultLeaseDuration * 60), nil
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
				logging.Debugf(ctx, "poll: error")
				return err
			}
			if success {
				logging.Debugf(ctx, "poll: success")
				return nil
			}
		case <-ctx.Done():
			logging.Debugf(ctx, "poll: context done")
			return ctx.Err()
		}
	}
}
