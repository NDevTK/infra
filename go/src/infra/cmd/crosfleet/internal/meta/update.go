// Copyright 2021 The Chromium Authors. All rights reserved.
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

	"infra/cmd/crosfleet/internal/common"
	"infra/cmd/crosfleet/internal/site"
)

// crosfleetLatest is a fragment of a cipd manifest that is used to install the latest version of the crosfleet
// command line tool.
const crosfleetLatest = "chromiumos/infra/crosfleet/${platform} latest"

// Update subcommand: Update crosfleet tool.
var Update = &subcommands.Command{
	UsageLine: "update",
	ShortDesc: "update crosfleet tool",
	LongDesc: `Update crosfleet tool.

This is just a thin wrapper around CIPD.`,
	CommandRun: func() subcommands.CommandRun {
		c := &updateRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.printer.Register(&c.Flags)
		return c
	},
}

type updateRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	printer   common.CLIPrinter
}

func (c *updateRun) Run(a subcommands.Application, _ []string, _ subcommands.Env) int {
	if err := c.innerRun(a); err != nil {
		c.printer.WriteTextStderr("%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *updateRun) innerRun(a subcommands.Application) error {
	d, err := executableDir()
	if err != nil {
		return err
	}
	root, err := findCIPDRootDir(d)
	if err != nil {
		return err
	}

	if err := cipdEnsureLatest(a, root, &c.printer); err != nil {
		return err
	}
	c.printer.WriteTextStderr("%s: You may need to run crosfleet login again after the update", a.GetName())
	c.printer.WriteTextStderr("%s: Run crosfleet whoami to check login status", a.GetName())
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

// cipdEnsureLatest takes an application and a directory and runs a command with
// arguments that will read a cipd manifest from stdin and then run "ensure".
//
// Without this function, you need to run `sudo env PATH="$PATH" crosfleet update` in order to update
// crosfleet if crosfleet was installed as root.
//
// cipdEnsureLatest assumes that the directory exists and that the [[dir]]/.cipd directory
// exists.
func cipdEnsureLatest(a subcommands.Application, dir string, printer *common.CLIPrinter) error {
	// We create two runnable command objects that update the cipd directory.
	// One runs as the current user and the other always runs as root.
	// If the command that runs as the current user fails, then we try the second command.
	asSelf := exec.Command("cipd", "ensure", "-root", dir, "-ensure-file", "-")
	asSelf.Stdin = strings.NewReader(crosfleetLatest)
	asSelf.Stdout = printer.GetOut()
	asSelf.Stderr = printer.GetErr()
	// Windows does not support sudo
	pathvar := fmt.Sprintf("PATH=%s", os.Getenv("PATH"))
	asRootUnix := exec.Command("sudo", "/usr/bin/env", pathvar, "cipd", "ensure", "-root", dir, "-ensure-file", "-")
	asRootUnix.Stdin = strings.NewReader(crosfleetLatest)
	asRootUnix.Stdout = printer.GetOut()
	asRootUnix.Stderr = printer.GetErr()

	if err := asSelf.Run(); err == nil {
		return nil
	}

	// We unconditionally run `sudo` on all OS's, however, we expect it to fail on Windows.
	printer.WriteTextStderr("Retrying as root. Updating crosfleet through cipd.\n")
	if err := asRootUnix.Run(); err != nil {
		return fmt.Errorf("updating cipd as root: %s", err)
	}
	return nil
}
