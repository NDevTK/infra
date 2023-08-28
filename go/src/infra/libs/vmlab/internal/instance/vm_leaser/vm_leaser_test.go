// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaser

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/config/go/test/api"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	vmlabpb "infra/libs/vmlab/api"
)

// mockVMLeaserClient mocks vmLeaserServiceClient for testing.
type mockVMLeaserClient struct {
	leaseVM    func() (*api.LeaseVMResponse, error)
	releaseVM  func() (*api.ReleaseVMResponse, error)
	listLeases func() (*api.ListLeasesResponse, error)
}

// LeaseVM mocks the LeaseVM method of the VM Leaser Client.
func (m *mockVMLeaserClient) LeaseVM(context.Context, *api.LeaseVMRequest, ...grpc.CallOption) (*api.LeaseVMResponse, error) {
	return m.leaseVM()
}

// ReleaseVM mocks the ReleaseVM method of the VM Leaser Client.
func (m *mockVMLeaserClient) ReleaseVM(context.Context, *api.ReleaseVMRequest, ...grpc.CallOption) (*api.ReleaseVMResponse, error) {
	return m.releaseVM()
}

// ListLeases mocks the ListLeases method of the VM Leaser Client.
func (m *mockVMLeaserClient) ListLeases(context.Context, *api.ListLeasesRequest, ...grpc.CallOption) (*api.ListLeasesResponse, error) {
	return m.listLeases()
}

func TestCreate(t *testing.T) {
	t.Parallel()
	Convey("Test Create", t, func() {
		Convey("Create - error: empty request", func() {
			vmLeaser, err := New()
			So(err, ShouldBeNil)

			ins, err := vmLeaser.Create(context.Background(), &vmlabpb.CreateVmInstanceRequest{})
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "no config found")
		})
	})
}

func TestLeaseVM(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Test leaseVM", t, func() {
		Convey("leaseVM - success", func() {
			client := &mockVMLeaserClient{
				leaseVM: func() (*api.LeaseVMResponse, error) {
					return &api.LeaseVMResponse{
						LeaseId: "vm-test-id",
						Vm: &api.VM{
							Id: "vm-test-id",
							Address: &api.VMAddress{
								Host: "1.2.3.4",
								Port: 99,
							},
							Type: api.VMType_VM_TYPE_DUT,
						},
						ExpirationTime: timestamppb.Now(),
					}, nil
				},
			}

			d, err := time.ParseDuration("60s")
			So(err, ShouldBeNil)

			vmLeaser := &vmLeaserInstanceApi{}
			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceImage:       "test-image",
							GceRegion:      "test-region",
							GceProject:     "test-project",
							GceMachineType: "test-machine-type",
							GceDiskSize:    100,
						},
						LeaseDuration: durationpb.New(d),
					},
				},
			}
			ins, err := vmLeaser.leaseVM(ctx, client, &vmlabpb.CreateVmInstanceRequest{
				Config: &cfg,
			})
			So(ins, ShouldResembleProto, &vmlabpb.VmInstance{
				Name: "vm-test-id",
				Ssh: &vmlabpb.AddressPort{
					Address: "1.2.3.4",
					Port:    99,
				},
				Config: &cfg,
			})
			So(err, ShouldBeNil)
		})
		Convey("leaseVM - error: failed to lease VM", func() {
			client := &mockVMLeaserClient{
				leaseVM: func() (*api.LeaseVMResponse, error) {
					return nil, errors.New("leasing error")
				},
			}

			d, err := time.ParseDuration("60s")
			So(err, ShouldBeNil)

			vmLeaser := &vmLeaserInstanceApi{}
			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceImage:       "test-image",
							GceRegion:      "test-region",
							GceProject:     "test-project",
							GceMachineType: "test-machine-type",
							GceDiskSize:    100,
						},
						LeaseDuration: durationpb.New(d),
					},
				},
			}
			ins, err := vmLeaser.leaseVM(ctx, client, &vmlabpb.CreateVmInstanceRequest{
				Config: &cfg,
			})
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "failed to lease VM: leasing error")
		})
	})
}

