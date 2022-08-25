// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package compilefailureanalysis

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/proto"

	"infra/appengine/gofindit/internal/buildbucket"
	"infra/appengine/gofindit/internal/gitiles"
	"infra/appengine/gofindit/internal/logdog"
	"infra/appengine/gofindit/model"
	gofindit "infra/appengine/gofindit/proto"
)

func TestAnalyzeFailure(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())
	cl := testclock.New(testclock.TestTimeUTC)
	c = clock.Set(c, cl)

	// Setup mock for buildbucket
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mc := buildbucket.NewMockedClient(c, ctl)
	c = mc.Ctx
	c = gitiles.MockedGitilesClientContext(c, map[string]string{})
	res := &bbpb.Build{
		Input: &bbpb.Build_Input{
			GitilesCommit: &bbpb.GitilesCommit{
				Host:    "host",
				Project: "proj",
				Id:      "id",
				Ref:     "ref",
			},
		},
		Steps: []*bbpb.Step{
			{
				Name: "compile",
				Logs: []*bbpb.Log{
					{
						Name:    "json.output[ninja_info]",
						ViewUrl: "https://logs.chromium.org/logs/ninja_log",
					},
					{
						Name:    "stdout",
						ViewUrl: "https://logs.chromium.org/logs/stdout_log",
					},
				},
			},
		},
	}
	mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()

	// Mock logdog
	ninjaLogJson := map[string]interface{}{
		"failures": []map[string]interface{}{
			{
				"output_nodes": []string{
					"obj/net/net_unittests__library/ssl_server_socket_unittest.o",
				},
			},
		},
	}
	ninjaLogStr, _ := json.Marshal(ninjaLogJson)
	c = logdog.MockClientContext(c, map[string]string{
		"https://logs.chromium.org/logs/ninja_log":  string(ninjaLogStr),
		"https://logs.chromium.org/logs/stdout_log": "stdout_log",
	})

	Convey("AnalyzeFailure analysis is created", t, func() {
		failed_build := &model.LuciFailedBuild{
			Id: 88128398584903,
			LuciBuild: model.LuciBuild{
				BuildId:     88128398584903,
				Project:     "chromium",
				Bucket:      "ci",
				Builder:     "android",
				BuildNumber: 123,
				StartTime:   cl.Now(),
				EndTime:     cl.Now(),
				CreateTime:  cl.Now(),
			},
			BuildFailureType: gofindit.BuildFailureType_COMPILE,
		}
		So(datastore.Put(c, failed_build), ShouldBeNil)

		compile_failure := &model.CompileFailure{
			Build: datastore.KeyForObj(c, failed_build),
		}
		So(datastore.Put(c, compile_failure), ShouldBeNil)

		compile_failure_analysis, err := AnalyzeFailure(c, compile_failure, 123, 456)
		So(err, ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		err = datastore.Get(c, compile_failure)
		So(err, ShouldBeNil)
		So(compile_failure.OutputTargets, ShouldResemble, []string{"obj/net/net_unittests__library/ssl_server_socket_unittest.o"})

		// Make sure that the analysis is created
		q := datastore.NewQuery("CompileFailureAnalysis").Eq("compile_failure", datastore.KeyForObj(c, compile_failure))
		analyses := []*model.CompileFailureAnalysis{}
		datastore.GetAll(c, q, &analyses)
		So(len(analyses), ShouldEqual, 1)

		// Make sure the heuristic analysis and nthsection analysis are run
		q = datastore.NewQuery("CompileHeuristicAnalysis").Ancestor(datastore.KeyForObj(c, compile_failure_analysis))
		heuristic_analyses := []*model.CompileHeuristicAnalysis{}
		datastore.GetAll(c, q, &heuristic_analyses)
		So(len(heuristic_analyses), ShouldEqual, 1)

		q = datastore.NewQuery("CompileNthSectionAnalysis").Ancestor(datastore.KeyForObj(c, compile_failure_analysis))
		nthsection_analyses := []*model.CompileNthSectionAnalysis{}
		datastore.GetAll(c, q, &nthsection_analyses)
		So(len(nthsection_analyses), ShouldEqual, 1)
	})
}

func TestFindRegressionRange(t *testing.T) {
	t.Parallel()
	// Setup mock for buildbucket
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	c := context.Background()

	Convey("No Gitiles Commit", t, func() {
		mc := buildbucket.NewMockedClient(c, ctl)
		c = mc.Ctx
		res := &bbpb.Build{}
		mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()
		_, e := findRegressionRange(c, 8001, 8000)
		So(e, ShouldNotBeNil)
	})

	Convey("Have Gitiles Commit", t, func() {
		mc := buildbucket.NewMockedClient(c, ctl)
		c = mc.Ctx
		res1 := &bbpb.Build{
			Input: &bbpb.Build_Input{
				GitilesCommit: &bbpb.GitilesCommit{
					Host:    "host1",
					Project: "proj1",
					Id:      "id1",
					Ref:     "ref1",
				},
			},
		}

		res2 := &bbpb.Build{
			Input: &bbpb.Build_Input{
				GitilesCommit: &bbpb.GitilesCommit{
					Host:    "host2",
					Project: "proj2",
					Id:      "id2",
					Ref:     "ref2",
				},
			},
		}

		// It is hard to match the exact GetBuildRequest. We use Times() to simulate different response.
		mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res1, nil).Times(1)
		mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res2, nil).Times(1)

		rr, e := findRegressionRange(c, 8001, 8000)
		So(e, ShouldBeNil)

		diff := cmp.Diff(rr.FirstFailed, &bbpb.GitilesCommit{
			Host:    "host1",
			Project: "proj1",
			Id:      "id1",
			Ref:     "ref1",
		}, cmp.Comparer(proto.Equal))
		So(diff, ShouldEqual, "")

		diff = cmp.Diff(rr.LastPassed, &bbpb.GitilesCommit{
			Host:    "host2",
			Project: "proj2",
			Id:      "id2",
			Ref:     "ref2",
		}, cmp.Comparer(proto.Equal))
		So(diff, ShouldEqual, "")
	})
}

func TestVerifyCulprit(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())
	datastore.GetTestable(c).AutoIndex(true)

	Convey("getHeuristicSuspectsToVerify", t, func() {
		heuristicAnalysis := &model.CompileHeuristicAnalysis{
			Status: gofindit.AnalysisStatus_FOUND,
		}

		So(datastore.Put(c, heuristicAnalysis), ShouldBeNil)

		suspect1 := &model.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          1,
		}
		suspect2 := &model.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          3,
		}
		suspect3 := &model.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          4,
		}
		suspect4 := &model.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          2,
		}
		So(datastore.Put(c, suspect1), ShouldBeNil)
		So(datastore.Put(c, suspect2), ShouldBeNil)
		So(datastore.Put(c, suspect3), ShouldBeNil)
		So(datastore.Put(c, suspect4), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		suspects, err := getHeuristicSuspectsToVerify(c, heuristicAnalysis)
		So(err, ShouldBeNil)
		So(len(suspects), ShouldEqual, 3)
		So(suspects[0].Score, ShouldEqual, 4)
		So(suspects[1].Score, ShouldEqual, 3)
		So(suspects[2].Score, ShouldEqual, 2)
	})
}
