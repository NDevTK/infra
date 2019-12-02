// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	apipb "go.chromium.org/luci/swarming/proto/api"
)

func getSwarmCmd(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "get-swarm <swarm task id>",
		ShortDesc: "obtain a JobDefinition from a swarming task",
		LongDesc:  `Obtains the task definition from swarming and produce a JobDefinition.`,

		CommandRun: func() subcommands.CommandRun {
			ret := &cmdGetSwarm{}
			ret.logCfg.Level = logging.Info

			ret.logCfg.AddFlags(&ret.Flags)
			ret.authFlags.Register(&ret.Flags, authOpts)

			ret.Flags.StringVar(&ret.swarmingHost, "S", "chromium-swarm.appspot.com",
				"the swarming `host` to get the task from.")

			ret.Flags.BoolVar(&ret.pinMachine, "pin-machine", false,
				"Pin the dimensions of the JobDefinition to run on the same machine.")

			return ret
		},
	}
}

type cmdGetSwarm struct {
	subcommands.CommandRunBase

	logCfg    logging.Config
	authFlags authcli.Flags

	taskID       string
	swarmingHost string
	pinMachine   bool
}

func (c *cmdGetSwarm) validateFlags(ctx context.Context, args []string) (authOpts auth.Options, err error) {
	if len(args) != 1 {
		err = errors.Reason("expected 1 positional argument: %q", args).Err()
		return
	}
	c.taskID = args[0]

	if err = validateHost(c.swarmingHost); err != nil {
		err = errors.Annotate(err, "SwarmingHostname").Err()
		return
	}

	return c.authFlags.Options()
}

// GetFromSwarmingTask retrieves and renders a JobDefinition from the given
// swarming task, printing it to stdout and returning an error.
func GetFromSwarmingTask(ctx context.Context, authOpts auth.Options, name, host, taskID string, pinMachine bool) error {
	logging.Infof(ctx, "getting task definition: %q", taskID)
	_, swarm, err := newSwarmClient(ctx, authOpts, host)
	if err != nil {
		return err
	}

	req, err := swarm.Task.Request(taskID).Do()
	if err != nil {
		return err
	}

	jd, err := JobDefinitionFromNewTaskRequest(
		taskRequestToNewTaskRequest(req), name, host)
	if err != nil {
		return err
	}

	logging.Infof(ctx, "getting task definition: done")

	if pinMachine {
		logging.Infof(ctx, "pinning swarming bot")

		rslt, err := swarm.Task.Result(taskID).Do()
		if err != nil {
			return err
		}
		if len(rslt.BotDimensions) == 0 {
			return errors.Reason("could not pin bot ID, task is %q", rslt.State).Err()
		}

		id := ""
		for _, d := range rslt.BotDimensions {
			if d.Key == "id" {
				id = d.Value[0]
				break
			}
		}

		if id == "" {
			return errors.New("could not pin bot ID (bot ID not found)")
		}

		jd.EditSwarming(ctx, authOpts, func(ejd *EditSWJobDefinition) {
			ejd.tweakSlices(func(slc *apipb.TaskSlice) error {
				if slc.Properties == nil {
					slc.Properties = &apipb.TaskProperties{}
				}

				var poolDim *apipb.StringListPair
				for _, dim := range slc.GetProperties().GetDimensions() {
					if dim.Key == "pool" {
						poolDim = dim
						break
					}
				}
				slc.Properties.Dimensions = []*apipb.StringListPair{
					{Key: "id", Values: []string{id}},
				}
				if poolDim != nil {
					slc.Properties.Dimensions = append(slc.Properties.Dimensions, poolDim)
				}
				return nil
			})
		})
	}

	return dumpJobDefinition(jd)
}

func (c *cmdGetSwarm) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := c.logCfg.Set(cli.GetContext(a, c, env))
	authOpts, err := c.validateFlags(ctx, args)
	if err != nil {
		logging.Errorf(ctx, "bad arguments: %s\n\n", err)
		c.GetFlags().Usage()
		return 1
	}

	name := fmt.Sprintf(`get-swarm %s`, c.taskID)
	if err = GetFromSwarmingTask(ctx, authOpts, name, c.swarmingHost, c.taskID, c.pinMachine); err != nil {
		errors.Log(ctx, err)
		return 1
	}

	return 0
}
