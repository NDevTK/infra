// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package subcommands includes subcommand logic that will be used for the CLI
// front end.
package subcommands

import (
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/metrics"
	"infra/cros/cmd/suite_scheduler/run"
)

type runCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	prod        bool
	newBuilds   bool
	timedEvents bool
}

// setFlags adds also CLI flags to the subcommand.
func (c *runCommand) setFlags() {
	// TODO(b/319463660): Allow for execution time to be set for TIMED_EVENTS
	// TODO(b/319273179): Allow a path to local configs to be passed in via CLI
	// TODO(TBD): Allow for execution of only specified types of TIMED_EVENTS.
	// E.g. (DAILY | WEEKLY), (DAILY), (FORTNIGHTLY | WEEKLY), etc.

	c.Flags.BoolVar(&c.prod, "prod", false, "Run using prod environments.")

	// TODO(b/319464677): Implement a backfill run command

	c.Flags.BoolVar(&c.newBuilds, "new-builds", false, "Check for new build images and launch NEW_BUILD type suites.")
	c.Flags.BoolVar(&c.prod, "timed-events", false, "Launch TIMED_EVENT suites which are eligible to be triggered.")

}

func (c *runCommand) validate() error {

	if !c.newBuilds && !c.timedEvents {
		return fmt.Errorf("-new-builds or -timed-events must be specified")
	}

	return nil
}

// Run is the "main" function of the subcommand.
func (c *runCommand) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	common.Stdout.Println("Validating flags")
	err := c.validate()
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	common.Stdout.Printf("Running SuiteSchedulerV1.5... ")

	// Set the RunID for the entire execution run.
	err = metrics.SetSuiteSchedulerRunID()
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}
	common.Stdout.Printf("runID: %s\n", metrics.GetRunID().Id)

	// Launch execution path for NEW_BUILD type configs
	if c.newBuilds {
		common.Stdout.Println("Launching NEW_BUILDS")
		err := run.NewBuilds(&c.authFlags)
		if err != nil {
			common.Stderr.Println(err)
			return 1
		}
		common.Stdout.Println("Done launching NEW_BUILDS")
	}

	// Launch execution path for all TIMED_EVENT configs
	if c.timedEvents {
		common.Stdout.Println("Launching TIMED_EVENTS")
		err := run.TimedEvents()
		if err != nil {
			common.Stderr.Println(err)
			return 1
		}
		common.Stdout.Println("Done launching TIMED_EVENTS")
	}

	return 0

}

// GetRunCommand forms and returns the encapsulated Run subcommand for CLI use.
func GetRunCommand(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "run",
		LongDesc:  "The run command is used to launch full SuiteScheduler executions.",
		CommandRun: func() subcommands.CommandRun {
			cmd := &runCommand{}
			cmd.authFlags = authcli.Flags{}
			cmd.authFlags.Register(cmd.GetFlags(), authOpts)
			cmd.setFlags()
			return cmd
		},
	}
}
