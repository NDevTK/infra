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

	"infra/cros/internal/cmd"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
)

func getCmdRetry() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "retry [flags]",
		ShortDesc: "Rerun the specified (release) build.",
		CommandRun: func() subcommands.CommandRun {
			c := &retryRun{}
			c.tryRunBase.cmdRunner = cmd.RealCommandRunner{}
			c.addDryrunFlag()
			c.Flags.StringVar(&c.originalBBID, "bbid", "", "Buildbucket ID of the builder to retry.")
			if flag.NArg() > 1 && flag.Args()[1] == "help" {
				fmt.Printf("Run `cros try help` or `cros try help ${subcomand}` for help.")
				os.Exit(0)
			}
			return c
		},
	}
}

// retryRun tracks relevant info for a given `try retry` run.
type retryRun struct {
	tryRunBase
	originalBBID string
	// Used for testing purposes. If set, props will be written to this file
	// rather than a temporary one.
	propsFile *os.File
}

// validate validates retry-specific args for the command.
func (r *retryRun) validate() error {
	return nil
}

// Run provides the logic for a `try retry` command run.
func (r *retryRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
	r.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	r.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	if err := r.validate(); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	ctx := context.Background()
	if ret, err := r.run(ctx); err != nil {
		r.LogErr(err.Error())
		return ret
	}

	buildData, err := r.GetBuild(ctx, r.originalBBID)
	if err != nil {
		r.LogErr(err.Error())
		return CmdError
	}
	builder := buildData.GetBuilder()
	propsStruct := buildData.GetInput().GetProperties()

	// Set up propsFile.
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

	// Write props to file and launch builder.
	if err := writeStructToFile(propsStruct, propsFile); err != nil {
		r.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if r.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	r.bbAddArgs = append(r.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	builderName := fmt.Sprintf("%s/%s/%s", builder.GetProject(), builder.GetBucket(), builder.GetBuilder())
	if err := r.BBAdd(ctx, append([]string{builderName}, r.bbAddArgs...)...); err != nil {
		r.LogErr(err.Error())
		return CmdError
	}

	return Success
}
