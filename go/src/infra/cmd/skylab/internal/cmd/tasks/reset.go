// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab/internal/cmd/utils"
	"infra/cmd/skylab/internal/site"
	"infra/cmdsupport/cmdlib"
)

// Reset subcommand: Reset hosts.
var Reset = &subcommands.Command{
	UsageLine: "reset [HOST...]",
	ShortDesc: "create reset tasks",
	LongDesc: `Create reset tasks.

This command does not wait for the task to start running.`,
	CommandRun: func() subcommands.CommandRun {
		c := &resetRun{}
		c.AuthFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.EnvFlags.Register(&c.Flags)
		c.Flags.IntVar(&c.expirationMins, "expiration-mins", 10, "The expiration minutes of the reset request.")
		return c
	},
}

type resetRun struct {
	subcommands.CommandRunBase
	utils.TaskFlags
	expirationMins int
}

func (c *resetRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *resetRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if c.expirationMins >= dayInMinutes {
		return cmdlib.NewUsageError(c.Flags, "Expiration minutes (%d minutes) cannot exceed 1 day [%d minutes]", c.expirationMins, dayInMinutes)
	}

	return errors.Reason("not implemeneted yet").Err()
}
