// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package culpritverification verifies if a suspect is a culprit.
package culpritverification

import (
	"context"
	"infra/appengine/gofindit/internal/gitiles"
	"infra/appengine/gofindit/rerun"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

// VerifyCulprit checks if a commit is the culprit of a build failure.
func VerifyCulprit(c context.Context, commit *buildbucketpb.GitilesCommit, failedBuildID int64) error {
	// Query Gitiles to get parent commit
	repoUrl := gitiles.GetRepoUrl(c, commit)
	p, err := gitiles.GetParentCommit(c, repoUrl, commit.Id)
	if err != nil {
		return err
	}
	parentCommit := &buildbucketpb.GitilesCommit{
		Host:    commit.Host,
		Project: commit.Project,
		Ref:     commit.Ref,
		Id:      p,
	}

	// Trigger a rerun with commit and parent commit
	rerun.TriggerRerun(c, commit, failedBuildID)
	rerun.TriggerRerun(c, parentCommit, failedBuildID)
	return nil
}
