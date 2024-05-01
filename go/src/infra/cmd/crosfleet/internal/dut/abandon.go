// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"fmt"
	"strings"

	"infra/cmd/crosfleet/internal/buildbucket"
	crosfleetcommon "infra/cmd/crosfleet/internal/common"
	"infra/cmd/crosfleet/internal/site"
	"infra/cros/cmd/common_lib/common"
	"infra/libs/skylab/common/heuristics"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
)

const abandonCmd = "abandon"

var abandon = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s [HOST...]", abandonCmd),
	ShortDesc: "abandon DUTs which were previously leased via 'dut lease'",
	LongDesc: `Abandon DUTs which were previously leased via 'dut lease'.

If no hostnames are entered, all pending or active leases by the current user
will be abandoned.

This command's behavior is subject to change without notice.
Do not build automation around this subcommand.`,
	CommandRun: func() subcommands.CommandRun {
		c := &abandonRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.printer.Register(&c.Flags)
		c.Flags.StringVar(&c.reason, "reason", "", "Optional reason for abandoning.")
		return c
	},
}

type abandonRun struct {
	subcommands.CommandRunBase
	reason    string
	authFlags authcli.Flags
	envFlags  crosfleetcommon.EnvFlags
	printer   crosfleetcommon.CLIPrinter
}

func (c *abandonRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		crosfleetcommon.PrintCmdError(a, err)
		return 1
	}
	return 0
}

func (c *abandonRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	userEmail, err := crosfleetcommon.GetUserEmail(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	swarmingBotsClient, err := newSwarmingBotsClient(ctx, c.envFlags.Env().SwarmingService, &c.authFlags)
	if err != nil {
		return err
	}
	leasesBBClient, err := buildbucket.NewClient(ctx, c.envFlags.Env().DUTLeaserBuilder, c.envFlags.Env().BuildbucketService, c.authFlags)
	if err != nil {
		return err
	}
	earliestCreateTime := crosfleetcommon.OffsetTimestamp(-1 * maxLeaseLengthMinutes)
	var botIDs []string
	var correctedDeviceNames []string
	for _, deviceName := range args {
		correctedDeviceName := heuristics.NormalizeBotNameToDeviceName(deviceName)
		id, err := hostnameToBotID(ctx, swarmingBotsClient, correctedDeviceName)
		if err != nil {
			return err
		}
		botIDs = append(botIDs, id)
		correctedDeviceNames = append(correctedDeviceNames, correctedDeviceName)
	}

	// Flow for non-Scheduke (legacy) leases. TODO(b/332370221): Delete this.
	err = leasesBBClient.CancelBuildsByUser(ctx, c.printer, earliestCreateTime, userEmail, botIDs, c.reason)
	if err != nil {
		return err
	}

	// Flow for Scheduke leases.
	authOpts, err := c.authFlags.Options()
	if err != nil {
		return err
	}
	err = common.Abandon(ctx, authOpts, correctedDeviceNames, c.envFlags.UseDev())
	if err != nil {
		return err
	}
	if len(correctedDeviceNames) > 0 {
		c.printer.WriteTextStdout("Cancelled all leases for devices %s by the current user", strings.Join(correctedDeviceNames, ", "))
	} else {
		c.printer.WriteTextStdout("Cancelled all leases by the current user")
	}

	return nil
}
