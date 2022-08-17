// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	commonFlags "infra/cmd/mallet/internal/cmd/cmdlib"
	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// SatlabRecovery works on a satlab that is on the same network as you are, but has devices that are invisible to you.
//
// This is a common setup if, for example, you are in an office and there's a satlab and some DUTs on your desk.
// SSHing into the satlab-remote-access container, exec'ing into the drone, and then remaining commands technically works,
// but it is not convenient.
//
// This command automates the above process to some degree.

var SatlabRecovery = &subcommands.Command{
	UsageLine: "satlab-recovery UNIT_NAME",
	ShortDesc: `Recover a DUT connected to a Satlab.`,
	LongDesc: `Recover a DUT connected to a Satlab.

This is primarily intended to test changes to recovery.`,
	CommandRun: func() subcommands.CommandRun {
		return &satlabRecoveryRun{}
	},
}

// satlabRecoveryRun stores the arguments for the satlab recovery run command.
type satlabRecoveryRun struct {
	subcommands.CommandRunBase
	commonFlags.CommonFlags
	authFlags authcli.Flags
	envFlags  site.EnvFlags
}

// Run takes command line arguments and returns an exit status.
func (c *satlabRecoveryRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// innerRun takes command line arguments and returns an error.
func (c *satlabRecoveryRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return errors.New("satlab recovery: not yet implemented")
}
