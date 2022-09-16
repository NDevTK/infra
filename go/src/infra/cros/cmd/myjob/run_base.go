// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"io"

	"github.com/maruel/subcommands"
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

// RunCmd execs (or mocks) a shell command.
func (m myjobRunBase) RunCmd(ctx context.Context, stdoutBuf, stderrBuf io.Writer, dir, name string, args ...string) error {
	return m.cmdRunner.RunCommand(ctx, stdoutBuf, stderrBuf, dir, name, args...)
}
