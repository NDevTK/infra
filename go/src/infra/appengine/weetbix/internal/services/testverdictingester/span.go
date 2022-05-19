// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/pkg/errors"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
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
	// The test variant realms to insert/update if they are stale.
	// Test variant realms should be inserted before any test results.
	testVariantRealms []testresults.TestVariantRealm
	// Test results to insert. Already prepared as Spanner mutations.
	testResults []*spanner.Mutation
}

func batchTestResults(inv *testresults.IngestedInvocation, inputC chan []*rdbpb.TestVariant, outputC chan batch) {
	// Must be selected such that no more than 20,000 mutations occur in
	// one transaction in the worst case.
	const batchSize = 1000

	var trs []*spanner.Mutation
	var tvrs []testresults.TestVariantRealm
	startBatch := func() {
		trs = make([]*spanner.Mutation, 0, batchSize)
		tvrs = make([]testresults.TestVariantRealm, 0, batchSize)
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
	for tvs := range inputC {
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
			tvrs = append(tvrs, tvr)

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
	}
	if len(trs) > 0 {
		outputBatch()
	}
}

var testResultIngestionRandomnessKey = "used in tests only for forcing ensureTestVariantRealms determinism"

// ensureTestVariantRealms ensures the specified test variant realm entries exist with
// a recent LastIngestionTime.
// The implementation assumes the Project and Subrealm of every TestVariantRealm
// passed to this method is the same.
func ensureTestVariantRealms(ctx context.Context, tvrs []testresults.TestVariantRealm) error {
	// Look up the test variant realms to check if they already exist.
	keys := make([]spanner.Key, 0, len(tvrs))
	for _, tvr := range tvrs {
		keys = append(keys, spanner.Key{tvr.Project, tvr.TestID, tvr.VariantHash, tvr.SubRealm})
	}

	now := clock.Now(ctx)

	updateProb := rand.Float64()
	if fixedProb, ok := ctx.Value(&testResultIngestionRandomnessKey).(float64); ok {
		updateProb = fixedProb
	}

	// No need to read back to the Project and SubRealm as it is the same for all
	// entries.
	ks := spanner.KeySetFromKeys(keys...)
	cols := []string{"TestId", "VariantHash", "LastIngestionTime"}
	it := span.Read(span.Single(ctx), "TestVariantRealms", ks, cols)

	// For each test, the list of variants that do not require an update to the
	// TestVariant realm entry. (Realm is not required to be stored, as it is the
	// same for all entries.)
	freshTestVariants := make(map[string][]string, len(tvrs))
	updatesRequired := len(tvrs)

	err := it.Do(func(r *spanner.Row) error {
		var testID, variantHash string
		var lastIngestionTime time.Time
		err := r.Columns(&testID, &variantHash, &lastIngestionTime)
		if err != nil {
			return err
		}
		d := now.Sub(lastIngestionTime)
		if durationSinceUpdateToUpdateProbability(d) >= updateProb {
			updatesRequired++
		} else {
			freshTestVariants[testID] = append(freshTestVariants[testID], variantHash)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if updatesRequired == 0 {
		return nil
	}

	// Insert/update the test variant realm entries which don't exist/are stale.
	ms := make([]*spanner.Mutation, 0, updatesRequired)
	for _, tvr := range tvrs {
		freshVariants := freshTestVariants[tvr.TestID]
		requiresUpdate := true
		for _, vh := range freshVariants {
			if vh == tvr.VariantHash {
				requiresUpdate = false
			}
		}
		if requiresUpdate {
			ms = append(ms, tvr.SaveUnverified())
		}
	}

	_, err = span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		span.BufferWrite(ctx, ms...)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// durationSinceUpdateToUpdateProbability computes the probability with which
// a test variant realm record should be updated.
// The returned probability is a value between 0.0 and 1.0. The returned
// probabilities are designed to spread out updates to the records in a way
// that minimises the risk of concurrent updates.
func durationSinceUpdateToUpdateProbability(d time.Duration) float64 {
	if d < 0 {
		// Never update if stored time is after now.
		return 0.0
	}
	if d > time.Hour*12 {
		// Always update if older than 12 hours.
		return 1.0
	}
	// As the time since last update increases from 0 hours to 12 hours,
	// gradually scale up the probability we will update the test variant,
	// from (1 in a million) to (certainty). Use exponential scaling as this
	// minimises the probability of contention over a wide range of test
	// result insert rates for the test variant.
	// (Lack of updates to a test variant realm record can be taken as
	// an indication there are relatively fewer tasks inserting test results
	// for it, so it is safe to increase the fraction of tasks that will
	// update it.)
	fractionUntilUpdate := d.Seconds() / (12 * 60 * 60)
	exp := -1 + fractionUntilUpdate
	return math.Pow(1000*1000, exp)
}

// recordTestResults records test results from an test-verdict-ingestion task.
func recordTestResults(ctx context.Context, inv *testresults.IngestedInvocation, inputC chan []*rdbpb.TestVariant) error {
	const workerCount = 8

	return parallel.WorkPool(workerCount, func(c chan<- func() error) {
		batchC := make(chan batch)

		c <- func() error {
			defer close(batchC)
			batchTestResults(inv, inputC, batchC)
			return nil
		}

		for batch := range batchC {
			// Bind to a local variable so it can be used in a goroutine without being
			// overwritten. See https://go.dev/doc/faq#closures_and_goroutines
			batch := batch

			c <- func() error {
				err := ensureTestVariantRealms(ctx, batch.testVariantRealms)
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
