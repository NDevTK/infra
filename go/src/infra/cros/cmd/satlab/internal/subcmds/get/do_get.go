// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package get

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/maruel/subcommands"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/common"
	"infra/cros/cmd/satlab/internal/paths"
)

// Get gets information about a DUT associated with the current satlab.
func Get(c *getDUT, a subcommands.Application, positionalArgs []string) error {
	// 'shivas get dut' will list all DUTs everywhere.
	// This command takes a while to execute and gives no immediate feedback, so provide an error message to the user.
	if len(positionalArgs) == 0 {
		// TODO(gregorynisbet): pick a default behavior for get DUT.
		return errors.New(`default "get dut" functionality not implemented`)
	}

	// No flags need to be annotated with the satlab prefix for get dut.
	// However, the positional arguments need to have the satlab prefix
	// prepended.
	for i, item := range positionalArgs {
		positionalArgs[i] = common.MaybePrepend(c.commonFlags.SatlabID, item)
	}
	flags := makeShivasFlags(c)
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasPath, "get", "dut"},
		Flags:          flags,
		PositionalArgs: positionalArgs,
	}).ToCommand()
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	out, err := command.Output()
	fmt.Printf("%s\n", string(out))
	return err
}
