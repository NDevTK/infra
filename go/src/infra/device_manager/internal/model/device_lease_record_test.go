// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"

	. "go.chromium.org/luci/common/testing/assertions"
)

func TestCreateDeviceLeaseRecord(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("CreateDeviceLeaseRecord", t, func() {
		Convey("CreateDeviceLeaseRecord: valid insert", func() {
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
				INSERT INTO "DeviceLeaseRecords"
					(id, idempotency_key, device_id, device_address, device_type, owner_id,
					leased_time, expiration_time, last_updated_time)
				VALUES
					($1, $2, $3, $4, $5, $6, $7, $8, $9);`)).
				WithArgs(
					"test-lease-record-1",
					"fe20140c-b1aa-4953-90fc-d15677df0c6a",
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"test-owner-id-1",
					timeNow,
					timeNow,
					timeNow,
				).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = CreateDeviceLeaseRecord(ctx, tx, DeviceLeaseRecord{
				ID:              "test-lease-record-1",
				IdempotencyKey:  "fe20140c-b1aa-4953-90fc-d15677df0c6a",
				DeviceID:        "test-device-1",
				DeviceAddress:   "1.1.1.1:1",
				DeviceType:      "DEVICE_TYPE_PHYSICAL",
				OwnerID:         "test-owner-id-1",
				LeasedTime:      timeNow,
				ExpirationTime:  timeNow,
				LastUpdatedTime: timeNow,
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestGetDeviceLeaseRecordByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("GetDeviceLeaseRecordByID", t, func() {
		Convey("GetDeviceLeaseRecordByID: valid return", func() {
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
				"idempotency_key",
				"device_id",
				"device_address",
				"device_type",
				"owner_id",
				"leased_time",
				"released_time",
				"expiration_time",
				"last_updated_time"}).
				AddRow(
					"test-lease-record-1",
					"fe20140c-b1aa-4953-90fc-d15677df0c6a",
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"test-owner-id-1",
					timeNow,
					timeNow,
					timeNow,
					timeNow,
				)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					idempotency_key,
					device_id,
					device_address,
					device_type,
					owner_id,
					leased_time,
					released_time,
					expiration_time,
					last_updated_time
				FROM "DeviceLeaseRecords"
				WHERE id=$1;`)).
				WithArgs("test-lease-record-1").
				WillReturnRows(rows)

			record, err := GetDeviceLeaseRecordByID(ctx, db, "test-lease-record-1")
			So(err, ShouldBeNil)
			So(record, ShouldEqual, DeviceLeaseRecord{
				ID:              "test-lease-record-1",
				IdempotencyKey:  "fe20140c-b1aa-4953-90fc-d15677df0c6a",
				DeviceID:        "test-device-1",
				DeviceAddress:   "1.1.1.1:1",
				DeviceType:      "DEVICE_TYPE_PHYSICAL",
				OwnerID:         "test-owner-id-1",
				LeasedTime:      timeNow,
				ReleasedTime:    timeNow,
				ExpirationTime:  timeNow,
				LastUpdatedTime: timeNow,
			})
		})
		Convey("GetDeviceLeaseRecordByID: no record found", func() {
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
				"idempotency_key",
				"device_id",
				"device_address",
				"device_type",
				"owner_id",
				"leased_time",
				"released_time",
				"expiration_time",
				"last_updated_time"})

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					idempotency_key,
					device_id,
					device_address,
					device_type,
					owner_id,
					leased_time,
					released_time,
					expiration_time,
					last_updated_time
				FROM "DeviceLeaseRecords"
				WHERE id=$1;`)).
				WithArgs("test-lease-record-1").
				WillReturnRows(rows)

			record, err := GetDeviceLeaseRecordByID(ctx, db, "test-lease-record-1")
			So(err, ShouldErrLike, "no rows in result set")
			So(record, ShouldEqual, DeviceLeaseRecord{})
		})
	})
}

func TestGetDeviceLeaseRecordByIdemKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("GetDeviceLeaseRecordByIdemKey", t, func() {
		Convey("GetDeviceLeaseRecordByIdemKey: valid return", func() {
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
				"idempotency_key",
				"device_id",
				"device_address",
				"device_type",
				"owner_id",
				"leased_time",
				"released_time",
				"expiration_time",
				"last_updated_time"}).
				AddRow(
					"test-lease-record-1",
					"fe20140c-b1aa-4953-90fc-d15677df0c6a",
					"test-device-1",
					"1.1.1.1:1",
					"DEVICE_TYPE_PHYSICAL",
					"test-owner-id-1",
					timeNow,
					timeNow,
					timeNow,
					timeNow,
				)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					idempotency_key,
					device_id,
					device_address,
					device_type,
					owner_id,
					leased_time,
					released_time,
					expiration_time,
					last_updated_time
				FROM "DeviceLeaseRecords"
				WHERE idempotency_key=$1;`)).
				WithArgs("fe20140c-b1aa-4953-90fc-d15677df0c6a").
				WillReturnRows(rows)

			record, err := GetDeviceLeaseRecordByIdemKey(ctx, db, "fe20140c-b1aa-4953-90fc-d15677df0c6a")
			So(err, ShouldBeNil)
			So(record, ShouldEqual, DeviceLeaseRecord{
				ID:              "test-lease-record-1",
				IdempotencyKey:  "fe20140c-b1aa-4953-90fc-d15677df0c6a",
				DeviceID:        "test-device-1",
				DeviceAddress:   "1.1.1.1:1",
				DeviceType:      "DEVICE_TYPE_PHYSICAL",
				OwnerID:         "test-owner-id-1",
				LeasedTime:      timeNow,
				ReleasedTime:    timeNow,
				ExpirationTime:  timeNow,
				LastUpdatedTime: timeNow,
			})
		})
		Convey("GetDeviceLeaseRecordByIdemKey: no record found", func() {
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
				"idempotency_key",
				"device_id",
				"device_address",
				"device_type",
				"owner_id",
				"leased_time",
				"released_time",
				"expiration_time",
				"last_updated_time"})

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					idempotency_key,
					device_id,
					device_address,
					device_type,
					owner_id,
					leased_time,
					released_time,
					expiration_time,
					last_updated_time
				FROM "DeviceLeaseRecords"
				WHERE idempotency_key=$1;`)).
				WithArgs("fe20140c-b1aa-4953-90fc-d15677df0c6a").
				WillReturnRows(rows)

			record, err := GetDeviceLeaseRecordByIdemKey(ctx, db, "fe20140c-b1aa-4953-90fc-d15677df0c6a")
			So(err, ShouldErrLike, "no rows in result set")
			So(record, ShouldEqual, DeviceLeaseRecord{})
		})
	})
}

func TestUpdateDeviceLeaseRecord(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("UpdateDeviceLeaseRecord", t, func() {
		Convey("UpdateDeviceLeaseRecord: valid update", func() {
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
					"DeviceLeaseRecords"
				SET
					released_time=COALESCE($2, released_time),
					expiration_time=COALESCE($3, expiration_time),
					last_updated_time=COALESCE($4, last_updated_time)
				WHERE
					id=$1;`)).
				WithArgs(
					"test-lease-record-1",
					timeNow.Add(time.Second*600),
					timeNow.Add(time.Second*600),
					timeNow).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = UpdateDeviceLeaseRecord(ctx, tx, DeviceLeaseRecord{
				ID:              "test-lease-record-1",
				ReleasedTime:    timeNow.Add(time.Second * 600),
				ExpirationTime:  timeNow.Add(time.Second * 600),
				LastUpdatedTime: timeNow,
			})
			So(err, ShouldBeNil)
		})
	})
}
