// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"regexp"
	"time"
)

// TestVariantBQExportJobDuration is the duration that test variant bq export
// job uses. In each job, it exports data of the test variants and their
// verdicts within the duration to BigQuery.
// The cron job of exporting test variant rows runs once per day.
const TestVariantBQExportJobDuration = 24 * time.Hour

// ProjectRe matches validly formed LUCI Project names.
// From https://source.chromium.org/chromium/infra/infra/+/main:luci/appengine/components/components/config/common.py?q=PROJECT_ID_PATTERN
var ProjectRe = regexp.MustCompile(`^[a-z0-9\-]{1,40}$`)
