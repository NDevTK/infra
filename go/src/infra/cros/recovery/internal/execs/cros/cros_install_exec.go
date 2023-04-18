// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/components/cros/firmware"
	"infra/cros/recovery/internal/components/cros/storage"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/internal/retry"
	"infra/cros/recovery/logger/metrics"
)

// Boot device from servo USB drive when device is in DEV mode.
func devModeBootFromServoUSBDriveExec(ctx context.Context, info *execs.ExecInfo) error {
	am := info.GetActionArgs(ctx)
	bootRetry := am.AsInt(ctx, "boot_retry", 1)
	waitBootTimeout := am.AsDuration(ctx, "boot_timeout", 1, time.Second)
	waitBootInterval := am.AsDuration(ctx, "retry_interval", 1, time.Second)
	verifyUSBDriveBoot := am.AsBool(ctx, "verify_usbkey_boot", false)
	if !verifyUSBDriveBoot && bootRetry > 1 {
		// if we retry then we will verify boot as that is reason to tell that device booted as expected.
		verifyUSBDriveBoot = true
	}
	servod := info.NewServod()
	run := info.NewRunner(info.GetDut().Name)
	ping := info.NewPinger(info.GetDut().Name)
	logger := info.NewLogger()
	retryBootFunc := func() error {
		logger.Infof("Boot in DEV-mode: staring...")
		if err := cros.BootFromServoUSBDriveInDevMode(ctx, waitBootTimeout, waitBootInterval, run, ping, servod, logger); err != nil {
			return errors.Annotate(err, "retry boot in dev-mode").Err()
		}
		if verifyUSBDriveBoot {
			if err := cros.IsBootedFromExternalStorage(ctx, run); err != nil {
				logger.Infof("Boot in DEV-mode: booted from internal storage.")
				return errors.Annotate(err, "retry boot in dev-mode").Err()
			}
			logger.Infof("Boot in DEV-mode: device successfully booted from USB-drive.")
		} else {
			logger.Infof("Boot in DEV-mode: device successfully booted.")
		}
		return nil
	}
	if retryErr := retry.LimitCount(ctx, bootRetry, waitBootInterval, retryBootFunc, "boot in dev mode"); retryErr != nil {
		return errors.Annotate(retryErr, "dev-mode boot from usb-drive").Err()
	}
	return nil
}

// Install ChromeOS from servo USB drive when booted from it.
func runChromeosInstallCommandWhenBootFromUSBDriveExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	actionArgs := info.GetActionArgs(ctx)
	err := cros.RunInstallOSCommand(ctx, info.GetExecTimeout(), run)
	if issueReason := cros.StorageIssuesExist(ctx, err); issueReason.NotEmpty() {
		info.GetDut().State = dutstate.NeedsReplacement
		info.GetDut().DutStateReason = issueReason
		log.Debugf(ctx, "Setting DUT state: %s", dutstate.NeedsReplacement)
		newAnnotator := errors.Annotate(err, "install from usb drive in recovery mode: storage needs replacement").Tag(retry.LoopBreakTag())
		if actionArgs.AsBool(ctx, "allowed_abort_plan", true) {
			newAnnotator = newAnnotator.Tag(execs.PlanAbortTag)
		}
		return newAnnotator.Err()
	}
	return errors.Annotate(err, "run install os after boot from USB-drive").Err()
}

