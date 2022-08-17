// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	gfim "infra/appengine/gofindit/model"
	gfipb "infra/appengine/gofindit/proto"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/router"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAnalysisDetails(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())
	datastore.GetTestable(c).AutoIndex(true)

	Convey("Test router requests correctly", t, func() {
		// Set up a test router to handle analysis details requests
		testRouter := router.New()
		testRouter.GET("/api/analysis/b/:bbid", nil, GetAnalysisDetails)
		get := func(bbid string) *http.Response {
			url := fmt.Sprintf("/api/analysis/b/%s", bbid)
			request, err := http.NewRequestWithContext(c, "GET", url, nil)
			So(err, ShouldBeNil)

			response := httptest.NewRecorder()
			testRouter.ServeHTTP(response, request)
			return response.Result()
		}

		Convey("Invalid Buildbucket ID handled", func() {
			response := get("N0tABuildNumb3r")
			So(response.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Build not found", func() {
			response := get("10002340")
			So(response.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Build exists but no analysis for it", func() {
			// Prepare datastore
			failedBuild := &gfim.LuciFailedBuild{
				Id: 10002341,
				LuciBuild: gfim.LuciBuild{
					BuildId: 10002341,
					Builder: "test-builder-x86",
				},
				FailureType: gfim.BuildFailureType_Compile,
			}
			So(datastore.Put(c, failedBuild), ShouldBeNil)

			datastore.GetTestable(c).CatchupIndexes()

			response := get("10002341")
			So(response.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		// TODO: Update expected values below once fetching of more attributes has
		// been implemented (e.g. for related bugs, revert CL, etc)

		Convey("Analysis found without heuristic results", func() {
			// Prepare datastore
			failedBuild := &gfim.LuciFailedBuild{
				Id: 10002342,
				LuciBuild: gfim.LuciBuild{
					BuildId: 10002342,
					Builder: "test-builder-x86",
				},
				FailureType: gfim.BuildFailureType_Compile,
			}
			So(datastore.Put(c, failedBuild), ShouldBeNil)

			compileFailure := &gfim.CompileFailure{
				Id:    10002342,
				Build: datastore.KeyForObj(c, failedBuild),
			}
			So(datastore.Put(c, compileFailure), ShouldBeNil)

			compileFailureAnalysis := &gfim.CompileFailureAnalysis{
				Id:                 40000002,
				Status:             gfipb.AnalysisStatus_CREATED,
				CompileFailure:     datastore.KeyForObj(c, compileFailure),
				FirstFailedBuildId: 10002342,
			}
			So(datastore.Put(c, compileFailureAnalysis), ShouldBeNil)

			datastore.GetTestable(c).CatchupIndexes()

			response := get("10002342")
			So(response.StatusCode, ShouldEqual, http.StatusOK)

			b, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)

			var responseBody AnalysisDetails
			So(json.Unmarshal(b, &responseBody), ShouldBeNil)
			expected := AnalysisDetails{
				AnalysisID:  40000002,
				Status:      gfipb.AnalysisStatus_CREATED.String(),
				BuildID:     10002342,
				FailureType: gfim.BuildFailureType_Compile,
				Builder:     "test-builder-x86",
				SuspectRange: &SuspectRange{
					LinkText: "",
					URL:      "",
				},
				RelatedBugs:   []*AssociatedBug{},
				PrimeSuspects: []*PrimeSuspect{},
				HeuristicResults: &HeuristicAnalysisDetails{
					IsComplete: false,
					Suspects:   []*HeuristicSuspectDetails{},
				},
			}
			So(responseBody, ShouldResemble, expected)
		})

		Convey("Analysis found with heuristic results", func() {
			// Prepare datastore
			failedBuild := &gfim.LuciFailedBuild{
				Id: 10002343,
				LuciBuild: gfim.LuciBuild{
					BuildId: 10002343,
					Builder: "android",
				},
				FailureType: gfim.BuildFailureType_Compile,
			}
			So(datastore.Put(c, failedBuild), ShouldBeNil)

			compileFailure := &gfim.CompileFailure{
				Id:    10002343,
				Build: datastore.KeyForObj(c, failedBuild),
			}
			So(datastore.Put(c, compileFailure), ShouldBeNil)

			compileFailureAnalysis := &gfim.CompileFailureAnalysis{
				Id:                 40000003,
				Status:             gfipb.AnalysisStatus_FOUND,
				CompileFailure:     datastore.KeyForObj(c, compileFailure),
				FirstFailedBuildId: 10002343,
			}
			So(datastore.Put(c, compileFailureAnalysis), ShouldBeNil)

			compileHeuristicAnalysis := &gfim.CompileHeuristicAnalysis{
				Id:             7000,
				ParentAnalysis: datastore.KeyForObj(c, compileFailureAnalysis),
				Status:         gfipb.AnalysisStatus_FOUND,
			}
			So(datastore.Put(c, compileHeuristicAnalysis), ShouldBeNil)

			suspect := &gfim.Suspect{
				ParentAnalysis: datastore.KeyForObj(c, compileHeuristicAnalysis),
				GitilesCommit: buildbucketpb.GitilesCommit{
					Project:  "test-project",
					Host:     "test-host",
					Ref:      "ref",
					Id:       "c67dea932d23b",
					Position: 173,
				},
				ReviewUrl: "https://chromium-review.googlesource.com/placeholder/test-review",
				Score:     20,
				Justification: `The file "dir/a/b/x.cc" was added and it was in the failure log.
The file "dir/a/b/y.cc" was added and it was in the failure log.`,
			}
			So(datastore.Put(c, suspect), ShouldBeNil)

			datastore.GetTestable(c).CatchupIndexes()

			response := get("10002343")
			So(response.StatusCode, ShouldEqual, http.StatusOK)

			b, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)

			var responseBody AnalysisDetails
			So(json.Unmarshal(b, &responseBody), ShouldBeNil)
			expected := AnalysisDetails{
				AnalysisID:  40000003,
				Status:      gfipb.AnalysisStatus_FOUND.String(),
				BuildID:     10002343,
				FailureType: gfim.BuildFailureType_Compile,
				Builder:     "android",
				SuspectRange: &SuspectRange{
					LinkText: "",
					URL:      "",
				},
				RelatedBugs:   []*AssociatedBug{},
				PrimeSuspects: []*PrimeSuspect{},
				HeuristicResults: &HeuristicAnalysisDetails{
					IsComplete: true,
					Suspects: []*HeuristicSuspectDetails{
						{
							CL: CL{
								CommitID:  "c67dea932d23b",
								Title:     "",
								ReviewURL: "https://chromium-review.googlesource.com/placeholder/test-review",
							},
							Score:      20,
							Confidence: heuristic.GetConfidenceLevel(20).String(),
							Justification: []string{
								`The file "dir/a/b/x.cc" was added and it was in the failure log.`,
								`The file "dir/a/b/y.cc" was added and it was in the failure log.`,
							},
						},
					},
				},
			}
			So(responseBody, ShouldResemble, expected)
		})
	})
}
