// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab/internal/bb"
	skycmdlib "infra/cmd/skylab/internal/cmd/cmdlib"
	"infra/cmd/skylab/internal/cmd/recipe"
	"infra/cmd/skylab/internal/site"
	"infra/cmdsupport/cmdlib"
)

// CreateSuite subcommand: create a suite task.
var CreateSuite = &subcommands.Command{
	UsageLine: "create-suite [FLAGS...] SUITE_NAME",
	ShortDesc: "create a suite task [DEPRECATED--please use crosfleet (go/crosfleet-cli)]",
	LongDesc: `[DEPRECATED--please use crosfleet (go/crosfleet-cli)]

Create a suite task, with the given suite name.

You must supply -board, -image, and -pool.

This command does not wait for the task to start running.`,
	CommandRun: func() subcommands.CommandRun {
		c := &createSuiteRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.createRunCommon.Register(&c.Flags)
		c.Flags.BoolVar(&c.orphan, "orphan", false, "Deprecated, do not use.")
		c.Flags.BoolVar(&c.json, "json", false, "Format output as JSON")
		c.Flags.StringVar(&c.taskName, "task-name", "", "Optional name to be used for the Swarming task.")
		return c
	},
}

type createSuiteRun struct {
	subcommands.CommandRunBase
	createRunCommon
	authFlags authcli.Flags
	envFlags  skycmdlib.EnvFlags
	orphan    bool
	json      bool
	taskName  string
}

func (c *createSuiteRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *createSuiteRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}

	ctx := cli.GetContext(a, c, env)
	suiteName := c.Flags.Arg(0)

	return c.innerRunBB(ctx, a, suiteName)
}

func (c *createSuiteRun) validateArgs() error {
	if err := c.createRunCommon.ValidateArgs(c.Flags); err != nil {
		return err
	}

	if c.Flags.NArg() == 0 {
		return cmdlib.NewUsageError(c.Flags, "missing suite name")
	}

	if c.orphan {
		return errors.Reason("-orphan is deprecated").Err()
	}

	return nil
}

func (c *createSuiteRun) innerRunBB(ctx context.Context, a subcommands.Application, suiteName string) error {
	client, err := bb.NewClient(ctx, c.envFlags.Env().CTPBuilderInfo, c.authFlags)
	if err != nil {
		return err
	}

	req, err := c.testPlatformRequest(suiteName)
	if err != nil {
		return err
	}
	m := map[string]*test_platform.Request{"default": req}
	buildID, err := client.ScheduleCTPBuild(ctx, m, c.buildTags(suiteName))
	if err != nil {
		return err
	}
	buildURL := client.BuildURL(buildID)
	if c.json {
		return printScheduledTaskJSON(a.GetOut(), "cros_test_platform", fmt.Sprintf("%d", buildID), buildURL)
	}
	fmt.Fprintf(a.GetOut(), "Created request at %s\n", buildURL)
	return nil
}

func (c *createSuiteRun) testPlatformRequest(suite string) (*test_platform.Request, error) {
	recipeArgs, err := c.RecipeArgs(c.buildTags(suite))
	if err != nil {
		return nil, err
	}
	recipeArgs.TestPlan = recipe.NewTestPlanForSuites(suite)
	return recipeArgs.TestPlatformRequest()
}

func (c *createSuiteRun) buildTags(suiteName string) []string {
	return append(c.createRunCommon.BuildTags(), "skylab-tool:create-suite", fmt.Sprintf("suite:%s", suiteName))
}
