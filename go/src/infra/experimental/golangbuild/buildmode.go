// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"infra/experimental/golangbuild/golangbuildpb"
	"path/filepath"

	"go.chromium.org/luci/luciexe/build"
)

// buildRunner ensures a prebuilt toolchain exists for the current platform and the
// the sourceSpec this build was invoked with. It builds one if necessary and uploads
// it to CAS.
//
// This implements "build mode" for golangbuild.
type buildRunner struct {
	props *golangbuildpb.BuildMode
}

// newBuildRunner creates a new BuildMode runner.
func newBuildRunner(props *golangbuildpb.BuildMode) *buildRunner {
	return &buildRunner{props: props}
}

// Run implements the runner interface for buildRunner.
func (r *buildRunner) Run(ctx context.Context, spec *buildSpec) error {
	// Grab a prebuilt toolchain or build one and upload it.
	return getGo(ctx, spec, false)
}

func getGo(ctx context.Context, spec *buildSpec, requirePrebuilt bool) (err error) {
	step, ctx := build.StartStep(ctx, "get go")
	defer endStep(step, &err)

	defer func() {
		if err != nil {
			return
		}

		// Run `go env` on the resulting toolchain for debugging purposes.
		_ = cmdStepRun(ctx, "go env", spec.goCmd(ctx, spec.goroot, "env"), true)
	}()

	// Check to see if we might have a prebuilt Go in CAS.
	digest, err := checkForPrebuiltGo(ctx, spec)
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
	if err := fetchRepo(ctx, spec.goSrc, spec.goroot, spec.inputs); err != nil {
		return err
	}

	// Write out the VERSION file.
	var version string
	switch {
	case spec.inputs.VersionFile != "":
		version = spec.inputs.VersionFile
	case spec.goSrc.change != nil:
		version = fmt.Sprintf("devel %d/%d", spec.goSrc.change.Change, spec.goSrc.change.Patchset)
	case spec.goSrc.commit != nil:
		version = fmt.Sprintf("devel %s", spec.goSrc.commit.Id)
	}
	if err := writeFile(ctx, filepath.Join(spec.goroot, "VERSION"), version); err != nil {
		return err
	}

	// Build Go.
	ext := scriptExt(spec.inputs.Host)
	if err := cmdStepRun(ctx, "make"+ext, spec.goScriptCmd(ctx, "make"+ext), false); err != nil {
		return err
	}

	// Upload to CAS.
	return uploadGoToCAS(ctx, spec, spec.goSrc, spec.goroot)
}

// scriptExt returns the extension to use for
// GOROOT/src/{make,all} scripts on this GOOS.
func scriptExt(host *golangbuildpb.Port) string {
	switch host.Goos {
	case "windows":
		return ".bat"
	case "plan9":
		return ".rc"
	default:
		return ".bash"
	}
}
