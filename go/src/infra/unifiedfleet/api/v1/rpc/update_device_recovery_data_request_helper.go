// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This file enriches UpdateDeviceRecoveryDataRequest to maintains backward and
// forward compatibility to support the legacy and new message format.
// TODO(b/236170648): Remove this file after integration of Android devices
// with PARIS is completed.

package ufspb

import (
	"go.chromium.org/chromiumos/config/go/test/lab/api"

	lab "infra/unifiedfleet/api/v1/models/chromeos/lab"
)

type DutRecoveryDataInterface interface {
	GetSerialNumber() string
	GetHwID() string
	GetDeviceSku() string
}

type PeripheralInterface interface {
	GetHostname() string
	GetState() lab.PeripheralState
}

type LabDataInterface interface {
	GetServoType() string
	GetServoTopology() *lab.ServoTopology
	GetServoUsbDrive() *api.UsbDrive
	GetSmartUsbhub() bool
}

func GetDutId(r *UpdateDeviceRecoveryDataRequest) string {
	dutId := r.GetDeviceId()
	if dutId == "" {
		return r.GetChromeosDeviceId()
	}
	return dutId
}

func GetChromeOsDutState(r *UpdateDeviceRecoveryDataRequest) *lab.DutState {
	switch r.GetResourceType() {
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE:
		return r.GetChromeos().GetDutState()
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_UNSPECIFIED:
		// Legacy case
		return r.GetDutState()
	}
	return nil
}

func GetDutRecoveryData(r *UpdateDeviceRecoveryDataRequest) DutRecoveryDataInterface {
	switch r.GetResourceType() {
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE:
		return r.GetChromeos().GetDutData()
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_UNSPECIFIED:
		// Legacy case
		return r.GetDutData()
	}
	return nil
}

func GetLabData(r *UpdateDeviceRecoveryDataRequest) LabDataInterface {
	switch r.GetResourceType() {
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_CHROMEOS_DEVICE:
		return r.GetChromeos().GetLabData()
	case UpdateDeviceRecoveryDataRequest_RESOURCE_TYPE_UNSPECIFIED:
		// Legacy case
		return r.GetLabData()
	}
	return nil
}

func GetWifiRouters(labData LabDataInterface) []PeripheralInterface {
	var s []PeripheralInterface
	switch v := labData.(type) {
	case *ChromeOsRecoveryData_LabData:
		s = make([]PeripheralInterface, len(v.GetWifiRouters()))
		for i, router := range v.GetWifiRouters() {
			s[i] = router
		}
	case *UpdateDeviceRecoveryDataRequest_LabData:
		s = make([]PeripheralInterface, len(v.GetWifiRouters()))
		for i, router := range v.GetWifiRouters() {
			s[i] = router
		}
	}
	return s
}

func GetBluetoothPeers(labData LabDataInterface) []PeripheralInterface {
	var s []PeripheralInterface
	switch v := labData.(type) {
	case *ChromeOsRecoveryData_LabData:
		s = make([]PeripheralInterface, len(v.GetBluetoothPeers()))
		for i, peer := range v.GetBluetoothPeers() {
			s[i] = peer
		}
	case *UpdateDeviceRecoveryDataRequest_LabData:
		s = make([]PeripheralInterface, len(v.GetBlueoothPeers()))
		for i, peer := range v.GetBlueoothPeers() {
			s[i] = peer
		}
	}
	return s
}
