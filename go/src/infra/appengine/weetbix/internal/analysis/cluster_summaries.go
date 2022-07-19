// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analysis

import (
	"context"
	"math"
	"time"

	"cloud.google.com/go/bigquery"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/trace"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"

	"infra/appengine/weetbix/internal/aip"
	"infra/appengine/weetbix/internal/bqutil"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms/rulesalgorithm"
	configpb "infra/appengine/weetbix/proto/config"
)

// ProjectNotExistsErr is returned if the dataset for the given project
// does not exist.
var ProjectNotExistsErr = errors.New("project does not exist in Weetbix or analysis is not yet available")

// InvalidArgumentTag is used to indicate that one of the query options
// is invalid.
var InvalidArgumentTag = errors.BoolTag{Key: errors.NewTagKey("invalid argument")}

// ImpactfulClusterReadOptions specifies options for ReadImpactfulClusters().
type ImpactfulClusterReadOptions struct {
	// Project is the LUCI Project for which analysis is being performed.
	Project string
	// Thresholds is the set of thresholds, which if any are met
	// or exceeded, should result in the cluster being returned.
	// Thresholds are applied based on the residual pre-Weetbix (exoneration)
	// cluster impact.
	Thresholds *configpb.ImpactThreshold
	// AlwaysIncludeBugClusters controls whether to include analysis for all
	// bug clusters.
	AlwaysIncludeBugClusters bool
}

// Cluster contains detailed information about a cluster, including
// a statistical summary of a cluster's failures, and their impact.
type Cluster struct {
	ClusterID clustering.ClusterID `json:"clusterId"`
	// Distinct user CLs with presubmit rejects.
	PresubmitRejects1d Counts `json:"presubmitRejects1d"`
	PresubmitRejects3d Counts `json:"presubmitRejects3d"`
	PresubmitRejects7d Counts `json:"presubmitRejects7d"`
	// Distinct test runs failed.
	TestRunFails1d Counts `json:"testRunFailures1d"`
	TestRunFails3d Counts `json:"testRunFailures3d"`
	TestRunFails7d Counts `json:"testRunFailures7d"`
	// Total test results with unexpected failures.
	Failures1d Counts `json:"failures1d"`
	Failures3d Counts `json:"failures3d"`
	Failures7d Counts `json:"failures7d"`
	// Test failures exonerated on critical builders, and for an
	// exoneration reason other than NOT_CRITICAL.
	CriticalFailuresExonerated1d Counts `json:"criticalFailuresExonerated1d"`
	CriticalFailuresExonerated3d Counts `json:"criticalFailuresExonerated3d"`
	CriticalFailuresExonerated7d Counts `json:"criticalFailuresExonerated7d"`

	ExampleFailureReason bigquery.NullString `json:"exampleFailureReason"`
	// Top Test IDs included in the cluster, up to 5. Unless the cluster
	// is empty, will always include at least one Test ID.
	TopTestIDs []TopCount `json:"topTestIds"`
	// Top Monorail Components indicates the top monorail components failures
	// in the cluster are associated with by number of failures, up to 5.
	TopMonorailComponents []TopCount `json:"topMonorailComponents"`
}

// ExampleTestID returns an example Test ID that is part of the cluster, or
// "" if the cluster is empty.
func (s *Cluster) ExampleTestID() string {
	if len(s.TopTestIDs) > 0 {
		return s.TopTestIDs[0].Value
	}
	return ""
}

// Counts captures the values of an integer-valued metric in different
// calculation bases.
type Counts struct {
	// The statistic value after impact has been reduced by exoneration.
	Nominal int64 `json:"nominal"`
	// The statistic value:
	// - excluding impact already counted under other higher-priority clusters
	//   (I.E. bug clusters.)
	// - after impact has been reduced by exoneration.
	Residual int64 `json:"residual"`
}

// TopCount captures the result of the APPROX_TOP_COUNT operator. See:
// https://cloud.google.com/bigquery/docs/reference/standard-sql/approximate_aggregate_functions#approx_top_count
type TopCount struct {
	// Value is the value that was frequently occurring.
	Value string `json:"value"`
	// Count is the frequency with which the value occurred.
	Count int64 `json:"count"`
}

// NewClient creates a new client for reading clusters. Close() MUST
// be called after you have finished using this client.
func NewClient(ctx context.Context, gcpProject string) (*Client, error) {
	client, err := bqutil.Client(ctx, gcpProject)
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

// ProjectsWithDataset returns the set of LUCI projects which have
// a BigQuery dataset created.
func (c *Client) ProjectsWithDataset(ctx context.Context) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	di := c.client.Datasets(ctx)
	for {
		d, err := di.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}
		project, err := bqutil.ProjectForDataset(d.DatasetID)
		if err != nil {
			return nil, err
		}
		result[project] = struct{}{}
	}
	return result, nil
}

