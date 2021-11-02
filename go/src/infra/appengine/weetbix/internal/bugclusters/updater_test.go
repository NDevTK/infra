// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugclusters

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/bugs"
	"infra/appengine/weetbix/internal/bugs/monorail"
	"infra/appengine/weetbix/internal/clustering"
	"infra/appengine/weetbix/internal/clustering/algorithms"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/testutil"

	"cloud.google.com/go/bigquery"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/config/validation"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/proto"
)

func TestRun(t *testing.T) {
	ctx := testutil.SpannerTestContext(t)
	Convey("Run bug updates", t, func() {
		setRules(ctx, nil)

		f := &monorail.FakeIssuesStore{
			NextID:            100,
			PriorityFieldName: "projects/chromium/fieldDefs/11",
		}
		user := monorail.AutomationUsers[0]
		mc, err := monorail.NewClient(monorail.UseFakeIssuesClient(ctx, f, user), "myhost")
		So(err, ShouldBeNil)

		clusters := []*analysis.ClusterSummary{
			makeCluster(0),
			makeCluster(1),
			makeCluster(2),
			makeCluster(3),
		}
		cc := &fakeClusterClient{
			clusters: clusters,
		}

		project := "chromium"
		mgrs := make(map[string]BugManager)
		monorailCfg := monorail.ChromiumTestConfig()
		mgrs[monorail.ManagerName] = monorail.NewBugManager(mc, monorailCfg)

		thres := &config.ImpactThreshold{
			// Should be more onerous than the "keep-open" thresholds
			// configured for each individual bug manager.
			UnexpectedFailures_1D: proto.Int64(100),
			UnexpectedFailures_3D: proto.Int64(300),
			UnexpectedFailures_7D: proto.Int64(700),
		}

		Convey("Configuration used for testing is valid", func() {
			c := validation.Context{Context: context.Background()}
			projectCfg := &config.ProjectConfig{
				Monorail:           monorailCfg,
				BugFilingThreshold: thres,
			}
			config.ValidateProjectConfig(&c, projectCfg)
			So(c.Finalize(), ShouldBeNil)
		})
		Convey("With no impactful clusters", func() {
			bu := NewBugUpdater(project, mgrs, cc, thres)
			err = bu.Run(ctx)
			So(err, ShouldBeNil)

			// No failure association rules.
			rules, err := ReadActive(span.Single(ctx), project)
			So(err, ShouldBeNil)
			So(rules, ShouldResemble, []*FailureAssociationRule{})

			// No monorail issues.
			So(f.Issues, ShouldBeNil)
		})
		Convey("With a cluster above impact thresold", func() {
			reasonAlg, err := algorithms.ByName("failurereason-v1")
			So(err, ShouldBeNil)

			sourceClusterID := clustering.ClusterID{
				Algorithm: "failurereason-v1",
				ID: hex.EncodeToString(reasonAlg.Cluster(&clustering.Failure{
					Reason: "Test failure reason 99.",
				})),
			}
			clusters[1].ClusterID = sourceClusterID
			clusters[1].ExampleFailureReason = bigquery.NullString{StringVal: "Test failure reason 105.", Valid: true}

			test := func() {
				bu := NewBugUpdater(project, mgrs, cc, thres)
				err = bu.Run(ctx)
				So(err, ShouldBeNil)

				rules, err := ReadActive(span.Single(ctx), project)
				So(err, ShouldBeNil)

				So(len(rules), ShouldEqual, 1)
				rule := rules[0]

				expected := &FailureAssociationRule{
					Project:        "chromium",
					RuleDefinition: `reason LIKE "Test failure reason %."`,
					Bug:            "monorail/chromium/100",
					IsActive:       true,
					SourceCluster:  sourceClusterID,
				}

				// Accept whatever bug cluster ID has been generated.
				So(len(rule.RuleID), ShouldEqual, 16)
				expected.RuleID = rule.RuleID

				// Accept creation and last updated times, as set by Spanner.
				So(rule.CreationTime, ShouldNotBeZeroValue)
				expected.CreationTime = rule.CreationTime
				So(rule.LastUpdated, ShouldNotBeZeroValue)
				expected.LastUpdated = rule.LastUpdated

				So(rule, ShouldResemble, expected)
				So(len(f.Issues), ShouldEqual, 1)
				So(f.Issues[0].Issue.Name, ShouldEqual, "projects/chromium/issues/100")
				So(f.Issues[0].Issue.Summary, ShouldContainSubstring, "Test failure reason.")
			}
			Convey("1d unexpected failures", func() {
				clusters[1].Failures1d.Residual = 100
				test()
			})
			Convey("3d unexpected failures", func() {
				clusters[1].Failures3d.Residual = 300
				test()
			})
			Convey("7d unexpected failures", func() {
				clusters[1].Failures7d.Residual = 700
				test()
			})
		})
		Convey("With multiple clusters above impact thresold", func() {
			expectBugClusters := func(count int) {
				bugClusters, err := ReadActive(span.Single(ctx), project)
				So(err, ShouldBeNil)
				So(len(bugClusters), ShouldEqual, count)
				So(len(f.Issues), ShouldEqual, count)
			}
			clusters[1].Failures1d.Residual = 200
			clusters[2].Failures3d.Residual = 300
			clusters[3].Failures7d.Residual = 700

			bu := NewBugUpdater(project, mgrs, cc, thres)
			// Limit to one bug filed each time, so that
			// we test change throttling.
			bu.MaxBugsFiledPerRun = 1

			err = bu.Run(ctx)
			So(err, ShouldBeNil)
			expectBugClusters(1)

			err = bu.Run(ctx)
			So(err, ShouldBeNil)
			expectBugClusters(2)

			err = bu.Run(ctx)
			So(err, ShouldBeNil)

			expectFinalBugClusters := func() {
				// Check final set of bugs is as expected.
				rules, err := ReadActive(span.Single(ctx), project)
				So(err, ShouldBeNil)
				for _, r := range rules {
					// Accept whatever values the implementation has set.
					r.RuleID = ""
					r.CreationTime = time.Time{}
					r.LastUpdated = time.Time{}
				}

				So(rules, ShouldResemble, []*FailureAssociationRule{
					{
						Project:        "chromium",
						RuleDefinition: `test = "testname-1"`,
						Bug:            "monorail/chromium/100",
						SourceCluster:  clusterID(1),
						IsActive:       true,
					},
					{
						Project:        "chromium",
						RuleDefinition: `test = "testname-2"`,
						Bug:            "monorail/chromium/101",
						SourceCluster:  clusterID(2),
						IsActive:       true,
					},
					{
						Project:        "chromium",
						RuleDefinition: `test = "testname-3"`,
						Bug:            "monorail/chromium/102",
						SourceCluster:  clusterID(3),
						IsActive:       true,
					},
				})
				So(len(f.Issues), ShouldEqual, 3)
			}
			expectFinalBugClusters()

			// Further updates do nothing.
			originalIssues := monorail.CopyIssuesStore(f)
			err = bu.Run(ctx)
			So(err, ShouldBeNil)
			So(f, monorail.ShouldResembleIssuesStore, originalIssues)
			expectFinalBugClusters()

			Convey("Changing cluster priority updates issue priority", func() {
				issue := f.Issues[2].Issue
				So(issue.Name, ShouldEqual, "projects/chromium/issues/102")
				So(monorail.ChromiumTestIssuePriority(issue), ShouldNotEqual, "0")

				bugs.SetResidualImpact(clusters[3], monorail.ChromiumP0Impact())
				err = bu.Run(ctx)
				So(err, ShouldBeNil)

				So(len(f.Issues), ShouldEqual, 3)
				issue = f.Issues[2].Issue
				So(issue.Name, ShouldEqual, "projects/chromium/issues/102")
				So(monorail.ChromiumTestIssuePriority(issue), ShouldEqual, "0")

				expectFinalBugClusters()
			})
			Convey("Deleting cluster closes issue", func() {
				issue := f.Issues[0].Issue
				So(issue.Name, ShouldEqual, "projects/chromium/issues/100")
				So(issue.Status.Status, ShouldEqual, monorail.UntriagedStatus)

				// Drop the cluster at index 1.
				cc.clusters = []*analysis.ClusterSummary{cc.clusters[0], cc.clusters[2], cc.clusters[3]}
				err = bu.Run(ctx)
				So(err, ShouldBeNil)

				So(len(f.Issues), ShouldEqual, 3)
				issue = f.Issues[0].Issue
				So(issue.Name, ShouldEqual, "projects/chromium/issues/100")
				So(issue.Status.Status, ShouldEqual, monorail.VerifiedStatus)
			})
		})
	})
}

