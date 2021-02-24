// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/chromiumos/infra/proto/go/device"
	"go.chromium.org/chromiumos/infra/proto/go/lab"
	"go.chromium.org/chromiumos/infra/proto/go/manufacturing"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/errors"
	. "go.chromium.org/luci/common/testing/assertions"
	ds "go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "infra/appengine/cros/lab_inventory/api/v1"
	"infra/appengine/cros/lab_inventory/app/config"
	"infra/appengine/cros/lab_inventory/app/external"
	"infra/appengine/cros/lab_inventory/app/frontend/fake"
	"infra/cros/lab_inventory/datastore"
	"infra/cros/lab_inventory/deviceconfig"
	"infra/cros/lab_inventory/hwid"
	invlibs "infra/cros/lab_inventory/protos"
	fleet "infra/libs/fleet/protos"
	ufs "infra/libs/fleet/protos/go"
	ufspb "infra/unifiedfleet/api/v1/models"
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
			// Get invalid argument error since we pass an empty request.
			So(status.Code(err), ShouldEqual, codes.InvalidArgument)
		})
	})
}

func TestAddCrosDevices(t *testing.T) {
	t.Parallel()
	dut1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{Hostname: "dut1"},
		},
	}
	dut2 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{Hostname: "dut2"},
		},
	}
	labstation1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{Hostname: "labstation1"},
		},
	}
	Convey("Add Chrome OS devices", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		Convey("Add new devices", func() {
			for _, d := range []*lab.ChromeOSDevice{&dut1, &labstation1} {
				req := &api.AddCrosDevicesRequest{
					Devices: []*lab.ChromeOSDevice{d},
				}
				resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
				So(err, ShouldBeNil)
				So(resp.PassedDevices, ShouldHaveLength, 1)
			}
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
				Devices: []*lab.ChromeOSDevice{&dut1, &dut2},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)
			So(resp.FailedDevices, ShouldHaveLength, 0)
			So(resp.PassedDevices[0].Hostname, ShouldEqual, "dut1")

			resp, err = tf.Inventory.AddCrosDevices(tf.C, req2)
			So(err, ShouldBeNil)
			So(resp.FailedDevices, ShouldHaveLength, 1)
			So(resp.PassedDevices, ShouldHaveLength, 1)

			So(resp.FailedDevices[0].Hostname, ShouldEqual, "dut1")
			So(resp.PassedDevices[0].Hostname, ShouldEqual, "dut2")
		})
	})
}