func TestDelete(t *testing.T) {
	vmLeaser, err := New()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	err = vmLeaser.Delete(context.Background(), &vmlabpb.VmInstance{})
	if err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Test List", t, func() {
		Convey("List - error when listing; no backend", func() {
			vmLeaser, err := New()
			So(err, ShouldBeNil)

			ins, err := vmLeaser.List(ctx, &vmlabpb.ListVmInstancesRequest{})
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "invalid argument: bad backend: want vm leaser")
		})
		Convey("List - error when listing; no gce project", func() {
			vmLeaser, err := New()
			So(err, ShouldBeNil)

			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceRegion: "test-region",
						},
					},
				},
			}

			ins, err := vmLeaser.List(ctx, &vmlabpb.ListVmInstancesRequest{
				Config: &cfg,
			})
			So(ins, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "project must be set")
		})
	})
}

func TestListLeases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Test listLeases", t, func() {
		Convey("listLeases - success", func() {
			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceRegion:  "test-region",
							GceProject: "test-project",
						},
					},
				},
			}

			client := &mockVMLeaserClient{
				listLeases: func() (*api.ListLeasesResponse, error) {
					return &api.ListLeasesResponse{
						Vms: []*api.VM{
							{
								Id: "vm-test-id",
								Address: &api.VMAddress{
									Host: "1.2.3.4",
									Port: 99,
								},
								Type:      api.VMType_VM_TYPE_DUT,
								GceRegion: "test-region",
							},
							{
								Id: "vm-test-id-2",
								Address: &api.VMAddress{
									Host: "2.3.4.5",
									Port: 99,
								},
								Type:      api.VMType_VM_TYPE_DUT,
								GceRegion: "test-region",
							},
						},
					}, nil
				},
			}

			vmLeaser := &vmLeaserInstanceApi{}
			ins, err := vmLeaser.listLeases(ctx, client, &vmlabpb.ListVmInstancesRequest{
				Config: &cfg,
			})
			So(ins, ShouldResembleProto, []*vmlabpb.VmInstance{
				{
					Name: "vm-test-id",
					Ssh: &vmlabpb.AddressPort{
						Address: "1.2.3.4",
						Port:    99,
					},
					Config:    &cfg,
					GceRegion: "test-region",
				},
				{
					Name: "vm-test-id-2",
					Ssh: &vmlabpb.AddressPort{
						Address: "2.3.4.5",
						Port:    99,
					},
					Config:    &cfg,
					GceRegion: "test-region",
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("listLeases - no results", func() {
			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceRegion:  "test-region",
							GceProject: "test-project",
						},
					},
				},
			}

			client := &mockVMLeaserClient{
				listLeases: func() (*api.ListLeasesResponse, error) {
					return &api.ListLeasesResponse{
						Vms: []*api.VM{},
					}, nil
				},
			}

			vmLeaser := &vmLeaserInstanceApi{}
			ins, err := vmLeaser.listLeases(ctx, client, &vmlabpb.ListVmInstancesRequest{
				Config: &cfg,
			})
			So(ins, ShouldResembleProto, []*vmlabpb.VmInstance{})
			So(err, ShouldBeNil)
		})
		Convey("listLeases - error when listing", func() {
			cfg := vmlabpb.Config{
				Backend: &vmlabpb.Config_VmLeaserBackend_{
					VmLeaserBackend: &vmlabpb.Config_VmLeaserBackend{
						Env: vmlabpb.Config_VmLeaserBackend_ENV_LOCAL,
						VmRequirements: &api.VMRequirements{
							GceRegion:  "test-region",
							GceProject: "test-project",
						},
					},
				},
			}

			client := &mockVMLeaserClient{
				listLeases: func() (*api.ListLeasesResponse, error) {
					return nil, fmt.Errorf("cannot list VMs")
				},
			}

			vmLeaser := &vmLeaserInstanceApi{}
			ins, err := vmLeaser.listLeases(ctx, client, &vmlabpb.ListVmInstancesRequest{
				Config: &cfg,
			})
			So(err, ShouldErrLike, "failed to list VMs")
			So(ins, ShouldBeNil)
		})
	})
}
