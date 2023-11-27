// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSrcConfig(t *testing.T) {
	t.Parallel()

	var testSrcConfig SrcConfig = SrcConfig{
		DefaultSpecs: map[string]ProblemSpec{
			"Unhealthy": {
				Thresholds: Thresholds{
					TestPendingTime: PercentileThresholds{P50Mins: 60, P95Mins: 120},
					PendingTime:     PercentileThresholds{P50Mins: 60, P95Mins: 120},
					BuildTime:       PercentileThresholds{P50Mins: 60, P95Mins: 120},
					FailRate:        AverageThresholds{Average: 0.2},
					InfraFailRate:   AverageThresholds{Average: 0.1},
				},
			},
			"Low Value": {
				Thresholds: Thresholds{
					FailRate:      AverageThresholds{Average: 0.99},
					InfraFailRate: AverageThresholds{Average: 0.99},
				},
			},
		},
		BucketSpecs: map[string]BuilderSpecs{
			"bucket": {
				"builder": BuilderSpec{
					ProblemSpecs: []ProblemSpec{
						{
							Name:  "Unhealthy",
							Score: UNHEALTHY_SCORE,
							Thresholds: Thresholds{
								Default: "Unhealthy",
							},
						},
						{
							Name:  "Low Value",
							Score: LOW_VALUE_SCORE,
							Thresholds: Thresholds{
								Default: "Low Value",
							},
						},
					},
				},
			},
			"slow-bucket": {
				"slow-builder": BuilderSpec{
					ProblemSpecs: []ProblemSpec{
						{
							Name:  "Low Value",
							Score: LOW_VALUE_SCORE,
							Thresholds: Thresholds{
								Default: "Low Value",
							},
						},
						{
							Name:  "Unhealthy",
							Score: UNHEALTHY_SCORE,
							Thresholds: Thresholds{
								TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								FailRate:        AverageThresholds{Average: 0.4},
								InfraFailRate:   AverageThresholds{Average: 0.3},
							},
						},
					},
				},
			},
			"custom-bucket": {
				"custom-builder": BuilderSpec{
					ProblemSpecs: []ProblemSpec{
						{
							Name:  "Unhealthy",
							Score: UNHEALTHY_SCORE,
							Thresholds: Thresholds{
								TestPendingTime: PercentileThresholds{P50Mins: 60, P95Mins: 120},
								PendingTime:     PercentileThresholds{P50Mins: 60, P95Mins: 120},
								BuildTime:       PercentileThresholds{P50Mins: 60, P95Mins: 120},
								FailRate:        AverageThresholds{Average: 0.4},
								InfraFailRate:   AverageThresholds{Average: 0.3},
							},
						},
						{
							Name:  "Low Value",
							Score: LOW_VALUE_SCORE,
							Thresholds: Thresholds{
								TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								FailRate:        AverageThresholds{Average: 0.99},
								InfraFailRate:   AverageThresholds{Average: 0.99},
							},
						},
					},
				},
			},
			"improper-bucket": {
				"improper-builder": BuilderSpec{
					ProblemSpecs: []ProblemSpec{
						{
							Name:  "Unhealthy",
							Score: UNHEALTHY_SCORE,
							Thresholds: Thresholds{
								Default:         "Unhealthy",
								TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								FailRate:        AverageThresholds{Average: 0.4},
								InfraFailRate:   AverageThresholds{Average: 0.3},
							},
						},
					},
				},
				"improper-builder2": BuilderSpec{
					ProblemSpecs: []ProblemSpec{
						{
							Name:  "Unhealthy",
							Score: UNHEALTHY_SCORE,
							Thresholds: Thresholds{
								Default:         "not_default",
								TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
								FailRate:        AverageThresholds{Average: 0.4},
								InfraFailRate:   AverageThresholds{Average: 0.3},
							},
						},
					},
				},
				"improper-builder3": BuilderSpec{
					ProblemSpecs: []ProblemSpec{},
				},
			},
		},
	}

	Convey("Healthy builder is healthy, default thresholds", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "bucket",
			Builder: "builder",
			Metrics: []*Metric{
				{Type: "build_mins_p50", Value: 59},
				{Type: "build_mins_p95", Value: 119},
				{Type: "pending_mins_p50", Value: 59},
				{Type: "pending_mins_p95", Value: 119},
				{Type: "fail_rate", Value: 0.05},
				{Type: "infra_fail_rate", Value: 0},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(savedThresholds, ShouldResemble, testSrcConfig)
		So(outputRows[0].HealthScore, ShouldEqual, HEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldBeEmpty)
	})
	Convey("P50 percentile above threshold, default thresholds", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "bucket",
			Builder: "builder",
			Metrics: []*Metric{
				{Type: "build_mins_p50", Value: 61},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "build_mins_p50")
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("P95 percentile above thresholds, default thresholds", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "bucket",
			Builder: "builder",
			Metrics: []*Metric{
				{Type: "pending_mins_p95", Value: 121},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "pending_mins_p95")
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Fail rate above thresholds, default thresholds", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "bucket",
			Builder: "builder",
			Metrics: []*Metric{
				{Type: "fail_rate", Value: 0.3},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "fail_rate")
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("P50 build time below thresholds, slow builder", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "slow-bucket",
			Builder: "slow-builder",
			Metrics: []*Metric{
				{Type: "build_mins_p50", Value: 200},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, HEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldBeEmpty)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Infra fail rate above thresholds, slow builder", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "slow-bucket",
			Builder: "slow-builder",
			Metrics: []*Metric{
				{Type: "infra_fail_rate", Value: 0.5},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "infra_fail_rate")
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Default thresholds with custom thresholds error", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "slow-bucket",
			Builder: "slow-builder",
			Metrics: []*Metric{
				{Type: "infra_fail_rate", Value: 0.5},
			},
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "infra_fail_rate")
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Multiple healthy builders", t, func() {
		ctx := context.Background()
		rows := []Row{
			{
				Bucket:  "bucket",
				Builder: "builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 59},
					{Type: "build_mins_p95", Value: 119},
					{Type: "pending_mins_p50", Value: 59},
					{Type: "pending_mins_p95", Value: 119},
					{Type: "fail_rate", Value: 0.05},
					{Type: "infra_fail_rate", Value: 0},
				},
			},
			{
				Bucket:  "slow-bucket",
				Builder: "slow-builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 59},
					{Type: "build_mins_p95", Value: 119},
					{Type: "pending_mins_p50", Value: 59},
					{Type: "pending_mins_p95", Value: 119},
					{Type: "fail_rate", Value: 0.05},
					{Type: "infra_fail_rate", Value: 0},
				},
			},
		}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 2)
		So(savedThresholds, ShouldResemble, testSrcConfig)
		So(outputRows[0].HealthScore, ShouldEqual, HEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldBeEmpty)
		So(outputRows[1].HealthScore, ShouldEqual, HEALTHY_SCORE)
		So(outputRows[1].ScoreExplanation, ShouldBeEmpty)
	})
	Convey("One healthy, one unhealthy builder", t, func() {
		ctx := context.Background()
		rows := []Row{
			{
				Bucket:  "bucket",
				Builder: "builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 61},
					{Type: "build_mins_p95", Value: 121},
					{Type: "pending_mins_p50", Value: 61},
					{Type: "pending_mins_p95", Value: 121},
					{Type: "fail_rate", Value: 0.3},
					{Type: "infra_fail_rate", Value: 0.2},
				},
			},
			{
				Bucket:  "slow-bucket",
				Builder: "slow-builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 59},
					{Type: "build_mins_p95", Value: 119},
					{Type: "pending_mins_p50", Value: 59},
					{Type: "pending_mins_p95", Value: 119},
					{Type: "fail_rate", Value: 0.05},
					{Type: "infra_fail_rate", Value: 0},
				},
			},
		}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 2)
		So(savedThresholds, ShouldResemble, testSrcConfig)
		So(outputRows[0].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "build_mins")
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "infra_fail_rate")
		So(outputRows[1].HealthScore, ShouldEqual, HEALTHY_SCORE)
		So(outputRows[1].ScoreExplanation, ShouldBeEmpty)
	})
	Convey("One low value, one unhealthy", t, func() {
		ctx := context.Background()
		rows := []Row{
			{
				Bucket:  "bucket",
				Builder: "builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 61},
					{Type: "build_mins_p95", Value: 121},
					{Type: "pending_mins_p50", Value: 61},
					{Type: "pending_mins_p95", Value: 121},
					{Type: "fail_rate", Value: 1.0},
					{Type: "infra_fail_rate", Value: 1.0},
				},
			},
			{
				Bucket:  "slow-bucket",
				Builder: "slow-builder",
				Metrics: []*Metric{
					{Type: "build_mins_p50", Value: 61},
					{Type: "build_mins_p95", Value: 121},
					{Type: "pending_mins_p50", Value: 61},
					{Type: "pending_mins_p95", Value: 121},
					{Type: "fail_rate", Value: 0.90},
					{Type: "infra_fail_rate", Value: 0.90},
				},
			},
		}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 2)
		So(savedThresholds, ShouldResemble, testSrcConfig)
		So(outputRows[0].HealthScore, ShouldEqual, LOW_VALUE_SCORE)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "fail_rate")
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "Low Value")
		So(outputRows[1].HealthScore, ShouldEqual, UNHEALTHY_SCORE)
		So(outputRows[1].ScoreExplanation, ShouldContainSubstring, "fail_rate")
		So(outputRows[1].ScoreExplanation, ShouldContainSubstring, "Unhealthy")
	})
	Convey("Improper threshold config, both default and custom thresholds", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "improper-bucket",
			Builder: "improper-builder",
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldNotBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "default")
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "custom")
		So(outputRows[0].HealthScore, ShouldEqual, UNSET_SCORE)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Improper threshold config, Default set to unknown sentinel value", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "improper-bucket",
			Builder: "improper-builder2",
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldNotBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "unknown sentinel")
		So(outputRows[0].HealthScore, ShouldEqual, UNSET_SCORE)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Improper ProblemSpecs, no ProblemSpecs", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "improper-bucket",
			Builder: "improper-builder3",
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "no ProblemSpecs")
		So(outputRows[0].HealthScore, ShouldEqual, UNSET_SCORE)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Unconfigured builder", t, func() {
		ctx := context.Background()
		rows := []Row{{
			Bucket:  "unconfigured-bucket",
			Builder: "unconfigured-builder",
		}}
		savedThresholds := testSrcConfig
		outputRows, err := calculateIndicators(ctx, rows, testSrcConfig)
		So(err, ShouldBeNil)
		So(len(outputRows), ShouldEqual, 1)
		So(outputRows[0].ScoreExplanation, ShouldBeBlank)
		So(outputRows[0].HealthScore, ShouldEqual, UNSET_SCORE)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
	Convey("Sort ProblemSpecs", t, func() {
		ps := []ProblemSpec{
			{
				Name:  "Low Value",
				Score: 1,
				Thresholds: Thresholds{
					Default: "Low Value",
				},
			},
			{
				Name:  "Unhealthy",
				Score: UNHEALTHY_SCORE,
				Thresholds: Thresholds{
					Default: "Unhealthy",
				},
			},
		}
		sortProblemSpecs(ps)
		So(ps[0].Score, ShouldEqual, UNHEALTHY_SCORE)
	})
	Convey("Compare Thresholds Helper", t, func() {
		const unhealthyIndex = 0
		const lowValueIndex = 1
		row := Row{
			Bucket:  "custom-bucket",
			Builder: "custom-builder",
			Metrics: []*Metric{
				{Type: "fail_rate", Value: 1.0},
			},
		}
		ps := testSrcConfig.BucketSpecs[row.Bucket][row.Builder].ProblemSpecs

		row.Metrics[0].Threshold = ps[unhealthyIndex].Thresholds.FailRate.Average
		compareThresholdsHelper(&row, &ps[unhealthyIndex], row.Metrics[0])
		So(row.HealthScore, ShouldEqual, UNHEALTHY_SCORE)

		row.Metrics[0].Threshold = ps[lowValueIndex].Thresholds.FailRate.Average
		compareThresholdsHelper(&row, &ps[lowValueIndex], row.Metrics[0])
		So(row.HealthScore, ShouldEqual, LOW_VALUE_SCORE)
	})
}
