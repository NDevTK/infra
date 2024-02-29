// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"

	"go.chromium.org/luci/common/logging"
)

// Device contains a single row from the Devices table in the database.
type Device struct {
	ID            string
	DeviceAddress string
	DeviceType    string
	DeviceState   string
}

// GetDeviceByName gets a Device from the database by name.
func GetDeviceByName(ctx context.Context, db *sql.DB, deviceName string) (Device, error) {
	stmt, err := db.PrepareContext(ctx, `
		SELECT
			id,
			device_address,
			device_type,
			device_state
		FROM "Devices"
		WHERE id=$1;`)
	if err != nil {
		logging.Errorf(ctx, "GetDeviceByName: failed to prepare select statement: %s", err)
		return Device{}, err
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			logging.Debugf(ctx, "GetDeviceByName: failed to close statement: %s", err)
		}
	}()

	var device Device
	err = stmt.QueryRowContext(ctx, deviceName).Scan(
		&device.ID,
		&device.DeviceAddress,
		&device.DeviceType,
		&device.DeviceState,
	)
	if err != nil {
		logging.Errorf(ctx, "GetDeviceByName: failed to get Device %s: %s", deviceName, err)
		return device, err
	}

	return device, nil
}

// UpdateDevice updates the state of a Device in a transaction.
//
// UpdateDevice uses COALESCE to only update fields with provided values. If
// there is no value provided, then it will use the current value of the device
// field in the db.
func UpdateDevice(ctx context.Context, tx *sql.Tx, updatedDevice Device) error {
	stmt, err := tx.PrepareContext(ctx, `
		UPDATE
			"Devices"
		SET
			device_address=COALESCE($2, device_address),
			device_type=COALESCE($3, device_type),
			device_state=COALESCE($4, device_state)
		WHERE
			id=$1;`)
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: failed to prepare update statement: %s", err)
		return err
	}
	defer func() {
		err = stmt.Close()
		if err != nil {
			logging.Debugf(ctx, "UpdateDevice: failed to close statement: %s", err)
		}
	}()

	result, err := stmt.ExecContext(ctx, updatedDevice.ID,
		updatedDevice.DeviceAddress, updatedDevice.DeviceType,
		updatedDevice.DeviceState)
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: failed to update Device %s: %s", updatedDevice.ID, err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logging.Errorf(ctx, "UpdateDevice: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.Errorf(ctx, "UpdateDevice: error getting rows affected: %s", err)
	}

	logging.Debugf(ctx, "UpdateDevice: Device %s updated successfully (%d row affected)", updatedDevice.ID, rowsAffected)
	return nil
}
