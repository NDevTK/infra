// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"fmt"

	"github.com/maruel/subcommands"
)

// GetDNSCmd is the command to print the current DNS entries for local satlab network
var GetDNSCmd = &subcommands.Command{
	UsageLine: "dns",
	ShortDesc: "get DNS entries for local satlab network",
	LongDesc:  "Get DNS entries for local satlab network",
	CommandRun: func() subcommands.CommandRun {
		c := &getDNSRun{}
		return c
	},
}

// getDNSRun struct contains the arguments needed to run GetDNSCmd
type getDNSRun struct {
	subcommands.CommandRunBase
}

// Run is what is called when a user inputs the getDNSRun command
func (c *getDNSRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun contains business logic
func (c *getDNSRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return fmt.Errorf("not implemented yet")
}
