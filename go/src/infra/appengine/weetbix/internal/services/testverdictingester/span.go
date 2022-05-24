// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/pkg/errors"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/sync/parallel"
	rdbpbutil "go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/ingestion/resultdb"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
	"infra/appengine/weetbix/utils"
)

func extractIngestedInvocation(task *taskspb.IngestTestVerdicts, build *bbpb.Build, inv *rdbpb.Invocation) (*testresults.IngestedInvocation, error) {
	invID, err := rdbpbutil.ParseInvocationName(inv.Name)
	if err != nil {
		// This should never happen. Inv was originated from ResultDB.
		panic(err)
	}

	proj, subRealm := utils.SplitRealm(inv.Realm)

	contributedToCLSubmission := false
	if task.PresubmitRun != nil {
		contributedToCLSubmission = task.PresubmitRun.PresubmitRunSucceeded
	}

	var buildStatus pb.BuildStatus
	switch build.Status {
	case bbpb.Status_CANCELED:
		buildStatus = pb.BuildStatus_BUILD_STATUS_CANCELED
	case bbpb.Status_SUCCESS:
		buildStatus = pb.BuildStatus_BUILD_STATUS_SUCCESS
	case bbpb.Status_FAILURE:
		buildStatus = pb.BuildStatus_BUILD_STATUS_FAILURE
	case bbpb.Status_INFRA_FAILURE:
		buildStatus = pb.BuildStatus_BUILD_STATUS_INFRA_FAILURE
	default:
		return nil, fmt.Errorf("build has unknown status: %v", build.Status)
	}

	gerritChanges := build.GetInput().GetGerritChanges()
	changelists := make([]testresults.Changelist, 0, len(gerritChanges))
	for _, change := range gerritChanges {
		if !strings.HasSuffix(change.Host, "-review.googlesource.com") {
			return nil, fmt.Errorf(`gerrit host %q does not end in expected suffix "-review.googlesource.com"`, change.Host)
		}
		host := strings.TrimSuffix(change.Host, "-review.googlesource.com")
		changelists = append(changelists, testresults.Changelist{
			Host:     host,
			Change:   change.Change,
			Patchset: change.Patchset,
		})
	}

	return &testresults.IngestedInvocation{
		Project:                      proj,
		IngestedInvocationID:         invID,
		SubRealm:                     subRealm,
		PartitionTime:                task.PartitionTime.AsTime(),
		BuildStatus:                  buildStatus,
		HasContributedToClSubmission: contributedToCLSubmission,
		Changelists:                  changelists,
	}, nil
}

func recordIngestedInvocation(ctx context.Context, inv *testresults.IngestedInvocation) error {
	// Update the IngestedInvocations table.
	m := inv.SaveUnverified()

	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		span.BufferWrite(ctx, m)
		return nil
	})
	return err
}

type batch struct {
	// The test variant realms to insert/update.
	// Test variant realms should be inserted before any test results.
	testVariantRealms []*spanner.Mutation
	// Test results to insert. Already prepared as Spanner mutations.
	testResults []*spanner.Mutation
}

func batchTestResults(inv *testresults.IngestedInvocation, tvs []*rdbpb.TestVariant, outputC chan batch) {
	// Must be selected such that no more than 20,000 mutations occur in
	// one transaction in the worst case.
	const batchSize = 1000

	var trs []*spanner.Mutation
	var tvrs []*spanner.Mutation
	startBatch := func() {
		trs = make([]*spanner.Mutation, 0, batchSize)
		tvrs = make([]*spanner.Mutation, 0, batchSize)
	}
	outputBatch := func() {
		if len(trs) == 0 {
			// This should never happen.
			panic("Pushing empty batch")
		}
		outputC <- batch{
			testVariantRealms: tvrs,
			testResults:       trs,
		}
	}

	startBatch()
	for _, tv := range tvs {
		// Limit batch size.
		// Keep all results for one test variant in one batch, so that the
		// TestVariantRealm record is kept together with the test results.
		if len(trs) > batchSize {
			outputBatch()
			startBatch()
		}

		tvr := testresults.TestVariantRealm{
			Project:           inv.Project,
			TestID:            tv.TestId,
			VariantHash:       tv.VariantHash,
			SubRealm:          inv.SubRealm,
			Variant:           pbutil.VariantFromResultDB(tv.Variant),
			LastIngestionTime: spanner.CommitTimestamp,
		}
		tvrs = append(tvrs, tvr.SaveUnverified())

		exonerationStatus := resultdb.StatusFromExonerations(tv.Exonerations)

		// Group results into test runs and order them by start time.
		resultsByRun := resultdb.GroupAndOrderTestResults(tv.Results)
		for runIndex, run := range resultsByRun {
			for resultIndex, inputTR := range run {
				tr := testresults.TestResult{
					Project:                      inv.Project,
					TestID:                       tv.TestId,
					PartitionTime:                inv.PartitionTime,
					VariantHash:                  tv.VariantHash,
					IngestedInvocationID:         inv.IngestedInvocationID,
					RunIndex:                     int64(runIndex),
					ResultIndex:                  int64(resultIndex),
					IsUnexpected:                 !inputTR.Result.Expected,
					Status:                       resultdb.TestResultStatus(inputTR.Result.Status),
					ExonerationStatus:            exonerationStatus,
					SubRealm:                     inv.SubRealm,
					BuildStatus:                  inv.BuildStatus,
					HasContributedToClSubmission: inv.HasContributedToClSubmission,
					Changelists:                  inv.Changelists,
				}
				if inputTR.Result.Duration != nil {
					d := new(time.Duration)
					*d = inputTR.Result.Duration.AsDuration()
					tr.RunDuration = d
				}
				// Convert the test result into a mutation immediately
				// to avoid storing both the TestResult object and
				// mutation object in memory until the transaction
				// commits.
				trs = append(trs, tr.SaveUnverified())
			}
		}
	}
	if len(trs) > 0 {
		outputBatch()
	}
}

// recordTestResults records test results from an test-verdict-ingestion task.
func recordTestResults(ctx context.Context, inv *testresults.IngestedInvocation, tvs []*rdbpb.TestVariant) error {
	const workerCount = 8

	return parallel.WorkPool(workerCount, func(c chan<- func() error) {
		batchC := make(chan batch)

		c <- func() error {
			defer close(batchC)
			batchTestResults(inv, tvs, batchC)
			return nil
		}

		for batch := range batchC {
			// Bind to a local variable so it can be used in a goroutine without being
			// overwritten. See https://go.dev/doc/faq#closures_and_goroutines
			batch := batch

			c <- func() error {
				_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					span.BufferWrite(ctx, batch.testVariantRealms...)
					return nil
				})
				if err != nil {
					return errors.Wrap(err, "inserting test variant realms")
				}
				_, err = span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					span.BufferWrite(ctx, batch.testResults...)
					return nil
				})
				if err != nil {
					return errors.Wrap(err, "inserting test results")
				}
				return nil
			}
		}
	})
}
