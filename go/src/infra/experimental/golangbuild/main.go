// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary golangbuild is a luciexe binary that builds and tests the code for the
// Go project. It supports building and testing go.googlesource.com/go as well as
// Go project subrepositories (e.g. go.googlesource.com/net) and on different branches.
//
// To build and run this locally end-to-end, follow these steps:
//
//	luci-auth login -scopes "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/gerritcodereview https://www.googleapis.com/auth/cloud-platform"
//	cat > build.jsonpb <<EOF
//	{
//		"builder": {
//			"project": "go",
//			"bucket": "ci",
//			"builder": "linux-amd64"
//		},
//		"input": {
//			"properties": {
//				"project": "go",
//				"tools_cache": "tools"
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
//	export SWARMING_SERVER=https://chromium-swarm.appspot.com
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
//				"tools_cache": "tools",
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
//   - Keep step presentation and command execution separate from logic where possible
//     (which will make it easier to write unit tests).
//   - Steps should be function-scoped. Steps should be created at the start of a function
//     with the step end immediately deferred to function exit.
//
// ## Experiments
//
// When adding new functionality, consider putting it behind an experiment. Experiments are
// made available in the buildSpec and are propagated from the builder definition.
// Experiments in the builder definition are given a probability of being enabled on any given
// build, but always manifest in the current build as either "on" or "off."
// Experiments should have a name like "golang.my_specific_new_functionality" and should
// be checked for via spec.experiment("golang.my_specific_new_functionality").
//
// Experiments can be disabled at first (no work needs to be done on the builder definition),
// rolled out, and then tested in a real build environment via `led`
//
//	led get-build ... | \
//	led edit \
//	  -experiment golang.my_specific_new_functionality=true | \
//	led launch
//
// or via `bb add -ex "+golang.my_specific_new_functionality" ...`.
//
// Experiments can be enabled on LUCIEXE_FAKEBUILD runs through the "experiments" property (array
// of strings) on "input."
//
// ### Current experiments
//
//   - golang.force_test_outside_repository: Can be used to force running tests
//     from outside the repository to catch accidental reads outside of module
//     boundaries despite the repository not having opted-in to this test
//     behavior.
//   - golang.no_network_in_short_test_mode: Disable network access in -short
//     test mode. In the process of being gradually rolled out to all repos.
//   - golang.no_network_in_short_test_mode_v2: Like above, but unshare uses
//     the -c (--map-current-user) flag instead of -r (--map-root-user).
package main

import (
	"context"
	"log"
	"os"

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

	// Collect enabled experiments.
	experiments := make(map[string]struct{})
	for _, ex := range st.Build().GetInput().GetExperiments() {
		experiments[ex] = struct{}{}
	}

	// Install some tools we'll need, including a bootstrap toolchain.
	toolsRoot, err := installTools(ctx, inputs)
	if err != nil {
		return err
	}
	log.Printf("installed tools")

	// Define working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return infraErrorf("get CWD")
	}

	spec, err := deriveBuildSpec(ctx, cwd, toolsRoot, experiments, st, inputs)
	if err != nil {
		return infraWrap(err)
	}

	// Set up environment.
	ctx = spec.setEnv(ctx)
	ctx, err = spec.installDatastoreClient(ctx)
	if err != nil {
		return err
	}

	// Select a runner based on the mode, then initialize and invoke it.
	var rn runner
	switch inputs.GetMode() {
	case golangbuildpb.Mode_MODE_ALL:
		rn = newAllRunner(inputs.GetAllMode())
	case golangbuildpb.Mode_MODE_COORDINATOR:
		rn = newCoordRunner(inputs.GetCoordMode())
	case golangbuildpb.Mode_MODE_BUILD:
		rn = newBuildRunner(inputs.GetBuildMode())
	case golangbuildpb.Mode_MODE_TEST:
		rn, err = newTestRunner(inputs.GetTestMode(), inputs.GetTestShard())
	}
	if err != nil {
		return infraErrorf("initializing runner: %v", err)
	}
	return rn.Run(ctx, spec)
}

// runner is an interface that provides an abstraction over golangbuild's various modes.
//
// Every mode basically requires the same setup at the beginning of the build; runner
// determines what to do once we have all that.
type runner interface {
	Run(ctx context.Context, spec *buildSpec) error
}
