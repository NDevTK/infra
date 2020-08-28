// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package query

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cmd/shivas/ufs/subcmds/chromeplatform"
	"infra/cmd/shivas/ufs/subcmds/drac"
	"infra/cmd/shivas/ufs/subcmds/host"
	"infra/cmd/shivas/ufs/subcmds/machine"
	"infra/cmd/shivas/ufs/subcmds/machineprototype"
	"infra/cmd/shivas/ufs/subcmds/nic"
	"infra/cmd/shivas/ufs/subcmds/rack"
	"infra/cmd/shivas/ufs/subcmds/rackprototype"
	"infra/cmd/shivas/ufs/subcmds/rpm"
	"infra/cmd/shivas/ufs/subcmds/switches"
	"infra/cmd/shivas/ufs/subcmds/vlan"
	"infra/cmd/shivas/ufs/subcmds/vm"
)

type list struct {
	subcommands.CommandRunBase
}

// ListCmd contains list command specification
var ListCmd = &subcommands.Command{
	UsageLine: "list <sub-command>",
	ShortDesc: "List entries for resources/entities",
	LongDesc: `List entries for:
	machine/rack/kvm/rpm/switch/drac/nic
	host/vm/vm-slots
	machine-prototype/rack-prototype/chromeplatform/vlan`,
	CommandRun: func() subcommands.CommandRun {
		c := &list{}
		return c
	},
}

type listApp struct {
	cli.Application
}

// Run implementing subcommands.CommandRun interface
func (c *list) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&listApp{*d}, args)
}

// GetCommands lists all the subcommands under list
func (c listApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
		machine.ListMachineCmd,
		host.ListHostCmd,
		rpm.ListRPMCmd,
		switches.ListSwitchCmd,
		drac.ListDracCmd,
		nic.ListNicCmd,
		vm.ListVMCmd,
		vm.ListVMSlotCmd,
		vlan.ListVlanCmd,
		rack.ListRackCmd,
		machineprototype.ListMachineLSEPrototypeCmd,
		rackprototype.ListRackLSEPrototypeCmd,
		chromeplatform.ListChromePlatformCmd,
	}
}

// GetName is cli.Application interface implementation
func (c listApp) GetName() string {
	return "list"
}
