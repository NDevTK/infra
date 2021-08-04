// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package get

import (
	"errors"

	"github.com/maruel/subcommands"

	"infra/cmdsupport/cmdlib"
	// "infra/cros/cmd/satlab/internal/site"
)

var DUTcmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Get a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := makeDefaultShivasCommand()
		registerShivasFlags(c)
		// site.RegisterCommonFlags(&c.CommonFlags, c.GetFlags())
		return c
	},
}

type getDUT struct {
	shivasGetDUT
	// Satlab-specific fields, if any exist, go here.
}

func (c *getDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *getDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return errors.New("not yet implemented")
}
