// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/cros/usb"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// The map of USB drive vendor:model pairings that currently support SMART.
var smartUSBs = map[string]string{
	"Corsair": "Voyager GTX",
}

func checkUSBSupportsSmart(ctx context.Context, run components.Runner, usbPath string) (bool, error) {
	usbBasename := path.Base(usbPath)
	vendorPath := fmt.Sprintf("/sys/block/%s/device/vendor", usbBasename)
	modelPath := fmt.Sprintf("/sys/block/%s/device/model", usbBasename)
	vendor, err := run(ctx, time.Minute, fmt.Sprintf("cat %s", vendorPath))
	if err != nil {
		return false, errors.Annotate(err, "get usb drive vendor on dut failed").Err()
	}
	model, err := run(ctx, time.Minute, fmt.Sprintf("cat %s", modelPath))
	if err != nil {
		return false, errors.Annotate(err, "get usb drive vendor on dut failed").Err()
	}
	v := strings.TrimSpace(vendor)
	if val, exists := smartUSBs[v]; exists {
		if strings.Contains(model, val) {
			return true, nil
		}
	}
	return false, nil
}

// getUSBDrivePathOnDut finds and returns the path of USB drive on a DUT.
func getUSBDrivePathOnDut(ctx context.Context, run components.Runner, s components.Servod) (string, error) {
	// switch USB on servo multiplexer to the DUT-side
	if err := s.Set(ctx, "image_usbkey_direction", "dut_sees_usbkey"); err != nil {
		return "", errors.Annotate(err, "get usb drive path on dut: could not switch USB to DUT").Err()
	}
	// A detection delay is required when attaching this USB drive to DUT.
	time.Sleep(5 * time.Second)
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
	timeout := actionArgs.AsDuration(ctx, "audit_timeout_min", 120, time.Minute)

	dut := info.GetDut()
	servoHost := info.GetChromeos().GetServo()
	log.Infof(ctx, "Begin servo audit USB from DUT side for %q %q", dut.Name, servoHost.GetName())

	dutRunner := info.NewRunner(dut.Name)
	dutUSB, err := getUSBDrivePathOnDut(ctx, dutRunner, info.NewServod())
	if err != nil {
		log.Errorf(ctx, "Failed to determine dut USB path: %s", err.Error())
		return errors.Annotate(err, "audit USB from DUT side").Err()
	}
	smartSupport, err := checkUSBSupportsSmart(ctx, dutRunner, dutUSB)
	if err != nil {
		log.Errorf(ctx, "Failed to determine if dut USB supports SMART: %s", err.Error())
		return errors.Annotate(err, "audit USB from DUT side").Err()
	}
	state, err := usb.RunCheckOnHost(ctx, dutRunner, dutUSB, smartSupport, timeout)
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
