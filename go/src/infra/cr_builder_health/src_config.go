// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"infra/libs/git"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"
)

type SrcConfig struct {
	Default BuilderSpec           `json:"_default"`
	Specs   map[string]BucketSpec `json:"specs"`
}
type BucketSpec map[string]BuilderSpec
type BuilderSpec struct {
	Thresholds       Thresholds `json:"thresholds"`
	ContactTeamEmail string     `json:"contact_team_email"`
}
type Thresholds struct {
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

const explanationPrefix = "Builder was above the"
const explanationSuffix = "threshold for the last 7 days of builds."

func getSrcConfig(buildCtx context.Context) (*SrcConfig, error) {
	var err error
	step, ctx := build.StartStep(buildCtx, "Get Src Config")
	defer func() { step.End(err) }()

	step.SetSummaryMarkdown("Reading src config from https://chromium.googlesource.com/chromium/src/+/refs/heads/main/infra/config/generated/health-specs/health-specs.json")

	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, auth.Options{Scopes: []string{gitiles.OAuthScope}})
	httpClient, err := authenticator.Client()
	if err != nil {
		return nil, errors.Annotate(err, "Initializing Auth").Err()
	}

	client, err := git.NewClient(ctx, httpClient, "chromium-review.googlesource.com", "chromium.googlesource.com", "chromium/src", "main")
	if err != nil {
		return nil, errors.Annotate(err, "Initializing Gitiles client").Err()
	}

	srcConfigString, err := client.GetFile(ctx, "infra/config/generated/health-specs/health-specs.json")
	if err != nil {
		return nil, errors.Annotate(err, "Downloading src config").Err()
	}

	var srcConfig SrcConfig
	err = json.Unmarshal([]byte(srcConfigString), &srcConfig)
	if err != nil {
		return nil, errors.Annotate(err, "Unmarshalling src config").Err()
	}

	return &srcConfig, nil
}

func compareThresholds(ctx context.Context, row *Row, builderConfig *BuilderSpec) error {
	row.HealthScore = 10
	var stepErr error
	for _, metric := range row.Metrics {
		switch metric.Type {
		case "build_mins_p50":
			metric.Threshold = builderConfig.Thresholds.BuildTime.P50Mins
		case "build_mins_p95":
			metric.Threshold = builderConfig.Thresholds.BuildTime.P95Mins
		case "fail_rate":
			metric.Threshold = builderConfig.Thresholds.FailRate.Average
		case "infra_fail_rate":
			metric.Threshold = builderConfig.Thresholds.InfraFailRate.Average
		case "pending_mins_p50":
			metric.Threshold = builderConfig.Thresholds.PendingTime.P50Mins
		case "pending_mins_p95":
			metric.Threshold = builderConfig.Thresholds.PendingTime.P95Mins
		// TODO: add checks for Test Pending Time once the data is added to the DB query
		default:
			metric.HealthScore = 0
			err := fmt.Errorf("Found unknown metric type %s in BigQuery", metric.Type)

			// Log all, return just the last
			logging.Errorf(ctx, "%s", err)
			stepErr = err
			continue
		}
		compareThresholdsHelper(row, metric)
	}
	row.ScoreExplanation = strings.TrimRight(row.ScoreExplanation, " ")

	return stepErr
}

func compareThresholdsHelper(row *Row, metric *Metric) {
	if metric.Threshold == 0 {
		metric.HealthScore = 0 // redundant but wanted to be explicit
		return
	}

	metric.HealthScore = 10
	if metric.Value > metric.Threshold {
		metric.HealthScore = 1
		row.HealthScore = 1
		row.ScoreExplanation += fmt.Sprintf("%s %s %s ", explanationPrefix, metric.Type, explanationSuffix)
	}
}
