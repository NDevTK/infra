// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type ResourceEntity struct {
	// Unique identifier of the resource
	ResourceId string `gae:"$id"`
	// Name of the resource
	Name string
	// Description of the resource
	Description string
	// Type of the resource
	Type string
	//  Operating system of the machine (If Type is machine or ad_joined_machine)
	OperatingSystem string
	// TODO: crbug/1328854 move the image info as part of property later phases
	// image associated to the machine or ad_joined_machine
	Image string
	// Gcp project where image resides
	ImageProject string
	// Family accociated with the gcp project where image resides
	ImageFamily string
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
