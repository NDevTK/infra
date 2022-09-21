// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
	"infra/cros/internal/cmd"
)

// myjobRunBase contains data for a single `myjob` command run.
type myjobRunBase struct {
	subcommands.CommandRunBase
	branch    string
	staging   bool
	cmdRunner cmd.CommandRunner
}

// addBranchFlag creates a `-branch` command-line flag to specify the branch.
func (m *myjobRunBase) addBranchFlag() {
	m.Flags.StringVar(&m.branch, "branch", "main", "Specify the branch on which to run the builder.")
}

// addStagingFlag creates a `-staging` command-line flag for a myjob subcommand.
func (m *myjobRunBase) addStagingFlag() {
	m.Flags.BoolVar(&m.staging, "staging", false, "Run a staging builder instead of a prod builder.")
}

// RunCmd executes a shell command.
func (m myjobRunBase) RunCmd(ctx context.Context, name string, args ...string) (stdout, stderr string, err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err = m.cmdRunner.RunCommand(ctx, &stdoutBuf, &stderrBuf, "", name, args...)
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		return stdout, stderr, errors.Annotate(err, fmt.Sprintf("running `%s %s`", name, strings.Join(args, " "))).Err()
	}
	return stdout, stderr, nil
}