func TestDeleteCrosDevices(t *testing.T) {
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
		Id: &lab.ChromeOSDeviceID{
			Value: "ASSET_ID_123",
		},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{
				Hostname: "labstation1",
			},
		},
	}
	labstationID1 := api.DeviceID{
		Id: &api.DeviceID_ChromeosDeviceId{ChromeosDeviceId: "ASSET_ID_123"},
	}
	devID2 := api.DeviceID{
		Id: &api.DeviceID_Hostname{Hostname: "dut1"},
	}
	devIDNonExisting := api.DeviceID{
		Id: &api.DeviceID_Hostname{Hostname: "ghost"},
	}
	Convey("Delete Chrome OS devices", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		Convey("Happy path", func() {
			for _, d := range []*lab.ChromeOSDevice{&dut1, &labstation1} {
				req := &api.AddCrosDevicesRequest{
					Devices: []*lab.ChromeOSDevice{d},
				}
				resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
				So(err, ShouldBeNil)
				So(resp.PassedDevices, ShouldHaveLength, 1)
			}

			reqDelete := &api.DeleteCrosDevicesRequest{
				Ids: []*api.DeviceID{&labstationID1, &devID2},
			}
			rsp, err := tf.Inventory.DeleteCrosDevices(tf.C, reqDelete)
			So(err, ShouldBeNil)
			So(rsp.RemovedDevices, ShouldHaveLength, 2)

			removedDeviceNames := make([]string, 2)
			for i, r := range rsp.RemovedDevices {
				removedDeviceNames[i] = r.Hostname
			}
			So("dut1", ShouldBeIn, removedDeviceNames)
			So("labstation1", ShouldBeIn, removedDeviceNames)

			So(rsp.FailedDevices, ShouldHaveLength, 0)
		})

		Convey("Bad request: duplicated ID", func() {
			req := &api.DeleteCrosDevicesRequest{Ids: []*api.DeviceID{&labstationID1, &labstationID1}}
			rsp, err := tf.Inventory.DeleteCrosDevices(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Duplicated id found")
		})

		Convey("Delete non existing device", func() {
			req := &api.DeleteCrosDevicesRequest{Ids: []*api.DeviceID{&devIDNonExisting}}
			rsp, err := tf.Inventory.DeleteCrosDevices(tf.C, req)
			So(rsp.RemovedDevices, ShouldHaveLength, 0)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetCrosDevices(t *testing.T) {
	t.Parallel()
	dut1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{Hostname: "test-dut"},
		},
	}
	labstation1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{Value: "test-machine-labstation"},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{Hostname: "test-labstation"},
		},
	}
	devID1 := api.DeviceID{
		Id: &api.DeviceID_ChromeosDeviceId{ChromeosDeviceId: "test-machine-labstation"},
	}
	devID2 := api.DeviceID{
		Id: &api.DeviceID_Hostname{Hostname: "test-dut"},
	}
	devIDNonExisting := api.DeviceID{
		Id: &api.DeviceID_Hostname{Hostname: "ghost"},
	}

	Convey("Get Chrome OS devices", t, func() {
		ctx := testingContext()
		ctx = external.WithTestingContext(ctx)
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		for _, d := range []*lab.ChromeOSDevice{&dut1, &labstation1} {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{d},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)
		}

		getHwidDataFunc = func(ctx context.Context, hwidstr string, secret string) (*hwid.Data, error) {
			return &hwid.Data{Sku: "test", Variant: "test"}, nil
		}
		getDeviceConfigFunc = func(ctx context.Context, ids []*device.ConfigId) ([]proto.Message, error) {
			fakeCfgs := make([]proto.Message, len(ids))
			for i := range fakeCfgs {
				fakeCfgs[i] = &device.Config{}
			}
			return fakeCfgs, nil
		}
		getManufacturingConfigFunc = func(ctx context.Context, ids []*manufacturing.ConfigID) ([]proto.Message, error) {
			fakeCfgs := make([]proto.Message, len(ids))
			for i := range fakeCfgs {
				fakeCfgs[i] = &manufacturing.Config{}
			}
			return fakeCfgs, nil
		}

		Convey("Happy path", func() {
			reqGet := &api.GetCrosDevicesRequest{
				Ids: []*api.DeviceID{&devID1, &devID2},
			}
			rsp, err := tf.Inventory.GetCrosDevices(tf.C, reqGet)
			So(err, ShouldBeNil)
			So(rsp.FailedDevices, ShouldBeEmpty)
			So(rsp.Data, ShouldHaveLength, 2)
			for _, d := range rsp.Data {
				var machine *ufspb.Machine
				if d.GetLabConfig().GetDut() != nil {
					nb, err := proto.Marshal(d.GetLabConfig().GetDut())
					So(err, ShouldBeNil)
					ob, err := proto.Marshal(fake.GetMockDUT().GetChromeosMachineLse().GetDeviceLse().GetDut())
					So(err, ShouldBeNil)
					So(nb, ShouldResemble, ob)

					nb, err = proto.Marshal(d.GetDutState())
					So(err, ShouldBeNil)
					fakeDutState := fake.GetMockDutStateForDUT()
					fakeDutState.Hostname = ""
					ob, err = proto.Marshal(fakeDutState)
					So(err, ShouldBeNil)
					So(nb, ShouldResemble, ob)

					machine = fake.GetMockMachineForDUT()
				} else {
					nb, err := proto.Marshal(d.GetLabConfig().GetLabstation())
					So(err, ShouldBeNil)
					ob, err := proto.Marshal(fake.GetMockLabstation().GetChromeosMachineLse().GetDeviceLse().GetLabstation())
					So(err, ShouldBeNil)
					So(nb, ShouldResemble, ob)

					nb, err = proto.Marshal(d.GetDutState())
					So(err, ShouldBeNil)
					fakeDutState := fake.GetMockDutStateForLabstation()
					fakeDutState.Hostname = ""
					ob, err = proto.Marshal(fakeDutState)
					So(err, ShouldBeNil)
					So(nb, ShouldResemble, ob)

					machine = fake.GetMockMachineForLabstation()
				}
				So(d.GetLabConfig().GetSerialNumber(), ShouldEqual, machine.GetSerialNumber())
				So(d.GetLabConfig().GetId().GetValue(), ShouldEqual, machine.GetName())
				So(d.GetLabConfig().GetDeviceConfigId().GetPlatformId().GetValue(), ShouldEqual, machine.GetChromeosMachine().GetBuildTarget())
				So(d.GetLabConfig().GetDeviceConfigId().GetModelId().GetValue(), ShouldEqual, machine.GetChromeosMachine().GetModel())
				So(d.GetLabConfig().GetDeviceConfigId().GetVariantId().GetValue(), ShouldEqual, machine.GetChromeosMachine().GetSku())
				So(d.GetLabConfig().GetManufacturingId().GetValue(), ShouldEqual, machine.GetChromeosMachine().GetHwid())
			}
		})

		Convey("Happy path with real device config", func() {
			realDutID1 := api.DeviceID{
				Id: &api.DeviceID_Hostname{Hostname: "test-dut"},
			}
			realDutID2 := api.DeviceID{
				Id: &api.DeviceID_Hostname{Hostname: "test-labstation"},
			}
			realDut1 := lab.ChromeOSDevice{
				Id: &lab.ChromeOSDeviceID{},
				DeviceConfigId: &device.ConfigId{
					PlatformId: &device.PlatformId{
						Value: "testdutplatform",
					},
					ModelId: &device.ModelId{
						Value: "testdutmodel",
					},
					VariantId: &device.VariantId{
						Value: "testdutvariant",
					},
				},
				Device: &lab.ChromeOSDevice_Dut{
					Dut: &lab.DeviceUnderTest{Hostname: "test-dut"},
				},
			}
			realDut2 := lab.ChromeOSDevice{
				Id: &lab.ChromeOSDeviceID{},
				DeviceConfigId: &device.ConfigId{
					PlatformId: &device.PlatformId{
						Value: "testlabstationplatform",
					},
					ModelId: &device.ModelId{
						Value: "testlabstationmodel",
					},
				},
				Device: &lab.ChromeOSDevice_Dut{
					Dut: &lab.DeviceUnderTest{Hostname: "test-labstation"},
				},
			}
			_, err := datastore.AddDevices(ctx, []*lab.ChromeOSDevice{&realDut1, &realDut2}, false)
			So(err, ShouldBeNil)
			reqGet := &api.GetCrosDevicesRequest{
				Ids: []*api.DeviceID{&realDutID1, &realDutID2},
			}

			getDeviceConfigFunc = func(ctx context.Context, ids []*device.ConfigId) ([]proto.Message, error) {
				m := map[string]*device.Config{
					"testdutplatform.testdutmodel.": {
						GpuFamily: "real_gpu",
					},
					"testlabstationplatform.testlabstationmodel.": {
						GpuFamily: "fake_gpu",
					},
				}
				var errs errors.MultiError
				fakeCfgs := make([]proto.Message, 0)
				for _, id := range ids {
					if dc, ok := m[deviceconfig.GetDeviceConfigIDStr(id)]; ok {
						fakeCfgs = append(fakeCfgs, dc)
						errs = append(errs, nil)
					} else {
						fakeCfgs = append(fakeCfgs, nil)
						errs = append(errs, errors.Reason("no such entity").Err())
					}
				}
				return fakeCfgs, errs
			}
			rsp, err := tf.Inventory.GetCrosDevices(tf.C, reqGet)
			So(err, ShouldBeNil)
			So(rsp.Data, ShouldHaveLength, 2)
			resultM := make(map[string]string, 0)
			for _, d := range rsp.Data {
				if d.GetLabConfig().GetDut() != nil {
					resultM[d.GetLabConfig().GetDut().GetHostname()] = d.GetDeviceConfig().GetGpuFamily()
				} else {
					resultM[d.GetLabConfig().GetLabstation().GetHostname()] = d.GetDeviceConfig().GetGpuFamily()
				}
			}
			So(resultM["test-dut"], ShouldEqual, "real_gpu")
			So(resultM["test-labstation"], ShouldEqual, "fake_gpu")
		})

		Convey("Bad hwid server", func() {
			getHwidDataFunc = hwid.GetHwidData
			reqGet := &api.GetCrosDevicesRequest{
				Ids: []*api.DeviceID{&devID1, &devID2},
			}
			rsp, err := tf.Inventory.GetCrosDevices(tf.C, reqGet)
			So(err, ShouldBeNil)
			So(rsp.Data, ShouldHaveLength, 2)
			So(rsp.FailedDevices, ShouldHaveLength, 0)
			So(rsp.Data[0].LabConfig, ShouldNotBeNil)
			So(rsp.Data[0].HwidData, ShouldBeNil)
			So(rsp.Data[1].LabConfig, ShouldNotBeNil)
			So(rsp.Data[1].HwidData, ShouldBeNil)
		})

		Convey("Failed to get device config", func() {
			getHwidDataFunc = func(ctx context.Context, hwidstr string, secret string) (*hwid.Data, error) {
				return &hwid.Data{Sku: "sku", Variant: "variant"}, nil
			}
			getDeviceConfigFunc = func(ctx context.Context, ids []*device.ConfigId) ([]proto.Message, error) {
				errs := make([]error, len(ids))
				for i := range ids {
					errs[i] = errors.New("get device config error")
				}
				return make([]proto.Message, len(ids)), errors.NewMultiError(errs...)
			}
			reqGet := &api.GetCrosDevicesRequest{
				Ids: []*api.DeviceID{&devID1, &devID2},
			}
			rsp, err := tf.Inventory.GetCrosDevices(tf.C, reqGet)
			So(err, ShouldBeNil)
			So(rsp.Data, ShouldHaveLength, 2)
			So(rsp.FailedDevices, ShouldHaveLength, 0)
			So(rsp.Data[0].LabConfig, ShouldNotBeNil)
			So(rsp.Data[0].DeviceConfig, ShouldBeNil)
			So(rsp.Data[1].LabConfig, ShouldNotBeNil)
			So(rsp.Data[1].DeviceConfig, ShouldBeNil)
		})

		Convey("Get non existing device", func() {
			reqGet := &api.GetCrosDevicesRequest{
				Ids: []*api.DeviceID{&devID1, &devID2, &devIDNonExisting},
			}
			rsp, err := tf.Inventory.GetCrosDevices(tf.C, reqGet)
			So(err, ShouldBeNil)
			So(rsp.FailedDevices, ShouldHaveLength, 1)
			So(rsp.FailedDevices[0].ErrorMsg, ShouldContainSubstring, "No MachineLSE found")
			So(rsp.Data, ShouldHaveLength, 2)
		})
	})
}

