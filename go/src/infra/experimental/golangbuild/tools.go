// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/experimental/golangbuild/golangbuildpb"
)

// CIPD dependencies for non-coordinator mode builds.
//
// N.B. We assume a few tools are already available on the machine we're
// running on. Namely:
// - For non-Windows, a C/C++ toolchain
const cipdBuildDeps = `
@Subdir
infra/3pp/tools/git/${platform} version:2@2.39.2.chromium.11
@Subdir cc/${os=windows}
golang/third_party/llvm-mingw-msvcrt/${platform} latest
`

// CIPD tool dependencies only. Used for coordinator builds.
const cipdToolDeps = `
@Subdir bin
infra/tools/bb/${platform} latest
infra/tools/rdb/${platform} latest
infra/tools/luci/cas/${platform} latest
infra/tools/result_adapter/${platform} latest
`

// CIPD dependency for XCode.
const cipdXCodeDep = `
@Subdir
infra/tools/mac_toolchain/${platform} latest
`

func installTools(ctx context.Context, inputs *golangbuildpb.Inputs, experiments map[string]struct{}) (toolsRoot string, err error) {
	step, ctx := build.StartStep(ctx, "install tools")
	defer endInfraStep(step, &err) // Any failure in this function is an infrastructure failure.

	// Construct the CIPD ensure file.
	var cipdDeps string
	gotXCode := false

	if inputs.GetMode() == golangbuildpb.Mode_MODE_COORDINATOR {
		cipdDeps = cipdToolDeps
	} else {
		// Don't install git from CIPD on less-common platforms. We'll get it from the external OS as needed.
		if _, bestEffortPlatform := experiments["luci.best_effort_platform"]; !bestEffortPlatform {
			cipdDeps = cipdBuildDeps
		}

		cipdDeps += cipdToolDeps + fmt.Sprintf(`
@Subdir go_bootstrap
golang/bootstrap-go/${platform} %v
`, inputs.BootstrapVersion)
		if inputs.XcodeVersion != "" {
			gotXCode = true
			cipdDeps += cipdXCodeDep
		}
	}

	io.WriteString(step.Log("ensure file"), cipdDeps)

	// Store in the named cache specified in Inputs. This is shared across
	// builder types, allowing reuse across builds if the dependencies
	// versions are the same.
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
	if gotXCode {
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
