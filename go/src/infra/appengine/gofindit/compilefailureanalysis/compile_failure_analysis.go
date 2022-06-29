// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package compilefailureanalysis is the component for analyzing
// compile failures.
// It has 2 main components: heuristic analysis and nth_section analysis
package compilefailureanalysis

import (
	"context"
	"fmt"
	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	"infra/appengine/gofindit/compilefailureanalysis/nthsection"
	"infra/appengine/gofindit/culpritverification"
	"infra/appengine/gofindit/internal/buildbucket"
	gfim "infra/appengine/gofindit/model"
	gfipb "infra/appengine/gofindit/proto"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
)

// AnalyzeFailure receives failure information and perform analysis.
// Note that this assumes that the failure is new (i.e. the client of this
// function should make sure this is not a duplicate analysis)
func AnalyzeFailure(
	c context.Context,
	cf *gfim.CompileFailure,
	first_failed_build_id int64,
	last_passed_build_id int64,
) (*gfim.CompileFailureAnalysis, error) {
	regression_range, e := findRegressionRange(c, first_failed_build_id, last_passed_build_id)
	if e != nil {
		return nil, e
	}

	logging.Infof(c, "Regression range: %v", regression_range)
	// Creates a new CompileFailureAnalysis entity in datastore
	analysis := &gfim.CompileFailureAnalysis{
		CompileFailure:         datastore.KeyForObj(c, cf),
		CreateTime:             clock.Now(c),
		Status:                 gfipb.AnalysisStatus_CREATED,
		FirstFailedBuildId:     first_failed_build_id,
		LastPassedBuildId:      last_passed_build_id,
		InitialRegressionRange: regression_range,
	}
	e = datastore.Put(c, analysis)
	if e != nil {
		return nil, e
	}

	// TODO (nqmtuan): run heuristic analysis and nth-section analysis in parallel
	// Nth-section analysis
	_, e = nthsection.Analyze(c, analysis, regression_range)
	if e != nil {
		logging.Errorf(c, "Error during nthsection analysis: %v", e)
	}

	// Heuristic analysis
	heuristicResult, e := heuristic.Analyze(c, analysis, regression_range)
	if e != nil {
		logging.Errorf(c, "Error during heuristic analysis: %v", e)
		// As we only run heuristic analysis now, returns the error if heuristic
		// analysis failed.
		return nil, e
	}

	// Verifies heuristic analysis result
	verifyHeuristicResults(c, heuristicResult, first_failed_build_id)

	// TODO: For now, just check heuristic analysis status
	// We need to implement nth-section analysis as well
	analysis.Status = heuristicResult.Status
	analysis.EndTime = heuristicResult.EndTime

	e = datastore.Put(c, analysis)
	if e != nil {
		return nil, fmt.Errorf("Failed saving analysis: %w", e)
	}
	return analysis, nil
}

func verifyHeuristicResults(c context.Context, heuristicAnalysis *gfim.CompileHeuristicAnalysis, failedBuildId int64) error {
	suspects, err := getHeuristicSuspectsToVerify(c, heuristicAnalysis)
	if err != nil {
		return err
	}
	for _, suspect := range suspects {
		culpritverification.VerifyCulprit(c, &suspect.GitilesCommit, failedBuildId)
	}
	return nil
}

// In case heuristic analysis returns too many results, we don't want to verify all of them.
// Instead, we want to be selective in what we want to verify.
// For now, we will just take top 3 results of heuristic analysis.
func getHeuristicSuspectsToVerify(c context.Context, heuristicAnalysis *gfim.CompileHeuristicAnalysis) ([]*gfim.Suspect, error) {
	// Getting the suspects for heuristic analysis
	suspects := []*gfim.Suspect{}
	q := datastore.NewQuery("Suspect").Ancestor(datastore.KeyForObj(c, heuristicAnalysis)).Order("-score")
	err := datastore.GetAll(c, q, &suspects)
	if err != nil {
		return nil, err
	}

	// Get top 3 suspects to verify
	nSuspects := 3
	if nSuspects > len(suspects) {
		nSuspects = len(suspects)
	}
	return suspects[:nSuspects], nil
}

// findRegressionRange takes in the first failed and last passed buildID
// and returns the regression range based on GitilesCommit.
func findRegressionRange(
	c context.Context,
	firstFailedBuildID int64,
	lastPassedBuildID int64,
) (*gfipb.RegressionRange, error) {
	firstFailedBuild, err := buildbucket.GetBuild(c, firstFailedBuildID, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting build %d: %w", firstFailedBuildID, err)
	}

	lastPassedBuild, err := buildbucket.GetBuild(c, lastPassedBuildID, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting build %d: %w", lastPassedBuildID, err)
	}

	if firstFailedBuild.GetInput().GetGitilesCommit() == nil || lastPassedBuild.GetInput().GetGitilesCommit() == nil {
		return nil, fmt.Errorf("couldn't get gitiles commit for builds (%d, %d)", lastPassedBuildID, firstFailedBuildID)
	}

	return &gfipb.RegressionRange{
		FirstFailed: firstFailedBuild.GetInput().GetGitilesCommit(),
		LastPassed:  lastPassedBuild.GetInput().GetGitilesCommit(),
	}, nil
}
