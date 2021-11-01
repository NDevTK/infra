// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package bugclusters

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestRead(t *testing.T) {
	ctx := testutil.SpannerTestContext(t)
	Convey(`ReadForAllProjects`, t, func() {
		Convey(`Empty`, func() {
			setBugClusters(ctx, nil)

			clusters, err := ReadActiveForAllProjects(span.Single(ctx))
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, []*BugCluster{})
		})
		Convey(`Multiple`, func() {
			clustersToCreate := []*BugCluster{
				newBugCluster(0).Build(),
				newBugCluster(1).withActive(false).Build(),
				newBugCluster(2).withProject(1).Build(),
			}
			setBugClusters(ctx, clustersToCreate)

			clusters, err := ReadActiveForAllProjects(span.Single(ctx))
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, []*BugCluster{
				clustersToCreate[0],
				clustersToCreate[2],
			})
		})
	})
	Convey(`Read`, t, func() {
		Convey(`Empty`, func() {
			setBugClusters(ctx, nil)

			clusters, err := ReadActive(span.Single(ctx), "chromium")
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, []*BugCluster{})
		})
		Convey(`Multiple`, func() {
			clustersToCreate := []*BugCluster{
				newBugCluster(0).Build(),
				newBugCluster(1).withActive(false).Build(),
				newBugCluster(2).Build(),
				newBugCluster(3).withProject(1).Build(),
			}

			setBugClusters(ctx, clustersToCreate)

			clusters, err := ReadActive(span.Single(ctx), clustersToCreate[0].Project)
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, []*BugCluster{
				clustersToCreate[0],
				clustersToCreate[2],
			})
		})
	})
	Convey(`ReadBugsForCluster`, t, func() {
		Convey(`NoBugsInCluster`, func() {
			cluster := newBugCluster(0).Build()
			setBugClusters(ctx, []*BugCluster{
				cluster,
			})

			clusters, err := ReadBugsForCluster(span.Single(ctx), cluster.Project, "non-existing-cluster-id")
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, []*BugCluster{})
		})
		Convey(`MultipleBugsInCluster`, func() {
			clustersToCreate := []*BugCluster{
				newBugCluster(0).Build(),
				newBugCluster(1).withProject(0).withActive(false).Build(),
			}
			clustersToCreate[1].AssociatedClusterID = clustersToCreate[0].AssociatedClusterID
			setBugClusters(ctx, clustersToCreate)

			clusters, err := ReadBugsForCluster(span.Single(ctx), clustersToCreate[0].Project, clustersToCreate[0].AssociatedClusterID)
			So(err, ShouldBeNil)
			So(clusters, ShouldResemble, clustersToCreate)
		})
	})
	Convey(`Create`, t, func() {
		testCreate := func(bc *BugCluster) error {
			_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
				return Create(ctx, bc)
			})
			return err
		}
		Convey(`Valid`, func() {
			bc := newBugCluster(103).Build()
			err := testCreate(bc)
			So(err, ShouldBeNil)
			// Create followed by read already tested as part of Read tests.
		})
		Convey(`With missing Project`, func() {
			bc := newBugCluster(100).Build()
			bc.Project = ""
			err := testCreate(bc)
			So(err, ShouldErrLike, "project must be specified")
		})
		Convey(`With missing Bug`, func() {
			bc := newBugCluster(101).Build()
			bc.Bug = ""
			err := testCreate(bc)
			So(err, ShouldErrLike, "bug must be specified")
		})
		Convey(`With missing Associated Cluster`, func() {
			bc := newBugCluster(102).Build()
			bc.AssociatedClusterID = ""
			err := testCreate(bc)
			So(err, ShouldErrLike, "associated cluster must be specified")
		})
	})
}

type BugClusterBuilder struct {
	bugCluster BugCluster
}

func newBugCluster(clusterID int) *BugClusterBuilder {
	return &BugClusterBuilder{
		bugCluster: BugCluster{
			Project:             "project0",
			Bug:                 fmt.Sprintf("monorail/project0/%v", clusterID),
			AssociatedClusterID: fmt.Sprintf("some-cluster-id%v", clusterID),
			IsActive:            true,
		},
	}
}

func (c *BugClusterBuilder) withProject(uniqifier int) *BugClusterBuilder {
	c.bugCluster.Project = fmt.Sprintf("project%v", uniqifier)
	c.bugCluster.Bug = fmt.Sprintf("monorail/project%v/%v", uniqifier, strings.TrimPrefix(c.bugCluster.AssociatedClusterID, "some-cluster-id"))
	return c
}

func (c *BugClusterBuilder) withActive(active bool) *BugClusterBuilder {
	c.bugCluster.IsActive = active
	return c
}

func (c *BugClusterBuilder) Build() *BugCluster {
	return &c.bugCluster
}

// setBugClusters replaces the set of stored bug clusters to match the given set.
func setBugClusters(ctx context.Context, bcs []*BugCluster) {
	testutil.MustApply(ctx,
		spanner.Delete("BugClusters", spanner.AllKeys()))
	// Insert some BugClusters.
	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		for _, bc := range bcs {
			if err := Create(ctx, bc); err != nil {
				return err
			}
		}
		return nil
	})
	So(err, ShouldBeNil)
}
