// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"errors"
	"testing"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/logging"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	pb "infra/vm_leaser/api/v1"
)

// mockComputeInstancesClient mocks compute.NewInstancesRESTClient for testing.
type mockComputeInstancesClient struct {
	getFunc    func() (*computepb.Instance, error)
	insertFunc func() (*compute.Operation, error)
}

// Get mocks the Get instance method of the compute client.
func (m *mockComputeInstancesClient) Get(context.Context, *computepb.GetInstanceRequest, ...gax.CallOption) (*computepb.Instance, error) {
	return m.getFunc()
}

// Insert mocks the Insert instance method of the compute client.
func (m *mockComputeInstancesClient) Insert(context.Context, *computepb.InsertInstanceRequest, ...gax.CallOption) (*compute.Operation, error) {
	return m.insertFunc()
}

func TestComputeExpirationTime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test computeExpirationTime", t, func() {
		Convey("Compute expiration time - no lease duration passed", func() {
			defaultExpTime := time.Now().Unix() + (DefaultLeaseDuration * 60)
			res, err := computeExpirationTime(ctx, nil)
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, defaultExpTime, defaultExpTime+1)
		})
		Convey("Compute expiration time - lease duration passed", func() {
			leaseDuration, err := time.ParseDuration("20m")
			So(err, ShouldBeNil)

			expTime := time.Now().Add(leaseDuration).Unix()
			logging.Errorf(ctx, "%s", durationpb.New(leaseDuration))
			res, err := computeExpirationTime(ctx, durationpb.New(leaseDuration))
			So(err, ShouldBeNil)
			So(res, ShouldBeBetweenOrEqual, expTime, expTime+1)
		})
	})
}

func TestCreateInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test createInstance", t, func() {
		Convey("createInstance - error: unable to create", func() {
			client := &mockComputeInstancesClient{
				insertFunc: func() (*compute.Operation, error) {
					return nil, errors.New("failed insert")
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := createInstance(ctx, client, "test-id", 100, hostReqs)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unable to create instance")
		})
		Convey("createInstance - error: no operation returned", func() {
			client := &mockComputeInstancesClient{
				insertFunc: func() (*compute.Operation, error) {
					return nil, nil
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := createInstance(ctx, client, "test-id", 100, hostReqs)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no operation returned for waiting")
		})
	})
}

func TestGetInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	Convey("Test getInstance", t, func() {
		Convey("getInstance - happy path", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{
						Name: proto.String("test-id"),
						NetworkInterfaces: []*computepb.NetworkInterface{
							{
								NetworkIP: proto.String("1.2.3.4"),
							},
						},
					}, nil
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := getInstance(ctx, client, "test-id", hostReqs)
			So(ins, ShouldNotBeNil)
			So(ins, ShouldResembleProto, &computepb.Instance{
				Name: proto.String("test-id"),
				NetworkInterfaces: []*computepb.NetworkInterface{
					{
						NetworkIP: proto.String("1.2.3.4"),
					},
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("getInstance - error: instance not found", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return nil, errors.New("failed get")
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := getInstance(ctx, client, "test-id", hostReqs)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance not found")
		})
		Convey("getInstance - error: no network interface", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{}, nil
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := getInstance(ctx, client, "test-id", hostReqs)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance does not have a network interface")
		})
		Convey("getInstance - error: no network ip", func() {
			client := &mockComputeInstancesClient{
				getFunc: func() (*computepb.Instance, error) {
					return &computepb.Instance{
						NetworkInterfaces: []*computepb.NetworkInterface{
							{},
						},
					}, nil
				},
			}
			hostReqs := &pb.VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			ins, err := getInstance(ctx, client, "test-id", hostReqs)
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "instance does not have a network IP")
		})
	})
}
