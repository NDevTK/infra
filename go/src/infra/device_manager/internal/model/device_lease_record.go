// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package model

import (
	"context"
	"database/sql"
	"time"

	"go.chromium.org/luci/common/logging"
)

// DeviceLeaseRecordData is used to pass data to the HTML template.
type DeviceLeaseRecordData struct {
	Records []DeviceLeaseRecord
}

// DeviceLeaseRecord contains a single row from the DeviceLeaseRecords table in
// the database.
type DeviceLeaseRecord struct {
	ID              string
	IdempotencyKey  string
	DeviceID        string
	DeviceAddress   string
	DeviceType      string
	OwnerID         string
	LeasedTime      time.Time
	ReleasedTime    time.Time
	ExpirationTime  time.Time
	LastUpdatedTime time.Time
}

// CreateDeviceLeaseRecord creates a DeviceLeaseRecord in the database.
func CreateDeviceLeaseRecord(ctx context.Context, tx *sql.Tx, record DeviceLeaseRecord) error {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO "DeviceLeaseRecords"
			(id, idempotency_key, device_id, device_address, device_type, owner_id,
			 leased_time, expiration_time, last_updated_time)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		record.ID,
		record.IdempotencyKey,
		record.DeviceID,
		record.DeviceAddress,
		record.DeviceType,
		record.OwnerID,
		record.LeasedTime,
		record.ExpirationTime,
		record.LastUpdatedTime,
	)
	if err != nil {
		logging.Errorf(ctx, "CreateDeviceLeaseRecord: error inserting into DeviceLeaseRecords: %s", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logging.Errorf(ctx, "CreateDeviceLeaseRecord: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.Errorf(ctx, "CreateDeviceLeaseRecord: error getting rows affected: %s", err)
	}

	logging.Debugf(ctx, "CreateDeviceLeaseRecord: DeviceLeaseRecord %s for Device %s created successfully (%d row affected)", record.ID, record.DeviceID, rowsAffected)
	return nil
}

// GetDeviceLeaseRecordByID gets a DeviceLeaseRecord from the database by name.
func GetDeviceLeaseRecordByID(ctx context.Context, db *sql.DB, recordID string) (DeviceLeaseRecord, error) {
	var (
		record          DeviceLeaseRecord
		leasedTime      sql.NullTime
		releasedTime    sql.NullTime
		expirationTime  sql.NullTime
		lastUpdatedTime sql.NullTime
	)

	err := db.QueryRowContext(ctx, `
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
		WHERE id=$1;`, recordID).Scan(
		&record.ID,
		&record.IdempotencyKey,
		&record.DeviceID,
		&record.DeviceAddress,
		&record.DeviceType,
		&record.OwnerID,
		&leasedTime,
		&releasedTime,
		&expirationTime,
		&lastUpdatedTime,
	)
	if err != nil {
		logging.Errorf(ctx, "GetDeviceLeaseRecordByID: failed to get DeviceLeaseRecord %s: %s", recordID, err)
		return record, err
	}

	// Handle possible null times
	if leasedTime.Valid {
		record.LeasedTime = leasedTime.Time
	}
	if releasedTime.Valid {
		record.ReleasedTime = releasedTime.Time
	}
	if expirationTime.Valid {
		record.ExpirationTime = expirationTime.Time
	}
	if lastUpdatedTime.Valid {
		record.LastUpdatedTime = lastUpdatedTime.Time
	}

	logging.Debugf(ctx, "GetDeviceLeaseRecordByID: success: %v", record)
	return record, nil
}

// GetDeviceLeaseRecordByIdemKey gets a DeviceLeaseRecord from the database by idempotency key.
func GetDeviceLeaseRecordByIdemKey(ctx context.Context, db *sql.DB, idemKey string) (DeviceLeaseRecord, error) {
	var (
		record          DeviceLeaseRecord
		leasedTime      sql.NullTime
		releasedTime    sql.NullTime
		expirationTime  sql.NullTime
		lastUpdatedTime sql.NullTime
	)

	err := db.QueryRowContext(ctx, `
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
		WHERE idempotency_key=$1;`, idemKey).Scan(
		&record.ID,
		&record.IdempotencyKey,
		&record.DeviceID,
		&record.DeviceAddress,
		&record.DeviceType,
		&record.OwnerID,
		&leasedTime,
		&releasedTime,
		&expirationTime,
		&lastUpdatedTime,
	)
	if err != nil {
		logging.Errorf(ctx, "GetDeviceLeaseRecordByIdemKey: failed to get DeviceLeaseRecord with Idempotency Key %s: %s", idemKey, err)
		return record, err
	}

	// Handle possible null times
	if leasedTime.Valid {
		record.LeasedTime = leasedTime.Time
	}
	if releasedTime.Valid {
		record.ReleasedTime = releasedTime.Time
	}
	if expirationTime.Valid {
		record.ExpirationTime = expirationTime.Time
	}
	if lastUpdatedTime.Valid {
		record.LastUpdatedTime = lastUpdatedTime.Time
	}

	logging.Debugf(ctx, "GetDeviceLeaseRecordByIdemKey: success: %v", record)
	return record, nil
}

// UpdateDeviceLeaseRecord updates a lease record in a transaction.
//
// UpdateDeviceLeaseRecord uses COALESCE to only update fields with provided
// values. If there is no value provided, then it will use the current value of
// the device field in the db.
func UpdateDeviceLeaseRecord(ctx context.Context, tx *sql.Tx, updatedRec DeviceLeaseRecord) error {
	var (
		releasedTime    sql.NullTime
		expirationTime  sql.NullTime
		lastUpdatedTime sql.NullTime
	)

	// Handle possible null times
	if !updatedRec.ReleasedTime.IsZero() {
		releasedTime.Time = updatedRec.ReleasedTime
		releasedTime.Valid = true
	}
	if !updatedRec.ExpirationTime.IsZero() {
		expirationTime.Time = updatedRec.ExpirationTime
		expirationTime.Valid = true
	}
	if !updatedRec.LastUpdatedTime.IsZero() {
		lastUpdatedTime.Time = updatedRec.LastUpdatedTime
		lastUpdatedTime.Valid = true
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE
			"DeviceLeaseRecords"
		SET
			released_time=COALESCE($2, released_time),
			expiration_time=COALESCE($3, expiration_time),
			last_updated_time=COALESCE($4, last_updated_time)
		WHERE
			id=$1;`,
		updatedRec.ID,
		releasedTime,
		expirationTime,
		lastUpdatedTime,
	)
	if err != nil {
		logging.Errorf(ctx, "UpdateDeviceLeaseRecord: failed to update DeviceLeaseRecord %s: %s", updatedRec.ID, err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logging.Errorf(ctx, "UpdateDeviceLeaseRecord: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logging.Errorf(ctx, "UpdateDeviceLeaseRecord: error getting rows affected: %s", err)
	}

	logging.Debugf(ctx, "UpdateDeviceLeaseRecord: DeviceLeaseRecord %s updated successfully (%d row affected)", updatedRec.ID, rowsAffected)
	return nil
}
