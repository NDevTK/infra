// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cron

import (
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/common/tsmon/types"
)

var (
	// Measures how many instance leases are expired.
	expiredLeaseCount = metric.NewInt(
		"vmlab/vm_leaser/leases/expired_count",
		"The total number of expired leases, by GCP project.",
		&types.MetricMetadata{Units: "leases"},
		// The GCP Project where the leases are created.
		field.String("project"),
	)

	// Measures how many instance leases are active in a GCP project.
	activeLeaseCount = metric.NewInt(
		"vmlab/vm_leaser/leases/active_count",
		"The total number of active leases, by GCP project.",
		&types.MetricMetadata{Units: "leases"},
		// The GCP Project where the leases are created.
		field.String("project"),
	)

	// Measures how many instance leases are currently managed in a GCP project.
	totalLeaseCount = metric.NewInt(
		"vmlab/vm_leaser/leases/total_count",
		"The total number of managed leases, by GCP project.",
		&types.MetricMetadata{Units: "leases"},
		// The GCP Project where the leases are created.
		field.String("project"),
	)
)
