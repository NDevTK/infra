// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package components

import (
	"go.chromium.org/luci/common/errors"
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
