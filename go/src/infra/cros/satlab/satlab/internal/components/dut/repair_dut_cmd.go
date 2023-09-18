// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"

	"infra/cmdsupport/cmdlib"
	"infra/cros/satlab/common/dut/shivas"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
)

// RepairDUTCmd is the command that repairs a satlab DUT.
var RepairDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Repair a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := &repairDUTCmd{}
		registerRepairShivasFlags(c)
		return c
	},
}

// RepairDUT is the 'satlab repair dut' command. Its fields are the command line arguments.
type repairDUTCmd struct {
	shivasRepairDUT

	// Deep repair
	Deep bool
}

// Run is the main entrypoint to 'satlab repair dut'.
func (c *repairDUTCmd) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of 'satlab repair {dut}'.
func (c *repairDUTCmd) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	dockerHostBoxIdentifier, err := getDockerHostBoxIdentifier(c.commonFlags)
	if err != nil {
		return errors.Annotate(err, "repair dut").Err()
	}

	qualifiedHostname := site.MaybePrepend(site.Satlab, dockerHostBoxIdentifier, args[0])
	action := shivas.Verify
	if c.Deep {
		action = shivas.DeepRepair
	}

	res, err := (&shivas.DUTRepairer{
		Name:     qualifiedHostname,
		Executor: &executor.ExecCommander{},
	}).Repair(context.Background(), action)

	fmt.Printf("Build Link: %v\n### Batch tasks URL ###\nTask Link: %v\n", res.BuildLink, res.TaskLink)

	return err
}
