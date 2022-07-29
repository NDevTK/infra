// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"
	"sort"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/auth/realms"
	"go.chromium.org/luci/server/caching"
	"go.chromium.org/luci/server/secrets"
	"go.chromium.org/luci/server/secrets/testsecrets"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/bugs"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms"
	"infra/appengine/weetbix/internal/clustering/algorithms/failurereason"
	"infra/appengine/weetbix/internal/clustering/algorithms/rulesalgorithm"
	"infra/appengine/weetbix/internal/clustering/algorithms/testname"
	"infra/appengine/weetbix/internal/clustering/rules"
	"infra/appengine/weetbix/internal/clustering/runs"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/config/compiledcfg"
	"infra/appengine/weetbix/internal/perms"
	"infra/appengine/weetbix/internal/testutil"
	configpb "infra/appengine/weetbix/proto/config"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestClusters(t *testing.T) {
	Convey("With a clusters server", t, func() {
		ctx := testutil.SpannerTestContext(t)
		ctx = caching.WithEmptyProcessCache(ctx)

		// For user identification.
		ctx = authtest.MockAuthConfig(ctx)
		authState := &authtest.FakeState{
			Identity:       "user:someone@example.com",
			IdentityGroups: []string{"weetbix-access"},
		}
		ctx = auth.WithState(ctx, authState)
		ctx = secrets.Use(ctx, &testsecrets.Store{})

		// Provides datastore implementation needed for project config.
		ctx = memory.Use(ctx)
		analysisClient := newFakeAnalysisClient()
		server := NewClustersServer(analysisClient)

		configVersion := time.Date(2025, time.August, 12, 0, 1, 2, 3, time.UTC)
		projectCfg := config.CreatePlaceholderProjectConfig()
		projectCfg.LastUpdated = timestamppb.New(configVersion)
		projectCfg.Monorail.DisplayPrefix = "crbug.com"
		projectCfg.Monorail.MonorailHostname = "bugs.chromium.org"

		configs := make(map[string]*configpb.ProjectConfig)
		configs["testproject"] = projectCfg
		err := config.SetTestProjectConfig(ctx, configs)
		So(err, ShouldBeNil)

		compiledTestProjectCfg, err := compiledcfg.NewConfig(projectCfg)
		So(err, ShouldBeNil)

		// Rules version is in microsecond granularity, consistent with
		// the granularity of Spanner commit timestamps.
		rulesVersion := time.Date(2021, time.February, 12, 1, 2, 4, 5000, time.UTC)
		rs := []*rules.FailureAssociationRule{
			rules.NewRule(0).
				WithProject("testproject").
				WithRuleDefinition(`test LIKE "%TestSuite.TestName%"`).
				WithPredicateLastUpdated(rulesVersion.Add(-1 * time.Hour)).
				WithBug(bugs.BugID{
					System: "monorail",
					ID:     "chromium/7654321",
				}).Build(),
			rules.NewRule(1).
				WithProject("testproject").
				WithRuleDefinition(`reason LIKE "my_file.cc(%): Check failed: false."`).
				WithPredicateLastUpdated(rulesVersion).
				WithBug(bugs.BugID{
					System: "buganizer",
					ID:     "82828282",
				}).Build(),
			rules.NewRule(2).
				WithProject("testproject").
				WithRuleDefinition(`test LIKE "%Other%"`).
				WithPredicateLastUpdated(rulesVersion.Add(-2 * time.Hour)).
				WithBug(bugs.BugID{
					System: "monorail",
					ID:     "chromium/912345",
				}).Build(),
		}
		err = rules.SetRulesForTesting(ctx, rs)
		So(err, ShouldBeNil)

		Convey("Unauthorised requests are rejected", func() {
			// Ensure no access to weetbix-access.
			ctx = auth.WithState(ctx, &authtest.FakeState{
				Identity: "user:someone@example.com",
				// Not a member of weetbix-access.
				IdentityGroups: []string{"other-group"},
			})

			// Make some request (the request should not matter, as
			// a common decorator is used for all requests.)
			request := &pb.ClusterRequest{
				Project: "testproject",
			}

			rule, err := server.Cluster(ctx, request)
			So(err, ShouldBeRPCPermissionDenied, "not a member of weetbix-access")
			So(rule, ShouldBeNil)
		})
		Convey("Cluster", func() {
			authState.IdentityPermissions = []authtest.RealmPermission{
				{
					Realm:      "testproject:@root",
					Permission: perms.PermGetClustersByFailure,
				},
				{
					Realm:      "testproject:@root",
					Permission: perms.PermGetRule,
				},
			}

			request := &pb.ClusterRequest{
				Project: "testproject",
				TestResults: []*pb.ClusterRequest_TestResult{
					{
						RequestTag: "my tag 1",
						TestId:     "ninja://chrome/test:interactive_ui_tests/TestSuite.TestName",
						FailureReason: &pb.FailureReason{
							PrimaryErrorMessage: "my_file.cc(123): Check failed: false.",
						},
					},
					{
						RequestTag: "my tag 2",
						TestId:     "Other_test",
					},
				},
			}

			Convey("Not authorised to cluster", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetClustersByFailure)

				response, err := server.Cluster(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.getByFailure")
				So(response, ShouldBeNil)
			})
			Convey("Not authorised to get rule", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetRule)

				response, err := server.Cluster(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.rules.get")
				So(response, ShouldBeNil)
			})
			Convey("With a valid request", func() {
				// Run
				response, err := server.Cluster(ctx, request)

				// Verify
				So(err, ShouldBeNil)
				So(response, ShouldResembleProto, &pb.ClusterResponse{
					ClusteredTestResults: []*pb.ClusterResponse_ClusteredTestResult{
						{
							RequestTag: "my tag 1",
							Clusters: sortClusterEntries([]*pb.ClusterResponse_ClusteredTestResult_ClusterEntry{
								{
									ClusterId: &pb.ClusterId{
										Algorithm: "rules",
										Id:        rs[0].RuleID,
									},
									Bug: &pb.AssociatedBug{
										System:   "monorail",
										Id:       "chromium/7654321",
										LinkText: "crbug.com/7654321",
										Url:      "https://bugs.chromium.org/p/chromium/issues/detail?id=7654321",
									},
								}, {
									ClusterId: &pb.ClusterId{
										Algorithm: "rules",
										Id:        rs[1].RuleID,
									},
									Bug: &pb.AssociatedBug{
										System:   "buganizer",
										Id:       "82828282",
										LinkText: "b/82828282",
										Url:      "https://issuetracker.google.com/issues/82828282",
									},
								},
								failureReasonClusterEntry(compiledTestProjectCfg, "my_file.cc(123): Check failed: false."),
								testNameClusterEntry(compiledTestProjectCfg, "ninja://chrome/test:interactive_ui_tests/TestSuite.TestName"),
							}),
						},
						{
							RequestTag: "my tag 2",
							Clusters: sortClusterEntries([]*pb.ClusterResponse_ClusteredTestResult_ClusterEntry{
								{
									ClusterId: &pb.ClusterId{
										Algorithm: "rules",
										Id:        rs[2].RuleID,
									},
									Bug: &pb.AssociatedBug{
										System:   "monorail",
										Id:       "chromium/912345",
										LinkText: "crbug.com/912345",
										Url:      "https://bugs.chromium.org/p/chromium/issues/detail?id=912345",
									},
								},
								testNameClusterEntry(compiledTestProjectCfg, "Other_test"),
							}),
						},
					},
					ClusteringVersion: &pb.ClusteringVersion{
						AlgorithmsVersion: algorithms.AlgorithmsVersion,
						RulesVersion:      timestamppb.New(rulesVersion),
						ConfigVersion:     timestamppb.New(configVersion),
					},
				})
			})
			Convey("With missing test ID", func() {
				request.TestResults[1].TestId = ""

				// Run
				response, err := server.Cluster(ctx, request)

				// Verify
				So(response, ShouldBeNil)
				So(err, ShouldBeRPCInvalidArgument, "test result 1: test ID must not be empty")
			})
			Convey("With too many test results", func() {
				var testResults []*pb.ClusterRequest_TestResult
				for i := 0; i < 1001; i++ {
					testResults = append(testResults, &pb.ClusterRequest_TestResult{
						TestId: "AnotherTest",
					})
				}
				request.TestResults = testResults

				// Run
				response, err := server.Cluster(ctx, request)

				// Verify
				So(response, ShouldBeNil)
				So(err, ShouldBeRPCInvalidArgument, "too many test results: at most 1000 test results can be clustered in one request")
			})
			Convey("With project not configured", func() {
				err := config.SetTestProjectConfig(ctx, map[string]*configpb.ProjectConfig{})
				So(err, ShouldBeNil)

				// Run
				response, err := server.Cluster(ctx, request)

				// Verify
				So(response, ShouldBeNil)
				So(err, ShouldBeRPCFailedPrecondition, "project does not exist in Weetbix")
			})
		})
		Convey("BatchGet", func() {
			authState.IdentityPermissions = []authtest.RealmPermission{
				{
					Realm:      "testproject:@root",
					Permission: perms.PermGetCluster,
				},
				{
					Realm:      "testproject:@root",
					Permission: perms.PermExpensiveClusterQueries,
				},
				{
					Realm:      "testproject:realm1",
					Permission: rdbperms.PermListTestResults,
				},
				{
					Realm:      "testproject:realm3",
					Permission: rdbperms.PermListTestResults,
				},
			}

			example := &clustering.Failure{
				TestID: "TestID_Example",
				Reason: &pb.FailureReason{
					PrimaryErrorMessage: "Example failure reason 123.",
				},
			}
			a := &failurereason.Algorithm{}
			reasonClusterID := a.Cluster(compiledTestProjectCfg, example)

			analysisClient.clustersByProject["testproject"] = []*analysis.Cluster{
				{
					ClusterID: clustering.ClusterID{
						Algorithm: rulesalgorithm.AlgorithmName,
						ID:        "11111100000000000000000000000000",
					},
					PresubmitRejects1d:           analysis.Counts{Nominal: 1},
					PresubmitRejects3d:           analysis.Counts{Nominal: 2},
					PresubmitRejects7d:           analysis.Counts{Nominal: 3},
					CriticalFailuresExonerated1d: analysis.Counts{Nominal: 4},
					CriticalFailuresExonerated3d: analysis.Counts{Nominal: 5},
					CriticalFailuresExonerated7d: analysis.Counts{Nominal: 6},
					Failures1d:                   analysis.Counts{Nominal: 7},
					Failures3d:                   analysis.Counts{Nominal: 8},
					Failures7d:                   analysis.Counts{Nominal: 9},
					ExampleFailureReason:         bigquery.NullString{Valid: true, StringVal: "Example failure reason."},
					TopTestIDs: []analysis.TopCount{
						{Value: "TestID 1", Count: 2},
						{Value: "TestID 2", Count: 1},
					},
					Realms: []string{"testproject:realm1", "testproject:realm2"},
				},
				{
					ClusterID: clustering.ClusterID{
						Algorithm: testname.AlgorithmName,
						ID:        "cccccc00000000000000000000000001",
					},
					PresubmitRejects7d:   analysis.Counts{Nominal: 11},
					ExampleFailureReason: bigquery.NullString{Valid: true, StringVal: "Example failure reason 2."},
					TopTestIDs: []analysis.TopCount{
						{Value: "TestID 3", Count: 2},
					},
					Realms: []string{"testproject:realm2", "testproject:realm3"},
				},
				{
					ClusterID: clustering.ClusterID{
						Algorithm: failurereason.AlgorithmName,
						ID:        hex.EncodeToString(reasonClusterID),
					},
					PresubmitRejects7d:   analysis.Counts{Nominal: 15},
					ExampleFailureReason: bigquery.NullString{Valid: true, StringVal: "Example failure reason 123."},
					TopTestIDs: []analysis.TopCount{
						{Value: "TestID_Example", Count: 10},
					},
					Realms: []string{"testproject:realm1", "testproject:realm3"},
				},
			}

			request := &pb.BatchGetClustersRequest{
				Parent: "projects/testproject",
				Names: []string{
					// Rule for which data exists.
					"projects/testproject/clusters/rules/11111100000000000000000000000000",

					// Rule for which no data exists.
					"projects/testproject/clusters/rules/1111110000000000000000000000ffff",

					// Suggested cluster for which cluster ID matches the example
					// provided for the cluster.
					"projects/testproject/clusters/" + failurereason.AlgorithmName + "/" + hex.EncodeToString(reasonClusterID),

					// Suggested cluster for which data exists, but cluster ID mismatches
					// the example provided for the cluster. This could be because
					// configuration has changed and re-clustering is not yet complete.
					"projects/testproject/clusters/" + testname.AlgorithmName + "/cccccc00000000000000000000000001",

					// Suggested cluster for which no impact data exists.
					"projects/testproject/clusters/reason-v3/cccccc0000000000000000000000ffff",
				},
			}

			expectedResponse := &pb.BatchGetClustersResponse{
				Clusters: []*pb.Cluster{
					{
						Name:       "projects/testproject/clusters/rules/11111100000000000000000000000000",
						HasExample: true,
						UserClsFailedPresubmit: &pb.Cluster_MetricValues{
							OneDay:   &pb.Cluster_MetricValues_Counts{Nominal: 1},
							ThreeDay: &pb.Cluster_MetricValues_Counts{Nominal: 2},
							SevenDay: &pb.Cluster_MetricValues_Counts{Nominal: 3},
						},
						CriticalFailuresExonerated: &pb.Cluster_MetricValues{
							OneDay:   &pb.Cluster_MetricValues_Counts{Nominal: 4},
							ThreeDay: &pb.Cluster_MetricValues_Counts{Nominal: 5},
							SevenDay: &pb.Cluster_MetricValues_Counts{Nominal: 6},
						},
						Failures: &pb.Cluster_MetricValues{
							OneDay:   &pb.Cluster_MetricValues_Counts{Nominal: 7},
							ThreeDay: &pb.Cluster_MetricValues_Counts{Nominal: 8},
							SevenDay: &pb.Cluster_MetricValues_Counts{Nominal: 9},
						},
					},
					{
						Name:                       "projects/testproject/clusters/rules/1111110000000000000000000000ffff",
						HasExample:                 false,
						UserClsFailedPresubmit:     emptyMetricValues(),
						CriticalFailuresExonerated: emptyMetricValues(),
						Failures:                   emptyMetricValues(),
					},
					{
						Name:       "projects/testproject/clusters/" + failurereason.AlgorithmName + "/" + hex.EncodeToString(reasonClusterID),
						Title:      "Example failure reason %.",
						HasExample: true,
						UserClsFailedPresubmit: &pb.Cluster_MetricValues{
							OneDay:   &pb.Cluster_MetricValues_Counts{},
							ThreeDay: &pb.Cluster_MetricValues_Counts{},
							SevenDay: &pb.Cluster_MetricValues_Counts{Nominal: 15},
						},
						CriticalFailuresExonerated:       emptyMetricValues(),
						Failures:                         emptyMetricValues(),
						EquivalentFailureAssociationRule: `reason LIKE "Example failure reason %."`,
					},
					{
						Name:       "projects/testproject/clusters/" + testname.AlgorithmName + "/cccccc00000000000000000000000001",
						Title:      "(definition unavailable due to ongoing reclustering)",
						HasExample: true,
						UserClsFailedPresubmit: &pb.Cluster_MetricValues{
							OneDay:   &pb.Cluster_MetricValues_Counts{},
							ThreeDay: &pb.Cluster_MetricValues_Counts{},
							SevenDay: &pb.Cluster_MetricValues_Counts{Nominal: 11},
						},
						CriticalFailuresExonerated:       emptyMetricValues(),
						Failures:                         emptyMetricValues(),
						EquivalentFailureAssociationRule: ``,
					},
					{
						Name:                       "projects/testproject/clusters/reason-v3/cccccc0000000000000000000000ffff",
						HasExample:                 false,
						UserClsFailedPresubmit:     emptyMetricValues(),
						CriticalFailuresExonerated: emptyMetricValues(),
						Failures:                   emptyMetricValues(),
					},
				},
			}

			Convey("Not authorised to get cluster", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetCluster)

				response, err := server.BatchGet(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.get")
				So(response, ShouldBeNil)
			})
			Convey("Not authorised to perform expensive queries", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermExpensiveClusterQueries)

				response, err := server.BatchGet(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.expensiveQueries")
				So(response, ShouldBeNil)
			})
			Convey("With a valid request", func() {
				Convey("No duplicate requests", func() {
					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
				Convey("No test result list permission", func() {
					authState.IdentityPermissions = removePermission(authState.IdentityPermissions, rdbperms.PermListTestResults)

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					for _, r := range expectedResponse.Clusters {
						r.Title = ""
						r.EquivalentFailureAssociationRule = ""
					}
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
				Convey("Duplicate requests", func() {
					// Even if request items are duplicated, the request
					// should still succeed and return correct results.
					request.Names = append(request.Names, request.Names...)
					expectedResponse.Clusters = append(expectedResponse.Clusters, expectedResponse.Clusters...)

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
			})
			Convey("With invalid request", func() {
				Convey("Invalid parent", func() {
					request.Parent = "blah"

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "parent: invalid project name, expected format: projects/{project}")
				})
				Convey("No names specified", func() {
					request.Names = []string{}

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "names must be specified")
				})
				Convey("Parent does not match request items", func() {
					// Request asks for project "blah" but parent asks for
					// project "testproject".
					So(request.Parent, ShouldEqual, "projects/testproject")
					request.Names[1] = "projects/blah/clusters/reason-v3/cccccc00000000000000000000000001"

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, `name 1: project must match parent project ("testproject")`)
				})
				Convey("Invalid name", func() {
					request.Names[1] = "invalid"

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "name 1: invalid cluster name, expected format: projects/{project}/clusters/{cluster_alg}/{cluster_id}")
				})
				Convey("Invalid cluster algorithm in name", func() {
					request.Names[1] = "projects/blah/clusters/reason/cccccc00000000000000000000000001"

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "name 1: invalid cluster name: algorithm not valid")
				})
				Convey("Invalid cluster ID in name", func() {
					request.Names[1] = "projects/blah/clusters/reason-v3/123"

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "name 1: invalid cluster name: ID is not valid lowercase hexadecimal bytes")
				})
				Convey("Too many request items", func() {
					var names []string
					for i := 0; i < 1001; i++ {
						names = append(names, "projects/testproject/clusters/rules/11111100000000000000000000000000")
					}
					request.Names = names

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "too many names: at most 1000 clusters can be retrieved in one request")
				})
				Convey("Dataset does not exist", func() {
					delete(analysisClient.clustersByProject, "testproject")

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCNotFound, "project dataset not provisioned in Weetbix or cluster analysis is not yet available")
				})
				Convey("With project not configured", func() {
					err := config.SetTestProjectConfig(ctx, map[string]*configpb.ProjectConfig{})
					So(err, ShouldBeNil)

					// Run
					response, err := server.BatchGet(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCFailedPrecondition, "project does not exist in Weetbix")
				})
			})
		})
		Convey("QueryClusterSummaries", func() {
			authState.IdentityPermissions = listTestResultsPermissions(
				"testproject:realm1",
				"testproject:realm2",
				"otherproject:realm3",
			)
			authState.IdentityPermissions = append(authState.IdentityPermissions, []authtest.RealmPermission{
				{
					Realm:      "testproject:@root",
					Permission: perms.PermListClusters,
				},
				{
					Realm:      "testproject:@root",
					Permission: perms.PermExpensiveClusterQueries,
				},
				{
					Realm:      "testproject:@root",
					Permission: perms.PermGetRule,
				},
				{
					Realm:      "testproject:@root",
					Permission: perms.PermGetRuleDefinition,
				},
			}...)

			analysisClient.clusterMetricsByProject["testproject"] = []*analysis.ClusterSummary{
				{
					ClusterID: clustering.ClusterID{
						Algorithm: rulesalgorithm.AlgorithmName,
						ID:        rs[0].RuleID,
					},
					PresubmitRejects:           1,
					CriticalFailuresExonerated: 2,
					Failures:                   3,
					ExampleFailureReason:       bigquery.NullString{Valid: true, StringVal: "Example failure reason."},
					ExampleTestID:              "TestID 1",
				},
				{
					ClusterID: clustering.ClusterID{
						Algorithm: "reason-v3",
						ID:        "cccccc00000000000000000000000001",
					},
					PresubmitRejects:           4,
					CriticalFailuresExonerated: 5,
					Failures:                   6,
					ExampleFailureReason:       bigquery.NullString{Valid: true, StringVal: "Example failure reason 2."},
					ExampleTestID:              "TestID 3",
				},
				{
					ClusterID: clustering.ClusterID{
						// Rule that is no longer active.
						Algorithm: rulesalgorithm.AlgorithmName,
						ID:        "01234567890abcdef01234567890abcdef",
					},
					PresubmitRejects:           7,
					CriticalFailuresExonerated: 8,
					Failures:                   9,
					ExampleFailureReason:       bigquery.NullString{Valid: true, StringVal: "Example failure reason."},
					ExampleTestID:              "TestID 1",
				},
			}
			analysisClient.expectedRealmsQueried = []string{"testproject:realm1", "testproject:realm2"}

			request := &pb.QueryClusterSummariesRequest{
				Project:       "testproject",
				FailureFilter: "test_id:\"pita.Boot\" failure_reason:\"failed to boot\"",
				OrderBy:       "presubmit_rejects desc, critical_failures_exonerated, failures desc",
			}
			Convey("Not authorised to list clusters", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermListClusters)

				response, err := server.QueryClusterSummaries(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.list")
				So(response, ShouldBeNil)
			})
			Convey("Not authorised to perform expensive queries", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermExpensiveClusterQueries)

				response, err := server.QueryClusterSummaries(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.expensiveQueries")
				So(response, ShouldBeNil)
			})
			Convey("Not authorised to get rules", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetRule)

				response, err := server.QueryClusterSummaries(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.rules.get")
				So(response, ShouldBeNil)
			})
			Convey("Valid request", func() {
				expectedResponse := &pb.QueryClusterSummariesResponse{
					ClusterSummaries: []*pb.ClusterSummary{
						{
							ClusterId: &pb.ClusterId{
								Algorithm: "rules",
								Id:        rs[0].RuleID,
							},
							Title: rs[0].RuleDefinition,
							Bug: &pb.AssociatedBug{
								System:   "monorail",
								Id:       "chromium/7654321",
								LinkText: "crbug.com/7654321",
								Url:      "https://bugs.chromium.org/p/chromium/issues/detail?id=7654321",
							},
							PresubmitRejects:           1,
							CriticalFailuresExonerated: 2,
							Failures:                   3,
						},
						{
							ClusterId: &pb.ClusterId{
								Algorithm: "reason-v3",
								Id:        "cccccc00000000000000000000000001",
							},
							Title:                      `Example failure reason 2.`,
							PresubmitRejects:           4,
							CriticalFailuresExonerated: 5,
							Failures:                   6,
						},
						{
							ClusterId: &pb.ClusterId{
								Algorithm: "rules",
								Id:        "01234567890abcdef01234567890abcdef",
							},
							Title:                      `(rule archived)`,
							PresubmitRejects:           7,
							CriticalFailuresExonerated: 8,
							Failures:                   9,
						},
					},
				}

				Convey("With filters and order by", func() {
					response, err := server.QueryClusterSummaries(ctx, request)
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
				Convey("Without filters or order", func() {
					request.FailureFilter = ""
					request.OrderBy = ""

					response, err := server.QueryClusterSummaries(ctx, request)
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
				Convey("Without rule definition get permission", func() {
					authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetRuleDefinition)

					// The RPC cannot return the rule definition as the
					// cluster title as the user is not authorised to see it.
					// Instead, it should generate a description of the
					// content of the cluster based on what the user can see.
					expectedResponse.ClusterSummaries[0].Title = "Selected failures in TestID 1"

					response, err := server.QueryClusterSummaries(ctx, request)
					So(err, ShouldBeNil)
					So(response, ShouldResembleProto, expectedResponse)
				})
			})
			Convey("Invalid request", func() {
				Convey("Dataset does not exist", func() {
					delete(analysisClient.clusterMetricsByProject, "testproject")

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCNotFound, "project dataset not provisioned in Weetbix or cluster analysis is not yet available")
				})
				Convey("Failure filter syntax is invalid", func() {
					request.FailureFilter = "test_id::"

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "failure_filter: expected arg after :")
				})
				Convey("Failure filter references non-existant column", func() {
					request.FailureFilter = `test:"pita.Boot"`

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, `failure_filter: no filterable field named "test"`)
				})
				Convey("Failure filter references unimplemented feature", func() {
					request.FailureFilter = "test<=\"blah\""

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "failure_filter: comparator operator not implemented yet")
				})
				Convey("Order by syntax invalid", func() {
					request.OrderBy = "presubmit_rejects asc"

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, `order_by: invalid ordering "presubmit_rejects asc"`)
				})
				Convey("Order by syntax references invalid column", func() {
					request.OrderBy = "not_exists desc"

					// Run
					response, err := server.QueryClusterSummaries(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, `order_by: no sortable field named "not_exists"`)
				})
			})
		})
		Convey("GetReclusteringProgress", func() {
			authState.IdentityPermissions = []authtest.RealmPermission{{
				Realm:      "testproject:@root",
				Permission: perms.PermGetCluster,
			}}

			request := &pb.GetReclusteringProgressRequest{
				Name: "projects/testproject/reclusteringProgress",
			}
			Convey("Not authorised to get cluster", func() {
				authState.IdentityPermissions = removePermission(authState.IdentityPermissions, perms.PermGetCluster)

				response, err := server.GetReclusteringProgress(ctx, request)
				So(err, ShouldBeRPCPermissionDenied, "caller does not have permission weetbix.clusters.get")
				So(response, ShouldBeNil)
			})
			Convey("With a valid request", func() {
				rulesVersion := time.Date(2021, time.January, 1, 1, 0, 0, 0, time.UTC)
				reference := time.Date(2020, time.February, 1, 1, 0, 0, 0, time.UTC)
				configVersion := time.Date(2019, time.March, 1, 1, 0, 0, 0, time.UTC)
				rns := []*runs.ReclusteringRun{
					runs.NewRun(0).
						WithProject("testproject").
						WithAttemptTimestamp(reference.Add(-5 * time.Minute)).
						WithRulesVersion(rulesVersion).
						WithAlgorithmsVersion(2).
						WithConfigVersion(configVersion).
						WithNoReportedProgress().
						Build(),
					runs.NewRun(1).
						WithProject("testproject").
						WithAttemptTimestamp(reference.Add(-10 * time.Minute)).
						WithRulesVersion(rulesVersion).
						WithAlgorithmsVersion(2).
						WithConfigVersion(configVersion).
						WithReportedProgress(500).
						Build(),
					runs.NewRun(2).
						WithProject("testproject").
						WithAttemptTimestamp(reference.Add(-20 * time.Minute)).
						WithRulesVersion(rulesVersion.Add(-1 * time.Hour)).
						WithAlgorithmsVersion(1).
						WithConfigVersion(configVersion.Add(-1 * time.Hour)).
						WithCompletedProgress().
						Build(),
				}
				err := runs.SetRunsForTesting(ctx, rns)
				So(err, ShouldBeNil)

				// Run
				response, err := server.GetReclusteringProgress(ctx, request)

				// Verify.
				So(err, ShouldBeNil)
				So(response, ShouldResembleProto, &pb.ReclusteringProgress{
					Name:             "projects/testproject/reclusteringProgress",
					ProgressPerMille: 500,
					Last: &pb.ClusteringVersion{
						AlgorithmsVersion: 1,
						ConfigVersion:     timestamppb.New(configVersion.Add(-1 * time.Hour)),
						RulesVersion:      timestamppb.New(rulesVersion.Add(-1 * time.Hour)),
					},
					Next: &pb.ClusteringVersion{
						AlgorithmsVersion: 2,
						ConfigVersion:     timestamppb.New(configVersion),
						RulesVersion:      timestamppb.New(rulesVersion),
					},
				})
			})
			Convey("With an invalid request", func() {
				Convey("Invalid name", func() {
					request.Name = "invalid"

					// Run
					response, err := server.GetReclusteringProgress(ctx, request)

					// Verify
					So(response, ShouldBeNil)
					So(err, ShouldBeRPCInvalidArgument, "name: invalid reclustering progress name, expected format: projects/{project}/reclusteringProgress")
				})
			})
		})
	})
}

