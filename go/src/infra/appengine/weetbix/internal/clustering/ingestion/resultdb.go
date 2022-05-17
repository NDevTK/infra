// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ingestion

import (
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	cpb "infra/appengine/weetbix/internal/clustering/proto"
	"infra/appengine/weetbix/internal/ingestion/resultdb"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

func failuresFromTestVariants(opts Options, tvs []*rdbpb.TestVariant) []*cpb.Failure {
	var failures []*cpb.Failure
	for _, tv := range tvs {
		// Whether there were any passed or expected results.
		var hasPass bool
		for _, tr := range tv.Results {
			if tr.Result.Status != rdbpb.TestStatus_SKIP &&
				(tr.Result.Status == rdbpb.TestStatus_PASS ||
					tr.Result.Expected) {
				hasPass = true
			}
		}

		// Group test results by run and sort in order of start time.
		resultsByRun := resultdb.GroupAndOrderTestResults(tv.Results)

		resultIndex := 0
		for _, run := range resultsByRun {
			// Whether there were any passed or expected results in the run.
			var testRunHasPass bool
			for _, tr := range run {
				if tr.Result.Status != rdbpb.TestStatus_SKIP &&
					(tr.Result.Status == rdbpb.TestStatus_PASS ||
						tr.Result.Expected) {
					testRunHasPass = true
				}
			}

			for i, tr := range run {
				if tr.Result.Expected || !isFailure(tr.Result.Status) {
					// Only unexpected failures are ingested for clustering.
					resultIndex++
					continue
				}

				exoneration := resultdb.StatusFromExonerations(tv.Exonerations)
				if !hasPass && opts.ImplicitlyExonerateBlockingFailures && exoneration == pb.ExonerationStatus_NOT_EXONERATED {
					// TODO(meiring): Replace with separate build status field.
					exoneration = pb.ExonerationStatus_IMPLICIT
				}

				testRun := resultdb.TestRunFromResult(tr.Result)
				failure := failureFromResult(tv, tr.Result, opts, exoneration, testRun)
				failure.IngestedInvocationResultIndex = int64(resultIndex)
				failure.IngestedInvocationResultCount = int64(len(tv.Results))
				failure.IsIngestedInvocationBlocked = !hasPass
				failure.TestRunResultIndex = int64(i)
				failure.TestRunResultCount = int64(len(run))
				failure.IsTestRunBlocked = !testRunHasPass
				failures = append(failures, failure)

				resultIndex++
			}
		}
	}
	return failures
}

func isFailure(s rdbpb.TestStatus) bool {
	return (s == rdbpb.TestStatus_ABORT ||
		s == rdbpb.TestStatus_CRASH ||
		s == rdbpb.TestStatus_FAIL)
}

func failureFromResult(tv *rdbpb.TestVariant, tr *rdbpb.TestResult, opts Options, exonerationStatus pb.ExonerationStatus, testRunID string) *cpb.Failure {
	result := &cpb.Failure{
		TestResultId:                  pbutil.TestResultIDFromResultDB(tr.Name),
		PartitionTime:                 timestamppb.New(opts.PartitionTime),
		ChunkIndex:                    -1, // To be populated by chunking.
		Realm:                         opts.Realm,
		TestId:                        tv.TestId,                              // Get from variant, as it is not populated on each result.
		Variant:                       pbutil.VariantFromResultDB(tv.Variant), // Get from variant, as it is not populated on each result.
		VariantHash:                   tv.VariantHash,                         // Get from variant, as it is not populated on each result.
		FailureReason:                 pbutil.FailureReasonFromResultDB(tr.FailureReason),
		BugTrackingComponent:          extractBugTrackingComponent(tr.Tags),
		StartTime:                     tr.StartTime,
		Duration:                      tr.Duration,
		ExonerationStatus:             exonerationStatus,
		IngestedInvocationId:          opts.InvocationID,
		IngestedInvocationResultIndex: -1,    // To be populated by caller.
		IngestedInvocationResultCount: -1,    // To be populated by caller.
		IsIngestedInvocationBlocked:   false, // To be populated by caller.
		TestRunId:                     testRunID,
		TestRunResultIndex:            -1,    // To be populated by caller.
		TestRunResultCount:            -1,    // To be populated by caller.
		IsTestRunBlocked:              false, // To be populated by caller.
		PresubmitRunId:                opts.PresubmitRunID,
		PresubmitRunOwner:             opts.PresubmitRunOwner,
		PresubmitRunCls:               opts.PresubmitRunCls,
	}
	// Copy the result to avoid the result aliasing any of the protos used as input.
	return proto.Clone(result).(*cpb.Failure)
}

func extractBugTrackingComponent(tags []*rdbpb.StringPair) *pb.BugTrackingComponent {
	var value string
	for _, tag := range tags {
		if tag.Key == "monorail_component" {
			value = tag.Value
			break
		}
	}
	if value != "" {
		return &pb.BugTrackingComponent{
			System:    "monorail",
			Component: value,
		}
	}
	return nil
}
