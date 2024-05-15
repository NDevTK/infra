// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.chromium.org/chromiumos/config/go/test/api"
	schedulingAPI "go.chromium.org/chromiumos/config/go/test/scheduling"
	. "go.chromium.org/luci/common/testing/assertions"

	"infra/device_manager/internal/model"
	"infra/libs/skylab/inventory/swarming"
)

func TestGetDevice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("GetDevice", t, func() {
		Convey("GetDevice: valid return", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			timeNow := time.Now()
			rows := sqlmock.NewRows([]string{
				"id",
				"device_address",
				"device_type",
				"device_state",
				"schedulable_labels",
				"created_time",
				"last_updated_time",
				"is_active"}).
				AddRow(
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"DEVICE_STATE_AVAILABLE",
					`{"label-test":{"Values":["test-value-1"]}}`,
					timeNow,
					timeNow,
					true).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					timeNow,
					timeNow,
					false)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state,
					schedulable_labels,
					created_time,
					last_updated_time,
					is_active
				FROM "Devices"
				WHERE id=$1;`)).
				WithArgs("test-device-1").
				WillReturnRows(rows)

			device, err := GetDevice(ctx, db, "test-device-1")
			So(err, ShouldBeNil)
			So(device, ShouldResembleProto, &api.Device{
				Id: "test-device-1",
				Address: &api.DeviceAddress{
					Host: "1.1.1.1",
					Port: 1,
				},
				Type:  api.DeviceType_DEVICE_TYPE_PHYSICAL,
				State: api.DeviceState_DEVICE_STATE_AVAILABLE,
				HardwareReqs: &api.HardwareRequirements{
					SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
						"label-test": {
							Values: []string{"test-value-1"},
						},
					},
				},
			})
		})
		Convey("GetDevice: invalid request; no device name match", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state,
					schedulable_labels,
					created_time,
					last_updated_time,
					is_active
				FROM "Devices"
				WHERE id=$1;`)).
				WithArgs("test-device-2").
				WillReturnError(fmt.Errorf("GetDevice: failed to get Device"))

			device, err := GetDevice(ctx, db, "test-device-2")
			So(err, ShouldNotBeNil)
			So(device, ShouldResembleProto, &api.Device{})
		})
	})
}

