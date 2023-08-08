// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type AssetInstanceEntity struct {

	// Unique identifier of the AssetInstance
	AssetInstanceId string `gae:"$id"`
	// AssetId associated with the AssetInstance
	AssetId string
	// Status of the AssetInstance
	Status string
	// Deployment Errors
	Errors string `gae:",noindex"`
	// Deployment Logs
	Logs string `gae:",noindex"`
	// Project Id associated with the AssetInstance
	ProjectId string
	// User who created the record.
	CreatedBy string
	// Timestamp for the creation of the record.
	CreatedAt time.Time
	// Timestamp for the last update of the record.
	ModifiedAt time.Time
	// User who modified the record.
	ModifiedBy string
	// Timestamp to delete the machines
	DeleteAt time.Time
}
