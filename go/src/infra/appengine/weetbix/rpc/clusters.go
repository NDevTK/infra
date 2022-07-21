// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/sync/parallel"
	"go.chromium.org/luci/grpc/appstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/aip"
	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms"
	"infra/appengine/weetbix/internal/clustering/reclustering"
	"infra/appengine/weetbix/internal/clustering/rules/cache"
	"infra/appengine/weetbix/internal/clustering/runs"
	"infra/appengine/weetbix/internal/config/compiledcfg"
	pb "infra/appengine/weetbix/proto/v1"
)

// MaxClusterRequestSize is the maximum number of test results to cluster in
// one call to Cluster(...).
const MaxClusterRequestSize = 1000

// MaxBatchGetClustersRequestSize is the maximum number of clusters to obtain
// impact for in one call to BatchGetClusters().
const MaxBatchGetClustersRequestSize = 1000

type AnalysisClient interface {
	ReadClusters(ctx context.Context, luciProject string, clusterIDs []clustering.ClusterID) ([]*analysis.Cluster, error)
	QueryClusterSummaries(ctx context.Context, luciProject string, options *analysis.QueryClusterSummariesOptions) ([]*analysis.ClusterSummary, error)
}

type clustersServer struct {
	analysisClient AnalysisClient
}

func NewClustersServer(analysisClient AnalysisClient) *pb.DecoratedClusters {
	return &pb.DecoratedClusters{
		Prelude:  checkAllowedPrelude,
		Service:  &clustersServer{analysisClient: analysisClient},
		Postlude: gRPCifyAndLogPostlude,
	}
}

// Cluster clusters a list of test failures. See proto definition for more.
func (*clustersServer) Cluster(ctx context.Context, req *pb.ClusterRequest) (*pb.ClusterResponse, error) {
	if len(req.TestResults) > MaxClusterRequestSize {
		return nil, invalidArgumentError(fmt.Errorf(
			"too many test results: at most %v test results can be clustered in one request", MaxClusterRequestSize))
	}

	failures := make([]*clustering.Failure, 0, len(req.TestResults))
	for i, tr := range req.TestResults {
		if err := validateTestResult(i, tr); err != nil {
			return nil, err
		}
		failures = append(failures, &clustering.Failure{
			TestID: tr.TestId,
			Reason: tr.FailureReason,
		})
	}

	// Fetch a recent project configuration.
	// (May be a recent value that was cached.)
	cfg, err := readProjectConfig(ctx, req.Project)
	if err != nil {
		return nil, err
	}

	// Fetch a recent ruleset.
	ruleset, err := reclustering.Ruleset(ctx, req.Project, cache.StrongRead)
	if err != nil {
		return nil, err
	}

	// Perform clustering from scratch. (Incremental clustering does not make
	// sense for this RPC.)
	existing := algorithms.NewEmptyClusterResults(len(req.TestResults))

	results := algorithms.Cluster(cfg, ruleset, existing, failures)

	// Construct the response proto.
	clusteredTRs := make([]*pb.ClusterResponse_ClusteredTestResult, 0, len(results.Clusters))
	for i, r := range results.Clusters {
		request := req.TestResults[i]

		entries := make([]*pb.ClusterResponse_ClusteredTestResult_ClusterEntry, 0, len(r))
		for _, clusterID := range r {
			entry := &pb.ClusterResponse_ClusteredTestResult_ClusterEntry{
				ClusterId: createClusterIdPB(clusterID),
			}
			if clusterID.IsBugCluster() {
				// For bug clusters, the ID of the cluster is also the ID of
				// the rule that defines it. Use this property to lookup the
				// associated rule.
				ruleID := clusterID.ID
				rule := ruleset.ActiveRulesByID[ruleID]
				entry.Bug = createAssociatedBugPB(rule.Rule.BugID, cfg.Config)
			}
			entries = append(entries, entry)
		}
		clusteredTR := &pb.ClusterResponse_ClusteredTestResult{
			RequestTag: request.RequestTag,
			Clusters:   entries,
		}
		clusteredTRs = append(clusteredTRs, clusteredTR)
	}

	version := &pb.ClusteringVersion{
		AlgorithmsVersion: int32(results.AlgorithmsVersion),
		RulesVersion:      timestamppb.New(results.RulesVersion),
		ConfigVersion:     timestamppb.New(results.ConfigVersion),
	}

	return &pb.ClusterResponse{
		ClusteredTestResults: clusteredTRs,
		ClusteringVersion:    version,
	}, nil
}

func validateTestResult(i int, tr *pb.ClusterRequest_TestResult) error {
	if tr.TestId == "" {
		return invalidArgumentError(fmt.Errorf("test result %v: test ID must not be empty", i))
	}
	return nil
}

