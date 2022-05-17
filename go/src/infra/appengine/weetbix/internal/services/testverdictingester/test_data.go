// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"time"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mockedGetBuildRsp(inv string) *bbpb.Build {
	return &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Project: "chromium",
			Bucket:  "ci",
			Builder: "builder",
		},
		Infra: &bbpb.BuildInfra{
			Resultdb: &bbpb.BuildInfra_ResultDB{
				Hostname:   "results.api.cr.dev",
				Invocation: inv,
			},
		},
		Status: bbpb.Status_FAILURE,
		Input: &bbpb.Build_Input{
			GerritChanges: []*bbpb.GerritChange{
				{
					Host:     "mygerrit-review.googlesource.com",
					Change:   12345,
					Patchset: 5,
				},
				{
					Host:     "anothergerrit-review.googlesource.com",
					Change:   77788,
					Patchset: 19,
				},
			},
		},
	}
}

func mockedQueryTestVariantsRsp() *rdbpb.QueryTestVariantsResponse {
	return &rdbpb.QueryTestVariantsResponse{
		TestVariants: []*rdbpb.TestVariant{
			{
				TestId:      "test_id_1",
				VariantHash: "hash_1",
				Status:      rdbpb.TestVariantStatus_UNEXPECTED,
				Variant:     pbutil.Variant("k1", "v1"),
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_1/results/run0-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)),
							Status:    rdbpb.TestStatus_FAIL,
							Expected:  false,
							Duration:  durationpb.New(time.Second * 10),
						},
					},
				},
			},
			{
				TestId:      "test_id_1",
				VariantHash: "hash_2",
				Status:      rdbpb.TestVariantStatus_FLAKY,
				Variant:     pbutil.Variant("k1", "v2"),
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_1/results/run0-second",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 10, 0, time.UTC)),
							Status:    rdbpb.TestStatus_FAIL,
							Expected:  false,
							Duration:  durationpb.New(time.Second * 10),
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_1/results/run0-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)),
							Status:    rdbpb.TestStatus_PASS,
							Expected:  true,
							Duration:  durationpb.New(time.Second),
						},
					},
				},
			},
			{
				TestId:      "test_id_2",
				VariantHash: "hash_1",
				Status:      rdbpb.TestVariantStatus_FLAKY,
				Variant:     pbutil.Variant("k1", "v1"),
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/b/tests/test_id_2/results/run1-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 10, 0, time.UTC)),
							Status:    rdbpb.TestStatus_FAIL,
							Expected:  false,
							Duration:  durationpb.New(time.Second * 10),
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_2/results/run0-second",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 20, 0, time.UTC)),
							Status:    rdbpb.TestStatus_PASS,
							Expected:  true,
							Duration:  durationpb.New(time.Second),
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_2/results/run0-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)),
							Status:    rdbpb.TestStatus_PASS,
							Expected:  true,
							Duration:  durationpb.New(time.Second * 3),
						},
					},
				},
			},
			{
				TestId:      "test_id_2",
				VariantHash: "hash_2",
				Status:      rdbpb.TestVariantStatus_EXONERATED,
				Variant:     pbutil.Variant("k1", "v2"),
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/a/tests/test_id_2/results/run0-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)),
							Status:    rdbpb.TestStatus_FAIL,
							Expected:  false,
							Duration:  durationpb.New(time.Second * 10),
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/b/tests/test_id_2/results/run1-second",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 10, 0, time.UTC)),
							Status:    rdbpb.TestStatus_PASS,
							Expected:  true,
							Duration:  durationpb.New(2 * time.Second),
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:      "invocations/b/tests/test_id_2/results/run1-first",
							StartTime: timestamppb.New(time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)),
							Status:    rdbpb.TestStatus_PASS,
							Expected:  true,
							Duration:  durationpb.New(time.Second),
						},
					},
				},
			},
		},
	}
}
