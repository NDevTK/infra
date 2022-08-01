// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analysis

import (
	"context"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/trace"
	"google.golang.org/api/iterator"

	"infra/appengine/weetbix/internal/bqutil"
	"infra/appengine/weetbix/internal/clustering"
)

type ClusterFailure struct {
	Realm             bigquery.NullString `json:"realm"`
	TestID            bigquery.NullString `json:"testId"`
	Variant           []*Variant          `json:"variant"`
	PresubmitRunID    *PresubmitRunID     `json:"presubmitRunId"`
	PresubmitRunOwner bigquery.NullString `json:"presubmitRunOwner"`
	PresubmitRunMode  bigquery.NullString `json:"presubmitRunMode"`
	// TODO(b/239768873): Remove when legacy cluster failures endpoint deleted.
	Changelist    *Changelist `json:"changelist"`
	Changelists   []*Changelist
	PartitionTime bigquery.NullTimestamp `json:"partitionTime"`
	Exonerations  []*Exoneration         `json:"exonerations"`
	// weetbix.v1.BuildStatus, without "BUILD_STATUS_" prefix.
	BuildStatus                 bigquery.NullString `json:"buildStatus"`
	IsBuildCritical             bigquery.NullBool   `json:"isBuildCritical"`
	IngestedInvocationID        bigquery.NullString `json:"ingestedInvocationId"`
	IsIngestedInvocationBlocked bigquery.NullBool   `json:"isIngestedInvocationBlocked"`
	Count                       int32               `json:"count"`
}

type Exoneration struct {
	// weetbix.v1.ExonerationReason value. E.g. "OCCURS_ON_OTHER_CLS".
	Reason bigquery.NullString `json:"reason"`
}

type Variant struct {
	Key   bigquery.NullString `json:"key"`
	Value bigquery.NullString `json:"value"`
}

type PresubmitRunID struct {
	System bigquery.NullString `json:"system"`
	ID     bigquery.NullString `json:"id"`
}

type Changelist struct {
	Host     bigquery.NullString `json:"host"`
	Change   bigquery.NullInt64  `json:"change"`
	Patchset bigquery.NullInt64  `json:"patchset"`
}

type ReadClusterFailuresOptions struct {
	// The LUCI Project.
	Project   string
	ClusterID clustering.ClusterID
	Realms    []string
}

// ReadClusterFailures reads the latest 2000 groups of failures for a single cluster for the last 7 days.
// A group of failures are failures that would be grouped together in MILO display, i.e.
// same ingested_invocation_id, test_id and variant.
func (c *Client) ReadClusterFailures(ctx context.Context, opts ReadClusterFailuresOptions) (cfs []*ClusterFailure, err error) {
	_, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/analysis/ReadClusterFailures")
	s.Attribute("project", opts.Project)
	defer func() { s.End(err) }()

	dataset, err := bqutil.DatasetForProject(opts.Project)
	if err != nil {
		return nil, errors.Annotate(err, "getting dataset").Err()
	}
	q := c.client.Query(`
		WITH latest_failures_7d AS (
			SELECT
				cluster_algorithm,
				cluster_id,
				test_result_system,
				test_result_id,
				ARRAY_AGG(cf ORDER BY cf.last_updated DESC LIMIT 1)[OFFSET(0)] as r
			FROM clustered_failures cf
			WHERE cf.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
			  AND cluster_algorithm = @clusterAlgorithm
			  AND cluster_id = @clusterID
			  AND realm IN UNNEST(@realms)
			GROUP BY cluster_algorithm, cluster_id, test_result_system, test_result_id
			HAVING r.is_included
		)
		SELECT
			r.realm as Realm,
			r.test_id as TestID,
			ANY_VALUE(r.variant) as Variant,
			ANY_VALUE(r.presubmit_run_id) as PresubmitRunID,
			ANY_VALUE(r.presubmit_run_owner) as PresubmitRunOwner,
			ANY_VALUE(r.presubmit_run_mode) as PresubmitRunMode,
			ANY_VALUE(IF(ARRAY_LENGTH(r.changelists)>0,
				r.changelists[OFFSET(0)], NULL)) as Changelist,
			ANY_VALUE(r.changelists) as Changelists,
			r.partition_time as PartitionTime,
			ANY_VALUE(r.exonerations) as Exonerations,
			ANY_VALUE(r.build_status) as BuildStatus,
			ANY_VALUE(r.build_critical) as IsBuildCritical,
			r.ingested_invocation_id as IngestedInvocationID,
			ANY_VALUE(r.is_ingested_invocation_blocked) as IsIngestedInvocationBlocked,
			count(*) as Count
		FROM latest_failures_7d
		GROUP BY
			r.realm,
			r.ingested_invocation_id,
			r.test_id,
			r.variant_hash,
			r.partition_time
		ORDER BY r.partition_time DESC
		LIMIT 2000
	`)
	q.DefaultDatasetID = dataset
	q.Parameters = []bigquery.QueryParameter{
		{Name: "clusterAlgorithm", Value: opts.ClusterID.Algorithm},
		{Name: "clusterID", Value: opts.ClusterID.ID},
		{Name: "realms", Value: opts.Realms},
	}
	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying cluster failures").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, handleJobReadError(err)
	}
	failures := []*ClusterFailure{}
	for {
		row := &ClusterFailure{}
		err := it.Next(row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next cluster failure row").Err()
		}
		failures = append(failures, row)
	}
	return failures, nil
}
