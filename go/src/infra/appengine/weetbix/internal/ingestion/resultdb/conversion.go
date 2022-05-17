// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resultdb

import (
	"regexp"
	"sort"

	rdbpb "go.chromium.org/luci/resultdb/proto/v1"

	pb "infra/appengine/weetbix/proto/v1"
)

// StatusFromExonerations returns the Weetbix exoneration status corresponding
// to the given ResultDB exonerations. This is NOT_EXONERATED if there are
// no exonerations, or the exoneration with the highest precedence value
// otherwise.
func StatusFromExonerations(es []*rdbpb.TestExoneration) pb.ExonerationStatus {
	result := pb.ExonerationStatus_NOT_EXONERATED
	for _, e := range es {
		alternativeStatus := statusFromExoneration(e)
		if alternativeStatus > result || result == pb.ExonerationStatus_NOT_EXONERATED {
			result = alternativeStatus
		}
	}
	return result
}

// TestResultStatus returns the Weetbix test result status corresponding
// to the given ResultDB test result status.
func TestResultStatus(s rdbpb.TestStatus) pb.TestResultStatus {
	switch s {
	case rdbpb.TestStatus_ABORT:
		return pb.TestResultStatus_ABORT
	case rdbpb.TestStatus_CRASH:
		return pb.TestResultStatus_CRASH
	case rdbpb.TestStatus_FAIL:
		return pb.TestResultStatus_FAIL
	case rdbpb.TestStatus_PASS:
		return pb.TestResultStatus_PASS
	case rdbpb.TestStatus_SKIP:
		return pb.TestResultStatus_SKIP
	default:
		return pb.TestResultStatus_TEST_RESULT_STATUS_UNSPECIFIED
	}
}

func statusFromExoneration(e *rdbpb.TestExoneration) pb.ExonerationStatus {
	switch e.Reason {
	case rdbpb.ExonerationReason_NOT_CRITICAL:
		return pb.ExonerationStatus_NOT_CRITICAL
	case rdbpb.ExonerationReason_OCCURS_ON_MAINLINE:
		return pb.ExonerationStatus_OCCURS_ON_MAINLINE
	case rdbpb.ExonerationReason_OCCURS_ON_OTHER_CLS:
		return pb.ExonerationStatus_OCCURS_ON_OTHER_CLS
	default:
		return pb.ExonerationStatus_EXONERATION_STATUS_UNSPECIFIED
	}
}

// GroupAndOrderTestResults groups test results into test runs, and orders
// them by start time. Test results are returned in sorted start time order
// within the runs, and runs are ordered based on the start time of the first
// test result that is inside them.
// The result order is guaranteed to be deterministic even if all test
// results have the same start time.
func GroupAndOrderTestResults(input []*rdbpb.TestResultBundle) [][]*rdbpb.TestResultBundle {
	var result [][]*rdbpb.TestResultBundle
	runIndexByName := make(map[string]int)

	// Process results in order of StartTime.
	// This is to ensure test result indexes are later
	// assigned correctly w.r.t the actual execution order.
	input = sortResultsByStartTime(input)

	// Process test results, creating runs as they are needed.
	// Runs will be created in the order of the first test result
	// that is inside them.
	for _, tr := range input {
		testRun := TestRunFromResult(tr.Result)
		idx, ok := runIndexByName[testRun]
		if !ok {
			// Create an empty run.
			idx = len(result)
			runIndexByName[testRun] = idx
			result = append(result, nil)
		}

		result[idx] = append(result[idx], tr)
	}
	return result
}

// testRunRe extracts the test run from the ResultDB test result name. This is
// the parent invocation the test result was included in, as distinct from
// the ingested invocation ID.
var testRunRe = regexp.MustCompile(`^invocations/([^/]+)/tests/[^/]+/results/[^/]+$`)

// TestRunFromResult extracts the invocation that the test result is
// immediately included inside.
func TestRunFromResult(r *rdbpb.TestResult) string {
	match := testRunRe.FindStringSubmatch(r.Name)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func sortResultsByStartTime(results []*rdbpb.TestResultBundle) []*rdbpb.TestResultBundle {
	// Copy the results to avoid modifying parameter slice, which
	// the caller to IngestFromResultDB may not expect.
	sortedResults := make([]*rdbpb.TestResultBundle, len(results))
	for i, r := range results {
		sortedResults[i] = r
	}

	sort.Slice(sortedResults, func(i, j int) bool {
		aResult := sortedResults[i].Result
		bResult := sortedResults[j].Result
		aTime := aResult.StartTime.AsTime()
		bTime := bResult.StartTime.AsTime()
		if aTime.Equal(bTime) {
			// If start time the same, order by Result Name.
			// Needed to ensure the output of this sort is
			// deterministic given the input.
			return aResult.Name < bResult.Name
		}
		return aTime.Before(bTime)
	})
	return sortedResults
}
