// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"time"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/types/known/durationpb"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/tlw"
)

type routerSshRunner struct {
	sshAccess           ssh.Access
	resource            string
	sshUsernameOverride string
}

func newRouterSshRunner(sshAccess ssh.Access, resource string, deviceType labapi.WifiRouterDeviceType) ssh.Runner {
	r := &routerSshRunner{
		sshAccess: sshAccess,
		resource:  resource,
	}
	if deviceType == labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_ASUSWRT {
		r.sshUsernameOverride = asusWrtSshUser
	}
	return r
}

// Run executes the given command with its arguments on a host over ssh. The
// stdout of the command execution is returned. If the command returns a
// non-zero exit code, a non-nil error is returned.
//
// Can be used as a components.Runner.
func (r *routerSshRunner) Run(ctx context.Context, timeout time.Duration, cmd string, args ...string) (string, error) {
	runResult := r.RunForResult(ctx, timeout, false, cmd, args...)
	if runResult.GetExitCode() != 0 {
		return "", errors.Reason("run of command %q result returned non-zero exit code: %s", runResult.GetCommand(), runResult.GetStderr()).Err()
	}
	return runResult.GetStdout(), nil
}

// RunInBackground starts to execute the given command with its arguments on a
// host over ssh, but does so in a background process and does not wait for it
// to finish before returning. A non-nil error is returned when it fails to
// start the command execution.
func (r *routerSshRunner) RunInBackground(ctx context.Context, timeout time.Duration, cmd string, args ...string) error {
	runResult := r.RunForResult(ctx, timeout, true, cmd, args...)
	if runResult.GetExitCode() != 0 {
		return errors.Reason("background run of command %q result returned non-zero exit code: %s", runResult.GetCommand(), runResult.GetStderr()).Err()
	}
	return nil
}

// RunForResult executes the given command with its arguments on a host over
// ssh. A RunResult is returned, which can be used to determine the success
// of the command execution.
func (r *routerSshRunner) RunForResult(ctx context.Context, timeout time.Duration, inBackground bool, cmd string, args ...string) ssh.RunResult {
	return r.sshAccess.Run(ctx, &tlw.RunRequest{
		SshUsername:  r.sshUsernameOverride,
		Resource:     r.resource,
		Timeout:      durationpb.New(timeout),
		InBackground: inBackground,
		Command:      cmd,
		Args:         args,
	})
}
