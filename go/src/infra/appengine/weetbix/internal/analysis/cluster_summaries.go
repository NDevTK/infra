// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analysis

import (
	"context"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/trace"
	"google.golang.org/api/iterator"

	"infra/appengine/weetbix/internal/aip"
	"infra/appengine/weetbix/internal/bqutil"
	"infra/appengine/weetbix/internal/clustering"
)

var ClusteredFailuresTable = aip.NewTable().WithColumns(
	aip.NewColumn().WithName("test_id").WithDatabaseName("test_id").FilterableImplicitly().Build(),
	aip.NewColumn().WithName("failure_reason").WithDatabaseName("failure_reason.primary_error_message").FilterableImplicitly().Build(),
	aip.NewColumn().WithName("realm").WithDatabaseName("realm").Filterable().Build(),
	aip.NewColumn().WithName("ingested_invocation_id").WithDatabaseName("ingested_invocation_id").Filterable().Build(),
	aip.NewColumn().WithName("cluster_algorithm").WithDatabaseName("cluster_algorithm").Filterable().Build(),
	aip.NewColumn().WithName("cluster_id").WithDatabaseName("cluster_id").Filterable().Build(),
	aip.NewColumn().WithName("variant_hash").WithDatabaseName("variant_hash").Filterable().Build(),
	aip.NewColumn().WithName("test_run_id").WithDatabaseName("test_run_id").Filterable().Build(),
).Build()

var ClusterSummariesTable = aip.NewTable().WithColumns(
	aip.NewColumn().WithName("presubmit_rejects").WithDatabaseName("PresubmitRejects").Sortable().Build(),
	aip.NewColumn().WithName("critical_failures_exonerated").WithDatabaseName("CriticalFailuresExonerated").Sortable().Build(),
	aip.NewColumn().WithName("failures").WithDatabaseName("Failures").Sortable().Build(),
).Build()

var ClusterSummariesDefaultOrder = []aip.OrderBy{
	{Name: "presubmit_rejects", Descending: true},
	{Name: "critical_failures_exonerated", Descending: true},
	{Name: "failures", Descending: true},
}

type QueryClusterSummariesOptions struct {
	// A filter on the underlying failures to include in the clusters.
	FailureFilter *aip.Filter
	OrderBy       []aip.OrderBy
	Realms        []string
}

// ClusterSummary represents a summary of the cluster's failures
// and their impact.
type ClusterSummary struct {
	ClusterID                       clustering.ClusterID
	ExampleFailureReason            bigquery.NullString
	ExampleTestID                   string
	PresubmitRejects                int64
	PresubmitRejectsByDay           []int64
	CriticalFailuresExonerated      int64
	CriticalFailuresExoneratedByDay []int64
	Failures                        int64
	FailuresByDay                   []int64
	UniqueTestIDs                   int64
	UniqueTestIDsByDay              []int64
}

