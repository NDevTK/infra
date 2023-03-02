// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package registration

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	. "infra/unifiedfleet/app/model/datastore"
)

func mockChromeOSMachine(id, lab, board string) *ufspb.Machine {
	return &ufspb.Machine{
		Name: id,
		Device: &ufspb.Machine_ChromeosMachine{
			ChromeosMachine: &ufspb.ChromeOSMachine{
				ReferenceBoard: board,
			},
		},
	}
}

func mockChromeBrowserMachine(id, lab, name string) *ufspb.Machine {
	return &ufspb.Machine{
		Name: id,
		Device: &ufspb.Machine_ChromeBrowserMachine{
			ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
				Description: name,
			},
		},
	}
}

func mockChromeBrowserMachineWithOwnership(id, lab, name string, ownership *ufspb.OwnershipData) *ufspb.Machine {
	machine := mockChromeBrowserMachine(id, lab, name)
	machine.Ownership = ownership
	return machine
}

func mockAttachedDevice(id, lab, buildTarget string) *ufspb.Machine {
	return &ufspb.Machine{
		Name: id,
		Device: &ufspb.Machine_AttachedDevice{
			AttachedDevice: &ufspb.AttachedDevice{
				BuildTarget: buildTarget,
			},
		},
	}
}

func assertMachineEqual(a *ufspb.Machine, b *ufspb.Machine) {
	So(a.GetName(), ShouldEqual, b.GetName())
	So(a.GetChromeBrowserMachine().GetDescription(), ShouldEqual,
		b.GetChromeBrowserMachine().GetDescription())
	So(a.GetChromeosMachine().GetReferenceBoard(), ShouldEqual,
		b.GetChromeosMachine().GetReferenceBoard())
	So(a.GetAttachedDevice().GetBuildTarget(), ShouldEqual,
		b.GetAttachedDevice().GetBuildTarget())
}

func assertMachineWithOwnershipEqual(a *ufspb.Machine, b *ufspb.Machine) {
	if a.GetOwnership() == nil && b.GetOwnership() == nil {
		return
	}
	assertMachineEqual(a, b)
	So(a.GetOwnership().PoolName, ShouldEqual, b.GetOwnership().PoolName)
	So(a.GetOwnership().SwarmingInstance, ShouldEqual, b.GetOwnership().SwarmingInstance)
	So(a.GetOwnership().Customer, ShouldEqual, b.GetOwnership().Customer)
	So(a.GetOwnership().SecurityLevel, ShouldEqual, b.GetOwnership().SecurityLevel)
}

func getMachineNames(machines []*ufspb.Machine) []string {
	names := make([]string, len(machines))
	for i, p := range machines {
		names[i] = p.GetName()
	}
	return names
}

func TestCreateMachine(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	chromeOSMachine1 := mockChromeOSMachine("chromeos-asset-1", "chromeoslab", "samus")
	chromeOSMachine2 := mockChromeOSMachine("", "chromeoslab", "samus")
	attchedDevice1 := mockAttachedDevice("attached-device-1", "chromeoslab", "goldfish")
	chromeBrowserMachine1 := mockChromeBrowserMachine("chrome-asset-1", "chromelab", "machine-1")

	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	chromeBrowserMachineWithOwnership := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	Convey("CreateMachine", t, func() {
		Convey("Create new os machine", func() {
			resp, err := CreateMachine(ctx, chromeOSMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine1)
		})
		Convey("Create new attached device", func() {
			resp, err := CreateMachine(ctx, attchedDevice1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, attchedDevice1)
		})
		Convey("Create existing machine", func() {
			resp, err := CreateMachine(ctx, chromeOSMachine1)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, AlreadyExists)
		})
		Convey("Create machine - invalid ID", func() {
			resp, err := CreateMachine(ctx, chromeOSMachine2)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Create new browser machine with ownership data - ownership is not saved", func() {
			resp, err := CreateMachine(ctx, chromeBrowserMachineWithOwnership)
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachine1)
		})
	})
}

