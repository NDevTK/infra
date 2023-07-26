// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/satlab/internal/commands"
	"infra/cros/satlab/satlab/internal/paths"
)

// DUTRepairer repairs a DUT with the given name.
type DUTRepairer struct {
	Name string
	// ShivasArgs map[string][]string
}

// repair invokes shivas with the required arguments to repair a DUT.
func (u *DUTRepairer) Repair() error {
	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "repair-duts", "-verify", u.Name},
	}).ToCommand()
	fmt.Fprintf(os.Stderr, "repair dut: run %s\n", args)
	command := exec.Command(args[0], args[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	return errors.Annotate(
		err,
		fmt.Sprintf(
			"repair dut: running %s",
			strings.Join(args, " "),
		),
	).Err()
}
