// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"

	ufspb "infra/unifiedfleet/api/v1/models"
	. "infra/unifiedfleet/app/model/datastore"
)

func mockVM(id string) *ufspb.VM {
	return &ufspb.VM{
		Name: id,
	}
}

func mockVMWithOwnership(id string, ownership *ufspb.OwnershipData) *ufspb.VM {
	machine := mockVM(id)
	machine.Ownership = ownership
	return machine
}

func assertVMWithOwnershipEqual(a *ufspb.VM, b *ufspb.VM) {
	if a.GetOwnership() == nil && b.GetOwnership() == nil {
		return
	}
	So(a.GetOwnership().PoolName, ShouldEqual, b.GetOwnership().PoolName)
	So(a.GetOwnership().SwarmingInstance, ShouldEqual, b.GetOwnership().SwarmingInstance)
	So(a.GetOwnership().Customer, ShouldEqual, b.GetOwnership().Customer)
	So(a.GetOwnership().SecurityLevel, ShouldEqual, b.GetOwnership().SecurityLevel)
}

func TestBatchUpdateVMs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	vm1 := mockVM("vm-1")
	vm2 := mockVM("vm-2")
	vm3 := mockVM("")
	Convey("Batch Update VM", t, func() {
		Convey("BatchUpdate all vms", func() {
			resp, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1, vm2})
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm1, vm2})
		})
		Convey("BatchUpdate existing vms", func() {
			vm2.MacAddress = "123"
			_, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1, vm2})
			So(err, ShouldBeNil)
			vm, err := GetVM(ctx, "vm-2")
			So(err, ShouldBeNil)
			So(vm.GetMacAddress(), ShouldEqual, "123")
		})
		Convey("BatchUpdate invalid vms", func() {
			resp, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1, vm2, vm3})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestUpdateVMOwnership(t *testing.T) {
	// Tests the ownership update scenarios for a VM
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
	vm1 := mockVM("vm-1")
	vm1_ownership := mockVMWithOwnership("vm-1", ownershipData)
	vm2 := mockVMWithOwnership("vm-1", ownershipData2)
	Convey("UpdateVM", t, func() {
		Convey("Update existing VM with ownership data", func() {
			resp, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm1})

			// Ownership data should be updated
			vmResp, err := UpdateVMOwnership(ctx, resp[0].Name, ownershipData)
			So(err, ShouldBeNil)
			So(vmResp.GetOwnership(), ShouldNotBeNil)
			assertVMWithOwnershipEqual(vmResp, vm1_ownership)

			// Regular Update calls should not override ownership data
			resp, err = BatchUpdateVMs(ctx, []*ufspb.VM{vm2})
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm2})

			vmResp, err = GetVM(ctx, "vm-1")
			So(err, ShouldBeNil)
			So(vmResp.GetOwnership(), ShouldNotBeNil)
			assertVMWithOwnershipEqual(vmResp, vm1_ownership)
		})
		Convey("Update non-existing VM with ownership", func() {
			resp, err := UpdateVMOwnership(ctx, "dummy", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update VM with ownership - invalid ID", func() {
			resp, err := UpdateVMOwnership(ctx, "", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestGetVM(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	Convey("GetVM", t, func() {
		Convey("Get machineLSE by non-existing ID", func() {
			resp, err := GetMachineLSE(ctx, "empty")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get machineLSE - invalid ID", func() {
			resp, err := GetMachineLSE(ctx, "")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestListVMs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	vm1 := &ufspb.VM{
		Name:          "vm-1",
		ResourceState: ufspb.State_STATE_DECOMMISSIONED,
	}
	vm2 := &ufspb.VM{
		Name: "vm-2",
		Tags: []string{"tag-1", "tag-2"},
	}
	vm3 := &ufspb.VM{
		Name:   "vm-3",
		Memory: 1234,
	}
	vm4 := mockVM("vm-4")
	vms := []*ufspb.VM{vm1, vm2, vm3, vm4}

	Convey("ListVMs", t, func() {
		_, err := BatchUpdateVMs(ctx, vms)
		So(err, ShouldBeNil)
		Convey("List vms - page_token invalid", func() {
			resp, nextPageToken, err := ListVMs(ctx, 5, 5, "abc", nil, false, nil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List vms - Full listing with no pagination", func() {
			resp, nextPageToken, err := ListVMs(ctx, 4, 4, "", nil, false, nil)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(resp, ShouldResembleProto, vms)
		})

		Convey("List vms - listing with pagination", func() {
			resp, nextPageToken, err := ListVMs(ctx, 3, 3, "", nil, false, nil)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, vms[:3])

			resp, _, err = ListVMs(ctx, 2, 2, nextPageToken, nil, false, nil)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, vms[3:])
		})
	})
	Convey("ListVMs with Filters", t, func() {
		_, err := BatchUpdateVMs(ctx, vms)
		So(err, ShouldBeNil)
		filterMap := make(map[string][]interface{})
		Convey("List vms - Filter by state", func() {
			filterMap["state"] = []interface{}{"STATE_DECOMMISSIONED"}
			resp, nextPageToken, err := ListVMs(ctx, 1, 2, "", filterMap, false, nil)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm1})
		})
		Convey("List vms - Filter by tags", func() {
			filterMap["tags"] = []interface{}{"tag-1"}
			resp, nextPageToken, err := ListVMs(ctx, 1, 2, "", filterMap, false, nil)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm2})
		})
		Convey("List vms - Filter by memory", func() {
			filterMap["memory"] = []interface{}{1234}
			resp, nextPageToken, err := ListVMs(ctx, 1, 2, "", filterMap, false, nil)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, []*ufspb.VM{vm3})
		})
	})
}

// TestListVMsByIdPrefixSearch tests the functionality for listing
// VMs by searching for name/id prefix
func TestListVMsByIdPrefixSearch(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	vm1 := &ufspb.VM{
		Name:          "vm-1",
		ResourceState: ufspb.State_STATE_DECOMMISSIONED,
	}
	vm2 := &ufspb.VM{
		Name: "vm-2",
		Tags: []string{"tag-1", "tag-2"},
	}
	vm3 := &ufspb.VM{
		Name:   "vm-3",
		Memory: 1234,
	}
	vm4 := mockVM("vm-4")
	vms := []*ufspb.VM{vm1, vm2, vm3, vm4}
	Convey("ListMachinesByIdPrefixSearch", t, func() {
		_, err := BatchUpdateVMs(ctx, vms)
		So(err, ShouldBeNil)
		Convey("List vms - page_token invalid", func() {
			resp, nextPageToken, err := ListVMsByIdPrefixSearch(ctx, 5, 2, "abc", "vm-", false, nil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List vms - Full listing with valid prefix and no pagination", func() {
			resp, nextPageToken, err := ListVMsByIdPrefixSearch(ctx, 4, 4, "", "vm-", false, nil)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, vms)
		})

		Convey("List vms - Full listing with invalid prefix", func() {
			resp, nextPageToken, err := ListVMsByIdPrefixSearch(ctx, 4, 2, "", "vm1-", false, nil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("List vms - listing with valid prefix and pagination", func() {
			resp, nextPageToken, err := ListVMsByIdPrefixSearch(ctx, 3, 3, "", "vm-", false, nil)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, vms[:3])

			resp, _, err = ListVMsByIdPrefixSearch(ctx, 2, 2, nextPageToken, "vm-", false, nil)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, vms[3:])
		})
	})
}

func TestDeleteVMs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	vm1 := mockVM("vm-delete1")
	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	vm1_ownership := mockVMWithOwnership("vm-delete1", ownershipData)
	Convey("DeleteVMs", t, func() {
		Convey("Delete VM by existing ID", func() {
			_, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
			So(err, ShouldBeNil)
			DeleteVMs(ctx, []string{"vm-delete1"})
			vm, err := GetVM(ctx, "vm-delete1")
			So(vm, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete vms by non-existing ID", func() {
			res := DeleteVMs(ctx, []string{"vm-delete2"})
			So(res.Failed(), ShouldHaveLength, 1)
		})
		Convey("Delete machineLSE - invalid ID", func() {
			res := DeleteVMs(ctx, []string{""})
			So(res.Failed(), ShouldHaveLength, 1)
		})
		Convey("Delete VM - with ownershipdata", func() {
			vmResp, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
			So(err, ShouldBeNil)

			// Ownership data should be updated
			resp, err := UpdateVMOwnership(ctx, vmResp[0].Name, ownershipData)
			So(err, ShouldBeNil)
			assertVMWithOwnershipEqual(resp, vm1_ownership)

			DeleteVMs(ctx, []string{"vm-delete1"})
			vm, err := GetVM(ctx, "vm-delete1")
			So(vm, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestQueryVMByPropertyName(t *testing.T) {
	t.Parallel()
	Convey("QueryVMByPropertyName", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		vm1 := mockVM("vm-queryByProperty1")
		vm1.MacAddress = "00:50:56:17:00:00"
		_, err := BatchUpdateVMs(ctx, []*ufspb.VM{vm1})
		So(err, ShouldBeNil)

		Convey("Query By existing mac address", func() {
			resp, err := QueryVMByPropertyName(ctx, "mac_address", "00:50:56:17:00:00", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 1)
			So(resp[0], ShouldResembleProto, vm1)
		})
		Convey("Query By non-existing mac address", func() {
			resp, err := QueryVMByPropertyName(ctx, "mac_address", "00:50:56:xx:yy:zz", false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
	})
}
