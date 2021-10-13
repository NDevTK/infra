// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clusteredfailures

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
)

// Entry describes a clustered failure. Specifically, it tracks the inclusion
// of a failure in a cluster. A LastUpdated time is included so that the
// entry may be versioned (e.g. so that a failure may be removed from the
// cluster if the cluster definition changes, by setting IsIncluded = false).
type Entry struct {
	// Primary key fields.
	Project          string    `bigquery:"project"`
	ClusterAlgorithm string    `bigquery:"cluster_algorithm"`
	ClusterID        string    `bigquery:"cluster_id"`
	TestResultID     string    `bigquery:"test_result_id"`
	LastUpdated      time.Time `bigquery:"last_updated"`

	// Partition fields.
	PartitionTime time.Time `bigquery:"partition_time"`

	// Inclusion Fields.
	IsIncluded                 bool `bigquery:"is_included"`
	IsIncludedWithHighPriority bool `bigquery:"is_included_with_high_priority"`

	// Fields assigned during ingestion.
	ChunkID    string `bigquery:"chunk_id"`
	ChunkIndex int64  `bigquery:"chunk_index"`

	Realm                     string              `bigquery:"realm"`
	TestID                    string              `bigquery:"test_id"`
	Variant                   []*Variant          `bigquery:"variant"`
	VariantHash               string              `bigquery:"variant_hash"`
	FailureReason             *FailureReason      `bigquery:"failure_reason"`
	Component                 string              `bigquery:"component"`
	StartTime                 time.Time           `bigquery:"start_time"`
	Duration                  time.Duration       `bigquery:"duration"`
	IsExonerated              bool                `bigquery:"is_exonerated"`
	RootInvocationID          string              `bigquery:"root_invocation_id"`
	RootInvocationResultSeq   int64               `bigquery:"root_invocation_result_seq"`
	RootInvocationResultCount int64               `bigquery:"root_invocation_result_count"`
	IsRootInvocationBlocked   bool                `bigquery:"is_root_invocation_blocked"`
	TaskID                    string              `bigquery:"task_id"`
	TaskResultSeq             int64               `bigquery:"task_result_seq"`
	TaskResultCount           int64               `bigquery:"task_result_count"`
	IsTaskBlocked             bool                `bigquery:"is_task_blocked"`
	CQRunID                   bigquery.NullString `bigquery:"cq_run_id"`
}

// Variant is a key-value pair describing how the test was executed, for tests
// that can be executed in multiple different ways.
// For example, the operating system.
type Variant struct {
	Key   string `bigquery:"key"`
	Value string `bigquery:"value"`
}

// FailureReason contains information about why a test failed.
type FailureReason struct {
	PrimaryErrorMessage string `bigquery:"primary_error_message"`
}

// NewClient creates a new client for exporting clustered failures.
// Call Close() after you have finished using this client.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// Client provides methods to export clustered failures to BigQuery.
type Client struct {
	client *bigquery.Client
}

// Close releases any resources held by the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Insert inserts the given rows in BigQuery.
func (c *Client) Insert(ctx context.Context, rows []*Entry) error {
	inserter := c.client.Dataset("analysis").Table("clustered_failures").Inserter()
	err := inserter.Put(ctx, rows)
	if err != nil {
		errors.Annotate(err, "inserting clustered failures").Err()
	}
	return nil
}
