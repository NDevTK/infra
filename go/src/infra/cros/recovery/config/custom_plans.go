// Copyright 2022 The Chromium OS Authors
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

// RestoreHWIDFromInventoryConfig reads the configuration from the inventory.
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

// RecoverCBIFromInventoryConfig restores backup CBI contents from UFS
func RecoverCBIFromInventoryConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanCrOS,
		},
		Plans: map[string]*Plan{
			PlanCrOS: {
				CriticalActions: []string{
					"Recover CBI With Contents From Inventory",
				},
				Actions: crosRepairActions(),
			},
		},
	}
}

func EraseMRCCacheConfig() *Configuration {
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			PlanCrOS,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			PlanServo: setAllowFail(servoRepairPlan(), false),
			PlanCrOS: {
				CriticalActions: []string{
					"Erase DUT MRC cache via servo",
				},
				Actions: crosRepairActions(),
			},
			PlanClosing: setAllowFail(crosClosePlan(), true),
		},
	}
}

// FixTPM54Config creates a custom configuration to address issue reported by b/271040435
//
// Implementatio based on b/272310645#comment33 and b/272310645#comment39.
func FixTPM54Config() *Configuration {
	customFixPlan := "cros_tpm54_fix"
	return &Configuration{
		PlanNames: []string{
			PlanServo,
			customFixPlan,
			PlanCrOS,
			PlanChameleon,
			PlanBluetoothPeer,
			PlanWifiRouter,
			PlanClosing,
		},
		Plans: map[string]*Plan{
			// Not allowed to fail as servo is critical for the fix plan.
			PlanServo: servoRepairPlan(),
			// If fix didn't work then no need to run repair plans.
			customFixPlan: {
				CriticalActions: []string{
					"Is servod running",
					"Download stable version OS image to servo usbkey if necessary (allow fail)",
					"Restore FW and reset GBB flags from USB drive",
					"Cold reset DUT by servo",
					"Sleep 60 seconds",
					"Restore from TPM 0x54 error",
					"Sleep 60 seconds",
					"Install OS in recovery mode by booting from servo USB-drive",
				},
				Actions: crosRepairActions(),
			},
			PlanCrOS:          setAllowFail(crosRepairPlan(), false),
			PlanChameleon:     setAllowFail(chameleonPlan(), true),
			PlanBluetoothPeer: setAllowFail(btpeerRepairPlan(), true),
			PlanWifiRouter:    setAllowFail(wifiRouterRepairPlan(), true),
			PlanClosing:       setAllowFail(crosClosePlan(), true),
		},
	}
}
