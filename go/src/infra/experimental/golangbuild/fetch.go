// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	lucierrors "go.chromium.org/luci/common/errors"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/experimental/golangbuild/golangbuildpb"
)

const (
	// N.B. Unfortunately Go still calls the main branch "master" due to technical issues.
	mainBranch = "master" // nocheck
)

// sourceSpec indicates a repository to fetch and what state to fetch it at.
//
// One of commit and change must be non-nil.
type sourceSpec struct {
	// project is a project in host. Must not be empty.
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

// asMarkdown returns a markdown-formatted representation of the source.
func (s *sourceSpec) asMarkdown() string {
	var linkText string
	switch {
	case s.commit != nil && s.change == nil:
		linkText = fmt.Sprintf("commit %s", s.commit.Id[:7])
	case s.commit == nil && s.change != nil:
		linkText = fmt.Sprintf("change %d", s.change.Change)
	}
	return fmt.Sprintf("%s on %s ([%s](%s))", s.project, s.branch, linkText, s.asURL())
}

// asSource returns a golangbuildpb.Source for the sourceSpec.
func (s *sourceSpec) asSource() *golangbuildpb.Source {
	var src golangbuildpb.Source
	switch {
	case s.commit != nil && s.change != nil:
		panic("sourceSpec has both a change and a commit")
	case s.commit != nil && s.change == nil:
		src.GitilesCommit = s.commit
	case s.commit == nil && s.change != nil:
		src.GerritChange = s.change
	}
	return &src
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

	return cloneFromCache(ctx, inputs, change.Host, change.Project, dst, sha)
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

	if err := runGit(ctx, "git cherry-pick", "-C", dst, "cherry-pick", "--allow-empty", changeSha); err != nil {
		return err
	}

	return nil
}

// fetchRepoAtCommit checks out a commit to be tested as is.
func fetchRepoAtCommit(ctx context.Context, inputs *golangbuildpb.Inputs, commit *bbpb.GitilesCommit, dst string) error {
	_, err := fetchObjIntoCache(ctx, inputs, commit.Host, commit.Project, commit.Id)
	if err != nil {
		return err
	}

	return cloneFromCache(ctx, inputs, commit.Host, commit.Project, dst, commit.Id)
}

// fetchDependencies uses 'go mod download' to fetch
// dependencies for the given modules.
func fetchDependencies(ctx context.Context, spec *buildSpec, modules []module) (err error) {
	step, ctx := build.StartStep(ctx, "fetch dependencies")
	defer endStep(step, &err)

	var errs []error
	for _, m := range modules {
		err := goModDownload(ctx, spec, fmt.Sprintf("fetch %q dependencies", m.Path), m.RootDir)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// goModDownload runs 'go mod download' in dir in a build step.
//
// Its output (in JSON format) is parsed to determine whether a
// failure in the build step is a test or infrastructure failure.
func goModDownload(ctx context.Context, spec *buildSpec, stepName, dir string) (err error) {
	cmd := spec.goCmd(ctx, dir, "mod", "download", "-json")
	// Invest reasonable effort to estimate whether the failure was likely due to
	// an infrastructure problem rather than a problem in the input being tested.
	//
	// It's viable to default to non-infra unless a known-infra error is seen, or vice versa,
	// so use whichever strikes a better balance of low false positives and maintenance costs.
	var infra bool
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return err
	}
	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(step.Log("stdout"), &stdout)
	cmd.Stderr = step.Log("stderr")
	err = cmd.Run()
	if ee := (*exec.ExitError)(nil); errors.As(err, &ee) && ee.ExitCode() == 1 {
		for dec := json.NewDecoder(&stdout); ; {
			var m struct{ Error string }
			err := dec.Decode(&m)
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("error decoding JSON object from go mod download -json: %w\n", err)
			}
			if strings.Contains(m.Error, "dial tcp") && strings.HasSuffix(m.Error, ": i/o timeout") {
				// An I/O timeout error to a Go module proxy is deemed to be an infrastructure failure.
				// See https://ci.chromium.org/b/8772399708036918561 for an example.
				infra = true
				break
			} else if strings.HasSuffix(m.Error, ": EOF") {
				// An EOF error from a Go module proxy is deemed to be an infrastructure failure.
				// See go.dev/issue/63684 and https://ci.chromium.org/b/8766452547270027457 for an example.
				infra = true
				break
			}
		}
	}
	if err != nil {
		return lucierrors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
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
func sourceForBranch(ctx context.Context, auth *auth.Authenticator, host, project, branch string) (*sourceSpec, error) {
	hc, err := auth.Client()
	if err != nil {
		return nil, fmt.Errorf("auth.Client: %w", err)
	}
	gc, err := gitiles.NewRESTClient(hc, host, true)
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
			Host:    host,
			Project: project,
			Id:      log.Log[0].Id,
			Ref:     ref,
		},
	}, nil
}

// sourceForGoBranchAndCommit produces a sourceSpec representing
// the specified branch and commit in the main Go repo.
func sourceForGoBranchAndCommit(host, branch, commit string) (*sourceSpec, error) {
	if branch == "" {
		return nil, fmt.Errorf("empty branch")
	}
	if len(commit) != len("4368e1cdfd37cbcdbc7a4fbcc78ad61139f7ba90") {
		return nil, fmt.Errorf("unsupported commit ID %q: length is %d, want 40", commit, len(commit))
	}
	for _, c := range commit {
		if ok := '0' <= c && c <= '9' || 'a' <= c && c <= 'f'; !ok {
			return nil, fmt.Errorf("unsupported commit ID %q: contains %q, want 0-9a-f only", commit, c)
		}
	}
	return &sourceSpec{
		project: "go",
		branch:  branch,
		commit: &bbpb.GitilesCommit{
			Host:    host,
			Project: "go",
			Id:      commit,
			Ref:     "refs/heads/" + branch,
		},
	}, nil
}
