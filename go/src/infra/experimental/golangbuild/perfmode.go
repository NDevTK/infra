// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"fmt"
	"infra/experimental/golangbuild/golangbuildpb"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/auth"
)

// perfRunner runs performance tests and optionally uploads their results to perfdata.golang.org.
//
// For the main Go repository, it acquires a baseline toolchain and an experiment toolchain
// (the former is configured in golangbuildpb.PerfMode while the latter comes from the build's
// input sources), then runs cmd/bench in the x/benchmarks repository.
//
// For golang.org/x repos, it acquires a toolchain for the latest commit on the Go branch specified
// in the input properties, checks out the repository at the baseline commit and the commit
// specified in the input sources, and finally runs benchmarks via `go test`.
//
// TODO(mknyszek): Subrepos are not yet supported.
type perfRunner struct {
	props *golangbuildpb.PerfMode
}

// newPerfRunner creates a new PerfMode runner.
func newPerfRunner(props *golangbuildpb.PerfMode) *perfRunner {
	return &perfRunner{props: props}
}

// Run implements the runner interface for perfRunner.
func (r *perfRunner) Run(ctx context.Context, spec *buildSpec) error {
	if !isGoProject(spec.inputs.Project) {
		return fmt.Errorf("MODE_PERF not supported for subrepositories yet")
	}

	// Get a built Go toolchain or build it if necessary. This will be
	// our experiment toolchain.
	if err := getGoFromSpec(ctx, spec, false); err != nil {
		return err
	}

	// Get the baseline Go.
	gorootBaseline := filepath.Join(spec.workdir, "go_baseline")
	goBaselineSrc, err := sourceForBaseline(ctx, spec.auth, spec.goSrc, r.props.Baseline)
	if err != nil {
		return err
	}
	if err := getGo(ctx, "get baseline go", gorootBaseline, goBaselineSrc, spec.inputs, false); err != nil {
		return err
	}

	// Get the tip of the benchmarks repo.
	benchmarksSrc, err := sourceForBranch(ctx, spec.auth, publicGoHost, "benchmarks", mainBranch)
	if err != nil {
		return err
	}

	// Fetch the benchmarks repository.
	benchmarksRoot := filepath.Join(spec.workdir, "benchmarks")
	if err := fetchRepo(ctx, benchmarksSrc, benchmarksRoot, spec.inputs); err != nil {
		return err
	}

	// Construct benchmark command.
	benchCmd := spec.goCmd(ctx, benchmarksRoot, "run",
		"./cmd/bench",
		"-goroot", spec.goroot,
		"-goroot-baseline", gorootBaseline,
		"-branch", spec.goSrc.branch,
		"-repository", "go",
	)

	// Run benchmarks.
	results, err := cmdStepOutput(ctx, "go run cmd/bench", benchCmd, false)
	if err != nil {
		return err
	}

	// Summarize results with benchstat.
	benchstatCmd := toolCmd(ctx, "benchstat", "-col", "toolchain@(baseline experiment)", "-ignore", "pkg,shortname", "-")
	benchstatCmd.Stdin = bytes.NewReader(results)
	formattedResults, err := cmdStepOutput(ctx, "benchstat", benchstatCmd, true)
	if err != nil {
		return err
	}
	topLevelLog(ctx, "benchmark results").Write(formattedResults)

	// TODO(mknyszek): Upload results to perfdata.golang.org.
	return nil
}

func sourceForBaseline(ctx context.Context, auth *auth.Authenticator, src *sourceSpec, baseline string) (*sourceSpec, error) {
	if baseline == "parent" {
		return sourceForParent(ctx, auth, src)
	}
	if baseline == "latest_go_release" {
		if src.project != "go" {
			return nil, fmt.Errorf("the latest_go_release baseline is only supported for the go project")
		}
		return sourceForLatestGoRelease(ctx, auth, src.branch)
	}
	if branch, ok := strings.CutPrefix(baseline, "refs/heads/"); ok {
		return sourceForBranch(ctx, auth, publicGoHost, src.project, branch)
	}
	return nil, fmt.Errorf("invalid baseline specification %q", baseline)
}
