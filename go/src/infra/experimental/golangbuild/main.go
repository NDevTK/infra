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
//
// ## Contributing
//
// To keep things organized and consistent, keep to the following guidelines:
//   - Only functions generate steps. Methods never generate steps.
//   - Keep step presentation and high-level ordering logic in main.go when possible.
//   - Steps should be function-scoped. Steps should be created at the start of a function
//     with the step end immediately deferred to function exit.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"

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

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx)
	if err != nil {
		return err
	}
	log.Printf("installed tools")

	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return build.AttachStatus(errors.Annotate(err, "Get CWD").Err(), bbpb.Status_INFRA_FAILURE, nil)
	}

	spec, err := deriveBuildSpec(ctx, cwd, toolsRoot, st, inputs)
	if err != nil {
		return build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
	}

	// Set up environment.
	ctx = spec.setEnv(ctx)

	// Fetch the main Go repository into goroot.
	if spec.invokedSrc.project == "go" {
		if err := fetchRepo(ctx, spec.invokedSrc, spec.goroot); err != nil {
			return err
		}
	} else {
		// We're fetching the Go repo for a subrepo build against a subrepo CL.
		if err := fetchRepo(ctx, &sourceSpec{project: "go", branch: inputs.GoBranch}, spec.goroot); err != nil {
			return err
		}
	}

	// Build Go.
	//
	// TODO(mknyszek): Support cross-compile-only modes, perhaps by having CompileGOOS
	// and CompileGOARCH repeated fields in the input proto to identify what to build.
	// TODO(mknyszek): Grab a prebuilt copy available.
	// TODO(mknyszek): Upload the result of make.bash somewhere that downstream builders can find.
	if err := runCommandAsStep(ctx, "make"+scriptExt(), spec.goScriptCmd(ctx, "make"+scriptExt()), false); err != nil {
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
		// TODO(mknyszek): Support sharding by running `go tool dist test -list` and
		// triggering N test builders with a subset of those tests in their properties.
		// Pass the newly-built toolchain via CAS.
		distTestArgs := []string{"tool", "dist", "test", "-no-rebuild"}
		if spec.inputs.RaceMode {
			distTestArgs = append(distTestArgs, "-race")
		}
		testCmd := spec.goCmd(ctx, spec.goroot, distTestArgs...)
		if err := runCommandAsStep(ctx, "go tool dist test", testCmd, false); err != nil {
			return err
		}
	} else {
		// Fetch the target repository into targetrepo.
		if spec.invokedSrc.project == "go" {
			if err := fetchRepo(ctx, &sourceSpec{project: inputs.Project, branch: mainBranch}, spec.subrepoDir); err != nil {
				return err
			}
		} else {
			// We're testing the tip of spec.inputs.Project against a Go commit.
			if err := fetchRepo(ctx, spec.invokedSrc, spec.subrepoDir); err != nil {
				return err
			}
		}

		// Test this specific subrepo.
		goArgs := []string{"test", "-json"}
		if spec.inputs.RaceMode {
			goArgs = append(goArgs, "-race")
		}
		goArgs = append(goArgs, "./...")
		testCmd := spec.goCmd(ctx, spec.subrepoDir, goArgs...)
		spec.wrapTestCmd(testCmd)
		if err := runCommandAsStep(ctx, "go test -json [-race] ./...", testCmd, false); err != nil {
			return err
		}
	}
	return nil
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
