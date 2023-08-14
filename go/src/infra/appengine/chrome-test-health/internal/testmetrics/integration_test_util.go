// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build integration
// +build integration

package testmetrics

import (
	"context"
	"infra/appengine/chrome-test-health/api"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"go.chromium.org/luci/common/errors"
)

var (
	testProject     = "chrome-test-health-staging"
	fakeChromiumRdb = "fake_chromium_rdb"
	fakeChromeRdb   = "fake_chrome_rdb"
	fakeAttempts    = "fake_attempts"
	testDataset     = "test"

	sqlDir = "../../"

	createQueries = []string{
		"sql/create_daily_file_summary_table.sql",
		"sql/create_daily_summary_table.sql",
		"sql/create_raw_table.sql",
		"sql/create_weekly_file_summary_table.sql",
		"sql/create_weekly_summary_table.sql",
	}

	defaultTestId = "ninja://test_id"
)

type fakeRdbResult struct {
	testId        string
	buildId       string
	variantHash   string
	expected      bool
	exonerated    bool
	partitionTime time.Time

	duration  float64
	filename  string
	repo      string
	name      string
	builder   string
	testSuite string
	platform  string
	component string

	offsetTime time.Duration
}

type resultFactory struct {
	// The day to create the results in
	timePartition civil.Date
}

func (f *resultFactory) createResult() *fakeRdbResult {
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return nil
	}

	partitionTime := f.timePartition.In(tz)

	return &fakeRdbResult{
		testId:        defaultTestId,
		buildId:       "build-id",
		variantHash:   "variant_hash",
		expected:      true,
		exonerated:    false,
		partitionTime: partitionTime,
		component:     "component",
		duration:      1.0,
		builder:       "builder",
		testSuite:     "test_suite",
	}
}

func (f *fakeRdbResult) ResultTime() time.Time {
	return f.partitionTime.Add(f.offsetTime)
}

func (f *fakeRdbResult) WithTestSuite(testSuite string) *fakeRdbResult {
	f.testSuite = testSuite
	return f
}

func (f *fakeRdbResult) WithBuilder(builder string) *fakeRdbResult {
	f.builder = builder
	return f
}

func (f *fakeRdbResult) AddTime(time time.Duration) *fakeRdbResult {
	f.offsetTime = time
	return f
}

func (f *fakeRdbResult) Failed() *fakeRdbResult {
	f.expected = false
	return f
}

func (f *fakeRdbResult) Duration(duration float64) *fakeRdbResult {
	f.duration = duration
	return f
}

func (f *fakeRdbResult) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return map[string]bigquery.Value{
		// Required by the schema
		"exported": map[string]bigquery.Value{
			"id":    "123",
			"realm": "project:bucket",
		},
		"parent": map[string]bigquery.Value{
			"id": "123",
		},
		"test_id":        f.testId,
		"result_id":      "fake_result_id",
		"variant_hash":   f.variantHash,
		"expected":       f.expected,
		"status":         "status",
		"summary_html":   "<summary_html>",
		"exonerated":     f.exonerated,
		"partition_time": f.ResultTime(),
		// Used from the schema
		"duration": f.duration,
		"test_metadata": map[string]bigquery.Value{
			"location": map[string]bigquery.Value{
				"file_name": f.filename,
				"repo":      f.repo,
			},
		},
		"name": f.name,
		"variant": []map[string]bigquery.Value{
			{
				"key":   "builder",
				"value": f.builder,
			},
			{
				"key":   "test_suite",
				"value": f.testSuite,
			},
		},
		"tags": []map[string]bigquery.Value{
			{
				"key":   "target_platform",
				"value": f.platform,
			},
			{
				"key":   "monorail_component",
				"value": f.component,
			},
		},
	}, "", nil
}

func setupClient(ctx context.Context, bqClient *bigquery.Client, dataSet string, project string) (*Client, error) {
	var client = &Client{
		BqClient:         bqClient,
		ProjectId:        project,
		DataSet:          dataSet,
		ChromiumRdbTable: testProject + "." + testDataset + "." + fakeChromiumRdb,
		ChromeRdbTable:   testProject + "." + testDataset + "." + fakeChromeRdb,
		AttemptsTable:    testProject + "." + testDataset + "." + fakeAttempts,
	}
	err := client.Init(sqlDir)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func createRollupFromResults(ctx context.Context, client *Client, project string, dataSet string, rdbTable string, date string, results []*fakeRdbResult) error {
	if date != "" {
		tz, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			return err
		}
		for _, result := range results {
			date, err := civil.ParseDate(date)
			if err != nil {
				return err
			}
			result.partitionTime = date.In(tz)
		}
	}

	createFakeRdb(ctx, client.BqClient, project, dataSet, rdbTable, results)

	if len(results) == 0 {
		return nil
	}

	maxTime := results[0].ResultTime()
	minTime := results[0].ResultTime()
	for _, result := range results {
		if result == nil {
			return errors.New("Received a nul result")
		}
		resultTime := result.ResultTime()
		if resultTime.Compare(maxTime) > 0 {
			maxTime = resultTime
		} else if resultTime.Compare(minTime) < 0 {
			minTime = resultTime
		}
	}

	return client.UpdateSummary(ctx, civil.DateOf(minTime), civil.DateOf(maxTime))
}

