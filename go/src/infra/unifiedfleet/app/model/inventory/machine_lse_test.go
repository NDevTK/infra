// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

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

func mockMachineLSE(id string) *ufspb.MachineLSE {
	return &ufspb.MachineLSE{
		Name: id,
	}
}

func mockMachineLSEWithOwnership(id string, ownership *ufspb.OwnershipData) *ufspb.MachineLSE {
	machine := mockMachineLSE(id)
	machine.Ownership = ownership
	return machine
}

func TestCreateMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machineLSE1 := mockMachineLSE("machineLSE-1")
	machineLSE2 := mockMachineLSE("")

	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
		Customer:         "test-customer",
		SecurityLevel:    "test-security-level",
	}
	machineLSE3Ownership := mockMachineLSEWithOwnership("machineLSE-3", ownershipData)
	Convey("CreateMachineLSE", t, func() {
		Convey("Create new machineLSE", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
		})
		Convey("Create existing machineLSE", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, AlreadyExists)
		})
		Convey("Create machineLSE - invalid ID", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE2)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Create machineLSE with ownership data - ownership is not saved", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE3Ownership)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE3Ownership)
			So(resp.Ownership, ShouldBeNil)
		})
	})
}

func TestUpdateMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	machineLSE1 := mockMachineLSE("machineLSE-1")
	machineLSE2 := mockMachineLSE("machineLSE-1")
	machineLSE2.Hostname = "Linux Server"
	machineLSE3 := mockMachineLSE("machineLSE-3")
	machineLSE4 := mockMachineLSE("")
	Convey("UpdateMachineLSE", t, func() {
		Convey("Update existing machineLSE", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)

			resp, err = UpdateMachineLSE(ctx, machineLSE2)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE2)
		})
		Convey("Update non-existing machineLSE", func() {
			resp, err := UpdateMachineLSE(ctx, machineLSE3)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update machineLSE - invalid ID", func() {
			resp, err := UpdateMachineLSE(ctx, machineLSE4)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
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
	machineLSE1 := mockMachineLSEWithOwnership("machineLSE-1", ownershipData)
	machineLSE2 := mockMachineLSEWithOwnership("machineLSE-1", ownershipData2)

	Convey("UpdateMachine", t, func() {
		Convey("Update existing machine with ownership data", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)

			// Ownership data should be updated
			resp, err = UpdateMachineLSEOwnership(ctx, resp.Name, ownershipData)
			So(err, ShouldBeNil)
			So(resp.GetOwnership(), ShouldResembleProto, ownershipData)

			// Regular Update calls should not override ownership data
			resp, err = UpdateMachineLSE(ctx, machineLSE2)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE2)

			resp, err = GetMachineLSE(ctx, "machineLSE-1")
			So(err, ShouldBeNil)
			So(resp.GetOwnership(), ShouldResembleProto, ownershipData)
		})
		Convey("Update non-existing machine with ownership", func() {
			resp, err := UpdateMachineLSEOwnership(ctx, "dummy", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update machine with ownership - invalid ID", func() {
			resp, err := UpdateMachineLSEOwnership(ctx, "", ownershipData)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestGetMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	machineLSE1 := mockMachineLSE("machineLSE-1")
	Convey("GetMachineLSE", t, func() {
		Convey("Get machineLSE by existing ID", func() {
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
			resp, err = GetMachineLSE(ctx, "machineLSE-1")
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
		})
		Convey("Get machineLSE by non-existing ID", func() {
			resp, err := GetMachineLSE(ctx, "machineLSE-2")
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

func TestListMachineLSEs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machineLSEs := make([]*ufspb.MachineLSE, 0, 4)
	for i := 0; i < 4; i++ {
		machineLSE1 := mockMachineLSE(fmt.Sprintf("machineLSE-%d", i))
		resp, _ := CreateMachineLSE(ctx, machineLSE1)
		machineLSEs = append(machineLSEs, resp)
	}
	Convey("ListMachineLSEs", t, func() {
		Convey("List machineLSEs - page_token invalid", func() {
			resp, nextPageToken, err := ListMachineLSEs(ctx, 5, "abc", nil, false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List machineLSEs - Full listing with no pagination", func() {
			resp, nextPageToken, err := ListMachineLSEs(ctx, 4, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(resp, ShouldResembleProto, machineLSEs)
		})

		Convey("List machineLSEs - listing with pagination", func() {
			resp, nextPageToken, err := ListMachineLSEs(ctx, 3, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs[:3])

			resp, _, err = ListMachineLSEs(ctx, 2, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs[3:])
		})
	})
}

// TestListMachineLSEsByIdPrefixSearch tests the functionality for listing
// machineLSEs by searching for name/id prefix
func TestListMachineLSEsByIdPrefixSearch(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machineLSEs := make([]*ufspb.MachineLSE, 0, 4)
	for i := 0; i < 4; i++ {
		machineLSE1 := mockMachineLSE(fmt.Sprintf("machineLSE-%d", i))
		resp, _ := CreateMachineLSE(ctx, machineLSE1)
		machineLSEs = append(machineLSEs, resp)
	}
	Convey("ListMachinesByIdPrefixSearch", t, func() {
		Convey("List machines - page_token invalid", func() {
			resp, nextPageToken, err := ListMachineLSEsByIdPrefixSearch(ctx, 5, "abc", "machineLSE-", false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List machines - Full listing with valid prefix and no pagination", func() {
			resp, nextPageToken, err := ListMachineLSEsByIdPrefixSearch(ctx, 4, "", "machineLSE-", false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs)
		})

		Convey("List machines - Full listing with invalid prefix", func() {
			resp, nextPageToken, err := ListMachineLSEsByIdPrefixSearch(ctx, 4, "", "machineLSE1-", false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("List machines - listing with valid prefix and pagination", func() {
			resp, nextPageToken, err := ListMachineLSEsByIdPrefixSearch(ctx, 3, "", "machineLSE-", false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs[:3])

			resp, _, err = ListMachineLSEsByIdPrefixSearch(ctx, 2, nextPageToken, "machineLSE-", false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs[3:])
		})
	})
}

func TestDeleteMachineLSE(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	machineLSE1 := mockMachineLSE("machineLSE-1")
	ownershipData := &ufspb.OwnershipData{
		PoolName:         "pool1",
		SwarmingInstance: "test-swarming",
	}
	machineLSE2 := mockMachineLSEWithOwnership("machineLSE-2", ownershipData)
	Convey("DeleteMachineLSE", t, func() {
		Convey("Delete machineLSE by existing ID", func() {
			resp, cerr := CreateMachineLSE(ctx, machineLSE1)
			So(cerr, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
			err := DeleteMachineLSE(ctx, "machineLSE-1")
			So(err, ShouldBeNil)
			res, err := GetMachineLSE(ctx, "machineLSE-1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machineLSE by non-existing ID", func() {
			err := DeleteMachineLSE(ctx, "machineLSE-2")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete machineLSE - invalid ID", func() {
			err := DeleteMachineLSE(ctx, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
		Convey("Delete machineLSE - with ownershipdata", func() {
			resp, cerr := CreateMachineLSE(ctx, machineLSE2)
			So(cerr, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE2)

			// Ownership data should be updated
			resp, err := UpdateMachineLSEOwnership(ctx, resp.Name, ownershipData)
			So(err, ShouldBeNil)
			So(resp.GetOwnership(), ShouldResembleProto, ownershipData)

			err = DeleteMachineLSE(ctx, "machineLSE-2")
			So(err, ShouldBeNil)
			res, err := GetMachineLSE(ctx, "machineLSE-2")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}

func TestBatchUpdateMachineLSEs(t *testing.T) {
	t.Parallel()
	Convey("BatchUpdateMachineLSEs", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		machineLSEs := make([]*ufspb.MachineLSE, 0, 4)
		for i := 0; i < 4; i++ {
			machineLSE1 := mockMachineLSE(fmt.Sprintf("machineLSE-%d", i))
			resp, err := CreateMachineLSE(ctx, machineLSE1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSE1)
			machineLSEs = append(machineLSEs, resp)
		}
		Convey("BatchUpdate all machineLSEs", func() {
			resp, err := BatchUpdateMachineLSEs(ctx, machineLSEs)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs)
		})
		Convey("BatchUpdate existing and invalid machineLSEs", func() {
			machineLSE5 := mockMachineLSE("")
			machineLSEs = append(machineLSEs, machineLSE5)
			resp, err := BatchUpdateMachineLSEs(ctx, machineLSEs)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestQueryMachineLSEByPropertyName(t *testing.T) {
	t.Parallel()
	Convey("QueryMachineLSEByPropertyName", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		dummymachineLSE := &ufspb.MachineLSE{
			Name: "machineLSE-1",
		}
		machineLSE1 := &ufspb.MachineLSE{
			Name:                "machineLSE-1",
			Machines:            []string{"machine-1", "machine-2"},
			MachineLsePrototype: "machineLsePrototype-1",
			LogicalZone:         ufspb.LogicalZone_DRILLZONE_SFO36,
		}
		resp, cerr := CreateMachineLSE(ctx, machineLSE1)
		So(cerr, ShouldBeNil)
		So(resp, ShouldResembleProto, machineLSE1)

		machineLSEs := make([]*ufspb.MachineLSE, 0, 1)
		machineLSEs = append(machineLSEs, machineLSE1)

		dummymachineLSEs := make([]*ufspb.MachineLSE, 0, 1)
		dummymachineLSEs = append(dummymachineLSEs, dummymachineLSE)
		Convey("Query By existing Machine", func() {
			resp, err := QueryMachineLSEByPropertyName(ctx, "machine_ids", "machine-1", false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs)
		})
		Convey("Query By non-existing Machine", func() {
			resp, err := QueryMachineLSEByPropertyName(ctx, "machine_ids", "machine-5", false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Query By existing MachineLsePrototype keysonly", func() {
			resp, err := QueryMachineLSEByPropertyName(ctx, "machinelse_prototype_id", "machineLsePrototype-1", true)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, dummymachineLSEs)
		})
		Convey("Query By non-existing MachineLsePrototype", func() {
			resp, err := QueryMachineLSEByPropertyName(ctx, "machinelse_prototype_id", "machineLsePrototype-2", true)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Query By LogicalZone", func() {
			resp, err := QueryMachineLSEByPropertyName(ctx, "logical_zone", "DRILLZONE_SFO36", false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, machineLSEs)
		})
	})
}

func TestListAllMachineLSEs(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	machineLSEs := make([]*ufspb.MachineLSE, 0, 4)
	for i := 0; i < 4; i++ {
		machineLSE1 := mockMachineLSE(fmt.Sprintf("machineLSE-%d", i))
		machineLSE1.Description = "Test machineLSE"
		resp, _ := CreateMachineLSE(ctx, machineLSE1)
		machineLSEs = append(machineLSEs, resp)
	}
	Convey("ListAllMachineLSEs", t, func() {
		Convey("List all machineLSEs - keysOnly", func() {
			resp, _ := ListAllMachineLSEs(ctx, true)
			So(resp, ShouldNotBeNil)
			So(len(resp), ShouldEqual, 4)
			for i := 0; i < 4; i++ {
				So(resp[i].GetName(), ShouldEqual, fmt.Sprintf("machineLSE-%d", i))
				So(resp[i].GetDescription(), ShouldBeEmpty)
			}
		})

		Convey("List all machineLSEs", func() {
			resp, _ := ListAllMachineLSEs(ctx, false)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, machineLSEs)
		})
	})
}
