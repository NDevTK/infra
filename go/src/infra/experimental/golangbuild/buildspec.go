// Copyright 2023 The Chromium Authors
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
	"strings"
	"time"

	clouddatastore "cloud.google.com/go/datastore"
	"google.golang.org/api/option"

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	lucierrors "go.chromium.org/luci/common/errors"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/gae/impl/cloud"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/luciexe/build"
	sauth "go.chromium.org/luci/server/auth"
)

// buildSpec specifies what a single build will begin doing.
type buildSpec struct {
	auth *auth.Authenticator

	builderName        string
	workdir            string
	goroot             string
	gopath             string
	gocacheDir         string
	toolsRoot          string
	casInstance        string
	priority           int32
	golangbuildVersion string

	inputs *golangbuildpb.Inputs

	goSrc      *sourceSpec // the Go repo spec
	subrepoSrc *sourceSpec // the x/ repo spec, or nil if inputs.Project == "go"
	invokedSrc *sourceSpec // the commit/change we were invoked with

	invocation string // current ResultDB invocation

	experiments map[string]struct{}
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
		changedProject = gitilesCommit.Project
		changedBranch = refToBranch(gitilesCommit.Ref)
	case gerritChange != nil && gitilesCommit == nil:
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
	if !isGoProject(inputs.Project) && changedProject != inputs.Project && !isGoProject(changedProject) {
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
		// TODO(mknyszek): Cherry-pick and/or rebase change(s) onto branch if it's
		// not a dry-run. Currently all we have are dry-runs, and we need to be
		// careful that child builds (build mode and test mode) also understand that
		// it's a dry run. Passing that information would be complicated and hacky
		// at present, but in the future go.chromium.org/luci/luciexe/build/cv could
		// be extended to expose the full CV proto, which we should then pass onto
		// child builds before changing this.
		cherryPick: false,
	}

	// For now, the internal repository is only supported for invokedSrc.
	// TODO: In the future, we will probably have subrepos in the security repository.
	// Which version of Go will we want to test against in that case?
	const publicGoHost = "go.googlesource.com"

	var goSrc, subrepoSrc *sourceSpec
	var err error
	if isGoProject(invokedSrc.project) {
		goSrc = invokedSrc
		if !isGoProject(inputs.Project) {
			// We're testing the tip of inputs.Project's main branch against a Go commit.
			subrepoSrc, err = sourceForBranch(ctx, authenticator, publicGoHost, inputs.Project, mainBranch)
			if err != nil {
				return nil, fmt.Errorf("sourceForBranch: %w", err)
			}
		}
		if inputs.GoCommit != "" {
			return nil, fmt.Errorf("GoCommit can be set only when invoked in a project other than 'go'")
		}
	} else {
		subrepoSrc = invokedSrc
		if inputs.GoCommit == "" {
			// We're testing the tip of inputs.GoBranch against a commit to inputs.Project.
			goSrc, err = sourceForBranch(ctx, authenticator, publicGoHost, "go", inputs.GoBranch)
			if err != nil {
				return nil, fmt.Errorf("sourceForBranch: %w", err)
			}
		} else {
			// We're testing a commit on inputs.GoBranch that was already selected for us.
			goSrc, err = sourceForGoBranchAndCommit(publicGoHost, inputs.GoBranch, inputs.GoCommit)
			if err != nil {
				return nil, fmt.Errorf("sourceForGoBranchAndCommit: %w", err)
			}
		}
	}

	// Validate BuildersToTriggerAfterToolchainBuild invariant.
	if !isGoProject(inputs.Project) && len(inputs.GetCoordMode().GetBuildersToTriggerAfterToolchainBuild()) != 0 {
		return nil, fmt.Errorf("specified builders to trigger for unsupported project")
	}

	// Get the CAS instance.
	casInst, err := casInstanceFromEnv()
	if err != nil {
		return nil, fmt.Errorf("casInstanceFromEnv: %w", err)
	}

	if inputs.NoNetwork {
		// Return a helpful error if the system requirements
		// for disabling external network are unmet.
		if inputs.Host.Goos != "linux" {
			return nil, fmt.Errorf("NoNetwork is not supported on %q", inputs.Host.Goos)
		}
		for _, cmd := range [...]string{"unshare", "ip", "sh"} {
			if _, err := exec.LookPath(cmd); err != nil {
				return nil, fmt.Errorf("NoNetwork needs %s in $PATH: %w", cmd, err)
			}
		}
	}

	return &buildSpec{
		auth:               authenticator,
		builderName:        st.Build().GetBuilder().GetBuilder(),
		workdir:            cwd,
		goroot:             filepath.Join(cwd, "goroot"),
		gopath:             filepath.Join(cwd, "gopath"),
		gocacheDir:         filepath.Join(cwd, "gocache"),
		toolsRoot:          toolsRoot,
		casInstance:        casInst,
		priority:           st.Build().GetInfra().GetSwarming().GetPriority(),
		golangbuildVersion: st.Build().GetExe().GetCipdVersion(),
		inputs:             inputs,
		invocation:         st.Build().GetInfra().GetResultdb().GetInvocation(),
		goSrc:              goSrc,
		subrepoSrc:         subrepoSrc,
		invokedSrc:         invokedSrc,
		experiments:        experiments,
	}, nil
}

