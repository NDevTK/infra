// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/googleoauth"
	"go.chromium.org/luci/common/logging"
)

func launchCmd(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "launch",
		ShortDesc: "launches a JobDefinition on swarming",

		CommandRun: func() subcommands.CommandRun {
			ret := &cmdLaunch{}
			ret.logCfg.Level = logging.Info

			ret.logCfg.AddFlags(&ret.Flags)
			ret.authFlags.Register(&ret.Flags, authOpts)

			ret.Flags.BoolVar(&ret.dump, "dump", false, "dump swarming task to stdout instead of running it.")

			return ret
		},
	}
}

type cmdLaunch struct {
	subcommands.CommandRunBase

	logCfg    logging.Config
	authFlags authcli.Flags

	dump bool
}

func (c *cmdLaunch) validateFlags(ctx context.Context, args []string) (authOpts auth.Options, err error) {
	if len(args) > 0 {
		err = errors.Reason("unexpected positional arguments: %q", args).Err()
		return
	}
	return c.authFlags.Options()
}

func (c *cmdLaunch) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := c.logCfg.Set(cli.GetContext(a, c, env))
	authOpts, err := c.validateFlags(ctx, args)
	if err != nil {
		logging.Errorf(ctx, "bad arguments: %s\n\n", err)
		c.GetFlags().Usage()
		return 1
	}

	jd, err := decodeJobDefinition(ctx)
	if err != nil {
		errors.Log(ctx, err)
		return 1
	}

	authenticator, authClient, swarm, err := newSwarmClient(ctx, authOpts, jd.SwarmingHostname)
	if err != nil {
		errors.Log(ctx, err)
		return 1
	}

	tok, err := authenticator.GetAccessToken(time.Minute)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "getting access token")
		return 1
	}
	info, err := googleoauth.GetTokenInfo(ctx, googleoauth.TokenInfoParams{
		AccessToken: tok.AccessToken,
	})
	uid := info.Email
	if uid == "" {
		uid = "uid:" + info.Sub
	}

	isoFlags, err := getIsolatedFlags(swarm)
	if err != nil {
		errors.Log(ctx, err)
		return 1
	}

	arc := mkArchiver(ctx, isoFlags, authClient)

	logging.Infof(ctx, "building swarming task")
	st, err := jd.GetSwarmingNewTask(ctx, uid, arc)
	if err != nil {
		errors.Log(ctx, err)
		return 1
	}
	logging.Infof(ctx, "building swarming task: done")

	if c.dump {
		err := json.NewEncoder(os.Stdout).Encode(st)
		if err != nil {
			errors.Log(ctx, err)
			return 1
		}
		return 0
	}

	logging.Infof(ctx, "launching swarming task")
	req, err := swarm.Tasks.New(st).Do()
	if err != nil {
		errors.Log(ctx, err)
		return 1
	}
	logging.Infof(ctx, "launching swarming task: done")

	logging.Infof(ctx, "Launched swarming task: https://%s/task?id=%s",
		jd.SwarmingHostname, req.TaskId)
	logging.Infof(ctx, "LUCI UI: https://ci.chromium.org/swarming/task/%s?server=%s",
		req.TaskId, jd.SwarmingHostname)
	return 0
}
