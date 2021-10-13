// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ingestion

import (
	"regexp"
	"sort"

	cpb "infra/appengine/weetbix/internal/clustering/proto"

	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func failuresFromTestVariants(opts Options, tvs []*rdbpb.TestVariant) []*cpb.Failure {
	var failures []*cpb.Failure
	for _, tv := range tvs {
		// Process results in order of StartTime.
		results := sortResultsByStartTime(tv.Results)

		// Stores the task (lowest-level invocation) for each test result.
		tasks := make([]string, len(results))

		// Whether there were any passed or expected results.
		var hasPass bool
		// Whether there were any passed or expected results for a task
		// (excluding skips).
		taskHasPass := make(map[string]bool)
		// Total number of results by task.
		countByTask := make(map[string]int64)
		for i, tr := range tv.Results {
			task := taskFromResult(tr.Result)
			tasks[i] = task
			countByTask[task] += 1

			if tr.Result.Status != rdbpb.TestStatus_SKIP &&
				(tr.Result.Status == rdbpb.TestStatus_PASS ||
					tr.Result.Expected) {
				hasPass = true
				taskHasPass[task] = true
			}
		}

		seqByTask := make(map[string]int64)
		for i, tr := range tv.Results {
			task := tasks[i]
			// Sequence values are one-based, not zero-based.
			seqByTask[task] += 1

			if !isUnexpectedFailure(tr.Result) {
				// Only unexpected failures are ingested for clustering.
				continue
			}

			exonerated := len(tv.Exonerations) > 0
			failure := failureFromResult(tr.Result, opts, exonerated, task)
			failure.RootInvocationResultSeq = int64(i + 1)
			failure.RootInvocationResultCount = int64(len(tv.Results))
			failure.IsRootInvocationBlocked = !hasPass
			failure.TaskResultSeq = seqByTask[task]
			failure.TaskResultCount = countByTask[task]
			failure.IsTaskBlocked = !taskHasPass[task]
			failures = append(failures, failure)
		}
	}
	return failures
}

// taskRe extracts the task from the ResultDB test result name. This is
// the original invocation the test result was included in, and is distinct from
// the root invocation ID. In Weetbix nomenclature, this original invocation is
// called the "task".
var taskRe = regexp.MustCompile(`^invocations/([^/]+)/tests/[^/]+/results/[^/]+$`)

func taskFromResult(r *rdbpb.TestResult) string {
	match := taskRe.FindStringSubmatch(r.Name)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func isUnexpectedFailure(r *rdbpb.TestResult) bool {
	return !r.Expected &&
		(r.Status == rdbpb.TestStatus_ABORT ||
			r.Status == rdbpb.TestStatus_CRASH ||
			r.Status == rdbpb.TestStatus_FAIL)
}

func sortResultsByStartTime(results []*rdbpb.TestResultBundle) []*rdbpb.TestResultBundle {
	// Copy the results to avoid modifying parameter slice, which
	// the caller to IngestFromResultDB may not expect.
	sortedResults := make([]*rdbpb.TestResultBundle, len(results))
	for i, r := range results {
		sortedResults[i] = proto.Clone(r).(*rdbpb.TestResultBundle)
	}

	sort.Slice(sortedResults, func(i, j int) bool {
		aResult := results[i].Result
		bResult := results[j].Result
		aTime := aResult.StartTime.AsTime()
		bTime := bResult.StartTime.AsTime()
		if aTime.Equal(bTime) {
			// If start time the same, order by Result Name.
			return aResult.Name < bResult.Name
		}
		return aTime.Before(bTime)
	})
	return sortedResults
}

func failureFromResult(tr *rdbpb.TestResult, opts Options, exonerated bool, taskId string) *cpb.Failure {
	return &cpb.Failure{
		TestResultId:              tr.Name,
		PartitionTime:             timestamppb.New(opts.PartitionTime),
		ChunkIndex:                -1, // To be populated by chunking.
		Realm:                     opts.Realm,
		TestId:                    tr.TestId,
		Variant:                   variant(tr.Variant),
		VariantHash:               tr.VariantHash,
		FailureReason:             failureReason(tr.FailureReason),
		Component:                 extractTagValue(tr.Tags, "monorail_component"),
		StartTime:                 tr.StartTime,
		Duration:                  tr.Duration,
		IsExonerated:              exonerated,
		RootInvocationId:          opts.RootInvocationID,
		RootInvocationResultSeq:   -1,    // To be populated by caller.
		RootInvocationResultCount: -1,    // To be populated by caller.
		IsRootInvocationBlocked:   false, // To be populated by caller.
		TaskId:                    taskId,
		TaskResultSeq:             -1,    // To be populated by caller.
		TaskResultCount:           -1,    // To be populated by caller.
		IsTaskBlocked:             false, // To be populated by caller.
		CqId:                      opts.CQRunID,
	}
}

func variant(v *rdbpb.Variant) *cpb.Variant {
	if v == nil {
		// Variant is optional in ResultDB.
		return &cpb.Variant{Def: make(map[string]string)}
	}
	return &cpb.Variant{Def: v.Def}
}

func failureReason(fr *rdbpb.FailureReason) *cpb.FailureReason {
	if fr == nil {
		return nil
	}
	return &cpb.FailureReason{
		PrimaryErrorMessage: fr.PrimaryErrorMessage,
	}
}

func extractTagValue(tags []*rdbpb.StringPair, key string) string {
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Value
		}
	}
	return ""
}
