// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package culpritverification verifies if a suspect is a culprit.
package culpritverification

import (
	"context"
	"infra/appengine/gofindit/internal/gitiles"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/logging"
)

// VerifyCulprit checks if a commit is the culprit of a build failure.
func VerifyCulprit(c context.Context, commit *buildbucketpb.GitilesCommit, failedBuildId int64) error {
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
	triggerRerun(c, commit, failedBuildId)
	triggerRerun(c, parentCommit, failedBuildId)
	return nil
}

func triggerRerun(c context.Context, commit *buildbucketpb.GitilesCommit, failedBuildId int64) {
	// TODO (nqmtuan): Trigger the rerun build for the commit
	logging.Infof(c, "triggerRerun with commit %s", commit.Id)
}
