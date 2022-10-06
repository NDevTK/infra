// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
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
	stdoutLog *log.Logger
	stderrLog *log.Logger
	bbAddArgs []string
	cmdRunner cmd.CommandRunner

	dryrun  bool
	branch  string
	staging bool
	// Patches of the form of "crrev.com/c/1234567", "crrev.com/i/1234567".
	patches      list
	buildTargets list
	buildspec    string
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

// addBuildTargetsFlag creates a `-build_targets` command-line flag for a try subcommand.
func (m *tryRunBase) addBuildTargetsFlag() {
	m.Flags.Var(&m.buildTargets, "build_targets", "(comma-separated) Build targets to run. If not set, the standard set of build targets will be used.")
}

// addBuildspecFlag creates a `-buildspec` command-line flag for a try command.
func (m *tryRunBase) addBuildspecFlag() {
	m.Flags.StringVar(&m.buildspec, "buildspec", "", "GS uri to the buildspec that the builder should sync to, e.g. gs://chromeos-manifest-versions/buildspecs/108/15159.0.0.xml.")
}

// addDryrunFlag creates a `-dryrun` command-line flag for a try command.
func (m *tryRunBase) addDryrunFlag() {
	m.Flags.BoolVar(&m.dryrun, "dryrun", false, "Dry run (i.e. don't actually run `bb add`).")
}

// validate validates base args for the command.
func (m *tryRunBase) validate() error {
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

	if m.buildspec != "" {
		if !m.staging {
			return fmt.Errorf("--buildspec is only supported with --staging")
		}
		if !strings.HasPrefix(m.buildspec, "gs://") {
			return fmt.Errorf("--buildspec must start with gs://")
		}
	}

	return nil
}

// LogOut logs to stdout.
func (m *tryRunBase) LogOut(format string, a ...interface{}) {
	if m.stdoutLog != nil {
		m.stdoutLog.Printf(format, a...)
	}
}

// LogErr logs to stderr.
func (m *tryRunBase) LogErr(format string, a ...interface{}) {
	if m.stderrLog != nil {
		m.stderrLog.Printf(format, a...)
	}
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