func (c *clustersServer) BatchGet(ctx context.Context, req *pb.BatchGetClustersRequest) (*pb.BatchGetClustersResponse, error) {
	project, err := parseProjectName(req.Parent)
	if err != nil {
		return nil, invalidArgumentError(errors.Annotate(err, "parent").Err())
	}

	if len(req.Names) > MaxBatchGetClustersRequestSize {
		return nil, invalidArgumentError(fmt.Errorf(
			"too many names: at most %v clusters can be retrieved in one request", MaxBatchGetClustersRequestSize))
	}
	if len(req.Names) == 0 {
		// Return INVALID_ARGUMENT if no names specified, as per google.aip.dev/231.
		return nil, invalidArgumentError(errors.New("names must be specified"))
	}

	cfg, err := readProjectConfig(ctx, project)
	if err != nil {
		return nil, err
	}

	// The cluster ID requested in each request item.
	clusterIDs := make([]clustering.ClusterID, 0, len(req.Names))

	for i, name := range req.Names {
		clusterProject, clusterID, err := parseClusterName(name)
		if err != nil {
			return nil, invalidArgumentError(errors.Annotate(err, "name %v", i).Err())
		}
		if clusterProject != project {
			return nil, invalidArgumentError(fmt.Errorf("name %v: project must match parent project (%q)", i, project))
		}
		clusterIDs = append(clusterIDs, clusterID)
	}

	clusters, err := c.analysisClient.ReadClusters(ctx, project, clusterIDs)
	if err != nil {
		if err == analysis.ProjectNotExistsErr {
			return nil, appstatus.Error(codes.NotFound,
				"project dataset not provisioned in Weetbix or cluster analysis is not yet available")
		}
		return nil, err
	}

	readClusterByID := make(map[clustering.ClusterID]*analysis.Cluster)
	for _, c := range clusters {
		readClusterByID[c.ClusterID] = c
	}

	// As per google.aip.dev/231, the order of responses must be the
	// same as the names in the request.
	results := make([]*pb.Cluster, 0, len(clusterIDs))
	for i, clusterID := range clusterIDs {
		c, ok := readClusterByID[clusterID]
		if !ok {
			c = &analysis.Cluster{
				ClusterID: clusterID,
				// No impact available for cluster (e.g. because no examples
				// in BigQuery). Use suitable default values (all zeros
				// for impact).
			}
		}

		result := &pb.Cluster{
			Name:       req.Names[i],
			HasExample: ok,
			UserClsFailedPresubmit: &pb.Cluster_MetricValues{
				OneDay:   newCounts(c.PresubmitRejects1d),
				ThreeDay: newCounts(c.PresubmitRejects3d),
				SevenDay: newCounts(c.PresubmitRejects7d),
			},
			CriticalFailuresExonerated: &pb.Cluster_MetricValues{
				OneDay:   newCounts(c.CriticalFailuresExonerated1d),
				ThreeDay: newCounts(c.CriticalFailuresExonerated3d),
				SevenDay: newCounts(c.CriticalFailuresExonerated7d),
			},
			Failures: &pb.Cluster_MetricValues{
				OneDay:   newCounts(c.Failures1d),
				ThreeDay: newCounts(c.Failures3d),
				SevenDay: newCounts(c.Failures7d),
			},
		}

		if !clusterID.IsBugCluster() && ok {
			example := &clustering.Failure{
				TestID: c.ExampleTestID(),
				Reason: &pb.FailureReason{
					PrimaryErrorMessage: c.ExampleFailureReason.StringVal,
				},
			}
			result.Title = suggestedClusterTitle(c.ClusterID, example, cfg)

			// Ignore error, it is only returned if algorithm cannot be found.
			alg, _ := algorithms.SuggestingAlgorithm(clusterID.Algorithm)
			if alg != nil {
				result.EquivalentFailureAssociationRule = alg.FailureAssociationRule(cfg, example)
			}
		}
		results = append(results, result)
	}
	return &pb.BatchGetClustersResponse{
		Clusters: results,
	}, nil
}

func newCounts(counts analysis.Counts) *pb.Cluster_MetricValues_Counts {
	return &pb.Cluster_MetricValues_Counts{Nominal: counts.Nominal}
}

func suggestedClusterTitle(clusterID clustering.ClusterID, exampleFailure *clustering.Failure, cfg *compiledcfg.ProjectConfig) string {
	var title string

	// Ignore error, it is only returned if algorithm cannot be found.
	alg, _ := algorithms.SuggestingAlgorithm(clusterID.Algorithm)
	switch {
	case alg != nil:
		title = alg.ClusterTitle(cfg, exampleFailure)
	case clusterID.IsTestNameCluster():
		// Fallback for old test name clusters.
		title = exampleFailure.TestID
	case clusterID.IsFailureReasonCluster():
		// Fallback for old reason-based clusters.
		title = exampleFailure.Reason.PrimaryErrorMessage
	default:
		// Fallback for all other cases.
		title = fmt.Sprintf("%s/%s", clusterID.Algorithm, clusterID.ID)
	}
	return title
}

