// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"infra/experimental/golangbuild/golangbuildpb"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	clouddatastore "cloud.google.com/go/datastore"
	"google.golang.org/api/option"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	"go.chromium.org/luci/common/errors"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/gae/impl/cloud"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/luciexe/build"
	"go.chromium.org/luci/luciexe/build/cv"
	sauth "go.chromium.org/luci/server/auth"
)

// buildSpec specifies what a single build will begin doing.
type buildSpec struct {
	auth *auth.Authenticator

	builderName string
	workdir     string
	goroot      string
	gopath      string
	gocacheDir  string
	toolsRoot   string
	casInstance string

	inputs *golangbuildpb.Inputs

	goSrc      *sourceSpec
	subrepoSrc *sourceSpec // nil if inputs.Project == "go"
	invokedSrc *sourceSpec // the commit/change we were invoked with

	invocation string // current ResultDB invocation

	experiments map[string]struct{}

	noNetworkCapable bool // whether the host OS can disable network access during test execution
}

func deriveBuildSpec(ctx context.Context, cwd, toolsRoot string, experiments map[string]struct{}, st *build.State, inputs *golangbuildpb.Inputs) (*buildSpec, error) {
	authOpts := chromeinfra.SetDefaultAuthOptions(auth.Options{
		Scopes: append([]string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gerritcodereview",
		}, sauth.CloudOAuthScopes...),
	})
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)

	// Build the sourceSpec we were invoked with.
	isDryRun := false
	if mode, err := cv.RunMode(ctx); err == nil {
		isDryRun = strings.HasSuffix(mode, "DRY_RUN")
	} else if err != cv.ErrNotActive {
		return nil, fmt.Errorf("cv.RunMode: %w", err)
	}
	gitilesCommit := st.Build().GetInput().GetGitilesCommit()
	var gerritChange *bbpb.GerritChange
	if changes := st.Build().GetInput().GetGerritChanges(); len(changes) > 1 {
		return nil, fmt.Errorf("no support for multiple GerritChanges")
	} else if len(changes) != 0 {
		gerritChange = changes[0]
	}
	var changedProject, changedBranch string
	switch {
	case gerritChange != nil && gitilesCommit != nil:
		return nil, fmt.Errorf("only a Gerrit change or a Gitiles commit is supported, not both")
	case gerritChange == nil && gitilesCommit != nil:
		if gitilesCommit.Host != goHost {
			return nil, fmt.Errorf("unsupported host %q, want %q", gitilesCommit.Host, goHost)
		}
		changedProject = gitilesCommit.Project
		changedBranch = refToBranch(gitilesCommit.Ref)
	case gerritChange != nil && gitilesCommit == nil:
		if gerritChange.Host != goReviewHost {
			return nil, fmt.Errorf("unsupported host %q, want %q", gerritChange.Host, goReviewHost)
		}
		changedProject = gerritChange.Project
		hc, err := authenticator.Client()
		if err != nil {
			return nil, fmt.Errorf("authenticator.Client: %w", err)
		}
		gc, err := gerrit.NewRESTClient(hc, gerritChange.Host, true)
		if err != nil {
			return nil, fmt.Errorf("gerrit.NewRESTClient: %w", err)
		}
		changeInfo, err := gc.GetChange(ctx, &gerritpb.GetChangeRequest{
			Number:  gerritChange.Change,
			Project: gerritChange.Project,
		})
		if err != nil {
			return nil, fmt.Errorf("gc.GetChange: %w", err)
		}
		changedBranch = changeInfo.Branch
	default:
		return nil, fmt.Errorf("no commit or change specified for build and test")
	}
	if inputs.Project != "go" && changedProject != inputs.Project && changedProject != "go" {
		// This case is something like a "build" commit for an "image" build, which
		// doesn't make any sense.
		return nil, fmt.Errorf("unexpected change and project pairing: %s vs. %s", changedProject, inputs.Project)
	}

	// Figure out what our Go and subrepo commits are, but retain
	// which one we were invoked with.
	invokedSrc := &sourceSpec{
		project: changedProject,
		branch:  changedBranch,
		commit:  gitilesCommit,
		change:  gerritChange,
		rebase:  !isDryRun && gerritChange != nil, // Rebase change onto branch if it's not a dry run.
	}
	var goSrc, subrepoSrc *sourceSpec
	var err error
	if invokedSrc.project == "go" {
		goSrc = invokedSrc
		if inputs.Project != "go" {
			// We're testing the tip of inputs.Project's main branch against a Go commit.
			subrepoSrc, err = sourceForBranch(ctx, authenticator, inputs.Project, mainBranch)
			if err != nil {
				return nil, fmt.Errorf("sourceForBranch: %w", err)
			}
		}
	} else {
		// We're testing the tip of inputs.GoBranch against a commit to inputs.Project.
		subrepoSrc = invokedSrc
		goSrc, err = sourceForBranch(ctx, authenticator, "go", inputs.GoBranch)
		if err != nil {
			return nil, fmt.Errorf("sourceForBranch: %w", err)
		}
	}

	// Validate BuildersToTrigger invariant.
	if inputs.Project != "go" && len(inputs.BuildersToTrigger) != 0 {
		return nil, fmt.Errorf("specified builders to trigger for unsupported project")
	}

	// Get the CAS instance.
	casInst, err := casInstanceFromEnv()
	if err != nil {
		return nil, fmt.Errorf("casInstanceFromEnv: %w", err)
	}

	var noNetworkCapable bool
	if _, err := exec.LookPath("unshare"); err == nil {
		noNetworkCapable = true
	}

	return &buildSpec{
		auth:        authenticator,
		builderName: st.Build().GetBuilder().GetBuilder(),
		workdir:     cwd,
		goroot:      filepath.Join(cwd, "goroot"),
		gopath:      filepath.Join(cwd, "gopath"),
		gocacheDir:  filepath.Join(cwd, "gocache"),
		toolsRoot:   toolsRoot,
		casInstance: casInst,
		inputs:      inputs,
		invocation:  st.Build().GetInfra().GetResultdb().GetInvocation(),
		goSrc:       goSrc,
		subrepoSrc:  subrepoSrc,
		invokedSrc:  invokedSrc,
		experiments: experiments,

		noNetworkCapable: noNetworkCapable,
	}, nil
}

