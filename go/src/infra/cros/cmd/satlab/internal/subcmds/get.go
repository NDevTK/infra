// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcmds

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cros/cmd/satlab/internal/commands/dns"
	"infra/cros/cmd/satlab/internal/components/dut"
	"infra/cros/cmd/satlab/internal/site"
)

// GetBase is a placeholder command for "get".
type getBase struct {
	subcommands.CommandRunBase
}

// GetCmd is a placeholder command for get.
var GetCmd = &subcommands.Command{
	UsageLine: "get <sub-command>",
	CommandRun: func() subcommands.CommandRun {
		c := &getBase{}
		return c
	},
}

// GetApp is an application tha tholds the get subcommands.
type getApp struct {
	cli.Application
}

// GetName fulfills the cli.Application interface's method call which lets us print the correct usage
// alternatively we could define another Application with the `satlab get` name like in the subcommands
// https://github.com/maruel/subcommands/blob/main/sample-complex/ask.go#L13
func (c getApp) GetName() string {
	return fmt.Sprintf("%s get", site.AppPrefix)
}

// Run transfers control to a subcommand.
func (c *getBase) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&getApp{*d}, args)
}

// GetCommands lists the available subcommands.
func (c getApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
		dut.GetDUTCmd,
		dns.GetDNSCmd,
	}
}
