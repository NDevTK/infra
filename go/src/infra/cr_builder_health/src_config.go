// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/api/gitiles"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/luciexe/build"

	"infra/libs/git"
)

type SrcConfig struct {
	DefaultSpecs []ProblemSpec           `json:"_default_specs"`
	BucketSpecs  map[string]BuilderSpecs `json:"specs"` // e.g. ci -> {}
}
type BuilderSpecs map[string]BuilderSpec // e.g. linux-rel -> {}
type BuilderSpec struct {
	ContactTeamEmail string        `json:"contact_team_email"`
	ProblemSpecs     []ProblemSpec `json:"problem_specs"` // e.g. UNHEALTHY -> {}
}
type ProblemSpec struct {
	Name       string     `json:"name"` // This name will be shown in Milo when a builder is affected by this Problem.
	Score      int        `json:"score"`
	Thresholds Thresholds `json:"thresholds"`
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

const HEALTHY_SCORE = 10
const UNHEALTHY_SCORE = 5
const LOW_VALUE_SCORE = 1
const UNSET_SCORE = 0

const UNSET_THRESHOLD = 0

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

func compareThresholds(ctx context.Context, row *Row, problemSpec *ProblemSpec) error {
	if row.HealthScore == UNSET_SCORE {
		row.HealthScore = HEALTHY_SCORE
	}
	// TODO: make metric.Threshold a list, right now it just takes the lowest problem spec score threshold
	var stepErr error
	for _, metric := range row.Metrics {
		switch metric.Type {
		case "build_mins_p50":
			metric.Threshold = problemSpec.Thresholds.BuildTime.P50Mins
		case "build_mins_p95":
			metric.Threshold = problemSpec.Thresholds.BuildTime.P95Mins
		case "fail_rate":
			metric.Threshold = problemSpec.Thresholds.FailRate.Average
		case "infra_fail_rate":
			metric.Threshold = problemSpec.Thresholds.InfraFailRate.Average
		case "pending_mins_p50":
			metric.Threshold = problemSpec.Thresholds.PendingTime.P50Mins
		case "pending_mins_p95":
			metric.Threshold = problemSpec.Thresholds.PendingTime.P95Mins
		// TODO: add checks for Test Pending Time once the data is added to the DB query
		default:
			metric.HealthScore = UNSET_SCORE
			err := fmt.Errorf("Found unknown metric type %s in BigQuery", metric.Type)

			// Log all, return just the last
			logging.Errorf(ctx, "%s", err)
			stepErr = err
			continue
		}
		compareThresholdsHelper(row, problemSpec, metric)
	}
	row.ScoreExplanation = strings.TrimRight(row.ScoreExplanation, " ")

	return stepErr
}

func compareThresholdsHelper(row *Row, problemSpec *ProblemSpec, metric *Metric) {
	if metric.Threshold == UNSET_THRESHOLD {
		return
	}

	metric.HealthScore = int(math.Min(HEALTHY_SCORE, float64(metric.HealthScore)))
	if metric.Value > metric.Threshold {
		metric.HealthScore = problemSpec.Score
		row.HealthScore = problemSpec.Score
		row.ScoreExplanation += fmt.Sprintf("%s %s threshold in the %s ProblemSpec", explanationPrefix, metric.Type, problemSpec.Name)
	}
}

// Used for problem precedence
func sortProblemSpecs(problemSpecs []ProblemSpec) {
	sort.Slice(problemSpecs, func(i, j int) bool {
		return problemSpecs[i].Score > problemSpecs[j].Score
	})
}