func createFakeRdb(ctx context.Context, client *bigquery.Client, project string, dataSet string, rdbTable string, results []*fakeRdbResult) error {
	inserter := client.Dataset(dataSet).Table(rdbTable).Inserter()

	// The table should be create
	var err error
	attempt := 0
	for attempt < 10 {
		err = inserter.Put(ctx, results)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
		attempt += 1
	}

	return err
}

func getSchema(ctx context.Context, project string, dataset string, table string) (bigquery.Schema, error) {
	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	metadata, err := client.Dataset(dataset).Table(table).Metadata(ctx)
	if err != nil {
		return nil, err
	}
	return metadata.Schema, nil
}

func ensureTable(ctx context.Context, bqClient *bigquery.Client, project string, dataset string, tableName string, testDataset *bigquery.Dataset, fakeTableName string) error {
	schema, err := getSchema(ctx, project, dataset, tableName)
	if err != nil {
		return err
	}

	table := testDataset.Table(fakeTableName)
	if _, err := table.Metadata(ctx); err == nil {
		table.Delete(ctx)
	}
	if err := table.Create(ctx,
		&bigquery.TableMetadata{
			Schema:         schema,
			ExpirationTime: time.Now().Add(1 * time.Hour),
		}); err != nil {
		return err
	}

	// Make sure the table propagated
	tableCreated := false
	attempt := 0
	for attempt < 60 {
		_, err := table.Metadata(ctx)
		tableCreated = err == nil
		if tableCreated {
			break
		}
		time.Sleep(time.Second)
		attempt += 1
	}

	return nil
}

func ensureTables(ctx context.Context, client *bigquery.Client) error {
	// Delete the dataset to avoid previous runs
	// TODO(sshrimp): This will fail if there is data in the streaming
	// buffer. A better way of ensuring we have a clean dataset should
	// created since there's no way to manually flush the streaming buffer
	// which can take up to 90 minutes to flush on it's own
	datasets := client.Datasets(ctx)
	var testSet *bigquery.Dataset
	for {
		dataset, err := datasets.Next()
		if err != nil {
			return err
		}
		if dataset.DatasetID == testDataset {
			testSet = dataset
			break
		}
	}
	if testSet != nil {
		if err := testSet.DeleteWithContents(ctx); err != nil {
			return err
		}
	}
	dataset := client.Dataset(testDataset)
	if err := dataset.Create(ctx, &bigquery.DatasetMetadata{}); err != nil {
		return err
	}

	// Create the summary tables
	r := strings.NewReplacer(
		"DATASET", testDataset,
		"APP_ID", testProject,
	)
	for _, queryFile := range createQueries {
		queryString, err := parseCreateQuery(r, filepath.Join(sqlDir, queryFile))
		if err != nil {
			return err
		}

		query := client.Query(queryString)

		job, err := query.Run(ctx)
		if err != nil {
			return errors.Annotate(err, "failed to start the job").Err()
		}

		jobStatus, err := job.Wait(ctx)
		if err != nil {
			return errors.Annotate(err, "failed to finish the query").Err()
		}
		err = jobStatus.Err()
		if err != nil {
			return err
		}
	}

	if err := ensureTable(ctx, client, "chrome-luci-data", "chromium", "try_test_results", testSet, fakeChromiumRdb); err != nil {
		return err
	}
	if err := ensureTable(ctx, client, "chrome-luci-data", "chrome", "try_test_results", testSet, fakeChromeRdb); err != nil {
		return err
	}
	if err := ensureTable(ctx, client, "commit-queue", "chromium", "attempts", testSet, fakeAttempts); err != nil {
		return err
	}
	return nil
}

func parseCreateQuery(r *strings.Replacer, fileName string) (string, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return r.Replace(string(bytes)), nil
}

func checkForRows(ctx context.Context, client *bigquery.Client, q string) (bool, error) {
	query := client.Query(q)
	job, err := query.Run(ctx)
	if err != nil {
		return true, err
	}
	it, err := job.Read(ctx)
	if err != nil {
		return true, err
	}

	return it.TotalRows > 0, nil
}

func checkForDuplicateRows(ctx context.Context, client *bigquery.Client) error {
	query := `
	SELECT COUNT(*) rowCount
	FROM ` + testProject + `.` + testDataset + `.raw_metrics
	GROUP BY
		date, test_id, repo, project, bucket, builder, test_suite, target_platform, variant_hash
	HAVING rowCount > 1`

	duplicates, err := checkForRows(ctx, client, query)
	if err != nil {
		return err
	}
	if duplicates {
		return errors.New("Duplicate rows created in raw_metrics")
	}
	return nil
}

func getTestIdFromResponse(resp *api.FetchTestMetricsResponse, testId string) *api.TestDateMetricData {
	var testResult *api.TestDateMetricData
	for _, t := range resp.Tests {
		if t.TestId == testId {
			testResult = t
			break
		}
	}
	return testResult
}

func getBuilderVariantFromTest(testSummary *api.TestDateMetricData, builder string) *api.TestVariantData {
	var variant *api.TestVariantData
	for _, v := range testSummary.Variants {
		if v.Builder == builder {
			variant = v
			break
		}
	}
	return variant
}
