// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.chromium.org/luci/common/logging"
)

// Device contains a single row from the Devices table in the database.
type Device struct {
	ID                string
	DeviceAddress     string
	DeviceType        string
	DeviceState       string
	SchedulableLabels SchedulableLabels `json:"SchedulableLabels"`

	LastUpdatedTime time.Time
	IsActive        bool
}

// LabelValues is the struct containing an array of label values.
type LabelValues struct {
	Values []string
}

// SchedulableLabels is made up of a label key and LabelValues.
type SchedulableLabels map[string]LabelValues

// GormDataType expresses SchedulableLabels as a gorm type to db.
func (SchedulableLabels) GormDataType() string {
	return "JSONB"
}

// Scan implements scanner interface for SchedulableLabels.
func (s *SchedulableLabels) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		bytes = []byte(`{}`)
	}
	err := json.Unmarshal(bytes, s)
	return err
}

// Value implements Valuer interface for SchedulableLabels.
func (s SchedulableLabels) Value() (driver.Value, error) {
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

// GetDeviceByName gets a Device from the database by name.
func GetDeviceByName(ctx context.Context, db *sql.DB, deviceName string) (Device, error) {
	var (
		device          Device
		lastUpdatedTime sql.NullTime
	)
	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			device_address,
			device_type,
			device_state,
			schedulable_labels,
			last_updated_time,
			is_active
		FROM "Devices"
		WHERE id=$1;`, deviceName).Scan(
		&device.ID,
		&device.DeviceAddress,
		&device.DeviceType,
		&device.DeviceState,
		&device.SchedulableLabels,
		&lastUpdatedTime,
		&device.IsActive,
	)
	// TODO (b/328662436): Collect metrics on results
	if err != nil {
		logging.Errorf(ctx, "GetDeviceByName: failed to get Device %s: %s", deviceName, err)
		return device, err
	}

	// Handle possible null times
	if lastUpdatedTime.Valid {
		device.LastUpdatedTime = lastUpdatedTime.Time
	}

	return device, nil
}

// ListDevices retrieves Devices with pagination.
func ListDevices(ctx context.Context, db *sql.DB, pageNumber, pageSize int) ([]Device, error) {
	// handle potential errors for negative page numbers or page sizes
	if pageNumber < 0 || pageSize <= 0 {
		return nil, errors.New("ListDevices: invalid pagination parameters")
	}
	offset := pageNumber * pageSize

	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			device_address,
			device_type,
			device_state,
			schedulable_labels,
			last_updated_time,
			is_active
		FROM "Devices"
		LIMIT $1
		OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Device
	for rows.Next() {
		var (
			device          Device
			lastUpdatedTime sql.NullTime
		)
		err := rows.Scan(
			&device.ID,
			&device.DeviceAddress,
			&device.DeviceType,
			&device.DeviceState,
			&device.SchedulableLabels,
			&lastUpdatedTime,
			&device.IsActive,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, device)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateDevice updates a Device in a transaction.
//
// UpdateDevice uses COALESCE to only update fields with provided values. If
// there is no value provided, then it will use the current value of the device
// field in the db.
func UpdateDevice(ctx context.Context, tx *sql.Tx, updatedDevice Device) error {
	result, err := tx.ExecContext(ctx, `
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
			id=$1;`,
		updatedDevice.ID,
		updatedDevice.DeviceAddress,
		updatedDevice.DeviceType,
		updatedDevice.DeviceState,
		updatedDevice.SchedulableLabels,
		updatedDevice.LastUpdatedTime,
		updatedDevice.IsActive,
	)
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

// UpsertDevice upserts a Device in a transaction.
//
// UpsertDevice will attempt to insert a Device into the db. On conflict of
// the ID, the old device record will be updated with the new information.
func UpsertDevice(ctx context.Context, db *sql.DB, device Device) error {
	result, err := db.ExecContext(ctx, `
		INSERT INTO "Devices"
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
			device_address=COALESCE($2, EXCLUDED.device_address),
			device_type=COALESCE($3, EXCLUDED.device_type),
			device_state=COALESCE($4, EXCLUDED.device_state),
			schedulable_labels=COALESCE($5, EXCLUDED.schedulable_labels),
			last_updated_time=COALESCE($6, EXCLUDED.last_updated_time),
			is_active=COALESCE($7, EXCLUDED.is_active);`,
		device.ID,
		device.DeviceAddress,
		device.DeviceType,
		device.DeviceState,
		device.SchedulableLabels,
		device.LastUpdatedTime,
		device.IsActive,
	)
	if err != nil {
		logging.Errorf(ctx, "UpsertDevice: failed to upsert Device %s: %s", device.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.Errorf(ctx, "UpsertDevice: error getting rows affected: %s", err)
	}

	logging.Debugf(ctx, "UpsertDevice: Device %s upserted successfully (%d row affected)", device.ID, rowsAffected)
	return nil
}
