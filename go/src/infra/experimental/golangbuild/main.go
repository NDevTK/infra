// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary golangbuild is a luciexe binary that builds and tests the code for the
// Go project. It supports building and testing go.googlesource.com/go as well as
// Go project subrepositories (e.g. go.googlesource.com/net) and on different branches.
//
// To build and run this locally end-to-end, follow these steps:
//
//	luci-auth login -scopes "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/gerritcodereview https://www.googleapis.com/auth/cloud-platform"
//	cat > build.jsonpb <<EOF
//	{
//		"builder": {
//			"project": "go",
//			"bucket": "ci",
//			"builder": "linux-amd64"
//		},
//		"input": {
//			"properties": {
//				"project": "go"
//			},
//			"gitiles_commit": {
//				"host": "go.googlesource.com",
//				"project": "go",
//				"id": "27301e8247580e456e712a07d68890dc1e857000",
//				"ref": "refs/heads/master"
//			}
//		}
//	}
//	EOF
//	export SWARMING_SERVER=https://chromium-swarm.appspot.com
//	LUCIEXE_FAKEBUILD=./build.jsonpb golangbuild
//
// Modify `build.jsonpb` as needed in order to try different paths. The format of
// `build.jsonpb` is a JSON-encoded protobuf with schema `go.chromium.org/luci/buildbucket/proto.Build`.
// The input.properties field of this protobuf follows the `infra/experimental/golangbuildpb.Inputs`
// schema which represents input parameters that are specific to this luciexe, but may also contain
// namespaced properties that are injected by different services. For instance, CV uses the
// "$recipe_engine/cq" namespace.
//
// As an example, to try out a "try bot" path, try the following `build.jsonpb`:
//
//	{
//		"builder": {
//			"project": "go",
//			"bucket": "try",
//			"builder": "linux-amd64"
//		},
//		"input": {
//			"properties": {
//				"project": "go",
//				"$recipe_engine/cq": {
//					"active": true,
//					"runMode": "DRY_RUN"
//				}
//			},
//			"gerrit_changes": [
//				{
//					"host": "go-review.googlesource.com",
//					"project": "go",
//					"change": 460376,
//					"patchset": 1
//				}
//			]
//		}
//	}
//
// NOTE: by default, a luciexe fake build will discard the temporary directory created to run
// the build. If you'd like to retain the contents of the directory, specify a working directory
// to the golangbuild luciexe via the `--working-dir` flag. Be careful about where this working
// directory lives; particularly, make sure it isn't a subdirectory of a Go module a directory
// containing a go.mod file.
//
// ## Contributing
//
// To keep things organized and consistent, keep to the following guidelines:
//   - Only functions generate steps. Methods never generate steps.
//   - Keep step presentation and high-level ordering logic in main.go when possible.
//   - Steps should be function-scoped. Steps should be created at the start of a function
//     with the step end immediately deferred to function exit.
//
// ## Experiments
//
// When adding new functionality, consider putting it behind an experiment. Experiments are
// made available in the buildSpec and are propagated from the builder definition.
// Experiments in the builder definition are given a probability of being enabled on any given
// build, but always manifest in the current build as either "on" or "off."
// Experiments should have a name like "golang.my_specific_new_functionality" and should
// be checked for via spec.experiment("golang.my_specific_new_functionality").
//
// Experiments can be disabled at first (no work needs to be done on the builder definition),
// rolled out, and then tested in a real build environment via `led`
//
//	led get-build ... | \
//	led edit \
//	  -experiment golang.my_specific_new_functionality=true | \
//	led launch
//
// or via `bb add -ex "+golang.my_specific_new_functionality" ...`.
//
// Experiments can be enabled on LUCIEXE_FAKEBUILD runs through the "experiments" property (array
// of strings) on "input."
//
// ### Current experiments
//
//   - golang.cache_tools_root: Cache the cipd tool installation root across
//     builds and builders. If the tool versions remain the same across builds,
//     this allows `cipd ensure` to become a no-op on subsequent builds. This
//     requires a named cache defined on each builder, whose name is provided
//     via golangbuildpb.Inputs.ToolsCache.
//   - golang.force_test_outside_repository: Can be used to force running tests
//     from outside the repository to catch accidental reads outside of module
//     boundaries despite the repository not having opted-in to this test
//     behavior.
//   - golang.no_network_in_short_test_mode: Disable network access in -short
//     test mode. In the process of being gradually rolled out to all repos.
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"
	resultdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"
	"google.golang.org/protobuf/encoding/protojson"

	"infra/experimental/golangbuild/golangbuildpb"
)

