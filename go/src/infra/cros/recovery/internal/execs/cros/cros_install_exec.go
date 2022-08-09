// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/components/cros/firmware"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/metrics"
	"infra/cros/recovery/internal/log"
)

// Boot device from servo USB drive when device is in DEV mode.
func devModeBootFromServoUSBDriveExec(ctx context.Context, info *execs.ExecInfo) error {
	am := info.GetActionArgs(ctx)
	waitBootTimeout := am.AsDuration(ctx, "boot_timeout", 1, time.Second)
	waitBootInterval := am.AsDuration(ctx, "retry_interval", 1, time.Second)
	servod := info.NewServod()
	run := info.NewRunner(info.RunArgs.DUT.Name)
	ping := info.NewPinger(info.RunArgs.DUT.Name)
	err := cros.BootFromServoUSBDriveInDevMode(ctx, waitBootTimeout, waitBootInterval, run, ping, servod, info.NewLogger())
	return errors.Annotate(err, "dev-mode boot from servo usb-drive").Err()
}

// Install ChromeOS from servo USB drive when booted from it.
func runChromeosInstallCommandWhenBootFromUSBDriveExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	err := cros.RunInstallOSCommand(ctx, info.ActionTimeout, run, info.NewLogger())
	return errors.Annotate(err, "run install os after boot from USB-drive").Err()
}

// storageErrors are all the possible error messages that can be
// generated if OS install process fails due to errors with the
// storage device.
var storageErrors = map[string]bool{
	"No space left on device":                    true,
	"I/O error when trying to write primary GPT": true,
	"Input/output error while writing out":       true,
	"cannot read GPT header":                     true,
	"can not determine destination device":       true,
	"wrong fs type":                              true,
	"bad superblock on":                          true,
}

// installFromUSBDriveInRecoveryModeExec re-installs a test image from USB.
//
// Also can flash firmware  as part of action.
func installFromUSBDriveInRecoveryModeExec(ctx context.Context, info *execs.ExecInfo) error {
	am := info.GetActionArgs(ctx)
	dut := info.RunArgs.DUT
	dutRun := info.NewRunner(dut.Name)
	dutBackgroundRun := info.NewBackgroundRunner(dut.Name)
	dutPing := info.NewPinger(dut.Name)
	servod := info.NewServod()
	logger := info.NewLogger()
	callback := func(_ context.Context) error {
		if am.AsBool(ctx, "run_tpm_reset", false) {
			// Clear TPM is not critical as can fail in some cases.
			tpmResetTimeout := am.AsDuration(ctx, "tpm_reset_timeout", 60, time.Second)
			if _, err := dutRun(ctx, tpmResetTimeout, "chromeos-tpm-recovery"); err != nil {
				logger.Debugf("Install from USB drive: (non-critical) fail to reset tmp: Error: %s", err)
			}
		}
		if am.AsBool(ctx, "run_os_install", false) {
			installTimeout := am.AsDuration(ctx, "install_timeout", 600, time.Second)
			if _, err := dutRun(ctx, installTimeout, "chromeos-install", "--yes"); err != nil {
				stdErr, ok := errors.TagValueIn(execs.StdErrTag, err)
				if ok {
					stdErrStr := stdErr.(string)
					if storageErrors[stdErrStr] {
						info.RunArgs.DUT.State = dutstate.NeedsReplacement
						log.Debugf(ctx, "Install from USB Drive in Recovery Mode: Failed to install ChromeOS due to storage error %s, setting DUT state to %s", stdErrStr, dutstate.NeedsReplacement)
					}
				} else {
					log.Debugf(ctx, "Install from USB Drive in Recovery Mode: std err not found.")
				}
				return errors.Annotate(err, "install from usb drive in recovery mode").Err()
			}
			logger.Debugf("Install from USB drive: finished install process")
		}
		if am.AsBool(ctx, "run_fw_update", false) {
			req := &firmware.FirmwareUpdaterRequest{
				// Options for the mode are: autoupdate, recovery, factory.
				Mode:           am.AsString(ctx, "fw_update_mode", "autoupdate"),
				Force:          am.AsBool(ctx, "fw_update_use_force", false),
				UpdaterTimeout: am.AsDuration(ctx, "fw_update_timeout", 600, time.Second),
			}
			if err := firmware.RunFirmwareUpdater(ctx, req, dutRun, logger); err != nil {
				return errors.Annotate(err, "install from usb drive in recovery mode").Err()
			}
			logger.Debugf("Install from USB drive: finished fw update")
		}
		return nil
	}
	req := &cros.BootInRecoveryRequest{
		DUT:          dut,
		BootTimeout:  am.AsDuration(ctx, "boot_timeout", 480, time.Second),
		BootInterval: am.AsDuration(ctx, "boot_interval", 10, time.Second),
		// Register that device booted and sshable.
		Callback:            callback,
		HaltTimeout:         am.AsDuration(ctx, "halt_timeout", 120, time.Second),
		IgnoreRebootFailure: am.AsBool(ctx, "ignore_reboot_failure", false),
	}
	if err := cros.BootInRecoveryMode(ctx, req, dutRun, dutBackgroundRun, dutPing, servod, logger); err != nil {
		return errors.Annotate(err, "install from usb drive in recovery mode").Err()
	}
	// Time to wait DUT boot up from post installation.
	postInstallationBootTime := am.AsDuration(ctx, "post_install_boot_time", 60, time.Second)
	logger.Debugf("Wait %s post installation for DUT to boot up.", postInstallationBootTime)
	time.Sleep(postInstallationBootTime)
	return nil
}

