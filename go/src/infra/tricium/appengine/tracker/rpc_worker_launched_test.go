// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tracker

import (
	"strings"
	"testing"

	ds "github.com/luci/gae/service/datastore"

	. "github.com/smartystreets/goconvey/convey"

	"infra/tricium/api/admin/v1"
	"infra/tricium/api/v1"
	trit "infra/tricium/appengine/common/testing"
	"infra/tricium/appengine/common/track"
)

func TestWorkerLaunchedRequest(t *testing.T) {
	Convey("Test Environment", t, func() {
		tt := &trit.Testing{}
		ctx := tt.Context()

		// Add pending run entry.
		run := &track.Run{}
		So(ds.Put(ctx, run), ShouldBeNil)
		runKey := ds.KeyForObj(ctx, run)
		So(ds.Put(ctx, &track.RunResult{
			ID:     "1",
			Parent: runKey,
			State:  tricium.State_PENDING,
		}), ShouldBeNil)

		// Mark workflow as launched and add tracking entries for workers.
		err := workflowLaunched(ctx, &admin.WorkflowLaunchedRequest{
			RunId: run.ID,
		}, mockWorkflowProvider{})
		So(err, ShouldBeNil)

		// Mark worker as launched.
		err = workerLaunched(ctx, &admin.WorkerLaunchedRequest{
			RunId:  run.ID,
			Worker: fileIsolator,
		})
		So(err, ShouldBeNil)

		Convey("Marks worker as launched", func() {
			analyzerName := strings.Split(fileIsolator, "_")[0]
			analyzerKey := ds.NewKey(ctx, "AnalyzerRun", analyzerName, 0, runKey)
			workerKey := ds.NewKey(ctx, "WorkerRun", fileIsolator, 0, analyzerKey)
			wr := &track.WorkerResult{
				ID:     "1",
				Parent: workerKey,
			}
			err = ds.Get(ctx, wr)
			So(err, ShouldBeNil)
			So(wr.State, ShouldEqual, tricium.State_RUNNING)
			ar := &track.AnalyzerResult{
				ID:     "1",
				Parent: analyzerKey,
			}
			err = ds.Get(ctx, ar)
			So(err, ShouldBeNil)
			So(ar.State, ShouldEqual, tricium.State_RUNNING)
		})
	})
}
