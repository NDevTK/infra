// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package nic

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/flag"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// UpdateNicCmd Update nic by given name.
var UpdateNicCmd = &subcommands.Command{
	UsageLine: "nic [Options...]",
	ShortDesc: "Update a nic on a machine",
	LongDesc:  cmdhelp.UpdateNicLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateNic{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.NicFileText)
		c.Flags.BoolVar(&c.interactive, "i", false, "enable interactive mode for input")

		c.Flags.StringVar(&c.machineName, "machine", "", "name of the machine to associate the nic")
		c.Flags.StringVar(&c.nicName, "name", "", "the name of the nic to update")
		c.Flags.StringVar(&c.macAddress, "mac", "", "the mac address of the nic to add."+cmdhelp.ClearFieldHelpText)
		c.Flags.StringVar(&c.state, "state", "", cmdhelp.StateHelp)
		c.Flags.StringVar(&c.switchName, "switch", "", "the name of the switch that this nic is connected to. "+cmdhelp.ClearFieldHelpText)
		c.Flags.StringVar(&c.switchPort, "switch-port", "", "the port of the switch that this nic is connected to. "+cmdhelp.ClearFieldHelpText)
		c.Flags.Var(flag.StringSlice(&c.tags), "tag", "Name(s) of tag(s). Can be specified multiple times. "+cmdhelp.ClearFieldHelpText)
		return c
	},
}

type updateNic struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile string
	interactive  bool

	machineName string
	nicName     string
	macAddress  string
	state       string
	switchName  string
	switchPort  string
	tags        []string
}

func (c *updateNic) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateNic) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ns, err := c.envFlags.Namespace(nil, "")
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
	var nic ufspb.Nic
	if c.interactive {
		utils.GetNicInteractiveInput(ctx, ic, &nic, true)
	} else {
		if c.newSpecsFile != "" {
			if err = utils.ParseJSONFile(c.newSpecsFile, &nic); err != nil {
				return err
			}
			if nic.GetMachine() == "" {
				return errors.New(fmt.Sprintf("machine field is empty in json. It is a required parameter for json input."))
			}
		} else {
			c.parseArgs(&nic)
		}
	}
	if err := utils.PrintExistingNic(ctx, ic, nic.Name); err != nil {
		return err
	}
	nic.Name = ufsUtil.AddPrefix(ufsUtil.NicCollection, nic.Name)
	if !ufsUtil.ValidateTags(nic.Tags) {
		return fmt.Errorf(ufsAPI.InvalidTags)
	}
	res, err := ic.UpdateNic(ctx, &ufsAPI.UpdateNicRequest{
		Nic: &nic,
		UpdateMask: utils.GetUpdateMask(&c.Flags, map[string]string{
			"machine":     "machine",
			"mac":         "macAddress",
			"state":       "resourceState",
			"switch":      "switch",
			"switch-port": "portName",
			"tag":         "tags",
		}),
	})
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	fmt.Println("The nic after update:")
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Printf("Successfully updated the nic %s\n", res.Name)
	return nil
}

func (c *updateNic) parseArgs(nic *ufspb.Nic) {
	nic.Name = c.nicName
	if c.macAddress == utils.ClearFieldValue {
		nic.MacAddress = ""
	} else {
		nic.MacAddress = c.macAddress
	}
	nic.ResourceState = ufsUtil.ToUFSState(c.state)
	nic.SwitchInterface = &ufspb.SwitchInterface{}
	nic.Machine = c.machineName
	if c.switchName == utils.ClearFieldValue {
		nic.GetSwitchInterface().Switch = ""
	} else {
		nic.GetSwitchInterface().Switch = c.switchName
	}
	if c.switchPort == utils.ClearFieldValue {
		nic.GetSwitchInterface().PortName = ""
	} else {
		nic.GetSwitchInterface().PortName = c.switchPort
	}
	if ufsUtil.ContainsAnyStrings(c.tags, utils.ClearFieldValue) {
		nic.Tags = nil
	} else {
		nic.Tags = c.tags
	}
}

func (c *updateNic) validateArgs() error {
	if c.newSpecsFile != "" && c.interactive {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive & JSON mode cannot be specified at the same time.")
	}
	if c.newSpecsFile != "" || c.interactive {
		if c.nicName != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-name' cannot be specified at the same time.")
		}
		if c.switchName != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-switchName' cannot be specified at the same time.")
		}
		if c.switchPort != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-switch-port' cannot be specified at the same time.")
		}
		if c.macAddress != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-mac' cannot be specified at the same time.")
		}
		if len(c.tags) > 0 {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-tag' cannot be specified at the same time.")
		}
		if c.machineName != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive/JSON mode is specified. '-machine' cannot be specified at the same time.")
		}
		if c.state != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe JSON input file is already specified. '-state' cannot be specified at the same time.")
		}
	}
	if c.newSpecsFile == "" && !c.interactive {
		if c.nicName == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required, no mode ('-f' or '-i') is specified.")
		}
		if c.machineName == "" && c.switchName == "" && c.switchPort == "" && c.macAddress == "" && len(c.tags) == 0 && c.state == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNothing to update. Please provide any field to update")
		}
		if c.state != "" && !ufsUtil.IsUFSState(ufsUtil.RemoveStatePrefix(c.state)) {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n%s is not a valid state, please check help info for '-state'.", c.state)
		}
	}
	return nil
}
