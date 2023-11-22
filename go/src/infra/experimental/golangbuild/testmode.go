// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"infra/experimental/golangbuild/golangbuildpb"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.chromium.org/luci/luciexe/build"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
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
	// Determine what ports to test.
	ports := []*golangbuildpb.Port{spec.inputs.Target}
	if spec.inputs.MiscPorts {
		// Note: There may be code changes in cmd/dist or cmd/go that have not
		// been fully reviewed yet, and it is a test error if goDistList fails.
		var err error
		ports, err = goDistList(ctx, spec, r.shard)
		if err != nil {
			return err
		}
	}
	// Run tests. (Also fetch dependencies if applicable.)
	if spec.inputs.Project == "go" {
		return runGoTests(ctx, spec, r.shard, ports)
	}
	return fetchSubrepoAndRunTests(ctx, spec, ports)
}

// testShard is a test shard identity that can be used to deterministically filter tests.
type testShard struct {
	shardID uint32 // The ID, in the range [0, nShards-1].
	nShards uint32 // Total number of shards (at least 1).
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

func runGoTests(ctx context.Context, spec *buildSpec, shard testShard, ports []*golangbuildpb.Port) (err error) {
	step, ctx := build.StartStep(ctx, "run tests")
	defer endStep(step, &err)

	if spec.inputs.Project != "go" {
		return infraErrorf("runGoTests called for a subrepo builder")
	}
	gorootSrc := filepath.Join(spec.goroot, "src")

	hasImprovedDistTestCompileOnly := spec.inputs.GoBranch != "release-branch.go1.20" && spec.inputs.GoBranch != "release-branch.go1.19"
	if spec.inputs.CompileOnly {
		// If compiling any one port fails, keep going and report all at the end.
		g := new(errgroup.Group)
		g.SetLimit(runtime.NumCPU())
		var testErrors = make([]error, len(ports))
		for i, p := range ports {
			i, p := i, p
			portContext := addPortEnv(ctx, p, "GOMAXPROCS="+fmt.Sprint(max(1, runtime.NumCPU()/len(ports))))
			testCmd := spec.wrapTestCmd(spec.distTestCmd(portContext, gorootSrc, "", nil, true))
			if !hasImprovedDistTestCompileOnly {
				// TODO(when Go 1.20 stops being supported): Delete this non-'dist test' path.
				testCmd = spec.wrapTestCmd(spec.goCmd(portContext, gorootSrc, spec.goTestArgs("std", "cmd")...))
			}
			g.Go(func() error {
				testErrors[i] = cmdStepRun(portContext, fmt.Sprintf("compile %s port", p), testCmd, false)
				return nil
			})
		}
		g.Wait()
		return errors.Join(testErrors...)
	}

	if len(ports) != 1 || !proto.Equal(ports[0], spec.inputs.Target) {
		return infraErrorf("testing multiple ports is only supported in CompileOnly mode")
	}

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

	// Determine what tests to run.
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
		return nil, fmt.Errorf("parsing test list from dist: %w", err)
	}
	testList := strings.Join(tests, "\n")
	if len(tests) == 0 {
		testList = "(no tests selected)"
	}
	io.WriteString(step.Log("tests"), testList)
	return tests, nil
}

func fetchSubrepoAndRunTests(ctx context.Context, spec *buildSpec, ports []*golangbuildpb.Port) (err error) {
	step, ctx := build.StartStep(ctx, "run tests")
	defer endStep(step, &err)

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
	// If testing any one nested module or port fails, keep going and report all at the end.
	modules, err := repoToModules(ctx, spec, repoDir)
	if err != nil {
		return err
	}
	// Fetch module dependencies ahead of time, to mark temporary network errors as an infra
	// failures and because 'go test' may not have network access (see spec.inputs.NoNetwork).
	if err := fetchDependencies(ctx, spec, modules); err != nil {
		return err
	}
	if spec.inputs.CompileOnly {
		return compileTestsInParallel(ctx, spec, modules, ports)
	} else if len(ports) != 1 || !proto.Equal(ports[0], spec.inputs.Target) {
		return infraErrorf("testing multiple ports is only supported in CompileOnly mode")
	}
	var testErrors []error
	for _, m := range modules {
		testCmd := spec.wrapTestCmd(spec.goCmd(ctx, m.RootDir, spec.goTestArgs("./...")...))
		if err := cmdStepRun(ctx, fmt.Sprintf("test %s module", m.Path), testCmd, false); err != nil {
			testErrors = append(testErrors, err)
		}
	}
	return errors.Join(testErrors...)
}

