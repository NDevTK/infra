// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/config/go/test/api"
	. "go.chromium.org/luci/common/testing/assertions"
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
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			rows := sqlmock.NewRows([]string{"id", "device_address", "device_type", "device_state"}).
				AddRow("test-device-1", "1.1.1.1:1", "DEVICE_TYPE_PHYSICAL", "DEVICE_STATE_AVAILABLE").
				AddRow("test-device-2", "2.2.2.2:2", "DEVICE_TYPE_VIRTUAL", "DEVICE_STATE_LEASED")

			mock.ExpectPrepare(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state
				FROM "Devices"
				WHERE id=$1;`)).
				ExpectQuery().
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
				err = db.Close()
				if err != nil {
					t.Fatalf("failed to close db: %s", err)
				}
			}()

			mock.ExpectPrepare(regexp.QuoteMeta(`
				SELECT
					id,
					device_address,
					device_type,
					device_state
				FROM "Devices"
				WHERE id=$1;`)).
				ExpectQuery().
				WithArgs("test-device-2").
				WillReturnError(fmt.Errorf("GetDevice: failed to get Device"))

			device, err := GetDevice(ctx, db, "test-device-2")
			So(err, ShouldNotBeNil)
			So(device, ShouldResembleProto, &api.Device{})
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
