// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"net/http"
	"strconv"

	gfim "infra/appengine/gofindit/model"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server/router"
)

type ChangeList struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Url            string `json:"url"`
	Status         string `json:"status"`
	SubmitTime     string `json:"submitTime"`
	CommitPosition int    `json:"commitPosition"`
}

type SuspectSummary struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Url           string `json:"url"`
	CulpritStatus string `json:"culpritStatus"`
	AccuseSource  string `json:"accuseSource"`
}

type FailureAnalysis struct {
	Id                int                           `json:"id"`
	Status            string                        `json:"status"`
	FailureType       string                        `json:"failureType"`
	BuildbucketId     int                           `json:"buildbucketId"`
	Builder           string                        `json:"builder"`
	SuspectRange      []string                      `json:"suspectRange"`
	RelatedBugs       []string                      `json:"bugs"`
	RevertChangeList  ChangeList                    `json:"revertChangeList"`
	Suspects          []SuspectSummary              `json:"suspects"`
	HeuristicAnalysis *gfim.HeuristicAnalysisResult `json:"heuristicAnalysis"`
}

func (h *Handlers) GetAnalysisDetails(ctx *router.Context) {
	var buildbucketIdParam = ctx.Params.ByName("buildbucketId")

	id, err := strconv.Atoi(buildbucketIdParam)
	if err != nil {
		logging.Errorf(ctx.Context, "Failed to convert Buildbucket ID '%s' to integer: %s", buildbucketIdParam, err)
		http.Error(ctx.Writer, "Buildbucket ID could not be converted to a number", http.StatusBadRequest)
		return
	}

	// TODO: replace this hardcoded response with actual analysis details
	// query results
	response := FailureAnalysis{
		Id:            10000001,
		Status:        "VERIFYING",
		FailureType:   "Compile failure",
		BuildbucketId: id,
		Builder:       "builder-type-amd-rhel-cc64",
		SuspectRange:  []string{"cd52ae", "cd52af"},
		RelatedBugs:   []string{"cr/23527"},
		RevertChangeList: ChangeList{
			Id:             "f23ade252",
			Title:          "Title of revert CL that was created by GoFindit",
			Url:            "https://not.a.real.link",
			Status:         "MERGED",
			SubmitTime:     "2022-02-02 16:21:13",
			CommitPosition: 77346,
		},
		Suspects: []SuspectSummary{
			{
				Id:            "b2f50452c",
				Title:         "Short title",
				Url:           "https://www.google.com",
				CulpritStatus: "VERIFYING",
				AccuseSource:  "Heuristic",
			},
		},
		HeuristicAnalysis: &gfim.HeuristicAnalysisResult{
			Items: []*gfim.HeuristicAnalysisResultItem{
				{
					Commit:    "wxyz",
					ReviewUrl: "https://chromium-review.googlesource.com/",
					Justification: &gfim.SuspectJustification{
						Items: []*gfim.SuspectJustificationItem{
							{
								Score:    5,
								FilePath: "dir/a/b/x.cc",
								Reason:   `The file "dir/a/b/x.cc" was added and it was in the failure log.`,
							},
							{
								Score:    5,
								FilePath: "dir/a/b/y.cc",
								Reason:   `The file "dir/a/b/y.cc" was added and it was in the failure log.`,
							},
							{
								Score:    5,
								FilePath: "dir/a/b/z.cc",
								Reason:   `The file "dir/a/b/z.cc" was added and it was in the failure log.`,
							},
						},
					},
				},
				{
					Commit:    "abcd",
					ReviewUrl: "https://chromium-review.googlesource.com/",
					Justification: &gfim.SuspectJustification{
						Items: []*gfim.SuspectJustificationItem{
							{
								Score:    2,
								FilePath: "content/util.c",
								Reason:   "The file \"content/util.c\" was modified. It was related to the file obj/content/util.o which was in the failure log.",
							},
						},
					},
				},
			},
		},
	}

	respondWithJSON(ctx, response)
}
