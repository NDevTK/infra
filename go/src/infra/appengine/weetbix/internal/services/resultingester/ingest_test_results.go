// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resultingester

import (
	"context"
	"fmt"
	"time"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/common/trace"
	"go.chromium.org/luci/common/tsmon/field"
	"go.chromium.org/luci/common/tsmon/metric"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/span"
	"go.chromium.org/luci/server/tq"
	"golang.org/x/sync/semaphore"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/analysis/clusteredfailures"
	"infra/appengine/weetbix/internal/buildbucket"
	"infra/appengine/weetbix/internal/clustering/chunkstore"
	"infra/appengine/weetbix/internal/clustering/ingestion"
	"infra/appengine/weetbix/internal/config"
	"infra/appengine/weetbix/internal/ingestion/control"
	"infra/appengine/weetbix/internal/resultdb"
	"infra/appengine/weetbix/internal/services/resultcollector"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/internal/testresults"
	pb "infra/appengine/weetbix/proto/v1"
)

const (
	resultIngestionTaskClass = "result-ingestion"
	resultIngestionQueue     = "result-ingestion"

	// ingestionEarliest is the oldest data that may be ingested by Weetbix.
	// This is an offset relative to the current time, and should be kept
	// in sync with the data retention period in Spanner and BigQuery.
	ingestionEarliest = -90 * 24 * time.Hour

	// ingestionLatest is the newest data that may be ingested by Weetbix.
	// This is an offset relative to the current time. It is designed to
	// allow for clock drift.
	ingestionLatest = 24 * time.Hour
)

var (
	taskCounter = metric.NewCounter(
		"weetbix/ingestion/task_completion",
		"The number of completed Weetbix ingestion tasks, by build project and outcome.",
		nil,
		// The LUCI Project.
		field.String("project"),
		// "success", "failed_validation",
		// "ignored_no_bb_access", "ignored_no_project_config",
		// "ignored_no_invocation", "ignored_has_ancestor".
		field.String("outcome"))

	ancestorCounter = metric.NewCounter(
		"weetbix/ingestion/ancestor_build_status",
		"The status retrieving ancestor builds in ingestion tasks, by build project.",
		nil,
		// The LUCI Project.
		field.String("project"),
		// "no_bb_access_to_ancestor",
		// "no_resultdb_invocation_on_ancestor",
		// "ok".
		field.String("ancestor_status"))

	testVariantReadMask = &fieldmaskpb.FieldMask{
		Paths: []string{
			"test_id",
			"variant_hash",
			"status",
			"variant",
			"test_metadata",
			"exonerations.*.reason",
			"results.*.result.name",
			"results.*.result.expected",
			"results.*.result.status",
			"results.*.result.start_time",
			"results.*.result.duration",
			"results.*.result.tags",
			"results.*.result.failure_reason",
		},
	}

	buildReadMask = &field_mask.FieldMask{
		Paths: []string{"builder", "infra.resultdb", "status", "input", "output", "ancestor_ids"},
	}
)

// Options configures test result ingestion.
type Options struct {
}

type resultIngester struct {
	clustering *ingestion.Ingester
}

var resultIngestion = tq.RegisterTaskClass(tq.TaskClass{
	ID:        resultIngestionTaskClass,
	Prototype: &taskspb.IngestTestResults{},
	Queue:     resultIngestionQueue,
	Kind:      tq.Transactional,
})

// RegisterTaskHandler registers the handler for result ingestion tasks.
func RegisterTaskHandler(srv *server.Server) error {
	ctx := srv.Context
	cfg, err := config.Get(ctx)
	if err != nil {
		return err
	}
	chunkStore, err := chunkstore.NewClient(ctx, cfg.ChunkGcsBucket)
	if err != nil {
		return err
	}
	srv.RegisterCleanup(func(ctx context.Context) {
		chunkStore.Close()
	})
	cf := clusteredfailures.NewClient(srv.Options.CloudProject)
	analysis := analysis.NewClusteringHandler(cf)
	ri := &resultIngester{
		clustering: ingestion.New(chunkStore, analysis),
	}
	handler := func(ctx context.Context, payload proto.Message) error {
		task := payload.(*taskspb.IngestTestResults)
		return ri.ingestTestResults(ctx, task)
	}
	resultIngestion.AttachHandler(handler)
	return nil
}

