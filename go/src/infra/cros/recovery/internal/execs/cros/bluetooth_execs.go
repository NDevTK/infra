// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"reflect"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	// Command to check whether the bluetooth device is powered-on and
	// responsive on system DBus. In case of successful Bluetooth
	// detection, the exit code will be 0 (success) and output string
	// will approximately be like '\s*variant\s+boolean\s+true'. In
	// case of failure, the output will either include 'false' instead
	// of 'true', or the exist code will be non-zero, and output will
	// be empty.
	bluetoothDetectionCmd = `dbus-send --print-reply --system --dest=org.bluez /org/bluez/hci0 ` +
		`org.freedesktop.DBus.Properties.Get string:"org.bluez.Adapter1" string:"Powered"`
)

// auditBluetoothExec will verify bluetooth on the host is detected correctly.
//
// Check if bluetooth on the host has been powered-on and is responding.
func auditBluetoothExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	r := args.NewRunner(args.ResourceName)
	output, err := r(ctx, bluetoothDetectionCmd)
	if err == nil {
		// dbus-send command completed with success
		lines := strings.Split(output, "\n")
		if len(lines) == 2 {
			btInfoArray := strings.Fields(lines[1])
			if reflect.DeepEqual(btInfoArray, []string{"variant", "boolean", "true"}) {
				args.DUT.Bluetooth.State = tlw.HardwareStateNormal
				return nil
			}
		}
	}
	if !args.DUT.Bluetooth.Expected {
		// If bluetooth is not detected, but was expected by setup info
		// then we set needs_replacement as it is probably a hardware issue.
		args.DUT.Bluetooth.State = tlw.HardwareStateNeedReplacement
		log.Info(ctx, "Detected problems with bluetooth on the DUT")
		return nil
	}
	// the bluetooth state cannot be determined due to cmd failed
	// therefore, set it to HardwareStateUnspecified.
	args.DUT.Bluetooth.State = tlw.HardwareStateUnspecified
	return errors.Annotate(err, "audit bluetooth: bluetooth state cannot be extracted").Err()
}

func init() {
	execs.Register("cros_audit_bluetooth", auditBluetoothExec)
}