func TestListCrosDevicesLabConfig(t *testing.T) {
	t.Parallel()

	dut1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{Hostname: "dut1"},
		},
	}
	dut2 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{Hostname: "dut2"},
		},
	}
	labstation1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{Value: "ASSET_ID_123"},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{Hostname: "labstation1"},
		},
	}

	Convey("List ChromeOS device lab configs", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		for _, d := range []*lab.ChromeOSDevice{&dut1, &dut2, &labstation1} {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{d},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)

		}
		for _, d := range []*lab.ChromeOSDevice{&dut1, &dut2} {
			resp2, err := tf.Inventory.UpdateDutsStatus(ctx, &api.UpdateDutsStatusRequest{
				States: []*lab.DutState{
					{
						Id: &lab.ChromeOSDeviceID{
							Value: d.GetId().GetValue(),
						},
						Servo: lab.PeripheralState_BROKEN,
					},
				},
			})
			So(err, ShouldBeNil)
			So(resp2.GetUpdatedDevices(), ShouldHaveLength, 1)
		}

		resp, err := tf.Inventory.ListCrosDevicesLabConfig(tf.C, &api.ListCrosDevicesLabConfigRequest{})
		So(err, ShouldBeNil)
		So(resp.GetLabConfigs(), ShouldHaveLength, 3)
		duts := make([]string, 0, 2)
		labstations := make([]string, 0, 1)
		for _, lc := range resp.GetLabConfigs() {
			c := lc.GetConfig()
			if c.GetDut() != nil {
				duts = append(duts, c.GetDut().GetHostname())
				So(lc.GetState().GetServo(), ShouldEqual, lab.PeripheralState_BROKEN)
				So(lc.GetState().GetId().GetValue(), ShouldEqual, c.GetId().GetValue())
			}
			if c.GetLabstation() != nil {
				labstations = append(labstations, c.GetLabstation().GetHostname())
				So(lc.GetState(), ShouldResembleProto, &lab.DutState{})
			}
		}
		So(duts, ShouldContain, "dut1")
		So(duts, ShouldContain, "dut2")
		So(labstations, ShouldContain, "labstation1")
	})
}

func TestUpdateCrosDevicesSetup(t *testing.T) {
	t.Parallel()
	getDut := func(servo *lab.Servo) *lab.ChromeOSDevice {
		if servo == nil {
			servo = &lab.Servo{
				ServoHostname: "labstation1",
				ServoPort:     8887,
				ServoSerial:   "SN0002",
				ServoType:     "v4",
			}
		}
		return &lab.ChromeOSDevice{
			Id: &lab.ChromeOSDeviceID{Value: "UUID:01"},
			Device: &lab.ChromeOSDevice_Dut{
				Dut: &lab.DeviceUnderTest{
					Hostname: "dut1",
					Peripherals: &lab.Peripherals{
						Servo: servo,
					},
				},
			},
		}
	}
	getLab := func() *lab.ChromeOSDevice {
		return &lab.ChromeOSDevice{
			Id: &lab.ChromeOSDeviceID{Value: "UUID:02"},
			Device: &lab.ChromeOSDevice_Labstation{
				Labstation: &lab.Labstation{Hostname: "labstation1"},
			},
		}
	}
	Convey("Update Chrome OS devices setup", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		for _, d := range []*lab.ChromeOSDevice{getLab(), getDut(nil)} {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{d},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)
		}

		Convey("Happy path", func() {
			servo := &lab.Servo{
				ServoHostname: "labstation1",
				ServoPort:     1234,
				ServoSerial:   "SN0002",
				ServoType:     "v4",
			}

			req := &api.UpdateCrosDevicesSetupRequest{Devices: []*lab.ChromeOSDevice{getDut(servo), getLab()}}
			resp, err := tf.Inventory.UpdateCrosDevicesSetup(tf.C, req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.UpdatedDevices, ShouldHaveLength, 2)
		})

		Convey("Fail update the DUT when update servo with existed serial number under labstation", func() {
			dutServo := &lab.Servo{
				ServoHostname: "labstation1",
				ServoPort:     1230,
				ServoSerial:   "SN0099",
				ServoType:     "v3",
			}
			labServo := &lab.Servo{
				ServoHostname: "labstation1",
				ServoPort:     1231,
				ServoSerial:   "SN0099",
				ServoType:     "v3",
			}
			dut1 := getDut(dutServo)
			labstation1 := getLab()
			labstation1.GetLabstation().Servos = []*lab.Servo{labServo}

			req := &api.UpdateCrosDevicesSetupRequest{Devices: []*lab.ChromeOSDevice{dut1, labstation1}}
			resp, err := tf.Inventory.UpdateCrosDevicesSetup(tf.C, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "the servo with serial number: \"SN0099\" is already attached to \"labstation1\"")
			So(resp, ShouldBeNil)
		})

		Convey("Fail update the DUT when update servo with existed port number under labstation", func() {
			dutServo := &lab.Servo{
				ServoHostname: "labstation1",
				ServoPort:     1230,
				ServoSerial:   "SN0091",
				ServoType:     "v3",
			}
			labServos := []*lab.Servo{
				{
					ServoHostname: "labstation1",
					ServoPort:     1230,
					ServoSerial:   "SN0092",
					ServoType:     "v3",
				},
			}
			dut1 := getDut(dutServo)
			labstation1 := getLab()
			labstation1.GetLabstation().Servos = labServos

			req := &api.UpdateCrosDevicesSetupRequest{Devices: []*lab.ChromeOSDevice{dut1, labstation1}}
			resp, err := tf.Inventory.UpdateCrosDevicesSetup(tf.C, req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "the servo port: '1230' is already used in \"labstation1\"")
			So(resp, ShouldBeNil)
		})

		Convey("Update non-existing devices", func() {
			ghost := lab.ChromeOSDevice{
				Id: &lab.ChromeOSDeviceID{Value: "UUID:ghost"},
				Device: &lab.ChromeOSDevice_Dut{
					Dut: &lab.DeviceUnderTest{Hostname: "dut1"},
				},
			}
			req := &api.UpdateCrosDevicesSetupRequest{Devices: []*lab.ChromeOSDevice{&ghost}}
			resp, err := tf.Inventory.UpdateCrosDevicesSetup(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.FailedDevices, ShouldHaveLength, 1)
			So(resp.FailedDevices[0].ErrorMsg, ShouldContainSubstring, "no such entity")
		})
	})
}

