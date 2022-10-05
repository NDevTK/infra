// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"infra/cros/internal/cmd"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"
)

type list []string

func (l *list) Set(value string) error {
	*l = strings.Split(strings.TrimSpace(value), ",")
	return nil
}

func (l *list) String() string {
	return strings.Join(*l, ",")
}

// tryRunBase contains data for a single `try` command run.
type tryRunBase struct {
	subcommands.CommandRunBase
	gitCookiesPath string
	branch         string
	staging        bool
	// Patches of the form of "crrev.com/c/1234567", "crrev.com/i/1234567".
	patches   list
	bbAddArgs []string
	cmdRunner cmd.CommandRunner
}

// addBranchFlag creates a `-branch` command-line flag to specify the branch.
func (m *tryRunBase) addBranchFlag() {
	m.Flags.StringVar(&m.branch, "branch", "main", "specify the branch on which to run the builder")
}

// addStagingFlag creates a `-staging` command-line flag for a try subcommand.
func (m *tryRunBase) addStagingFlag() {
	m.Flags.BoolVar(&m.staging, "staging", false, "run a staging builder instead of a prod builder")
}

// addStagingFlag creates a `-staging` command-line flag for a try subcommand.
func (m *tryRunBase) addPatchesFlag() {
	m.Flags.Var(&m.patches, "gerrit-patches", "(comma-separated) patches to apply to the build, e.g. crrev.com/c/1234567,crrev.com/i/1234567.")
	m.Flags.Var(&m.patches, "g", "alias for --gerrit-patches")
}

// validate validates base args for the command.
func (m *tryRunBase) validate(ctx context.Context) error {
	if len(m.patches) > 0 {
		patchSpec := regexp.MustCompile(`^crrev\.com\/[ci]\/\d{7,8}$`)
		for _, patch := range m.patches {
			if !patchSpec.MatchString(patch) {
				return fmt.Errorf(`invalid patch "%s". patches must be of the format crrev.com/[ci]/<number>.`, patch)
			}
		}

		if !m.staging {
			return fmt.Errorf("-g/--gerrit-patches is only supported with --staging")
		}
	}
	return nil
}

// RunCmd executes a shell command.
func (m tryRunBase) RunCmd(ctx context.Context, name string, args ...string) (stdout, stderr string, err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err = m.cmdRunner.RunCommand(ctx, &stdoutBuf, &stderrBuf, "", name, args...)
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		return stdout, stderr, errors.Annotate(err, fmt.Sprintf("running `%s %s`", name, strings.Join(args, " "))).Err()
	}
	return stdout, stderr, nil
}
