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

// DockerHostBoxIdentifierGetter is a function type fulfilled by commands.GetDockerHostBoxIdentifier
// we do this because the function assumes we are in a real Satlab env and we can't unit test using this func
type DockerHostBoxIdentifierGetter func() (string, error)

// HostfileUpdater is a function type fulfilled by UpdateRecord
// it should take care of the e2e flow of updating a record and applying those changes
// any implementation should be atomic
type HostfileUpdater func(host string, addr string) (string, error)

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

// innerRun gathers all needed function and interface implementations and calls the business logic
func (c *updateDNSRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return c.runCmdInjected(args, commands.GetDockerHostBoxIdentifier, UpdateRecord)
}

// runCmdInjected executes actual logic of command
// it takes in mockable functions and interfaces to make it easier to unit test the logic of the update dns cmd
func (c *updateDNSRun) runCmdInjected(args []string, dhbIDFunc DockerHostBoxIdentifierGetter, updateRecordFunc HostfileUpdater) error {
	satlabID, err := dhbIDFunc()
	if err != nil {
		return err
	}

	err = c.validate(args, satlabID)
	if err != nil {
		return err
	}

	_, err = updateRecordFunc(c.host, c.address)
	if err != nil {
		return err
	}

	return nil
}

// validate checks for required and unexpected args + formats hostname
func (c *updateDNSRun) validate(args []string, satlabId string) error {
	if c.host == "" {
		return errors.Reason("host must be specified").Err()
	}
	if c.address == "" {
		return errors.Reason("address must be specified").Err()
	}
	if len(args) > 0 {
		return errors.Reason("unrecognized positional argument(s): %+v", args).Err()
	}

	c.host = site.MaybePrepend(site.Satlab, satlabId, c.host)
	return nil
}
