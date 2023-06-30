// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execution

import (
	"context"
	"errors"
	"strings"
	"time"

	"infra/cmd/cros_test_platform/internal/execution/build"
	trservice "infra/cmd/cros_test_platform/internal/execution/testrunner/service"
	"infra/cmd/cros_test_platform/internal/execution/types"

	buildbucket "go.chromium.org/luci/buildbucket/proto"

	"go.chromium.org/luci/common/logging"
)

var suiteLimitError = errors.New("TestExecutionLimit: Maximum suite execution runtime exceeded.")

type suiteFilter struct {
	suiteName    string
	expiration   time.Time
	neverExpires bool
}

// exceptions stores all granted exceptions from the SuiteLimits project.
var exceptions = []suiteFilter{
	{
		suiteName:    "arc-cts",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-cts-camera-opendut",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-cts-hardware",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-cts-qual",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-cts-vm-stable",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-gts",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-gts-qual",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-sts-full",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-sts-full-r",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-sts-full-t",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "arc-sts-incremental-r",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: true,
	},
	{
		suiteName:    "bvt-tast-arc",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: false,
	},
	{
		suiteName:    "bvt-tast-cq",
		expiration:   time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
		neverExpires: false,
	},
}

func cancelExceededTests(ctx context.Context, client trservice.Client, taskSetName string, taskSet *RequestTaskSet) error {
	// Aggregate all the references of tasks to be cancelled.
	taskIds := []trservice.TaskReference{}
	for _, task := range taskSet.activeTasks {
		taskIds = append(taskIds, task.TaskReference)
	}
	err := client.CancelTasks(ctx, taskIds, "SUITE EXECUTION TIME LIMIT EXCEEDED")
	if err != nil {
		return err
	}

	// Remove all tasks from the active task set
	for iid := range taskSet.activeTasks {
		taskSet.step.Close(buildbucket.Status_FAILURE, build.ExceededExecutionTimeText)
		taskSet.invocationSteps[iid].AddCancelledSummary()

		delete(taskSet.activeTasks, iid)
	}

	return err
}

func checkForExceptionNoLogging(taskSetName string) bool {
	// Iterate through the exceptions list.
	for _, exception := range exceptions {

		// Continue early if current exception doesn't apply.
		if !strings.HasSuffix(taskSetName, exception.suiteName) {
			continue
		}

		// Exception has expired but has not been removed from the list. Explain
		// the failure in the summary markdown.
		if exception.expiration.Before(time.Now()) {
			return false
		}
		// Active exception found.
		return true
	}

	// No Exception Found
	return false
}

// checkForException iterates through the exceptions list to see if we should
// allow the suite to continue running.
func checkForException(ctx context.Context, taskSetName string, request *RequestTaskSet) bool {
	// If exception already granted then quickly return with an OK.
	if request.SuiteLimitExceptionGranted {
		return true
	}

	// Iterate through the exceptions list.
	for _, exception := range exceptions {
		// Continue early if current exception doesn't apply.
		if !strings.HasSuffix(taskSetName, exception.suiteName) {
			continue
		}

		// Exception has expired but has not been removed from the list. Explain
		// the failure in the summary markdown.
		if exception.expiration.Before(time.Now()) {
			request.step.DisplayExceptionExpiredSummary(exception.expiration)
			return false
		}

		logging.Infof(ctx, "SuiteLimits: Exception found for taskSetName: %s\n", exception.suiteName)

		// Mark as exception granted so next checks will not require a full list search.
		request.SuiteLimitExceptionGranted = true

		// Show exception information in the summary markdown.
		request.step.DisplayExceptionSummary(exception.expiration)

		// Active exception found.
		return true
	}

	// No Exception Found
	return false
}

// updateTestExecutionTracking calculates the amount of time it's been since
// this task was last seen. It then increase the global value tracking test
// execution and returns the updated map with a new timestamp for the current iid.
func updateTestExecutionTracking(ctx context.Context, iid types.InvocationID, lastSeen time.Time, taskSetName string, request *RequestTaskSet, completed bool, logChan chan trackingMetric) error {

	// Mark the current time for calculation of the duration since last seen.
	currentlySeenAt := time.Now()
	lastSeenRuntimePerTask[taskSetName].lastSeenMap[iid] = currentlySeenAt

	// Calculate the duration of time it has been since the last time we've seen
	// this iid running.
	delta := lastSeenRuntimePerTask[taskSetName].lastSeenMap[iid].Sub(lastSeen)

	// Increase the total test execution time for the suite.
	lastSeenRuntimePerTask[taskSetName].totalSuiteTrackingTime += delta

	// Add update to the log.
	logChan <- trackingMetric{
		suiteName:       taskSetName,
		taskName:        string(iid),
		lastSeen:        lastSeen,
		currentlySeenAt: currentlySeenAt,
		delta:           delta,
		completed:       completed,
	}

	// Check if we've exceeded the maximum time allowed for test execution.
	// TODO(b/254114334): Activate this once the feature is dropped.
	// if lastSeenRuntimePerTask[taskSetName].totalSuiteTrackingTime.Seconds() > SuiteTestExecutionMaximumSeconds && !checkForException(ctx, taskSetName, request) {
	// 	logging.Infof(ctx, "Suite %s exceeded execution runtime limit. %d seconds allowed, %d seconds used.", taskSetName, SuiteTestExecutionMaximumSeconds, int(lastSeenRuntimePerTask[taskSetName].totalSuiteTrackingTime.Seconds()))

	// 	return suiteLimitError
	// }

	return nil
}
