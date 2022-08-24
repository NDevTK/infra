// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	gfim "infra/appengine/gofindit/model"
	gfipb "infra/appengine/gofindit/proto"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func TestUpdateAnalysisProgress(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())
	cl := testclock.New(testclock.TestTimeUTC)
	c = clock.Set(c, cl)

	// For some reasons, AutoIndex does not work in this case
	// and it requires an explicit index
	// datastore.GetTestable(c).AutoIndex(true)
	datastore.GetTestable(c).AddIndexes(&datastore.IndexDefinition{
		Kind: "SingleRerun",
		SortBy: []datastore.IndexColumn{
			{
				Property: "rerun_build",
			},
			{
				Property: "start_time",
			},
		},
	})

	Convey("UpdateAnalysisProgress", t, func() {
		// Setup the models
		// Set up suspects
		suspect := &gfim.Suspect{
			Score: 10,
			GitilesCommit: bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3425",
			},
		}
		So(datastore.Put(c, suspect), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		// Set up reruns
		rerunBuildModel := &gfim.CompileRerunBuild{
			Id:      8800,
			Type:    gfim.RerunBuildType_CulpritVerification,
			Suspect: datastore.KeyForObj(c, suspect),
		}
		So(datastore.Put(c, rerunBuildModel), ShouldBeNil)

		parentRerunBuildModel := &gfim.CompileRerunBuild{
			Id:      8801,
			Type:    gfim.RerunBuildType_CulpritVerification,
			Suspect: datastore.KeyForObj(c, suspect),
		}
		So(datastore.Put(c, parentRerunBuildModel), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		suspect.SuspectRerunBuild = datastore.KeyForObj(c, rerunBuildModel)
		suspect.ParentRerunBuild = datastore.KeyForObj(c, parentRerunBuildModel)
		So(datastore.Put(c, suspect), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		// Setup single rerun
		singleRerun1 := &gfim.SingleRerun{
			RerunBuild: datastore.KeyForObj(c, rerunBuildModel),
			GitilesCommit: bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3425",
			},
			Status: gfipb.RerunStatus_IN_PROGRESS,
		}

		singleRerun2 := &gfim.SingleRerun{
			RerunBuild: datastore.KeyForObj(c, parentRerunBuildModel),
			GitilesCommit: bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3426",
			},
			Status: gfipb.RerunStatus_IN_PROGRESS,
		}
		So(datastore.Put(c, singleRerun1), ShouldBeNil)
		So(datastore.Put(c, singleRerun2), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		// Update analysis
		req1 := &gfipb.UpdateAnalysisProgressRequest{
			AnalysisId: 1234,
			Bbid:       8800,
			GitilesCommit: &bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3425",
			},
			RerunResult: &gfipb.RerunResult{
				RerunStatus: gfipb.RerunStatus_FAILED,
			},
		}

		req2 := &gfipb.UpdateAnalysisProgressRequest{
			AnalysisId: 1234,
			Bbid:       8801,
			GitilesCommit: &bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3426",
			},
			RerunResult: &gfipb.RerunResult{
				RerunStatus: gfipb.RerunStatus_PASSED,
			},
		}

		server := &GoFinditBotServer{}
		_, err := server.UpdateAnalysisProgress(c, req1)
		So(err, ShouldBeNil)
		datastore.Get(c, singleRerun1)
		So(singleRerun1.Status, ShouldEqual, gfipb.RerunStatus_FAILED)
		datastore.Get(c, suspect)
		So(suspect.VerificationStatus, ShouldEqual, gfim.SuspectVerificationStatus_UnderVerification)

		_, err = server.UpdateAnalysisProgress(c, req2)
		So(err, ShouldBeNil)
		datastore.Get(c, singleRerun2)
		So(singleRerun2.Status, ShouldEqual, gfipb.RerunStatus_PASSED)
		datastore.Get(c, suspect)
		So(suspect.VerificationStatus, ShouldEqual, gfim.SuspectVerificationStatus_ConfirmedCulprit)
	})

	Convey("verifyUpdateAnalysisProgressRequest", t, func() {
		req := &gfipb.UpdateAnalysisProgressRequest{}
		So(verifyUpdateAnalysisProgressRequest(c, req), ShouldNotBeNil)
		req.AnalysisId = 123
		So(verifyUpdateAnalysisProgressRequest(c, req), ShouldNotBeNil)
		req.Bbid = 8888
		So(verifyUpdateAnalysisProgressRequest(c, req), ShouldNotBeNil)
		req.GitilesCommit = &bbpb.GitilesCommit{}
		So(verifyUpdateAnalysisProgressRequest(c, req), ShouldNotBeNil)
		req.RerunResult = &gfipb.RerunResult{
			RerunStatus: gfipb.RerunStatus_FAILED,
		}
		So(verifyUpdateAnalysisProgressRequest(c, req), ShouldBeNil)
	})
}
