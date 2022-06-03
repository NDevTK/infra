// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testresults

import (
	"time"

	pb "infra/appengine/weetbix/proto/v1"
)

// TestResultBuilder provides methods to build a test result for testing.
type TestResultBuilder struct {
	result TestResult
}

func NewTestResult() *TestResultBuilder {
	d := time.Hour
	result := TestResult{
		Project:              "proj",
		TestID:               "test_id",
		PartitionTime:        time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		VariantHash:          "hash",
		IngestedInvocationID: "inv-id",
		RunIndex:             2,
		ResultIndex:          3,
		IsUnexpected:         true,
		RunDuration:          &d,
		Status:               pb.TestResultStatus_PASS,
		ExonerationStatus:    pb.ExonerationStatus_OCCURS_ON_OTHER_CLS,
		SubRealm:             "realm",
		BuildStatus:          pb.BuildStatus_BUILD_STATUS_SUCCESS,
		Changelists: []Changelist{
			{
				Host:     "mygerrit",
				Change:   12345678,
				Patchset: 9,
			},
			{
				Host:     "anothergerrit",
				Change:   234568790,
				Patchset: 1,
			},
		},
	}
	return &TestResultBuilder{
		result: result,
	}
}

func (b *TestResultBuilder) WithProject(project string) *TestResultBuilder {
	b.result.Project = project
	return b
}

func (b *TestResultBuilder) WithTestID(testID string) *TestResultBuilder {
	b.result.TestID = testID
	return b
}

func (b *TestResultBuilder) WithPartitionTime(partitionTime time.Time) *TestResultBuilder {
	b.result.PartitionTime = partitionTime
	return b
}

func (b *TestResultBuilder) WithVariantHash(variantHash string) *TestResultBuilder {
	b.result.VariantHash = variantHash
	return b
}

func (b *TestResultBuilder) WithIngestedInvocationID(invID string) *TestResultBuilder {
	b.result.IngestedInvocationID = invID
	return b
}

func (b *TestResultBuilder) WithRunIndex(runIndex int64) *TestResultBuilder {
	b.result.RunIndex = runIndex
	return b
}

func (b *TestResultBuilder) WithResultIndex(resultIndex int64) *TestResultBuilder {
	b.result.ResultIndex = resultIndex
	return b
}

func (b *TestResultBuilder) WithIsUnexpected(unexpected bool) *TestResultBuilder {
	b.result.IsUnexpected = unexpected
	return b
}

func (b *TestResultBuilder) WithRunDuration(duration time.Duration) *TestResultBuilder {
	b.result.RunDuration = &duration
	return b
}

func (b *TestResultBuilder) WithoutRunDuration() *TestResultBuilder {
	b.result.RunDuration = nil
	return b
}

func (b *TestResultBuilder) WithStatus(status pb.TestResultStatus) *TestResultBuilder {
	b.result.Status = status
	return b
}

func (b *TestResultBuilder) WithExonerationStatus(exonerationStatus pb.ExonerationStatus) *TestResultBuilder {
	b.result.ExonerationStatus = exonerationStatus
	return b
}

func (b *TestResultBuilder) WithSubRealm(subRealm string) *TestResultBuilder {
	b.result.SubRealm = subRealm
	return b
}

func (b *TestResultBuilder) WithBuildStatus(buildStatus pb.BuildStatus) *TestResultBuilder {
	b.result.BuildStatus = buildStatus
	return b
}

func (b *TestResultBuilder) WithPresubmitRun(run *PresubmitRun) *TestResultBuilder {
	if run != nil {
		// Copy run to stop changes the caller may make to run
		// after this call propagating into the resultant test result.
		r := new(PresubmitRun)
		*r = *run
		run = r
	}
	b.result.PresubmitRun = run
	return b
}

func (b *TestResultBuilder) WithChangelists(changelist []Changelist) *TestResultBuilder {
	// Copy changelist to stop changes the caller may make to changelist
	// after this call propagating into the resultant test result.
	cls := make([]Changelist, len(changelist))
	copy(cls, changelist)

	// Ensure changelists are stored sorted.
	SortChangelists(cls)
	b.result.Changelists = cls
	return b
}

func (b *TestResultBuilder) Build() *TestResult {
	// Copy the result, so that calling further methods on the builder does
	// not change the returned test verdict.
	result := new(TestResult)
	*result = b.result

	if b.result.PresubmitRun != nil {
		run := new(PresubmitRun)
		*run = *b.result.PresubmitRun
		result.PresubmitRun = run
	}
	cls := make([]Changelist, len(b.result.Changelists))
	copy(cls, b.result.Changelists)
	result.Changelists = cls
	return result
}

// TestVerdictBuilder provides methods to build a test variant for testing.
type TestVerdictBuilder struct {
	baseResult        TestResult
	status            *pb.TestVerdictStatus
	runStatuses       []RunStatus
	passedAvgDuration *time.Duration
}

type RunStatus int64

const (
	Unexpected RunStatus = iota
	Flaky
	Expected
)