// RebuildAnalysis re-builds the cluster summaries analysis from
// clustered test results.
func (c *Client) RebuildAnalysis(ctx context.Context, project string) error {
	datasetID, err := bqutil.DatasetForProject(project)
	if err != nil {
		return errors.Annotate(err, "getting dataset").Err()
	}
	dataset := c.client.Dataset(datasetID)

	dstTable := dataset.Table("cluster_summaries")

	q := c.client.Query(clusterSummariesAnalysis)
	q.DefaultDatasetID = dataset.DatasetID
	q.Dst = dstTable
	q.CreateDisposition = bigquery.CreateIfNeeded
	q.WriteDisposition = bigquery.WriteTruncate
	job, err := q.Run(ctx)
	if err != nil {
		return errors.Annotate(err, "starting cluster summary analysis").Err()
	}

	waitCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	js, err := job.Wait(waitCtx)
	if err != nil {
		return errors.Annotate(err, "waiting for cluster summary analysis to complete").Err()
	}
	if js.Err() != nil {
		return errors.Annotate(err, "cluster summary analysis failed").Err()
	}
	return nil
}

// ReadImpactfulClusters reads clusters exceeding specified impact metrics, or are otherwise
// nominated to be read.
func (c *Client) ReadImpactfulClusters(ctx context.Context, opts ImpactfulClusterReadOptions) (cs []*Cluster, err error) {
	_, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/analysis/ReadImpactfulClusters")
	s.Attribute("project", opts.Project)
	defer func() { s.End(err) }()

	if opts.Thresholds == nil {
		return nil, errors.New("thresholds must be specified")
	}

	dataset, err := bqutil.DatasetForProject(opts.Project)
	if err != nil {
		return nil, errors.Annotate(err, "getting dataset").Err()
	}

	whereCriticalFailuresExonerated, cfeParams := whereThresholdsMet("critical_failures_exonerated", opts.Thresholds.CriticalFailuresExonerated)
	whereFailures, failuresParams := whereThresholdsMet("failures", opts.Thresholds.TestResultsFailed)
	whereTestRuns, testRunsParams := whereThresholdsMet("test_run_fails", opts.Thresholds.TestRunsFailed)
	wherePresubmits, presubmitParams := whereThresholdsMet("presubmit_rejects", opts.Thresholds.PresubmitRunsFailed)

	q := c.client.Query(`
		SELECT
			STRUCT(cluster_algorithm AS Algorithm, cluster_id as ID) as ClusterID,` +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "1d") +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "3d") +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "7d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "1d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "3d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "7d") +
		selectCounts("test_run_fails", "TestRunFails", "1d") +
		selectCounts("test_run_fails", "TestRunFails", "3d") +
		selectCounts("test_run_fails", "TestRunFails", "7d") +
		selectCounts("failures", "Failures", "1d") +
		selectCounts("failures", "Failures", "3d") +
		selectCounts("failures", "Failures", "7d") + `
			example_failure_reason.primary_error_message as ExampleFailureReason,
			top_test_ids as TopTestIDs,
			ARRAY(
				SELECT AS STRUCT value, count
				FROM UNNEST(top_monorail_components)
				WHERE value IS NOT NULL
			) as TopMonorailComponents
		FROM cluster_summaries
		WHERE (` + whereCriticalFailuresExonerated + `) OR (` + whereFailures + `)
		    OR (` + whereTestRuns + `) OR (` + wherePresubmits + `)
		    OR (@alwaysIncludeBugClusters AND cluster_algorithm = @ruleAlgorithmName)
		ORDER BY
			presubmit_rejects_residual_1d DESC,
			critical_failures_exonerated_residual_1d DESC,
			test_run_fails_residual_1d DESC,
			failures_residual_1d DESC
	`)
	q.DefaultDatasetID = dataset

	params := []bigquery.QueryParameter{
		{
			Name:  "ruleAlgorithmName",
			Value: rulesalgorithm.AlgorithmName,
		},
		{
			Name:  "alwaysIncludeBugClusters",
			Value: opts.AlwaysIncludeBugClusters,
		},
	}
	params = append(params, cfeParams...)
	params = append(params, failuresParams...)
	params = append(params, testRunsParams...)
	params = append(params, presubmitParams...)
	q.Parameters = params

	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying clusters").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, handleJobReadError(err)
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

