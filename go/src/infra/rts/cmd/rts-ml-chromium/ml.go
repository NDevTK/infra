// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/iterator"
)

// A single test's stability for a day (Timestamp) including its identify
// information
type bqStabilityRow struct {
	TestId      string    `bigquery:"test_id"`
	VariantHash string    `bigquery:"variant_hash"`
	Stability   stability `bigquery:"stability"`
	TestSuite   string    `bigquery:"test_suite"`
	Timestamp   time.Time `bigquery:"day"`
}

// The stability information for a single test on a single day
type stability struct {
	SixMonthFailCount int64 `bigquery:"six_month_fail_count"`
	SixMonthRunCount  int64 `bigquery:"six_month_run_count"`
	OneMonthFailCount int64 `bigquery:"one_month_fail_count"`
	OneMonthRunCount  int64 `bigquery:"one_month_run_count"`
	OneWeekFailCount  int64 `bigquery:"one_week_fail_count"`
	OneWeekRunCount   int64 `bigquery:"one_week_run_count"`
}

// A single example that the model will make a prediction against
type mlExample struct {
	TestId            string
	TestSuite         string
	SixMonthFailCount int64
	SixMonthRunCount  int64
	OneMonthFailCount int64
	OneMonthRunCount  int64
	OneWeekFailCount  int64
	OneWeekRunCount   int64
	GitDistance       float64
	UseGitDistance    bool
	FileDistance      float64
	UseFileDistance   bool
}

// Convert the stability information into an mlExample for the ML model
func (r bqStabilityRow) mlExample() mlExample {
	return mlExample{
		TestId:            r.TestId,
		TestSuite:         r.TestSuite,
		SixMonthFailCount: r.Stability.SixMonthFailCount,
		SixMonthRunCount:  r.Stability.SixMonthRunCount,
		OneMonthFailCount: r.Stability.OneMonthFailCount,
		OneMonthRunCount:  r.Stability.OneMonthRunCount,
		OneWeekFailCount:  r.Stability.OneWeekFailCount,
		OneWeekRunCount:   r.Stability.OneWeekRunCount,
		GitDistance:       0,
		UseGitDistance:    false,
		FileDistance:      0,
		UseFileDistance:   false,
	}
}

// Uses the ml cli to make predictions. Passes the dataframes to the cli through
// a file to avoid command line argument limits
func fileInferMlModel(rows []mlExample, savedModelDir string, cli string) ([]float64, error) {
	featuresFile, err := ioutil.TempFile("predictions", "PsFeatures_thread*.csv")
	if err != nil {
		return nil, err
	}
	featureFileName := featuresFile.Name()
	defer os.Remove(featureFileName)

	predictionsFileName := strings.Replace(featureFileName, ".csv", "_predict.csv", -1)
	err = os.MkdirAll("predictions", os.ModePerm)
	if err != nil {
		return nil, err
	}

	csvWriter := csv.NewWriter(featuresFile)

	csvData := [][]string{{
		"GitDistance",
		"FileDistance",
		"OneWeekFailCount",
		"OneWeekRunCount",
		"OneMonthFailCount",
		"OneMonthRunCount",
		"SixMonthFailCount",
		"SixMonthRunCount"},
	}

	for _, row := range rows {
		gitDistance := ""
		if row.UseGitDistance {
			gitDistance = fmt.Sprintf("%v", row.GitDistance)
		}
		fileDistance := ""
		if row.UseFileDistance {
			fileDistance = fmt.Sprintf("%v", row.FileDistance)
		}
		csvData = append(csvData,
			[]string{
				gitDistance,
				fileDistance,
				fmt.Sprintf("%v", row.OneWeekFailCount),
				fmt.Sprintf("%v", row.OneWeekRunCount),
				fmt.Sprintf("%v", row.OneMonthFailCount),
				fmt.Sprintf("%v", row.OneMonthRunCount),
				fmt.Sprintf("%v", row.SixMonthFailCount),
				fmt.Sprintf("%v", row.SixMonthRunCount),
			})
	}
	csvWriter.WriteAll(csvData)

	cmd := exec.Command("python3",
		cli,
		"predict",
		"--file",
		featureFileName,
		"--output",
		predictionsFileName,
		"--model",
		savedModelDir,
	)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	err = cmd.Run()

	fmt.Print("stdout from cli:\n")
	fmt.Print(string(outBuf.String()))

	if err != nil {
		fmt.Print("stderr from cli:\n")
		fmt.Print(string(errBuf.String()))
		return nil, err
	}

	fileText, err := ioutil.ReadFile(predictionsFileName)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(fileText)), "\n")

	if len(rows) != len(lines) {
		fmt.Print("ML inference returned too many predictions\n")
		return nil, errors.New("ML inference returned too many predictions")
	}

	outDistances := make([]float64, len(rows))
	for i, line := range lines {

		distance, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return nil, errors.Annotate(err, "failed to parse a prediction from %q", line).Err()
		}

		// "Distance" and "Failure" have reverse meaning on whether to run
		outDistances[i] = 1. - distance
	}
	return outDistances, nil
}

