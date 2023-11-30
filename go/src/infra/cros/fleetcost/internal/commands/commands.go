// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"

	"infra/cros/fleetcost/internal/site"
)

// PingCommand pings the service.
var PingCommand *subcommands.Command = &subcommands.Command{
	UsageLine: "ping [options...]",
	ShortDesc: "ping a fleet cost instance",
	LongDesc:  "Ping a fleet cost instance",
	CommandRun: func() subcommands.CommandRun {
		c := &pingCommand{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		return c
	},
}

type pingCommand struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
}

// Run is the main entrypoint to the ping.
func (c *pingCommand) Run(subcommands.Application, []string, subcommands.Env) int {
	return 1
}
