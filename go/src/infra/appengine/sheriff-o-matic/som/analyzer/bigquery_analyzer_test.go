// Copyright 2017 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging/gologger"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"infra/appengine/sheriff-o-matic/som/analyzer/step"
	"infra/appengine/sheriff-o-matic/som/model"
	"infra/monitoring/messages"
)

type mockResults struct {
	failures []failureRow
	err      error
	curr     int
}

func (m *mockResults) Next(dst interface{}) error {
	if m.curr >= len(m.failures) {
		return iterator.Done
	}
	fdst := dst.(*failureRow)
	*fdst = m.failures[m.curr]
	m.curr++
	return m.err
}

func TestMockBQResults(t *testing.T) {
	Convey("no results", t, func() {
		mr := &mockResults{}
		r := &failureRow{}
		So(mr.Next(r), ShouldEqual, iterator.Done)
	})
	Convey("copy op works", t, func() {
		mr := &mockResults{
			failures: []failureRow{
				{
					StepName: "foo",
				},
			},
		}
		r := failureRow{}
		err := mr.Next(&r)
		So(err, ShouldBeNil)
		So(r.StepName, ShouldEqual, "foo")
		So(mr.Next(&r), ShouldEqual, iterator.Done)
	})

}

func TestGenerateBuilderURL(t *testing.T) {
	Convey("Test builder with no space", t, func() {
		project := "chromium"
		bucket := "ci"
		builderName := "Win"
		url := generateBuilderURL(project, bucket, builderName)
		So(url, ShouldEqual, "https://ci.chromium.org/p/chromium/builders/ci/Win")
	})
	Convey("Test builder with some spaces", t, func() {
		project := "chromium"
		bucket := "ci"
		builderName := "Win 7 Test"
		url := generateBuilderURL(project, bucket, builderName)
		So(url, ShouldEqual, "https://ci.chromium.org/p/chromium/builders/ci/Win%207%20Test")
	})
	Convey("Test builder with special characters", t, func() {
		project := "chromium"
		bucket := "ci"
		builderName := "Mac 10.13 Tests (dbg)"
		url := generateBuilderURL(project, bucket, builderName)
		So(url, ShouldEqual, "https://ci.chromium.org/p/chromium/builders/ci/Mac%2010.13%20Tests%20%28dbg%29")
	})
}

func TestGenerateBuildURL(t *testing.T) {
	Convey("Test build url with build ID", t, func() {
		project := "chromium"
		bucket := "ci"
		builderName := "Win"
		buildID := bigquery.NullInt64{Int64: 8127364737474, Valid: true}
		url := generateBuildURL(project, bucket, builderName, buildID)
		So(url, ShouldEqual, "https://ci.chromium.org/p/chromium/builders/ci/Win/b8127364737474")
	})
	Convey("Test build url with empty buildID", t, func() {
		project := "chromium"
		bucket := "ci"
		builderName := "Win"
		buildID := bigquery.NullInt64{}
		url := generateBuildURL(project, bucket, builderName, buildID)
		So(url, ShouldEqual, "")
	})
}

// Make SQL query uniform, for the purpose of testing
func formatQuery(query string) string {
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	query = regexp.MustCompile(`\s?\(\s?`).ReplaceAllString(query, "(")
	query = regexp.MustCompile(`\s?\)\s?`).ReplaceAllString(query, ")")
	return query
}

