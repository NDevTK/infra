// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"fmt"
	"io"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/commands"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/common/utils/misc"
)

// DUT contains all the information necessary to add a DUT.
type DUT struct {
	// TODO(gregorynisbet): remove the namespace field.
	Namespace  string
	Zone       string
	Name       string
	Servo      string
	Rack       string
	ShivasArgs map[string][]string
}

// CheckAndAdd adds a DUT if it does not already exist.
func (d *DUT) CheckAndAdd(executor executor.IExecCommander, w io.Writer) (bool, error) {
	exists, err := d.check(executor, w)
	if err != nil {
		return false, errors.Annotate(err, "check and update").Err()
	}
	if !exists {
		return false, d.add(executor, w)
	} else {
		fmt.Fprintf(w, "DUT already added\n\n")
	}
	return true, nil
}

// Check checks for the existnce of a UFS DUT.
func (d *DUT) check(executor executor.IExecCommander, w io.Writer) (bool, error) {
	flags := map[string][]string{
		"namespace": {d.Namespace},
		"zone":      {d.Zone},
	}

	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "dut"},
		Flags:          flags,
		PositionalArgs: []string{d.Name},
		AuthRequired:   true,
	}).ToCommand()
	fmt.Fprintf(w, "Check dut exists: run %s\n", args)

	command := exec.Command(args[0], args[1:]...)
	// Don't use `CombinedOutput` here because it returns
	// `rpc error: code = NotFound` that means the DUT doesn't exist.
	// As we check only the length of output.
	stdout, err := executor.Output(command)

	if err != nil {
		return false, errors.Annotate(err, "add dut").Err()
	}

	// if DUT not found, shivas returns output in stderr, and stdout is empty.
	exists := (len(stdout) != 0)

	return exists, nil
}

// Add a DUT to UFS.
func (d *DUT) add(executor executor.IExecCommander, w io.Writer) error {
	fmt.Fprintf(w, "Adding DUT\n")

	flags := make(map[string][]string)
	for k, v := range d.ShivasArgs {
		flags[k] = v
	}

	flags["name"] = []string{d.Name}
	// This flag must have the form labstation:port.
	// Do not validate this flag here since we don't want to potentially drift
	// out of sync with the format that shivas expects.
	// TODO(gregorynisbet): Consider pre-populating it.
	flags["servo"] = []string{d.Servo}

	// These flags control where the deploy task is run.
	flags["deploy-project"] = []string{site.GetLUCIProject()}
	flags["deploy-bucket"] = []string{site.GetDeployBucket()}

	// TODO(gregorynisbet): Consider a different strategy for tracking flags
	// that cannot be passed to shivas add dut.
	args := (&commands.CommandWithFlags{
		Commands:     []string{paths.ShivasCLI, "add", "dut"},
		Flags:        flags,
		AuthRequired: true,
	}).ToCommand()
	fmt.Fprintf(w, "Add dut: run %s\n", args)

	command := exec.Command(args[0], args[1:]...)
	out, err := executor.CombinedOutput(command)

	fmt.Fprintln(w, misc.TrimOutput(out))

	return err
}
