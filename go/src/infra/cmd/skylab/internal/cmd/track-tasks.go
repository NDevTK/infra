// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/flag"

	"infra/cmd/cros/ipcpubsub/pubsublib"
	"infra/cmd/skylab/internal/cmd/recipe"
	"infra/cmd/skylab/internal/site"

	"cloud.google.com/go/pubsub"
	"github.com/dchest/uniuri"
)

const projectName = "chromeos-swarming"

// TrackTasks subcommand.
var TrackTasks = &subcommands.Command{
	UsageLine: "track-tasks [-task-id TASK_ID...] [-tag TAG...]",
	ShortDesc: "Deprecated, do not use.",
	LongDesc:  `Create copies of tasks to run again.`,
	CommandRun: func() subcommands.CommandRun {
		c := &trackTasksRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.Var(flag.StringSlice(&c.statusIds), "status-id", "Identifier string for locating tests to track. May be specified multiple times.")
		c.Flags.BoolVar(&c.includeInterim, "include-interim", false, "If true, print state updates other than completion notifications.")
		c.Flags.BoolVar(&c.printURL, "print-url", false, "Print full URLs to skylab builds rather than just the task IDs.")
		//c.Flags.BoolVar(&c.isolate, "link-isolate", false, "Print URL to the Isolate outputs of the tests.")
		return c
	},
}

type trackTasksRun struct {
	subcommands.CommandRunBase
	authFlags      authcli.Flags
	envFlags       envFlags
	statusIds      []string
	includeInterim bool
	isolate        bool
	namesOnly      bool
	printURL       bool
	idChannel      chan string
	msgsReceived   int
}

func (c *trackTasksRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		PrintError(a.GetErr(), err)
		return 1
	}
	return 0
}

func (c *trackTasksRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	topic := recipe.PubsubStatusTopic
	ctx, can := context.WithCancel(cli.GetContext(a, c, env))
	defer can()
	c.idChannel = make(chan string, 1)

	sub, err := createSubscription(ctx, topic)
	if err != nil {
		return err
	}
	// There is no easy way to know when all tasks have been found and
	// printed, so we instead handle cancellation gracefully
	cancelWhenSIGINT(can)
	ch := pubsublib.ReceiveToChannel(ctx, sub)
	go c.fetchCompleteTasks(ctx)
	go c.receiveMessages(ctx, ch)

	c.printOutput(ctx)
	return nil
}

func (c *trackTasksRun) fetchCompleteTasks(ctx context.Context) error {
	taskIds, err := c.queryCompleteTasks(ctx)
	if err != nil {
		return err
	}
	for n := range taskIds {
		select {
		case <-ctx.Done():
			return nil
		default:
			c.idChannel <- taskIds[n]
		}
	}
	return nil
}

// Checks whether a message (JSON object) has one of the flags we're expecting
func (c *trackTasksRun) shouldActOnMessage(message map[string]string) bool {
	subtopic := message["userdata"]
	for _, id := range c.statusIds {
		if subtopic == id {
			return true
		}
	}
	return false
}

// Pull messages from the MOE channel, filter, and pass to output channel
func (c *trackTasksRun) receiveMessages(ctx context.Context, ch <-chan pubsublib.MessageOrError) {
	for {
		select {
		case <-ctx.Done():
			return
		case moe := <-ch:
			if moe.Message == nil {
				continue
			}
			m := make(map[string]string)
			json.Unmarshal(moe.Message.Body(), &m)
			if !c.shouldActOnMessage(m) {
				moe.Message.Ack()
				continue
			}
			id := m["task_id"]
			if len(id) == 0 {
				moe.Message.Ack()
				continue
			}
			if c.includeInterim || c.isTaskComplete(ctx, id) {
				c.idChannel <- id
				moe.Message.Ack()
			}
		}
	}
}

func (c *trackTasksRun) printOutput(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-c.idChannel:
			if c.printURL {
				fmt.Printf("https://chromeos-swarming.appspot.com/task?id=%s\n", id)
			} else if c.includeInterim {
				fmt.Printf("Task %s updated\n", id)
			} else {
				fmt.Printf("Task %s complete\n", id)
			}
			//if c.isolate {
			//	fmt.Println("Linking Isolate not yet supported.")
			//}
		}
	}
}

func (c *trackTasksRun) isTaskComplete(ctx context.Context, id string) bool {
	ser, err := swarming.NewService(ctx)
	if err != nil {
		return false
	}
	ts := swarming.NewTaskService(ser)
	res, err := ts.Result(id).Do()
	if err != nil {
		return false
	}
	state := res.State
	switch state {
	case "CANCELED", "COMPLETED", "KILLED", "TIMED_OUT":
		return true
	default:
		return false
	}
}

func (c *trackTasksRun) queryCompleteTasks(ctx context.Context) ([]string, error) {
	var completeIDs []string
	ser, err := swarming.NewService(ctx)
	if err != nil {
		return nil, err
	}
	ts := swarming.NewTasksService(ser)
	for _, state := range [...]string{"CANCELED", "COMPLETED", "KILLED", "TIMED_OUT"} {
		for _, id := range c.statusIds {
			list := ts.List()
			list.Tags(fmt.Sprintf("results-label:%s", id))
			list.State(state)
			res, err := list.Do()
			if err != nil {
				return nil, err
			}
			for _, r := range res.Items {
				completeIDs = append(completeIDs, r.TaskId)
			}
		}
	}
	return completeIDs, nil
}

func cancelWhenSIGINT(can context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		_ = <-ch
		can()
	}()
}

func createSubscription(ctx context.Context, topic string) (*pubsub.Subscription, error) {
	client, err := pubsub.NewClient(ctx, projectName)
	if err != nil {
		return nil, err
	}
	t := client.Topic(topic)
	conf := pubsub.SubscriptionConfig{Topic: t}
	sub, err := client.CreateSubscription(ctx, uniuri.New(), conf)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