func TestGenerateSQLQuery(t *testing.T) {
	c := gaetesting.TestingContext()

	Convey("Test generate SQL query for project", t, func() {
		expected := `
			SELECT
			  Project,
			  Bucket,
			  Builder,
			  BuilderGroup,
			  SheriffRotations,
			  Critical,
			  StepName,
			  TestNamesFingerprint,
			  TestNamesTrunc,
			  TestsTrunc,
			  NumTests,
			  BuildIdBegin,
			  BuildIdEnd,
			  BuildNumberBegin,
			  BuildNumberEnd,
			  CPRangeOutputBegin,
			  CPRangeOutputEnd,
			  CPRangeInputBegin,
			  CPRangeInputEnd,
			  CulpritIdRangeBegin,
			  CulpritIdRangeEnd,
			  StartTime,
			  BuildStatus
			FROM
				` + "`sheriff-o-matic.chrome.sheriffable_failures`"
		actual := generateQueryForProject("sheriff-o-matic", "chrome")
		So(formatQuery(actual), ShouldEqual, formatQuery(expected))
	})

	Convey("Test generate SQL query for chromeos", t, func() {
		treeName := "chromeos"
		tree := &model.Tree{
			Name: treeName,
		}
		So(datastore.Put(c, tree), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()
		expected := `
			SELECT
			  Project,
			  Bucket,
			  Builder,
			  BuilderGroup,
			  SheriffRotations,
			  Critical,
			  StepName,
			  TestNamesFingerprint,
			  TestNamesTrunc,
			  TestsTrunc,
			  NumTests,
			  BuildIdBegin,
			  BuildIdEnd,
			  BuildNumberBegin,
			  BuildNumberEnd,
			  CPRangeOutputBegin,
			  CPRangeOutputEnd,
			  CPRangeInputBegin,
			  CPRangeInputEnd,
			  CulpritIdRangeBegin,
			  CulpritIdRangeEnd,
			  StartTime,
			  BuildStatus
			FROM
				` + "`sheriff-o-matic.chromeos.sheriffable_failures`" + `
			WHERE Project = "chromeos"
				AND ((Bucket IN ("postsubmit"))
                                     OR (Bucket IN ("release")
					 AND Builder LIKE "%-release-main"))
				AND (Critical != "NO" OR Critical is NULL)
		`
		actual, err := generateSQLQuery(c, treeName, "sheriff-o-matic")
		So(formatQuery(actual), ShouldEqual, formatQuery(expected))
		So(err, ShouldBeNil)
	})

	Convey("Test generate SQL query for fuchsia", t, func() {
		treeName := "fuchsia"
		tree := &model.Tree{
			Name:                     treeName,
			BuildBucketProjectFilter: "fuchsia-test",
		}
		So(datastore.Put(c, tree), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()
		expected := `
			SELECT
			  Project,
			  Bucket,
			  Builder,
			  BuilderGroup,
			  SheriffRotations,
			  Critical,
			  StepName,
			  TestNamesFingerprint,
			  TestNamesTrunc,
			  TestsTrunc,
			  NumTests,
			  BuildIdBegin,
			  BuildIdEnd,
			  BuildNumberBegin,
			  BuildNumberEnd,
			  CPRangeOutputBegin,
			  CPRangeOutputEnd,
			  CPRangeInputBegin,
			  CPRangeInputEnd,
			  CulpritIdRangeBegin,
			  CulpritIdRangeEnd,
			  StartTime,
			  BuildStatus
			FROM
				` + "`sheriff-o-matic.fuchsia.sheriffable_failures`" + `
			WHERE
				Project = "fuchsia-test"
				AND Bucket = "global.ci"
			LIMIT
				1000
		`
		actual, err := generateSQLQuery(c, treeName, "sheriff-o-matic")
		So(formatQuery(actual), ShouldEqual, formatQuery(expected))
		So(err, ShouldBeNil)
	})

	Convey("Test generate SQL query for angle", t, func() {
		treeName := "angle"
		tree := &model.Tree{
			Name:                     treeName,
			BuildBucketProjectFilter: "angle-test",
		}
		So(datastore.Put(c, tree), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()
		expected := `
			SELECT
			  Project,
			  Bucket,
			  Builder,
			  BuilderGroup,
			  SheriffRotations,
			  Critical,
			  StepName,
			  TestNamesFingerprint,
			  TestNamesTrunc,
			  TestsTrunc,
			  NumTests,
			  BuildIdBegin,
			  BuildIdEnd,
			  BuildNumberBegin,
			  BuildNumberEnd,
			  CPRangeOutputBegin,
			  CPRangeOutputEnd,
			  CPRangeInputBegin,
			  CPRangeInputEnd,
			  CulpritIdRangeBegin,
			  CulpritIdRangeEnd,
			  StartTime,
			  BuildStatus
			FROM
				` + "`sheriff-o-matic.angle.sheriffable_failures`" + `
			WHERE
				"angle" in UNNEST(SheriffRotations)
		`
		actual, err := generateSQLQuery(c, treeName, "sheriff-o-matic")
		So(formatQuery(actual), ShouldEqual, formatQuery(expected))
		So(err, ShouldBeNil)
	})

	Convey("Test generate SQL query for invalid tree", t, func() {
		_, err := generateSQLQuery(c, "abc", "sheriff-o-matic")
		So(err, ShouldNotBeNil)
	})
}

type mockBuildersClient struct{}

func (mbc mockBuildersClient) ListBuilders(c context.Context, req *buildbucketpb.ListBuildersRequest, opts ...grpc.CallOption) (*buildbucketpb.ListBuildersResponse, error) {
	if req.Bucket == "ci" {
		return &buildbucketpb.ListBuildersResponse{
			Builders: []*buildbucketpb.BuilderItem{
				{
					Id: &buildbucketpb.BuilderID{
						Project: "chromium",
						Bucket:  "ci",
						Builder: "ci_1",
					},
				},
				{
					Id: &buildbucketpb.BuilderID{
						Project: "chromium",
						Bucket:  "ci",
						Builder: "ci_2",
					},
				},
			},
		}, nil
	}
	if req.Bucket == "try" {
		return &buildbucketpb.ListBuildersResponse{
			Builders: []*buildbucketpb.BuilderItem{
				{
					Id: &buildbucketpb.BuilderID{
						Project: "chromium",
						Bucket:  "try",
						Builder: "try_1",
					},
				},
				{
					Id: &buildbucketpb.BuilderID{
						Project: "chromium",
						Bucket:  "try",
						Builder: "try_2",
					},
				},
			},
		}, nil
	}
	if req.Bucket == "err" {
		return nil, fmt.Errorf("some infra error")
	}
	if req.Bucket == "notfound" {
		return nil, status.Error(codes.NotFound, "Not found")
	}

	return nil, nil
}

func TestFilterDeletedBuilders(t *testing.T) {
	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)
	cl := mockBuildersClient{}

	Convey("no builder", t, func() {
		failureRows := []failureRow{}
		filtered, err := filterDeletedBuildersWithClient(ctx, cl, failureRows)
		So(err, ShouldBeNil)
		So(filtered, ShouldBeEmpty)
	})

	Convey("builders belong to one bucket", t, func() {
		failureRows := []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_3",
			},
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_2",
			},
		}
		filtered, err := filterDeletedBuildersWithClient(ctx, cl, failureRows)
		So(err, ShouldBeNil)
		So(filtered, ShouldResemble, []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_2",
			},
		})
	})

	Convey("builders belong to more than one buckets", t, func() {
		failureRows := []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_3",
			},
			{
				Project: "chromium",
				Bucket:  "try",
				Builder: "try_3",
			},
			{
				Project: "chromium",
				Bucket:  "try",
				Builder: "try_1",
			},
		}
		filtered, err := filterDeletedBuildersWithClient(ctx, cl, failureRows)
		So(err, ShouldBeNil)
		So(filtered, ShouldResemble, []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "try",
				Builder: "try_1",
			},
		})
	})

	Convey("rpc returns errors", t, func() {
		failureRows := []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "err",
				Builder: "err_1",
			},
		}
		_, err := filterDeletedBuildersWithClient(ctx, cl, failureRows)
		So(err, ShouldNotBeNil)
	})

	Convey("rpc returns NotFound", t, func() {
		failureRows := []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
			{
				Project: "chromium",
				Bucket:  "notfound",
				Builder: "notfound_1",
			},
		}
		filtered, err := filterDeletedBuildersWithClient(ctx, cl, failureRows)
		So(err, ShouldBeNil)
		So(filtered, ShouldResemble, []failureRow{
			{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ci_1",
			},
		})
	})
}

