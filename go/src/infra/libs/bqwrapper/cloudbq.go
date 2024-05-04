// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bqwrapper

import (
	"context"

	"cloud.google.com/go/bigquery"
)

// CloudBQ wraps the prod client.
type CloudBQ struct {
	client *bigquery.Client
}

// Assert that CloudBQ satisfies the right interface.
var _ BQIf = &CloudBQ{}

// NewCloudBQ makes a new one.
func NewCloudBQ(client *bigquery.Client) *CloudBQ {
	return &CloudBQ{
		client: client,
	}
}

// Put writes a record to BigQuery.
func (cbq *CloudBQ) Put(ctx context.Context, projectID string, dataset string, table string, data []bigquery.ValueSaver) error {
	return cbq.client.DatasetInProject(projectID, dataset).Table(table).Inserter().Put(ctx, data)
}
