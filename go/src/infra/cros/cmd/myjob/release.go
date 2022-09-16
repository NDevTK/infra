// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"fmt"
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
	if bbAuthed, err := r.IsBBAuthed(ctx); err != nil {
		fmt.Println(errors.Annotate(err, "determining whether `bb` is authed").Err())
		return AuthError
	} else if !bbAuthed {
		fmt.Println("bb CLI is not logged in. Please run the following command, then try again:\n\tbb auth-login")
		return AuthError
	}

	if err := r.runReleaseOrchestrator(ctx); err != nil {
		fmt.Println(err.Error())
		return BBError
	}

	return Success
}

// getReleaseOrchestratorName finds the name of the release orchestrator matching the myjob CLI flags.
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
	var stdoutBuf, stderrBuf bytes.Buffer
	err := r.RunCmd(ctx, &stdoutBuf, &stderrBuf, "", "bb", "add", orchName)
	if err != nil {
		fmt.Printf("`bb add %s` had stderr:\n%s\n", orchName, stderrBuf.String())
		return errors.Annotate(err, "running bb add command").Err()
	}
	fmt.Println(stdoutBuf.String())
	return nil
}