func valueOrDefault(value *int64, defaultValue int64) int64 {
	if value != nil {
		return *value
	}
	return defaultValue
}

// selectCounts generates SQL to select a set of Counts.
func selectCounts(sqlPrefix, fieldPrefix, suffix string) string {
	return `STRUCT(` +
		sqlPrefix + `_` + suffix + ` AS Nominal,` +
		sqlPrefix + `_residual_` + suffix + ` AS Residual` +
		`) AS ` + fieldPrefix + suffix + `,`
}

// whereThresholdsMet generates a SQL Where clause to query
// where a particular metric meets a given threshold.
func whereThresholdsMet(sqlPrefix string, threshold *configpb.MetricThreshold) (string, []bigquery.QueryParameter) {
	if threshold == nil {
		threshold = &configpb.MetricThreshold{}
	}
	sql := sqlPrefix + "_residual_1d >= @" + sqlPrefix + "_1d OR " +
		sqlPrefix + "_residual_3d >= @" + sqlPrefix + "_3d OR " +
		sqlPrefix + "_residual_7d >= @" + sqlPrefix + "_7d"
	parameters := []bigquery.QueryParameter{
		{
			Name:  sqlPrefix + "_1d",
			Value: valueOrDefault(threshold.OneDay, math.MaxInt64),
		},
		{
			Name:  sqlPrefix + "_3d",
			Value: valueOrDefault(threshold.ThreeDay, math.MaxInt64),
		},
		{
			Name:  sqlPrefix + "_7d",
			Value: valueOrDefault(threshold.SevenDay, math.MaxInt64),
		},
	}
	return sql, parameters
}

// ReadCluster reads information about a list of clusters.
// If the dataset for the LUCI project does not exist, returns ProjectNotExistsErr.
func (c *Client) ReadClusters(ctx context.Context, luciProject string, clusterIDs []clustering.ClusterID) (cs []*Cluster, err error) {
	_, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/analysis/ReadClusters")
	s.Attribute("project", luciProject)
	defer func() { s.End(err) }()

	dataset, err := bqutil.DatasetForProject(luciProject)
	if err != nil {
		return nil, errors.Annotate(err, "getting dataset").Err()
	}

	q := c.client.Query(`
		SELECT
			STRUCT(cluster_algorithm AS Algorithm, cluster_id as ID) as ClusterID,` +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "1d") +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "3d") +
		selectCounts("critical_failures_exonerated", "CriticalFailuresExonerated", "7d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "1d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "3d") +
		selectCounts("presubmit_rejects", "PresubmitRejects", "7d") +
		selectCounts("test_run_fails", "TestRunFails", "1d") +
		selectCounts("test_run_fails", "TestRunFails", "3d") +
		selectCounts("test_run_fails", "TestRunFails", "7d") +
		selectCounts("failures", "Failures", "1d") +
		selectCounts("failures", "Failures", "3d") +
		selectCounts("failures", "Failures", "7d") + `
			example_failure_reason.primary_error_message as ExampleFailureReason,
			top_test_ids as TopTestIDs
		FROM cluster_summaries
		WHERE STRUCT(cluster_algorithm AS Algorithm, cluster_id as ID) IN UNNEST(@clusterIDs)
	`)
	q.DefaultDatasetID = dataset
	q.Parameters = []bigquery.QueryParameter{
		{Name: "clusterIDs", Value: clusterIDs},
	}
	job, err := q.Run(ctx)
	if err != nil {
		return nil, errors.Annotate(err, "querying cluster").Err()
	}
	it, err := job.Read(ctx)
	if err != nil {
		return nil, handleJobReadError(err)
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
}

