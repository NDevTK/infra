// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcmds

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cros/cmd/satlab/internal/servod"
)

// servodBase is the type for the servod placeholder command.
type servodBase struct {
	subcommands.CommandRunBase
}

// Run transfers control to the servod subcommands.
func (c *servodBase) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&servodApp{*d}, args)
}

// ServodCmd is the servod placeholder command.
var ServodCmd = &subcommands.Command{
	UsageLine: "servod <sub-command>",
	ShortDesc: "Commands related to servod containers",
	CommandRun: func() subcommands.CommandRun {
		c := &servodBase{}
		return c
	},
}

// ServodApp is the placeholder application for the servod command.
type servodApp struct {
	cli.Application
}

// GetCommands lists the servod subcommands.
func (c servodApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
		servod.StartServodCmd,
	}
}
