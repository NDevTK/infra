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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"
	"golang.org/x/exp/slices"
	"golang.org/x/mod/modfile"

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
		return build.AttachStatus(errors.Annotate(err, "Get CWD").Err(), bbpb.Status_INFRA_FAILURE, nil)
	}

	spec, err := deriveBuildSpec(ctx, cwd, toolsRoot, experiments, st, inputs)
	if err != nil {
		return build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
	}

	// Set up environment.
	ctx = spec.setEnv(ctx)
	ctx, err = spec.installDatastoreClient(ctx)
	if err != nil {
		return err
	}

	// Get a built Go toolchain.
	if err := getGo(ctx, spec); err != nil {
		return err
	}

	if spec.inputs.Project == "go" {
		// Trigger downstream builders (subrepo builders) with the commit and/or Gerrit change we got.
		if len(spec.inputs.BuildersToTrigger) > 0 {
			if err := triggerBuilders(ctx, spec); err != nil {
				return err
			}
		}

		// Test Go.
		//
		// We have two paths, unfortunately: a simple one for Go 1.21+ that uses dist test -json,
		// and a two-step path for Go 1.20 and older that uses go test -json and dist test (without JSON).
		// TODO(when Go 1.20 stops being supported): Delete the latter path.
		//
		// TODO(mknyszek): Support sharding by running `go tool dist test -list` and/or `go list std cmd` and
		// triggering N test builders with a subset of those tests in their properties.
		// Pass the newly-built toolchain via CAS.
		gorootSrc := filepath.Join(spec.goroot, "src")
		hasDistTestJSON := spec.inputs.GoBranch != "release-branch.go1.20" && spec.inputs.GoBranch != "release-branch.go1.19"
		if hasDistTestJSON {
			testCmd := spec.wrapTestCmd(spec.goCmd(ctx, gorootSrc, spec.distTestArgs()...))
			if err := runCommandAsStep(ctx, "all"+scriptExt()+" -json", testCmd, false); err != nil {
				return err
			}
		} else {
			// To have structured all.bash output on 1.20/1.19 release branches without dist test -json,
			// we divide Go tests into two parts:
			//   - the large remaining set with structured output support (uploaded to ResultDB)
			//   - a small set of unstructured tests (this part is fully eliminated in Go 1.21!)
			// While maintaining the property that their union doesn't fall short of all.bash.
			jsonOnPart := spec.wrapTestCmd(spec.goCmd(ctx, gorootSrc, spec.goTestArgs("std", "cmd")...))
			if err := runCommandAsStep(ctx, "run std and cmd tests", jsonOnPart, false); err != nil {
				return err
			}
			const allButStdCmd = "!^go_test:.+$" // Pattern that works in Go 1.20 and 1.19.
			jsonOffPart := spec.goCmd(ctx, gorootSrc, spec.distTestNoJSONArgs(allButStdCmd)...)
			if err := runCommandAsStep(ctx, "run various dist tests", jsonOffPart, false); err != nil {
				return err
			}
		}
	} else {
		// Fetch the target repository.
		repoDir, err := os.MkdirTemp(cwd, "targetrepo") // Use a non-predictable base directory name.
		if err != nil {
			return err
		}
		if err := fetchRepo(ctx, spec.subrepoSrc, repoDir); err != nil {
			return err
		}

		// Test this specific subrepo.
		// If testing any one nested module fails, keep going and report all the end.
		modules, err := repoToModules(ctx, spec, cwd, repoDir)
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
			if err := runCommandAsStep(ctx, fmt.Sprintf("test %q module", m.Path), testCmd, false); err != nil {
				testErrors = append(testErrors, err)
			}
		}
		if len(testErrors) > 0 {
			return errors.Join(testErrors...)
		}
	}
	return nil
}

// A module is a Go module located on disk.
type module struct {
	RootDir string // Module root directory on disk.
	Path    string // Module path specified in go.mod.
}

// repoToModules discovers and reports modules in repoDir to be tested.
func repoToModules(ctx context.Context, spec *buildSpec, cwd, repoDir string) (modules []module, err error) {
	step, ctx := build.StartStep(ctx, "discover modules")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

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
			randomDir, err := os.MkdirTemp(cwd, "nestedmod")
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
var cipdDeps = `
infra/3pp/tools/git/${platform} version:2@2.39.2.chromium.11
@Subdir bin
infra/tools/bb/${platform} latest
infra/tools/rdb/${platform} latest
infra/tools/luci/cas/${platform} latest
infra/tools/result_adapter/${platform} latest
@Subdir go_bootstrap
infra/3pp/tools/go/${platform} version:2@1.19.3
@Subdir cc/${os=windows}
golang/third_party/llvm-mingw-msvcrt/${platform} latest
`

func installTools(ctx context.Context, inputs *golangbuildpb.Inputs, experiments map[string]struct{}) (toolsRoot string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

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
	if err := runCommandAsStep(ctx, "cipd ensure", cmd, true); err != nil {
		return "", err
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

func getGo(ctx context.Context, spec *buildSpec) (err error) {
	step, ctx := build.StartStep(ctx, "get go")
	defer step.End(err)

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

	// There was no prebuilt toolchain we could grab. Fetch Go and build it.

	// Fetch the main Go repository into goroot.
	if err := fetchRepo(ctx, spec.goSrc, spec.goroot); err != nil {
		return err
	}

	// Build Go.
	if err := runCommandAsStep(ctx, "make"+scriptExt(), spec.goScriptCmd(ctx, "make"+scriptExt()), false); err != nil {
		return err
	}

	// Upload to CAS.
	return uploadGoToCAS(ctx, spec, spec.goSrc, spec.goroot)
}

func triggerBuilders(ctx context.Context, spec *buildSpec) (err error) {
	step, ctx := build.StartStep(ctx, "trigger downstream builders")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

	// Scribble down the builders we're triggering.
	buildersLog := step.Log("builders")
	if _, err := io.WriteString(buildersLog, strings.Join(spec.inputs.BuildersToTrigger, "\n")+"\n"); err != nil {
		return err
	}

	// Figure out the arguments to bb.
	bbArgs := []string{"add"}
	if spec.invokedSrc.commit != nil {
		commit := spec.invokedSrc.commit
		bbArgs = append(bbArgs, "-commit", fmt.Sprintf("https://%s/%s/+/%s", commit.Host, commit.Project, commit.Id))
	}
	if spec.invokedSrc.change != nil {
		change := spec.invokedSrc.change
		bbArgs = append(bbArgs, "-cl", fmt.Sprintf("https://%s/c/%s/+/%d/%d", change.Host, change.Project, change.Change, change.Patchset))
	}
	bbArgs = append(bbArgs, spec.inputs.BuildersToTrigger...)

	return runCommandAsStep(ctx, "bb add", spec.toolCmd(ctx, "bb", bbArgs...), true)
}

// runCommandAsStep runs the provided command as a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func runCommandAsStep(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (err error) {
	step, ctx := build.StartStep(ctx, stepName)
	defer func() {
		if infra {
			// Any failure in this function is an infrastructure failure.
			err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		}
		step.End(err)
	}()

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
		return err
	}

	// Run the command.
	//
	// Combine output because it's annoying to pick one of stdout and stderr
	// in the UI and be wrong.
	output := step.Log("output")
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}
