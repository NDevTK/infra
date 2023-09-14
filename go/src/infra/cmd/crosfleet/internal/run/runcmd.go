// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/cmd/crosfleet/internal/common"
)

const runCmdName = "run"

var testApplication = &cli.Application{
	Name:  fmt.Sprintf("crosfleet %s", runCmdName),
	Title: "Run tests.",
	Commands: []*subcommands.Command{
		backfill,
		test,
		suite,
		testplan,
		subcommands.CmdHelp,
	},
}

// CmdRun is the parent command for all `crosfleet run <subcommand>` commands.
var CmdRun = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s <subcommand>", runCmdName),
	ShortDesc: "runs tests and other executable tasks on DUTs in ChromeOS hardware labs",
	LongDesc: fmt.Sprintf(`Runs individual tests, test suites, or custom test plan files, depending on the subcommand given.

Run 'crosfleet %s' to see list of all subcommands.`, runCmdName),
	CommandRun: func() subcommands.CommandRun {
		c := &runCmd{}
		return c
	},
}

type runCmd struct {
	subcommands.CommandRunBase
}

func (c *runCmd) Run(a subcommands.Application, args []string, _ subcommands.Env) int {
	status := subcommands.Run(testApplication, args)
	if status == 0 {
		common.WriteCrosfleetUIPromptStderr(args)
	}
	return status
}
