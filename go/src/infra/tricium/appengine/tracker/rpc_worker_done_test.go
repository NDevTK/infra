// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	ds "github.com/luci/gae/service/datastore"

	. "github.com/smartystreets/goconvey/convey"

	"golang.org/x/net/context"

	"infra/tricium/api/admin/v1"
	"infra/tricium/api/v1"
	trit "infra/tricium/appengine/common/testing"
	"infra/tricium/appengine/common/track"
)

// MockIsolator mocks the Isolator interface for testing.
type mockIsolator struct{}

func (*mockIsolator) IsolateGitFileDetails(c context.Context, serverURL string, d *tricium.Data_GitFileDetails) (string, error) {
	return "mockmockmock", nil
}
func (*mockIsolator) IsolateWorker(c context.Context, serverURL string, worker *admin.Worker, inputIsolate string) (string, error) {
	return "mockmockmock", nil
}
func (*mockIsolator) LayerIsolates(c context.Context, serverURL, isolatedInput, isolatedOutput string) (string, error) {
	return "mockmockmock", nil
}
func (*mockIsolator) FetchIsolatedResults(c context.Context, serverURL, isolatedOutput string) (string, error) {
	result := &tricium.Data_Results{
		Comments: []*tricium.Data_Comment{
			{
				Message: "Hello",
			},
		},
	}
	res, err := json.Marshal(result)
	if err != nil {
		return "", errors.New("failed to marshall mock result")
	}
	return string(res), nil
}

func TestWorkerDoneRequest(t *testing.T) {
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

		// Mark workflow as launched.
		err := workflowLaunched(ctx, &admin.WorkflowLaunchedRequest{
			RunId: run.ID,
		}, mockWorkflowProvider{})
		So(err, ShouldBeNil)

		// test
		var analyzerResults []*track.AnalyzerResult
		So(ds.GetAll(ctx, ds.NewQuery("AnalyzerResult").Ancestor(runKey), &analyzerResults), ShouldBeNil)
		for _, a := range analyzerResults {
			fmt.Printf("1.AnalyzerResults in test: %v\n", a)
		}

		// Mark worker as launched.
		err = workerLaunched(ctx, &admin.WorkerLaunchedRequest{
			RunId:  run.ID,
			Worker: fileIsolator,
		})
		So(err, ShouldBeNil)

		// test
		var analyzerResults2 []*track.AnalyzerResult
		So(ds.GetAll(ctx, ds.NewQuery("AnalyzerResult").Ancestor(runKey), &analyzerResults2), ShouldBeNil)
		for _, a := range analyzerResults2 {
			fmt.Printf("2.AnalyzerResults in test: %v\n", a)
		}

		// Mark worker as done.
		err = workerDone(ctx, &admin.WorkerDoneRequest{
			RunId:    run.ID,
			Worker:   fileIsolator,
			ExitCode: 0,
		}, &mockIsolator{})
		So(err, ShouldBeNil)

		// test
		var analyzerResults3 []*track.AnalyzerResult
		So(ds.GetAll(ctx, ds.NewQuery("AnalyzerResult").Ancestor(runKey), &analyzerResults3), ShouldBeNil)
		for _, a := range analyzerResults3 {
			fmt.Printf("3.AnalyzerResults in test: %v\n", a)
		}

		Convey("Marks worker as done", func() {
			analyzerName := strings.Split(fileIsolator, "_")[0]
			analyzerKey := ds.NewKey(ctx, "AnalyzerRun", analyzerName, 0, runKey)
			workerKey := ds.NewKey(ctx, "WorkerRun", fileIsolator, 0, analyzerKey)
			wr := &track.WorkerResult{
				ID:     "1",
				Parent: workerKey,
			}
			err = ds.Get(ctx, wr)
			So(err, ShouldBeNil)
			So(wr.State, ShouldEqual, tricium.State_SUCCESS)
			ar := &track.AnalyzerResult{
				ID:     "1",
				Parent: analyzerKey,
			}
			err = ds.Get(ctx, ar)
			So(err, ShouldBeNil)
			So(ar.State, ShouldEqual, tricium.State_SUCCESS)
		})
		// TODO(emso): multi-platform analyzer is half done, analyzer stays launched
	})
}
