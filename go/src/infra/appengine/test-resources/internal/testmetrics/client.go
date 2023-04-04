// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testmetrics

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"

	"infra/appengine/test-resources/api"
)

// Client is used to fetch metrics from a given data source.
type Client struct {
	BqClient              *bigquery.Client
	updateDailySummarySql string
}

// Initializes the testmetric client
func (c *Client) Init(project_id string) error {
	bytes, err := os.ReadFile("sql/update_test_metrics.sql")
	if err != nil {
		return err
	}
	c.updateDailySummarySql = fmt.Sprintf(string(bytes), project_id)
	return nil
}

// Updates the summary tables between the days in the provided
// UpdateMetricsTableRequest. All rollups (e.g. weekly/monthly) will be updated
// as well
func (c *Client) UpdateSummary(ctx context.Context, req *api.UpdateMetricsTableRequest) (*api.UpdateMetricsTableResponse, error) {
	fromDate, err := civil.ParseDate(req.FromDate)
	if err != nil {
		return nil, err
	}
	toDate, err := civil.ParseDate(req.ToDate)
	if err != nil {
		return nil, err
	}
	// toDate should be inclusive
	for date := fromDate; !toDate.Before(date); date = date.AddDays(1) {
		err := c.UpdateDateSummary(ctx, date)
		if err != nil {
			return nil, err
		}
	}
	return &api.UpdateMetricsTableResponse{}, nil
}

// Updates the summary tables for a single date. All rollups
// (e.g. weekly/monthly) will be updated as well
func (c *Client) UpdateDateSummary(ctx context.Context, date civil.Date) error {
	q := c.BqClient.Query(c.updateDailySummarySql)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "run_date", Value: date},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return err
	}

	job_status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := job_status.Err(); err != nil {
		return err
	}

	return err
}
