// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package failure_detection analyses a failed build and determines if it needs
// to trigger a new analysis for it
package compilefailuredetection

import (
	"context"
	"fmt"

	"infra/appengine/gofindit/internal/buildbucket"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// AnalyzeBuild analyzes a build and trigger an analysis if necessary.
// Returns true if a new analysis is triggered, returns false otherwise.
func AnalyzeBuild(c context.Context, bbid int64) (bool, error) {
	build, err := buildbucket.GetBuild(c, bbid, &buildbucketpb.BuildMask{
		Fields: &fieldmaskpb.FieldMask{
			Paths: []string{"id", "builder", "input", "status", "steps"},
		},
	})
	if err != nil {
		return false, err
	}

	// We only care about builds with compile failure
	if !hasCompileStepStatus(c, build, buildbucketpb.Status_FAILURE) {
		return false, nil
	}

	lastPassedBuild, firstFailedBuild, err := getLastPassedFirstFailedBuilds(c, build)

	// Could not find last passed build, skip the analysis
	if err != nil {
		logging.Infof(c, "Could not find last passed/first failed builds for failure of build %d. Exiting...", bbid)
		return false, nil
	}

	return triggerAnalysisIfNeeded(c, lastPassedBuild, firstFailedBuild)
}

// Search builds older than refBuild to find the last passed and first failed builds
func getLastPassedFirstFailedBuilds(c context.Context, refBuild *buildbucketpb.Build) (*buildbucketpb.Build, *buildbucketpb.Build, error) {
	// Query buildbucket for the first build with compile failure
	// We only consider maximum of 100 builds before the failed build.
	// If we cannot find the regression range within 100 builds, the failure is
	// too old for the analysis to be useful.
	olderBuilds, err := buildbucket.SearchOlderBuilds(c, refBuild, &buildbucketpb.BuildMask{
		Fields: &fieldmaskpb.FieldMask{
			Paths: []string{"id", "builder", "input", "status", "steps"},
		},
	}, 100)

	if err != nil {
		logging.Errorf(c, "Could not search for older builds: %s", err)
		return nil, nil, err
	}

	var lastPassedBuild *buildbucketpb.Build = nil
	firstFailedBuild := refBuild
	for _, oldBuild := range olderBuilds {
		// We found the last passed build, break
		if oldBuild.Status == buildbucketpb.Status_SUCCESS || hasCompileStepStatus(c, oldBuild, buildbucketpb.Status_SUCCESS) {
			lastPassedBuild = oldBuild
			break
		}
		if hasCompileStepStatus(c, oldBuild, buildbucketpb.Status_FAILURE) {
			firstFailedBuild = oldBuild
		}
	}
	if lastPassedBuild == nil {
		return nil, nil, fmt.Errorf("could not find last passed build")
	}
	return lastPassedBuild, firstFailedBuild, nil
}

// triggerAnalysisIfNeeded checks if there has been an analysis with the regression range.
// if not, it will trigger an analysis.
func triggerAnalysisIfNeeded(c context.Context, lastPassedBuild *buildbucketpb.Build, firstFailedBuild *buildbucketpb.Build) (bool, error) {
	logging.Infof(c, "triggerAnalysisIfNeeded for range (%d, %d)", lastPassedBuild.Id, firstFailedBuild.Id)
	// TODO (nqmtuan): Implement this
	// Search in datastore if there is already an analysis with the same regression range
	// If not, trigger an analysis
	// For now, it just triggers a heuristic analysis, which runs very fast
	// In the future, once we have nth-section analysis, we should put in a task
	// queue and return immediately
	return false, nil
}

// hasCompileStepStatus checks if the compile step for a build has the specified status.
func hasCompileStepStatus(c context.Context, build *buildbucketpb.Build, status buildbucketpb.Status) bool {
	for _, step := range build.Steps {
		if step.Name == "compile" && step.Status == status {
			return true
		}
	}
	return false
}