func TestListDevices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("ListDevices", t, func() {
		Convey("ListDevices: valid return; page token returned", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			var (
				pageSize int32 = 1
				timeNow        = time.Now()
			)

			rows := sqlmock.NewRows([]string{
				"id",
				"device_address",
				"device_type",
				"device_state",
				"schedulable_labels",
				"last_updated_time",
				"is_active"}).
				AddRow(
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"DEVICE_STATE_AVAILABLE",
					`{"label-test":{"Values":["test-value-1"]}}`,
					timeNow,
					true).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					timeNow,
					false)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state,
					schedulable_labels,
					last_updated_time,
					is_active
				FROM "Devices"
				ORDER BY id
				LIMIT $1;`)).
				WithArgs(pageSize + 1).
				WillReturnRows(rows)

			devices, err := ListDevices(ctx, db, &api.ListDevicesRequest{
				PageSize: pageSize,
			})
			So(err, ShouldBeNil)
			So(devices, ShouldResembleProto, &api.ListDevicesResponse{
				Devices: []*api.Device{
					{
						Id: "test-device-1",
						Address: &api.DeviceAddress{
							Host: "1.1.1.1",
							Port: 1,
						},
						Type:  api.DeviceType_DEVICE_TYPE_PHYSICAL,
						State: api.DeviceState_DEVICE_STATE_AVAILABLE,
						HardwareReqs: &api.HardwareRequirements{
							SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
								"label-test": {
									Values: []string{"test-value-1"},
								},
							},
						},
					},
				},
				NextPageToken: "dGVzdC1kZXZpY2UtMQ==",
			})

			decodedToken, err := model.DecodePageToken(ctx, model.PageToken(devices.GetNextPageToken()))
			So(err, ShouldBeNil)
			So(decodedToken, ShouldEqual, "test-device-1")
		})
		Convey("ListDevices: valid return; no page token returned", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			var (
				pageSize int32 = 2
				timeNow        = time.Now()
			)

			rows := sqlmock.NewRows([]string{
				"id",
				"device_address",
				"device_type",
				"device_state",
				"schedulable_labels",
				"last_updated_time",
				"is_active"}).
				AddRow(
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"DEVICE_STATE_AVAILABLE",
					`{"label-test":{"Values":["test-value-1"]}}`,
					timeNow,
					true).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					timeNow,
					false)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state,
					schedulable_labels,
					last_updated_time,
					is_active
				FROM "Devices"
				ORDER BY id
				LIMIT $1;`)).
				WithArgs(pageSize + 1).
				WillReturnRows(rows)

			devices, err := ListDevices(ctx, db, &api.ListDevicesRequest{
				PageSize: pageSize,
			})
			So(err, ShouldBeNil)
			So(devices.GetNextPageToken(), ShouldEqual, "")
			So(devices, ShouldResembleProto, &api.ListDevicesResponse{
				Devices: []*api.Device{
					{
						Id: "test-device-1",
						Address: &api.DeviceAddress{
							Host: "1.1.1.1",
							Port: 1,
						},
						Type:  api.DeviceType_DEVICE_TYPE_PHYSICAL,
						State: api.DeviceState_DEVICE_STATE_AVAILABLE,
						HardwareReqs: &api.HardwareRequirements{
							SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
								"label-test": {
									Values: []string{"test-value-1"},
								},
							},
						},
					},
					{
						Id: "test-device-2",
						Address: &api.DeviceAddress{
							Host: "2.2.2.2",
							Port: 2,
						},
						Type:  api.DeviceType_DEVICE_TYPE_VIRTUAL,
						State: api.DeviceState_DEVICE_STATE_LEASED,
						HardwareReqs: &api.HardwareRequirements{
							SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
								"label-test": {
									Values: []string{"test-value-2"},
								},
							},
						},
					},
				},
			})
		})
		Convey("ListDevices: valid request using page token", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			var (
				pageSize  int32 = 1
				pageToken       = "dGVzdC1kZXZpY2UtMQ=="
				timeNow         = time.Now()
			)

			// only add rows after test-device-1
			rows := sqlmock.NewRows([]string{
				"id",
				"device_address",
				"device_type",
				"device_state",
				"schedulable_labels",
				"last_updated_time",
				"is_active"}).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					timeNow,
					false)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state,
					schedulable_labels,
					last_updated_time,
					is_active
				FROM "Devices"
				WHERE id > $1
				ORDER BY id
				LIMIT $2;`)).
				WithArgs("test-device-1", pageSize+1).
				WillReturnRows(rows)

			devices, err := ListDevices(ctx, db, &api.ListDevicesRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			So(err, ShouldBeNil)
			So(devices.GetNextPageToken(), ShouldEqual, "")
			So(devices, ShouldResembleProto, &api.ListDevicesResponse{
				Devices: []*api.Device{
					{
						Id: "test-device-2",
						Address: &api.DeviceAddress{
							Host: "2.2.2.2",
							Port: 2,
						},
						Type:  api.DeviceType_DEVICE_TYPE_VIRTUAL,
						State: api.DeviceState_DEVICE_STATE_LEASED,
						HardwareReqs: &api.HardwareRequirements{
							SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
								"label-test": {
									Values: []string{"test-value-2"},
								},
							},
						},
					},
				},
			})
		})
	})
}

func TestUpdateDevice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Set up fake PubSub server
	srv := pstest.NewServer()
	defer func() {
		err := srv.Close()
		if err != nil {
			t.Logf("failed to close fake pubsub server: %s", err)
		}
	}()

	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("could not start fake pubsub server")
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			t.Logf("failed to close fake pubsub connection: %s", err)
		}
	}()

	psClient, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("could not connect to fake pubsub server")
	}
	defer func() {
		err = psClient.Close()
		if err != nil {
			t.Logf("failed to close fake pubsub client: %s", err)
		}
	}()

	_, err = psClient.CreateTopic(ctx, DeviceEventsPubSubTopic)
	if err != nil {
		t.Fatalf("failed to create fake pubsub topic")
	}

	Convey("UpdateDevice", t, func() {
		Convey("UpdateDevice: valid update", func() {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				mock.ExpectClose()
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			mock.ExpectBegin()

			var txOpts *sql.TxOptions
			tx, err := db.BeginTx(ctx, txOpts)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub db transaction", err)
			}

			timeNow := time.Now()
			mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE
					"Devices"
				SET
					device_address=COALESCE($2, device_address),
					device_type=COALESCE($3, device_type),
					device_state=COALESCE($4, device_state),
					schedulable_labels=COALESCE($5, schedulable_labels),
					last_updated_time=COALESCE($6, last_updated_time),
					is_active=COALESCE($7, is_active)
				WHERE
					id=$1;`)).
				WithArgs(
					"test-device-1",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-1"]}}`,
					timeNow,
					false).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = UpdateDevice(ctx, tx, psClient, model.Device{
				ID:            "test-device-1",
				DeviceAddress: "2.2.2.2:2",
				DeviceType:    "DEVICE_TYPE_VIRTUAL",
				DeviceState:   "DEVICE_STATE_LEASED",
				SchedulableLabels: model.SchedulableLabels{
					"label-test": model.LabelValues{
						Values: []string{"test-value-1"},
					},
				},
				LastUpdatedTime: timeNow,
				IsActive:        false,
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestIsDeviceAvailable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("IsDeviceAvailable", t, func() {
		Convey("IsDeviceAvailable: device is available", func() {
			rsp := IsDeviceAvailable(ctx, api.DeviceState_DEVICE_STATE_AVAILABLE)
			So(rsp, ShouldEqual, true)
		})
		Convey("IsDeviceAvailable: device is not available", func() {
			rsp := IsDeviceAvailable(ctx, api.DeviceState_DEVICE_STATE_LEASED)
			So(rsp, ShouldEqual, false)
		})
	})
}

func Test_stringToDeviceAddress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("stringToDeviceAddress", t, func() {
		Convey("stringToDeviceAddress: valid address", func() {
			addr, err := stringToDeviceAddress(ctx, "1.1.1.1:1")
			So(err, ShouldBeNil)
			So(addr, ShouldResembleProto, &api.DeviceAddress{
				Host: "1.1.1.1",
				Port: 1,
			})
		})
		Convey("stringToDeviceAddress: invalid address; no port", func() {
			addr, err := stringToDeviceAddress(ctx, "1.1.1.1.1.1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "failed to split host and port")
			So(addr, ShouldResembleProto, &api.DeviceAddress{})
		})
		Convey("stringToDeviceAddress: invalid address; bad port", func() {
			addr, err := stringToDeviceAddress(ctx, "1.1.1.1:abc")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "port abc is not convertible to integer")
			So(addr, ShouldResembleProto, &api.DeviceAddress{})
		})
	})
}

