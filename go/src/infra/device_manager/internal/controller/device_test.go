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

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"go.chromium.org/chromiumos/config/go/test/api"
	. "go.chromium.org/luci/common/testing/assertions"

	"infra/device_manager/internal/model"
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

			rows := sqlmock.NewRows([]string{"id", "device_address", "device_type", "device_state"}).
				AddRow("test-device-1", "1.1.1.1:1", "DEVICE_TYPE_PHYSICAL", "DEVICE_STATE_AVAILABLE").
				AddRow("test-device-2", "2.2.2.2:2", "DEVICE_TYPE_VIRTUAL", "DEVICE_STATE_LEASED")

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state
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
					device_state
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

	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
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

	_, err = psClient.CreateTopic(ctx, deviceEventsPubsubTopic)
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

			mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE
					"Devices"
				SET
					device_address=COALESCE($2, device_address),
					device_type=COALESCE($3, device_type),
					device_state=COALESCE($4, device_state)
				WHERE
					id=$1;`)).
				WithArgs("test-device-1", "2.2.2.2:2", "DEVICE_TYPE_VIRTUAL", "DEVICE_STATE_LEASED").
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = UpdateDevice(ctx, tx, psClient, model.Device{
				ID:            "test-device-1",
				DeviceAddress: "2.2.2.2:2",
				DeviceType:    "DEVICE_TYPE_VIRTUAL",
				DeviceState:   "DEVICE_STATE_LEASED",
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

func TestConvertDeviceAddressToAPIFormat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("convertDeviceAddressToAPIFormat", t, func() {
		Convey("convertDeviceAddressToAPIFormat: valid address", func() {
			addr, err := convertDeviceAddressToAPIFormat(ctx, "1.1.1.1:1")
			So(err, ShouldBeNil)
			So(addr, ShouldResembleProto, &api.DeviceAddress{
				Host: "1.1.1.1",
				Port: 1,
			})
		})
		Convey("convertDeviceAddressToAPIFormat: invalid address; no port", func() {
			addr, err := convertDeviceAddressToAPIFormat(ctx, "1.1.1.1.1.1")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "failed to split host and port")
			So(addr, ShouldResembleProto, &api.DeviceAddress{})
		})
		Convey("convertDeviceAddressToAPIFormat: invalid address; bad port", func() {
			addr, err := convertDeviceAddressToAPIFormat(ctx, "1.1.1.1:abc")
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "port abc is not convertible to integer")
			So(addr, ShouldResembleProto, &api.DeviceAddress{})
		})
	})
}

func TestConvertDeviceAddressToDBFormat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("convertAPIDeviceAddressToDBFormat", t, func() {
		Convey("convertAPIDeviceAddressToDBFormat: valid address", func() {
			addr := convertAPIDeviceAddressToDBFormat(ctx, &api.DeviceAddress{
				Host: "1.1.1.1",
				Port: 1,
			})
			So(addr, ShouldEqual, "1.1.1.1:1")
		})
		Convey("convertAPIDeviceAddressToDBFormat: ipv6 address", func() {
			addr := convertAPIDeviceAddressToDBFormat(ctx, &api.DeviceAddress{
				Host: "1:2:3",
				Port: 1,
			})
			So(addr, ShouldEqual, "[1:2:3]:1")
		})
	})
}

func TestConvertDeviceTypeToAPIFormat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("convertDeviceTypeToAPIFormat", t, func() {
		Convey("convertDeviceTypeToAPIFormat: valid types", func() {
			for _, deviceType := range []string{
				"DEVICE_TYPE_UNSPECIFIED",
				"DEVICE_TYPE_VIRTUAL",
				"DEVICE_TYPE_PHYSICAL",
			} {
				apiType := convertDeviceTypeToAPIFormat(ctx, deviceType)
				So(apiType, ShouldEqual, api.DeviceType_value[deviceType])
			}
		})
		Convey("convertDeviceTypeToAPIFormat: unknown type", func() {
			apiType := convertDeviceTypeToAPIFormat(ctx, "UNKNOWN_TYPE")
			So(apiType, ShouldEqual, api.DeviceType_DEVICE_TYPE_UNSPECIFIED)
		})
	})
}

func TestConvertDeviceStateToAPIFormat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("convertDeviceStateToAPIFormat", t, func() {
		Convey("convertDeviceStateToAPIFormat: valid types", func() {
			for _, deviceState := range []string{
				"DEVICE_STATE_UNSPECIFIED",
				"DEVICE_STATE_AVAILABLE",
				"DEVICE_STATE_LEASED",
			} {
				apiState := convertDeviceStateToAPIFormat(ctx, deviceState)
				So(apiState, ShouldEqual, api.DeviceState_value[deviceState])
			}
		})
		Convey("convertDeviceStateToAPIFormat: unknown state", func() {
			apiState := convertDeviceStateToAPIFormat(ctx, "UNKNOWN_STATE")
			So(apiState, ShouldEqual, api.DeviceState_DEVICE_STATE_UNSPECIFIED)
		})
	})
}