func TestUpdateDutsStatus(t *testing.T) {
	t.Parallel()
	dut1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{Value: "UUID:01"},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{
				Hostname: "dut1",
				Peripherals: &lab.Peripherals{
					Servo: &lab.Servo{
						ServoType: "",
					},
					SmartUsbhub: false,
				},
			},
		},
	}
	labstation1 := lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{Value: "UUID:02"},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{Hostname: "labstation1"},
		},
	}
	Convey("Update Dut status", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		for _, d := range []*lab.ChromeOSDevice{&dut1, &labstation1} {
			req := &api.AddCrosDevicesRequest{
				Devices: []*lab.ChromeOSDevice{d},
			}
			resp, err := tf.Inventory.AddCrosDevices(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.PassedDevices, ShouldHaveLength, 1)
		}
		dutStateGood := &lab.DutState{
			Id: &lab.ChromeOSDeviceID{Value: "UUID:01"},
		}
		dutMetaGood := &api.DutMeta{
			ChromeosDeviceId: "UUID:01",
			SerialNumber:     "serial2",
			HwID:             "hwid2",
		}
		labMetaGood := &api.LabMeta{
			ChromeosDeviceId: "UUID:01",
			ServoType:        "servo_v4_with_ccd_cr50",
			ServoTopology: &lab.ServoTopology{
				Main: &lab.ServoTopologyItem{
					Type:         "servo_v4",
					Serial:       "SomeSerial",
					SysfsProduct: "1-4.6.5",
				},
			},
			SmartUsbhub: true,
		}
		getUpdatedDevice := func(ctx context.Context) *lab.ChromeOSDevice {
			r := datastore.GetDevicesByIds(ctx, []string{"UUID:01"})
			So(r, ShouldHaveLength, 1)
			var p lab.ChromeOSDevice
			r[0].Entity.GetCrosDeviceProto(&p)
			return &p
		}

		Convey("Happy path", func() {
			req := &api.UpdateDutsStatusRequest{
				States: []*lab.DutState{dutStateGood},
			}
			resp, err := tf.Inventory.UpdateDutsStatus(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.UpdatedDevices, ShouldHaveLength, 1)
			So(resp.UpdatedDevices[0].Id, ShouldEqual, "UUID:01")

			p := getUpdatedDevice(ctx)
			So(p.GetSerialNumber(), ShouldEqual, "")
			So(p.GetManufacturingId().GetValue(), ShouldEqual, "")
			So(p.GetDut().GetPeripherals().GetServo().GetServoType(), ShouldEqual, "")
			So(p.GetDut().GetPeripherals().GetServo().GetServoTopology(), ShouldBeNil)
			So(p.GetDut().GetPeripherals().GetSmartUsbhub(), ShouldEqual, false)
		})

		Convey("Happy path with dut meta", func() {
			req := &api.UpdateDutsStatusRequest{
				States:   []*lab.DutState{dutStateGood},
				DutMetas: []*api.DutMeta{dutMetaGood},
			}
			resp, err := tf.Inventory.UpdateDutsStatus(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.UpdatedDevices, ShouldHaveLength, 1)
			So(resp.UpdatedDevices[0].Id, ShouldEqual, "UUID:01")

			p := getUpdatedDevice(ctx)
			So(p.GetSerialNumber(), ShouldEqual, "serial2")
			So(p.GetManufacturingId().GetValue(), ShouldEqual, "hwid2")
			So(p.GetDut().GetPeripherals().GetServo().GetServoType(), ShouldEqual, "")
			So(p.GetDut().GetPeripherals().GetServo().GetServoTopology(), ShouldBeNil)
			So(p.GetDut().GetPeripherals().GetSmartUsbhub(), ShouldEqual, false)
		})

		Convey("Happy path with lab meta", func() {
			req := &api.UpdateDutsStatusRequest{
				States:   []*lab.DutState{dutStateGood},
				LabMetas: []*api.LabMeta{labMetaGood},
			}
			resp, err := tf.Inventory.UpdateDutsStatus(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.UpdatedDevices, ShouldHaveLength, 1)
			So(resp.UpdatedDevices[0].Id, ShouldEqual, "UUID:01")

			p := getUpdatedDevice(ctx)
			So(p.GetSerialNumber(), ShouldEqual, "")
			So(p.GetManufacturingId().GetValue(), ShouldEqual, "")
			So(p.GetDut().GetPeripherals().GetServo().GetServoType(), ShouldEqual, "servo_v4_with_ccd_cr50")
			So(proto.MarshalTextString(p.GetDut().GetPeripherals().GetServo().GetServoTopology()), ShouldEqual, proto.MarshalTextString(labMetaGood.GetServoTopology()))
			So(p.GetDut().GetPeripherals().GetSmartUsbhub(), ShouldEqual, true)
		})

		Convey("Happy path with dut and lab meta", func() {
			req := &api.UpdateDutsStatusRequest{
				States:   []*lab.DutState{dutStateGood},
				DutMetas: []*api.DutMeta{dutMetaGood},
				LabMetas: []*api.LabMeta{labMetaGood},
			}
			resp, err := tf.Inventory.UpdateDutsStatus(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.UpdatedDevices, ShouldHaveLength, 1)
			So(resp.UpdatedDevices[0].Id, ShouldEqual, "UUID:01")

			p := getUpdatedDevice(ctx)
			So(p.GetSerialNumber(), ShouldEqual, "serial2")
			So(p.GetManufacturingId().GetValue(), ShouldEqual, "hwid2")
			So(p.GetDut().GetPeripherals().GetServo().GetServoType(), ShouldEqual, "servo_v4_with_ccd_cr50")
			So(proto.MarshalTextString(p.GetDut().GetPeripherals().GetServo().GetServoTopology()), ShouldEqual, proto.MarshalTextString(labMetaGood.GetServoTopology()))
			So(p.GetDut().GetPeripherals().GetSmartUsbhub(), ShouldEqual, true)
		})

		Convey("Cannot update a labstation", func() {
			req := &api.UpdateDutsStatusRequest{
				States: []*lab.DutState{
					{
						Id: &lab.ChromeOSDeviceID{Value: "UUID:02"},
					},
				}}
			resp, err := tf.Inventory.UpdateDutsStatus(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.FailedDevices, ShouldHaveLength, 1)
			So(resp.FailedDevices[0].ErrorMsg, ShouldContainSubstring, "labstation")
		})
	})
}
func mockAsset(id, lab string) *fleet.ChopsAsset {
	return &fleet.ChopsAsset{
		Id: id,
		Location: &ufs.Location{
			Lab:      lab,
			Row:      "Phobos-3",
			Rack:     "Deimos-0",
			Shelf:    "Olympus-Mons",
			Position: "Amazonis-Planitia",
		},
	}
}

func assertLocationEqual(a, b *ufs.Location) {
	So(a.GetLab(), ShouldEqual, b.GetLab())
	So(a.GetRow(), ShouldEqual, b.GetRow())
	So(a.GetRack(), ShouldEqual, b.GetRack())
	So(a.GetShelf(), ShouldEqual, b.GetShelf())
	So(a.GetPosition(), ShouldEqual, b.GetPosition())
}

func TestAddAsset(t *testing.T) {
	t.Parallel()

	asset1 := mockAsset("4675636B596F75", "lab1")
	asset2 := mockAsset("4675636B596F76", "lab2")
	asset3 := mockAsset("", "")

	Convey("AddAsset", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		Convey("Add asset with tag", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Failed, ShouldHaveLength, 0)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
		})
		Convey("Add asset without tag", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset3},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Passed, ShouldHaveLength, 0)
		})
		Convey("Add existing asset", func() {
			req1 := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 0)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
			req2 := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1, asset2},
			}
			resp, err = tf.Inventory.AddAssets(tf.C, req2)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset2.GetId())
			So(resp.Failed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
		})
	})
}

func TestUpdateAsset(t *testing.T) {
	t.Parallel()

	asset1 := mockAsset("4675636B596F75", "lab1")
	asset2 := mockAsset("4675636B596F76", "lab2")
	asset3 := mockAsset("4675636B596F76", "lab3")
	asset4 := mockAsset("", "")

	Convey("UpdateAsset", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()

		Convey("Update non-existing asset", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1},
			}
			resp, err := tf.Inventory.UpdateAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Passed, ShouldHaveLength, 0)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Failed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
		})
		Convey("Update asset without tag", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset4},
			}
			resp, err := tf.Inventory.UpdateAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Passed, ShouldHaveLength, 0)
		})
		Convey("Update existing asset", func() {
			req1 := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset2},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 0)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset2.GetId())
			req2 := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1, asset3},
			}
			resp, err = tf.Inventory.UpdateAssets(tf.C, req2)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Failed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset3.GetId())

			req := &api.AssetIDList{Id: []string{asset2.GetId()}}
			resp, err = tf.Inventory.GetAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.Passed[0].Asset.GetLocation().GetLab(), ShouldEqual, asset3.GetLocation().GetLab())
		})
	})
}

func TestGetAsset(t *testing.T) {
	t.Parallel()

	asset1 := mockAsset("4675636B596F75", "lab1")

	Convey("GetAsset", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		req := &api.AssetList{
			Asset: []*fleet.ChopsAsset{asset1},
		}
		resp, err := tf.Inventory.AddAssets(tf.C, req)
		So(err, ShouldBeNil)
		So(resp.Passed, ShouldHaveLength, 1)
		So(resp.Passed[0].ErrorMsg, ShouldEqual, "")
		So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
		assertLocationEqual(resp.Passed[0].Asset.GetLocation(), asset1.Location)
		Convey("Get Asset from ID", func() {
			req := &api.AssetIDList{Id: []string{asset1.GetId()}}
			resp, err = tf.Inventory.GetAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].ErrorMsg, ShouldEqual, "")
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
			assertLocationEqual(resp.Passed[0].Asset.GetLocation(), asset1.Location)
		})
		Convey("Get Asset from non existent ID", func() {
			req := &api.AssetIDList{Id: []string{
				"4675636B596F76",
			}}
			resp, err = tf.Inventory.GetAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.Failed, ShouldHaveLength, 1)
			So(resp.Failed[0].ErrorMsg, ShouldNotEqual, "")
			So(resp.Failed[0].Asset.GetId(), ShouldEqual, "4675636B596F76")
		})
	})
}

