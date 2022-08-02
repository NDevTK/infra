// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package api contains the utility functions and APIs required to populate the
// GoFindit frontend
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	gfim "infra/appengine/gofindit/model"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

type AssociatedBug struct {
	BugSystem string `json:"bugSystem"`
	Project   string `json:"project"`
	ID        string `json:"id"`
	LinkText  string `json:"linkText"`
	URL       string `json:"url"`
}

type SuspectRange struct {
	LinkText string `json:"linkText"`
	URL      string `json:"url"`
}

type ChangeList struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	Status         string `json:"status"`
	SubmitTime     string `json:"submitTime"`
	CommitPosition int    `json:"commitPosition"`
}

type SuspectSummary struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	CulpritStatus string `json:"culpritStatus"`
	AccuseSource  string `json:"accuseSource"`
}

type FailureAnalysis struct {
	ID                int                `json:"id"`
	Status            string             `json:"status"`
	FailureType       string             `json:"failureType"`
	BuildID           int                `json:"buildID"`
	Builder           string             `json:"builder"`
	SuspectRange      *SuspectRange      `json:"suspectRange"`
	RelatedBugs       []*AssociatedBug   `json:"bugs"`
	RevertChangeList  *ChangeList        `json:"revertChangeList"`
	Suspects          []*SuspectSummary  `json:"suspects"`
	HeuristicAnalysis []*DetailedSuspect `json:"heuristicAnalysis"`
}

type DetailedSuspect struct {
	CommitID      string   `json:"commitID"`
	ReviewURL     string   `json:"reviewURL"`
	Score         int      `json:"score"`
	Confidence    string   `json:"confidence"`
	Justification []string `json:"justification"`
}

func newSuspectRange(
	repoURL string,
	startRevision string,
	endRevision string) *SuspectRange {
	return &SuspectRange{
		LinkText: fmt.Sprintf("%s ... %s", startRevision, endRevision),
		URL:      fmt.Sprintf("%s/+log/%s..%s", repoURL, startRevision, endRevision),
	}
}

func GetAnalysisDetails(ctx *router.Context) {
	var bbid = ctx.Params.ByName("bbid")

	buildID, err := strconv.Atoi(bbid)
	if err != nil {
		logging.Errorf(ctx.Context, "Failed to convert Buildbucket ID '%s' to integer: %s", bbid, err)
		http.Error(ctx.Writer, "Buildbucket ID could not be converted to a number", http.StatusBadRequest)
		return
	}

	// TODO: replace these hardcoded response with actual analysis details
	// query results

	var response *FailureAnalysis
	if buildID%2 == 0 {
		response = getFauxEmptyAnalysisDetails(buildID)
	} else {
		response = getFauxAnalysisDetails(buildID)
	}

	respondWithJSON(ctx, &response)
}

// Returns a faux failure analysis that doesn't have much data, which is useful
// for checking how the UI handles lack of data
func getFauxEmptyAnalysisDetails(buildID int) *FailureAnalysis {
	suspectRange := newSuspectRange("https://chromium.googlesource.com/placeholder", "cd52ae", "cd52af")

	return &FailureAnalysis{
		ID:                10000000,
		Status:            "ANALYSING",
		FailureType:       "Compile failure",
		BuildID:           buildID,
		Builder:           "builder-type-amd-rhel-cc64",
		SuspectRange:      suspectRange,
		RelatedBugs:       []*AssociatedBug{},
		Suspects:          []*SuspectSummary{},
		HeuristicAnalysis: []*DetailedSuspect{},
	}
}

// Returns a faux failure analysis that has most fields, which is useful
// for checking how the UI renders all details
func getFauxAnalysisDetails(buildID int) *FailureAnalysis {
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

	detailedSuspects := make([]*DetailedSuspect, len(suspects))
	for i, suspect := range suspects {
		detailedSuspect := &DetailedSuspect{
			CommitID:      "c9e3a" + strconv.Itoa(i),
			ReviewURL:     suspect.ReviewUrl,
			Score:         suspect.Score,
			Confidence:    heuristic.GetConfidenceLevel(suspect.Score).String(),
			Justification: strings.Split(suspect.Justification, "\n"),
		}
		detailedSuspects[i] = detailedSuspect
	}

	return &FailureAnalysis{
		ID:           10000001,
		Status:       "VERIFYING",
		FailureType:  "Compile failure",
		BuildID:      buildID,
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
		RevertChangeList: &ChangeList{
			ID:             "f23ade252",
			Title:          "Title of revert CL that was created by GoFindit",
			URL:            "https://not.a.real.link",
			Status:         "MERGED",
			SubmitTime:     "2022-02-02 16:21:13",
			CommitPosition: 77346,
		},
		Suspects: []*SuspectSummary{
			{
				ID:            "b2f50452c",
				Title:         "Short title",
				URL:           "https://www.google.com",
				CulpritStatus: "VERIFYING",
				AccuseSource:  "Heuristic",
			},
		},
		HeuristicAnalysis: detailedSuspects,
	}
}
