// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"

	chromeosLab "infra/unifiedfleet/api/v1/models/chromeos/lab"
	"infra/unifiedfleet/app/config"
	ufsds "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/util"
)

func mockDutState(id string) *chromeosLab.DutState {
	return &chromeosLab.DutState{
		Id:       &chromeosLab.ChromeOSDeviceID{Value: id},
		Hostname: fmt.Sprintf("hostname-%s", id),
		Servo:    chromeosLab.PeripheralState_NOT_CONNECTED,
	}
}

func mockDutStateWithRealm(id string, realm string) *chromeosLab.DutState {
	return &chromeosLab.DutState{
		Id:       &chromeosLab.ChromeOSDeviceID{Value: id},
		Hostname: fmt.Sprintf("hostname-%s", id),
		Servo:    chromeosLab.PeripheralState_NOT_CONNECTED,
		Realm:    realm,
	}
}

func TestUpdateDutState(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	Convey("UpdateDutState", t, func() {
		Convey("Update existing dut state", func() {
			dutState1 := mockDutState("existing-dut-id")
			resp, err := UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1})
			So(err, ShouldBeNil)
			So(resp[0], ShouldResembleProto, dutState1)

			dutState1.Servo = chromeosLab.PeripheralState_BAD_RIBBON_CABLE
			resp, err = UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1})
			So(err, ShouldBeNil)
			So(resp[0], ShouldResembleProto, dutState1)

			getRes, err := GetDutState(ctx, "existing-dut-id")
			So(err, ShouldBeNil)
			So(getRes, ShouldResembleProto, dutState1)
		})
		Convey("Update non-existing dut state", func() {
			dutState1 := mockDutState("non-existing-dut-id")
			resp, err := UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1})
			So(resp[0], ShouldResembleProto, dutState1)
			So(err, ShouldBeNil)

			getRes, err := GetDutState(ctx, "non-existing-dut-id")
			So(err, ShouldBeNil)
			So(getRes, ShouldResembleProto, dutState1)
		})
		Convey("Update dut state - invalid ID", func() {
			dutState1 := mockDutState("")
			resp, err := UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1})
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, ufsds.InternalError)
		})
	})
}

func TestDeleteDutState(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	Convey("DeleteDutStates", t, func() {
		Convey("Delete dut state by existing ID", func() {
			dutState1 := mockDutState("delete-dut-id1")
			dutState2 := mockDutState("delete-dut-id2")
			_, err := UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1, dutState2})
			So(err, ShouldBeNil)

			resp, err := GetAllDutStates(ctx)
			So(err, ShouldBeNil)
			So(resp.Passed(), ShouldHaveLength, 2)

			resp2 := DeleteDutStates(ctx, []string{"delete-dut-id2"})
			So(resp2.Passed(), ShouldHaveLength, 1)

			resp, err = GetAllDutStates(ctx)
			So(err, ShouldBeNil)
			So(resp.Passed(), ShouldHaveLength, 1)
			So(resp.Passed()[0].Data.(*chromeosLab.DutState).GetId().GetValue(), ShouldEqual, "delete-dut-id1")
			So(resp.Passed()[0].Data.(*chromeosLab.DutState).GetHostname(), ShouldEqual, "hostname-delete-dut-id1")
		})

		Convey("Delete dut state by non-existing ID", func() {
			resp := DeleteDutStates(ctx, []string{"delete-dut-non-existing-id"})
			So(resp.Failed(), ShouldHaveLength, 1)
		})
	})
}

