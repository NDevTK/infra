// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testmetrics

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/iterator"

	"infra/appengine/test-resources/api"
)

var (
	periodToTestMetricTable = map[api.Period]string{
		api.Period_DAY:  "test_metrics",
		api.Period_WEEK: "weekly_test_metrics",
	}
	periodToFileMetricTable = map[api.Period]string{
		api.Period_DAY:  "file_metrics",
		api.Period_WEEK: "weekly_file_metrics",
	}
	sortTypeSqlLookup = map[api.SortType]string{
		api.SortType_SORT_NUM_RUNS:      "num_runs",
		api.SortType_SORT_NUM_FAILURES:  "num_failures",
		api.SortType_SORT_AVG_RUNTIME:   "avg_runtime",
		api.SortType_SORT_TOTAL_RUNTIME: "total_runtime",
	}
)

// Client is used to fetch metrics from a given data source.
type Client struct {
	BqClient                   *bigquery.Client
	ProjectId                  string
	DataSet                    string
	updateDailySummarySql      string
	updateWeeklySummarySql     string
	updateFileSummarySql       string
	updateWeeklyFileSummarySql string
}

// Initializes the testmetric client
func (c *Client) Init() error {
	if c.ProjectId == "" {
		c.ProjectId = "chrome-resources-staging"
	}
	if c.DataSet == "" {
		c.DataSet = "test_results"
	}
	bytes, err := os.ReadFile("sql/update_test_metrics.sql")
	if err != nil {
		return err
	}
	c.updateDailySummarySql = fmt.Sprintf(string(bytes), c.ProjectId, c.DataSet)
	bytes, err = os.ReadFile("sql/update_weekly_test_metrics.sql")
	if err != nil {
		return err
	}
	c.updateWeeklySummarySql = fmt.Sprintf(string(bytes), c.ProjectId, c.DataSet, c.ProjectId, c.DataSet)
	bytes, err = os.ReadFile("sql/update_file_metrics.sql")
	if err != nil {
		return err
	}
	c.updateFileSummarySql = fmt.Sprintf(string(bytes), c.ProjectId, c.DataSet, c.ProjectId, c.DataSet)
	bytes, err = os.ReadFile("sql/update_weekly_file_metrics.sql")
	if err != nil {
		return err
	}
	c.updateWeeklyFileSummarySql = fmt.Sprintf(string(bytes), c.ProjectId, c.DataSet, c.ProjectId, c.DataSet)
	return nil
}

func bqToDateArray(dates []string) ([]civil.Date, error) {
	ret := make([]civil.Date, len(dates))
	for i, date := range dates {
		d, err := civil.ParseDate(date)
		if err != nil {
			return nil, err
		}
		ret[i] = d
	}
	return ret, nil
}

func (c *Client) ListComponents(ctx context.Context, req *api.ListComponentsRequest) (*api.ListComponentsResponse, error) {
	query := "SELECT DISTINCT component FROM chrome-metadata.chrome.monorail_component_owners ORDER BY component"
	q := c.BqClient.Query(query)

	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}
	response := &api.ListComponentsResponse{}
	type row struct {
		Component string `bigquery:"component"`
	}
	for {
		var rowVals row
		err = it.Next(&rowVals)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next component row").Err()
		}
		response.Components = append(response.Components, rowVals.Component)
	}
	return response, nil
}

