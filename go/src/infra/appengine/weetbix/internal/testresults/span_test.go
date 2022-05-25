// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testresults

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestReadTestHistory(t *testing.T) {
	Convey("ReadTestHistory", t, func() {
		ctx := testutil.SpannerTestContext(t)

		referenceTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)

		var1 := pbutil.Variant("key1", "val1", "key2", "val1")
		var2 := pbutil.Variant("key1", "val2", "key2", "val1")
		var3 := pbutil.Variant("key1", "val2", "key2", "val2")
		var4 := pbutil.Variant("key1", "val1", "key2", "val2")

		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			insertTVR := func(subRealm string, variant *pb.Variant) {
				span.BufferWrite(ctx, (&TestVariantRealm{
					Project:     "project",
					TestID:      "test_id",
					SubRealm:    subRealm,
					Variant:     variant,
					VariantHash: pbutil.VariantHash(variant),
				}).SaveUnverified())
			}

			insertTVR("realm", var1)
			insertTVR("realm", var2)
			insertTVR("realm", var3)
			insertTVR("realm2", var4)

			insertTV := func(partitionTime time.Time, variant *pb.Variant, invId string, status pb.TestVerdictStatus, hasUnsubmittedChanges bool, avgDuration *time.Duration) {
				baseTestResult := NewTestResult().
					WithProject("project").
					WithTestID("test_id").
					WithVariantHash(pbutil.VariantHash(variant)).
					WithPartitionTime(partitionTime).
					WithIngestedInvocationID(invId).
					WithSubRealm("realm").
					WithStatus(pb.TestResultStatus_PASS)
				if hasUnsubmittedChanges {
					baseTestResult = baseTestResult.WithChangelists([]Changelist{
						{
							Host:     "mygerrit",
							Change:   4321,
							Patchset: 5,
						},
						{
							Host:     "anothergerrit",
							Change:   5471,
							Patchset: 6,
						},
					})
				} else {
					baseTestResult = baseTestResult.WithChangelists(nil)
				}

				trs := NewTestVerdict().
					WithBaseTestResult(baseTestResult.Build()).
					WithStatus(status).
					WithPassedAvgDuration(avgDuration).
					Build()
				for _, tr := range trs {
					span.BufferWrite(ctx, tr.SaveUnverified())
				}
			}

			insertTV(referenceTime.Add(-1*time.Hour), var1, "inv1", pb.TestVerdictStatus_EXPECTED, false, newDuration(11111*time.Microsecond))
			insertTV(referenceTime.Add(-1*time.Hour), var1, "inv2", pb.TestVerdictStatus_EXONERATED, false, newDuration(1234567890123456*time.Microsecond))
			insertTV(referenceTime.Add(-1*time.Hour), var2, "inv1", pb.TestVerdictStatus_FLAKY, false, nil)

			insertTV(referenceTime.Add(-2*time.Hour), var1, "inv1", pb.TestVerdictStatus_UNEXPECTED, false, newDuration(33333*time.Microsecond))
			insertTV(referenceTime.Add(-2*time.Hour), var1, "inv2", pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED, true, nil)
			insertTV(referenceTime.Add(-2*time.Hour), var2, "inv1", pb.TestVerdictStatus_EXPECTED, true, nil)

			insertTV(referenceTime.Add(-3*time.Hour), var3, "inv1", pb.TestVerdictStatus_EXONERATED, true, newDuration(88888*time.Microsecond))

			return nil
		})
		So(err, ShouldBeNil)

		opts := ReadTestHistoryOptions{
			Project: "project",
			TestID:  "test_id",
		}

		Convey("pagination works", func() {
			opts.PageSize = 5
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: durationpb.New(11111 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv2",
					Status:            pb.TestVerdictStatus_EXONERATED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: durationpb.New(1234567890123456 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_FLAKY,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_UNEXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: durationpb.New(33333 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv2",
					Status:            pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
			})

			opts.PageToken = nextPageToken
			verdicts, nextPageToken, err = ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var3),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXONERATED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-3 * time.Hour)),
					PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
				},
			})
		})

		Convey("with partition_time_range", func() {
			opts.TimeRange = &pb.TimeRange{
				// Inclusive.
				Earliest: timestamppb.New(referenceTime.Add(-2 * time.Hour)),
				// Exclusive.
				Latest: timestamppb.New(referenceTime.Add(-1 * time.Hour)),
			}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_UNEXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: durationpb.New(33333 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv2",
					Status:            pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
			})
		})

		Convey("with contains variant_predicate", func() {
			Convey("with single key-value pair", func() {
				opts.VariantPredicate = &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2"),
					},
				}
				verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
					{
						TestId:            "test_id",
						VariantHash:       pbutil.VariantHash(var2),
						InvocationId:      "inv1",
						Status:            pb.TestVerdictStatus_FLAKY,
						PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
						PassedAvgDuration: nil,
					},
					{
						TestId:            "test_id",
						VariantHash:       pbutil.VariantHash(var2),
						InvocationId:      "inv1",
						Status:            pb.TestVerdictStatus_EXPECTED,
						PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
						PassedAvgDuration: nil,
					},
					{
						TestId:            "test_id",
						VariantHash:       pbutil.VariantHash(var3),
						InvocationId:      "inv1",
						Status:            pb.TestVerdictStatus_EXONERATED,
						PartitionTime:     timestamppb.New(referenceTime.Add(-3 * time.Hour)),
						PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
					},
				})
			})

			Convey("with multiple key-value pairs", func() {
				opts.VariantPredicate = &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2", "key2", "val2"),
					},
				}
				verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
					{
						TestId:            "test_id",
						VariantHash:       pbutil.VariantHash(var3),
						InvocationId:      "inv1",
						Status:            pb.TestVerdictStatus_EXONERATED,
						PartitionTime:     timestamppb.New(referenceTime.Add(-3 * time.Hour)),
						PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
					},
				})
			})
		})

		Convey("with equals variant_predicate", func() {
			opts.VariantPredicate = &pb.VariantPredicate{
				Predicate: &pb.VariantPredicate_Equals{
					Equals: var2,
				},
			}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_FLAKY,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
			})
		})

		Convey("with submitted_filter", func() {
			opts.SubmittedFilter = pb.SubmittedFilter_ONLY_UNSUBMITTED
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv2",
					Status:            pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var3),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXONERATED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-3 * time.Hour)),
					PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
				},
			})

			opts.SubmittedFilter = pb.SubmittedFilter_ONLY_SUBMITTED
			verdicts, nextPageToken, err = ReadTestHistory(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_EXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: durationpb.New(11111 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv2",
					Status:            pb.TestVerdictStatus_EXONERATED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: durationpb.New(1234567890123456 * time.Microsecond),
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var2),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_FLAKY,
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					PassedAvgDuration: nil,
				},
				{
					TestId:            "test_id",
					VariantHash:       pbutil.VariantHash(var1),
					InvocationId:      "inv1",
					Status:            pb.TestVerdictStatus_UNEXPECTED,
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					PassedAvgDuration: durationpb.New(33333 * time.Microsecond),
				},
			})
		})
	})
}

