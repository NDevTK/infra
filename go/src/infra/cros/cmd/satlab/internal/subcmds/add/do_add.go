// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"fmt"
	"infra/cros/cmd/satlab/internal/adders"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/commands/dns"
	"infra/cros/cmd/satlab/internal/common"
	"os"
	"strings"

	"github.com/maruel/subcommands"
)

// Add is the implementation of the add command.
func Add(c *addDUT, a subcommands.Application, args []string) error {
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
	}).Run(); err != nil {
		return err
	}

	if err := (&adders.Asset{
		Asset: c.asset,
		Rack:  c.qualifiedRack,
		Zone:  c.zone,
		Model: c.model,
		Board: c.board,
	}).Run(); err != nil {
		return err
	}

	if err := (&adders.DUT{
		Namespace:  c.envFlags.Namespace,
		Zone:       c.zone,
		Host:       c.qualifiedHostname,
		Servo:      c.qualifiedServo,
		ShivasArgs: makeShivasFlags(c),
	}).Run(); err != nil {
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
