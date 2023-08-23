// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gitiles"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/experimental/golangbuild/golangbuildpb"
)

// TODO(yifany): `goHost` is used in the implementation for deriving the build
// specs in the file `buildspec.go`. The current implementation works under the
// the situation that:
//   - development of subrepos is in the public gerrit host,
//   - tests from subrepos are always against corresponding branch in the go
//     repo in the public gerrit host.
//
// If these conditions changed, the implementation need to adjust accordingly.
// At that point, the const declarations here may no longer be needed.
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

	// cherryPick controls whether to cherry-pick change onto branch.
	//
	// This field may only be true when change is non-nil.
	//
	// Note that this will cherry-pick change without any of its parent
	// changes (if any), thus testing the change in isolation.
	cherryPick bool
}

// asURL returns a URL string for the sourceSpec.
func (s *sourceSpec) asURL() string {
	switch {
	case s.commit != nil && s.change != nil:
		panic("sourceSpec has both a change and a commit")
	case s.commit != nil && s.change == nil:
		return fmt.Sprintf("https://%s/%s/+/%s", s.commit.Host, s.commit.Project, s.commit.Id)
	case s.commit == nil && s.change != nil:
		return fmt.Sprintf("https://%s/c/%s/+/%d/%d", s.change.Host, s.change.Project, s.change.Change, s.change.Patchset)
	}
	panic("no commit or change in sourceSpec")
}

// fetchRepo fetches a repository according to src and places it at dst.
func fetchRepo(ctx context.Context, src *sourceSpec, dst string, inputs *golangbuildpb.Inputs) (err error) {
	step, ctx := build.StartStep(ctx, "fetch "+src.project)
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.

	switch {
	case src.change != nil && !src.cherryPick:
		return fetchRepoChangeAsIs(ctx, inputs, src.change, dst)
	case src.change != nil && src.cherryPick:
		return fetchRepoChangeWithCherryPick(ctx, inputs, src.branch, src.change, dst)
	case src.commit != nil:
		if src.cherryPick {
			return fmt.Errorf("cherryPick is unexpectedly set in the commit case")
		}
		return fetchRepoAtCommit(ctx, inputs, src.commit, dst)
	}
	return fmt.Errorf("one of change or commit must be non-nil")
}

// cachedRepoPath returns the path to the cached bare repo for this gerrit
// change, performing initial clone if it is missing from cache.
func cachedRepoPath(ctx context.Context, inputs *golangbuildpb.Inputs, host, project string) (string, error) {
	cache := inputs.GitCache
	if cache == "" {
		return "", fmt.Errorf("inputs missing GitCache: %+v", inputs)
	}
	if !filepath.IsLocal(cache) {
		return "", fmt.Errorf("GitCache %q must be relative", cache)
	}

	// Double check that the host and project don't form malformed paths.
	if !filepath.IsLocal(host) {
		return "", fmt.Errorf("host %q must make a relative path", host)
	}
	if !filepath.IsLocal(project) {
		return "", fmt.Errorf("project %q must make a relative path", project)
	}

	luciExe := lucictx.GetLUCIExe(ctx)
	if luciExe == nil {
		return "", fmt.Errorf("missing LUCI_CONTEXT")
	}

	// N.B. host is included in the path so that we don't mix contents from
	// potentially conflicting repos with the same project name. e.g.,
	// go/go and go-internal/go.
	//
	// Once annoyance with this approach is that GitilesChange uses host
	// "go.googlesource.com" while GerritChange uses
	// "go-review.googlesource.com", so those repos will be duplicated.
	repo := filepath.Join(luciExe.GetCacheDir(), cache, host, project)

	_, err := os.Stat(repo)
	if err == nil {
		// Repo already exists.
		return repo, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		// Some other error.
		return "", fmt.Errorf("error checking for repo existence: %w", err)
	}

	// Repo does not exist; perform initial clone.

	if err := runGit(ctx, "git clone upstream to cache", "clone", "--bare", urlFromHostProject(host, project), repo); err != nil {
		return "", err
	}

	return repo, nil
}

// Fetch a remote object (a ref or explicit commit SHA) from host/project into
// the cache repo. Returns the explicit commit SHA that obj resolves to
// (unnecessary if obj is already an explicit commit SHA).
//
// N.B. We don't need to worry about git gc pruning these unreferenced objects
// before we get a chance to clone to repo because git gc never prunes objects
// less than 2 weeks old.
func fetchObjIntoCache(ctx context.Context, inputs *golangbuildpb.Inputs, host, project, obj string) (string, error) {
	cacheRepo, err := cachedRepoPath(ctx, inputs, host, project)
	if err != nil {
		return "", err
	}

	// Fetch object into cached repo (may be a no-op).
	if err := runGit(ctx, "git fetch into cache", "--git-dir", cacheRepo, "fetch", urlFromHostProject(host, project), obj); err != nil {
		return "", err
	}

	// Resolve fetched commit for subsequent fetch. We need this because
	// the fetched refs aren't named in the cache repo, so they need to be
	// fetched by explicit sha.
	//
	// This is useless if obj is already a sha, but it doesn't hurt.
	out, err := runGitOutput(ctx, "git rev-parse", "--git-dir", cacheRepo, "rev-parse", "FETCH_HEAD")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(out))

	return sha, nil
}

