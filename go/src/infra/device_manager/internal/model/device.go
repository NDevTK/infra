// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

	CreatedTime     time.Time
	LastUpdatedTime time.Time
	IsActive        bool
}

// DeviceIDType indicates the type of ID used to identify a Device in DB.
type DeviceIDType string

const (
	IDTypeHostname DeviceIDType = "hostname"
	IDTypeDutID    DeviceIDType = "dut_id"
)

// PageToken is a string containing a page token to a database query.
type PageToken string

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

// GetDeviceByID gets a Device from the database by a type of ID.
func GetDeviceByID(ctx context.Context, db *sql.DB, idType DeviceIDType, deviceID string) (Device, error) {
	var (
		device          Device
		createdTime     sql.NullTime
		lastUpdatedTime sql.NullTime
	)
	query := `
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

	switch idType {
	case IDTypeDutID:
		// Use DUT ID type also known as Asset Tag.
		query += `
			WHERE
				jsonb_path_query_array(
					schedulable_labels,
					'$.dut_id.Values[0]'
				) @> to_jsonb($1::text);`
	case IDTypeHostname:
		// Use hostname which is how they are stored in DB.
		query += `
			WHERE id=$1;`
	default:
		return Device{}, fmt.Errorf("GetDeviceByID: unsupported Device ID type: %s", idType)
	}

	err := db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID,
		&device.DeviceAddress,
		&device.DeviceType,
		&device.DeviceState,
		&device.SchedulableLabels,
		&createdTime,
		&lastUpdatedTime,
		&device.IsActive,
	)

	// TODO (b/328662436): Collect metrics on results
	if err != nil {
		logging.Errorf(ctx, "GetDeviceByID: failed to get Device %s: %s", deviceID, err)
		return device, err
	}

	// Handle possible null times
	if createdTime.Valid {
		device.CreatedTime = createdTime.Time
	}
	if lastUpdatedTime.Valid {
		device.LastUpdatedTime = lastUpdatedTime.Time
	}

	return device, nil
}

// ListDevices retrieves Devices with pagination.
func ListDevices(ctx context.Context, db *sql.DB, pageToken PageToken, pageSize int) ([]Device, PageToken, error) {
	// TODO (b/337086313): Implement filtering
	// handle potential errors for negative page numbers or page sizes
	if pageSize <= 0 {
		return nil, "", errors.New("ListDevices: invalid pagination parameters")
	}

	query := `
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
	args := []interface{}{}
	currArgPosition := 1

	if pageToken != "" {
		decodedTime, err := DecodePageToken(ctx, pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("ListDevices: %w", err)
		}

		query += fmt.Sprintf(`
			WHERE created_time > $%d`, currArgPosition)
		args = append(args, decodedTime)
		currArgPosition += 1
	}

	query += fmt.Sprintf(`
		ORDER BY created_time
		LIMIT $%d;`, currArgPosition)
	args = append(args, pageSize+1) // fetch one extra to check for 'next page'

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("ListDevices: %w", err)
	}
	defer rows.Close()

	var results []Device
	for rows.Next() {
		var (
			device          Device
			createdTime     sql.NullTime
			lastUpdatedTime sql.NullTime
		)
		err := rows.Scan(
			&device.ID,
			&device.DeviceAddress,
			&device.DeviceType,
			&device.DeviceState,
			&device.SchedulableLabels,
			&createdTime,
			&lastUpdatedTime,
			&device.IsActive,
		)
		if err != nil {
			return nil, "", fmt.Errorf("ListDevices: %w", err)
		}

		// handle possible null times
		if createdTime.Valid {
			device.CreatedTime = createdTime.Time
		}
		if lastUpdatedTime.Valid {
			device.LastUpdatedTime = lastUpdatedTime.Time
		}

		results = append(results, device)
	}

	if err := rows.Close(); err != nil {
		return nil, "", fmt.Errorf("ListDevices: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("ListDevices: %w", err)
	}

	// truncate results and use last Device ID as next page token
	var nextPageToken PageToken
	if len(results) > pageSize {
		lastDevice := results[pageSize-1]
		nextPageToken = EncodePageToken(ctx, lastDevice.CreatedTime.Format(time.RFC3339Nano))
		results = results[0:pageSize] // trim results to page size
	}
	return results, nextPageToken, nil
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
			device_address=COALESCE(NULLIF($2, ''), device_address),
			device_type=COALESCE(NULLIF($3, ''), device_type),
			device_state=COALESCE(NULLIF($4, ''), device_state),
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
			device_address=COALESCE(NULLIF(EXCLUDED.device_address, ''), d.device_address),
			device_type=COALESCE(NULLIF(EXCLUDED.device_type, ''), d.device_type),
			device_state=COALESCE(NULLIF(EXCLUDED.device_state, ''), NULLIF(d.device_state, ''), 'DEVICE_STATE_AVAILABLE'),
			schedulable_labels=COALESCE(EXCLUDED.schedulable_labels, d.schedulable_labels),
			last_updated_time=COALESCE(EXCLUDED.last_updated_time, d.last_updated_time),
			is_active=COALESCE(EXCLUDED.is_active, d.is_active);`,
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

// EncodePageToken encodes a Device CreatedTime as a base64 PageToken.
func EncodePageToken(ctx context.Context, createdTime string) PageToken {
	return PageToken(base64.StdEncoding.EncodeToString([]byte(createdTime)))
}

// DecodePageToken decodes a base64 PageToken as a string.
func DecodePageToken(ctx context.Context, token PageToken) (string, error) {
	createdTime, err := base64.StdEncoding.DecodeString(string(token))
	if err != nil {
		return "", fmt.Errorf("DecodePageToken: %w", err)
	}
	return string(createdTime), nil
}
