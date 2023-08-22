// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"fmt"
	"os"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/satlab/internal/commands"
)

// Rack is a group of arguments for adding a rack.
type Rack struct {
	Name      string
	Namespace string
	Zone      string
}

// CheckAndAdd runs check and then update if the item does not exist.
func (r *Rack) CheckAndAdd() error {
	rackMsg, err := r.check()
	if err != nil {
		return errors.Annotate(err, "check and update").Err()
	}
	if len(rackMsg) == 0 {
		return r.add()
	} else {
		fmt.Fprintf(os.Stderr, "Rack already added\n")
	}
	return nil
}

// Check checks if a rack exists.
func (r *Rack) check() (string, error) {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "rack"},
		PositionalArgs: []string{r.Name},
		Flags: map[string][]string{
			"namespace": {r.Namespace},
		},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add rack: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	rackMsgBytes, err := command.Output()
	rackMsg := satlabcommands.TrimOutput(rackMsgBytes)
	if err != nil {
		return "", errors.Annotate(err, "add rack").Err()
	}
	return rackMsg, nil
}

// Add adds a rack unconditionally to UFS.
func (r *Rack) add() error {
	fmt.Fprintf(os.Stderr, "Adding rack\n")
	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "add", "rack"},
		Flags: map[string][]string{
			// TODO(gregorynisbet): Default to OS for everything.
			"namespace": {r.Namespace},
			"name":      {r.Name},
			"zone":      {r.Zone},
		},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add rack: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	return errors.Annotate(err, "add rack").Err()
}
