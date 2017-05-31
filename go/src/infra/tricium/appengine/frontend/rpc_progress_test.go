// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"testing"

	ds "github.com/luci/gae/service/datastore"

	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/authtest"
	"github.com/luci/luci-go/server/auth/identity"
	. "github.com/smartystreets/goconvey/convey"

	"infra/tricium/api/v1"
	trit "infra/tricium/appengine/common/testing"
	"infra/tricium/appengine/common/track"
)

func TestProgress(t *testing.T) {
	Convey("Test Environment", t, func() {

		tt := &trit.Testing{}
		ctx := tt.Context()

		// Add completed run entry.
		run := &track.Run{}
		So(ds.Put(ctx, run), ShouldBeNil)
		runKey := ds.KeyForObj(ctx, run)
		So(ds.Put(ctx, &track.RunResult{
			ID:     "1",
			Parent: runKey,
			State:  tricium.State_SUCCESS,
		}), ShouldBeNil)
		analyzerName := "Hello"
		platform := tricium.Platform_UBUNTU
		analyzerKey := ds.NewKey(ctx, "AnalyzerRun", analyzerName, 0, runKey)
		So(ds.Put(ctx, &track.AnalyzerRun{
			ID:     analyzerName,
			Parent: runKey,
		}), ShouldBeNil)
		So(ds.Put(ctx, &track.AnalyzerResult{
			ID:     "1",
			Parent: analyzerKey,
			State:  tricium.State_SUCCESS,
		}), ShouldBeNil)
		workerName := analyzerName + "_UBUNTU"
		workerKey := ds.NewKey(ctx, "WorkerRun", workerName, 0, analyzerKey)
		worker := &track.WorkerRun{
			ID:       workerName,
			Parent:   analyzerKey,
			Platform: platform,
		}
		So(ds.Put(ctx, worker), ShouldBeNil)
		workerKey = ds.KeyForObj(ctx, worker)
		So(ds.Put(ctx, &track.WorkerResult{
			ID:          "1",
			Parent:      workerKey,
			State:       tricium.State_SUCCESS,
			NumComments: 1,
		}), ShouldBeNil)

		Convey("Progress request", func() {
			ctx = auth.WithState(ctx, &authtest.FakeState{
				Identity: identity.Identity(okACLUser),
			})

			state, progress, err := progress(ctx, run.ID)
			So(err, ShouldBeNil)
			So(state, ShouldEqual, tricium.State_SUCCESS)
			So(len(progress), ShouldEqual, 1)
			So(progress[0].Analyzer, ShouldEqual, analyzerName)
			So(progress[0].Platform, ShouldEqual, platform)
			So(progress[0].NumComments, ShouldEqual, 1)
			So(progress[0].State, ShouldEqual, tricium.State_SUCCESS)
		})
	})
}
