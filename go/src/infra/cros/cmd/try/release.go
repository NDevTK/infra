// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"infra/cros/internal/cmd"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
)

func getCmdRelease() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "release [comma-separated build-targets|*] [-staging]",
		ShortDesc: "Run a release builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &releaseRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addBranchFlag()
			c.addStagingFlag()
			c.addPatchesFlag()
			c.Flags.BoolVar(&c.useProdTests, "prod_tests", false, "Use the production testing config even if staging.")
			return c
		},
	}
}

// releaseRun tracks relevant info for a given `try release` run.
type releaseRun struct {
	tryRunBase
	useProdTests bool
}

// validate validates release-specific args for the command.
func (r *releaseRun) validate(ctx context.Context) error {
	if err := r.tryRunBase.validate(ctx); err != nil {
		return err
	}
	return nil
}

// Run provides the logic for a `try release` command run.
func (r *releaseRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	if !r.staging {
		fmt.Println("Non-staging release builds are currently unsupported. Please try again with --staging.")
		return NotImplementedError
	}

	ctx := context.Background()
	if err := r.validate(ctx); err != nil {
		fmt.Println(err.Error())
		return CmdError
	}

	if err := r.EnsureLUCIToolsAuthed(ctx, "bb", "led"); err != nil {
		fmt.Println(err)
		return AuthError
	}

	propsStruct, err := r.GetBuilderInputProps(ctx, r.getReleaseOrchestratorName())
	if err != nil {
		fmt.Println(err)
		return CmdError
	}

	if len(r.patches) > 0 {
		// If gerrit patches are set, the orchestrator by default will try to do
		// build planning, which is meaningless for release builds and drops
		// all children. This property skips pruning.
		if err := setProperty(propsStruct, "$chromeos/build_plan.disable_build_plan_pruning", true); err != nil {
			fmt.Println(err)
			return CmdError
		}
		for _, patch := range r.patches {
			r.bbAddArgs = append(r.bbAddArgs, []string{"-cl", patch}...)
		}
	}

	if r.useProdTests {
		if err := setProperty(propsStruct, "$chromeos/cros_test_plan.use_prod_config", true); err != nil {
			fmt.Println(err)
			return CmdError
		}
	}

	propsFile, err := writeStructToFile(propsStruct)
	if err != nil {
		fmt.Println(errors.Annotate(err, "writing input properties to tempfile").Err())
		return UnspecifiedError
	}
	defer os.Remove(propsFile.Name())
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	if err := r.runReleaseOrchestrator(ctx); err != nil {
		fmt.Println(err.Error())
		return CmdError
	}

	return Success
}

// getReleaseOrchestratorName finds the full name of the release orchestrator matching the try CLI flags.
func (r *releaseRun) getReleaseOrchestratorName() string {
	const project = "chromeos"
	var bucket, builder, stagingPrefix string
	if r.staging {
		bucket = "staging"
		stagingPrefix = "staging-"
	} else {
		bucket = "release"
	}
	if strings.HasPrefix(r.branch, "release-") {
		builder = fmt.Sprintf("%s%s-orchestrator", stagingPrefix, r.branch)
	} else {
		builder = fmt.Sprintf("%srelease-%s-orchestrator", stagingPrefix, r.branch)
	}
	return fmt.Sprintf("%s/%s/%s", project, bucket, builder)
}

// runReleaseOrchestrator creates a release orchestrator build via `bb add`, and reports it to the user.
func (r *releaseRun) runReleaseOrchestrator(ctx context.Context) error {
	orchName := r.getReleaseOrchestratorName()
	return r.BBAdd(ctx, append([]string{orchName}, r.bbAddArgs...)...)
}
