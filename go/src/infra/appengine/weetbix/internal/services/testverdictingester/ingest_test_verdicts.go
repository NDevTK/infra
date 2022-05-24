// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"context"
	"fmt"
	"time"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/server/tq"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"infra/appengine/weetbix/internal/buildbucket"
	"infra/appengine/weetbix/internal/resultdb"
	"infra/appengine/weetbix/internal/tasks/taskspb"
	"infra/appengine/weetbix/utils"
)

const (
	taskClass = "test-verdict-ingestion"
	queue     = "test-verdict-ingestion"

	// ingestionEarliest is the oldest data that may be ingested by Weetbix.
	// This is an offset relative to the current time, and should be kept
	// in sync with the data retention period in Spanner and BigQuery.
	ingestionEarliest = -90 * 24 * time.Hour

	// ingestionLatest is the newest data that may be ingested by Weetbix.
	// This is an offset relative to the current time. It is designed to
	// allow for clock drift.
	ingestionLatest = 24 * time.Hour

	// maxResultDBPages is the maximum number of pages of test verdicts to ingest
	// from ResultDB, per build. The page size is 1000 test verdicts.
	maxResultDBPages = int(^uint(0) >> 1) // set to max int
)

var testVerdictIngestion = tq.RegisterTaskClass(tq.TaskClass{
	ID:        taskClass,
	Prototype: &taskspb.IngestTestVerdicts{},
	Queue:     queue,
	Kind:      tq.Transactional,
	Handler: func(ctx context.Context, payload proto.Message) error {
		task := payload.(*taskspb.IngestTestVerdicts)
		return ingestTestResults(ctx, task)
	},
})

// Schedule enqueues a task to get all the test results from an invocation,
// group them into test verdicts, and save them to the TestVerdicts table.
func Schedule(ctx context.Context, task *taskspb.IngestTestVerdicts) {
	tq.MustAddTask(ctx, &tq.Task{
		Title:   fmt.Sprintf("%s-%d-page-%d", task.Build.Host, task.Build.Id, task.PageIndex),
		Payload: task,
	})
}

func ingestTestResults(ctx context.Context, payload *taskspb.IngestTestVerdicts) error {
	if err := validateRequest(ctx, payload); err != nil {
		return err
	}

	// Buildbucket build only has input.gerrit_changes, infra.resultdb, status populated.
	build, err := retrieveBuild(ctx, payload)
	code := status.Code(err)
	if code == codes.NotFound {
		// Build not found, end the task gracefully.
		logging.Warningf(ctx, "Buildbucket build %s/%d for project %s not found (or Weetbix does not have access to read it).",
			payload.Build.Host, payload.Build.Id, payload.Build.Project)
		return nil
	}
	if err != nil {
		return err
	}

	if build.Infra.GetResultdb().GetInvocation() == "" {
		// Build does not have a ResultDB invocation to ingest.
		logging.Debugf(ctx, "Skipping ingestion of build %s-%d because it has no ResultDB invocation.",
			payload.Build.Host, payload.Build.Id)
		return nil
	}

	rdbHost := build.Infra.Resultdb.Hostname
	invName := build.Infra.Resultdb.Invocation
	rc, err := resultdb.NewClient(ctx, rdbHost)
	if err != nil {
		return err
	}
	inv, err := rc.GetInvocation(ctx, invName)
	if err != nil {
		return err
	}
	project, _ := utils.SplitRealm(inv.Realm)
	if project == "" {
		return fmt.Errorf("invocation has invalid realm: %q", inv.Realm)
	}

	ingestedInv, err := extractIngestedInvocation(payload, build, inv)
	if err != nil {
		return err
	}
	if payload.PageIndex == 0 {
		err = recordIngestedInvocation(ctx, ingestedInv)
		if err != nil {
			return err
		}
	}

	// Query test variants from ResultDB.
	req := &rdbpb.QueryTestVariantsRequest{
		Invocations: []string{invName},
		PageSize:    10000,
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"test_id",
				"variant_hash",
				"status",
				"variant",
				"results.*.result.name",
				"results.*.result.start_time",
				"results.*.result.status",
				"results.*.result.expected",
				"results.*.result.duration",
			},
		},
		PageToken: payload.PageToken,
	}
	rsp, err := rc.QueryTestVariants(ctx, req)
	if err != nil {
		return err
	}

	// Schedule a task to deal with the next page of results (if needed).
	// Do this immediately, so that task can commence while we are still
	// inserting the results for this page.
	if rsp.NextPageToken != "" {
		if err := scheduleNextTask(ctx, payload, rsp.NextPageToken); err != nil {
			return errors.Annotate(err, "schedule next task").Err()
		}
	}

	// Record the test results.
	err = recordTestResults(ctx, ingestedInv, rsp.TestVariants)
	if err != nil {
		// If any transaction failed, the task will be retried and the tables will be
		// eventual-consistent.
		return errors.Annotate(err, "record test results").Err()
	}

	return nil
}

func validateRequest(ctx context.Context, payload *taskspb.IngestTestVerdicts) error {
	if !payload.PartitionTime.IsValid() {
		return tq.Fatal.Apply(errors.New("partition time must be specified and valid"))
	}
	t := payload.PartitionTime.AsTime()
	now := clock.Now(ctx)
	if t.Before(now.Add(ingestionEarliest)) {
		return tq.Fatal.Apply(fmt.Errorf("partition time (%v) is too long ago", t))
	} else if t.After(now.Add(ingestionLatest)) {
		return tq.Fatal.Apply(fmt.Errorf("partition time (%v) is too far in the future", t))
	}
	if payload.Build == nil {
		return tq.Fatal.Apply(errors.New("build must be specified"))
	}
	return nil
}

func retrieveBuild(ctx context.Context, payload *taskspb.IngestTestVerdicts) (*bbpb.Build, error) {
	bbHost := payload.Build.Host
	id := payload.Build.Id
	bc, err := buildbucket.NewClient(ctx, bbHost)
	if err != nil {
		return nil, err
	}
	request := &bbpb.GetBuildRequest{
		Id: id,
		Mask: &bbpb.BuildMask{
			Fields: &field_mask.FieldMask{
				Paths: []string{"input.gerrit_changes", "infra.resultdb", "status"},
			},
		},
	}
	b, err := bc.GetBuild(ctx, request)
	switch {
	case err != nil:
		return nil, err
	}
	return b, nil
}
