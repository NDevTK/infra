// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package tlw provides an abstract representation of the TLW API which is used by recovery
package tlw

import (
	"context"

	"infra/cros/dutstate"
)

// Access represent TLW level to access to the devices and inventory.
// Each device in the lab is representing as resource with name.
// For now the resource name matche to host-name but later can became different.
// Examples:
// 	Hostname: lab1-row1-rack1-device1, lab1-row1-rack1-ap1
// 	Resource Name: TestDevice256, CustomApV3.0
type Access interface {
	// Ping performs ping by resource name.
	Ping(ctx context.Context, resourceName string, count int) error
	// Run executes command on device by SSH related to resource name.
	Run(ctx context.Context, req *RunRequest) *RunResult
	// InitServod initiates servod daemon on servo-host.
	InitServod(ctx context.Context, req *InitServodRequest) error
	// StopServod stops servod daemon on servo-host.
	StopServod(ctx context.Context, resourceName string) error
	// CallServod executes a command on servod related to resource name.
	// Commands will be run against servod on servo-host.
	CallServod(ctx context.Context, req *CallServodRequest) *CallServodResponse
	// Commands will be run against servod on servo-host.
	CallBluetoothPeer(ctx context.Context, req *CallBluetoothPeerRequest) *CallBluetoothPeerResponse
	// CopyFileTo copies file to remote device from local.
	CopyFileTo(ctx context.Context, req *CopyRequest) error
	// CopyFileFrom copies file from remote device to local.
	CopyFileFrom(ctx context.Context, req *CopyRequest) error
	// CopyDirectoryTo copies directory to remote device from local, recursively.
	CopyDirectoryTo(ctx context.Context, req *CopyRequest) error
	// CopyDirectoryFrom copies directory from remote device to local, recursively.
	CopyDirectoryFrom(ctx context.Context, req *CopyRequest) error
	// RunRPMAction performs power action on RPM outlet per request.
	RunRPMAction(ctx context.Context, req *RunRPMActionRequest) error
	// ListResourcesForUnit provides list of resources names related to target unit.
	// All test and task scheduling against the target unit which can link to 1 or more resources.
	ListResourcesForUnit(ctx context.Context, unitName string) ([]string, error)
	// GetDut provides DUT info per requested resource name from inventory.
	GetDut(ctx context.Context, resourceName string) (*Dut, error)
	// UpdateDut updates DUT info into inventory.
	UpdateDut(ctx context.Context, dut *Dut) error
	// Version provides versions for requested device and type of versions.
	Version(ctx context.Context, req *VersionRequest) (*VersionResponse, error)
	// GetCacheUrl provides URL to download requested path to file.
	// URL will use to download image to USB-drive and provisioning.
	GetCacheUrl(ctx context.Context, resourceName, filePath string) (string, error)
	// Provision triggers provisioning of the device.
	Provision(ctx context.Context, req *ProvisionRequest) error
	// Close closes all used resources.
	Close(ctx context.Context) error
}

// DUTSetupType describes different DUT setups.
type DUTSetupType string

const (
	DUTSetupTypeUnspecified DUTSetupType = "UNSPECIFIED"
	// Special setup of servo-host represented as labstation.
	DUTSetupTypeLabstation DUTSetupType = "LABSTATION"
	// Special setup for routers.
	DUTSetupTypeJetstream DUTSetupType = "JETSTREAM"
	// Special setup for ChromeOS devices.
	DUTSetupTypeCros DUTSetupType = "CROS"
	// Special setup for Android devices.
	DUTSetupTypeAndroid DUTSetupType = "ANDROID"
	// Special setup for IOS devices.
	DUTSetupTypeIOS DUTSetupType = "IOS"
)

// PowerSupplyType describes different power supply types for a DUT.
type PowerSupplyType string

const (
	PowerSupplyTypeUnspecified PowerSupplyType = "UNSPECIFIED"
	// Primary power source of the devices is wall-power. Devices does not have a battery.
	PowerSupplyTypeACOnly PowerSupplyType = "AC_ONLY"
	// Primary power source of the devices is battery. Devices still connected to wall-power to charge it.
	PowerSupplyTypeBattery PowerSupplyType = "BATTERY"
)

// Cr50Phase describes different phases of CR50 firmware used on DUT.
// Label indicating the cr50 image type.
type Cr50Phase string

const (
	Cr50PhaseUnspecified Cr50Phase = "UNSPECIFIED"
	Cr50PhasePVT         Cr50Phase = "CR50_PHASE_PVT"
	Cr50PhasePREPVT      Cr50Phase = "CR50_PHASE_PREPVT"
)

// Cr50KeyEnv describes key env for cr50 RW version.
// Label indicating the cr50 RW version.
type Cr50KeyEnv string

const (
	Cr50KeyEnvUnspecified Cr50KeyEnv = "UNSPECIFIED"
	Cr50KeyEnvProd        Cr50KeyEnv = "CR50_KEYENV_PROD"
	Cr50KeyEnvDev         Cr50KeyEnv = "CR50_KEYENV_DEV"
)

const (
	// Extra attributes for DUT to provide custom info.
	ExtraAttributePools          = "POOLS"
	ExtraAttributeServoSetup     = "SERVO_SETUP"
	ExtraAttributeServoSetupDual = "SERVO_SETUP_DUAL"
)

// Dut holds info about setup used as testbed.
type Dut struct {
	// Unique identifier in inventory.
	// Should never be changed in processes.
	Id string
	// Name is the resource name for the DUT.
	Name string
	// SetupType describes the setup of the DUT, which affects how it is verified/repaired.
	SetupType DUTSetupType
	// PowerSupplyType describes the DUT's power supply type.
	PowerSupplyType PowerSupplyType
	// Audio info.
	Audio *DUTAudio
	// State of the DUT.
	State dutstate.State

	// Peripheral devices.
	// ServoHost of the DUT setup.
	ServoHost *ServoHost
	// Peripheral Wifi state of testbed
	PeripheralWifiState PeripheralWifiState
	// ProvisionedInfo tells provisioned info for the DUT.
	ProvisionedInfo *ProvisionedInfo
	// Extra attrubes of the DUT.
	// Used to provide any other mapping data which can be used for custom actions.
	// All values has to be converted to string.
	// Example: pools, force_flashing, restrictions and special abilities.
	ExtraAttributes map[string][]string
	// Chromeos hold specific data for ChromeOS device's data.
	Chromeos *ChromeOS
	// Android hold specific data for Android device's data.
	Android *Android
}

// GetAndroid returns Android device.
// The method created to mimic proto syntactic.
func (d *Dut) GetAndroid() *Android {
	return d.Android
}

// GetChromeos returns ChromeOS device.
// The method created to mimic proto syntactic.
func (d *Dut) GetChromeos() *ChromeOS {
	return d.Chromeos
}

// PeripheralWifiState describes the state of peripheral wifi testbed.
type PeripheralWifiState string

const (
	PeripheralWifiStateUnspecified PeripheralWifiState = "UNSPECIFIED"
	// All wifirouters are working as expected.
	PeripheralWifiStateWorking PeripheralWifiState = "WORKING"
	// Not all wifirouters are working as expected.
	PeripheralWifiStateBroken PeripheralWifiState = "BROKEN"
)
