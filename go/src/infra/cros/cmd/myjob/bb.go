// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"os/exec"

	"go.chromium.org/luci/common/errors"
)

// IsBBAuthed checks whether the `bb` CLI is authed.
func (m myjobRunBase) IsBBAuthed(ctx context.Context) (bool, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err := m.RunCmd(ctx, &stdoutBuf, &stderrBuf, "", "bb", "auth-info")
	if err == nil {
		// If bb is authed, then `bb auth-info` returns exit status 0, and Run() should give no error.
		return true, nil
	} else if exiterr, ok := err.(*exec.ExitError); ok && exiterr.ExitCode() == 1 {
		// If bb is not authed, then `bb auth-info` returns exit status 1.
		return false, nil
	}
	// Other problems return other exit statuses:
	// `bb fake-subcommand` returns exit status 2;
	// if the `bb` command is not found, then it'll return 127.
	return false, errors.Annotate(err, "running `bb auth-info`").Err()
}
