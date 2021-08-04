// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcmds

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cros/cmd/satlab/internal/subcmds/delete"
)

type deleteBase struct {
	subcommands.CommandRunBase
}

var DeleteCmd = &subcommands.Command{
	UsageLine: "delete <sub-command>",
	CommandRun: func() subcommands.CommandRun {
		c := &deleteBase{}
		return c
	},
}

type deleteApp struct {
	cli.Application
}

func (c *deleteBase) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&deleteApp{*d}, args)
}

func (c deleteApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
		delete.DUTCmd,
	}
}
