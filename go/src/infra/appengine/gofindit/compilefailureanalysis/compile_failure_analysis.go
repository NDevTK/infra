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
	firstFailedBuildID int64,
	lastPassedBuildID int64,
) (*gfim.CompileFailureAnalysis, error) {
	logging.Infof(c, "AnalyzeFailure firstFailed = %d", firstFailedBuildID)
	regression_range, e := findRegressionRange(c, firstFailedBuildID, lastPassedBuildID)
	if e != nil {
		return nil, e
	}

	logging.Infof(c, "Regression range: %v", regression_range)
	// Creates a new CompileFailureAnalysis entity in datastore
	analysis := &gfim.CompileFailureAnalysis{
		CompileFailure:         datastore.KeyForObj(c, cf),
		CreateTime:             clock.Now(c),
		Status:                 gfipb.AnalysisStatus_CREATED,
		FirstFailedBuildId:     firstFailedBuildID,
		LastPassedBuildId:      lastPassedBuildID,
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
		logging.Errorf(c, "Error during heuristic analysis for build %d: %v", e)
		// As we only run heuristic analysis now, returns the error if heuristic
		// analysis failed.
		return nil, e
	}

	// TODO: For now, just check heuristic analysis status
	// We need to implement nth-section analysis as well
	analysis.Status = heuristicResult.Status
	analysis.EndTime = heuristicResult.EndTime

	e = datastore.Put(c, analysis)
	if e != nil {
		return nil, fmt.Errorf("Failed saving analysis: %w", e)
	}

	// Verifies heuristic analysis result.
	// TODO (nqmtuan): Enable verifyHeuristicResults when we fully implemented
	// the culprit verification. Enabling it now will create a lot of noises.
	// verifyHeuristicResults(c, heuristicResult, firstFailedBuildID)

	return analysis, nil
}

// verifyHeuristicResults verifies if the suspects of heuristic analysis are the real culprit.
// analysisID is CompileFailureAnalysis ID. It is meant to be propagated all the way to the
// recipe, so we can identify the analysis in buildbucket.
func verifyHeuristicResults(c context.Context, heuristicAnalysis *gfim.CompileHeuristicAnalysis, failedBuildID int64, analysisID int64) error {
	// TODO (nqmtuan): Move the verification into a task queue
	suspects, err := getHeuristicSuspectsToVerify(c, heuristicAnalysis)
	if err != nil {
		return err
	}
	for _, suspect := range suspects {
		err := culpritverification.VerifySuspect(c, suspect, failedBuildID, analysisID)
		if err != nil {
			// Just log the error and continue for other suspects
			logging.Errorf(c, "Error in verifying suspect %d for analysis %d", suspect.Id, analysisID)
		}
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
