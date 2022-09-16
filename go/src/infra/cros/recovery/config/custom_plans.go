// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package config

import (
	"fmt"
)

// DownloadImageToServoUSBDrive creates configuration to download image to USB-drive connected to the servo.
func DownloadImageToServoUSBDrive(gsImagePath, imageName string) *Configuration {
	rc := CrosRepairConfig()
	rc.PlanNames = []string{
		PlanServo,
		PlanCrOS,
	}
	// Servo plan is not critical we just care to start servod.
	rc.Plans[PlanServo].AllowFail = true
	// Remove closing plan as we do not collect any logs or update states kin spacial ways.
	delete(rc.Plans, PlanClosing)
	cp := rc.Plans[PlanCrOS]
	const targetAction = "Download stable image to USB-key"
	cp.CriticalActions = []string{targetAction}
	var newArgs []string
	if gsImagePath != "" {
		newArgs = append(newArgs, fmt.Sprintf("os_image_path:%s", gsImagePath))
	} else if imageName != "" {
		newArgs = append(newArgs, fmt.Sprintf("os_name:%s", imageName))
	}
	cp.GetActions()[targetAction].ExecExtraArgs = newArgs
	return rc
}

// ReserveDutConfig creates configuration to reserve a dut
func ReserveDutConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"dut_state_reserved",
				},
			},
		},
	}
}

// Restore HWID from inventory.
func RestoreHWIDFromInventoryConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"dut_has_hwid",
					"cros_ssh",
					"Disable software-controlled write-protect for 'host'",
					"Disable software-controlled write-protect for 'ec'",
					"cros_update_hwid_from_inventory_to_host",
					"Simple reboot",
					"Sleep 1s",
					"Wait to be SSHable (normal boot)",
					"cros_match_hwid_to_inventory",
				},
				Actions: crosRepairActions(),
			},
		},
	}
}
