// Copyright 2024 The Chromium Authors
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
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/common/errors"
)

// TODO(b/318522770): Use the production builder.
const keyManagerBuilderName = "chromeos/staging/staging-key-manager"

func GetCmdCreatePreMPKeys(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "create_premp_keys --board BOARD [flags]",
		ShortDesc: "Create PreMP Keys for the given build target.",
		CommandRun: func() subcommands.CommandRun {
			c := &createPreMPKeysRun{}
			c.cmdRunner = cmd.RealCommandRunner{}
			c.tryRunBase.authOpts = authOpts
			c.addDryrunFlag()
			c.Flags.StringVar(&c.buildTarget, "build_target", "", "build targets to create keys for")
			return c
		},
	}
}

// firmwareRun tracks relevant info for a given `try firmware` run.
type createPreMPKeysRun struct {
	tryRunBase
	propsFile   *os.File
	buildTarget string
}

// Run provides the logic for a `try firmware` command run.
func (f *createPreMPKeysRun) Run(_ subcommands.Application, _ []string, _ subcommands.Env) int {
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

	propsStruct, err := f.bbClient.GetBuilderInputProps(ctx, keyManagerBuilderName)
	if err != nil {
		f.LogErr(err.Error())
		return CmdError
	}

	// TODO(b/318522770): Set create_premp_keys_requests input property.

	var propsFile *os.File
	if f.propsFile != nil {
		propsFile = f.propsFile
	} else {
		propsFile, err = os.CreateTemp("", "input_props")
		if err != nil {
			f.LogErr(err.Error())
			return CmdError
		}
	}
	if err := bb.WriteStructToFile(propsStruct, propsFile); err != nil {
		f.LogErr(errors.Annotate(err, "writing input properties to tempfile").Err().Error())
		return UnspecifiedError
	}
	if f.propsFile == nil {
		defer os.Remove(propsFile.Name())
	}
	f.bbAddArgs = append(f.bbAddArgs, "-p", fmt.Sprintf("@%s", propsFile.Name()))

	if err := f.runKeyManagerBuilder(ctx); err != nil {
		f.LogErr(err.Error())
		return CmdError
	}
	return Success
}

// validate validates firmware-specific args for the command.
func (f *createPreMPKeysRun) validate(ctx context.Context) error {
	if f.buildTarget == "" {
		return errors.New("must provide a build target with --build_target")
	}
	if err := f.tryRunBase.validate(); err != nil {
		return err
	}
	return nil
}

// runFWBuilder creates a firmware build via `bb add`, and reports it to the user.
func (f *createPreMPKeysRun) runKeyManagerBuilder(ctx context.Context) error {
	_, err := f.bbClient.BBAdd(ctx, f.dryrun, append([]string{keyManagerBuilderName}, f.bbAddArgs...)...)
	return err
}
