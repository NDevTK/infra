// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"log"
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
			c.addBranchFlag("main")
			c.addStagingFlag()
			c.addPatchesFlag()
			c.addBuildTargetsFlag()
			c.Flags.BoolVar(&c.useProdTests, "prod_tests", false, "Use the production testing config even if staging.")
			return c
		},
	}
}

// releaseRun tracks relevant info for a given `try release` run.
type releaseRun struct {
	tryRunBase
	useProdTests bool
	// Used for testing purposes. If set, props will be written to this file
	// rather than a temporary one.
	propsFile *os.File
}

// validate validates release-specific args for the command.
func (r *releaseRun) validate() error {
	if !r.staging {
		return fmt.Errorf("Non-staging release builds are currently unsupported. Please try again with --staging.")
	}

	if err := r.tryRunBase.validate(); err != nil {
		return err
	}
	return nil
}

// Run provides the logic for a `try release` command run.
func (r *releaseRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	if err := r.validate(); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	ctx := context.Background()
	if err := r.EnsureLUCIToolsAuthed(ctx, "bb", "led"); err != nil {
		r.LogErr(err.Error())
		return AuthError
	}

	propsStruct, err := r.GetBuilderInputProps(ctx, r.getReleaseOrchestratorName())
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	if len(r.patches) > 0 {
		// If gerrit patches are set, the orchestrator by default will try to do
		// build planning, which is meaningless for release builds and drops
		// all children. This property skips pruning.
		if err := setProperty(propsStruct, "$chromeos/build_plan.disable_build_plan_pruning", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
		for _, patch := range r.patches {
			r.bbAddArgs = append(r.bbAddArgs, []string{"-cl", patch}...)
		}
	}

	if r.useProdTests {
		if err := setProperty(propsStruct, "$chromeos/cros_test_plan.use_prod_config", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	if len(r.buildTargets) > 0 {
		if err := setProperty(propsStruct, "$chromeos/orch_menu.child_builds", r.getReleaseBuilderNames()); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	var propsFile *os.File
	if r.propsFile != nil {
		propsFile = r.propsFile
	} else {
		propsFile, err = os.CreateTemp("", "input_props")
		if err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}
	if err := writeStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	if err := r.runReleaseOrchestrator(ctx); err != nil {
		r.LogErr(err.Error())
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

func (r *releaseRun) getReleaseBuilderNames() []string {
	const project = "chromeos"
	var builder, stagingPrefix string
	if r.staging {
		stagingPrefix = "staging-"
	}
	builderNames := make([]string, len(r.buildTargets))
	for i, buildTarget := range r.buildTargets {
		if strings.HasPrefix(r.branch, "release-") {
			builder = fmt.Sprintf("%s%s-%s", stagingPrefix, buildTarget, r.branch)
		} else {
			builder = fmt.Sprintf("%s%s-release-%s", stagingPrefix, buildTarget, r.branch)
		}
		builderNames[i] = builder
	}

	return builderNames
}

// runReleaseOrchestrator creates a release orchestrator build via `bb add`, and reports it to the user.
func (r *releaseRun) runReleaseOrchestrator(ctx context.Context) error {
	orchName := r.getReleaseOrchestratorName()
	return r.BBAdd(ctx, append([]string{orchName}, r.bbAddArgs...)...)
}
