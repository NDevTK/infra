// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package delete

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/paths"
)

// Delete deletes a DUT associated with a satlab instance.
func Delete(c *deleteDUT, a subcommands.Application, positionalArgs []string) error {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasPath, "delete", "dut"},
		PositionalArgs: positionalArgs,
		// TODO(gregorynisbet): Consider replacing.
		Flags: nil,
	}).ToCommand()
	command := exec.Command(args[0], args[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return errors.Annotate(
			err,
			fmt.Sprintf(
				"delete dut: running %s",
				strings.Join(args, " "),
			),
		).Err()
	}
	return nil
}
