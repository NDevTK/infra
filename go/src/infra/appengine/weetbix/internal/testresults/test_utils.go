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
		Project:                      "proj",
		TestID:                       "test_id",
		PartitionTime:                time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		VariantHash:                  "hash",
		IngestedInvocationID:         "inv-id",
		RunIndex:                     2,
		ResultIndex:                  3,
		IsUnexpected:                 true,
		RunDuration:                  &d,
		Status:                       pb.TestResultStatus_PASS,
		ExonerationStatus:            pb.ExonerationStatus_OCCURS_ON_OTHER_CLS,
		SubRealm:                     "realm",
		BuildStatus:                  pb.BuildStatus_BUILD_STATUS_SUCCESS,
		HasContributedToClSubmission: true,
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

func (b *TestResultBuilder) WithHasContributedToClSubmission(hasContributedToClSubmission bool) *TestResultBuilder {
	b.result.HasContributedToClSubmission = hasContributedToClSubmission
	return b
}

func (b *TestResultBuilder) WithChangelists(changelist []Changelist) *TestResultBuilder {
	// Copy changelist to stop changes the caller may make to changelist
	// after this call propagating into the resultant test result.
	cls := make([]Changelist, len(changelist))
	copy(cls, changelist)
	b.result.Changelists = cls
	return b
}

func (b *TestResultBuilder) Build() *TestResult {
	// Copy the result, so that calling further methods on the builder does
	// not change the returned test verdict.
	result := new(TestResult)
	*result = b.result

	cls := make([]Changelist, len(b.result.Changelists))
	copy(cls, b.result.Changelists)
	result.Changelists = cls
	return result
}

// TestVariantBuilder provides methods to build a test variant for testing.
type TestVariantBuilder struct {
	baseResult        TestResult
	status            pb.TestVerdictStatus
	passedAvgDuration *time.Duration
}

func NewTestVariant() *TestVariantBuilder {
	result := new(TestVariantBuilder)
	result.baseResult = *NewTestResult().WithStatus(pb.TestResultStatus_PASS).Build()
	result.status = pb.TestVerdictStatus_FLAKY
	d := 919191 * time.Microsecond
	result.passedAvgDuration = &d
	return result
}

// WithBaseTestResult specifies a test result to use as the template for
// the test variant's test results.
func (b *TestVariantBuilder) WithBaseTestResult(testResult *TestResult) *TestVariantBuilder {
	b.baseResult = *testResult
	return b
}

// WithPassedAvgDuration specifies the average duration to use for
// passed test results. If setting to a non-nil value, make sure
// to set the result status as passed on the base test result if
// using this option.
func (b *TestVariantBuilder) WithPassedAvgDuration(duration *time.Duration) *TestVariantBuilder {
	b.passedAvgDuration = duration
	return b
}

// WithPassedAvgDuration specifies the status of the test verdict.
func (b *TestVariantBuilder) WithStatus(status pb.TestVerdictStatus) *TestVariantBuilder {
	b.status = status
	return b
}

func (b *TestVariantBuilder) Build() []*TestResult {
	result := make([]*TestResult, 3)
	result[0] = new(TestResult)
	*result[0] = b.baseResult
	result[1] = new(TestResult)
	*result[1] = b.baseResult
	result[2] = new(TestResult)
	*result[2] = b.baseResult

	// Set status.
	for _, tr := range result {
		tr.IsUnexpected = true
		tr.ExonerationStatus = pb.ExonerationStatus_NOT_EXONERATED
	}
	result[0].RunIndex = 0
	result[0].ResultIndex = 0
	result[1].RunIndex = 1
	result[1].ResultIndex = 0
	result[2].RunIndex = 1
	result[2].ResultIndex = 1

	switch b.status {
	case pb.TestVerdictStatus_EXONERATED:
		result[0].ExonerationStatus = pb.ExonerationStatus_OCCURS_ON_MAINLINE
		result[1].ExonerationStatus = pb.ExonerationStatus_OCCURS_ON_MAINLINE
		result[2].ExonerationStatus = pb.ExonerationStatus_OCCURS_ON_MAINLINE
	case pb.TestVerdictStatus_UNEXPECTED:
		// No changes required.
	case pb.TestVerdictStatus_EXPECTED:
		result[0].IsUnexpected = false
		result[1].IsUnexpected = false
		result[2].IsUnexpected = false
		// Make sure not all test results are SKIPPED, to avoid the status
		// UNEXPECTEDLY_SKIPPED.
		if result[0].Status == pb.TestResultStatus_SKIP {
			result[0].Status = pb.TestResultStatus_CRASH
		}
	case pb.TestVerdictStatus_UNEXPECTEDLY_SKIPPED:
		result[0].Status = pb.TestResultStatus_SKIP
		result[1].Status = pb.TestResultStatus_SKIP
		result[2].Status = pb.TestResultStatus_SKIP
	case pb.TestVerdictStatus_FLAKY:
		result[0].IsUnexpected = false
		result[1].IsUnexpected = true
		result[2].IsUnexpected = false
	default:
		panic("status must be specified")
	}

	// Set average passed duration.
	passCount := 0
	for _, tr := range result {
		if tr.Status == pb.TestResultStatus_PASS {
			passCount++
		}
	}
	for i, tr := range result {
		if tr.Status == pb.TestResultStatus_PASS {
			if b.passedAvgDuration == nil {
				tr.RunDuration = nil
			} else {
				d := *b.passedAvgDuration
				if passCount == 1 {
					tr.RunDuration = &d
				} else { // passCount == 2 or 3
					if i == 0 {
						// Set one test result to have twice
						// the average duration.
						d = d * 2
						tr.RunDuration = &d
					} else if i == 1 {
						// Set the second to have zero
						// duration.
						d = 0
						tr.RunDuration = &d
					} else if i == 2 {
						// Set the last to have no duration.
						tr.RunDuration = nil
					}
				}
			}
		}
	}
	return result
}
