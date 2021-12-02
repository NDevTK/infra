// Copyright 2018 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inventory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/retry"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	fleet "infra/appengine/crosskylabadmin/api/fleet/v1"
	"infra/appengine/crosskylabadmin/internal/app/config"
	"infra/appengine/crosskylabadmin/internal/app/frontend/internal/datastore/dronecfg"
	dsinventory "infra/appengine/crosskylabadmin/internal/app/frontend/internal/datastore/inventory"
	dssv "infra/appengine/crosskylabadmin/internal/app/frontend/internal/datastore/stableversion"
	"infra/appengine/crosskylabadmin/internal/app/gitstore/fakes"
	"infra/libs/skylab/inventory"
)

func TestGetDutInfoWithConsistentDatastoreAndSplitInventory(t *testing.T) {
	Convey("On happy path and 3 DUTs in the inventory", t, func() {
		ctx := testingContext()
		ctx = withSplitInventory(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		setSplitGitilesDuts(tf.C, tf.FakeGitiles, []testInventoryDut{
			{id: "dut1_id", hostname: "jetstream-host", model: "link", pool: "DUT_POOL_SUITES"},
			{id: "dut2_id", hostname: "jetstream-host", model: "peppy", pool: "DUT_POOL_SUITES"},
			{id: "dut3_id", hostname: "chromeos15-rack1-row2-host3", model: "link", pool: "DUT_POOL_SUITES"},
		})

		Convey("initial GetDutInfo (by Id) returns NotFound", func() {
			_, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "dut1_id"})
			So(status.Code(err), ShouldEqual, codes.NotFound)
		})

		Convey("initial GetDutInfo (by Hostname) returns NotFound", func() {
			_, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Hostname: "jetstream-host"})
			So(status.Code(err), ShouldEqual, codes.NotFound)
		})

		Convey("after a call to UpdateCachedInventory", func() {
			_, err := tf.Inventory.UpdateCachedInventory(tf.C, &fleet.UpdateCachedInventoryRequest{})
			So(err, ShouldBeNil)

			Convey("Dut with same hostname will be overwritten", func() {
				resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "dut1_id"})
				So(status.Code(err), ShouldEqual, codes.NotFound)

				resp, err = tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "dut2_id"})
				So(err, ShouldBeNil)
				dut := getDutInfo(t, resp)
				So(dut.GetCommon().GetId(), ShouldEqual, "dut2_id")
				So(dut.GetCommon().GetHostname(), ShouldEqual, "jetstream-host")
			})

			Convey("GetDutInfo (by ID) returns the DUT", func() {
				resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "dut3_id"})
				So(err, ShouldBeNil)
				dut := getDutInfo(t, resp)
				So(dut.GetCommon().GetId(), ShouldEqual, "dut3_id")
				So(dut.GetCommon().GetHostname(), ShouldEqual, "chromeos15-rack1-row2-host3")
			})

			Convey("GetDutInfo (by Hostname) returns the DUT", func() {
				resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Hostname: "jetstream-host"})
				So(err, ShouldBeNil)
				dut := getDutInfo(t, resp)
				So(dut.GetCommon().GetId(), ShouldEqual, "dut2_id")
				So(dut.GetCommon().GetHostname(), ShouldEqual, "jetstream-host")
			})
		})
	})
}

