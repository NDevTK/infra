// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmds

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/iterator"

	"infra/cros/cmd/crosgrep/internal/base"
	"infra/cros/cmd/crosgrep/internal/swarming/logging"
	"infra/cros/cmd/crosgrep/internal/swarming/query"
)

// StatusLog gets the status.log associated with a specific swarming task if a
// task ID is provided. Otherwise, it returns the status log associated with
// an arbitrary swarming task.
var StatusLog = &subcommands.Command{
	UsageLine: `status-log`,
	ShortDesc: "get the status log",
	LongDesc:  "Get the status log for a specified task or an arbitrary log if no task is specified.",
	CommandRun: func() subcommands.CommandRun {
		c := &statusLogCmd{}
		c.InitFlags()
		c.Flags.StringVar(&c.taskID, "task-id", "", "The task ID to search")
		return c
	},
}

// StatusLogCmd holds the arguments needed to get the status log of a task.
// There are the crosgrep common arguments and the task ID.
type statusLogCmd struct {
	base.Command
	taskID string
}

// Run is a wrapper around the main entrypoint for the status log command.
func (c *statusLogCmd) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// InnerRun is the main entrypoint for the status-log commad.
func (c *statusLogCmd) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	ctx = logging.SetContextVerbosity(ctx, c.Verbose())
	bqClient, err := bigquery.NewClient(ctx, c.GetBQProject())
	if err != nil {
		return errors.Annotate(err, "status-log: getting BigQuery client with project %q", c.GetBQProject()).Err()
	}
	_, err = storage.NewClient(ctx)
	if err != nil {
		return errors.Annotate(err, "status-log: getting Google Storage client").Err()
	}
	it, err := query.RunStatusLogQuery(
		ctx,
		bqClient,
		&query.GetStatusLogParams{
			SwarmingTaskID: c.taskID,
		},
	)
	if err != nil {
		return errors.Annotate(err, "status-log: getting result set").Err()
	}
	for {
		var item map[string]bigquery.Value
		err := it.Next(&item)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Annotate(err, "status-log: extracting item from result set").Err()
		}
		record, ok := item["bb_output_properties"]
		if !ok {
			return errors.New("status-log: bb_output_properties field not present")
		}
		encodedRecord, ok := record.(string)
		if !ok {
			return errors.New("status-log: bb_output_properties field is not string")
		}
		parsed, err := query.UnmarshalBBRecord(encodedRecord)
		if err != nil {
			return errors.Annotate(err, "status-log: extracting record").Err()
		}
		item["bb_output_properties"] = parsed
		// TODO(gregorynisbet): Replace this print with a better strategy for printing records
		// to users is an easily readable way.
		fmt.Fprintf(a.GetOut(), "%#v\n", item)
		jsonEncoded, err := query.JSONEncodeBBRecord(parsed)
		if err != nil {
			return err
		}
		item["bb_output_properties"] = jsonEncoded
		fmt.Fprintf(a.GetOut(), "%s\n", jsonEncoded)
	}
	return nil
}
