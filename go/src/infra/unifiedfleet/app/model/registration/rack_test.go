// Copyright 2020 The Chromium Authors
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
	"infra/unifiedfleet/app/config"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/util"
)

func mockRack(id string, rackCapactiy int32, zone ufspb.Zone) *ufspb.Rack {
	return &ufspb.Rack{
		Name:       id,
		CapacityRu: rackCapactiy,
		Location: &ufspb.Location{
			Zone: zone,
		},
	}
}

func TestCreateRack(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	rack1 := mockRack("Rack-1", 5, ufspb.Zone_ZONE_CHROMEOS4)
	rack2 := mockRack("", 10, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("CreateRack", t, func() {
		Convey("Create new rack", func() {
			resp, err := CreateRack(ctx, rack1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rack1)
		})
		Convey("Create existing rack", func() {
			resp, err := CreateRack(ctx, rack1)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, AlreadyExists)
		})
		Convey("Create rack - invalid ID", func() {
			resp, err := CreateRack(ctx, rack2)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestUpdateRack(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	rack1 := mockRack("Rack-1", 5, ufspb.Zone_ZONE_CHROMEOS4)
	rack2 := mockRack("Rack-1", 10, ufspb.Zone_ZONE_CHROMEOS4)
	rack3 := mockRack("Rack-3", 15, ufspb.Zone_ZONE_CHROMEOS4)
	rack4 := mockRack("", 20, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("UpdateRack", t, func() {
		Convey("Update existing rack", func() {
			resp, err := CreateRack(ctx, rack1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rack1)

			resp, err = UpdateRack(ctx, rack2)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rack2)
		})
		Convey("Update non-existing rack", func() {
			resp, err := UpdateRack(ctx, rack3)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Update rack - invalid ID", func() {
			resp, err := UpdateRack(ctx, rack4)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestGetRack(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	rack1 := mockRack("Rack-1", 5, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("GetRack", t, func() {
		Convey("Get rack by existing ID", func() {
			resp, err := CreateRack(ctx, rack1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rack1)
			resp, err = GetRack(ctx, "Rack-1")
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, rack1)
		})
		Convey("Get rack by non-existing ID", func() {
			resp, err := GetRack(ctx, "rack-2")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Get rack - invalid ID", func() {
			resp, err := GetRack(ctx, "")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestGetRackACL(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	ctx = config.Use(ctx, &config.Config{
		ExperimentalAPI: &config.ExperimentalAPI{
			GetRackACL: 99,
		},
	})

	// realm "@internal:ufs/os-acs"
	rack := mockRack("rack-123", 100, ufspb.Zone_ZONE_CHROMEOS5)
	_, err := CreateRack(ctx, rack)
	if err != nil {
		t.Errorf("failed to create rack: %s", err)
	}

	Convey("When a rack is created in a certain realm", t, func() {

		Convey("No user is rejected", func() {
			resp, err := getRackACL(ctx, "rack-123")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Internal")
			So(resp, ShouldBeNil)
		})
		Convey("A user without perms is rejected", func() {
			userCtx := mockUser(ctx, "email@google.com")
			resp, err := getRackACL(userCtx, "rack-123")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Permission")
			So(resp, ShouldBeNil)
		})
		Convey("A user without the correct perm is rejected", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.ConfigurationsGet)
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.RegistrationsCreate)
			resp, err := getRackACL(userCtx, "rack-123")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Permission")
			So(resp, ShouldBeNil)
		})
		Convey("A user without the correct realm is rejected", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.SatLabInternalUserRealm, util.RegistrationsGet)
			resp, err := getRackACL(userCtx, "rack-123")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Permission")
			So(resp, ShouldBeNil)
		})
		Convey("A user with the correct realm and permission is accepted", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.RegistrationsGet)
			resp, err := getRackACL(userCtx, "rack-123")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, rack)
		})
	})
}

func TestBatchGetRackACL(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	ctx = config.Use(ctx, &config.Config{
		ExperimentalAPI: &config.ExperimentalAPI{
			GetRackACL: 99,
		},
	})

	// realm "@internal:ufs/os-acs"
	rack1 := mockRack("rack-1", 100, ufspb.Zone_ZONE_CHROMEOS5)
	_, err := CreateRack(ctx, rack1)
	if err != nil {
		t.Errorf("failed to create rack: %s", err)
	}

	// realm "@internal:ufs/os-atl"
	rack2 := mockRack("rack-2", 100, ufspb.Zone_ZONE_CHROMEOS4)
	_, err = CreateRack(ctx, rack2)
	if err != nil {
		t.Errorf("failed to create rack: %s", err)
	}

	Convey("When two racks are created", t, func() {
		Convey("No user is rejected", func() {
			resp, err := BatchGetRacksACL(ctx, []string{"rack-1", "rack-2"})
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Internal")
			So(resp, ShouldBeNil)
		})
		Convey("A user without perms is rejected", func() {
			userCtx := mockUser(ctx, "email@google.com")
			resp, err := BatchGetRacksACL(userCtx, []string{"rack-1", "rack-2"})
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Permission")
			So(resp, ShouldBeNil)
		})
		Convey("A user requesting racks without for at least one is rejected", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.RegistrationsGet)
			resp, err := BatchGetRacksACL(userCtx, []string{"rack-1", "rack-2"})
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Permission")
			So(resp, ShouldBeNil)
		})
		Convey("A user requesting only racks they can access succeeds", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.RegistrationsGet)
			resp, err := BatchGetRacksACL(userCtx, []string{"rack-1"})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, []*ufspb.Rack{rack1})
		})
		Convey("A user with all realm perms can see racks in multiple realms", func() {
			userCtx := mockUser(ctx, "email@google.com")
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.RegistrationsGet)
			mockRealmPerms(userCtx, util.AtlLabAdminRealm, util.RegistrationsGet)
			resp, err := BatchGetRacksACL(userCtx, []string{"rack-1", "rack-2"})
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, []*ufspb.Rack{rack1, rack2})
		})
	})
}

