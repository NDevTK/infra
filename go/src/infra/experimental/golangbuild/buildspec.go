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

	"go.chromium.org/luci/auth"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/api/gerrit"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/luciexe/build"
	"go.chromium.org/luci/luciexe/build/cv"
)

type buildSpec struct {
	auth *auth.Authenticator

	builderName string
	goroot      string
	subrepoDir  string
	gopath      string
	gocacheDir  string
	toolsRoot   string

	inputs *golangbuildpb.Inputs

	invokedSrc *sourceSpec
}

func deriveBuildSpec(ctx context.Context, cwd, toolsRoot string, st *build.State, inputs *golangbuildpb.Inputs) (*buildSpec, error) {
	authOpts := chromeinfra.SetDefaultAuthOptions(auth.Options{
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gerritcodereview",
		},
	})
	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)

	// Build the sourceSpec we were invoked with.
	isDryRun := false
	if mode, err := cv.RunMode(ctx); err == nil {
		isDryRun = strings.HasSuffix(mode, "DRY_RUN")
	} else if err != cv.ErrNotActive {
		return nil, err
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
			return nil, err
		}
		gc, err := gerrit.NewRESTClient(hc, gerritChange.Host, true)
		if err != nil {
			return nil, err
		}
		changeInfo, err := gc.GetChange(ctx, &gerritpb.GetChangeRequest{
			Number:  gerritChange.Change,
			Project: gerritChange.Project,
		})
		if err != nil {
			return nil, err
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
	src := &sourceSpec{
		project: changedProject,
		branch:  changedBranch,
		commit:  gitilesCommit,
		change:  gerritChange,
		rebase:  !isDryRun && gerritChange != nil, // Rebase change onto branch if it's not a dry run.
	}

	// Validate BuildersToTrigger invariant.
	if inputs.Project != "go" && len(inputs.BuildersToTrigger) != 0 {
		return nil, fmt.Errorf("specified builders to trigger for unsupported project")
	}

	var subrepoDir string
	if inputs.Project != "go" {
		subrepoDir = filepath.Join(cwd, "targetrepo")
	}
	return &buildSpec{
		auth:        authenticator,
		builderName: st.Build().GetBuilder().GetBuilder(),
		goroot:      filepath.Join(cwd, "goroot"),
		gopath:      filepath.Join(cwd, "gopath"),
		gocacheDir:  filepath.Join(cwd, "gocache"),
		subrepoDir:  subrepoDir,
		toolsRoot:   toolsRoot,
		inputs:      inputs,
		invokedSrc:  src,
	}, nil
}

func (b *buildSpec) setEnv(ctx context.Context) context.Context {
	env := environ.FromCtx(ctx)
	env.Load(b.inputs.Env)
	env.Set("GOROOT_BOOTSTRAP", filepath.Join(b.toolsRoot, "go_bootstrap"))
	env.Set("GOPATH", b.gopath) // Explicitly set to an empty per-build directory, to avoid reusing the implicit default one.
	env.Set("GOBIN", "")
	env.Set("GOROOT", "") // Clear GOROOT because it's likely someone has one set locally, e.g. for luci-go development.
	env.Set("GOCACHE", b.gocacheDir)
	env.Set("GO_BUILDER_NAME", b.builderName)
	// Use our tools before the system tools. Notably, use raw Git rather than the Chromium wrapper.
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.toolsRoot, "bin"), os.PathListSeparator, env.Get("PATH")))

	if runtime.GOOS == "windows" {
		env.Set("GOBUILDEXIT", "1") // On Windows, emit exit codes from .bat scripts. See go.dev/issue/9799.
		// TODO(heschi): select gcc32 for GOARCH=i386
		env.Set("PATH", fmt.Sprintf("%v%c%v", env.Get("PATH"), os.PathListSeparator, filepath.Join(b.toolsRoot, "cc/windows/gcc64/bin")))
	}
	return env.SetInCtx(ctx)
}

func (b *buildSpec) toolPath(tool string) string {
	return filepath.Join(b.toolsRoot, "bin", tool)
}

func (b *buildSpec) toolCmd(ctx context.Context, tool string, args ...string) *exec.Cmd {
	return command(ctx, b.toolPath(tool), args...)
}

func (b *buildSpec) wrapTestCmd(cmd *exec.Cmd) {
	cmd.Path = b.toolPath("rdb")
	cmd.Args = append([]string{
		cmd.Path, "stream", "--",
		b.toolPath("result_adapter"), "go", "--",
	}, cmd.Args...)
}

func (b *buildSpec) goScriptCmd(ctx context.Context, script string) *exec.Cmd {
	dir := filepath.Join(b.goroot, "src")
	cmd := command(ctx, filepath.FromSlash("./"+script))
	cmd.Dir = dir
	return cmd
}

func (b *buildSpec) goCmd(ctx context.Context, dir string, args ...string) *exec.Cmd {
	env := environ.FromCtx(ctx)
	// Ensure the go binary found in PATH is the same as the one we're about to execute.
	env.Set("PATH", fmt.Sprintf("%v%c%v", filepath.Join(b.goroot, "bin"), os.PathListSeparator, env.Get("PATH")))
	cmd := command(env.SetInCtx(ctx), filepath.Join(b.goroot, "bin", "go"), args...)
	cmd.Dir = dir
	return cmd
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
