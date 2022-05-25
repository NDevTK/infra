// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testresults

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/testutil"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func TestQueryFailureRate(t *testing.T) {
	Convey("QueryFailureRate", t, func() {
		ctx := testutil.SpannerTestContext(t)

		referenceTime := time.Date(2022, time.January, 1, 0, 0, 0, 1000, time.UTC)

		var1 := pbutil.Variant("key1", "val1", "key2", "val1")
		var3 := pbutil.Variant("key1", "val2", "key2", "val2")

		err := CreateQueryFailureRateTestData(ctx, referenceTime)
		So(err, ShouldBeNil)

		project, tvs := QueryFailureRateSampleRequest()
		opts := QueryFailureRateOptions{
			Project:            project,
			TestVariants:       tvs,
			AfterPartitionTime: referenceTime.Add(-6 * time.Hour),
		}
		expectedResult := QueryFailureRateSampleResponse(referenceTime)

		Convey("Baseline", func() {
			result, err := QueryFailureRate(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(result, ShouldResembleProto, expectedResult)
		})
		Convey("Project filter works correctly", func() {
			opts.Project = "none"
			expectedResult = []*pb.TestVariantFailureRateAnalysis{
				emptyResult("test_id", var1),
				emptyResult("test_id", var3),
			}

			result, err := QueryFailureRate(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(result, ShouldResembleProto, expectedResult)
		})
		Convey("Works for tests without data", func() {
			notExistsVariant := pbutil.Variant("key1", "val1", "key2", "not_exists")
			opts.TestVariants = append(opts.TestVariants,
				&pb.TestVariantIdentifier{
					TestId:  "not_exists_test_id",
					Variant: var1,
				},
				&pb.TestVariantIdentifier{
					TestId:  "test_id",
					Variant: notExistsVariant,
				})

			expectedResult = append(expectedResult,
				emptyResult("not_exists_test_id", var1),
				emptyResult("test_id", notExistsVariant))

			result, err := QueryFailureRate(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(result, ShouldResembleProto, expectedResult)
		})
		Convey("AfterPartitionTime works correctly", func() {
			opts.AfterPartitionTime = referenceTime.Add(-2 * time.Hour)
			expectedResult[0] = emptyResult("test_id", var1)
			expectedResult[1].Sample.VerdictsPreDeduplication = 1

			result, err := QueryFailureRate(span.Single(ctx), opts)
			So(err, ShouldBeNil)
			So(result, ShouldResembleProto, expectedResult)
		})
	})
}

func emptyResult(testID string, variant *pb.Variant) *pb.TestVariantFailureRateAnalysis {
	return &pb.TestVariantFailureRateAnalysis{
		TestId:  testID,
		Variant: variant,
		FailingRunRatio: &pb.FailingRunRatio{
			Numerator:   0,
			Denominator: 0,
		},
		FlakyVerdictRatio: &pb.FlakyVerdictRatio{
			Numerator:   0,
			Denominator: 0,
		},
		Sample: &pb.TestVariantFailureRateAnalysis_Sample{
			Verdicts:                 0,
			VerdictsPreDeduplication: 0,
		},
	}
}
