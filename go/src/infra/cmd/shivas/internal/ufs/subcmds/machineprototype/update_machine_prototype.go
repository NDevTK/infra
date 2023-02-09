// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package machineprototype

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// UpdateMachineLSEPrototypeCmd update MachineLSEPrototype by given name.
var UpdateMachineLSEPrototypeCmd = &subcommands.Command{
	UsageLine: "machine-prototype",
	ShortDesc: "Update prototype for a host",
	LongDesc:  cmdhelp.UpdateMachineLSEPrototypeLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateMachineLSEPrototype{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.MachineLSEPrototypeFileText)
		c.Flags.BoolVar(&c.interactive, "i", false, "enable interactive mode for input")
		return c
	},
}

type updateMachineLSEPrototype struct {
	subcommands.CommandRunBase
	authFlags    authcli.Flags
	envFlags     site.EnvFlags
	newSpecsFile string
	interactive  bool
}

func (c *updateMachineLSEPrototype) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateMachineLSEPrototype) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace(site.AllNamespaces, "")
	if err != nil {
		return err
	}
	ctx = utils.SetupContext(ctx, ns)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	fmt.Printf("Using UnifiedFleet service %s\n", e.UnifiedFleetService)
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	var machinelsePrototype ufspb.MachineLSEPrototype
	if c.interactive {
		utils.GetMachinelsePrototypeInteractiveInput(ctx, ic, &machinelsePrototype, true)
	} else {
		err = utils.ParseJSONFile(c.newSpecsFile, &machinelsePrototype)
		if err != nil {
			return err
		}
	}
	if err := utils.PrintExistingMachinePrototype(ctx, ic, machinelsePrototype.Name); err != nil {
		return err
	}
	machinelsePrototype.Name = ufsUtil.AddPrefix(ufsUtil.MachineLSEPrototypeCollection, machinelsePrototype.Name)
	res, err := ic.UpdateMachineLSEPrototype(ctx, &ufsAPI.UpdateMachineLSEPrototypeRequest{
		MachineLSEPrototype: &machinelsePrototype,
	})
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	fmt.Println("The machine lse prototype after update:")
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Println()
	return nil
}

func (c *updateMachineLSEPrototype) validateArgs() error {
	if !c.interactive && c.newSpecsFile == "" {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNeither JSON input file specified nor in interactive mode to accept input.")
	}
	return nil
}
