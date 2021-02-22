// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"fmt"
	"go.chromium.org/luci/common/cli"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/luci/auth/client/authcli"
	"infra/cmd/crosfleet/internal/buildbucket"
	"infra/cmd/crosfleet/internal/common"
	"infra/cmd/crosfleet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// testPlanCmdName is the name of the `crosfleet run testplan` command.
var testPlanCmdName = "testplan"

var testplan = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s [FLAGS...] PLAN_FILE", testPlanCmdName),
	ShortDesc: "runs a test plan",
	LongDesc: `Launches a test plan task for a given test plan file.

You must supply -board, -image and -pool.

This command is more general than "run test" or "run suite". The supplied
test plan should conform to the TestPlan proto as defined here:
https://chromium.googlesource.com/chromiumos/infra/proto/+/master/src/test_platform/request.proto

This command does not wait for the task to start running.

This command's behavior is subject to change without notice.
Do not build automation around this subcommand.`,
	CommandRun: func() subcommands.CommandRun {
		c := &planRun{}
		c.envFlags.Register(&c.Flags)
		c.testCommonFlags.Register(&c.Flags)
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		return c
	},
}

type planRun struct {
	subcommands.CommandRunBase
	testCommonFlags
	authFlags authcli.Flags
	envFlags  common.EnvFlags
}

func (c *planRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *planRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(&c.Flags, testPlanCmdName); err != nil {
		return err
	}
	if len(args) > 1 {
		return fmt.Errorf("expected exactly one arg for the test plan file; found %v", args)
	}
	testPlan, err := readTestPlan(args[0])
	if err != nil {
		return err
	}
	// Don't create a tag for the user's test plan file.
	buildTags := c.buildTags(testCmdName, "")

	ctx := cli.GetContext(a, c, env)
	bbClient, err := buildbucket.NewClient(ctx, c.envFlags.Env().CTPBuilderInfo, c.authFlags)
	if err != nil {
		return err
	}
	buildID, err := launchCTPBuild(ctx, bbClient, testPlan, buildTags, &c.testCommonFlags)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.GetOut(), "Running %s at %s\n", testPlanCmdName, bbClient.BuildURL(buildID))
	return nil
}

func readTestPlan(path string) (*test_platform.Request_TestPlan, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error reading test plan: %v", err)
	}
	defer file.Close()

	testPlan := &test_platform.Request_TestPlan{}
	if err := common.JSONPBUnmarshaler.Unmarshal(file, testPlan); err != nil {
		return nil, fmt.Errorf("error reading test plan: %v", err)
	}
	return testPlan, nil
}
