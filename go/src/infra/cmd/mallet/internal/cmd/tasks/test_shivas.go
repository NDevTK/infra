// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// Perform an integration test on shivas
var TestShivas = &subcommands.Command{
	UsageLine: "test-shivas",
	ShortDesc: "Test shivas CLI",
	LongDesc:  "Test shivas CLI",
	CommandRun: func() subcommands.CommandRun {
		c := &testShivasRun{}
		c.Flags.StringVar(&c.dir, "dir", "", `directory where shivas command is located`)
		return c
	},
}

type testShivasRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  site.EnvFlags

	dir string
}

func (c *testShivasRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *testShivasRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return errors.Reason("test shivas: not yet implemented").Err()
}
