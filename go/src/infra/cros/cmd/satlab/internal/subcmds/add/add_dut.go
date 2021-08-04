// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"github.com/maruel/subcommands"

	"infra/cmdsupport/cmdlib"
)

var DUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Deploy a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := makeDefaultShivasCommand()
		c.Flags.StringVar(&c.address, "address", "", "IP address of host")
		c.Flags.BoolVar(&c.skipDNS, "skipDNS", false, "whether to skip updating the DNS")
		registerShivasFlags(c)
		return c
	},
}

type addDUT struct {
	shivasAddDUT
	// Satlab-specific fields, if any exist, go here.
	address string
	skipDNS bool
}

func (c *addDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *addDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return Add(c, a, args)
}