// Queries a summary of clusters in the project.
// The subset of failures included in the clustering may be filtered.
// If the dataset for the LUCI project does not exist, returns
// ProjectNotExistsErr.
// If options.FailuresFilter or options.OrderBy is invalid with respect to the
// query schema, returns an error tagged with InvalidArgumentTag so that the
// appropriate gRPC error can be returned to the client (if applicable).
func (c *Client) QueryClusterSummaries(ctx context.Context, luciProject string, options *QueryClusterSummariesOptions) (cs []*ClusterSummary, err error) {
	_, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/analysis/QueryClusterSummaries")
	s.Attribute("project", luciProject)
	defer func() { s.End(err) }()

	// Note that the content of the filter and order_by clause is untrusted
	// user input and is validated as part of the Where/OrderBy clause
	// generation here.
	const parameterPrefix = "w_"
	whereClause, parameters, err := ClusteredFailuresTable.WhereClause(options.FailureFilter, parameterPrefix)
	if err != nil {
		return nil, errors.Annotate(err, "failure_filter").Tag(InvalidArgumentTag).Err()
	}

	order := aip.MergeWithDefaultOrder(ClusterSummariesDefaultOrder, options.OrderBy)
	orderByClause, err := ClusterSummariesTable.OrderByClause(order)
	if err != nil {
		return nil, errors.Annotate(err, "order_by").Tag(InvalidArgumentTag).Err()
	}

	dataset, err := bqutil.DatasetForProject(luciProject)
	if err != nil {
		return nil, errors.Annotate(err, "getting dataset").Err()
	}
	// The following query does not take into account removals of test failures
	// from clusters as this dramatically slows down the query. Instead, we
	// rely upon a periodic job to purge these results from the table.
	// We avoid double-counting the test failures (e.g. in case of addition
	// deletion, re-addition) by using APPROX_COUNT_DISTINCT to count the
	// number of distinct failures in the cluster.
	sql := `
	WITH
  failures AS (
  SELECT
    cluster_algorithm,
    cluster_id,
    test_id,
    failure_reason,
    CONCAT(chunk_id, '/', chunk_index) AS unique_test_result_id,
    (build_critical AND
      -- Exonerated for a reason other than NOT_CRITICAL or UNEXPECTED_PASS.
      -- Passes are not ingested by Weetbix, but if a test has both an unexpected pass
      -- and an unexpected failure, it will be exonerated for the unexpected pass.
      (STRUCT('OCCURS_ON_MAINLINE' AS Reason) IN UNNEST(exonerations)
        OR STRUCT('OCCURS_ON_OTHER_CLS' AS Reason) IN UNNEST(exonerations))) AS is_critical_and_exonerated,
  IF
    (is_ingested_invocation_blocked
      AND build_critical
      AND presubmit_run_mode = 'FULL_RUN'
      AND ARRAY_LENGTH(exonerations) = 0
      AND build_status = 'FAILURE'
      AND presubmit_run_owner = 'user',
    IF
      (ARRAY_LENGTH(changelists)>0
        AND presubmit_run_owner='user',
        CONCAT(changelists[
        OFFSET
          (0)].host, changelists[
        OFFSET
          (0)].change),
        NULL),
      NULL) AS presubmit_cl_blocked,
    DATE(partition_time) AS day,
  FROM
    clustered_failures cf
  WHERE
    is_included_with_high_priority
    AND partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
    AND ` + whereClause + ` 
    ),
  clusters AS (
  SELECT
    cluster_algorithm,
    cluster_id,
    APPROX_COUNT_DISTINCT(test_id) AS UniqueTestIDs,
    ANY_VALUE(failure_reason.primary_error_message) AS ExampleFailureReason,
    ANY_VALUE(test_id) AS ExampleTestID,
    APPROX_COUNT_DISTINCT(presubmit_cl_blocked) AS PresubmitRejects,
    APPROX_COUNT_DISTINCT(
    IF
      (is_critical_and_exonerated,
        unique_test_result_id,
        NULL)) AS CriticalFailuresExonerated,
    APPROX_COUNT_DISTINCT(unique_test_result_id) AS Failures,
  FROM
    failures
  GROUP BY
    cluster_algorithm,
    cluster_id),
  daily_clusters AS (
  SELECT
    cluster_algorithm,
    cluster_id,
    day,
    APPROX_COUNT_DISTINCT(test_id) AS UniqueTestIDs,
    APPROX_COUNT_DISTINCT(presubmit_cl_blocked) AS PresubmitRejects,
    APPROX_COUNT_DISTINCT(
    IF
      (is_critical_and_exonerated,
        unique_test_result_id,
        NULL)) AS CriticalFailuresExonerated,
    APPROX_COUNT_DISTINCT(unique_test_result_id) AS Failures,
  FROM
    failures
  GROUP BY
    cluster_algorithm,
    cluster_id,
    day),
  cluster_days AS (
  SELECT
    cluster_algorithm,
    cluster_id,
    day,
  FROM
    clusters
  CROSS JOIN
    UNNEST(GENERATE_DATE_ARRAY(DATE_SUB(CURRENT_DATE(), INTERVAL 7 DAY), CURRENT_DATE())) AS day ),
  daily_clusters_no_gaps AS (
  SELECT
    d.cluster_algorithm,
    d.cluster_id,
    d.day,
    COALESCE(c.UniqueTestIDs,
      0) AS UniqueTestIDs,
    COALESCE(c.PresubmitRejects,
      0) AS PresubmitRejects,
    COALESCE(c.CriticalFailuresExonerated,
      0) AS CriticalFailuresExonerated,
    COALESCE(c.Failures,
      0) AS Failures,
  FROM
    cluster_days d
  LEFT OUTER JOIN
    daily_clusters c
  ON
    d.cluster_algorithm = c.cluster_algorithm
    AND d.cluster_id = c.cluster_id
    AND d.day = c.day ),
daily_clusters_agg AS (SELECT
  cluster_algorithm,
  cluster_id,
  ARRAY_AGG(UniqueTestIDs
  ORDER BY
    day DESC) AS UniqueTestIDsByDay,
  ARRAY_AGG(PresubmitRejects
  ORDER BY
    day DESC) AS PresubmitRejectsByDay,
  ARRAY_AGG(CriticalFailuresExonerated
  ORDER BY
    day DESC) AS CriticalFailuresExoneratedByDay,
  ARRAY_AGG(Failures
  ORDER BY
    day DESC) AS FailuresByDay,
FROM
  daily_clusters_no_gaps
GROUP BY
  cluster_algorithm,
  cluster_id)
SELECT
STRUCT(c.cluster_algorithm AS Algorithm,
    c.cluster_id AS ID) AS ClusterID,
    ExampleFailureReason,
    ExampleTestID,
    UniqueTestIDs,
    UniqueTestIDsByDay,
    PresubmitRejects,
    PresubmitRejectsByDay,
    CriticalFailuresExonerated,
    CriticalFailuresExoneratedByDay,
    Failures,
    FailuresByDay
FROM
  clusters c JOIN daily_clusters_agg d ON c.cluster_algorithm = d.cluster_algorithm AND c.cluster_id = d.cluster_id
  ` + orderByClause + `
LIMIT
  1000
	`

	q := c.client.Query(sql)
	q.DefaultDatasetID = dataset
	q.Parameters = toBigQueryParameters(parameters)
	q.Parameters = append(q.Parameters, bigquery.QueryParameter{
		Name:  "realms",
		Value: options.Realms,
	})

	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying cluster summaries").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, handleJobReadError(err)
	}
	clusters := []*ClusterSummary{}
	for {
		row := &ClusterSummary{}
		err := it.Next(row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Annotate(err, "obtain next cluster summary row").Err()
		}
		clusters = append(clusters, row)
	}
	return clusters, nil
}

func toBigQueryParameters(pars []aip.QueryParameter) []bigquery.QueryParameter {
	result := make([]bigquery.QueryParameter, 0, len(pars))
	for _, p := range pars {
		result = append(result, bigquery.QueryParameter{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return result
}
