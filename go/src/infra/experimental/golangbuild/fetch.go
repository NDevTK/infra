// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gitiles"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/luciexe/build"
)

const (
	goHost       = "go.googlesource.com"
	goReviewHost = "go-review.googlesource.com"
	// N.B. Unfortunately Go still calls the main branch "master" due to technical issues.
	mainBranch = "master" // nocheck
)

// sourceSpec indicates a repository to fetch and what state to fetch it at.
//
// One of commit and change must be non-nil.
type sourceSpec struct {
	// project is a go.googlesource.com project. Must not be empty.
	project string

	// branch is the branch of project that change and/or commit are on. Must not be empty.
	// branch is derived from and lines up with commit.Ref if commit != nil.
	branch string

	// change is a Gerrit CL to fetch. If this is non-nil, commit must be nil.
	change *bbpb.GerritChange

	// commit is a Gitiles commit to fetch. If this is non-nil, change must be nil.
	commit *bbpb.GitilesCommit

	// rebase controls whether to rebase onto branch.
	// This field applies when change is non-nil only, and must be false otherwise.
	rebase bool
}

// fetchRepo fetches a repository according to src and places it at dst.
func fetchRepo(ctx context.Context, src *sourceSpec, dst string) (err error) {
	step, ctx := build.StartStep(ctx, "fetch "+src.project)
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

	switch {
	case src.change != nil && !src.rebase:
		return fetchRepoChangeAsIs(ctx, dst, src.change)
	case src.change != nil && src.rebase:
		return fetchRepoChangeWithRebase(ctx, src.branch, dst, src.change)
	case src.commit != nil:
		if src.rebase {
			return fmt.Errorf("rebase is unexpectedly set in the commit case")
		}
		return fetchRepoAtCommit(ctx, dst, src.commit)
	}
	return fmt.Errorf("one of change or commit must be non-nil")
}

// fetchRepoChangeAsIs checks out a change to be tested as is, without rebasing.
func fetchRepoChangeAsIs(ctx context.Context, dst string, change *bbpb.GerritChange) error {
	// TODO(mknyszek): We're cloning tip here then fetching what we actually want because git doesn't
	// provide a good way to clone at a specific ref or commit. Is there a way to speed this up?
	// Maybe caching is sufficient?
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "https://"+change.Host+"/"+change.Project, dst); err != nil {
		return err
	}
	ref := refFromChange(change)
	if err := runGit(ctx, "git fetch", "-C", dst, "fetch", "https://"+change.Host+"/"+change.Project, ref); err != nil {
		return err
	}
	if err := runGit(ctx, "git checkout", "-C", dst, "checkout", "FETCH_HEAD"); err != nil {
		return err
	}
	if change.Project == "go" {
		if err := writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset)); err != nil {
			return err
		}
	}
	return nil
}

// fetchRepoChangeWithRebase checks out a change, rebasing it on top of its branch.
func fetchRepoChangeWithRebase(ctx context.Context, branch, dst string, change *bbpb.GerritChange) error {
	// For submit, fetch HEAD for the branch this change is for, fetch the CL, and cherry-pick it.
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "-b", branch, "https://"+change.Host+"/"+change.Project, dst); err != nil {
		return err
	}
	ref := refFromChange(change)
	if err := runGit(ctx, "git fetch", "-C", dst, "fetch", "https://"+change.Host+"/"+change.Project, ref); err != nil {
		return err
	}
	if err := runGit(ctx, "git cherry-pick", "-C", dst, "cherry-pick", "FETCH_HEAD"); err != nil {
		return err
	}
	if change.Project == "go" {
		if err := writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset)); err != nil {
			return err
		}
	}
	return nil
}

// fetchRepoAtCommit checks out a commit to be tested as is.
func fetchRepoAtCommit(ctx context.Context, dst string, commit *bbpb.GitilesCommit) error {
	// TODO(mknyszek): This is a full git checkout, which is wasteful. Consider caching.
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "https://"+commit.Host+"/"+commit.Project, dst); err != nil {
		return err
	}
	if err := runGit(ctx, "git checkout", "-C", dst, "checkout", commit.Id); err != nil {
		return err
	}
	if commit.Project == "go" {
		if err := writeVersionFile(ctx, dst, commit.Id); err != nil {
			return err
		}
	}
	return nil
}

func writeVersionFile(ctx context.Context, dst, version string) error {
	return writeFile(ctx, filepath.Join(dst, "VERSION"), "devel "+version)
}

func writeFile(ctx context.Context, path, data string) (err error) {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("write %s", filepath.Base(path)))
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()
	contentsLog := step.Log("contents")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		r := f.Close()
		if err == nil {
			err = r
		} else {
			if _, r2 := io.WriteString(step.Log("close error"), r.Error()); r2 != nil {
				log.Printf("%v", r2)
			}
		}
	}()
	_, err = io.WriteString(io.MultiWriter(contentsLog, f), data)
	return err
}

func refFromChange(change *bbpb.GerritChange) string {
	return fmt.Sprintf("refs/changes/%02d/%d/%d", change.Change%100, change.Change, change.Patchset)
}

func runGit(ctx context.Context, stepName string, args ...string) (err error) {
	return runCommandAsStep(ctx, stepName, exec.CommandContext(ctx, "git", args...), true)
}

// sourceForBranch produces a sourceSpec representing the tip of a branch for a project.
func sourceForBranch(ctx context.Context, auth *auth.Authenticator, project, branch string) (*sourceSpec, error) {
	hc, err := auth.Client()
	if err != nil {
		return nil, err
	}
	gc, err := gitiles.NewRESTClient(hc, goHost, true)
	if err != nil {
		return nil, err
	}
	ref := fmt.Sprintf("refs/heads/%s", branch)
	log, err := gc.Log(ctx, &gitilespb.LogRequest{
		Project:    project,
		Committish: ref,
		PageSize:   1,
	})
	if err != nil {
		return nil, err
	}
	if len(log.Log) == 0 {
		return nil, fmt.Errorf("no commits found for project %s on branch %s", project, branch)
	}
	return &sourceSpec{
		project: project,
		branch:  branch,
		commit: &bbpb.GitilesCommit{
			Host:    goHost,
			Project: project,
			Id:      log.Log[0].Id,
			Ref:     ref,
		},
	}, nil
}

func (s *sourceSpec) prebuiltID() (string, error) {
	if s.project != "go" {
		return "", fmt.Errorf("prebuilt Go ID only applies if project is 'go'; project is %q", s.project)
	}
	var rev string
	if s.commit != nil {
		rev = s.commit.Id
	} else if s.change != nil {
		rev = fmt.Sprintf("%d-%d", s.change.Change, s.change.Patchset)
	} else {
		return "", fmt.Errorf("source for (%s, %s) has no change or commit", s.project, s.branch)
	}
	return fmt.Sprintf("%s-%s-%s", hostGOOS, hostGOARCH, rev), nil
}
