// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

// TODO(gregorynisbet): Validate existence of required flags.

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/satlab/internal/commands"
	"infra/cros/cmd/satlab/internal/parse"
	"infra/cros/cmd/satlab/internal/paths"
)

// DeleteDUT deletes a DUT. It takes a possibly-empty path to the service account credentials,
// the prefix for the satlab box in question, and the result of parsing the command line flags given to the
// command initially.
func DeleteDUT(serviceAccountJSON string, satlabPrefix string, p *parse.CommandParseResult) error {
	if p == nil {
		return errors.New("command parse cannot be nil")
	}

	positionalArgs := []string{}
	for _, item := range p.PositionalArgs {
		positionalArgs = append(positionalArgs, fmt.Sprintf("%s-%s", satlabPrefix, item))
	}
	flags := make(map[string][]string)
	for k, v := range p.Flags {
		flags[k] = []string{v}
	}
	for k := range p.NullaryFlags {
		flags[k] = nil
	}

	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasPath, "delete", "dut"},
		PositionalArgs: positionalArgs,
		Flags:          flags,
	}).ApplyFlagFilter(true, map[string]bool{
		"satlab-id": false,
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
