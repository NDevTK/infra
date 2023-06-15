// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package controller defines a common router controller class and
// implementations of this class per router type.
package controller

import (
	"context"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/tlw"
)

// IdentifyRouterDeviceType probes the router host over ssh to determine its
// device type.
func IdentifyRouterDeviceType(ctx context.Context, sshAccess ssh.Access, resource string) (labapi.WifiRouterDeviceType, error) {
	if sshAccess == nil {
		return 0, errors.Reason("sshAccess must not be nil").Err()
	}

	sshRunner := newRouterSshRunner(sshAccess, resource)
	isChromeOSGale, err := hostIsChromeOSGaleRouter(ctx, sshRunner)
	if err != nil {
		return 0, errors.Annotate(err, "failed to check if host has the device type of %s", labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_CHROMEOS_GALE).Err()
	}
	if isChromeOSGale {
		return labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_CHROMEOS_GALE, nil
	}

	return labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_INVALID, nil
}

// RouterController is the common interface that all router controllers must
// implement so that router execs that are not dependent upon the device type
// may rely just on RouterController for usage and NewRouterDeviceController for
// implementation selection.
type RouterController interface {
	// WifiRouterHost returns the corresponding tlw.WifiRouterHost instance for
	// this router. Changes to this instance are persisted across execs.
	WifiRouterHost() *tlw.WifiRouterHost

	// DeviceType returns the labapi.WifiRouterDeviceType of the router.
	DeviceType() labapi.WifiRouterDeviceType

	// Model returns a unique name for the router model.
	Model() (string, error)

	// Features returns the router features that this router supports.
	Features() ([]labapi.WifiRouterFeature, error)

	// Reboot will reboot the router and wait for it to come back up. A non-nil
	// error indicates that the router was rebooted and is ssh-able again.
	Reboot(ctx context.Context) error
}

// NewRouterDeviceController creates a new router controller instance for the
// specified router host. The controller implementation used is dependent upon
// the wifiRouter.DeviceType, so this must be populated.
func NewRouterDeviceController(sshAccess ssh.Access, resource string, wifiRouter *tlw.WifiRouterHost) (RouterController, error) {
	if sshAccess == nil {
		return nil, errors.Reason("sshAccess must not be nil").Err()
	}
	if wifiRouter == nil {
		return nil, errors.Reason("wifiRouter must not be nil").Err()
	}
	sshRunner := newRouterSshRunner(sshAccess, resource)
	switch wifiRouter.DeviceType {
	case labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_UNKNOWN:
		return nil, errors.Reason("cannot control unknown router; it must be analyzed first").Err()
	case labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_INVALID:
		return nil, errors.Reason("cannot control invalid router").Err()
	case labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_CHROMEOS_GALE:
		return newChromeOSGaleRouterController(sshRunner, wifiRouter), nil
	}
	return nil, errors.Reason("unsupported DeviceType %q", wifiRouter.DeviceType).Err()
}
