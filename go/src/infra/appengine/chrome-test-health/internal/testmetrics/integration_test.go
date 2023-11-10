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

	Convey("avg cores unfiltered", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table. Use a Sunday to make weekly tests easier
		testPartition := "2023-04-02"

		// Setup defaults for rdb data
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}
		rf.defaultFilename = "//dir/name/filename.go"
		// Make tests run all day for 7 days so all avg cores will be 1
		rf.defaultRuntime = (time.Hour * 24).Seconds()

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			rf.createResult().AddTime(time.Hour * 24 * 0),
			rf.createResult().AddTime(time.Hour * 24 * 1),
			rf.createResult().AddTime(time.Hour * 24 * 2),
			rf.createResult().AddTime(time.Hour * 24 * 3),
			rf.createResult().AddTime(time.Hour * 24 * 4),
			rf.createResult().AddTime(time.Hour * 24 * 5),
			rf.createResult().AddTime(time.Hour * 24 * 6),
			// Force the weekly cores to be 2
			rf.createResult().AddTime(time.Hour * 24 * 6).WithDuration(time.Hour * 24 * 7),
		}); err != nil {
			t.Fail()
		}

		// Start checking the fetches
		testResp, err := client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-04-08"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				PageSize: 10,
			},
		)
		So(err, ShouldBeNil)

		testSummary := getTestIdFromResponse(testResp, defaultTestId)
		metric, err := getMetric(testSummary.Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The test ran for 24 hours on the Sunday consuming 1 core the whole time
		So(metric, ShouldEqual, 1)
		metric, err = getMetric(testSummary.Metrics["2023-04-08"], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The test ran for 1 + 7 days on the Saturday over 2 results
		So(metric, ShouldEqual, 8)

		// Verify weekly
		testResp, err = client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_WEEK,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				PageSize: 10,
			},
		)
		So(err, ShouldBeNil)

		testSummary = getTestIdFromResponse(testResp, defaultTestId)
		metric, err = getMetric(testSummary.Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The total runtime should be 14 days which over 7 days is 2 cores
		So(metric, ShouldEqual, 2)

		dirResp, err := client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-04-08"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The test ran for 24 hours on the Sunday consuming 1 core the whole time
		So(metric, ShouldEqual, 1)
		metric, err = getMetric(dirResp.Nodes[0].Metrics["2023-04-08"], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The test ran for 1 + 7 days on the Saturday over 2 results
		So(metric, ShouldEqual, 8)

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_WEEK,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-04-08"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The total runtime should be 14 days which over 7 days is 2 cores
		So(metric, ShouldEqual, 2)
	})

	Convey("avg cores filtered", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table. Use a Sunday to make weekly tests easier
		testPartition := "2023-03-19"

		// Setup defaults for rdb data
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}
		rf.defaultFilename = "//dir/name/filename.go"
		// Make tests run all day for 7 days so all avg cores will be 1
		rf.defaultRuntime = (time.Hour * 24).Seconds()

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			rf.createResult().AddTime(time.Hour * 24 * 6),
			// Force the weekly cores to be 1 split when filter is "other_builder"
			rf.createResult().AddTime(time.Hour * 24 * 6).WithDuration(time.Hour * 24 * 7).WithBuilder("other_builder"),
		}); err != nil {
			t.Fail()
		}

		testResp, err := client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-03-25"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				PageSize: 10,
				Filter:   "other_builder",
			},
		)
		So(err, ShouldBeNil)

		testSummary := getTestIdFromResponse(testResp, defaultTestId)
		// other_builder variant did not run Sunday, we shouldn't get anything for this day
		So(testSummary.Metrics, ShouldNotContainKey, testPartition)
		metric, err := getMetric(testSummary.Metrics["2023-03-25"], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The other_builder test ran for 7 days on the Saturday over 1 day is 7 cores
		So(metric, ShouldEqual, 7)

		// Verify weekly
		testResp, err = client.FetchMetrics(ctx,
			&api.FetchTestMetricsRequest{
				Period:     api.Period_WEEK,
				Components: []string{"component"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				PageSize: 10,
				Filter:   "other_builder",
			},
		)
		So(err, ShouldBeNil)

		testSummary = getTestIdFromResponse(testResp, defaultTestId)
		metric, err = getMetric(testSummary.Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The total runtime for other_builder should be 7 days over 7 days which is 1 core
		So(metric, ShouldEqual, 1)

		dirResp, err := client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-03-25"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				ParentIds: []string{"/"},
				Filter:    "other_builder",
			},
		)
		So(err, ShouldBeNil)
		// other_builder variant did not run Sunday, we shouldn't get anything for this day
		So(dirResp.Nodes[0].Metrics, ShouldNotContainKey, testPartition)
		metric, err = getMetric(dirResp.Nodes[0].Metrics["2023-03-25"], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The other_builder test ran for 7 days on the Saturday over 1 day is 7 cores
		So(metric, ShouldEqual, 7)

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_WEEK,
				Components: []string{"component"},
				Dates:      []string{testPartition, "2023-03-25"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_CORES,
				},
				ParentIds: []string{"/"},
				Filter:    "other_builder",
			},
		)
		So(err, ShouldBeNil)
		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_CORES)
		So(err, ShouldBeNil)
		// The total runtime for other_builder should be 7 days over 7 days which is 1 core
		So(metric, ShouldEqual, 1)
	})

	Convey("file based component aggregations", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table. Use a Sunday to make weekly tests easier
		testPartition := "2023-03-12"

		// Setup defaults for rdb data
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}
		rf.defaultFilename = "//dir/name/filename.go"
		rf.defaultRuntime = (time.Hour * 24).Seconds()

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			// All tests exist in the same file but with different component/builder combinations
			rf.createResult().WithId("test1").WithComponent("component1").WithDuration(time.Minute * 1),
			rf.createResult().WithId("test2").WithComponent("component1").WithDuration(time.Minute * 5).WithBuilder("other_builder"),
			rf.createResult().WithId("test3").WithComponent("component2").WithDuration(time.Minute * 10).WithBuilder("other_builder"),
			// Add an entry that shouldn't affect the previous day
			rf.createResult().WithId("test1").WithComponent("component1").WithDuration(time.Hour).AddTime(time.Hour * 24),
			rf.createResult().WithId("test3").WithComponent("component2").WithDuration(time.Hour).AddTime(time.Hour * 24).WithBuilder("other_builder"),
		}); err != nil {
			t.Fail()
		}

		dirResp, err := client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)

		metric, err := getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// 3 tests in the same file on the first day make the avg runtime of
		// the file the sum of those 3 tests (1 + 5 + 10)
		So(metric, ShouldEqual, (time.Minute * 16).Seconds())

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{},
				// Add the next day to exercise the multi-day query
				Dates: []string{testPartition, "2023-03-13"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)
		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// The first day should not be affected by fetching the second
		So(metric, ShouldEqual, (time.Minute * 16).Seconds())
		metric, err = getMetric(dirResp.Nodes[0].Metrics["2023-03-13"], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		So(metric, ShouldEqual, (time.Hour * 2).Seconds())

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				ParentIds: []string{"/"},
				Filter:    "other_builder",
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// 2 tests belong to other_builder on the first day (5 + 10)
		So(metric, ShouldEqual, (time.Minute * 15).Seconds())

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{},
				// Add the next day to exercise the multi-day query
				Dates: []string{testPartition, "2023-03-13"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				ParentIds: []string{"/"},
				Filter:    "other_builder",
			},
		)
		So(err, ShouldBeNil)
		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// The first day should not be affected by fetching the second
		So(metric, ShouldEqual, (time.Minute * 15).Seconds())
		metric, err = getMetric(dirResp.Nodes[0].Metrics["2023-03-13"], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// Only 1 test on the 2nd day for other_builder
		So(metric, ShouldEqual, (time.Hour).Seconds())

		dirResp, err = client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"component1"},
				Dates:      []string{testPartition, "2023-03-13"},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				ParentIds: []string{"/"},
			},
		)
		So(err, ShouldBeNil)

		metric, err = getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// 2 tests have component1 on the first day make the avg runtime of
		// the file the sum of those 2 tests (1 + 5)
		So(metric, ShouldEqual, (time.Minute * 6).Seconds())
		metric, err = getMetric(dirResp.Nodes[0].Metrics["2023-03-13"], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// Only 1 test on the 2nd day for component1
		So(metric, ShouldEqual, (time.Hour).Seconds())
	})

	Convey("invalid file name summaries", t, func() {
		// Deleting rows with a streaming buffer doesn't work well, instead
		// partition the fake table. Use a Sunday to make weekly tests easier
		testPartition := "2023-02-26"

		// Setup defaults for rdb data
		rf.timePartition, err = civil.ParseDate(testPartition)
		if err != nil {
			t.Fail()
		}
		rf.defaultRuntime = (time.Hour * 24).Seconds()

		// Generate the rollups from fake rdb data.
		if err := createRollupFromResults(ctx, client, testProject, testDataset, fakeChromiumTryRdb, testPartition, []*fakeRdbResult{
			// All tests exist in the same file but with different component/builder combinations
			rf.createResult().WithId("test1").WithComponent("Unknown").WithDuration(time.Minute * 1).WithFilename("//dir/name/filename.go"),
			rf.createResult().WithId("test2").WithComponent("Unknown").WithDuration(time.Minute * 3).WithFilename("Unknown File"),
			rf.createResult().WithId("test3").WithComponent("Unknown").WithDuration(time.Minute * 7).WithFilename("Unknown File"),
		}); err != nil {
			t.Fail()
		}

		dirResp, err := client.FetchDirectoryMetrics(ctx,
			&api.FetchDirectoryMetricsRequest{
				Period:     api.Period_DAY,
				Components: []string{"Unknown"},
				Dates:      []string{testPartition},
				Metrics: []api.MetricType{
					api.MetricType_AVG_RUNTIME,
				},
				// "" Should return any file name not in the root "//"
				ParentIds: []string{""},
			},
		)
		So(err, ShouldBeNil)

		So(len(dirResp.Nodes), ShouldEqual, 1)
		metric, err := getMetric(dirResp.Nodes[0].Metrics[testPartition], api.MetricType_AVG_RUNTIME)
		So(err, ShouldBeNil)
		// Only test2 and test3 with "Unknown File" should appear
		So(metric, ShouldEqual, (time.Minute * 10).Seconds())
	})
}
