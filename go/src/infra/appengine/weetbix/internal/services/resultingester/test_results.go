// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resultingester

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/sync/parallel"
	"go.chromium.org/luci/common/trace"
	rdbpbutil "go.chromium.org/luci/resultdb/pbutil"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/ingestion/resultdb"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testresults"
	"infra/appengine/weetbix/internal/testresults/gitreferences"
	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
	"infra/appengine/weetbix/utils"
)

// maximumCLs is the maximum number of changelists to capture from any
// buildbucket run, after which the CL list is truncated. This avoids builds
// with an excessive number of included CLs from storing an excessive amount
// of data per failure.
const maximumCLs = 10

// extractGitReference extracts the git reference used to number the commit
// tested by the given build.
func extractGitReference(project string, commit *bbpb.GitilesCommit) *gitreferences.GitReference {
	return &gitreferences.GitReference{
		Project:          project,
		GitReferenceHash: gitreferences.GitReferenceHash(commit.Host, commit.Project, commit.Ref),
		Hostname:         commit.Host,
		Repository:       commit.Project,
		Reference:        commit.Ref,
	}
}

// extractIngestionContext extracts the ingested invocation and
// the git reference tested (if any).
func extractIngestionContext(task *taskspb.IngestTestResults, build *bbpb.Build, inv *rdbpb.Invocation) (*testresults.IngestedInvocation, *gitreferences.GitReference, error) {
	invID, err := rdbpbutil.ParseInvocationName(inv.Name)
	if err != nil {
		// This should never happen. Inv was originated from ResultDB.
		panic(err)
	}

	proj, subRealm := utils.SplitRealm(inv.Realm)
	if proj == "" {
		return nil, nil, errors.Reason("invocation has invalid realm: %q", inv.Realm).Err()
	}
	if proj != task.Build.Project {
		return nil, nil, errors.Reason("invocation project (%q) does not match build project (%q) for build %s-%d",
			proj, task.Build.Project, task.Build.Host, task.Build.Id).Err()
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
		return nil, nil, fmt.Errorf("build has unknown status: %v", build.Status)
	}

	gerritChanges := build.GetInput().GetGerritChanges()
	changelists := make([]testresults.Changelist, 0, len(gerritChanges))
	for _, change := range gerritChanges {
		if !strings.HasSuffix(change.Host, testresults.GerritHostnameSuffix) {
			return nil, nil, fmt.Errorf(`gerrit host %q does not end in expected suffix %q`, change.Host, testresults.GerritHostnameSuffix)
		}
		host := strings.TrimSuffix(change.Host, testresults.GerritHostnameSuffix)
		changelists = append(changelists, testresults.Changelist{
			Host:     host,
			Change:   change.Change,
			Patchset: change.Patchset,
		})
	}
	// Store the tested changelists in sorted order. This ensures that for
	// the same combination of CLs tested, the arrays are identical.
	testresults.SortChangelists(changelists)

	// Truncate the list of changelists to avoid storing an excessive number.
	// Apply truncation after sorting to ensure a stable set of changelists.
	if len(changelists) > maximumCLs {
		changelists = changelists[:maximumCLs]
	}

	var presubmitRun *testresults.PresubmitRun
	if task.PresubmitRun != nil {
		presubmitRun = &testresults.PresubmitRun{
			Mode:  task.PresubmitRun.Mode,
			Owner: task.PresubmitRun.Owner,
		}
	}

	invocation := &testresults.IngestedInvocation{
		Project:              proj,
		IngestedInvocationID: invID,
		SubRealm:             subRealm,
		PartitionTime:        task.PartitionTime.AsTime(),
		BuildStatus:          buildStatus,
		PresubmitRun:         presubmitRun,
		Changelists:          changelists,
	}

	commit := build.Output.GetGitilesCommit()
	if commit == nil {
		commit = build.Input.GetGitilesCommit()
	}
	var gitRef *gitreferences.GitReference
	if commit != nil {
		gitRef = extractGitReference(proj, commit)
		invocation.GitReferenceHash = gitRef.GitReferenceHash
		invocation.CommitPosition = int64(commit.Position)
		invocation.CommitHash = strings.ToLower(commit.Id)
	}
	return invocation, gitRef, nil
}

func recordIngestionContext(ctx context.Context, inv *testresults.IngestedInvocation, gitRef *gitreferences.GitReference) error {
	// Update the IngestedInvocations table.
	m := inv.SaveUnverified()

	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		span.BufferWrite(ctx, m)

		if gitRef != nil {
			// Ensure the git reference (if any) exists in the GitReferences table.
			if err := gitreferences.EnsureExists(ctx, gitRef); err != nil {
				return errors.Annotate(err, "ensuring git reference").Err()
			}
		}
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
	const batchSize = 900

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

		exonerationReasons := make([]pb.ExonerationReason, 0, len(tv.Exonerations))
		for _, ex := range tv.Exonerations {
			exonerationReasons = append(exonerationReasons, pbutil.ExonerationReasonFromResultDB(ex.Reason))
		}

		// Group results into test runs and order them by start time.
		resultsByRun := resultdb.GroupAndOrderTestResults(tv.Results)
		for runIndex, run := range resultsByRun {
			for resultIndex, inputTR := range run {
				tr := testresults.TestResult{
					Project:              inv.Project,
					TestID:               tv.TestId,
					PartitionTime:        inv.PartitionTime,
					VariantHash:          tv.VariantHash,
					IngestedInvocationID: inv.IngestedInvocationID,
					RunIndex:             int64(runIndex),
					ResultIndex:          int64(resultIndex),
					IsUnexpected:         !inputTR.Result.Expected,
					Status:               pbutil.TestResultStatusFromResultDB(inputTR.Result.Status),
					ExonerationReasons:   exonerationReasons,
					SubRealm:             inv.SubRealm,
					BuildStatus:          inv.BuildStatus,
					PresubmitRun:         inv.PresubmitRun,
					GitReferenceHash:     inv.GitReferenceHash,
					CommitPosition:       inv.CommitPosition,
					Changelists:          inv.Changelists,
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
func recordTestResults(ctx context.Context, inv *testresults.IngestedInvocation, tvs []*rdbpb.TestVariant) (err error) {
	ctx, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/services/resultingester.recordTestResults")
	defer func() { s.End(err) }()

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
					return errors.Annotate(err, "inserting test variant realms").Err()
				}
				_, err = span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
					span.BufferWrite(ctx, batch.testResults...)
					return nil
				})
				if err != nil {
					return errors.Annotate(err, "inserting test results").Err()
				}
				return nil
			}
		}
	})
}
