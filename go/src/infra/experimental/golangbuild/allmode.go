// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"infra/experimental/golangbuild/golangbuildpb"
)

// allRunner gets Go (building it if no prebuilt toolchain is available for the current
// platform), then runs tests for the current project.
//
// This is the most basic form of serial execution and represents the simplest Go build.
type allRunner struct {
	props *golangbuildpb.AllMode
}

// newAllRunner creates a new AllMode runner.
func newAllRunner(props *golangbuildpb.AllMode) *allRunner {
	return &allRunner{props: props}
}

// Run implements the runner interface for allRunner.
func (r *allRunner) Run(ctx context.Context, spec *buildSpec) error {
	// Get a built Go toolchain or build it if necessary.
	if err := getGo(ctx, spec, false); err != nil {
		return err
	}
	if spec.inputs.Project == "go" {
		return runGoTests(ctx, spec, noSharding)
	}
	return fetchSubrepoAndRunTests(ctx, spec)
}
