// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"fmt"
	"infra/cros/cmd/satlab/internal/adders"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/common"
	"os"
	"strings"

	"github.com/maruel/subcommands"
)

func Add(c *addDUT, a subcommands.Application, args []string) error {
	var err error
	dockerHostBoxIdentifier := strings.ToLower(c.satlabID)
	if dockerHostBoxIdentifier == "" {
		dockerHostBoxIdentifier, err = commands.GetDockerHostBoxIdentifier()
		fmt.Fprintf(os.Stderr, "Unable to determine satlab prefix, use %s to pass explicitly\n", common.SatlabID)
		return err
	}

	satlabPrefix := common.MaybePrepend("satlab-%s", dockerHostBoxIdentifier)

	if c.rack == "" {
		c.rack = common.MaybePrepend(satlabPrefix, "rack")
	} else {
		c.rack = common.MaybePrepend(satlabPrefix, c.rack)
	}

	if c.zone == "" {
		c.zone = "satlab"
	}

	if err := (&adders.Rack{
		Rack:      c.rack,
		Namespace: c.namespace,
	}).Run(); err != nil {
		return err
	}

	if err := (&adders.Asset{
		Asset: c.asset,
		Rack:  c.rack,
		Zone:  c.zone,
		Model: c.model,
		Board: c.board,
	}).Run(); err != nil {
		return err
	}

	if err := (&adders.DUT{
		ShivasArgs: makeShivasFlags(c),
	}).Run(); err != nil {
		return err
	}

	return nil
}
