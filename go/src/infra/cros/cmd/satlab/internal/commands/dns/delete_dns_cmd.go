// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"fmt"

	"github.com/maruel/subcommands"
)

// DeleteDNSCmd is the command to delete a hostname from the hostsfile of the DNS container
var DeleteDNSCmd = &subcommands.Command{
	UsageLine: "dns -host <hostname>",
	ShortDesc: "delete DNS entry in local satlab network",
	LongDesc:  "Delete DNS entry in local satlab network",
	CommandRun: func() subcommands.CommandRun {
		c := &deleteDNSRun{}
		c.Flags.StringVar(&c.host, "host", "", "hostname to delete")
		return c
	},
}

// deleteDNSRun struct contains the arguments needed to run DeleteDNSCmd
type deleteDNSRun struct {
	subcommands.CommandRunBase
	host string
}

// Run is what is called when a user inputs the deleteDNSRun command
func (c *deleteDNSRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun executes actual logic
func (c *deleteDNSRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return fmt.Errorf("not implemented yet")
}
