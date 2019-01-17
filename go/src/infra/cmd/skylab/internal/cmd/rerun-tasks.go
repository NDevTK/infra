// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab/internal/site"
)

// RerunTasks subcommand.
var RerunTasks = &subcommands.Command{
	UsageLine: "rerun-tasks [TASK_ID...]",
	ShortDesc: "Rerun Skylab tasks",
	LongDesc:  "Rerun Skylab tasks with new tasks.",
	CommandRun: func() subcommands.CommandRun {
		c := &rerunTasksRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		return c
	},
}

type rerunTasksRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  envFlags
}

func (c *rerunTasksRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		PrintError(a.GetErr(), err)
		return 1
	}
	return 0
}

func (c *rerunTasksRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if c.Flags.NArg() == 0 {
		return NewUsageError(c.Flags, "missing task ID")
	}
	originalIDs := c.Flags.Args()
	siteEnv := c.envFlags.Env()
	ctx := cli.GetContext(a, c, env)
	s, err := newSwarmingService(ctx, c.authFlags, siteEnv)
	if err != nil {
		return err
	}

	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()

	originalRequests, err := getSwarmingRequestsForIds(ctx, originalIDs, s)
	if err != nil {
		return err
	}

	// TODO(akeshet/maruel): Use a batched task create call if one exists.
	for i, original := range originalRequests {
		originalID := originalIDs[i]
		newRequest, err := createRerunRequest(original, siteEnv)
		if err != nil {
			return errors.Annotate(err, fmt.Sprintf("rerun task %s", originalID)).Err()
		}

		resp, err := s.Tasks.New(newRequest).Context(ctx).Do()
		if err != nil {
			return errors.Annotate(err, fmt.Sprintf("rerun task %s", originalID)).Err()
		}
		fmt.Fprintf(a.GetOut(), "Rerunning %s\tCreated Swarming task %s\n", originalID, swarmingTaskURL(siteEnv, resp.TaskId))
	}

	return nil
}

func getSwarmingRequestsForIds(ctx context.Context, IDs []string, s *swarming.Service) ([]*swarming.SwarmingRpcsTaskRequest, error) {
	// TODO(akeshet/maruel): Use a batched call to get all tasks with the given IDs, if one exists.
	// (I couldn't find one in my initial perusal of the exported API).
	requests := make([]*swarming.SwarmingRpcsTaskRequest, len(IDs))
	for i, ID := range IDs {
		request, err := s.Task.Request(ID).Context(ctx).Do()
		if err != nil {
			return nil, errors.Annotate(err, fmt.Sprintf("rerun task %s", ID)).Err()
		}
		requests[i] = request
	}
	return requests, nil
}

// createRerunRequest creates a new request to rerun a previous Skylab task.
func createRerunRequest(original *swarming.SwarmingRpcsTaskRequest, siteEnv site.Environment) (*swarming.SwarmingRpcsNewTaskRequest, error) {
	newURL := generateAnnotationURL(siteEnv)
	slices := make([]*swarming.SwarmingRpcsTaskSlice, len(original.TaskSlices))
	for i, s := range original.TaskSlices {
		cmd := s.Properties.Command
		if cmd[0] != "/opt/infra-tools/skylab_swarming_worker" {
			return nil, fmt.Errorf("task was not a Skylab task")
		}

		newCmd := make([]string, len(cmd))
		copy(newCmd, cmd)
		for j, c := range newCmd {
			if c == "-logdog-annotation-url" {
				newCmd[j+1] = newURL
			}
		}

		newSlice := &swarming.SwarmingRpcsTaskSlice{
			ExpirationSecs:  s.ExpirationSecs,
			ForceSendFields: s.ForceSendFields,
			NullFields:      s.NullFields,
			Properties:      s.Properties,
			WaitForCapacity: s.WaitForCapacity,
		}
		newSlice.Properties.Command = newCmd

		slices[i] = s
	}

	return newTaskRequest(original.Name, original.Tags, slices, original.Priority), nil
}
