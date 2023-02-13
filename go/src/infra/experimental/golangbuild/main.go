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

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx)
	if err != nil {
		return err
	}

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
	env.Set("GOCACHE", gocacheDir)
	env.Set("GO_BUILDER_NAME", st.Build().GetBuilder().GetBuilder()) // TODO(mknyszek): This is underspecified. We may need Project and Bucket.
	if runtime.GOOS == "windows" {
		// TODO(heschi): select gcc32 for GOARCH=i386
		env.Set("PATH", fmt.Sprintf("%v%v%v", env.Get("PATH"), os.PathListSeparator, filepath.Join(toolsRoot, "cc/windows/gcc64/bin")))
	}
	ctx = env.SetInCtx(ctx)

	inputPb := st.Build().GetInput()

	// Fetch the main Go repository into goroot.
	isDryRun := false
	if mode, err := cv.RunMode(ctx); err == nil {
		isDryRun = strings.HasSuffix(mode, "DRY_RUN")
	} else if err != cv.ErrNotActive {
		return err
	}
	if err := fetchRepo(ctx, httpClient, inputs.Project, goroot, inputPb.GetGitilesCommit(), inputPb.GetGerritChanges(), isDryRun); err != nil {
		return err
	}

	if inputs.Project == "go" {
		// Build and test Go.
		//
		// TODO(mknyszek): Support cross-compile-only modes, perhaps by having CompileGOOS
		// and CompileGOARCH repeated fields in the input proto to identify what to build.
		// TODO(mknyszek): Support split make/run and sharding.
		allScript := "all"
		if inputs.RaceMode {
			allScript = "race"
		}
		if err := runGoScript(ctx, goroot, allScript+scriptExt()); err != nil {
			return err
		}

		// Test the latest version of some subrepos.
		if err := runSubrepoTests(ctx, goroot); err != nil {
			return err
		}
	} else {
		// TODO(dmitshur): Build (only) the Go toolchain to use.
		// TODO(dmitshur): Test this specific subrepo.
		return fmt.Errorf("subrepository build/test is unimplemented")
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
// - git
// - For non-Windows, a C/C++ toolchain
//
// TODO(mknyszek): Make sure Go 1.17 still works as the bootstrap toolchain since
// it's our published minimum.
var cipdDeps = `
@Subdir go_bootstrap
infra/3pp/tools/go/${platform} version:2@1.19.3
@Subdir cc/${os=windows}
golang/third_party/llvm-mingw-msvcrt/${platform} latest
`

func installTools(ctx context.Context) (toolsRoot string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
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
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
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
		return fetchRepoForTry(ctx, hc, project, dst, change)
	case change != nil && !isDryRun:
		return fetchRepoForSubmit(ctx, hc, dst, change)
	case commit != nil:
		return fetchRepoForCI(ctx, hc, project, dst, commit)
	}
	// TODO(mknyszek): Fetch repo at HEAD here for subrepo tests.
	return fmt.Errorf("no commit or change specified for build and test")
}

func fetchRepoForTry(ctx context.Context, hc *http.Client, project, dst string, change *bbpb.GerritChange) (err error) {
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
	return writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset))
}

func fetchRepoForSubmit(ctx context.Context, hc *http.Client, dst string, change *bbpb.GerritChange) (err error) {
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
	return writeVersionFile(ctx, dst, fmt.Sprintf("%d/%d", change.Change, change.Patchset))
}

func fetchRepoForCI(ctx context.Context, hc *http.Client, project, dst string, commit *bbpb.GitilesCommit) (err error) {
	// TODO(mknyszek): This is a full git checkout, which is wasteful. Consider caching.
	if err := runGit(ctx, "git clone", "-C", ".", "clone", "https://"+commit.Host+"/"+commit.Project, dst); err != nil {
		return err
	}
	if err := runGit(ctx, "git checkout", "-C", dst, "checkout", commit.Id); err != nil {
		return err
	}
	return writeVersionFile(ctx, dst, commit.Id)
}

func writeVersionFile(ctx context.Context, dst, version string) error {
	return writeFile(ctx, filepath.Join(dst, "VERSION"), "devel "+version)
}

func writeFile(ctx context.Context, path, data string) (err error) {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("write %s", filepath.Base(path)))
	defer func() {
		// Any failure in this function is an infrastructure failure.
		step.End(build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil))
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

// runSubrepoTests tests the latest version of some subrepos
// using the Go toolchain at goroot.
func runSubrepoTests(ctx context.Context, goroot string) error {
	if err := os.Mkdir("subrepo", 0755); err != nil {
		return err
	}

	infra := true
	runGo := func(args ...string) error {
		cmd := exec.CommandContext(ctx, filepath.Join(goroot, "bin", "go"), args...)
		cmd.Dir = "subrepo"
		return runCommandAsStep(ctx, "step name", cmd, infra)
	}
	if err := runGo("mod", "init", "test"); err != nil {
		return err
	}
	// TODO(dmitshur): Think about the optimal general test strategy.
	if err := writeFile(ctx, filepath.Join("subrepo", "test.go"), `//go:build test

package p

import (
	_ "golang.org/x/mod/zip"
	_ "golang.org/x/term"
)
`); err != nil {
		return err
	}
	if err := runGo("mod", "tidy"); err != nil {
		return err
	}
	if err := runGo("get", "-t", "golang.org/x/mod/zip"); err != nil {
		return err
	}

	infra = false
	if err := runGo("test", "-json", "golang.org/x/mod/...", "golang.org/x/term/..."); err != nil {
		return err
	}

	return nil
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
