// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package subcmds

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cros/cmd/satlab/internal/components/run"
	"infra/cros/cmd/satlab/internal/site"
)

// RunBase is the placeholder for the run command.
type runBase struct {
	subcommands.CommandRunBase
}

// RunCmd contains the usage and implementation for the run command.
var RunCmd = &subcommands.Command{
	UsageLine: "run <sub-command>",
	CommandRun: func() subcommands.CommandRun {
		c := &runBase{}
		return c
	},
}

// runApp is an application for the run commands. Control is transferred here
// when consuming the "run" subcommand.
type runApp struct {
	cli.Application
}

// GetName fulfills the cli.Application interface's method call which lets us print the correct usage
// alternatively we could define another Application with the `satlab get` name like in the subcommands
// https://github.com/maruel/subcommands/blob/main/sample-complex/ask.go#L13
func (c runApp) GetName() string {
	return fmt.Sprintf("%s run", site.AppPrefix)
}

// Run transfers control to the subcommands of run.
func (c *runBase) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	d := a.(*cli.Application)
	return subcommands.Run(&runApp{*d}, args)
}

// GetCommands lists the subcommands of run.
func (c runApp) GetCommands() []*subcommands.Command {
	return []*subcommands.Command{
		subcommands.CmdHelp,
		run.RunSuiteCmd,
	}
}
