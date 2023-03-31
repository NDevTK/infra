// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

const (
	// Threshold of messages log size we can keep. It should use expression(bcwkMG) that supported by `find` cli.
	currentMessagesLogSizeThreshold = "300M"
	oldMessagesLogSizeThreshold     = "50M"
	// A specific firmware for servo dock, applies to servo_v4.1 only.
	genesysLogicFirmwarePath = "/usr/share/fwupd/remotes.d/vendor/firmware/be2c9146ff4cfac5d647376c39ce0b78151e9f1a785a287e93ac3968aff2ed50-GenesysLogic_GL3590_64.17.cab"
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
	dur, err := cros.Uptime(ctx, info.DefaultRunner())
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

// logCleanupExec rotate and cleanup stale log files on labstation.
func logCleanupExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()

	// Clean up stale(> 7 days) servod logs.
	run(ctx, info.GetExecTimeout(), "find /var/log/servo* -type f -mtime +7 | xargs rm")

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

	// Remove anything in /var/log if large than 500M.
	run(ctx, info.GetExecTimeout(), "find /var/log/ -type f -size +500M | xargs rm")

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

// checkGenesysLogicFirmwareImageExists checks if the OS image on labstation contains a specific GenesysLogic firmware image.
func checkGenesysLogicFirmwareImageExists(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	if _, err := run(ctx, 20*time.Second, fmt.Sprintf("test -f %s", genesysLogicFirmwarePath)); err != nil {
		return errors.Reason("check genesys logic firmware image exists: current labstation image does not contains target firmware.").Err()
	}
	return nil
}

// updateGenesysLogicFirmwareForServos updates a specific version of GenesysLogic firmware for all
// servo_v4p1 on the labstation. The update is a no-op for servos that already updated to the given
// firmware, and servos that doesn't have the applicable chip(e.g. servo_v4).
func updateGenesysLogicFirmwareForServos(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	if _, err := run(ctx, 20*time.Minute, fmt.Sprintf("fwupdtool install %s", genesysLogicFirmwarePath)); err != nil {
		// We expected non-zero exit code in no-op cases, so just log the error here.
		log.Debugf(ctx, "(Non-critical)fwupdtool run returns non-zero exit code, %s", err.Error())
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
	execs.Register("cros_update_genesys_logic_firmware", updateGenesysLogicFirmwareForServos)
	execs.Register("cros_genesys_logic_firmware_image_exists", checkGenesysLogicFirmwareImageExists)
}
