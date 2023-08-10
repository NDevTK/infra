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

	"infra/vm_leaser/internal/constants"
	"infra/vm_leaser/internal/validation"
	"infra/vm_leaser/internal/zone_selector"
)

// computeInstancesClient interfaces the GCE instance client API.
type computeInstancesClient interface {
	Delete(ctx context.Context, r *computepb.DeleteInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error)
	Get(ctx context.Context, r *computepb.GetInstanceRequest, opts ...gax.CallOption) (*computepb.Instance, error)
	Insert(ctx context.Context, r *computepb.InsertInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error)
	List(ctx context.Context, r *computepb.ListInstancesRequest, opts ...gax.CallOption) *compute.InstanceIterator
}

// Prove that Server implements pb.VMLeaserServiceServer by instantiating a Server
var _ api.VMLeaserServiceServer = (*Server)(nil)

// Server is a struct implements the pb.VMLeaserServiceServer
type Server struct {
	api.UnimplementedVMLeaserServiceServer
	Env string

	// retry defaults
	initialRetryBackoff time.Duration
	maxRetries          int
}

// NewServer returns a new Server
func NewServer(env string) *Server {
	return &Server{
		Env:                 env,
		initialRetryBackoff: 1 * time.Second,
		maxRetries:          3,
	}
}

// LeaseVM leases a VM defined by LeaseVMRequest
func (s *Server) LeaseVM(ctx context.Context, r *api.LeaseVMRequest) (*api.LeaseVMResponse, error) {
	logging.Infof(ctx, "LeaseVM start")

	// Set defaults for LeaseVMRequest if needed.
	r, err := setDefaultLeaseVMRequest(ctx, r, s.Env)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to set lease request: %s", err)
	}

	if err := validation.ValidateLeaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate lease request: %s", err)
	}

	// Appending "vm-" to satisfy GCE regex
	leaseID := fmt.Sprintf("vm-%s", uuid.New().String())
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create NewInstancesRESTClient: %s", err)
	}
	defer instancesClient.Close()

	retry := 0
	for {
		err = createInstance(ctx, instancesClient, s.Env, leaseID, r)
		if err == nil {
			break
		}

		logging.Errorf(ctx, "retry #%d - error when creating instance: %s", retry, err)
		if retry >= s.maxRetries {
			return nil, status.Errorf(codes.Internal, "failed to create instance after %d retries: %s", s.maxRetries, err)
		}

		if ctx.Err() != nil {
			return nil, status.Errorf(codes.DeadlineExceeded, "context error when creating instance: %s", ctx.Err())
		}

		time.Sleep(s.initialRetryBackoff * (1 << retry))
		retry++
		logging.Debugf(ctx, "LeaseVM: retrying %d time createInstance", retry)
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
			GceRegion: r.GetHostReqs().GetGceRegion(),
		},
	}, nil
}

// ExtendLease extends a VM lease
func (s *Server) ExtendLease(ctx context.Context, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	logging.Infof(ctx, "ExtendLease start")
	return nil, status.Errorf(codes.Unimplemented, "ExtendLease is not implemented")
}

// ReleaseVM releases a VM lease
func (s *Server) ReleaseVM(ctx context.Context, r *api.ReleaseVMRequest) (*api.ReleaseVMResponse, error) {
	logging.Infof(ctx, "ReleaseVM start")

	// Set default values for ReleaseVMRequest if needed.
	r = setDefaultReleaseVMRequest(ctx, r, s.Env)

	if err := validation.ValidateReleaseVMRequest(r); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate release request: %s", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create NewInstancesRESTClient: %s", err)
	}
	defer instancesClient.Close()

	retry := 0
	for {
		err := deleteInstance(ctx, instancesClient, r)
		if err == nil {
			break
		}

		logging.Errorf(ctx, "retry #%d - error when deleting instance: %s", retry, err)
		if retry >= s.maxRetries {
			return nil, status.Errorf(codes.Internal, "failed to delete instance after %d retries: %s", s.maxRetries, err)
		}

		if ctx.Err() != nil {
			return nil, status.Errorf(codes.DeadlineExceeded, "context error when deleting instance: %s", ctx.Err())
		}

		time.Sleep(s.initialRetryBackoff * (1 << retry))
		retry++
	}

	return &api.ReleaseVMResponse{
		LeaseId: r.GetLeaseId(),
	}, nil
}

func (s *Server) ListLeases(ctx context.Context, r *api.ListLeasesRequest) (*api.ListLeasesResponse, error) {
	logging.Infof(ctx, "ListLeases start")
	return nil, status.Errorf(codes.Unimplemented, "ListLeases is not implemented")
}

