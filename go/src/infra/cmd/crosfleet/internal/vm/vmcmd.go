// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vm

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
)

const vmCmdName = "vm"

var vmApplication = &cli.Application{
	Name:  fmt.Sprintf("crosfleet %s", vmCmdName),
	Title: "Interact with VMs.",
	Commands: []*subcommands.Command{
		abandon,
		lease,
		leases,
		subcommands.CmdHelp,
	},
}

// CmdVm is the parent command for all `crosfleet vm <subcommand>` commands.
var CmdVm = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s <subcommand>", vmCmdName),
	ShortDesc: "Creates and manages VMs leased by the current user.",
	LongDesc: fmt.Sprintf(`Creates and manages VMs leased by the current user.

Run 'crosfleet %s' to see list of all subcommands.`, vmCmdName),
	CommandRun: func() subcommands.CommandRun {
		c := &vmCmdRun{}
		return c
	},
}

type vmCmdRun struct {
	subcommands.CommandRunBase
}

func (c *vmCmdRun) Run(a subcommands.Application, args []string, _ subcommands.Env) int {
	return subcommands.Run(vmApplication, args)
}
