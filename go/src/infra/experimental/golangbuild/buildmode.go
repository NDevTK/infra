// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"infra/experimental/golangbuild/golangbuildpb"

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
	if err := fetchRepo(ctx, spec.goSrc, spec.goroot, spec.inputs, spec.experiments); err != nil {
		return err
	}

	// Build Go.
	if err := cmdStepRun(ctx, "make"+scriptExt(), spec.goScriptCmd(ctx, "make"+scriptExt()), false); err != nil {
		return err
	}

	// Upload to CAS.
	return uploadGoToCAS(ctx, spec, spec.goSrc, spec.goroot)
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
