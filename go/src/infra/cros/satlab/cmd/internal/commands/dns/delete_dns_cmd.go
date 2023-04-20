// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dns

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/cmd/internal/commands"
	"infra/cros/satlab/cmd/internal/site"
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

// innerRun calls underlying business logic with appropriate functions and interfaces injected
// extra abstraction layer allows us to test `runCmdInjected` with fake implementations
func (c *deleteDNSRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return c.runCmdInjected(args, commands.GetDockerHostBoxIdentifier)
}

// runCmdInjected executes business logic
func (c *deleteDNSRun) runCmdInjected(args []string, dhbIDFunc DockerHostBoxIdentifierGetter) error {
	satlabID, err := dhbIDFunc()
	if err != nil {
		return err
	}

	err = c.validate(args, satlabID)
	if err != nil {
		return err
	}

	_, err = DeleteRecord(ensureRecords, readContents, c.host)
	return err
}

// validate checks for required and unexpected args + formats hostname
func (c *deleteDNSRun) validate(args []string, satlabId string) error {
	if c.host == "" {
		return errors.Reason("host must be specified").Err()
	}
	if len(args) > 0 {
		return errors.Reason("unrecognized positional argument(s): %+v", args).Err()
	}

	c.host = site.MaybePrepend(site.Satlab, satlabId, c.host)
	return nil
}