// Schedule enqueues a task to ingest test results from a build.
func Schedule(ctx context.Context, task *taskspb.IngestTestResults) {
	tq.MustAddTask(ctx, &tq.Task{
		Title:   fmt.Sprintf("%s-%s-%d-page-%v", task.Build.Project, task.Build.Host, task.Build.Id, task.TaskIndex),
		Payload: task,
	})
}

// requestLimiter limits the number of concurrent result ingestion requests.
// This is to ensure the instance remains within GAE memory limits.
// These requests are larger than others and latency is not critical,
// so using a semaphore to limit throughput was deemed overall better than
// limiting the number of concurrent requests to the instance as a whole.
var requestLimiter = semaphore.NewWeighted(5)

func (i *resultIngester) ingestTestResults(ctx context.Context, payload *taskspb.IngestTestResults) error {
	if err := validateRequest(ctx, payload); err != nil {
		project := "(unknown)"
		if payload.GetBuild().GetProject() != "" {
			project = payload.Build.Project
		}
		taskCounter.Add(ctx, 1, project, "failed_validation")
		return tq.Fatal.Apply(err)
	}

	// Limit the number of concurrent requests in the following section.
	err := requestLimiter.Acquire(ctx, 1)
	if err != nil {
		return transient.Tag.Apply(err)
	}
	defer requestLimiter.Release(1)

	// Buildbucket build only has builder, infra.resultdb, status populated.
	build, err := retrieveBuild(ctx, payload.Build.Host, payload.Build.Id)
	code := status.Code(err)
	if code == codes.NotFound {
		// Build not found, end the task gracefully.
		logging.Warningf(ctx, "Buildbucket build %s/%d for project %s not found (or Weetbix does not have access to read it).",
			payload.Build.Host, payload.Build.Id, payload.Build.Project)
		taskCounter.Add(ctx, 1, payload.Build.Project, "ignored_no_bb_access")
		return nil
	}
	if err != nil {
		return transient.Tag.Apply(err)
	}

	if build.Infra.GetResultdb().GetInvocation() == "" {
		// Build does not have a ResultDB invocation to ingest.
		logging.Debugf(ctx, "Skipping ingestion of build %s-%d because it has no ResultDB invocation.",
			payload.Build.Host, payload.Build.Id)
		taskCounter.Add(ctx, 1, payload.Build.Project, "ignored_no_invocation")
		return nil
	}

	if payload.TaskIndex == 0 {
		// Before ingesting any of the build. If we are already ingesting the
		// build (TaskIndex > 0), we made it past this check before.
		if len(build.AncestorIds) > 0 {
			// If the build has an ancestor build, see if its immediate
			// ancestor is accessible by Weetbix and has a ResultDB invocation
			// (likely indicating it includes the test results from this
			// build).
			included, err := includedByAncestorBuild(ctx, payload.Build.Host, build.AncestorIds[len(build.AncestorIds)-1], payload.Build.Project)
			if err != nil {
				return transient.Tag.Apply(err)
			}
			if included {
				// Yes. Do not ingest this build to avoid ingesting the same test
				// results multiple times.
				taskCounter.Add(ctx, 1, payload.Build.Project, "ignored_has_ancestor")
				return nil
			}
		}
	}

	rdbHost := build.Infra.Resultdb.Hostname
	invName := build.Infra.Resultdb.Invocation
	builder := build.Builder.Builder
	rc, err := resultdb.NewClient(ctx, rdbHost)
	if err != nil {
		return transient.Tag.Apply(err)
	}
	inv, err := rc.GetInvocation(ctx, invName)
	code = status.Code(err)
	if code == codes.NotFound {
		// Invocation not found, end the task gracefully.
		logging.Warningf(ctx, "Invocation %s for project %s not found (or Weetbix does not have access to read it).",
			invName, payload.Build.Project)
		taskCounter.Add(ctx, 1, payload.Build.Project, "ignored_no_resultdb_access")
		return nil
	}
	if err != nil {
		return transient.Tag.Apply(err)
	}

	ingestedInv, gitRef, err := extractIngestionContext(payload, build, inv)
	if err != nil {
		return err
	}

	if payload.TaskIndex == 0 {
		// The first task should create the ingested invocation record
		// and git reference record referenced from the invocation record
		// (if any).
		err = recordIngestionContext(ctx, ingestedInv, gitRef)
		if err != nil {
			return err
		}
	}

	// Query test variants from ResultDB.
	req := &rdbpb.QueryTestVariantsRequest{
		Invocations: []string{inv.Name},
		PageSize:    10000,
		ReadMask:    testVariantReadMask,
		PageToken:   payload.PageToken,
	}
	rsp, err := rc.QueryTestVariants(ctx, req)
	if err != nil {
		err = errors.Annotate(err, "query test variants").Err()
		return transient.Tag.Apply(err)
	}

	// Schedule a task to deal with the next page of results (if needed).
	// Do this immediately, so that task can commence while we are still
	// inserting the results for this page.
	if rsp.NextPageToken != "" {
		if err := scheduleNextTask(ctx, payload, rsp.NextPageToken); err != nil {
			err = errors.Annotate(err, "schedule next task").Err()
			return transient.Tag.Apply(err)
		}
	}

	// Record the test results for test history.
	err = recordTestResults(ctx, ingestedInv, rsp.TestVariants)
	if err != nil {
		// If any transaction failed, the task will be retried and the tables will be
		// eventual-consistent.
		return errors.Annotate(err, "record test results").Err()
	}

	failingTVs := filterToTestVariantsWithUnexpectedFailures(rsp.TestVariants)
	nextPageToken := rsp.NextPageToken
	// Allow garbage collector to free test variants except for those that are
	// unexpected.
	rsp = nil

	// Insert the test results for clustering.
	err = ingestForClustering(ctx, i.clustering, payload, ingestedInv, failingTVs)
	if err != nil {
		return err
	}

	// Ingest for test variant analysis.
	realmCfg, err := config.Realm(ctx, inv.Realm)
	if err != nil && err != config.RealmNotExistsErr {
		return transient.Tag.Apply(err)
	}

	ingestForTestVariantAnalysis := realmCfg != nil &&
		shouldIngestForTestVariants(realmCfg, payload)

	if ingestForTestVariantAnalysis {
		if err := createOrUpdateAnalyzedTestVariants(ctx, inv.Realm, builder, failingTVs); err != nil {
			err = errors.Annotate(err, "ingesting for test variant analysis").Err()
			return transient.Tag.Apply(err)
		}

		if nextPageToken == "" {
			// In the last task, after all test variants ingested.
			isPreSubmit := payload.PresubmitRun != nil
			contributedToCLSubmission := payload.PresubmitRun != nil &&
				payload.PresubmitRun.Mode == pb.PresubmitRunMode_FULL_RUN &&
				payload.PresubmitRun.Status == pb.PresubmitRunStatus_PRESUBMIT_RUN_STATUS_SUCCEEDED
			if err = resultcollector.Schedule(ctx, inv, rdbHost, build.Builder.Builder, isPreSubmit, contributedToCLSubmission); err != nil {
				return transient.Tag.Apply(err)
			}
		}
	}

	if nextPageToken == "" {
		// In the last task.
		taskCounter.Add(ctx, 1, payload.Build.Project, "success")
	}
	return nil
}