func (b *buildSpec) setEnv(ctx context.Context) context.Context {
	env := environ.FromCtx(ctx)
	env.Load(b.inputs.Env)
	env.Set("GOOS", b.inputs.Target.Goos)
	env.Set("GOARCH", b.inputs.Target.Goarch)
	env.Set("GOHOSTOS", b.inputs.Host.Goos)
	env.Set("GOHOSTARCH", b.inputs.Host.Goarch)
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

	if b.inputs.Host.Goos == "windows" {
		env.Set("GOBUILDEXIT", "1") // On Windows, emit exit codes from .bat scripts. See go.dev/issue/9799.
		ccPath := filepath.Join(b.toolsRoot, "cc/windows/gcc64/bin")
		if b.inputs.Target.Goarch == "386" { // Not obvious whether this should check host or target. As of writing they never differ.
			ccPath = filepath.Join(b.toolsRoot, "cc/windows/gcc32/bin")
		}
		env.Set("PATH", fmt.Sprintf("%v%c%v", env.Get("PATH"), os.PathListSeparator, ccPath))
	}
	if b.inputs.Target.Goarch == "wasm" {
		// Add go_*_wasm_exec and the appropriate Wasm runtime to PATH.
		env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.goroot, "misc/wasm"), os.PathListSeparator, env.Get("PATH")))
		switch {
		case b.inputs.Target.Goos == "js":
			env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.toolsRoot, "nodejs/bin"), os.PathListSeparator, env.Get("PATH")))
		case b.inputs.Target.Goos == "wasip1" && env.Get("GOWASIRUNTIME") == "wasmtime",
			b.inputs.Target.Goos == "wasip1" && env.Get("GOWASIRUNTIME") == "wazero":
			env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.toolsRoot, env.Get("GOWASIRUNTIME")), os.PathListSeparator, env.Get("PATH")))
		}
	}

	return env.SetInCtx(ctx)
}

func addEnv(ctx context.Context, add ...string) context.Context {
	env := environ.FromCtx(ctx)
	for _, e := range add {
		env.SetEntry(e)
	}
	return env.SetInCtx(ctx)
}

func addPortEnv(ctx context.Context, target *golangbuildpb.Port, extraEnv ...string) context.Context {
	return addEnv(ctx, append(extraEnv, "GOOS="+target.Goos, "GOARCH="+target.Goarch)...)
}