// verifyBootInRecoveryModeExec verify that device can boot in recovery mode and reboot to normal mode again.
func verifyBootInRecoveryModeExec(ctx context.Context, info *execs.ExecInfo) error {
	am := info.GetActionArgs(ctx)
	dut := info.RunArgs.DUT
	dutRun := info.NewRunner(dut.Name)
	dutBackgroundRun := info.NewBackgroundRunner(dut.Name)
	dutPing := info.NewPinger(dut.Name)
	servod := info.NewServod()
	// Flag to notice when device booted and sshable.
	var successBooted bool
	callback := func(_ context.Context) error {
		successBooted = true
		return nil
	}
	req := &cros.BootInRecoveryRequest{
		DUT:          dut,
		BootTimeout:  am.AsDuration(ctx, "boot_timeout", 480, time.Second),
		BootInterval: am.AsDuration(ctx, "boot_interval", 10, time.Second),
		// Register that device booted and sshable.
		Callback:            callback,
		HaltTimeout:         am.AsDuration(ctx, "halt_timeout", 120, time.Second),
		IgnoreRebootFailure: am.AsBool(ctx, "ignore_reboot_failure", false),
	}
	if err := cros.BootInRecoveryMode(ctx, req, dutRun, dutBackgroundRun, dutPing, servod, info.NewLogger()); err != nil {
		return errors.Annotate(err, "verify boot in recovery mode").Err()
	}
	if !successBooted {
		return errors.Reason("verify boot in recovery mode: did not booted").Err()
	}
	return nil
}

// isTimeToForceDownloadImageToUsbKeyExec verifies if we want to force download image to usbkey.
//
// @params: actionArgs should be in the format of:
// Ex: ["task_name:xxx", "repair_failed_count:1", "repair_failed_interval:10"]
func isTimeToForceDownloadImageToUsbKeyExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	taskName := argsMap.AsString(ctx, "task_name", "")
	repairFailedCountTarget := argsMap.AsInt(ctx, "repair_failed_count", -1)
	repairFailedInterval := argsMap.AsInt(ctx, "repair_failed_interval", 10)
	repairFailedCount, err := metrics.CountFailedRepairFromMetrics(ctx, taskName, info)
	if err != nil {
		return errors.Annotate(err, "is time to force download image to usbkey").Err()
	}
	// The previous repair task was successful, and the user didn't specify
	// when repair_failed_count == 0 to flash usbkey image.
	if repairFailedCount == 0 && repairFailedCountTarget != 0 {
		return errors.Reason("is time to force download image to usbkey: the number of failed repair is 0, will not force to install os iamge").Err()
	}
	if repairFailedCount == repairFailedCountTarget || repairFailedCount%repairFailedInterval == 0 {
		log.Infof(ctx, "Required re-download image to usbkey as a previous repair failed. Fail count: %d", repairFailedCount)
		return nil
	}
	return errors.Reason("is time to force download image to usbkey: Fail count: %d", repairFailedCount).Err()
}

func init() {
	execs.Register("cros_dev_mode_boot_from_servo_usb_drive", devModeBootFromServoUSBDriveExec)
	execs.Register("cros_run_chromeos_install_command_after_boot_usbdrive", runChromeosInstallCommandWhenBootFromUSBDriveExec)
	execs.Register("cros_install_in_recovery_mode", installFromUSBDriveInRecoveryModeExec)
	execs.Register("cros_verify_boot_in_recovery_mode", verifyBootInRecoveryModeExec)
	execs.Register("cros_is_time_to_force_download_image_to_usbkey", isTimeToForceDownloadImageToUsbKeyExec)
}
