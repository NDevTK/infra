// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package servo

import (
	"context"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/cros"
	"infra/cros/recovery/internal/log"
)

const (
	// mac address command
	macAddressServoCmd = "macaddr"
	// base/start path for the HUB in order to find the single usb device dir.
	hubBasePath = "/sys/bus/usb/devices/"
)

// servoAuditNICMacAddressExec retrieve and audits the servo NIC MAC address by comparing the mac address
// cached in the servo side and the mac address from the DUT. In addition, it will attempt to
// update the value of servo command "macaddr" to the mac address from DUT if these two are not equal.
func servoAuditNICMacAddressExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	// This part confirms the path to NIC is located in the path to HUB.
	// eg.
	// HUB: /sys/bus/usb/devices/1-1
	// NIC: /sys/bus/usb/devices/1-1.1
	r := args.NewRunner(args.ResourceName)
	hubPath, err := cros.FindSingleUsbDeviceFSDir(ctx, r, hubBasePath, cros.SERVO_DUT_HUB_VID, cros.SERVO_DUT_HUB_PID)
	if err != nil {
		return errors.Annotate(err, "servo audit nic mac address").Err()
	}
	log.Debug(ctx, "Path to the servo HUB device: %s", hubPath)
	nicPath, err := cros.FindSingleUsbDeviceFSDir(ctx, r, hubPath, cros.SERVO_DUT_NIC_VID, cros.SERVO_DUT_NIC_PID)
	if err != nil {
		return errors.Annotate(err, "servo audit nic mac address").Err()
	}
	log.Debug(ctx, "Path to the servo NIC device: %s", nicPath)
	if hubPath == nicPath || !strings.HasPrefix(nicPath, hubPath) {
		return errors.Reason("servo audit nic mac address: the servo nic path was detect out of servo hub path").Err()
	}
	// get mac address from the DUT.
	macAddressFromDUT, err := cros.ServoNICMacAddress(ctx, r, nicPath)
	if err != nil {
		return errors.Annotate(err, "servo audit nic mac address").Err()
	}
	// get the mac address from the servo cache.
	// TODO: to use ServodGetString help function.
	res, err := ServodCallGet(ctx, args, macAddressServoCmd)
	if err != nil {
		return errors.Annotate(err, "servo audit nic mac address").Err()
	}
	cachedMacAddressFromServo := res.Value.GetString_()
	if cachedMacAddressFromServo == "" || cachedMacAddressFromServo != macAddressFromDUT {
		if _, err := ServodCallSet(ctx, args, macAddressServoCmd, macAddressFromDUT); err != nil {
			log.Debug(ctx, `Fail to update "macaddr" to value: %s`, macAddressFromDUT)
			return errors.Annotate(err, "servo audit nic mac address").Err()
		}
		log.Info(ctx, `Successfully updated the servo "macaddr" to be: %s`, macAddressFromDUT)
		return nil
	}
	log.Info(ctx, `The servo "macaddr" does not need update.`)
	return nil
}

func init() {
	execs.Register("servo_audit_nic_mac_address", servoAuditNICMacAddressExec)
}
