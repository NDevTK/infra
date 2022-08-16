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
	cmd := fmt.Sprintf(adbEnableHarnessModeCmd, serialNumber)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "enable device harness mode").Err()
	}
	log.Debugf("Harness mode is enabled for attached device: %q", serialNumber)
	return nil
}

// WaitForDevice waits until the device is online.
func WaitForDevice(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	const adbWaitForDeviceCmd = "adb -s %s wait-for-device"
	cmd := fmt.Sprintf(adbWaitForDeviceCmd, serialNumber)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "wait for device").Err()
	}
	// DUT may still be flaky after a reset even success in wait-for-device, so we need an additional check here
	// to ensure we can get correct DUT state at least 3 times in a row.
	waitForStableCount := 30
	successCount := 0
	for waitForStableCount > 0 {
		waitForStableCount -= 1
		if err := IsDeviceAccessible(ctx, run, log, serialNumber); err != nil {
			successCount = 0
		} else {
			successCount += 1
			log.Debugf("Device is accessible, current success count %d", successCount)
		}
		if successCount > 2 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if successCount < 3 {
		return errors.Reason("failed to wait DUT become stable").Err()
	}
	log.Debugf("Attached device is available: %q", serialNumber)
	return nil
}
