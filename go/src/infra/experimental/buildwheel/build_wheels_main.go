package main

import (
	"context"
	"os/exec"

	"go.chromium.org/luci/luciexe/build"
)

// Entry-point for compiled build-wheels binary.
func main() {

	// `InputProps` is a self-defined proto.message with fields that match
	//  your top-level properties. It is populated with the contents of the
	//  build's input.properties inside build.Main.
	//
	// I.e. in this case we get access to the properties from build_wheels.py
	input := &InputProps{}

	build.Main(input, nil, nil, func(ctx context.Context, userArgs []string, state *build.State) error {
		executor := (func(*exec.Cmd) error)(nil)
		if input.ExperimentalDryrun {
			executor = CreateDryRunExecutor(true)
		}
		return RunDockerBuild(ctx, userArgs, state, executor)
	})
}
