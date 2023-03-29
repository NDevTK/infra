// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary golangbuild is a luciexe binary that builds and tests the code for the
// Go project. It supports building and testing go.googlesource.com/go as well as
// Go project subrepositories (e.g. go.googlesource.com/net) and on different branches.
//
// To build and run this locally end-to-end, follow these steps:
//
//	luci-auth login -scopes "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/gerritcodereview"
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
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/errors"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/luciexe/build"
	"go.chromium.org/luci/luciexe/build/cv"

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
	authOpts := chromeinfra.SetDefaultAuthOptions(auth.Options{
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gerritcodereview",
		},
	})
	httpClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts).Client()
	if err != nil {
		return err
	}
	log.Printf("auth created")

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx)
	if err != nil {
		return err
	}
	log.Printf("installed tools")

	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Annotate(err, "Get CWD").Err()
	}
	goroot := filepath.Join(cwd, "goroot")
	gocacheDir := filepath.Join(cwd, "gocache")

	// Set up environment.
	env := environ.FromCtx(ctx)
	env.Load(inputs.Env)
	env.Set("GOROOT_BOOTSTRAP", filepath.Join(toolsRoot, "go_bootstrap"))
	env.Set("GOPATH", filepath.Join(cwd, "gopath")) // Explicitly set to an empty per-build directory, to avoid reusing the implicit default one.
	env.Set("GOBIN", "")
	env.Set("GOROOT", "") // Clear GOROOT because it's likely someone has one set locally, e.g. for luci-go development.
	env.Set("GOCACHE", gocacheDir)
	env.Set("GO_BUILDER_NAME", st.Build().GetBuilder().GetBuilder()) // TODO(mknyszek): This is underspecified. We may need Project and Bucket.
	// Use our tools before the system tools. Notably, use raw Git rather than the Chromium wrapper.
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(toolsRoot, "bin"), os.PathListSeparator, env.Get("PATH")))

	if runtime.GOOS == "windows" {
		// TODO(heschi): select gcc32 for GOARCH=i386
		env.Set("PATH", fmt.Sprintf("%v%c%v", env.Get("PATH"), os.PathListSeparator, filepath.Join(toolsRoot, "cc/windows/gcc64/bin")))
	}
	ctx = env.SetInCtx(ctx)

	// Grab commit/change/presubmit state.
	isDryRun := false
	if mode, err := cv.RunMode(ctx); err == nil {
		isDryRun = strings.HasSuffix(mode, "DRY_RUN")
	} else if err != cv.ErrNotActive {
		return err
	}
	gitilesCommit := st.Build().GetInput().GetGitilesCommit()
	gerritChanges := st.Build().GetInput().GetGerritChanges()

	if inputs.Project == "go" {
		// Fetch the main Go repository into goroot.
		if err := fetchRepo(ctx, httpClient, "go", goroot, gitilesCommit, gerritChanges, isDryRun); err != nil {
			return err
		}

		// Build Go.
		//
		// TODO(mknyszek): Support cross-compile-only modes, perhaps by having CompileGOOS
		// and CompileGOARCH repeated fields in the input proto to identify what to build.
		if err := runGoScript(ctx, goroot, "make"+scriptExt()); err != nil {
			return err
		}
		// TODO(mknyszek): Upload the result of make.bash somewhere that downstream builders can find.

		// Trigger downstream builders (subrepo builders) with the commit and/or Gerrit change we got.
		if len(inputs.BuildersToTrigger) > 0 {
			err := triggerBuilders(ctx,
				filepath.Join(toolsRoot, "bin", "bb"),
				gitilesCommit,
				gerritChanges,
				inputs.BuildersToTrigger...,
			)
			if err != nil {
				return err
			}
		}

		// Test Go.
		//
		// TODO(mknyszek): Support sharding by running `go tool dist test -list` and
		// triggering N test builders with a subset of those tests in their properties.
		// Pass the newly-built toolchain via CAS.
		distTestArgs := []string{"tool", "dist", "test", "-no-rebuild"}
		if inputs.RaceMode {
			distTestArgs = append(distTestArgs, "-race")
		}
		if err := runGo(ctx, "go tool dist test", goroot, goroot, distTestArgs...); err != nil {
			return err
		}
	} else {
		if len(inputs.BuildersToTrigger) != 0 {
			return fmt.Errorf("specified builders to trigger for unsupported project")
		}

		// Fetch the main and target repositories.
		//
		// TODO(mknyszek): Grab a prebuilt Go toolchain from a shared location instead.
		if err := fetchGoRepoAtBranch(ctx, goroot, inputs.GoBranch); err != nil {
			return err
		}
		if err := fetchRepo(ctx, httpClient, inputs.Project, "targetrepo", gitilesCommit, gerritChanges, isDryRun); err != nil {
			return err
		}

		// Build (only) the Go toolchain to use.
		if err := runGoScript(ctx, goroot, "make"+scriptExt()); err != nil {
			return err
		}

		// Test this specific subrepo.
		tryResultAdapter := inputs.Project == "build"
		if tryResultAdapter {
			if err := runSubrepoTestsWithResultAdapter(ctx, goroot, "targetrepo", inputs.RaceMode,
				filepath.Join(toolsRoot, "bin", "rdb"),
				filepath.Join(toolsRoot, "bin", "result_adapter")); err != nil {
				return err
			}
		} else {
			if err := runSubrepoTests(ctx, goroot, "targetrepo", inputs.RaceMode); err != nil {
				return err
			}
		}
	}
	return nil
}

