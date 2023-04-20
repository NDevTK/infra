// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/satlab/internal/commands"
	"infra/cros/satlab/satlab/internal/paths"
	"infra/cros/satlab/satlab/internal/site"
)

// commandRunnerFunc is a type allowing us to monkey patch command execution
// for testing.
type commandRunnerFunc func(*exec.Cmd) error

// execCommand is a function of type `commandRunnerFunc` that just calls the
// existing Cmd.Run().
func execCommand(c *exec.Cmd) error {
	return c.Run()
}

// commandRunner is a package level variable controlling the behavior of
// executing commands. Should be overridden when testing.
var commandRunner commandRunnerFunc = execCommand

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

// Run adds a DUT if it does not already exist.
func (d *DUT) CheckAndAdd() error {
	dutMsg, err := d.check()
	if err != nil {
		return errors.Annotate(err, "check and update").Err()
	}
	if len(dutMsg) == 0 {
		return d.add()
	} else {
		fmt.Fprintf(os.Stderr, "DUT already added\n")
	}
	return nil
}

// Check checks for the existnce of a UFS DUT.
func (d *DUT) check() (string, error) {
	flags := map[string][]string{
		"namespace": {d.Namespace},
		"zone":      {d.Zone},
	}

	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "dut"},
		Flags:          flags,
		PositionalArgs: []string{d.Name},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add dut: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = os.Stderr
	var stdout bytes.Buffer
	command.Stdout = &stdout

	err := commandRunner(command)

	dutMsg := commands.TrimOutput(stdout.Bytes())
	if err != nil {
		return "", errors.Annotate(err, "check DUT in UFS: running %s", strings.Join(args, " ")).Err()
	}
	return dutMsg, nil
}

// Add a DUT to UFS.
func (d *DUT) add() error {
	fmt.Fprintf(os.Stderr, "Adding DUT\n")

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
		Commands: []string{paths.ShivasCLI, "add", "dut"},
		Flags:    flags,
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "Add dut: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := commandRunner(command)
	return errors.Annotate(
		err,
		fmt.Sprintf(
			"add dut: running %s",
			strings.Join(args, " "),
		),
	).Err()
}
