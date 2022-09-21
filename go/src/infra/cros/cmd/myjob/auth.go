// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// IsLUCIToolAuthed checks whether the named LUCI CLI tool is logged in.
func (m myjobRunBase) IsLUCIToolAuthed(ctx context.Context, tool string) (bool, error) {
	_, stderr, err := m.RunCmd(ctx, tool, "auth-info")
	if err == nil {
		// If the tool is authed, then `auth-info` returns exit status 0, and Run() should give no error.
		return true, nil
	} else if exiterr, ok := errors.Unwrap(err).(*exec.ExitError); ok && exiterr.ExitCode() == 1 {
		// If the tool is not authed, then `auth-info` returns exit status 1.
		return false, nil
	}
	// Other problems return other exit statuses:
	// `bb fake-subcommand` and `led fake-subcommand` return exit status 2;
	// if the tool is altogether not found, then it'll return 127.
	fmt.Println(stderr)
	return false, err
}

// EnsureLUCIToolAuthed checks whether the named LUCI CLI tool is logged in.
// If not, it instructs the user to log in, and returns an error.
func (m myjobRunBase) EnsureLUCIToolAuthed(ctx context.Context, tool string) error {
	if authed, err := m.IsLUCIToolAuthed(ctx, tool); err != nil {
		return errors.Annotate(err, fmt.Sprintf("determining whether `%s` is authed", tool)).Err()
	} else if !authed {
		return fmt.Errorf("%s CLI not logged in. Please run `%s auth-login`, then try again.", tool, tool)
	}
	return nil
}

// EnsureLUCIToolsAuthed ensures that multiple LUCI CLI tools are logged in.
// If any tools are not authed, it will return an error instructing the user to log into each unauthed tool.
func (m myjobRunBase) EnsureLUCIToolsAuthed(ctx context.Context, tools ...string) error {
	var unauthedTools []string
	var authCommands []string
	for _, tool := range tools {
		if authed, err := m.IsLUCIToolAuthed(ctx, tool); err != nil {
			return err
		} else if !authed {
			unauthedTools = append(unauthedTools, tool)
			authCommands = append(authCommands, fmt.Sprintf("%s auth-login", tool))
		}
	}
	if len(unauthedTools) != 0 {
		return fmt.Errorf(
			"The following tools were not logged in: %s. Please run the following commands, then try again:\n\t%s",
			strings.Join(unauthedTools, ", "),
			strings.Join(authCommands, "\n\t"))
	}
	return nil
}
