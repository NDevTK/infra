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
	ufsds "infra/unifiedfleet/internal/app/model/datastore"
)

func mockDutState(id string) *chromeosLab.DutState {
	return &chromeosLab.DutState{
		Id:       &chromeosLab.ChromeOSDeviceID{Value: id},
		Hostname: fmt.Sprintf("hostname-%s", id),
		Servo:    chromeosLab.PeripheralState_NOT_CONNECTED,
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
