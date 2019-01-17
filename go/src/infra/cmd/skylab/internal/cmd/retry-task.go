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

// RetryTasks subcommand.
var RetryTasks = &subcommands.Command{
	UsageLine: "retry-tasks [TASK_ID...]",
	ShortDesc: "Retry Skylab tasks",
	LongDesc:  "Retry Skylab tasks with new tasks.",
	CommandRun: func() subcommands.CommandRun {
		c := &retryTasksRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		return c
	},
}

type retryTasksRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  envFlags
}

func (c *retryTasksRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		PrintError(a.GetErr(), err)
		return 1
	}
	return 0
}

func (c *retryTasksRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	originalTaskIDs := c.Flags.Args()
	e := c.envFlags.Env()
	ctx := cli.GetContext(a, c, env)
	s, err := newSwarmingService(ctx, c.authFlags, e)
	if err != nil {
		return err
	}

	// TODO(akeshet/maruel): Use a batched call to get all tasks with the given IDs, if one exists.
	// (I couldn't find one in my initial perusal of the exported API).
	requests := make([]*swarming.SwarmingRpcsNewTaskRequest, len(originalTaskIDs))

	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()

	for i, ID := range originalTaskIDs {
		request, err := s.Task.Request(ID).Context(ctx).Do()
		if err != nil {
			return errors.Annotate(err, fmt.Sprintf("retry task %s", ID)).Err()
		}
		newRequest, err := modifyForRetry(request, e)
		if err != nil {
			return errors.Annotate(err, fmt.Sprintf("retry task %s", ID)).Err()
		}
		requests[i] = newRequest
	}

	// TODO(akeshet/maruel): Use a batched task create call if one exists.
	for i, req := range requests {
		originalID := originalTaskIDs[i]
		resp, err := s.Tasks.New(req).Context(ctx).Do()
		if err != nil {
			return errors.Annotate(err, fmt.Sprintf("retry task %s", originalID)).Err()
		}
		fmt.Fprintf(a.GetOut(), "Retrying %s | Created Swarming task %s\n", originalID, swarmingTaskURL(e, resp.TaskId))
	}

	return nil
}

// modifyForRetry modifies a previously set task request for Skylab retry (by changing nothing but
// the annotation URL).
func modifyForRetry(original *swarming.SwarmingRpcsTaskRequest, e site.Environment) (*swarming.SwarmingRpcsNewTaskRequest, error) {
	newURL := generateAnnotationURL(e)
	slices := make([]*swarming.SwarmingRpcsTaskSlice, len(original.TaskSlices))
	for i, s := range original.TaskSlices {
		cmd := s.Properties.Command
		if cmd[0] != "/opt/infra-tools/skylab_swarming_worker" {
			return nil, fmt.Errorf("task with was not a Skylab task")
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

	return testTaskRequest(original.Name, original.Tags, slices), nil
}
