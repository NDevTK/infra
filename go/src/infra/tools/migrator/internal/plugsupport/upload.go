// Copyright 2021 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package plugsupport

import (
	"context"
	"io/ioutil"
	"strings"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"infra/tools/migrator"
)

// ExecuteUpload implements "upload" subcommand.
func ExecuteUpload(ctx context.Context, projectDir ProjectDir, force bool) (*migrator.ReportDump, error) {
	tweaks, err := LoadTweaks(projectDir)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load tweaks").Err()
	}

	blob, err := ioutil.ReadFile(projectDir.CommitMessageFile())
	if err != nil {
		return nil, errors.Annotate(err, "failed to read the commit message").Err()
	}
	message := string(blob)

	return visitReposInParallel(ctx, projectDir, projectDir.UploadReportPath(), func(ctx context.Context, r *repo) {
		reviewers := stringset.New(0)
		cc := stringset.New(0)
		for _, proj := range r.projects {
			projectTweaks := tweaks.ProjectTweaks(proj.Id)
			reviewers.AddAll(projectTweaks.Reviewers)
			cc.AddAll(projectTweaks.CC)
		}

		git := r.git(ctx)
		defer func() {
			if git.err != nil {
				logging.Errorf(ctx, "%s", git.err)
				r.report(ctx, "GIT_ERROR", git.err.Error())
			}
		}()

		// Examine the current state of the checkout.
		uncommittedDiff := git.read("diff", "HEAD", "--name-only")
		localCommit := git.read("rev-list", "--count", "@{u}..HEAD") != "0"
		remoteCL := git.gerritCL()
		if git.err != nil {
			return
		}

		// Skip completely untouched checkouts.
		if uncommittedDiff == "" && !localCommit {
			return
		}

		// Skip if the change has already been uploaded.
		if !force && uncommittedDiff == "" && remoteCL != "" {
			r.report(ctx, "UNCHANGED", "No new changes", migrator.MetadataOption("CL", remoteCL))
			return
		}

		// Prepare the local commit or amend the existing one (if any).
		commitCmd := []string{
			"commit", "--quiet", "--all", "--no-edit", "--message", message,
		}
		if localCommit {
			commitCmd = append(commitCmd, "--amend")
		}
		git.run(commitCmd...)
		if git.err != nil {
			return
		}

		// Upload it as a CL.
		uploadArgs := []string{
			"cl", "upload",
			"--force", "--bypass-hooks",
			"--message", message,
			"--title", "migrator change",
		}
		if len(reviewers) != 0 {
			uploadArgs = append(uploadArgs, "--reviewers", strings.Join(reviewers.ToSortedSlice(), ","))
		} else {
			uploadArgs = append(uploadArgs, "--r-owners")
		}
		if len(cc) != 0 {
			uploadArgs = append(uploadArgs, "--cc", strings.Join(cc.ToSortedSlice(), ","))
		}
		git.run(uploadArgs...)

		// We should have a CL link now for sure.
		clMD := migrator.MetadataOption("CL", git.gerritCL())

		if remoteCL != "" {
			// We already had a CL and it was updated.
			r.report(ctx, "UPDATED", "Updated the CL", clMD)
		} else {
			// We just created a completely new CL.
			r.report(ctx, "UPLOADED", "Created the CL", clMD)
		}
	})
}
