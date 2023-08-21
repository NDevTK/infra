// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import (
	"time"
)

type FinditConfig struct {
	ActionSettings            string    `datastore:"action_settings"`
	BuildersToTrybots         string    `datastore:"builders_to_trybots"`
	CheckFlakeSettings        string    `datastore:"check_flake_settings"`
	CheckFlakeTryJobSettings  string    `datastore:"check_flake_try_job_settings"`
	CodeCoverageSettings      []byte    `datastore:"code_coverage_settings"`
	CodeReviewSettings        string    `datastore:"code_review_settings"`
	DownloadBuildDataSettings string    `datastore:"download_build_data_settings"`
	FlakeDetectionSettings    string    `datastore:"flake_detection_settings"`
	MastersToBlacklistedSteps string    `datastore:"masters_to_blacklisted_steps"`
	Message                   string    `datastore:"message"`
	StepsForMastersRules      string    `datastore:"steps_for_masters_rules"`
	SwarmingSettings          string    `datastore:"swarming_settings"`
	TryJobSettings            string    `datastore:"try_job_settings"`
	UpdatedBy                 string    `datastore:"updated_by"`
	UpdatedTs                 time.Time `datastore:"updated_ts"`
}
