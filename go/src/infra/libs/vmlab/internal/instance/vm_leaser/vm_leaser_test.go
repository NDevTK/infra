// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaser

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	vmlabpb "infra/libs/vmlab/api"
	vmleaserpb "infra/vm_leaser/api/v1"
)

// mockVMLeaserClient mocks vmLeaserServiceClient for testing.
type mockVMLeaserClient struct {
	leaseVM func() (*vmleaserpb.LeaseVMResponse, error)
}

// LeaseVM mocks the LeaseVM method of the VM Leaser Client.
func (m *mockVMLeaserClient) LeaseVM(context.Context, *vmleaserpb.LeaseVMRequest, ...grpc.CallOption) (*vmleaserpb.LeaseVMResponse, error) {
	return m.leaseVM()
}

func TestCreate(t *testing.T) {
	t.Parallel()
	Convey("Test Create", t, func() {
		Convey("Create - error: empty request", func() {
			vmLeaser, err := New()
			So(err, ShouldBeNil)

			ins, err := vmLeaser.Create(&vmlabpb.CreateVmInstanceRequest{})
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
				leaseVM: func() (*vmleaserpb.LeaseVMResponse, error) {
					return &vmleaserpb.LeaseVMResponse{
						LeaseId: "vm-test-id",
						Vm: &vmleaserpb.VM{
							Id: "vm-test-id",
							Address: &vmleaserpb.VMAddress{
								Host: "1.2.3.4",
								Port: 99,
							},
							Type: vmleaserpb.VMType_VM_TYPE_DUT,
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
						VmRequirements: &vmleaserpb.VMRequirements{
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
				leaseVM: func() (*vmleaserpb.LeaseVMResponse, error) {
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
						VmRequirements: &vmleaserpb.VMRequirements{
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

	err = vmLeaser.Delete(&vmlabpb.VmInstance{})
	if err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestList(t *testing.T) {
	vmLeaser, err := New()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	_, err = vmLeaser.List(&vmlabpb.ListVmInstancesRequest{})
	if err == nil {
		t.Errorf("error should not be nil")
	}
}