func newDuration(value time.Duration) *time.Duration {
	d := new(time.Duration)
	*d = value
	return d
}

func TestReadTestHistoryStats(t *testing.T) {
	Convey("ReadTestHistoryStats", t, func() {
		ctx := testutil.SpannerTestContext(t)

		referenceTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)

		day := 24 * time.Hour

		var1 := pbutil.Variant("key1", "val1", "key2", "val1")
		var2 := pbutil.Variant("key1", "val2", "key2", "val1")
		var3 := pbutil.Variant("key1", "val2", "key2", "val2")
		var4 := pbutil.Variant("key1", "val1", "key2", "val2")

		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			insertTVR := func(subRealm string, variant *pb.Variant) {
				span.BufferWrite(ctx, (&TestVariantRealm{
					Project:     "project",
					TestID:      "test_id",
					SubRealm:    subRealm,
					Variant:     variant,
					VariantHash: pbutil.VariantHash(variant),
				}).SaveUnverified())
			}

			insertTVR("realm", var1)
			insertTVR("realm", var2)
			insertTVR("realm", var3)
			insertTVR("realm2", var4)

			insertTV := func(partitionTime time.Time, variant *pb.Variant, invId string, status pb.TestVerdictStatus, hasUnsubmittedChanges bool, avgDuration *time.Duration) {
				baseTestResult := NewTestResult().
					WithProject("project").
					WithTestID("test_id").
					WithVariantHash(pbutil.VariantHash(variant)).
					WithPartitionTime(partitionTime).
					WithIngestedInvocationID(invId).
					WithSubRealm("realm").
					WithStatus(pb.TestResultStatus_PASS)
				if hasUnsubmittedChanges {
					baseTestResult = baseTestResult.WithChangelists([]Changelist{
						{
							Host:     "mygerrit",
							Change:   4321,
							Patchset: 5,
						},
					})
				} else {
					baseTestResult = baseTestResult.WithChangelists(nil)
				}

				trs := NewTestVerdict().
					WithBaseTestResult(baseTestResult.Build()).
					WithStatus(status).
					WithPassedAvgDuration(avgDuration).
					Build()
				for _, tr := range trs {
					span.BufferWrite(ctx, tr.SaveUnverified())
				}
			}

			insertTV(referenceTime.Add(-1*time.Hour), var1, "inv1", pb.TestVerdictStatus_EXPECTED, false, newDuration(22222*time.Microsecond))
			insertTV(referenceTime.Add(-12*time.Hour), var1, "inv2", pb.TestVerdictStatus_EXONERATED, false, newDuration(1234567890123456*time.Microsecond))
			insertTV(referenceTime.Add(-24*time.Hour), var2, "inv1", pb.TestVerdictStatus_FLAKY, false, nil)

			insertTV(referenceTime.Add(-day-1*time.Hour), var1, "inv1", pb.TestVerdictStatus_UNEXPECTED, false, newDuration(33333*time.Microsecond))
			insertTV(referenceTime.Add(-day-12*time.Hour), var1, "inv2", pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED, true, nil)
			insertTV(referenceTime.Add(-day-24*time.Hour), var2, "inv1", pb.TestVerdictStatus_EXPECTED, true, nil)

			insertTV(referenceTime.Add(-2*day-3*time.Hour), var3, "inv1", pb.TestVerdictStatus_EXONERATED, true, newDuration(88888*time.Microsecond))
			return nil
		})
		So(err, ShouldBeNil)

		opts := ReadTestHistoryOptions{
			Project: "project",
			TestID:  "test_id",
		}

		Convey("pagination works", func() {
			opts.PageSize = 3
			verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
					VariantHash:       pbutil.VariantHash(var1),
					ExpectedCount:     1,
					ExoneratedCount:   1,
					PassedAvgDuration: durationpb.New(((22222 + 1234567890123456) / 2) * time.Microsecond),
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					FlakyCount:        1,
					PassedAvgDuration: nil,
				},
				{
					PartitionTime:            timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:              pbutil.VariantHash(var1),
					UnexpectedCount:          1,
					UnexpectedlySkippedCount: 1,
					PassedAvgDuration:        durationpb.New(33333 * time.Microsecond),
				},
			})

			opts.PageToken = nextPageToken
			verdicts, nextPageToken, err = ReadTestHistoryStats(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					ExpectedCount:     1,
					PassedAvgDuration: nil,
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-3 * day)),
					VariantHash:       pbutil.VariantHash(var3),
					ExoneratedCount:   1,
					PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
				},
			})
		})

		Convey("with partition_time_range", func() {
			Convey("day boundaries", func() {
				opts.TimeRange = &pb.TimeRange{
					// Inclusive.
					Earliest: timestamppb.New(referenceTime.Add(-2 * day)),
					// Exclusive.
					Latest: timestamppb.New(referenceTime.Add(-1 * day)),
				}
				verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
					{
						PartitionTime:            timestamppb.New(referenceTime.Add(-2 * day)),
						VariantHash:              pbutil.VariantHash(var1),
						UnexpectedCount:          1,
						UnexpectedlySkippedCount: 1,
						PassedAvgDuration:        durationpb.New(33333 * time.Microsecond),
					},
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
						VariantHash:       pbutil.VariantHash(var2),
						ExpectedCount:     1,
						PassedAvgDuration: nil,
					},
				})
			})
			Convey("part-day boundaries", func() {
				opts.TimeRange = &pb.TimeRange{
					// Inclusive.
					Earliest: timestamppb.New(referenceTime.Add(-2*day - 3*time.Hour)),
					// Exclusive.
					Latest: timestamppb.New(referenceTime.Add(-1*day - 1*time.Hour)),
				}
				verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
					{
						PartitionTime:            timestamppb.New(referenceTime.Add(-2 * day)),
						VariantHash:              pbutil.VariantHash(var1),
						UnexpectedlySkippedCount: 1,
						PassedAvgDuration:        nil,
					},
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
						VariantHash:       pbutil.VariantHash(var2),
						ExpectedCount:     1,
						PassedAvgDuration: nil,
					},
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-3 * day)),
						VariantHash:       pbutil.VariantHash(var3),
						ExoneratedCount:   1,
						PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
					},
				})
			})
		})

		Convey("with contains variant_predicate", func() {
			Convey("with single key-value pair", func() {
				opts.VariantPredicate = &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2"),
					},
				}
				verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
						VariantHash:       pbutil.VariantHash(var2),
						FlakyCount:        1,
						PassedAvgDuration: nil,
					},
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
						VariantHash:       pbutil.VariantHash(var2),
						ExpectedCount:     1,
						PassedAvgDuration: nil,
					},
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-3 * day)),
						VariantHash:       pbutil.VariantHash(var3),
						ExoneratedCount:   1,
						PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
					},
				})
			})

			Convey("with multiple key-value pairs", func() {
				opts.VariantPredicate = &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2", "key2", "val2"),
					},
				}
				verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
					{
						PartitionTime:     timestamppb.New(referenceTime.Add(-3 * day)),
						VariantHash:       pbutil.VariantHash(var3),
						ExoneratedCount:   1,
						PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
					},
				})
			})
		})

		Convey("with equals variant_predicate", func() {
			opts.VariantPredicate = &pb.VariantPredicate{
				Predicate: &pb.VariantPredicate_Equals{
					Equals: var2,
				},
			}
			verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					FlakyCount:        1,
					PassedAvgDuration: nil,
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					ExpectedCount:     1,
					PassedAvgDuration: nil,
				},
			})
		})

		Convey("with submitted_filter", func() {
			opts.SubmittedFilter = pb.SubmittedFilter_ONLY_UNSUBMITTED
			verdicts, nextPageToken, err := ReadTestHistoryStats(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
				{
					PartitionTime:            timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:              pbutil.VariantHash(var1),
					UnexpectedlySkippedCount: 1,
					PassedAvgDuration:        nil,
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					ExpectedCount:     1,
					PassedAvgDuration: nil,
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-3 * day)),
					VariantHash:       pbutil.VariantHash(var3),
					ExoneratedCount:   1,
					PassedAvgDuration: durationpb.New(88888 * time.Microsecond),
				},
			})

			opts.SubmittedFilter = pb.SubmittedFilter_ONLY_SUBMITTED
			verdicts, nextPageToken, err = ReadTestHistoryStats(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.QueryTestHistoryStatsResponse_Group{
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
					VariantHash:       pbutil.VariantHash(var1),
					ExpectedCount:     1,
					ExoneratedCount:   1,
					PassedAvgDuration: durationpb.New(((22222 + 1234567890123456) / 2) * time.Microsecond),
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-1 * day)),
					VariantHash:       pbutil.VariantHash(var2),
					FlakyCount:        1,
					PassedAvgDuration: nil,
				},
				{
					PartitionTime:     timestamppb.New(referenceTime.Add(-2 * day)),
					VariantHash:       pbutil.VariantHash(var1),
					UnexpectedCount:   1,
					PassedAvgDuration: durationpb.New(33333 * time.Microsecond),
				},
			})
		})
	})
}