// installFromUSBDriveInRecoveryModeExec re-installs a test image from USB.
//
// Also can flash firmware  as part of action.
func installFromUSBDriveInRecoveryModeExec(ctx context.Context, info *execs.ExecInfo) error {
	am := info.GetActionArgs(ctx)
	dut := info.GetDut()
	dutRun := info.NewRunner(dut.Name)
	dutBackgroundRun := info.NewBackgroundRunner(dut.Name)
	dutPing := info.NewPinger(dut.Name)
	servod := info.NewServod()
	logger := info.NewLogger()
	// Record if device booted in recovery mode.
	bootedInrecoveryMode := "no"
	finishedTPMReset := "no"
	finishedOSInstall := "no"
	finishedFWUpdate := "no"
	defer func() {
		info.AddObservation(metrics.NewStringObservation("bootedInrecoveryMode", bootedInrecoveryMode))
		info.AddObservation(metrics.NewStringObservation("finishedTPMReset", finishedTPMReset))
		info.AddObservation(metrics.NewStringObservation("finishedOSInstall", finishedOSInstall))
		info.AddObservation(metrics.NewStringObservation("finishedFWUpdate", finishedFWUpdate))
	}()
	callback := func(_ context.Context) error {
		bootedInrecoveryMode = "yes"
		if am.AsBool(ctx, "run_custom_commands", false) {
			allowedToFail := am.AsBool(ctx, "custom_command_allowed_to_fail", false)
			commandTimeout := am.AsDuration(ctx, "custom_command_timeout", 60, time.Second)
			customCommands := am.AsString(ctx, "custom_commands", "")
			if customCommands != "" {
				for _, customCommand := range strings.Split(customCommands, "##") {
					logger.Debugf("Prepare run custom command: %q", customCommand)
					if _, err := dutRun(ctx, commandTimeout, customCommand); err != nil {
						if allowedToFail {
							logger.Debugf("Run custom command allowed to continue after fail with error: %s", err)
						} else {
							return errors.Annotate(err, "run custom command").Err()
						}
					}
				}
			}
		}
		if am.AsBool(ctx, "run_tpm_reset", false) {
			// Clear TPM is not critical as can fail in some cases.
			tpmResetTimeout := am.AsDuration(ctx, "tpm_reset_timeout", 60, time.Second)
			if _, err := dutRun(ctx, tpmResetTimeout, "chromeos-tpm-recovery"); err != nil {
				finishedTPMReset = "failed"
				logger.Debugf("Install from USB drive: (non-critical) fail to reset tmp: Error: %s", err)
			} else {
				finishedTPMReset = "yes"
			}
		}
		if am.AsBool(ctx, "run_os_install", false) {
			installTimeout := am.AsDuration(ctx, "install_timeout", 600, time.Second)
			if err := cros.RunInstallOSCommand(ctx, installTimeout, dutRun); err != nil {
				finishedOSInstall = "failed"
				if issueReason := cros.StorageIssuesExist(ctx, err); issueReason.NotEmpty() {
					info.GetDut().State = dutstate.NeedsReplacement
					info.GetDut().DutStateReason = issueReason
					log.Debugf(ctx, "Setting DUT state: %s", dutstate.NeedsReplacement)
					newAnnotator := errors.Annotate(err, "install from usb drive in recovery mode: storage needs replacement").Tag(retry.LoopBreakTag())
					if am.AsBool(ctx, "allowed_abort_plan", true) {
						newAnnotator = newAnnotator.Tag(execs.PlanAbortTag)
					}
					return newAnnotator.Err()
				}
				log.Debugf(ctx, "Install from usb drive fail: %s", err)
				log.Debugf(ctx, "Will try to check storage if that is bad!")
				// When install fail it can be because of bad storage.
				// Following the logic in legacy repair, we will now
				// attempt a storage audit on the DUT.
				if err := storage.AuditStorageSMART(ctx, dutRun, info.GetChromeos().GetStorage(), dut); err != nil {
					return errors.Annotate(err, "install from usb drive in recovery mode").Tag(retry.LoopBreakTag()).Err()
				}
				// Default values for these variables have also been
				// included in the action to document their availability
				// for modification. As we booted from USB-drive we can check
				// internal storage for read-write.
				bbMode := storage.AuditMode(am.AsString(ctx, "badblocks_mode", "rw"))
				timeoutRO := am.AsDuration(ctx, "rw_badblocks_timeout", 5400, time.Second)
				timeoutRW := am.AsDuration(ctx, "ro_badblocks_timeout", 3600, time.Second)
				bbArgs := storage.BadBlocksArgs{
					AuditMode: bbMode,
					Run:       dutRun,
					Storage:   info.GetChromeos().GetStorage(),
					Dut:       info.GetDut(),
					Metrics:   info.GetMetrics(),
					TimeoutRW: timeoutRW,
					TimeoutRO: timeoutRO,
					NewMetric: info.NewMetric,
				}
				if err := storage.CheckBadblocks(ctx, &bbArgs); err != nil {
					if execs.SSHErrorInternal.In(err) {
						log.Debugf(ctx, "Install from usb drive: bad blocks check command returned a negative error code, not setting needs replacement state for the DUT.")
					} else {
						log.Debugf(ctx, "The new DUT state: %q, reason: %q", info.GetDut().State, info.GetDut().DutStateReason)
					}
					newAnnotator := errors.Annotate(err, "install from usb drive in recovery mode").Tag(retry.LoopBreakTag())
					if am.AsBool(ctx, "allowed_abort_plan", true) {
						newAnnotator = newAnnotator.Tag(execs.PlanAbortTag)
					}
					return newAnnotator.Err()
				}
			}
			haltTimeout := am.AsDuration(ctx, "halt_timeout", 120, time.Second)
			if _, err := dutRun(ctx, haltTimeout, "halt"); err != nil {
				logger.Debugf("Install from USB drive: Halt the DUT failed: %s", err)
			}
			logger.Debugf("Install from USB drive: finished install process")
			finishedOSInstall = "yes"
		}
		if am.AsBool(ctx, "run_fw_update", false) {
			req := &firmware.FirmwareUpdaterRequest{
				// Options for the mode are: autoupdate, recovery, factory.
				Mode:            am.AsString(ctx, "fw_update_mode", "autoupdate"),
				Force:           am.AsBool(ctx, "fw_update_use_force", false),
				UpdaterTimeout:  am.AsDuration(ctx, "fw_update_timeout", 600, time.Second),
				WriteProtection: am.AsBool(ctx, "fw_update_wp", false),
			}
			isCritical := am.AsBool(ctx, "fw_update_critical", true)
			if err := firmware.RunFirmwareUpdater(ctx, req, dutRun, logger); err != nil {
				finishedFWUpdate = "failed"
				if isCritical {
					return errors.Annotate(err, "install from usb drive in recovery mode").Err()
				} else {
					logger.Debugf("Failed to update fw on the DUT: %s", err)
				}
			} else {
				finishedFWUpdate = "true"
			}
			logger.Debugf("Install from USB drive: finished fw update")
		}
		return nil
	}
	req := &cros.BootInRecoveryRequest{
		DUT:          dut,
		BootRetry:    am.AsInt(ctx, "boot_retry", 1),
		BootTimeout:  am.AsDuration(ctx, "boot_timeout", 480, time.Second),
		BootInterval: am.AsDuration(ctx, "boot_interval", 10, time.Second),
		// Register that device booted and sshable.
		Callback:            callback,
		AddObservation:      info.AddObservation,
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
	dut := info.GetDut()
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
		BootRetry:    am.AsInt(ctx, "boot_retry", 1),
		BootTimeout:  am.AsDuration(ctx, "boot_timeout", 480, time.Second),
		BootInterval: am.AsDuration(ctx, "boot_interval", 10, time.Second),
		// Register that device booted and sshable.
		Callback:            callback,
		AddObservation:      info.AddObservation,
		IgnoreRebootFailure: am.AsBool(ctx, "ignore_reboot_failure", false),
		// After reboot action settings.
		AfterRebootVerify:             am.AsBool(ctx, "after_reboot_check", false),
		AfterRebootTimeout:            am.AsDuration(ctx, "after_reboot_timeout", 150, time.Second),
		AfterRebootAllowUseServoReset: am.AsBool(ctx, "after_reboot_allow_use_servo_reset", false),
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
	repairFailedCount, err := metrics.CountFailedRepairFromMetrics(ctx, info.GetDut().Name, taskName, info.GetMetrics())
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