func makeCluster(uniqifier int) *analysis.ClusterSummary {
	return &analysis.ClusterSummary{
		ClusterID:     clusterID(uniqifier),
		Failures1d:    analysis.Counts{Residual: 9},
		Failures3d:    analysis.Counts{Residual: 29},
		Failures7d:    analysis.Counts{Residual: 69},
		ExampleTestID: fmt.Sprintf("testname-%v", uniqifier),
	}
}

func clusterID(uniqifier int) clustering.ClusterID {
	return clustering.ClusterID{
		Algorithm: "testname-v1",
		ID:        hex.EncodeToString([]byte(fmt.Sprintf("cluster-id-%v", uniqifier))),
	}
}

type fakeClusterClient struct {
	clusters []*analysis.ClusterSummary
}

func (f *fakeClusterClient) ReadImpactfulClusters(ctx context.Context, opts analysis.ImpactfulClusterReadOptions) ([]*analysis.ClusterSummary, error) {
	var results []*analysis.ClusterSummary
	for _, c := range f.clusters {
		include := (opts.Thresholds.UnexpectedFailures_1D != nil && int64(c.Failures1d.Residual) >= *opts.Thresholds.UnexpectedFailures_1D) ||
			(opts.Thresholds.UnexpectedFailures_3D != nil && int64(c.Failures3d.Residual) >= *opts.Thresholds.UnexpectedFailures_3D) ||
			(opts.Thresholds.UnexpectedFailures_7D != nil && int64(c.Failures7d.Residual) >= *opts.Thresholds.UnexpectedFailures_7D)
		if include {
			results = append(results, c)
		}
	}
	return results, nil
}

func containsValue(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