func compileTestsInParallel(ctx context.Context, spec *buildSpec, modules []module, ports []*golangbuildpb.Port) error {
	g := new(errgroup.Group)
	g.SetLimit(runtime.NumCPU())
	var testErrors = make([]error, len(ports)*len(modules))
	for i, p := range ports {
		i := i
		portContext := addPortEnv(ctx, p, "GOMAXPROCS="+fmt.Sprint(max(1, runtime.NumCPU()/(len(ports)*len(modules)))))
		for _, m := range modules {
			stepName := fmt.Sprintf("test %s module", m.Path)
			if len(ports) > 1 || !proto.Equal(p, spec.inputs.Target) {
				stepName += fmt.Sprintf(" for %s", p)
			}
			testCmd := spec.wrapTestCmd(spec.goCmd(portContext, m.RootDir, spec.goTestArgs("./...")...))
			if spec.inputs.CompileOnly && compileOptOut(spec.inputs.Project, p, m.Path) {
				stepName += " (skipped)"
				testCmd = command(portContext, "echo", "(skipped)")
			}
			g.Go(func() error {
				testErrors[i] = cmdStepRun(portContext, stepName, testCmd, false)
				return nil
			})
		}
	}
	g.Wait()
	return errors.Join(testErrors...)
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

// goDistList uses 'go tool dist list' to get a list of all non-broken ports,
// excluding ones that definitely already have a pre-submit builder,
// and returns those that match the provided shard.
func goDistList(ctx context.Context, spec *buildSpec, shard testShard) (ports []*golangbuildpb.Port, err error) {
	step, ctx := build.StartStep(ctx, "list ports")
	defer endStep(step, &err)

	// Run go tool dist list -json.
	//
	// Notably, we leave out -broken flag to get only non-broken ports.
	listCmd := spec.distListCmd(ctx, spec.goroot)
	listOutput, err := cmdStepOutput(ctx, "go tool dist list -json", listCmd, false)
	if err != nil {
		return nil, err
	}

	// Parse the JSON output and collect available ports.
	var allPorts []struct {
		GOOS, GOARCH string
		FirstClass   bool
	}
	err = json.Unmarshal(listOutput, &allPorts)
	if err != nil {
		return nil, fmt.Errorf("parsing port list from dist: %w", err)
	}
	for _, p := range allPorts {
		if p.GOOS == "" || p.GOARCH == "" {
			return nil, fmt.Errorf("go tool dist list returned an invalid GOOS/GOARCH pair: %#v", p)
		}
		switch {
		case p.FirstClass && p.GOOS != "darwin":
			// There's enough machine capacity and speed for almost
			// all first-class ports to have a pre-submit builder,
			// and there's not enough benefit to include them here.
			continue
		case p.GOOS == "ios" && p.GOARCH == "arm64":
			// TODO(go.dev/issue/61761): Add misc-compile coverage for the ios/arm64 port (iOS).
			continue
		case p.GOOS == "ios" && p.GOARCH == "amd64":
			// TODO(go.dev/issue/61760): Add misc-compile coverage for the ios/amd64 port (iOS Simulator).
			continue
		case p.GOOS == "android":
			// TODO(go.dev/issue/61762): Add misc-compile coverage for the GOOS=android ports (Android).
			continue
		case spec.inputs.GoBranch == "release-branch.go1.20" && p.GOOS == "openbsd" && p.GOARCH == "mips64":
			// The openbsd/mips64 port is marked broken at tip as of 2023-08-10.
			// It's not marked as broken in cmd/dist on release-branch.go1.20,
			// but it still fails to compile a number of golang.org/x repos.
			// So treat it as a broken port for our purposes since its negative signal
			// is not actionable until known issues with the port at tip are resolved.
			//
			// TODO(go.dev/issue/61546, go.dev/issue/58110): If the port gets fixed, drop this case.
			continue
		}
		ports = append(ports, &golangbuildpb.Port{Goos: p.GOOS, Goarch: p.GOARCH})
	}
	// Split up the ports into buckets, and pick one for this shard.
	bucketSize := len(ports) / int(shard.nShards)
	if bucketSize*int(shard.nShards) < len(ports) {
		// Round up when the number of ports doesn't divide evenly by shard count.
		bucketSize++
	}
	i := min(bucketSize*int(shard.shardID), len(ports))
	j := min(bucketSize*int(shard.shardID+1), len(ports))
	ports = ports[i:j]

	portList := fmt.Sprint(ports)
	if len(ports) == 0 {
		portList = "(no ports selected)"
	}
	io.WriteString(step.Log("ports"), portList)
	return ports, nil
}

// compileOptOut is a policy function that reports whether the provided
// port and module pair is considered opted out of compile-only testing.
//
// TODO(dmitshur,heschi): Ideally we want to have policy configured in
// one place, more likely in main.star than here. If so, factor it out.
func compileOptOut(project string, p *golangbuildpb.Port, modulePath string) bool {
	const (
		optOut                           = true
		performCompileOnlyTestingAsUsual = false // Long name so that it stands out. It's the rare case here.
	)
	ps := p.Goos + "-" + p.Goarch
	switch project {
	case "benchmarks":
		if p.Goos == "plan9" {
			// Dependency "github.com/coreos/go-systemd/v22/journal" fails to build on Plan 9.
			return optOut
		}
		if p.Goarch == "wasm" {
			// Dependencies "github.com/blevesearch/mmap-go", "go.etcd.io/bbolt", and "github.com/coreos/go-systemd/v22/journal"
			// fail to build. Also "golang.org/x/benchmarks/driver" fails to build.
			return optOut
		}
	case "build":
		// build is a special repository for internal Go build infrastructure needs.
		// It relies only on real pre- and post-submit testing, not compile-only testing.
		if p.Goos == "darwin" {
			// Except darwin, which doesn't yet have pre-submit coverage,
			// so use compile-only coverage to help out.
			return performCompileOnlyTestingAsUsual
		}
		return optOut
	case "debug":
		if p.Goarch == "wasm" {
			// Dependency "github.com/chzyer/readline" fails to build.
			return optOut
		}
	case "exp":
		switch modulePath {
		case "golang.org/x/exp/event":
			if ps == "wasip1-wasm" {
				// Dependency "github.com/sirupsen/logrus" fails to build on wasip1/wasm.
				return optOut
			}
		case "golang.org/x/exp/shiny":
			switch ps {
			case "darwin-arm64", "darwin-amd64", "linux-mips64", "linux-mips64le",
				"linux-ppc64", "linux-ppc64le", "linux-s390x", "openbsd-amd64":
				return performCompileOnlyTestingAsUsual
			default:
				// x/exp/shiny fails to build on most cross-compile platforms, largely because
				// of x/mobile dependencies.
				return optOut
			}
		}
	case "mobile":
		// mobile fails to build on all cross-compile platforms. This is somewhat expected
		// given the nature of the repository. Leave this as a blanket policy for now.
		return optOut
	case "pkgsite":
		// See go.dev/issue/61341.
		if p.Goos == "plan9" {
			// Dependency "github.com/lib/pq" fails to build on Plan 9.
			return optOut
		}
		if ps == "wasip1-wasm" {
			// Dependency "github.com/lib/pq" fails to build on wasip1/wasm.
			return optOut
		}
	case "pkgsite-metrics":
		if ps == "wasip1-wasm" {
			// Dependency "github.com/lib/pq" fails to build on wasip1/wasm.
			return optOut
		}
		if ps == "aix-ppc64" || p.Goos == "plan9" {
			// Dependency "github.com/apache/thrift/lib/go/thrift" fails to build on aix/ppc64 and Plan 9.
			return optOut
		}
	case "vuln":
		if p.Goos == "plan9" {
			// Dependency "github.com/google/go-cmdtest" fails to build on Plan 9.
			return optOut
		}
	case "vulndb":
		if ps == "aix-ppc64" {
			// Dependency "github.com/go-git/go-billy/v5/osfs" fails to build on aix/ppc64.
			// See go.dev/issue/58308.
			return optOut
		}
		if ps == "wasip1-wasm" {
			// Dependency "github.com/go-git/go-billy/v5/osfs" fails to build on wasip1/wasm.
			return optOut
		}
	}
	// The default policy decision is not to opt out.
	return performCompileOnlyTestingAsUsual
}

func min(x, y int) int { // TODO: Drop once go.mod's version is 1.21 or newer.
	if x < y {
		return x
	}
	return y
}
func max(x, y int) int { // TODO: Drop once go.mod's version is 1.21 or newer.
	if x > y {
		return x
	}
	return y
}