func TestReadVariants(t *testing.T) {
	Convey("ReadVariants", t, func() {
		ctx := testutil.SpannerTestContext(t)

		var1 := pbutil.Variant("key1", "val1", "key2", "val1")
		var2 := pbutil.Variant("key1", "val2", "key2", "val1")
		var3 := pbutil.Variant("key1", "val2", "key2", "val2")
		var4 := pbutil.Variant("key1", "val1", "key2", "val2")

		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			insertTVR := func(subRealm string, variant *pb.Variant) {
				span.BufferWrite(ctx, (&TestVariantRealm{
					Project:     "project",
					TestID:      "test_id",
					SubRealm:    subRealm,
					Variant:     variant,
					VariantHash: pbutil.VariantHash(variant),
				}).SaveUnverified())
			}

			insertTVR("realm1", var1)
			insertTVR("realm1", var2)

			insertTVR("realm2", var2)
			insertTVR("realm2", var3)

			insertTVR("realm3", var4)

			return nil
		})
		So(err, ShouldBeNil)

		Convey("pagination works", func() {
			opts := ReadVariantsOptions{PageSize: 3}
			variants, nextPageToken, err := ReadVariants(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(variants, ShouldResembleProto, []*pb.QueryVariantsResponse_VariantInfo{
				{
					VariantHash: pbutil.VariantHash(var1),
					Variant:     var1,
				},
				{
					VariantHash: pbutil.VariantHash(var3),
					Variant:     var3,
				},
				{
					VariantHash: pbutil.VariantHash(var4),
					Variant:     var4,
				},
			})

			opts.PageToken = nextPageToken
			variants, nextPageToken, err = ReadVariants(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(variants, ShouldResembleProto, []*pb.QueryVariantsResponse_VariantInfo{
				{
					VariantHash: pbutil.VariantHash(var2),
					Variant:     var2,
				},
			})
		})

		Convey("multi-realm works", func() {
			opts := ReadVariantsOptions{SubRealms: []string{"realm1", "realm2"}}
			variants, nextPageToken, err := ReadVariants(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(variants, ShouldResembleProto, []*pb.QueryVariantsResponse_VariantInfo{
				{
					VariantHash: pbutil.VariantHash(var1),
					Variant:     var1,
				},
				{
					VariantHash: pbutil.VariantHash(var3),
					Variant:     var3,
				},
				{
					VariantHash: pbutil.VariantHash(var2),
					Variant:     var2,
				},
			})
		})

		Convey("single-realm works", func() {
			opts := ReadVariantsOptions{SubRealms: []string{"realm2"}}
			variants, nextPageToken, err := ReadVariants(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(variants, ShouldResembleProto, []*pb.QueryVariantsResponse_VariantInfo{
				{
					VariantHash: pbutil.VariantHash(var3),
					Variant:     var3,
				},
				{
					VariantHash: pbutil.VariantHash(var2),
					Variant:     var2,
				},
			})
		})
	})
}
