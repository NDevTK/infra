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
	const (
		adbWaitForDeviceCmd     = "timeout %d adb -s %s wait-for-device"
		maxWaitForDeviceSeconds = 120
	)
	cmd := fmt.Sprintf(adbWaitForDeviceCmd, maxWaitForDeviceSeconds, serialNumber)
	if _, err := run(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "wait for device").Err()
	}
	log.Debugf("Attached device is available: %q", serialNumber)
	return nil
}