func (b *buildSpec) setEnv(ctx context.Context) context.Context {
	env := environ.FromCtx(ctx)
	env.Load(b.inputs.Env)
	env.Set("GOROOT_BOOTSTRAP", filepath.Join(b.toolsRoot, "go_bootstrap"))
	env.Set("GOPATH", b.gopath) // Explicitly set to an empty per-build directory, to avoid reusing the implicit default one.
	env.Set("GOBIN", "")
	env.Set("GOROOT", "")           // Clear GOROOT because it's possible someone has one set locally, e.g. for luci-go development.
	env.Set("GOTOOLCHAIN", "local") // golangbuild scope includes selecting the exact Go toolchain version, so always use that local one.
	env.Set("GOCACHE", b.gocacheDir)
	env.Set("GO_BUILDER_NAME", b.builderName)
	if b.inputs.LongTest {
		env.Set("GO_TEST_SHORT", "0") // Tell 'dist test' to operate in longtest mode. See go.dev/issue/12508.
	}
	// Use our tools before the system tools. Notably, use raw Git rather than the Chromium wrapper.
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.toolsRoot, "bin"), os.PathListSeparator, env.Get("PATH")))

	if b.targetGOOS() != hostGOOS {
		env.Set("GOHOSTOS", hostGOOS)
	}
	if b.targetGOARCH() != hostGOARCH {
		env.Set("GOHOSTARCH", hostGOARCH)
	}
	if hostGOOS == "windows" {
		env.Set("GOBUILDEXIT", "1") // On Windows, emit exit codes from .bat scripts. See go.dev/issue/9799.
		// TODO(heschi): select gcc32 for GOARCH=i386
		env.Set("PATH", fmt.Sprintf("%v%c%v", env.Get("PATH"), os.PathListSeparator, filepath.Join(b.toolsRoot, "cc/windows/gcc64/bin")))
	}
	return env.SetInCtx(ctx)
}

// goTestArgs returns go command arguments that test the specified import path patterns.
func (b *buildSpec) goTestArgs(patterns ...string) []string {
	args := []string{"test", "-json"}
	if b.inputs.CompileOnly {
		hasGoIssue15513Fix := b.inputs.GoBranch != "release-branch.go1.20" && b.inputs.GoBranch != "release-branch.go1.19"
		if !hasGoIssue15513Fix { // TODO: Delete after 1.20 drops off.
			// In Go 1.20 and older, go test -c did not support multiple packages,
			// so use the next best thing of -run that matches no tests.
			return append(append(args, "-run=^$"), patterns...)
		}
		return append(append(args, "-c", "-o", os.DevNull), patterns...)
	}
	if !b.inputs.LongTest {
		args = append(args, "-short")
	}
	if b.inputs.LongTest && b.inputs.Project == "go" { // TODO(dmitshur): Delete after 1.20 drops off.
		const (
			goTestDefaultTimeout = 10 * time.Minute // Default value taken from Go 1.20.
			scale                = 5                // An approximation of GO_TEST_TIMEOUT_SCALE.
		)
		args = append(args, "-timeout="+(goTestDefaultTimeout*scale).String())
	}
	if b.inputs.RaceMode {
		args = append(args, "-race")
	}
	return append(args, patterns...)
}

// distTestListCmd returns an exec.Cmd for executing `go tool dist test -list`.
//
// dir is the directory to run the command from.
//
// It automatically applies additional dist flags based on the buildSpec (e.g. -race).
func (b *buildSpec) distTestListCmd(ctx context.Context, dir string) *exec.Cmd {
	args := []string{"tool", "dist", "test", "-list"}
	args = append(args, b.distTestFlags()...)
	return b.goCmd(ctx, dir, args...)
}

