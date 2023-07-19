// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"hash/crc32"
	"infra/experimental/golangbuild/golangbuildpb"
)

// runner is an interface that provides an abstraction over golangbuild's various modes.
//
// Every mode basically requires the same setup at the beginning of the build; runner
// determines what to do once we have all that.
type runner interface {
	Run(ctx context.Context, spec *buildSpec) error
}

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
		// Trigger downstream builds (of subrepo builders) with the commit and/or Gerrit change we got.
		//
		// TODO(mknyszek): This is for backwards compatibility. Soon only coordinator mode invocations
		// of golangbuild will be allowed to schedule builds, at which point this will fail to work.
		if builders := spec.inputs.BuildersToTrigger; len(builders) > 0 {
			if err := triggerDownstreamBuilds(ctx, spec, builders...); err != nil {
				return err
			}
		}
		return runGoTests(ctx, spec, noSharding)
	}
	return fetchSubrepoAndRunTests(ctx, spec)
}

// coordRunner ensures a prebuilt Go toolchain exists (launching a build to build one if
// necessary) and then launches test builds, potentially more than one (to shard test execution).
//
// This implements "coordinator mode" for golangbuild. It's called coordinator mode because it
// coordinates build and test from afar, and provides part of the functionality that the old Go
// CI system's coordinator provided.
type coordRunner struct {
	props *golangbuildpb.CoordinatorMode
}

// newCoordRunner creates a new CoordinatorMode runner.
func newCoordRunner(props *golangbuildpb.CoordinatorMode) *coordRunner {
	return &coordRunner{props: props}
}

// Run implements the runner interface for coordRunner.
func (r *coordRunner) Run(ctx context.Context, spec *buildSpec) error {
	// Ensure prebuilt Go exists.
	if err := ensurePrebuiltGoExists(ctx, spec, r.props.BuildBuilder); err != nil {
		return err
	}
	if spec.inputs.Project == "go" {
		// Trigger downstream builders (subrepo builders) with the commit and/or Gerrit change we got.
		if builders := r.props.GetBuildersToTriggerAfterToolchainBuild(); len(builders) > 0 {
			if err := triggerDownstreamBuilds(ctx, spec, builders...); err != nil {
				return err
			}
		}
	}
	// Launch and wait on test shards.
	return runTestShards(ctx, spec, r.props.NumTestShards, r.props.TestBuilder)
}

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

// testRunner runs a non-strict subset of available tests. It requires a prebuilt
// toolchain to be available (it will not create one on-demand).
//
// This implements "test mode" for golangbuild.
type testRunner struct {
	props *golangbuildpb.TestMode
	shard testShard
}

// newTestRunner creates a new TestMode runner.
func newTestRunner(props *golangbuildpb.TestMode, gotShard *golangbuildpb.TestShard) (*testRunner, error) {
	shard := noSharding
	if gotShard != nil {
		shard = testShard{
			shardID: gotShard.ShardId,
			nShards: gotShard.NumShards,
		}
		if shard.shardID >= shard.nShards {
			return nil, fmt.Errorf("invalid test shard designation: shard ID is %d, num shards is %d", shard.shardID, shard.nShards)
		}
	}
	return &testRunner{props: props, shard: shard}, nil
}

// Run implements the runner interface for testRunner.
func (r *testRunner) Run(ctx context.Context, spec *buildSpec) error {
	// Get a built Go toolchain and require it to be prebuilt.
	if err := getGo(ctx, spec, true); err != nil {
		return err
	}
	if spec.inputs.Project == "go" {
		return runGoTests(ctx, spec, r.shard)
	}
	return fetchSubrepoAndRunTests(ctx, spec)
}

// testShard is a test shard identity that can be used to deterministically filter tests.
type testShard struct {
	shardID uint32
	nShards uint32
}

// shouldRunTest deterministically returns true for whether the shard identity should run
// the test by name. The name of the test doesn't matter, as long as it's consistent across
// test shards.
func (s testShard) shouldRunTest(name string) bool {
	return crc32.ChecksumIEEE([]byte(name))%s.nShards == s.shardID
}

// noSharding indicates that no sharding should take place. It represents executing the entirety
// of the test suite.
var noSharding = testShard{shardID: 0, nShards: 1}
