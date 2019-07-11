// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"infra/cmd/cros/ipcpubsub/pubsublib"

	"go.chromium.org/luci/common/flag"

	"time"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
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
	return 0
}

func (c *subscribeRun) newSubscription(ctx context.Context, cli pubsublib.Client, t pubsublib.Topic, id string, f pubsublib.Filter) (pubsublib.Subscription, error) {
	s, err := cli.CreateSubscription(ctx, t, id)
	if err != nil {
		return nil, err
	}
	if f != nil {
		if err = s.SetFilter(ctx, f); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Subscribe pulls messageCount messages from the subscription sub, respecting any filtering sub has.
//   It returns each message as a string of plain bytes.
func Subscribe(ctx context.Context, sub pubsublib.Subscription, messageCount int) ([][]byte, error) {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	storedMessages := map[string]pubsublib.Message{}
	msgChannel := make(chan pubsublib.Message, messageCount)

	errs := make(chan error, 1)
	handleMessage := func(ctx context.Context, msg pubsublib.Message) {
		select {
		case <-ctx.Done():
			//cancelled, noop
		case msgChannel <- msg:
			//export message to channel and then noop
		}
	}
	for {
		go func() {
			errs <- sub.Receive(cctx, handleMessage)
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errs:
			if err != nil {
				return nil, err
			}
		case m := <-msgChannel:
			storedMessages[m.ID()] = m
			if len(storedMessages) >= messageCount {
				return extractBodiesFromMap(storedMessages), nil
			}
		}
	}
}

func extractBodiesFromMap(m map[string]pubsublib.Message) [][]byte {
	l := make([][]byte, 0, len(m))
	for _, v := range m {
		l = append(l, v.Body())
	}
	return l
}