func main() {
	inputs := new(golangbuildpb.Inputs)
	build.Main(inputs, nil, nil, func(ctx context.Context, args []string, st *build.State) error {
		return run(ctx, args, st, inputs)
	})
}

func run(ctx context.Context, args []string, st *build.State, inputs *golangbuildpb.Inputs) (err error) {
	log.Printf("run starting")

	// Collect enabled experiments.
	experiments := make(map[string]struct{})
	for _, ex := range st.Build().GetInput().GetExperiments() {
		experiments[ex] = struct{}{}
	}

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx, inputs, experiments)
	if err != nil {
		return err
	}
	log.Printf("installed tools")

	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return infraErrorf("get CWD")
	}

	spec, err := deriveBuildSpec(ctx, cwd, toolsRoot, experiments, st, inputs)
	if err != nil {
		return infraWrap(err)
	}

	// Set up environment.
	ctx = spec.setEnv(ctx)
	ctx, err = spec.installDatastoreClient(ctx)
	if err != nil {
		return err
	}

	// Select a runner based on the mode, then initialize and invoke it.
	var rn runner
	switch inputs.GetMode() {
	case golangbuildpb.Mode_MODE_ALL:
		rn = newAllRunner(inputs.GetAllMode())
	case golangbuildpb.Mode_MODE_COORDINATOR:
		rn = newCoordRunner(inputs.GetCoordMode())
	case golangbuildpb.Mode_MODE_BUILD:
		rn = newBuildRunner(inputs.GetBuildMode())
	case golangbuildpb.Mode_MODE_TEST:
		rn, err = newTestRunner(inputs.GetTestMode(), inputs.GetTestShard())
	}
	if err != nil {
		return infraErrorf("initializing runner: %v", err)
	}
	return rn.Run(ctx, spec)
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

// cipdDeps is an ensure file that describes all our CIPD dependencies.
//
// N.B. We assume a few tools are already available on the machine we're
// running on. Namely:
// - For non-Windows, a C/C++ toolchain
//
// TODO(mknyszek): Make sure Go 1.17 still works as the bootstrap toolchain since
// it's our published minimum.
const cipdDeps = `
infra/3pp/tools/git/${platform} version:2@2.39.2.chromium.11
@Subdir bin
infra/tools/bb/${platform} latest
infra/tools/rdb/${platform} latest
infra/tools/luci/cas/${platform} latest
infra/tools/result_adapter/${platform} latest
@Subdir go_bootstrap
infra/3pp/tools/go/${platform} version:2@BOOTSTRAP_VERSION
@Subdir cc/${os=windows}
golang/third_party/llvm-mingw-msvcrt/${platform} latest
@Subdir
`

func installTools(ctx context.Context, inputs *golangbuildpb.Inputs, experiments map[string]struct{}) (toolsRoot string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.

	bootstrap := inputs.BootstrapVersion
	if bootstrap == "" {
		bootstrap = "1.19.3"
	}
	cipdDeps := strings.ReplaceAll(cipdDeps, "BOOTSTRAP_VERSION", bootstrap)

	if inputs.XcodeVersion != "" {
		cipdDeps += "infra/tools/mac_toolchain/${platform} latest\n"
	}

	io.WriteString(step.Log("ensure file"), cipdDeps)

	if _, ok := experiments["golang.cache_tools_root"]; ok {
		// Store in a named cache. This is shared across builder types,
		// allowing reuse across builds if the dependencies versions
		// are the same.
		luciExe := lucictx.GetLUCIExe(ctx)
		if luciExe == nil {
			return "", fmt.Errorf("missing LUCI_CONTEXT")
		}

		cache := inputs.ToolsCache
		if cache == "" {
			return "", fmt.Errorf("inputs missing ToolsCache: %+v", inputs)
		}
		if !filepath.IsLocal(cache) {
			return "", fmt.Errorf("ToolsCache %q must be relative", cache)
		}

		toolsRoot = filepath.Join(luciExe.GetCacheDir(), cache)
	} else {
		// Store under CWD. This will be deleted after each build.
		toolsRoot, err = os.Getwd()
		if err != nil {
			return "", err
		}
		toolsRoot = filepath.Join(toolsRoot, "tools")
	}

	io.WriteString(step.Log("tools root"), toolsRoot)

	// Install packages.
	cmd := exec.CommandContext(ctx, "cipd",
		"ensure", "-root", toolsRoot, "-ensure-file", "-",
		"-json-output", filepath.Join(os.TempDir(), "ensure_results.json"))
	cmd.Stdin = strings.NewReader(cipdDeps)
	if err := cmdStepRun(ctx, "cipd ensure", cmd, true); err != nil {
		return "", err
	}

	// Set up XCode.
	// See https://source.corp.google.com/h/chromium/infra/infra/+/main:go/src/infra/cmd/mac_toolchain/README.md and
	// https://chromium.googlesource.com/chromium/tools/depot_tools/+/HEAD/recipes/recipe_modules/osx_sdk/api.py
	if inputs.XcodeVersion != "" {
		xcodeInstall := exec.CommandContext(ctx, filepath.Join(toolsRoot, "mac_toolchain"), "install", "-xcode-version", inputs.XcodeVersion, "-output-dir", filepath.Join(toolsRoot, "XCode.app"))
		if err := cmdStepRun(ctx, "install XCode "+inputs.XcodeVersion, xcodeInstall, true); err != nil {
			return "", err
		}
		xcodeSelect := exec.CommandContext(ctx, "sudo", "xcode-select", "--switch", filepath.Join(toolsRoot, "XCode.app"))
		if err := cmdStepRun(ctx, "select XCode "+inputs.XcodeVersion, xcodeSelect, true); err != nil {
			return "", err
		}
	}
	return toolsRoot, nil
}

