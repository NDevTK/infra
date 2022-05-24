// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package testverdictingester

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/span"

	"infra/appengine/weetbix/internal/ingestion/control"
	"infra/appengine/weetbix/internal/tasks/taskspb"
)

// scheduleNextTask schedules a task to continue the ingestion,
// starting at the given page token.
// If a continuation task for this task has been previously scheduled
// (e.g. in a previous try of this task), this method does nothing.
func scheduleNextTask(ctx context.Context, task *taskspb.IngestTestVerdicts, nextPageToken string) error {
	if nextPageToken == "" {
		panic("next page token cannot be the initial page token")
	}
	buildID := control.BuildID(task.Build.Host, task.Build.Id)
	nextPageIndex := task.PageIndex + 1

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
		if entry.TaskCount != nextPageIndex {
			// Task has previously been created for this page. Do not create
			// it again.
			// This can happen if the ingestion task for a page failed after
			// it already scheduled the ingestion task for the next page, and
			// was subsequently retried.
			return nil
		}
		entry.TaskCount = entry.TaskCount + 1
		if err := control.InsertOrUpdate(ctx, entry); err != nil {
			return errors.Annotate(err, "update ingestion record").Err()
		}

		itvTask := &taskspb.IngestTestVerdicts{
			PartitionTime: task.PartitionTime,
			Build:         task.Build,
			PresubmitRun:  task.PresubmitRun,
			PageToken:     nextPageToken,
			PageIndex:     nextPageIndex,
		}
		Schedule(ctx, itvTask)

		return nil
	})
	return err
}