// Clone the cache repo to dst and checkout sha.
func cloneFromCache(ctx context.Context, inputs *golangbuildpb.Inputs, host, project, dst, sha string) error {
	cacheRepo, err := cachedRepoPath(ctx, inputs, host, project)
	if err != nil {
		return err
	}

	// It is very tempting to be clever and try to "optimize" this. e.g.,
	// by creating a new repo and fetching only the objects we need from
	// the cache. However, that ends up being less efficient because fetch
	// is fairly expensive, while a local clone is optimized to a direct
	// copy of the .git directory (see --local in git help clone), which is
	// much faster.
	if err := runGit(ctx, "git clone from cache", "clone", cacheRepo, dst); err != nil {
		return err
	}

	if err := runGit(ctx, "git checkout", "-C", dst, "checkout", sha); err != nil {
		return err
	}

	return nil
}

// fetchRepoChangeAsIs checks out a change to be tested as is, without
// rebasing.
func fetchRepoChangeAsIs(ctx context.Context, inputs *golangbuildpb.Inputs, change *bbpb.GerritChange, dst string) error {
	ref := refFromChange(change)
	sha, err := fetchObjIntoCache(ctx, inputs, change.Host, change.Project, ref)
	if err != nil {
		return err
	}

	if err := cloneFromCache(ctx, inputs, change.Host, change.Project, dst, sha); err != nil {
		return err
	}

	if change.Project == "go" {
		if err := writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset)); err != nil {
			return err
		}
	}

	return nil
}

// fetchRepoChangeWithCherryPick checks out a change by cherry-picking it on
// top of branch.
func fetchRepoChangeWithCherryPick(ctx context.Context, inputs *golangbuildpb.Inputs, branch string, change *bbpb.GerritChange, dst string) error {
	// Fetch branch and change into cache so they will be both be available
	// in the clone.
	branchRef := "refs/heads/" + branch
	branchSha, err := fetchObjIntoCache(ctx, inputs, change.Host, change.Project, branchRef)
	if err != nil {
		return err
	}
	changeRef := refFromChange(change)
	changeSha, err := fetchObjIntoCache(ctx, inputs, change.Host, change.Project, changeRef)
	if err != nil {
		return err
	}

	if err := cloneFromCache(ctx, inputs, change.Host, change.Project, dst, branchSha); err != nil {
		return err
	}

	if err := runGit(ctx, "git cherry-pick", "-C", dst, "cherry-pick", changeSha); err != nil {
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
func fetchRepoAtCommit(ctx context.Context, inputs *golangbuildpb.Inputs, commit *bbpb.GitilesCommit, dst string) error {
	_, err := fetchObjIntoCache(ctx, inputs, commit.Host, commit.Project, commit.Id)
	if err != nil {
		return err
	}

	if err := cloneFromCache(ctx, inputs, commit.Host, commit.Project, dst, commit.Id); err != nil {
		return err
	}

	if commit.Project == "go" {
		if err := writeVersionFile(ctx, dst, commit.Id); err != nil {
			return err
		}
	}
	return nil
}

// fetchDependencies uses 'go mod download' to fetch
// dependencies for the given modules.
func fetchDependencies(ctx context.Context, spec *buildSpec, modules []module) (err error) {
	step, ctx := build.StartStep(ctx, "fetch dependencies")
	defer endStep(step, &err)

	var errs []error
	for _, m := range modules {
		dlCmd := spec.goCmd(ctx, m.RootDir, "mod", "download", "-json")
		err := goModDownloadStep(ctx, fmt.Sprintf("fetch %q dependencies", m.Path), dlCmd)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func writeVersionFile(ctx context.Context, dst, version string) error {
	return writeFile(ctx, filepath.Join(dst, "VERSION"), "devel "+version)
}

func writeFile(ctx context.Context, path, data string) (err error) {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("write %s", filepath.Base(path)))
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.
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

func urlFromHostProject(host, project string) string {
	return "https://" + host + "/" + project
}

func refFromChange(change *bbpb.GerritChange) string {
	return fmt.Sprintf("refs/changes/%02d/%d/%d", change.Change%100, change.Change, change.Patchset)
}

func runGit(ctx context.Context, stepName string, args ...string) (err error) {
	return cmdStepRun(ctx, stepName, exec.CommandContext(ctx, "git", args...), true)
}

func runGitOutput(ctx context.Context, stepName string, args ...string) (output []byte, err error) {
	return cmdStepOutput(ctx, stepName, exec.CommandContext(ctx, "git", args...), true)
}

// sourceForBranch produces a sourceSpec representing the tip of a branch for a project.
func sourceForBranch(ctx context.Context, auth *auth.Authenticator, project, branch string) (*sourceSpec, error) {
	hc, err := auth.Client()
	if err != nil {
		return nil, fmt.Errorf("auth.Client: %w", err)
	}
	gc, err := gitiles.NewRESTClient(hc, goHost, true)
	if err != nil {
		return nil, fmt.Errorf("gitiles.NewRESTClient: %w", err)
	}
	ref := fmt.Sprintf("refs/heads/%s", branch)
	log, err := gc.Log(ctx, &gitilespb.LogRequest{
		Project:    project,
		Committish: ref,
		PageSize:   1,
	})
	if err != nil {
		return nil, fmt.Errorf("gc.Log: %w", err)
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