func TestDeleteAsset(t *testing.T) {
	t.Parallel()

	asset1 := mockAsset("4675636B596F75", "lab1")
	asset2 := mockAsset("4675636B596F76", "lab2")

	Convey("DeleteAsset", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		Convey("Delete Asset from ID", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
			req1 := &api.AssetIDList{Id: []string{asset1.GetId()}}
			resp1, err := tf.Inventory.DeleteAssets(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp1.Passed, ShouldHaveLength, 1)
			So(resp1.Passed[0].Id, ShouldEqual, asset1.GetId())
		})
		Convey("Delete non-existent Asset", func() {
			req := &api.AssetList{
				Asset: []*fleet.ChopsAsset{asset1},
			}
			resp, err := tf.Inventory.AddAssets(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.Passed, ShouldHaveLength, 1)
			So(resp.Passed[0].Asset.GetId(), ShouldEqual, asset1.GetId())
			req1 := &api.AssetIDList{
				Id: []string{asset1.GetId(), asset2.GetId()},
			}
			resp2, err := tf.Inventory.DeleteAssets(tf.C, req1)
			So(err, ShouldBeNil)
			So(resp2.Failed, ShouldHaveLength, 1)
			So(resp2.Passed, ShouldHaveLength, 1)
			So(resp2.Failed[0].Id, ShouldEqual, asset2.GetId())
			So(resp2.Passed[0].Id, ShouldEqual, asset1.GetId())
		})
	})
}

type devcfgEntity struct {
	_kind     string `gae:"$kind,DevConfig"`
	ID        string `gae:"$id"`
	DevConfig []byte `gae:",noindex"`
	Updated   time.Time
}

func TestDeviceConfigsExists(t *testing.T) {
	t.Parallel()

	Convey("Test exists device config in datastore", t, func() {
		ctx := testingContext()
		tf, validate := newTestFixtureWithContext(ctx, t)
		defer validate()
		err := ds.Put(ctx, []devcfgEntity{
			{ID: "kunimitsu.lars.variant1"},
			{ID: "sarien.arcada.variant2"},
			{
				ID:        "platform.model.variant3",
				DevConfig: []byte("bad data"),
			},
		})
		So(err, ShouldBeNil)

		Convey("Happy path", func() {
			resp, err := tf.Inventory.DeviceConfigsExists(ctx, &api.DeviceConfigsExistsRequest{
				ConfigIds: []*device.ConfigId{
					{
						PlatformId: &device.PlatformId{Value: "lars"},
						ModelId:    &device.ModelId{Value: "lars"},
						VariantId:  &device.VariantId{Value: "variant1"},
					},
					{
						PlatformId: &device.PlatformId{Value: "arcada"},
						ModelId:    &device.ModelId{Value: "arcada"},
						VariantId:  &device.VariantId{Value: "variant2"},
					},
				},
			})
			So(err, ShouldBeNil)
			So(resp.Exists[0], ShouldBeTrue)
			So(resp.Exists[1], ShouldBeTrue)
		})

		Convey("check for nonexisting data", func() {
			resp, err := tf.Inventory.DeviceConfigsExists(ctx, &api.DeviceConfigsExistsRequest{
				ConfigIds: []*device.ConfigId{
					{
						PlatformId: &device.PlatformId{Value: "platform"},
						ModelId:    &device.ModelId{Value: "model"},
						VariantId:  &device.VariantId{Value: "variant-nonexisting"},
					},
				},
			})
			So(err, ShouldBeNil)
			So(resp.Exists[0], ShouldBeFalse)
		})

		Convey("check for existing and nonexisting data", func() {
			resp, err := tf.Inventory.DeviceConfigsExists(ctx, &api.DeviceConfigsExistsRequest{
				ConfigIds: []*device.ConfigId{
					{
						PlatformId: &device.PlatformId{Value: "platform"},
						ModelId:    &device.ModelId{Value: "model"},
						VariantId:  &device.VariantId{Value: "variant-nonexisting"},
					},
					{
						PlatformId: &device.PlatformId{Value: "arcada"},
						ModelId:    &device.ModelId{Value: "arcada"},
						VariantId:  &device.VariantId{Value: "variant2"},
					},
				},
			})
			So(err, ShouldBeNil)
			So(resp.Exists[0], ShouldBeFalse)
			So(resp.Exists[1], ShouldBeTrue)
		})
	})
}

func mockDeviceManualRepairRecord(hostname string, assetTag string, createdTime int64, completed bool) *invlibs.DeviceManualRepairRecord {
	var state invlibs.DeviceManualRepairRecord_RepairState
	var updatedTime timestamp.Timestamp
	var completedTime timestamp.Timestamp
	if completed {
		state = invlibs.DeviceManualRepairRecord_STATE_COMPLETED
		updatedTime = timestamp.Timestamp{Seconds: 444, Nanos: 0}
		completedTime = timestamp.Timestamp{Seconds: 444, Nanos: 0}
	} else {
		state = invlibs.DeviceManualRepairRecord_STATE_IN_PROGRESS
		updatedTime = timestamp.Timestamp{Seconds: 222, Nanos: 0}
		completedTime = timestamp.Timestamp{Seconds: 444, Nanos: 0}
	}

	return &invlibs.DeviceManualRepairRecord{
		Hostname:                        hostname,
		AssetTag:                        assetTag,
		RepairTargetType:                invlibs.DeviceManualRepairRecord_TYPE_DUT,
		RepairState:                     state,
		BuganizerBugUrl:                 "https://b/12345678",
		ChromiumBugUrl:                  "https://crbug.com/12345678",
		DutRepairFailureDescription:     "Mock DUT repair failure description.",
		DutVerifierFailureDescription:   "Mock DUT verifier failure description.",
		ServoRepairFailureDescription:   "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis:                       "Mock diagnosis.",
		RepairProcedure:                 "Mock repair procedure.",
		LabstationRepairActions: []invlibs.LabstationRepairAction{
			invlibs.LabstationRepairAction_LABSTATION_POWER_CYCLE,
			invlibs.LabstationRepairAction_LABSTATION_REIMAGE,
			invlibs.LabstationRepairAction_LABSTATION_UPDATE_CONFIG,
			invlibs.LabstationRepairAction_LABSTATION_REPLACE,
		},
		IssueFixed:    true,
		UserLdap:      "testing-account",
		TimeTaken:     15,
		CreatedTime:   &timestamp.Timestamp{Seconds: createdTime, Nanos: 0},
		UpdatedTime:   &updatedTime,
		CompletedTime: &completedTime,
	}
}

func mockServo(servoHost string) *lab.Servo {
	return &lab.Servo{
		ServoHostname: servoHost,
		ServoPort:     8888,
		ServoSerial:   "SERVO1",
		ServoType:     "v3",
	}
}

func mockDut(hostname, id, servoHost string) *lab.ChromeOSDevice {
	return &lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{
			Value: id,
		},
		Device: &lab.ChromeOSDevice_Dut{
			Dut: &lab.DeviceUnderTest{
				Hostname: hostname,
				Peripherals: &lab.Peripherals{
					Servo:       mockServo(servoHost),
					SmartUsbhub: false,
				},
			},
		},
	}
}