func Test_deviceAddressToString(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("deviceAddressToString", t, func() {
		Convey("deviceAddressToString: valid address", func() {
			addr := deviceAddressToString(ctx, &api.DeviceAddress{
				Host: "1.1.1.1",
				Port: 1,
			})
			So(addr, ShouldEqual, "1.1.1.1:1")
		})
		Convey("deviceAddressToString: ipv6 address", func() {
			addr := deviceAddressToString(ctx, &api.DeviceAddress{
				Host: "1:2:3",
				Port: 1,
			})
			So(addr, ShouldEqual, "[1:2:3]:1")
		})
	})
}

func Test_stringToDeviceType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("stringToDeviceType", t, func() {
		Convey("stringToDeviceType: valid types", func() {
			for _, deviceType := range []string{
				"DEVICE_TYPE_UNSPECIFIED",
				"DEVICE_TYPE_VIRTUAL",
				"DEVICE_TYPE_PHYSICAL",
			} {
				apiType := stringToDeviceType(ctx, deviceType)
				So(apiType, ShouldEqual, api.DeviceType_value[deviceType])
			}
		})
		Convey("stringToDeviceType: unknown type", func() {
			apiType := stringToDeviceType(ctx, "UNKNOWN_TYPE")
			So(apiType, ShouldEqual, api.DeviceType_DEVICE_TYPE_UNSPECIFIED)
		})
	})
}

func Test_stringToDeviceState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("stringToDeviceState", t, func() {
		Convey("stringToDeviceState: valid types", func() {
			for _, deviceState := range []string{
				"DEVICE_STATE_UNSPECIFIED",
				"DEVICE_STATE_AVAILABLE",
				"DEVICE_STATE_LEASED",
			} {
				apiState := stringToDeviceState(ctx, deviceState)
				So(apiState, ShouldEqual, api.DeviceState_value[deviceState])
			}
		})
		Convey("stringToDeviceState: unknown state", func() {
			apiState := stringToDeviceState(ctx, "UNKNOWN_STATE")
			So(apiState, ShouldEqual, api.DeviceState_DEVICE_STATE_UNSPECIFIED)
		})
	})
}

func Test_labelsToHardwareReqs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("labelsToHardwareReqs", t, func() {
		Convey("labelsToHardwareReqs: valid labels", func() {
			labels := model.SchedulableLabels{
				"label-test": model.LabelValues{
					Values: []string{
						"test-value-1",
						"test-value-2",
					},
				},
			}
			dims := labelsToHardwareReqs(ctx, labels)
			So(dims, ShouldResembleProto, api.HardwareRequirements{
				SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
					"label-test": {
						Values: []string{
							"test-value-1",
							"test-value-2",
						},
					},
				},
			})
		})
		Convey("labelsToHardwareReqs: empty labels", func() {
			labels := model.SchedulableLabels{}
			dims := labelsToHardwareReqs(ctx, labels)
			So(dims, ShouldEqual, &api.HardwareRequirements{
				SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{},
			})
		})
	})
}

