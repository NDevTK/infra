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

func failuresFromTestVariant(opts Options, tv *rdbpb.TestVariant) []*cpb.Failure {
	var failures []*cpb.Failure
	if tv.Status == rdbpb.TestVariantStatus_EXPECTED {
		// Short circuit: There will be nothing in the test variant to
		// ingest, as everything is expected.
		return nil
	}

	// Whether there were any (non-skip) passed or expected results.
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

			failure := failureFromResult(tv, tr.Result, opts)
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
	return failures
}

func isFailure(s rdbpb.TestStatus) bool {
	return (s == rdbpb.TestStatus_ABORT ||
		s == rdbpb.TestStatus_CRASH ||
		s == rdbpb.TestStatus_FAIL)
}

func failureFromResult(tv *rdbpb.TestVariant, tr *rdbpb.TestResult, opts Options) *cpb.Failure {
	exonerations := make([]*cpb.TestExoneration, 0, len(tv.Exonerations))
	for _, e := range tv.Exonerations {
		exonerations = append(exonerations, exonerationFromResultDB(e))
	}

	var presubmitRun *cpb.PresubmitRun
	var buildCritical *bool
	if opts.PresubmitRun != nil {
		presubmitRun = &cpb.PresubmitRun{
			PresubmitRunId: opts.PresubmitRun.ID,
			Owner:          opts.PresubmitRun.Owner,
			Mode:           opts.PresubmitRun.Mode,
			Status:         opts.PresubmitRun.Status,
		}
		buildCritical = &opts.BuildCritical
	}

	testRunID, err := resultdb.InvocationFromTestResultName(tr.Name)
	if err != nil {
		// Should never happen, as the result name from ResultDB
		// should be valid.
		panic(err)
	}

	result := &cpb.Failure{
		TestResultId:                  pbutil.TestResultIDFromResultDB(tr.Name),
		PartitionTime:                 timestamppb.New(opts.PartitionTime),
		ChunkIndex:                    -1, // To be populated by chunking.
		Realm:                         opts.Realm,
		TestId:                        tv.TestId,                              // Get from variant, as it is not populated on each result.
		Variant:                       pbutil.VariantFromResultDB(tv.Variant), // Get from variant, as it is not populated on each result.
		Tags:                          pbutil.StringPairFromResultDB(tr.Tags),
		VariantHash:                   tv.VariantHash, // Get from variant, as it is not populated on each result.
		FailureReason:                 pbutil.FailureReasonFromResultDB(tr.FailureReason),
		BugTrackingComponent:          extractBugTrackingComponent(tr.Tags),
		StartTime:                     tr.StartTime,
		Duration:                      tr.Duration,
		Exonerations:                  exonerations,
		PresubmitRun:                  presubmitRun,
		BuildStatus:                   opts.BuildStatus,
		BuildCritical:                 buildCritical,
		Changelists:                   opts.Changelists,
		IngestedInvocationId:          opts.InvocationID,
		IngestedInvocationResultIndex: -1,    // To be populated by caller.
		IngestedInvocationResultCount: -1,    // To be populated by caller.
		IsIngestedInvocationBlocked:   false, // To be populated by caller.
		TestRunId:                     testRunID,
		TestRunResultIndex:            -1,    // To be populated by caller.
		TestRunResultCount:            -1,    // To be populated by caller.
		IsTestRunBlocked:              false, // To be populated by caller.
	}

	// Copy the result to avoid the result aliasing any of the protos used as input.
	return proto.Clone(result).(*cpb.Failure)
}

func exonerationFromResultDB(e *rdbpb.TestExoneration) *cpb.TestExoneration {
	return &cpb.TestExoneration{
		Reason: pbutil.ExonerationReasonFromResultDB(e.Reason),
	}
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