func TestListRacks(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	racks := make([]*ufspb.Rack, 0, 4)
	for i := 0; i < 4; i++ {
		rack1 := mockRack(fmt.Sprintf("rack-%d", i), 5, ufspb.Zone_ZONE_CHROMEOS4)
		resp, _ := CreateRack(ctx, rack1)
		racks = append(racks, resp)
	}
	Convey("ListRacks", t, func() {
		Convey("List racks - page_token invalid", func() {
			resp, nextPageToken, err := ListRacks(ctx, 5, "abc", nil, false)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InvalidPageToken)
		})

		Convey("List racks - Full listing with no pagination", func() {
			resp, nextPageToken, err := ListRacks(ctx, 4, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks)
		})

		Convey("List racks - listing with pagination", func() {
			resp, nextPageToken, err := ListRacks(ctx, 3, "", nil, false)
			So(resp, ShouldNotBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks[:3])

			resp, _, err = ListRacks(ctx, 2, nextPageToken, nil, false)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks[3:])
		})
	})
}

func TestListRacksACL(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	racks := make([]*ufspb.Rack, 0, 8)
	for i := 0; i < 4; i++ {
		rack := mockRack(fmt.Sprintf("rack-1%d", i), 4, ufspb.Zone_ZONE_CHROMEOS4)
		resp, _ := CreateRack(ctx, rack)
		racks = append(racks, resp)
	}
	for i := 0; i < 4; i++ {
		rack := mockRack(fmt.Sprintf("rack-2%d", i), 4, ufspb.Zone_ZONE_SFO36_BROWSER)
		resp, _ := CreateRack(ctx, rack)
		racks = append(racks, resp)
	}

	noPermUserCtx := mockUser(ctx, "none@google.com")

	somePermUserCtx := mockUser(ctx, "some@google.com")
	mockRealmPerms(somePermUserCtx, util.BrowserLabAdminRealm, util.RegistrationsList)

	allPermUserCtx := mockUser(ctx, "all@google.com")
	mockRealmPerms(allPermUserCtx, util.BrowserLabAdminRealm, util.RegistrationsList)
	mockRealmPerms(allPermUserCtx, util.AtlLabAdminRealm, util.RegistrationsList)

	Convey("ListRacks", t, func() {
		Convey("List racks - anonymous call rejected", func() {
			resp, nextPageToken, err := ListRacksACL(ctx, 100, "", nil, false)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List racks - filter on realm rejected", func() {
			resp, nextPageToken, err := ListRacksACL(allPermUserCtx, 100, "", map[string][]interface{}{"realm": nil}, false)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List racks - happy path with no perms returns no results", func() {
			resp, nextPageToken, err := ListRacksACL(noPermUserCtx, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List racks - happy path with partial perms returns partial results", func() {
			resp, nextPageToken, err := ListRacksACL(somePermUserCtx, 2, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks[4:6])
			So(nextPageToken, ShouldNotBeEmpty)

			resp2, nextPageToken2, err2 := ListRacksACL(somePermUserCtx, 100, nextPageToken, nil, false)
			So(err2, ShouldBeNil)
			So(resp2, ShouldResembleProto, racks[6:])
			So(nextPageToken2, ShouldBeEmpty)
		})
		Convey("List racks - happy path with all perms returns all results", func() {
			resp, nextPageToken, err := ListRacksACL(allPermUserCtx, 4, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks[:4])
			So(nextPageToken, ShouldNotBeEmpty)

			resp2, nextPageToken2, err2 := ListRacksACL(allPermUserCtx, 100, nextPageToken, nil, false)
			So(err2, ShouldBeNil)
			So(resp2, ShouldResembleProto, racks[4:])
			So(nextPageToken2, ShouldBeEmpty)
		})
		Convey("List racks - happy path with all perms and filters returns filtered results", func() {
			resp, nextPageToken, err := ListRacksACL(allPermUserCtx, 100, "", map[string][]interface{}{"zone": {"ZONE_CHROMEOS4"}}, false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, racks[:4])
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List racks - happy path with all perms and filters with no matches returns no results", func() {
			resp, nextPageToken, err := ListRacksACL(allPermUserCtx, 100, "", map[string][]interface{}{"zone": {"fake"}}, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
	})
}

func TestDeleteRack(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	rack1 := mockRack("rack-1", 5, ufspb.Zone_ZONE_CHROMEOS4)
	Convey("DeleteRack", t, func() {
		Convey("Delete rack by existing ID", func() {
			resp, cerr := CreateRack(ctx, rack1)
			So(cerr, ShouldBeNil)
			So(resp, ShouldResembleProto, rack1)
			err := DeleteRack(ctx, "rack-1")
			So(err, ShouldBeNil)
			res, err := GetRack(ctx, "rack-1")
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete rack by non-existing ID", func() {
			err := DeleteRack(ctx, "rack-2")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
		Convey("Delete rack - invalid ID", func() {
			err := DeleteRack(ctx, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestBatchUpdateRacks(t *testing.T) {
	t.Parallel()
	Convey("BatchUpdateRacks", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		Racks := make([]*ufspb.Rack, 0, 4)
		for i := 0; i < 4; i++ {
			Rack1 := mockRack(fmt.Sprintf("Rack-%d", i), 10, ufspb.Zone_ZONE_CHROMEOS4)
			resp, err := CreateRack(ctx, Rack1)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, Rack1)
			Racks = append(Racks, resp)
		}
		Convey("BatchUpdate all Racks", func() {
			resp, err := BatchUpdateRacks(ctx, Racks)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, Racks)
		})
		Convey("BatchUpdate existing and invalid Racks", func() {
			Rack5 := mockRack("", 10, ufspb.Zone_ZONE_CHROMEOS4)
			Racks = append(Racks, Rack5)
			resp, err := BatchUpdateRacks(ctx, Racks)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, InternalError)
		})
	})
}

func TestQueryRackByPropertyName(t *testing.T) {
	t.Parallel()
	Convey("QueryRackByPropertyName", t, func() {
		ctx := gaetesting.TestingContextWithAppID("go-test")
		datastore.GetTestable(ctx).Consistent(true)
		dummyRack := &ufspb.Rack{
			Name: "Rack-1",
		}
		Rack1 := &ufspb.Rack{
			Name: "Rack-1",
			Rack: &ufspb.Rack_ChromeBrowserRack{
				ChromeBrowserRack: &ufspb.ChromeBrowserRack{},
			},
			Tags: []string{"tag-1"},
		}
		resp, cerr := CreateRack(ctx, Rack1)
		So(cerr, ShouldBeNil)
		So(resp, ShouldResembleProto, Rack1)

		Racks := make([]*ufspb.Rack, 0, 1)
		Racks = append(Racks, Rack1)

		dummyRacks := make([]*ufspb.Rack, 0, 1)
		dummyRacks = append(dummyRacks, dummyRack)
		Convey("Query By existing Rack", func() {
			resp, err := QueryRackByPropertyName(ctx, "tags", "tag-1", false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, Racks)
		})
		Convey("Query By non-existing Rack", func() {
			resp, err := QueryRackByPropertyName(ctx, "tags", "tag-5", false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
		Convey("Query By existing RackPrototype keysonly", func() {
			resp, err := QueryRackByPropertyName(ctx, "tags", "tag-1", true)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, dummyRacks)
		})
		Convey("Query By non-existing RackPrototype", func() {
			resp, err := QueryRackByPropertyName(ctx, "tags", "tag-5", true)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
		})
	})
}
