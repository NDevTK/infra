// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package culpritverification verifies if a suspect is a culprit.
package culpritverification

import (
	"context"

	"infra/appengine/gofindit/internal/gitiles"
	gfim "infra/appengine/gofindit/model"
	"infra/appengine/gofindit/rerun"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
)

// VerifySuspect verifies if a suspect is indeed the culprit.
// analysisID is CompileFailureAnalysis ID. It is meant to be propagated all the way to the
// recipe, so we can identify the analysis in buildbucket.
func VerifySuspect(c context.Context, suspect *gfim.Suspect, failedBuildID int64, analysisID int64) error {
	logging.Infof(c, "Verifying suspect %d for build %d", datastore.KeyForObj(c, suspect).IntID())
	suspectBuild, parentBuild, err := VerifyCommit(c, &suspect.GitilesCommit, failedBuildID, analysisID)
	if err != nil {
		logging.Errorf(c, "Error triggering rerun for build %d", failedBuildID)
		return err
	}
	suspectRerunBuildModel, err := createRerunBuildModel(c, suspectBuild, suspect)
	if err != nil {
		return err
	}

	parentRerunBuildModel, err := createRerunBuildModel(c, parentBuild, suspect)
	if err != nil {
		return err
	}

	suspect.VerificationStatus = gfim.SuspectVerificationStatus_UnderVerification
	suspect.SuspectRerunBuild = datastore.KeyForObj(c, suspectRerunBuildModel)
	suspect.ParentRerunBuild = datastore.KeyForObj(c, parentRerunBuildModel)
	err = datastore.Put(c, suspect)
	if err != nil {
		return err
	}
	return nil
}

func createRerunBuildModel(c context.Context, build *buildbucketpb.Build, suspect *gfim.Suspect) (*gfim.CompileRerunBuild, error) {
	rerunBuild := &gfim.CompileRerunBuild{
		Id:      build.GetId(),
		Type:    gfim.RerunBuildType_CulpritVerification,
		Suspect: datastore.KeyForObj(c, suspect),
		LuciBuild: gfim.LuciBuild{
			BuildId:       build.GetId(),
			Project:       build.Builder.Project,
			Bucket:        build.Builder.Bucket,
			Builder:       build.Builder.Builder,
			CreateTime:    build.CreateTime.AsTime(),
			StartTime:     build.StartTime.AsTime(),
			Status:        build.GetStatus(),
			GitilesCommit: *build.GetInput().GetGitilesCommit(),
		},
	}
	err := datastore.Put(c, rerunBuild)
	if err != nil {
		logging.Errorf(c, "Error in creating CompileRerunBuild model for build %d", build.GetId())
		return nil, err
	}
	return rerunBuild, nil
}

// VerifyCommit checks if a commit is the culprit of a build failure.
// Returns 2 builds:
// - The 1st build is the rerun build for the commit
// - The 2nd build is the rerun build for the parent commit
func VerifyCommit(c context.Context, commit *buildbucketpb.GitilesCommit, failedBuildID int64, analysisID int64) (*buildbucketpb.Build, *buildbucketpb.Build, error) {
	// Query Gitiles to get parent commit
	repoUrl := gitiles.GetRepoUrl(c, commit)
	p, err := gitiles.GetParentCommit(c, repoUrl, commit.Id)
	if err != nil {
		return nil, nil, err
	}
	parentCommit := &buildbucketpb.GitilesCommit{
		Host:    commit.Host,
		Project: commit.Project,
		Ref:     commit.Ref,
		Id:      p,
	}

	// Trigger a rerun with commit and parent commit
	build1, err := rerun.TriggerRerun(c, commit, failedBuildID, analysisID)
	if err != nil {
		return nil, nil, err
	}

	build2, err := rerun.TriggerRerun(c, parentCommit, failedBuildID, analysisID)
	if err != nil {
		return nil, nil, err
	}

	return build1, build2, nil
}
