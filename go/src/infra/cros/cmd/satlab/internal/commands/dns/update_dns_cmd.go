// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"fmt"

	"github.com/maruel/subcommands"
)

// UpdateDNSCmd is the command to upsert a hostname-ip pairing in /etc/dut_hosts/hosts used in DNS container
var UpdateDNSCmd = &subcommands.Command{
	UsageLine: "dns -host <hostname> -address <address>",
	ShortDesc: "upsert DNS entry in local satlab network",
	LongDesc:  "Upsert DNS entry in local satlab network",
	CommandRun: func() subcommands.CommandRun {
		c := &updateDNSRun{}
		c.Flags.StringVar(&c.host, "host", "", "hostname to update")
		c.Flags.StringVar(&c.address, "address", "", "address to associate with hostname")
		return c
	},
}

// updateDNSRun struct contains the arguments needed to run UpdateDNSCmd
type updateDNSRun struct {
	subcommands.CommandRunBase
	host    string
	address string
}

// Run is what is called when a user inputs the updateDNSRun command
func (c *updateDNSRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun contains the actual logic of the updateDNSRun command
func (c *updateDNSRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return fmt.Errorf("not implemented yet")
}
