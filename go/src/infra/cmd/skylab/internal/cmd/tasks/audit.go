// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"fmt"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	skycmdlib "infra/cmd/skylab/internal/cmd/cmdlib"
	"infra/cmd/skylab/internal/cmd/utils"
	"infra/cmd/skylab/internal/site"
	"infra/cmdsupport/cmdlib"
)

// Audit subcommand: Audit hosts.
var Audit = &subcommands.Command{
	UsageLine: "audit [HOST...]",
	ShortDesc: "create audit tasks",
	LongDesc: `Create audit tasks.

This command does not wait for the tasks to start running.`,
	CommandRun: func() subcommands.CommandRun {
		c := &auditRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.Flags.IntVar(&c.expirationMins, "expiration-mins", 10, "The expiration minutes of the repair request.")
		c.Flags.BoolVar(&c.skipVerifyServoUSB, "skip-verify-servo-usb", false, "Do not run verifyer for servo usb drive.")
		c.Flags.BoolVar(&c.skipVerifyDUTStorage, "skip-verify-dut-storage", false, "Do not run verifyer for DUT storage.")
		return c
	},
}

type auditRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  skycmdlib.EnvFlags

	expirationMins       int
	skipVerifyServoUSB   bool
	skipVerifyDUTStorage bool
}

func (c *auditRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *auditRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if c.expirationMins >= dayInMinutes {
		return cmdlib.NewUsageError(c.Flags, "Expiration minutes (%d minutes) cannot exceed 1 day [%d minutes]", c.expirationMins, dayInMinutes)
	}
	if len(args) == 0 {
		return errors.Reason("at least one host has to provided").Err()
	}

	actions, err := c.getActions()
	if err != nil {
		return err
	}

	ctx := cli.GetContext(a, c, env)
	creator, err := utils.NewTaskCreator(ctx, &c.authFlags, c.envFlags)
	if err != nil {
		return err
	}

	for _, host := range args {
		dutName := skycmdlib.FixSuspiciousHostname(host)
		if dutName != host {
			fmt.Fprintf(a.GetErr(), "correcting (%s) to (%s)\n", host, dutName)
		}
		task, err := creator.AuditTask(ctx, dutName, actions, c.expirationMins*60)
		if err != nil {
			return err
		}
		fmt.Fprintf(a.GetOut(), "Created Swarming task %s for host %s\n", task.TaskURL, dutName)
	}
	if len(args) > 1 {
		fmt.Fprintf(a.GetOut(), "\nBatch tasks URL: %s\n\n", creator.GetSessionTasksURL())
	}
	return nil
}

func (c *auditRun) getActions() (string, error) {
	a := make([]string, 0, 2)
	if !c.skipVerifyDUTStorage {
		a = append(a, "verify-dut-storage")
	}
	if !c.skipVerifyServoUSB {
		a = append(a, "verify-servo-usb-drive")
	}
	if len(a) == 0 {
		return "", errors.Reason("At least one action has to be allowed").Err()
	}
	return strings.Join(a, ","), nil
}
