// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package usb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

const (
	// The prefix of the badblocks command for verifying USB
	// drives. The USB-drive path will be attached to it when
	// badblocks needs to be executed on a drive.
	badBlocksCommandPrefix = "badblocks -w -e 300 -b 4096 -t random %s"
	// The prefix of the smartctl command for running the health test
	// for USB drives that support SMART. The USB drive path will be attached
	// to it when the command needs to be executed on a drive.
	smartHealthCommandPrefix = "smartctl -H %s | awk '/SMART overall-health self-assessment test result:/ {print $6}'"
	// Expected output of a passing SMART health test.
	smartPass = "PASSED"
)

// UsbReset resets USB devices. usbId is one of the following
// - PPPP:VVVV - product and vendor id
// - BBB/DDD   - bus and device number
// - "Product" - product name
func UsbReset(ctx context.Context, run components.Runner, log logger.Logger, usbId string) error {
	usbResetCmd := "usbreset " + usbId
	_, err := run(ctx, time.Minute, usbResetCmd)
	if err != nil {
		return errors.Annotate(err, "usb reset").Err()
	}
	log.Debugf("USB is successfully reset: %s", usbId)
	return nil
}

// RunCheckOnHost generates new state for USB-drive by running check on DUT.
func RunCheckOnHost(ctx context.Context, run components.Runner, usbPath string, isSmartDevice bool, timeout time.Duration) (tlw.HardwareState, error) {
	command := fmt.Sprintf(badBlocksCommandPrefix, usbPath)
	if isSmartDevice {
		command = fmt.Sprintf(smartHealthCommandPrefix, usbPath)
	}
	log.Debugf(ctx, "Run Check On Host: Executing %q", command)
	// The execution timeout for this audit job is configured at the
	// level of the action. So the execution of this command will be
	// bound by that.
	out, err := run(ctx, timeout, command)
	if !isSmartDevice && err != nil && strings.Contains(err.Error(), "not safe to run badblocks") {
		log.Debugf(ctx, "Check fail due system find USB-drive used by it. Let's retry!")
		metrics.DefaultActionAddObservations(ctx, metrics.NewStringObservation("usbkey_audit_restarted", "yes"))
		// Sometime it happening, so we can retry.
		time.Sleep(2)
		out, err = run(ctx, timeout, command)
	}
	// Register error for following analysis.
	if err != nil {
		metrics.DefaultActionAddObservations(ctx, metrics.NewStringObservation("usbkey_audit_failure", err.Error()))
	}
	switch {
	case err == nil:
		if isSmartDevice {
			if strings.Contains(out, smartPass) {
				return tlw.HardwareState_HARDWARE_NORMAL, nil
			}
			return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
		}
		// TODO(vkjoshi@): recheck if this is required, or does stderr need to be examined.
		if len(out) > 0 {
			return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
		}
		return tlw.HardwareState_HARDWARE_NORMAL, nil
	case strings.Contains(err.Error(), "not safe to run badblocks"):
		log.Debugf(ctx, "Check fail due system find USB-drive used by it! Skip as something stramge with this DUT.")
		fallthrough
	case components.SSHErrorLinuxTimeout.In(err): // 124 timeout
		fallthrough
	case components.SSHErrorCLINotFound.In(err): // 127 badblocks
		return tlw.HardwareState_HARDWARE_UNSPECIFIED, errors.Annotate(err, "run check on host: could not successfully complete check").Err()
	default:
		return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
	}
}
