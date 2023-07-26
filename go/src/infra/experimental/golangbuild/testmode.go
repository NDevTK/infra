// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"infra/experimental/golangbuild/golangbuildpb"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"
)

// testRunner runs a non-strict subset of available tests. It requires a prebuilt
// toolchain to be available (it will not create one on-demand).
//
// This implements "test mode" for golangbuild.
type testRunner struct {
	props *golangbuildpb.TestMode
	shard testShard
}

// newTestRunner creates a new TestMode runner.
func newTestRunner(props *golangbuildpb.TestMode, gotShard *golangbuildpb.TestShard) (*testRunner, error) {
	shard := noSharding
	if gotShard != nil {
		shard = testShard{
			shardID: gotShard.ShardId,
			nShards: gotShard.NumShards,
		}
		if shard.shardID >= shard.nShards {
			return nil, fmt.Errorf("invalid test shard designation: shard ID is %d, num shards is %d", shard.shardID, shard.nShards)
		}
	}
	return &testRunner{props: props, shard: shard}, nil
}

// Run implements the runner interface for testRunner.
func (r *testRunner) Run(ctx context.Context, spec *buildSpec) error {
	// Get a built Go toolchain and require it to be prebuilt.
	if err := getGo(ctx, spec, true); err != nil {
		return err
	}
	if spec.inputs.Project == "go" {
		return runGoTests(ctx, spec, r.shard)
	}
	return fetchSubrepoAndRunTests(ctx, spec)
}

// testShard is a test shard identity that can be used to deterministically filter tests.
type testShard struct {
	shardID uint32
	nShards uint32
}

// shouldRunTest deterministically returns true for whether the shard identity should run
// the test by name. The name of the test doesn't matter, as long as it's consistent across
// test shards.
func (s testShard) shouldRunTest(name string) bool {
	return crc32.ChecksumIEEE([]byte(name))%s.nShards == s.shardID
}

// noSharding indicates that no sharding should take place. It represents executing the entirety
// of the test suite.
var noSharding = testShard{shardID: 0, nShards: 1}

func runGoTests(ctx context.Context, spec *buildSpec, shard testShard) error {
	if spec.inputs.Project != "go" {
		return infraErrorf("runGoTests called for a subrepo builder")
	}
	gorootSrc := filepath.Join(spec.goroot, "src")

	// We have two paths, unfortunately: a simple one for Go 1.21+ that uses dist test -json,
	// and a two-step path for Go 1.20 and older that uses go test -json and dist test (without JSON).
	hasDistTestJSON := spec.inputs.GoBranch != "release-branch.go1.20" && spec.inputs.GoBranch != "release-branch.go1.19"
	if !hasDistTestJSON {
		if shard != noSharding {
			return fmt.Errorf("test sharding is not supported for Go version 1.20 and earlier")
		}
		// TODO(when Go 1.20 stops being supported): Delete this path.
		//
		// To have structured all.bash output on 1.20/1.19 release branches without dist test -json,
		// we divide Go tests into two parts:
		//   - the large remaining set with structured output support (uploaded to ResultDB)
		//   - a small set of unstructured tests (this part is fully eliminated in Go 1.21!)
		// While maintaining the property that their union doesn't fall short of all.bash.
		jsonOnPart := spec.wrapTestCmd(spec.goCmd(ctx, gorootSrc, spec.goTestArgs("std", "cmd")...))
		if err := cmdStepRun(ctx, "run std and cmd tests", jsonOnPart, false); err != nil {
			return err
		}
		const allButStdCmd = "!^go_test:.+$" // Pattern that works in Go 1.20 and 1.19.
		jsonOffPart := spec.distTestCmd(ctx, gorootSrc, allButStdCmd, nil, false)
		return cmdStepRun(ctx, "run various dist tests", jsonOffPart, false)
	}
	// Go 1.21+ path.

	// Determine what to run.
	//
	// If noSharding is true, tests will be left as the empty slice, which means to
	// use dist test's default behavior of running all tests.
	var tests []string
	if shard != noSharding {
		// Collect the list of tests for this shard.
		var err error
		tests, err = goDistTestList(ctx, spec, shard)
		if err != nil {
			return err
		}
		if len(tests) == 0 {
			// No tests were selected to run. Explicitly return early instead
			// of needlessly calling dist test and telling it to run no tests.
			return nil
		}
	}

	// Invoke go tool dist test.
	testCmd := spec.wrapTestCmd(spec.distTestCmd(ctx, gorootSrc, "", tests, true))
	return cmdStepRun(ctx, "go tool dist test -json", testCmd, false)
}

