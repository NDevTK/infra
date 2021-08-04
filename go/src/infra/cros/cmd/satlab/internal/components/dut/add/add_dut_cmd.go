// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/maruel/subcommands"

	"infra/cmdsupport/cmdlib"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/commands/dns"
	"infra/cros/cmd/satlab/internal/common"
	"infra/cros/cmd/satlab/internal/components/dut/internal/adders"
)

// DUTCmd is the command that deploys a Satlab DUT.
var DUTCmd = &subcommands.Command{
	UsageLine: "dut [options ...]",
	ShortDesc: "Deploy a Satlab DUT",
	CommandRun: func() subcommands.CommandRun {
		c := makeDefaultShivasCommand()
		c.Flags.StringVar(&c.address, "address", "", "IP address of host")
		c.Flags.BoolVar(&c.skipDNS, "skip-dns", false, "whether to skip updating the DNS")
		registerShivasFlags(c)
		return c
	},
}

// AddDUT contains the arguments for "satlab add dut ...". It also contains additional
// qualified arguments that are the result of adding the satlab prefix to "raw" arguments.
type addDUT struct {
	shivasAddDUT
	// Satlab-specific fields, if any exist, go here.
	address           string
	skipDNS           bool
	qualifiedHostname string
	qualifiedServo    string
	qualifiedRack     string
}

// Run adds a DUT and returns an exit status.
func (c *addDUT) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// InnerRun is the implementation of run.
func (c *addDUT) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	var err error
	dockerHostBoxIdentifier := strings.ToLower(c.commonFlags.SatlabID)
	if dockerHostBoxIdentifier == "" {
		dockerHostBoxIdentifier, err = commands.GetDockerHostBoxIdentifier()
		fmt.Fprintf(os.Stderr, "Unable to determine -satlab prefix, use %s to pass explicitly\n", c.commonFlags.SatlabID)
		return err
	}

	// The qualified name of a rack if no information is given is "satlab-...-rack".
	if c.rack == "" {
		c.rack = "rack"
	}

	satlabPrefix := common.MaybePrepend("satlab", dockerHostBoxIdentifier)
	c.qualifiedHostname = common.MaybePrepend(satlabPrefix, c.hostname)
	c.qualifiedRack = common.MaybePrepend(satlabPrefix, c.rack)
	c.qualifiedServo = common.MaybePrepend(satlabPrefix, c.servo)

	if c.zone == "" {
		c.zone = "satlab"
	}

	if err := (&adders.Rack{
		Rack:      c.qualifiedRack,
		Namespace: c.envFlags.Namespace,
		Zone:      c.zone,
	}).CheckAndUpdate(); err != nil {
		return err
	}

	if err := (&adders.Asset{
		Asset: c.asset,
		Rack:  c.qualifiedRack,
		Zone:  c.zone,
		Model: c.model,
		Board: c.board,
	}).CheckAndUpdate(); err != nil {
		return err
	}

	if err := (&adders.DUT{
		Namespace:  c.envFlags.Namespace,
		Zone:       c.zone,
		Host:       c.qualifiedHostname,
		Servo:      c.qualifiedServo,
		ShivasArgs: makeShivasFlags(c),
	}).CheckAndUpdate(); err != nil {
		return err
	}

	if err := dns.UpdateRecord(
		c.qualifiedHostname,
		c.address,
	); err != nil {
		return err
	}

	return nil
}