func TestProcessBQResults(t *testing.T) {
	ctx := context.Background()
	ctx = gologger.StdConfig.Use(ctx)

	Convey("smoke", t, func() {
		failureRows := []failureRow{}
		got, err := processBQResults(ctx, failureRows)
		So(err, ShouldEqual, nil)
		So(got, ShouldBeEmpty)
	})

	Convey("single result, only start/end build numbers", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "some builder",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 1)
	})

	Convey("single result, only end build number", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "some builder",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 1)
	})

	Convey("single result, start/end build numbers, single test name", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "some builder",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 1)
		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 1,
			Tests:           makeTestWithResults("1"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)
	})

	Convey("multiple results, start/end build numbers, same step, same test name", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 1",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 2",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestNamesTrunc: bigquery.NullString{
					StringVal: "some/test/name",
					Valid:     true,
				},
				TestsTrunc: makeTestFailures("1"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)
		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 1,
			Tests:           makeTestWithResults("1"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)
		So(len(got[1].Builders), ShouldEqual, 1)
	})

	Convey("multiple results, start/end build numbers, different steps, different sets of test names", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step 1",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 1",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1", "2"),
				NumTests: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
			},
			{
				StepName: "some step 2",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 2",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("3"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		sort.Sort(byStepName(got))
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)

		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 1",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 2,
			Tests:           makeTestWithResults("1", "2"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)

		reason = got[1].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 2",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 1,
			Tests:           makeTestWithResults("3"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)
	})

	Convey("multiple results, start/end build numbers, same step, different sets of test names", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step 1",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 1",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1", "2"),
				NumTests: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
			},
			{
				StepName: "some step 1",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 2",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("3"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		sort.Sort(byTests(got))
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)

		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 1",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 2,
			Tests:           makeTestWithResults("1", "2"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)
		So(got[0].Builders[0].Name, ShouldEqual, "builder 1")

		reason = got[1].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 1",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 1,
			Tests:           makeTestWithResults("3"),
		})
		So(len(got[1].Builders), ShouldEqual, 1)
		So(got[1].Builders[0].Name, ShouldEqual, "builder 2")
	})

	Convey("chromium.perf case: multiple results, different start build numbers, same end build number, same step, different sets of test names", t, func() {
		failureRows := []failureRow{
			{
				StepName: "performance_test_suite",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "win-10-perf",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 110,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("A1", "A2", "A3"),
				NumTests: bigquery.NullInt64{
					Int64: 3,
					Valid: true,
				},
			},
			{
				StepName: "performance_test_suite",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "win-10-perf",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 102,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 110,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("B1", "B2", "B3"),
				NumTests: bigquery.NullInt64{
					Int64: 3,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		sort.Sort(byTests(got))
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)

		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "performance_test_suite",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 3,
			Tests:           makeTestWithResults("A1", "A2", "A3"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)
		So(got[0].Builders[0].Name, ShouldEqual, "win-10-perf")
		So(got[0].Builders[0].FirstFailure, ShouldEqual, 100)
		So(got[0].Builders[0].LatestFailure, ShouldEqual, 110)

		reason = got[1].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "performance_test_suite",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 3,
			Tests:           makeTestWithResults("B1", "B2", "B3"),
		})
		So(len(got[1].Builders), ShouldEqual, 1)
		So(got[1].Builders[0].Name, ShouldEqual, "win-10-perf")
		So(got[1].Builders[0].FirstFailure, ShouldEqual, 102)
		So(got[1].Builders[0].LatestFailure, ShouldEqual, 110)
	})

	Convey("chromium.perf case: multiple results, same step, same truncated list of test names, different test name fingerprints", t, func() {
		failureRows := []failureRow{
			{
				StepName: "performance_test_suite",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "win-10-perf",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 110,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("A1", "A2", "A3"),
				NumTests: bigquery.NullInt64{
					Int64: 3,
					Valid: true,
				},
			},
			{
				StepName: "performance_test_suite",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "win-10-perf",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 102,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 110,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("A1", "A2", "A3"),
				NumTests: bigquery.NullInt64{
					Int64: 3,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		sort.Sort(byFirstFailure(got))
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)

		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "performance_test_suite",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 3,
			Tests:           makeTestWithResults("A1", "A2", "A3"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)

		So(got[0].Builders[0].Name, ShouldEqual, "win-10-perf")
		So(got[0].Builders[0].FirstFailure, ShouldEqual, 100)
		So(got[0].Builders[0].LatestFailure, ShouldEqual, 110)

		reason = got[1].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "performance_test_suite",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 3,
			Tests:           makeTestWithResults("A1", "A2", "A3"),
		})
		So(len(got[1].Builders), ShouldEqual, 1)
		So(got[1].Builders[0].Name, ShouldEqual, "win-10-perf")
		So(got[1].Builders[0].FirstFailure, ShouldEqual, 102)
		So(got[1].Builders[0].LatestFailure, ShouldEqual, 110)
	})

	Convey("multiple results, start/end build numbers, different steps, same set of test names", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step 1",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 1",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1", "2"),
				NumTests: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
			},
			{
				StepName: "some step 2",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder: "builder 2",
				Project: "some project",
				Bucket:  "some bucket",
				BuildIDBegin: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				BuildIDEnd: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestsTrunc: makeTestFailures("1", "2"),
				NumTests: bigquery.NullInt64{
					Int64: 2,
					Valid: true,
				},
			},
		}
		got, err := processBQResults(ctx, failureRows)
		sort.Sort(byStepName(got))
		So(err, ShouldEqual, nil)
		So(len(got), ShouldEqual, 2)
		reason := got[0].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 1",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 2,
			Tests:           makeTestWithResults("1", "2"),
		})
		So(len(got[0].Builders), ShouldEqual, 1)

		reason = got[1].Reason
		So(reason, ShouldNotBeNil)
		So(reason.Raw, ShouldResemble, &BqFailure{
			Name:            "some step 2",
			kind:            "test",
			severity:        messages.ReliableFailure,
			NumFailingTests: 2,
			Tests:           makeTestWithResults("1", "2"),
		})
		So(len(got[1].Builders), ShouldEqual, 1)
	})

	Convey("process changepoint result", t, func() {
		failureRows := []failureRow{
			{
				StepName: "some step",
				BuilderGroup: bigquery.NullString{
					StringVal: "some builder group",
					Valid:     true,
				},
				Builder:    "some builder",
				Project:    "some project",
				Bucket:     "some bucket",
				TestsTrunc: makeTestFailures("1"),
				NumTests: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
				TestNamesFingerprint: bigquery.NullInt64{
					Int64: 1,
					Valid: true,
				},
			},
		}
		Convey("one segment", func() {
			failureRows[0].TestsTrunc[0].Segments = []*segment{{
				StartHour: bigquery.NullTimestamp{
					Timestamp: time.Unix(3600*11, 0),
					Valid:     true,
				},
				CountUnexpectedResults: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
				CountTotalResults: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
			}}
			result := step.TestWithResult{
				TestName:                "some/test/1",
				TestID:                  "ninja://some/test/1",
				VariantHash:             "12341",
				ClusterName:             "reason-v3/1",
				Realm:                   "chromium:1",
				RefHash:                 "ref-hash/1",
				RegressionStartPosition: 0,
				RegressionEndPosition:   0,
				CurStartHour:            time.Unix(3600*11, 0),
				PrevEndHour:             time.Time{},
				CurCounts: step.Counts{
					UnexpectedResults: 10,
					TotalResults:      100,
				},
				PrevCounts: step.Counts{},
			}
			got, err := processBQResults(ctx, failureRows)
			So(err, ShouldEqual, nil)
			So(len(got), ShouldEqual, 1)
			reason := got[0].Reason
			So(reason, ShouldNotBeNil)
			So(reason.Raw, ShouldResembleProto, &BqFailure{
				Name:            "some step",
				kind:            "test",
				severity:        messages.ReliableFailure,
				NumFailingTests: 1,
				Tests:           []step.TestWithResult{result},
			})
			So(len(got[0].Builders), ShouldEqual, 1)
		})
		Convey("two segments, deterministic failure", func() {
			failureRows[0].TestsTrunc[0].Segments = []*segment{{
				StartHour: bigquery.NullTimestamp{
					Timestamp: time.Unix(3600*11, 0),
					Valid:     true,
				},
				CountUnexpectedResults: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				CountTotalResults: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				StartPosition: bigquery.NullInt64{
					Int64: 6,
					Valid: true,
				},
				StartPositionLowerBound99Th: bigquery.NullInt64{
					Int64: 5,
					Valid: true,
				},
				StartPositionUpperBound99Th: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
				{
					EndHour: bigquery.NullTimestamp{
						Timestamp: time.Unix(3600*10, 0),
						Valid:     true,
					},
					EndPosition: bigquery.NullInt64{
						Int64: 4,
						Valid: true,
					},
					CountUnexpectedResults: bigquery.NullInt64{
						Int64: 0,
						Valid: true,
					},
					CountTotalResults: bigquery.NullInt64{
						Int64: 100,
						Valid: true,
					},
				}}
			result := step.TestWithResult{
				TestName:                "some/test/1",
				TestID:                  "ninja://some/test/1",
				VariantHash:             "12341",
				ClusterName:             "reason-v3/1",
				Realm:                   "chromium:1",
				RefHash:                 "ref-hash/1",
				RegressionStartPosition: 4,
				RegressionEndPosition:   6,
				CurStartHour:            time.Unix(3600*11, 0),
				PrevEndHour:             time.Unix(3600*10, 0),
				CurCounts: step.Counts{
					UnexpectedResults: 100,
					TotalResults:      100,
				},
				PrevCounts: step.Counts{
					UnexpectedResults: 0,
					TotalResults:      100,
				},
			}
			got, err := processBQResults(ctx, failureRows)
			So(err, ShouldEqual, nil)
			So(len(got), ShouldEqual, 1)
			reason := got[0].Reason
			So(reason, ShouldNotBeNil)
			So(reason.Raw, ShouldResembleProto, &BqFailure{
				Name:            "some step",
				kind:            "test",
				severity:        messages.ReliableFailure,
				NumFailingTests: 1,
				Tests:           []step.TestWithResult{result},
			})
			So(len(got[0].Builders), ShouldEqual, 1)
		})
		Convey("two segments, non-deterministic failure", func() {
			failureRows[0].TestsTrunc[0].Segments = []*segment{{
				StartHour: bigquery.NullTimestamp{
					Timestamp: time.Unix(3600*11, 0),
					Valid:     true,
				},
				CountUnexpectedResults: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				CountTotalResults: bigquery.NullInt64{
					Int64: 100,
					Valid: true,
				},
				StartPosition: bigquery.NullInt64{
					Int64: 6,
					Valid: true,
				},
				StartPositionLowerBound99Th: bigquery.NullInt64{
					Int64: 5,
					Valid: true,
				},
				StartPositionUpperBound99Th: bigquery.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
				{
					EndHour: bigquery.NullTimestamp{
						Timestamp: time.Unix(3600*10, 0),
						Valid:     true,
					},
					EndPosition: bigquery.NullInt64{
						Int64: 4,
						Valid: true,
					},
					CountUnexpectedResults: bigquery.NullInt64{
						Int64: 1,
						Valid: true,
					},
					CountTotalResults: bigquery.NullInt64{
						Int64: 100,
						Valid: true,
					},
				}}
			result := step.TestWithResult{
				TestName:                "some/test/1",
				TestID:                  "ninja://some/test/1",
				VariantHash:             "12341",
				ClusterName:             "reason-v3/1",
				Realm:                   "chromium:1",
				RefHash:                 "ref-hash/1",
				RegressionStartPosition: 4,
				RegressionEndPosition:   10,
				CurStartHour:            time.Unix(3600*11, 0),
				PrevEndHour:             time.Unix(3600*10, 0),
				CurCounts: step.Counts{
					UnexpectedResults: 100,
					TotalResults:      100,
				},
				PrevCounts: step.Counts{
					UnexpectedResults: 1,
					TotalResults:      100,
				},
			}
			got, err := processBQResults(ctx, failureRows)
			So(err, ShouldEqual, nil)
			So(len(got), ShouldEqual, 1)
			reason := got[0].Reason
			So(reason, ShouldNotBeNil)
			So(reason.Raw, ShouldResembleProto, &BqFailure{
				Name:            "some step",
				kind:            "test",
				severity:        messages.ReliableFailure,
				NumFailingTests: 1,
				Tests:           []step.TestWithResult{result},
			})
			So(len(got[0].Builders), ShouldEqual, 1)
		})
	})
}

type byFirstFailure []*messages.BuildFailure

func (f byFirstFailure) Len() int      { return len(f) }
func (f byFirstFailure) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f byFirstFailure) Less(i, j int) bool {
	return f[i].Builders[0].FirstFailure < f[j].Builders[0].FirstFailure
}

