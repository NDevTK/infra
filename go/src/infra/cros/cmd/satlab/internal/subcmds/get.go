// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcmds

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
)

type getBase struct {
	subcommands.CommandRunBase
}

var GetCmd = &subcommands.Command{
	UsageLine: "get <sub-command>",
	CommandRun: func() subcommands.CommandRun {
		c := &addBase{}
		return c
	},
}

type getApp struct {
	cli.Application
}

func (c *getBase) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&getApp{*d}, args)
}

func (c getApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
	}
}