func TestUpdateMachine(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	chromeOSMachine1 := mockChromeOSMachine("chromeos-asset-1", "chromeoslab", "samus")
	chromeOSMachine2 := mockChromeOSMachine("chromeos-asset-1", "chromeoslab", "veyron")
	chromeBrowserMachine1 := mockChromeBrowserMachine("chrome-asset-1", "chromelab", "machine-1")

	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	chromeBrowserMachineWithOwnership := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	chromeOSMachine3 := mockChromeOSMachine("", "chromeoslab", "samus")
	Convey("UpdateMachine", t, func() {
		Convey("Update existing machine", func() {
			resp, err := CreateMachine(ctx, chromeOSMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine1)

			resp, err = UpdateMachine(ctx, chromeOSMachine2)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine2)
		})
		Convey("Update non-existing machine", func() {
			resp, err := UpdateMachine(ctx, chromeBrowserMachine1)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update machine - invalid ID", func() {
			resp, err := UpdateMachine(ctx, chromeOSMachine3)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Update existing machine - does not update ownership", func() {
			resp, err := CreateMachine(ctx, chromeBrowserMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeBrowserMachine1)

			resp, err = UpdateMachine(ctx, chromeBrowserMachineWithOwnership)
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachine1)
			So(resp.GetOwnership(), ShouldBeNil)
		})
	})
}