type byTests []*messages.BuildFailure

func (f byTests) Len() int      { return len(f) }
func (f byTests) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f byTests) Less(i, j int) bool {
	iTests, jTests := []string{}, []string{}
	for _, t := range f[i].Reason.Raw.(*BqFailure).Tests {
		iTests = append(iTests, t.TestName)
	}
	for _, t := range f[j].Reason.Raw.(*BqFailure).Tests {
		jTests = append(jTests, t.TestName)
	}

	return strings.Join(iTests, "\n") < strings.Join(jTests, "\n")
}

func TestFilterHierarchicalSteps(t *testing.T) {
	Convey("smoke", t, func() {
		failures := []*messages.BuildFailure{}
		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 0)
	})

	Convey("single step, single builder", t, func() {
		failures := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results",
					},
				},
			},
		}

		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 1)
		So(len(got[0].Builders), ShouldEqual, 1)
	})

	Convey("nested step, single builder", t, func() {
		failures := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results|chromeos.postsubmit.beaglebone_servo-postsubmit",
					},
				},
			},
		}

		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 1)
		So(len(got[0].Builders), ShouldEqual, 1)
	})

	Convey("single step, multiple builders", t, func() {
		failures := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results",
					},
				},
			},
		}

		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 1)
		So(len(got[0].Builders), ShouldEqual, 2)
	})

	Convey("nested step, multiple builder", t, func() {
		failures := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results|chromeos.postsubmit.beaglebone_servo-postsubmit",
					},
				},
			},
		}

		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 1)
		So(len(got[0].Builders), ShouldEqual, 2)
		So(got[0].StepAtFault.Step.Name, ShouldEqual, "check build results|build results|chromeos.postsubmit.beaglebone_servo-postsubmit")
	})

	Convey("mixed nested steps, multiple builder", t, func() {
		failures := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "test foo",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "test bar",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "test baz",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results",
					},
				},
			},
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name A",
					},
					{
						Project: "project",
						Bucket:  "bucket",
						Name:    "builder name B",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "check build results|build results|chromeos.postsubmit.beaglebone_servo-postsubmit",
					},
				},
			},
		}

		got := filterHierarchicalSteps(failures)
		So(len(got), ShouldEqual, 4)
		So(len(got[0].Builders), ShouldEqual, 2)
		So(got[0].StepAtFault.Step.Name, ShouldEqual, "test foo")
		So(len(got[1].Builders), ShouldEqual, 1)
		So(got[1].StepAtFault.Step.Name, ShouldEqual, "test bar")
		So(len(got[2].Builders), ShouldEqual, 1)
		So(got[2].StepAtFault.Step.Name, ShouldEqual, "test baz")
		So(len(got[3].Builders), ShouldEqual, 2)
		So(got[3].StepAtFault.Step.Name, ShouldEqual, "check build results|build results|chromeos.postsubmit.beaglebone_servo-postsubmit")
	})
}

