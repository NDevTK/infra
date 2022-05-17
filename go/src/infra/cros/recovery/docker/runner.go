// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package docker

// TODO: Move package to common lib when developing finished.

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/log"
)

// runResult holds info of execution.
type runResult struct {
	FinishedOnTime bool
	ExitCode       int
	Stdout         string
	Stderr         string
}

// runWithTimeout runs command with a time limit.
//
// runResult.FinishedOnTime will be set to true if and only if a command was run and it completed within the limit.
// err is nil if and only if the command actually ran, completed with the time limit, and exited successfully.
//
// Note that exceeding the timeout does not cause the process that was started to stop.
func runWithTimeout(ctx context.Context, timeout time.Duration, command string, args ...string) (res runResult, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cw := make(chan error, 1)
	var se, so bytes.Buffer
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stderr = &se
	cmd.Stdout = &so
	defer func() {
		res.Stdout = so.String()
		res.Stderr = se.String()
	}()
	go func() {
		log.Debugf(ctx, "Run cmd: %s", cmd)
		cw <- cmd.Run()
	}()
	select {
	case e := <-cw:
		res.FinishedOnTime = true
		if exitError, ok := e.(*exec.ExitError); ok {
			res.ExitCode = exitError.ExitCode()
		} else if e != nil {
			res.ExitCode = 1
		}
		err = errors.Annotate(e, "run with timeout %s", timeout).Err()
		return
	case <-ctx.Done():
		res.ExitCode = 124
		err = errors.Reason("run with timeout %s: exceeded timeout", timeout).Err()
		return
	}
}