func includedByAncestorBuild(ctx context.Context, buildHost string, buildID int64, project string) (bool, error) {
	// Retrieve the ancestor build.
	rootBuild, err := retrieveBuild(ctx, buildHost, buildID)
	code := status.Code(err)
	if code == codes.NotFound {
		logging.Warningf(ctx, "Buildbucket ancestor build %s/%d for project %s not found (or Weetbix does not have access to read it).",
			buildHost, buildID, project)
		// Weetbix won't be able to retrieve the ancestor build to ingest it,
		// even if it did include the test results from this build.

		ancestorCounter.Add(ctx, 1, project, "no_bb_access_to_ancestor")
		return false, nil
	}
	if err != nil {
		return false, errors.Annotate(err, "retrieving ancestor build").Err()
	}
	if rootBuild.Infra.GetResultdb().GetInvocation() == "" {
		ancestorCounter.Add(ctx, 1, project, "no_resultdb_invocation_on_ancestor")
		return false, nil
	}

	// The ancestor build also has a ResultDB invocation. This is what
	// we expected. We will ingest the ancestor build only
	// to avoid ingesting the same test results multiple times.
	ancestorCounter.Add(ctx, 1, project, "ok")
	return true, nil
}

// filterToTestVariantsWithUnexpectedFailures filters the given list of
// test variants to only those with unexpected failures.
func filterToTestVariantsWithUnexpectedFailures(tvs []*rdbpb.TestVariant) []*rdbpb.TestVariant {
	var results []*rdbpb.TestVariant
	for _, tv := range tvs {
		if hasUnexpectedFailures(tv) {
			results = append(results, tv)
		}
	}
	return results
}