// ClusterSummary represents a summary of the cluster's failures
// and their impact.
type ClusterSummary struct {
	ClusterID                  clustering.ClusterID `json:"clusterId"`
	PresubmitRejects           int64                `json:"presubmitRejects"`
	CriticalFailuresExonerated int64                `json:"criticalFailuresExonerated"`
	Failures                   int64                `json:"failures"`
	ExampleFailureReason       bigquery.NullString  `json:"exampleFailureReason"`
	ExampleTestID              string               `json:"exampleTestId"`
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
		SELECT
			STRUCT(cluster_algorithm AS Algorithm,
				cluster_id AS ID) AS ClusterID,
			ANY_VALUE(failure_reason.primary_error_message) AS ExampleFailureReason,
			ANY_VALUE(test_id) AS ExampleTestID,
			APPROX_COUNT_DISTINCT(presubmit_cl_blocked) AS PresubmitRejects,
			APPROX_COUNT_DISTINCT(IF(is_critical_and_exonerated,unique_test_result_id, NULL)) AS CriticalFailuresExonerated,
			APPROX_COUNT_DISTINCT(unique_test_result_id) AS Failures,
		FROM (
			SELECT
				cluster_algorithm,
				cluster_id,
				test_id,
				failure_reason,
				CONCAT(chunk_id, '/', chunk_index) as unique_test_result_id,
				(build_critical AND
				-- Exonerated for a reason other than NOT_CRITICAL or UNEXPECTED_PASS.
				-- Passes are not ingested by Weetbix, but if a test has both an unexpected pass
				-- and an unexpected failure, it will be exonerated for the unexpected pass.
				(STRUCT('OCCURS_ON_MAINLINE' as Reason) in UNNEST(exonerations) OR
					STRUCT('OCCURS_ON_OTHER_CLS' as Reason) in UNNEST(exonerations)))
				AS is_critical_and_exonerated,
				IF(is_ingested_invocation_blocked AND build_critical AND presubmit_run_mode = 'FULL_RUN' AND
				ARRAY_LENGTH(exonerations) = 0 AND build_status = 'FAILURE' AND presubmit_run_owner = 'user',
					IF(ARRAY_LENGTH(changelists)>0 AND presubmit_run_owner='user',
					CONCAT(changelists[OFFSET(0)].host, changelists[OFFSET(0)].change),
					NULL),
					NULL)
				AS presubmit_cl_blocked,
			FROM clustered_failures cf
			WHERE
				is_included_with_high_priority
				AND partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
				AND ` + whereClause + `
		)
		GROUP BY
			cluster_algorithm,
			cluster_id
		` + orderByClause + `
		LIMIT 1000
	`

	q := c.client.Query(sql)
	q.DefaultDatasetID = dataset
	q.Parameters = toBigQueryParameters(parameters)

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

type ClusterFailure struct {
	Realm                       bigquery.NullString    `json:"realm"`
	TestID                      bigquery.NullString    `json:"testId"`
	Variant                     []*Variant             `json:"variant"`
	PresubmitRunID              *PresubmitRunID        `json:"presubmitRunId"`
	PresubmitRunOwner           bigquery.NullString    `json:"presubmitRunOwner"`
	PresubmitRunMode            bigquery.NullString    `json:"presubmitRunMode"`
	Changelist                  *Changelist            `json:"changelist"`
	PartitionTime               bigquery.NullTimestamp `json:"partitionTime"`
	Exonerations                []*Exoneration         `json:"exonerations"`
	BuildStatus                 bigquery.NullString    `json:"buildStatus"`
	IsBuildCritical             bigquery.NullBool      `json:"isBuildCritical"`
	IngestedInvocationID        bigquery.NullString    `json:"ingestedInvocationId"`
	IsIngestedInvocationBlocked bigquery.NullBool      `json:"isIngestedInvocationBlocked"`
	Count                       int32                  `json:"count"`
}

type Exoneration struct {
	// One of OCCURS_ON_OTHER_CLS, OCCURS_ON_MAINLINE, NOT_CRITICAL.
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

// ReadClusterFailures reads the latest 2000 groups of failures for a single cluster for the last 7 days.
// A group of failures are failures that would be grouped together in MILO display, i.e.
// same ingested_invocation_id, test_id and variant.
func (c *Client) ReadClusterFailures(ctx context.Context, luciProject string, clusterID clustering.ClusterID) (cfs []*ClusterFailure, err error) {
	_, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/analysis/ReadClusterFailures")
	s.Attribute("project", luciProject)
	defer func() { s.End(err) }()

	dataset, err := bqutil.DatasetForProject(luciProject)
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
			r.partition_time as PartitionTime,
			ANY_VALUE(r.exonerations) as Exonerations,
			ANY_VALUE(r.build_status) as BuildStatus,
			ANY_VALUE(r.build_critical) as IsBuildCritical,
			r.ingested_invocation_id as IngestedInvocationID,
			ANY_VALUE(r.is_ingested_invocation_blocked) as IsIngestedInvocationBlocked,
			count(*) as Count
		FROM latest_failures_7d
		WHERE cluster_algorithm = @clusterAlgorithm
		  AND cluster_id = @clusterID
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
		{Name: "clusterAlgorithm", Value: clusterID.Algorithm},
		{Name: "clusterID", Value: clusterID.ID},
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

func handleJobReadError(err error) error {
	switch e := err.(type) {
	case *googleapi.Error:
		if e.Code == 404 {
			return ProjectNotExistsErr
		}
	}
	return errors.Annotate(err, "obtain result iterator").Err()
}
