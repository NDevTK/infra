// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"

	"infra/libs/git"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
)

type Thresholds struct {
	Default    BuilderThresholds           `json:"_default"`
	Thresholds map[string]BucketThresholds `json:"thresholds"`
}
type BucketThresholds map[string]BuilderThresholds
type BuilderThresholds struct {
	Default         string               `json:"_default"` // if set to the sentinel value "_default", then use the defaults
	BuildTime       PercentileThresholds `json:"build_time"`
	FailRate        AverageThresholds    `json:"fail_rate"`
	InfraFailRate   AverageThresholds    `json:"infra_fail_rate"`
	PendingTime     PercentileThresholds `json:"pending_time"`
	TestPendingTime PercentileThresholds `json:"test_pending_time"`
}

type PercentileThresholds struct {
	P50Mins float32 `json:"p50_mins"`
	P95Mins float32 `json:"p95_mins"`
	P99Mins float32 `json:"p99_mins"`
}

type AverageThresholds struct {
	Average float32 `json:"average"`
}

func getThresholds(buildCtx context.Context) (*Thresholds, error) {
	var err error
	step, ctx := build.StartStep(buildCtx, "Get thresholds")
	defer func() { step.End(err) }()

	step.SetSummaryMarkdown("Reading thresholds from https://chromium.googlesource.com/chromium/src/+/refs/heads/main/infra/config/generated/health-specs/health-specs.json")

	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, auth.Options{Scopes: []string{gitiles.OAuthScope}})
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "Initializing Auth").Err()
	}

	client, err := git.NewClient(ctx, httpClient, "chromium-review.googlesource.com", "chromium.googlesource.com", "chromium/src", "main")
	if err != nil {
		return nil, errors.Annotate(err, "Initializing Gitiles client").Err()
	}

	thresholdsString, err := client.GetFile(ctx, "infra/config/generated/health-specs/health-specs.json")
	if err != nil {
		return nil, errors.Annotate(err, "Downloading thresholds").Err()
	}

	var thresholds Thresholds
	err = json.Unmarshal([]byte(thresholdsString), &thresholds)
	if err != nil {
		return nil, errors.Annotate(err, "Unmarshalling thresholds").Err()
	}

	return &thresholds, nil
}

func belowThresholds(row Row, thresholds BuilderThresholds) (bool, string) {
	const suffix = " for the last 7 days of builds. "
	explanation := ""
	healthy := true
	if row.BuildMinsP50 > thresholds.BuildTime.P50Mins {
		healthy = false
		explanation += "Builder was above the P50 build time threshold" + suffix
	}
	if row.BuildMinsP95 > thresholds.BuildTime.P95Mins {
		healthy = false
		explanation += "Builder was above the P95 build time threshold" + suffix
	}
	if row.FailRate > thresholds.FailRate.Average {
		healthy = false
		explanation += "Builder was above the average fail rate threshold" + suffix
	}
	if row.InfraFailRate > thresholds.InfraFailRate.Average {
		healthy = false
		explanation += "Builder was above the average infra fail rate threshold" + suffix
	}
	if row.PendingMinsP50 > thresholds.PendingTime.P50Mins {
		healthy = false
		explanation += "Builder was above the P50 pending time threshold" + suffix
	}
	if row.PendingMinsP95 > thresholds.PendingTime.P95Mins {
		healthy = false
		explanation += "Builder was above the P95 pending time threshold" + suffix
	}
	// TODO: add checks for Test Pending Time once the data is added to the DB query

	return healthy, explanation
}
