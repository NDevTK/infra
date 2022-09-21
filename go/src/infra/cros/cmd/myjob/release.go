// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
	"infra/cros/internal/cmd"
)

func getCmdRelease() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "release [comma-separated build-targets|*] [-staging]",
		ShortDesc: "Run a release builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &releaseRun{}
			c.myjobRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addBranchFlag()
			c.addStagingFlag()
			return c
		},
	}
}

// releaseRun tracks relevant info for a given `myjob release` run.
type releaseRun struct {
	myjobRunBase
}

// Run provides the logic for a `myjob release` command run.
func (r *releaseRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	if !r.staging {
		fmt.Println("Non-staging release builds are currently unsupported. Please try again with -staging.")
		return NotImplementedError
	}

	ctx := context.Background()
	if err := r.EnsureLUCIToolsAuthed(ctx, "bb", "led"); err != nil {
		fmt.Println(err)
		return AuthError
	}

	propsStruct, err := r.GetBuilderInputProps(ctx, r.getReleaseOrchestratorName())
	if err != nil {
		fmt.Println(err)
		return CmdError
	}
	propsFile, err := writeStructToFile(propsStruct)
	if err != nil {
		fmt.Println(errors.Annotate(err, "writing input properties to tempfile").Err())
		return UnspecifiedError
	}
	defer os.Remove(propsFile.Name())

	if err := r.runReleaseOrchestrator(ctx); err != nil {
		fmt.Println(err.Error())
		return CmdError
	}

	return Success
}

// getReleaseOrchestratorName finds the full name of the release orchestrator matching the myjob CLI flags.
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
	return r.BBAdd(ctx, orchName)
}
