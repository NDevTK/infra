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
	var device Device
	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			device_address,
			device_type,
			device_state
		FROM "Devices"
		WHERE id=$1;`, deviceName).Scan(
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
