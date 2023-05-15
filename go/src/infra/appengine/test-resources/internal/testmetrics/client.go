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
// as well. The dates are inclusive
func (c *Client) UpdateSummary(ctx context.Context, fromDate civil.Date, toDate civil.Date) error {
	q := c.BqClient.Query(c.updateDailySummarySql)

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
	if err := job_status.Err(); err != nil {
		return err
	}

	return nil
}
