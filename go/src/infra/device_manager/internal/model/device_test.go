// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/common/testing/typed"
)

func TestGetDeviceByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	baseQuery := `
		SELECT
			id,
			device_address,
			device_type,
			device_state,
			schedulable_labels,
			created_time,
			last_updated_time,
			is_active
		FROM "Devices"`

	timeNow := time.Now()
	validCases := []struct {
		name           string
		idType         DeviceIDType
		expectedDevice Device
		err            error
	}{
		{
			name:   "GetDeviceByID: valid return; search by hostname",
			idType: IDTypeHostname,
			expectedDevice: Device{
				ID:            "test-device-1",
				DeviceAddress: "1.1.1.1:1",
				DeviceType:    "DEVICE_TYPE_PHYSICAL",
				DeviceState:   "DEVICE_STATE_AVAILABLE",
				SchedulableLabels: SchedulableLabels{
					"label-test": LabelValues{
						Values: []string{"test-value-1"},
					},
				},
				CreatedTime:     timeNow,
				LastUpdatedTime: timeNow,
				IsActive:        true,
			},
			err: nil,
		},
		{
			name:   "GetDeviceByID: valid return; search by DUT ID",
			idType: IDTypeDutID,
			expectedDevice: Device{
				ID:            "test-device-1",
				DeviceAddress: "1.1.1.1:1",
				DeviceType:    "DEVICE_TYPE_PHYSICAL",
				DeviceState:   "DEVICE_STATE_AVAILABLE",
				SchedulableLabels: SchedulableLabels{
					"label-test": LabelValues{
						Values: []string{"test-value-1"},
					},
				},
				CreatedTime:     timeNow,
				LastUpdatedTime: timeNow,
				IsActive:        true,
			},
			err: nil,
		},
	}

	for _, tt := range validCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

			query := baseQuery
			switch tt.idType {
			case IDTypeDutID:
				query += `
					WHERE
						jsonb_path_query_array(
							schedulable_labels,
							'$.dut_id.Values[0]'
						) @> to_jsonb($1::text);`
			case IDTypeHostname:
				query += `
					WHERE id=$1;`
			default:
				t.Errorf("unexpected error: id type %s is not supported", tt.idType)
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs("test-device-1").
				WillReturnRows(rows)

			device, err := GetDeviceByID(ctx, db, tt.idType, "test-device-1")
			if !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %v; want: %v", err, tt.err)
			}
			if diff := typed.Got(device).Want(tt.expectedDevice).Diff(); diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}

	failedCases := []struct {
		name           string
		idType         DeviceIDType
		expectedDevice Device
		err            error
	}{
		{
			name:           "invalid request; search by hostname, no device name match",
			idType:         IDTypeHostname,
			expectedDevice: Device{},
			err:            fmt.Errorf("GetDeviceByID: failed to get Device"),
		},
		{
			name:           "invalid request; search by DUT ID, no device name match",
			idType:         IDTypeDutID,
			expectedDevice: Device{},
			err:            fmt.Errorf("GetDeviceByID: failed to get Device"),
		},
	}

	for _, tt := range failedCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

			query := baseQuery
			switch tt.idType {
			case IDTypeDutID:
				query += `
					WHERE
						jsonb_path_query_array(
							schedulable_labels,
							'$.dut_id.Values[0]'
						) @> to_jsonb($1::text);`
			case IDTypeHostname:
				query += `
					WHERE id=$1;`
			default:
				t.Errorf("unexpected error: id type %s is not supported", tt.idType)
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs("test-device-1").
				WillReturnError(fmt.Errorf("GetDeviceByID: failed to get Device"))

			device, err := GetDeviceByID(ctx, db, tt.idType, "test-device-1")
			if err.Error() != tt.err.Error() {
				t.Errorf("unexpected error: %s", err)
			}
			if diff := typed.Got(device).Want(tt.expectedDevice).Diff(); diff != "" {
				t.Errorf("unexpected diff: %s", diff)
			}
		})
	}
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
				pageSize = 1
				timeNow  = time.Now()
			)

			createdTime, err := time.Parse("2006-01-02 15:04:05", "2024-01-01 12:00:00")
			So(err, ShouldBeNil)

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
					createdTime,
					timeNow,
					true).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					createdTime,
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
				ORDER BY created_time
				LIMIT $1;`)).
				WithArgs(pageSize + 1).
				WillReturnRows(rows)

			devices, nextPageToken, err := ListDevices(ctx, db, "", pageSize)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldEqual, PageToken("MjAyNC0wMS0wMVQxMjowMDowMFo="))
			So(devices, ShouldEqual, []Device{
				{
					ID:            "test-device-1",
					DeviceAddress: "1.1.1.1:1",
					DeviceType:    "DEVICE_TYPE_PHYSICAL",
					DeviceState:   "DEVICE_STATE_AVAILABLE",
					SchedulableLabels: SchedulableLabels{
						"label-test": LabelValues{
							Values: []string{"test-value-1"},
						},
					},
					CreatedTime:     createdTime,
					LastUpdatedTime: timeNow,
					IsActive:        true,
				},
			})

			decodedToken, err := DecodePageToken(ctx, PageToken(nextPageToken))
			So(err, ShouldBeNil)
			So(decodedToken, ShouldEqual, createdTime.Format(time.RFC3339Nano))
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
				pageSize = 2
				timeNow  = time.Now()
			)

			createdTime, err := time.Parse("2006-01-02 15:04:05", "2024-01-01 12:00:00")
			So(err, ShouldBeNil)

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
					createdTime,
					timeNow,
					true).
				AddRow(
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					createdTime,
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
				ORDER BY created_time
				LIMIT $1;`)).
				WithArgs(pageSize + 1).
				WillReturnRows(rows)

			devices, nextPageToken, err := ListDevices(ctx, db, "", pageSize)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldEqual, PageToken(""))
			So(devices, ShouldEqual, []Device{
				{
					ID:            "test-device-1",
					DeviceAddress: "1.1.1.1:1",
					DeviceType:    "DEVICE_TYPE_PHYSICAL",
					DeviceState:   "DEVICE_STATE_AVAILABLE",
					SchedulableLabels: SchedulableLabels{
						"label-test": LabelValues{
							Values: []string{"test-value-1"},
						},
					},
					CreatedTime:     createdTime,
					LastUpdatedTime: timeNow,
					IsActive:        true,
				},
				{
					ID:            "test-device-2",
					DeviceAddress: "2.2.2.2:2",
					DeviceType:    "DEVICE_TYPE_VIRTUAL",
					DeviceState:   "DEVICE_STATE_LEASED",
					SchedulableLabels: SchedulableLabels{
						"label-test": LabelValues{
							Values: []string{"test-value-2"},
						},
					},
					CreatedTime:     createdTime,
					LastUpdatedTime: timeNow,
					IsActive:        false,
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
				pageSize  = 1
				pageToken = "MjAyNC0wMS0wMVQxMjowMDowMFo="
				timeNow   = time.Now()
			)

			createdTime, err := time.Parse("2006-01-02 15:04:05", "2024-01-01 12:00:00")
			So(err, ShouldBeNil)

			// only add rows after test-device-1
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
					"test-device-2",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-2"]}}`,
					createdTime,
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
				WHERE created_time > $1
				ORDER BY created_time
				LIMIT $2;`)).
				WithArgs(createdTime.Format(time.RFC3339Nano), pageSize+1).
				WillReturnRows(rows)

			devices, nextPageToken, err := ListDevices(ctx, db, PageToken(pageToken), pageSize)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldEqual, PageToken(""))
			So(devices, ShouldEqual, []Device{
				{
					ID:            "test-device-2",
					DeviceAddress: "2.2.2.2:2",
					DeviceType:    "DEVICE_TYPE_VIRTUAL",
					DeviceState:   "DEVICE_STATE_LEASED",
					SchedulableLabels: SchedulableLabels{
						"label-test": LabelValues{
							Values: []string{"test-value-2"},
						},
					},
					CreatedTime:     createdTime,
					LastUpdatedTime: timeNow,
					IsActive:        false,
				},
			})
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

func TestUpsertDevice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("UpsertDevice", t, func() {
		Convey("UpsertDevice: valid upsert", func() {
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
			mock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO "Devices" AS d
					(
						id,
						device_address,
						device_type,
						device_state,
						schedulable_labels,
						last_updated_time,
						is_active
					)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT(id)
				DO UPDATE SET
					device_address=COALESCE(EXCLUDED.device_address, d.device_address),
					device_type=COALESCE(EXCLUDED.device_type, d.device_type),
					device_state=COALESCE(EXCLUDED.device_state, d.device_state),
					schedulable_labels=COALESCE(EXCLUDED.schedulable_labels, d.schedulable_labels),
					last_updated_time=COALESCE(EXCLUDED.last_updated_time, d.last_updated_time),
					is_active=COALESCE(EXCLUDED.is_active, d.is_active);`)).
				WithArgs(
					"test-device-1",
					"2.2.2.2:2",
					"DEVICE_TYPE_VIRTUAL",
					"DEVICE_STATE_LEASED",
					`{"label-test":{"Values":["test-value-1"]}}`,
					timeNow,
					false).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = UpsertDevice(ctx, db, Device{
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
