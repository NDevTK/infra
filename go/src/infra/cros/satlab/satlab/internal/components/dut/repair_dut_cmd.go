// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"infra/cmdsupport/cmdlib"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/satlab/internal/components/dut/shivas"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
)

// RepairDUTCmd is the command that repairs a satlab DUT.
var RepairDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Repair a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := &repairDUT{}
		registerRepairShivasFlags(c)
		return c
	},
}

// RepairDUT is the 'satlab repair dut' command. Its fields are the command line arguments.
type repairDUT struct {
	shivasRepairDUT
}

// Run is the main entrypoint to 'satlab repair dut'.
func (c *repairDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of 'satlab repair {dut}'.
func (c *repairDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	dockerHostBoxIdentifier, err := getDockerHostBoxIdentifier(ctx, c.commonFlags)
	if err != nil {
		return errors.Annotate(err, "repair dut").Err()
	}

	qualifiedHostname := site.MaybePrepend(site.Satlab, dockerHostBoxIdentifier, args[0])

	return (&shivas.DUTRepairer{
		Name: qualifiedHostname,
		// ShivasArgs: makeRepairShivasFlags(c), # no additional flags to shivas for now #
	}).Repair()
}
