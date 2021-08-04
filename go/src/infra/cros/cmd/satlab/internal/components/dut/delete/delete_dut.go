// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package delete

import (
	"github.com/maruel/subcommands"

	"infra/cmdsupport/cmdlib"
)

// DUTCmd is the implementation of the "satlab delete DUT" command.
var DUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Delete a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := makeDefaultShivasCommand()
		registerShivasFlags(c)
		return c
	},
}

// DeleteDUT holds the arguments that are needed for the delete DUT command.
type deleteDUT struct {
	shivasDeleteDUT
	// Satlab-specific fields, if any exist, go here.
}

// Run attempts to delete a DUT and returns an exit status.
func (c *deleteDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of the delete command.
func (c *deleteDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return Delete(c, a, args)
}
