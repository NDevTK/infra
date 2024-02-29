// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/luci/common/logging"

	"infra/device_manager/internal/model"
)

// LeaseDevice leases a device specified by the request.
//
// The function executes as a transaction. It attempts to create a lease record
// with an available device. Then it updates the Device's state to LEASED
// and publishes to a PubSub stream. The transaction is then committed.
func LeaseDevice(ctx context.Context, db *sql.DB, r *api.LeaseDeviceRequest, device *api.Device) (*api.LeaseDeviceResponse, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.New("LeaseDevice: failed to start database transaction")
	}

	timeNow := time.Now()
	newRecord := model.DeviceLeaseRecord{
		ID:              uuid.New().String(),
		IdempotencyKey:  r.GetIdempotencyKey(),
		DeviceID:        device.GetId(),
		DeviceAddress:   convertAPIDeviceAddressToDBFormat(ctx, device.GetAddress()),
		DeviceType:      device.GetType().String(),
		LeasedTime:      timeNow,
		ExpirationTime:  timeNow.Add(r.GetLeaseDuration().AsDuration()),
		LastUpdatedTime: timeNow,
	}

	err = model.CreateDeviceLeaseRecord(ctx, tx, newRecord)
	if err != nil {
		logging.Errorf(ctx, "LeaseDevice: failed to create DeviceLeaseRecord %s", err)
		return nil, err
	}

	updatedDevice := model.Device{
		ID:            device.GetId(),
		DeviceAddress: convertAPIDeviceAddressToDBFormat(ctx, device.GetAddress()),
		DeviceType:    device.GetType().String(),
		DeviceState:   api.DeviceState_DEVICE_STATE_LEASED.String(),
	}
	err = UpdateDevice(ctx, tx, updatedDevice)
	if err != nil {
		logging.Errorf(ctx, "LeaseDevice: failed to update device state %s", err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	logging.Debugf(ctx, "LeaseDevice: created DeviceLeaseRecord %v", newRecord)
	return &api.LeaseDeviceResponse{
		DeviceLease: &api.DeviceLeaseRecord{
			Id:             newRecord.ID,
			IdempotencyKey: newRecord.IdempotencyKey,
			DeviceId:       newRecord.DeviceID,
			DeviceAddress: &api.DeviceAddress{
				Host: newRecord.DeviceAddress,
			},
			DeviceType:      api.DeviceType_DEVICE_TYPE_PHYSICAL,
			LeasedTime:      timestamppb.New(newRecord.LeasedTime),
			ReleasedTime:    timestamppb.New(newRecord.ReleasedTime),
			ExpirationTime:  timestamppb.New(newRecord.ExpirationTime),
			LastUpdatedTime: timestamppb.New(newRecord.LastUpdatedTime),
		},
	}, nil
}

// CheckLeaseIdempotency checks if there is a record with the same idempotency key.
//
// If there is an unexpired record, it will return the record. If it is expired,
// it will error. If there is no record, it will return an empty response and no
// error.
func CheckLeaseIdempotency(ctx context.Context, db *sql.DB, idemKey string) (*api.LeaseDeviceResponse, error) {
	timeNow := time.Now()
	existingRecord, err := model.GetDeviceLeaseRecordByIdemKey(ctx, db, idemKey)
	if err == nil {
		if existingRecord.ExpirationTime.After(timeNow) {
			addr, err := convertDeviceAddressToAPIFormat(ctx, existingRecord.DeviceAddress)
			if err != nil {
				addr = &api.DeviceAddress{}
			}

			return &api.LeaseDeviceResponse{
				DeviceLease: &api.DeviceLeaseRecord{
					Id:              existingRecord.ID,
					IdempotencyKey:  existingRecord.IdempotencyKey,
					DeviceId:        existingRecord.DeviceID,
					DeviceAddress:   addr,
					DeviceType:      api.DeviceType_DEVICE_TYPE_PHYSICAL,
					LeasedTime:      timestamppb.New(existingRecord.LeasedTime),
					ReleasedTime:    timestamppb.New(existingRecord.ReleasedTime),
					ExpirationTime:  timestamppb.New(existingRecord.ExpirationTime),
					LastUpdatedTime: timestamppb.New(existingRecord.LastUpdatedTime),
				},
			}, nil
		} else {
			return &api.LeaseDeviceResponse{}, errors.New("CheckLeaseIdempotency: DeviceLeaseRecord found with same idempotency key but is already expired")
		}
	}
	return &api.LeaseDeviceResponse{}, nil
}
