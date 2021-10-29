// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clustering

import (
	"context"
	"fmt"
	"infra/appengine/weetbix/internal/config"
	"math"

	"go.chromium.org/luci/common/errors"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// ImpactfulClusterReadOptions specifies options for ReadImpactfulClusters().
type ImpactfulClusterReadOptions struct {
	// Project is the LUCI Project for which analysis is being performed.
	Project string
	// Thresholds is the set of thresholds, which if any are met
	// or exceeded, should result in the cluster being returned.
	Thresholds *config.ImpactThreshold
	// AlwaysIncludeClusterIDs is the set of clusters for which analysis
	// should always be read, if available. This is typically the set of
	// clusters for which bugs have been filed.
	AlwaysIncludeClusterIDs []string
}

// Cluster represents a group of failures, with associated impact metrics.
type Cluster struct {
	ClusterID              string              `json:"clusterId"`
	UnexpectedFailures1d   int64               `json:"unexpectedFailures1d"`
	UnexpectedFailures3d   int64               `json:"unexpectedFailures3d"`
	UnexpectedFailures7d   int64               `json:"unexpectedFailures7d"`
	UnexoneratedFailures1d int64               `json:"unexoneratedFailures1d"`
	UnexoneratedFailures3d int64               `json:"unexoneratedFailures3d"`
	UnexoneratedFailures7d int64               `json:"unexoneratedFailures7d"`
	AffectedRuns1d         int64               `json:"affectedRuns1d"`
	AffectedRuns3d         int64               `json:"affectedRuns3d"`
	AffectedRuns7d         int64               `json:"affectedRuns7d"`
	AffectedTests1d        []SubCluster        `json:"affectedTests1d"`
	AffectedTests3d        []SubCluster        `json:"affectedTests3d"`
	AffectedTests7d        []SubCluster        `json:"affectedTests7d"`
	ExampleFailureReason   bigquery.NullString `json:"exampleFailureReason"`
}

type SubCluster struct {
	Value     string `json:"value"`
	Num_Fails int    `json:"numFails"`
}

// NewClient creates a new client for reading clusters. Close() MUST
// be called after you have finished using this client.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// Client may be used to read Weetbix clusters.
type Client struct {
	client *bigquery.Client
}

// Close releases any resources held by the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// ReadImpactfulClusters reads clusters exceeding specified impact metrics, or are otherwise
// nominated to be read.
func (c *Client) ReadImpactfulClusters(ctx context.Context, opts ImpactfulClusterReadOptions) ([]*Cluster, error) {
	if opts.Project != "chromium" {
		return nil, errors.New("chromium is the only project for which analysis is currently supported")
	}
	if opts.Thresholds == nil {
		return nil, errors.New("thresholds must be specified")
	}

	q := c.client.Query(`
	SELECT cluster_id as ClusterID,
		unexpected_failures_1d as UnexpectedFailures1d,
		unexpected_failures_3d as UnexpectedFailures3d,
		unexpected_failures_7d as UnexpectedFailures7d,
		unexonerated_failures_1d as UnexoneratedFailures1d,
		unexonerated_failures_3d as UnexoneratedFailures3d,
		unexonerated_failures_7d as UnexoneratedFailures7d,
		affected_runs_1d as AffectedRuns1d,
		affected_runs_3d as AffectedRuns3d,
		affected_runs_7d as AffectedRuns7d,
		example_failure_reason.primary_error_message as ExampleFailureReason
	FROM chromium.clusters
	WHERE (unexpected_failures_1d > @unexpFailThreshold1d
		OR unexpected_failures_3d > @unexpFailThreshold3d
		OR unexpected_failures_7d > @unexpFailThreshold7d)
		OR cluster_id IN UNNEST(@alwaysSelectClusters)
	ORDER BY unexpected_failures_1d DESC, unexpected_failures_3d DESC, unexpected_failures_7d DESC
	`)
	unexpectedFailures1d := int64(math.MaxInt64)
	if opts.Thresholds.UnexpectedFailures_1D != nil {
		unexpectedFailures1d = *opts.Thresholds.UnexpectedFailures_1D
	}
	unexpectedFailures3d := int64(math.MaxInt64)
	if opts.Thresholds.UnexpectedFailures_3D != nil {
		unexpectedFailures3d = *opts.Thresholds.UnexpectedFailures_3D
	}
	unexpectedFailures7d := int64(math.MaxInt64)
	if opts.Thresholds.UnexpectedFailures_7D != nil {
		unexpectedFailures7d = *opts.Thresholds.UnexpectedFailures_7D
	}
	// TODO(crbug.com/1243174): This will not scale if the set of
	// cluster IDs to always select grows too large.
	q.Parameters = []bigquery.QueryParameter{
		{Name: "unexpFailThreshold1d", Value: unexpectedFailures1d},
		{Name: "unexpFailThreshold3d", Value: unexpectedFailures3d},
		{Name: "unexpFailThreshold7d", Value: unexpectedFailures7d},
		{Name: "alwaysSelectClusters", Value: opts.AlwaysIncludeClusterIDs},
	}
	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying clusters").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "obtain cluster iterator").Err()
	}
	clusters := []*Cluster{}
	for {
		row := &Cluster{}
		err := it.Next(row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next cluster row").Err()
		}
		clusters = append(clusters, row)
	}
	return clusters, nil
}

// ReadCluster reads information about a single cluster.
func (c *Client) ReadCluster(ctx context.Context, clusterID string) (*Cluster, error) {
	q := c.client.Query(`
	SELECT
		cluster_id as ClusterID,
		unexpected_failures_1d as UnexpectedFailures1d,
		unexpected_failures_3d as UnexpectedFailures3d,
		unexpected_failures_7d as UnexpectedFailures7d,
		unexonerated_failures_1d as UnexoneratedFailures1d,
		unexonerated_failures_3d as UnexoneratedFailures3d,
		unexonerated_failures_7d as UnexoneratedFailures7d,
		affected_runs_1d as AffectedRuns1d,
		affected_runs_3d as AffectedRuns3d,
		affected_runs_7d as AffectedRuns7d,
		affected_tests_1d as AffectedTests1d,
		affected_tests_3d as AffectedTests3d,
		affected_tests_7d as AffectedTests7d,
		example_failure_reason.primary_error_message as ExampleFailureReason,
		example_result_id as ExampleResultID
		FROM chromium.clusters
		WHERE cluster_id = @clusterID
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "clusterID", Value: clusterID},
	}
	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying cluster").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "obtain cluster iterator").Err()
	}
	clusters := []*Cluster{}
	for {
		row := &Cluster{}
		err := it.Next(row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next cluster row").Err()
		}
		clusters = append(clusters, row)
	}
	if len(clusters) == 0 {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}
	return clusters[0], nil
}
