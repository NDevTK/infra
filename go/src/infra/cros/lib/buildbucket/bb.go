// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package buildbucket

import (
	"context"
	gerr "errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"infra/cros/internal/cmd"
	"infra/cros/internal/util"

	"go.chromium.org/luci/common/errors"
)

type Client struct {
	cmdRunner cmd.CommandRunner
	stdoutLog *log.Logger
	stderrLog *log.Logger
}

// NewClient creates a new Buildbucket client.
func NewClient(cmdRunner cmd.CommandRunner, stdoutLog *log.Logger, stderrLog *log.Logger) *Client {
	return &Client{
		cmdRunner: cmdRunner,
		stdoutLog: stdoutLog,
		stderrLog: stderrLog,
	}
}

// LogOut logs to stdout.
func (c *Client) LogOut(format string, a ...interface{}) {
	if c.stdoutLog != nil {
		c.stdoutLog.Printf(format, a...)
	}
}

// LogErr logs to stderr.
func (c *Client) LogErr(format string, a ...interface{}) {
	if c.stderrLog != nil {
		c.stderrLog.Printf(format, a...)
	}
}

// FakeAuthInfoRunner creates a FakeCommandRunner for `{tool} auth-info` (like bb or led).
func FakeAuthInfoRunner(tool string, exitCode int) cmd.FakeCommandRunner {
	return cmd.FakeCommandRunner{
		ExpectedCmd: []string{tool, "auth-info"},
		FailCommand: exitCode != 0,
		FailError:   createCmdFailError(exitCode),
	}
}

// createCmdFailError creates an error with the desired exit status.
func createCmdFailError(exitCode int) error {
	return exec.Command("bash", "-c", fmt.Sprintf("exit %d", exitCode)).Run()
}

// IsLUCIToolAuthed checks whether the named LUCI CLI tool is logged in.
func (c *Client) IsLUCIToolAuthed(ctx context.Context, tool string) (bool, error) {
	_, stderr, err := c.runCmd(ctx, tool, "auth-info")
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
func (c *Client) EnsureLUCIToolAuthed(ctx context.Context, tool string) error {
	if authed, err := c.IsLUCIToolAuthed(ctx, tool); err != nil {
		return errors.Annotate(err, fmt.Sprintf("determining whether `%s` is authed", tool)).Err()
	} else if !authed {
		return fmt.Errorf("%s CLI not logged in. Please run `%s auth-login`, then try again.", tool, tool)
	}
	return nil
}

// EnsureLUCIToolsAuthed ensures that multiple LUCI CLI tools are logged in.
// If any tools are not authed, it will return an error instructing the user to log into each unauthed tool.
func (c *Client) EnsureLUCIToolsAuthed(ctx context.Context, tools ...string) error {
	var unauthedTools []string
	var authCommands []string
	for _, tool := range tools {
		if authed, err := c.IsLUCIToolAuthed(ctx, tool); err != nil {
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

// runBBCmd runs a `bb` subcommand.
func (c *Client) runBBCmd(ctx context.Context, dryRun bool, subcommand string, args ...string) (stdout, stderr string, err error) {
	if dryRun {
		c.LogOut("would have run `bb %s`", strings.Join(args, " "))
		return "", "", nil
	}
	return c.runCmd(ctx, "bb", util.PrependString(subcommand, args)...)
}

// BBAdd runs a `bb add` command, and prints stdout to the user. Returns the
// bbid of the build and an error (if any).
func (c *Client) BBAdd(ctx context.Context, dryRun bool, args ...string) (string, error) {
	stdout, stderr, err := c.runBBCmd(ctx, dryRun, "add", args...)
	if err != nil {
		c.LogErr(stderr)
		return "", err
	}
	if dryRun {
		return "dry_run_bbid", nil
	}
	c.LogOut("\n" + strings.Split(stdout, "\n")[0])
	bbidRegexp := regexp.MustCompile(`http:\/\/ci.chromium.org\/b\/(?P<bbid>\d+) `)
	matches := bbidRegexp.FindStringSubmatch(stdout)
	if matches == nil {
		return "", gerr.New("could not parse BBID from `bb add` stdout.")
	}
	return matches[1], nil
}

// getBuilders runs the `bb builders` command to get all builders in the given bucket.
// The bucket param should not include the project prefix (normally "chromeos/").
func (c *Client) BBBuilders(ctx context.Context, bucket string) ([]string, error) {
	stdout, stderr, err := c.runBBCmd(ctx, false, "builders", fmt.Sprintf("chromeos/%s", bucket))
	if err != nil {
		c.LogErr(stderr)
		return []string{}, err
	}
	return strings.Split(stdout, "\n"), nil
}
