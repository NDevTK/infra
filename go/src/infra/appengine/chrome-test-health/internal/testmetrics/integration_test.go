// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build integration
// +build integration

package testmetrics

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

	"infra/appengine/chrome-test-health/api"

	. "github.com/smartystreets/goconvey/convey"
)

// Runs the integration tests returning an error if any fail to run or a
// check fails
func TestIntegrationTest(t *testing.T) {
	ctx := context.Background()

	// Setup the test environment
	bqClient, err := bigquery.NewClient(ctx, testProject)
	if err != nil {
		t.Fail()
	}

	client, err := setupClient(ctx, bqClient, testDataset, testProject)
	if err != nil {
		t.Fail()
	}

	if err := ensureTables(ctx, bqClient); err != nil {
		t.Fail()
	}

	rf := &resultFactory{}

	Convey("no duplicate rows are created", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table
		testPartition := "2023-07-02"
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}

		// Generate the fake rdb data.
		if err := createFakeRdb(ctx, bqClient, testProject, testDataset, fakeChromiumTryRdb, []*fakeRdbResult{
			rf.createResult(),
			rf.createResult(),
			rf.createResult().AddTime(time.Hour * 24),
			rf.createResult().AddTime(time.Hour * 24 * 2),
			rf.createResult().AddTime(time.Hour * 24 * 6),
		}); err != nil {
			t.Fail()
		}

		err = client.UpdateSummary(ctx, rf.timePartition, rf.timePartition.AddDays(6))
		So(err, ShouldBeNil)
		// Even if new data appears after the first roll up, that data needs to
		// be included in existing rows, not as new ones
		if err := createFakeRdb(ctx, bqClient, testProject, testDataset, fakeChromiumTryRdb, []*fakeRdbResult{
			rf.createResult(),
			rf.createResult().AddTime(time.Hour * 24),
			rf.createResult().AddTime(time.Hour * 24 * 2),
			rf.createResult().AddTime(time.Hour * 24 * 6),
		}); err != nil {
			t.Fail()
		}
		// Run the updates again to ensure nothing changes
		err = client.UpdateSummary(ctx, rf.timePartition, rf.timePartition.AddDays(6))
		So(err, ShouldBeNil)

		// Check the rolled up tables
		err := checkForDuplicateRows(ctx, bqClient)
		So(err, ShouldBeNil)
	})

	Convey("fetch", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table
		testPartition := "2023-06-01"
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			rf.createResult().AddTime(-time.Second),
			rf.createResult(),
			rf.createResult(),
			rf.createResult().Failed(),
			rf.createResult().WithBuilder("different_builder"),
			rf.createResult().AddTime(time.Hour * 24),
		}); err != nil {
			t.Fail()
		}

		// Start checking the fetches
		resp, err := client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics:    []api.MetricType{api.MetricType_NUM_RUNS},
				PageSize:   10,
			},
		)

		So(err, ShouldBeNil)

		// Check the test_id rollup is correct
		testSummary := getTestIdFromResponse(resp, defaultTestId)
		testMetricData := testSummary.Metrics[testPartition].Data[0]
		So(testMetricData.MetricType, ShouldEqual, api.MetricType_NUM_RUNS)
		So(testMetricData.MetricValue, ShouldEqual, 4)

		variant := getBuilderVariantFromTest(testSummary, "builder")
		variantMetricData := variant.Metrics[testPartition].Data[0]
		So(variantMetricData.MetricType, ShouldEqual, api.MetricType_NUM_RUNS)
		So(variantMetricData.MetricValue, ShouldEqual, 3)

		variant = getBuilderVariantFromTest(testSummary, "different_builder")
		variantMetricData = variant.Metrics[testPartition].Data[0]
		So(variantMetricData.MetricType, ShouldEqual, api.MetricType_NUM_RUNS)
		So(variantMetricData.MetricValue, ShouldEqual, 1)
	})

	Convey("total runtime", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table. Use a Sunday to make weekly tests easier
		testPartition := "2023-05-07"

		// Setup defaults for rdb data
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}
		rf.defaultFilename = "//dir/name/filename.go"
		rf.defaultRuntime = 1.0

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			rf.createResult(),
			rf.createResult(),
			rf.createResult().WithFilename("//dir/other/name/filename.go"),
			rf.createResult().WithBuilder("different_builder"),
			rf.createResult().WithBuilder("different_builder").WithFilename("//dir/other/name/filename.go"),
		}); err != nil {
			t.Fail()
		}

		// Start checking the fetches
		testResp, err := client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_TOTAL_RUNTIME,
				},
				PageSize: 10,
			},
		)
		So(err, ShouldBeNil)

		testSummary := getTestIdFromResponse(testResp, defaultTestId)
		metric, err := getMetric(testSummary.Metrics[testPartition], api.MetricType_TOTAL_RUNTIME)
		So(err, ShouldBeNil)
		// Each test is 1 second, 5 tests on this day should be 5 total runtime
		So(metric, ShouldEqual, 5)

		testResp, err = client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_TOTAL_RUNTIME,
				},
				Filter:   "different_builder",
				PageSize: 10,
			},
		)
		So(err, ShouldBeNil)

		testSummary = getTestIdFromResponse(testResp, defaultTestId)
		metric, err = getMetric(testSummary.Metrics[testPartition], api.MetricType_TOTAL_RUNTIME)
		So(err, ShouldBeNil)
		// Each test is 1 second, only 2 tests on this day use
		// "different_builder" should be 2 total runtime
		So(metric, ShouldEqual, 2)

		// Start checking the fetches
		dirResp, err := client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_TOTAL_RUNTIME,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_TOTAL_RUNTIME)
		So(err, ShouldBeNil)
		// Each test is 1 second, 5 tests on this day should be 5 total runtime
		So(metric, ShouldEqual, 5)

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_TOTAL_RUNTIME,
				},
				ParentIds: []string{"/"},
				Filter:    "different_builder",
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_TOTAL_RUNTIME)
		So(err, ShouldBeNil)
		// Each test is 1 second, only 2 tests on this day use
		// "different_builder" should be 2 total runtime
		So(metric, ShouldEqual, 2)
	})
}
