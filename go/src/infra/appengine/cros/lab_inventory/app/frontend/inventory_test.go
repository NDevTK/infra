// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/chromiumos/infra/proto/go/lab"
	api "infra/appengine/cros/lab_inventory/api/v1"
	"infra/appengine/cros/lab_inventory/app/config"
)

type testFixture struct {
	T *testing.T
	C context.Context

	Inventory          *InventoryServerImpl
	DecoratedInventory *api.DecoratedInventory
}

func newTestFixtureWithContext(ctx context.Context, t *testing.T) (testFixture, func()) {
	tf := testFixture{T: t, C: ctx}
	mc := gomock.NewController(t)

	tf.Inventory = &InventoryServerImpl{}
	tf.DecoratedInventory = &api.DecoratedInventory{
		Service: tf.Inventory,
		Prelude: checkAccess,
	}

	validate := func() {
		mc.Finish()
	}
	return tf, validate
}

func testingContext() context.Context {
	c := gaetesting.TestingContextWithAppID("dev~infra-lab-inventory")
	c = config.Use(c, &config.Config{
		Readers: &config.LuciAuthGroup{
			Value: "fake_group",
		},
	})
	return c
}

func TestACL(t *testing.T) {
	t.Parallel()

	Convey("Get Chrome OS devices with ACL check", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		req := &api.GetCrosDevicesRequest{}
		Convey("Unknown user", func() {
			_, err := tf.DecoratedInventory.GetCrosDevices(tf.C, req)
			So(err, ShouldNotBeNil)
			So(status.Code(err), ShouldEqual, codes.Internal)
		})
		Convey("Non authorized user", func() {
			ctx := auth.WithState(tf.C, &authtest.FakeState{
				Identity:       "user:abc@def.com",
				IdentityGroups: []string{"abc"},
			})
			_, err := tf.DecoratedInventory.GetCrosDevices(ctx, req)
			So(err, ShouldNotBeNil)
			So(status.Code(err), ShouldEqual, codes.PermissionDenied)
		})
		Convey("Happy path", func() {
			ctx := auth.WithState(tf.C, &authtest.FakeState{
				Identity:       "user:abc@def.com",
				IdentityGroups: []string{"fake_group"},
			})
			_, err := tf.DecoratedInventory.GetCrosDevices(ctx, req)
			So(err, ShouldBeNil)
		})
	})
}

func TestAddCrosDevices(t *testing.T) {
	t.Parallel()
	dut1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{
				Hostname: "dut1",
			},
		},
	}
	labstation1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{
				Hostname: "labstation1",
			},
		},
	}
	Convey("Add Chrome OS devices", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		Convey("Add new devices", func() {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{&dut1, &labstation1},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 2)
		})
		Convey("Fail the input validation check", func() {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{&dut1, &dut1},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
		Convey("Add existing devices", func() {
			req1 := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{&dut1},
			}
			req2 := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{&dut1, &labstation1},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)
			So(resp.FailedDevices, ShouldHaveLength, 0)
			So(resp.PassedDevices[0].Hostname, ShouldEqual, "dut1")

			resp, err = tf.Inventory.AddCrosDevices(tf.C, req2)
			So(err, ShouldNotBeNil)
			So(resp.FailedDevices, ShouldHaveLength, 1)
			So(resp.PassedDevices, ShouldHaveLength, 1)

			So(resp.FailedDevices[0].Hostname, ShouldEqual, "dut1")
			So(resp.PassedDevices[0].Hostname, ShouldEqual, "labstation1")
		})
	})
}