func TestGetDutInfoWithEventuallyConsistentDatastoreAndSplitInventory(t *testing.T) {
	Convey("With eventually consistent datastore and a single DUT in the inventory", t, func() {
		ctx := testingContext()
		ctx = withSplitInventory(ctx)
		ctx = withDutInfoCacheValidity(ctx, 100*time.Second)
		datastore.GetTestable(ctx).Consistent(false)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		setSplitGitilesDuts(tf.C, tf.FakeGitiles, []testInventoryDut{
			{id: "dut1_id", hostname: "jetstream-host", model: "link", pool: "DUT_POOL_SUITES"},
		})

		Convey("after a call to UpdateCachedInventory", func() {
			_, err := tf.Inventory.UpdateCachedInventory(tf.C, &fleet.UpdateCachedInventoryRequest{})
			So(err, ShouldBeNil)

			Convey("GetDutInfo (by ID) returns the DUT", func() {
				resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "dut1_id"})
				So(err, ShouldBeNil)
				dut := getDutInfo(t, resp)
				So(dut.GetCommon().GetId(), ShouldEqual, "dut1_id")
				So(dut.GetCommon().GetHostname(), ShouldEqual, "jetstream-host")
			})

			Convey("GetDutInfo (by Hostname) returns NotFound", func() {
				_, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "jetstream-host"})
				So(status.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey("after index update, GetDutInfo (by Hostname) returns the DUT", func() {
				datastore.GetTestable(ctx).CatchupIndexes()
				resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Hostname: "jetstream-host"})
				So(err, ShouldBeNil)
				dut := getDutInfo(t, resp)
				So(dut.GetCommon().GetId(), ShouldEqual, "dut1_id")
				So(dut.GetCommon().GetHostname(), ShouldEqual, "jetstream-host")

				Convey("after a Hostname update, GetDutInfo (by Hostname) returns NotFound", func() {
					setSplitGitilesDuts(tf.C, tf.FakeGitiles, []testInventoryDut{
						{id: "dut1_id", hostname: "jetstream-host-2", model: "link", pool: "DUT_POOL_SUITES"},
					})
					_, err := tf.Inventory.UpdateCachedInventory(tf.C, &fleet.UpdateCachedInventoryRequest{})
					So(err, ShouldBeNil)

					_, err = tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Id: "jetstream-host"})
					So(status.Code(err), ShouldEqual, codes.NotFound)

					Convey("after index update, GetDutInfo (by Hostname) returns the DUT for the new Hostname", func() {
						datastore.GetTestable(ctx).CatchupIndexes()
						resp, err := tf.Inventory.GetDutInfo(tf.C, &fleet.GetDutInfoRequest{Hostname: "jetstream-host-2"})
						So(err, ShouldBeNil)
						dut := getDutInfo(t, resp)
						So(dut.GetCommon().GetId(), ShouldEqual, "dut1_id")
						So(dut.GetCommon().GetHostname(), ShouldEqual, "jetstream-host-2")
					})
				})
			})
		})
	})
}

func TestInvalidDutID(t *testing.T) {
	Convey("DutID with empty hostname won't go to drone config datastore", t, func() {
		ctx := testingContext()
		ctx = withDutInfoCacheValidity(ctx, 100*time.Minute)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		err := tf.FakeGitiles.SetInventory(config.Get(tf.C).Inventory, fakes.InventoryData{
			Lab: inventoryBytesFromDUTs([]testInventoryDut{
				{"dut1_id", "dut1_hostname", "link", "DUT_POOL_SUITES"},
			}),
			Infrastructure: inventoryBytesFromServers([]testInventoryServer{
				{
					hostname:    "fake-drone.google.com",
					environment: inventory.Environment_ENVIRONMENT_STAGING,
					dutIDs:      []string{"dut1_id", "empty_id"},
				},
			}),
		})
		So(err, ShouldBeNil)

		_, err = tf.Inventory.UpdateCachedInventory(tf.C, &fleet.UpdateCachedInventoryRequest{})
		So(err, ShouldBeNil)
		e, err := dronecfg.Get(tf.C, "fake-drone.google.com")
		So(err, ShouldBeNil)
		So(e.DUTs, ShouldHaveLength, 1)
		duts := make([]string, len(e.DUTs))
		for i, d := range e.DUTs {
			duts[i] = d.Hostname
		}
		So(duts, ShouldResemble, []string{"dut1_hostname"})
	})
}

var testLooksLikeFakeServoTests = []struct {
	in   string
	good bool
}{
	{``, false},
	{`dummy_host`, false},
	{`FAKE_SERVO_HOST`, false},
	{`chromeos6-row3-rack11-labstation`, true},
}

func TestLooksLikeFakeServo(t *testing.T) {
	for _, tt := range testLooksLikeFakeServoTests {
		name := fmt.Sprintf("(%s)", tt.in)
		t.Run(name, func(t *testing.T) {
			good := !looksLikeFakeServo(tt.in)
			if good != tt.good {
				t.Errorf("wanted: (%t) got: (%t)", tt.good, good)
			}
		})
	}
}

