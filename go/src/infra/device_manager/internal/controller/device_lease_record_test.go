// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/chromiumos/config/go/test/api"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestCheckLeaseIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("CheckLeaseIdempotency", t, func() {
		Convey("CheckLeaseIdempotency: valid request", func() {
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
					timeNow.Add(time.Hour*1),
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

			rsp, err := CheckLeaseIdempotency(ctx, db, "fe20140c-b1aa-4953-90fc-d15677df0c6a")
			So(err, ShouldBeNil)
			So(rsp, ShouldResemble, &api.LeaseDeviceResponse{
				DeviceLease: &api.DeviceLeaseRecord{
					Id:             "test-lease-record-1",
					IdempotencyKey: "fe20140c-b1aa-4953-90fc-d15677df0c6a",
					DeviceId:       "test-device-1",
					DeviceAddress: &api.DeviceAddress{
						Host: "1.1.1.1",
						Port: 1,
					},
					DeviceType:      api.DeviceType_DEVICE_TYPE_PHYSICAL,
					LeasedTime:      timestamppb.New(timeNow),
					ReleasedTime:    timestamppb.New(timeNow),
					ExpirationTime:  timestamppb.New(timeNow.Add(time.Hour * 1)),
					LastUpdatedTime: timestamppb.New(timeNow),
				},
			})
		})
		Convey("CheckLeaseIdempotency: invalid request; expired record", func() {
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

			rsp, err := CheckLeaseIdempotency(ctx, db, "fe20140c-b1aa-4953-90fc-d15677df0c6a")
			So(err, ShouldErrLike, "DeviceLeaseRecord found with same idempotency key but is already expired")
			So(rsp, ShouldResemble, &api.LeaseDeviceResponse{})
		})
	})
}

func TestCheckExtensionIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("CheckExtensionIdempotency", t, func() {
		Convey("CheckExtensionIdempotency: valid request", func() {
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
				"lease_id",
				"idempotency_key",
				"extend_duration",
				"request_time",
				"expiration_time"}).
				AddRow(
					"test-extend-record-1",
					"test-lease-record-1",
					"fe20140c-b1aa-4953-90fc-d15677df0c6a",
					600,
					timeNow,
					timeNow.Add(time.Minute*10),
				)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT
					id,
					lease_id,
					idempotency_key,
					extend_duration,
					request_time,
					expiration_time
				FROM "ExtendLeaseRequests"
				WHERE idempotency_key=$1;`)).
				WithArgs("fe20140c-b1aa-4953-90fc-d15677df0c6a").
				WillReturnRows(rows)

			rsp, err := CheckExtensionIdempotency(ctx, db, "fe20140c-b1aa-4953-90fc-d15677df0c6a")
			So(err, ShouldBeNil)
			So(rsp, ShouldResemble, &api.ExtendLeaseResponse{
				LeaseId:        "test-lease-record-1",
				ExpirationTime: timestamppb.New(timeNow.Add(time.Minute * 10)),
			})
		})
	})
}
