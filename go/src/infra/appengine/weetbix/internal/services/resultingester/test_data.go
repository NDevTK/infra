// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resultingester

import (
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
)

var sampleVar = pbutil.Variant("k1", "v1")
var sampleTmd = &rdbpb.TestMetadata{
	Name: "test_new_failure",
}

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
	}
}

func mockedQueryTestVariantsRsp() *rdbpb.QueryTestVariantsResponse {
	return &rdbpb.QueryTestVariantsResponse{
		TestVariants: []*rdbpb.TestVariant{
			{
				TestId:       "ninja://test_new_failure",
				VariantHash:  "hash",
				Status:       rdbpb.TestVariantStatus_UNEXPECTED,
				Variant:      pbutil.Variant("k1", "v1"),
				TestMetadata: sampleTmd,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_new_failure/results/one",
							Status: rdbpb.TestStatus_FAIL,
							Tags:   pbutil.StringPairs("random_tag", "random_tag_value", "monorail_component", "Monorail>Component"),
						},
					},
				},
			},
			{
				TestId:      "ninja://test_known_flake",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_UNEXPECTED,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_known_flake/results/one",
							Status: rdbpb.TestStatus_FAIL,
							Tags:   pbutil.StringPairs("os", "Mac", "monorail_component", "Monorail>Component"),
						},
					},
				},
			},
			{
				TestId:      "ninja://test_consistent_failure",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_UNEXPECTED,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_consistent_failure/results/one",
							Status: rdbpb.TestStatus_FAIL,
						},
					},
				},
			},
			{
				TestId:      "ninja://test_no_new_results",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_UNEXPECTED,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_no_new_results/results/one",
							Status: rdbpb.TestStatus_FAIL,
						},
					},
				},
			},
			// Should ignore.
			{
				TestId:      "ninja://test_skip",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_UNEXPECTEDLY_SKIPPED,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_skip/results/one",
							Status: rdbpb.TestStatus_SKIP,
						},
					},
				},
			},
			{
				TestId:      "ninja://test_new_flake",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_FLAKY,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_new_flake/results/one",
							Status: rdbpb.TestStatus_FAIL,
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_new_flake/results/two",
							Status: rdbpb.TestStatus_PASS,
						},
					},
				},
			},
			{
				TestId:      "ninja://test_has_unexpected",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_FLAKY,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_has_unexpected/results/one",
							Status: rdbpb.TestStatus_FAIL,
						},
					},
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_has_unexpected/results/two",
							Status: rdbpb.TestStatus_PASS,
						},
					},
				},
			},
			{
				TestId:      "ninja://test_unexpected_pass",
				VariantHash: "hash",
				Status:      rdbpb.TestVariantStatus_UNEXPECTED,
				Results: []*rdbpb.TestResultBundle{
					{
						Result: &rdbpb.TestResult{
							Name:   "invocations/build-1234/tests/ninja%3A%2F%2Ftest_unexpected_pass/results/one",
							Status: rdbpb.TestStatus_PASS,
						},
					},
				},
			},
		},
	}
}