// distTestRunCmd returns an exec.Cmd for executing `go tool dist test -run`.
//
// dir is the directory to run the command from.
// run controls with dist tests are run, using dist test's interface for controlling
// which tests to run:
//
//	-run string
//	  	run only those tests matching the regular expression; empty means to run all.
//	  	Special exception: if the string begins with '!', the match is inverted.
//
// If json is true, passes the -json flag, producing `go test -json`-compatible output.
// Note: -json is not supported before Go 1.21.
//
// TODO(go.dev/issue/59990): Delete the json argument when it becomes always true.
//
// It automatically applies additional dist flags based on the buildSpec (e.g. -race).
func (b *buildSpec) distTestRunCmd(ctx context.Context, dir, run string, json bool) *exec.Cmd {
	args := []string{"tool", "dist", "test"}
	if json {
		args = append(args, "-json")
	}
	args = append(args, b.distTestFlags()...)
	args = append(args, "-run", run)
	return b.goCmd(ctx, dir, args...)
}

// distTestFlags returns just the flags that we should pass to `go tool dist test`
// based on the spec.
func (b *buildSpec) distTestFlags() []string {
	var args []string
	if b.inputs.CompileOnly {
		args = append(args, "-compile-only")
	}
	if b.inputs.LongTest {
		// dist test doesn't have a flag to control longtest mode,
		// so this is handled in buildSpec.setEnv instead of here.
	}
	if b.inputs.RaceMode {
		args = append(args, "-race")
	}
	return args
}

const cloudProject = "golang-ci-luci"

func (b *buildSpec) installDatastoreClient(ctx context.Context) (context.Context, error) {
	// TODO(mknyszek): Enable auth only when not in a fake build.
	ts, err := b.auth.TokenSource()
	if err != nil {
		return nil, errors.Annotate(err, "failed to initialize the token source").Err()
	}
	client, err := clouddatastore.NewClient(ctx, cloudProject, option.WithTokenSource(ts))
	if err != nil {
		return nil, errors.Annotate(err, "failed to instantiate the datastore client").Err()
	}
	cfg := &cloud.ConfigLite{
		ProjectID: cloudProject,
		DS:        client,
	}
	return cfg.Use(ctx), nil
}

func (b *buildSpec) targetGOOS() string {
	if goos, ok := b.inputs.Env["GOOS"]; ok {
		return goos
	}
	return hostGOOS
}

func (b *buildSpec) targetGOARCH() string {
	if goarch, ok := b.inputs.Env["GOARCH"]; ok {
		return goarch
	}
	return hostGOARCH
}

const (
	hostGOOS   = runtime.GOOS
	hostGOARCH = runtime.GOARCH
)

func (b *buildSpec) toolPath(tool string) string {
	return filepath.Join(b.toolsRoot, "bin", tool)
}

func (b *buildSpec) toolCmd(ctx context.Context, tool string, args ...string) *exec.Cmd {
	return command(ctx, b.toolPath(tool), args...)
}

// wrapTestCmd wraps cmd with 'rdb' and 'result_adapter' to send test results
// to ResultDB. cmd must be a test command that emits a JSON stream in the
// https://go.dev/cmd/test2json#hdr-Output_Format format.
//
// On Linux host OS, if long test mode is off, cmd is also prefixed with 'unshare'
// to disable network access and catch tests that forget to check testing.Short().
//
// It edits cmd in place but for convenience also returns cmd back to its caller.
func (b *buildSpec) wrapTestCmd(cmd *exec.Cmd) *exec.Cmd {
	if b.noNetworkCapable && !b.inputs.LongTest && b.experiment("golang.no_network_in_short_test_mode") {
		// Disable external network access for the test command.
		// Permit internal loopback access.
		cmd.Args = []string{
			"unshare", "-r", "-n", "--",
			"sh", "-c", "ip link set dev lo up && " + strings.Join(cmd.Args, " "),
		}
	}

	cmd.Path = b.toolPath("rdb")
	cmd.Args = append([]string{
		cmd.Path, "stream", "--",
		b.toolPath("result_adapter"), "go", "--",
	}, cmd.Args...)
	return cmd
}

func (b *buildSpec) goScriptCmd(ctx context.Context, script string) *exec.Cmd {
	dir := filepath.Join(b.goroot, "src")
	cmd := command(ctx, filepath.FromSlash("./"+script))
	cmd.Dir = dir
	return cmd
}

// goCmd creates a command for running 'go {args}' in dir.
func (b *buildSpec) goCmd(ctx context.Context, dir string, args ...string) *exec.Cmd {
	env := environ.FromCtx(ctx)
	// Ensure the go binary found in PATH is the same as the one we're about to execute.
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.goroot, "bin"), os.PathListSeparator, env.Get("PATH")))
	cmd := command(env.SetInCtx(ctx), filepath.Join(b.goroot, "bin", "go"), args...)
	cmd.Dir = dir
	return cmd
}

func (b *buildSpec) experiment(ex string) bool {
	_, ok := b.experiments[ex]
	return ok
}

func command(ctx context.Context, bin string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = environ.FromCtx(ctx).Sorted()
	return cmd
}

func refToBranch(ref string) string {
	const branchRefPrefix = "refs/heads/"
	if strings.HasPrefix(ref, branchRefPrefix) {
		return ref[len(branchRefPrefix):]
	}
	return ""
}