func TestGetDutStateACL(t *testing.T) {
	t.Parallel()
	// manually turn on config
	alwaysUseACLConfig := config.Config{
		ExperimentalAPI: &config.ExperimentalAPI{
			GetDutStateACL: 99,
		},
	}

	ctx := gaetesting.TestingContextWithAppID("go-test")
	ctx = config.Use(ctx, &alwaysUseACLConfig)

	dutState1 := mockDutStateWithRealm("dut-state-1", util.BrowserLabAdminRealm)
	dutState2 := mockDutStateWithRealm("dut-state-2", util.SatLabInternalUserRealm)
	_, err := UpdateDutStates(ctx, []*chromeosLab.DutState{dutState1, dutState2})
	if err != nil {
		fmt.Println("Not able to instantiate DutStates in TestGetDutStateACL")
	}

	Convey("GetDutStateACL", t, func() {
		Convey("GetDutStateACL - no user", func() {
			resp, err := GetDutStateACL(ctx, "dut-state-1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "Internal")
			So(resp, ShouldBeNil)
		})
		Convey("GetDutStateACL - no perms", func() {
			userCtx := mockUser(ctx, "nombre@chromium.org")
			resp, err := GetDutStateACL(userCtx, "dut-state-1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "PermissionDenied")
			So(resp, ShouldBeNil)
		})
		Convey("GetDutStateACL - missing perms", func() {
			userCtx := mockUser(ctx, "nombre@chromium.org")
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.RegistrationsList)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.RegistrationsGet)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.RegistrationsDelete)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.RegistrationsCreate)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.InventoriesList)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.InventoriesDelete)
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.InventoriesCreate)
			resp, err := GetDutStateACL(userCtx, "dut-state-1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "PermissionDenied")
			So(resp, ShouldBeNil)
		})
		Convey("GetDutStateACL - missing realms", func() {
			userCtx := mockUser(ctx, "nombre@chromium.org")
			mockRealmPerms(userCtx, util.AtlLabAdminRealm, util.ConfigurationsGet)
			mockRealmPerms(userCtx, util.AtlLabChromiumAdminRealm, util.ConfigurationsGet)
			mockRealmPerms(userCtx, util.AcsLabAdminRealm, util.ConfigurationsGet)
			mockRealmPerms(userCtx, util.AtlLabAdminRealm, util.ConfigurationsGet)
			mockRealmPerms(userCtx, util.SatLabInternalUserRealm, util.ConfigurationsGet)
			resp, err := GetDutStateACL(userCtx, "dut-state-1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "PermissionDenied")
			So(resp, ShouldBeNil)
			// 2nd dut-state of different realm
			user2Ctx := mockUser(ctx, "name@chromium.org")
			mockRealmPerms(user2Ctx, util.BrowserLabAdminRealm, util.ConfigurationsGet)
			resp, err = GetDutStateACL(user2Ctx, "dut-state-2")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "PermissionDenied")
			So(resp, ShouldBeNil)
		})
		Convey("GetDutStateACL - happy path", func() {
			userCtx := mockUser(ctx, "nombre@chromium.org")
			mockRealmPerms(userCtx, util.BrowserLabAdminRealm, util.ConfigurationsGet)
			resp, err := GetDutStateACL(userCtx, "dut-state-1")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, dutState1)
			// 2nd dut-state of different realm
			user2Ctx := mockUser(ctx, "name@chromium.org")
			mockRealmPerms(user2Ctx, util.SatLabInternalUserRealm, util.ConfigurationsGet)
			resp, err = GetDutStateACL(user2Ctx, "dut-state-2")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, dutState2)
		})
	})
}
func TestListDutStatesACL(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)
	dutStates := make([]*chromeosLab.DutState, 0, 8)
	for i := 0; i < 4; i++ {
		broswerDutState := mockDutStateWithRealm(fmt.Sprintf("dut-state-%d", i), util.BrowserLabAdminRealm)
		dutStates = append(dutStates, broswerDutState)
	}
	for i := 0; i < 4; i++ {
		satlabDutState := mockDutStateWithRealm(fmt.Sprintf("dut-state-%d", i+4), util.SatLabInternalUserRealm)
		dutStates = append(dutStates, satlabDutState)
	}
	_, err := UpdateDutStates(ctx, dutStates)
	if err != nil {
		fmt.Println("Not able to instantiate DutStates in TestGetDutStateACL")
	}

	noPermUserCtx := mockUser(ctx, "none@google.com")

	somePermUserCtx := mockUser(ctx, "some@google.com")
	mockRealmPerms(somePermUserCtx, util.BrowserLabAdminRealm, util.ConfigurationsList)

	allPermUserCtx := mockUser(ctx, "all@google.com")
	mockRealmPerms(allPermUserCtx, util.BrowserLabAdminRealm, util.ConfigurationsList)
	mockRealmPerms(allPermUserCtx, util.SatLabInternalUserRealm, util.ConfigurationsList)
	Convey("ListDutStates", t, func() {
		Convey("List DutStates - anonymous call rejected", func() {
			resp, nextPageToken, err := ListDutStatesACL(ctx, 100, "", nil, false)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List DutStates - filter on realm rejected", func() {
			resp, nextPageToken, err := ListDutStatesACL(allPermUserCtx, 100, "", map[string][]interface{}{"realm": nil}, false)
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List DutStates - happy path with no perms returns no results", func() {
			resp, nextPageToken, err := ListDutStatesACL(noPermUserCtx, 100, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
		Convey("List DutStates - happy path with partial perms returns partial results", func() {
			resp, nextPageToken, err := ListDutStatesACL(somePermUserCtx, 2, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, dutStates[:2])
			So(nextPageToken, ShouldNotBeEmpty)

			resp2, nextPageToken2, err2 := ListDutStatesACL(somePermUserCtx, 100, nextPageToken, nil, false)
			So(err2, ShouldBeNil)
			So(resp2, ShouldResembleProto, dutStates[2:4])
			So(nextPageToken2, ShouldBeEmpty)
		})
		Convey("List DutStates - happy path with all perms returns all results", func() {
			resp, nextPageToken, err := ListDutStatesACL(allPermUserCtx, 4, "", nil, false)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, dutStates[:4])
			So(nextPageToken, ShouldNotBeEmpty)

			resp2, nextPageToken2, err2 := ListDutStatesACL(allPermUserCtx, 100, nextPageToken, nil, false)
			So(err2, ShouldBeNil)
			So(resp2, ShouldResembleProto, dutStates[4:])
			So(nextPageToken2, ShouldBeEmpty)
		})
		Convey("List DutStates - happy path with all perms and filters with no matches returns no results", func() {
			resp, nextPageToken, err := ListDutStatesACL(allPermUserCtx, 100, "", map[string][]interface{}{"hostname": {"fake"}}, false)
			So(err, ShouldBeNil)
			So(resp, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
		})
	})
}
