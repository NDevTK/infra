// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	. "go.chromium.org/luci/common/testing/assertions"
	rdbpbutil "go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"
	_ "go.chromium.org/luci/server/tq/txn/spanner"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/buildbucket"
	"infra/appengine/weetbix/internal/ingestion/control"
	ctrlpb "infra/appengine/weetbix/internal/ingestion/control/proto"
	"infra/appengine/weetbix/internal/resultdb"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/pbutil"
	weetbixpb "infra/appengine/weetbix/proto/v1"
)

func TestSchedule(t *testing.T) {
	Convey(`TestSchedule`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		ctx, skdr := tq.TestingContext(ctx, nil)

		task := &taskspb.IngestTestVerdicts{
			Build:         &ctrlpb.BuildResult{},
			PartitionTime: timestamppb.New(time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)),
		}
		expected := proto.Clone(task).(*taskspb.IngestTestVerdicts)

		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			Schedule(ctx, task)
			return nil
		})
		So(err, ShouldBeNil)
		So(skdr.Tasks().Payloads()[0], ShouldResembleProto, expected)
	})
}

func TestIngestTestVerdicts(t *testing.T) {
	Convey(`TestIngestTestVerdicts`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		ctx, skdr := tq.TestingContext(ctx, nil)

		Convey(`partition time`, func() {
			payload := &taskspb.IngestTestVerdicts{
				Build: &ctrlpb.BuildResult{
					Host: "host",
					Id:   13131313,
				},
				PartitionTime: timestamppb.New(clock.Now(ctx).Add(-1 * time.Hour)),
			}
			Convey(`too early`, func() {
				payload.PartitionTime = timestamppb.New(clock.Now(ctx).Add(25 * time.Hour))
				err := ingestTestResults(ctx, payload)
				So(err, ShouldErrLike, "too far in the future")
			})
			Convey(`too late`, func() {
				payload.PartitionTime = timestamppb.New(clock.Now(ctx).Add(-91 * 24 * time.Hour))
				err := ingestTestResults(ctx, payload)
				So(err, ShouldErrLike, "too long ago")
			})
		})

		partitionTime := clock.Now(ctx).Add(-1 * time.Hour)
		verifyTestResults := func() {
			trBuilder := testresults.NewTestResult().
				WithProject("chromium").
				WithPartitionTime(timestamppb.New(partitionTime).AsTime()).
				WithIngestedInvocationID("build-87654321").
				WithSubRealm("ci").
				WithBuildStatus(weetbixpb.BuildStatus_BUILD_STATUS_FAILURE).
				WithChangelists([]testresults.Changelist{
					{
						Host:     "mygerrit",
						Change:   12345,
						Patchset: 5,
					},
					{
						Host:     "anothergerrit",
						Change:   77788,
						Patchset: 19,
					},
				}).
				WithHasContributedToClSubmission(true)
			expectedTRs := []*testresults.TestResult{
				trBuilder.WithTestID("test_id_1").
					WithVariantHash("hash_1").
					WithRunIndex(0).
					WithResultIndex(0).
					WithIsUnexpected(true).
					WithStatus(weetbixpb.TestResultStatus_FAIL).
					WithRunDuration(10 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_1").
					WithVariantHash("hash_2").
					WithRunIndex(0).
					WithResultIndex(0).
					WithIsUnexpected(false).
					WithStatus(weetbixpb.TestResultStatus_PASS).
					WithRunDuration(time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_1").
					WithVariantHash("hash_2").
					WithRunIndex(0).
					WithResultIndex(1).
					WithIsUnexpected(true).
					WithStatus(weetbixpb.TestResultStatus_FAIL).
					WithRunDuration(10 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_1").
					WithRunIndex(0).
					WithResultIndex(0).
					WithIsUnexpected(false).
					WithStatus(weetbixpb.TestResultStatus_PASS).
					WithRunDuration(3 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_1").
					WithRunIndex(0).
					WithResultIndex(1).
					WithIsUnexpected(false).
					WithStatus(weetbixpb.TestResultStatus_PASS).
					WithRunDuration(time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_1").
					WithRunIndex(1).
					WithResultIndex(0).
					WithIsUnexpected(true).
					WithStatus(weetbixpb.TestResultStatus_FAIL).
					WithRunDuration(10 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_2").
					WithRunIndex(0).
					WithResultIndex(0).
					WithIsUnexpected(true).
					WithStatus(weetbixpb.TestResultStatus_FAIL).
					WithRunDuration(10 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_2").
					WithRunIndex(1).
					WithResultIndex(0).
					WithIsUnexpected(false).
					WithStatus(weetbixpb.TestResultStatus_PASS).
					WithRunDuration(time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
				trBuilder.WithTestID("test_id_2").
					WithVariantHash("hash_2").
					WithRunIndex(1).
					WithResultIndex(1).
					WithIsUnexpected(false).
					WithStatus(weetbixpb.TestResultStatus_PASS).
					WithRunDuration(2 * time.Second).
					WithExonerationStatus(weetbixpb.ExonerationStatus_NOT_EXONERATED).
					Build(),
			}

			// Validate TestResults table is populated.
			var actualTRs []*testresults.TestResult
			err := testresults.ReadTestResults(span.Single(ctx), spanner.AllKeys(), func(tr *testresults.TestResult) error {
				actualTRs = append(actualTRs, tr)
				return nil
			})
			So(err, ShouldBeNil)

			So(actualTRs, ShouldResemble, expectedTRs)

			// Validate TestVariantRealms table is populated.
			tvrs := make([]*testresults.TestVariantRealm, 0)
			err = testresults.ReadTestVariantRealms(span.Single(ctx), spanner.AllKeys(), func(tvr *testresults.TestVariantRealm) error {
				tvrs = append(tvrs, tvr)
				return nil
			})
			So(err, ShouldBeNil)
			So(tvrs, ShouldHaveLength, 4)
			So(tvrs[0].LastIngestionTime, ShouldNotBeZeroValue)
			So(tvrs[1].LastIngestionTime, ShouldNotBeZeroValue)
			So(tvrs[2].LastIngestionTime, ShouldNotBeZeroValue)
			So(tvrs[3].LastIngestionTime, ShouldNotBeZeroValue)
			So(tvrs, ShouldResemble, []*testresults.TestVariantRealm{
				{
					Project:           "chromium",
					TestID:            "test_id_1",
					VariantHash:       "hash_1",
					SubRealm:          "ci",
					Variant:           pbutil.VariantFromResultDB(rdbpbutil.Variant("k1", "v1")),
					LastIngestionTime: tvrs[0].LastIngestionTime,
				},
				{
					Project:           "chromium",
					TestID:            "test_id_1",
					VariantHash:       "hash_2",
					SubRealm:          "ci",
					Variant:           pbutil.VariantFromResultDB(rdbpbutil.Variant("k1", "v2")),
					LastIngestionTime: tvrs[1].LastIngestionTime,
				},
				{
					Project:           "chromium",
					TestID:            "test_id_2",
					VariantHash:       "hash_1",
					SubRealm:          "ci",
					Variant:           pbutil.VariantFromResultDB(rdbpbutil.Variant("k1", "v1")),
					LastIngestionTime: tvrs[2].LastIngestionTime,
				},
				{
					Project:           "chromium",
					TestID:            "test_id_2",
					VariantHash:       "hash_2",
					SubRealm:          "ci",
					Variant:           pbutil.VariantFromResultDB(rdbpbutil.Variant("k1", "v2")),
					LastIngestionTime: tvrs[3].LastIngestionTime,
				},
			})
		}
		verifyIngestedInvocation := func(expected *testresults.IngestedInvocation) {
			var invs []*testresults.IngestedInvocation
			// Validate IngestedInvocations table is populated.
			err := testresults.ReadIngestedInvocations(span.Single(ctx), spanner.AllKeys(), func(inv *testresults.IngestedInvocation) error {
				invs = append(invs, inv)
				return nil
			})
			So(err, ShouldBeNil)
			if expected == nil {
				So(invs, ShouldHaveLength, 0)
			} else {
				So(invs, ShouldHaveLength, 1)
				So(invs[0], ShouldResemble, expected)
			}
		}

		Convey(`valid payload`, func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			bHost := "cr-buildbucket-dev.appspot.com"
			bID := int64(87654321)
			inv := "invocations/build-87654321"
			realm := "chromium:ci"

			mrc := resultdb.NewMockedClient(ctx, ctl)
			mbc := buildbucket.NewMockedClient(mrc.Ctx, ctl)
			ctx = mbc.Ctx

			request := &bbpb.GetBuildRequest{
				Id: bID,
				Mask: &bbpb.BuildMask{
					Fields: &field_mask.FieldMask{
						Paths: []string{"input.gerrit_changes", "infra.resultdb", "status"},
					},
				},
			}
			mbc.GetBuild(request, mockedGetBuildRsp(inv))

			invReq := &rdbpb.GetInvocationRequest{
				Name: inv,
			}
			invRes := &rdbpb.Invocation{
				Name:  inv,
				Realm: realm,
			}
			mrc.GetInvocation(invReq, invRes)

			payload := &taskspb.IngestTestVerdicts{
				Build: &ctrlpb.BuildResult{
					Host:         bHost,
					Id:           bID,
					CreationTime: timestamppb.New(time.Date(2020, time.April, 1, 2, 3, 4, 5, time.UTC)),
				},
				PartitionTime: timestamppb.New(partitionTime),
				PresubmitRun: &ctrlpb.PresubmitResult{
					PresubmitRunId: &weetbixpb.PresubmitRunId{
						System: "luci-cv",
						Id:     "infra/12345",
					},
					PresubmitRunSucceeded: true,
					CreationTime:          timestamppb.New(time.Date(2021, time.April, 1, 2, 3, 4, 5, time.UTC)),
				},
				PageToken: "expected_token",
				PageIndex: 0,
			}

			ingestionRecord :=
				control.NewEntry(0).
					WithBuildID(control.BuildID(bHost, bID)).
					WithBuildResult(payload.Build).
					WithPresubmitResult(payload.PresubmitRun).
					WithTaskCount(1).
					Build()

			tvReq := &rdbpb.QueryTestVariantsRequest{
				Invocations: []string{inv},
				PageSize:    10000,
				ReadMask: &fieldmaskpb.FieldMask{
					Paths: []string{
						"test_id",
						"variant_hash",
						"status",
						"variant",
						"results.*.result.name",
						"results.*.result.start_time",
						"results.*.result.status",
						"results.*.result.expected",
						"results.*.result.duration",
					},
				},
				PageToken: "expected_token",
			}
			rsp := mockedQueryTestVariantsRsp()

			expectedInvocation := &testresults.IngestedInvocation{
				Project:                      "chromium",
				IngestedInvocationID:         "build-87654321",
				SubRealm:                     "ci",
				PartitionTime:                timestamppb.New(partitionTime).AsTime(),
				BuildStatus:                  weetbixpb.BuildStatus_BUILD_STATUS_FAILURE,
				HasContributedToClSubmission: true,
				Changelists: []testresults.Changelist{
					{
						Host:     "mygerrit",
						Change:   12345,
						Patchset: 5,
					},
					{
						Host:     "anothergerrit",
						Change:   77788,
						Patchset: 19,
					},
				},
			}

			Convey("First task", func() {
				rsp.NextPageToken = "continuation_token"
				mrc.QueryTestVariants(tvReq, rsp)

				_, err := control.SetEntriesForTesting(ctx, ingestionRecord)
				So(err, ShouldBeNil)

				// Run ingestion. Clone the input to avoid changes the
				// implementation may make to payload flowing back out.
				clonedPayload := proto.Clone(payload).(*taskspb.IngestTestVerdicts)
				err = ingestTestResults(ctx, clonedPayload)
				So(err, ShouldBeNil)

				// Ensure continuation task is created.
				So(skdr.Tasks().Payloads(), ShouldHaveLength, 1)
				task := skdr.Tasks().Payloads()[0].(*taskspb.IngestTestVerdicts)
				So(task, ShouldResembleProto, &taskspb.IngestTestVerdicts{
					Build:         payload.Build,
					PartitionTime: payload.PartitionTime,
					PresubmitRun:  payload.PresubmitRun,
					PageToken:     "continuation_token",
					PageIndex:     1,
				})

				verifyTestResults()
				verifyIngestedInvocation(expectedInvocation)
			})
			Convey("Final task", func() {
				ingestionRecord.TaskCount = 10
				payload.PageIndex = 9
				rsp.NextPageToken = "" // No more results.

				mrc.QueryTestVariants(tvReq, rsp)

				_, err := control.SetEntriesForTesting(ctx, ingestionRecord)
				So(err, ShouldBeNil)

				// Run ingestion.
				err = ingestTestResults(ctx, payload)
				So(err, ShouldBeNil)

				// No continuation task scheduled.
				So(skdr.Tasks().Payloads(), ShouldHaveLength, 0)

				verifyTestResults()
				// As the task was not the first task, ensure the invocation
				// record has not been created again.
				verifyIngestedInvocation(nil)
			})
			Convey("Retried task", func() {
				payload.PageIndex = 1
				ingestionRecord.TaskCount = 3
				rsp.NextPageToken = "continuation_token"

				mrc.QueryTestVariants(tvReq, rsp)

				_, err := control.SetEntriesForTesting(ctx, ingestionRecord)
				So(err, ShouldBeNil)

				// Run ingestion.
				err = ingestTestResults(ctx, payload)
				So(err, ShouldBeNil)

				// No continuation task scheduled.
				So(skdr.Tasks().Payloads(), ShouldHaveLength, 0)
			})
		})
	})
}
