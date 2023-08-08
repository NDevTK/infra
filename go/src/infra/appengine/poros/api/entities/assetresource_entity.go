// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type AssetResourceEntity struct {
	// Unique identifier of the entity
	AssetResourceId string `gae:"$id"`
	// Identifier of the asset associated with the entity
	AssetId string
	// Identifier of the resource associated with the entity
	ResourceId string
	// Alias name of the entity
	AliasName string
	// User who created the record.
	CreatedBy string
	// Timestamp for the creation of the record.
	CreatedAt time.Time
	// Timestamp for the last update of the record.
	ModifiedAt time.Time
	// User who modified the record.
	ModifiedBy string
	// Flag to denote whether this AssetResource is default
	Default bool
}
