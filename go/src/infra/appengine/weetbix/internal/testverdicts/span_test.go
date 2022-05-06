// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdicts

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestReadTestHistory(t *testing.T) {
	t.Parallel()

	Convey("ReadTestHistory", t, func() {
		ctx := testutil.SpannerTestContext(t)
		ctx, _ = testclock.UseTime(ctx, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC))

		now := clock.Now(ctx)

		var1 := pbutil.Variant("key1", "val1", "key2", "val1")
		var2 := pbutil.Variant("key1", "val2", "key2", "val1")
		var3 := pbutil.Variant("key1", "val2", "key2", "val2")
		var4 := pbutil.Variant("key1", "val1", "key2", "val2")

		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			insertTVR := func(subRealm string, variant *pb.Variant) {
				(&TestVariantRealm{
					Project:     "project",
					TestID:      "test_id",
					SubRealm:    subRealm,
					Variant:     variant,
					VariantHash: pbutil.VariantHash(variant),
				}).SaveUnverified(ctx)
			}

			insertTVR("realm", var1)
			insertTVR("realm", var2)
			insertTVR("realm", var3)
			insertTVR("realm2", var4)

			insertTV := func(partitionTime time.Time, variant *pb.Variant, invId string, hasUnsubmittedChanges bool) {
				(&TestVerdict{
					Project:               "project",
					TestID:                "test_id",
					SubRealm:              "realm",
					PartitionTime:         partitionTime,
					VariantHash:           pbutil.VariantHash(variant),
					IngestedInvocationID:  invId,
					HasUnsubmittedChanges: hasUnsubmittedChanges,
				}).SaveUnverified(ctx)
			}

			insertTV(now.Add(-1*time.Hour), var1, "inv1", false)
			insertTV(now.Add(-1*time.Hour), var1, "inv2", false)
			insertTV(now.Add(-1*time.Hour), var2, "inv1", false)

			insertTV(now.Add(-2*time.Hour), var1, "inv1", false)
			insertTV(now.Add(-2*time.Hour), var1, "inv2", true)
			insertTV(now.Add(-2*time.Hour), var2, "inv1", true)

			insertTV(now.Add(-3*time.Hour), var3, "inv1", true)

			return nil
		})
		So(err, ShouldBeNil)

		Convey("pagination works", func() {
			opts := ReadTestHistoryOptions{PageSize: 5}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldNotBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv2",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv2",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
			})

			opts.PageToken = nextPageToken
			verdicts, nextPageToken, err = ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var3),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-3 * time.Hour)),
				},
			})
		})

		Convey("with partition_time_range", func() {
			opts := ReadTestHistoryOptions{TimeRange: &pb.TimeRange{
				// Exclusive.
				Earliest: timestamppb.New(now.Add(-3 * time.Hour)),
				// Inclusive.
				Latest: timestamppb.New(now.Add(-2 * time.Hour)),
			}}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv2",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
			})
		})

		Convey("with contains variant_predicate", func() {
			Convey("with single key-value pair", func() {
				opts := ReadTestHistoryOptions{VariantPredicate: &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2"),
					},
				}}
				verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
					{
						TestId:        "test_id",
						VariantHash:   pbutil.VariantHash(var2),
						InvocationId:  "inv1",
						Status:        pb.TestVerdictStatus_EXPECTED,
						PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
					},
					{
						TestId:        "test_id",
						VariantHash:   pbutil.VariantHash(var2),
						InvocationId:  "inv1",
						Status:        pb.TestVerdictStatus_EXPECTED,
						PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
					},
					{
						TestId:        "test_id",
						VariantHash:   pbutil.VariantHash(var3),
						InvocationId:  "inv1",
						Status:        pb.TestVerdictStatus_EXPECTED,
						PartitionTime: timestamppb.New(now.Add(-3 * time.Hour)),
					},
				})
			})

			Convey("with multiple key-value pairs", func() {
				opts := ReadTestHistoryOptions{VariantPredicate: &pb.VariantPredicate{
					Predicate: &pb.VariantPredicate_Contains{
						Contains: pbutil.Variant("key1", "val2", "key2", "val2"),
					},
				}}
				verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
				So(err, ShouldBeNil)
				So(nextPageToken, ShouldBeEmpty)
				So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
					{
						TestId:        "test_id",
						VariantHash:   pbutil.VariantHash(var3),
						InvocationId:  "inv1",
						Status:        pb.TestVerdictStatus_EXPECTED,
						PartitionTime: timestamppb.New(now.Add(-3 * time.Hour)),
					},
				})
			})
		})

		Convey("with equals variant_predicate", func() {
			opts := ReadTestHistoryOptions{VariantPredicate: &pb.VariantPredicate{
				Predicate: &pb.VariantPredicate_Equals{
					Equals: var2,
				},
			}}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
			})
		})

		Convey("with submitted_filter", func() {
			opts := ReadTestHistoryOptions{SubmittedFilter: pb.SubmittedFilter_ONLY_UNSUBMITTED}
			verdicts, nextPageToken, err := ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv2",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var3),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-3 * time.Hour)),
				},
			})

			opts = ReadTestHistoryOptions{SubmittedFilter: pb.SubmittedFilter_ONLY_SUBMITTED}
			verdicts, nextPageToken, err = ReadTestHistory(span.Single(ctx), "project", "test_id", opts)
			So(err, ShouldBeNil)
			So(nextPageToken, ShouldBeEmpty)
			So(verdicts, ShouldResembleProto, []*pb.TestVerdict{
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv2",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var2),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-1 * time.Hour)),
				},
				{
					TestId:        "test_id",
					VariantHash:   pbutil.VariantHash(var1),
					InvocationId:  "inv1",
					Status:        pb.TestVerdictStatus_EXPECTED,
					PartitionTime: timestamppb.New(now.Add(-2 * time.Hour)),
				},
			})
		})
	})
}