// Fetches requested metrics for the provided days and filters
func (c *Client) FetchMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	dates, err := bqToDateArray(req.GetDates())
	if err != nil {
		return nil, err
	}

	// If there's a filter we have to aggregate it now, otherwise use the
	// pre-aggregated metric
	metricNames := make([]string, len(req.Metrics))
	for i := 0; i < len(req.Metrics); i++ {
		name := MetricSqlName(req.Metrics[i])
		if req.Filter == "" {
			metricNames[i] = name
		} else {
			aggFunc := "SUM"
			if req.Metrics[i] == api.MetricType_AVG_RUNTIME {
				aggFunc = "AVG"
			}
			metricNames[i] = `(SELECT ` + aggFunc + `(f.` + name + `) FROM UNNEST(m.variant_summaries) f
		WHERE
			REGEXP_CONTAINS(test_name, @string_filter) OR
			REGEXP_CONTAINS(file_name, @string_filter) OR
			# Only display variant matches if the variant is what matches
			REGEXP_CONTAINS(builder, @string_filter) OR
			REGEXP_CONTAINS(test_suite, @string_filter) LIMIT 1) AS ` + name
		}
	}

	// Terms for converting the rolled up variants (including thing like
	// project, bucket etc) into the builder and suite variant
	metricAggregations := make([]string, len(req.Metrics))
	for i := 0; i < len(req.Metrics); i++ {
		name := MetricSqlName(req.Metrics[i])
		if req.Metrics[i] == api.MetricType_AVG_RUNTIME {
			metricAggregations[i] = `AVG(` + name + `) AS ` + name
		} else {
			metricAggregations[i] = `SUM(` + name + `) AS ` + name
		}
	}

	table, ok := periodToTestMetricTable[req.Period]
	if !ok {
		return nil, errors.Reason("Received unsupported period request: '%s'", req.Period).Err()
	}

	sortMetric := "test_id"
	// A default value of 0 maps to the name which for test based fetches is
	// test_id. Other values map to metrics in both file and directory tables
	if req.Sort != nil && req.Sort.Metric != api.SortType_SORT_NAME {
		sortMetric = sortTypeSqlLookup[req.Sort.Metric]
	}
	sortDirection := "ASC"
	if req.Sort != nil && !req.Sort.Ascending {
		sortDirection = "DESC"
	}

	string_filter := strings.Join(strings.Split(req.Filter, " "), "|")

	query := `
WITH raw AS (
	SELECT
		m.date,
		m.test_id,
		m.test_name,
		m.file_name,
		` + strings.Join(metricNames, ",\n\t\t") + `,
		(
			SELECT ARRAY_AGG(builder_suite_summary) FROM (
				SELECT STRUCT(
					v.builder AS builder,
					v.test_suite AS test_suite,
					` + strings.Join(metricAggregations, ",\n\t\t\t\t\t") + `
				) AS builder_suite_summary
				FROM m.variant_summaries v
				WHERE (@string_filter = "" OR
					# If it matches the parent ID display everything
					REGEXP_CONTAINS(test_name, @string_filter) OR
					REGEXP_CONTAINS(file_name, @string_filter) OR
					# Only display variant matches if the variant is what matches
					REGEXP_CONTAINS(builder, @string_filter) OR
					REGEXP_CONTAINS(test_suite, @string_filter))
				GROUP BY builder, test_suite
				)
			) AS variants
	FROM
		` + c.ProjectId + `.` + c.DataSet + `.` + table + ` AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
		AND component = @component
)
SELECT * FROM raw WHERE ARRAY_LENGTH(raw.variants) > 0
ORDER BY ` + sortMetric + ` ` + sortDirection + `
LIMIT @page_size OFFSET @page_offset`

	q := c.BqClient.Query(query)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "dates", Value: dates},
		{Name: "page_size", Value: req.PageSize + 1},
		{Name: "page_offset", Value: req.PageOffset},
		{Name: "component", Value: req.Component},
		{Name: "string_filter", Value: string_filter},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}

	// maps for quick lookup of the test id
	testIdToTestDateMetricData := make(map[string]*api.TestDateMetricData)
	variantHashToTestDateMetricData := make(map[string]map[string]*api.TestVariantData)

	response := &api.FetchTestMetricsResponse{
		LastPage: int64(it.TotalRows) != req.PageSize+1,
	}
	for i := int64(0); i < req.PageSize; i++ {
		var rowVals rowLoader
		err = it.Next(&rowVals)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next test summary row").Err()
		}
		testId := rowVals.String("test_id")
		testIdData, ok := testIdToTestDateMetricData[testId]
		if !ok {
			testIdData = &api.TestDateMetricData{
				TestId:   testId,
				TestName: rowVals.NullString("test_name").StringVal,
				FileName: rowVals.NullString("file_name").StringVal,
				Metrics:  make(map[string]*api.TestMetricsArray),
			}
			testIdToTestDateMetricData[testId] = testIdData
			variantHashToTestDateMetricData[testId] = make(map[string]*api.TestVariantData)
			// Add to the actual struct being returned
			response.Tests = append(response.Tests, testIdData)
		}

		// Each row is the summary of a test id for a day
		date := rowVals.Date("date").String()

		// Handle the test's rollup metrics
		testIdData.Metrics[date] = &api.TestMetricsArray{
			Data: rowVals.Metrics(req.Metrics),
		}

		// Handle the indvidiual variant metrics
		repeated := true
		val, err := rowVals.valueWithType("variants", bigquery.RecordFieldType, repeated)
		if err != nil {
			return nil, err
		}

		i, _ := rowVals.fieldIndex("variants")
		rowSchema := rowVals.schema[i].Schema

		variantRows := val.([]bigquery.Value)
		for _, variantRow := range variantRows {
			var variantRowVals rowLoader
			if err := variantRowVals.Load(variantRow.([]bigquery.Value), rowSchema); err != nil {
				return nil, err
			}

			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, errors.Annotate(err, "obtain next variant summary row").Err()
			}

			builder := variantRowVals.NullString("builder").StringVal
			suite := variantRowVals.NullString("test_suite").StringVal
			builderSuite := builder + ":" + suite
			builderSuiteData, ok := variantHashToTestDateMetricData[testId][builderSuite]
			if !ok {
				fields := ""
				for _, field := range rowSchema {
					fields += field.Name
				}
				builderSuiteData = &api.TestVariantData{
					Builder: variantRowVals.NullString("builder").StringVal,
					Suite:   variantRowVals.NullString("test_suite").StringVal,
					Metrics: make(map[string]*api.TestMetricsArray),
				}
				variantHashToTestDateMetricData[testId][builderSuite] = builderSuiteData
				testIdData.Variants = append(testIdData.Variants, builderSuiteData)
			}
			builderSuiteData.Metrics[date] = &api.TestMetricsArray{
				Data: variantRowVals.Metrics(req.Metrics),
			}

			if err := variantRowVals.Error(); err != nil {
				return nil, err
			}
		}
		if err := rowVals.Error(); err != nil {
			return nil, err
		}
	}

	return response, nil
}