func mockLabstation(hostname, id string) *lab.ChromeOSDevice {
	return &lab.ChromeOSDevice{
		Id: &lab.ChromeOSDeviceID{
			Value: id,
		},
		Device: &lab.ChromeOSDevice_Labstation{
			Labstation: &lab.Labstation{
				Hostname: hostname,
			},
		},
	}
}

func TestGetDeviceManualRepairRecord(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).Consistent(true)

	record1 := mockDeviceManualRepairRecord("chromeos-getRecords-aa", "getRecords-111", 1, false)
	record2 := mockDeviceManualRepairRecord("chromeos-getRecords-bb", "getRecords-222", 1, false)
	record3 := mockDeviceManualRepairRecord("chromeos-getRecords-bb", "getRecords-333", 1, false)
	records := []*invlibs.DeviceManualRepairRecord{record1, record2, record3}

	// Set up records in datastore
	datastore.AddDeviceManualRepairRecords(ctx, records)

	Convey("Test get device manual repair records", t, func() {
		Convey("Get record using single hostname", func() {
			req := &api.GetDeviceManualRepairRecordRequest{
				Hostname: "chromeos-getRecords-aa",
			}
			resp, err := tf.Inventory.GetDeviceManualRepairRecord(tf.C, req)
			So(err, ShouldBeNil)
			So(resp.DeviceRepairRecord, ShouldNotBeNil)
		})
		Convey("Get first record when hostname has multiple active records", func() {
			req := &api.GetDeviceManualRepairRecordRequest{
				Hostname: "chromeos-getRecords-bb",
			}
			resp, err := tf.Inventory.GetDeviceManualRepairRecord(tf.C, req)
			So(resp.DeviceRepairRecord, ShouldNotBeNil)
			So(resp.DeviceRepairRecord.GetAssetTag(), ShouldEqual, "getRecords-222")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "More than one active record found")
		})
		Convey("Get record using non-existent hostname", func() {
			req := &api.GetDeviceManualRepairRecordRequest{
				Hostname: "chromeos-getRecords-cc",
			}
			resp, err := tf.Inventory.GetDeviceManualRepairRecord(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "No record found")
		})
		Convey("Get record using empty hostname", func() {
			req := &api.GetDeviceManualRepairRecordRequest{
				Hostname: "",
			}
			resp, err := tf.Inventory.GetDeviceManualRepairRecord(tf.C, req)
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "No record found")
		})
	})
}

func TestCreateDeviceManualRepairRecord(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).Consistent(true)

	// Empty datastore
	record1 := mockDeviceManualRepairRecord("chromeos-createRecords-aa", "", 1, false)
	record2 := mockDeviceManualRepairRecord("", "", 1, false)

	// Set up records in datastore
	Convey("Test add devices using an empty datastore", t, func() {
		Convey("Add single record", func() {
			propFilter := map[string]string{"hostname": record1.Hostname}
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record1}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Check added record
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-aa")
			So(getRes[0].Record.GetAssetTag(), ShouldEqual, "n/a")
			So(getRes[0].Record.GetCreatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 1, Nanos: 0})
		})
		Convey("Add single record without hostname", func() {
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record2}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Hostname cannot be empty")

			// No record should be added
			propFilter := map[string]string{"hostname": record2.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 0)
		})
		Convey("Add single record to a host with an open record", func() {
			// Check existing record
			propFilter := map[string]string{"hostname": record1.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-aa")

			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record1}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "A record already exists for host chromeos-createRecords-aa")
		})
	})

	// Datastore with DeviceEntity
	record3 := mockDeviceManualRepairRecord("chromeos-createRecords-bb", "", 1, false)
	record4 := mockDeviceManualRepairRecord("chromeos-createRecords-cc", "", 1, false)
	record5 := mockDeviceManualRepairRecord("", "", 1, false)
	record6 := mockDeviceManualRepairRecord("chromeos-createRecords-ee", "", 1, true)

	Convey("Test add devices using an non-empty datastore", t, func() {
		dut1 := mockDut("chromeos-createRecords-bb", "mockDutAssetTag-111", "labstation1")
		dut2 := mockDut("chromeos-createRecords-cc", "", "labstation1")
		dut3 := mockDut("chromeos-createRecords-ee", "mockDutAssetTag-222", "labstation1")
		labstation1 := mockLabstation("labstation1", "assetId-111")
		dut1.DeviceConfigId = &device.ConfigId{ModelId: &device.ModelId{Value: "model1"}}
		dut2.DeviceConfigId = &device.ConfigId{ModelId: &device.ModelId{Value: "model2"}}
		dut3.DeviceConfigId = &device.ConfigId{ModelId: &device.ModelId{Value: "model3"}}
		labstation1.DeviceConfigId = &device.ConfigId{
			ModelId: &device.ModelId{Value: "model5"},
		}
		devsToAdd := []*lab.ChromeOSDevice{dut1, dut2, dut3, labstation1}
		_, err := datastore.AddDevices(ctx, devsToAdd, false)
		if err != nil {
			t.Fatal(err)
		}
		Convey("Add single record", func() {
			propFilter := map[string]string{"hostname": record3.Hostname}
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record3}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Check added record
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-bb")
			So(getRes[0].Record.GetAssetTag(), ShouldEqual, "mockDutAssetTag-111")
			So(getRes[0].Record.GetCreatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 1, Nanos: 0})
		})
		Convey("Add single record using dut without asset tag", func() {
			propFilter := map[string]string{"hostname": record4.Hostname}
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record4}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Asset tag should be uuid generated for dut
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-cc")
			So(getRes[0].Record.GetAssetTag(), ShouldNotEqual, "")
			So(getRes[0].Record.GetAssetTag(), ShouldNotEqual, "n/a")
			So(getRes[0].Record.GetCreatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 1, Nanos: 0})
		})
		Convey("Add single record with no hostname", func() {
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record5}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Hostname cannot be empty")
		})
		Convey("Add single record with completed repair state", func() {
			propFilter := map[string]string{"hostname": record6.Hostname}
			req := &api.CreateDeviceManualRepairRecordRequest{DeviceRepairRecord: record6}
			rsp, err := tf.Inventory.CreateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Completed time should be same as created
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-ee")
			So(getRes[0].Record.GetAssetTag(), ShouldEqual, "mockDutAssetTag-222")
			So(getRes[0].Record.GetCreatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 1, Nanos: 0})
			So(getRes[0].Record.GetCreatedTime(), ShouldResembleProto, getRes[0].Record.GetCompletedTime())
		})
	})
}

