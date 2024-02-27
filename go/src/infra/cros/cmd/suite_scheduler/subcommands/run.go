// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package subcommands includes subcommand logic that will be used for the CLI
// front end.
package subcommands

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/cmd/suite_scheduler/common"
	"infra/cros/cmd/suite_scheduler/metrics"
	"infra/cros/cmd/suite_scheduler/pubsub"
	"infra/cros/cmd/suite_scheduler/run"
	"infra/cros/cmd/suite_scheduler/totmanager"
)

type runCommand struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	runID       string
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

	c.Flags.StringVar(&c.runID, "run-id", common.DefaultString, "Used to manually set the runID. Should only be used by the recipe builder.")

	c.Flags.BoolVar(&c.prod, "prod", false, "Run using prod environments.")

	// TODO(b/319464677): Implement a backfill run command

	c.Flags.BoolVar(&c.newBuilds, "new-builds", false, "Check for new build images and launch NEW_BUILD type suites.")
	c.Flags.BoolVar(&c.prod, "timed-events", false, "Launch TIMED_EVENT suites which are eligible to be triggered.")

}

// validate ensures that the provided flags are being used in an expected
// manner.
func (c *runCommand) validate() error {
	if !c.newBuilds && !c.timedEvents {
		return fmt.Errorf("-new-builds or -timed-events must be specified")
	}

	if totmanager.GetTot() == 0 {
		return fmt.Errorf("totmanager was not initialized")
	}

	return nil
}

// endRun ends the timer and publishes the message to pubsub.
func endRun() error {
	err := metrics.SetEndTime()
	if err != nil {
		return err
	}

	run := metrics.GenerateRunMessage()

	data, err := protojson.Marshal(run)
	if err != nil {
		return err
	}

	client, err := pubsub.InitPublishClient(context.Background(), common.StagingProjectID, common.RunsPubSubTopic)
	if err != nil {
		return err
	}

	err = client.PublishMessage(context.Background(), data)
	if err != nil {
		return err
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

	// Initialize the ToT Manager at the start of the run. If this isn't
	// initialized then no builds will be targeted as ToT will be set to 0 by
	// default.
	common.Stdout.Printf("Initializing ToTManager")
	err = totmanager.InitTotManager()
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}

	common.Stdout.Printf("Running SuiteSchedulerV1.5... ")

	// Set the RunID for the entire execution run.
	err = metrics.SetSuiteSchedulerRunID(c.runID)
	if err != nil {
		common.Stderr.Println(err)
		return 1
	}
	common.Stdout.Printf("runID: %s\n", metrics.GetRunID().Id)

	// Start the clock for the run metrics
	err = metrics.SetStartTime()
	if err != nil {
		// Stop run timer and publish the message to pubsub
		endRunErr := endRun()
		if endRunErr != nil {
			common.Stderr.Println(err)
		}

		common.Stderr.Println(err)
		return 1
	}

	// Launch execution path for NEW_BUILD type configs
	if c.newBuilds {
		common.Stdout.Println("Launching NEW_BUILDS")
		err := run.NewBuilds(&c.authFlags)
		if err != nil {
			// Stop run timer and publish the message to pubsub
			endRunErr := endRun()
			if endRunErr != nil {
				common.Stderr.Println(err)
			}

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
			// Stop run timer and publish the message to pubsub
			endRunErr := endRun()
			if endRunErr != nil {
				common.Stderr.Println(err)
			}

			common.Stderr.Println(err)
			return 1
		}
		common.Stdout.Println("Done launching TIMED_EVENTS")
	}

	// Stop run timer and publish the message to pubsub
	endRunErr := endRun()
	if endRunErr != nil {
		common.Stderr.Println(err)
		return 1
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
