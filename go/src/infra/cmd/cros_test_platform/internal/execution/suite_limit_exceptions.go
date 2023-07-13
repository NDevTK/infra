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
		suiteName:  "appcompat_default",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "appcompat_release",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "appcompat_top_apps",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-camera-opendut",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-hardware",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-qual",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-cts-vm-stable",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-gts-qual",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full-r",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-full-t",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "arc-sts-incremental-r",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bluetooth_sa",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bluetooth_standalone_cq",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "borealis_per-day",
		expiration: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "borealis_per-week",
		expiration: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-perbuild",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},

	{
		suiteName:  "bvt-tast-arc",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq",
		expiration: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-cft-crostini",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-crostini",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-hw",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-criticalstaging",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-informational",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-cq-non-arc-non-crostini",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "bvt-tast-parallels-informational",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_nightly",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_perbuild",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "crosbolt_perf_weekly",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "flex-perbuild",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-build",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-day",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "graphics_per-week",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "stress",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
	{
		suiteName:  "chrome-uprev-hw",
		expiration: time.Date(2023, time.August, 10, 0, 0, 0, 0, time.UTC),
	},
}