func TestSliceContains(t *testing.T) {
	Convey("slice contains", t, func() {
		haystack := []string{"a", "b", "c"}
		So(sliceContains(haystack, "a"), ShouldBeTrue)
		So(sliceContains(haystack, "b"), ShouldBeTrue)
		So(sliceContains(haystack, "c"), ShouldBeTrue)
		So(sliceContains(haystack, "d"), ShouldBeFalse)
	})
}

func TestZipUnzipData(t *testing.T) {
	Convey("zip and unzip data", t, func() {
		data := []byte("abcdef")
		zippedData, err := zipData(data)
		So(err, ShouldBeNil)
		unzippedData, err := unzipData(zippedData)
		So(err, ShouldBeNil)
		So(unzippedData, ShouldResemble, data)
	})
}

func TestGetFilterFuncForTree(t *testing.T) {
	Convey("get filter func for tree", t, func() {
		_, err := getFilterFuncForTree("android")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("chromium")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("chromium.gpu")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("chromium.perf")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("ios")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("chrome_browser_release")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("chromium.clang")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("dawn")
		So(err, ShouldBeNil)
		_, err = getFilterFuncForTree("another")
		So(err, ShouldNotBeNil)
	})
}

func makeTestFailures(uniquifiers ...string) []*TestFailure {
	failures := make([]*TestFailure, 0, len(uniquifiers))
	for _, u := range uniquifiers {
		failures = append(failures, makeTestFailure(u))
	}
	return failures
}