// scriptExt returns the extension to use for
// GOROOT/src/{make,all} scripts on this GOOS.
func scriptExt() string {
	switch runtime.GOOS {
	case "windows":
		return ".bat"
	case "plan9":
		return ".rc"
	default:
		return ".bash"
	}
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
infra/tools/result_adapter/${platform} latest
@Subdir go_bootstrap
infra/3pp/tools/go/${platform} version:2@1.19.3
@Subdir cc/${os=windows}
golang/third_party/llvm-mingw-msvcrt/${platform} latest
`

func installTools(ctx context.Context) (toolsRoot string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

	io.WriteString(step.Log("ensure file"), cipdDeps)

	toolsRoot, err = os.Getwd()
	if err != nil {
		return "", err
	}
	toolsRoot = filepath.Join(toolsRoot, "tools")

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

const (
	goHost       = "go.googlesource.com"
	goReviewHost = "go-review.googlesource.com"
	// N.B. Unfortunately Go still calls the main branch "master" due to technical issues.
	tipBranch = "master" // nocheck
)

func fetchRepo(ctx context.Context, hc *http.Client, project, dst string, commit *bbpb.GitilesCommit, changes []*bbpb.GerritChange, isDryRun bool) (err error) {
	step, ctx := build.StartStep(ctx, "fetch repo")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

	// Get the GerritChange.
	var change *bbpb.GerritChange
	if len(changes) > 1 {
		return fmt.Errorf("no support for multiple GerritChanges")
	} else if len(changes) != 0 {
		change = changes[0]
	}

	// Validate change and commit.
	if change != nil {
		if change.Host != goReviewHost {
			return fmt.Errorf("unsupported host %q, want %q", change.Host, goReviewHost)
		}
		if change.Project != project {
			return fmt.Errorf("subrepo tests do not support cross-project triggers for trybots: triggered by %q", project)
		}
	}
	if commit != nil {
		if commit.Host != goHost {
			return fmt.Errorf("unsupported host %q, want %q", commit.Host, goHost)
		}
		if commit.Project != project {
			if commit.Project != "go" {
				return fmt.Errorf("unsupported trigger project for subrepo tests: %s", commit.Project)
			}
			// Subrepo test triggered by a change from a different project. Fetch at HEAD
			// and download Go toolchain for this commit.
			return fmt.Errorf("subrepo tests unimplemented")
		}
	}
	switch {
	case change != nil && isDryRun:
		return fetchRepoChangeAsIs(ctx, hc, dst, change)
	case change != nil && !isDryRun:
		return fetchRepoChangeWithRebase(ctx, hc, dst, change)
	case commit != nil:
		if isDryRun {
			return fmt.Errorf("DRY_RUN is unexpectedly set in the commit case")
		}
		return fetchRepoAtCommit(ctx, hc, dst, commit)
	default:
		return fmt.Errorf("no commit or change specified for build and test")
	}
}

// fetchRepoChangeAsIs checks out a change to be tested as is, without rebasing.
func fetchRepoChangeAsIs(ctx context.Context, hc *http.Client, dst string, change *bbpb.GerritChange) error {
	// TODO(mknyszek): We're cloning tip here then fetching what we actually want because git doesn't
	// provide a good way to clone at a specific ref or commit. Is there a way to speed this up?
	// Maybe caching is sufficient?
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "https://"+change.Host+"/"+change.Project, dst); err != nil {
		return err
	}
	ref := fmt.Sprintf("refs/changes/%d/%d/%d", change.Change%100, change.Change, change.Patchset)
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
func fetchRepoChangeWithRebase(ctx context.Context, hc *http.Client, dst string, change *bbpb.GerritChange) error {
	// For submit, fetch HEAD for the branch this change is for, fetch the CL, and cherry-pick it.
	gc, err := gerrit.NewRESTClient(hc, change.Host, true)
	if err != nil {
		return err
	}
	changeInfo, err := gc.GetChange(ctx, &gerritpb.GetChangeRequest{
		Number:  change.Change,
		Project: change.Project,
	})
	if err != nil {
		return err
	}
	branch := changeInfo.Branch
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "-b", branch, "https://"+change.Host+"/"+change.Project, dst); err != nil {
		return err
	}
	ref := fmt.Sprintf("refs/changes/%d/%d/%d", change.Change%100, change.Change, change.Patchset)
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
func fetchRepoAtCommit(ctx context.Context, hc *http.Client, dst string, commit *bbpb.GitilesCommit) error {
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

// fetchGoRepoAtBranch checks out the head of the specified branch of the main Go repository.
func fetchGoRepoAtBranch(ctx context.Context, dst, branch string) error {
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "--depth", "1", "-b", branch, "https://"+goHost+"/go", dst); err != nil {
		return err
	}
	if branch == tipBranch {
		// Write a VERSION file when testing the main branch.
		// Release branches have a checked-in VERSION file, reuse it as is for now.
		if err := writeVersionFile(ctx, dst, "tip"); err != nil {
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
			io.WriteString(step.Log("close error"), r.Error())
		}
	}()
	_, err = io.WriteString(io.MultiWriter(contentsLog, f), data)
	return err
}

func runGit(ctx context.Context, stepName string, args ...string) (err error) {
	return runCommandAsStep(ctx, stepName, exec.CommandContext(ctx, "git", args...), true)
}

func runGoScript(ctx context.Context, goroot, script string) (err error) {
	dir := filepath.Join(goroot, "src")
	cmd := exec.CommandContext(ctx, filepath.FromSlash("./"+script))
	cmd.Dir = dir
	return runCommandAsStep(ctx, script, cmd, false)
}

// runGo runs the Go command from goroot in dir as a step.
func runGo(ctx context.Context, stepName, goroot, dir string, args ...string) error {
	// Add the Go binary we're about to execute to PATH. There are a whole bunch of cases
	// like `go tool dist test` where the expectation is that the invoked `go` binary is
	// the one found in PATH.
	env := environ.FromCtx(ctx)
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(goroot, "bin"), os.PathListSeparator, env.Get("PATH")))

	// Run the command.
	cmd := exec.CommandContext(ctx, filepath.Join(goroot, "bin", "go"), args...)
	cmd.Dir = dir
	cmd.Env = env.Sorted()
	return runCommandAsStep(ctx, stepName, cmd, false)
}

// runSubrepoTests runs tests for Go packages in the module at dir
// using the Go toolchain at goroot. It prints test results to stdout/stderr.
//
// TODO(dmitshur): For final version, don't forget to also test packages in nested modules.
// TODO(dmitshur): Improve coverage (at cost of setup complexity) by running tests outside their repositories. See go.dev/issue/34352.
func runSubrepoTests(ctx context.Context, goroot, dir string, race bool) error {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	args = append(args, "./...")
	return runGo(ctx, "go test", goroot, dir, args...)
}

// runSubrepoTestsWithResultAdapter runs tests for Go packages in the module at dir
// using the Go toolchain at goroot. It uses the provided result_adapter (go/result-sink#result-adapter)
// and rdb (go/result-sink#resultsink-on-ci) to stream test results to ResultSink.
func runSubrepoTestsWithResultAdapter(ctx context.Context, goroot, dir string, race bool, rdb, resultAdapter string) error {
	args := []string{rdb, "stream", "--",
		resultAdapter, "go", "--",
		filepath.Join(goroot, "bin", "go"), "test", "-json"}
	if race {
		args = append(args, "-race")
	}
	args = append(args, "./...")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	return runCommandAsStep(ctx, "go test -json [-race] ./... (with rdb+result_adapter)", cmd, false)
}

// triggerBuilders triggers builds for downstream builders using the same commit
// and/or changes. Note: commit or changes must be specified, but not both.
func triggerBuilders(ctx context.Context, bbPath string, commit *bbpb.GitilesCommit, changes []*bbpb.GerritChange, builders ...string) (err error) {
	step, ctx := build.StartStep(ctx, "trigger downstream builders")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		err = build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
		step.End(err)
	}()

	// Scribble down the builders we're triggering.
	buildersLog := step.Log("builders")
	if _, err := io.WriteString(buildersLog, strings.Join(builders, "\n")+"\n"); err != nil {
		return err
	}

	// Figure out the arguments to bb.
	bbArgs := []string{"add"}
	switch {
	case commit != nil && len(changes) == 0:
		bbArgs = append(bbArgs, "-commit", fmt.Sprintf("https://%s/%s/+/%s", commit.Host, commit.Project, commit.Id))
	case commit == nil && len(changes) != 0:
		if len(changes) > 1 {
			return fmt.Errorf("specifying more than one Gerrit change is unsupported")
		}
		change := changes[0]
		bbArgs = append(bbArgs, "-cl", fmt.Sprintf("https://%s/c/%s/+/%d/%d", change.Host, change.Project, change.Change, change.Patchset))
	case commit == nil && len(changes) == 0:
		return fmt.Errorf("no source information specified")
	default:
		return fmt.Errorf("specifying both a commit and a Gerrit change is unsupported")
	}
	bbArgs = append(bbArgs, builders...)

	// Run bb add.
	return runCommandAsStep(ctx, "bb add", exec.CommandContext(ctx, bbPath, bbArgs...), true)
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
	var fullCmd bytes.Buffer
	envs := environ.FromCtx(ctx).Sorted()
	for _, env := range envs {
		fullCmd.WriteString(env)
		fullCmd.WriteString(" ")
	}
	if cmd.Dir != "" {
		fullCmd.WriteString("PWD=")
		fullCmd.WriteString(cmd.Dir)
		fullCmd.WriteString(" ")
	}
	fullCmd.WriteString(cmd.String())
	io.Copy(step.Log("commands"), &fullCmd)

	// Run the command.
	stdout := step.Log("stdout")
	stderr := step.Log("stderr")
	cmd.Env = envs
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}

// Copied from go.googlesource.com/build/internal/untar/untar.go
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