// Fetches requested metrics for the provided days and filters for a given
// directory node. A directory node represents a directory or file and the
// metrics are the combined metrics of the tests in these locations
func (c *Client) FetchDirectoryMetrics(ctx context.Context, req *api.FetchDirectoryMetricsRequest) (*api.FetchDirectoryMetricsResponse, error) {
	dates, err := bqToDateArray(req.GetDates())
	if err != nil {
		return nil, err
	}

	// If there's a filter we have to aggregate it now, otherwise use the
	// pre-aggregated metric
	metricNames := make([]string, len(req.Metrics))
	for i := 0; i < len(req.Metrics); i++ {
		name := MetricSqlName(req.Metrics[i])
		if req.Filter == "" {
			metricNames[i] = name
		} else {
			metricNames[i] = `(SELECT SUM(f.` + name + `) FROM UNNEST(child_file_summaries) f WHERE REGEXP_CONTAINS(f.file_name, @string_filter) LIMIT 1) AS ` + name
		}
	}

	table, ok := periodToFileMetricTable[req.Period]
	if !ok {
		return nil, errors.Reason("Received unsupported period request: '%s'", req.Period).Err()
	}

	sortMetric := "node_name"
	// A default value of 0 maps to the name which for file based fetches is
	// node_name. Other values map to metrics in both file and directory tables
	if req.Sort != nil && req.Sort.Metric != api.SortType_SORT_NAME {
		sortMetric = sortTypeSqlLookup[req.Sort.Metric]
	}
	sortDirection := "ASC"
	if req.Sort != nil && !req.Sort.Ascending {
		sortDirection = "DESC"
	}

	string_filter := strings.Join(strings.Split(req.Filter, " "), "|")

	query := `
SELECT
	date,
	node_name,
	ARRAY_REVERSE(SPLIT(node_name, '/'))[SAFE_OFFSET(0)] AS display_name,
	is_file,
	` + strings.Join(metricNames, ",\n\t") + `,
FROM ` + c.ProjectId + `.` + c.DataSet + `.` + table + `
WHERE
	STARTS_WITH(node_name, @parent || "/") AND
	-- The child folders and files can't have a / after the parent's name
	REGEXP_CONTAINS(SUBSTR(node_name, LENGTH(@parent) + 2), "^[^/]*$")
	AND DATE(date) IN UNNEST(@dates)
	AND component = @component
	AND (@string_filter = "" OR EXISTS(SELECT 0 FROM UNNEST(child_file_summaries) AS f WHERE REGEXP_CONTAINS(f.file_name, @string_filter)))
ORDER BY ` + sortMetric + ` ` + sortDirection

	q := c.BqClient.Query(query)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "dates", Value: dates},
		{Name: "component", Value: req.Component},
		{Name: "parent", Value: req.ParentId},
		{Name: "string_filter", Value: string_filter},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}

	// maps for quick lookup of the node id
	filenameToTestDateMetricData := make(map[string]*api.DirectoryNode)

	response := &api.FetchDirectoryMetricsResponse{}
	for {
		var rowVals rowLoader
		err = it.Next(&rowVals)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next test summary row").Err()
		}
		nodeName := rowVals.String("node_name")
		dirNode, ok := filenameToTestDateMetricData[nodeName]
		if !ok {
			nodeType := api.DirectoryNodeType_DIRECTORY
			if rowVals.Bool("is_file") {
				nodeType = api.DirectoryNodeType_FILENAME
			}
			dirNode = &api.DirectoryNode{
				Id:      nodeName,
				Metrics: make(map[string]*api.TestMetricsArray),
				Name:    rowVals.String("display_name"),
				Type:    nodeType,
			}
			filenameToTestDateMetricData[nodeName] = dirNode
			// Add to the actual struct being returned
			response.Node = append(response.Node, dirNode)
		}

		// Each row is the summary of a node for a day
		date := rowVals.Date("date").String()

		// Handle the metric rollup metrics
		dirNode.Metrics[date] = &api.TestMetricsArray{
			Data: rowVals.Metrics(req.Metrics),
		}

		if err := rowVals.Error(); err != nil {
			return nil, err
		}
	}

	return response, nil
}

// Updates the summary tables between the days in the provided
// UpdateMetricsTableRequest. All rollups (e.g. weekly/monthly) will be updated
// as well. The dates are inclusive
func (c *Client) UpdateSummary(ctx context.Context, fromDate civil.Date, toDate civil.Date) error {
	err := c.runUpdateSummary(ctx, fromDate, toDate, c.updateDailySummarySql)
	if err != nil {
		return err
	}
	err = c.runUpdateSummary(ctx, fromDate, toDate, c.updateWeeklySummarySql)
	if err != nil {
		return err
	}
	err = c.runUpdateSummary(ctx, fromDate, toDate, c.updateFileSummarySql)
	if err != nil {
		return err
	}
	err = c.runUpdateSummary(ctx, fromDate, toDate, c.updateWeeklyFileSummarySql)
	return err
}

func (c *Client) runUpdateSummary(ctx context.Context, fromDate civil.Date, toDate civil.Date, query string) error {
	q := c.BqClient.Query(query)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "from_date", Value: fromDate},
		{Name: "to_date", Value: toDate},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return err
	}

	job_status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	return job_status.Err()
}
