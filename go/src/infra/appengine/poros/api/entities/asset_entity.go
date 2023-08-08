// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type AssetEntity struct {

	// Unique identifier of the asset
	AssetId string `gae:"$id"`
	// Name of the asset
	Name string
	// Description of the asset
	Description string
	//Type of the Asset (active_directory, etc)
	AssetType string
	// User who created the record.
	CreatedBy string
	// Timestamp for the creation of the record.
	CreatedAt time.Time
	// Timestamp for the last update of the record.
	ModifiedAt time.Time
	// User who modified the record.
	ModifiedBy string
	// Flag to denote whether this Resource is deleted
	Deleted bool
}
