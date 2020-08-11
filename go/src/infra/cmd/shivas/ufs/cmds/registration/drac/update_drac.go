// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package drac

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
	ufspb "infra/unifiedfleet/api/v1/proto"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// UpdateDracCmd Update drac by given name.
var UpdateDracCmd = &subcommands.Command{
	UsageLine: "update-drac [Options...]",
	ShortDesc: "Update a drac by name",
	LongDesc:  cmdhelp.UpdateDracLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateDrac{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DracFileText)
		c.Flags.BoolVar(&c.interactive, "i", false, "enable interactive mode for input")

		c.Flags.StringVar(&c.machineName, "machine", "", "name of the machine to update the association of the drac to a different machine")
		c.Flags.StringVar(&c.dracName, "name", "", "name of the drac to update")
		c.Flags.StringVar(&c.macAddress, "mac-address", "", "the mac address of the drac to add.")
		c.Flags.StringVar(&c.switchName, "switch", "", "the name of the switch that this drac is connected to. "+cmdhelp.ClearFieldHelpText)
		c.Flags.IntVar(&c.switchPort, "switch-port", 0, "the port of the switch that this drac is connected to. "+"To clear this field set it to -1.")
		c.Flags.StringVar(&c.tags, "tags", "", "comma separated tags. You can only append/add new tags here. "+cmdhelp.ClearFieldHelpText)
		c.Flags.StringVar(&c.vlanName, "vlan", "", "the vlan to assign the drac to")
		c.Flags.BoolVar(&c.deleteVlan, "delete-vlan", false, "if deleting the ip assignment for the drac")
		c.Flags.StringVar(&c.ip, "ip", "", "the ip to assign the drac to")
		return c
	},
}

type updateDrac struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	newSpecsFile string
	interactive  bool
	commonFlags  site.CommonFlags

	machineName string
	dracName    string
	macAddress  string
	switchName  string
	switchPort  int
	tags        string
	vlanName    string
	deleteVlan  bool
	ip          string
}

func (c *updateDrac) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateDrac) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
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
	var drac ufspb.Drac
	if c.interactive {
		c.machineName = utils.GetDracInteractiveInput(ctx, ic, &drac, true)
	} else {
		if c.newSpecsFile != "" {
			if err = utils.ParseJSONFile(c.newSpecsFile, &drac); err != nil {
				return err
			}
		} else {
			c.parseArgs(&drac)
		}
	}
	drac.Name = ufsUtil.AddPrefix(ufsUtil.DracCollection, drac.Name)
	res, err := ic.UpdateDrac(ctx, &ufsAPI.UpdateDracRequest{
		Drac:    &drac,
		Machine: c.machineName,
		NetworkOption: &ufsAPI.NetworkOption{
			Vlan:   c.vlanName,
			Delete: c.deleteVlan,
			Ip:     c.ip,
		},
		UpdateMask: utils.GetUpdateMask(&c.Flags, map[string]string{
			"machine":     "machine",
			"mac-address": "macAddress",
			"switch":      "switch",
			"switch-port": "port",
			"tags":        "tags",
		}),
	})
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	utils.PrintProtoJSON(res)
	if c.deleteVlan {
		fmt.Printf("Successfully deleted vlan of drac %s\n", res.Name)
	}
	if c.vlanName != "" {
		// Log the assigned IP
		if dhcp, err := ic.GetDHCPConfig(ctx, &ufsAPI.GetDHCPConfigRequest{
			Hostname: res.Name,
		}); err == nil {
			utils.PrintProtoJSON(dhcp)
			fmt.Println("Successfully added dhcp config to drac: ", res.Name)
		}
	}
	return nil
}

func (c *updateDrac) parseArgs(drac *ufspb.Drac) {
	drac.Name = c.dracName
	drac.MacAddress = c.macAddress
	drac.SwitchInterface = &ufspb.SwitchInterface{}
	if c.switchName == utils.ClearFieldValue {
		drac.GetSwitchInterface().Switch = ""
	} else {
		drac.GetSwitchInterface().Switch = c.switchName
	}
	if c.switchPort == -1 {
		drac.GetSwitchInterface().Port = 0
	} else {
		drac.GetSwitchInterface().Port = int32(c.switchPort)
	}
	if c.tags == utils.ClearFieldValue {
		drac.Tags = nil
	} else {
		drac.Tags = utils.GetStringSlice(c.tags)
	}
}

func (c *updateDrac) validateArgs() error {
	if c.newSpecsFile != "" && c.interactive {
		return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe interactive & JSON mode cannot be specified at the same time.")
	}
	if c.newSpecsFile == "" && !c.interactive {
		if c.dracName == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required, no mode ('-f' or '-i') is specified.")
		}
		if c.vlanName == "" && !c.deleteVlan && c.ip == "" &&
			c.machineName == "" && c.switchName == "" && c.switchPort == 0 &&
			c.macAddress == "" && c.tags == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNothing to update. Please provide any field to update")
		}
	}
	return nil
}