func listTestResultsPermissions(realms ...string) []authtest.RealmPermission {
	var result []authtest.RealmPermission
	for _, r := range realms {
		result = append(result, authtest.RealmPermission{
			Realm:      r,
			Permission: rdbperms.PermListTestResults,
		})
		result = append(result, authtest.RealmPermission{
			Realm:      r,
			Permission: rdbperms.PermListTestExonerations,
		})
	}
	return result
}

func removePermission(perms []authtest.RealmPermission, permission realms.Permission) []authtest.RealmPermission {
	var result []authtest.RealmPermission
	for _, p := range perms {
		if p.Permission != permission {
			result = append(result, p)
		}
	}
	return result
}

func emptyMetricValues() *pb.Cluster_MetricValues {
	return &pb.Cluster_MetricValues{
		OneDay:   &pb.Cluster_MetricValues_Counts{},
		ThreeDay: &pb.Cluster_MetricValues_Counts{},
		SevenDay: &pb.Cluster_MetricValues_Counts{},
	}
}

func failureReasonClusterEntry(projectcfg *compiledcfg.ProjectConfig, primaryErrorMessage string) *pb.ClusterResponse_ClusteredTestResult_ClusterEntry {
	alg := &failurereason.Algorithm{}
	clusterID := alg.Cluster(projectcfg, &clustering.Failure{
		Reason: &pb.FailureReason{
			PrimaryErrorMessage: primaryErrorMessage,
		},
	})
	return &pb.ClusterResponse_ClusteredTestResult_ClusterEntry{
		ClusterId: &pb.ClusterId{
			Algorithm: failurereason.AlgorithmName,
			Id:        hex.EncodeToString(clusterID),
		},
	}
}

