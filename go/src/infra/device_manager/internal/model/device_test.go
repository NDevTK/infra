// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDeviceByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("GetDeviceByName", t, func() {
		Convey("GetDeviceByName: valid return", func() {
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
				WHERE id=$1;`)).
				WithArgs("test-device-1").
				WillReturnRows(rows)

			device, err := GetDeviceByName(ctx, db, "test-device-1")
			So(err, ShouldBeNil)
			So(device, ShouldEqual, Device{
				ID:            "test-device-1",
				DeviceAddress: "1.1.1.1:1",
				DeviceType:    "DEVICE_TYPE_PHYSICAL",
				DeviceState:   "DEVICE_STATE_AVAILABLE",
				SchedulableLabels: SchedulableLabels{
					"label-test": LabelValues{
						Values: []string{"test-value-1"},
					},
				},
				LastUpdatedTime: timeNow,
				IsActive:        true,
			})
		})
		Convey("GetDeviceByName: invalid request; no device name match", func() {
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
					last_updated_time,
					is_active
				FROM "Devices"
				WHERE id=$1;`)).
				WithArgs("test-device-2").
				WillReturnError(fmt.Errorf("GetDeviceByName: failed to get Device"))

			device, err := GetDeviceByName(ctx, db, "test-device-2")
			So(err, ShouldNotBeNil)
			So(device, ShouldEqual, Device{})
		})
	})
}

func TestUpdateDevice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

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

			err = UpdateDevice(ctx, tx, Device{
				ID:            "test-device-1",
				DeviceAddress: "2.2.2.2:2",
				DeviceType:    "DEVICE_TYPE_VIRTUAL",
				DeviceState:   "DEVICE_STATE_LEASED",
				SchedulableLabels: SchedulableLabels{
					"label-test": LabelValues{
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
