// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testvariantupdator

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/resultdb/pbutil"
	"go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"

	"infra/appengine/weetbix/internal/analyzedtestvariants"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/internal/testutil/insert"
	pb "infra/appengine/weetbix/proto/v1"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/clock"
	. "go.chromium.org/luci/common/testing/assertions"
)

func init() {
	RegisterTaskClass()
}

func TestSchedule(t *testing.T) {
	Convey(`TestSchedule`, t, func() {
		ctx, skdr := tq.TestingContext(testutil.SpannerTestContext(t), nil)

		realm := "chromium:ci"
		testID := "ninja://test"
		variantHash := "deadbeef"
		now := clock.Now(ctx)
		task := &taskspb.UpdateTestVariant{
			TestVariantKey: &taskspb.TestVariantKey{
				Realm:       realm,
				TestId:      testID,
				VariantHash: variantHash,
			},
			EnqueueTime: pbutil.MustTimestampProto(now),
		}
		_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
			Schedule(ctx, realm, testID, variantHash, now)
			return nil
		})
		So(err, ShouldBeNil)
		So(skdr.Tasks().Payloads()[0], ShouldResembleProto, task)
	})
}

func TestCheckTask(t *testing.T) {
	Convey(`checkTask`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		realm := "chromium:ci"
		tID := "ninja://test"
		vh := "varianthash"
		now := clock.Now(ctx)
		ms := []*spanner.Mutation{
			insert.AnalyzedTestVariant(realm, tID, vh, pb.AnalyzedTestVariantStatus_CONSISTENTLY_EXPECTED,
				map[string]interface{}{
					"NextUpdateTaskEnqueueTime": now,
				}),
			insert.AnalyzedTestVariant(realm, "anothertest", vh, pb.AnalyzedTestVariantStatus_CONSISTENTLY_EXPECTED, nil),
		}
		testutil.MustApply(ctx, ms...)

		task := &taskspb.UpdateTestVariant{
			TestVariantKey: &taskspb.TestVariantKey{
				Realm:       realm,
				TestId:      tID,
				VariantHash: vh,
			},
		}
		Convey(`match`, func() {
			task.EnqueueTime = pbutil.MustTimestampProto(now)
			_, err := checkTask(span.Single(ctx), task)
			So(err, ShouldBeNil)
		})

		Convey(`mismatch`, func() {
			anotherTime := now.Add(time.Hour)
			task.EnqueueTime = pbutil.MustTimestampProto(anotherTime)
			_, err := checkTask(span.Single(ctx), task)
			So(err, ShouldEqual, errUnknownTask)
		})

		Convey(`no schedule`, func() {
			task.TestVariantKey.TestId = "anothertest"
			_, err := checkTask(span.Single(ctx), task)
			So(err, ShouldEqual, errShouldNotSchedule)
		})
	})
}

func TestUpdateTestVariantStatus(t *testing.T) {
	Convey(`updateTestVariant`, t, func() {
		ctx, skdr := tq.TestingContext(testutil.SpannerTestContext(t), nil)
		realm := "chromium:ci"
		tID := "ninja://test"
		vh := "varianthash"
		now := clock.Now(ctx)
		ms := []*spanner.Mutation{
			insert.AnalyzedTestVariant(realm, tID, vh, pb.AnalyzedTestVariantStatus_CONSISTENTLY_EXPECTED, map[string]interface{}{
				"NextUpdateTaskEnqueueTime": now,
			}),
			insert.Verdict(realm, tID, vh, "build-1", pb.VerdictStatus_VERDICT_FLAKY, clock.Now(ctx).UTC().Add(-2*time.Hour), nil),
		}
		testutil.MustApply(ctx, ms...)

		task := &taskspb.UpdateTestVariant{
			TestVariantKey: &taskspb.TestVariantKey{
				Realm:       realm,
				TestId:      tID,
				VariantHash: vh,
			},
			EnqueueTime: pbutil.MustTimestampProto(now),
		}
		err := updateTestVariant(ctx, task)
		So(err, ShouldBeNil)

		// Read the test variant to confirm the updates.
		var status pb.AnalyzedTestVariantStatus
		var enqTime spanner.NullTime
		err = analyzedtestvariants.Read(span.Single(ctx), spanner.KeySets(spanner.Key{realm, tID, vh}), func(atv *pb.AnalyzedTestVariant, t spanner.NullTime) error {
			status = atv.Status
			enqTime = t
			return nil
		})
		So(err, ShouldBeNil)
		So(status, ShouldEqual, pb.AnalyzedTestVariantStatus_FLAKY)
		So(len(skdr.Tasks().Payloads()), ShouldEqual, 1)
		nextTask := skdr.Tasks().Payloads()[0].(*taskspb.UpdateTestVariant)
		So(pbutil.MustTimestamp(nextTask.EnqueueTime), ShouldEqual, enqTime.Time)

	})
}