func testNameClusterEntry(projectcfg *compiledcfg.ProjectConfig, testID string) *pb.ClusterResponse_ClusteredTestResult_ClusterEntry {
	alg := &testname.Algorithm{}
	clusterID := alg.Cluster(projectcfg, &clustering.Failure{
		TestID: testID,
	})
	return &pb.ClusterResponse_ClusteredTestResult_ClusterEntry{
		ClusterId: &pb.ClusterId{
			Algorithm: testname.AlgorithmName,
			Id:        hex.EncodeToString(clusterID),
		},
	}
}

// sortClusterEntries sorts clusters by ascending Cluster ID.
func sortClusterEntries(entries []*pb.ClusterResponse_ClusteredTestResult_ClusterEntry) []*pb.ClusterResponse_ClusteredTestResult_ClusterEntry {
	result := make([]*pb.ClusterResponse_ClusteredTestResult_ClusterEntry, len(entries))
	copy(result, entries)
	sort.Slice(result, func(i, j int) bool {
		if result[i].ClusterId.Algorithm != result[j].ClusterId.Algorithm {
			return result[i].ClusterId.Algorithm < result[j].ClusterId.Algorithm
		}
		return result[i].ClusterId.Id < result[j].ClusterId.Id
	})
	return result
}

type fakeAnalysisClient struct {
	clustersByProject       map[string][]*analysis.Cluster
	clusterMetricsByProject map[string][]*analysis.ClusterSummary
	expectedRealmsQueried   []string
}