func goDistTestList(ctx context.Context, spec *buildSpec, shard testShard) (tests []string, err error) {
	step, ctx := build.StartStep(ctx, "list tests")
	defer endStep(step, &err)

	// Run go tool dist test -list.
	listCmd := spec.distTestListCmd(ctx, spec.goroot)
	listOutput, err := cmdStepOutput(ctx, "go tool dist test -list", listCmd, false)
	if err != nil {
		return nil, err
	}

	// Parse the output—each line is a test name,
	// and select ones matching this shard.
	scanner := bufio.NewScanner(bytes.NewReader(listOutput))
	for scanner.Scan() {
		name := scanner.Text()
		if shard.shouldRunTest(name) {
			tests = append(tests, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parsing test list from dist: %v", err)
	}
	testList := strings.Join(tests, "\n")
	if len(tests) == 0 {
		testList = "(no tests selected)"
	}
	io.WriteString(step.Log("tests"), testList)
	return tests, nil
}

func fetchSubrepoAndRunTests(ctx context.Context, spec *buildSpec) error {
	if spec.inputs.Project == "go" {
		return infraErrorf("fetchSubrepoAndRunTests called for a main Go repo builder")
	}

	// Fetch the target repository.
	repoDir, err := os.MkdirTemp(spec.workdir, "targetrepo") // Use a non-predictable base directory name.
	if err != nil {
		return err
	}
	if err := fetchRepo(ctx, spec.subrepoSrc, repoDir, spec.inputs); err != nil {
		return err
	}

	// Test this specific subrepo.
	// If testing any one nested module fails, keep going and report all the end.
	modules, err := repoToModules(ctx, spec, repoDir)
	if err != nil {
		return err
	}
	if spec.noNetworkCapable && !spec.inputs.LongTest && spec.experiment("golang.no_network_in_short_test_mode") {
		// Fetch module dependencies ahead of time since 'go test' will not have network access.
		err := fetchDependencies(ctx, spec, modules)
		if err != nil {
			return err
		}
	}
	var testErrors []error
	for _, m := range modules {
		testCmd := spec.wrapTestCmd(spec.goCmd(ctx, m.RootDir, spec.goTestArgs("./...")...))
		if err := cmdStepRun(ctx, fmt.Sprintf("test %q module", m.Path), testCmd, false); err != nil {
			testErrors = append(testErrors, err)
		}
	}
	if len(testErrors) > 0 {
		return errors.Join(testErrors...)
	}
	return nil
}

// A module is a Go module located on disk.
type module struct {
	RootDir string // Module root directory on disk.
	Path    string // Module path specified in go.mod.
}

// repoToModules discovers and reports modules in repoDir to be tested.
func repoToModules(ctx context.Context, spec *buildSpec, repoDir string) (modules []module, err error) {
	step, ctx := build.StartStep(ctx, "discover modules")
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.

	// Discover all modules that we wish to test. See go.dev/issue/32528.
	if err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") || d.Name() == "testdata") {
			// Skip directories that we're not looking to support having testable modules in.
			return fs.SkipDir
		}
		if goModFile := d.Name() == "go.mod" && !d.IsDir(); goModFile {
			modPath, err := modPath(path)
			if err != nil {
				return err
			}
			modules = append(modules, module{
				RootDir: filepath.Dir(path),
				Path:    modPath,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	keepNestedModsInsideRepo := map[string]bool{
		"tools":     true, // A local replace directive in x/tools/gopls as of 2023-06-08.
		"telemetry": true, // A local replace directive in x/telemetry/godev as of 2023-06-08.
		"exp":       true, // A local replace directive in x/exp/slog/benchmarks/{zap,zerolog}_benchmarks as of 2023-06-08.
	}
	if !keepNestedModsInsideRepo[spec.inputs.Project] || spec.experiment("golang.force_test_outside_repository") {
		// Move nested modules to directories that aren't predictably-relative to each other
		// to catch accidental reads across nested module boundaries. See go.dev/issue/34352.
		//
		// Sort modules by increasing nested-ness, and do this
		// in reverse order for all but the first (root) module.
		slices.SortFunc(modules, func(a, b module) bool {
			return strings.Count(a.RootDir, string(filepath.Separator)) < strings.Count(b.RootDir, string(filepath.Separator))
		})
		for i := len(modules) - 1; i >= 1; i-- {
			randomDir, err := os.MkdirTemp(spec.workdir, "nestedmod")
			if err != nil {
				return nil, err
			}
			newDir := filepath.Join(randomDir, filepath.Base(randomDir)) // Use a non-predictable base directory name.
			if err := os.Rename(modules[i].RootDir, newDir); err != nil {
				return nil, err
			}
			modules[i].RootDir = newDir
		}
	}

	return modules, nil
}

// modPath reports the module path in the given go.mod file.
func modPath(goModFile string) (string, error) {
	b, err := os.ReadFile(goModFile)
	if err != nil {
		return "", err
	}
	f, err := modfile.ParseLax(goModFile, b, nil)
	if err != nil {
		return "", err
	} else if f.Module == nil {
		return "", fmt.Errorf("go.mod file %q has no module statement", goModFile)
	}
	return f.Module.Mod.Path, nil
}
