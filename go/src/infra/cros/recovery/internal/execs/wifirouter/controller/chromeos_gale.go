// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/tlw"
)

const (
	lsbReleaseFilePath    = "/etc/lsb-release"
	lsbReleaseMatchIfGale = "(?m)^CHROMEOS_RELEASE_BOARD=gale$"
)

// hostIsChromeOSGaleRouter checks if the remote host is a ChromeOS Gale router.
func hostIsChromeOSGaleRouter(ctx context.Context, sshRunner ssh.Runner) (bool, error) {
	matches, err := RemoteFileContentsMatch(ctx, sshRunner, lsbReleaseFilePath, lsbReleaseMatchIfGale)
	if err != nil {
		return false, errors.Annotate(err, "failed to check if remote file %q contents match %q", lsbReleaseFilePath, lsbReleaseMatchIfGale).Err()
	}
	return matches, nil
}

// ChromeOSGaleRouterController is the RouterController implementation for
// ChromeOS Gale router devices.
type ChromeOSGaleRouterController struct {
	sshRunner      ssh.Runner
	wifiRouterHost *tlw.WifiRouterHost
}

func newChromeOSGaleRouterController(sshRunner ssh.Runner, wifiRouterHost *tlw.WifiRouterHost) *ChromeOSGaleRouterController {
	return &ChromeOSGaleRouterController{
		wifiRouterHost: wifiRouterHost,
		sshRunner:      sshRunner,
	}
}

// WifiRouterHost returns the corresponding tlw.WifiRouterHost instance for
// this router. Changes to this instance are persisted across execs.
func (c *ChromeOSGaleRouterController) WifiRouterHost() *tlw.WifiRouterHost {
	return c.wifiRouterHost
}

// DeviceType returns the labapi.WifiRouterDeviceType of the router.
func (c *ChromeOSGaleRouterController) DeviceType() labapi.WifiRouterDeviceType {
	return labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_CHROMEOS_GALE
}

// Model returns a unique name for the router model.
//
// All ChromeOS Gale routers have the same model name, since only one model is
// supported.
func (c *ChromeOSGaleRouterController) Model() (string, error) {
	// This matches the manually set value.
	return "gale", nil
}

// Features returns the router features that this router supports.
//
// ChromeOS Gale routers have a static list of known features. This list is
// maintained by the controller and is not retrieved from the device itself.
func (c *ChromeOSGaleRouterController) Features() ([]labapi.WifiRouterFeature, error) {
	return []labapi.WifiRouterFeature{
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_A,
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_B,
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_G,
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_N,
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_IEEE_802_11_AC,
	}, nil
}

// Reboot will reboot the router and wait for it to come back up. A non-nil
// error indicates that the router was rebooted and is ssh-able again.
func (c *ChromeOSGaleRouterController) Reboot(ctx context.Context) error {
	return ssh.Reboot(ctx, c.sshRunner)
}
