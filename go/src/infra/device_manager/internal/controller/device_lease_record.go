// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cloud.google.com/go/pubsub"
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
func LeaseDevice(ctx context.Context, db *sql.DB, psClient *pubsub.Client, r *api.LeaseDeviceRequest, device *api.Device) (*api.LeaseDeviceResponse, error) {
	// TODO (b/328662436): Collect metrics
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
	err = UpdateDevice(ctx, tx, psClient, updatedDevice)
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

// ExtendLease attempts to extend the lease on a device.
//
// ExtendLease checks the requested lease to verify that it is unexpired. If
// unexpired, it will extend the lease by the requested duration. This maintains
// the leased state on a device.
func ExtendLease(ctx context.Context, db *sql.DB, r *api.ExtendLeaseRequest) (*api.ExtendLeaseResponse, error) {
	// TODO (b/328662436): Collect metrics
	record, err := model.GetDeviceLeaseRecordByID(ctx, db, r.GetLeaseId())
	if err != nil {
		return &api.ExtendLeaseResponse{}, err
	}

	timeNow := time.Now()
	if record.ExpirationTime.Before(timeNow) {
		return &api.ExtendLeaseResponse{
			LeaseId:        r.GetLeaseId(),
			ExpirationTime: timestamppb.New(record.ExpirationTime),
		}, errors.New("ExtendLease: lease is already expired")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.New("ExtendLease: failed to start database transaction")
	}

	// Record ExtendLeaseRequest in DB
	extendDur := r.GetExtendDuration().GetSeconds()
	newExpirationTime := record.ExpirationTime.Add(time.Second * time.Duration(extendDur))
	newRequest := model.ExtendLeaseRequest{
		ID:             uuid.New().String(),
		LeaseID:        r.GetLeaseId(),
		IdempotencyKey: r.GetIdempotencyKey(),
		ExtendDuration: extendDur,
		RequestTime:    timeNow,
		ExpirationTime: newExpirationTime,
	}

	err = model.CreateExtendLeaseRequest(ctx, tx, newRequest)
	if err != nil {
		logging.Errorf(ctx, "ExtendLease: failed to create ExtendLeaseRequest %s", err)
		return nil, err
	}

	// Update DeviceLeaseRecord with new expiration time
	updatedRec := model.DeviceLeaseRecord{
		ID:              r.GetLeaseId(),
		ExpirationTime:  newExpirationTime,
		LastUpdatedTime: timeNow,
	}

	err = model.UpdateDeviceLeaseRecord(ctx, tx, updatedRec)
	if err != nil {
		logging.Errorf(ctx, "ExtendLease: failed to update DeviceLeaseRecord %s: %s", updatedRec.ID, err)
		return nil, err
	}
	logging.Debugf(ctx, "ExtendLease: updated Device %s successfully", updatedRec.ID)

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	logging.Debugf(ctx, "ExtendLease: created ExtendLeaseRequest %v", newRequest)
	return &api.ExtendLeaseResponse{
		LeaseId:        r.GetLeaseId(),
		ExpirationTime: timestamppb.New(newRequest.ExpirationTime),
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

// CheckExtensionIdempotency checks if there is a extend request with the same
// idempotency key.
//
// If there is a duplicate request, it will return the request. If there is no
// record, it will return an empty response and no error.
func CheckExtensionIdempotency(ctx context.Context, db *sql.DB, idemKey string) (*api.ExtendLeaseResponse, error) {
	existingRecord, err := model.GetExtendLeaseRequestByIdemKey(ctx, db, idemKey)
	if err == nil {
		return &api.ExtendLeaseResponse{
			LeaseId:        existingRecord.LeaseID,
			ExpirationTime: timestamppb.New(existingRecord.ExpirationTime),
		}, nil
	}
	return &api.ExtendLeaseResponse{}, nil
}
