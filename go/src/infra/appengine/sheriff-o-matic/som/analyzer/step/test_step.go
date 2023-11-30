// Copyright 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package step

import (
	"time"
)

// ArtifactLink is a link to a test artifact left by perf tests.
type ArtifactLink struct {
	// Name is the name of the artifact.
	Name string `json:"name"`
	// Location is the location of the artifact.
	Location string `json:"location"`
}

// TestWithResult stores the information for a specific test,
// for example if the test is flaky or is there a culprit for the test failure.
// Also contains test-specific details like expectations and any artifacts
// produced by the test run.
type TestWithResult struct {
	TestName    string `json:"test_name"`
	TestID      string `json:"test_id"`
	Realm       string `json:"realm"`
	VariantHash string `json:"variant_hash"`
	RefHash     string `json:"ref_hash"`
	ClusterName string `json:"cluster_name"`
	// Start commit position of the regression range exclusive.
	RegressionStartPosition int64 `json:"regression_start_position"`
	// End commit position of the regression range inclusive.
	RegressionEndPosition int64 `json:"regression_end_position"`
	// The approximation of the start hour of the current segment.
	// See https://source.chromium.org/chromium/infra/infra/+/main:go/src/go.chromium.org/luci/analysis/proto/bq/test_variant_branch_row.proto;l=113
	CurStartHour time.Time `json:"cur_start_hour"`
	// The approximation of the end hour of the previous segment.
	// See https://source.chromium.org/chromium/infra/infra/+/main:go/src/go.chromium.org/luci/analysis/proto/bq/test_variant_branch_row.proto;l=124
	PrevEndHour time.Time `json:"prev_end_hour"`
	// Statistics for the current segments from changepoint analysis.
	CurCounts Counts `json:"cur_counts"`
	// Statistics for the previous segments from changepoint analysis.
	PrevCounts Counts `json:"prev_counts"`
}

type Counts struct {
	UnexpectedResults int64 `json:"unexpected_results"`
	TotalResults      int64 `json:"total_results"`
}
