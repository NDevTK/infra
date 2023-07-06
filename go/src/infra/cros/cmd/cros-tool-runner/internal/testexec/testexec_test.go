// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package testexec runs tests.
package testexec

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestPrepareTestResponse_NilTestCaseResults(t *testing.T) {
	resultRootDir := "/path/to/results"

	testCaseResult := &api.TestCaseResult{
		ResultDirPath: nil,
	}
	testCaseResults := []*api.TestCaseResult{testCaseResult}

	res, err := prepareTestResponse(resultRootDir, testCaseResults)

	// Verify that the returned response is not nil.
	if res == nil {
		t.Errorf("Expected non-nil response, got nil")
	}

	// Verify that the TestCaseResults field in the response is an empty slice.
	if len(res.TestCaseResults) != 0 {
		t.Errorf("Expected empty TestCaseResults, got %d elements", len(res.TestCaseResults))
	}

	// Verify that the returned error is nil.
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