func TestUpdateMachineOwnership(t *testing.T) {
	// Tests the ownership update scenarios for a machine
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	ownershipData2 := &ufspb.OwnershipData{
		PoolName:         "pool2",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	chromeBrowserMachine1 := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	chromeBrowserMachine1copy := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	chromeBrowserMachine2 := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-2", ownershipData2)
	chromeBrowserMachine2_oldOwnership := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-2", ownershipData)

	Convey("UpdateMachine", t, func() {
		Convey("Update existing machine with ownership data", func() {
			resp, err := CreateMachine(ctx, chromeBrowserMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeBrowserMachine1)

			// Ownership data should be updated
			resp, err = UpdateMachineOwnership(ctx, resp.Name, ownershipData)
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachine1copy)

			// Regular Update calls should not override ownership data
			resp, err = UpdateMachine(ctx, chromeBrowserMachine2)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeBrowserMachine2)

			resp, err = GetMachine(ctx, "chrome-asset-1")
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachine2_oldOwnership)
		})
		Convey("Update non-existing machine with ownership", func() {
			resp, err := UpdateMachineOwnership(ctx, "dummy", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update machine with ownership - invalid ID", func() {
			resp, err := UpdateMachineOwnership(ctx, "", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestGetMachine(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	chromeOSMachine1 := mockChromeOSMachine("chromeos-asset-3", "chromeoslab", "samus")

	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	chromeBrowserMachine1 := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	chromeBrowserMachinecopy := mockChromeBrowserMachineWithOwnership("chrome-asset-1", "chromelab", "machine-1", ownershipData)
	Convey("GetMachine", t, func() {
		Convey("Get machine by existing ID", func() {
			resp, err := CreateMachine(ctx, chromeOSMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine1)
			resp, err = GetMachine(ctx, "chromeos-asset-3")
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine1)
		})
		Convey("Get machine by non-existing ID", func() {
			resp, err := GetMachine(ctx, "chrome-asset-1")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get machine - invalid ID", func() {
			resp, err := GetMachine(ctx, "")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Get machine with ownership by existing ID", func() {
			resp, err := CreateMachine(ctx, chromeBrowserMachine1)
			So(err, ShouldBeNil)
			assertMachineEqual(resp, chromeBrowserMachine1)
			So(resp.GetOwnership(), ShouldBeNil)

			// Ownership data should be updated
			resp, err = UpdateMachineOwnership(ctx, resp.Name, ownershipData)
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachinecopy)

			resp, err = GetMachine(ctx, "chrome-asset-1")
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachinecopy)
		})
	})
}

func TestListMachines(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machines := make([]*ufspb.Machine, 0, 4)
	for i := 0; i < 4; i++ {
		chromeOSMachine1 := mockChromeOSMachine(fmt.Sprintf("chromeos-%d", i), "chromeoslab", "samus")
		resp, _ := CreateMachine(ctx, chromeOSMachine1)
		machines = append(machines, resp)
	}
	Convey("ListMachines", t, func() {
		Convey("List machines - page_token invalid", func() {
			resp, nextPageToken, err := ListMachines(ctx, 5, "abc", nil, false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List machines - Full listing with no pagination", func() {
			resp, nextPageToken, err := ListMachines(ctx, 4, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines)
		})

		Convey("List machines - listing with pagination", func() {
			resp, nextPageToken, err := ListMachines(ctx, 3, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines[:3])

			resp, _, err = ListMachines(ctx, 2, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines[3:])
		})
	})
}

// TestListMachinesByIdPrefixSearch tests the functionality for listing
// machines by seraching for name/id prefix
func TestListMachinesByIdPrefixSearch(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machines := make([]*ufspb.Machine, 0, 4)
	for i := 0; i < 4; i++ {
		chromeOSMachine1 := mockChromeOSMachine(fmt.Sprintf("chromeos-%d", i), "chromeoslab", "samus")
		resp, _ := CreateMachine(ctx, chromeOSMachine1)
		machines = append(machines, resp)
	}
	Convey("ListMachinesByIdPrefixSearch", t, func() {
		Convey("List machines - page_token invalid", func() {
			resp, nextPageToken, err := ListMachinesByIdPrefixSearch(ctx, 5, "abc", "chromeos-", false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List machines - Full listing with valid prefix and no pagination", func() {
			resp, nextPageToken, err := ListMachinesByIdPrefixSearch(ctx, 4, "", "chromeos-", false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines)
		})

		Convey("List machines - Full listing with invalid prefix", func() {
			resp, nextPageToken, err := ListMachinesByIdPrefixSearch(ctx, 4, "", "chromeos1-", false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("List machines - listing with valid prefix and pagination", func() {
			resp, nextPageToken, err := ListMachinesByIdPrefixSearch(ctx, 3, "", "chromeos-", false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines[:3])

			resp, _, err = ListMachinesByIdPrefixSearch(ctx, 2, nextPageToken, "chromeos-", false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines[3:])
		})
	})
}

func TestDeleteMachine(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	chromeOSMachine2 := mockChromeOSMachine("chromeos-asset-2", "chromeoslab", "samus")

	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	chromeBrowserMachine1 := mockChromeBrowserMachineWithOwnership("chrome-asset-3", "chromelab", "machine-1", ownershipData)
	chromeBrowserMachinecopy := mockChromeBrowserMachineWithOwnership("chrome-asset-3", "chromelab", "machine-1", ownershipData)
	Convey("DeleteMachine", t, func() {
		Convey("Delete machine by existing ID", func() {
			resp, cerr := CreateMachine(ctx, chromeOSMachine2)
			So(cerr, ShouldBeNil)
			assertMachineEqual(resp, chromeOSMachine2)
			err := DeleteMachine(ctx, "chromeos-asset-2")
			So(err, ShouldBeNil)
			res, err := GetMachine(ctx, "chromeos-asset-2")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machine by non-existing ID", func() {
			err := DeleteMachine(ctx, "chrome-asset-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machine - invalid ID", func() {
			err := DeleteMachine(ctx, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Delete machine - with ownershipdata", func() {
			resp, cerr := CreateMachine(ctx, chromeBrowserMachine1)
			So(cerr, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachine1)

			// Ownership data should be updated
			resp, err := UpdateMachineOwnership(ctx, resp.Name, ownershipData)
			So(err, ShouldBeNil)
			assertMachineWithOwnershipEqual(resp, chromeBrowserMachinecopy)

			err = DeleteMachine(ctx, "chrome-asset-3")
			So(err, ShouldBeNil)
		})
	})
}

func TestBatchUpdateMachines(t *testing.T) {
	t.Parallel()
	Convey("BatchUpdateMachines", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		machines := make([]*ufspb.Machine, 0, 4)
		for i := 0; i < 4; i++ {
			chromeOSMachine1 := mockChromeOSMachine(fmt.Sprintf("chromeos-%d", i), "chromeoslab", "samus")
			resp, err := CreateMachine(ctx, chromeOSMachine1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, chromeOSMachine1)
			machines = append(machines, resp)
		}
		Convey("BatchUpdate all machines", func() {
			resp, err := BatchUpdateMachines(ctx, machines)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines)
		})
		Convey("BatchUpdate existing and non-existing machines", func() {
			chromeOSMachine5 := mockChromeOSMachine("", "chromeoslab", "samus")
			machines = append(machines, chromeOSMachine5)
			resp, err := BatchUpdateMachines(ctx, machines)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestQueryMachineByPropertyName(t *testing.T) {
	t.Parallel()
	Convey("QueryMachineByPropertyName", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		dummyMachine := &ufspb.Machine{
			Name: "machine-1",
		}
		machine1 := &ufspb.Machine{
			Name: "machine-1",
			Device: &ufspb.Machine_ChromeBrowserMachine{
				ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
					ChromePlatform: "chromePlatform-1",
					KvmInterface: &ufspb.KVMInterface{
						Kvm: "kvm-1",
					},
					RpmInterface: &ufspb.RPMInterface{
						Rpm: "rpm-1",
					},
				},
			},
		}
		resp, cerr := CreateMachine(ctx, machine1)
		So(cerr, ShouldBeNil)
		So(resp, ShouldResembleProto, machine1)

		machines := make([]*ufspb.Machine, 0, 1)
		machines = append(machines, dummyMachine)

		machines1 := make([]*ufspb.Machine, 0, 1)
		machines1 = append(machines1, machine1)
		Convey("Query By existing ChromePlatform", func() {
			resp, err := QueryMachineByPropertyName(ctx, "chrome_platform_id", "chromePlatform-1", true)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines)
		})
		Convey("Query By non-existing ChromePlatform", func() {
			resp, err := QueryMachineByPropertyName(ctx, "chrome_platform_id", "chromePlatform-2", true)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Query By existing rpm", func() {
			resp, err := QueryMachineByPropertyName(ctx, "rpm_id", "rpm-1", false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines1)
		})
		Convey("Query By non-existing rpm", func() {
			resp, err := QueryMachineByPropertyName(ctx, "rpm_id", "rpm-2", false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Query By existing kvm", func() {
			resp, err := QueryMachineByPropertyName(ctx, "kvm_id", "kvm-1", true)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machines)
		})
		Convey("Query By non-existing kvm", func() {
			resp, err := QueryMachineByPropertyName(ctx, "kvm_id", "kvm-2", true)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
	})
}

/*
func TestGetAllMachines(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	Convey("GetAllMachines", t, func() {
		Convey("Get empty machines", func() {
			resp, err := GetAllMachines(ctx)
			So(err, ShouldBeNil)
			So(resp.Passed(), ShouldHaveLength, 0)
			So(resp.Failed(), ShouldHaveLength, 0)
		})
		Convey("Get all the machines", func() {
			chromeOSMachine1 := mockChromeOSMachine("chromeos-asset-1", "chromeoslab", "samus")
			chromeMachine1 := mockChromeMachine("chrome-asset-1", "chromelab", "machine-1")
			input := []*fleet.Machine{chromeMachine1, chromeOSMachine1}
			resp, err := CreateMachines(ctx, input)
			So(err, ShouldBeNil)
			So(resp.Passed(), ShouldHaveLength, 2)
			So(resp.Failed(), ShouldHaveLength, 0)
			assertMachineEqual(resp.Passed()[0].Data.(*fleet.Machine), chromeMachine1)
			assertMachineEqual(resp.Passed()[1].Data.(*fleet.Machine), chromeOSMachine1)

			resp, err = GetAllMachines(ctx)
			So(err, ShouldBeNil)
			So(resp.Passed(), ShouldHaveLength, 2)
			So(resp.Failed(), ShouldHaveLength, 0)
			output := []*fleet.Machine{
				resp.Passed()[0].Data.(*fleet.Machine),
				resp.Passed()[1].Data.(*fleet.Machine),
			}
			wants := getMachineNames(input)
			gets := getMachineNames(output)
			So(wants, ShouldResemble, gets)
		})
	})
}
*/
