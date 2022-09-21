// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

const (
	// Threshold of messages log size we can keep. It should use expression(bcwkMG) that supported by `find` cli.
	currentMessagesLogSizeThreshold = "300M"
	oldMessagesLogSizeThreshold     = "50M"
)

// cleanTmpOwnerRequestExec cleans tpm owner requests.
func cleanTmpOwnerRequestExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	_, err := run(ctx, time.Minute, "crossystem clear_tpm_owner_request=1")
	return errors.Annotate(err, "clear tpm owner request").Err()
}

// validateUptime validate that host is up for more than a threshold
// number of hours.
func validateUptime(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	maxDuration := argsMap.AsDuration(ctx, "max_duration", 0, time.Hour)
	minDuration := argsMap.AsDuration(ctx, "min_duration", 0, time.Hour)
	if maxDuration == 0 && minDuration == 0 {
		return errors.Reason("validate uptime: neither min nor max duration is specified").Err()
	}
	dur, err := uptime(ctx, info.DefaultRunner())
	if err != nil {
		return errors.Annotate(err, "validate uptime").Err()
	}
	if maxDuration != 0 && *dur >= maxDuration {
		return errors.Reason("validate uptime: uptime %s equals or exceeds the expected maximum threshold %s", dur, maxDuration).Err()
	}
	if minDuration != 0 && *dur < minDuration {
		return errors.Reason("validate uptime: uptime %s is less than the expected minimum threshold %s", dur, minDuration).Err()
	}
	log.Debugf(ctx, "Validate Uptime: current uptime: %s, min threshold: %s, max threshold: %s: all good.", dur, minDuration, maxDuration)
	return nil
}

const (
	// The flag-file indicates the host should not to be rebooted.
	noRebootFlagFile = "/tmp/no_reboot"
)

// allowedRebootExec checks if DUT is allowed to reboot.
// If system has /tmp/no_reboot file then reboot is not allowed.
func allowedRebootExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	cmd := fmt.Sprintf("test %s", noRebootFlagFile)
	_, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return errors.Annotate(err, "has no-reboot request").Err()
	}
	log.Debugf(ctx, "No-reboot request file found.")
	return nil
}

// filesystemIoNotBlockedExec check if the labstation's filesystem IO is blocked.
func filesystemIoNotBlockedExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	cmd := "ps axl | awk '$10 ~ /D/'"
	output, err := run(ctx, info.GetExecTimeout(), cmd)
	if err != nil {
		return errors.Annotate(err, "filesystem is not blocked").Err()
	}
	// Good labstation may occasionally have an process in uninterruptible
	// sleep state transiently, so we look for these who have 2+ processes
	// stuck in such a state.
	if len(output) > 1 {
		return errors.Reason("filesystem is not blocked: more than one processes in uninterruptible sleep state, I/O is likely blocked.").Err()
	}
	return nil
}

// logCleanupExec rotate and cleanup messages logs if they're large than the threshold.
func logCleanupExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	// First we want to check if the current messages log larger than the threshold, and if it is
	// we need rotate logs before we can safely remove it as other process may still writing logs into it.
	checkCurrentCmd := fmt.Sprintf("find /var/log/messages -size +%s", currentMessagesLogSizeThreshold)
	if out, _ := run(ctx, info.GetExecTimeout(), checkCurrentCmd); out != "" {
		log.Debugf(ctx, "Log cleanup: current messages log larger than %s, will rotate logs.", currentMessagesLogSizeThreshold)
		if _, err := run(ctx, info.GetExecTimeout(), "/usr/sbin/chromeos-cleanup-logs"); err != nil {
			log.Debugf(ctx, "Log cleanup: failed to execute chromeos-cleanup-logs, %v", err)
		}
	}
	// Checking if there are any old logs that larger than the threshold, and if true remove all old logs.
	checkOldCmd := fmt.Sprintf("find /var/log/messages.* -size +%s", oldMessagesLogSizeThreshold)
	if out, _ := run(ctx, info.GetExecTimeout(), checkOldCmd); out != "" {
		log.Debugf(ctx, "Log cleanup: detected old messages log that larger than %s", oldMessagesLogSizeThreshold)
		if _, err := run(ctx, info.GetExecTimeout(), "rm /var/log/messages.*"); err != nil {
			return errors.Reason("log cleanup: failed to remove old messages log.").Err()
		}
		log.Debugf(ctx, "Log cleanup: successfully removed old messages log.")
	}
	return nil
}

// removeBluetoothDeviceExec removes bluetooth device from the labstation.
func removeBluetoothDeviceExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	out, err := run(ctx, info.GetExecTimeout(), "bluetoothctl devices")
	if err != nil {
		return errors.Reason("remote bluetooth device: failed to get available devices.").Err()
	}
	// The output of device info will looks like "Device F4:60:77:0C:7C:39 F4-60-77-0C-7C-39",
	// and we need the middle part uuit as identifier to remove it.
	s := strings.Fields(out)
	if len(s) > 1 {
		log.Debugf(ctx, "Removing bluetooth device %s", s[1])
		if _, err := run(ctx, info.GetExecTimeout(), "bluetoothctl", "remove", s[1]); err != nil {
			return errors.Reason("remote bluetooth device: failed to remove bluetooth device.").Err()
		}
	}
	return nil
}

func init() {
	execs.Register("cros_clean_tmp_owner_request", cleanTmpOwnerRequestExec)
	execs.Register("cros_validate_uptime", validateUptime)
	execs.Register("cros_allowed_reboot", allowedRebootExec)
	execs.Register("cros_filesystem_io_not_blocked", filesystemIoNotBlockedExec)
	execs.Register("cros_log_clean_up", logCleanupExec)
	execs.Register("cros_remove_bt_devices", removeBluetoothDeviceExec)
}
