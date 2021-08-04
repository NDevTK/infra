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

	"infra/cmdsupport/cmdlib"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/common"
	"infra/cros/cmd/satlab/internal/paths"
)

// DUTCmd is the implementation of "satlab get dut ...".
var DUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Get a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := makeDefaultShivasCommand()
		registerShivasFlags(c)
		return c
	},
}

// GetDUT holds the arguments for "satlab get dut ...".
type getDUT struct {
	shivasGetDUT
	// Satlab-specific fields, if any exist, go here.
}

// Run runs the get DUT subcommand.
func (c *getDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *getDUT) innerRun(a subcommands.Application, positionalArgs []string, env subcommands.Env) error {
	// 'shivas get dut' will list all DUTs everywhere.
	// This command takes a while to execute and gives no immediate feedback, so provide an error message to the user.
	if len(positionalArgs) == 0 {
		// TODO(gregorynisbet): pick a default behavior for get DUT.
		return errors.New(`default "get dut" functionality not implemented`)
	}

	if c.commonFlags.SatlabID == "" {
		c.commonFlags.SatlabID, _ = commands.GetDockerHostBoxIdentifier()
	}

	// No flags need to be annotated with the satlab prefix for get dut.
	// However, the positional arguments need to have the satlab prefix
	// prepended.
	for i, item := range positionalArgs {
		positionalArgs[i] = common.MaybePrepend("satlab", c.commonFlags.SatlabID, item)
	}
	flags := makeShivasFlags(c)
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "dut"},
		Flags:          flags,
		PositionalArgs: positionalArgs,
	}).ToCommand()
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	out, err := command.Output()
	fmt.Printf("%s\n", string(out))
	return err
}
