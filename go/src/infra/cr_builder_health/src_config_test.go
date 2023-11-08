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
		Default: BuilderSpec{
			Thresholds: Thresholds{
				TestPendingTime: PercentileThresholds{P50Mins: 60, P95Mins: 120},
				PendingTime:     PercentileThresholds{P50Mins: 60, P95Mins: 120},
				BuildTime:       PercentileThresholds{P50Mins: 60, P95Mins: 120},
				FailRate:        AverageThresholds{Average: 0.2},
				InfraFailRate:   AverageThresholds{Average: 0.1},
			},
		},
		Specs: map[string]BucketSpec{
			"bucket": {
				"builder": BuilderSpec{
					Thresholds: Thresholds{
						Default: "_default",
					},
				},
			},
			"slow-bucket": {
				"slow-builder": BuilderSpec{
					Thresholds: Thresholds{
						TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						FailRate:        AverageThresholds{Average: 0.4},
						InfraFailRate:   AverageThresholds{Average: 0.3},
					},
				},
			},
			"improper-bucket": {
				"improper-builder": BuilderSpec{
					Thresholds: Thresholds{
						Default:         "_default",
						TestPendingTime: PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						PendingTime:     PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						BuildTime:       PercentileThresholds{P50Mins: 600, P95Mins: 1200},
						FailRate:        AverageThresholds{Average: 0.4},
						InfraFailRate:   AverageThresholds{Average: 0.3},
					},
				},
				"improper-builder2": BuilderSpec{
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
		So(outputRows[0].HealthScore, ShouldEqual, 10)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
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
		So(outputRows[0].HealthScore, ShouldEqual, 10)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
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
		So(outputRows[0].HealthScore, ShouldEqual, 10)
		So(outputRows[0].ScoreExplanation, ShouldBeEmpty)
		So(outputRows[1].HealthScore, ShouldEqual, 10)
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
		So(outputRows[0].HealthScore, ShouldEqual, 1)
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "build_mins")
		So(outputRows[0].ScoreExplanation, ShouldContainSubstring, "infra_fail_rate")
		So(outputRows[1].HealthScore, ShouldEqual, 10)
		So(outputRows[1].ScoreExplanation, ShouldBeEmpty)
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
		So(outputRows[0].HealthScore, ShouldEqual, 0)
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
		So(outputRows[0].HealthScore, ShouldEqual, 0)
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
		So(outputRows[0].HealthScore, ShouldEqual, 0)
		So(savedThresholds, ShouldResemble, testSrcConfig)
	})
}
