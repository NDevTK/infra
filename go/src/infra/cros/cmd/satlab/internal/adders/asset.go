// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adders

import (
	"fmt"
	"os"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/paths"
)

type Asset struct {
	Asset string
	Rack  string
	Zone  string
	Model string
	Board string
}

func (a *Asset) Run() error {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasPath, "get", "asset"},
		PositionalArgs: []string{a.Asset},
		Flags: map[string][]string{
			"json":  nil,
			"rack":  {a.Rack},
			"zone":  {a.Zone},
			"model": {a.Model},
			"board": {a.Board},
		},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add asset if applicable: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	assetMsgBytes, err := command.Output()
	assetMsg := commands.TrimOutput(assetMsgBytes)
	if err != nil {
		return errors.Annotate(err, "add asset if applicable").Err()
	}

	if len(assetMsg) == 0 {
		// Add the asset.
		fmt.Fprintf(os.Stderr, "Adding asset\n")
		args := (&commands.CommandWithFlags{
			Commands: []string{paths.ShivasPath, "add", "asset"},
			Flags: map[string][]string{
				"model": {a.Model},
				"board": {a.Board},
				"rack":  {a.Rack},
				"zone":  {a.Zone},
				"name":  {a.Asset},
			},
		}).ToCommand()
		fmt.Fprintf(os.Stderr, "Add asset if applicable: run %s\n", args)
		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return errors.Annotate(err, "add asset if applicable").Err()
		}
	} else {
		fmt.Fprintf(os.Stderr, "Asset already added\n")
	}
	return nil
}
