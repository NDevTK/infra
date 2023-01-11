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

func getCmdFirmware() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "firmware --branch BRANCH [flags]",
		ShortDesc: "Run a firmware branch builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &firmwareRun{}
			c.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.addBranchFlag("")
			c.addProductionFlag()
			c.addPatchesFlag()
			return c
		},
	}
}

// firmwareRun tracks relevant info for a given `try firmware` run.
type firmwareRun struct {
	tryRunBase
}

// Run provides the logic for a `try firmware` command run.
func (f *firmwareRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	f.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	f.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	ctx := context.Background()
	// Need to call run first to do LUCI auth / set up other shared constructs.
	if ret, err := f.run(ctx); err != nil {
		f.LogErr(err.Error())
		return ret
	}
	if err := f.validate(ctx); err != nil {
		f.LogErr(err.Error())
		return CmdError
	}

	if len(f.patches) > 0 {
		f.bbAddArgs = patchListToBBAddArgs(f.patches)
	}
	if err := f.runFirmwareBuilder(ctx); err != nil {
		f.LogErr(err.Error())
		return CmdError
	}
	return Success
}

// validate validates firmware-specific args for the command.
func (f *firmwareRun) validate(ctx context.Context) error {
	if f.branch == "" {
		return errors.New("must provide a firmware branch with --branch")
	}
	if !strings.HasPrefix(f.branch, "firmware-") || !strings.HasSuffix(f.branch, ".B") {
		return fmt.Errorf("provided branch does not look like a firmware branch: %s", f.branch)
	}
	if builderExists, err := f.doesFWBranchHaveBuilder(ctx, f.branch, !f.production); err != nil {
		return err
	} else if !builderExists {
		return fmt.Errorf("firmware builder does not seem to exist for branch %s", f.branch)
	}
	if err := f.tryRunBase.validate(); err != nil {
		return err
	}
	return nil
}

// doesFWBranchHaveBuilder checks whether the given branch has a firmware builder configured.
func (f *firmwareRun) doesFWBranchHaveBuilder(ctx context.Context, branch string, staging bool) (bool, error) {
	bucket := "firmware"
	if staging {
		bucket = "staging"
	}
	allFWBuilders, err := f.bbClient.BBBuilders(ctx, bucket)
	if err != nil {
		return false, errors.Annotate(err, "querying bb for firmware builders").Err()
	}
	return sliceContainsStr(allFWBuilders, getFWBuilderFullName(branch, staging)), nil
}

// getFWBuilderFullName finds the full name (<project>/<bucket>/<builder>) for the given firmware branch.
func getFWBuilderFullName(branch string, staging bool) string {
	var bucket, stagingPrefix string
	if staging {
		bucket = "staging"
		stagingPrefix = "staging-"
	} else {
		bucket = "firmware"
	}
	return fmt.Sprintf("chromeos/%s/%s%s-branch", bucket, stagingPrefix, branch)
}

// runFWBuilder creates a firmware build via `bb add`, and reports it to the user.
func (f *firmwareRun) runFirmwareBuilder(ctx context.Context) error {
	builderName := getFWBuilderFullName(f.branch, !f.production)
	return f.bbClient.BBAdd(ctx, f.dryrun, append([]string{builderName}, f.bbAddArgs...)...)
}
