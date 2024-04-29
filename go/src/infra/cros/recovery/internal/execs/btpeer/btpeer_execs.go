// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/btpeer"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// setStateBrokenExec sets state as BROKEN.
func setStateBrokenExec(ctx context.Context, info *execs.ExecInfo) error {
	if h, err := activeHost(info.GetActiveResource(), info.GetChromeos()); err != nil {
		return errors.Annotate(err, "set state broken").Err()
	} else {
		h.State = tlw.BluetoothPeer_BROKEN
	}
	return nil
}

// setStateWorkingExec sets state as WORKING.
func setStateWorkingExec(ctx context.Context, info *execs.ExecInfo) error {
	if h, err := activeHost(info.GetActiveResource(), info.GetChromeos()); err != nil {
		return errors.Annotate(err, "set state working").Err()
	} else {
		h.State = tlw.BluetoothPeer_WORKING
	}
	return nil
}

// rebootExec reboots the device over ssh and waits for the device to become
// ssh-able again.
func rebootExec(ctx context.Context, info *execs.ExecInfo) error {
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	if err := ssh.Reboot(ctx, sshRunner, 10*time.Second, 10*time.Second, 3*time.Minute); err != nil {
		return errors.Annotate(err, "failed to reboot btpeer").Err()
	}
	return nil
}

// assertUptimeIsLessThanDurationExec checks the uptime of the device and
// fails if the uptime is not less than the duration in minutes provided in
// the "duration_min" action arg.
func assertUptimeIsLessThanDurationExec(ctx context.Context, info *execs.ExecInfo) error {
	// Parse duration arg.
	actionArgs := info.GetActionArgs(ctx)
	const durationMinArgKey = "duration_min"
	durationArg := actionArgs.AsDuration(ctx, durationMinArgKey, 24*60, time.Minute)

	// Get uptime from device.
	sshRunner := btpeer.NewSshRunner(info.GetAccess(), info.GetActiveResource())
	uptime, err := cros.Uptime(ctx, sshRunner.Run)
	if err != nil {
		return errors.Annotate(err, "assert uptime is less than duration: failed to get uptime from device").Err()
	}

	// Evaluate assertion.
	if !(*uptime < durationArg) {
		return errors.Reason("assert uptime is less than duration: device uptime of %s is not less than %s", *uptime, durationArg).Err()
	}
	log.Debugf(ctx, "Device uptime of %s is less than %s", *uptime, durationArg)
	return nil
}

func init() {
	execs.Register("btpeer_state_broken", setStateBrokenExec)
	execs.Register("btpeer_state_working", setStateWorkingExec)
	execs.Register("btpeer_reboot", rebootExec)
	execs.Register("btpeer_assert_uptime_is_less_than_duration", assertUptimeIsLessThanDurationExec)
}
