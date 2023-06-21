// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/internal/retry"
	"infra/cros/recovery/logger"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

// BootInRecoveryRequest holds info to boot device in recovery mode.
type BootInRecoveryRequest struct {
	DUT *tlw.Dut
	// Booting time value to verify when device booted and available for SSH.
	BootRetry    int
	BootTimeout  time.Duration
	BootInterval time.Duration
	// Call function to cal after device booted in recovery mode.
	Callback            func(context.Context) error
	AddObservation      func(*metrics.Observation)
	IgnoreRebootFailure bool
	// After reboot params specified to check if device booted or not.
	AfterRebootVerify             bool
	AfterRebootTimeout            time.Duration
	AfterRebootAllowUseServoReset bool
}

// BootInRecoveryMode perform boot device in recovery mode.
//
// Boot in recovery mode performed by RO firmware and in some cases required stopPD negotiation.
// Please specify callback function to perform needed actions when device booted in recovery mode.
func BootInRecoveryMode(ctx context.Context, req *BootInRecoveryRequest, dutRun, dutBackgroundRun components.Runner, dutPing components.Pinger, servod components.Servod, log logger.Logger) (rErr error) {
	if req.BootRetry < 1 {
		// We retry at least once when method called.
		req.BootRetry = 1
	}
	// If observation is not provided then we create fake to print to logs
	if req.AddObservation == nil {
		req.AddObservation = func(observation *metrics.Observation) {
			if observation != nil {
				log.Debugf("Observation created kind:%q with %v", observation.MetricKind, observation.Value)
			}
		}
	}
	needSink, err := RecoveryModeRequiredPDOff(ctx, dutRun, servod, req.DUT)
	if err != nil {
		return errors.Annotate(err, "boot in recovery mode").Err()
	}
	defer func() {
		// Record the label at the end as it can be changed.
		req.AddObservation(metrics.NewStringObservation("need_snk_power", fmt.Sprintf("%v", needSink)))
	}()
	req.AddObservation(metrics.NewStringObservation("need_snk_expected", fmt.Sprintf("%v", needSink)))
	if needSink {
		if batteryLevel, err := servo.BatteryChargePercent(ctx, servod); err != nil {
			req.AddObservation(metrics.NewInt64Observation("battery_level", -1))
			log.Debugf("Fail to read battery level from device %s.", err)
			log.Debugf("We will not set PD to snk mode when boot in recovery mode.")
			needSink = false
		} else {
			req.AddObservation(metrics.NewInt64Observation("battery_level", int64(batteryLevel)))
			// If device has less 30% of battery then we will not try to recover it.
			// If device lost power in middle of install it damage the disk.
			const minBatterLevel = int32(30)
			if batteryLevel < minBatterLevel {
				log.Debugf("Battery level %d%% is lower minimum expectation of %d%%.", batteryLevel, minBatterLevel)
				log.Debugf("We will not set PD to snk mode when boot in recovery mode.")
				needSink = false
			}
		}
	}
	log.Debugf("Servo OS Install Repair: needSink :%t", needSink)
	restoreServoState := func() error {
		log.Debugf("Boot in recovery mode: recover servo states...")
		// Register turn off for the DUT if at the end.
		// All errors just logging as the action to clean up the state.
		if needSink {
			if err := servo.SetPDRole(ctx, servod, servo.PD_ON, false); err != nil {
				log.Debugf("Restore PD for DUT failed: %s", err)
			}
		}
		// Waiting 10 seconds for USB re-enumerate after PD role switch.
		time.Sleep(10 * time.Second)
		if err := servo.SetPowerState(ctx, servod, servo.PowerStateValueOFF); err != nil {
			return errors.Annotate(err, "boot in recovery mode").Err()
		}
		if err := servo.UpdateUSBVisibility(ctx, servo.USBVisibleOff, servod); err != nil {
			log.Debugf("Turn off USB drive on servo failed: %s", err)
		}
		// Waiting 10 seconds before turn it on as the device can be still in transition to off.
		time.Sleep(10 * time.Second)
		if err := servo.SetPowerState(ctx, servod, servo.PowerStateValueON); err != nil {
			return errors.Annotate(err, "boot in recovery mode").Err()
		}
		// Waiting 3 seconds before allowed followeing commands to try something else.
		time.Sleep(3 * time.Second)
		log.Debugf("Boot in recovery mode: servo states recovered.")
		return nil
	}
	// Always restore servo state by the end!
	defer func() {
		if err := restoreServoState(); err != nil {
			log.Debugf("Boot in recovery mode: %s", err)
			// Don't override the original error.
			if !req.IgnoreRebootFailure && rErr == nil {
				// We cannot return it, so we set it.
				// If we fail when restored the states then we have issues.
				rErr = err
				return
			}
		}
		// Verify the boot only if pass the execution or restore states.
		if rErr == nil && req.AfterRebootVerify {
			log.Debugf("Boot in recovery mode: starting verification of the boot...")
			for {
				if err := WaitUntilSSHable(ctx, req.AfterRebootTimeout, req.BootInterval, dutRun, log); err != nil {
					if req.AfterRebootAllowUseServoReset {
						req.AfterRebootAllowUseServoReset = false
						if err := servo.SetPowerState(ctx, servod, servo.PowerStateValueReset); err != nil {
							log.Infof("Fail to reset by servo: %s", err)
						}
						continue
					}
					log.Debugf("Device is not SSH-able after reboot!")
					rErr = err
				} else {
					log.Debugf("Device is SSH-able!")
				}
				break
			}
		}
	}()
	retryBootFunc := func() error {
		log.Infof("Boot in Recovery Mode: starting retry...")
		// Turn power off.
		if err := servo.SetPowerState(ctx, servod, servo.PowerStateValueOFF); err != nil {
			return errors.Annotate(err, "retry boot").Err()
		}
		// Next:Boot in recovery mode. The steps are:
		// Step 1. Switch the USB to DUT on the servo multiplexer
		if err := servo.UpdateUSBVisibility(ctx, servo.USBVisibleDUT, servod); err != nil {
			return errors.Annotate(err, "retry boot").Err()
		}
		// Step 2. For servo V4, switch power delivery to sink mode. c.f.:
		// crbug.com/1129165.
		if needSink {
			if err := servo.SetPDRole(ctx, servod, servo.PD_OFF, false); err != nil {
				return errors.Annotate(err, "retry boot").Err()
			}
		}
		// Sleep a few seconds to allowed apply all previous states before boot in recovery mode.
		time.Sleep(1 * time.Second)
		log.Infof("Boot in Recovery Mode: Started try to boot in recovery mode by power_state:rec.")
		if err := servo.SetPowerState(ctx, servod, servo.PowerStateValueRecoveryMode); err != nil {
			log.Debugf("Boot in Recovery Mode: Failure when trying to set power_state:rec with error: %s", err)
		}
		log.Debugf("Boot in Recovery Mode: Waiting to device to be SSH-able.")
		if err := WaitUntilSSHable(ctx, req.BootTimeout, req.BootInterval, dutRun, log); err != nil {
			return errors.Annotate(err, "retry boot").Err()
		}
		if err := IsBootedFromExternalStorage(ctx, dutRun); err != nil {
			log.Infof("Device booted from internal storage.")
			return errors.Annotate(err, "retry boot").Err()
		}
		log.Infof("Device successfully booted in recovery mode from USB-drive.")
		return nil
	}
	if retryErr := retry.LimitCount(ctx, req.BootRetry, req.BootInterval, retryBootFunc, "boot in recovery mode"); retryErr != nil {
		return errors.Annotate(retryErr, "boot in recovery mode").Err()
	}
	if req.Callback != nil {
		log.Infof("Boot in recovery mode: passing control to call back.")
		// List information about block devices.
		// This informcation helps to understand which devices present and visible on the DUT.
		if _, err := dutRun(ctx, 10*time.Second, "lsblk"); err != nil {
			log.Infof("Fail to list device of the DUT: %s", err)
		}
		if err := req.Callback(ctx); err != nil {
			return errors.Annotate(err, "boot in recovery mode: callback").Err()
		}
		log.Infof("Boot in recovery mode: control returned.")
	}
	return nil
}
