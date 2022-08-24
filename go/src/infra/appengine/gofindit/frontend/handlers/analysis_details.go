// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	gfim "infra/appengine/gofindit/model"
	gfipb "infra/appengine/gofindit/proto"
	gfis "infra/appengine/gofindit/server"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

type AssociatedBug struct {
	// System the bug is from, e.g. "monorail" or "buganizer"
	BugSystem string `json:"bugSystem"`
	// Project of the bug, e.g. "chromium"
	Project string `json:"project"`
	// Bug ID, e.g. "1234567"
	ID string `json:"id"`
	// Text to display for the bug link, e.g. "crbug.com/1234567"
	LinkText string `json:"linkText"`
	// URL for the bug, e.g. "https://bugs.chromium.org/p/chromium/issues/detail?id=1234567"
	URL string `json:"url"`
}

type SuspectRange struct {
	// Text to display for the suspect range link, e.g. "10000c0 .. 1000c6"
	LinkText string `json:"linkText"`
	// URL for the suspect range, e.g. "https://chromium.googlesource.com/chromium/src/+log/10000c0..10000c6"
	URL string `json:"url"`
}

type CL struct {
	// Commit ID of the change list (CL), e.g. "1234567"
	CommitID string `json:"commitID"`
	// Title of the CL, e.g. "[GoFindit] UI for analysis details"
	Title string `json:"title"`
	// URL to the code review of the CL, e.g. "https://chromium-review.googlesource.com/c/infra/infra/+/1234567"
	ReviewURL string `json:"reviewURL"`
}

// A CL to revert a merged CL
type RevertCL struct {
	CL `json:"cl"`
	// Status of the CL, e.g. "Merged"
	Status string `json:"status"`
	// Submit time of the CL, e.g. "2022-02-02 10:11:47 AM UTC+10:00"
	SubmitTime string `json:"submitTime"`
	// Commit position, e.g. 50571
	CommitPosition int `json:"commitPosition"`
}

// PrimeSuspect is a suspect CL that was identified as the most likely culprit
type PrimeSuspect struct {
	// The suspect CL
	CL `json:"cl"`
	// Culprit verification status, e.g. "VERIFYING", "VERIFIED AS CULPRIT", "UNSUCCESSFUL VERIFICATION"
	CulpritStatus string `json:"culpritStatus"`
	// Source analysis which identified the CL, e.g. "Heuristic", "Nth Section"
	AccuseSource string `json:"accuseSource"`
}

type HeuristicSuspectDetails struct {
	// The suspect CL
	CL `json:"cl"`
	// The heuristic analysis score for the CL (higher score means more likely to be the culprit)
	Score int `json:"score"`
	// Confidence that the suspect CL is the culprit, e.g. "LOW", "MEDIUM", "HIGH"
	Confidence string `json:"confidence"`
	// Justification for the suspect confidence; can be one or more reasons
	Justification []string `json:"justification"`
}

type HeuristicAnalysisDetails struct {
	// Whether the heuristic analysis has finished
	IsComplete bool `json:"isComplete"`
	// Suspects identified from the heuristic analysis
	Suspects []*HeuristicSuspectDetails `json:"suspects"`
}

type AnalysisDetails struct {
	// ID of the failure analysis, e.g. "100000"
	AnalysisID int64 `json:"analysisID,string"`
	// Status of the failure analysis, e.g. "FOUND"
	Status string `json:"status"`
	// Buildbucket ID for the first failed build associated with the failure analysis, e.g. "80000000001"
	BuildID int64 `json:"buildID,string"`
	// Type of failure, e.g. "Compile"
	FailureType gfim.BuildFailureType `json:"failureType"`
	// Name of the builder for the first failed build
	Builder string `json:"builder"`
	// Range of suspected CLs
	SuspectRange *SuspectRange `json:"suspectRange"`
	// Bugs related to this failure
	RelatedBugs []*AssociatedBug `json:"bugs"`
	// Revert CL for the culprit CL, if a culprit has been found
	RevertCL *RevertCL `json:"revertCL"`
	// Suspect CLs that are most likely to be the culprit
	PrimeSuspects []*PrimeSuspect `json:"primeSuspects"`
	// Heuristic analysis results
	HeuristicResults *HeuristicAnalysisDetails `json:"heuristicResults"`
	// TODO: add nth section results
	// TODO: add culprit verification results
}

