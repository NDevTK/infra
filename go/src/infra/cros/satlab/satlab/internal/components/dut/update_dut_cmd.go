// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"infra/cmdsupport/cmdlib"
	"infra/cros/satlab/common/dut"
	"infra/cros/satlab/common/site"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
)

// UpdateDUTCmd is the command that updates fields for a satlab DUT.
var UpdateDUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Update a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := &updateDUT{}
		registerUpdateShivasFlags(c)
		return c
	},
}

// UpdateDUT is the 'satlab update dut' command. Its fields are the command line arguments.
type updateDUT struct {
	subcommands.CommandRunBase

	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	dut.UpdateDUT
}

// Run is the main entrypoint to 'satlab update dut'.
func (c *updateDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of 'satlab update dut'.
func (c *updateDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
  c.SatlabID = c.commonFlags.SatlabID
  c.Namespace = c.envFlags.GetNamespace()

	ctx := context.Background()
	return c.TriggerRun(ctx)
}