// goTestArgs returns go command arguments that test the specified import path patterns.
func (b *buildSpec) goTestArgs(patterns ...string) []string {
	args := []string{"test", "-json"}
	if b.inputs.CompileOnly {
		hasGoIssue15513Fix := b.inputs.GoBranch != "release-branch.go1.20" && b.inputs.GoBranch != "release-branch.go1.19"
		if !hasGoIssue15513Fix { // TODO: Delete after 1.20 drops off.
			// In Go 1.20 and older, go test -c did not support multiple packages,
			// so use the next best thing of -exec=true to not run the test binary.
			// Note that 'true' here refers not to a boolean value, but a binary
			// (e.g., /usr/bin/true) that ignores parameters and exits with code 0.
			return append(append(args, "-exec=true"), patterns...)
		}
		return append(append(args, "-c", "-o", os.DevNull), patterns...)
	}
	if !b.inputs.LongTest {
		args = append(args, "-short")
	}
	if b.inputs.LongTest && isGoProject(b.inputs.Project) { // TODO(dmitshur): Delete after 1.20 drops off.
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

// distTestCmd returns an exec.Cmd for executing `go tool dist test`.
//
// dir is the directory to run the command from.
//
// runRx and testNames are optional, mutually exclusive flags that select
// a subset of dist tests to run using dist test's command-line interface.
// (See 'go tool dist test -help'.)
//
// If json is true, passes the -json flag, producing `go test -json`-compatible output.
// Note: -json is not supported before Go 1.21.
//
// TODO(go.dev/issue/59990): Delete the json argument when it becomes always true.
//
// It automatically applies additional dist flags based on the buildSpec (e.g. -race).
func (b *buildSpec) distTestCmd(ctx context.Context, dir, runRx string, testNames []string, json bool) *exec.Cmd {
	args := []string{"tool", "dist", "test"}
	if json {
		args = append(args, "-json")
	}
	args = append(args, b.distTestFlags()...)
	if runRx != "" {
		args = append(args, "-run="+runRx)
	}
	args = append(args, testNames...)
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

// distListCmd returns an exec.Cmd for executing `go tool dist list -json`.
// dir is the directory to run the command from.
func (b *buildSpec) distListCmd(ctx context.Context, dir string) *exec.Cmd {
	return b.goCmd(ctx, dir, "tool", "dist", "list", "-json")
}

const cloudProject = "golang-ci-luci"

func (b *buildSpec) installDatastoreClient(ctx context.Context) (context.Context, error) {
	// TODO(mknyszek): Enable auth only when not in a fake build.
	ts, err := b.auth.TokenSource()
	if err != nil {
		return nil, lucierrors.Annotate(err, "failed to initialize the token source").Err()
	}
	client, err := clouddatastore.NewClient(ctx, cloudProject, option.WithTokenSource(ts))
	if err != nil {
		return nil, lucierrors.Annotate(err, "failed to instantiate the datastore client").Err()
	}
	cfg := &cloud.ConfigLite{
		ProjectID: cloudProject,
		DS:        client,
	}
	return cfg.Use(ctx), nil
}

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
// If external network access is to be disabled, cmd is also prefixed with 'unshare'.
//
// It edits cmd in place but for convenience also returns cmd back to its caller.
func (b *buildSpec) wrapTestCmd(cmd *exec.Cmd) *exec.Cmd {
	if b.inputs.NoNetwork {
		// Disable external network access for the test command.
		// Permit internal loopback access.
		cmd.Args = []string{
			"unshare", "--net", "--map-root-user", "--",
			"sh", "-c", "ip link set dev lo up && " + strings.Join(cmd.Args, " "),
		}
	}

	// Compute all the test tags and variants we want to send to ResultDB.
	rdbArgs := []string{
		"-var", fmt.Sprintf("goos:%s", b.inputs.Target.Goos),
		"-var", fmt.Sprintf("goarch:%s", b.inputs.Target.Goarch),
		"-var", fmt.Sprintf("host_goos:%s", b.inputs.Host.Goos),
		"-var", fmt.Sprintf("host_goarch:%s", b.inputs.Host.Goarch),
		"-var", fmt.Sprintf("builder:%s", b.builderName),
		"-var", fmt.Sprintf("go_branch:%s", b.inputs.GoBranch),
		"-tag", fmt.Sprintf("bootstrap_version:%s", b.inputs.BootstrapVersion),
	}
	if b.inputs.RaceMode {
		rdbArgs = append(rdbArgs, "-tag", "run_mod:race")
	}
	if b.inputs.LongTest {
		rdbArgs = append(rdbArgs, "-tag", "run_mod:longtest")
	}
	if b.inputs.XcodeVersion != "" {
		rdbArgs = append(rdbArgs, "-tag", "xcode_version:"+b.inputs.XcodeVersion)
	}
	if b.inputs.NodeVersion != "" {
		rdbArgs = append(rdbArgs, "-tag", "node_version:"+b.inputs.NodeVersion)
	}

	// Assemble args and update the command.
	cmd.Path = b.toolPath("rdb")
	args := []string{cmd.Path, "stream"}
	args = append(args, rdbArgs...)
	args = append(args, "--", b.toolPath("result_adapter"), "go", "--")
	args = append(args, cmd.Args...)
	cmd.Args = args
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

// isGoProject reports whether a project should be treated as the main Go repo.
// For the moment, we develop security fixes in a project named golang/go-private.
func isGoProject(name string) bool {
	return name == "go" || name == "golang/go-private"
}
