// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package analyzedtestvariants

import (
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"go.chromium.org/luci/common/clock"

	"infra/appengine/weetbix/internal"
	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/internal/testutil/insert"
	atvpb "infra/appengine/weetbix/proto/analyzedtestvariant"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPurge(t *testing.T) {
	Convey(`TestPurge`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		const realm = "chromium:ci"
		const tID1 = "ninja://test1"
		const tID2 = "ninja://test2"
		const tID3 = "ninja://test3"
		const tID4 = "ninja://test4"
		const tID5 = "ninja://test5"
		const vh = "varianthash"
		now := clock.Now(ctx)
		ms := []*spanner.Mutation{
			// Active flaky test variants are not deleted.
			insert.AnalyzedTestVariant(realm, tID1, vh, atvpb.Status_FLAKY, map[string]interface{}{
				"StatusUpdateTime": now.Add(-time.Hour),
			}),
			// Active flaky test variants are not deleted, even if it has been in the
			// status for a long time.
			insert.AnalyzedTestVariant(realm, tID2, vh, atvpb.Status_FLAKY, map[string]interface{}{
				"StatusUpdateTime": now.Add(-2 * 31 * 24 * time.Hour),
			}),
			// No new results, but was newly updated.
			insert.AnalyzedTestVariant(realm, tID3, vh, atvpb.Status_NO_NEW_RESULTS, map[string]interface{}{
				"StatusUpdateTime": now.Add(-time.Hour),
			}),
			// No new results for over a month, should delete.
			insert.AnalyzedTestVariant(realm, tID4, vh, atvpb.Status_NO_NEW_RESULTS, map[string]interface{}{
				"StatusUpdateTime": now.Add(-2 * 31 * 24 * time.Hour),
			}),
			// consistently expected for over a month, should delete.
			insert.AnalyzedTestVariant(realm, tID5, vh, atvpb.Status_CONSISTENTLY_EXPECTED, map[string]interface{}{
				"StatusUpdateTime": now.Add(-2 * 31 * 24 * time.Hour),
			}),
			insert.Verdict(realm, tID1, vh, "build-0", internal.VerdictStatus_EXPECTED, now.Add(-time.Hour), nil),
			insert.Verdict(realm, tID4, vh, "build-1", internal.VerdictStatus_VERDICT_FLAKY, now.Add(-5*30*24*time.Hour), nil),
			insert.Verdict(realm, tID4, vh, "build-2", internal.VerdictStatus_EXPECTED, now.Add(-2*30*24*time.Hour), nil),
			insert.Verdict(realm, tID5, vh, "build-1", internal.VerdictStatus_EXPECTED, now.Add(-2*30*24*time.Hour), nil),
			insert.Verdict(realm, tID5, vh, "build-2", internal.VerdictStatus_VERDICT_FLAKY, now.Add(-5*24*time.Hour), nil),
		}
		testutil.MustApply(ctx, ms...)

		rowCount, err := purge(ctx)
		So(err, ShouldBeNil)
		So(rowCount, ShouldEqual, 2)
	})
}
