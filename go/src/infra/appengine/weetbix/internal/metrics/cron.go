// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package metrics

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/tsmon"
	"go.chromium.org/luci/common/tsmon/distribution"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	"go.chromium.org/luci/common/tsmon/types"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/clustering/rules"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/ingestion/control"
)

var (
	activeRulesGauge = metric.NewInt(
		"weetbix/clustering/active_rules",
		"The total number of active rules, by LUCI project.",
		&types.MetricMetadata{Units: "rules"},
		// The LUCI Project.
		field.String("project"))

	joinToBuildGauge = metric.NewNonCumulativeDistribution(
		"weetbix/ingestion/join/to_build_result_by_hour",
		fmt.Sprintf(
			"The age distribution of presubmit builds with a presubmit"+
				" result recorded, broken down by project of the presubmit "+
				" run and whether the builds are joined to a buildbucket "+
				" build result."+
				" Age is measured as hours since the presubmit run result was"+
				" recorded. Only recent data (age < %v hours) is included."+
				" Used to measure Weetbix's performance joining to"+
				" buildbucket builds.", control.JoinStatsHours),
		&types.MetricMetadata{Units: "hours ago"},
		distribution.FixedWidthBucketer(1, control.JoinStatsHours),
		// The LUCI Project.
		field.String("project"),
		field.Bool("joined"))

	joinToPresubmitGauge = metric.NewNonCumulativeDistribution(
		"weetbix/ingestion/join/to_presubmit_result_by_hour",
		fmt.Sprintf(
			"The age distribution of presubmit builds with a buildbucket"+
				" build result recorded, broken down by project of the"+
				" buildbucket build and whether the builds are joined to"+
				" a presubmit run result."+
				" Age is measured as hours since the buildbucket build"+
				" result was recorded. Only recent data (age < %v hours)"+
				" is included."+
				" Used to measure Weetbix's performance joining to presubmit"+
				" runs.", control.JoinStatsHours),
		&types.MetricMetadata{Units: "hours ago"},
		distribution.FixedWidthBucketer(1, control.JoinStatsHours),
		// The LUCI Project.
		field.String("project"),
		field.Bool("joined"))
)

func init() {
	// Register metrics as global metrics, which has the effort of
	// resetting them after every flush.
	tsmon.RegisterGlobalCallback(func(ctx context.Context) {
		// Do nothing -- the metrics will be populated by the cron
		// job itself and does not need to be triggered externally.
	}, activeRulesGauge, joinToBuildGauge, joinToPresubmitGauge)
}

// GlobalMetrics handles the "global-metrics" cron job. It reports
// metrics related to overall system state (that are not logically
// reported as part of individual task or cron job executions).
func GlobalMetrics(ctx context.Context) error {
	projectConfigs, err := config.Projects(ctx)
	if err != nil {
		return errors.Annotate(err, "obtain project configs").Err()
	}

	// Total number of active rules, broken down by project.
	activeRules, err := rules.ReadTotalActiveRules(span.Single(ctx))
	if err != nil {
		return errors.Annotate(err, "collect total active rules").Err()
	}
	for project := range projectConfigs {
		// If there is no entry in activeRules for this project
		// (e.g. because there are no rules in that project),
		// the read count defaults to zero, which is the correct
		// behaviour.
		count := activeRules[project]
		activeRulesGauge.Set(ctx, count, project)
	}

	// Performance joining to buildbucket builds in ingestion.
	buildJoinStats, err := control.ReadBuildJoinStatistics(span.Single(ctx))
	if err != nil {
		return errors.Annotate(err, "collect buildbucket build join statistics").Err()
	}
	reportJoinStats(ctx, joinToBuildGauge, buildJoinStats)

	// Performance joining to presubmit runs in ingestion.
	psRunJoinStats, err := control.ReadPresubmitRunJoinStatistics(span.Single(ctx))
	if err != nil {
		return errors.Annotate(err, "collect presubmit run join statistics").Err()
	}
	reportJoinStats(ctx, joinToPresubmitGauge, psRunJoinStats)

	return nil
}

func reportJoinStats(ctx context.Context, metric metric.NonCumulativeDistribution, resultsByProject map[string]control.JoinStatistics) {
	for project, stats := range resultsByProject {
		joinedDist := distribution.New(metric.Bucketer())
		unjoinedDist := distribution.New(metric.Bucketer())

		for hoursAgo := 0; hoursAgo < control.JoinStatsHours; hoursAgo++ {
			joinedBuilds := stats.JoinedByHour[hoursAgo]
			unjoinedBuilds := stats.TotalByHour[hoursAgo] - joinedBuilds
			for i := int64(0); i < joinedBuilds; i++ {
				joinedDist.Add(float64(hoursAgo))
			}
			for i := int64(0); i < unjoinedBuilds; i++ {
				unjoinedDist.Add(float64(hoursAgo))
			}
		}

		joined := true
		metric.Set(ctx, joinedDist, project, joined)
		joined = false
		metric.Set(ctx, unjoinedDist, project, joined)
	}
}
