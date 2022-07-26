// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package api contains the utility functions and APIs required to populate the
// GoFindit frontend
package api

import (
	"net/http"
	"strconv"
	"strings"

	"infra/appengine/gofindit/compilefailureanalysis/heuristic"
	gfim "infra/appengine/gofindit/model"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

type ChangeList struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Url            string `json:"url"`
	Status         string `json:"status"`
	SubmitTime     string `json:"submitTime"`
	CommitPosition int    `json:"commitPosition"`
}

type SuspectSummary struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Url           string `json:"url"`
	CulpritStatus string `json:"culpritStatus"`
	AccuseSource  string `json:"accuseSource"`
}

type FailureAnalysis struct {
	ID                int                `json:"id"`
	Status            string             `json:"status"`
	FailureType       string             `json:"failureType"`
	BuildID           int                `json:"buildID"`
	Builder           string             `json:"builder"`
	SuspectRange      []string           `json:"suspectRange"`
	RelatedBugs       []string           `json:"bugs"`
	RevertChangeList  ChangeList         `json:"revertChangeList"`
	Suspects          []SuspectSummary   `json:"suspects"`
	HeuristicAnalysis []*DetailedSuspect `json:"heuristicAnalysis"`
}

type DetailedSuspect struct {
	CommitID      string   `json:"commitID"`
	ReviewUrl     string   `json:"reviewURL"`
	Score         int      `json:"score"`
	Confidence    string   `json:"confidence"`
	Justification []string `json:"justification"`
}

func GetAnalysisDetails(ctx *router.Context) {
	var bbid = ctx.Params.ByName("bbid")

	buildID, err := strconv.Atoi(bbid)
	if err != nil {
		logging.Errorf(ctx.Context, "Failed to convert Buildbucket ID '%s' to integer: %s", bbid, err)
		http.Error(ctx.Writer, "Buildbucket ID could not be converted to a number", http.StatusBadRequest)
		return
	}

	// TODO: replace this hardcoded response with actual analysis details
	// query results

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
			ReviewUrl:     suspect.ReviewUrl,
			Score:         suspect.Score,
			Confidence:    heuristic.GetConfidenceLevel(suspect.Score).String(),
			Justification: strings.Split(suspect.Justification, "\n"),
		}
		detailedSuspects[i] = detailedSuspect
	}

	response := FailureAnalysis{
		ID:           10000001,
		Status:       "VERIFYING",
		FailureType:  "Compile failure",
		BuildID:      buildID,
		Builder:      "builder-type-amd-rhel-cc64",
		SuspectRange: []string{"cd52ae", "cd52af"},
		RelatedBugs:  []string{"crbug.com/23527", "crbug.com/23528"},
		RevertChangeList: ChangeList{
			ID:             "f23ade252",
			Title:          "Title of revert CL that was created by GoFindit",
			Url:            "https://not.a.real.link",
			Status:         "MERGED",
			SubmitTime:     "2022-02-02 16:21:13",
			CommitPosition: 77346,
		},
		Suspects: []SuspectSummary{
			{
				ID:            "b2f50452c",
				Title:         "Short title",
				Url:           "https://www.google.com",
				CulpritStatus: "VERIFYING",
				AccuseSource:  "Heuristic",
			},
		},
		HeuristicAnalysis: detailedSuspects,
	}

	respondWithJSON(ctx, response)
}
