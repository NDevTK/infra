// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// mockComputeInstancesClient mocks compute.NewInstancesRESTClient for testing.
type mockComputeInstancesClient struct {
	deleteFunc         func() (*compute.Operation, error)
	getFunc            func() (*computepb.Instance, error)
	insertFunc         func(r *computepb.InsertInstanceRequest) (*compute.Operation, error)
	listFunc           func() *compute.InstanceIterator
	aggregatedListFunc func() *compute.InstancesScopedListPairIterator
}

// Delete mocks the Delete instance method of the compute client.
func (m *mockComputeInstancesClient) Delete(context.Context, *computepb.DeleteInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return m.deleteFunc()
}

// Get mocks the Get instance method of the compute client.
func (m *mockComputeInstancesClient) Get(context.Context, *computepb.GetInstanceRequest, ...gax.CallOption) (*computepb.Instance, error) {
	return m.getFunc()
}

// Insert mocks the Insert instance method of the compute client.
func (m *mockComputeInstancesClient) Insert(ctx context.Context, r *computepb.InsertInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error) {
	return m.insertFunc(r)
}

// List mocks the List instance method of the compute client.
func (m *mockComputeInstancesClient) List(context.Context, *computepb.ListInstancesRequest, ...gax.CallOption) *compute.InstanceIterator {
	return m.listFunc()
}

// AggregateList mocks the AggregateList instance method of the compute client.
func (m *mockComputeInstancesClient) AggregatedList(context.Context, *computepb.AggregatedListInstancesRequest, ...gax.CallOption) *compute.InstancesScopedListPairIterator {
	return m.aggregatedListFunc()
}

func TestCheckIdempotencyKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test CheckIdempotencyKey", t, func() {
		Convey("CheckIdempotencyKey - no instances found", func() {
			client := &mockComputeInstancesClient{
				aggregatedListFunc: func() *compute.InstancesScopedListPairIterator {
					return nil
				},
			}
			in := CheckIdempotencyKey(ctx, client, "test-project", "test-key")
			So(in, ShouldBeNil)
		})
	})
}

func TestCreateInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test CreateInstance", t, func() {
		Convey("CreateInstance - error: unable to create", func() {
			client := &mockComputeInstancesClient{
				insertFunc: func(*computepb.InsertInstanceRequest) (*compute.Operation, error) {
					return nil, errors.New("failed insert")
				},
			}
			leaseReq := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceNetwork:     "test-network",
					GceDiskSize:    100,
				},
			}
			err := CreateInstance(ctx, client, "dev", "test-id", leaseReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unable to create instance")
		})
		Convey("CreateInstance - error: no operation returned", func() {
			client := &mockComputeInstancesClient{
				insertFunc: func(*computepb.InsertInstanceRequest) (*compute.Operation, error) {
					return nil, nil
				},
			}
			leaseReq := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceNetwork:     "test-network",
					GceDiskSize:    100,
				},
			}
			err := CreateInstance(ctx, client, "dev", "test-id", leaseReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no operation returned for waiting")
		})
		Convey("CreateInstance - with labels - error: no operation returned", func() {
			labels := map[string]string{"k": "v"}
			client := &mockComputeInstancesClient{
				insertFunc: func(r *computepb.InsertInstanceRequest) (*compute.Operation, error) {
					if !reflect.DeepEqual(r.InstanceResource.GetLabels(), labels) {
						return nil, fmt.Errorf("Expected labels to be %v, but is %v", labels, r.InstanceResource.GetLabels())
					}
					return nil, nil
				},
			}
			leaseReq := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceNetwork:     "test-network",
					GceDiskSize:    100,
				},
				Labels: labels,
			}
			err := CreateInstance(ctx, client, "dev", "test-id", leaseReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no operation returned for waiting")
		})
		Convey("CreateInstance - error: failed to get network interface", func() {
			client := &mockComputeInstancesClient{
				insertFunc: func(*computepb.InsertInstanceRequest) (*compute.Operation, error) {
					return nil, nil
				},
			}

			leaseReq := &api.LeaseVMRequest{
				HostReqs: &api.VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceDiskSize:    100,
				},
			}
			err := CreateInstance(ctx, client, "dev", "test-id", leaseReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to get network interface")
		})
	})
}

func TestDeleteInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test DeleteInstance", t, func() {
		Convey("DeleteInstance - error: unable to delete", func() {
			client := &mockComputeInstancesClient{
				deleteFunc: func() (*compute.Operation, error) {
					return nil, errors.New("failed delete")
				},
			}
			releaseReq := &api.ReleaseVMRequest{
				LeaseId:    "test-id",
				GceProject: "test-project",
				GceRegion:  "test-region",
			}
			err := DeleteInstance(ctx, client, releaseReq)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unable to delete instance")
		})
		Convey("DeleteInstance - success", func() {
			client := &mockComputeInstancesClient{
				deleteFunc: func() (*compute.Operation, error) {
					return &compute.Operation{}, nil
				},
			}
			releaseReq := &api.ReleaseVMRequest{
				LeaseId:    "test-id",
				GceProject: "test-project",
				GceRegion:  "test-region",
			}
			err := DeleteInstance(ctx, client, releaseReq)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test GetInstance", t, func() {
		Convey("GetInstance - happy path", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{
						Name: proto.String("test-id"),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								AccessConfigs: []*computepb.AccessConfig{
									{
										NatIP: proto.String("1.2.3.4"),
									},
								},
							},
						},
					}, nil
				},
			}
			hostReqs := &api.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := GetInstance(ctx, client, "test-id", hostReqs, false)
			So(ins, ShouldNotBeNil)
			So(ins, ShouldResembleProto, &computepb.Instance{
				Name: proto.String("test-id"),
				NetworkInterfaces: []*computepb.NetworkInterface{
					{
						AccessConfigs: []*computepb.AccessConfig{
							{
								NatIP: proto.String("1.2.3.4"),
							},
						},
					},
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("GetInstance - error: failed get", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return nil, errors.New("failed get")
				},
			}
			hostReqs := &api.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := GetInstance(ctx, client, "test-id", hostReqs, false)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed get")
		})
		Convey("GetInstance - error: no network interface", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{}, nil
				},
			}
			hostReqs := &api.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := GetInstance(ctx, client, "test-id", hostReqs, false)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance does not have a network interface")
		})
		Convey("GetInstance - error: no access config", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{
						NetworkInterfaces: []*computepb.NetworkInterface{
							{},
						},
					}, nil
				},
			}
			hostReqs := &api.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := GetInstance(ctx, client, "test-id", hostReqs, false)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance does not have an access config")
		})
		Convey("GetInstance - error: no nat ip", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								AccessConfigs: []*computepb.AccessConfig{
									{},
								},
							},
						},
					}, nil
				},
			}
			hostReqs := &api.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := GetInstance(ctx, client, "test-id", hostReqs, false)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance does not have a nat ip")
		})
	})
}

func TestListInstances(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test ListInstances", t, func() {
		Convey("listAllInstances - nil iterator returned", func() {
			client := &mockComputeInstancesClient{
				aggregatedListFunc: func() *compute.InstancesScopedListPairIterator {
					return nil
				},
			}
			listReq := &api.ListLeasesRequest{
				Parent:    "projects/test-project",
				PageSize:  5,
				PageToken: "test-token",
			}
			_, err := listAllInstances(ctx, client, "test-project", listReq)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "listAllInstances: cannot get instances")
		})
		Convey("listZoneInstances - nil iterator returned", func() {
			client := &mockComputeInstancesClient{
				listFunc: func() *compute.InstanceIterator {
					return nil
				},
			}
			listReq := &api.ListLeasesRequest{
				Parent:    "projects/test-project/zones/test-zone",
				PageSize:  5,
				PageToken: "test-token",
			}
			_, err := listZoneInstances(ctx, client, "test-project", "test-zone", listReq)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "listZoneInstances: cannot get instances")
		})
	})
}

func TestComputeExpirationTime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test computeExpirationTime", t, func() {
		Convey("Compute expiration time - no lease duration passed", func() {
			defaultExpTime := time.Now().Unix() + (600 * 60)
			res, err := computeExpirationTime(ctx, nil, "dev")
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, defaultExpTime, defaultExpTime+1)
		})
		Convey("Compute expiration time - lease duration passed", func() {
			leaseDuration, err := time.ParseDuration("20m")
			So(err, ShouldBeNil)

			expTime := time.Now().Add(leaseDuration).Unix()
			logging.Errorf(ctx, "%s", durationpb.New(leaseDuration))
			res, err := computeExpirationTime(ctx, durationpb.New(leaseDuration), "dev")
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, expTime, expTime+1)
		})
	})
}

func TestGetInstanceNetworkInterfaces(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test getInstanceNetworkInterfaces", t, func() {
		Convey("getInstanceNetworkInterfaces - happy path", func() {
			hostReqs := &api.VMRequirements{
				GceNetwork: "test-network",
				GceSubnet:  "test-subnet",
			}
			n, err := getInstanceNetworkInterfaces(ctx, hostReqs)
			So(n, ShouldResembleProto, []*computepb.NetworkInterface{
				{
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name: proto.String("External NAT"),
						},
					},
					Network:    proto.String("test-network"),
					Subnetwork: proto.String("test-subnet"),
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("getInstanceNetworkInterfaces - error: no network", func() {
			hostReqs := &api.VMRequirements{}
			n, err := getInstanceNetworkInterfaces(ctx, hostReqs)
			So(n, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "gce network cannot be empty")
		})
		Convey("getInstanceNetworkInterfaces - no subnet", func() {
			hostReqs := &api.VMRequirements{
				GceNetwork: "test-network",
			}
			n, err := getInstanceNetworkInterfaces(ctx, hostReqs)
			So(n, ShouldResembleProto, []*computepb.NetworkInterface{
				{
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name: proto.String("External NAT"),
						},
					},
					Network: proto.String("test-network"),
				},
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestPoll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test poll", t, func() {
		Convey("poll - no context deadline", func() {
			f := func(ctx context.Context) (bool, error) {
				return false, nil
			}
			interval := time.Duration(1)
			err := poll(ctx, f, interval)
			So(err, ShouldNotBeNil)
		})
		Convey("poll - quit on error", func() {
			expected := 2
			count := 1
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			f := func(ctx context.Context) (bool, error) {
				count++
				if count == 2 {
					return false, errors.New("error on 2")
				}
				return false, nil
			}
			err := poll(ctx, f, 100*time.Millisecond)
			actual := count
			So(err, ShouldNotBeNil)
			So(actual, ShouldEqual, expected)
		})
		Convey("poll - quit on success", func() {
			expected := 3
			count := 1
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			f := func(ctx context.Context) (bool, error) {
				count++
				if count == 3 {
					return true, nil
				}
				return false, nil
			}
			err := poll(ctx, f, 100*time.Millisecond)
			actual := count

			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		})
	})
}
