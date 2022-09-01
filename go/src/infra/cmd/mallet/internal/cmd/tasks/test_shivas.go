// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
)

// mallet test-shivas runs an integration test on shivas.
//
// It uses the `./dev-shivas` (which is `shivas` built with a dev tag) to manipulate
// a UFS instance running locally.
//
// The `dev-shivas` target is incapable of manipulating prod data, so this tool, too, should
// only be capable of talking to the dev shivas.

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
	if len(args) != 0 {
		return errors.Reason("test shivas: positional arguments are not supported").Err()
	}
	if c.dir == "" {
		return errors.Reason("test shivas: argument -dir must be provided").Err()
	}
	if err := exec.Command(filepath.Join(c.dir, "dev-shivas"), "help").Run(); err != nil {
		return errors.Annotate(err, `test shivas: "dev-shivas help" failed`).Err()
	}
	fmt.Fprintf(a.GetErr(), "OK\n")
	return nil
}