func Test_labelsToSwarmingDims(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("labelsToSwarmingDims", t, func() {
		Convey("labelsToSwarmingDims: valid labels", func() {
			labels := model.SchedulableLabels{
				"label-test": model.LabelValues{
					Values: []string{
						"test-value-1",
						"test-value-2",
					},
				},
			}
			dims := labelsToSwarmingDims(ctx, labels)
			So(dims, ShouldResembleProto, schedulingAPI.SwarmingDimensions{
				DimsMap: map[string]*schedulingAPI.DimValues{
					"label-test": {
						Values: []string{
							"test-value-1",
							"test-value-2",
						},
					},
				},
			})
		})
		Convey("labelsToSwarmingDims: empty labels", func() {
			labels := model.SchedulableLabels{}
			dims := labelsToSwarmingDims(ctx, labels)
			So(dims, ShouldEqual, &schedulingAPI.SwarmingDimensions{
				DimsMap: map[string]*schedulingAPI.DimValues{},
			})
		})
	})
}

func TestSwarmingDimsToLabels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("SwarmingDimsToLabels", t, func() {
		Convey("SwarmingDimsToLabels: valid dims", func() {
			dims := swarming.Dimensions{
				"label-test": []string{
					"test-value-1",
					"test-value-2",
				},
			}
			labels := SwarmingDimsToLabels(ctx, dims)
			So(labels, ShouldEqual, model.SchedulableLabels{
				"label-test": model.LabelValues{
					Values: []string{
						"test-value-1",
						"test-value-2",
					},
				},
			})
		})
		Convey("SwarmingDimsToLabels: empty dims", func() {
			dims := swarming.Dimensions{}
			labels := SwarmingDimsToLabels(ctx, dims)
			So(labels, ShouldEqual, model.SchedulableLabels{})
		})
	})
}

func Test_deviceModelToAPIDevice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("deviceModelToAPIDevice", t, func() {
		Convey("deviceModelToAPIDevice: valid device", func() {
			modelDevice := model.Device{
				ID:            "test-device-1",
				DeviceAddress: "1.1.1.1:1",
				DeviceType:    "DEVICE_TYPE_PHYSICAL",
				DeviceState:   "DEVICE_STATE_AVAILABLE",
				SchedulableLabels: model.SchedulableLabels{
					"label-test": model.LabelValues{
						Values: []string{"test-value-1"},
					},
				},
			}

			apiDevice := deviceModelToAPIDevice(ctx, modelDevice)

			So(apiDevice, ShouldResembleProto, &api.Device{
				Id: "test-device-1",
				Address: &api.DeviceAddress{
					Host: "1.1.1.1",
					Port: 1,
				},
				Type:  api.DeviceType_DEVICE_TYPE_PHYSICAL,
				State: api.DeviceState_DEVICE_STATE_AVAILABLE,
				HardwareReqs: &api.HardwareRequirements{
					SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{
						"label-test": {
							Values: []string{"test-value-1"},
						},
					},
				},
			})
		})
		Convey("deviceModelToAPIDevice: invalid fields", func() {
			modelDevice := model.Device{
				ID:                "test-device-invalid",
				DeviceAddress:     "1.1",
				DeviceType:        "UNKNOWN",
				DeviceState:       "UNKNOWN",
				SchedulableLabels: model.SchedulableLabels{},
			}

			apiDevice := deviceModelToAPIDevice(ctx, modelDevice)

			So(apiDevice, ShouldResembleProto, &api.Device{
				Id:      "test-device-invalid",
				Address: &api.DeviceAddress{},
				HardwareReqs: &api.HardwareRequirements{
					SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{},
				},
			})
		})
		Convey("deviceModelToAPIDevice: empty device", func() {
			modelDevice := model.Device{
				ID: "test-device-empty",
			}

			apiDevice := deviceModelToAPIDevice(ctx, modelDevice)

			So(apiDevice, ShouldResembleProto, &api.Device{
				Id:      "test-device-empty",
				Address: &api.DeviceAddress{},
				HardwareReqs: &api.HardwareRequirements{
					SchedulableLabels: map[string]*api.HardwareRequirements_LabelValues{},
				},
			})
		})
	})
}
