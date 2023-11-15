// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execution

import (
	"time"
)

type suiteFilter struct {
	suiteName  string
	expiration time.Time
}

// exceptions stores all granted exceptions from the SuiteLimits project. go/sl-tracking-sheet for more information.
var exceptions = []suiteFilter{
	{
		suiteName:  "arc-cts",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-long",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-camera-opendut",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-hardware",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-qual",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-qual-long",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-vm-stable",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-vm-stable-long",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts-long",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts-qual",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts-qual-long",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full-r",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full-t",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-incremental-r",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-perbuild",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-arc",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-cft-crostini",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-crostini",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-hw",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-criticalstaging",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-informational",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-non-arc-non-crostini",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-parallels-informational",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "fieldtrial-testing-config-on-weekly",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_nightly",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_perbuild",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_weekly",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "flex-perbuild",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "chrome-uprev-hw",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-build",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-day",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-week",
		expiration: time.Date(2024, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	// Release specific exemptions, giving an extra year of time so the
	// exemption doesn't unexpectedly expire.
	{
		suiteName:  "paygen_au_stable",
		expiration: time.Date(2025, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "paygen_au_dev",
		expiration: time.Date(2025, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "paygen_au_beta",
		expiration: time.Date(2025, time.January, 30, 0, 0, 0, 0, time.UTC),
	},
}