// Creates a SuspectRange based on the repo and start & end revisions
func newSuspectRange(
	repoURL string,
	startRevision string,
	endRevision string) *SuspectRange {
	return &SuspectRange{
		LinkText: fmt.Sprintf("%s ... %s", startRevision, endRevision),
		URL:      fmt.Sprintf("%s/+log/%s..%s", repoURL, startRevision, endRevision),
	}
}

// Responds to HTTP Get requests for the details of the failure analysis
// associated with a given build
func GetAnalysisDetails(ctx *router.Context) {
	bbid := ctx.Params.ByName("bbid")

	buildID, err := strconv.ParseInt(bbid, 10, 64)
	if err != nil {
		logging.Errorf(ctx.Context, "Failed to convert Buildbucket ID '%s' to integer: %s", bbid, err)
		http.Error(ctx.Writer, "Buildbucket ID could not be converted to a number", http.StatusBadRequest)
		return
	}

	// TODO: remove the hardcoded response cases and eliminate this switch
	var analysisDetails *AnalysisDetails
	switch buildID {
	case 0:
		analysisDetails = getFauxEmptyAnalysisDetails(buildID)
	case 1:
		analysisDetails = getFauxAnalysisDetails(buildID)
	default:
		analysisDetails, err = lookUpAnalysisDetails(ctx, buildID)
		if err != nil {
			http.Error(ctx.Writer, fmt.Sprintf("Encountered error: %s", err), http.StatusInternalServerError)
			return
		}
		if analysisDetails == nil {
			http.Error(ctx.Writer, fmt.Sprintf("No analysis found related to build with ID %d", buildID), http.StatusNotFound)
			return
		}
	}

	respondWithJSON(ctx, &analysisDetails)
}

// Returns a faux failure analysis that doesn't have much data, which is useful
// for checking how the UI handles lack of data
func getFauxEmptyAnalysisDetails(bbid int64) *AnalysisDetails {
	suspectRange := newSuspectRange("https://chromium.googlesource.com/placeholder", "cd52ae", "cd52af")

	return &AnalysisDetails{
		AnalysisID:    10000000,
		Status:        gfipb.AnalysisStatus_CREATED.String(),
		BuildID:       bbid,
		FailureType:   gfim.BuildFailureType_Compile,
		Builder:       "builder-type-amd-rhel-cc64",
		SuspectRange:  suspectRange,
		RelatedBugs:   []*AssociatedBug{},
		PrimeSuspects: []*PrimeSuspect{},
		HeuristicResults: &HeuristicAnalysisDetails{
			IsComplete: false,
			Suspects:   []*HeuristicSuspectDetails{},
		},
	}
}

// Returns a faux failure analysis that has most fields, which is useful
// for checking how the UI renders all details
func getFauxAnalysisDetails(bbid int64) *AnalysisDetails {
	suspectRange := newSuspectRange("https://chromium.googlesource.com/placeholder", "cd52ae", "cd52af")

	suspects := []*gfim.Suspect{
		{
			ReviewUrl: "https://chromium-review.googlesource.com/placeholder1",
			Score:     15,
			Justification: `The file "dir/a/b/x.cc" was added and it was in the failure log.
The file "dir/a/b/y.cc" was added and it was in the failure log.
The file "dir/a/b/z.cc" was added and it was in the failure log.`,
		},
		{
			ReviewUrl:     "https://chromium-review.googlesource.com/placeholder2",
			Score:         2,
			Justification: `The file "content/util.c" was modified. It was related to the file obj/content/util.o which was in the failure log.`,
		},
	}

	heuristicSuspects := make([]*HeuristicSuspectDetails, len(suspects))
	for i, suspect := range suspects {
		heuristicSuspects[i] = &HeuristicSuspectDetails{
			CL: CL{
				CommitID:  "c9e3a" + strconv.Itoa(i),
				Title:     "",
				ReviewURL: suspect.ReviewUrl,
			},
			Score:         suspect.Score,
			Confidence:    heuristic.GetConfidenceLevel(suspect.Score).String(),
			Justification: strings.Split(suspect.Justification, "\n"),
		}
	}

	return &AnalysisDetails{
		AnalysisID:   10000001,
		Status:       gfipb.AnalysisStatus_CREATED.String(),
		BuildID:      bbid,
		FailureType:  gfim.BuildFailureType_Compile,
		Builder:      "builder-type-amd-rhel-cc64",
		SuspectRange: suspectRange,
		RelatedBugs: []*AssociatedBug{
			{
				BugSystem: "monorail",
				Project:   "chromium",
				ID:        "23527",
				LinkText:  "crbug.com/23527",
				URL:       "https://bugs.chromium.org/placeholder/chromium/issues/detail?id=23527",
			},
			{
				BugSystem: "monorail",
				Project:   "chromium",
				ID:        "23528",
				LinkText:  "crbug.com/23528",
				URL:       "https://bugs.chromium.org/placeholder/chromium/issues/detail?id=23528",
			},
		},
		RevertCL: &RevertCL{
			CL: CL{
				CommitID:  "f23ade252",
				Title:     "Title of revert CL that was created by GoFindit",
				ReviewURL: "https://not.a.real.link",
			},
			Status:         "Merged",
			SubmitTime:     "2022-02-02 10:21:13 AM UTC+10:00",
			CommitPosition: 77346,
		},
		PrimeSuspects: []*PrimeSuspect{
			{
				CL: CL{
					CommitID:  "b2f50452c",
					Title:     "Short title",
					ReviewURL: "https://www.google.com",
				},
				CulpritStatus: "VERIFYING",
				AccuseSource:  "Heuristic",
			},
		},
		HeuristicResults: &HeuristicAnalysisDetails{
			IsComplete: true,
			Suspects:   heuristicSuspects,
		},
	}
}