// scriptExt returns the extension to use for
// GOROOT/src/{make,all} scripts on this GOOS.
func scriptExt() string {
	switch hostGOOS {
	case "windows":
		return ".bat"
	case "plan9":
		return ".rc"
	default:
		return ".bash"
	}
}

func getGo(ctx context.Context, spec *buildSpec, requirePrebuilt bool) (err error) {
	step, ctx := build.StartStep(ctx, "get go")
	defer endStep(step, &err)

	// Check to see if we might have a prebuilt Go in CAS.
	digest, err := checkForPrebuiltGo(ctx, spec.goSrc)
	if err != nil {
		return err
	}
	if digest != "" {
		// Try to fetch from CAS. Note this might fail if the digest is stale enough.
		ok, err := fetchGoFromCAS(ctx, spec, digest, spec.goroot)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	if requirePrebuilt {
		return infraErrorf("no prebuilt Go found, but this builder requires it")
	}

	// There was no prebuilt toolchain we could grab. Fetch Go and build it.

	// Fetch the main Go repository into goroot.
	if err := fetchRepo(ctx, spec.goSrc, spec.goroot); err != nil {
		return err
	}

	// Build Go.
	if err := cmdStepRun(ctx, "make"+scriptExt(), spec.goScriptCmd(ctx, "make"+scriptExt()), false); err != nil {
		return err
	}

	// Upload to CAS.
	return uploadGoToCAS(ctx, spec, spec.goSrc, spec.goroot)
}

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
		jsonOffPart := spec.distTestRunCmd(ctx, gorootSrc, allButStdCmd, false)
		return cmdStepRun(ctx, "run various dist tests", jsonOffPart, false)
	}
	// Go 1.21+ path.

	// Determine what to run.
	//
	// If noSharding is true, runRegexp will be the empty string, which actually means to
	// run *all* tests.
	runRegexp := ""
	if shard != noSharding {
		// Collect the list of tests for this shard.
		testList, err := goDistTestList(ctx, spec, shard)
		if err != nil {
			return err
		}
		if len(testList) == 0 {
			// Explicitly disable all tests. We can't leave runRegexp blank because that will
			// run all tests.
			runRegexp = "^$"
		} else {
			// Transform each test name into a regular expression matching only that name.
			for i := range testList {
				testList[i] = "^" + regexp.QuoteMeta(testList[i]) + "$"
			}
			// Construct a regexp string for the test list.
			runRegexp = strings.Join(testList, "|")
		}
	}

	// Invoke go tool dist test.
	testCmd := spec.wrapTestCmd(spec.distTestRunCmd(ctx, gorootSrc, runRegexp, true))
	return cmdStepRun(ctx, "go tool dist test -json", testCmd, false)
}