func (c *clustersServer) GetReclusteringProgress(ctx context.Context, req *pb.GetReclusteringProgressRequest) (*pb.ReclusteringProgress, error) {
	project, err := parseReclusteringProgressName(req.Name)
	if err != nil {
		return nil, invalidArgumentError(errors.Annotate(err, "name").Err())
	}

	progress, err := runs.ReadReclusteringProgress(ctx, project)
	if err != nil {
		return nil, err
	}

	return &pb.ReclusteringProgress{
		Name:             req.Name,
		ProgressPerMille: int32(progress.ProgressPerMille),
		Last: &pb.ClusteringVersion{
			AlgorithmsVersion: int32(progress.Last.AlgorithmsVersion),
			RulesVersion:      timestamppb.New(progress.Last.RulesVersion),
			ConfigVersion:     timestamppb.New(progress.Last.ConfigVersion),
		},
		Next: &pb.ClusteringVersion{
			AlgorithmsVersion: int32(progress.Next.AlgorithmsVersion),
			RulesVersion:      timestamppb.New(progress.Next.RulesVersion),
			ConfigVersion:     timestamppb.New(progress.Next.ConfigVersion),
		},
	}, nil
}

func (c *clustersServer) QueryClusterSummaries(ctx context.Context, req *pb.QueryClusterSummariesRequest) (*pb.QueryClusterSummariesResponse, error) {
	var cfg *compiledcfg.ProjectConfig
	var ruleset *cache.Ruleset
	var clusters []*analysis.ClusterSummary
	var bqErr error
	// Parallelise call to Biquery (slow call)
	// with the datastore/spanner calls to reduce the critical path.
	err := parallel.FanOutIn(func(ch chan<- func() error) {
		ch <- func() error {
			start := time.Now()
			var err error
			// Fetch a recent project configuration.
			// (May be a recent value that was cached.)
			cfg, err = readProjectConfig(ctx, req.Project)
			if err != nil {
				return err
			}

			// Fetch a recent ruleset.
			ruleset, err = reclustering.Ruleset(ctx, req.Project, cache.StrongRead)
			if err != nil {
				return err
			}
			logging.Infof(ctx, "QueryClusterSummaries: Ruleset part took %v", time.Since(start))
			return nil
		}
		ch <- func() error {
			start := time.Now()
			// To avoid the error returned from the service being non-deterministic
			// if both goroutines error, populate any error encountered here
			// into bqErr and return no error.
			opts := &analysis.QueryClusterSummariesOptions{}
			var err error
			opts.FailureFilter, err = aip.ParseFilter(req.FailureFilter)
			if err != nil {
				bqErr = invalidArgumentError(errors.Annotate(err, "failure_filter").Err())
				return nil
			}
			opts.OrderBy, err = aip.ParseOrderBy(req.OrderBy)
			if err != nil {
				bqErr = invalidArgumentError(errors.Annotate(err, "order_by").Err())
				return nil
			}

			clusters, err = c.analysisClient.QueryClusterSummaries(ctx, req.Project, opts)
			if err != nil {
				if err == analysis.ProjectNotExistsErr {
					bqErr = appstatus.Error(codes.NotFound,
						"project dataset not provisioned in Weetbix or cluster analysis is not yet available")
					return nil
				}
				if analysis.InvalidArgumentTag.In(err) {
					bqErr = invalidArgumentError(err)
					return nil
				}
				bqErr = errors.Annotate(err, "query clusters for failures").Err()
				return nil
			}
			logging.Infof(ctx, "QueryClusterSummaries: BigQuery part took %v", time.Since(start))
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	// To avoid the error returned from the service being non-deterministic
	// if both goroutines error, return error from bigQuery part after any other errors.
	if bqErr != nil {
		return nil, bqErr
	}

	result := []*pb.ClusterSummary{}
	for _, c := range clusters {
		cs := &pb.ClusterSummary{
			ClusterId:                  createClusterIdPB(c.ClusterID),
			PresubmitRejects:           c.PresubmitRejects,
			CriticalFailuresExonerated: c.CriticalFailuresExonerated,
			Failures:                   c.Failures,
		}
		if c.ClusterID.IsBugCluster() {
			ruleID := c.ClusterID.ID
			rule := ruleset.ActiveRulesByID[ruleID]
			if rule != nil {
				cs.Title = rule.Rule.RuleDefinition
				cs.Bug = createAssociatedBugPB(rule.Rule.BugID, cfg.Config)
			} else {
				// Rule is inactive / in process of being archived.
				cs.Title = "(rule archived)"
			}
		} else {
			example := &clustering.Failure{
				TestID: c.ExampleTestID,
				Reason: &pb.FailureReason{
					PrimaryErrorMessage: c.ExampleFailureReason.StringVal,
				},
			}
			cs.Title = suggestedClusterTitle(c.ClusterID, example, cfg)
		}

		result = append(result, cs)
	}
	return &pb.QueryClusterSummariesResponse{ClusterSummaries: result}, nil
}
