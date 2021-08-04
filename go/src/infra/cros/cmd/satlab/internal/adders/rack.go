// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adders

import (
	"fmt"
	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/paths"
	"os"
	"os/exec"

	"go.chromium.org/luci/common/errors"
)

// Rack is a group of arguments for adding a rack.
type Rack struct {
	Rack      string
	Namespace string
	Zone      string
}

// Run adds a rack if it does not already exist.
func (r *Rack) Run() error {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasPath, "get", "rack"},
		PositionalArgs: []string{r.Rack},
		Flags:          nil,
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add rack if applicable: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	rackMsgBytes, err := command.Output()
	rackMsg := commands.TrimOutput(rackMsgBytes)
	if err != nil {
		return errors.Annotate(err, "add rack if applicable").Err()
	}

	if len(rackMsg) == 0 {
		fmt.Fprintf(os.Stderr, "Adding rack\n")
		args := (&commands.CommandWithFlags{
			Commands: []string{paths.ShivasPath, "add", "rack"},
			Flags: map[string][]string{
				// TODO(gregorynisbet): Default to OS for everything.
				"namespace": {r.Namespace},
				"name":      {r.Rack},
				"zone":      {r.Zone},
			},
		}).ToCommand()
		fmt.Fprintf(os.Stderr, "Add rack if applicable: run %s\n", args)
		command := exec.Command(args[0], args[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return errors.Annotate(err, "add rack if applicable").Err()
		}
	} else {
		fmt.Fprintf(os.Stderr, "Rack already added\n")
	}
	return nil
}