// scheduleNextTask schedules a task to continue the ingestion,
// starting at the given page token.
// If a continuation task for this task has been previously scheduled
// (e.g. in a previous try of this task), this method does nothing.
func scheduleNextTask(ctx context.Context, task *taskspb.IngestTestResults, nextPageToken string) error {
	if nextPageToken == "" {
		// If the next page token is "", it means ResultDB returned the
		// last page. We should not schedule a continuation task.
		panic("next page token cannot be the empty page token")
	}
	buildID := control.BuildID(task.Build.Host, task.Build.Id)

	// Schedule the task transactionally, conditioned on it not having been
	// scheduled before.
	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		entries, err := control.Read(ctx, []string{buildID})
		if err != nil {
			return errors.Annotate(err, "read ingestion record").Err()
		}

		entry := entries[0]
		if entry == nil {
			return errors.Reason("build %v does not have ingestion record", buildID).Err()
		}
		if task.TaskIndex >= entry.TaskCount {
			// This should nver happen.
			panic("current ingestion task not recorded on ingestion control record")
		}
		nextTaskIndex := task.TaskIndex + 1
		if nextTaskIndex != entry.TaskCount {
			// Next task has already been created in the past. Do not create
			// it again.
			// This can happen if the ingestion task failed after
			// it scheduled the ingestion task for the next page,
			// and was subsequently retried.
			return nil
		}
		entry.TaskCount = entry.TaskCount + 1
		if err := control.InsertOrUpdate(ctx, entry); err != nil {
			return errors.Annotate(err, "update ingestion record").Err()
		}

		itrTask := &taskspb.IngestTestResults{
			PartitionTime: task.PartitionTime,
			Build:         task.Build,
			PresubmitRun:  task.PresubmitRun,
			PageToken:     nextPageToken,
			TaskIndex:     nextTaskIndex,
		}
		Schedule(ctx, itrTask)

		return nil
	})
	return err
}

