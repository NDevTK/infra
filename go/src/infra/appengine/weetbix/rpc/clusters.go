// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/appstatus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms"
	"infra/appengine/weetbix/internal/clustering/reclustering"
	"infra/appengine/weetbix/internal/clustering/rules/cache"
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
	ReadClusters(ctx context.Context, luciProject string, clusterIDs []clustering.ClusterID) ([]*analysis.ClusterSummary, error)
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
		AlgorithmsVersion: results.AlgorithmsVersion,
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

	readClusterByID := make(map[clustering.ClusterID]*analysis.ClusterSummary)
	for _, c := range clusters {
		readClusterByID[c.ClusterID] = c
	}

	// As per google.aip.dev/231, the order of responses must be the
	// same as the names in the request.
	results := make([]*pb.Cluster, 0, len(clusterIDs))
	for i, clusterID := range clusterIDs {
		cs, ok := readClusterByID[clusterID]
		if !ok {
			cs = &analysis.ClusterSummary{
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
				OneDay:   newCounts(cs.PresubmitRejects1d),
				ThreeDay: newCounts(cs.PresubmitRejects3d),
				SevenDay: newCounts(cs.PresubmitRejects7d),
			},
			CriticalFailuresExonerated: &pb.Cluster_MetricValues{
				OneDay:   newCounts(cs.CriticalFailuresExonerated1d),
				ThreeDay: newCounts(cs.CriticalFailuresExonerated3d),
				SevenDay: newCounts(cs.CriticalFailuresExonerated7d),
			},
			Failures: &pb.Cluster_MetricValues{
				OneDay:   newCounts(cs.Failures1d),
				ThreeDay: newCounts(cs.Failures3d),
				SevenDay: newCounts(cs.Failures7d),
			},
		}

		if !clusterID.IsBugCluster() && ok {
			result.Title = suggestedClusterTitle(cs, cfg)

			// Ignore error, it is only returned if algorithm cannot be found.
			alg, _ := algorithms.SuggestingAlgorithm(clusterID.Algorithm)
			if alg != nil {
				example := &clustering.Failure{
					TestID: cs.ExampleTestID(),
					Reason: &pb.FailureReason{
						PrimaryErrorMessage: cs.ExampleFailureReason.StringVal,
					},
				}
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

func suggestedClusterTitle(cs *analysis.ClusterSummary, cfg *compiledcfg.ProjectConfig) string {
	var title string

	// Ignore error, it is only returned if algorithm cannot be found.
	alg, _ := algorithms.SuggestingAlgorithm(cs.ClusterID.Algorithm)
	switch {
	case alg != nil:
		example := &clustering.Failure{
			TestID: cs.ExampleTestID(),
			Reason: &pb.FailureReason{
				PrimaryErrorMessage: cs.ExampleFailureReason.StringVal,
			},
		}
		title = alg.ClusterTitle(cfg, example)
	case cs.ClusterID.IsTestNameCluster():
		// Fallback for old test name clusters.
		title = cs.ExampleTestID()
	case cs.ClusterID.IsFailureReasonCluster():
		// Fallback for old reason-based clusters.
		title = cs.ExampleFailureReason.StringVal
	default:
		// Fallback for all other cases.
		title = fmt.Sprintf("%s/%s", cs.ClusterID.Algorithm, cs.ClusterID.ID)
	}
	return title
}