type stabilityMapKey struct {
	date   time.Time
	testID string
}

// Returns the test fail rates as they were between the provided time period
// as a map of the day
func getTestIdToStabilityRowMap(ctx context.Context, bqClient *bigquery.Client, builder string, testSuite string, start time.Time, stop time.Time) (map[stabilityMapKey]mlExample, error) {
	testStability, err := queryStability(ctx, bqClient, builder, testSuite, start, stop)

	if err != nil {
		return nil, err
	}

	testStabilityMap := make(map[stabilityMapKey]mlExample)
	for _, row := range testStability {
		key := stabilityMapKey{date: row.Timestamp, testID: row.TestId}
		testStabilityMap[key] = row.mlExample()
	}
	return testStabilityMap, err
}

// Get the historic stability for the given time range
// TODO(sshrimp): Use a callback to avoid duplicate memory allocationg
func queryStability(ctx context.Context, bqClient *bigquery.Client, builder string, testSuite string, startTime time.Time, endTime time.Time) ([]bqStabilityRow, error) {
	q := bqClient.Query(evalStabilityQuery)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start_day", Value: startTime},
		{Name: "end_day", Value: endTime},
		{Name: "query_builder", Value: builder},
		{Name: "query_test_suite", Value: testSuite},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	rows := []bqStabilityRow{}
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Read the next row.
		row := &bqStabilityRow{}
		switch err := it.Next(row); {
		case err == iterator.Done:
			return rows, ctx.Err()
		case err != nil:
			return nil, err
		}

		rows = append(rows, *row)
	}
}

const evalStabilityQuery = `
-- Unique test ids that were run during the time period of interest, ignoring
-- the day on which it happened. We can't just look at the tests in the summary
-- on that day since if the test didn't run on a day, it wouldn't appear in the
-- daily summary and we would get gaps
WITH test_ids as (
	SELECT
		ds.test_id test_id,
		ds.variant_hash variant_hash,
		ANY_VALUE(SUBSTR((SELECT v FROM ds.testVariant.variant v WHERE v LIKE 'builder:%'), 9)) as builder,
		ANY_VALUE(SUBSTR((SELECT v FROM ds.testVariant.variant v WHERE v LIKE 'test_suite:%'), 12)) as test_suite,
	FROM chrome-trooper-analytics.test_results.daily_summary ds
	GROUP BY ds.test_id, variant_hash
	HAVING
		("" = @query_test_suite OR test_suite = @query_test_suite)
		AND ("" = @query_builder OR builder = @query_builder)
),

-- The unique days for the period of interest (from 6 months before the earliest
-- day to the day before the latest day)
day_range as (
	SELECT *
	FROM
		UNNEST(GENERATE_TIMESTAMP_ARRAY(TIMESTAMP_SUB(TIMESTAMP_TRUNC(@start_day, DAY), INTERVAL 181 DAY), @end_day, INTERVAL 1 DAY)) AS day
)

-- Collect unique test variants and the stability info for that test over the
-- requested days
SELECT
	dt.day day,
	tid.test_id test_id,
	tid.variant_hash variant_hash,
	tid.builder as builder,
	tid.test_suite as test_suite,
	(
		# Aggregate the daily summary leading up to the day in the outer query (dt.day)
		SELECT AS STRUCT
			IFNULL(SUM(ARRAY_LENGTH(dds.patchsets_with_failures)), 0) AS six_month_fail_count,
			IFNULL(SUM(dds.run_count), 0) AS six_month_run_count,
			# Note the extra day is to account for lag between collecting the data and a new model being generated
			IFNULL(SUM(IF(dds.day >= TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 31 DAY), ARRAY_LENGTH(dds.patchsets_with_failures), 0)), 0) AS one_month_fail_count,
			IFNULL(SUM(IF(dds.day >= TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 31 DAY), dds.run_count, 0)), 0) AS one_month_run_count,
			IFNULL(SUM(IF(dds.day >= TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 8 DAY), ARRAY_LENGTH(dds.patchsets_with_failures), 0)), 0) AS one_week_fail_count,
			IFNULL(SUM(IF(dds.day >= TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 8 DAY), dds.run_count, 0)), 0) AS one_week_run_count,
		FROM chrome-trooper-analytics.test_results.daily_summary dds
		WHERE
			day BETWEEN
				TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 181 DAY) AND
				TIMESTAMP_SUB(TIMESTAMP_TRUNC(dt.day, DAY), INTERVAL 1 DAY)
			AND dds.test_id = tid.test_id
			AND dds.variant_hash = tid.variant_hash
	) as stability,
FROM test_ids tid, day_range dt
WHERE
	dt.day BETWEEN @start_day AND @end_day
`
