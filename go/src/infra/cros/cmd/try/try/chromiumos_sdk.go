// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package try

import (
	"context"
	"fmt"
	"log"
	"os"

	"infra/cros/internal/cmd"
	bb "infra/cros/lib/buildbucket"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
)

func GetCmdChromiumOSSDK() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "chromiumos_sdk --branch BRANCH [flags]",
		ShortDesc: "Run a ChromiumOS SDK builder.",
		CommandRun: func() subcommands.CommandRun {
			c := &chromiumOSSDKRun{}
			c.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.addPatchesFlag()
			c.addProductionFlag()
			return c
		},
	}
}

// chromiumOSSDKRun tracks relevant info for a given `try chromiumos_sdk` run.
type chromiumOSSDKRun struct {
	tryRunBase
	propsFile *os.File
}

// Run provides the logic for a `try chromiumos_sdk` command run.
func (r *chromiumOSSDKRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	ctx := context.Background()

	// Need to call run first to do LUCI auth / set up other shared constructs.
	if ret, err := r.run(ctx); err != nil {
		r.LogErr(err.Error())
		return ret
	}

	if err := r.validate(ctx); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	propsStruct, err := r.bbClient.GetBuilderInputProps(ctx, r.getBuilderFullName())
	fmt.Println(r.getBuilderFullName())
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
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
	if err := bb.WriteStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	if len(r.patches) > 0 {
		r.bbAddArgs = append(r.bbAddArgs, patchListToBBAddArgs(r.patches)...)
	}
	if err := r.runBuilder(ctx); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	return Success
}

// validate validates args for the command.
func (r *chromiumOSSDKRun) validate(ctx context.Context) error {
	return r.tryRunBase.validate()
}

// getBuilderFullName finds the full builder name (<project>/<bucket>/<builder>).
func (r *chromiumOSSDKRun) getBuilderFullName() string {
	var bucket, stagingPrefix string
	if r.production {
		bucket = "infra"
	} else {
		bucket = "staging"
		stagingPrefix = "staging-"
	}
	return fmt.Sprintf("chromeos/%s/%sbuild-chromiumos-sdk", bucket, stagingPrefix)
}

// runBuilder creates a ChromiumOS SDK build via `bb add`, and reports it to the user.
func (r *chromiumOSSDKRun) runBuilder(ctx context.Context) error {
	builderName := r.getBuilderFullName()
	_, err := r.bbClient.BBAdd(ctx, r.dryrun, append([]string{builderName}, r.bbAddArgs...)...)
	return err
}