// Returns the analysis details using the given build ID as the search key
func lookUpAnalysisDetails(ctx *router.Context, bbid int64) (*AnalysisDetails, error) {
	// Get the failure analysis related to the build
	failureAnalysis, err := gfis.GetAnalysisForBuild(ctx.Context, bbid)
	if (err != nil) || (failureAnalysis == nil) {
		return nil, err
	}

	// Get the failure analysis's first failed build
	firstFailedBuild, err := gfis.GetBuild(ctx.Context, failureAnalysis.FirstFailedBuildId)
	if (err != nil) || (firstFailedBuild == nil) {
		return nil, err
	}

	// TODO: replace this suspect range with the actual suspect range
	suspectRange := &SuspectRange{
		LinkText: "",
		URL:      "",
	}

	heuristicDetails, err := getHeuristicDetails(ctx.Context, failureAnalysis)
	if err != nil {
		return nil, err
	}

	// TODO: get nth section analysis results once it has been implemented

	// TODO: get compile verification results once it has been implemented

	details := &AnalysisDetails{
		AnalysisID:   failureAnalysis.Id,
		Status:       failureAnalysis.Status.String(),
		BuildID:      firstFailedBuild.BuildId,
		FailureType:  firstFailedBuild.FailureType,
		Builder:      firstFailedBuild.Builder,
		SuspectRange: suspectRange,
		// TODO: get related bugs
		RelatedBugs: []*AssociatedBug{},
		// TODO: get info for revert CL if one has been created
		RevertCL: nil,
		// TODO: get the top suspects from the analysis results
		PrimeSuspects:    []*PrimeSuspect{},
		HeuristicResults: heuristicDetails,
		// TODO: add nth section results
		// TODO: add culprit verification results
	}
	return details, nil
}

func getHeuristicDetails(c context.Context, analysis *gfim.CompileFailureAnalysis) (*HeuristicAnalysisDetails, error) {
	heuristicAnalysis, err := gfis.GetHeuristicAnalysis(c, analysis)
	if err != nil {
		return nil, err
	}

	if heuristicAnalysis == nil {
		// No heuristic analysis for this compile failure analysis
		return &HeuristicAnalysisDetails{
			IsComplete: false,
			Suspects:   []*HeuristicSuspectDetails{},
		}, nil
	}

	suspects, err := gfis.GetSuspects(c, heuristicAnalysis)
	if err != nil {
		return nil, err
	}

	suspectDetails := make([]*HeuristicSuspectDetails, len(suspects))
	for i, suspect := range suspects {
		suspectDetails[i] = &HeuristicSuspectDetails{
			CL: CL{
				CommitID: suspect.GitilesCommit.Id,
				// TODO: get the title of the suspect CL
				Title:     "",
				ReviewURL: suspect.ReviewUrl,
			},
			Score:         suspect.Score,
			Confidence:    heuristic.GetConfidenceLevel(suspect.Score).String(),
			Justification: strings.Split(suspect.Justification, "\n"),
		}
	}

	return &HeuristicAnalysisDetails{
		IsComplete: true,
		Suspects:   suspectDetails,
	}, nil
}
