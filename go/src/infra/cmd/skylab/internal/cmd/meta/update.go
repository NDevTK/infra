// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package meta

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/skylab/internal/site"
)

// Update subcommand: Update skylab tool.
var Update = &subcommands.Command{
	UsageLine: "update",
	ShortDesc: "update skylab tool",
	LongDesc: `Update skylab tool.

If you installed the skylab tool as a part of lab tools, you should
use update_lab_tools instead of this.

This is just a thin wrapper around CIPD.`,
	CommandRun: func() subcommands.CommandRun {
		c := &updateRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		return c
	},
}

type updateRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
}

func (c *updateRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *updateRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	d, err := executableDir()
	if err != nil {
		return err
	}
	root, err := findCIPDRootDir(d)
	if err != nil {
		return err
	}
	cmd := exec.Command("cipd", "ensure", "-root", root, "-ensure-file", "-")
	cmd.Stdin = strings.NewReader("chromiumos/infra/skylab/${platform} latest")
	cmd.Stdout = a.GetOut()
	cmd.Stderr = a.GetErr()
	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), " failed to update packages") {
			fmt.Printf("skylab has insufficient permissions to update itself.\n")
			fmt.Printf("please run 'sudo env PATH=\"$PATH\" skylab update'\n")
		}
		return err
	}
	fmt.Fprintf(a.GetErr(), "%s: You may need to run skylab login again after the update\n", a.GetName())
	fmt.Fprintf(a.GetErr(), "%s: Run skylab whoami to check login status\n", a.GetName())
	return nil
}

// executableDir returns the directory the current executable came
// from.
func executableDir() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", errors.Annotate(err, "get executable directory").Err()
	}
	return filepath.Dir(p), nil
}

func findCIPDRootDir(dir string) (string, error) {
	a, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.Annotate(err, "find CIPD root dir").Err()
	}
	for d := a; d != "/"; d = filepath.Dir(d) {
		if isCIPDRootDir(d) {
			return d, nil
		}
	}
	return "", errors.Reason("find CIPD root dir: no CIPD root above %s", dir).Err()
}

func isCIPDRootDir(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, ".cipd"))
	if err != nil {
		return false
	}
	return fi.Mode().IsDir()
}
