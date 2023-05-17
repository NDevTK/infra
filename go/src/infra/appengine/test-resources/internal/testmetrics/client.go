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

// Client is used to fetch metrics from a given data source.
type Client struct {
	BqClient               *bigquery.Client
	ProjectId              string
	updateDailySummarySql  string
	updateWeeklySummarySql string
}

// Initializes the testmetric client
func (c *Client) Init() error {
	if c.ProjectId == "" {
		c.ProjectId = "chrome-resources-staging"
	}
	bytes, err := os.ReadFile("sql/update_test_metrics.sql")
	if err != nil {
		return err
	}
	c.updateDailySummarySql = fmt.Sprintf(string(bytes), c.ProjectId)
	bytes, err = os.ReadFile("sql/update_weekly_test_metrics.sql")
	if err != nil {
		return err
	}
	c.updateWeeklySummarySql = fmt.Sprintf(string(bytes), c.ProjectId)
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

// Fetches requested metrics for the provided days and filters
func (c *Client) FetchMetrics(ctx context.Context, req *api.FetchTestMetricsRequest) (*api.FetchTestMetricsResponse, error) {
	dates, err := bqToDateArray(req.GetDates())
	if err != nil {
		return nil, err
	}

	metricNames := make([]string, len(req.Metrics))
	for i := 0; i < len(req.Metrics); i++ {
		// Our column names aren't quite the metric type string
		metricNames[i] = MetricSqlName(req.Metrics[i])
	}

	query := `
	SELECT
		m.date,
		m.test_id,
		m.test_name,
		m.file_name,
		` + strings.Join(metricNames, ",\n") + `,
		((SELECT ARRAY_AGG(STRUCT(
			v.variant_hash AS variant_hash,
			v.target_platform AS target_platform,
			v.builder AS builder,
			v.test_suite AS test_suite,
			` + strings.Join(metricNames, ",\n") + `
		)) FROM m.variant_summaries v)) AS variants
	FROM
		` + c.ProjectId + `.test_results.test_metrics AS m
	WHERE
		DATE(date) IN UNNEST(@dates)
	LIMIT @page_size OFFSET @page`

	q := c.BqClient.Query(query)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "dates", Value: dates},
		{Name: "page_size", Value: req.PageSize},
		{Name: "page", Value: req.Page},
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

	response := &api.FetchTestMetricsResponse{}
	for {
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

			variantHash := variantRowVals.String("variant_hash")
			variantData, ok := variantHashToTestDateMetricData[testId][variantHash]
			if !ok {
				fields := ""
				for _, field := range rowSchema {
					fields += field.Name
				}
				variantData = &api.TestVariantData{
					Builder: variantRowVals.NullString("builder").StringVal,
					Suite:   variantRowVals.NullString("test_suite").StringVal,
					Metrics: make(map[string]*api.TestMetricsArray),
				}
				variantHashToTestDateMetricData[testId][variantHash] = variantData
				testIdData.Variants = append(testIdData.Variants, variantData)
			}
			variantData.Metrics[date] = &api.TestMetricsArray{
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

// Updates the summary tables between the days in the provided
// UpdateMetricsTableRequest. All rollups (e.g. weekly/monthly) will be updated
// as well. The dates are inclusive
func (c *Client) UpdateSummary(ctx context.Context, fromDate civil.Date, toDate civil.Date) error {
	err := c.runUpdateSummary(ctx, fromDate, toDate, c.updateDailySummarySql)
	if err != nil {
		return err
	}
	err = c.runUpdateSummary(ctx, fromDate, toDate, c.updateWeeklySummarySql)
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
