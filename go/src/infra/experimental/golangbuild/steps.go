// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/luciexe/build"
)

// cmdStepRun calls Run on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepRun(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return err
	}

	// Combine output because it's annoying to pick one of stdout and stderr
	// in the UI and be wrong.
	output := step.Log("output")
	cmd.Stdout = output
	cmd.Stderr = output

	// Run the command.
	if err := cmd.Run(); err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}

// cmdStepOutput calls Output on the provided command and wraps it in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStepOutput(ctx context.Context, stepName string, cmd *exec.Cmd, infra bool) (output []byte, err error) {
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return nil, err
	}

	// Make sure we log stderr.
	cmd.Stderr = step.Log("stderr")

	// Run the command and capture stdout.
	output, err = cmd.Output()

	// Log stdout before we do anything else.
	step.Log("stdout").Write(output)

	// Check for errors.
	if err != nil {
		return output, errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return output, nil
}

// goModDownloadStep runs a 'go mod download -json' command in a build step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func goModDownloadStep(ctx context.Context, stepName string, cmd *exec.Cmd) (err error) {
	var infra bool
	step, ctx, err := cmdStartStep(ctx, stepName, cmd)
	defer func() {
		if infra {
			err = infraWrap(err) // Failure is deemed to be an infrastructure failure.
		}
		step.End(err)
	}()
	if err != nil {
		return err
	}

	// Run 'go mod download -json' and process its output
	// to determine if this was an infrastructure failure.
	var stdout bytes.Buffer
	cmd.Stdout = io.MultiWriter(step.Log("stdout"), &stdout)
	cmd.Stderr = step.Log("stderr")
	err = cmd.Run()
	if ee := (*exec.ExitError)(nil); errors.As(err, &ee) && ee.ExitCode() == 1 {
		for dec := json.NewDecoder(&stdout); ; {
			var m struct{ Error string }
			err := dec.Decode(&m)
			if err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("error decoding JSON object from go mod download -json: %v\n", err)
			}
			if strings.Contains(m.Error, "dial tcp") && strings.HasSuffix(m.Error, ": i/o timeout") {
				// An I/O timeout error to the Go module proxy is deemed to be an infrastructure failure.
				// See https://ci.chromium.org/b/8772399708036918561 for an example.
				infra = true
				break
			}
		}
	}
	if err != nil {
		return errors.Annotate(err, "Failed to run %q", stepName).Err()
	}
	return nil
}

// cmdStartStep sets up a command step.
//
// It overwrites cmd.Stdout and cmd.Stderr to redirect into step logs.
// It runs the command with the environment from the context, so change
// the context's environment to alter the command's environment.
func cmdStartStep(ctx context.Context, stepName string, cmd *exec.Cmd) (*build.Step, context.Context, error) {
	step, ctx := build.StartStep(ctx, stepName)

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
		return step, ctx, err
	}
	return step, ctx, nil
}

func infraErrorf(s string, args ...any) error {
	return build.AttachStatus(fmt.Errorf(s, args...), bbpb.Status_INFRA_FAILURE, nil)
}

func infraWrap(err error) error {
	return build.AttachStatus(err, bbpb.Status_INFRA_FAILURE, nil)
}

func endStep(step *build.Step, errp *error) {
	step.End(*errp)
}

func endInfraStep(step *build.Step, errp *error) {
	step.End(infraWrap(*errp))
}