func NewTestVerdict() *TestVerdictBuilder {
	result := new(TestVerdictBuilder)
	result.baseResult = *NewTestResult().WithStatus(pb.TestResultStatus_PASS).Build()
	status := pb.TestVerdictStatus_FLAKY
	result.status = &status
	result.runStatuses = nil
	d := 919191 * time.Microsecond
	result.passedAvgDuration = &d
	return result
}

// WithBaseTestResult specifies a test result to use as the template for
// the test variant's test results.
func (b *TestVerdictBuilder) WithBaseTestResult(testResult *TestResult) *TestVerdictBuilder {
	b.baseResult = *testResult
	return b
}

// WithPassedAvgDuration specifies the average duration to use for
// passed test results. If setting to a non-nil value, make sure
// to set the result status as passed on the base test result if
// using this option.
func (b *TestVerdictBuilder) WithPassedAvgDuration(duration *time.Duration) *TestVerdictBuilder {
	b.passedAvgDuration = duration
	return b
}

// WithStatus specifies the status of the test verdict.
func (b *TestVerdictBuilder) WithStatus(status pb.TestVerdictStatus) *TestVerdictBuilder {
	b.status = &status
	return b
}

// WithRunStatus specifies the status of runs of the test verdict.
func (b *TestVerdictBuilder) WithRunStatus(runStatuses ...RunStatus) *TestVerdictBuilder {
	b.runStatuses = runStatuses
	return b
}

func applyStatus(trs []*TestResult, status pb.TestVerdictStatus) {
	// Set all test results to unexpected, not exonerated by default.
	for _, tr := range trs {
		tr.IsUnexpected = true
		tr.ExonerationStatus = pb.ExonerationStatus_NOT_EXONERATED
	}
	switch status {
	case pb.TestVerdictStatus_EXONERATED:
		for _, tr := range trs {
			tr.ExonerationStatus = pb.ExonerationStatus_OCCURS_ON_MAINLINE
		}
	case pb.TestVerdictStatus_UNEXPECTED:
		// No changes required.
	case pb.TestVerdictStatus_EXPECTED:
		allSkipped := true
		for _, tr := range trs {
			tr.IsUnexpected = false
			if tr.Status != pb.TestResultStatus_SKIP {
				allSkipped = false
			}
		}
		// Make sure not all test results are SKIPPED, to avoid the status
		// UNEXPECTEDLY_SKIPPED.
		if allSkipped {
			trs[0].Status = pb.TestResultStatus_CRASH
		}
	case pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED:
		for _, tr := range trs {
			tr.Status = pb.TestResultStatus_SKIP
		}
	case pb.TestVerdictStatus_FLAKY:
		trs[0].IsUnexpected = false
	default:
		panic("status must be specified")
	}
}

// applyRunStatus applies the given run status to the given test results.
func applyRunStatus(trs []*TestResult, runStatus RunStatus) {
	for _, tr := range trs {
		tr.IsUnexpected = true
	}
	switch runStatus {
	case Expected:
		for _, tr := range trs {
			tr.IsUnexpected = false
		}
	case Flaky:
		trs[0].IsUnexpected = false
	case Unexpected:
		// All test results already unexpected.
	}
}

func applyAvgPassedDuration(trs []*TestResult, passedAvgDuration *time.Duration) {
	if passedAvgDuration == nil {
		for _, tr := range trs {
			if tr.Status == pb.TestResultStatus_PASS {
				tr.RunDuration = nil
			}
		}
		return
	}

	passCount := 0
	for _, tr := range trs {
		if tr.Status == pb.TestResultStatus_PASS {
			passCount++
		}
	}
	passIndex := 0
	for _, tr := range trs {
		if tr.Status == pb.TestResultStatus_PASS {
			d := *passedAvgDuration
			if passCount == 1 {
				// If there is only one pass, assign it the
				// set duration.
				tr.RunDuration = &d
				break
			}
			if passIndex == 0 && passCount%2 == 1 {
				// If there are an odd number of passes, and
				// more than one pass, assign the first pass
				// a nil duration.
				tr.RunDuration = nil
			} else {
				// Assigning alternating passes 2*d the duration
				// and 0 duration, to keep the average correct.
				if passIndex%2 == 0 {
					d = d * 2
					tr.RunDuration = &d
				} else {
					d = 0
					tr.RunDuration = &d
				}
			}
			passIndex++
		}
	}
}

func (b *TestVerdictBuilder) Build() []*TestResult {
	runs := 2
	if len(b.runStatuses) > 0 {
		runs = len(b.runStatuses)
	}

	// Create two test results per run, to allow
	// for all expected, all unexpected and
	// flaky (mixed expected+unexpected) statuses
	// to be represented.
	trs := make([]*TestResult, 0, runs*2)
	for i := 0; i < runs*2; i++ {
		tr := new(TestResult)
		*tr = b.baseResult
		tr.RunIndex = int64(i / 2)
		tr.ResultIndex = int64(i % 2)
		trs = append(trs, tr)
	}

	// Normally only one of these should be set.
	// If both are set, run statuses has precedence.
	if b.status != nil {
		applyStatus(trs, *b.status)
	}
	for i, runStatus := range b.runStatuses {
		runTRs := trs[i*2 : (i+1)*2]
		applyRunStatus(runTRs, runStatus)
	}

	applyAvgPassedDuration(trs, b.passedAvgDuration)
	return trs
}
