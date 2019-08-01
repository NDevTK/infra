// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"time"

	"infra/cmd/cros/ipcpubsub/pubsublib"

	"cloud.google.com/go/pubsub"
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/flag"
)

type subscribeRun struct {
	baseRun
	attributes   map[string]string
	messageCount int
	subName      string
	timeout      time.Duration
	outputDir    string
}

// CmdSubscribe describes the subcommand flags for subscribing to messages
var CmdSubscribe = &subcommands.Command{
	UsageLine: "subscribe -project [PROJECT] -topic [TOPIC] -output [PATH/TO/OUTPUT/DIR] [OPTIONS]",
	ShortDesc: "subscribe to a filtered topic",
	CommandRun: func() subcommands.CommandRun {
		c := &subscribeRun{}
		c.registerCommonFlags(&c.Flags)
		c.Flags.Var(flag.JSONMap(&c.attributes), "attributes", "map of attributes to filter for")
		c.Flags.IntVar(&c.messageCount, "count", 1, "number of messages to read before returning")
		c.Flags.StringVar(&c.outputDir, "output", "", "path to directory to store output")
		c.Flags.StringVar(&c.subName, "sub-name", "", "name of subscription: must be 3-255 characters, start with a letter, and composed of alphanumerics and -_.~+% only")
		c.Flags.DurationVar(&c.timeout, "timeout", time.Hour, "timeout to stop waiting, ex. 10s, 5m, 1h30m")
		return c
	},
}

func (c *subscribeRun) validateArgs(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	if c.messageCount < 1 {
		return errors.Reason("message-count must be >0").Err()
	}
	if c.subName == "" {
		return errors.Reason("subscription name is required").Err()
	}
	minTimeout := 10 * time.Second
	if c.timeout < minTimeout {
		return errors.Reason("timeout must be >= 10s").Err()
	}
	return nil
}

func (c *subscribeRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.validateArgs(ctx, a, args, env); err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		c.Flags.Usage()
		return 1
	}
	client, err := pubsub.NewClient(ctx, c.project)
	if err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		return 1
	}
	sub := client.Subscription(c.subName)
	received, err := pubsublib.PubsubSubscribe(ctx, sub, c.messageCount, c.attributes)
	if err != nil {
		fmt.Fprintln(a.GetErr(), err.Error())
		return 1
	}
	// Do something with received messages
	_ = received
	return 0
}

// Subscribe pulls messageCount messages from the message stream msgs. If it receives an error or
//  has received messageCount messages which match the filter, it will close the 'done' channel and exit
// Subscribe returns each message as a string of unformatted bytes.
func Subscribe(ctx context.Context, msgs <-chan pubsublib.Message, errs <-chan error, done chan<- interface{}, messageCount int, filter map[string]string) ([][]byte, error) {
	storedMessages := map[string]pubsublib.Message{}

	defer close(done)

	for m := range msgs {
		if _, present := storedMessages[m.ID()]; present {
			m.Ack()
			continue
		}
		if !matchesFilter(filter, m) {
			m.Ack()
			continue
		}
		storedMessages[m.ID()] = m
		m.Ack()
		if len(storedMessages) >= messageCount {
			return extractBodiesFromMap(storedMessages), nil
		}
	}
	select {
	case e := <-errs:
		if e != nil {
			return nil, e
		}
	default:
	}
	return nil, errors.Reason("Subscribe ended without sufficient messages.").Err()
}

func matchesFilter(f map[string]string, m pubsublib.Message) bool {
	a := m.Attributes()
	if len(f) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	for k, v := range f {
		if a[k] != v {
			return false
		}
	}
	return true
}

func extractBodiesFromMap(m map[string]pubsublib.Message) [][]byte {
	lst := make([][]byte, 0, len(m))
	for _, v := range m {
		lst = append(lst, v.Body())
	}
	return lst
}