func TestUpdateDeviceManualRepairRecord(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).Consistent(true)

	// Empty datastore
	record1 := mockDeviceManualRepairRecord("chromeos-updateRecords-aa", "updateRec-111", 1, false)
	record1Complete := mockDeviceManualRepairRecord("chromeos-updateRecords-aa", "updateRec-111", 1, true)
	record2 := mockDeviceManualRepairRecord("chromeos-updateRecords-bb", "updateRec-222", 1, false)
	record2Complete := mockDeviceManualRepairRecord("chromeos-updateRecords-bb", "updateRec-222", 1, true)
	record3 := mockDeviceManualRepairRecord("chromeos-updateRecords-cc", "updateRec-333", 1, false)
	record3Update := mockDeviceManualRepairRecord("chromeos-updateRecords-cc", "updateRec-333", 1, false)
	record4 := mockDeviceManualRepairRecord("chromeos-updateRecords-dd", "updateRec-444", 1, false)

	// Set up records in datastore
	records := []*invlibs.DeviceManualRepairRecord{record1, record2, record3}
	datastore.AddDeviceManualRepairRecords(ctx, records)

	Convey("Test update devices using an non-empty datastore", t, func() {
		Convey("Update single record with completed repair state", func() {
			propFilter := map[string]string{"hostname": record1.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			req := &api.UpdateDeviceManualRepairRecordRequest{
				Id:                 getRes[0].Entity.ID,
				DeviceRepairRecord: record1Complete,
			}
			rsp, err := tf.Inventory.UpdateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Check updated record
			getRes, err = datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-updateRecords-aa")
			So(getRes[0].Record.GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_COMPLETED)
			So(getRes[0].Record.GetUpdatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 222, Nanos: 0})
			So(getRes[0].Record.GetUpdatedTime(), ShouldResembleProto, getRes[0].Record.GetCompletedTime())
		})
		Convey("Update single record with no id", func() {
			req := &api.UpdateDeviceManualRepairRecordRequest{
				Id:                 "",
				DeviceRepairRecord: record2Complete,
			}
			rsp, err := tf.Inventory.UpdateDeviceManualRepairRecord(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ID cannot be empty")

			// Check updated record and make sure it is unchanged
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, map[string]string{"hostname": record2.Hostname}, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-updateRecords-bb")
			So(getRes[0].Record.GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_IN_PROGRESS)
			So(getRes[0].Record.GetUpdatedTime(), ShouldResembleProto, &timestamp.Timestamp{Seconds: 222, Nanos: 0})
		})
		Convey("Update single record", func() {
			propFilter := map[string]string{"hostname": record3.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			record3Update.TimeTaken = 20
			req := &api.UpdateDeviceManualRepairRecordRequest{
				Id:                 getRes[0].Entity.ID,
				DeviceRepairRecord: record3Update,
			}
			rsp, err := tf.Inventory.UpdateDeviceManualRepairRecord(tf.C, req)
			So(rsp.String(), ShouldEqual, "")
			So(err, ShouldBeNil)

			// Check updated record and make sure fields are changed properly
			getRes, err = datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-updateRecords-cc")
			So(getRes[0].Record.GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_IN_PROGRESS)
			So(getRes[0].Record.GetTimeTaken(), ShouldEqual, 20)
			So(getRes[0].Record.GetUpdatedTime(), ShouldNotResemble, &timestamp.Timestamp{Seconds: 222, Nanos: 0})
			So(getRes[0].Record.GetCompletedTime(), ShouldResembleProto, &timestamp.Timestamp{Seconds: 444, Nanos: 0})
		})
		Convey("Update single non-existent record", func() {
			propFilter := map[string]string{"hostname": record4.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 0)

			req := &api.UpdateDeviceManualRepairRecordRequest{
				Id:                 "test-id",
				DeviceRepairRecord: record4,
			}
			rsp, err := tf.Inventory.UpdateDeviceManualRepairRecord(tf.C, req)
			So(rsp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "No open record exists for host chromeos-updateRecords-dd")
		})
	})
}

func TestListManualRepairRecords(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).AutoIndex(true)
	ds.GetTestable(ctx).Consistent(true)

	// Updated times should go in descending order of record1 > record2 = record3
	record1 := mockDeviceManualRepairRecord("chromeos-getRecords-aa", "getRecords-111", 1, true)
	record2 := mockDeviceManualRepairRecord("chromeos-getRecords-aa", "getRecords-111", 2, false)
	record3 := mockDeviceManualRepairRecord("chromeos-getRecords-aa", "getRecords-222", 3, false)
	records := []*invlibs.DeviceManualRepairRecord{record1, record2, record3}

	// Set up records in datastore
	datastore.AddDeviceManualRepairRecords(ctx, records)

	Convey("Test list device manual repair records", t, func() {
		Convey("List records using hostname and asset tag", func() {
			req := &api.ListManualRepairRecordsRequest{
				Hostname: "chromeos-getRecords-aa",
				AssetTag: "getRecords-111",
				Limit:    5,
			}
			resp, err := tf.Inventory.ListManualRepairRecords(tf.C, req)

			So(err, ShouldBeNil)
			So(resp.RepairRecords, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 2)
			So(resp.RepairRecords[0].GetHostname(), ShouldEqual, "chromeos-getRecords-aa")
			So(resp.RepairRecords[0].GetAssetTag(), ShouldEqual, "getRecords-111")
			So(resp.RepairRecords[0].GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_IN_PROGRESS)
			So(resp.RepairRecords[1].GetHostname(), ShouldEqual, "chromeos-getRecords-aa")
			So(resp.RepairRecords[1].GetAssetTag(), ShouldEqual, "getRecords-111")
			So(resp.RepairRecords[1].GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_COMPLETED)
		})
		Convey("List records using hostname and asset tag with offset", func() {
			req := &api.ListManualRepairRecordsRequest{
				Hostname: "chromeos-getRecords-aa",
				AssetTag: "getRecords-111",
				Limit:    1,
				Offset:   1,
			}
			resp, err := tf.Inventory.ListManualRepairRecords(tf.C, req)

			So(err, ShouldBeNil)
			So(resp.RepairRecords, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 1)
			So(resp.RepairRecords[0].GetHostname(), ShouldEqual, "chromeos-getRecords-aa")
			So(resp.RepairRecords[0].GetAssetTag(), ShouldEqual, "getRecords-111")
			So(resp.RepairRecords[0].GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_COMPLETED)
		})
		Convey("List records using all filters", func() {
			req := &api.ListManualRepairRecordsRequest{
				Hostname:    "chromeos-getRecords-aa",
				AssetTag:    "getRecords-111",
				Limit:       5,
				UserLdap:    "testing-account",
				RepairState: "STATE_COMPLETED",
			}
			resp, err := tf.Inventory.ListManualRepairRecords(tf.C, req)

			So(err, ShouldBeNil)
			So(resp.RepairRecords, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 1)
			So(resp.RepairRecords[0].GetHostname(), ShouldEqual, "chromeos-getRecords-aa")
			So(resp.RepairRecords[0].GetAssetTag(), ShouldEqual, "getRecords-111")
			So(resp.RepairRecords[0].GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_COMPLETED)
		})
		Convey("List records using hostname and asset tag with limit 1", func() {
			req := &api.ListManualRepairRecordsRequest{
				Hostname: "chromeos-getRecords-aa",
				AssetTag: "getRecords-111",
				Limit:    1,
			}
			resp, err := tf.Inventory.ListManualRepairRecords(tf.C, req)

			So(err, ShouldBeNil)
			So(resp.RepairRecords, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 1)
			So(resp.RepairRecords[0].GetHostname(), ShouldEqual, "chromeos-getRecords-aa")
			So(resp.RepairRecords[0].GetAssetTag(), ShouldEqual, "getRecords-111")
			So(resp.RepairRecords[0].GetRepairState(), ShouldEqual, invlibs.DeviceManualRepairRecord_STATE_IN_PROGRESS)
		})
		Convey("List records that do not exist", func() {
			req := &api.ListManualRepairRecordsRequest{
				Hostname: "chromeos-getRecords-bb",
				AssetTag: "getRecords-111",
				Limit:    5,
			}
			resp, err := tf.Inventory.ListManualRepairRecords(tf.C, req)

			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 0)
		})
	})
}