func makeTestFailure(uniquifier string) *TestFailure {
	return &TestFailure{
		TestName:    bigquery.NullString{StringVal: fmt.Sprintf("some/test/%s", uniquifier), Valid: true},
		TestID:      bigquery.NullString{StringVal: fmt.Sprintf("ninja://some/test/%s", uniquifier), Valid: true},
		Realm:       bigquery.NullString{StringVal: fmt.Sprintf("chromium:%s", uniquifier), Valid: true},
		VariantHash: bigquery.NullString{StringVal: fmt.Sprintf("1234%s", uniquifier), Valid: true},
		ClusterName: bigquery.NullString{StringVal: fmt.Sprintf("reason-v3/%s", uniquifier), Valid: true},
		RefHash:     bigquery.NullString{StringVal: fmt.Sprintf("ref-hash/%s", uniquifier), Valid: true},
	}
}

func makeTestWithResults(uniquifiers ...string) []step.TestWithResult {
	failures := make([]step.TestWithResult, 0, len(uniquifiers))
	for _, u := range uniquifiers {
		failures = append(failures, makeTestWithResult(u))
	}
	return failures
}

func makeTestWithResult(uniquifier string) step.TestWithResult {
	return step.TestWithResult{
		TestName:    fmt.Sprintf("some/test/%s", uniquifier),
		TestID:      fmt.Sprintf("ninja://some/test/%s", uniquifier),
		VariantHash: fmt.Sprintf("1234%s", uniquifier),
		ClusterName: fmt.Sprintf("reason-v3/%s", uniquifier),
		Realm:       fmt.Sprintf("chromium:%s", uniquifier),
		RefHash:     fmt.Sprintf("ref-hash/%s", uniquifier),
	}
}
