// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"flag"
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
		UsageLine: "release [flags]",
		ShortDesc: "Run a release builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &releaseRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.addBranchFlag("main")
			c.addProductionFlag()
			c.addPatchesFlag()
			c.addBuildTargetsFlag()
			c.addBuildspecFlag()
			c.Flags.BoolVar(&c.useProdTests, "prod_tests", false, "Use the production testing config even if staging.")
			c.Flags.BoolVar(&c.skipPaygen, "skip_paygen", false, "Skip payload generation. Only supported for staging builds.")
			if flag.NArg() > 1 && flag.Args()[1] == "help" {
				fmt.Printf("Run `cros try help` or `cros try help ${subcomand}` for help.")
				os.Exit(0)
			}
			return c
		},
	}
}

// releaseRun tracks relevant info for a given `try release` run.
type releaseRun struct {
	tryRunBase
	useProdTests bool
	skipPaygen   bool
	// Used for testing purposes. If set, props will be written to this file
	// rather than a temporary one.
	propsFile *os.File
}

// validate validates release-specific args for the command.
func (r *releaseRun) validate() error {
	if r.skipPaygen && r.production {
		return fmt.Errorf("--skip_paygen is not supported for production builds")
	}

	if strings.HasPrefix(r.branch, "stabilize-") && !r.production {
		return fmt.Errorf("can only run production builds for stabilize branches")
	}

	if err := r.tryRunBase.validate(); err != nil {
		return err
	}
	return nil
}

// checkChildrenExist checks that any explicitly requested build targets
// have a builder for the relevant branch.
func (r *releaseRun) checkChildrenExist(ctx context.Context) error {
	if len(r.buildTargets) > 0 {
		builderNames := r.getReleaseBuilderNames()
		bucket := "staging"
		if r.production {
			bucket = "release"
		}
		for i, builderName := range builderNames {
			fullBuilderName := fmt.Sprintf("chromeos/%s/%s", bucket, builderName)
			_, err := r.GetBuilderInputProps(ctx, fullBuilderName)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("%s is not a valid build target for %s", r.buildTargets[i], r.branch)
			}
		}
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

	if r.production && !r.skipProductionPrompt {
		if yes, err := r.promptYes(); err != nil {
			r.LogErr(err.Error())
			return CmdError
		} else if !yes {
			r.LogOut("Exiting.")
			return Success
		}
	}

	ctx := context.Background()
	if ret, err := r.run(ctx); err != nil {
		r.LogErr(err.Error())
		return ret
	}

	if err := r.checkChildrenExist(ctx); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	propsStruct, err := r.GetBuilderInputProps(ctx, r.getReleaseOrchestratorName())
	if err != nil {
		r.LogErr(err.Error())
		if strings.Contains(err.Error(), "not found") {
			if strings.HasPrefix(r.branch, "stabilize-") {
				r.LogErr(fmt.Sprintf("Builder not found, is '%s' defined in stabilize_builders.textpb?", r.branch))
			}
		}
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
		r.bbAddArgs = append(r.bbAddArgs, patchListToBBAddArgs(r.patches)...)
	}

	if r.useProdTests {
		if err := setProperty(propsStruct, "$chromeos/cros_test_plan.use_prod_config", true); err != nil {
			r.LogErr(err.Error())
			return CmdError
		}
	}

	if r.skipPaygen {
		if err := setProperty(propsStruct, "$chromeos/orch_menu.skip_paygen", true); err != nil {
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

	if r.buildspec != "" {
		if err := setProperty(propsStruct, "$chromeos/cros_source.syncToManifest.manifestGsPath", r.buildspec); err != nil {
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
	if r.production {
		bucket = "release"
	} else {
		bucket = "staging"
		stagingPrefix = "staging-"
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
	if !r.production {
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
