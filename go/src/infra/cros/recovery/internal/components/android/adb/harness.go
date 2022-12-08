// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adb

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/logger"
)

// EnableDeviceTestHarnessMode resets device (https://developer.android.com/studio/command-line/adb#test_harness).
func EnableDeviceTestHarnessMode(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	const adbEnableHarnessModeCmd = "adb -s %s shell cmd testharness enable"
	// TODO(b/259746452): use shell.QuoteUnix for quoting
	cmd := fmt.Sprintf(adbEnableHarnessModeCmd, serialNumber)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "enable device harness mode").Err()
	}
	log.Debugf("Harness mode is enabled for attached device: %q", serialNumber)
	return nil
}

// WaitForDeviceState waits until the device gets into the expected state.
func WaitForDeviceState(ctx context.Context, expectedState State, stateCount int, waitTimeout time.Duration, run components.Runner, log logger.Logger, serialNumber string) error {
	waitInRetry := 5 * time.Second
	retryCount := int(waitTimeout / waitInRetry)
	log.Debugf("Waiting for DUT %s state '%s'. Retry count: %d", serialNumber, expectedState, retryCount)
	if stateCount == 0 {
		stateCount = 1
	}
	// Ensure the consistent device state at least <stateCount> times in a row.
	successCount, failureCount := 0, 0
	for {
		if ds, err := GetDeviceState(ctx, run, log, serialNumber); err != nil {
			successCount = 0
			failureCount += 1
		} else {
			if ds == expectedState {
				successCount += 1
				failureCount = 0
				log.Debugf("DUT %s is in '%s' state. Current success count: %d", serialNumber, ds, successCount)
				if successCount >= stateCount {
					break
				}
			} else {
				successCount = 0
				if ds == Unauthorized {
					failureCount += 1
					// If device is in unauthorized state for more than 90 seconds, return error.
					// The device either broken or public key is missing.
					if failureCount >= 16 {
						return errors.Reason("dut state is '%s': %s", ds, serialNumber).Err()
					}
				}
			}
		}
		retryCount -= 1
		if retryCount <= 0 {
			break
		}
		time.Sleep(waitInRetry)
	}
	if successCount < stateCount {
		return errors.Reason("failed to wait for dut '%s' state: %q", expectedState, serialNumber).Err()
	}
	log.Debugf("Attached device in '%s' state: %q", expectedState, serialNumber)
	return nil
}
