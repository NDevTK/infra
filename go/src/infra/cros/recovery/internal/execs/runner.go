// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/durationpb"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/tlw"
)

var (
	// ErrCodeTag is the key value pair for storing the error code for the linux command.
	ErrCodeTag = errors.NewTagKey("error_code")

	// StdErrTag is the key value pair for storing the error code
	// associated with the standard error
	StdErrTag = errors.NewTagKey("std_error")

	// 127: linux command line error of command not found.
	SSHErrorCLINotFound = errors.BoolTag{Key: errors.NewTagKey("ssh_error_cli_not_found")}

	// 124: linux command line error of command timeout.
	SSHErrorLinuxTimeout = errors.BoolTag{Key: errors.NewTagKey("linux_timeout")}

	// other linux error tag.
	GeneralError = errors.BoolTag{Key: errors.NewTagKey("general_error")}

	// internal error tag.
	SSHErrorInternal = errors.BoolTag{Key: errors.NewTagKey("ssh_error_internal")}

	// -1: fail to create ssh session.
	FailToCreateSSHErrorInternal = errors.BoolTag{Key: errors.NewTagKey("fail_to_create_ssh_error_internal")}

	// -2: session is down, but the server sends no confirmation of the exit status.
	NoExitStatusErrorInternal = errors.BoolTag{Key: errors.NewTagKey("no_exit_status_error_internal")}

	// other internal error tag.
	OtherErrorInternal = errors.BoolTag{Key: errors.NewTagKey("other_error_internal")}
)

// Runner defines the type for a function that will execute a command
// on a host, and returns the result as a single line.
// TODO: Remove as we do not need extra type.
type Runner = components.Runner

// NewBackgroundRunner returns runner for requested resource specified
// per plan.
//
// TODO(b/222698101): At this time this method is a
// placeholder. This will eventually be replaced with an
// implementation that will submit a command for background execution,
// and will return without waiting for it to complete.
func (ei *ExecInfo) NewBackgroundRunner(resource string) components.Runner {
	return ei.NewRunner(resource)
}

// DefaultRunner returns runner for current resource name specified per plan.
func (ei *ExecInfo) DefaultRunner() components.Runner {
	return ei.NewRunner(ei.GetActiveResource())
}

// NewRunner returns a function of type Runner that executes a command
// on a host and returns the results as a single line. This function
// defines the specific host on which the command will be
// executed. Examples of such specific hosts can be the DUT, or the
// servo-host etc.
func (a *ExecInfo) NewRunner(host string) components.Runner {
	runner := func(ctx context.Context, timeout time.Duration, cmd string, args ...string) (string, error) {
		fullCmd := cmd
		if len(args) > 0 {
			fullCmd += " " + strings.Join(args, " ")
		}
		log := a.NewLogger()
		log.Debugf("Prepare to run command: %q", fullCmd)
		r := a.GetAccess().Run(ctx, &tlw.RunRequest{
			Resource: host,
			Timeout:  durationpb.New(timeout),
			Command:  cmd,
			Args:     args,
		})
		log.Debugf("Run %q completed with exit code %d", r.Command, r.ExitCode)
		exitCode := r.ExitCode
		out := strings.TrimSpace(r.Stdout)
		log.Debugf("Run output:\n%s", out)
		if exitCode != 0 {
			errAnnotator := errors.Reason("runner: command %q completed with exit code %d", r.Command, r.ExitCode)
			errCodeTagValue := errors.TagValue{Key: ErrCodeTag, Value: exitCode}
			errAnnotator.Tag(errCodeTagValue)
			errAnnotator.Tag(errors.TagValue{Key: StdErrTag, Value: r.Stderr})
			log.Debugf("Run stderr:\n%s", r.Stderr)
			// different kinds of internal errors
			if exitCode < 0 {
				errAnnotator.Tag(SSHErrorInternal)
				if exitCode == -1 {
					errAnnotator.Tag(FailToCreateSSHErrorInternal)
				} else if exitCode == -2 {
					errAnnotator.Tag(NoExitStatusErrorInternal)
				} else if exitCode == -3 {
					errAnnotator.Tag(OtherErrorInternal)
				}
				// general linux errors
			} else if exitCode == 124 {
				errAnnotator.Tag(SSHErrorLinuxTimeout)
			} else if exitCode == 127 {
				errAnnotator.Tag(SSHErrorCLINotFound)
			} else {
				errAnnotator.Tag(GeneralError)
			}
			return out, errAnnotator.Err()
		}
		return out, nil
	}
	return runner
}