func goDistTestList(ctx context.Context, spec *buildSpec, shard testShard) ([]string, error) {
	// Run go tool dist test -list.
	listCmd := spec.distTestListCmd(ctx, spec.goroot)
	testList, err := cmdStepOutput(ctx, "go tool dist test -list", listCmd, false)
	if err != nil {
		return nil, err
	}

	// Parse the output: each line is a test name.
	var tests []string
	scanner := bufio.NewScanner(bytes.NewReader(testList))
	for scanner.Scan() {
		name := scanner.Text()
		if shard.shouldRunTest(name) {
			tests = append(tests, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parsing test list from dist: %v", err)
	}
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
	if err := fetchRepo(ctx, spec.subrepoSrc, repoDir); err != nil {
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

// triggerDownstreamBuilds triggers a single build for each of a bunch of builders,
// and does not wait on them to complete.
func triggerDownstreamBuilds(ctx context.Context, spec *buildSpec, builders ...string) (err error) {
	step, ctx := build.StartStep(ctx, "trigger downstream builds")
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.

	// Scribble down the builders we're triggering.
	buildersLog := step.Log("builders")
	if _, err := io.WriteString(buildersLog, strings.Join(builders, "\n")+"\n"); err != nil {
		return err
	}

	// Figure out the arguments to bb.
	bbArgs := []string{"add"}
	if spec.invokedSrc.commit != nil {
		bbArgs = append(bbArgs, "-commit", spec.invokedSrc.asURL())
	}
	if spec.invokedSrc.change != nil {
		bbArgs = append(bbArgs, "-cl", spec.invokedSrc.asURL())
	}
	bbArgs = append(bbArgs, builders...)

	return cmdStepRun(ctx, "bb add", spec.toolCmd(ctx, "bb", bbArgs...), true)
}

// ensurePrebuiltGoExists checks if a prebuilt Go exists for the invoked source, and if
// not, spawns a new build for the provided builder to generate that prebuilt Go.
func ensurePrebuiltGoExists(ctx context.Context, spec *buildSpec, builder string) (err error) {
	step, ctx := build.StartStep(ctx, "ensure prebuilt go exists")
	defer endStep(step, &err)

	// Check to see if we might have a prebuilt Go in CAS.
	digest, err := checkForPrebuiltGo(ctx, spec.goSrc)
	if err != nil {
		return err
	}
	if digest != "" {
		// Try to fetch from CAS. Note this might fail if the digest is stale enough.
		//
		// TODO(mknyszek): Rather than download the toolchain, it would be nice to check
		// this more directly.
		ok, err := fetchGoFromCAS(ctx, spec, digest, spec.goroot)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}

	// There was no prebuilt toolchain we could grab. Launch a build.
	//
	// N.B. We can theoretically just launch a build without checking, since the build
	// will already back out if it turns out there's a prebuilt Go already hanging around.
	// But it's worthwhile to check first because we don't have to wait to acquire the
	// resources for a build.
	build, err := triggerBuild(ctx, spec, noSharding, builder)
	if err != nil {
		return err
	}

	// Wait on build to finish.
	return waitOnBuilds(ctx, spec, "wait for make.bash", build.Id)
}

// runTestShards spawns `shards` builds from the provided `builder` and waits on them to complete.
//
// It passes the current build's source information to the child builds and includes the child builds'
// ResultDB invocations in the current invocation.
func runTestShards(ctx context.Context, spec *buildSpec, shards uint32, builder string) (err error) {
	step, ctx := build.StartStep(ctx, "run tests")
	defer endStep(step, &err)

	// Trigger test shards.
	buildIDs, err := triggerTestShards(ctx, spec, shards, builder)
	if err != nil {
		return err
	}

	// Wait on test shards to finish.
	return waitOnBuilds(ctx, spec, "wait for test shards", buildIDs...)
}

// triggerTestShards spawns `shards` builds from the provided `builder`.
//
// It passes the current build's source information to the child builds and includes the child builds'
// ResultDB invocations in the current invocation.
func triggerTestShards(ctx context.Context, spec *buildSpec, shards uint32, builder string) (ids []int64, err error) {
	step, ctx := build.StartStep(ctx, "trigger test shards")
	defer endStep(step, &err)

	// Start N shards and collect their build IDs and invocation IDs.
	buildIDs := make([]int64, 0, shards)
	invocationIDs := make([]string, 0, shards)
	for i := uint32(0); i < shards; i++ {
		shardBuild, err := triggerBuild(ctx, spec, testShard{shardID: i, nShards: shards}, builder)
		if err != nil {
			return nil, err
		}
		buildIDs = append(buildIDs, shardBuild.Id)
		invocationIDs = append(invocationIDs, shardBuild.GetInfra().GetResultdb().GetInvocation())
	}
	// Include the ResultDB invocations in ours.
	if err := includeResultDBInvocations(ctx, spec, invocationIDs...); err != nil {
		return nil, infraWrap(err)
	}
	return buildIDs, nil
}

// triggerBuild spawns a single build from the provided `builder`.
//
// If shard is not noSharding, then this function will pass the test shard identity
// as a set of properties to the build.
func triggerBuild(ctx context.Context, spec *buildSpec, shard testShard, builder string) (b *bbpb.Build, err error) {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("trigger %s (%d of %d)", builder, shard.shardID+1, shard.nShards))
	defer endStep(step, &err)

	// Construct args.
	prop, err := protojson.Marshal(&golangbuildpb.TestShard{
		ShardId:   shard.shardID,
		NumShards: shard.nShards,
	})
	if err != nil {
		return nil, infraErrorf("marshalling shard identity: %v", err)
	}
	bbArgs := []string{"add", "-json"}
	if shard != noSharding {
		bbArgs = append(bbArgs, "-p", fmt.Sprintf(`test_shard=%s`, string(prop)))
	}
	if spec.invokedSrc.commit != nil {
		bbArgs = append(bbArgs, "-commit", spec.invokedSrc.asURL())
	}
	if spec.invokedSrc.change != nil {
		bbArgs = append(bbArgs, "-cl", spec.invokedSrc.asURL())
	}
	bbArgs = append(bbArgs, builder)

	// Execute `bb add` for this shard and collect the output.
	stepName := fmt.Sprintf("bb add (%d of %d)", shard.shardID+1, shard.nShards)
	out, err := cmdStepOutput(ctx, stepName, spec.toolCmd(ctx, "bb", bbArgs...), true)
	if err != nil {
		return nil, err
	}

	// Unmarshal the output as a bbpb.Build.
	var build bbpb.Build
	if err := protojson.Unmarshal(out, &build); err != nil {
		return nil, infraWrap(err)
	}
	step.SetSummaryMarkdown(fmt.Sprintf(`[build link](https://ci.chromium.org/b/%d)`, build.Id))
	return &build, nil
}

// includeResultDBInvocations includes the provided ResultDB invocation IDs in the
// current invocation, as found in the buildSpec.
func includeResultDBInvocations(ctx context.Context, spec *buildSpec, ids ...string) (err error) {
	step, ctx := build.StartStep(ctx, "include ResultDB invocations")
	defer endStep(step, &err)

	// Set up the request and marshal it as JSON.
	updateReq := resultdbpb.UpdateIncludedInvocationsRequest{
		IncludingInvocation: spec.invocation,
		AddInvocations:      ids,
	}
	out, err := protojson.Marshal(&updateReq)
	if err != nil {
		return infraWrap(err)
	}

	// Write out the JSON as a log for debugging.
	reqLog := step.Log("request json")
	reqLog.Write(out)

	// Set up the `rdb rpc` command and pass the request through stdin.
	//
	// TODO(mknyszek): It's a bit silly to shell out for something that is
	// overtly just making an RPC call. However, there's currently no API
	// for pulling some of the ResultDB information out of LUCI_CONTEXT,
	// so we'd have to hard-code that and copy it here, or send a patch
	// to luci-go. The latter is preferable and should be considered as
	// part of a more general unit testing story for golangbuild.
	// For now, just shell out.
	cmd := spec.toolCmd(ctx, "rdb", "rpc", "-include-update-token", "luci.resultdb.v1.Recorder", "UpdateIncludedInvocations")
	cmd.Stdin = bytes.NewReader(out)
	return cmdStepRun(ctx, "rdb rpc", cmd, true)
}

// waitOnBuilds polls until the provided builds (by int64 ID) complete and
// reports a step that represents the result of those builds. Returns an error if
// any of the builds fail or if for some reason it fails to wait on the builds. The error
// returned by this function reflects the "worst" state of all builds. More specifically,
// an infra failure takes precedence over a regular test failure among the builds.
func waitOnBuilds(ctx context.Context, spec *buildSpec, stepName string, buildIDs ...int64) (err error) {
	step, ctx := build.StartStep(ctx, stepName)
	defer endStep(step, &err)

	// Run `bb collect`.
	collectArgs := []string{
		"collect",
		"-json",
		"-interval", "20s",
	}
	for _, id := range buildIDs {
		collectArgs = append(collectArgs, strconv.FormatInt(id, 10))
	}
	collectCmd := spec.toolCmd(ctx, "bb", collectArgs...)
	out, err := cmdStepOutput(ctx, "bb collect", collectCmd, true)
	if err != nil {
		return err
	}

	// Presentation state.
	var summary strings.Builder
	writeSummaryLine := func(shardID int, buildID int64, result string) {
		summary.WriteString(fmt.Sprintf("[shard %d %s](https://ci.chromium.org/b/%d)\n", shardID, result, buildID))
	}

	// Parse the protojson output: one per line.
	buildsBytes := bytes.Split(out, []byte{'\n'})
	var foundFailure, foundInfraFailure bool
	for i, buildBytes := range buildsBytes {
		build := new(bbpb.Build)
		if err := protojson.Unmarshal(buildBytes, build); err != nil {
			return infraWrap(err)
		}
		switch build.Status {
		case bbpb.Status_SUCCESS:
			// Tests passed. Nothing to do.
		case bbpb.Status_FAILURE:
			// Something was wrong with the change being tested.
			writeSummaryLine(i+1, build.Id, "failed")
			foundFailure = true
		case bbpb.Status_INFRA_FAILURE:
			// Something was wrong with the infrastructure.
			writeSummaryLine(i+1, build.Id, "infra-failed")
			foundInfraFailure = true
		case bbpb.Status_CANCELED:
			// Build got cancelled, which is very unexpected. Call it out.
			writeSummaryLine(i+1, build.Id, "cancelled")
			foundInfraFailure = true
		default:
			return infraErrorf("unexpected build status from `bb collect` for build %d: %s", build.Id, build.Status)
		}
	}
	step.SetSummaryMarkdown(summary.String())

	// Report an error for regular test failure or infra failure.
	if foundInfraFailure {
		return infraErrorf("one or more test shards experienced an infra failure")
	} else if foundFailure {
		return fmt.Errorf("one or more tests failed")
	}
	return nil
}

// cmdStepRun calls Run on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepRun(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			// Any failure in this function is an infrastructure failure.
			err = infraWrap(err)
		}
		step.End(err)
	}()
	if err != nil {
		return err
	}

	// Combine output because it's annoying to pick one of stdout and stderr
	// in the UI and be wrong.
	output := step.Log("output")
	cmd.Stdout = output
	cmd.Stderr = output

	// Run the command.
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}

