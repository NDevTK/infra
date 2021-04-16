// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// RenameDUTCmd rename dut by given name.
var RenameDUTCmd = &subcommands.Command{
	UsageLine: "dut ...",
	ShortDesc: "Rename dut with new name",
	LongDesc: `Rename dut with new name.

Example:

shivas rename dut -name {oldName} -new-name {newName}

Renames the dut and prints the output in the user-specified format.

WARNING: Ensure that the dut is in required state after renaming it.
`,
	CommandRun: func() subcommands.CommandRun {
		c := &renameDUT{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)
		c.outputFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.name, "name", "", "the name of the dut to rename")
		c.Flags.StringVar(&c.newName, "new-name", "", "the new name of the dut")
		return c
	},
}

type renameDUT struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags
	outputFlags site.OutputFlags

	name    string
	newName string
}

func (c *renameDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *renameDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace()
	if err != nil {
		return err
	}
	ctx = utils.SetupContext(ctx, ns)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UFS service %s\n", e.UnifiedFleetService)
	}
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	if err := utils.PrintExistingDUT(ctx, ic, c.name); err != nil {
		return err
	}
	_, err = ic.RenameMachineLSE(ctx, &ufsAPI.RenameMachineLSERequest{
		Name:    ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, c.name),
		NewName: ufsUtil.AddPrefix(ufsUtil.MachineLSECollection, c.newName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *renameDUT) validateArgs() error {
	if c.name == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required")
	}
	if c.newName == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-new-name' is required")
	}
	return nil
}
