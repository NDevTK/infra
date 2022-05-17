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
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/sync/parallel"
	rdbpbutil "go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/internal/testverdicts"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
	"infra/appengine/weetbix/utils"
)

func recordIngestedInvocation(ctx context.Context, task *taskspb.IngestTestVerdicts, build *bbpb.Build, inv *rdbpb.Invocation) error {
	invID, err := rdbpbutil.ParseInvocationName(inv.Name)
	if err != nil {
		// This should never happen. Inv was originated from ResultDB.
		panic(err)
	}

	proj, subRealm := utils.SplitRealm(inv.Realm)
	contributedToCLSubmission := task.GetPresubmitRun().GetPresubmitRunSucceeded()
	gerritChanges := build.GetInput().GetGerritChanges()

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
		return fmt.Errorf("build has unknown status: %v", build.Status)
	}

	// Update the IngestedInvocations table.
	ingestedInv := &testresults.IngestedInvocation{
		Project:                      proj,
		IngestedInvocationID:         invID,
		SubRealm:                     subRealm,
		PartitionTime:                task.PartitionTime.AsTime(),
		BuildStatus:                  buildStatus,
		HasContributedToClSubmission: contributedToCLSubmission,
	}
	if len(gerritChanges) > 0 {
		// TODO(meiring): Ingest all gerrit changes.
		change := gerritChanges[0]
		if !strings.HasSuffix(change.Host, "-review.googlesource.com") {
			return fmt.Errorf(`Gerrit host %q does not end in expected suffix "-review.googlesource.com"`, change.Host)
		}
		host := strings.TrimSuffix(change.Host, "-review.googlesource.com")
		ingestedInv.ChangelistHost = spanner.NullString{StringVal: host, Valid: true}
		ingestedInv.ChangelistChange = spanner.NullInt64{Int64: change.Change, Valid: true}
		ingestedInv.ChangelistPatchset = spanner.NullInt64{Int64: change.Patchset, Valid: true}
	}

	_, err = span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		ingestedInv.SaveUnverified(ctx)
		return nil
	})
	return err
}

// recordTestVerdicts records test verdicts from an test-verdict-ingestion task.
func recordTestVerdicts(ctx context.Context, task *taskspb.IngestTestVerdicts, build *bbpb.Build, inv *rdbpb.Invocation, tvsC chan []*rdbpb.TestVariant) error {
	const (
		workerCount = 8
		batchSize   = 1000
	)

	invId, err := rdbpbutil.ParseInvocationName(inv.Name)
	if err != nil {
		return err
	}

	proj, subRealm := utils.SplitRealm(inv.Realm)
	contributedToCLSubmission := task.GetPresubmitRun().GetPresubmitRunSucceeded()
	hasUnsubmittedChanges := len(build.GetInput().GetGerritChanges()) != 0

	// recordBatch updates TestVerdicts table and TestVariantRealms table from a
	// batch of test variants. Must be called in a spanner RW transactional
	// context.
	recordBatch := func(ctx context.Context, batch []*rdbpb.TestVariant) error {
		for _, tv := range batch {
			// Record the test verdict.
			expectedCount, unexpectedCount, skippedCount := countResults(tv)
			verdict := &testverdicts.TestVerdict{
				Project:                      proj,
				TestID:                       tv.TestId,
				PartitionTime:                task.PartitionTime.AsTime(),
				VariantHash:                  tv.VariantHash,
				IngestedInvocationID:         invId,
				SubRealm:                     subRealm,
				ExpectedCount:                expectedCount,
				UnexpectedCount:              unexpectedCount,
				SkippedCount:                 skippedCount,
				IsExonerated:                 tv.Status == rdbpb.TestVariantStatus_EXONERATED,
				PassedAvgDuration:            calcPassedAvgDuration(tv),
				HasUnsubmittedChanges:        hasUnsubmittedChanges,
				HasContributedToClSubmission: contributedToCLSubmission,
			}
			verdict.SaveUnverified(ctx)

			// Record the test variant realm.
			tvr := &testresults.TestVariantRealm{
				Project:           proj,
				TestID:            tv.TestId,
				VariantHash:       tv.VariantHash,
				SubRealm:          subRealm,
				Variant:           pbutil.VariantFromResultDB(tv.Variant),
				LastIngestionTime: spanner.CommitTimestamp,
			}
			tvr.SaveUnverified(ctx)
		}
		return nil
	}

	return parallel.WorkPool(workerCount, func(c chan<- func() error) {
		batchC := make(chan []*rdbpb.TestVariant, workerCount)

		// Split test variants into smaller batches so we have less than 20k
		// mutations in a single spanner transaction.
		c <- func() error {
			defer close(batchC)
			for tvs := range tvsC {
				for i := 0; i < len(tvs); i += batchSize {
					end := i + batchSize
					if end > len(tvs) {
						end = len(tvs)
					}
					batchC <- tvs[i:end]
				}
			}
			return nil
		}

		for batch := range batchC {
			// Bind to a local variable so it can be used in a goroutine without being
			// overwritten. See https://go.dev/doc/faq#closures_and_goroutines
			batch := batch

			c <- func() error {
				_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					return recordBatch(ctx, batch)
				})
				return err
			}
		}
	})
}

func countResults(tv *rdbpb.TestVariant) (expected, unexpected, skipped int64) {
	for _, trb := range tv.Results {
		tr := trb.Result
		if tr.Status == rdbpb.TestStatus_SKIP {
			skipped++
		}
		if tr.Expected {
			expected++
		} else {
			unexpected++
		}
	}
	return
}

// calcPassedAvgDuration calculates the average duration of passed results.
// Return nil if there's no passed results.
func calcPassedAvgDuration(tv *rdbpb.TestVariant) *time.Duration {
	count := 0
	totalDuration := time.Duration(0)
	for _, trb := range tv.Results {
		tr := trb.Result
		if tr.Status != rdbpb.TestStatus_PASS {
			// Only calculate passed test results
			continue
		}
		count++
		totalDuration += tr.Duration.AsDuration()
	}
	if count == 0 {
		return nil
	}
	avgDuration := totalDuration / time.Duration(count)
	return &avgDuration
}
