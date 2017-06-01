// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package frontend

import (
	"encoding/json"
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

func TestResults(t *testing.T) {
	Convey("Test Environment", t, func() {
		tt := &trit.Testing{}
		ctx := tt.Context()

		// Add run->analyzer->worker->comments
		run := &track.Run{}
		So(ds.Put(ctx, run), ShouldBeNil)
		runKey := ds.KeyForObj(ctx, run)
		So(ds.Put(ctx, &track.RunResult{
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
		So(ds.Put(ctx, &track.WorkerRun{
			ID:       workerName,
			Parent:   analyzerKey,
			Platform: platform,
		}), ShouldBeNil)
		So(ds.Put(ctx, &track.WorkerResult{
			ID:          "1",
			Parent:      workerKey,
			State:       tricium.State_SUCCESS,
			NumComments: 1,
		}), ShouldBeNil)
		json, err := json.Marshal(tricium.Data_Comment{
			Category: analyzerName,
			Message:  "Hello",
		})
		So(err, ShouldBeNil)
		comment := &track.Comment{
			Parent:    workerKey,
			Category:  analyzerName,
			Comment:   string(json),
			Platforms: 0,
		}
		So(ds.Put(ctx, comment), ShouldBeNil)
		commentKey := ds.KeyForObj(ctx, comment)
		So(ds.Put(ctx, &track.CommentResult{
			ID:       "1",
			Parent:   commentKey,
			Included: true,
		}), ShouldBeNil)
		comment = &track.Comment{
			Parent:    workerKey,
			Category:  analyzerName,
			Comment:   string(json),
			Platforms: 0,
		}
		So(ds.Put(ctx, comment), ShouldBeNil)
		commentKey = ds.KeyForObj(ctx, comment)
		So(ds.Put(ctx, &track.CommentResult{
			ID:       "1",
			Parent:   commentKey,
			Included: false,
		}), ShouldBeNil)

		Convey("Merged results request", func() {
			ctx = auth.WithState(ctx, &authtest.FakeState{
				Identity: identity.Identity(okACLUser),
			})

			results, isMerged, err := results(ctx, run.ID)
			So(err, ShouldBeNil)
			So(len(results.Comments), ShouldEqual, 1)
			So(isMerged, ShouldBeTrue)
			comment := results.Comments[0]
			So(comment.Category, ShouldEqual, analyzerName)
		})
	})
}
