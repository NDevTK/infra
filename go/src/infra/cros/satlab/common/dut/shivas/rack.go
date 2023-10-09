// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"bytes"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/commands"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/satlabcommands"
	e "infra/cros/satlab/common/utils/errors"
	"infra/cros/satlab/common/utils/executor"
)

// Rack is a group of arguments for adding a rack.
type Rack struct {
	Name      string
	Namespace string
	Zone      string
}

// CheckAndAdd runs check and then update if the item does not exist.
func (r *Rack) CheckAndAdd(executor executor.IExecCommander) (string, error) {
	rackMsg, err := r.check(executor)
	if err != nil {
		return "", errors.Annotate(err, "check and update").Err()
	}

	if len(rackMsg) == 0 {
		return r.add(executor)
	} else {
		return "", e.RackExist
	}
}

// check checks if a rack exists.
func (r *Rack) check(executor executor.IExecCommander) (string, error) {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "rack"},
		PositionalArgs: []string{r.Name},
		Flags: map[string][]string{
			"namespace": {r.Namespace},
		},
	}).ToCommand()
	var b bytes.Buffer

	command := exec.Command(args[0], args[1:]...)
	command.Stderr = &b
	rackMsgBytes, err := executor.Exec(command)

	if err != nil {
		return "", errors.Annotate(err, "check rack - %s", b.String()).Err()
	}
	rackMsg := satlabcommands.TrimOutput(rackMsgBytes)

	return rackMsg, nil
}

// add adds a rack unconditionally to UFS.
func (r *Rack) add(executor executor.IExecCommander) (string, error) {
	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "add", "rack"},
		Flags: map[string][]string{
			// TODO(gregorynisbet): Default to OS for everything.
			"namespace": {r.Namespace},
			"name":      {r.Name},
			"zone":      {r.Zone},
		},
	}).ToCommand()

	var b bytes.Buffer
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = &b
	out, err := executor.Exec(command)
	if err != nil {
		return "", errors.Annotate(err, "add rack - %s", b.String()).Err()
	}

	return string(out), nil
}