// cmdStepOutput calls Output on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepOutput(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (output []byte, err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			// Any failure in this function is an infrastructure failure.
			err = infraWrap(err)
		}
		step.End(err)
	}()
	if err != nil {
		return nil, err
	}

	// Make sure we log stderr.
	cmd.Stderr = step.Log("stderr")

	// Run the command and capture stdout.
	output, err = cmd.Output()

	// Log stdout before we do anything else.
	step.Log("stdout").Write(output)

	// Check for errors.
	if err != nil {
		return output, errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return output, nil
}

// cmdStartStep sets up a command step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStartStep(ctx context.Context, stepName string, cmd *exec.Cmd) (*build.Step, context.Context, error) {
	step, ctx := build.StartStep(ctx, stepName)

	// Log the full command we're executing.
	//
	// Put each env var on its own line to actually make this readable.
	envs := cmd.Env
	if envs == nil {
		envs = os.Environ()
	}
	var fullCmd bytes.Buffer
	for _, env := range envs {
		fullCmd.WriteString(env)
		fullCmd.WriteString("\n")
	}
	if cmd.Dir != "" {
		fullCmd.WriteString("PWD=")
		fullCmd.WriteString(cmd.Dir)
		fullCmd.WriteString("\n")
	}
	fullCmd.WriteString(cmd.String())
	if _, err := io.Copy(step.Log("command"), &fullCmd); err != nil {
		return step, ctx, err
	}
	return step, ctx, nil
}

func infraErrorf(s string, args ...any) error {
	return build.AttachStatus(fmt.Errorf(s, args...), bbpb.Status_INFRA_FAILURE, nil)
}

func infraWrap(err error) error {
	return build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
}

func endStep(step *build.Step, errp *error) {
	step.End(*errp)
}

func endInfraStep(step *build.Step, errp *error) {
	step.End(infraWrap(*errp))
}