func TestGetStableVersion(t *testing.T) {
	Convey("Test GetStableVersion RPC -- stable versions exist", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		err := dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)
		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				BuildTarget: "xxx-build-target",
				Model:       "xxx-model",
			},
		)
		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-cros-version")
		So(resp.FaftVersion, ShouldEqual, "xxx-faft-version")
		So(resp.FirmwareVersion, ShouldEqual, "xxx-firmware-version")
	})

	Convey("Test GetStableVersion RPC -- look up by hostname beaglebone", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		// use a fake beaglebone servo
		duts := []*inventory.DeviceUnderTest{
			{
				Common: &inventory.CommonDeviceSpecs{
					Attributes: []*inventory.KeyValue{
						{
							Key:   strptr("servo_host"),
							Value: strptr("xxx-beaglebone-servo"),
						},
					},
					Id:       strptr("xxx-id"),
					Hostname: strptr("xxx-hostname"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-model"),
						Board: strptr("xxx-build-target"),
					},
				},
			},
		}

		err := dsinventory.UpdateDUTs(ctx, duts)
		So(err, ShouldBeNil)

		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleCrosStableVersion(ctx, beagleboneServo, beagleboneServo, "xxx-beaglebone-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)

		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				Hostname: "xxx-hostname",
			},
		)

		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-cros-version")
		So(resp.FaftVersion, ShouldEqual, "xxx-faft-version")
		So(resp.FirmwareVersion, ShouldEqual, "xxx-firmware-version")
		So(resp.ServoCrosVersion, ShouldEqual, "xxx-beaglebone-cros-version")
	})

	Convey("Test GetStableVersion RPC -- look up by hostname labstation", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		// use a fake labstation
		duts := []*inventory.DeviceUnderTest{
			{
				Common: &inventory.CommonDeviceSpecs{
					Attributes: []*inventory.KeyValue{
						{
							Key:   strptr("servo_host"),
							Value: strptr("xxx-labstation"),
						},
					},
					Id:       strptr("xxx-id"),
					Hostname: strptr("xxx-hostname"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-model"),
						Board: strptr("xxx-build-target"),
					},
				},
			},
			{
				Common: &inventory.CommonDeviceSpecs{
					Id:       strptr("xxx-labstation-id"),
					Hostname: strptr("xxx-labstation"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-labstation-model"),
						Board: strptr("xxx-labstation-board"),
					},
				},
			},
		}

		err := dsinventory.UpdateDUTs(ctx, duts)
		So(err, ShouldBeNil)

		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-labstation-board", "xxx-labstation-model", "xxx-labstation-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)

		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				Hostname: "xxx-hostname",
			},
		)

		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-cros-version")
		So(resp.FaftVersion, ShouldEqual, "xxx-faft-version")
		So(resp.FirmwareVersion, ShouldEqual, "xxx-firmware-version")
		So(resp.ServoCrosVersion, ShouldEqual, "xxx-labstation-cros-version")
	})

	Convey("Test GetStableVersion RPC -- look up labstation proper", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		// use a fake labstation
		duts := []*inventory.DeviceUnderTest{
			{
				Common: &inventory.CommonDeviceSpecs{
					Attributes: []*inventory.KeyValue{
						{
							Key:   strptr("servo_host"),
							Value: strptr("xxx-labstation"),
						},
					},
					Id:       strptr("xxx-id"),
					Hostname: strptr("xxx-hostname"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-model"),
						Board: strptr("xxx-build-target"),
					},
				},
			},
			{
				Common: &inventory.CommonDeviceSpecs{
					Id:       strptr("xxx-labstation-id"),
					Hostname: strptr("xxx-labstation"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-labstation-model"),
						Board: strptr("xxx-labstation-board"),
					},
				},
			},
		}

		err := dsinventory.UpdateDUTs(ctx, duts)
		So(err, ShouldBeNil)

		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-labstation-board", "xxx-labstation-model", "xxx-labstation-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-labstation-board", "xxx-labstation-model", "xxx-labstation-firmware-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)

		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				Hostname: "xxx-labstation",
			},
		)

		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-labstation-cros-version")
		So(resp.FaftVersion, ShouldEqual, "")
		So(resp.FirmwareVersion, ShouldEqual, "xxx-labstation-firmware-version")
		So(resp.ServoCrosVersion, ShouldEqual, "")
	})

	Convey("Test GetStableVersion RPC -- look up beaglebone proper", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		// use a fake beaglebone servo
		duts := []*inventory.DeviceUnderTest{
			{
				Common: &inventory.CommonDeviceSpecs{
					Attributes: []*inventory.KeyValue{
						{
							Key:   strptr("servo_host"),
							Value: strptr("xxx-beaglebone-servo"),
						},
					},
					Id:       strptr("xxx-id"),
					Hostname: strptr("xxx-hostname"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-model"),
						Board: strptr("xxx-build-target"),
					},
				},
			},
		}

		err := dsinventory.UpdateDUTs(ctx, duts)
		So(err, ShouldBeNil)

		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleCrosStableVersion(ctx, beagleboneServo, "", "xxx-beaglebone-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)

		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				Hostname: "xxx-beaglebone-servo",
			},
		)

		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-beaglebone-cros-version")
		So(resp.FaftVersion, ShouldEqual, "")
		So(resp.FirmwareVersion, ShouldEqual, "")
		So(resp.ServoCrosVersion, ShouldEqual, "")
	})

	Convey("Test GetStableVersion RPC -- hostname with dummy_host", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		// use a fake labstation
		duts := []*inventory.DeviceUnderTest{
			{
				Common: &inventory.CommonDeviceSpecs{
					Attributes: []*inventory.KeyValue{
						{
							Key:   strptr("servo_host"),
							Value: strptr("dummy_host"),
						},
					},
					Id:       strptr("xxx-id"),
					Hostname: strptr("xxx-hostname"),
					Labels: &inventory.SchedulableLabels{
						Model: strptr("xxx-model"),
						Board: strptr("xxx-build-target"),
					},
				},
			},
		}

		err := dsinventory.UpdateDUTs(ctx, duts)
		So(err, ShouldBeNil)

		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleCrosStableVersion(ctx, "xxx-labstation-board", "xxx-labstation-model", "xxx-labstation-cros-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFaftStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-faft-version")
		So(err, ShouldBeNil)
		err = dssv.PutSingleFirmwareStableVersion(ctx, "xxx-build-target", "xxx-model", "xxx-firmware-version")
		So(err, ShouldBeNil)

		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				Hostname: "xxx-hostname",
			},
		)

		So(err, ShouldBeNil)
		So(resp.CrosVersion, ShouldEqual, "xxx-cros-version")
		So(resp.FaftVersion, ShouldEqual, "xxx-faft-version")
		So(resp.FirmwareVersion, ShouldEqual, "xxx-firmware-version")
		So(resp.ServoCrosVersion, ShouldEqual, "")
	})

	Convey("Test GetStableVersion RPC -- no stable versions exist", t, func() {
		ctx := testingContext()
		datastore.GetTestable(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		resp, err := tf.Inventory.GetStableVersion(
			ctx,
			&fleet.GetStableVersionRequest{
				BuildTarget: "xxx-build-target",
				Model:       "xxx-model",
			},
		)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})
}

