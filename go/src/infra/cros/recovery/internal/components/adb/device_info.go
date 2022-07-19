// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package adb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/logger"
)

// GetDeviceState reads device state (offline | bootloader | device).
func GetDeviceState(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) (string, error) {
	const adbGetStateCmd = "adb -s %s get-state"
	cmd := fmt.Sprintf(adbGetStateCmd, serialNumber)
	state, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return "", errors.Annotate(err, "get attached device state").Err()
	}
	log.Debugf("Attached device state: %q", state)
	return state, nil
}

// GetDevicePath reads device path.
func GetDevicePath(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) (string, error) {
	const adbGetDevPathCmd = "adb -s %s get-devpath"
	cmd := fmt.Sprintf(adbGetDevPathCmd, serialNumber)
	path, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return "", errors.Annotate(err, "get attached device path").Err()
	}
	log.Debugf("Attached device path: %q", path)
	return path, nil
}

// IsDeviceAccessible verifies that DUT is accessible through the associated host.
func IsDeviceAccessible(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	state, err := GetDeviceState(ctx, run, log, serialNumber)
	if err != nil {
		return errors.Annotate(err, "dut is accessible").Err()
	}
	log.Debugf("Attached DUT %q: state %q", serialNumber, state)
	if state != "device" {
		return errors.Reason("invalid attached dut %q state %q", serialNumber, state).Err()
	}
	return nil
}

// IsDeviceRooted validates whether device is rooted. Refer to go/abp-security/rooted-devices for info on device rooting.
func IsDeviceRooted(ctx context.Context, run components.Runner, log logger.Logger, serialNumber string) error {
	const adbIsDeviceRootedCmd = "adb -s %s shell su root whoami>/dev/null 2>&1; echo $?"
	cmd := fmt.Sprintf(adbIsDeviceRootedCmd, serialNumber)
	output, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return errors.Annotate(err, "is device rooted").Err()
	}
	if strings.TrimSpace(output) != "0" {
		return errors.Reason("attached device %q is not rooted", serialNumber).Err()
	}
	log.Debugf("Attached device %q is rooted", serialNumber)
	return nil
}
