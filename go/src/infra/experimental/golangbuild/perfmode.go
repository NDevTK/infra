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
type perfRunner struct {
	props *golangbuildpb.PerfMode
}

// newPerfRunner creates a new PerfMode runner.
func newPerfRunner(props *golangbuildpb.PerfMode) *perfRunner {
	return &perfRunner{props: props}
}

// Run implements the runner interface for perfRunner.
func (r *perfRunner) Run(ctx context.Context, spec *buildSpec) error {
	var results []byte
	var err error
	if isGoProject(spec.inputs.Project) {
		results, err = runGoBenchmarks(ctx, spec, r.props)
	} else {
		results, err = runSubrepoBenchmarks(ctx, spec, r.props)
	}
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

func runGoBenchmarks(ctx context.Context, spec *buildSpec, perfProps *golangbuildpb.PerfMode) ([]byte, error) {
	// Get a built Go toolchain or build it if necessary. This will be
	// our experiment toolchain.
	if err := getGoFromSpec(ctx, spec, false); err != nil {
		return nil, err
	}

	// Get the baseline Go.
	gorootBaseline := filepath.Join(spec.workdir, "go_baseline")
	goBaselineSrc, err := sourceForBaseline(ctx, spec.auth, spec.goSrc, perfProps.Baseline)
	if err != nil {
		return nil, err
	}
	if err := getGo(ctx, "get baseline go", gorootBaseline, goBaselineSrc, spec.inputs, false); err != nil {
		return nil, err
	}

	// Get the tip of the benchmarks repo.
	benchmarksSrc, err := sourceForBranch(ctx, spec.auth, publicGoHost, "benchmarks", mainBranch)
	if err != nil {
		return nil, err
	}

	// Fetch the benchmarks repository.
	benchmarksRoot := filepath.Join(spec.workdir, "benchmarks")
	if err := fetchRepo(ctx, benchmarksSrc, benchmarksRoot, spec.inputs); err != nil {
		return nil, err
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
	return cmdStepOutput(ctx, "go run cmd/bench", benchCmd, false)
}

func runSubrepoBenchmarks(ctx context.Context, spec *buildSpec, perfProps *golangbuildpb.PerfMode) ([]byte, error) {
	// Fetch the subrepo at whatever we were triggered on.
	subrepoExperimentDir := filepath.Join(spec.workdir, spec.inputs.Project)
	if err := fetchRepo(ctx, spec.subrepoSrc, subrepoExperimentDir, spec.inputs); err != nil {
		return nil, err
	}

	// Fetch the subrepo at whatever baseline is in the builder configuration.
	subrepoBaselineSrc, err := sourceForBaseline(ctx, spec.auth, spec.subrepoSrc, perfProps.Baseline)
	if err != nil {
		return nil, err
	}
	subrepoBaselineDir := filepath.Join(spec.workdir, spec.inputs.Project+"_baseline")
	if err := fetchRepo(ctx, subrepoBaselineSrc, subrepoBaselineDir, spec.inputs); err != nil {
		return nil, err
	}

	// Pick the baseline Go toolchain we're going to use, which is just the latest release
	// for the Go branch this builder is building against.
	goBaselineSrc, err := sourceForLatestGoRelease(ctx, spec.auth, spec.inputs.GoBranch)
	if err != nil {
		return nil, err
	}

	// Get the baseline Go.
	gorootBaseline := filepath.Join(spec.workdir, "go_baseline")
	if err := getGo(ctx, "get baseline go", gorootBaseline, goBaselineSrc, spec.inputs, false); err != nil {
		return nil, err
	}

	// Get the tip of the benchmarks repo.
	benchmarksSrc, err := sourceForBranch(ctx, spec.auth, publicGoHost, "benchmarks", mainBranch)
	if err != nil {
		return nil, err
	}

	// Fetch the benchmarks repository.
	benchmarksRoot := filepath.Join(spec.workdir, "benchmarks")
	if err := fetchRepo(ctx, benchmarksSrc, benchmarksRoot, spec.inputs); err != nil {
		return nil, err
	}

	// Construct benchmark command.
	benchCmd := goCmd(ctx, gorootBaseline, benchmarksRoot, "run",
		"./cmd/bench",
		"-goroot-baseline", gorootBaseline,
		"-subrepo", subrepoExperimentDir,
		"-subrepo-baseline", subrepoBaselineDir,
		"-branch", spec.goSrc.branch,
		"-repository", spec.inputs.Project,
	)

	// Run benchmarks.
	return cmdStepOutput(ctx, "go run cmd/bench", benchCmd, false)
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
	return sourceForRef(ctx, auth, publicGoHost, src.project, baseline)
}
