// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/localtlw/servod"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// Time between an usb disk plugged-in and detected in the system.
const usbDetectionDelay = 5

// The prefix of the badblocks command for verifying USB
// drives. The USB-drive path will be attached to it when
// badblocks needs to be executed on a drive.
const badBlocksCommandPrefix = "badblocks -w -e 300 -b 4096 -t random %s"

func runCheckOnHost(ctx context.Context, run execs.Runner, usbPath string, timeout time.Duration) (tlw.HardwareState, error) {
	command := fmt.Sprintf(badBlocksCommandPrefix, usbPath)
	log.Debugf(ctx, "Run Check On Host: Executing %q", command)
	// The execution timeout for this audit job is configured at the
	// level of the action. So the execution of this command will be
	// bound by that.
	out, err := run(ctx, timeout, command)
	switch {
	case err == nil:
		// TODO(vkjoshi@): recheck if this is required, or does stderr need to be examined.
		if len(out) > 0 {
			return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
		}
		return tlw.HardwareState_HARDWARE_NORMAL, nil
	case execs.SSHErrorLinuxTimeout.In(err): // 124 timeout
		fallthrough
	case execs.SSHErrorCLINotFound.In(err): // 127 badblocks
		return tlw.HardwareState_HARDWARE_UNSPECIFIED, errors.Annotate(err, "run check on host: could not successfully complete check").Err()
	default:
		return tlw.HardwareState_HARDWARE_NEED_REPLACEMENT, nil
	}
}

// getUSBDrivePathOnDut finds and returns the path of USB drive on a DUT.
func getUSBDrivePathOnDut(ctx context.Context, run components.Runner, s components.Servod) (string, error) {
	// switch USB on servo multiplexer to the DUT-side
	if err := s.Set(ctx, servod.ImageUsbkeyDirection, servod.ImageUsbkeyTowardsDUT); err != nil {
		return "", errors.Annotate(err, "get usb drive path on dut: could not switch USB to DUT").Err()
	}
	// A detection delay is required when attaching this USB drive to DUT
	time.Sleep(usbDetectionDelay * time.Second)
	if out, err := run(ctx, time.Minute, "ls /dev/sd[a-z]"); err != nil {
		return "", errors.Annotate(err, "get usb drive path on dut").Err()
	} else {
		for _, p := range strings.Split(out, "\n") {
			dtOut, dtErr := run(ctx, time.Minute, fmt.Sprintf(". /usr/share/misc/chromeos-common.sh; get_device_type %s", p))
			if dtErr != nil {
				return "", errors.Annotate(dtErr, "get usb drive path on dut: could not check %q", p).Err()
			}
			if dtOut == "USB" {
				if _, fErr := run(ctx, time.Minute, fmt.Sprintf("fdisk -l %s", p)); fErr == nil {
					return p, nil
				} else {
					log.Debugf(ctx, "Get USB-drive path on dut: checked candidate usb drive path %q and found it incorrect.", p)
				}
			}
		}
		log.Debugf(ctx, "Get USB-drive path on dut: did not find any valid USB drive path on the DUT.")
	}
	return "", errors.Reason("get usb drive path on dut: did not find any USB Drive connected to the DUT as we checked that DUT is up").Err()
}

// auditUSBFromDUTSideKeyExec initiates an audit of the servo USB key strictly from the DUT side.
func auditUSBFromDUTSideKeyExec(ctx context.Context, info *execs.ExecInfo) error {
	actionArgs := info.GetActionArgs(ctx)
	timeout := actionArgs.AsDuration(ctx, "audit_timeout", 2, time.Hour)

	dut := info.RunArgs.DUT
	servoHost := info.GetChromeos().GetServo()
	log.Infof(ctx, "Begin servo audit USB from DUT side for %q %q", dut.Name, servoHost.GetName())

	dutRunner := info.NewRunner(dut.Name)
	dutUSB, err := getUSBDrivePathOnDut(ctx, dutRunner, info.NewServod())
	if err != nil {
		log.Errorf(ctx, "Failed to determine dut USB path: %s", err.Error())
		return errors.Annotate(err, "audit USB from DUT side").Err()
	}
	state, err := runCheckOnHost(ctx, dutRunner, dutUSB, timeout)
	if err != nil {
		log.Errorf(ctx, "DUT check failed")
		return errors.Reason("audit USB from DUT side: could not check DUT usb path %q", dutUSB).Err()
	}
	servoHost.UsbkeyState = state
	log.Infof(ctx, "Successfully end servo audit USB from DUT side for %q %q", dut.Name, servoHost.GetName())
	return nil
}

func init() {
	execs.Register("audit_usb_from_dut_side", auditUSBFromDUTSideKeyExec)
}