// getInstance gets a GCE instance based on lease id and GCE configs.
//
// getInstance returns a GCE instance with valid network interface and network
// IP. If no network is available, it does not return the instance.
func getInstance(parentCtx context.Context, client computeInstancesClient, leaseID string, hostReqs *api.VMRequirements, shouldPoll bool) (*computepb.Instance, error) {
	// Implement a 30 second deadline for polling for the instance
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()

	getReq := &computepb.GetInstanceRequest{
		Instance: leaseID,
		Project:  hostReqs.GetGceProject(),
		Zone:     hostReqs.GetGceRegion(),
	}

	var ins *computepb.Instance
	var err error
	if shouldPoll {
		logging.Debugf(ctx, "getInstance: polling for instance")
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
		logging.Debugf(ctx, "getInstance: getting instance without polling")
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
func createInstance(parentCtx context.Context, client computeInstancesClient, env, leaseID string, r *api.LeaseVMRequest) error {
	// Implement a 180 second deadline for creating the instance
	ctx, cancel := context.WithTimeout(parentCtx, 180*time.Second)
	defer cancel()

	hostReqs := r.GetHostReqs()
	zone := hostReqs.GetGceRegion()
	networkInterfaces, err := getInstanceNetworkInterfaces(ctx, hostReqs)
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %v", err)
	}
	metadata, err := getMetadata(ctx, env, r)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %v", err)
	}

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
			MachineType:       proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", zone, hostReqs.GetGceMachineType())),
			Metadata:          metadata,
			NetworkInterfaces: networkInterfaces,
		},
	}

	logging.Debugf(ctx, "createInstance: InsertInstanceRequest payload: %v", req)
	op, err := client.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %v", err)
	}
	if op == nil {
		return errors.New("no operation returned for waiting")
	}

	logging.Debugf(ctx, "createInstance: waiting for operation completion")
	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	logging.Infof(ctx, "createInstance: instance scheduled for creation: %s", leaseID)
	return nil
}

// deleteInstance sends an instance deletion request to the Compute Engine API.
func deleteInstance(ctx context.Context, c computeInstancesClient, r *api.ReleaseVMRequest) error {
	req := &computepb.DeleteInstanceRequest{
		Instance: r.GetLeaseId(),
		Project:  r.GetGceProject(),
		Zone:     r.GetGceRegion(),
	}
	logging.Debugf(ctx, "deleteInstance: DeleteInstanceRequest payload: %v", req)

	// We omit checking the returned operation or calling Wait so that this call
	// becomes non-blocking. This saves callers time and lets the clean up cron
	// job take care of any stale instances instead. See b/287524018.
	_, err := c.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete instance: %v", err)
	}

	logging.Infof(ctx, "deleteInstance: instance delete request received by GCP")
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
func setDefaultLeaseVMRequest(ctx context.Context, r *api.LeaseVMRequest, env string) (*api.LeaseVMRequest, error) {
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
		hostReqs.GceRegion = zone_selector.SelectZone(ctx, r, time.Now().UnixNano())
	}
	if hostReqs.GetSubnetModeNetworkEnabled() {
		var err error
		hostReqs.GceSubnet, err = zone_selector.GetZoneSubnet(ctx, hostReqs.GetGceRegion())
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

// setDefaultReleaseVMRequest sets default values for ReleaseVMRequest.
func setDefaultReleaseVMRequest(ctx context.Context, r *api.ReleaseVMRequest, env string) *api.ReleaseVMRequest {
	defaultParams := getDefaultParams(env)
	if r.GetGceProject() == "" {
		r.GceProject = defaultParams.DefaultProject
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

// getMetadata gets the Metadata based on VM reqs.
func getMetadata(ctx context.Context, env string, r *api.LeaseVMRequest) (*computepb.Metadata, error) {
	expirationTime, err := computeExpirationTime(ctx, r.GetLeaseDuration(), env)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to compute expiration time: %s", err)
	}

	metadataItems := []*computepb.Items{
		{
			Key:   proto.String("expiration_time"),
			Value: proto.String(strconv.FormatInt(expirationTime, 10)),
		},
	}

	if r.GetIdempotencyKey() != "" {
		metadataItems = append(metadataItems, &computepb.Items{
			Key:   proto.String("idempotency_key"),
			Value: proto.String(r.GetIdempotencyKey()),
		})
	}

	return &computepb.Metadata{
		Items: metadataItems,
	}, nil
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
