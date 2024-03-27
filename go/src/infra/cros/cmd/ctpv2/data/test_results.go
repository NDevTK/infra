// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"fmt"
	"strings"

	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
)

type TestResults struct {
	Suite         string
	Key           string // board-model-variant
	TopLevelError *error
	Results       *skylab_test_runner.Result
	Attempt       int // 0 means no retry
	BuildUrl      string
}

func (t *TestResults) GetFailureErr() error {
	if t.TopLevelError != nil {
		return *t.TopLevelError
	}

	testResults, ok := t.Results.GetAutotestResults()["original_test"]
	if !ok {
		// the test results from trv2 should be here, if not,
		// something else failed before test execution. so fail.
		return fmt.Errorf("no test result found")
	}

	for _, testCase := range testResults.GetTestCases() {
		if testCase.GetVerdict() != skylab_test_runner.Result_Autotest_TestCase_VERDICT_PASS {
			return fmt.Errorf("test(s) failure")
		}
	}

	return nil
}

type ByAttempt []*TestResults

func (a ByAttempt) Len() int           { return len(a) }
func (a ByAttempt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAttempt) Less(i, j int) bool { return a[i].Attempt < a[j].Attempt }

type BotParamsRejectedError struct {
	Key          string
	RejectedDims []string
}

func (e *BotParamsRejectedError) Error() string {
	return fmt.Sprintf("rejected params: %s", strings.Join(e.RejectedDims, ", "))
}

type EnumerationError struct {
	SuiteName string
}

func (e *EnumerationError) Error() string {
	return fmt.Sprintf("no test found for suite '%s'", e.SuiteName)
}