// TestLooksLikeServo tests that looks like servo heuristic.
func TestLooksLikeServo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		hostname string
		isServo  bool
	}{
		{
			name:     "empty string",
			hostname: "",
			isServo:  false,
		},
		{
			name:     "a servo",
			hostname: "chromeos32-servo",
			isServo:  true,
		},
		{
			name:     "servov4p1 is *DUT* not servo",
			hostname: "servo-servo4p1",
			isServo:  false,
		},
		{
			name:     "legitimate-seeming DUT hostname",
			hostname: "chromeos100-row100-rack100-host100",
			isServo:  false,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expected := tt.isServo
			actual := looksLikeServo(tt.hostname)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("unexpected diff (-want +got): %s", diff)
			}
		})
	}
}

func withDutInfoCacheValidity(ctx context.Context, v time.Duration) context.Context {
	cfg := config.Get(ctx)
	cfg.Inventory.DutInfoCacheValidity = durationpb.New(v)
	return config.Use(ctx, cfg)
}

func withSplitInventory(ctx context.Context) context.Context {
	cfg := config.Get(ctx)
	cfg.Inventory.Multifile = true
	return config.Use(ctx, cfg)
}

func getDutInfo(t *testing.T, di *fleet.GetDutInfoResponse) *inventory.DeviceUnderTest {
	t.Helper()

	var dut inventory.DeviceUnderTest
	So(di.Spec, ShouldNotBeNil)
	err := proto.Unmarshal(di.Spec, &dut)
	So(err, ShouldBeNil)
	return &dut
}

func getDutInfoBasic(t *testing.T, di *fleet.GetDutInfoResponse) *inventory.DeviceUnderTest {
	t.Helper()
	var dut inventory.DeviceUnderTest
	if di.Spec == nil {
		t.Fatalf("Got nil spec")
	}
	err := proto.Unmarshal(di.Spec, &dut)
	if err != nil {
		t.Fatalf("Unmarshal DutInfo returned non-nil error: %s", err)
	}
	return &dut
}

// Maximum time to failure: (2^7 - 1)*(50/1000) = 6.35 seconds
var testRetriesTemplate = retry.ExponentialBackoff{
	Limited: retry.Limited{
		Delay:   50 * time.Millisecond,
		Retries: 7,
	},
	MaxDelay:   5 * time.Second,
	Multiplier: 2,
}

func testRetryIteratorFactory() retry.Iterator {
	it := testRetriesTemplate
	return &it
}

func strptr(x string) *string {
	return &x
}
