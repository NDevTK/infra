// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package server

import (
	"context"
	"testing"

	gfim "infra/appengine/gofindit/model"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func TestGetBuild(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())

	Convey("No build found", t, func() {
		buildModel, err := GetBuild(c, 100)
		So(err, ShouldBeNil)
		So(buildModel, ShouldBeNil)
	})

	Convey("Build found", t, func() {
		// Prepare datastore
		failed_build := &gfim.LuciFailedBuild{
			Id: 101,
		}
		So(datastore.Put(c, failed_build), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		buildModel, err := GetBuild(c, 101)
		So(err, ShouldBeNil)
		So(buildModel, ShouldNotBeNil)
		So(buildModel.Id, ShouldEqual, 101)
	})
}

func TestGetAnalysisForBuild(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())

	Convey("No build found", t, func() {
		analysis, err := GetAnalysisForBuild(c, 100)
		So(err, ShouldBeNil)
		So(analysis, ShouldBeNil)
	})

	Convey("No analysis found", t, func() {
		// Prepare datastore
		failedBuild := &gfim.LuciFailedBuild{
			Id: 101,
		}
		So(datastore.Put(c, failedBuild), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		analysis, err := GetAnalysisForBuild(c, 101)
		So(err, ShouldBeNil)
		So(analysis, ShouldBeNil)
	})

	Convey("Analysis found", t, func() {
		// Prepare datastore
		failedBuild := &gfim.LuciFailedBuild{
			Id: 101,
		}
		So(datastore.Put(c, failedBuild), ShouldBeNil)

		compileFailure := &gfim.CompileFailure{
			Id:    101,
			Build: datastore.KeyForObj(c, failedBuild),
		}
		So(datastore.Put(c, compileFailure), ShouldBeNil)

		compileFailureAnalysis := &gfim.CompileFailureAnalysis{
			Id:                 1230001,
			CompileFailure:     datastore.KeyForObj(c, compileFailure),
			FirstFailedBuildId: 101,
		}
		So(datastore.Put(c, compileFailureAnalysis), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		analysis, err := GetAnalysisForBuild(c, 101)
		So(err, ShouldBeNil)
		So(analysis, ShouldNotBeNil)
		So(analysis.Id, ShouldEqual, 1230001)
		So(analysis.FirstFailedBuildId, ShouldEqual, 101)
	})

	Convey("Related analysis found", t, func() {
		// Prepare datastore
		firstFailedBuild := &gfim.LuciFailedBuild{
			Id: 200,
		}
		So(datastore.Put(c, firstFailedBuild), ShouldBeNil)

		firstCompileFailure := &gfim.CompileFailure{
			Id:    200,
			Build: datastore.KeyForObj(c, firstFailedBuild),
		}
		So(datastore.Put(c, firstCompileFailure), ShouldBeNil)

		failedBuild := &gfim.LuciFailedBuild{
			Id: 201,
		}
		So(datastore.Put(c, failedBuild), ShouldBeNil)

		compileFailure := &gfim.CompileFailure{
			Id:               201,
			Build:            datastore.KeyForObj(c, failedBuild),
			MergedFailureKey: datastore.KeyForObj(c, firstCompileFailure),
		}
		So(datastore.Put(c, compileFailure), ShouldBeNil)

		compileFailureAnalysis := &gfim.CompileFailureAnalysis{
			Id:             1230002,
			CompileFailure: datastore.KeyForObj(c, firstCompileFailure),
		}
		So(datastore.Put(c, compileFailureAnalysis), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		analysis, err := GetAnalysisForBuild(c, 201)
		So(err, ShouldBeNil)
		So(analysis, ShouldNotBeNil)
		So(analysis.Id, ShouldEqual, 1230002)
	})
}

func TestGetHeuristicAnalysis(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())

	Convey("No heuristic analysis found", t, func() {
		compileFailureAnalysis := &gfim.CompileFailureAnalysis{
			Id: 1230003,
		}
		heuristicAnalysis, err := GetHeuristicAnalysis(c, compileFailureAnalysis)
		So(err, ShouldBeNil)
		So(heuristicAnalysis, ShouldBeNil)
	})

	Convey("Heuristic analysis found", t, func() {
		// Prepare datastore
		compileFailureAnalysis := &gfim.CompileFailureAnalysis{
			Id: 1230003,
		}
		So(datastore.Put(c, compileFailureAnalysis), ShouldBeNil)

		compileHeuristicAnalysis := &gfim.CompileHeuristicAnalysis{
			Id:             4560001,
			ParentAnalysis: datastore.KeyForObj(c, compileFailureAnalysis),
		}
		So(datastore.Put(c, compileHeuristicAnalysis), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		heuristicAnalysis, err := GetHeuristicAnalysis(c, compileFailureAnalysis)
		So(err, ShouldBeNil)
		So(heuristicAnalysis, ShouldNotBeNil)
		So(heuristicAnalysis.Id, ShouldEqual, 4560001)
	})
}

func TestGetSuspects(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())
	datastore.GetTestable(c).AutoIndex(true)

	Convey("No suspects found", t, func() {
		// Prepare datastore
		heuristicAnalysis := &gfim.CompileHeuristicAnalysis{
			Id: 700,
		}
		So(datastore.Put(c, heuristicAnalysis), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		suspects, err := GetSuspects(c, heuristicAnalysis)
		So(err, ShouldBeNil)
		So(len(suspects), ShouldEqual, 0)
	})

	Convey("All suspects found", t, func() {
		// Prepare datastore
		heuristicAnalysis := &gfim.CompileHeuristicAnalysis{
			Id: 701,
		}
		So(datastore.Put(c, heuristicAnalysis), ShouldBeNil)

		suspect1 := &gfim.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          1,
		}
		suspect2 := &gfim.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          3,
		}
		suspect3 := &gfim.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          4,
		}
		suspect4 := &gfim.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, heuristicAnalysis),
			Score:          2,
		}
		So(datastore.Put(c, suspect1), ShouldBeNil)
		So(datastore.Put(c, suspect2), ShouldBeNil)
		So(datastore.Put(c, suspect3), ShouldBeNil)
		So(datastore.Put(c, suspect4), ShouldBeNil)

		// Add a different heuristic analysis with its own suspect
		otherHeuristicAnalysis := &gfim.CompileHeuristicAnalysis{
			Id: 702,
		}
		So(datastore.Put(c, heuristicAnalysis), ShouldBeNil)
		otherSuspect := &gfim.Suspect{
			ParentAnalysis: datastore.KeyForObj(c, otherHeuristicAnalysis),
			Score:          5,
		}
		So(datastore.Put(c, otherSuspect), ShouldBeNil)

		datastore.GetTestable(c).CatchupIndexes()

		suspects, err := GetSuspects(c, heuristicAnalysis)
		So(err, ShouldBeNil)
		So(len(suspects), ShouldEqual, 4)
		So(suspects[0].Score, ShouldEqual, 4)
		So(suspects[1].Score, ShouldEqual, 3)
		So(suspects[2].Score, ShouldEqual, 2)
		So(suspects[3].Score, ShouldEqual, 1)
	})
}