func TestBatchGetManualRepairRecords(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).Consistent(true)

	record1 := mockDeviceManualRepairRecord("chromeos-getRecords-xx", "getRecords-111", 1, false)
	record2 := mockDeviceManualRepairRecord("chromeos-getRecords-yy", "getRecords-222", 1, false)
	record3 := mockDeviceManualRepairRecord("chromeos-getRecords-zz", "getRecords-333", 1, false)
	record4 := mockDeviceManualRepairRecord("chromeos-getRecords-zz", "getRecords-444", 1, false)
	records := []*invlibs.DeviceManualRepairRecord{record1, record2, record3, record4}

	// Set up records in datastore
	datastore.AddDeviceManualRepairRecords(ctx, records)

	Convey("Test batch get manual repair records", t, func() {
		Convey("Get record using multiple hostnames", func() {
			req := &api.BatchGetManualRepairRecordsRequest{
				Hostnames: []string{
					"chromeos-getRecords-xx",
					"chromeos-getRecords-yy",
				},
			}
			resp, err := tf.Inventory.BatchGetManualRepairRecords(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 2)
			So(resp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(resp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(resp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-getRecords-xx")
			So(resp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-getRecords-xx")
			So(resp.RepairRecords[1].ErrorMsg, ShouldBeEmpty)
			So(resp.RepairRecords[1].RepairRecord, ShouldNotBeNil)
			So(resp.RepairRecords[1].RepairRecord.Hostname, ShouldEqual, "chromeos-getRecords-yy")
			So(resp.RepairRecords[1].Hostname, ShouldEqual, "chromeos-getRecords-yy")
		})
		Convey("Get first record when hostname has multiple active records", func() {
			req := &api.BatchGetManualRepairRecordsRequest{
				Hostnames: []string{
					"chromeos-getRecords-zz",
				},
			}
			resp, err := tf.Inventory.BatchGetManualRepairRecords(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 1)
			So(resp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-getRecords-zz")
			So(resp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-getRecords-zz")
		})
		Convey("Get record using a non-existent hostname", func() {
			req := &api.BatchGetManualRepairRecordsRequest{
				Hostnames: []string{
					"chromeos-getRecords-xx",
					"chromeos-getRecords-cc",
				},
			}
			resp, err := tf.Inventory.BatchGetManualRepairRecords(tf.C, req)
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp.RepairRecords, ShouldHaveLength, 2)
			So(resp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(resp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(resp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-getRecords-xx")
			So(resp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-getRecords-xx")
			So(resp.RepairRecords[1].ErrorMsg, ShouldContainSubstring, "No record found")
			So(resp.RepairRecords[1].RepairRecord, ShouldBeNil)
			So(resp.RepairRecords[1].Hostname, ShouldEqual, "chromeos-getRecords-cc")
		})
	})
}

func TestBatchCreateManualRepairRecords(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	tf, validate := newTestFixtureWithContext(ctx, t)
	defer validate()

	ds.GetTestable(ctx).Consistent(true)

	// Empty datastore
	record1 := mockDeviceManualRepairRecord("chromeos-createRecords-zz", "", 1, false)
	record2 := mockDeviceManualRepairRecord("chromeos-createRecords-yy", "", 1, false)
	record3 := mockDeviceManualRepairRecord("chromeos-createRecords-xx", "", 1, false)
	record4 := mockDeviceManualRepairRecord("chromeos-createRecords-ww", "", 1, false)
	record5 := mockDeviceManualRepairRecord("", "", 1, false)

	// Set up records in datastore
	Convey("Test add devices using an empty datastore", t, func() {
		Convey("Add single record", func() {
			createReq := &api.BatchCreateManualRepairRecordsRequest{
				RepairRecords: []*invlibs.DeviceManualRepairRecord{record1},
			}
			createRsp, err := tf.Inventory.BatchCreateManualRepairRecords(tf.C, createReq)
			So(err, ShouldBeNil)
			So(createRsp, ShouldNotBeNil)
			So(createRsp.RepairRecords, ShouldHaveLength, 1)
			So(createRsp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(createRsp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(createRsp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-zz")
			So(createRsp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-createRecords-zz")

			// Check added record
			getReq := &api.BatchGetManualRepairRecordsRequest{
				Hostnames: []string{
					"chromeos-createRecords-zz",
				},
			}
			getRsp, err := tf.Inventory.BatchGetManualRepairRecords(tf.C, getReq)
			So(err, ShouldBeNil)
			So(getRsp, ShouldNotBeNil)
			So(getRsp.RepairRecords, ShouldHaveLength, 1)
			So(getRsp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(getRsp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(getRsp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-zz")
			So(getRsp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-createRecords-zz")
		})
		Convey("Add multiple records", func() {
			createReq := &api.BatchCreateManualRepairRecordsRequest{
				RepairRecords: []*invlibs.DeviceManualRepairRecord{record2, record3},
			}
			createRsp, err := tf.Inventory.BatchCreateManualRepairRecords(tf.C, createReq)
			So(err, ShouldBeNil)
			So(createRsp, ShouldNotBeNil)
			So(createRsp.RepairRecords, ShouldHaveLength, 2)
			So(createRsp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(createRsp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(createRsp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-xx")
			So(createRsp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-createRecords-xx")
			So(createRsp.RepairRecords[1].ErrorMsg, ShouldBeEmpty)
			So(createRsp.RepairRecords[1].RepairRecord, ShouldNotBeNil)
			So(createRsp.RepairRecords[1].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-yy")
			So(createRsp.RepairRecords[1].Hostname, ShouldEqual, "chromeos-createRecords-yy")

			// Check added record
			getReq := &api.BatchGetManualRepairRecordsRequest{
				Hostnames: []string{
					"chromeos-createRecords-yy",
					"chromeos-createRecords-xx",
				},
			}
			getRsp, err := tf.Inventory.BatchGetManualRepairRecords(tf.C, getReq)
			So(err, ShouldBeNil)
			So(getRsp, ShouldNotBeNil)
			So(getRsp.RepairRecords, ShouldHaveLength, 2)
			So(getRsp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(getRsp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(getRsp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-yy")
			So(getRsp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-createRecords-yy")
			So(getRsp.RepairRecords[1].ErrorMsg, ShouldBeEmpty)
			So(getRsp.RepairRecords[1].RepairRecord, ShouldNotBeNil)
			So(getRsp.RepairRecords[1].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-xx")
			So(getRsp.RepairRecords[1].Hostname, ShouldEqual, "chromeos-createRecords-xx")
		})
		Convey("Add multiple records; one with an open record", func() {
			// Check existing record
			propFilter := map[string]string{"hostname": record1.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 1)
			So(getRes[0].Record.GetHostname(), ShouldEqual, "chromeos-createRecords-zz")

			createReq := &api.BatchCreateManualRepairRecordsRequest{
				RepairRecords: []*invlibs.DeviceManualRepairRecord{record1, record4},
			}
			createRsp, err := tf.Inventory.BatchCreateManualRepairRecords(tf.C, createReq)
			So(err, ShouldBeNil)
			So(createRsp, ShouldNotBeNil)
			So(createRsp.RepairRecords, ShouldHaveLength, 2)
			So(createRsp.RepairRecords[0].ErrorMsg, ShouldBeEmpty)
			So(createRsp.RepairRecords[0].RepairRecord, ShouldNotBeNil)
			So(createRsp.RepairRecords[0].RepairRecord.Hostname, ShouldEqual, "chromeos-createRecords-ww")
			So(createRsp.RepairRecords[0].Hostname, ShouldEqual, "chromeos-createRecords-ww")
			So(createRsp.RepairRecords[1].ErrorMsg, ShouldContainSubstring, "A record already exists for host chromeos-createRecords-zz")
			So(createRsp.RepairRecords[1].RepairRecord, ShouldBeNil)
			So(createRsp.RepairRecords[1].Hostname, ShouldEqual, "chromeos-createRecords-zz")
		})
		Convey("Add single record without hostname", func() {
			createReq := &api.BatchCreateManualRepairRecordsRequest{
				RepairRecords: []*invlibs.DeviceManualRepairRecord{record5},
			}
			createRsp, err := tf.Inventory.BatchCreateManualRepairRecords(tf.C, createReq)
			So(err, ShouldBeNil)
			So(createRsp, ShouldNotBeNil)
			So(createRsp.RepairRecords, ShouldHaveLength, 1)
			So(createRsp.RepairRecords[0].ErrorMsg, ShouldContainSubstring, "Hostname cannot be empty")
			So(createRsp.RepairRecords[0].Hostname, ShouldBeEmpty)

			// No record should be added
			propFilter := map[string]string{"hostname": record5.Hostname}
			getRes, err := datastore.GetRepairRecordByPropertyName(ctx, propFilter, -1, 0, []string{})
			So(getRes, ShouldHaveLength, 0)
		})
	})
}
