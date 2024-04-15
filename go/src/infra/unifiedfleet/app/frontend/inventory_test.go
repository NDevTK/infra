// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	code "google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	"infra/unifiedfleet/app/model/configuration"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"
	"infra/unifiedfleet/app/model/state"
	"infra/unifiedfleet/app/util"
)

func mockMachineLSE(id string) *ufspb.MachineLSE {
	return &ufspb.MachineLSE{
		Name:     util.AddPrefix(util.MachineLSECollection, id),
		Hostname: id,
	}
}

func mockDutMachineLSE(id string) *ufspb.MachineLSE {
	dut := &chromeosLab.DeviceUnderTest{
		Hostname: id,
	}
	device := &ufspb.ChromeOSDeviceLSE_Dut{
		Dut: dut,
	}
	deviceLse := &ufspb.ChromeOSDeviceLSE{
		Device: device,
	}
	chromeosLse := &ufspb.ChromeOSMachineLSE_DeviceLse{
		DeviceLse: deviceLse,
	}
	chromeOSMachineLse := &ufspb.ChromeOSMachineLSE{
		ChromeosLse: chromeosLse,
	}
	lse := &ufspb.MachineLSE_ChromeosMachineLse{
		ChromeosMachineLse: chromeOSMachineLse,
	}
	return &ufspb.MachineLSE{
		Name:     util.AddPrefix(util.MachineLSECollection, id),
		Hostname: id,
		Lse:      lse,
	}
}

func mockRackLSE(id string) *ufspb.RackLSE {
	return &ufspb.RackLSE{
		Name: util.AddPrefix(util.RackLSECollection, id),
	}
}

func mockSchedulingUnit(name string) *ufspb.SchedulingUnit {
	return &ufspb.SchedulingUnit{
		Name: util.AddPrefix(util.SchedulingUnitCollection, name),
	}
}

func mockAttachedDeviceMachineLSE(name string) *ufspb.MachineLSE {
	return &ufspb.MachineLSE{
		Name:     util.AddPrefix(util.MachineLSECollection, name),
		Hostname: name,
		Lse: &ufspb.MachineLSE_AttachedDeviceLse{
			AttachedDeviceLse: &ufspb.AttachedDeviceLSE{
				OsVersion: &ufspb.OSVersion{
					Value: "test",
				},
			},
		},
	}
}

func TestUpdateNetworkOpt(t *testing.T) {
	input := &ufsAPI.NetworkOption{
		Vlan: "vlan1",
		Ip:   "",
		Nic:  "eth0",
	}
	Convey("No vlan & ip, empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("", "", nil)
		So(nwOpt, ShouldBeNil)
	})
	Convey("No vlan & ip, non-empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("", "", input)
		So(nwOpt, ShouldResembleProto, input)
	})
	Convey("Have vlan, no ip, empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("vlan1", "", nil)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Vlan: "vlan1",
		})
	})
	Convey("Have vlan, no ip, non-empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("vlan2", "", input)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Vlan: "vlan2",
			Nic:  "eth0",
		})
	})
	Convey("no vlan, have ip, empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("", "0.0.0.0", nil)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Ip: "0.0.0.0",
		})
	})
	Convey("no vlan, have ip, non-empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("", "0.0.0.0", input)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Ip:  "0.0.0.0",
			Nic: "eth0",
		})
	})
	Convey("have vlan, have ip, empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("vlan1", "0.0.0.0", nil)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Ip:   "0.0.0.0",
			Vlan: "vlan1",
		})
	})
	Convey("have vlan, have ip, non-empty nwOpt", t, func() {
		nwOpt := updateNetworkOpt("vlan2", "0.0.0.0", input)
		So(nwOpt, ShouldResembleProto, &ufsAPI.NetworkOption{
			Ip:   "0.0.0.0",
			Vlan: "vlan2",
			Nic:  "eth0",
		})
	})
}

func TestCreateMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("CreateMachineLSEs", t, func() {
		Convey("Create new machineLSE with machineLSE_id", func() {
			machine := &ufspb.Machine{
				Name: "machine-1",
			}
			_, err := registration.CreateMachine(ctx, machine)
			So(err, ShouldBeNil)

			mlsePrototype := &ufspb.MachineLSEPrototype{
				Name: "browser:no-vm",
			}
			_, err = configuration.CreateMachineLSEPrototype(ctx, mlsePrototype)
			So(err, ShouldBeNil)

			machineLSE1 := mockMachineLSE("machineLSE-1")
			machineLSE1.MachineLsePrototype = "browser:no-vm"
			machineLSE1.Lse = &ufspb.MachineLSE_ChromeBrowserMachineLse{
				ChromeBrowserMachineLse: &ufspb.ChromeBrowserMachineLSE{},
			}
			machineLSE1.Machines = []string{"machine-1"}
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE:   machineLSE1,
				MachineLSEId: "machinelse-1",
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
		})

		Convey("Create new machineLSE - Invalid input nil", func() {
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE: nil,
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Create new machineLSE - Invalid input empty ID", func() {
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE:   mockMachineLSE("machineLSE-3"),
				MachineLSEId: "",
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyID)
		})

		Convey("Create new machineLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE:   mockMachineLSE("machineLSE-4"),
				MachineLSEId: "a.b)7&",
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidHostname)
		})

		Convey("Create new machineLSE - Invalid input nil machines", func() {
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE:   mockMachineLSE("machineLSE-4"),
				MachineLSEId: "machineLSE-4",
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyMachineName)
		})

		Convey("Create new machineLSE - Invalid input empty machines", func() {
			mlse := mockMachineLSE("machineLSE-4")
			mlse.Machines = []string{""}
			req := &ufsAPI.CreateMachineLSERequest{
				MachineLSE:   mlse,
				MachineLSEId: "machineLSE-4",
			}
			resp, err := tf.Fleet.CreateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyMachineName)
		})
	})
}

func TestUpdateMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("UpdateMachineLSEs", t, func() {
		Convey("Update existing machineLSEs", func() {
			machine := &ufspb.Machine{
				Name: "machine-0",
			}
			_, err := registration.CreateMachine(ctx, machine)
			So(err, ShouldBeNil)

			_, err = inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name:     "machinelse-1",
				Machines: []string{"machine-0"},
			})
			So(err, ShouldBeNil)

			machineLSE := mockMachineLSE("machineLSE-1")
			machineLSE.Machines = []string{"machine-0"}
			req := &ufsAPI.UpdateMachineLSERequest{
				MachineLSE: machineLSE,
			}
			resp, err := tf.Fleet.UpdateMachineLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE)
		})

		Convey("Update existing machineLSEs with states", func() {
			machine := &ufspb.Machine{
				Name: "machine-1",
			}
			_, err := registration.CreateMachine(ctx, machine)
			So(err, ShouldBeNil)

			_, err = inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name:     "machinelse-state",
				Machines: []string{"machine-1"},
			})
			So(err, ShouldBeNil)

			machineLSE := mockMachineLSE("machineLSE-state")
			machineLSE.Machines = []string{"machine-1"}
			machineLSE.ResourceState = ufspb.State_STATE_DEPLOYED_TESTING
			req := &ufsAPI.UpdateMachineLSERequest{
				MachineLSE: machineLSE,
			}
			resp, err := tf.Fleet.UpdateMachineLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE)
			s, err := state.GetStateRecord(ctx, "hosts/machinelse-state")
			So(err, ShouldBeNil)
			So(s.GetState(), ShouldEqual, ufspb.State_STATE_DEPLOYED_TESTING)
		})

		Convey("Update machineLSE - Invalid input nil", func() {
			req := &ufsAPI.UpdateMachineLSERequest{
				MachineLSE: nil,
			}
			resp, err := tf.Fleet.UpdateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Update machineLSE - Invalid input empty name", func() {
			machineLSE4 := mockMachineLSE("")
			machineLSE4.Name = ""
			req := &ufsAPI.UpdateMachineLSERequest{
				MachineLSE: machineLSE4,
			}
			resp, err := tf.Fleet.UpdateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Update machineLSE - Invalid input invalid characters", func() {
			machineLSE5 := mockMachineLSE("a.b)7&")
			req := &ufsAPI.UpdateMachineLSERequest{
				MachineLSE: machineLSE5,
			}
			resp, err := tf.Fleet.UpdateMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestGetMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("GetMachineLSE", t, func() {
		Convey("Get machineLSE by existing ID", func() {
			machineLSE1, err := inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name: "machinelse-1",
			})
			So(err, ShouldBeNil)
			machineLSE1.Name = util.AddPrefix(util.MachineLSECollection, machineLSE1.Name)

			req := &ufsAPI.GetMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "machinelse-1"),
			}
			resp, err := tf.Fleet.GetMachineLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
		})
		Convey("Get machineLSE by non-existing ID", func() {
			req := &ufsAPI.GetMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "machineLSE-2"),
			}
			resp, err := tf.Fleet.GetMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get machineLSE - Invalid input empty name", func() {
			req := &ufsAPI.GetMachineLSERequest{
				Name: "",
			}
			resp, err := tf.Fleet.GetMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Get machineLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.GetMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.GetMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestCreateVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("CreateVM", t, func() {
		registration.CreateMachine(ctx, &ufspb.Machine{
			Name: "inventory-create-machine",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS3,
			},
		})
		inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
			Name:     "inventory-create-host",
			Zone:     ufspb.Zone_ZONE_CHROMEOS3.String(),
			Machines: []string{"inventory-create-machine"},
		})
		Convey("Create new VM - happy path", func() {
			vm1 := &ufspb.VM{
				Name:         "inventory-create-vm1",
				MachineLseId: "inventory-create-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
			})
			So(err, ShouldBeNil)

			resp, err := tf.Fleet.GetVM(ctx, &ufsAPI.GetVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-create-vm1"),
			})
			So(err, ShouldBeNil)
			So(resp.GetName(), ShouldEqual, "vms/inventory-create-vm1")
			So(resp.GetZone(), ShouldEqual, ufspb.Zone_ZONE_CHROMEOS3.String())
			So(resp.GetMachineLseId(), ShouldEqual, "inventory-create-host")
			So(resp.GetResourceState(), ShouldEqual, ufspb.State_STATE_REGISTERED)

			s, err := state.GetStateRecord(ctx, "vms/inventory-create-vm1")
			So(err, ShouldBeNil)
			So(s.GetState(), ShouldEqual, ufspb.State_STATE_REGISTERED)
		})

		Convey("Create new VM - missing host ", func() {
			vm1 := &ufspb.VM{
				Name: "missing",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyHostName)
		})

		Convey("Create vm - Invalid mac", func() {
			resp, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: &ufspb.VM{
					Name:         "createvm-0",
					MacAddress:   "123",
					MachineLseId: "inventory-create-host",
				},
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidMac)
		})

		Convey("Create new VM - assign ip", func() {
			setupTestVlan(ctx)
			vm1 := &ufspb.VM{
				Name:         "inventory-create-vm2",
				MachineLseId: "inventory-create-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
				NetworkOption: &ufsAPI.NetworkOption{
					Vlan: "vlan-1",
				},
			})
			So(err, ShouldBeNil)

			resp, err := tf.Fleet.GetVM(ctx, &ufsAPI.GetVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-create-vm2"),
			})
			So(err, ShouldBeNil)
			So(resp.GetVlan(), ShouldEqual, "vlan-1")
			dhcp, err := configuration.GetDHCPConfig(ctx, "inventory-create-vm2")
			So(err, ShouldBeNil)
			ip, err := configuration.QueryIPByPropertyName(ctx, map[string]string{"ipv4_str": dhcp.GetIp()})
			So(err, ShouldBeNil)
			So(ip, ShouldHaveLength, 1)
			So(ip[0].GetOccupied(), ShouldBeTrue)
		})
	})
}

func TestUpdateVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("UpdateVM", t, func() {
		registration.CreateMachine(ctx, &ufspb.Machine{
			Name: "inventory-update-machine",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS3,
			},
		})
		inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
			Name:     "inventory-update-host",
			Zone:     ufspb.Zone_ZONE_CHROMEOS3.String(),
			Machines: []string{"inventory-update-machine"},
		})
		Convey("Update existing VM", func() {
			vm1 := &ufspb.VM{
				Name:         "inventory-update-vm1",
				MachineLseId: "inventory-update-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
			})
			So(err, ShouldBeNil)
			vm, err := tf.Fleet.GetVM(ctx, &ufsAPI.GetVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-update-vm1"),
			})
			So(err, ShouldBeNil)
			vm.UpdateTime = nil

			req := &ufsAPI.UpdateVMRequest{
				Vm: vm1,
			}
			resp, err := tf.Fleet.UpdateVM(tf.C, req)
			So(err, ShouldBeNil)
			resp.UpdateTime = nil
			So(resp, ShouldResembleProto, vm)
		})

		Convey("Update existing VMs with states", func() {
			vm1 := &ufspb.VM{
				Name:         "inventory-update-vm2",
				MachineLseId: "inventory-update-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
			})
			So(err, ShouldBeNil)

			vm1.ResourceState = ufspb.State_STATE_DEPLOYED_TESTING
			req := &ufsAPI.UpdateVMRequest{
				Vm: vm1,
			}
			resp, err := tf.Fleet.UpdateVM(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.GetResourceState(), ShouldEqual, ufspb.State_STATE_DEPLOYED_TESTING)
			s, err := state.GetStateRecord(ctx, "vms/inventory-update-vm2")
			So(err, ShouldBeNil)
			So(s.GetState(), ShouldEqual, ufspb.State_STATE_DEPLOYED_TESTING)
		})

		Convey("Update VM - Invalid input nil", func() {
			req := &ufsAPI.UpdateVMRequest{
				Vm: nil,
			}
			resp, err := tf.Fleet.UpdateVM(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Update vm - Invalid mac", func() {
			resp, err := tf.Fleet.UpdateVM(ctx, &ufsAPI.UpdateVMRequest{
				Vm: &ufspb.VM{
					Name:       "updatevm-0",
					MacAddress: "123",
				},
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidMac)
		})

		Convey("Update VM - Invalid input empty name", func() {
			req := &ufsAPI.UpdateVMRequest{
				Vm: &ufspb.VM{
					Name:         "",
					MachineLseId: "inventory-update-host",
				},
			}
			resp, err := tf.Fleet.UpdateVM(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Update VM - Invalid input invalid characters", func() {
			req := &ufsAPI.UpdateVMRequest{
				Vm: &ufspb.VM{
					Name:         util.AddPrefix(util.VMCollection, "a.b)7&"),
					MachineLseId: "inventory-update-host",
				},
			}
			resp, err := tf.Fleet.UpdateVM(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestDeleteVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("DeleteVM", t, func() {
		registration.CreateMachine(ctx, &ufspb.Machine{
			Name: "inventory-delete-machine",
			Location: &ufspb.Location{
				Zone: ufspb.Zone_ZONE_CHROMEOS3,
			},
		})
		inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
			Name:     "inventory-delete-host",
			Zone:     ufspb.Zone_ZONE_CHROMEOS3.String(),
			Machines: []string{"inventory-delete-machine"},
		})
		Convey("Delete vm by existing ID", func() {
			vm1 := &ufspb.VM{
				Name:         "inventory-delete-vm1",
				MachineLseId: "inventory-delete-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
			})
			So(err, ShouldBeNil)

			req := &ufsAPI.DeleteVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-delete-vm1"),
			}
			_, err = tf.Fleet.DeleteVM(tf.C, req)
			So(err, ShouldBeNil)

			res, err := inventory.GetVM(tf.C, "inventory-delete-vm1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
			s, err := state.GetStateRecord(ctx, "vms/inventory-delete-vm1")
			So(s, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete vm by existing ID with assigned ip", func() {
			setupTestVlan(ctx)
			vm1 := &ufspb.VM{
				Name:         "inventory-delete-vm2",
				MachineLseId: "inventory-delete-host",
			}
			_, err := tf.Fleet.CreateVM(ctx, &ufsAPI.CreateVMRequest{
				Vm: vm1,
				NetworkOption: &ufsAPI.NetworkOption{
					Ip: "192.168.40.18",
				},
			})
			So(err, ShouldBeNil)

			req := &ufsAPI.DeleteVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-delete-vm2"),
			}
			_, err = tf.Fleet.DeleteVM(tf.C, req)
			So(err, ShouldBeNil)

			res, err := inventory.GetVM(tf.C, "inventory-delete-vm2")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
			dhcp, err := configuration.GetDHCPConfig(ctx, "inventory-delete-vm2")
			So(dhcp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
			ips, err := configuration.QueryIPByPropertyName(ctx, map[string]string{"ipv4_str": "192.168.40.18"})
			So(err, ShouldBeNil)
			So(ips, ShouldHaveLength, 1)
			So(ips[0].GetOccupied(), ShouldBeFalse)
		})
		Convey("Delete vm by non-existing ID", func() {
			req := &ufsAPI.DeleteVMRequest{
				Name: util.AddPrefix(util.VMCollection, "inventory-delete-vm3"),
			}
			_, err := tf.Fleet.DeleteVM(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete vm - Invalid input empty name", func() {
			req := &ufsAPI.DeleteVMRequest{
				Name: "",
			}
			resp, err := tf.Fleet.DeleteVM(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Delete vm - Invalid input invalid characters", func() {
			req := &ufsAPI.DeleteVMRequest{
				Name: util.AddPrefix(util.VMCollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.DeleteVM(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestListVMs(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	vms := []*ufspb.VM{
		{
			Name: "vm-list-1",
			OsVersion: &ufspb.OSVersion{
				Value: "os-1",
			},
			Vlan:          "vlan-1",
			ResourceState: ufspb.State_STATE_SERVING,
		},
		{
			Name: "vm-list-2",
			OsVersion: &ufspb.OSVersion{
				Value: "os-1",
			},
			Vlan:          "vlan-2",
			ResourceState: ufspb.State_STATE_SERVING,
		},
		{
			Name: "vm-list-3",
			OsVersion: &ufspb.OSVersion{
				Value: "os-2",
			},
			Vlan:          "vlan-1",
			ResourceState: ufspb.State_STATE_SERVING,
		},
		{
			Name: "vm-list-4",
			OsVersion: &ufspb.OSVersion{
				Value: "os-2",
			},
			Zone:          ufspb.Zone_ZONE_CHROMEOS3.String(),
			Vlan:          "vlan-2",
			ResourceState: ufspb.State_STATE_DEPLOYED_TESTING,
		},
	}
	Convey("ListVMs", t, func() {
		_, err := inventory.BatchUpdateVMs(ctx, vms)
		So(err, ShouldBeNil)
		Convey("ListVMs - page_size negative - error", func() {
			resp, err := tf.Fleet.ListVMs(tf.C, &ufsAPI.ListVMsRequest{
				PageSize: -5,
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidPageSize)
		})

		Convey("ListVMs - invalid filter - error", func() {
			resp, err := tf.Fleet.ListVMs(tf.C, &ufsAPI.ListVMsRequest{
				Filter: "os=os-1 | state=serving",
			})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidFilterFormat)
		})

		Convey("List VMs - happy path", func() {
			resp, err := tf.Fleet.ListVMs(tf.C, &ufsAPI.ListVMsRequest{
				Filter:   "os=os-1 & state=serving",
				PageSize: 5,
			})
			So(err, ShouldBeNil)
			So(resp.GetVms(), ShouldHaveLength, 2)
			So(ufsAPI.ParseResources(resp.GetVms(), "Name"), ShouldResemble, []string{"vm-list-1", "vm-list-2"})
		})
	})
}

func TestListMachineLSEs(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	machineLSEs := make([]*ufspb.MachineLSE, 0, 4)
	for i := 0; i < 4; i++ {
		resp, _ := inventory.CreateMachineLSE(tf.C, &ufspb.MachineLSE{
			Name:     fmt.Sprintf("machineLSEFilter-%d", i),
			Machines: []string{"mac-1"},
		})
		resp.Name = util.AddPrefix(util.MachineLSECollection, resp.Name)
		machineLSEs = append(machineLSEs, resp)
	}
	Convey("ListMachineLSEs", t, func() {
		Convey("ListMachineLSEs - page_size negative - error", func() {
			req := &ufsAPI.ListMachineLSEsRequest{
				PageSize: -5,
			}
			resp, err := tf.Fleet.ListMachineLSEs(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidPageSize)
		})

		Convey("ListMachineLSEs - Full listing - happy path", func() {
			req := &ufsAPI.ListMachineLSEsRequest{}
			resp, err := tf.Fleet.ListMachineLSEs(tf.C, req)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.MachineLSEs, ShouldResembleProto, machineLSEs)
		})

		Convey("ListMachineLSEs - filter format invalid format OR - error", func() {
			req := &ufsAPI.ListMachineLSEsRequest{
				Filter: "machine=mac-1|rpm=rpm-2",
			}
			_, err := tf.Fleet.ListMachineLSEs(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidFilterFormat)
		})

		Convey("ListMachineLSEs - filter format valid AND", func() {
			req := &ufsAPI.ListMachineLSEsRequest{
				Filter: "machine=mac-1 & machineprototype=mlsep-1",
			}
			resp, err := tf.Fleet.ListMachineLSEs(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.MachineLSEs, ShouldBeNil)
		})

		Convey("ListMachineLSEs - filter format valid", func() {
			req := &ufsAPI.ListMachineLSEsRequest{
				Filter: "machine=mac-1",
			}
			resp, err := tf.Fleet.ListMachineLSEs(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.MachineLSEs, ShouldResembleProto, machineLSEs)
		})
	})
}

func TestDeleteMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("DeleteMachineLSE", t, func() {
		Convey("Delete machineLSE by existing ID", func() {
			machine := &ufspb.Machine{
				Name: "machine-1",
			}
			_, err := registration.CreateMachine(ctx, machine)
			So(err, ShouldBeNil)

			_, err = inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name:     "machinelse-1",
				Hostname: "machinelse-1",
				Machines: []string{"machine-1"},
			})
			So(err, ShouldBeNil)

			req := &ufsAPI.DeleteMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "machineLSE-1"),
			}
			_, err = tf.Fleet.DeleteMachineLSE(tf.C, req)
			So(err, ShouldBeNil)

			res, err := inventory.GetMachineLSE(tf.C, "machineLSE-1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machineLSE by existing ID with assigned ip", func() {
			machine := &ufspb.Machine{
				Name: "machine-with-ip",
			}
			_, err := registration.CreateMachine(ctx, machine)
			So(err, ShouldBeNil)

			_, err = inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
				Name:     "machinelse-with-ip",
				Hostname: "machinelse-with-ip",
				Nic:      "eth0",
				Machines: []string{"machine-1"},
			})
			So(err, ShouldBeNil)
			_, err = configuration.BatchUpdateDHCPs(ctx, []*ufspb.DHCPConfig{
				{
					Hostname: "machinelse-with-ip",
					Ip:       "1.2.3.4",
				},
			})
			So(err, ShouldBeNil)
			_, err = configuration.BatchUpdateIPs(ctx, []*ufspb.IP{
				{
					Id:       "vlan:1234",
					Ipv4:     1234,
					Ipv4Str:  "1.2.3.4",
					Vlan:     "vlan",
					Occupied: true,
				},
			})
			So(err, ShouldBeNil)

			req := &ufsAPI.DeleteMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "machineLSE-with-ip"),
			}
			_, err = tf.Fleet.DeleteMachineLSE(tf.C, req)
			So(err, ShouldBeNil)

			res, err := inventory.GetMachineLSE(tf.C, "machineLSE-1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
			dhcp, err := configuration.GetDHCPConfig(ctx, "machineLSE-with-ip")
			So(err, ShouldNotBeNil)
			So(dhcp, ShouldBeNil)
			s, _ := status.FromError(err)
			So(s.Code(), ShouldEqual, codes.NotFound)
			ips, err := configuration.QueryIPByPropertyName(ctx, map[string]string{"ipv4_str": "1.2.3.4"})
			So(err, ShouldBeNil)
			So(ips, ShouldHaveLength, 1)
			So(ips[0].GetOccupied(), ShouldBeFalse)
		})
		Convey("Delete machineLSE by non-existing ID", func() {
			req := &ufsAPI.DeleteMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "machineLSE-2"),
			}
			_, err := tf.Fleet.DeleteMachineLSE(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machineLSE - Invalid input empty name", func() {
			req := &ufsAPI.DeleteMachineLSERequest{
				Name: "",
			}
			resp, err := tf.Fleet.DeleteMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Delete machineLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.DeleteMachineLSERequest{
				Name: util.AddPrefix(util.MachineLSECollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.DeleteMachineLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestRenameMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("RenameMachineLSE", t, func() {
		Convey("Rename an empty machineLSE name - returns error", func() {
			_, err := tf.Fleet.RenameMachineLSE(tf.C, &ufsAPI.RenameMachineLSERequest{
				Name: "",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Rename a machineLSE to an empty name - returns error", func() {
			_, err := tf.Fleet.RenameMachineLSE(tf.C, &ufsAPI.RenameMachineLSERequest{
				Name:    "oldMachineLSE",
				NewName: "",
			})
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCreateRackLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	rackLSE1 := mockRackLSE("rackLSE-1")
	rackLSE2 := mockRackLSE("rackLSE-2")
	Convey("CreateRackLSEs", t, func() {
		Convey("Create new rackLSE with rackLSE_id", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE:   rackLSE1,
				RackLSEId: "rackLSE-1",
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rackLSE1)
		})

		Convey("Create existing rackLSEs", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE:   rackLSE1,
				RackLSEId: "rackLSE-1",
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, AlreadyExists)
		})

		Convey("Create new rackLSE - Invalid input nil", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE: nil,
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Create new rackLSE - Invalid input empty ID", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE:   rackLSE2,
				RackLSEId: "",
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyID)
		})

		Convey("Create new rackLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE:   rackLSE2,
				RackLSEId: "a.b)7&",
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestUpdateRackLSE(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	rackLSE1 := mockRackLSE("rackLSE-1")
	rackLSE2 := mockRackLSE("rackLSE-1")
	rackLSE3 := mockRackLSE("rackLSE-3")
	rackLSE4 := mockRackLSE("")
	rackLSE5 := mockRackLSE("a.b)7&")
	Convey("UpdateRackLSEs", t, func() {
		Convey("Update existing rackLSEs", func() {
			req := &ufsAPI.CreateRackLSERequest{
				RackLSE:   rackLSE1,
				RackLSEId: "rackLSE-1",
			}
			resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rackLSE1)
			ureq := &ufsAPI.UpdateRackLSERequest{
				RackLSE: rackLSE2,
			}
			resp, err = tf.Fleet.UpdateRackLSE(tf.C, ureq)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rackLSE2)
		})

		Convey("Update non-existing rackLSEs", func() {
			ureq := &ufsAPI.UpdateRackLSERequest{
				RackLSE: rackLSE3,
			}
			resp, err := tf.Fleet.UpdateRackLSE(tf.C, ureq)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})

		Convey("Update rackLSE - Invalid input nil", func() {
			req := &ufsAPI.UpdateRackLSERequest{
				RackLSE: nil,
			}
			resp, err := tf.Fleet.UpdateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Update rackLSE - Invalid input empty name", func() {
			rackLSE4.Name = ""
			req := &ufsAPI.UpdateRackLSERequest{
				RackLSE: rackLSE4,
			}
			resp, err := tf.Fleet.UpdateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Update rackLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.UpdateRackLSERequest{
				RackLSE: rackLSE5,
			}
			resp, err := tf.Fleet.UpdateRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestGetRackLSE(t *testing.T) {
	t.Parallel()
	Convey("GetRackLSE", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		rackLSE1 := mockRackLSE("rackLSE-1")
		req := &ufsAPI.CreateRackLSERequest{
			RackLSE:   rackLSE1,
			RackLSEId: "rackLSE-1",
		}
		resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, rackLSE1)
		Convey("Get rackLSE by existing ID", func() {
			req := &ufsAPI.GetRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "rackLSE-1"),
			}
			resp, err := tf.Fleet.GetRackLSE(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rackLSE1)
		})
		Convey("Get rackLSE by non-existing ID", func() {
			req := &ufsAPI.GetRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "rackLSE-2"),
			}
			resp, err := tf.Fleet.GetRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get rackLSE - Invalid input empty name", func() {
			req := &ufsAPI.GetRackLSERequest{
				Name: "",
			}
			resp, err := tf.Fleet.GetRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Get rackLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.GetRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.GetRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestListRackLSEs(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	rackLSEs := make([]*ufspb.RackLSE, 0, 4)
	for i := 0; i < 4; i++ {
		resp, _ := inventory.CreateRackLSE(tf.C, &ufspb.RackLSE{
			Name:  fmt.Sprintf("rackLSE-%d", i),
			Racks: []string{"rack-1"},
		})
		resp.Name = util.AddPrefix(util.RackLSECollection, resp.Name)
		rackLSEs = append(rackLSEs, resp)
	}
	Convey("ListRackLSEs", t, func() {
		Convey("ListRackLSEs - page_size negative - error", func() {
			req := &ufsAPI.ListRackLSEsRequest{
				PageSize: -5,
			}
			resp, err := tf.Fleet.ListRackLSEs(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidPageSize)
		})

		Convey("ListRackLSEs - Full listing - happy path", func() {
			req := &ufsAPI.ListRackLSEsRequest{}
			resp, err := tf.Fleet.ListRackLSEs(tf.C, req)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.RackLSEs, ShouldResembleProto, rackLSEs)
		})

		Convey("ListRackLSEs - filter format invalid format OR - error", func() {
			req := &ufsAPI.ListRackLSEsRequest{
				Filter: "rack=mac-1|rpm=rpm-2",
			}
			_, err := tf.Fleet.ListRackLSEs(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidFilterFormat)
		})

		Convey("ListRackLSEs - filter format valid AND", func() {
			req := &ufsAPI.ListRackLSEsRequest{
				Filter: "rack=rack-1 & rackprototype=mlsep-1",
			}
			resp, err := tf.Fleet.ListRackLSEs(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.RackLSEs, ShouldBeNil)
		})

		Convey("ListRackLSEs - filter format valid", func() {
			req := &ufsAPI.ListRackLSEsRequest{
				Filter: "rack=rack-1",
			}
			resp, err := tf.Fleet.ListRackLSEs(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.RackLSEs, ShouldResembleProto, rackLSEs)
		})
	})
}

func TestDeleteRackLSE(t *testing.T) {
	t.Parallel()
	Convey("DeleteRackLSE", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		rackLSE1 := mockRackLSE("")
		req := &ufsAPI.CreateRackLSERequest{
			RackLSE:   rackLSE1,
			RackLSEId: "rackLSE-1",
		}
		resp, err := tf.Fleet.CreateRackLSE(tf.C, req)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, rackLSE1)
		Convey("Delete rackLSE by existing ID", func() {
			req := &ufsAPI.DeleteRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "rackLSE-1"),
			}
			_, err := tf.Fleet.DeleteRackLSE(tf.C, req)
			So(err, ShouldBeNil)
			greq := &ufsAPI.GetRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "rackLSE-1"),
			}
			res, err := tf.Fleet.GetRackLSE(tf.C, greq)
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete rackLSE by non-existing ID", func() {
			req := &ufsAPI.DeleteRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "rackLSE-2"),
			}
			_, err := tf.Fleet.DeleteRackLSE(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete rackLSE - Invalid input empty name", func() {
			req := &ufsAPI.DeleteRackLSERequest{
				Name: "",
			}
			resp, err := tf.Fleet.DeleteRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Delete rackLSE - Invalid input invalid characters", func() {
			req := &ufsAPI.DeleteRackLSERequest{
				Name: util.AddPrefix(util.RackLSECollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.DeleteRackLSE(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestImportOSMachineLSEs(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("Import ChromeOS machine lses", t, func() {
		Convey("happy path", func() {
			req := &ufsAPI.ImportOSMachineLSEsRequest{
				Source: &ufsAPI.ImportOSMachineLSEsRequest_MachineDbSource{
					MachineDbSource: &ufsAPI.MachineDBSource{
						Host: "fake_host",
					},
				},
			}
			tf.Fleet.importPageSize = 25
			res, err := tf.Fleet.ImportOSMachineLSEs(ctx, req)
			So(err, ShouldBeNil)
			So(res.Code, ShouldEqual, code.Code_OK)

			// Verify machine lse prototypes
			lps, _, err := configuration.ListMachineLSEPrototypes(ctx, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(ufsAPI.ParseResources(lps, "Name"), ShouldResemble, []string{"acs:camera", "acs:wificell", "atl:labstation", "atl:standard"})

			// Verify machine lses
			machineLSEs, _, err := inventory.ListMachineLSEs(ctx, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(ufsAPI.ParseResources(machineLSEs, "Name"), ShouldResemble, []string{"chromeos2-test_host", "chromeos3-test_host", "chromeos5-test_host", "test_servo"})
			// Spot check some fields
			for _, r := range machineLSEs {
				switch r.GetName() {
				case "test_host", "chromeos1-test_host", "chromeos3-test_host":
					So(r.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPools(), ShouldResemble, []string{"DUT_POOL_QUOTA", "hotrod"})
					So(r.GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetSmartUsbhub(), ShouldBeTrue)
				case "test_servo":
					So(r.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetPools(), ShouldResemble, []string{"labstation_main"})
					So(r.GetChromeosMachineLse().GetDeviceLse().GetLabstation().GetRpm().GetPowerunitName(), ShouldEqual, "test_power_unit_name")
				}
			}
			lse, err := inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", "mock_dut_id", false)
			So(err, ShouldBeNil)
			So(lse, ShouldHaveLength, 1)
			So(lse[0].GetMachineLsePrototype(), ShouldEqual, "atl:standard")
			So(lse[0].GetHostname(), ShouldEqual, "chromeos2-test_host")
			lse, err = inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", "mock_camera_dut_id", false)
			So(err, ShouldBeNil)
			So(lse, ShouldHaveLength, 1)
			So(lse[0].GetMachineLsePrototype(), ShouldEqual, "acs:camera")
			So(lse[0].GetHostname(), ShouldEqual, "chromeos3-test_host")
			lse, err = inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", "mock_wifi_dut_id", false)
			So(err, ShouldBeNil)
			So(lse, ShouldHaveLength, 1)
			So(lse[0].GetMachineLsePrototype(), ShouldEqual, "acs:wificell")
			So(lse[0].GetHostname(), ShouldEqual, "chromeos5-test_host")
			lse, err = inventory.QueryMachineLSEByPropertyName(ctx, "machine_ids", "mock_labstation_id", false)
			So(err, ShouldBeNil)
			So(lse, ShouldHaveLength, 1)
			So(lse[0].GetMachineLsePrototype(), ShouldEqual, "atl:labstation")
			So(lse[0].GetHostname(), ShouldEqual, "test_servo")

			// Verify dut states
			resp, err := state.GetAllDutStates(ctx)
			So(err, ShouldBeNil)
			// Labstation doesn't have dut state
			So(resp.Passed(), ShouldHaveLength, 3)
			ds, err := state.GetDutState(ctx, "mock_dut_id")
			So(err, ShouldBeNil)
			So(ds.GetServo(), ShouldEqual, chromeosLab.PeripheralState_WORKING)
			So(ds.GetWorkingBluetoothBtpeer(), ShouldEqual, 1)
			So(ds.GetStorageState(), ShouldEqual, chromeosLab.HardwareState_HARDWARE_NORMAL)
			So(ds.GetCr50Phase(), ShouldEqual, chromeosLab.DutState_CR50_PHASE_PVT)
			So(ds.GetHostname(), ShouldEqual, "chromeos2-test_host")

			ds, err = state.GetDutState(ctx, "mock_camera_dut_id")
			So(err, ShouldBeNil)
			So(ds.GetServo(), ShouldEqual, chromeosLab.PeripheralState_SERVOD_ISSUE)
			So(ds.GetWorkingBluetoothBtpeer(), ShouldEqual, 1)
			So(ds.GetStorageState(), ShouldEqual, chromeosLab.HardwareState_HARDWARE_NORMAL)
			So(ds.GetCr50Phase(), ShouldEqual, chromeosLab.DutState_CR50_PHASE_PVT)
			So(ds.GetHostname(), ShouldEqual, "chromeos3-test_host")

			ds, err = state.GetDutState(ctx, "mock_wifi_dut_id")
			So(err, ShouldBeNil)
			So(ds.GetServo(), ShouldEqual, chromeosLab.PeripheralState_NOT_CONNECTED)
			So(ds.GetWorkingBluetoothBtpeer(), ShouldEqual, 1)
			So(ds.GetStorageState(), ShouldEqual, chromeosLab.HardwareState_HARDWARE_NORMAL)
			So(ds.GetCr50Phase(), ShouldEqual, chromeosLab.DutState_CR50_PHASE_PVT)
			So(ds.GetHostname(), ShouldEqual, "chromeos5-test_host")

			ds, err = state.GetDutState(ctx, "mock_labstation_id")
			So(ds, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestGetMachineLSEDeployment(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("GetMachineLSEDeployment", t, func() {
		Convey("Get machine lse deployment record by existing ID", func() {
			dr1, err := inventory.UpdateMachineLSEDeployments(ctx, []*ufspb.MachineLSEDeployment{
				{
					SerialNumber: "dr-get-1",
				},
			})
			So(err, ShouldBeNil)
			So(dr1, ShouldHaveLength, 1)
			dr1[0].SerialNumber = util.AddPrefix(util.MachineLSEDeploymentCollection, dr1[0].SerialNumber)

			req := &ufsAPI.GetMachineLSEDeploymentRequest{
				Name: util.AddPrefix(util.MachineLSEDeploymentCollection, "dr-get-1"),
			}
			resp, err := tf.Fleet.GetMachineLSEDeployment(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, dr1[0])
		})
		Convey("Get machine lse deployment record by non-existing ID", func() {
			req := &ufsAPI.GetMachineLSEDeploymentRequest{
				Name: util.AddPrefix(util.MachineLSEDeploymentCollection, "dr-get-2"),
			}
			resp, err := tf.Fleet.GetMachineLSEDeployment(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get machine lse deployment record - Invalid input empty name", func() {
			req := &ufsAPI.GetMachineLSEDeploymentRequest{
				Name: "",
			}
			resp, err := tf.Fleet.GetMachineLSEDeployment(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})
		Convey("Get machine lse deployment record - Invalid input invalid characters", func() {
			req := &ufsAPI.GetMachineLSEDeploymentRequest{
				Name: util.AddPrefix(util.MachineLSEDeploymentCollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.GetMachineLSEDeployment(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestCreateSchedulingUnit(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("CreateSchedulingUnit", t, func() {
		Convey("Create new SchedulingUnit with schedulingUnitId - happy path", func() {
			su := mockSchedulingUnit("")
			req := &ufsAPI.CreateSchedulingUnitRequest{
				SchedulingUnit:   su,
				SchedulingUnitId: "su-1",
			}
			resp, err := tf.Fleet.CreateSchedulingUnit(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, su)
		})

		Convey("Create new SchedulingUnit with nil entity", func() {
			req := &ufsAPI.CreateSchedulingUnitRequest{
				SchedulingUnit:   nil,
				SchedulingUnitId: "su-2",
			}
			_, err := tf.Fleet.CreateSchedulingUnit(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Create new SchedulingUnit without schedulingUnitId", func() {
			su := mockSchedulingUnit("")
			req := &ufsAPI.CreateSchedulingUnitRequest{
				SchedulingUnit: su,
			}
			_, err := tf.Fleet.CreateSchedulingUnit(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyID)
		})
	})
}

func TestUpdateSchedulingUnit(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	Convey("UpdateSchedulingUnit", t, func() {
		Convey("Update existing SchedulingUnit - happy path", func() {
			inventory.CreateSchedulingUnit(ctx, &ufspb.SchedulingUnit{
				Name: "su-1",
			})

			su1 := mockSchedulingUnit("su-1")
			su1.Description = "Updatedesc"
			ureq := &ufsAPI.UpdateSchedulingUnitRequest{
				SchedulingUnit: su1,
			}
			resp, err := tf.Fleet.UpdateSchedulingUnit(tf.C, ureq)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, su1)
		})

		Convey("Update SchedulingUnit - Invalid input nil", func() {
			req := &ufsAPI.UpdateSchedulingUnitRequest{
				SchedulingUnit: nil,
			}
			resp, err := tf.Fleet.UpdateSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.NilEntity)
		})

		Convey("Update SchedulingUnit - Invalid input empty name", func() {
			su := mockSchedulingUnit("")
			su.Name = ""
			req := &ufsAPI.UpdateSchedulingUnitRequest{
				SchedulingUnit: su,
			}
			resp, err := tf.Fleet.UpdateSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Update SchedulingUnit - Invalid input invalid name", func() {
			su := mockSchedulingUnit("a.b)7&")
			req := &ufsAPI.UpdateSchedulingUnitRequest{
				SchedulingUnit: su,
			}
			resp, err := tf.Fleet.UpdateSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestGetSchedulingUnit(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	su, _ := inventory.CreateSchedulingUnit(ctx, &ufspb.SchedulingUnit{
		Name: "su-1",
	})
	Convey("GetSchedulingUnit", t, func() {
		Convey("Get SchedulingUnit by existing ID - happy path", func() {
			req := &ufsAPI.GetSchedulingUnitRequest{
				Name: util.AddPrefix(util.SchedulingUnitCollection, "su-1"),
			}
			resp, _ := tf.Fleet.GetSchedulingUnit(tf.C, req)
			So(resp, ShouldNotBeNil)
			resp.Name = util.RemovePrefix(resp.Name)
			So(resp, ShouldResembleProto, su)
		})

		Convey("Get SchedulingUnit - Invalid input empty name", func() {
			req := &ufsAPI.GetSchedulingUnitRequest{
				Name: "",
			}
			resp, err := tf.Fleet.GetSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Get SchedulingUnit - Invalid input invalid characters", func() {
			req := &ufsAPI.GetSchedulingUnitRequest{
				Name: util.AddPrefix(util.SchedulingUnitCollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.GetSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestDeleteSchedulingUnit(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	inventory.CreateSchedulingUnit(ctx, &ufspb.SchedulingUnit{
		Name: "su-1",
	})
	Convey("DeleteSchedulingUnit", t, func() {
		Convey("Delete SchedulingUnit by existing ID - happy path", func() {
			req := &ufsAPI.DeleteSchedulingUnitRequest{
				Name: util.AddPrefix(util.SchedulingUnitCollection, "su-1"),
			}
			_, err := tf.Fleet.DeleteSchedulingUnit(tf.C, req)
			So(err, ShouldBeNil)

			res, err := inventory.GetSchedulingUnit(tf.C, "su-1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})

		Convey("Delete SchedulingUnit - Invalid input empty name", func() {
			req := &ufsAPI.DeleteSchedulingUnitRequest{
				Name: "",
			}
			resp, err := tf.Fleet.DeleteSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.EmptyName)
		})

		Convey("Delete SchedulingUnit - Invalid input invalid characters", func() {
			req := &ufsAPI.DeleteSchedulingUnitRequest{
				Name: util.AddPrefix(util.SchedulingUnitCollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.DeleteSchedulingUnit(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidCharacters)
		})
	})
}

func TestListSchedulingUnits(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()
	schedulingUnits := make([]*ufspb.SchedulingUnit, 0, 4)
	for i := 0; i < 4; i++ {
		su := mockSchedulingUnit("")
		su.Name = fmt.Sprintf("su-%d", i)
		resp, _ := inventory.CreateSchedulingUnit(tf.C, su)
		resp.Name = util.AddPrefix(util.SchedulingUnitCollection, resp.Name)
		schedulingUnits = append(schedulingUnits, resp)
	}
	Convey("ListSchedulingUnits", t, func() {
		Convey("ListSchedulingUnits - page_size negative - error", func() {
			req := &ufsAPI.ListSchedulingUnitsRequest{
				PageSize: -5,
			}
			resp, err := tf.Fleet.ListSchedulingUnits(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidPageSize)
		})

		Convey("ListSchedulingUnits - Full listing with no pagination - happy path", func() {
			req := &ufsAPI.ListSchedulingUnitsRequest{}
			resp, err := tf.Fleet.ListSchedulingUnits(tf.C, req)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp.SchedulingUnits, ShouldResembleProto, schedulingUnits)
		})

		Convey("ListSchedulingUnits - filter format invalid format OR - error", func() {
			req := &ufsAPI.ListSchedulingUnitsRequest{
				Filter: "state=x|state=y",
			}
			_, err := tf.Fleet.ListSchedulingUnits(tf.C, req)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsAPI.InvalidFilterFormat)
		})
	})
}

func TestGetDeviceData(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Debug)
	osCtx, _ := util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	su, _ := inventory.CreateSchedulingUnit(osCtx, &ufspb.SchedulingUnit{
		Name: "su-1",
	})

	machine := &ufspb.Machine{
		Name: "machine-1",
		Device: &ufspb.Machine_ChromeosMachine{
			ChromeosMachine: &ufspb.ChromeOSMachine{
				ReferenceBoard: "test",
				BuildTarget:    "test",
				Model:          "test",
				Hwid:           "test",
			},
		},
	}
	registration.CreateMachine(osCtx, machine)

	machinelse := mockDutMachineLSE("lse-1")
	// In datastore, the name doesn't contain prefix
	machinelse.Name = "lse-1"
	machinelse.Machines = []string{"machine-1"}
	inventory.CreateMachineLSE(osCtx, machinelse)

	adm := &ufspb.Machine{
		Name: "adm-1",
		Device: &ufspb.Machine_AttachedDevice{
			AttachedDevice: &ufspb.AttachedDevice{
				Manufacturer: "Apple",
				DeviceType:   ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_APPLE_PHONE,
				BuildTarget:  "test",
				Model:        "test",
			},
		},
	}
	registration.CreateMachine(osCtx, adm)

	admlse := mockAttachedDeviceMachineLSE("admlse-1")
	admlse.Name = "admlse-1"
	admlse.Machines = []string{"adm-1"}
	inventory.CreateMachineLSE(osCtx, admlse)

	// Add Browser host/vm
	inventory.CreateMachineLSE(ctx, &ufspb.MachineLSE{
		Name:     "browser-host1",
		Hostname: "browser-host1",
		Lse: &ufspb.MachineLSE_ChromeBrowserMachineLse{
			ChromeBrowserMachineLse: &ufspb.ChromeBrowserMachineLSE{},
		},
	})
	vms := []*ufspb.VM{
		{
			Name: "browser-vm1",
			OsVersion: &ufspb.OSVersion{
				Value: "os-1",
			},
			ResourceState: ufspb.State_STATE_SERVING,
		},
	}
	inventory.BatchUpdateVMs(ctx, vms)

	Convey("GetDeviceData", t, func() {
		Convey("Get Browser Host by hostname - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: "browser-host1",
			}
			resp, err := tf.Fleet.GetDeviceData(ctx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_BROWSER_DEVICE)
			So(resp.GetBrowserDeviceData().GetHost().GetHostname(), ShouldEqual, "browser-host1")
		})

		Convey("Get Browser VM by hostname - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: "browser-vm1",
			}
			resp, err := tf.Fleet.GetDeviceData(ctx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_BROWSER_DEVICE)
			So(resp.GetBrowserDeviceData().GetVm().GetName(), ShouldEqual, "browser-vm1")
		})

		Convey("Get SchedulingUnit by existing ID - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.SchedulingUnitCollection, "su-1"),
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			resp.GetSchedulingUnit().Name = util.RemovePrefix(resp.GetSchedulingUnit().Name)
			So(resp.GetSchedulingUnit(), ShouldResembleProto, su)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_SCHEDULING_UNIT)
		})

		Convey("Get ChromeOSDeviceData by existing hostname - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.MachineLSECollection, "lse-1"),
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetChromeOsDeviceData().GetLabConfig(), ShouldResembleProto, machinelse)
			So(resp.GetChromeOsDeviceData().GetMachine(), ShouldResembleProto, machine)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE)
		})

		Convey("Get ChromeOSDeviceData by existing hostname in os namespace", func() {
			machineOs := &ufspb.Machine{
				Name: "machine-os-1",
				Device: &ufspb.Machine_ChromeosMachine{
					ChromeosMachine: &ufspb.ChromeOSMachine{
						ReferenceBoard: "test",
						BuildTarget:    "test",
						Model:          "test",
						Hwid:           "test",
					},
				},
			}
			registration.CreateMachine(osCtx, machineOs)

			machineOsLse := mockDutMachineLSE("lse-os-1")
			machineOsLse.Name = "lse-os-1"
			machineOsLse.Machines = []string{"machine-os-1"}
			inventory.CreateMachineLSE(osCtx, machineOsLse)

			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.MachineLSECollection, "lse-os-1"),
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetChromeOsDeviceData().GetLabConfig(), ShouldResembleProto, machineOsLse)
			So(resp.GetChromeOsDeviceData().GetMachine(), ShouldResembleProto, machineOs)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE)

			// Should not exist in Browser namespace
			respBrowser, err := tf.Fleet.GetDeviceData(tf.C, req)
			So(respBrowser, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no valid device found")
		})

		Convey("Get ChromeOSDeviceData by existing asset tag - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				DeviceId: "machine-1",
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetChromeOsDeviceData().GetLabConfig(), ShouldResembleProto, machinelse)
			So(resp.GetChromeOsDeviceData().GetMachine(), ShouldResembleProto, machine)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_CHROMEOS_DEVICE)
		})

		Convey("Get AttachedDeviceData by existing hostname - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.MachineLSECollection, "admlse-1"),
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetAttachedDeviceData().GetLabConfig(), ShouldResembleProto, admlse)
			So(resp.GetAttachedDeviceData().GetMachine(), ShouldResembleProto, adm)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE)
		})

		Convey("Get AttachedDeviceData by existing hostname in os namespace", func() {
			adm := &ufspb.Machine{
				Name: "adm-os-1",
				Device: &ufspb.Machine_AttachedDevice{
					AttachedDevice: &ufspb.AttachedDevice{
						Manufacturer: "Apple",
						DeviceType:   ufspb.AttachedDeviceType_ATTACHED_DEVICE_TYPE_APPLE_PHONE,
						BuildTarget:  "test",
						Model:        "test",
					},
				},
			}
			registration.CreateMachine(osCtx, adm)

			admlse := mockAttachedDeviceMachineLSE("admlse-os-1")
			admlse.Name = "admlse-os-1"
			admlse.Machines = []string{"adm-os-1"}
			inventory.CreateMachineLSE(osCtx, admlse)

			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.MachineLSECollection, "admlse-os-1"),
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetAttachedDeviceData().GetLabConfig(), ShouldResembleProto, admlse)
			So(resp.GetAttachedDeviceData().GetMachine(), ShouldResembleProto, adm)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE)

			// Should not exist in Browser namespace
			respBrowser, err := tf.Fleet.GetDeviceData(tf.C, req)
			So(respBrowser, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no valid device found")
		})

		Convey("Get AttachedDeviceData by existing asset tag - happy path", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				DeviceId: "adm-1",
			}
			resp, err := tf.Fleet.GetDeviceData(osCtx, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.GetAttachedDeviceData().GetLabConfig(), ShouldResembleProto, admlse)
			So(resp.GetAttachedDeviceData().GetMachine(), ShouldResembleProto, adm)
			So(resp.GetResourceType(), ShouldEqual, ufsAPI.GetDeviceDataResponse_RESOURCE_TYPE_ATTACHED_DEVICE)
		})

		Convey("Get device data - Invalid input empty name", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: "",
			}
			resp, err := tf.Fleet.GetDeviceData(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Both Id and hostname are empty")
		})

		Convey("Get device data - Invalid input invalid characters", func() {
			req := &ufsAPI.GetDeviceDataRequest{
				Hostname: util.AddPrefix(util.SchedulingUnitCollection, "a.b)7&"),
			}
			resp, err := tf.Fleet.GetDeviceData(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no valid device found")
		})
	})
}

func mustParseMachineLSE(input string) *ufspb.MachineLSE {
	out := &ufspb.MachineLSE{}
	err := protojson.Unmarshal([]byte(input), out)
	if err != nil {
		panic(err.Error())
	}
	return out
}

// TestGetDUTsForLabstation tests extracting DUT names from servo names on a labstation object.
func TestGetDUTsForLabstation(t *testing.T) {
	t.Parallel()
	ctx := memory.Use(context.Background())
	datastore.GetTestable(ctx).Consistent(true)
	ctx = logging.SetLevel(ctx, logging.Debug)
	ctx, _ = util.SetupDatastoreNamespace(ctx, util.OSNamespace)
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	const labstation = `
{
        "name": "fake-labstation",
        "hostname": "fake-labstation",
        "chromeosMachineLse": {
                "deviceLse": {
                        "networkDeviceInterface": null,
                        "labstation": {
                                "hostname": "fake-labstation",
                                "servos": [
                                        {
                                                "servoHostname": "fake-labstation",
                                                "servoPort": 9999,
                                                "servoSerial": "XXXXXXXXXXX"
                                        }
                                ],
                                "rpm": {
                                        "powerunitName": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
                                        "powerunitOutlet": "XXX"
                                },
                                "pools": [
                                        "labstation_main"
                                ]
                        }
                }
        },
        "machines": [
                "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
        ]
}
`

	const dut = `
{
        "name":  "fake-dut",
        "machineLsePrototype":  "atl:standard",
        "hostname":  "fake-dut",
        "chromeosMachineLse":  {
                "deviceLse":  {
                        "dut":  {
                                "hostname":  "fake-dut",
                                "peripherals":  {
                                        "servo":  {
                                                "servoHostname":  "fake-labstation",
                                                "servoPort":  9999
                                        }
                                }
                        }
                }
        },
        "machines":  [
                "XXXXXXX"
        ]
}
`

	Convey("smoke test", t, func() {
		_, err := inventory.CreateMachineLSE(ctx, mustParseMachineLSE(labstation))
		if err != nil {
			panic(err.Error())
		}

		req := &ufsAPI.GetDUTsForLabstationRequest{
			Hostname: []string{"fake-labstation"},
		}

		resp, err := tf.Fleet.GetDUTsForLabstation(tf.C, req)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, &ufsAPI.GetDUTsForLabstationResponse{
			Items: []*ufsAPI.GetDUTsForLabstationResponse_LabstationMapping{
				{
					Hostname: "fake-labstation",
					DutName:  []string{},
				},
			},
		})

		_, err = inventory.CreateMachineLSE(ctx, mustParseMachineLSE(dut))
		if err != nil {
			panic(err.Error())
		}

		resp, err = tf.Fleet.GetDUTsForLabstation(tf.C, req)
		So(err, ShouldBeNil)
		So(resp, ShouldResembleProto, &ufsAPI.GetDUTsForLabstationResponse{
			Items: []*ufsAPI.GetDUTsForLabstationResponse_LabstationMapping{
				{
					Hostname: "fake-labstation",
					DutName:  []string{"fake-dut"},
				},
			},
		})

	})
}