func newFakeAnalysisClient() *fakeAnalysisClient {
	return &fakeAnalysisClient{
		clustersByProject:       make(map[string][]*analysis.Cluster),
		clusterMetricsByProject: make(map[string][]*analysis.ClusterSummary),
	}
}

func (f *fakeAnalysisClient) ReadClusters(ctx context.Context, project string, clusterIDs []clustering.ClusterID) ([]*analysis.Cluster, error) {
	clusters, ok := f.clustersByProject[project]
	if !ok {
		return nil, analysis.ProjectNotExistsErr
	}

	var results []*analysis.Cluster
	for _, c := range clusters {
		include := false
		for _, ci := range clusterIDs {
			if ci == c.ClusterID {
				include = true
			}
		}
		if include {
			results = append(results, c)
		}
	}
	return results, nil
}

func (f *fakeAnalysisClient) QueryClusterSummaries(ctx context.Context, project string, options *analysis.QueryClusterSummariesOptions) ([]*analysis.ClusterSummary, error) {
	clusters, ok := f.clusterMetricsByProject[project]
	if !ok {
		return nil, analysis.ProjectNotExistsErr
	}

	set := stringset.NewFromSlice(options.Realms...)
	if set.Len() != len(f.expectedRealmsQueried) || !set.HasAll(f.expectedRealmsQueried...) {
		panic("realms passed to QueryClusterSummaries do not match expected")
	}

	_, _, err := analysis.ClusteredFailuresTable.WhereClause(options.FailureFilter, "w_")
	if err != nil {
		return nil, analysis.InvalidArgumentTag.Apply(errors.Annotate(err, "failure_filter").Err())
	}
	_, err = analysis.ClusterSummariesTable.OrderByClause(options.OrderBy)
	if err != nil {
		return nil, analysis.InvalidArgumentTag.Apply(errors.Annotate(err, "order_by").Err())
	}

	var results []*analysis.ClusterSummary
	for _, c := range clusters {
		results = append(results, c)
	}
	return results, nil
}
