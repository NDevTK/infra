package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"go.chromium.org/luci/luciexe/build"
)

func CreateDryRunExecutor(mockPass bool) func(*exec.Cmd) error {
	return func(cmd *exec.Cmd) error {
		cmdStr := fmt.Sprintf("dryrun mode for command: %v", cmd.Args)

		_, err := cmd.Stdout.Write([]byte(fmt.Sprintf("stdout %v", cmdStr)))
		if err != nil {
			panic(err)
		}

		_, err = cmd.Stderr.Write([]byte(fmt.Sprintf("stderr %v", cmdStr)))
		if err != nil {
			panic(err)
		}

		if !mockPass {
			return errors.New("Fake error from DryRunExecutor to mock failure")
		}

		return nil
	}
}

// Runs the docker build command for a luciexe binary.
// It takes in an executor function or nil for dry-run and testing purposes.
func RunDockerBuild(ctx context.Context, userArgs []string, state *build.State, executor func(*exec.Cmd) error) error {

	step, _ := build.StartStep(ctx, "dockerbuild")
	var err error
	defer func() { step.End(err) }()

	// TODO (ranijiang): Add execution details log
	// Example: https://logs.chromium.org/logs/infra-internal/buildbucket/cr-buildbucket/8795032547863921041/+/u/gclient_verify/execution_details

	dockerBuildCmd := exec.Command("vpython3", "-m", "infra.tools.dockerbuild")
	dockerBuildCmd.Args = append(dockerBuildCmd.Args, userArgs...)

	stdoutLog := step.Log("stdout")
	stderrLog := step.Log("stderr")
	dockerBuildCmd.Stdout = stdoutLog
	dockerBuildCmd.Stderr = stderrLog

	if executor == nil {
		err = dockerBuildCmd.Run()
	} else {
		err = executor(dockerBuildCmd)
	}

	return err
}