func ingestForClustering(ctx context.Context, clustering *ingestion.Ingester, payload *taskspb.IngestTestResults, inv *testresults.IngestedInvocation, tvs []*rdbpb.TestVariant) (err error) {
	if payload.PresubmitRun != nil && payload.PresubmitRun.Mode != pb.PresubmitRunMode_FULL_RUN {
		// Do not ingest dry run data.
		return nil
	}

	ctx, s := trace.StartSpan(ctx, "infra/appengine/weetbix/internal/services/resultingester.ingestForClustering")
	defer func() { s.End(err) }()

	if _, err := config.Project(ctx, payload.Build.Project); err != nil {
		if err == config.NotExistsErr {
			// Project not configured in Weetbix, ignore it.
			return nil
		} else {
			// Transient error.
			return transient.Tag.Apply(errors.Annotate(err, "get project config").Err())
		}
	}

	changelists := make([]*pb.Changelist, 0, len(inv.Changelists))
	for _, cl := range inv.Changelists {
		changelists = append(changelists, &pb.Changelist{
			Host:     cl.Host + "-review.googlesource.com",
			Change:   cl.Change,
			Patchset: int32(cl.Patchset),
		})
	}

	// Setup clustering ingestion.
	opts := ingestion.Options{
		TaskIndex:     payload.TaskIndex,
		Project:       inv.Project,
		PartitionTime: inv.PartitionTime,
		Realm:         inv.Project + ":" + inv.SubRealm,
		InvocationID:  inv.IngestedInvocationID,
		BuildStatus:   inv.BuildStatus,
		Changelists:   changelists,
	}

	if payload.PresubmitRun != nil {
		opts.PresubmitRun = &ingestion.PresubmitRun{
			ID:     payload.PresubmitRun.PresubmitRunId,
			Owner:  payload.PresubmitRun.Owner,
			Mode:   payload.PresubmitRun.Mode,
			Status: payload.PresubmitRun.Status,
		}
		opts.BuildCritical = payload.PresubmitRun.Critical
		if payload.PresubmitRun.Critical && inv.BuildStatus == pb.BuildStatus_BUILD_STATUS_FAILURE &&
			payload.PresubmitRun.Status == pb.PresubmitRunStatus_PRESUBMIT_RUN_STATUS_SUCCEEDED {
			logging.Warningf(ctx, "Inconsistent data from LUCI CV: build %v/%v was critical to presubmit run %v/%v and failed, but presubmit run succeeded.",
				payload.Build.Host, payload.Build.Id, payload.PresubmitRun.PresubmitRunId.System, payload.PresubmitRun.PresubmitRunId.Id)
		}
	}
	// Clustering ingestion is designed to behave gracefully in case of
	// a task retry. Given the same options and same test variants (in
	// the same order), the IDs and content of the chunks it writes is
	// designed to be stable. If chunks already exist, it will skip them.
	if err := clustering.Ingest(ctx, opts, tvs); err != nil {
		err = errors.Annotate(err, "ingesting for clustering").Err()
		return transient.Tag.Apply(err)
	}
	return nil
}

func validateRequest(ctx context.Context, payload *taskspb.IngestTestResults) error {
	if !payload.PartitionTime.IsValid() {
		return errors.New("partition time must be specified and valid")
	}
	t := payload.PartitionTime.AsTime()
	now := clock.Now(ctx)
	if t.Before(now.Add(ingestionEarliest)) {
		return fmt.Errorf("partition time (%v) is too long ago", t)
	} else if t.After(now.Add(ingestionLatest)) {
		return fmt.Errorf("partition time (%v) is too far in the future", t)
	}
	if payload.Build == nil {
		return errors.New("build must be specified")
	}
	if payload.Build.Project == "" {
		return errors.New("project must be specified")
	}
	return nil
}

func retrieveBuild(ctx context.Context, bbHost string, id int64) (*bbpb.Build, error) {
	bc, err := buildbucket.NewClient(ctx, bbHost)
	if err != nil {
		return nil, err
	}
	request := &bbpb.GetBuildRequest{
		Id: id,
		Mask: &bbpb.BuildMask{
			Fields: buildReadMask,
		},
	}
	b, err := bc.GetBuild(ctx, request)
	switch {
	case err != nil:
		return nil, err
	}
	return b, nil
}
