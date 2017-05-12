// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/maruel/subcommands"

	"github.com/luci/luci-go/client/authcli"
	swarming "github.com/luci/luci-go/common/api/swarming/swarming/v1"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/isolatedclient"
	"github.com/luci/luci-go/common/logging"
)

const defaultSwarmingServer = "https://chromium-swarm.appspot.com"

func launchRawCmd(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "launch-raw",
		ShortDesc: "Reads a builder definition from stdin and launches it on swarming.",

		CommandRun: func() subcommands.CommandRun {
			ret := &cmdLaunchRaw{}
			ret.logCfg.Level = logging.Info

			ret.logCfg.AddFlags(&ret.Flags)
			ret.authFlags.Register(&ret.Flags, authOpts)
			ret.isolatedFlags.Init(&ret.Flags)

			ret.Flags.StringVar(&ret.swarmingServer, "S", defaultSwarmingServer,
				"The swarming server to launch the task on.")

			return ret
		},
	}
}

type cmdLaunchRaw struct {
	subcommands.CommandRunBase

	logCfg         logging.Config
	authFlags      authcli.Flags
	isolatedFlags  isolatedclient.Flags
	swarmingServer string
}

func (c *cmdLaunchRaw) validateFlags(ctx context.Context, args []string) (authOpts auth.Options, err error) {
	if len(args) > 0 {
		err = errors.Reason("unexpected positional arguments: %(args)q").D("args", args).Err()
		return
	}
	if c.isolatedFlags.ServerURL == "" {
		c.isolatedFlags.ServerURL = defaultIsolateServer
	}
	if err = c.isolatedFlags.Parse(); err != nil {
		err = errors.Annotate(err).Reason("bad isolate flags").Err()
		return
	}
	return c.authFlags.Options()
}

func (c *cmdLaunchRaw) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := c.logCfg.Set(cli.GetContext(a, c, env))
	authOpts, err := c.validateFlags(ctx, args)
	if err != nil {
		logging.Errorf(ctx, "bad arguments: %s", err)
		fmt.Fprintln(os.Stderr)
		subcommands.CmdHelp.CommandRun().Run(a, args, env)
		return 1
	}

	jd := &JobDefinition{}
	if err := json.NewDecoder(os.Stdin).Decode(jd); err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}

	authenticator := auth.NewAuthenticator(ctx, auth.SilentLogin, authOpts)
	authClient, err := authenticator.Client()
	if err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}
	swarm, err := swarming.New(authClient)
	if err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}
	swarm.BasePath = c.swarmingServer + "/api/swarming/v1/"

	arc := mkArchiver(ctx, c.isolatedFlags, authClient)

	logging.Infof(ctx, "building swarming task")
	st, err := jd.GetSwarmingNewTask(ctx, arc)
	if err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}
	logging.Infof(ctx, "building swarming task: done")

	logging.Infof(ctx, "launching swarming task")
	req, err := swarm.Tasks.New(st).Do()
	if err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}
	logging.Infof(ctx, "launching swarming task: done")

	logging.Infof(ctx, "Launched swarming task: %s/task?id=%s",
		c.swarmingServer, req.TaskId)
	return 0
}
